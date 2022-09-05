package baseball

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

// FavoriteTeamGameListHandler is now commented
func FavoriteTeamGameListHandler(id string, homePageMap map[int]Team) []GameDay {
	var favTeamGames []GameDay

	if id != "" {
		// fetch all the games for this team in last 7-30 days from now
		date1 := ""
		for i := 0; i > -15; i-- {
			gameDate, dates, games := getDatesAndGames(date1, "-1", homePageMap, id)

			date1 = gameDate.Format("2006/month_01/day_02")

			favTeamGames = append(favTeamGames, GameDay{Date: dates, ReadableDate: gameDate.Format("Mon, Jan _2 2006"), Games: games})
		}
	}

	// log.Printf("%d: %v", len(favTeamGames), favTeamGames)
	return favTeamGames
}

// AllGames is now commented
type AllGames struct {
	Date              string
	VideoCountStorage string
	BallgameVideoURLs []string
	BallgameCount     int
	Mobile            bool
}

// PlayAllGamesOfDayHandler is now commented
func PlayAllGamesOfDayHandler(date1 string, offset string, homePageMap map[int]Team) AllGames {
	gameDate, _, games := getDatesAndGames(date1, offset, homePageMap, "")

	var gameUrls []string
	if len(games) > 1 {
		log.Printf(games[0][0])
	}
	for _, stringArray := range games {
		actualGameURL := FetchGameURLFromID(stringArray[10])

		log.Printf("gameID: %s -> URL: %s", stringArray[10], actualGameURL)
		gameUrls = append(gameUrls, actualGameURL)
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
	Mobile       bool
}

// GameDayListingHandler is now commented
func GameDayListingHandler(date1 string, offset string, homePageMap map[int]Team) GameDay {
	gameDate, dates, games := getDatesAndGames(date1, offset, homePageMap, "")

	return GameDay{Date: dates, ReadableDate: gameDate.Format("Mon, Jan _2 2006"), Games: games}
}

func getDatesAndGames(date1 string, offset string, homePageMap map[int]Team, favTeam string) (time.Time, string, map[int][]string) {
	gameDate := time.Now().AddDate(0, 0, -1)
	if date1 != "" {
		date1 = strings.TrimLeft(date1, "year_")
		location, _ := time.LoadLocation("UTC")
		monthDayString, err := time.ParseInLocation("2006/month_01/day_02", date1, location)
		if err != nil {
			log.Print("HERE: " + err.Error())
		} else {
			i, _ := strconv.Atoi(offset)
			gameDate = monthDayString.AddDate(0, 0, i)
		}
	}

	dates := gameDate.Format("01/02/2006")
	//dates := "year_" + gameDate.Format("2006/month_01/day_02")
	games := make(map[int][]string)
	games = searchMLBGames(dates, games, homePageMap, favTeam)

	return gameDate, dates, games
}

// SearchMLBGames is now commented
func searchMLBGames(dates string, games map[int][]string, homePageMap map[int]Team, favTeam string) map[int][]string {
	domain := "https://statsapi.mlb.com"
	suffix := "/api/v1/schedule?sportId=1&date="
	URL := domain + suffix + url.PathEscape(dates)
	var awayTeamID, homeTeamID string
	k := 0

	resp, err := http.Get(URL)
	if err != nil {
		log.Print(err)
		return nil
	}
	defer resp.Body.Close()

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
	}

	scheduleJSON := string(raw)

	gamesJSON := gjson.Get(scheduleJSON, "dates.0.games")
	for _, game := range gamesJSON.Array() {
		awayTeamID = game.Get("teams.away.team.id").String()
		homeTeamID = game.Get("teams.home.team.id").String()

		if favTeam != "" && (favTeam != awayTeamID && favTeam != homeTeamID) {
			//log.Printf("[%s]: %s != %s OR %s", dates, favTeam, awayTeamID, homeTeamID)
			continue
		}

		contentURL := game.Get("content.link").String()
		if contentURL != "" && contentURL != "http://baseball.theater" {
			awayTeam := LookupTeamInfo(homePageMap, awayTeamID)
			homeTeam := LookupTeamInfo(homePageMap, homeTeamID)
			gameURL := contentURL
			gameID := game.Get("gamePK").String()
			games[k] = []string{
				awayTeam.Name, awayTeam.HomePage, strconv.Itoa(awayTeam.ID), awayTeam.Abbreviation,
				homeTeam.Name, homeTeam.HomePage, strconv.Itoa(homeTeam.ID), homeTeam.Abbreviation,
				gameID, dates, gameURL}
			k++
		}
	}

	return games
}

