// vim: set ts=4 sw=4 tw=99 noet:
//
// Blaster (C) Copyright 2014 AlliedModders LLC
// Licensed under the GNU General Public License, version 3 or higher.
// See LICENSE.txt for more details.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/alliedmodders/blaster/valve"
	"github.com/kylelemons/go-gypsy/yaml"
)

var kTimeout time.Duration = time.Second * 3

func main() {
	flag_config := flag.String("config", "config.yml", "Config file path")
	flag_game := flag.String("game", "", "Game to query (hl1 or hl2)")
	flag.Parse()

	if *flag_config == "" {
		fmt.Fprintf(os.Stderr, "Must specify a config file via -config.\n")
		os.Exit(1)
	}

	game_id := int64(0)
	switch *flag_game {
	case "hl1":
		game_id = 1
	case "hl2":
		game_id = 2
	case "":
		fmt.Fprintf(os.Stderr, "Must specify a game via -game.\n")
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "Unrecognized game: %s\n", *flag_game)
		os.Exit(1)
	}

	cfg, err := yaml.ReadFile(*flag_config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read config file: %s\n", err.Error())
		os.Exit(1)
	}

	// Get the timeout value.
	if timeoutStr, err := cfg.Get("timeout"); err == nil {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			kTimeout = timeout
		}
	}

	// Get a database connection.
	db := getDatabase(cfg, "database")
	defer db.Close()

	// Make sure multithreading is enabled.
	runtime.GOMAXPROCS(runtime.NumCPU())

	queryStats(db, game_id)
}

func queryStats(db *Database, game_id int64) {
	lock := sync.Mutex{}

	collector := NewStatsCollector(db, game_id)

	queryMaster(db, game_id, func(server *Server) {
		lock.Lock()
		defer lock.Unlock()

		if server == nil {
			collector.global.DeadCount++
			return
		}

		modId := collector.getMod(server)

		tables := []*Stats{
			// Aggregate over game+mod.
			collector.get(StatsKey{ModId: modId, Type: server.DbType()}),
		}

		// Find any addons that apply.
		addons := map[int64]bool{}
		for key, value := range server.Rules {
			addonVar := collector.getAddonByVarName(key)
			if addonVar == nil {
				continue
			}

			found, _ := addons[addonVar.AddonId]
			if !found {
				// Aggregate over addon and over mod+type+addon.
				tables = append(tables,
					collector.get(StatsKey{AddonId: addonVar.AddonId}),
					collector.get(StatsKey{
						ModId:   modId,
						Type:    server.DbType(),
						AddonId: addonVar.AddonId,
					}),
				)
				addons[addonVar.AddonId] = true
			}

			// Aggregate over value and mod+type+value.
			valueId := collector.getValue(addonVar.VariableId, value)
			tables = append(tables,
				collector.get(StatsKey{ValueId: valueId}),
				collector.get(StatsKey{
					ModId:   modId,
					Type:    server.DbType(),
					ValueId: valueId,
				}),
			)
		}

		// Aggregate.
		for _, table := range tables {
			table.ServerCount++
			table.TotalPlayers += int64(server.Info.Players)
			table.MaxPlayers += int64(server.Info.MaxPlayers)
			table.TotalBots += int64(server.Info.Bots)
		}
		collector.global.AliveCount++
		collector.global.TotalPlayers += int64(server.Info.Players)
		collector.global.MaxPlayers += int64(server.Info.MaxPlayers)
		collector.global.TotalBots += int64(server.Info.Bots)

		// Global table needs extra data.
		switch server.Info.OS {
		case valve.ServerOS_Linux:
			collector.global.LinuxServers++
		case valve.ServerOS_Windows:
			collector.global.WindowsServers++
		}

		if server.Info.Type == valve.ServerType_Listen {
			collector.global.ListenServers++
		}
	})

	collector.finish()
}
