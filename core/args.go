// Copyright 2017 Gavin Chun Jin. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Package core provides command line parsing & DNS lookup conveniences.
package core

import (
	"errors"
	"flag"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/idna"
	"golang.org/x/net/ipv4"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"text/template"
	//"time"
)

var lookupIPfunc = net.LookupIP
var lookupAddrfunc = net.LookupAddr

// Cache saves the last reverse ip lookup.
type Cache struct {
	m map[string]string
}

// ErrMissingPeer means the client provided nil as a peer argument.
var ErrMissingPeer = errors.New("peer argument is nil")

// ErrPeerNotResolving means no reverse IP for the peer was found.
var ErrPeerNotResolving = errors.New("peer not resolving")

// NewCache returns a new Cache instance.
func NewCache() *Cache {
	return &Cache{
		m: make(map[string]string),
	}
}

// Reverse caches reverse IP resolution
func (c *Cache) Reverse(peer net.Addr) (string, error) {
	if peer == nil {
		return "", ErrMissingPeer
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

	return "", ErrPeerNotResolving
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
func NewEcho(payload string, seq int) icmp.Message {
	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: seq,
			Data: []byte(payload),
		},
	}
	return wm
}

// ErrUnknownHost means we cannot parse
// the IP address or the FQDN that was provided.
var ErrUnknownHost = errors.New("Name or service not known")

// ErrNoTarget means the target was not supplied
var ErrNoTarget = errors.New("Missing target")

// ErrBadCount signifies that count packets must be greater than or equal to 1.
var ErrBadCount = errors.New("bad number of packets to transmit")

// Arg holds the command line arguments.
type Arg struct {
	Host      string
	Interface string
	Help      bool
	Extra     bool
	Count     uint64
}

const defaultInterface = "0.0.0.0"

// ParseOption parses command line arguments
func ParseOption(options []string) (bool, bool, uint64, *net.IPAddr, string, string, error) {
	if len(options) == 0 {
		return false, false, 0, nil, "", defaultInterface, ErrUnknownHost
	}
	bucket := new(Arg)

	f := flag.NewFlagSet("goping", flag.ContinueOnError)
	f.SetOutput(ioutil.Discard)
	f.BoolVar(&bucket.Help, "h", false, "")
	f.BoolVar(&bucket.Extra, "v", false, "")
	f.StringVar(&bucket.Interface, "I", "0.0.0.0", "")
	f.Uint64Var(&bucket.Count, "c", 5, "")

	if err := f.Parse(options); err != nil {
		return false, false, 0, nil, "", defaultInterface, err
	}

	if bucket.Extra {
		fmt.Printf("Interface selected: %v\n", bucket.Interface)
	}

	if bucket.Help {
		return bucket.Help, bucket.Extra, 0, nil, "", defaultInterface, nil
	}

	if len(f.Args()) == 0 {
		return false, false, 0, nil, "", defaultInterface, ErrNoTarget
	} else {
		bucket.Host = f.Args()[0]
	}

	if bucket.Count == 0 {
		return false, false, 0, nil, "", defaultInterface, ErrBadCount
	}

	//start := time.Now()
	fmt.Fprintf(os.Stderr, ".\n")
	ipAddr := ParseAddr(bucket.Host)
	//elapsed := time.Since(start)
	//fmt.Fprintf(os.Stderr, "%v\n\n", elapsed)

	if ipAddr == nil {
		return false, bucket.Extra, 0, nil, "", bucket.Interface, ErrUnknownHost
	}
	return bucket.Help, bucket.Extra, bucket.Count, ipAddr, TryConvertPunycode(GetCNAME(bucket.Host)), bucket.Interface, nil
}

const step uint64 = 1

// Counter keeps track of messages sent & received
type Counter struct {
	Sent   uint64
	Recvd  uint64
	Loss   uint32
	Err    bool
	Errors uint64
	lock   sync.Mutex
	tmpl   *template.Template
}

// NewCounter constructs a new Counter
func NewCounter() *Counter {
	t, err := template.New("stat1").Parse("{{.Sent}} packets transmitted, {{.Recvd}} received,{{if .Err}} +{{.Errors}} errors,{{end}} {{.Loss}}% packet loss\n")
	if err != nil {
		panic(err)
	}
	return &Counter{tmpl: t}
}

// OnSent remembers how many ICMP Echo was sent.
func (c *Counter) OnSent() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.Sent += step
}

// OnReception remembers how many ICMP Echo Reply was received.
func (c *Counter) OnReception() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.Recvd += step
}

func (c *Counter) NoteAnError() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.Errors += 1
}

func (c *Counter) gotError() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.Errors > 0 {
		c.Err = true
	}
}

// CalculateLoss side effect is to calculate the Loss percentage.
func (c *Counter) calculateLoss() {
	c.lock.Lock()
	defer c.lock.Unlock()
	//r := atomic.LoadUint64(&c.Recvd)
	//s := atomic.LoadUint64(&c.Sent)
	if c.Recvd == 0 {
		if c.Sent != 0 {
			c.Loss = 100
		}
	} else {
		if c.Sent != 0 {
			c.Loss = uint32(((float32(c.Sent) - float32(c.Recvd)) / float32(c.Sent)) * float32(100))
		}
	}
}

// Render calculates detailed packet loss percentages
// and writes them to the supplied Writer.
func (c *Counter) Render(w io.Writer, header string) {
	c.calculateLoss()
	c.gotError()
	fmt.Fprintln(w, header)
	_ = c.tmpl.Execute(w, c)
}

// NeedStatistics informs the caller if more statistics can be printed.
func (c *Counter) NeedStatistics() bool {
	return c.Sent > 0 && c.Recvd > 0 && c.Recvd <= c.Sent
}

// GetCNAME returns an empty string or the actual CNAME
func GetCNAME(h string) string {
	//We want the final CNAME
	cname, e := net.LookupCNAME(h)
	if e != nil {
		return ""
	}
	return cname
}

// TryConvertPunycode attempts to convert the domain if it is punycode encoded
func TryConvertPunycode(domain string) string {
	if len(domain) == 0 {
		return domain
	}
	p := idna.New()
	if strings.HasPrefix(domain, "xn--") {
		if result, err := p.ToUnicode(domain); err != nil {
			return domain
		} else {
			return result
		}

	} else {
		return domain
	}
}

// Peer encapsulates a displayable DNS name.
type Peer struct {
	FQDN string
	IP   string
}

// ChoosePeer decides between the supplied target host or the responding peer.
func ChoosePeer(hostFQDN string, host *net.IPAddr, hostErr error, peerFQDN string, peer net.Addr, peerErr error) Peer {
	if peer != nil {
		if peerErr != nil {
			return Peer{FQDN: peer.String(), IP: peer.String()}
		} else {
			return Peer{FQDN: peerFQDN, IP: peer.String()}
		}
	}

	if host != nil {
		if hostErr != nil {
			return Peer{FQDN: host.String(), IP: host.String()}
		} else {
			return Peer{FQDN: hostFQDN, IP: host.IP.String()}
		}
	}

	return Peer{FQDN: "Unknown", IP: "?.?.?.?"}
}
