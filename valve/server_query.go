// vim: set ts=4 sw=4 tw=99 noet:
//
// Blaster (C) Copyright 2014 AlliedModders LLC
// Licensed under the GNU General Public License, version 3 or higher.
// See LICENSE.txt for more details.
package valve

import (
	"bytes"
	"compress/bzip2"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"time"
)

var ErrBadPacketHeader = errors.New("bad packet header")
var ErrMistakenReply = errors.New("mistaken reply")
var ErrUnknownInfoVersion = errors.New("unknown A2S_INFO version")
var ErrImmediateRulesReply = errors.New("immediate rules reply")
var ErrBadChallengeResponse = errors.New("bad challenge response")
var ErrUnknownGameEngine = errors.New("must query A2S_INFO first")
var ErrDuplicatePacket = errors.New("received duplicate numbered packets")
var ErrBadPacketNumber = errors.New("packet number is out of sequence")
var ErrConfusedChallengeReply = errors.New("challenge reply is for the wrong query")
var ErrBadRulesReply = errors.New("bad rules reply")
var ErrWrongBz2Size = errors.New("bad bz2 decompression size")
var ErrWrongBz2Checksum = errors.New("bad bz2 checksum")

// Always import fmt for debugging.
var _ = fmt.Println

// A ServerQuerier is used to issue A2S queries against an HL1/HL2 server.
type ServerQuerier struct {
	socket  *UdpSocket
	timeout time.Duration
	info    *ServerInfo
}

// Create a new server querying object.
func NewServerQuerier(hostAndPort string, timeout time.Duration) (*ServerQuerier, error) {
	socket, err := NewUdpSocket(hostAndPort, timeout)
	if err != nil {
		return nil, err
	}
	return &ServerQuerier{
		socket:  socket,
		timeout: timeout,
	}, nil
}

// Close the socket used to query.
func (this *ServerQuerier) Close() {
	this.socket.Close()
}

// Query a server's info via A2S_INFO.
func (this *ServerQuerier) QueryInfo() (*ServerInfo, error) {
	this.info = &ServerInfo{
		Address: this.socket.RemoteAddr().String(),
	}

	err := Try(func() error {
		return this.a2s_info(this.info)
	})
	if err != nil && err != ErrMistakenReply {
		return nil, err
	}

	// Mysteriously, Half-Life 1 servers will often reply to an A2S_INFO with
	// two extra packets: A2S_PLAYERS and then a newer A2S_INFO. We peek for
	// up to three extra packets with a very small timeout.
	if err == ErrMistakenReply || this.info.InfoVersion == S2A_INFO_GOLDSRC {
		err := Try(func() error {
			return this.check_bad_a2s_info(this.info)
		})
		if err == nil {
			return this.info, nil
		}
	}

	if err != nil {
		return nil, err
	}
	return this.info, nil
}

func (this *ServerQuerier) check_bad_a2s_info(info *ServerInfo) error {
	this.socket.SetTimeout(time.Millisecond * 250)
	defer this.socket.SetTimeout(this.timeout)

	data1, err := this.socket.Recv()
	if err != nil {
		return err
	}

	data2, err := this.socket.Recv()
	if err != nil {
		return err
	}

	// Try to decode either packet as an A2S_INFO.
	if this.parse_a2s_info_reply(this.info, data1) == nil {
		return nil
	}
	return this.parse_a2s_info_reply(this.info, data2)
}

func (this *ServerQuerier) a2s_info(info *ServerInfo) error {
	var packet PacketBuilder
	packet.WriteBytes([]byte{0xff, 0xff, 0xff, 0xff, A2S_INFO})
	packet.WriteCString("Source Engine Query")
	if err := this.socket.Send(packet.Bytes()); err != nil {
		return err
	}

	data, err := this.socket.Recv()
	if err != nil {
		return err
	}

	switch data[4] {
	case S2C_CHALLENGE:
		// The newer protocol requires A2S_INFO requests to contain a challenge,
		// servers that expected a challenge will have sent us a S2C_CHALLENGE response instead.
		// Re-send the query with the challenge we received.
		packet.WriteBytes([]byte{
			data[5], data[6], data[7], data[8],
		})
		if err := this.socket.Send(packet.Bytes()); err != nil {
			return err
		}

		data, err = this.socket.Recv()
		if err != nil {
			return err
		}
	}

	return this.parse_a2s_info_reply(info, data)
}

