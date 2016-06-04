package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
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
	homePageMap := InitHomePageMap()

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

	mux.HandleFunc("/bbStream", func(w http.ResponseWriter, r *http.Request) {
		URL := r.URL.Query().Get("url")
		log.Print("render URL: " + URL)

		render := render.New(render.Options{
			IsDevelopment: false,
		})
		render.HTML(w, http.StatusOK, "stream", URL)
	})

	mux.HandleFunc("/bbAll", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			date := time.Now().AddDate(0, 0, -1)
			date1 := r.URL.Query().Get("date1")
			if date1 != "" {
				date1 = strings.TrimLeft(date1, "year_")
				location, _ := time.LoadLocation("UTC")
				monthDayString, err := time.ParseInLocation("2006/month_01/day_02", date1, location)
				log.Print("found: " + monthDayString.Format(time.RFC3339))
				if err != nil {
					log.Print(err)
				} else {
					i, _ := strconv.Atoi(r.URL.Query().Get("offset"))
					date = monthDayString.AddDate(0, 0, i)
				}
			}
			w.Header().Set("Cache-Control", "max-age=10800")
			DayHandler(w, r, date, homePageMap)
		}
	})

	mux.HandleFunc("/bbDay", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			date := time.Now().AddDate(0, 0, -1)
			date1 := r.URL.Query().Get("date1")
			if date1 != "" {
				date1 = strings.TrimLeft(date1, "year_")
				location, _ := time.LoadLocation("UTC")
				monthDayString, err := time.ParseInLocation("2006/month_01/day_02", date1, location)
				if err != nil {
					log.Print(err)
				} else {
					i, _ := strconv.Atoi(r.URL.Query().Get("offset"))
					date = monthDayString.AddDate(0, 0, i)
				}
			}
			w.Header().Set("Cache-Control", "max-age=10800")
			GameHandler(w, r, date, homePageMap)
		}
	})

	mux.HandleFunc("/bb", func(w http.ResponseWriter, req *http.Request) {
		// prepare response page
		date := time.Now().AddDate(0, 0, -1)
		dates := "year_" + date.Format("2006/month_01/day_02")
		readableDates := date.Format("Mon, Jan _2 2006")
		games := make(map[int][]string)
		games = SearchMLBGames(dates, games, homePageMap)

		r := render.New(render.Options{
			Layout:        "content",
			IsDevelopment: false,
		})

		w.Header().Set("Cache-Control", "max-age=10800")
		r.HTML(w, http.StatusOK, "mlbResponse", GameDay{Date: dates, ReadableDate: readableDates, Games: games})
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

// DayHandler is now commented
func DayHandler(w http.ResponseWriter, req *http.Request, gameDate time.Time, homePageMap map[int]Team) {
	dates := "year_" + gameDate.Format("2006/month_01/day_02")
	games := make(map[int][]string)
	games = SearchMLBGames(dates, games, homePageMap)

	// prepare response page
	r := render.New(render.Options{
		IsDevelopment: true,
	})
	
	var game_urls []string
	for _, stringArray := range games {
    game_urls = append(game_urls, stringArray[10])
  }
  sort.Strings(game_urls)
	readableDates := gameDate.Format("Mon, Jan _2 2006")

	r.HTML(w, http.StatusOK, "playlist", AllGames{
		Date: readableDates, 
		VideoCountStorage: gameDate.Format("2006-01-02"), 
		BallgameVideoURLs: game_urls,
		BallgameCount: len(games) })
}

// AllGames is now commented
type AllGames struct {
	Date         			string
	VideoCountStorage	string
	BallgameVideoURLs []string
	BallgameCount     int
}

// GameHandler is now commented
func GameHandler(w http.ResponseWriter, req *http.Request, gameDate time.Time, homePageMap map[int]Team) {
	dates := "year_" + gameDate.Format("2006/month_01/day_02")
	games := make(map[int][]string)
	games = SearchMLBGames(dates, games, homePageMap)

	// prepare response page
	r := render.New(render.Options{
		IsDevelopment: false,
	})

	readableDates := gameDate.Format("Mon, Jan _2 2006")

	r.HTML(w, http.StatusOK, "mlbResponse", GameDay{Date: dates, ReadableDate: readableDates, Games: games})
}

// GameDay is now commented
type GameDay struct {
	Date         string
	ReadableDate string
	Games        map[int][]string
}

// SearchMLBGames is now commented
func SearchMLBGames(dates string, games map[int][]string, homePageMap map[int]Team) map[int][]string {
	domain := "http://gd2.mlb.com/components/game/mlb/"
	suffix := "/grid_ce.xml"
	url := domain + dates + suffix

	log.Printf(url)
	resp, err := http.Get(url)
	if err != nil {
		log.Print(err)
		return nil
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

		if gameID != "" {
			awayTeamID := aGameVal["-away_team_id"].(string)
			awayTeamName, awayTeamHomePage := LookupTeamInfo(homePageMap, awayTeamID)
			awayAbbrev := aGameVal["-away_name_abbrev"].(string)

			homeTeamID := aGameVal["-home_team_id"].(string)
			homeTeamName, homeTeamHomePage := LookupTeamInfo(homePageMap, homeTeamID)
			homeAbbrev := aGameVal["-home_name_abbrev"].(string)

			detailURL := "http://m.mlb.com/gen/multimedia/detail" + generateDetailURL(gameID)
			gameURL := fetchGameURL(detailURL, "FLASH_2500K_1280X720")

			games[k] = []string{awayTeamName, awayTeamHomePage, awayTeamID, awayAbbrev, homeTeamName, homeTeamHomePage, homeTeamID, homeAbbrev, gameID, dates, gameURL}
		}
	}

	log.Println("Media games:", mediaGames)
	log.Println("Condensed games:", condensedGames)

	return games
}

// fetchGameURL is now commented
func fetchGameURL(detailURL string, desiredQuality string) string {
	gameURL := "MickeyMouse.mp4"

	resp, err := http.Get(detailURL)
	if err != nil {
		log.Print(err)
	}
	defer resp.Body.Close()
	xml, err := ioutil.ReadAll(resp.Body)
	m, err := mxj.NewMapXml(xml)

	URLs, err := m.ValuesForKey("url")
	if err != nil {
		log.Fatal("err:", err.Error())
		return ""
	}

	// now just manipulate Map entries returned as []interface{} array.
	for _, v := range URLs {
		aGameVal, _ := v.(map[string]interface{})
		if aGameVal["-playback_scenario"].(string) == desiredQuality {
			return aGameVal["#text"].(string)
		}
	}

	return gameURL
}

// generateDetailURL is now commented
func generateDetailURL(gameID string) string {
	// given gameID 605442983 return "/9/8/3/605442983.xml"
	return 	"/" + gameID[len(gameID)-3:len(gameID)-2] + 
					"/" + gameID[len(gameID)-2:len(gameID)-1] + 
					"/" + gameID[len(gameID)-1:] + 
					"/" + gameID + ".xml"
}

// Team is now commented
type Team struct {
	Name, HomePage string
}

// LookupTeamInfo is now commented
func LookupTeamInfo(homePageMap map[int]Team, teamIDS string) (string, string) {
	teamID, _ := strconv.Atoi(teamIDS)
	return homePageMap[teamID].Name, homePageMap[teamID].HomePage
}

// InitHomePageMap is now commented
func InitHomePageMap() map[int]Team {
	homePageMap := make(map[int]Team)

	homePageMap[110] = Team{"Baltimore Orioles", "http://m.orioles.mlb.com/roster"}
	homePageMap[145] = Team{"Chicago Whitesox", "http://m.whitesox.mlb.com/roster"}
	homePageMap[117] = Team{"Houston Astros", "http://m.astros.mlb.com/roster"}
	homePageMap[144] = Team{"Atlanta Braves", "http://m.braves.mlb.com/roster"}
	homePageMap[112] = Team{"Chicago Cubs", "http://m.cubs.mlb.com/roster"}
	homePageMap[109] = Team{"Arizona Diamond Backs", "http://m.dbacks.mlb.com/roster"}
	homePageMap[111] = Team{"Boston Red Sox", "http://m.redsox.mlb.com/roster"}
	homePageMap[114] = Team{"Cleveland Indians", "http://m.indians.mlb.com/roster"}
	homePageMap[108] = Team{"Los Angeles Angels", "http://m.angels.mlb.com/roster"}
	homePageMap[146] = Team{"Miami Marlins", "http://m.marlins.mlb.com/roster"}
	homePageMap[113] = Team{"Cincinnati Reds", "http://m.reds.mlb.com/roster"}
	homePageMap[115] = Team{"Colorado Rockies", "http://www.rockies.com/roster"}
	homePageMap[147] = Team{"New York Yankees", "http://m.yankees.mlb.com/roster"}
	homePageMap[116] = Team{"Detroit Tigers", "http://www.tigers.com/roster"}
	homePageMap[133] = Team{"Oakland Athletics", "http://m.athletics.mlb.com/roster"}
	homePageMap[121] = Team{"New York Mets", "http://m.mets.mlb.com/roster"}
	homePageMap[158] = Team{"Milwaukee Brewers", "http://m.brewers.mlb.com/roster"}
	homePageMap[119] = Team{"LA Dodgers", "http://m.dodgers.mlb.com/roster"}
	homePageMap[139] = Team{"Tampa Bay Rays", "http://m.rays.mlb.com/roster"}
	homePageMap[118] = Team{"Kansas City Royals", "http://m.royals.mlb.com/roster"}
	homePageMap[136] = Team{"Seattle Mariners", "http://m.mariners.mlb.com/roster"}
	homePageMap[143] = Team{"Philadelphia Phillies", "http://m.phillies.mlb.com/roster"}
	homePageMap[138] = Team{"St Louis Cardinals", "http://m.cardinals.mlb.com/roster"}
	homePageMap[135] = Team{"San Diego Padres", "http://m.padres.mlb.com/roster"}
	homePageMap[141] = Team{"Toronto Blue Jays", "http://m.bluejays.mlb.com/roster"}
	homePageMap[142] = Team{"Minnesota Twins", "http://m.twins.mlb.com/roster"}
	homePageMap[140] = Team{"Texas Rangers", "http://m.rangers.mlb.com/roster"}
	homePageMap[120] = Team{"Washington Nationals", "http://m.nationals.mlb.com/roster"}
	homePageMap[134] = Team{"Pittsburgh Pirates", "http://m.pirates.mlb.com/roster"}
	homePageMap[137] = Team{"San Francisco Giants", "http://m.giants.mlb.com/roster"}

	return homePageMap
}
