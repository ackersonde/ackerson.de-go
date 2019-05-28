package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/danackerson/ackerson.de-go/baseball"
	"github.com/danackerson/ackerson.de-go/structures"
	"github.com/danackerson/digitalocean/common"
	sessions "github.com/goincremental/negroni-sessions"
	"github.com/goincremental/negroni-sessions/cookiestore"
	"github.com/gorilla/mux"
	"github.com/otium/ytdl"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

var httpPort = ":8080"
var downloadDir = "bender/"

func getHTTPPort() string {
	return httpPort
}

func main() {
	parseEnvVariables()

	r := mux.NewRouter()
	setUpRoutes(r)
	n := negroni.Classic()

	store := cookiestore.New([]byte(secret))
	n.Use(sessions.Sessions("gurkherpaderp", store))
	n.UseHandler(r)
	http.ListenAndServe(httpPort, n)
}

var mongo string
var secret string
var joinAPIKey string
var poem string
var darksky string
var version string
var port string
var fileToken string
var dropboxToken string
var spacesKey, spacesSecret, spacesNamePublic string
var post = "POST"

func parseEnvVariables() {
	secret = os.Getenv("ackSecret")
	joinAPIKey = os.Getenv("joinAPIKey")
	darksky = os.Getenv("ackWunder")
	version = os.Getenv("CIRCLE_BUILD_NUM")
	fileToken = os.Getenv("FILE_TOKEN")
	dropboxToken = os.Getenv("DROPBOX_TOKEN")
	spacesKey = os.Getenv("SPACES_KEY")
	spacesSecret = os.Getenv("SPACES_SECRET")
	spacesNamePublic = os.Getenv("SPACES_NAME_PUBLIC")
}

func setUpRoutes(router *mux.Router) {
	homePageMap = baseball.InitHomePageMap()

	// handlers
	router.HandleFunc("/date", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == post {
			DateHandler(w, r)
		}
	})
	router.HandleFunc("/whoami", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == post {
			WhoAmIHandler(w, r)
		}
	})
	router.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == post {
			VersionHandler(w, r)
		}
	})
	router.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == post {
			WeatherHandler(w, r)
		}
	})
	router.HandleFunc("/poems", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			session := sessions.GetSession(r)
			pass := session.Get("pass")

			if pass == nil && r.FormValue("sesam") != poem {
				http.NotFound(w, r)
			} else if r.FormValue("sesam") == poem || pass != nil {
				session.Set("pass", "true")
			}
		}
	})

	// favTeamGameListing shows all games of selected team for last 30 days
	router.HandleFunc("/bbFavoriteTeam", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		favTeamGameListing := baseball.FavoriteTeamGameListHandler(id, homePageMap)

		w.Header().Set("Cache-Control", "max-age=10800")
		render := render.New(render.Options{
			Layout:        "content",
			IsDevelopment: false,
		})

		teamID, _ := strconv.Atoi(id)
		favTeam := homePageMap[teamID]
		render.HTML(w, http.StatusOK, "bbFavoriteTeamGameList", FavGames{FavGamesList: favTeamGameListing, FavTeam: favTeam})
	})

	// gameDayListing for yesterday (default 'homepage')
	router.HandleFunc("/bb", bbHome)

	// ajax request for gameDayListing
	router.HandleFunc("/bbAjaxDay", bbAjaxDay)

	// play a single game
	router.HandleFunc("/bbStream", bbStream)

	// play all games of the day
	router.HandleFunc("/bbAll", bbAll)

	// download gameURL to ~/downloads (& eventually send to Join Push App)
	router.HandleFunc("/bb_download", bbDownloadPush)

	router.HandleFunc("/bb_download_status", bbDownloadStatus)

	icon := "https://ackerson.de/images/baseballSmall.png"
	smallIcon := "https://connect.baseball.trackman.com/Images/spinner.png"
	router.HandleFunc("/bb_resend_join_push",
		func(w http.ResponseWriter, r *http.Request) {
			response := sendPayloadToJoinAPI(downloadDir+r.URL.Query().Get("title"),
				r.URL.Query().Get("title"), icon, smallIcon)

			w.Write([]byte(response))
		})

	router.HandleFunc("/down/{file:.*}", func(w http.ResponseWriter, r *http.Request) {
		common.DownloadFromDOSpaces(spacesNamePublic, w, r)
	})
}

