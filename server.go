package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ackersonde/ackerson.de-go/baseball"
	"github.com/ackersonde/ackerson.de-go/structures"
	"github.com/mssola/user_agent"
	"github.com/urfave/negroni"
)

var (
	//go:embed templates/index.html templates/*.gohtml templates/layouts/bb.gohtml
	tmpl         embed.FS
	templates    map[string]*template.Template
	templatesDir = "templates"
	layoutsDir   = "templates/layouts"
	post         = "POST"
	httpPort     = ":8080"
	darksky      string
	version      string
)

func getHTTPPort() string {
	return httpPort
}

func main() {
	parseEnvVariables()
	err := parseTemplates()
	if err != nil {
		log.Fatalf("ARRGGH: %v", err)
	}

	r := http.NewServeMux()
	setUpRoutes(r)
	n := negroni.Classic()
	n.UseHandler(r)

	if err := http.ListenAndServe(httpPort, n); err != nil {
		log.Printf("HANDLER ERR: %v", err)
	}
}

func parseEnvVariables() {
	darksky = os.Getenv("DARKSKY_API_KEY")
	version = os.Getenv("GITHUB_RUN_ID")
}

func parseTemplates() error {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}
	tmplFiles, err := fs.ReadDir(tmpl, templatesDir)
	if err != nil {
		return err
	}

	for _, file := range tmplFiles {
		if file.IsDir() {
			continue
		}

		pt, err := template.ParseFS(tmpl, templatesDir+"/"+file.Name(), layoutsDir+"/bb.gohtml")
		if err != nil {
			return err
		}

		templates[file.Name()] = pt
	}
	return nil
}

func setUpRoutes(router *http.ServeMux) {
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
		w.Write([]byte(r.Header.Get("X-Forwarded-For") + "\n"))
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

	// gameDayListing for yesterday (default 'homepage')
	router.HandleFunc("/bb", bbHome)

	// ajax request for gameDayListing
	router.HandleFunc("/bbAjaxDay", bbAjaxDay)

	// play a single game
	router.HandleFunc("/bbStream", bbStream)

	// play all games of the day
	router.HandleFunc("/bbAll", bbAll)

	// favTeamGameListing shows all games of selected team for last 30 days
	router.HandleFunc("/bbFavoriteTeam", bbFavorite)

	// homepage
	router.HandleFunc("/", handleIndex)
}

func bbFavorite(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	favTeamGameListing := baseball.FavoriteTeamGameListHandler(id, homePageMap)

	w.Header().Set("Cache-Control", "max-age=10800")

	teamID, _ := strconv.Atoi(id)
	favTeam := homePageMap[teamID]

	ua := user_agent.New(r.UserAgent())
	if err := templates["bbFavoriteTeamGameList.gohtml"].Execute(w, FavGames{Mobile: ua.Mobile(), FavGamesList: favTeamGameListing, FavTeam: favTeam}); err != nil {
		log.Printf("Failed to parse bbHome page: %v", err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=30800")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	ua := user_agent.New(r.UserAgent())
	data := struct{ Mobile bool }{Mobile: ua.Mobile()}
	if err := templates["index.html"].Execute(w, data); err != nil {
		log.Printf("Failed to parse index page: %v", err)
	}
}

// FavGames is now commented
type FavGames struct {
	FavGamesList []baseball.GameDay
	FavTeam      baseball.Team
	Mobile       bool
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
		t := time.Now()
		date1 = t.Format("year_2006/month_01/day_02")
		offset = "-1"
	}

	gameDayListing := baseball.GameDayListingHandler(date1, offset, homePageMap)
	w.Header().Set("Cache-Control", "max-age=10800")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	ua := user_agent.New(r.UserAgent())
	gameDayListing.Mobile = ua.Mobile()
	if err := templates["bbGameDayListing.gohtml"].Execute(w, gameDayListing); err != nil {
		log.Printf("Failed to parse bbHome page: %v", err)
	}
}

func bbStream(w http.ResponseWriter, r *http.Request) {
	URL := r.URL.Query().Get("url")
	log.Print("render BB Game URL: " + URL)

	if strings.HasPrefix(URL, "/api/v1/game") {
		// take care and check that the incoming URL is what we expect it to be
		URL = baseball.FetchGameURLFromID(URL)
		if err := templates["bbPlaySingleGameOfDay.gohtml"].Execute(w, URL); err != nil {
			log.Printf("Failed to parse bbStream page: %v", err)
		}
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
	ua := user_agent.New(r.UserAgent())
	allGames.Mobile = ua.Mobile()
	if err := templates["bbPlayAllGamesOfDay.gohtml"].Execute(w, allGames); err != nil {
		log.Printf("Failed to parse bbAll page: %v", err)
	}
}

func bbAjaxDay(w http.ResponseWriter, r *http.Request) {
	date1 := r.URL.Query().Get("date1")
	offset := r.URL.Query().Get("offset")
	gameDayListing := baseball.GameDayListingHandler(date1, offset, homePageMap)

	// prepare response page
	w.Header().Set("Cache-Control", "max-age=10800")
	w.Header().Set("Content-Type", "text/html;charset=utf-8")

	ua := user_agent.New(r.UserAgent())
	gameDayListing.Mobile = ua.Mobile()
	if err := templates["bbGameDayListing.gohtml"].Execute(w, gameDayListing); err != nil {
		log.Printf("Failed to parse bbAjax page: %v", err)
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
