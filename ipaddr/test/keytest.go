package test

import (
	"fmt"
	"github.com/seancfoley/ipaddress-go/ipaddr"
	"sync/atomic"
)

type keyTester struct {
	testBase
}

var didKeyTest int32

func (t keyTester) run() {
	didIt := atomic.LoadInt32(&didKeyTest)
	if didIt == 0 {
		cached := t.getAllCached()
		cachedMAC := t.getAllMACCached()
		if len(cached) > 0 || len(cachedMAC) > 0 {
			swapped := atomic.CompareAndSwapInt32(&didKeyTest, 0, 1)
			if swapped {
				if len(cached) > 0 {
					zeroAddr := ipaddr.Address{}
					zeroIPAddr := ipaddr.IPAddress{}
					zero4Addr := ipaddr.IPv4Address{}
					zero6Addr := ipaddr.IPv6Address{}
					cached = append(cached, &zeroIPAddr, zero4Addr.ToIP(), zero6Addr.ToIP())

					//fmt.Printf("testing %d IPs\n", len(cached))
					testGenericKeys[*ipaddr.IPAddress](t, cached)

					addrs := make([]*ipaddr.Address, 0, len(cached)+4)
					for _, addr := range cached {
						addrs = append(addrs, addr.ToAddressBase())
					}
					addrs = append(addrs, &zeroAddr, zeroIPAddr.ToAddressBase(), zero4Addr.ToAddressBase(), zero6Addr.ToAddressBase())
					testGenericKeys[*ipaddr.Address](t, addrs)
					t.testKeys(addrs)

					ipv4Addrs := make([]*ipaddr.IPv4Address, 0, len(cached)+1)
					for _, addr := range cached {
						if addr.IsIPv4() {
							ipv4Addrs = append(ipv4Addrs, addr.ToIPv4())
						}
					}
					ipv4Addrs = append(ipv4Addrs, &zero4Addr)
					testGenericKeys[*ipaddr.IPv4Address](t, ipv4Addrs)

					ipv6Addrs := make([]*ipaddr.IPv6Address, 0, len(cached)+1)
					for _, addr := range cached {
						if addr.IsIPv6() {
							ipv6Addrs = append(ipv6Addrs, addr.ToIPv6())
						}
					}
					ipv6Addrs = append(ipv6Addrs, &zero6Addr)
					testGenericKeys[*ipaddr.IPv6Address](t, ipv6Addrs)

					t.testNetNetIPs(cached)
					t.testNetIPAddrs(cached)
					t.testNetIPs(cached)
				}
				if len(cachedMAC) > 0 {
					zeroAddr := ipaddr.Address{}
					zeroMACAddr := ipaddr.MACAddress{}
					cachedMAC = append(cachedMAC, &zeroMACAddr)

					//fmt.Printf("testing %d MACS\n", len(cachedMAC))
					testGenericKeys[*ipaddr.MACAddress](t, cachedMAC)

					addrs := make([]*ipaddr.Address, 0, len(cached)+1)
					for _, addr := range cached {
						addrs = append(addrs, addr.ToAddressBase())
					}
					addrs = append(addrs, &zeroAddr, zeroMACAddr.ToAddressBase())
					testGenericKeys[*ipaddr.Address](t, addrs)
					t.testKeys(addrs)
				}
			}
		}
	}

	key4 := ipaddr.Key[*ipaddr.IPv4Address]{}
	key6 := ipaddr.Key[*ipaddr.IPv6Address]{}
	ipKey := ipaddr.Key[*ipaddr.IPAddress]{}
	macKey := ipaddr.Key[*ipaddr.MACAddress]{}
	key := ipaddr.Key[*ipaddr.Address]{}

	zeroAddr := ipaddr.Address{}

	zeroMACAddr := ipaddr.MACAddress{}

	keyEquals(t, key4, zero4Addr.ToGenericKey())
	keyEquals(t, key6, zero6Addr.ToGenericKey())
	keyEquals(t, ipKey, zeroIPAddr.ToGenericKey())
	keyEquals(t, key, zeroAddr.ToGenericKey())
	keyEquals(t, macKey, zeroMACAddr.ToGenericKey())

	equals(t, key4.ToAddress(), &zero4Addr)
	equals(t, key6.ToAddress(), &zero6Addr)
	equals(t, macKey.ToAddress(), &zeroMACAddr)
	equals(t, ipKey.ToAddress(), &zeroIPAddr)
	equals(t, key.ToAddress(), &zeroAddr)

	ipv4key := ipaddr.IPv4AddressKey{}
	ipv6key := ipaddr.IPv6AddressKey{}
	macAddrKey := ipaddr.MACAddressKey{}

	keyEquals(t, ipv4key, zero4Addr.ToKey())
	keyEquals(t, ipv6key, zero6Addr.ToKey())
	keyEquals(t, macAddrKey, zeroMACAddr.ToKey())

	equals(t, ipv4key.ToAddress(), &zero4Addr)
	equals(t, ipv6key.ToAddress(), &zero6Addr)
	equals(t, macAddrKey.ToAddress(), &zeroMACAddr)
}