func (this *ServerQuerier) parse_a2s_info_reply(info *ServerInfo, data []byte) error {
	reader := NewPacketReader(data)
	if reader.ReadInt32() != -1 {
		return ErrBadPacketHeader
	}

	info.InfoVersion = reader.ReadUint8()
	switch info.InfoVersion {
	case S2A_PLAYER:
		// Non-steam servers seem to reply with a corrupted packet. If we send
		// the query again, we often get the right thing, so propagate a special
		// error up.
		return ErrMistakenReply
	case S2A_INFO_SOURCE:
		this.parseNewInfo(reader, info)
	case S2A_INFO_GOLDSRC:
		this.parseOldInfo(reader, info)
	default:
		return ErrUnknownInfoVersion
	}
	return nil
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
		info.Ext.GameId = gameId
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
	if isMod == 1 {
		info.Mod = &ModInfo{}
		info.Mod.Url = reader.ReadString()
		info.Mod.DwlUrl = reader.ReadString()
		reader.ReadUint8() // Ignore a null byte.
		info.Mod.Version = reader.ReadUint32()
		info.Mod.Size = reader.ReadUint32()
		info.Mod.Type = reader.ReadUint8()
		info.Mod.Dll = reader.ReadUint8()
	}

	info.Vac = reader.ReadUint8()
	info.Bots = reader.ReadUint8()
}

// Send an A2S_RULES query to the server. This returns a mapping of cvar names
// to values.
func (this *ServerQuerier) QueryRules() (map[string]string, error) {
	var rules map[string]string
	var err error

	// Note: must assign |err| in case there's a panic.
	err = Try(func() error {
		rules, err = this.queryRules()
		return err
	})

	return rules, err
}

func (this *ServerQuerier) queryRules() (map[string]string, error) {
	// Try to get a successful challenge.
	rechallenges := 0
	data, err := this.a2s_rules()
	for err == ErrConfusedChallengeReply && rechallenges < 3 {
		data, err = this.a2s_rules()
		rechallenges++
	}

	// Challenge failed - abort.
	if err != nil {
		return nil, err
	}

	switch int32(binary.LittleEndian.Uint32(data)) {
	case -1:
		return this.processRules(data, false)
	case -2:
		full, compressed, err := this.waitForMultiPacketReply(data)
		if err != nil {
			return nil, err
		}
		return this.processRules(full, compressed)
	default:
		return nil, ErrBadPacketHeader
	}
}

func (this *ServerQuerier) a2s_rules() ([]byte, error) {
	data := []byte{
		0xff, 0xff, 0xff, 0xff,
		A2S_RULES,
		0xff, 0xff, 0xff, 0xff,
	}
	if err := this.socket.Send(data); err != nil {
		return nil, err
	}

	data, err := this.socket.Recv()
	if err != nil {
		return nil, err
	}

	switch int32(binary.LittleEndian.Uint32(data[0:4])) {
	case -2:
		// AgeOfChivalry (appid 17510 had an instance of immediately reporting
		// a rules reply in response to a challenge. Maybe in a rare case the
		// server comes up with a -1 challenge?
		return data, nil
	case -1:
		// Ok, continue.
	default:
		panic(ErrBadPacketHeader)
	}

	switch data[4] {
	case S2A_RULES:
		// Some servers report an immediate, very truncated A2S_RULES reply.
		// It's not clear why - either a bug or some sort of information
		// hiding tactic, but we support this anyway.
		return data, nil
	case S2A_INFO_SOURCE, S2A_PLAYER:
		// Some servers reply with the wrong kind of query. For these, we retry.
		return nil, ErrConfusedChallengeReply
	case S2C_CHALLENGE:
		// Ok, continue.
	default:
		panic(ErrBadChallengeResponse)
	}

	// Send the rules query now that we've got a challenge sequence.
	reply := []byte{
		0xff, 0xff, 0xff, 0xff,
		A2S_RULES,
		data[5], data[6], data[7], data[8],
	}
	if err := this.socket.Send(reply); err != nil {
		return nil, err
	}
	return this.socket.Recv()
}

