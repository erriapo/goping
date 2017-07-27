package core

import (
	"errors"
	"flag"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"io/ioutil"
	"log"
	"net"
	"os"
	"reflect"
)

var lookupIPfunc = net.LookupIP
var lookupAddrfunc = net.LookupAddr

type Cache struct {
	lookupErr error
	Peer      string
}

var NoPeerArg = errors.New("peer argument is nil")
var NoPeerResult = errors.New("peer not resolving")

// New returns a new Cache instance.
func NewCache() *Cache {
	return &Cache{}
}

// Reverse cache reverse IP resolution
// @TODO caching logic is missing.
func (c *Cache) Reverse(peer net.Addr) (string, error) {
	if peer == nil {
		return "", NoPeerArg
	}
	fmt.Printf("peer.String=%v\n", peer.String())
	names, err := lookupAddrfunc(peer.String())
	if err == nil {
		//		fmt.Printf("\tpeer = %v\n", names)
		if len(names) > 0 {
			c.Peer = names[0]
			return c.Peer, nil
		}
	} else {
		c.lookupErr = err
	}
	return "", NoPeerResult
}

// ParseAddr returns the IPv4 address.
// Potentially could be a nil return value.
func ParseAddr(input string) *net.IPAddr {
	ip := net.ParseIP(input)
	if ip != nil {
		return &net.IPAddr{IP: ip}
	}

	// if input is 'localhost', the candidates can contain
	// both ipv4 & ipv6 address.
	// i.e. [::1 127.0.0.1 fe80::1]
	candidates, err := lookupIPfunc(input)
	if err != nil {
		log.Printf("%v\n", err)
		return nil
	}

	fmt.Printf("%v\n", candidates)
	fmt.Printf("type 0 = %v\n", reflect.TypeOf(candidates[0]))

	var result net.IP
	for _, candidate := range candidates {
		// First ipv4 address wins out.
		if candidate.To4() != nil {
			result = candidate
			break
		}
	}
	if result == nil {
		return nil
	}
	return &net.IPAddr{IP: result}
}

func NewEcho(payload string) icmp.Message {
	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: []byte(payload),
		},
	}
	return wm
}

var hostUnknown = errors.New("Unknown host")

type Arg struct {
	Help  bool
	Host  string
	Count uint64
}

func ParseOption(options []string) (error, bool, uint64, *net.IPAddr) {
	fmt.Printf("%v options\n", options)
	bucket := new(Arg)

	f := flag.NewFlagSet("goping", flag.ContinueOnError)
	f.SetOutput(ioutil.Discard)
	f.BoolVar(&bucket.Help, "h", false, "")
	f.Uint64Var(&bucket.Count, "c", 5, "")
	f.StringVar(&bucket.Host, "d", bucket.Host, "")
	fmt.Printf("%v bucket\n", bucket)
	if err := f.Parse(options); err != nil {
		return err, false, 0, nil
	}

	if bucket.Help {
		return nil, bucket.Help, 0, nil
	}

	fmt.Printf("parseaddr in\n")
	ipAddr := ParseAddr(bucket.Host)
	fmt.Printf("parseaddr out\n")
	if ipAddr == nil {
		return hostUnknown, false, 0, nil
	}
	return nil, bucket.Help, bucket.Count, ipAddr
}
