package main

import (
	"fmt"
	"github.com/erriapo/icmp/core"
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

func main() {
	fmt.Println("OS=" + runtime.GOOS)
	fmt.Println("arch=" + runtime.GOARCH)

	fmt.Printf("ParseAddr(\"localhost\") returns %v\n", core.ParseAddr("localhost"))
	//fmt.Printf("ParseAddr(\"www.google.com\") returns %v\n", core.ParseAddr("www.google.com"))
	fmt.Printf("ParseAddr(\"127.0.0.1\") returns %v\n", core.ParseAddr("127.0.0.1"))

	//@TODO valid argument or address to ping
	/*
	   a, errdns := net.LookupIP("www.google.com")
	   if errdns != nil {
	       log.Fatal(errdns)
	   }
	   fmt.Printf("%v \n", a)
	*/

	//c, err := icmp.ListenPacket("udp4", "0.0.0.0")
	// This returns an error when I write to it:
	// 2017/06/26 19:11:30 write udp 0.0.0.0:28->8.8.8.8:0: sendto: invalid argument
	//
	c, err := icmp.ListenPacket("ip4:icmp", net.IPv4zero.String())
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	//debug
	fmt.Printf("typeof c = %v\n", reflect.TypeOf(c))

	wm := core.NewEcho(payload)
	fmt.Printf("typeof wm = %v\n", reflect.TypeOf(wm))
	wb, err := wm.Marshal(nil)
	fmt.Println(wb)
	if err != nil {
		log.Fatal(err)
	}
	rb := make([]byte, 1500)

	var t1 time.Time
nn:
	for i := 1; i <= 10; i++ {
		//t1, _ := time.Parse(time.RFC3339, "2017-06-28T19:55:50+00:00")
		fmt.Println(i)
		t1 = time.Now().Add(time.Second * 6)
		fmt.Printf("timeout at %v\n", t1)
		if err := c.SetDeadline(t1); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to set read & write Deadline.")
			log.Fatal("Unable to continue. Halted")
		}

		start := time.Now()
		if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP("8.8.8.8")}); err != nil {
			fmt.Println("..exit from WriteTo")
			elapsed := time.Since(start)
			fmt.Printf("..%v\n", elapsed)
			continue nn
			//log.Fatal(err)
		}
		n, peer, err := c.ReadFrom(rb)
		if err != nil {
			fmt.Println("$$exit from ReadFrom")
			elapsed := time.Since(start)
			fmt.Printf("$$%v\n", elapsed)
			log.Fatal(err)
		}
		elapsed := time.Since(start)
		fmt.Printf("happy path from %v %v\n", peer, elapsed)

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
			log.Println("unexpected reply")
		}
		time.Sleep(1 * time.Millisecond)
	}
	os.Exit(run())
}

func run() int {
	return 0
}
