package filters

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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

func (f Filter) ResolveFilter() {
	var ans string
	new, err := rw.LoadMatches(matchesFile)
	if err != nil {
		log.Fatal(err)
	}
	if f.Player != "" {
		ans = playerToHash(f.Player)
	} else {
		ans = ""
	}

	for k, v := range new {
		if ans != "" {

			fmt.Println(k, v.Players[ans].KillsPaths)
			for i, j := range v.Players[ans].KillsPaths {
				fmt.Println(i, j)
			}
		} else {
			for _, j := range v.Players {
				for _, h := range j.KillsPaths {
					fmt.Println(h)
				}
			}
		}
	}

}

func playerToHash(name string) string {
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
