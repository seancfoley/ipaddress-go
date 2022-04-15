# ipaddress-go

[Go](https://golang.org/) library for handling IP addresses and subnets, both IPv4 and IPv6

IP network manipulation, operations, iterations, containment operations and data structures, with polymporphic code. 

[View Project Page](https://seancfoley.github.io/IPAddress/)

Note: godoc is actively being worked

[View Godoc](https://pkg.go.dev/github.com/seancfoley/ipaddress-go) [![Go Reference](https://pkg.go.dev/badge/github.com/seancfoley/ipaddress-go/ipaddr.svg)](https://pkg.go.dev/github.com/seancfoley/ipaddress-go)

[View Code Examples](https://github.com/seancfoley/ipaddress-go/wiki/Code-Examples)

Also available as a [Java](https://www.oracle.com/java/) library from the [IPAddress repository](https://github.com/seancfoley/IPAddress)

Version | Notes         |
------- | ------------- |
[1.1.2](https://github.com/seancfoley/ipaddress-go/releases/tag/v1.1.2) | Requires Go 1.12 or higher |

go get github.com/seancfoley/ipaddress-go@v1.1.2

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