type MultiPacketHeader struct {
	// Size of the packet header itself.
	Size int

	// Packet sequence id.
	Id uint32

	// Packet number out of Total Packets.
	PacketNumber uint8

	// Total number of packets to receive.
	TotalPackets uint8

	// Packet size (0 if not present).
	PacketSize uint16

	// Compression information.
	Compressed bool

	Payload []byte
}

func (this *ServerQuerier) decodeMultiPacketHeader(data []byte) *MultiPacketHeader {
	reader := NewPacketReader(data)
	if reader.ReadInt32() != -2 {
		panic(ErrBadPacketHeader)
	}
	if this.info == nil {
		panic(ErrUnknownGameEngine)
	}

	header := &MultiPacketHeader{}
	header.Id = reader.ReadUint32()

	switch this.info.GameEngine() {
	case GOLDSRC:
		pkt := reader.ReadUint8()
		header.PacketNumber = (pkt >> 4) & 0xf
		header.TotalPackets = (pkt & 0xf)

	case SOURCE:
		header.Compressed = (header.Id & uint32(0x80000000)) != 0
		header.TotalPackets = reader.ReadUint8()
		header.PacketNumber = reader.ReadUint8()
		if !this.info.IsPreOrangeBox() {
			header.PacketSize = reader.ReadUint16()
		}

	default:
		panic(ErrUnknownGameEngine)
	}

	header.Size = reader.Pos()
	header.Payload = data[header.Size:]
	return header
}

func (this *ServerQuerier) waitForMultiPacketReply(data []byte) ([]byte, bool, error) {
	header := this.decodeMultiPacketHeader(data)
	packets := make([]*MultiPacketHeader, header.TotalPackets)
	received := 0
	fullSize := 0

	for {
		if int(header.PacketNumber) >= len(packets) {
			panic(ErrBadPacketNumber)
		}
		if packets[header.PacketNumber] != nil {
			panic(ErrDuplicatePacket)
		}

		packets[header.PacketNumber] = header
		fullSize += len(header.Payload)
		received++

		if received == len(packets) {
			break
		}

		data, err := this.socket.Recv()
		if err != nil {
			return nil, false, err
		}

		header = this.decodeMultiPacketHeader(data)
	}

	payload := make([]byte, fullSize)
	cursor := 0
	for _, header := range packets {
		copy(payload[cursor:cursor+len(header.Payload)], header.Payload)
		cursor += len(header.Payload)
	}

	return payload, packets[0].Compressed, nil
}

func (this *ServerQuerier) processRules(data []byte, compressed bool) (map[string]string, error) {
	reader := NewPacketReader(data)

	if compressed {
		decompressedSize := reader.ReadUint32()
		checksum := reader.ReadUint32()

		// Sanity check so we don't allocate and zero 3GB of memory by accident.
		if decompressedSize > uint32(1024*1024) {
			return nil, ErrWrongBz2Size
		}

		decompressed := make([]byte, decompressedSize)
		bz2Reader := bzip2.NewReader(bytes.NewReader(data[reader.Pos():]))
		n, err := bz2Reader.Read(decompressed)
		if err != nil {
			return nil, err
		}
		if n != int(decompressedSize) {
			return nil, ErrWrongBz2Size
		}
		if crc32.ChecksumIEEE(decompressed) != checksum {
			return nil, ErrWrongBz2Checksum
		}

		// Switch to the decompressed stream.
		data = decompressed
		reader = NewPacketReader(data)
	}

	if reader.ReadInt32() != -1 {
		panic(ErrBadPacketHeader)
	}
	if reader.ReadUint8() != S2A_RULES {
		panic(ErrBadRulesReply)
	}

	count := int(reader.ReadUint16())

	rules := map[string]string{}
	for i := 0; i < count; i++ {
		key, ok := reader.TryReadString()
		if !ok {
			break
		}
		val, ok := reader.TryReadString()
		if !ok {
			break
		}
		rules[key] = val
	}

	return rules, nil
}
