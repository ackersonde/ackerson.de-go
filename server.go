package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/danackerson/ackerson.de-go/baseball"
	"github.com/danackerson/ackerson.de-go/structures"
	"github.com/goincremental/negroni-sessions"
	"github.com/goincremental/negroni-sessions/cookiestore"
	"github.com/otium/ytdl"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"golang.org/x/net/http2"
)

var port80only = false
var prodSession = false
var httpPort = ":8080"
var httpsPort = ":8443"
var gameDownloadDir = "/app/public/bb_games/"

func getHTTPPort() string {
	return httpPort
}

func getHTTPSPort() string {
	return httpsPort
}

func main() {
	parseEnvVariables()

	mux := http.NewServeMux()
	setUpMuxHandlers(mux)
	n := negroni.Classic()

	store := cookiestore.New([]byte(secret))
	n.Use(sessions.Sessions("gurkherpaderp", store))
	n.UseHandler(mux)

	sslCertPath := "/root/certs/"
	if !prodSession {
		sslCertPath = ""
	}

	if _, certErr := os.Stat(sslCertPath + "server.pem"); os.IsNotExist(certErr) {
		port80only = true
		http.ListenAndServe(httpPort, n)
	} else {

		// keep an ear on the http port and fwd accordingly
		go func() {
			errHTTP := http.ListenAndServe(httpPort, http.HandlerFunc(redirectToHTTPS))
			if errHTTP != nil {
				log.Fatal("Web server (HTTP): ", errHTTP)
			}
		}()
	}
	if !port80only {
		// HTTP2
		srv := &http.Server{
			Addr:    httpsPort,
			Handler: n,
		}
		http2.ConfigureServer(srv, &http2.Server{})
		log.Fatal(srv.ListenAndServeTLS(sslCertPath+"server.pem", sslCertPath+"server.key"))
	}
}

// redirectToHttps now commented
func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	if prodSession { // meaning we are running behind a docker container fwd'ing to 443
		http.Redirect(w, r, "https://ackerson.de"+r.RequestURI, http.StatusMovedPermanently)
	} else {
		http.Redirect(w, r, "https://localhost"+httpsPort+r.RequestURI, http.StatusMovedPermanently)
	}
}

var mongo string
var secret string
var joinAPIKey string
var poem string
var wunderground string
var version string
var port string

func parseEnvVariables() {
	secret = os.Getenv("ackSecret")
	joinAPIKey = os.Getenv("joinAPIKey")
	wunderground = os.Getenv("ackWunder")
	version = os.Getenv("CIRCLE_BUILD_NUM")
	prodSession, _ = strconv.ParseBool(os.Getenv("prodSession"))
}

func setUpMuxHandlers(mux *http.ServeMux) {
	post := "POST"
	homePageMap = baseball.InitHomePageMap()

	// handlers
	mux.HandleFunc("/date", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == post {
			DateHandler(w, r)
		}
	})
	mux.HandleFunc("/whoami", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == post {
			WhoAmIHandler(w, r)
		}
	})
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == post {
			VersionHandler(w, r)
		}
	})
	mux.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			WeatherHandler(w, r)
		}
	})
	mux.HandleFunc("/poems", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/bbFavoriteTeam", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/bb", bbHome)

	// ajax request for gameDayListing
	mux.HandleFunc("/bbAjaxDay", bbAjaxDay)

	// play a single game
	mux.HandleFunc("/bbStream", bbStream)

	// play all games of the day
	mux.HandleFunc("/bbAll", bbAll)

	// download gameURL to ~/bb_games (& eventually send to Join Push App)
	mux.HandleFunc("/bb_download", bbDownloadPush)

	mux.HandleFunc("/bb_download_status", bbDownloadStatus)

	icon := "https://ackerson.de/images/baseballSmall.png"
	smallIcon := "https://connect.baseball.trackman.com/Images/spinner.png"
	mux.HandleFunc("/bb_resend_join_push", func(w http.ResponseWriter, r *http.Request) {
		response := sendPayloadToJoinAPI(r.URL.Query().Get("title"), icon, smallIcon)

		w.Write([]byte(response))
	})
}

// FavGames is now commented
type FavGames struct {
	FavGamesList []baseball.GameDay
	FavTeam      baseball.Team
}

var homePageMap map[int]baseball.Team

func bbDownloadStatus(w http.ResponseWriter, req *http.Request) {
	var size int64

	title := req.URL.Query().Get("title")

	filepath := gameDownloadDir + title
	file, err := os.Open(filepath)
	if err != nil {
		log.Printf("%s", err)
	}
	fi, err := file.Stat()
	if err != nil {
		log.Printf("%s", err)
		size = -10
	} else {
		size = fi.Size()
	}
	v := map[string]int64{"size": size}

	data, _ := json.Marshal(v)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	w.Write(data)
}

