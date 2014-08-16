package valve

import (
	"errors"
	"fmt"
	"time"
)

var ErrBadPacketHeader = errors.New("bad packet header")
var ErrNonSteamReply = errors.New("non-steam reply")
var ErrUnknownInfoVersion = errors.New("unknown A2S_INFO version")

// Always import fmt for debugging.
var _ = fmt.Println

type ServerQuerier struct {
	socket *UdpSocket
}

func NewServerQuerier(hostAndPort string, timeout time.Duration) (*ServerQuerier, error) {
	socket, err := NewUdpSocket(hostAndPort, timeout)
	if err != nil {
		return nil, err
	}
	return &ServerQuerier{
		socket: socket,
	}, nil
}

func (this *ServerQuerier) Close() {
	this.socket.Close()
}

func (this *ServerQuerier) QueryInfo() (*ServerInfo, error) {
	info := &ServerInfo{
		Address: this.socket.RemoteAddr().String(),
	}

	err := Try(func() error {
		return this.a2s_info(info)
	})
	if err == ErrNonSteamReply {
		// Try again.
		err = Try(func() error {
			return this.a2s_info(info)
		})
	}

	if err != nil {
		return nil, err
	}
	return info, nil
}

func (this *ServerQuerier) a2s_info(info *ServerInfo) error {
	if err := this.sendInfoQuery(); err != nil {
		return err
	}

	data, err := this.socket.Recv()
	if err != nil {
		return err
	}

	reader := NewPacketReader(data)
	if reader.ReadInt32() != -1 {
		return ErrBadPacketHeader
	}

	info.InfoVersion = reader.ReadUint8()
	switch info.InfoVersion {
	case 0x44:
		// Non-steam servers seem to reply with a corrupted packet. If we send
		// the query again, we often get the right thing, so propagate a special
		// error up.
		return ErrNonSteamReply
	case A2S_INFO_SOURCE:
		this.parseNewInfo(reader, info)
	case A2S_INFO_GOLDSRC:
		this.parseOldInfo(reader, info)
	default:
		return ErrUnknownInfoVersion
	}
	return nil
}

func (this *ServerQuerier) sendInfoQuery() error {
	var packet PacketBuilder
	packet.WriteBytes([]byte{0xff, 0xff, 0xff, 0xff, 0x54})
	packet.WriteCString("Source Engine Query")
	return this.socket.Send(packet.Bytes())
}

func (this *ServerQuerier) parseNewInfo(reader *PacketReader, info *ServerInfo) {
	info.Protocol = reader.ReadUint8()
	info.Name = reader.ReadString()
	info.MapName = reader.ReadString()
	info.Folder = reader.ReadString()
	info.Game = reader.ReadString()

	// This gets extended later, potentially.
	appId := AppId(reader.ReadUint16())

	info.Players = reader.ReadUint8()
	info.MaxPlayers = reader.ReadUint8()
	info.Bots = reader.ReadUint8()

	serverType := reader.ReadUint8()
	switch serverType {
	case uint8('l'):
		info.Type = ServerType_Listen
	case uint8('d'):
		info.Type = ServerType_Dedicated
	default:
		info.Type = ServerType_Unknown
	}

	serverOS := reader.ReadUint8()
	switch serverOS {
	case uint8('l'):
		info.OS = ServerOS_Linux
	case uint8('w'):
		info.OS = ServerOS_Windows
	case uint8('m'):
		info.OS = ServerOS_Mac
	default:
		info.OS = ServerOS_Unknown
	}

	info.Visibility = reader.ReadUint8()
	info.Vac = reader.ReadUint8()

	// Read TheShip information.
	if AppId(appId) == App_TheShip {
		info.TheShip = &TheShipInfo{}
		info.TheShip.Mode = reader.ReadUint8()
		info.TheShip.Witnesses = reader.ReadUint8()
		info.TheShip.Duration = reader.ReadUint8()
	}

	info.Ext = &ExtendedInfo{
		AppId: appId,
	}

	// Start reading extended information.
	info.Ext.GameVersion = reader.ReadString()
	if !reader.More() {
		return
	}

	edf := reader.ReadUint8()
	if (edf & 0x80) != 0 {
		info.Ext.Port = reader.ReadUint16()
	}
	if (edf & 0x10) != 0 {
		info.Ext.SteamId = reader.ReadUint64()
	}
	if (edf & 0x40) != 0 {
		info.SpecTv = &SpecTvInfo{}
		info.SpecTv.Port = reader.ReadUint16()
		info.SpecTv.Name = reader.ReadString()
	}
	if (edf & 0x20) != 0 {
		info.Ext.GameModeDescription = reader.ReadString()
	}
	if (edf & 0x01) != 0 {
		gameId := reader.ReadUint64()

		// bits 0-23: true app id (original could be truncated)
		// bits 24-31: type
		// bits 32-63: mod id
		info.Ext.AppId = AppId(gameId & uint64(0xffffffff))
		info.Ext.AppInfo = &AppInfo{}
		info.Ext.AppInfo.AppType = uint8((gameId >> 24) & 0xff)
		info.Ext.AppInfo.ModId = uint32((gameId >> 32) & 0xffffffff)
	}
}

func (this *ServerQuerier) parseOldInfo(reader *PacketReader, info *ServerInfo) {
	info.Address = reader.ReadString()
	info.Name = reader.ReadString()
	info.MapName = reader.ReadString()
	info.Folder = reader.ReadString()
	info.Game = reader.ReadString()
	info.Players = reader.ReadUint8()
	info.MaxPlayers = reader.ReadUint8()
	info.Protocol = reader.ReadUint8()

	serverType := reader.ReadUint8()
	switch serverType {
	case uint8('l'):
		info.Type = ServerType_Listen
	case uint8('d'):
		info.Type = ServerType_Dedicated
	default:
		info.Type = ServerType_Unknown
	}

	serverOS := reader.ReadUint8()
	switch serverOS {
	case uint8('l'):
		info.OS = ServerOS_Linux
	case uint8('w'):
		info.OS = ServerOS_Windows
	default:
		info.OS = ServerOS_Unknown
	}

	info.Visibility = reader.ReadUint8()

	isMod := reader.ReadUint8()
	if isMod != 1 {
		return
	}

	info.Mod = &ModInfo{}
	info.Mod.Url = reader.ReadString()
	info.Mod.DwlUrl = reader.ReadString()
	reader.ReadUint8() // Ignore a null byte.
	info.Mod.Version = reader.ReadUint32()
	info.Mod.Size = reader.ReadUint32()
	info.Mod.Type = reader.ReadUint8()
	info.Mod.Dll = reader.ReadUint8()

	// This data is only exposed through is_mod == 1, but we put it in general
	// info anyway.
	info.Vac = reader.ReadUint8()
	info.Bots = reader.ReadUint8()
}
