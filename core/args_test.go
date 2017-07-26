package core

import (
	"fmt"
	"net"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	oldLookupIP := lookupIPfunc
	defer func() { lookupIPfunc = oldLookupIP }()
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
	code := m.Run()
	os.Exit(code)
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
