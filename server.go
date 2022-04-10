package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ackersonde/ackerson.de-go/baseball"
	"github.com/ackersonde/ackerson.de-go/structures"
	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/mux"
	"github.com/mssola/user_agent"
	"github.com/shurcooL/httpgzip"
	"github.com/urfave/negroni"
)

var httpPort = ":8080"
var darksky string
var version string
var post = "POST"

var tmpl = packr.New("templates", "./templates")
var static = packr.New("static", "./public")
var root = template.New("root")

func parseHTMLTemplateFiles() {
	// Go thru ./templates dir and load them for rendering
	for _, path := range tmpl.List() {
		//files = append(files, path)

		bytes, err := tmpl.Find(path)
		if err != nil || len(bytes) == 0 {
			log.Printf("Couldn't find %s: %s", path, err.Error())
		}

		// most important line in the whole goddamn program :(
		path := strings.TrimSuffix(path, ".tmpl")
		t, err2 := template.New(path).Parse(string(bytes))
		if err2 != nil {
			log.Printf("WTF? %s", err.Error())
		}
		root, err = root.AddParseTree(path, t.Tree)

		if err != nil {
			panic("OHH NOEESSS: " + err.Error())
		}
	}
}

func getHTTPPort() string {
	return httpPort
}

func main() {
	parseEnvVariables()
	parseHTMLTemplateFiles()

	r := mux.NewRouter()
	setUpRoutes(r)
	n := negroni.Classic()
	n.UseHandler(r)

	http.ListenAndServe(httpPort, n)
}

func parseEnvVariables() {
	darksky = os.Getenv("DARKSKY_API_KEY")
	version = os.Getenv("GITHUB_RUN_ID")
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
		WhoAmIHandler(w, r)
	})
	router.HandleFunc("/ip", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.Header.Get("X-Forwarded-For")))
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

	// favTeamGameListing shows all games of selected team for last 30 days
	router.HandleFunc("/bbFavoriteTeam", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		favTeamGameListing := baseball.FavoriteTeamGameListHandler(id, homePageMap)

		w.Header().Set("Cache-Control", "max-age=10800")

		teamID, _ := strconv.Atoi(id)
		favTeam := homePageMap[teamID]

		ua := user_agent.New(r.UserAgent())
		if ua.Mobile() {
			root.ExecuteTemplate(w, "bbFavoriteTeamGameListMobile", FavGames{FavGamesList: favTeamGameListing, FavTeam: favTeam})
		} else {
			root.ExecuteTemplate(w, "bbFavoriteTeamGameList", FavGames{FavGamesList: favTeamGameListing, FavTeam: favTeam})
		}
	})

	// gameDayListing for yesterday (default 'homepage')
	router.HandleFunc("/bb", bbHome)

	// ajax request for gameDayListing
	router.HandleFunc("/bbAjaxDay", bbAjaxDay)

	// play a single game
	router.HandleFunc("/bbStream", bbStream)

	// play all games of the day
	router.HandleFunc("/bbAll", bbAll)

	// catch all static file requests
	router.HandleFunc("/", handleIndex)
	router.PathPrefix("/").Handler(httpgzip.FileServer(
		static,
		httpgzip.FileServerOptions{
			IndexHTML: true,
		}))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	indexPage, _ := static.Find("index.html")
	w.Header().Set("Cache-Control", "max-age=30800")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexPage)
}

// FavGames is now commented
type FavGames struct {
	FavGamesList []baseball.GameDay
	FavTeam      baseball.Team
}

var homePageMap map[int]baseball.Team

func bbHome(w http.ResponseWriter, r *http.Request) {
	date1 := r.URL.Query().Get("date1")
	offset := r.URL.Query().Get("offset")

	if date1 == "" {
		/* for out of season, display Day 1 of last World Series
		// from https://en.wikipedia.org/wiki/2022_Major_League_Baseball_season
		date1 = "year_2021/month_10/day_30" // Day 1 of World Series 2021
		offset = "0"
		*/

		// this is for regular season operations
		year, month, day := time.Now().Date()
		date1 = "year_" + strconv.Itoa(year) + "/month_" +
			strconv.Itoa(int(month)) + "/day_" + strconv.Itoa(day)
		offset = "-1"
	}
	gameDayListing := baseball.GameDayListingHandler(date1, offset, homePageMap)

	w.Header().Set("Cache-Control", "max-age=10800")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	ua := user_agent.New(r.UserAgent())
	if ua.Mobile() {
		root.ExecuteTemplate(w, "bbGameDayListingMobile", gameDayListing)
	} else {
		root.ExecuteTemplate(w, "bbGameDayListing", gameDayListing)
	}
}

func bbStream(w http.ResponseWriter, r *http.Request) {
	URL := r.URL.Query().Get("url")
	log.Print("render BB Game URL: " + URL)

	if (strings.HasPrefix(URL, "https://mlb-cuts-diamond.mlb.com/FORGE/") ||
		strings.HasPrefix(URL, "https://mediadownloads.mlb.com")) &&
		strings.HasSuffix(URL, ".mp4") {
		// take care and check that the incoming URL is what we expect it to be
		root.ExecuteTemplate(w, "bbPlaySingleGameOfDay", URL)
	} else {
		// else send them on their merry way
		http.Redirect(w, r, URL, http.StatusBadRequest)
	}
}

func bbAll(w http.ResponseWriter, r *http.Request) {
	date1 := r.URL.Query().Get("date1")
	// time for GottaCatchEmAll isn't formatted how we expect
	date1Formatted, _ := time.Parse("01/02/2006", date1)
	date1 = date1Formatted.Format("year_2006/month_01/day_02")

	offset := r.URL.Query().Get("offset")
	allGames := baseball.PlayAllGamesOfDayHandler(date1, offset, homePageMap)

	// prepare response page
	w.Header().Set("Cache-Control", "max-age=10800")
	w.Header().Set("Content-Type", "text/html;charset=utf-8")
	root.ExecuteTemplate(w, "bbPlayAllGamesOfDay", allGames)
}

func bbAjaxDay(w http.ResponseWriter, r *http.Request) {
	date1 := r.URL.Query().Get("date1")
	offset := r.URL.Query().Get("offset")
	gameDayListing := baseball.GameDayListingHandler(date1, offset, homePageMap)

	// prepare response page
	w.Header().Set("Cache-Control", "max-age=10800")
	w.Header().Set("Content-Type", "text/html;charset=utf-8")

	ua := user_agent.New(r.UserAgent())
	if ua.Mobile() {
		root.ExecuteTemplate(w, "bbGameDayListingMobile", gameDayListing)
	} else {
		root.ExecuteTemplate(w, "bbGameDayListing", gameDayListing)
	}
}

// WhoAmIHandler now commented
func WhoAmIHandler(w http.ResponseWriter, req *http.Request) {
	s := []string{"[[g;#FFFF00;]Your Browser:] " + req.UserAgent(),
		"[[g;#FFFF00;]Your IP:] " + req.Header.Get("X-Forwarded-For")}
	rawData := strings.Join(s, "\r\n")

	rawDataJSON := map[string]string{"whoami": rawData}
	data, _ := json.Marshal(rawDataJSON)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	w.Write(data)
}

// VersionHandler now commented
func VersionHandler(w http.ResponseWriter, req *http.Request) {
	if strings.HasPrefix(version, "vg") {
		version = strings.TrimLeft(version, "vg")
	}
	buildURL := "https://github.com/ackersonde/ackerson.de-go/actions/runs/" + version
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
