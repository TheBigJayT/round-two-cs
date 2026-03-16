package readwrite

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/TheBigJayT/round-two-cs/protos"

	demos "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	"google.golang.org/protobuf/proto"

	"github.com/djherbis/times"
	// ex "github.com/markus-wa/demoinfocs-golang/v5/examples"
	events "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
	msg "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/msg"
)

const (
	teamsFile   = "data/teams.json"
	playersFile = "data/players.json"
	killsDir    = "data/kills"
	matchesFile = "data/matches.jsonl"
	mapsFile    = "data/minimap.json"
)

// PlayerInfo is the value stored in data/players.json, keyed by SteamID32 string.
type PlayerInfo struct {
	Name    string `json:"name"`
	SteamID uint32 `json:"steam_id"`
}

type MapInfo struct {
	PosX   int     `json:"pos_x"`
	PosY   int     `json:"pos_y"`
	Scale  float32 `json:"scale"`
	Rotate int     `json:"rotate"`
	Zoom   int     `json:"zoom"`
}

// PlayerMetadata is the per-player record inside MatchMetadata.Players.
// The map is keyed by SteamID32 string (e.g. "76561198...").
// KillsPaths maps side ("T" or "CT") to the path of the player's .pb kill file.
type PlayerMetadata struct {
	Name       string            `json:"name"`
	TeamID     string            `json:"team_id"`
	KillsPaths map[string]string `json:"kills_paths"` // side -> .pb path
}

type MatchMetadata struct {
	MatchID string                    `json:"match_id"`
	Date    string                    `json:"date"`
	Map     string                    `json:"map"`
	Players map[string]PlayerMetadata `json:"players"` // key = SteamID32 string
}

func TestReadMapInfo() {
	setPosX := -2009
	setPosY := 369
	maps := make(map[string]MapInfo)

	data, err := os.ReadFile(mapsFile)
	if err == nil {
		json.Unmarshal(data, &maps)
	}
	for k, v := range maps {
		fmt.Println(k, v)
		// for pos = pos.X -> 0
		// |pos| slightly > |pos.X| -> slightly larger than 0.
		fmt.Println((float32(setPosX - v.PosX)) / v.Scale)
		fmt.Println((float32(v.PosY - setPosY)) / v.Scale)

	}
}

func ReadPos(filename string) (*protos.Positions, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	positions := &protos.Positions{}

	if err := proto.Unmarshal(data, positions); err != nil {
		return nil, err
	}
	return positions, nil
}