// FetchGameURLFromID is now commented
func FetchGameURLFromID(contentURL string) string {
	gameURL := "http://baseball.theater"
	resp, err := http.Get("https://statsapi.mlb.com" + contentURL)
	if err != nil {
		log.Print(err)
	}

	raw, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Print(err)
	}

	mediaJSON := string(raw)

	video := gjson.Get(mediaJSON, `media.epgAlternate.#[title="Extended Highlights"].items.#.playbacks.#[name="mp4Avc"].url`)
	if len(video.Array()) <= 0 {
		video = gjson.Get(mediaJSON, `media.epg.#[title="Extended Highlights"].items.#.playbacks.#[name="mp4Avc"].url`)
	}
	if len(video.Array()) <= 0 {
		video = gjson.Get(mediaJSON, `media.epgAlternate.#[title="Extended Highlights"].items.#.playbacks.#[name="FLASH_2500K_1280X720"].url`)
	}
	if len(video.Array()) <= 0 {
		video = gjson.Get(mediaJSON, `media.epg.#[title="Extended Highlights"].items.#.playbacks.#[name="FLASH_2500K_1280X720"].url`)
	}

	//elapsedX1 := time.Since(startX1)
	//log.Printf("JSON parsing of subsection %s took %s", contentURL, elapsedX1)

	if len(video.Array()) > 0 {
		gameURL = video.Array()[0].String()
	} else {
		/* TODO: go grab the video from https://baseball.theater/search/condensed-game/2021-10-30 ?
		<a href="https://mlb-cuts-diamond.mlb.com/FORGE/2021/2021-10/30/856f2bda-dfe3a2e3-603b1c29-csvm-diamondx64-asset_1280x720_59_4000K.mp4" target="_blank">
		<div class="..." title="CG: HOU@ATL Gm4 - 10/30/21" ...</a>
		*/
	}

	return gameURL
}

// Team is now commented
type Team struct {
	ID                           int
	Abbreviation, Name, HomePage string
}

// LookupTeamInfo is now commented
func LookupTeamInfo(homePageMap map[int]Team, teamIDS string) Team {
	teamID, _ := strconv.Atoi(teamIDS)
	return Team{teamID, homePageMap[teamID].Abbreviation, homePageMap[teamID].Name, homePageMap[teamID].HomePage}
}

// InitHomePageMap is now commented
func InitHomePageMap() map[int]Team {
	homePageMap := make(map[int]Team)

	homePageMap[110] = Team{110, "BAL", "Baltimore Orioles", "http://m.orioles.mlb.com/roster"}
	homePageMap[145] = Team{145, "CHW", "Chicago Whitesox", "http://m.whitesox.mlb.com/roster"}
	homePageMap[117] = Team{117, "HOU", "Houston Astros", "http://m.astros.mlb.com/roster"}
	homePageMap[144] = Team{144, "ATL", "Atlanta Braves", "http://m.braves.mlb.com/roster"}
	homePageMap[112] = Team{112, "CHC", "Chicago Cubs", "http://m.cubs.mlb.com/roster"}
	homePageMap[109] = Team{109, "ARI", "Arizona Diamond Backs", "http://m.dbacks.mlb.com/roster"}
	homePageMap[111] = Team{111, "BOS", "Boston Red Sox", "http://m.redsox.mlb.com/roster"}
	homePageMap[114] = Team{114, "CLE", "Cleveland Indians", "http://m.indians.mlb.com/roster"}
	homePageMap[108] = Team{108, "LAA", "Los Angeles Angels", "http://m.angels.mlb.com/roster"}
	homePageMap[146] = Team{146, "MIA", "Miami Marlins", "http://m.marlins.mlb.com/roster"}
	homePageMap[113] = Team{113, "CIN", "Cincinnati Reds", "http://m.reds.mlb.com/roster"}
	homePageMap[115] = Team{115, "COL", "Colorado Rockies", "http://www.rockies.com/roster"}
	homePageMap[147] = Team{147, "NYY", "New York Yankees", "http://m.yankees.mlb.com/roster"}
	homePageMap[116] = Team{116, "DET", "Detroit Tigers", "http://www.tigers.com/roster"}
	homePageMap[133] = Team{133, "OAK", "Oakland Athletics", "http://m.athletics.mlb.com/roster"}
	homePageMap[121] = Team{121, "NYM", "New York Mets", "http://m.mets.mlb.com/roster"}
	homePageMap[158] = Team{158, "MIL", "Milwaukee Brewers", "http://m.brewers.mlb.com/roster"}
	homePageMap[119] = Team{119, "LAD", "LA Dodgers", "http://m.dodgers.mlb.com/roster"}
	homePageMap[139] = Team{139, "TB", "Tampa Bay Rays", "http://m.rays.mlb.com/roster"}
	homePageMap[118] = Team{118, "KC", "Kansas City Royals", "http://m.royals.mlb.com/roster"}
	homePageMap[136] = Team{136, "SEA", "Seattle Mariners", "http://m.mariners.mlb.com/roster"}
	homePageMap[143] = Team{143, "PHI", "Philadelphia Phillies", "http://m.phillies.mlb.com/roster"}
	homePageMap[138] = Team{138, "STL", "St Louis Cardinals", "http://m.cardinals.mlb.com/roster"}
	homePageMap[135] = Team{135, "SD", "San Diego Padres", "http://m.padres.mlb.com/roster"}
	homePageMap[141] = Team{141, "TOR", "Toronto Blue Jays", "http://m.bluejays.mlb.com/roster"}
	homePageMap[142] = Team{142, "MIN", "Minnesota Twins", "http://m.twins.mlb.com/roster"}
	homePageMap[140] = Team{140, "TEX", "Texas Rangers", "http://m.rangers.mlb.com/roster"}
	homePageMap[120] = Team{120, "WAS", "Washington Nationals", "http://m.nationals.mlb.com/roster"}
	homePageMap[134] = Team{134, "PIT", "Pittsburgh Pirates", "http://m.pirates.mlb.com/roster"}
	homePageMap[137] = Team{137, "SF", "San Francisco Giants", "http://m.giants.mlb.com/roster"}

	return homePageMap
}
