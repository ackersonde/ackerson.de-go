package main

import (
	"encoding/json"
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
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"golang.org/x/net/http2"
)

// ClockCheckHandler now commented
func ClockCheckHandler(w http.ResponseWriter, req *http.Request) {
	panelID := req.URL.Query().Get("panel")

	log.Printf("req: %s", panelID)
	render := render.New(render.Options{
		Layout:        "content",
		IsDevelopment: false,
	})

	render.HTML(w, http.StatusOK, "test", nil)
}

var prodSession = false
var httpPort = ":8080"
var httpsPort = ":8443"

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

	// keep an ear on the http port and fwd accordingly
	go func() {
		errHTTP := http.ListenAndServe(httpPort, http.HandlerFunc(redirectToHTTPS))
		if errHTTP != nil {
			log.Fatal("Web server (HTTP): ", errHTTP)
		}
	}()

	// HTTP2
	srv := &http.Server{
		Addr:    httpsPort,
		Handler: n,
	}
	http2.ConfigureServer(srv, &http2.Server{})
	log.Fatal(srv.ListenAndServeTLS(sslCertPath+"server.pem", sslCertPath+"server.key"))
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
var poem string
var wunderground string
var version string
var port string

func parseEnvVariables() {
	secret = os.Getenv("ackSecret")
	wunderground = os.Getenv("ackWunder")
	version = os.Getenv("CIRCLE_BUILD_NUM")
	prodSession, _ = strconv.ParseBool(os.Getenv("prodSession"))
}

func setUpMuxHandlers(mux *http.ServeMux) {
	post := "POST"
	homePageMap = baseball.InitHomePageMap()

	mux.HandleFunc("/clockCheck", func(w http.ResponseWriter, r *http.Request) {
		ClockCheckHandler(w, r)
	})
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
		/*
			id := r.Proto

			clientGone := w.(http.CloseNotifier).CloseNotify()
			w.Header().Set("Content-Type", "text/plain")
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()
			fmt.Fprintf(w, "# ~1KB of junk to force browsers to start rendering immediately: \n")
			io.WriteString(w, strings.Repeat("# xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n", 13))

			for {
				fmt.Fprintf(w, "%v [%s]\n", time.Now(), r.Proto)
				w.(http.Flusher).Flush()
				select {
				case <-ticker.C:
				case <-clientGone:
					log.Printf("Client %v disconnected from the clock", r.RemoteAddr)
					return
				}
			}
		*/
	})

	// gameDayListing for yesterday (default 'homepage')
	mux.HandleFunc("/bb", bbHome)

	// ajax request for gameDayListing
	mux.HandleFunc("/bbAjaxDay", bbAjaxDay)

	// play a single game
	mux.HandleFunc("/bbStream", bbStream)

	// play all games of the day
	mux.HandleFunc("/bbAll", bbAll)
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
		//TODO 2016 is over - make default /bb goto Game 7, 2016 World Series
		date1 = "year_2016/month_11/day_02"
		offset = "0"
	}
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
