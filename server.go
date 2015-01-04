package main

import (
	"encoding/json"
	"github.com/codegangsta/negroni"
	"github.com/danackerson/ackerson.de-go/structures"
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

	n := negroni.Classic()
	n.UseHandler(mux)
	n.Run(":3000")
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
		//log.Printf("%s\n", string(currentWeather))
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