// FavGames is now commented
type FavGames struct {
	FavGamesList []baseball.GameDay
	FavTeam      baseball.Team
}

var homePageMap map[int]baseball.Team

// TODO: check file @ DO Spaces
func bbDownloadStatus(w http.ResponseWriter, req *http.Request) {
	var size int64

	title := req.URL.Query().Get("title")

	filepath := downloadDir + title
	file, err := os.Open(filepath)
	if err != nil {
		log.Printf("%s\n", err)
	}
	fi, err := file.Stat()
	if err != nil {
		log.Printf("%s\n", err)
		size = -10
	} else {
		size = fi.Size()
	}
	v := map[string]int64{"size": size}

	data, _ := json.Marshal(v)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	w.Write(data)
}

// TODO: move the download dir to DO Spaces
func bbDownloadPush(w http.ResponseWriter, r *http.Request) {
	gameTitle := r.URL.Query().Get("gameTitle")
	gameURL := r.URL.Query().Get("gameURL")
	fileType := r.URL.Query().Get("fileType")
	var gameLength int64
	icon := "https://ackerson.de/images/baseballSmall.png"
	smallIcon := "https://connect.baseball.trackman.com/Images/spinner.png"

	downloadFilename := ""
	humanFilename := ""
	type GameMeta struct {
		GameTitle         string
		GameDownloadTitle string
		GameLength        int64
		GameDate          string
		GameFile          string
	}

	if fileType == "bb" {
		//gameTitle: 112-114__Wed, Nov  2 2016
		result := strings.Split(gameTitle, "__")
		gameDate, _ := time.Parse("Mon, Jan _2 2006", result[1])
		humanDate := gameDate.Format("Mon, Jan _2 2006")
		formattedDate := gameDate.Format("Mon_Jan02_2006")

		// parse out teams "<awayID>-<homeID>"
		teams := strings.Split(result[0], "-")
		awayID, _ := strconv.Atoi(teams[0])
		homeID, _ := strconv.Atoi(teams[1])
		awayTeam := homePageMap[awayID]
		homeTeam := homePageMap[homeID]

		// create download filename: "Cubs@Diamondbacks-Wed_Nov02_2017.mp4"
		downloadFilename = awayTeam.Name + "@" + homeTeam.Name + "-" + formattedDate + ".mp4"
		downloadFilename = strings.Replace(downloadFilename, " ", "_", -1)
		humanFilename = downloadFilename
		res, err := http.Head(gameURL)
		if err != nil {
			log.Printf("ERR: unable to find game size\n")
		}
		gameLength = res.ContentLength

		render := render.New(render.Options{IsDevelopment: true})
		render.HTML(w, http.StatusOK, "bbDownloadGameAndPushPhone",
			GameMeta{
				GameTitle:         awayTeam.Name + "@" + homeTeam.Name,
				GameDownloadTitle: gameURL,
				GameDate:          humanDate,
			})
	} else if fileType == "vpn" {
		icon = "http://www.setaram.com/wp-content/themes/setaram/library/images/lock.png"
		smallIcon = "http://www.setaram.com/wp-content/themes/setaram/library/images/lock.png"

		sendPayloadToJoinAPI(gameURL, gameTitle, icon, smallIcon)
		return
	} else if fileType == "dl" {
		from, err := os.Open("/app/public/downloads/" + gameURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			fi, _ := from.Stat()
			gameLength = fi.Size()

			w.Header().Set("Content-Length", strconv.FormatInt(gameLength, 10))
			w.Header().Set("Content-Disposition", `attachment; filename="`+gameTitle+`"`)
			w.Header().Set("Cache-Control", "private")
			w.Header().Set("Pragma", "private")
			w.Header().Set("Expires", "Mon, 26 Jul 1997 05:00:00 GMT")

			http.ServeContent(w, r, gameTitle, time.Now(), from)
		}
		return
	} else {
		// will be a YouTube video
		vid, _ := ytdl.GetVideoInfo(gameURL)
		URI, _ := vid.GetDownloadURL(vid.Formats[0])
		log.Println(URI.String())
		res, err := http.Head(URI.String())
		if err != nil {
			log.Printf("ERR: unable to find video size\n")
		} else {
			log.Println(strconv.FormatInt(res.ContentLength, 10) + " bytes")
		}
		icon = "https://emoji.slack-edge.com/T092UA8PR/youtube/a9a89483b7536f8a.png"
		smallIcon = "http://icons.iconarchive.com/icons/iconsmind/outline/16/Youtube-icon.png"
		gameLength = res.ContentLength
		downloadFilename = vid.ID + "." + vid.Formats[0].Extension
		humanFilename = vid.Title + "." + vid.Formats[0].Extension
		// TODO: split this download up into humanFileName and diskFileID (e.g. YouTube ID)
		gameURL = URI.String()
	}

	// and download it to ~/downloads/
	filepath := downloadDir + downloadFilename

	go func() {
		err := common.CopyFileToDOSpaces(spacesNamePublic, filepath, gameURL, gameLength)
		if err != nil {
			// Check if file was already downloaded & don't resend to Join!
			log.Printf("ERR: unable to download/save %s: %s\n", gameURL, err.Error())
		} else {
			log.Printf("Finished downloading %s\n", filepath)
			sendPayloadToJoinAPI(filepath, humanFilename, icon, smallIcon)
		}
	}()
}

