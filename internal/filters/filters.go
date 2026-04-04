package filters

// Somehow want to keep my files independent of each other
// so the rw import below has to go at some point.
// or specific functions from rw would have to move to a
// different file.

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	rw "github.com/TheBigJayT/round-two-cs/internal/readwrite"
)

const (
	teamsFile   = "data/teams.json"
	playersFile = "data/players.json"
	killsDir    = "data/kills"
	matchesFile = "data/matches.jsonl"
	mapsFile    = "data/minimap.json"
)

// In the future I expect these will become slices so one can
// filter by multiple players and teams and maps etc etc.
type Filter struct {
	Player   string
	Map      string
	DateFrom string
	DateTo   string
	Side     string
	Team     string
}

var NoMatches = errors.New("No files match your filter")

func (f Filter) ResolveFilter() ([]string, error) {
	var playerName, player32, mapName, teamOne, teamTwo, playerTeam, playerSide, exactDate string
	if f.Player != "" {
		var err error
		playerName = strings.ToLower(f.Player)
		player32, err = playerTo32(playerName)
		if err != nil {
			return []string{}, err
		}
	} else if f.Player == "" {
		player32 = "*"
	}

	if f.Map != "" {
		mapName = strings.ToLower(f.Map)
		mapName, _ = strings.CutPrefix(mapName, "de_")
		err := isMap(mapName)
		if err != nil {
			return []string{}, err
		}
		fmt.Println(mapName)
	} else if f.Map == "" {
		mapName = "*"
	}

	if f.DateFrom != "" && f.DateTo != "" {
		// do something with exact date.
	}

	if f.Side != "" {
		side, err := upperSide(f.Side)
		if err == nil {
			playerSide = side
		} else {
			return []string{}, err
		}
	} else {
		playerSide = "*"
	}

	// Playing for TEAM
	if f.Team != "" {
		var err error
		playerTeam, err = teamToHash(f.Team)
		if err != nil {
			return []string{}, err
		}

	} else {
		playerTeam = "*"
	}

	// filters not implemented yet go here
	if true {
		exactDate = "*"
		teamOne = "*"
		teamTwo = "*"
		// playerTeam = "*"
	}
	var stringy string
	// 																		/data/kills/<date>_<time>_<team_one>_<team_two>_<map>_team-<team_hash>_side-<side>_player-<STEAM32ID>.pb
	stringy = fmt.Sprintf("%s/%s_%s_%s_%s_%s_team-%s_side-%s_player-%s.pb", killsDir, exactDate, "*", teamOne, teamTwo, mapName, playerTeam, playerSide, player32)
	// fmt.Printf("\n%s\n\n", stringy)

	matches, err := filepath.Glob(stringy)
	if err != nil {
		log.Fatal(err)
	}
	if len(matches) == 0 {
		return matches, NoMatches
	}
	// for _, match := range matches {
	// 	matchNoPrefix, _ := strings.CutPrefix(match, "data/kills/")
	// 	// for the moment this prints the file paths.
	// 	fmt.Println(strings.Split(matchNoPrefix, "_"))
	// }
	// fmt.Println(len(matches))
	// I imagine this would be easy to "inject" or abuse which would mean
	// I'd have to implement some verification or something so someone doesn't
	// request the entire database or something idk.
	return matches, nil
}

var PlayerFileError = errors.New("Error opening players.json")
var PlayerNotFound = errors.New("No player found with that name")

func playerTo32(name string) (string, error) {
	players := make(map[string]rw.PlayerInfo)
	name = strings.ToLower(name)
	file, err := os.ReadFile(playersFile)
	if err != nil {
		return "", PlayerFileError
	}
	err = json.Unmarshal(file, &players)
	for id, info := range players {
		if strings.ToLower(info.Name) == name {
			return id, nil
		}
	}
	return "", PlayerNotFound
}

var MapNotFound = errors.New("Map not found")

func isMap(mapname string) error {
	maps := make(map[string]rw.MapInfo)
	file, err := os.ReadFile(mapsFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(file, &maps)
	for k := range maps {
		if strings.EqualFold(strings.TrimPrefix(mapname, "de_"), strings.TrimPrefix(k, "de_")) {
			return nil
		}
	}
	return MapNotFound
}

var SideNotValid = errors.New("That is not a valid side. (T or CT).")

func upperSide(side string) (string, error) {
	side = strings.ToUpper(side)
	switch side {
	case "T":
		return "T", nil
	case "CT":
		return "CT", nil
	default:
		return "", SideNotValid
	}
}

var TeamNotFound = errors.New("Team not found in database")

func teamToHash(team string) (string, error) {
	file, err := os.ReadFile(teamsFile)
	if err != nil {
		return "", err
	}
	var teams = make(map[string]string)
	err = json.Unmarshal(file, &teams)
	if teams[team] != "" {
		return teams[team], nil
	}
	return "", TeamNotFound
}
