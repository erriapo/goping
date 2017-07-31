## Dependencies

`goping` only works on unixes & is only tested with golang 1.8.

* [golang.org/x/net/icmp](https://godoc.org/golang.org/x/net/icmp)
* [golang.org/x/net/ipv4](https://godoc.org/golang.org/x/net/ipv4)
* [github.com/erriapo/stats](https://github.com/erriapo/stats)

## Usage example

You must supply the target host with a -d option.

```bash
$ bin/goping -d localhost
.Reverse lookup took 139.777418ms

PING 127.0.0.1 (localhost) 56(64) bytes of data.
64 bytes from (127.0.0.1): icmp_req=1 time=594.843µs
64 bytes from (127.0.0.1): icmp_req=2 time=278.673µs
64 bytes from (127.0.0.1): icmp_req=3 time=280.854µs
64 bytes from (127.0.0.1): icmp_req=4 time=317.852µs
64 bytes from (127.0.0.1): icmp_req=5 time=313.576µs
64 bytes from (127.0.0.1): icmp_req=6 time=379.35µs
64 bytes from (127.0.0.1): icmp_req=7 time=388.265µs
64 bytes from (127.0.0.1): icmp_req=8 time=317.853µs
64 bytes from (127.0.0.1): icmp_req=9 time=387.401µs
64 bytes from (127.0.0.1): icmp_req=10 time=532.417µs

--- localhost ping statistics ---
10 packets transmitted, 5 received, 50% packet loss
rtt min/avg/max/mdev = 0.279/0.365/0.532/0.1 ms
```

Additionally, the `goping` binary needs the CAP_NET_RAWIO capability. 
Or if you prefer, you can execute it set-uid root.

## TODOs

* Fix elephant-in-the-room bug goping#2
* Better test code coverage.
* Support IPV6 addresses.
