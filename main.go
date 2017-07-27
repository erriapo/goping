package main

import (
	"fmt"
	"github.com/erriapo/goping/core"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"os"
	"reflect"
	"runtime"
	"time"
)

// A quote by Arthur Schopenhauer
const payload = "A high degree of intellect tends to make a man unsocial."

var cache = core.NewCache()

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
	//	peerHost, errReversing := cache.Reverse(host)
	//	if errReversing != nil {
	//		peerHost = host
	//	}
	//	fmt.Printf("\tpeerHost = %v\n", peerHost)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

	c, err := icmp.ListenPacket("ip4:icmp", net.IPv4zero.String())
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	//debug
	fmt.Printf("typeof c = %v\n", reflect.TypeOf(c))
	wm := core.NewEcho(payload)
	//fmt.Printf("typeof wm = %v\n", reflect.TypeOf(wm))
	wb, err := wm.Marshal(nil)
	//fmt.Println(wb)
	if err != nil {
		log.Fatal(err)
	}
	rb := make([]byte, 1500)

	var t1 time.Time
	fmt.Printf("entering loop\n")
nn:
	for i := 1; i <= int(count); i++ {
		//t1, _ := time.Parse(time.RFC3339, "2017-06-28T19:55:50+00:00")
		fmt.Println(i)
		t1 = time.Now().Add(time.Second * 6)
		fmt.Printf("timeout at %v\n", t1)
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

		// TODO we need to loop until we receive an echo reply
		n, peer, err := c.ReadFrom(rb)
		fmt.Printf("Reading from peer %v\n", peer)

		if err != nil {
			fmt.Println("$$exit from ReadFrom")
			elapsed := time.Since(start)
			fmt.Printf("$$%v\n", elapsed)
			log.Fatal(err)
		}
		elapsed := time.Since(start)
		fmt.Printf("happy path from %v %v %v read\n", peer, elapsed, n)

		rm, err := icmp.ParseMessage(1, rb[:n])
		fmt.Println(rm.Type)
		if err != nil {
			log.Fatal(err)
		}
		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			log.Printf("got %+v; want echo reply", rm)
		case ipv4.ICMPTypeDestinationUnreachable:
			log.Printf("got %+v; Destination Net Prohibited", rm)
		default:
			log.Printf("DEFAULT got %+v;", rm)
		}
		time.Sleep(1 * time.Millisecond)
	}
	os.Exit(run())
}

func run() int {
	return 0
}
