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

				t.testKeys(cached)
				t.testNetNetIPs(cached)
				t.testNetIPAddrs(cached)
				t.testNetIPs(cached)
				//fmt.Println("trying the keys ", len(cached), " ipv4 ", ipv4Count, " ipv6 ", ipv6Count)

				//../../sdk/go1.19.2/bin/go list -f '{{.Version}}' -m github.com/seancfoley/ipaddress-go@master
			}
		}
	}
}

func (t keyTester) testNetNetIPs(cached []*ipaddr.IPAddress) {
	// test that key creation and address creation from keys works
	for _, addr := range cached {
		addr1Lower := addr.GetLower()
		addr1Upper := addr.GetLower()

		addr2Lower := ipaddr.NewIPAddressFromNetNetIPAddr(addr1Lower.GetNetNetIPAddr())
		addr2Upper := ipaddr.NewIPAddressFromNetNetIPAddr(addr1Upper.GetNetNetIPAddr())

		equals(t, addr1Lower, addr2Lower)
		equals(t, addr1Upper, addr2Upper)

		if addrv4 := addr.ToIPv4(); addrv4 != nil {
			addr1Lower := addrv4.GetLower()
			addr1Upper := addrv4.GetLower()

			addr2Lower := ipaddr.NewIPAddressFromNetNetIPAddr(addr1Lower.GetNetNetIPAddr()).ToIPv4()
			addr2Upper := ipaddr.NewIPAddressFromNetNetIPAddr(addr1Upper.GetNetNetIPAddr()).ToIPv4()

			equals(t, addr1Lower, addr2Lower)
			equals(t, addr1Upper, addr2Upper)
		}
		if addrv6 := addr.ToIPv6(); addrv6 != nil {
			addr1Lower := addrv6.GetLower()
			addr1Upper := addrv6.GetLower()

			addr2Lower := ipaddr.NewIPAddressFromNetNetIPAddr(addr1Lower.GetNetNetIPAddr()).ToIPv6()
			addr2Upper := ipaddr.NewIPAddressFromNetNetIPAddr(addr1Upper.GetNetNetIPAddr()).ToIPv6()

			equals(t, addr1Lower, addr2Lower)
			equals(t, addr1Upper, addr2Upper)
		}
	}
}

//var m sync.Mutex

func (t keyTester) testNetIPAddrs(cached []*ipaddr.IPAddress) {
	// test that key creation and address creation from keys works
	for _, addr := range cached {
		addr1Lower := addr.GetLower()
		addr1Upper := addr.GetLower()

		if addr.IsIPv6() {
			if addr1Lower.ToIPv6().IsIPv4Mapped() { // net.IP will switch to IPv4, so we might as well just do that ourselves
				addr4, _ := addr1Lower.ToIPv6().GetEmbeddedIPv4Address()
				addr1Lower = addr4.ToIP()
			}
			if addr1Upper.ToIPv6().IsIPv4Mapped() {
				addr4, _ := addr1Upper.ToIPv6().GetEmbeddedIPv4Address()
				addr1Upper = addr4.ToIP()
			}
		}

		addr2Lower, _ := ipaddr.NewIPAddressFromNetIPAddr(addr1Lower.GetNetIPAddr())
		addr2Upper, _ := ipaddr.NewIPAddressFromNetIPAddr(addr1Upper.GetUpperNetIPAddr())

		equals(t, addr1Lower, addr2Lower)
		equals(t, addr1Upper, addr2Upper)

		if addrv4 := addr.ToIPv4(); addrv4 != nil {
			addr1Lower := addrv4.GetLower()
			addr1Upper := addrv4.GetLower()

			addr2Lower, _ := ipaddr.NewIPAddressFromNetIPAddr(addr1Lower.GetNetIPAddr())
			addr2Upper, _ := ipaddr.NewIPAddressFromNetIPAddr(addr1Upper.GetUpperNetIPAddr())

			addr2Lower3 := addr2Lower.ToIPv4()
			addr2Upper3 := addr2Upper.ToIPv4()

			equals(t, addr1Lower, addr2Lower3)
			equals(t, addr1Upper, addr2Upper3)
		}
		if addrv6 := addr.ToIPv6(); addrv6 != nil {
			addr1Lower := addrv6.GetLower()
			addr1Upper := addrv6.GetLower()

			if !addr1Lower.IsIPv4Mapped() { // net.IP will switch to IPv4, so we might as well just do that ourselves
				addr2Lower, _ := ipaddr.NewIPAddressFromNetIPAddr(addr1Lower.GetNetIPAddr())
				addr2Lower3 := addr2Lower.ToIPv6()
				equals(t, addr1Lower, addr2Lower3)
			}
			if !addr1Upper.IsIPv4Mapped() {
				addr2Upper, _ := ipaddr.NewIPAddressFromNetIPAddr(addr1Upper.GetUpperNetIPAddr())
				addr2Upper3 := addr2Upper.ToIPv6()
				equals(t, addr1Upper, addr2Upper3)
			}
		}
		t.incrementTestCount()
	}
}