func sendPayloadToJoinAPI(downloadFilename string, humanFilename string, icon string, smallIcon string) string {
	response := "Sorry, couldn't resend..."
	humanFilenameEnc := &url.URL{Path: humanFilename}
	humanFilenameEncoded := humanFilenameEnc.String()
	// NOW send this URL to the Join Push App API
	pushURL := "https://joinjoaomgcd.appspot.com/_ah/api/messaging/v1/sendPush"
	defaultParams := "?deviceId=d888b2e9a3a24a29a15178b2304a40b3&icon=" + icon + "&smallicon=" + smallIcon
	fileOnPhone := "&title=" + humanFilenameEncoded
	fileURL := spacesNamePublic + ".ams3.digitaloceanspaces.com/" + downloadFilename
	apiKey := "&apikey=" + joinAPIKey

	fileURLEnc := &url.URL{Path: fileURL}
	fileURL = fileURLEnc.String()
	completeURL := pushURL + defaultParams + apiKey + fileOnPhone + "&file=https://" + fileURL
	// Get the data
	log.Printf("joinPushURL: %s\n", completeURL)
	resp, err := http.Get(completeURL)
	if err != nil {
		log.Printf("ERR: unable to call Join Push\n")
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		log.Printf("successfully sent payload to Join!\n")
		response = "Success!"
	}

	return response
}

func downloadFile(filepath string, url string, filesize int64) (err error) {
	// check if file exists and is same size as MLB.com (meaning we already downloaded it)
	fi, err := os.Stat(filepath)
	if err != nil {
		// Create the file
		out, err := os.Create(filepath)
		if err != nil {
			return err
		}
		defer out.Close()

		// Get the data
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Write the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return err
		}
	} else {
		if fi.Size() == filesize {
			return errors.New("file exists")
		}
	}
	return nil
}

func bbHome(w http.ResponseWriter, r *http.Request) {
	date1 := r.URL.Query().Get("date1")
	offset := r.URL.Query().Get("offset")

	/*if date1 == "" {
		// from https://en.wikipedia.org/wiki/2019_Major_League_Baseball_season
		// HINT: replace year with current to figure out Game 1 date
		//2019 is over - make default /bb goto Game 1, 2019 World Series
		date1 = "year_2019/month_10/day_22"
		offset = "0"
	}*/
	gameDayListing := baseball.GameDayListingHandler(date1, offset, homePageMap)

	w.Header().Set("Cache-Control", "max-age=10800")
	render := render.New(render.Options{
		Layout:        "content",
		IsDevelopment: false,
	})

	render.HTML(w, http.StatusOK, "bbGameDayListing", gameDayListing)
}

func bbStream(w http.ResponseWriter, r *http.Request) {
	URL := r.URL.Query().Get("url")
	log.Print("render URL: " + URL)

	render := render.New(render.Options{
		IsDevelopment: false,
	})

	if strings.Contains(URL, "youtube") {
		http.Redirect(w, r, URL, http.StatusFound)
	} else {
		render.HTML(w, http.StatusOK, "bbPlaySingleGameOfDay", URL)
	}
}

