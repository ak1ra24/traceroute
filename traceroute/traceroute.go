package traceroute

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"golang.org/x/net/ipv6"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func isIPv4(ip net.IP) bool {
	return len(ip.To4()) == net.IPv4len
}

func isIPv6(ip net.IP) bool {
	return len(ip) == net.IPv6len
}

func Traceroute(host string) {
	// const host = "1.1.1.1"
	ipaddrs, _ := net.LookupIP(host)
	addr := ipaddrs[rand.Intn(len(ipaddrs))]

	var version int

	if isIPv4(addr) {
		version = 4
		// p.IPVersion = 4
	} else if isIPv6(addr) {
		version = 6
		// p.IPVersion = 6
	}

	var conn *icmp.PacketConn
	var err error
	if version == 4 {
		conn, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	} else if version == 6 {
		conn, err = icmp.ListenPacket("ip6:ipv6-icmp", "::")
	}

	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	var typ icmp.Type
	if version == 4 {
		conn.IPv4PacketConn().SetControlMessage(ipv4.FlagTTL|ipv4.FlagSrc|ipv4.FlagDst|ipv4.FlagInterface, true)
		typ = ipv4.ICMPTypeEcho
	} else if version == 6 {
		conn.IPv6PacketConn().SetControlMessage(ipv6.FlagHopLimit|ipv6.FlagSrc|ipv6.FlagDst|ipv6.FlagInterface, true)
		typ = ipv6.ICMPTypeEchoRequest
	}

	wm := icmp.Message{
		Type: typ, Code: 0,
		Body: &icmp.Echo{
			ID:   0,
			Seq:  1,
			Data: []byte("HELLO-R-U-THERE"),
		},
	}

	rb := make([]byte, 1500)

	var wcm ipv6.ControlMessage

	for i := 1; i <= 64; i++ {
		wm.Body.(*icmp.Echo).Seq = i
		wb, err := wm.Marshal(nil)
		if err != nil {
			log.Fatal(err)
		}
		if version == 4 {
			if err := conn.IPv4PacketConn().SetTTL(i); err != nil {
				log.Fatal(err)
			}
		} else if version == 6 {
			wcm.HopLimit = i
		}

		dst := &net.IPAddr{IP: addr}
		begin := time.Now()
		if _, err := conn.WriteTo(wb, dst); err != nil {
			log.Fatal(err)
		}

		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		var n int
		var cm *ipv4.ControlMessage
		var cm6 *ipv6.ControlMessage
		var peer net.Addr
		if version == 4 {
			n, cm, peer, err = conn.IPv4PacketConn().ReadFrom(rb)
		} else if version == 6 {
			n, cm6, peer, err = conn.IPv6PacketConn().ReadFrom(rb)
		}
		if err != nil {
			err, ok := err.(net.Error)
			if ok && err.Timeout() {
				fmt.Printf("%v\t*\n", i)
				continue
			}
			log.Fatal(err)
		}
		rm, err := icmp.ParseMessage(1, rb[:n])
		if err != nil {
			log.Fatal(err)
		}
		rtt := time.Since(begin)

		switch rm.Type {
		case ipv4.ICMPTypeTimeExceeded:
			names, _ := net.LookupAddr(peer.String())
			fmt.Printf("%d\t%v %+v %v\n\t%+v\n", i, peer, names, rtt, cm)
		case ipv4.ICMPTypeEchoReply:
			names, _ := net.LookupAddr(peer.String())
			fmt.Printf("%d\t%v %+v %v\n\t%+v\n", i, peer, names, rtt, cm)
			return
		case ipv6.ICMPTypeEchoReply:
			names, _ := net.LookupAddr(peer.String())
			fmt.Printf("%d\t%v %+v %v\n\t%+v\n", i, peer, names, rtt, cm6)
			return
		case ipv6.ICMPTypeTimeExceeded:
			names, _ := net.LookupAddr(peer.String())
			fmt.Printf("%d\t%v %+v %v\n\t%+v\n", i, peer, names, rtt, cm6)
		default:
			log.Printf("unknown ICMP message: %+v\n", rm)
		}
	}
}