func WritePos(filename string, data []byte) error {
	err := os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func create() *protos.Positions {
	p := &protos.Positions{}
	return p
}

func ReadDemo(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	t, err := times.Stat(filename)
	if err != nil {
		log.Fatal(err.Error())
	}
	mod_time := t.ModTime()
	fmt.Println(mod_time.Format("20060102"))
	fmt.Printf("%T\n", mod_time)

	defer f.Close()
	demo := demos.NewParser(f)
	defer demo.Close()
	kills := make(map[string][]*protos.Position)

	demo.RegisterEventHandler(func(e events.Kill) {

		killer := e.Killer

		dead := e.Victim

		var new_position *protos.Position
		if killer != nil {
			new_position = &protos.Position{
				X:          float32(killer.Position().X),
				Y:          float32(killer.Position().Y),
				Z:          float32(killer.Position().Z),
				ID32:       int32(killer.SteamID32()),
				KillerName: string(killer.Name),
				DeadX:      float32(dead.Position().X),
				DeadY:      float32(dead.Position().Y),
				DeadZ:      float32(dead.Position().Z),
				DeadName:   string(dead.Name),
			}
			kills[killer.Name] = append(kills[killer.Name], new_position)
		}

	})

	demo.ParseToEnd()
	rounds := demo.GameState().TotalRoundsPlayed()
	jdata := create()
	for k, v := range kills {
		var key int
		for _, m := range v {

			jdata.Pos = append(jdata.Pos, m)
			key = rounds * len(v)
		}

		marshaled, err := proto.Marshal(jdata)
		if err != nil {
			log.Fatal(err)
		}
		filename := fmt.Sprintf("%s_%08d_%s.pb", mod_time.Format("20060102"), key, k)
		fmt.Println(filename)
		err = WritePos(filename, marshaled)

		jdata.Reset()
	}

	return nil
}

func ExtractKillsData(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalln("error opening", err)
		return err
	}
	defer f.Close()

	t, err := times.Stat(filename)
	if err != nil {
		log.Fatalln("error stating", err)
		return err
	}
	mod_time := t.ModTime()
	dateStr := mod_time.Format("2006-01-02")

	demo := demos.NewParser(f)
	defer demo.Close()

	playerKills := make(map[string]*protos.KillDataList)

	getTeamID := func(teamName string) string {
		safeTeamName := strings.ReplaceAll(strings.ToLower(teamName), " ", "-")
		if safeTeamName == "" {
			safeTeamName = "unknown"
		}

		teams := make(map[string]string)

		data, err := os.ReadFile(teamsFile)
		if err == nil {
			json.Unmarshal(data, &teams)
		}

		if id, exists := teams[safeTeamName]; exists {
			return id
		}

		hash := md5.Sum([]byte(safeTeamName))
		id := hex.EncodeToString(hash[:])[:6]

		teams[safeTeamName] = id
		newData, _ := json.MarshalIndent(teams, "", "    ")
		os.WriteFile(teamsFile, newData, 0644)

		return id
	}

	// getPlayerID upserts a player into data/players.json (keyed by SteamID32 string)
	// and returns the SteamID32 as a string identifier for use in filenames.
	getPlayerID := func(steamID32 uint32, playerName string) string {
		idStr := fmt.Sprintf("%d", steamID32)

		players := make(map[string]PlayerInfo)

		data, err := os.ReadFile(playersFile)
		if err == nil {
			json.Unmarshal(data, &players)
		}

		// Always update the name in case it changed (e.g. player renamed)
		players[idStr] = PlayerInfo{
			Name:    playerName,
			SteamID: steamID32,
		}

		newData, _ := json.MarshalIndent(players, "", "    ")
		os.WriteFile(playersFile, newData, 0644)

		return idStr
	}

	var matchStarted bool

	demo.RegisterNetMessageHandler(func(m *msg.CDemoFileInfo) {
		fmt.Println(m.GetGameInfo())
	})
	// events.MatchStart is not available in every demo thus events.AnnouncementMatchStarted is used instead...
	demo.RegisterEventHandler(func(e events.AnnouncementMatchStarted) {
		matchStarted = true
	})

	demo.RegisterEventHandler(func(e events.Kill) {
		killer := e.Killer
		dead := e.Victim
		selfKill := killer == dead
		if matchStarted {

			if killer != nil && dead != nil && !selfKill && killer.Team != dead.Team {
				weaponName := ""
				if e.Weapon != nil {
					weaponName = e.Weapon.String()
				}

				killerTeamName := ""
				if killer.TeamState != nil {
					killerTeamName = killer.TeamState.ClanName()
					if killerTeamName == "" {
						if killer.Team == 2 {
							killerTeamName = "T"
						} else if killer.Team == 3 {
							killerTeamName = "CT"
						}
					}
				}

				side := "SPEC"
				if killer.Team == 2 {
					side = "T"
				} else if killer.Team == 3 {
					side = "CT"
				}

				teamID := getTeamID(killerTeamName)
				playerID := getPlayerID(killer.SteamID32(), killer.Name)

				killData := &protos.KillData{
					KillerName:   killer.Name,
					KillerID:     int32(killer.SteamID32()),
					KillerX:      float32(killer.Position().X),
					KillerY:      float32(killer.Position().Y),
					KillerZ:      float32(killer.Position().Z),
					VictimName:   dead.Name,
					VictimID:     int32(dead.SteamID32()),
					VictimX:      float32(dead.Position().X),
					VictimY:      float32(dead.Position().Y),
					VictimZ:      float32(dead.Position().Z),
					Weapon:       weaponName,
					IsHeadshot:   e.IsHeadshot,
					RoundNum:     int32(demo.GameState().TotalRoundsPlayed()),
					KillerTeam:   killerTeamName,
					KillerTeamID: teamID,
				}

				// Map key: playerID (SteamID32) | playerName | teamID | side  — always unique per player per side
				mapKey := fmt.Sprintf("%s|%s|%s|%s", playerID, killer.Name, teamID, side)

				if _, ok := playerKills[mapKey]; !ok {
					playerKills[mapKey] = &protos.KillDataList{}
				}
				playerKills[mapKey].Kills = append(playerKills[mapKey].Kills, killData)
			}
		}
	})
	var mapName string
	demo.RegisterNetMessageHandler(func(m *msg.CDemoFileHeader) {
		mapName = m.GetMapName()
	})
	err = demo.ParseToEnd()
	if err != nil {
		return err
	}

	basename := filepath.Base(filename)
	basename = strings.TrimSuffix(basename, ".dem")
	timestampPrefix := mod_time.Format("20060102_150405")

	// Ensure the kills directory exists
	err = os.MkdirAll(killsDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create kills directory: %v", err)
	}

	matchMeta := MatchMetadata{
		MatchID: basename,
		Date:    dateStr,
		Map:     mapName,
		Players: make(map[string]PlayerMetadata),
	}

	for mapKey, killDataList := range playerKills {

		parts := strings.Split(mapKey, "|")

		playerID := parts[0]   // SteamID32 string
		playerName := parts[1] // human-readable name
		teamID := parts[2]
		side := parts[3]

		marshaled, err := proto.Marshal(killDataList)
		if err != nil {
			log.Printf("failed to marshal kills for %s: %v", playerName, err)
			continue
		}

		// Format: data/kills/{YYYYMMDD_HHMMSS}_{demoname}_team-{teamID}_side-{side}_player-{steamID32}.pb
		pbFilename := filepath.Join(killsDir, fmt.Sprintf("%s_%s_team-%s_side-%s_player-%s.pb", timestampPrefix, basename, teamID, side, playerID))

		err = WritePos(pbFilename, marshaled)
		if err != nil {
			log.Printf("failed to write file %s: %v", pbFilename, err)
			continue
		}
		fmt.Printf("  wrote: %s\n", pbFilename)

		// Accumulate both sides per player — initialise the entry if needed.
		entry, exists := matchMeta.Players[playerID]
		if !exists {
			entry = PlayerMetadata{
				Name:       playerName,
				TeamID:     teamID,
				KillsPaths: make(map[string]string),
			}
		}
		entry.KillsPaths[side] = pbFilename
		matchMeta.Players[playerID] = entry
	}

	// Append metadata to data/matches.jsonl
	fIndex, err := os.OpenFile(matchesFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open metadata index %s: %v", matchesFile, err)
	}
	defer fIndex.Close()

	metaBytes, err := json.Marshal(matchMeta)
	if err != nil {
		return fmt.Errorf("failed to marshal match metadata: %v", err)
	}

	_, err = fIndex.Write(append(metaBytes, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write to metadata index: %v", err)
	}

	return nil
}

func ReadKills(filename string) (*protos.KillDataList, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	kills := &protos.KillDataList{}
	if err := proto.Unmarshal(data, kills); err != nil {
		return nil, err
	}
	return kills, nil
}

// LoadMatches reads data/matches.jsonl and returns all MatchMetadata records.
// Reads sequentially line by line — fast for thousands of matches, no full
// directory listing required.
func LoadMatches(indexFile string) ([]MatchMetadata, error) {
	f, err := os.Open(indexFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var matches []MatchMetadata
	dec := json.NewDecoder(f)
	for dec.More() {
		var m MatchMetadata
		if err := dec.Decode(&m); err != nil {
			return nil, fmt.Errorf("malformed line in %s: %v", indexFile, err)
		}
		matches = append(matches, m)
	}
	return matches, nil
}

func PrintDemo(filename string) error {
	var killCount int
	var deathCount int

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	// var mapMetadata ex.Map
	demo := demos.NewParser(f)
	defer demo.Close()
	demo.RegisterNetMessageHandler(func(msg *msg.CDemoFileHeader) {
		// Get metadata for the map that the game was played on for coordinate translations
		fmt.Println(msg.GetMapName())
		// mapMetadata = ex.GetMapMetadata(msg.GetMapName())

	})
	demo.RegisterEventHandler(func(e events.RoundStart) {
		gs := demo.GameState()
		fmt.Printf("Round %d started\n\n", gs.TotalRoundsPlayed()+1)
	})
	demo.RegisterEventHandler(func(e events.Kill) {
		var printString string
		fmt.Printf("KILL\t")
		if e.Killer == nil {
			printString = fmt.Sprint(printString, "\x1b[31mnil killer\x1b[0m")
		} else {
			printString = fmt.Sprint(printString, e.Killer)
			killCount++
		}
		printString = fmt.Sprint(printString, " killed ")
		if e.Victim == nil {
			printString = fmt.Sprint(printString, "nil victim")
		} else {
			printString = fmt.Sprint(printString, e.Victim)
			deathCount++
		}
		printString = fmt.Sprint(printString, " with ")
		if e.Weapon == nil {
			printString = fmt.Sprint(printString, "nil weapon")
		} else {
			printString = fmt.Sprint(printString, e.Weapon)
		}
		printString = fmt.Sprint(printString, "\n")
		fmt.Print(printString)
		// if e.Killer != nil && e.Victim != nil {
		// 	fmt.Printf("%s killed %s with %s\n", e.Killer.Name, e.Victim.Name, e.Weapon)
		// } else if e.Killer == nil && e.Victim != nil {
		// 	fmt.Printf("Killer does not exist but Victim does %s\n", e.Victim.Name)
		// }

	})
	demo.RegisterEventHandler(func(e events.BombEvent) {
		fmt.Printf("BOMB\t")
		fmt.Printf("Bomb event at %s\n", string(e.Site))
	})
	demo.RegisterEventHandler(func(e events.MatchStart) {
		fmt.Printf("Match started\n\n")
	})
	demo.RegisterEventHandler(func(e events.AnnouncementMatchStarted) {
		fmt.Println("ANNOUNCEMENT MATCH STARTED")
	})
	demo.RegisterEventHandler(func(e events.OtherDeath) {
		fmt.Printf("OTHER DEATH %s killed %s\n", e.Killer.Name, e.OtherType)
	})
	demo.RegisterEventHandler(func(e events.RoundEnd) {
		fmt.Printf("\nRound end\n")
	})
	demo.RegisterEventHandler(func(e events.BombPlanted) {
		fmt.Printf("BOMB\t")
		fmt.Printf("Bomb planted at %s by %s\n", string(e.Site), string(e.Player.Name))
	})
	demo.RegisterEventHandler(func(e events.BombExplode) {
		fmt.Printf("BOMB\t")
		fmt.Printf("Bomb exploded at %s by %s	\n", string(e.Site), string(e.Player.Name))
	})

	err = demo.ParseToEnd()
	if err != nil {
		return err
	}
	// fmt.Println(mapMetadata.PosX, mapMetadata.PosY, mapMetadata.Scale)
	fmt.Printf("\x1b[38;5;196m%20s\t%-6d\x1b[0m\n", "Kill Count: ", killCount)
	fmt.Printf("\x1b[38;5;220m%20s\t%-6d\x1b[0m\n", "Death Count: ", deathCount)

	return nil
}