var (
	zeroIPAddr = ipaddr.IPAddress{}
	zero4Addr  = ipaddr.IPv4Address{}
	zero6Addr  = ipaddr.IPv6Address{}
)

func keyEquals[TE interface {
	addFailure(failure)
}, T interface {
	comparable
	fmt.Stringer
}](t TE, one, two T) {
	if two != one {
		f := newAddrFailure("comparison of "+one.String()+" with "+two.String(), nil)
		t.addFailure(f)
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

func (t keyTester) testKeys(cached []*ipaddr.Address) {
	// test that key creation and address creation from keys works
	//var ipv4Count, ipv6Count, macCount uint64
	for _, addr := range cached {
		addr2 := addr.ToKey().ToAddress()
		equals(t, addr, addr2)
		if ipAddr := addr.ToIP(); ipAddr != nil {
			other := ipAddr.ToKey().ToAddress()
			equals(t, ipAddr, other)

			ipRange := ipaddr.NewSequentialRange(ipAddr, &zeroIPAddr)
			ipRangeBack := ipRange.ToKey().ToSeqRange()
			equals(t, ipRangeBack.GetLower(), &zeroIPAddr)
			equals(t, ipRangeBack.GetUpper(), ipAddr)

			if !ipAddr.IsMax() {
				oneUp := ipAddr.Increment(1)
				ipRange := ipaddr.NewSequentialRange(ipAddr, oneUp)
				ipRangeBack := ipRange.ToKey().ToSeqRange()
				equals(t, ipRangeBack.GetUpper(), oneUp)
				equals(t, ipRangeBack.GetLower(), ipAddr)
			}
		}
		if addrv4 := addr.ToIPv4(); addrv4 != nil {
			other := addrv4.ToKey().ToAddress()
			equals(t, addrv4, other)

			ipRange := ipaddr.NewSequentialRange(addrv4, &zero4Addr)
			ipRangeBack := ipRange.ToKey().ToSeqRange()
			equals(t, ipRangeBack.GetLower(), &zero4Addr)
			equals(t, ipRangeBack.GetUpper(), addrv4)

			if !addrv4.IsMax() {
				oneUp := addrv4.Increment(1)
				ipRange := ipaddr.NewSequentialRange(addrv4, oneUp)
				ipRangeBack := ipRange.ToKey().ToSeqRange()
				equals(t, ipRangeBack.GetUpper(), oneUp)
				equals(t, ipRangeBack.GetLower(), addrv4)
			}

			//ipv4Count++
		}
		if addrv6 := addr.ToIPv6(); addrv6 != nil {
			other := addrv6.ToKey().ToAddress()
			equals(t, addrv6, other)

			ipRange := ipaddr.NewSequentialRange(addrv6, &zero6Addr)
			ipRangeBack := ipRange.ToKey().ToSeqRange()
			equals(t, ipRangeBack.GetLower(), &zero6Addr)
			equals(t, ipRangeBack.GetUpper(), addrv6)

			if !addrv6.IsMax() {
				oneUp := addrv6.Increment(1)
				ipRange := ipaddr.NewSequentialRange(addrv6, oneUp)
				ipRangeBack := ipRange.ToKey().ToSeqRange()
				equals(t, ipRangeBack.GetUpper(), oneUp)
				equals(t, ipRangeBack.GetLower(), addrv6)
			}

			//ipv6Count++
		}
		if addrmac := addr.ToMAC(); addrmac != nil {
			other := addrmac.ToKey().ToAddress()
			equals(t, addrmac, other)
			//macCount++
		}
		t.incrementTestCount()
	}
}

type AddrConstraint[T ipaddr.KeyConstraint[T]] interface {
	ipaddr.GenericKeyConstraint[T]
	ipaddr.AddressType
}

func testGenericKeys[T AddrConstraint[T]](t keyTester, cached []T) {
	for _, addr := range cached {
		addr2 := addr.ToGenericKey().ToAddress()
		equals(t, addr, addr2)
		t.incrementTestCount()
	}
}

func equals[TE interface{ addFailure(failure) }, T ipaddr.AddressType](t TE, one, two T) {
	if !one.Equal(two) || !two.Equal(one) {
		f := newAddrFailure("comparison of "+one.String()+" with "+two.String(), two.ToAddressBase())
		t.addFailure(f)
	} else if one.Compare(two) != 0 || two.Compare(one) != 0 {
		f := newAddrFailure("comparison of "+one.String()+" with "+two.String(), two.ToAddressBase())
		t.addFailure(f)
	}
}
