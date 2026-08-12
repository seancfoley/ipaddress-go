# ipaddress-go

[Go](https://golang.org/) library for handling IP addresses and subnets, both IPv4 and IPv6

IP address and network manipulations, CIDR, address and subnet operations, address ranges, iterators, containment checks, longest prefix match, subnetting, address tries, address collections, and more, with polymorphic code

[Docs Site](https://seancfoley.github.io/IPAddress/)

[Code Examples](https://github.com/seancfoley/ipaddress-go/wiki/Code-Examples)

[Godoc](https://pkg.go.dev/github.com/seancfoley/ipaddress-go/ipaddr) [![Go Reference](https://pkg.go.dev/badge/github.com/seancfoley/ipaddress-go/ipaddr.svg)](https://pkg.go.dev/github.com/seancfoley/ipaddress-go/ipaddr)

[List of Users](https://github.com/seancfoley/ipaddress-go/wiki)

| Version | Notes         |
| ------- | ------------- |
| [1.2.1](https://github.com/seancfoley/ipaddress-go/releases/tag/v1.2.1) | Requires Go 1.12 or higher |
| [1.4.1](https://github.com/seancfoley/ipaddress-go/releases/tag/v1.4.1) | Requires Go 1.13 or higher |
| [1.8.3](https://github.com/seancfoley/ipaddress-go/releases/tag/v1.8.3) | Requires Go 1.18 or higher |

In your go.mod file:\
require github.com/seancfoley/ipaddress-go v1.8.3

In your source file:\
import "github.com/seancfoley/ipaddress-go/ipaddr"

Packaged as a [Linux Fedora rpm](https://src.fedoraproject.org/rpms/golang-github-seancfoley-ipaddress) and as a [Linux Debian](https://tracker.debian.org/pkg/golang-github-seancfoley-ipaddress-go), available from [Ubuntu](https://launchpad.net/ubuntu/+source/golang-github-seancfoley-ipaddress-go)

Also available as a [Java](https://www.oracle.com/java/) library from the [IPAddress repository](https://github.com/seancfoley/IPAddress)

## Getting Started

starting with address or subnet strings
```go
import "github.com/seancfoley/ipaddress-go/ipaddr"

ipv6AddrStr := ipaddr.NewIPAddressString("a:b:c:d::a:b/64")
if ipAddr, err := ipv6AddrStr.ToAddress(); err != nil {
	// err.Error() has validation error
} else {
	// use the address
}
```
...or avoid errors, checking for nil:
```go
str := ipaddr.NewIPAddressString("a:b:c:d:e-f:f:1.2-3.3.4/64")
addr := str.GetAddress()
if addr != nil {
	// use address
}
```
starting with host name strings
```go
hostStr := "[::1]"

host := ipaddr.NewHostName(hostStr)
err := host.Validate()
if err == nil {
	if host.IsAddress() {
		fmt.Println("address: " + host.AsAddress().String())
	} else {
		fmt.Println("host name: " + host.String())
	}
	// use host
} else {
	fmt.Println(err.Error())
}
```


