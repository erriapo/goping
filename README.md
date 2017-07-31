[![Code Climate](https://codeclimate.com/github/erriapo/goping/badges/gpa.svg)](https://codeclimate.com/github/erriapo/goping)

## Dependencies

* golang.org/x/net
* github.com/erriapo/stats 

## Usage example

You must supply the target host with a -d option.

```bash
$ bin/goping -d localhost
.Reverse lookup took 139.777418ms

PING 127.0.0.1 (localhost) 56(64) bytes of data.
64 bytes from (127.0.0.1): icmp_req=1 time=218.105µs
64 bytes from (127.0.0.1): icmp_req=2 time=202.187µs
64 bytes from (127.0.0.1): icmp_req=3 time=130.121µs
64 bytes from (127.0.0.1): icmp_req=4 time=121.453µs
64 bytes from (127.0.0.1): icmp_req=5 time=120.818µs

--- localhost ping statistics ---
5 packets transmitted, 5 received, 0% packet loss
```

Additionally, the `goping` binary needs the CAP_NET_RAWIO capability. 
Or if you prefer, you can execute it set-uid root.

## TODOs

* Better test code coverage.
* Support IPV6 addresses.
