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
	"reflect"
	"runtime"
	"time"
)

// A quote by Arthur Schopenhauer
const payload = "A high degree of intellect tends to make a man unsocial."

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
	fmt.Println("OS=" + runtime.GOOS)
	fmt.Println("arch=" + runtime.GOARCH)

	fmt.Printf("%v\n", os.Args[1:])
	fmt.Printf("%v\n", reflect.TypeOf(os.Args))

	if len(os.Args) == 1 {
		fmt.Printf("%s", core.Usage)
		os.Exit(2)
	}

	err, help, count, host := core.ParseOption(os.Args[1:])
	fmt.Printf("%v // %v // %v // %v\n", err, help, count, host)

	if help {
		fmt.Printf("%s", core.Usage)
		os.Exit(2)
	}

	// FIXME host
	peerHost, _ := cache.Reverse(host)
	//if errReversing != nil {
	//		peerHost = host
	//}
	fmt.Printf("\t!!peerHost = %v\n", peerHost)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

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
		fmt.Println(".")
		t1 = time.Now().Add(time.Second * 6)
		if err := c.SetDeadline(t1); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to set read & write Deadline.")
			log.Fatal("Unable to continue. Halted")
		}
		time.Sleep(time.Second * 1)
		start := time.Now()
		//if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP("8.8.8.8")}); err != nil {
		if _, err := c.WriteTo(wb, host); err != nil {
			fmt.Println("..exit from WriteTo")
			elapsed := time.Since(start)
			fmt.Printf("..%v\n", elapsed)
			// TODO "network is unreachable"
			fmt.Fprintf(os.Stderr, "connect: Network is unreachable\n")
			continue nn
			//log.Fatal(err)
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
			log.Printf("got %+v; echo", rm)
		case ipv4.ICMPTypeEchoReply:
			counter.OnReception()
			log.Printf("got %+v; want echo reply", rm)
		case ipv4.ICMPTypeDestinationUnreachable:
			log.Printf("got %+v; Destination Net Prohibited", rm)
		default:
			log.Printf("DEFAULT got %+v;", rm)
		}
		time.Sleep(1 * time.Millisecond) // pause between sending each ping
	}
	counter.String(heading(choose(peerHost, peer2)))
	os.Exit(run())
}

func run() int {
	return 0
}