func bbDownloadPush(w http.ResponseWriter, r *http.Request) {
	gameTitle := r.URL.Query().Get("gameTitle")
	gameURL := r.URL.Query().Get("gameURL")
	fileType := r.URL.Query().Get("fileType")
	var gameLength int64
	icon := "https://ackerson.de/images/baseballSmall.png"
	smallIcon := "https://connect.baseball.trackman.com/Images/spinner.png"

	downloadFilename := ""
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

		res, err := http.Head(gameURL)
		if err != nil {
			log.Printf("ERR: unable to find game size")
		}
		gameLength = res.ContentLength

		render := render.New(render.Options{IsDevelopment: true})
		render.HTML(w, http.StatusOK, "bbDownloadGameAndPushPhone",
			GameMeta{
				GameTitle:         awayTeam.Name + "@" + homeTeam.Name,
				GameDownloadTitle: downloadFilename,
				GameLength:        res.ContentLength,
				GameDate:          humanDate,
				GameFile:          downloadFilename,
			})
	} else {
		// will be a YouTube video
		vid, _ := ytdl.GetVideoInfo(gameURL)
		URI, _ := vid.GetDownloadURL(vid.Formats[0])
		log.Println(URI.String())
		res, err := http.Head(URI.String())
		if err != nil {
			log.Printf("ERR: unable to find game size")
		} else {
			log.Println(strconv.FormatInt(res.ContentLength, 10) + " bytes")
		}
		icon = "https://emoji.slack-edge.com/T092UA8PR/youtube/a9a89483b7536f8a.png"
		smallIcon = "http://icons.iconarchive.com/icons/iconsmind/outline/16/Youtube-icon.png"
		gameLength = res.ContentLength
		downloadFilename = vid.Title + "." + vid.Formats[0].Extension
		gameURL = URI.String()
	}

	// and download it to ~/bb_games/
	filepath := gameDownloadDir + downloadFilename

	go func() {
		err := downloadFile(filepath, gameURL, gameLength)
		if err != nil {
			// Check if file was already downloaded & don't resend to Join!
			if err.Error() != "file exists" {
				log.Printf("ERR: unable to download & save file %v\n", err)
			}
		} else {
			log.Printf("Finished downloading %s\n", filepath)
			sendPayloadToJoinAPI(downloadFilename, icon, smallIcon)
		}
	}()
}

func sendPayloadToJoinAPI(downloadFilename string, icon string, smallIcon string) string {
	response := "Sorry, couldn't resend..."

	// NOW send this URL to the Join Push App API
	pushURL := "https://joinjoaomgcd.appspot.com/_ah/api/messaging/v1/sendPush"
	defaultParams := "?deviceId=007e5b72192c420d9115334d1f177c4c&icon=" + icon + "&smallicon=" + smallIcon
	fileOnPhone := "&title=" + downloadFilename
	fileURL := "&file=https://ackerson.de/bb_games/" + downloadFilename
	apiKey := "&apikey=" + joinAPIKey

	completeURL := pushURL + defaultParams + fileOnPhone + fileURL + apiKey
	// Get the data
	log.Printf("joinPushURL: %s\n", completeURL)
	resp, err := http.Get(completeURL)
	if err != nil {
		log.Printf("ERR: unable to call Join Push")
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		log.Printf("successfully sent payload to Join!")
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

		// Writer the body to file
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
		//TODO 2016 is over - make default /bb goto Game 7, 2016 World Series
		date1 = "year_2016/month_11/day_02"
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
	ip := r.Header.Get("X-Forwarded-For")

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
		log.Printf("%s: %s", header, value)
	}
	data, _ := json.Marshal(rawDataJSON)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	w.Write(data)
}

// VersionHandler now commenteds
func VersionHandler(w http.ResponseWriter, req *http.Request) {
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
		log.Printf("%s", err)
	}

	geoLocation := new(structures.JSONGeoLocationRequest)
	json.Unmarshal([]byte(body), &geoLocation)

	latString := strconv.FormatFloat(float64(geoLocation.Params.Lat), 'f', 15, 32)
	lngString := strconv.FormatFloat(float64(geoLocation.Params.Lng), 'f', 15, 32)

	// call wunderground API for Conditions & Forecast
	conditionsURI := "http://api.wunderground.com/api/" + wunderground + "/conditions/q/"
	forecastURI := "http://api.wunderground.com/api/" + wunderground + "/forecast/q/"
	locationParams := latString + "," + lngString + ".json"

	currentWeather := new(structures.CurrentWeatherConditions)
	currentWeatherResp, err := http.Get(conditionsURI + locationParams)
	if err != nil {
		log.Printf("%s", err)
	} else {
		defer currentWeatherResp.Body.Close()
		currentWeatherJSON, err2 := ioutil.ReadAll(currentWeatherResp.Body)
		if err2 != nil {
			log.Printf("%s", err2)
		}
		json.Unmarshal([]byte(currentWeatherJSON), &currentWeather)
		log.Printf("%v\n", currentWeather)
	}

	currentForecast := new(structures.CurrentWeatherForecast)
	currentForecastResp, err := http.Get(forecastURI + locationParams)
	if err != nil {
		log.Printf("%s", err)
	} else {
		defer currentForecastResp.Body.Close()
		currentForecastJSON, err := ioutil.ReadAll(currentForecastResp.Body)
		if err != nil {
			log.Printf("%s", err)
		}
		json.Unmarshal([]byte(currentForecastJSON), &currentForecast)
	}

	code := map[string]interface{}{"current": currentWeather, "forecastday": currentForecast}
	data, _ := json.Marshal(code)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}
