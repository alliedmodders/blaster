// vim: set ts=4 sw=4 tw=99 noet:
//
// Blaster (C) Copyright 2014 AlliedModders LLC
// Licensed under the GNU General Public License, version 3 or higher.
// See LICENSE.txt for more details.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	batch "github.com/alliedmodders/blaster/batch"
	valve "github.com/alliedmodders/blaster/valve"
)

var (
	sOutputLock   sync.Mutex
	sOutputBuffer io.Writer
	sOutputFormat string
	sNumServers   int64
)

type ErrorObject struct {
	Ip    string `json:"ip"`
	Error string `json:"error"`
}

type ServerObject struct {
	Address      string `json:"ip"`
	LocalAddress string `json:"local_ip,omitempty"`
	Protocol     uint8  `json:"protocol"`
	Name         string `json:"name"`
	MapName      string `json:"map"`
	Folder       string `json:"folder"`
	Game         string `json:"game"`
	Players      uint8  `json:"players"`
	MaxPlayers   uint8  `json:"max_players"`
	Bots         uint8  `json:"bots"`
	Type         string `json:"type"`
	Os           string `json:"os"`
	Visibility   string `json:"visibility"`
	Vac          bool   `json:"vac"`

	// Only available from The Ship.
	Ship *valve.TheShipInfo `json:"theship,omitempty"`

	// Only available on Source.
	AppId       valve.AppId `json:"appid,omitempty"`
	GameVersion string      `json:"game_version,omitempty"`
	Port        uint16      `json:"port,omitempty"`
	SteamId     string      `json:"steamid,omitempty"`
	GameMode    string      `json:"game_mode,omitempty"`
	GameId      string      `json:"gameid,omitempty"`
	SpecTvPort  uint16      `json:"spectv_port,omitempty"`
	SpecTvName  string      `json:"spectv_name,omitempty"`

	// Only available on Half-Life 1.
	Mod *valve.ModInfo `json:"mod,omitempty"`

	Rules map[string]string `json:"rules"`
}

func addJson(hostAndPort string, obj interface{}) {
	buf, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	sOutputLock.Lock()
	defer sOutputLock.Unlock()

	switch sOutputFormat {
	case "lines":
		if sNumServers != 0 {
			sOutputBuffer.Write([]byte("\n"))
		}

		sOutputBuffer.Write(buf)

		sNumServers++
		return
	}

	if sNumServers != 0 {
		sOutputBuffer.Write([]byte(",\n"))
	}
	sOutputBuffer.Write([]byte("\t"))

	var indented bytes.Buffer
	json.Indent(&indented, buf, "\t", "\t")

	switch sOutputFormat {
	case "map":
		sOutputBuffer.Write([]byte(fmt.Sprintf("\"%s\": ", hostAndPort)))
	}

	indented.WriteTo(sOutputBuffer)
	sNumServers++
}

func addError(hostAndPort string, err error) {
	addJson(hostAndPort, &ErrorObject{
		Ip:    hostAndPort,
		Error: err.Error(),
	})
}