func (t keyTester) testNetIPs(cached []*ipaddr.IPAddress) {
	// test that key creation and address creation from keys works
	for _, addr := range cached {
		addr1Lower := addr.GetLower()
		addr1Upper := addr.GetLower()
		if addr.IsIPv6() {
			if addr.ToIPv6().HasZone() { // net.IP cannot store zone, so we need to drop it to check equality
				addr1Lower = addr1Lower.ToIPv6().WithoutZone().ToIP()
				addr1Upper = addr1Upper.ToIPv6().WithoutZone().ToIP()
			}
			if addr1Lower.ToIPv6().IsIPv4Mapped() { // net.IP will switch to IPv4, so we might as well just do that ourselves
				addr4, _ := addr1Lower.ToIPv6().GetEmbeddedIPv4Address()
				addr1Lower = addr4.ToIP()
			}
			if addr1Upper.ToIPv6().IsIPv4Mapped() {
				addr4, _ := addr1Upper.ToIPv6().GetEmbeddedIPv4Address()
				addr1Upper = addr4.ToIP()
			}
		}

		addr2Lower, _ := ipaddr.NewIPAddressFromNetIP(addr1Lower.GetNetIP())
		addr2Upper, _ := ipaddr.NewIPAddressFromNetIP(addr1Upper.GetUpperNetIP())
		equals(t, addr1Lower, addr2Lower)

		equals(t, addr1Upper, addr2Upper)

		if addrv4 := addr.ToIPv4(); addrv4 != nil {
			addr1Lower := addrv4.GetLower()
			addr1Upper := addrv4.GetLower()

			addr2Lower, _ := ipaddr.NewIPAddressFromNetIPAddr(addr1Lower.GetNetIPAddr())
			addr2Upper, _ := ipaddr.NewIPAddressFromNetIPAddr(addr1Upper.GetUpperNetIPAddr())

			addr2Lower3 := addr2Lower.ToIPv4()
			addr2Upper3 := addr2Upper.ToIPv4()

			equals(t, addr1Lower, addr2Lower3)
			equals(t, addr1Upper, addr2Upper3)
		}
		if addrv6 := addr.ToIPv6(); addrv6 != nil {
			addr1Lower := addrv6.GetLower()
			addr1Upper := addrv6.GetLower()

			if addrv6.HasZone() { // net.IP cannot store zone, so we need to drop it to check equality
				addr1Lower = addr1Lower.WithoutZone()
				addr1Upper = addr1Upper.WithoutZone()
			}
			if !addr1Lower.IsIPv4Mapped() { // net.IP will switch to IPv4, so we might as well just do that ourselves
				addr2Lower, _ := ipaddr.NewIPAddressFromNetIPAddr(addr1Lower.GetNetIPAddr())
				addr2Lower3 := addr2Lower.ToIPv6()
				equals(t, addr1Lower, addr2Lower3)
			}
			if !addr1Upper.IsIPv4Mapped() {
				addr2Upper, _ := ipaddr.NewIPAddressFromNetIPAddr(addr1Upper.GetUpperNetIPAddr())
				addr2Upper3 := addr2Upper.ToIPv6()
				equals(t, addr1Upper, addr2Upper3)
			}
		}
		t.incrementTestCount()
	}
}

func (t keyTester) testKeys(cached []*ipaddr.IPAddress) {
	// test that key creation and address creation from keys works
	//ipv4Count, ipv6Count := 0, 0
	for _, addr := range cached {
		addr2 := addr.ToKey().ToAddress()
		equals(t, addr, addr2)

		if addrv4 := addr.ToIPv4(); addrv4 != nil {
			other := addrv4.ToKey().ToAddress()
			equals(t, addrv4, other)
			//ipv4Count++
		}
		if addrv6 := addr.ToIPv6(); addrv6 != nil {
			other := addrv6.ToKey().ToAddress()
			equals(t, addrv6, other)
			//ipv6Count++
		}
		t.incrementTestCount()
	}
}

func equals[TE interface{ addFailure(failure) }, T ipaddr.IPAddressType](t TE, one, two T) {
	if !one.Equal(two) || !two.Equal(one) {
		f := newIPAddrFailure("comparison of "+one.String()+" with "+two.String(), two.ToIP())
		t.addFailure(f)
	} else if one.Compare(two) != 0 || two.Compare(one) != 0 {
		f := newIPAddrFailure("comparison of "+one.String()+" with "+two.String(), two.ToIP())
		t.addFailure(f)
	}
}
