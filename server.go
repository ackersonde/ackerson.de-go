package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/clbanning/mxj"
	"github.com/codegangsta/negroni"
	"github.com/danackerson/ackerson.de-go/structures"
	"github.com/goincremental/negroni-sessions"
	"github.com/goincremental/negroni-sessions/cookiestore"
	"github.com/unrolled/render"
	"gopkg.in/mgo.v2"
)

func main() {
	mux := http.NewServeMux()
	post := "POST"
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

				PoemsHandler(w, r)
			}
		}
	})

	mux.HandleFunc("/bb", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			date := time.Now().AddDate(0, 0, -1)
			date1 := r.URL.Query().Get("date1")
			if date1 != "" {
				i, _ := strconv.Atoi(r.URL.Query().Get("offset"))
				date = date.AddDate(0, 0, i)
			}
			GameHandler(w, r, date)
		}
	})

	n := negroni.Classic()

	readInCreds()

	store := cookiestore.New([]byte(secret))
	n.Use(sessions.Sessions("gurkherpaderp", store))
	n.UseHandler(mux)
	n.Run(":" + port)
}

var mongo string
var secret string
var poem string
var wunderground string
var version string
var port string

func readInCreds() {
	mongo = os.Getenv("ackMongo")
	secret = os.Getenv("ackSecret")
	poem = os.Getenv("ackPoems")
	wunderground = os.Getenv("ackWunder")
	version = os.Getenv("CIRCLE_BUILD_NUM")
	port = os.Getenv("NEGRONI_PORT")

	if port == "" {
		port = "3001"
	}
}

func loadWritings(w http.ResponseWriter) [](structures.Writing) {
	writings := [](structures.Writing){}
	session, err := mgo.Dial(mongo)

	if err != nil {
		fmt.Fprintf(w, err.Error())
	} else {
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		c := session.DB("ackersonde").C("writings")

		iter := c.Find(nil).Iter()
		iter.All(&writings)
		session.Close()
	}

	return writings
}

// PoemsHandler now commented
func PoemsHandler(w http.ResponseWriter, req *http.Request) {
	writings := loadWritings(w)
	for _, writing := range writings {
		fmt.Fprintf(w, "%1.0f: %s", writing.ID, writing.Content)
		fmt.Fprintf(w, "\r\n")
	}
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
		panic("ioutil.ReadAll")
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

// GameHandler is now commented
func GameHandler(w http.ResponseWriter, req *http.Request, gameDate time.Time) {
	yesterday := gameDate
	// Reference date: Mon Jan 2 15:04:05 MST 2006
	dates := yesterday.Format("year_2006/month_01/day_02")

	games := make(map[int][]string)
	games = SearchMLBGames(dates, games)

	// prepare response page
	r := render.New(render.Options{
		IsDevelopment: true,
	})

	r.HTML(w, http.StatusOK, "content", GameDay{Date: dates, Games: games})
}

// GameDay is now commented
type GameDay struct {
	Date  string
	Games map[int][]string
}

// SearchMLBGames is now commented
func SearchMLBGames(dates string, games map[int][]string) map[int][]string {
	// http://gd2.mlb.com/components/game/mlb/year_2016/month_03/day_21/grid_ce.xml
	// 11 games here - 0 condensed

	// http://gd2.mlb.com/components/game/mlb/year_2016/month_03/day_21/grid_ce.xml
	// total of 13 games here - but only 1 condensed!

	// http://gd2.mlb.com/components/game/mlb/year_2016/month_03/day_30/grid_ce.xml
	// total of X games: 6 condensed
	domain := "http://gd2.mlb.com/components/game/mlb/"
	suffix := "/grid_ce.xml"
	url := domain + dates + suffix

	log.Printf(url)
	resp, err := http.Get(url)
	if err != nil {
		log.Print(err)
	}
	defer resp.Body.Close()
	xml, err := ioutil.ReadAll(resp.Body)

	// log.Printf(string(xml))

	m, err := mxj.NewMapXml(xml)

	gameInfos, err := m.ValuesForKey("game")
	if err != nil {
		log.Fatal("err:", err.Error())
		log.Printf("MLB site '%s' response empty", domain)
		games[0] = []string{"Error connecting to " + domain}
		return games
	}

	log.Printf("Found %d games", len(gameInfos))
	mediaGames := 0
	condensedGames := 0
	// now just manipulate Map entries returned as []interface{} array.
	for k, v := range gameInfos {
		gameID := ""
		aGameVal, _ := v.(map[string]interface{})
		if aGameVal["-media_state"].(string) == "media_dead" {
			continue
		} else {
			mediaGames++
		}

		// rescan looking for keys with data: Values or Value
		gm := aGameVal["game_media"].(map[string]interface{})
		hb := gm["homebase"].(map[string]interface{})
		media := hb["media"].([]interface{})
		for _, val := range media {
			aMediaVal, _ := val.(map[string]interface{})
			if aMediaVal["-type"].(string) != "condensed_game" {
				continue
			} else {
				condensedGames++
				gameID = aMediaVal["-id"].(string)
				continue
			}
		}
		// create file to hold csv data sets
		title := aGameVal["-calendar_event_id"].(string)
		awayTeamID := aGameVal["-away_team_id"].(string)
		awayAbbrev := aGameVal["-away_name_abbrev"].(string)
		homeTeamID := aGameVal["-home_team_id"].(string)
		homeAbbrev := aGameVal["-home_name_abbrev"].(string)

		if gameID != "" {
			games[k] = []string{title, awayTeamID, awayAbbrev, homeTeamID, homeAbbrev, gameID, dates}
		}
	}

	log.Println("Media games:", mediaGames)
	log.Println("Condensed games:", condensedGames)

	return games
}
