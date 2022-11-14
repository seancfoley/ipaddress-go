package test

import (
	"github.com/seancfoley/ipaddress-go/ipaddr"
	"sync/atomic"
)

type keyTester struct {
	testBase
}

var didKeyTest int32

func (t keyTester) run() {

	didKeyTest := atomic.LoadInt32(&didKeyTest)
	if didKeyTest == 0 {
		cached := t.getAllCached()
		if len(cached) > 0 {
			swapped := atomic.CompareAndSwapInt32(&didKeyTest, 0, 1)
			if swapped {
				zeroAddr := ipaddr.IPAddress{}
				zero4Addr := ipaddr.IPv4Address{}
				zero6Addr := ipaddr.IPv6Address{}
				cached = append(cached, &zeroAddr, zero4Addr.ToIP(), zero6Addr.ToIP())

				// test that key creation and address creation from keys works
				//ipv4Count, ipv6Count := 0, 0
				for _, addr := range cached {
					addr2 := addr.ToKey().ToAddress()
					if !addr.Equal(addr2) || !addr2.Equal(addr) {
						t.addFailure(newIPAddrFailure("key compare of "+addr.String(), addr2))
					} else if addr.Compare(addr2) != 0 || addr2.Compare(addr) != 0 {
						t.addFailure(newIPAddrFailure("key compare of "+addr.String(), addr2))
					}

					if addrv4 := addr.ToIPv4(); addrv4 != nil {
						other := addrv4.ToKey().ToAddress()
						if !addrv4.Equal(other) || !other.Equal(addrv4) {
							t.addFailure(newIPAddrFailure("key compare of "+addrv4.String(), other.ToIP()))
						} else if addrv4.Compare(other) != 0 || other.Compare(addrv4) != 0 {
							t.addFailure(newIPAddrFailure("key compare of "+addrv4.String(), other.ToIP()))
						}
						//ipv4Count++
					}
					if addrv6 := addr.ToIPv6(); addrv6 != nil {
						other := addrv6.ToKey().ToAddress()
						if !addrv6.Equal(other) || !other.Equal(addrv6) {
							t.addFailure(newIPAddrFailure("key compare of "+addrv6.String(), other.ToIP()))
						} else if addrv6.Compare(other) != 0 || other.Compare(addrv6) != 0 {
							t.addFailure(newIPAddrFailure("key compare of "+addrv6.String(), other.ToIP()))
						}
						//ipv6Count++
					}
				}
				//fmt.Println("trying the keys ", len(cached), " ipv4 ", ipv4Count, " ipv6 ", ipv6Count)
			}
		}
	}
	t.incrementTestCount()

}
