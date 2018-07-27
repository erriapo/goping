// Copyright 2017 Gavin Chun Jin. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"github.com/erriapo/goping/core"
	"github.com/erriapo/goping/thirdparty"
	"github.com/erriapo/stats"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

// A quote by Epictetus
const payload = "First learn the meaning of what you say, and then speak."
const payloadLen = len(payload)
const ipheader = 20
const icmpheader = 8
const payloadAndHeader = payloadLen + ipheader + icmpheader

// milliseconds to pause between sending each ICMP request
const pause = 1

var accountant = stats.NewSink()
var cache = core.NewCache()
var counter = core.NewCounter()
var pingHeading bool

// return the first non empty arg or "unknown"
func choose(option1 string, option2 net.Addr) string {
	if len(option1) != 0 {
		return option1
	}
	if option2 != nil {
		return option2.String()
	}
	return "unknown"
}

func heading(node string) string {
	return fmt.Sprintf("\n--- %[1]s ping statistics ---", node)
}

func nanoToMilli(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / float64(1000000)
}

func main() {
	help, verbose, count, host, cname, err := core.ParseOption(os.Args[1:])
	if help {
		fmt.Fprintf(os.Stderr, "%s", core.Usage)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Aborted: %v\n", err)
		fmt.Fprintf(os.Stderr, "%s", core.Usage)
		os.Exit(2)
	}

	// It is safe to ignore the error as we will fallback
	// to the supplied Host
	peerHost, _ := cache.Reverse(host)
	if verbose {
		fmt.Printf("Reversed lookup of supplied host = %v\n", peerHost)
		fmt.Printf("CNAME = %v\n", cname)
	}

	// trap CTRL+C
	exitchan := make(chan os.Signal, 1)
	signal.Notify(exitchan, os.Interrupt) // SIGINT
	go func() {
		<-exitchan
		counter.String(heading(choose(cname, host)))
		if counter.NeedStatistics() {
			fmt.Printf("%s\n", thirdparty.Format(accountant))
		}
		os.Exit(1)
	}()

	c, err := icmp.ListenPacket("ip4:icmp", net.IPv4zero.String())
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	var wm icmp.Message
	var wb []byte

	rb := make([]byte, 1500)

	var t1 time.Time
	var peer2 net.Addr
nn:
	for i := 1; i <= int(count); i++ {
		wm = core.NewEcho(payload, i)
		wb, err = wm.Marshal(nil)
		if err != nil {
			log.Fatal(err)
		}
		//t1, _ := time.Parse(time.RFC3339, "2017-06-28T19:55:50+00:00")
		t1 = time.Now().Add(time.Second * 6)
		if err := c.SetDeadline(t1); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to set read & write Deadline.")
			log.Fatal("Unable to continue. Halted")
		}
		time.Sleep(pause * time.Second)
		start := time.Now()
		if _, err := c.WriteTo(wb, host); err != nil {
			fmt.Fprintf(os.Stderr, "%d connect: Network is unreachable\n", i)
			continue nn
		}
		counter.OnSent()

		// TODO we need to loop until we receive an echo reply
		//hh := c.IPv4PacketConn()
		//k, cm, _, errgg := hh.ReadFrom(rb2)
		//fmt.Printf("!!! %v --- %v %v\n", cm, errgg, k)

		n, peer, err := c.ReadFrom(rb)
		if verbose {
			fmt.Printf("peer %v vs host %v\n", peer, host)
		}
		if peer != nil {
			peer2 = peer
		}

		if !pingHeading {
			fmt.Printf("PING %v (%v) %v(%v) bytes of data.\n", choose(cname, peer), host, payloadLen, payloadAndHeader)
			pingHeading = true
		}

		if err != nil {
			fmt.Printf("%v bytes from %v (%v): icmp_seq=%v No response\n", 0, choose(peerHost, host), host, i)
			if verbose {
				fmt.Fprintf(os.Stderr, "\t%+v\n", err)
			}
			continue nn
		}
		elapsed := time.Since(start)

		fmt.Printf("%v bytes from %v (%v): icmp_seq=%v time=%v\n", n, choose(peerHost, host), host.IP, i, elapsed)
		if verbose {
			fmt.Printf("RTT %d ns\n", elapsed.Nanoseconds())
		}

		rm, err := icmp.ParseMessage(1, rb[:n])
		//hder, _ := icmp.ParseIPv4Header(rb[:n])
		//fmt.Printf("REPLY %v -> %v\n", rm, hder)
		if err != nil {
			log.Fatal(err)
		}
		switch rm.Type {
		case ipv4.ICMPTypeEcho:
			if verbose {
				log.Printf("\t%+v; echo", rm)
			}
		case ipv4.ICMPTypeEchoReply:
			counter.OnReception()
			accountant.Push(nanoToMilli(elapsed))
			if verbose {
				log.Printf("\t%+v; echo reply", rm)
			}
		case ipv4.ICMPTypeDestinationUnreachable:
			fmt.Fprintf(os.Stderr, "\tDestination Net Prohibited\n")
			if verbose {
				log.Printf("%+v;", rm)
			}
		default:
			if verbose {
				log.Printf("\tunexpected %+v;", rm)
			}
		}
	}
	counter.String(heading(choose(cname, peer2)))
	if counter.NeedStatistics() {
		fmt.Printf("%s\n", thirdparty.Format(accountant))
	}
	os.Exit(0)
}
