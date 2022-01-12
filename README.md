# ipaddress-go

The [**Go**](https://golang.org/) IPAddress library

go get github.com/seancfoley/ipaddress-go/ipaddr

[View Project Page](https://seancfoley.github.io/IPAddress/)

View Godoc - coming soon

[View Code Examples](https://github.com/seancfoley/ipaddress-go/wiki/Code-Examples)

The IPAddress library is also [available for Java](https://github.com/seancfoley/IPAddress).  The API is similar to the Go implementation, the primary differences being the nuances of the Go language in comparison to Java.

Version | Notes         |
------- | ------------- |
[1.0.0](https://github.com/seancfoley/ipaddress-go/releases/tag/v1.0.0) | Requires Go 12.7 higher |

## Getting Started

starting with address or subnet strings
```go
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


