package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	internal "github.com/TheBigJayT/round-two-cs/internal"
	filters "github.com/TheBigJayT/round-two-cs/internal/filters"
	rw "github.com/TheBigJayT/round-two-cs/internal/readwrite"

	"github.com/golang/geo/r3"
)

func main() {
	mode := flag.String("mode", "read", "Mode of operation: 'extract', 'debug', or 'read'")
	demoFile := flag.String("demo", "", "Path to the demo file to extract kills from (required for extract and debug mode)")
	teamFlag := flag.String("team", "", "Team name to filter by (resolves via data/teams.json)")
	playerFlag := flag.String("player", "", "Player name or SteamID32 to filter by (resolves via data/players.json)")
	sideFlag := flag.String("side", "", "Side to filter by: T or CT")
	mapFlag := flag.String("map", "", "Map name to filter by (substring match on actual map name, e.g. dust2, inferno)")
	dateFlag := flag.String("date", "", "Exact date to filter by (YYYY-MM-DD)")
	dateFromFlag := flag.String("date-from", "", "Start of date range, inclusive (YYYY-MM-DD)")
	dateToFlag := flag.String("date-to", "", "End of date range, inclusive (YYYY-MM-DD)")
	outputKills := flag.Bool("list", true, "Print matching kills to stdout")
	isNewFeature := flag.Bool("new", false, "INTERNALLY used for testing before implementing")
	showPositions := flag.Bool("pos", false, "Show positions (mainly for debugging)")
	flag.Parse()

	if *isNewFeature {
		filter := filters.Filter{Player: *playerFlag, Side: *sideFlag, Map: *mapFlag, Team: *teamFlag}
		matches, err := filter.ResolveFilter()
		if err != nil {
			log.Fatal(err)
		}
		testerr := rw.ToCSV(matches)
		if testerr != nil {
			log.Fatal(testerr)
		}
		return
	}

	if *mode == "read" {
		filter := filters.Filter{Player: *playerFlag, Side: *sideFlag, Map: *mapFlag, Team: *teamFlag}
		matches, err := filter.ResolveFilter()
		if err != nil {
			log.Fatal(err)
		}
		var count int32
		for _, match := range matches {
			list, _ := rw.ReadKills(match)
			for _, j := range list.GetKills() {
				count++
				fmt.Printf("%-20s killed %-20s using %-20s assisted by %-20s\n", j.GetKillerName(), j.GetVictimName(), j.GetWeapon(), j.GetAssisterName())
			}
		}
		fmt.Println("\x1b[1;31mKills: ", count)
		return
	}

	// Debug mode prints kills in plain text to stdout, SUGGEST piping to a .txt
	if *mode == "debug" {
		if *demoFile == "" {
			log.Fatalf("debug mode requires -demo <path/to/file.dem>")

		}
		fmt.Printf("Outputing entire demo: %s\n", *demoFile)
		rw.PrintDemo(*demoFile)
		return
	}

	// Extract takes the path specified by -demo, parses said demo, and extracts kills to .pb files.
	if *mode == "extract" {
		if *demoFile == "" {
			log.Fatalf("extract mode requires -demo <path/to/file.dem>")
		}
		fmt.Printf("Extracting kills from demo: %s\n", *demoFile)
		err := rw.ExtractKillsData(*demoFile)
		if err != nil {
			log.Fatalf("failed to extract kills: %v", err)
		}
		fmt.Println("Success!")
		return
	}

	// Read mode is the default, generally does not have to be specified.
	// Reads kills from the .pb files according to certain filters defined
	// by other flags.
	if *mode != "read" {
		log.Fatalf("Unknown mode: %s. Use 'extract' or 'read' or 'debug'", *mode)
	}

	// --- resolve team filter ---
	// *teamFlag is the string of the team specified using -team
	teamID := ""
	if *teamFlag != "" {
		safeTeamName := strings.ReplaceAll(strings.ToLower(*teamFlag), " ", "-")
		data, err := os.ReadFile("data/teams.json")
		if err == nil {
			var teams map[string]string
			json.Unmarshal(data, &teams)
			if id, ok := teams[safeTeamName]; ok {
				teamID = id
				fmt.Printf("Team filter : '%s' → ID %s\n", safeTeamName, teamID)
			} else {
				teamID = safeTeamName
				fmt.Printf("Team filter : '%s' (no mapping found, using raw)\n", safeTeamName)
			}
		}
	}

	// --- resolve player filter ---
	playerID := ""
	playerDisplayName := ""
	if *playerFlag != "" {
		// Check if it looks like a numeric SteamID32
		isNumeric := true
		// Check the ascii for each character in the string
		for _, c := range *playerFlag {
			if c < '0' || c > '9' {
				isNumeric = false
				break
			}
		}
		if isNumeric {
			playerID = *playerFlag
			playerDisplayName = *playerFlag
		} else {
			// Search data/players.json by name
			data, err := os.ReadFile("data/players.json")
			if err == nil {
				var players map[string]internal.PlayerInfo
				json.Unmarshal(data, &players)
				nameLower := strings.ToLower(*playerFlag)
				for id, info := range players {
					if strings.ToLower(info.Name) == nameLower {
						playerID = id
						playerDisplayName = info.Name
						break
					}
				}
			}
			if playerID == "" {
				log.Fatalf("Player '%s' not found in data/players.json", *playerFlag)
			}
		}
		fmt.Printf("Player filter: '%s' -> SteamID32 %s\n", playerDisplayName, playerID)
	}

	// --- resolve side filter ---
	sideFilter := strings.ToUpper(*sideFlag)
	if sideFilter != "" && sideFilter != "T" && sideFilter != "CT" {
		log.Fatalf("Invalid side '%s'. Use T or CT", *sideFlag)
	}
	if sideFilter != "" {
		fmt.Printf("Side filter  : %s\n", sideFilter)
	}

	// --- resolve map filter ---
	mapFilter := strings.ToLower(*mapFlag)
	if mapFilter != "" {
		fmt.Printf("Map filter   : %s\n", mapFilter)
	}

	// --- resolve date range filter ---
	dateExact := *dateFlag
	dateFrom := *dateFromFlag
	dateTo := *dateToFlag
	if dateExact != "" {
		// --date is shorthand for --date-from=X --date-to=X
		dateFrom = dateExact
		dateTo = dateExact
		fmt.Printf("Date filter  : %s\n", dateExact)
	} else {
		if dateFrom != "" {
			fmt.Printf("Date from    : %s\n", dateFrom)
		}
		if dateTo != "" {
			fmt.Printf("Date to      : %s\n", dateTo)
		}
	}

	// --- load matches index ---
	indexFile := filepath.Join("data", "matches.jsonl")
	matches, err := rw.LoadMatches(indexFile)
	if err != nil {
		log.Fatalf("failed to load matches index %s: %v", indexFile, err)
	}
	if len(matches) == 0 {
		fmt.Println("No matches found in index. Run -mode=extract first.")
		return
	}

	fmt.Println("\nKills Summary:")
	fmt.Println(strings.Repeat("_", 110))
	fmt.Println()

	totalKills := 0
	fileCount := 0

	for _, match := range matches {
		// Date range filter — ISO 8601 strings compare lexicographically
		if dateFrom != "" && match.Date < dateFrom {
			continue
		}
		if dateTo != "" && match.Date > dateTo {
			continue
		}
		// Map filter — substring match on actual in-game map name (e.g. "dust2" matches "de_dust2")
		if mapFilter != "" && !strings.Contains(strings.ToLower(match.Map), mapFilter) {
			continue
		}

		for pid, playerMeta := range match.Players {
			// Player filter
			if playerID != "" && pid != playerID {
				continue
			}
			// Team filter
			if teamID != "" && playerMeta.TeamID != teamID {
				continue
			}

			for side, killsPath := range playerMeta.KillsPaths {
				// Side filter
				if sideFilter != "" && side != sideFilter {
					continue
				}

				kills, err := rw.ReadKills(killsPath)
				if err != nil || kills == nil || len(kills.Kills) == 0 {
					continue
				}

				if *outputKills {
					for _, y := range kills.GetKills() {
						killerV := r3.Vector{
							X: float64(y.GetKillerX()),
							Y: float64(y.GetKillerY()),
							Z: float64(y.GetKillerZ()),
						}
						victimV := r3.Vector{
							X: float64(y.GetVictimX()),
							Y: float64(y.GetVictimY()),
							Z: float64(y.GetVictimZ()),
						}
						distance := victimV.Distance(killerV)
						fmt.Println("Pixel: ", y.GetKillerXPixel(), y.GetKillerYPixel())
						fmt.Printf("%-20s killed\t%-20s using %-20s at %-20f units\n", y.GetKillerName(), y.GetVictimName(), y.GetWeapon(), distance)
						if *showPositions {
							fmt.Println(killerV)
						}

					}
				}
				fmt.Println(strings.Repeat("_", 110))
				fmt.Printf("\n%-20s  %3d kills       side:%-3s     map:%-12s\n[%s]\n",
					playerMeta.Name, len(kills.Kills), side, match.Map, filepath.Base(killsPath))
				fmt.Println(strings.Repeat("=", 110))
				fmt.Println()
				totalKills += len(kills.Kills)
				fileCount++
			}
		}
	}

	fmt.Printf("  %d file(s) matched — %d total kills\n", fileCount, totalKills)
}
