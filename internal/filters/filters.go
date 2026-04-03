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

type Filter struct {
	Player   string
	Map      string
	DateFrom string
	DateTo   string
	Side     string
	Team     string
}

func (f Filter) ResolveFilter() ([]string, error) {
	var playerName, player32 string
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
	// playerName := strings.ToLower(f.Player)
	// mapName := strings.ToLower(f.Map)
	// teamName := strings.ToLower(f.Team)
	// player32 = playerTo32(playerName)
	fmt.Println(player32)
	stringy := fmt.Sprintf("%s/%s_%s_%s_%s_%s_team-%s_side-%s_player-%s.pb", killsDir, "*", "*", "*", "*", "*", "*", "*", player32)
	fmt.Println(stringy)
	// I think the way the data is stored/fetched is fundamentally bad
	// A goal was not to use SQL of any kind so I'll stick to that
	// I had originally thought it would just search through every file
	// and find files that matched the filters.
	// File names would have to be structured (which they are)

	// <date>_<time>_<team_one>-vs-<team_two>_<map>_team-<team_hash>_side-<side>_player-<STEAM32ID>.pb

	// The below block is a bit of a test of that. Why AI didn't use this
	// I'm not sure... Maybe it's worse, maybe it's better but I'll see...
	// matches, err := filepath.Glob(killsDir + "/*.pb")
	matches, err := filepath.Glob(stringy)
	if err != nil {
		log.Fatal(err)
	}
	for _, match := range matches {
		matchNoPrefix, _ := strings.CutPrefix(match, "data/kills/")

		fmt.Println(strings.Split(matchNoPrefix, "_"))
	}
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
	// fmt.Println(players)
	// fmt.Println(players["81417650"])
	for id, info := range players {
		if strings.ToLower(info.Name) == name {
			return id, nil
		}
	}
	return "", PlayerNotFound
}