func bbAll(w http.ResponseWriter, r *http.Request) {
	date1 := r.URL.Query().Get("date1")
	offset := r.URL.Query().Get("offset")
	allGames := baseball.PlayAllGamesOfDayHandler(date1, offset, homePageMap)

	// prepare response page
	w.Header().Set("Cache-Control", "max-age=10800")

	render := render.New(render.Options{
		IsDevelopment: false,
	})
	render.HTML(w, http.StatusOK, "bbPlayAllGamesOfDay", allGames)
}

func bbAjaxDay(w http.ResponseWriter, r *http.Request) {
	date1 := r.URL.Query().Get("date1")
	offset := r.URL.Query().Get("offset")
	gameDayListing := baseball.GameDayListingHandler(date1, offset, homePageMap)

	// prepare response page
	w.Header().Set("Cache-Control", "max-age=10800")
	render := render.New(render.Options{
		IsDevelopment: false,
	})

	render.HTML(w, http.StatusOK, "bbGameDayListing", gameDayListing)
}

// GetIP now commented
func GetIP(r *http.Request) string {
	requestDump, _ := httputil.DumpRequest(r, false)
	log.Printf("GetIP Request: %v\n", requestDump)

	ip := r.Header.Get("X-Real-Ip")

	if len(ip) <= 0 {
		ipRemote, _, _ := net.SplitHostPort(r.RemoteAddr)
		return ipRemote
	}

	return ip
}

// WhoAmIHandler now commented
func WhoAmIHandler(w http.ResponseWriter, req *http.Request) {
	s := []string{"[[g;#FFFF00;]Your IP:] " + GetIP(req), "[[g;#FFFF00;]Your Browser:] " + req.UserAgent()}
	rawData := strings.Join(s, "\r\n")
	rawDataJSON := map[string]string{"whoami": rawData}
	for header, value := range req.Header {
		log.Printf("%s: %s\n", header, value)
	}
	data, _ := json.Marshal(rawDataJSON)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	w.Write(data)
}

// VersionHandler now commenteds
func VersionHandler(w http.ResponseWriter, req *http.Request) {
	if strings.HasPrefix(version, "vc") {
		version = strings.TrimLeft(version, "vc")
	}
	buildURL := "https://circleci.com/gh/danackerson/ackerson.de-go/" + version
	v := map[string]string{"version": buildURL, "build": version}

	data, _ := json.Marshal(v)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	w.Write(data)
}

// DateHandler now commented
func DateHandler(w http.ResponseWriter, req *http.Request) {
	now := "[[g;#FFFF00;]" + time.Now().Format("Mon Jan _2 15:04:05 2006") + "]"
	date := map[string]string{"date": now}

	data, _ := json.Marshal(date)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	w.Write(data)
}

// WeatherHandler now commented
func WeatherHandler(w http.ResponseWriter, req *http.Request) {
	// handle JSON POST request
	//body := string(structures.TestGeoLocationPost) // in case you are testing :)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("err: %s\n", err)
	} else {
		log.Printf("body: %s\n", body)
	}

	geoLocation := new(structures.JSONGeoLocationRequest)
	json.Unmarshal([]byte(body), &geoLocation)
	log.Printf("location: %v\n", geoLocation)
	latString := strconv.FormatFloat(float64(geoLocation.Params.Lat), 'f', 15, 32)
	lngString := strconv.FormatFloat(float64(geoLocation.Params.Lng), 'f', 15, 32)

	// call DarkSky.net for Conditions & Forecast
	encodedURLParams := "?units=auto&exclude=minutely,hourly,alerts"
	locationParams := latString + "," + lngString + "/" + encodedURLParams
	conditionsURI := "https://api.darksky.net/forecast/" + darksky + "/"

	currentWeather := new(structures.CurrentWeatherConditions)
	currentWeatherResp, err := http.Get(conditionsURI + locationParams)
	if err != nil {
		log.Printf("darksky ERR: %s\n", err)
	} else {
		defer currentWeatherResp.Body.Close()
		currentWeatherJSON, err2 := ioutil.ReadAll(currentWeatherResp.Body)
		if err2 != nil {
			log.Printf("darksky ERR2: %s\n", err2)
		}
		json.Unmarshal([]byte(currentWeatherJSON), &currentWeather)
	}

	// Go thru the response and overwrite the Summary fields with "Mon, Dec 25"
	// taken from the int Time fields
	code := map[string]interface{}{
		"current":     currentWeather.Currently,
		"forecastday": currentWeather.Daily,
		"flags":       currentWeather.Flags,
	}
	data, _ := json.Marshal(code)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}
