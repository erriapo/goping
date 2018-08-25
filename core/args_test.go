// Copyright 2017 Gavin Chun Jin. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package core

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	oldLookupIP := lookupIPfunc
	oldLookupAddr := lookupAddrfunc
	defer func() {
		lookupIPfunc = oldLookupIP
		lookupAddrfunc = oldLookupAddr
	}()
	lookupIPfunc = func(host string) (ips []net.IP, err error) {
		switch host {
		case "localhost":
			return []net.IP{net.ParseIP("::1"),
				net.ParseIP("fe80::1"),
				net.ParseIP("127.0.0.1")}, nil
		case "127.0.0.1":
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		case "www.google.com":
			return []net.IP{net.ParseIP("216.58.193.68")}, nil
		default:
			return []net.IP{}, &net.DNSError{Err: "no such host",
				Name: "placeholder", Server: "127.0.0.1:53"}
		}

	}
	lookupAddrfunc = func(addr string) (names []string, err error) {
		switch addr {
		case "127.0.0.1":
			return []string{"localhost", "frodo"}, nil
		default:
			return []string{}, &net.DNSError{Err: "unrecognized address"}
		}
	}

	code := m.Run()
	os.Exit(code)
}

type MockAddr struct {
	net.Addr
}

func (a MockAddr) String() string {
	return "127.0.0.1"
}

func TestResolutionToLocalhost(t *testing.T) {
	cache := NewCache()

	_, err := cache.Reverse(nil)
	if err != ErrMissingPeer {
		t.Errorf("expected %v ; got %v\n", ErrMissingPeer, err)
	}
	host, err := cache.Reverse(MockAddr{})
	if err != nil {
		t.Errorf("expected %v ; got %v\n", nil, err)
	}
	if host != "localhost" {
		t.Errorf("expected %v ; got %v\n", "localhost", host)
	}
}

func TestParseZeroOptions(t *testing.T) {
	fixture1 := []string{}
	var fixture2 []string

	if _, _, _, _, _, _, err := ParseOption(fixture1); err == nil {
		t.Errorf("expected errUnknownHost ; got %v\n", err)
	}
	if _, _, _, _, _, _, err := ParseOption(fixture2); err == nil {
		t.Errorf("expected errUnknownHost ; got %v\n", err)
	}
	if _, _, _, _, _, _, err := ParseOption(nil); err == nil {
		t.Errorf("expected errUnknownHost ; got %v\n", err)
	}
}

func TestReturnOnlyIPv4(t *testing.T) {
	ipv4 := ParseAddr("localhost")
	fmt.Printf("%v\n", ipv4)
	if ipv4.IP.To4() == nil {
		t.Errorf("expected %v ; got %v\n",
			net.ParseIP("127.0.0.1"), ipv4)
	}
}

func TestReturnDNSError(t *testing.T) {
	ipv4 := ParseAddr("babihutan")
	fmt.Printf("%v\n", ipv4)
	if ipv4 != nil {
		t.Errorf("expected %v ; got %v\n", nil, ipv4)
	}
}

var counterFixtures = []struct {
	header      string
	changesFunc func(*Counter)
	expected    string
}{
	{"CAFEBABE",
		func(c *Counter) {},
		"CAFEBABE\n0 packets transmitted, 0 received, 0% packet loss\n"},
	{"CAFEBABE",
		func(c *Counter) {
			c.OnSent()
		},
		"CAFEBABE\n1 packets transmitted, 0 received, 100% packet loss\n"},
	{"CAFEBABE",
		func(c *Counter) {
			c.OnSent()
			c.OnSent()
			c.OnReception()
			c.NoteAnError()
		},
		"CAFEBABE\n2 packets transmitted, 1 received, +1 errors, 50% packet loss\n"},
}

func TestCounter(t *testing.T) {
	for _, tt := range counterFixtures {
		var b bytes.Buffer
		counter := NewCounter()
		tt.changesFunc(counter)
		counter.Render(&b, tt.header)
		if b.String() != tt.expected {
			t.Errorf("expected <%v> ; got <%v>\n", tt.expected, b.String())
		}
	}
}

var idnaFixtures = []struct {
	punycode string
	expected string
}{
	{"xn--pdc.ca", "ઊ.ca"},
	{"xn--m5c.ws", "๛.ws"},
	{"xn--bdk.ws", "ツ.ws"},
	{"www.google.ca", "www.google.ca"},
}

func TestPunycodeConversion(t *testing.T) {
	for _, tt := range idnaFixtures {
		reality := TryConvertPunycode(tt.punycode)
		if reality != tt.expected {
			t.Errorf("TryConvertPunycode(%v): expected %s, actual %s", tt.punycode, tt.expected, reality)
		}
	}
}

func BenchmarkParseLocalhostFqdn(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseAddr("localhost")
	}
}

func BenchmarkParseLocalhostIpv4(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseAddr("127.0.0.1")
	}
}

func BenchmarkParseLocalhostIpv6(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseAddr("::1")
	}
}

func BenchmarkReverseLookup(b *testing.B) {
	cache := NewCache()
	b.ResetTimer()
	cache.Reverse(MockAddr{})
}
