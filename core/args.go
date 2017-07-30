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
	"sync/atomic"
	"text/template"
	"time"
)

var lookupIPfunc = net.LookupIP
var lookupAddrfunc = net.LookupAddr

// Cache saves the last reverse ip lookup.
type Cache struct {
	m map[string]string
}

// NoPeerArg means the client provided nil as a peer argument.
var NoPeerArg = errors.New("peer argument is nil")

// NoPeerResult means no reverse IP for the peer was found.
var NoPeerResult = errors.New("peer not resolving")

// New returns a new Cache instance.
func NewCache() *Cache {
	return &Cache{
		m: make(map[string]string),
	}
}

// Reverse caches reverse IP resolution
func (c *Cache) Reverse(peer net.Addr) (string, error) {
	if peer == nil {
		return "", NoPeerArg
	}

	v, ok := c.m[peer.String()]
	if ok {
		return v, nil
	}

	names, err := lookupAddrfunc(peer.String())
	if err == nil {
		//fmt.Printf("\tpeer = %v\n", names)
		if len(names) > 0 {
			c.m[peer.String()] = names[0]
			return c.m[peer.String()], nil
		}
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
	// e.g. [::1 127.0.0.1 fe80::1]
	candidates, err := lookupIPfunc(input)
	if err != nil {
		log.Printf("%v\n", err)
		return nil
	}

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

// NewEcho constructs an ICMP packet.
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

// Arg holds the command line arguments.
type Arg struct {
	Help  bool
	Host  string
	Extra bool
	Count uint64
}

func ParseOption(options []string) (error, bool, bool, uint64, *net.IPAddr) {
	bucket := new(Arg)

	f := flag.NewFlagSet("goping", flag.ContinueOnError)
	f.SetOutput(ioutil.Discard)
	f.BoolVar(&bucket.Help, "h", false, "")
	f.BoolVar(&bucket.Extra, "v", false, "")
	f.Uint64Var(&bucket.Count, "c", 5, "")
	f.StringVar(&bucket.Host, "d", bucket.Host, "")

	if err := f.Parse(options); err != nil {
		return err, false, false, 0, nil
	}

	if bucket.Help {
		return nil, bucket.Help, bucket.Extra, 0, nil
	}

	start := time.Now()
	fmt.Fprintf(os.Stderr, ".")
	ipAddr := ParseAddr(bucket.Host)
	elapsed := time.Since(start)
	fmt.Fprintf(os.Stderr, "Reverse lookup took %v\n\n", elapsed)

	if ipAddr == nil {
		return hostUnknown, false, bucket.Extra, 0, nil
	}
	return nil, bucket.Help, bucket.Extra, bucket.Count, ipAddr
}

// Counter keeps track of messages sent & received
type Counter struct {
	Sent  uint64
	Recvd uint64
	Loss  uint32
	tmpl  *template.Template
}

func NewCounter() *Counter {
	t, err := template.New("stat1").Parse("{{.Sent}} packets transmitted, {{.Recvd}} received, {{.Loss}}% packet loss\n")
	if err != nil {
		panic(err)
	}
	return &Counter{tmpl: t}
}

func (c *Counter) OnSent() {
	atomic.AddUint64(&c.Sent, 1)
}

func (c *Counter) OnReception() {
	atomic.AddUint64(&c.Recvd, 1)
}

func (c *Counter) String(header string) {
	r := atomic.LoadUint64(&c.Recvd)
	s := atomic.LoadUint64(&c.Sent)
	if r == 0 {
		if s != 0 {
			c.Loss = 100
		}
	} else {
		if s != 0 {
			c.Loss = uint32(((float32(s) - float32(r)) / float32(s)) * float32(100))
		}
	}
	fmt.Println(header)
	_ = c.tmpl.Execute(os.Stdout, c)
}
