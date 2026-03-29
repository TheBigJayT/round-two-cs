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
	playerToHash(f.Player)
}

func playerToHash(name string) {
	players := make(map[string]rw.PlayerInfo)
	name = strings.ToLower(name)
	file, err := os.ReadFile(playersFile)
	if err != nil {
		log.Fatalln(err)
	}
	err = json.Unmarshal(file, &players)
	fmt.Println(players)
}