func main() {
	flag_game := flag.String("game", "", "Game (hl1, hl2)")
	flag_appid := flag.Int("appid", 0, "Query a single AppID")
	flag_appids := flag.String("appids", "", "Comma-delimited list of AppIDs")
	flag_master := flag.String("master", valve.MasterServer, "Master server address")
	flag_j := flag.Int("j", 20, "Number of concurrent requests (more will introduce more timeouts)")
	flag_timeout := flag.Duration("timeout", time.Second*3, "Timeout for querying servers")
	flag_format := flag.String("format", "list", "JSON format (list, map, or lines)")
	flag_outfile := flag.String("outfile", "", "Output to a file")
	flag_norules := flag.Bool("norules", false, "Don't query server rules")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: -game or -appids\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	appids := []valve.AppId{}

	switch *flag_format {
	case "list", "map", "lines":
		sOutputFormat = *flag_format
	default:
		fmt.Fprintf(os.Stderr, "Unknown format type.\n")
		os.Exit(1)
	}

	if *flag_outfile != "" {
		file, err := os.Create(*flag_outfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not open %s for writing: %s\n", *flag_outfile, err.Error())
			os.Exit(1)
		}
		defer file.Close()

		sOutputBuffer = file
	} else {
		sOutputBuffer = os.Stdout
	}

	if *flag_game != "" {
		switch *flag_game {
		case "hl1":
			appids = append(appids, valve.HL1Apps...)
		case "hl2":
			appids = append(appids, valve.HL2Apps...)
		default:
			fmt.Fprintf(os.Stderr, "Unrecognized game: %s", *flag_game)
			os.Exit(1)
		}
	}

	if *flag_appids != "" {
		for _, part := range strings.Split(*flag_appids, ",") {
			appid, err := strconv.Atoi(part)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\"%s\" is not a valid AppID\n", part)
				os.Exit(1)
			}
			appids = append(appids, valve.AppId(appid))
		}
	}

	if *flag_appid != 0 {
		appids = append(appids, valve.AppId(*flag_appid))
	}

	if len(appids) == 0 {
		fmt.Fprintf(os.Stderr, "At least one AppID or game must be specified.\n")
		os.Exit(1)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	// Create a connection to the master server.
	master, err := valve.NewMasterServerQuerier(*flag_master)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not query master: %s", err.Error())
	}
	defer master.Close()

	// Set up the filter list.
	master.FilterAppIds(appids)

	// Initialize our batch processor, which will receive servers and query them
	// concurrently.
	bp := batch.NewBatchProcessor(func(item interface{}) {
		addr := item.(*net.TCPAddr)
		query, err := valve.NewServerQuerier(addr.String(), *flag_timeout)
		if err != nil {
			addError(addr.String(), err)
			return
		}
		defer query.Close()

		info, err := query.QueryInfo()
		if err != nil {
			addError(addr.String(), err)
			return
		}

		out := &ServerObject{
			Address:    addr.String(),
			Protocol:   info.Protocol,
			Name:       info.Name,
			MapName:    info.MapName,
			Folder:     info.Folder,
			Game:       info.Game,
			Players:    info.Players,
			MaxPlayers: info.MaxPlayers,
			Bots:       info.Bots,
			Type:       info.Type.String(),
			Os:         info.OS.String(),
			Ship:       info.TheShip,
			Mod:        info.Mod,
		}
		if info.Vac == 1 {
			out.Vac = true
		}
		if info.Visibility == 0 {
			out.Visibility = "public"
		} else {
			out.Visibility = "private"
		}
		if info.Ext != nil {
			out.AppId = info.Ext.AppId
			out.GameVersion = info.Ext.GameVersion
			out.Port = info.Ext.Port
			out.SteamId = fmt.Sprintf("%d", info.Ext.SteamId)
			out.GameMode = info.Ext.GameModeDescription
			out.GameId = fmt.Sprintf("%d", info.Ext.GameId)
		}
		if info.InfoVersion == valve.A2S_INFO_GOLDSRC {
			out.LocalAddress = info.Address
		}
		if info.SpecTv != nil {
			out.SpecTvPort = info.SpecTv.Port
			out.SpecTvName = info.SpecTv.Name
		}

		// We can't query rules for CSGO servers anymore because Valve.
		csgo := (info.Ext != nil && info.Ext.AppId == valve.App_CSGO)
		if !csgo && !*flag_norules {
			rules, err := query.QueryRules()
			if err != nil {
				out.Rules = map[string]string{
					"error": err.Error(),
				}
			} else {
				out.Rules = rules
			}
		}

		addJson(addr.String(), out)
	}, *flag_j)
	defer bp.Terminate()

	switch sOutputFormat {
	case "list":
		sOutputBuffer.Write([]byte("[\n"))
	case "map":
		sOutputBuffer.Write([]byte("{\n"))
	}

	// Query the master.
	err = master.Query(func(servers valve.ServerList) error {
		bp.AddBatch(servers)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not query the master: %s\n", err.Error())
		os.Exit(1)
	}

	// Wait for batch processing to complete.
	bp.Finish()

	if sNumServers != 0 {
		sOutputBuffer.Write([]byte("\n"))
	}

	switch sOutputFormat {
	case "list":
		sOutputBuffer.Write([]byte("]\n"))
	case "map":
		sOutputBuffer.Write([]byte("}\n"))
	}
}
