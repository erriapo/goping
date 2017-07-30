package main

import (
	"fmt"
	"github.com/erriapo/goping/core"
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

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintf(os.Stderr, "%s", core.Usage)
		os.Exit(2)
	}

	err, help, verbose, count, host := core.ParseOption(os.Args[1:])

	if help {
		fmt.Fprintf(os.Stderr, "%s", core.Usage)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Aborted: %v\n", err)
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
			// TODO print no reponse for packet #n
			fmt.Printf("%v bytes from (%v): icmp_req=%v No response\n", 0, peer2, i)
			elapsed := time.Since(start)
			fmt.Printf("$$%v\n", elapsed)
			continue nn
		}
		elapsed := time.Since(start)

		fmt.Printf("%v bytes from (%v): icmp_req=%v time=%v\n", n, peer, i, elapsed)

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
			if verbose {
				log.Printf("\t%+v; echo reply", rm)
			}
		case ipv4.ICMPTypeDestinationUnreachable:
			if verbose {
				log.Printf("\t%+v; Destination Net Prohibited", rm)
			}
		default:
			if verbose {
				log.Printf("\tunexpected %+v;", rm)
			}
		}
	}
	counter.String(heading(choose(peerHost, peer2)))
	os.Exit(run())
}

func run() int {
	return 0
}
