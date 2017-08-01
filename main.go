// Copyright 2017 Gavin Bong. All rights reserved.
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

// A quote by Arthur Schopenhauer
const payload = "A high degree of intellect tends to make a man unsocial."

// milliseconds to pause between sending each ICMP request
const pause = 1

var accountant = stats.NewSink()
var cache = core.NewCache()
var counter = core.NewCounter()
var pingHeading bool

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
	help, verbose, count, host, err := core.ParseOption(os.Args[1:])
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
	}

	// trap CTRL+C
	exitchan := make(chan os.Signal, 1)
	signal.Notify(exitchan, os.Interrupt) // SIGINT
	go func() {
		<-exitchan
		counter.String(heading(choose(peerHost, host)))
		if counter.Sent > 0 && counter.Recvd > 0 && counter.Recvd <= counter.Sent {
			fmt.Printf("%s\n", thirdparty.Format(accountant))
		}
		os.Exit(1)
	}()

	c, err := icmp.ListenPacket("ip4:icmp", net.IPv4zero.String())
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	wm := core.NewEcho(payload)
	wb, err := wm.Marshal(nil)
	if err != nil {
		log.Fatal(err)
	}
	rb := make([]byte, 1500)

	var t1 time.Time
	var peer2 net.Addr
nn:
	for i := 1; i <= int(count); i++ {
		//t1, _ := time.Parse(time.RFC3339, "2017-06-28T19:55:50+00:00")
		t1 = time.Now().Add(time.Second * 6)
		if err := c.SetDeadline(t1); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to set read & write Deadline.")
			log.Fatal("Unable to continue. Halted")
		}
		time.Sleep(pause * time.Second)
		start := time.Now()
		//if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP("8.8.8.8")}); err != nil {
		if _, err := c.WriteTo(wb, host); err != nil {
			fmt.Fprintf(os.Stderr, "%d connect: Network is unreachable\n", i)
			continue nn
		}
		counter.OnSent()

		// TODO we need to loop until we receive an echo reply
		n, peer, err := c.ReadFrom(rb)
		if peer != nil {
			peer2 = peer
		}

		if !pingHeading {
			fmt.Printf("PING %v (%v) 56(64) bytes of data.\n", host, choose(peerHost, peer))
			pingHeading = true
		}

		if err != nil {
			fmt.Printf("%v bytes from (%v): icmp_req=%v No response\n", 0, choose(peerHost, host), i)
			if verbose {
				fmt.Fprintf(os.Stderr, "\t%+v\n", err)
			}
			continue nn
		}
		elapsed := time.Since(start)

		fmt.Printf("%v bytes from (%v): icmp_req=%v time=%v\n", n, peer, i, elapsed)
		if verbose {
			fmt.Printf("RTT %d ns\n", elapsed.Nanoseconds())
		}

		rm, err := icmp.ParseMessage(1, rb[:n])
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
	counter.String(heading(choose(peerHost, peer2)))
	if counter.Sent > 0 && counter.Recvd > 0 && counter.Recvd <= counter.Sent {
		fmt.Printf("%s\n", thirdparty.Format(accountant))
	}
	os.Exit(0)
}
