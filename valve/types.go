package valve

import (
	"net"
)

// A list of IP addresses and ports.
type ServerList []*net.TCPAddr

// Implements Batch.Len().
func (this ServerList) Len() int {
	return len(this)
}

// Implements Batch.Item().
func (this ServerList) Item(index int) interface{} {
	return this[index]
}

// The game engine (either HL1 or HL2).
type GameEngine int

const (
	GOLDSRC GameEngine = GameEngine(1)
	SOURCE  GameEngine = GameEngine(2)
)

// The server type (either dedicated or listen).
type ServerType int

const (
	ServerType_Unknown ServerType = iota
	ServerType_Dedicated
	ServerType_Listen
)

// Returns the server type as a string.
func (this ServerType) String() string {
	switch this {
	case ServerType_Dedicated:
		return "dedicated"
	case ServerType_Listen:
		return "listen"
	default:
		return "unknown"
	}
}

// The server operating system (windows, linux, or mac).
type ServerOS int

const (
	ServerOS_Unknown ServerOS = iota
	ServerOS_Windows
	ServerOS_Linux
	ServerOS_Mac
)

// Returns the operating system as a string.
func (this ServerOS) String() string {
	switch this {
	case ServerOS_Windows:
		return "windows"
	case ServerOS_Linux:
		return "linux"
	case ServerOS_Mac:
		return "mac"
	default:
		return "unknown"
	}
}

// Official versions of the A2S_INFO reply.
const A2S_INFO_GOLDSRC uint8 = 0x6d
const A2S_INFO_SOURCE uint8 = 0x49

// Optional mod information returned by A2S_INFO_GOLDSRC.
type ModInfo struct {
	Url     string
	DwlUrl  string
	Version uint32
	Size    uint32
	Type    uint8
	Dll     uint8
}

// Optional information returned by App_TheShip.
type TheShipInfo struct {
	Mode      uint8
	Witnesses uint8
	Duration  uint8
}

// Optional information available with A2S_INFO_SOURCE.
type SpecTvInfo struct {
	Port uint16
	Name string
}

// Optional information available with A2S_INFO_SOURCE.
type AppInfo struct {
	AppType uint8
	ModId   uint32
}

// Optional information available with A2S_INFO_SOURCE. This isa  grab-bag
// of various optional bits. If some are not present they are left as 0.
// In the future this may change to distinguish from being present as 0.
type ExtendedInfo struct {
	AppId               AppId
	GameVersion         string
	Port                uint16 // 0 if not present.
	SteamId             uint64 // 0 if not present.
	GameModeDescription string // "" if not present.
	AppInfo             *AppInfo
}

// Information returned by an A2S_INFO query. Most of this is returned as-is
// from the wire, except where otherwise noted.
type ServerInfo struct {
	// The address can be arbitrary in older replies; for Source servers, it is
	// computed as the address and port used to connect. It should not be relied
	// on for reconnecting to the server.
	Address string

	// One of the A2S_INFO constants.
	InfoVersion uint8

	Protocol   uint8
	Name       string
	MapName    string
	Folder     string
	Game       string
	Players    uint8
	MaxPlayers uint8
	Bots       uint8
	Type       ServerType
	OS         ServerOS
	Visibility uint8
	Vac        uint8
	Mod        *ModInfo
	TheShip    *TheShipInfo
	SpecTv     *SpecTvInfo
	Ext        *ExtendedInfo
}

// Attempt to guess the game engine version.
func (this *ServerInfo) GameEngine() GameEngine {
	if this.InfoVersion == A2S_INFO_GOLDSRC || this.Ext == nil {
		return GOLDSRC
	}
	if uint32(this.Ext.AppId) < 80 {
		return GOLDSRC
	}
	return SOURCE
}
