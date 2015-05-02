package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/danackerson/ackerson.de-go/structures"
	"github.com/goincremental/negroni-sessions"
  "github.com/goincremental/negroni-sessions/cookiestore"
  "gopkg.in/mgo.v2"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	mux := http.NewServeMux()

	// handlers
	mux.HandleFunc("/date", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			DateHandler(w, r)
		}
	})
	mux.HandleFunc("/whoami", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			WhoAmIHandler(w, r)
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

			if pass == nil && r.FormValue("sesam") != "taco" {
				http.NotFound(w, r)
			} else if r.FormValue("sesam") == "taco" || pass != nil {
  			session.Set("pass", "true")

  			PoemsHandler(w, r)
  		}
		}
	})

	n := negroni.Classic()

	readInCreds()

	store := cookiestore.New([]byte(secret))  
  n.Use(sessions.Sessions("gurkherpaderp", store))
	n.UseHandler(mux)
	n.Run(":3001")
}

var mongo string
var secret string

func readInCreds() {
	content, _ := ioutil.ReadFile("/opt/creds.txt")
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		values := strings.Split(string(line), "=")
		if values[0] == "mongo" {
			mongo = values[1]
		} else if values[0] == "secret" {
			secret = values[1]
		}
  }
}

func loadWritings(w http.ResponseWriter) ([](structures.Writing)) {
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

func PoemsHandler(w http.ResponseWriter, req *http.Request) {
	writings := loadWritings(w)
  for _, writing := range writings {
    fmt.Fprintf(w, "%1.0f: %s", writing.ID, writing.Content)
    fmt.Fprintf(w, "\r\n")
  }
}

func GetIP(r *http.Request) string {
	if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
		return ipProxy
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func WhoAmIHandler(w http.ResponseWriter, req *http.Request) {
	s := []string{"[[g;#FFFF00;]Your IP:] " + GetIP(req), "[[g;#FFFF00;]Your Browser:] " + req.UserAgent()}
	rawData := strings.Join(s, "\r\n")
	rawDataJson := map[string]string{"whoami": rawData}

	data, _ := json.Marshal(rawDataJson)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	w.Write(data)
}

func DateHandler(w http.ResponseWriter, req *http.Request) {
	now := "[[g;#FFFF00;]" + time.Now().Format("Mon Jan _2 15:04:05 2006") + "]"
	date := map[string]string{"date": now}

	data, _ := json.Marshal(date)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	w.Write(data)
}

func WeatherHandler(w http.ResponseWriter, req *http.Request) {
	// handle JSON POST request
	//body := string(structures.TestGeoLocationPost) // in case you are testing :)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic("ioutil.ReadAll")
	}

	log.Printf("RAW /weather POST: %s", string(body))
	geoLocation := new(structures.JsonGeoLocationRequest)
	json.Unmarshal([]byte(body), &geoLocation)

	latString := strconv.FormatFloat(float64(geoLocation.Params.Lat), 'f', 15, 32)
	lngString := strconv.FormatFloat(float64(geoLocation.Params.Lng), 'f', 15, 32)

	// call wunderground API for Conditions & Forecast
	conditionsURI := "http://api.wunderground.com/api/c5060046bbda0736/conditions/q/"
	forecastURI := "http://api.wunderground.com/api/c5060046bbda0736/forecast/q/"
	locationParams := latString + "," + lngString + ".json"

	currentWeather := new(structures.CurrentWeatherConditions)
	currentWeatherResp, err := http.Get(conditionsURI + locationParams)
	if err != nil {
		log.Printf("%s", err)
	} else {
		defer currentWeatherResp.Body.Close()
		currentWeatherJSON, err := ioutil.ReadAll(currentWeatherResp.Body)
		if err != nil {
			log.Printf("%s", err)
		}
		json.Unmarshal([]byte(currentWeatherJSON), &currentWeather)
		log.Printf("%s\n", currentWeather)
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
		//log.Printf("%s\n", string(currentForecast))
	}

	code := map[string]interface{}{"current": currentWeather, "forecastday": currentForecast}
	data, _ := json.Marshal(code)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}
