blaster
=======

Source and GoldSrc Query Tool

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
