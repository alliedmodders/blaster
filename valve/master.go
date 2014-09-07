// vim: set ts=4 sw=4 tw=99 noet:
//
// Blaster (C) Copyright 2014 AlliedModders LLC
// Licensed under the GNU General Public License, version 3 or higher.
// See LICENSE.txt for more details.
package valve

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

const kMaxFilterLength = 190
const kDefaultMasterTimeout = time.Minute * 5

var ErrBadResponseHeader = fmt.Errorf("bad response header")
var kMasterResponseHeader = []byte{0xff, 0xff, 0xff, 0xff, 0x66, 0x0a}
var kNullIP = net.IP([]byte{0, 0, 0, 0})

// The callback the master query tool uses to notify of a batch of servers that
// has just been received.
type MasterQueryCallback func(batch ServerList) error

// Class for querying the master server.
type MasterServerQuerier struct {
	cn          *UdpSocket
	hostAndPort string
	filters     []string
}

// Create a new master server querier on the given host and port.
func NewMasterServerQuerier(hostAndPort string) (*MasterServerQuerier, error) {
	cn, err := NewUdpSocket(hostAndPort, kDefaultMasterTimeout)
	if err != nil {
		return nil, err
	}

	// 20 queries per minute, according to Valve. 15 seems to work best for us.
	cn.SetRateLimit(15)

	return &MasterServerQuerier{
		cn:          cn,
		hostAndPort: hostAndPort,
	}, nil
}

// Adds by AppIds to the filter list.
func (this *MasterServerQuerier) FilterAppIds(appIds []AppId) {
	for _, appId := range appIds {
		this.filters = append(this.filters, fmt.Sprintf("\\appid\\%d", appId))
	}
}

func (this *MasterServerQuerier) ClearFilters() {
	this.filters = []string{}
}

func computeNextFilterList(filters []string) ([]string, []string) {
	return filters[0:1], filters[1:]
}

// Query the master. Since the master server has timeout problems with lots of
// subsequent requests, we sleep for two seconds in between each batch request.
// This means the querying process is quite slow.
func (this *MasterServerQuerier) Query(callback MasterQueryCallback) error {
	filters, remaining := computeNextFilterList(this.filters)
	for {
		if err := this.tryQuery(callback, filters); err != nil {
			return err
		}

		if len(remaining) == 0 {
			break
		}
		filters, remaining = computeNextFilterList(remaining)
	}
	return nil
}

// Build a packet to query the master server, given an initial starting server
// ("0.0.0.0:0" for the initial batch) and an optional list of filter strings.
func BuildMasterQuery(hostAndPort string, filters []string) []byte {
	packet := PacketBuilder{}
	packet.WriteByte(0x31) // Magic number
	packet.WriteByte(0xFF) // All regions.
	packet.WriteCString(hostAndPort)

	if len(filters) == 0 {
		packet.WriteByte(0)
		packet.WriteByte(0)
	} else if len(filters) == 1 {
		packet.WriteCString(filters[0])
	} else {
		header := fmt.Sprintf("\\or\\%d", len(filters))
		packet.WriteBytes([]byte(header))
		for _, filter := range filters {
			packet.WriteBytes([]byte(filter))
		}
		packet.WriteByte(0)
	}
	return packet.Bytes()
}

func (this *MasterServerQuerier) tryQuery(callback MasterQueryCallback, filters []string) error {
	query := BuildMasterQuery("0.0.0.0:0", filters)
	if err := this.cn.Send(query); err != nil {
		return err
	}

	packet, err := this.cn.Recv()
	if err != nil {
		return err
	}

	// Sanity check the header.
	if len(packet) < 6 || bytes.Compare(packet[0:6], kMasterResponseHeader) != 0 {
		return ErrBadResponseHeader
	}

	// Chop off the response header.
	packet = packet[6:]

	seen := map[string]bool{}

	done := false
	ip := kNullIP
	port := uint16(0)
	for {
		reader := NewPacketReader(packet)
		serverCount := len(packet) / 6

		if serverCount == 0 {
			break
		}

		servers := ServerList{}
		for i := 0; i < serverCount; i++ {
			ip, err = reader.ReadIPv4()
			if err != nil {
				return err
			}
			port, err = reader.ReadPort()
			if err != nil {
				return err
			}

			// The list is terminated with 0s.
			if ip.Equal(kNullIP) && port == 0 {
				done = true
				break
			}

			addr := &net.TCPAddr{
				IP:   ip,
				Port: int(port),
			}
			if _, found := seen[addr.String()]; found {
				continue
			}

			servers = append(servers, addr)
			seen[addr.String()] = true
		}

		if err := callback(servers); err != nil {
			return err
		}

		if done {
			break
		}

		// Attempt to get the next batch 4 more times.
		for i := 1; ; i++ {
			address := fmt.Sprintf("%s:%d", ip.String(), port)
			query := BuildMasterQuery(address, filters)
			if err = this.cn.Send(query); err != nil {
				return err
			}

			if packet, err = this.cn.Recv(); err == nil {
				// Ok, keep going.
				break
			}

			// Maximum number of retries before we give up.
			if i == 4 {
				return err
			}
		}
	}

	return nil
}

func (this *MasterServerQuerier) Close() {
	this.cn.Close()
}
