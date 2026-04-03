package filters

// Somehow want to keep my files independent of each other
// so the rw import below has to go at some point.
// or specific functions from rw would have to move to a
// different file.

import (
	"encoding/json"
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

func (f Filter) ResolveFilter() string {
	playerName := strings.ToLower(f.Player)
	// mapName := strings.ToLower(f.Map)
	// teamName := strings.ToLower(f.Team)
	ans1 := playerTo32(playerName)
	fmt.Println(ans1)
	stringy := fmt.Sprintf("%s/%s_%s_%s_%s_%s_team-%s_side-%s_player-%s*.pb", killsDir, "*", "*", "*", "*", "*", "*", "*", ans1)
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
	return "0"
}

func playerTo32(name string) string {
	players := make(map[string]rw.PlayerInfo)
	name = strings.ToLower(name)
	file, err := os.ReadFile(playersFile)
	if err != nil {
		log.Fatalln(err)
	}
	err = json.Unmarshal(file, &players)
	// fmt.Println(players)
	// fmt.Println(players["81417650"])
	for id, info := range players {
		if strings.ToLower(info.Name) == name {
			return id
		}
	}
	return "0"
}
