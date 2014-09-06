blaster
=======

Blaster is a tool for querying servers from the Valve Master Server List. There are three components: a set of libraries for querying Valve protocols (which have many edge cases), a concurrenct batch-processing library, and a command-line tool for getting query results as JSON.

Valve's master server has a rate limit of about 15 queries per minute, and returns a batch of ~220 servers for each query. For a popular game, it can take a long time (around ten minutes) to retrieve its entire server list. Blaster will query individual game servers in the background to lessen the overall waiting time. At the moment it will process 20 servers in the background, concurrently. Go's scheduler is still weak so it's not recommended to use more.

Usage
-----
You can use blaster either across all of Half-Life 1 or Half-Life 2, or with a specific list of Application IDs. For a full list of Application IDs, see: https://developer.valvesoftware.com/wiki/Steam_Application_IDs

All output is in JSON format.

For example, querying "Bloody Good Time":
```
$ go run blaster.go -appids 2450 -norules -format=list
[
	{
		"ip": "168.62.205.3:27016",
		"protocol": 17,
		"name": "DMServer",
		"map": "horrorhouse",
		"folder": "pm",
		"game": "Bloody Good Time",
		"players": 9,
		"max_players": 16,
		"bots": 6,
		"type": "dedicated",
		"os": "windows",
		"visibility": "public",
		"vac": true,
		"appid": 2450,
		"game_version": "1.0.0.0",
		"port": 27016,
		"steamid": "90091830459546624",
		"gameid": "2450",
		"rules": null
	}
]
```

Building
--------

1. Make sure you have Golang installed, (see: http://golang.org/)
2. Make sure your Go environment is set up. Example:

        export GOROOT=~/tools/go
        export GOPATH=~/go
        export PATH="$PATH:$GOROOT/bin:$GOPATH/bin"

3. Get the source code and its dependencies:

        go get github.com/alliedmodders/blaster

4. Build:

        go install

5. The `blaster` binary wll be in `$GOPATH/bin/`.

Resources
---------
https://developer.valvesoftware.com/wiki/Master_Server_Query_Protocol
https://developer.valvesoftware.com/wiki/Server_queries
