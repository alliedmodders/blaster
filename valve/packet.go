// vim: set ts=4 sw=4 tw=99 noet:
//
// Blaster (C) Copyright 2014 AlliedModders LLC
// Licensed under the GNU General Public License, version 3 or higher.
// See LICENSE.txt for more details.
package valve

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

const kMaxPacketSize = 1400

var ErrOutOfBounds = errors.New("read out of bounds")

type PacketBuilder struct {
	bytes.Buffer
}

func (this *PacketBuilder) WriteCString(str string) {
	this.WriteString(str)
	this.WriteByte(0)
}

func (this *PacketBuilder) WriteBytes(bytes []byte) {
	this.Write(bytes)
}

type PacketReader struct {
	buffer []byte
	pos    int
}

func NewPacketReader(packet []byte) *PacketReader {
	return &PacketReader{
		buffer: packet,
		pos:    0,
	}
}

// Debugging only.
func (this *PacketReader) dumpRemaining() {
	fmt.Printf("%+v\n", this.buffer[this.pos:])
}

func (this *PacketReader) canRead(size int) error {
	if size+this.pos > len(this.buffer) {
		return ErrOutOfBounds
	}
	return nil
}

func (this *PacketReader) Slice(count int) []byte {
	if this.canRead(count) != nil {
		return nil
	}
	bytes := this.buffer[this.pos : this.pos+count]
	this.pos += count
	return bytes
}

func (this *PacketReader) Pos() int {
	return this.pos
}

func (this *PacketReader) ReadIPv4() (net.IP, error) {
	if err := this.canRead(net.IPv4len); err != nil {
		return nil, err
	}

	ip := net.IP(this.buffer[this.pos : this.pos+net.IPv4len])
	this.pos += net.IPv4len
	return ip, nil
}

func (this *PacketReader) ReadPort() (uint16, error) {
	if err := this.canRead(2); err != nil {
		return 0, err
	}

	port := binary.BigEndian.Uint16(this.buffer[this.pos:])
	this.pos += 2
	return port, nil
}

func (this *PacketReader) ReadUint8() uint8 {
	b := this.buffer[this.pos]
	this.pos++
	return b
}

func (this *PacketReader) ReadUint16() uint16 {
	u16 := binary.LittleEndian.Uint16(this.buffer[this.pos:])
	this.pos += 2
	return u16
}

func (this *PacketReader) ReadUint32() uint32 {
	u32 := binary.LittleEndian.Uint32(this.buffer[this.pos:])
	this.pos += 4
	return u32
}

func (this *PacketReader) ReadInt32() int32 {
	return int32(this.ReadUint32())
}

func (this *PacketReader) ReadUint64() uint64 {
	u64 := binary.LittleEndian.Uint64(this.buffer[this.pos:])
	this.pos += 8
	return u64
}

func (this *PacketReader) TryReadString() (string, bool) {
	start := this.pos
	for this.pos < len(this.buffer) {
		if this.buffer[this.pos] == 0 {
			this.pos++
			return string(this.buffer[start : this.pos-1]), true
		}
		this.pos++
	}
	return "", false
}

func (this *PacketReader) ReadString() string {
	start := this.pos
	for {
		// Note: it's intended that we panic for strings that are not null
		// terminated.
		if this.buffer[this.pos] == 0 {
			this.pos++
			break
		}
		this.pos++
	}
	return string(this.buffer[start : this.pos-1])
}

func (this *PacketReader) More() bool {
	return this.pos < len(this.buffer)
}

type UdpSocket struct {
	timeout time.Duration
	cn      net.Conn
	buffer  [kMaxPacketSize]byte
	wait    time.Duration
	next    time.Time
}

func NewUdpSocket(address string, timeout time.Duration) (*UdpSocket, error) {
	cn, err := net.Dial("udp", address)
	if err != nil {
		return nil, err
	}

	return &UdpSocket{
		timeout: timeout,
		cn:      cn,
	}, nil
}

func (this *UdpSocket) SetTimeout(timeout time.Duration) {
	this.timeout = timeout
}

func (this *UdpSocket) RemoteAddr() net.Addr {
	return this.cn.RemoteAddr()
}

func (this *UdpSocket) SetRateLimit(ratePerMinute int) {
	this.wait = (time.Minute / time.Duration(ratePerMinute)) + time.Second
}

func (this *UdpSocket) extendedDeadline() time.Time {
	return time.Now().Add(this.timeout)
}

func (this *UdpSocket) enforceRateLimit() {
	if this.wait == 0 {
		return
	}

	wait := this.next.Sub(time.Now())
	if wait > 0 {
		time.Sleep(wait)
	}
}

func (this *UdpSocket) setNextQueryTime() {
	if this.wait != 0 {
		this.next = time.Now().Add(this.wait)
	}
}

func (this *UdpSocket) Send(bytes []byte) error {
	this.enforceRateLimit()
	defer this.setNextQueryTime()

	// Set timeout.
	if this.timeout > 0 {
		this.cn.SetWriteDeadline(this.extendedDeadline())
	}

	// UDP is all or nothing.
	_, err := this.cn.Write(bytes)
	return err
}

func (this *UdpSocket) Recv() ([]byte, error) {
	defer this.setNextQueryTime()

	// Set timeout.
	if this.timeout > 0 {
		this.cn.SetReadDeadline(this.extendedDeadline())
	}

	n, err := this.cn.Read(this.buffer[0:kMaxPacketSize])
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, n)
	copy(buffer, this.buffer[:n])
	return buffer, nil
}

func (this *UdpSocket) Close() {
	this.cn.Close()
}
