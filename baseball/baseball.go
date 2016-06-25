package baseball

import (
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/clbanning/mxj"
)

// FavoriteTeamGameListHandler is now commented
func FavoriteTeamGameListHandler(id string, homePageMap map[int]Team) GameDay {
	// TODO

	return GameDay{Date: id}
}

// AllGames is now commented
type AllGames struct {
	Date              string
	VideoCountStorage string
	BallgameVideoURLs []string
	BallgameCount     int
}

// PlayAllGamesOfDayHandler is now commented
func PlayAllGamesOfDayHandler(date1 string, offset string, homePageMap map[int]Team) AllGames {
	gameDate, _, games := getDatesAndGames(date1, offset, homePageMap)

	var gameUrls []string
	for _, stringArray := range games {
		gameUrls = append(gameUrls, stringArray[10])
	}
	sort.Strings(gameUrls)

	return AllGames{
		Date:              gameDate.Format("Mon, Jan _2 2006"),
		VideoCountStorage: gameDate.Format("2006-01-02"),
		BallgameVideoURLs: gameUrls,
		BallgameCount:     len(games)}
}

// GameDay is now commented
type GameDay struct {
	Date         string
	ReadableDate string
	Games        map[int][]string
}

// GameDayListingHandler is now commented
func GameDayListingHandler(date1 string, offset string, homePageMap map[int]Team) GameDay {
	gameDate, dates, games := getDatesAndGames(date1, offset, homePageMap)

	return GameDay{Date: dates, ReadableDate: gameDate.Format("Mon, Jan _2 2006"), Games: games}
}

func getDatesAndGames(date1 string, offset string, homePageMap map[int]Team) (time.Time, string, map[int][]string) {
	gameDate := time.Now().AddDate(0, 0, -1)
	if date1 != "" {
		date1 = strings.TrimLeft(date1, "year_")
		location, _ := time.LoadLocation("UTC")
		monthDayString, err := time.ParseInLocation("2006/month_01/day_02", date1, location)
		if err != nil {
			log.Print(err)
		} else {
			i, _ := strconv.Atoi(offset)
			gameDate = monthDayString.AddDate(0, 0, i)
		}
	}

	dates := "year_" + gameDate.Format("2006/month_01/day_02")
	games := make(map[int][]string)
	games = searchMLBGames(dates, games, homePageMap)

	return gameDate, dates, games
}

// SearchMLBGames is now commented
func searchMLBGames(dates string, games map[int][]string, homePageMap map[int]Team) map[int][]string {
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
	m, err := mxj.NewMapXml(xml)

	gameInfos, err := m.ValuesForKey("game")
	if err != nil {
		log.Fatal("err:", err.Error())
		log.Printf("MLB site '%s' response empty", domain)
		games[0] = []string{"Error connecting to " + domain}
		return games
	}

	// now just manipulate Map entries returned as []interface{} array.
	for k, v := range gameInfos {
		gameID := ""
		aGameVal, _ := v.(map[string]interface{})
		if aGameVal["-media_state"].(string) == "media_dead" {
			continue
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

	return games
}

// fetchGameURL is now commented
func fetchGameURL(detailURL string, desiredQuality string) string {
	fallbackURL := "https://www.youtube.com/user/MLB"

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
			if aGameVal["#text"] != nil {
				return aGameVal["#text"].(string)
			}
			log.Print("ERROR: this game has no videoURL: " + detailURL)
		}
	}

	return fallbackURL
}

// generateDetailURL is now commented
func generateDetailURL(gameID string) string {
	// given gameID 605442983 return "/9/8/3/605442983.xml"
	return "/" + gameID[len(gameID)-3:len(gameID)-2] +
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
