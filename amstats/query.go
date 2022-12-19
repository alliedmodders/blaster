// vim: set ts=4 sw=4 tw=99 noet:
//
// Blaster (C) Copyright 2014 AlliedModders LLC
// Licensed under the GNU General Public License, version 3 or higher.
// See LICENSE.txt for more details.
package main

import (
	"fmt"
	"net"
	"os"

	"github.com/alliedmodders/blaster/batch"
	"github.com/alliedmodders/blaster/valve"
)

type ServerCallback func(server *Server)

type Server struct {
	Info  *valve.ServerInfo
	Rules map[string]string
}

func (this *Server) DbType() int64 {
	switch this.Info.Type {
	case valve.ServerType_Listen:
		return 3
	case valve.ServerType_Dedicated:
		switch this.Info.OS {
		case valve.ServerOS_Windows:
			return 2
		case valve.ServerOS_Linux:
			return 1
		case valve.ServerOS_Mac:
			return 4
		}
	}
	return 0
}

func queryMaster(db *Database, game_id int64, callback ServerCallback) {
	// Create a connection to the master server.
	master, err := valve.NewMasterServerQuerier(valve.MasterServer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not query master: %s", err.Error())
	}
	defer master.Close()

	// Set up the filter list.
	switch game_id {
	case 1:
		master.FilterAppIds(valve.HL1Apps)
	case 2:
		master.FilterAppIds(valve.HL2Apps)
	default:
		panic("unknown game_id")
	}

	bp := batch.NewBatchProcessor(func(item interface{}) {
		server, err := queryServer(item.(*net.TCPAddr))
		if server == nil {
			if err != nil {
				callback(nil)
			}
			return
		}
		callback(server)
	}, 20)
	defer bp.Terminate()

	// Query the master.
	err = master.Query(func(servers valve.ServerList) error {
		bp.AddBatch(servers)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not query master: %s\n", err.Error())
		os.Exit(1)
	}

	// Wait for back processing to complete.
	bp.Finish()
}

func queryServer(addr *net.TCPAddr) (*Server, error) {
	query, err := valve.NewServerQuerier(addr.String(), kTimeout)
	if err != nil {
		return nil, err
	}
	defer query.Close()

	info, err := query.QueryInfo()
	if err != nil {
		return nil, err
	}

	rules, err := query.QueryRules()
	if err != nil {
		return nil, err
	}

	return &Server{
		Info:  info,
		Rules: rules,
	}, nil
}
