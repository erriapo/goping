## Dependencies

`goping`only works on unixes & is only tested with golang 1.8 and 1.10

* [golang.org/x/net/icmp](https://godoc.org/golang.org/x/net/icmp)
* [golang.org/x/net/ipv4](https://godoc.org/golang.org/x/net/ipv4)
* [github.com/erriapo/stats](https://github.com/erriapo/stats)
* [golang.org/x/net/idna](https://godoc.org/golang.org/x/net/idna)

## Usage example

```bash
$ goping -c 2 8.8.4.4
.
PING 8.8.4.4 (8.8.4.4) 56(84) bytes of data.
64 bytes from google-public-dns-b.google.com. (8.8.4.4): icmp_seq=1 time=838.022µs
64 bytes from google-public-dns-b.google.com. (8.8.4.4): icmp_seq=2 time=978.804µs

--- 8.8.4.4 ping statistics ---
2 packets transmitted, 2 received, 0% packet loss
rtt min/avg/max/mdev = 0.838/0.908/0.979/0.1 ms



$ goping -c 4 xn--bdk.ws
.                                                                                              
PING ツ.ws. (132.148.137.119) 56(84) bytes of data.                                            
64 bytes from ip-132-148-137-119.ip.secureserver.net. (132.148.137.119): icmp_seq=1 time=33.892283ms
64 bytes from ip-132-148-137-119.ip.secureserver.net. (132.148.137.119): icmp_seq=2 time=33.402274ms
64 bytes from ip-132-148-137-119.ip.secureserver.net. (132.148.137.119): icmp_seq=3 time=33.361368ms
64 bytes from ip-132-148-137-119.ip.secureserver.net. (132.148.137.119): icmp_seq=4 time=33.486581ms                                                                                          
--- ツ.ws. ping statistics ---                                                                 
4 packets transmitted, 4 received, 0% packet loss                                              
rtt min/avg/max/mdev = 33.361/33.536/33.892/0.243 ms              
```

Additionally, the `goping` binary needs the CAP_NET_RAWIO capability. 
Or if you prefer, you can execute it set-uid root.

## TODOs

* Fix bug [#2](../../issues/2)
* Better test code coverage.
* Support IPV6 addresses.
