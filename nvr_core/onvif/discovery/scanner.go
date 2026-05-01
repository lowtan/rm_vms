package discovery

import (
	"context"
	"fmt"
	"net"
	"golang.org/x/net/ipv4"
	"nvr_core/onvif"
	"nvr_core/logger"
)

var LOG = logger.NewLogger("[nvr_core]","[onvif]","discovery")

const (
	multicastAddress = "239.255.255.250:3702"
	maxDatagramSize  = 8192
)

// Scanner handles the UDP networking for discovering ONVIF devices.
type Scanner struct {
	multicastAddr *net.UDPAddr
}

// NewScanner initializes the network dependencies for discovery.
func NewScanner() (*Scanner, error) {
	addr, err := net.ResolveUDPAddr("udp4", multicastAddress)
	if err != nil {
		return nil, fmt.Errorf("resolving multicast address: %w", err)
	}
	return &Scanner{multicastAddr: addr}, nil
}

// Network Configuration Requirements
// Once the code is updated with a TTL >= 2, the network infrastructure takes over. To ensure the packet actually reaches the target subnet and the unicast replies route back to your management layer:
// 	1.	Multicast Routing: The router connecting the subnets must have multicast routing enabled (e.g., PIM Sparse/Dense mode) or an IGMP Proxy configured to forward traffic destined for 239.255.255.250.
// 	2.	Firewall Rules: Ensure that UDP port 3702 is allowed to traverse the subnets.
// 	3.	Unicast Return Path: The cameras will respond to the management server's IP address directly via unicast. Your routing tables must allow standard TCP/UDP traffic to flow back from the camera VLAN to the server VLAN.
func (s *Scanner) Scan(ctx context.Context) ([]onvif.DiscoveredCamera, error) {

	log := LOG.Lin("[Scan]")

	// Create standard UDP connection
	conn, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		return nil, fmt.Errorf("starting UDP listener: %w", err)
	}
	defer conn.Close()

	// Wrap it with the IPv4 packet connection layer
	packetConn := ipv4.NewPacketConn(conn)

	// Increase the Multicast TTL to cross subnets
	// A TTL of 2 allows crossing 1 router. Increase this if your network topology requires more hops.
	if err := packetConn.SetMulticastTTL(2); err != nil {
		return nil, fmt.Errorf("setting multicast TTL: %w", err)
	}

	log.Debug("Building Probe Message")
	probePayload := BuildProbeMessage()

	// Send the payload using the standard connection
	if _, err := conn.WriteTo([]byte(probePayload), s.multicastAddr); err != nil {
		return nil, fmt.Errorf("sending multicast probe: %w", err)
	}

	results := make(chan []byte)

	log.Debug("Firing up UDP datagram read")

	// Background worker to read UDP datagrams
	go func() {
		for {
			buffer := make([]byte, maxDatagramSize)
			
			if deadline, ok := ctx.Deadline(); ok {
				conn.SetReadDeadline(deadline)
			}
			
			n, _, err := conn.ReadFrom(buffer)
			if err != nil {
				close(results) 
				return
			}
			
			bufCopy := make([]byte, n)
			copy(bufCopy, buffer[:n])
			results <- bufCopy
		}
	}()

	var discovered []onvif.DiscoveredCamera
	seenMap := make(map[string]bool)

	log.Debug("Start scanning...")

	for {
		select {
		case <-ctx.Done():
			return discovered, nil
		case raw, ok := <-results:
			if !ok {
				return discovered, nil
			}

			log.Debug("[for] %v", raw)

			cameras, err := ParseProbeMatch(raw)
			if err != nil {
				continue 
			}

			log.Debug("[for] cameras number:", len(cameras))

			for _, cam := range cameras {
				if !seenMap[cam.XAddrs] {
					seenMap[cam.XAddrs] = true
					discovered = append(discovered, cam)
				}
			}
		}
	}
}