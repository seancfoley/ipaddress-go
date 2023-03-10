//
// Copyright 2020-2022 Sean C Foley
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"fmt"
	"math"
	"net"
	"reflect"
	"strings"

	//"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"

	"github.com/seancfoley/ipaddress-go/ipaddr"
	"github.com/seancfoley/ipaddress-go/ipaddr/addrstrparam"
)

// this is just a test program used for trying out code
func main() {

	zeroipaddressString := ipaddr.IPAddressString{}
	fmt.Println(zeroipaddressString.GetAddress())

	fmt.Println(ipaddr.IPv4Address{})
	seg := ipaddr.IPv4AddressSegment{}

	seg.GetSegmentValue()

	fmt.Printf("%v\n", seg.GetBitCount())
	fmt.Printf("%v\n", seg.GetByteCount())

	grouping := ipaddr.IPv4AddressSection{}
	grouping.GetSegmentCount()

	builder := addrstrparam.IPAddressStringParamsBuilder{}
	params := builder.AllowAll(false).ToParams()
	fmt.Printf("%+v\n", params)

	//params := ipaddr.ipAddressStringParameters{}
	////fmt.Printf("%+v\n", params)
	//init := ipaddr.IPAddressStringParamsBuilder{}
	//params2 := init.AllowAll(false).ToParams()
	//params = *params2
	//_ = params
	////fmt.Printf("%+v\n", params)

	i := -1
	b := byte(i)
	fmt.Printf("byte is %+v\n", b)

	var slc []int
	fmt.Printf("%+v\n", slc) // expecting []
	fmt.Printf("%v\n", slc)  // expecting []
	fmt.Printf("%v\n", slc)  // expecting []

	addr := ipaddr.IPv6Address{}
	fmt.Printf("zero addr is %+v\n", addr)
	fmt.Printf("zero addr is %+v\n", &addr)

	addr4 := ipaddr.IPv4Address{}
	fmt.Printf("zero addr is %+v\n", addr4)
	addr2 := addr4.ToIP()
	fmt.Printf("zero addr is %+v\n", addr2)
	_ = addr2.String()
	_ = addr2.GetSection()
	fmt.Printf("zero addr is %+v\n", addr2.String())
	//fmt.Printf("%+v\n", &addr2)

	ipv4Prefixed := addr4.ToPrefixBlockLen(16)
	fmt.Printf("16 block is %+v\n", ipv4Prefixed)
	fmt.Printf("lower is %+v\n", ipv4Prefixed.GetLower())
	fmt.Printf("upper is %+v\n", ipv4Prefixed.GetUpper())
	fmt.Printf("lower is %+v\n", ipv4Prefixed.GetLower())
	fmt.Printf("upper is %+v\n", ipv4Prefixed.GetUpper())

	_ = addr.GetPrefixCount() // an inherited method

	addr5 := ipaddr.IPAddress{} // expecting []
	fmt.Printf("%+v\n", addr5)
	addr5Upper := addr5.GetUpper()
	fmt.Printf("%+v\n", addr5Upper) // expecting []
	addr6 := addr5Upper.ToIPv4()
	fmt.Printf("%+v\n", addr6) // expecting <nil>

	addrSection := ipaddr.AddressSection{}
	fmt.Printf("%+v\n", addrSection) // expecting [] or <nil>

	ipAddrSection := ipaddr.IPAddressSection{}
	fmt.Printf("%+v\n", ipAddrSection) // expecting [] or <nil>

	ipv4AddrSection := ipaddr.IPv4AddressSection{}
	fmt.Printf("%+v\n", ipv4AddrSection) // expecting [] or <nil>

	//addrStr := ipaddr.IPAddressString{}
	addrStr := ipaddr.NewIPAddressString("1.2.3.4")
	pAddr := addrStr.GetAddress()
	fmt.Printf("%+v\n", *pAddr)
	fmt.Printf("%+v\n", pAddr)

	//fmt.Printf("All the formats: %v %x %X %o %O %b %d %#x %#o %#b\n",
	//	pAddr, pAddr, pAddr, pAddr, pAddr, pAddr, pAddr, pAddr, pAddr, pAddr)
	fmt.Printf("All the formats: default %v\nstring %s\nquoted %q\nquoted backtick %#q\nlowercase hex %x\nuppercase hex %X\nlower hex prefixed %#x\nupper hex prefixed %#X\noctal no prefix %o\noctal prefixed %O\noctal 0 prefix %#o\nbinary %b\nbinary prefixed %#b\ndecimal %d\n\n",
		pAddr, pAddr, pAddr, pAddr, pAddr, pAddr, pAddr, pAddr, pAddr, pAddr, pAddr, pAddr, pAddr, pAddr)
	//fmt.Printf("All the formats: %v %x %X %o %O %b %d %#x %#o %#b\n",
	//	*pAddr, *pAddr, *pAddr, *pAddr, *pAddr, *pAddr, *pAddr, *pAddr, *pAddr, *pAddr)
	//fmt.Printf("octal no prefix %o\n", *pAddr)
	//fmt.Printf("octal prefixed %O\n", *pAddr)
	//fmt.Printf("octal 0 prefix %#o\n", *pAddr)
	//fmt.Printf("binary no prefix %b\n", *pAddr)
	//fmt.Printf("binary prefixed %#b\n", *pAddr)

	pAddr = addrStr.GetAddress() // test getting it a second time from the cache
	fmt.Printf("%+v\n", *pAddr)
	fmt.Printf("%+v\n", pAddr)

	cidrStr := ipaddr.NewIPAddressString("255.2.0.0/16")
	cidr := cidrStr.GetAddress()
	fmt.Printf("All the formats: default %v\nstring %s\nquoted %q\nquoted backtick %#q\nlowercase hex %x\nuppercase hex %X\nlower hex prefixed %#x\nupper hex prefixed %#X\noctal no prefix %o\noctal prefixed %O\noctal 0 prefix %#o\nbinary %b\nbinary prefixed %#b\ndecimal %d\n\n",
		cidr, cidr, cidr, cidr, cidr, cidr, cidr, cidr, cidr, cidr, cidr, cidr, cidr, cidr)

	pZeroSec := ipaddr.IPv4AddressSection{}
	//fmt.Printf("octal no prefix %o\noctal prefixed %O\noctal 0 prefix %#o\ndecimal %d\n\n",
	//	pZeroSec, pZeroSec, pZeroSec, pZeroSec)

	fmt.Printf("All the formats for zero section: default %v\nstring %s\nquoted %q\nquoted backtick %#q\nlowercase hex %x\nuppercase hex %X\nlower hex prefixed %#x\nupper hex prefixed %#X\noctal no prefix %o\noctal prefixed %O\noctal 0 prefix %#o\nbinary %b\nbinary prefixed %#b\ndecimal %d\n\n",
		pZeroSec, pZeroSec, pZeroSec, pZeroSec, pZeroSec, pZeroSec, pZeroSec, pZeroSec, pZeroSec, pZeroSec, pZeroSec, pZeroSec, pZeroSec, pZeroSec)

	addrStr = ipaddr.NewIPAddressString("abc.2.3.4")
	noAddr, err := addrStr.ToAddress()
	fmt.Printf("invalid string abc.2.3.4 is %v with err %v\n", noAddr, err)

	ipv4Prefixed2 := pAddr.ToPrefixBlockLen(19)
	fmt.Printf("19 block is %+v\n", ipv4Prefixed2)

	addrStr = ipaddr.NewIPAddressString("a:b:c:d:e:f:a:b")
	pAddr = addrStr.GetAddress()
	fmt.Printf("%+v\n", *pAddr)
	fmt.Printf("%+v\n", pAddr)

	addrStr = ipaddr.NewIPAddressString("a:b:c:d:e:f:a:b%eth0")
	pAddr = addrStr.GetAddress()
	fmt.Printf("%+v\n", *pAddr)
	fmt.Printf("%+v\n", pAddr)

	addrStr = ipaddr.NewIPAddressString("a:b:c:d:e:f:1.2.3.4")
	pAddr = addrStr.GetAddress()
	fmt.Printf("%+v\n", *pAddr)
	fmt.Printf("%+v\n", pAddr)

	ipv4Addr, _ := ipaddr.NewIPv4AddressFromBytes([]byte{1, 0, 1, 0})
	fmt.Printf("%+v\n", ipv4Addr)
	fmt.Printf("%+v\n", *ipv4Addr)

	ipv4Addr, ipv4Err := ipaddr.NewIPv4AddressFromBytes([]byte{1, 1, 0, 1, 0})
	fmt.Printf("%+v %+v\n", ipv4Addr, ipv4Err)

	ipv6Addr, ipv6Err := ipaddr.NewIPv6AddressFromBytes(net.IP{1, 0, 1, 0, 0xff, 0xa, 0xb, 0xc, 1, 0, 1, 0, 0xff, 0xa, 0xb, 0xc})
	fmt.Printf("%+v %+v\n", ipv6Addr, ipv6Err)
	fmt.Printf("%+v\n", *ipv6Addr)
	fmt.Printf("All the formats: default %v\nstring %s\nlowercase hex %x\nuppercase hex %X\nlower hex prefixed %#x\nupper hex prefixed %#X\noctal no prefix %o\noctal prefixed %O\noctal 0 prefix %#o\nbinary %b\nbinary prefixed %#b\ndecimal %d\n\n",
		ipv6Addr, ipv6Addr, ipv6Addr, ipv6Addr, ipv6Addr, ipv6Addr, ipv6Addr, ipv6Addr, ipv6Addr, ipv6Addr, ipv6Addr, ipv6Addr)
	//ipv6Addr = nil
	//fmt.Printf("All the formats: %v %x %X %o %O %b %#x %#o %#b\n",
	//	ipv6Addr, ipv6Addr, ipv6Addr, ipv6Addr, ipv6Addr, ipv6Addr, ipv6Addr, ipv6Addr, ipv6Addr)

	fmt.Println(ipv6Addr)
	ipv6Addr.ForEachSegment(func(i int, seg *ipaddr.IPv6AddressSegment) bool {
		fmt.Printf("visiting %d seg %s\n", i, seg)
		return false
	})
	base85Str, _ := ipv6Addr.ToBase85String()
	fmt.Println("Base 85 string is", base85Str, "for", ipv6Addr)

	ipv4Addr, _ = ipaddr.NewIPv4AddressFromBytes([]byte{1, 0, 1, 0})
	fmt.Println()
	fmt.Println(ipv4Addr)
	ipv4Addr.ForEachSegment(func(i int, seg *ipaddr.IPv4AddressSegment) bool {
		fmt.Printf("visiting %d seg %s\n", i, seg)
		return false
	})

	fmt.Println()
	fmt.Println(cidr)
	cidr.ForEachSegment(func(i int, seg *ipaddr.IPAddressSegment) bool {
		fmt.Printf("visiting %d seg %s\n", i, seg)
		return false
	})
	fmt.Println()

	ipv6Prefixed := ipv6Addr.ToPrefixBlockLen(32)
	fmt.Printf("32 block is %+v\n", ipv6Prefixed)
	ipv6Prefixed = ipv6Addr.ToPrefixBlockLen(40)
	fmt.Printf("40 block is %+v\n", ipv6Prefixed)

	mixedGrouping, _ := ipv6Addr.GetMixedAddressGrouping()
	fmt.Printf("mixed grouping of %v is %v and again %s\n", ipv6Addr, mixedGrouping.String(), mixedGrouping)
	mixedGrouping, _ = ipv6Prefixed.GetMixedAddressGrouping()
	fmt.Printf("mixed grouping of 40 block %v is %v and again %s\n", ipv6Prefixed, mixedGrouping.String(), mixedGrouping)

	addrDown := ipv6Prefixed.ToAddressBase()
	fmt.Printf("addr down converted 40 block is %+v\n", addrDown)

	addrUp := addrDown.ToIPv6()
	fmt.Printf("addr up converted 40 block is %+v\n", addrUp)

	addrUpNil := addrDown.ToIPv4()
	fmt.Printf("addr up converted nil is %+v\n", addrUpNil)

	ht := ipaddr.NewHostName("bla.com")
	fmt.Printf("%v\n", ht.ToNormalizedString())
	fmt.Printf("%v\n", ht.GetHost())
	//ip := net.IP{1, 0, 1, 0, 0xff, 0xa, 0xb, 0xc, 1, 0, 1, 0, 0xff, 0xa, 0xb, 0xc}
	//foo(ip)
	//foo2(ip)
	//foo3(net.IPAddr{IP: ip})

	//bytes := []byte{1, 0, 1, 0, 0xff, 0xa, 0xb, 0xc, 1, 0, 1, 0, 0xff, 0xa, 0xb, 0xc}
	//foo(bytes)
	//foo2(bytes)
	//foo3(net.IPAddr{IP: bytes})

	fmt.Printf("iterate a segment:\n")
	iter := addrUp.GetSegment(ipaddr.IPv6SegmentCount - 1).PrefixedBlockIterator(5)
	for iter.HasNext() {
		fmt.Printf("%v ", iter.Next())
	}
	fmt.Printf("\niterate another segment:\n")
	iter = addrUp.GetSegment(ipaddr.IPv6SegmentCount - 1).PrefixedBlockIterator(0)
	for iter.HasNext() {
		fmt.Printf("%v ", iter.Next())
	}

	addrStrPref := ipaddr.NewIPAddressString("1.2-11.0.0/15")
	pAddr = addrStrPref.GetAddress()
	newIter := pAddr.GetSection().PrefixBlockIterator()
	fmt.Println()
	fmt.Printf("to iterate: %+v", pAddr)
	fmt.Println("iterate prefix blocks (prefix len 15):")
	for newIter.HasNext() {
		fmt.Printf("%v ", newIter.Next())
	}
	addrStrPref = ipaddr.NewIPAddressString("1.2-11.0.0/16")
	pAddr = addrStrPref.GetAddress()
	fmt.Println()
	fmt.Printf("to iterate: %+v", pAddr)
	newIter = pAddr.GetSection().BlockIterator(2)
	fmt.Println("iterate a section's first two blocks:")
	for newIter.HasNext() {
		fmt.Printf("%v ", newIter.Next())
	}
	newIter = pAddr.GetSection().SequentialBlockIterator()
	fmt.Printf("\nsequential block iterator:\n")
	for newIter.HasNext() {
		fmt.Printf("%v ", newIter.Next())
	}

	addrStrPref1 := ipaddr.NewIPAddressString("1.2.3.4")
	addrStrPref2 := ipaddr.NewIPAddressString("1.2.4.1")
	rng := addrStrPref1.GetAddress().ToIPv4().SpanWithRange(addrStrPref2.GetAddress().ToIPv4())
	riter := rng.Iterator()
	fmt.Printf("\nsequential range iterator:\n")
	for riter.HasNext() {
		fmt.Printf("%v ", riter.Next())
	}
	riter = rng.PrefixBlockIterator(28)
	fmt.Printf("\nsequential range pref block iterator:\n")
	for riter.HasNext() {
		fmt.Printf("%v ", riter.Next())
	}

	sect := addrStrPref1.GetAddress().ToIPv4().GetSection()
	str := sect.ToCanonicalString()
	fmt.Printf("\nString is %s", str)
	addrStrPref6 := ipaddr.NewIPAddressString("1.2.3.4/16")
	sect = addrStrPref6.GetAddress().ToIPv4().GetSection()
	str = sect.ToCanonicalString()
	fmt.Printf("\nString with prefix length is %s", str)

	ipv4Addr = addrStrPref6.GetAddress().ToIPv4()
	str, _ = ipv4Addr.ToInetAtonJoinedString(ipaddr.Inet_aton_radix_hex, 2)
	fmt.Printf("\nInet Aton string with prefix length is %s", str)
	str, _ = ipv4Addr.ToInetAtonJoinedString(ipaddr.Inet_aton_radix_hex, 1)
	fmt.Printf("\nInet Aton string with prefix length is %s", str)
	str, _ = ipv4Addr.ToInetAtonJoinedString(ipaddr.Inet_aton_radix_hex, 0)
	fmt.Printf("\nInet Aton string with prefix length is %s", str)

	addrStrPref7 := ipaddr.NewIPAddressString("1:2:3:4::/64")
	ipv6Sect := addrStrPref7.GetAddress().ToIPv6().GetSection()
	str = ipv6Sect.ToCanonicalString()
	fmt.Printf("\nIPv6 string with prefix length is %s", str)
	str, _ = addrStrPref7.GetAddress().ToIPv6().ToMixedString()
	fmt.Printf("\nIPv6 mixed string with prefix length is %s", str)
	str, _ = addrStrPref7.GetAddress().ToBinaryString(true)
	fmt.Printf("\nIPv6 binary string is %s", str)

	str = addrStrPref7.GetAddress().ToSegmentedBinaryString()
	fmt.Printf("\nIPv6 segmented binary string is %s", str)

	addrStrPref8 := ipaddr.NewIPAddressString("1::4:5:6:7:8fff/64")
	ipv6Sect = addrStrPref8.GetAddress().ToIPv6().GetSection()
	str = ipv6Sect.ToCanonicalString()
	fmt.Printf("\nIPv6 string with prefix length is %s", str)
	str, _ = addrStrPref8.GetAddress().ToIPv6().ToMixedString()
	fmt.Printf("\nIPv6 mixed string with prefix length is %s", str)

	rangiter := rng.PrefixIterator(28)
	fmt.Printf("\nsequential range pref iterator:\n")
	for rangiter.HasNext() {
		fmt.Printf("%v ", rangiter.Next())
	}

	addrStrIPv6Pref1 := ipaddr.NewIPAddressString("1:2:3:4::")
	addrStrIPv6Pref2 := ipaddr.NewIPAddressString("1:2:4:1::")
	rng2 := addrStrIPv6Pref1.GetAddress().ToIPv6().SpanWithRange(addrStrIPv6Pref2.GetAddress().ToIPv6())
	rangeres := rng.Join(rng)
	fmt.Printf("\n\njoined ranges: %v\n", rangeres)
	rangeres2 := rng.ToIP().Join(rng2.ToIP())
	fmt.Printf("\n\njoined ranges: %v\n", rangeres2)
	rangeres3 := rng2.Join(rng2)
	fmt.Printf("\n\njoined ranges: %v\n", rangeres3)
	rangeres4 := rng2.ToIP().Join(rng.ToIP())
	fmt.Printf("\n\njoined ranges: %v\n", rangeres4)

	addrStrPref3 := ipaddr.NewIPAddressString("1-4::1/125")
	addrIter := addrStrPref3.GetAddress().PrefixBlockIterator()
	fmt.Printf("\naddress pref block iterator:\n")
	for addrIter.HasNext() {
		fmt.Printf("%v ", addrIter.Next())
	}

	addrStrPref4 := ipaddr.NewIPAddressString("1::1/125")
	addrIter = addrStrPref4.GetAddress().Iterator()
	fmt.Printf("\naddress iterator:\n")
	for addrIter.HasNext() {
		fmt.Printf("%v ", addrIter.Next())
	}

	addrStrPref5 := ipaddr.NewIPAddressString("1::/125")
	addrIter = addrStrPref5.GetAddress().Iterator()
	fmt.Printf("\naddress iterator:\n")
	for addrIter.HasNext() {
		fmt.Printf("%v ", addrIter.Next())
	}

	macStrPref1 := ipaddr.NewMACAddressString("1:2:3:4:5:6")
	mAddr := macStrPref1.GetAddress()
	fmt.Printf("\nmac addr is %+v\n", mAddr)

	macStrPref1 = ipaddr.NewMACAddressString("1:2:3:4:5:*")
	mAddr = macStrPref1.GetAddress()
	fmt.Printf("\nmac addr is %+v\n", mAddr)
	mAddrIter := mAddr.Iterator()
	fmt.Printf("\nmac address iterator:\n")
	for mAddrIter.HasNext() {
		fmt.Printf("%v ", mAddrIter.Next())
	}

	fmt.Printf("\nincremented by 1 mac addr %+v is %+v\n", mAddr, mAddr.Increment(1))
	fmt.Printf("\nincremented by -1 mac addr %+v is %+v\n", mAddr, mAddr.Increment(-1))
	fmt.Printf("\nincremented by -1 and then by +1 mac addr %+v is %+v\n", mAddr, mAddr.Increment(-1).Increment(1))
	fmt.Printf("\nincremented by +1 and then by -1 mac addr %+v is %+v\n", mAddr, mAddr.Increment(1).Increment(-1))

	splitIntoBlocks("0.0.0.0", "0.0.0.254")
	splitIntoBlocks("0.0.0.1", "0.0.0.254")
	splitIntoBlocks("0.0.0.0", "0.0.0.254") // 16 8 4 2 1
	splitIntoBlocks("0.0.0.10", "0.0.0.21")

	splitIntoBlocks("1.2.3.4", "1.2.3.3-5")
	splitIntoBlocks("1.2-3.4.5-6", "2.0.0.0")
	splitIntoBlocks("1.2.3.4", "1.2.4.4") // 16 8 4 2 1
	splitIntoBlocks("0.0.0.0", "255.0.0.0")

	fmt.Printf("\n\n")

	splitIntoBlocksSeq("0.0.0.0", "0.0.0.254")
	splitIntoBlocksSeq("0.0.0.1", "0.0.0.254")
	splitIntoBlocksSeq("0.0.0.0", "0.0.0.254") // 16 8 4 2 1
	splitIntoBlocksSeq("0.0.0.10", "0.0.0.21")

	splitIntoBlocksSeq("1.2.3.4", "1.2.3.3-5")
	splitIntoBlocksSeq("1.2-3.4.5-6", "2.0.0.0")
	splitIntoBlocksSeq("1.2-3.4.5-6", "1.3.4.6")
	splitIntoBlocksSeq("1.2.3.4", "1.2.4.4") // 16 8 4 2 1
	splitIntoBlocksSeq("0.0.0.0", "255.0.0.0")

	ipZero := &ipaddr.IPAddress{}
	ipZeroAgain := &ipaddr.IPAddress{}
	merged := ipZero.MergeToPrefixBlocks(ipZeroAgain, ipZero)
	//mergedOld := ipZero.MergeToPrefixBlocksOld(ipZeroAgain, ipZero)
	//fmt.Printf("new %v len %d\nold %v len %d", merged, len(merged), mergedOld, len(mergedOld))
	fmt.Printf("new %v len %d\n", merged, len(merged))
	merged = ipZero.MergeToPrefixBlocks(ipZeroAgain, ipZero, addrStrIPv6Pref1.GetAddress().ToIP())
	fmt.Printf("new %v len %d\n", merged, len(merged))

	fmt.Printf("%v\n\n", merge("209.152.214.112/30", "209.152.214.116/31", "209.152.214.118/31"))
	fmt.Printf("%v\n\n", merge("209.152.214.112/30", "209.152.214.116/32", "209.152.214.118/31"))
	fmt.Printf("%v\n\n", merge("1:2:3:4:8000::/65", "1:2:3:4::/66", "1:2:3:4:4000::/66", "1:2:3:5:4000::/66", "1:2:3:5::/66", "1:2:3:5:8000::/65"))

	delim := "1:2,3,4:3:6:4:5,6fff,7,8,99:6:8"
	delims := ipaddr.DelimitedAddressString(delim).ParseDelimitedSegments()
	delimCount := ipaddr.DelimitedAddressString(delim).CountDelimitedAddresses()
	i = 0
	for delims.HasNext() {
		i++
		fmt.Printf("%d of %d is %v, from %v\n", i, delimCount, delims.Next(), delim)
	}
	fmt.Println()
	delim = "1:3:6:4:5,6fff,7,8,99:6:2,3,4:8"
	delims = ipaddr.DelimitedAddressString(delim).ParseDelimitedSegments()
	delimCount = ipaddr.DelimitedAddressString(delim).CountDelimitedAddresses()
	//delims = ipaddr.ParseDelimitedSegments(delim)
	//delimCount = ipaddr.CountDelimitedAddresses(delim)
	i = 0
	for delims.HasNext() {
		i++
		fmt.Printf("%d of %d is %v, from %v\n", i, delimCount, delims.Next(), delim)
	}
	//bitsPerSegment := 8
	//prefBits := 7
	//maxVal := ^ipaddr.DivInt(0)
	//mask := ^(maxVal << (bitsPerSegment - prefBits))
	//masker := ipaddr.TestMaskRange(0, 4, mask, maxVal)
	//fmt.Printf("masked vals 0 to 4 masked with %v (should be 0 to 1): %v %v\n", mask, masker.GetMaskedLower(0, mask), masker.GetMaskedUpper(4, mask))
	//
	//prefBits = 4
	//mask = ^(maxVal << (bitsPerSegment - prefBits))
	//masker = ipaddr.TestMaskRange(17, 32, mask, maxVal)
	//fmt.Printf("masked vals 17 to 32 masked with %v (should be 0 to 15): %v %v\n", mask, masker.GetMaskedLower(17, mask), masker.GetMaskedUpper(32, mask))
	//
	//masker = ipaddr.TestMaskRange(16, 32, mask, maxVal)
	//fmt.Printf("masked vals 16 to 32 masked with %v (should be 0 to 15): %v %v\n", mask, masker.GetMaskedLower(16, mask), masker.GetMaskedUpper(32, mask))

	// iterate on nil - just checking what happens.  it panics, not surprisingly.
	//var niladdr *ipaddr.IPAddress
	//itr := niladdr.Iterator()
	//for itr.hasNext() {
	//	fmt.Printf("%v ", itr.Next())
	//}

	s := ipaddr.IPv4AddressSegment{}
	res := s.PrefixContains(&s, 6)
	fmt.Printf("Zero seg pref contains %v\n", res)

	// check is we need to "override" methods like ToHexString
	str, _ = ipaddr.NewIPv4Segment(3).ToHexString(true)
	fmt.Println("leading zeros?  Hope not: " + str)
	str, _ = (&ipaddr.IPv4AddressSegment{}).ToHexString(true)
	fmt.Println("leading zeros?  Hope not: " + str)

	// check is we need to "override" methods like ToNormalizedString
	str = ipaddr.NewIPv4Segment(3).ToNormalizedString()
	fmt.Println("leading zeros?  Hope not: " + str)
	str = (&ipaddr.IPv4AddressSegment{}).ToNormalizedString()
	fmt.Println("leading zeros?  Hope not: " + str)

	sega := ipaddr.NewIPv4Segment(128)
	segb := ipaddr.NewIPv4Segment(127)
	seg1 := ipaddr.NewIPv4Segment(3)
	seg2 := ipaddr.NewIPv4Segment(0)
	seg3 := &ipaddr.IPv4AddressSegment{}

	fmt.Printf("compare values: 1? %v nil? %v nil? %v 0? %v 0? %v nil? %v 1? %v 6? %v 8? %v 8? %v\n",
		sega.GetBlockMaskPrefixLen(true),  // should be 1
		segb.GetBlockMaskPrefixLen(true),  // should be nil
		seg1.GetBlockMaskPrefixLen(true),  // should be nil
		seg2.GetBlockMaskPrefixLen(true),  // should be 0 - either 0 or nil
		seg3.GetBlockMaskPrefixLen(true),  // should be 0 - either 0 or nil
		sega.GetBlockMaskPrefixLen(false), // should be nil
		segb.GetBlockMaskPrefixLen(false), // should be 1
		seg1.GetBlockMaskPrefixLen(false), // should be 6
		seg2.GetBlockMaskPrefixLen(false), // should be 8 - either 8 or nil
		seg3.GetBlockMaskPrefixLen(false), // should be 8 - either 8 or nil
	)

	ToPrefixLen := func(i ipaddr.PrefixBitCount) ipaddr.PrefixLen {
		return &i
	}
	p1 := ToPrefixLen(1)
	p2 := ToPrefixLen(2)
	fmt.Printf("%v %v\n", p1, p2)
	*p1 = *p2
	fmt.Printf("%v %v\n", p1, p2)
	p1 = ToPrefixLen(1)
	p2 = ToPrefixLen(2)
	fmt.Printf("%v %v\n", p1, p2)

	ToPort := func(i ipaddr.PortNum) ipaddr.Port {
		return &i
	}
	pr1 := ToPort(3)
	pr2 := ToPort(4)
	fmt.Printf("%p %p %v %v\n", pr1, pr2, pr1, pr2)
	*pr1 = *pr2
	fmt.Printf("%p %p %v %v\n", pr1, pr2, pr1, pr2)
	pr1 = ToPort(3)
	pr2 = ToPort(4)
	fmt.Printf("%v %v\n", pr1, pr2)

	fmt.Printf("\n\n")
	// _ = getDoc()

	bn := NewAddressTrieNode()
	_ = bn

	addrStr = ipaddr.NewIPAddressString("1.2.0.0/32")
	pAddr = addrStr.GetAddress()
	fmt.Printf("bit count pref len is pref block: %t\n", pAddr.IsPrefixBlock())

	trie := NewIPv4AddressTrie()
	addrStr = ipaddr.NewIPAddressString("1.2.0.0/16")
	trie.Add(pAddr.ToIPv4())
	addrStr = ipaddr.NewIPAddressString("1.2.3.4")
	pAddr = addrStr.GetAddress()
	fmt.Printf("no pref len is pref block: %t\n", pAddr.IsPrefixBlock())
	trie.Add(pAddr.ToIPv4())
	str = trie.String()
	fmt.Printf("%s", str)
	fmt.Printf("trie default: %v", trie)
	fmt.Printf("decimal: %d\n", trie)
	fmt.Printf("hex: %#x\n", trie)
	fmt.Printf("node default: %v\n", *trie.GetRoot())
	fmt.Printf("node decimal: %d\n", *trie.GetRoot())
	fmt.Printf("node hex: %#x\n", *trie.GetRoot())

	trie2 := ipaddr.IPv4AddressTrie{}
	fmt.Println(ipaddr.TreesString[*ipaddr.IPv4Address](true, &trie, &trie2, &trie))
	fmt.Println("zero trie\n", trie2)
	var ptraddr *ipaddr.IPv4Address
	fmt.Printf("nil addr %s\n", ptraddr)
	var trie3 *ipaddr.IPv4AddressTrie
	fmt.Printf("nil trie %s\n", trie3)
	fmt.Println("nil trie\n", trie3)
	fmt.Println(ipaddr.TreesString(true, &trie, &trie2, &trie, trie3, &trie))
	trie = ipaddr.IPv4AddressTrie{}
	fmt.Printf("%v %d %d %t %t",
		trie,
		trie.Size(),
		trie.NodeSize(),
		trie.BlockSizeAllNodeIterator(true).HasNext(),
		trie.ContainedFirstAllNodeIterator(true).HasNext())
	//fmt.Printf("%v %d %d %t %v",
	//	trie,
	//	trie.Size(),
	//	trie.NodeSize(),
	//	trie.BlockSizeAllNodeIterator(true).hasNext(),
	//	trie.BlockSizeAllNodeIterator(true).Next())
	fmt.Printf("%v %d %d %v %v",
		trie,
		trie.Size(),
		trie.NodeSize(),
		trie.BlockSizeAllNodeIterator(true).Next(),
		trie.ContainedFirstAllNodeIterator(true).Next())

	testers := []string{
		"1.2.3.4",
		"1.2.*.*",
		"1.2.*.0/24",
		"1.2.*.4",
		"1.2.0-1.*",
		"1.2.1-2.*",
		"1.2.252-255.*",
		"1.2.3.4/16",
	}
	fmt.Println()
	fmt.Println()
	fmt.Println("AssignPrefixForSingleBlock")
	for _, t := range testers {
		addr := ipaddr.NewIPAddressString(t).GetAddress()
		fmt.Printf("%s\n", addr.AssignPrefixForSingleBlock())
	}
	fmt.Println()
	fmt.Println("AssignMinPrefixForBlock")
	for _, t := range testers {
		addr := ipaddr.NewIPAddressString(t).GetAddress()
		fmt.Printf("%s\n", addr.AssignMinPrefixForBlock())
	}

	p4 := ToPrefixLen(4)
	segp := ipaddr.NewIPv4PrefixedSegment(1, p4)
	segp2 := ipaddr.NewIPv4Segment(2)
	p12 := ToPrefixLen(12)
	newSec := ipaddr.NewIPv4PrefixedSection([]*ipaddr.IPv4AddressSegment{segp, segp2, segp2, segp2}, p12)
	fmt.Println("the section is", newSec) // should be 1.2.2.2/4
	sgs := newSec.GetSegments()
	fmt.Println("the segs are", sgs)
	sg := sgs[0]
	fmt.Println("the first seg is", sg, "with prefix", sg.GetSegmentPrefixLen())
	sg = sgs[1]
	fmt.Println("the second seg is", sg, "with prefix", sg.GetSegmentPrefixLen())
	sg = sgs[2]
	fmt.Println("the third seg is", sg, "with prefix", sg.GetSegmentPrefixLen())

	newSec = ipaddr.NewIPv4PrefixedSection([]*ipaddr.IPv4AddressSegment{segp2, segp2, segp2, segp2}, p12)
	fmt.Println("the section is", newSec) // should be 1.2.2.2/12
	sgs = newSec.GetSegments()
	fmt.Println("the segs are", sgs)
	sg = sgs[0]
	fmt.Println("the first seg is", sg, "with prefix", sg.GetSegmentPrefixLen())
	sg = sgs[1]
	fmt.Println("the second seg is", sg, "with prefix", sg.GetSegmentPrefixLen())
	sg = sgs[2]
	fmt.Println("the third seg is", sg, "with prefix", sg.GetSegmentPrefixLen())

	fmt.Printf("decimal IPv4 address: %d\n", pAddr)
	fmt.Printf("decimal IPv4 address: %d\n", ipaddr.NewIPAddressString("255.255.255.255").GetAddress())
	fmt.Printf("decimal IPv6 address: %d\n", ipv6Addr)
	fmt.Printf("decimal IPv6 address: %d\n", ipaddr.NewIPAddressString("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff").GetAddress())

	allocator := ipaddr.IPPrefixBlockAllocator{}
	fmt.Println(allocator)
	allocator.AddAvailable(ipaddr.NewIPAddressString("192.168.10.0/24").GetAddress())
	fmt.Println(allocator)
	allocator.SetReserved(2)
	blocks := allocator.AllocateSizes(50, 30, 20, 2, 2, 2)
	fmt.Println("allocated blocks are:", blocks)
	fmt.Println(allocator)

	allocator = ipaddr.IPPrefixBlockAllocator{}
	fmt.Println(allocator)
	allocator.AddAvailable(ipaddr.NewIPAddressString("192.168.10.0/24").GetAddress())
	fmt.Println(allocator)
	allocator.SetReserved(2)
	blocks = allocator.AllocateSizes(60, 12, 12, 28)
	fmt.Println("allocated blocks are:", blocks)
	fmt.Println(allocator)

	//ipaddr.NewIPAddressString("1.2.3.16/28")
	//almostBlockStr := "1.2.3.17-31/28" xxxx
	// when switching from /30 to /28, we get  1.2.3.0/28 1.2.3.16/28 1.2.3.32/28 1.2.3.48/28
	almostBlockStr := "1.2.3.16-48/28"
	fmt.Println("Splitting " + almostBlockStr + " to range then back to iterator")
	almostBlock := ipaddr.NewIPAddressString(almostBlockStr).GetAddress()
	almostBlockRng := almostBlock.ToSequentialRange()
	fmt.Println("Range is " + almostBlockRng.String())
	fmt.Println("Range lower is " + almostBlockRng.GetLower().ToSegmentedBinaryString())
	fmt.Println("Range upper is " + almostBlockRng.GetUpper().ToSegmentedBinaryString())
	almostBlockIterRng := almostBlockRng.PrefixBlockIterator(almostBlockRng.GetMinPrefixLenForBlock())
	for almostBlockIterRng.HasNext() {
		fmt.Println(almostBlockIterRng.Next())
	}

	// Same as above, but instead of starting from "1.2.3.16-48/28", starts from "1.2.3.0/26"
	fmt.Println("and again")
	block := ipaddr.NewIPAddressString("1.2.3.0/26").GetAddress().SetPrefixLen(28)
	fmt.Println("count of prefixes of " + block.String() + " is " + block.GetPrefixCount().String())
	almostBlockIter := block.PrefixBlockIterator()
	almostBlockIter.Next()
	almostBlockRng = almostBlockIter.Next().GetLower().SpanWithRange(block.GetUpper())
	fmt.Println("Range is " + almostBlockRng.String())
	almostBlockIterRng = almostBlockRng.PrefixBlockIterator(almostBlockRng.GetMinPrefixLenForBlock())
	for almostBlockIterRng.HasNext() {
		fmt.Println(almostBlockIterRng.Next())
	}
	// the above shows we can take an iterator, the 1.2.3.16/28 prefix block iterator, peel off the first, then convert to sequential range,
	// then from the sequential range recover that iterator using almostBlockRng.PrefixBlockIterator(almostBlockRng.GetMinPrefixLenForBlock())

	// but we can get large blocks instead, by spanning again:
	fmt.Println(almostBlockRng.SpanWithPrefixBlocks())

	// Let's try this with IPv6
	originalBlock := ipaddr.NewIPAddressString("::/64").GetAddress()
	shrinkIt := originalBlock.SetPrefixLen(126)
	shrinkIter := shrinkIt.PrefixBlockIterator()
	shrinkIter.Next()
	low := shrinkIter.Next().GetLower()
	up := originalBlock.GetUpper()
	shrunkRange := low.SpanWithRange(up)
	fmt.Println("low " + low.String() + " to " + up.String() + " size " + shrunkRange.GetCount().String())
	fmt.Println(shrunkRange.SpanWithPrefixBlocks())

	//fmt.Println(almostBlockRng.SpanWithSequentialBlocks())

	alloc := ipaddr.IPPrefixBlockAllocator{}
	fmt.Println(alloc)
	alloc.AddAvailable(ipaddr.NewIPAddressString("192.168.10.0/24").GetAddress())
	fmt.Println(alloc)
	alloc.SetReserved(2)
	blocks = alloc.AllocateSizes(50, 30, 20, 2, 2, 2)
	fmt.Println("allocated blocks are:", blocks)
	fmt.Println(alloc)

	// put em back and see what happens
	for _, allocated := range blocks {
		alloc.AddAvailable(allocated.GetAddress())
		//fmt.Println(alloc)
	}
	fmt.Println(alloc)

	myaddr := ipaddr.IPAddress{}
	addr1Lower := myaddr.GetLower()
	fmt.Println("one is " + addr1Lower.String())
	//fmt.Println("one to bytes is ", addr1Lower.Bytes())
	naddr := addr1Lower.GetNetIPAddr()
	fmt.Println("one to ipaddr is " + naddr.String())
	faddr, _ := ipaddr.NewIPAddressFromNetIPAddr(naddr)
	fmt.Println("and back is " + faddr.String())
	//log2()

	addedTree := ipaddr.AddedTree[*ipaddr.IPv4Address]{}
	fmt.Println("\nzero tree is " + addedTree.String())
	fmt.Println("root is " + addedTree.GetRoot().String())
	fmt.Println("root key is " + addedTree.GetRoot().GetKey().String())
	fmt.Println("root subnodes are ", addedTree.GetRoot().GetSubNodes())
	fmt.Println("root tree string is " + addedTree.GetRoot().TreeString())

	addedTreeNode := ipaddr.AddedTreeNode[*ipaddr.IPv4Address]{}
	fmt.Println("node is " + addedTreeNode.String())
	fmt.Println("node key is " + addedTreeNode.GetKey().String())
	fmt.Println("node subnodes are ", addedTreeNode.GetSubNodes())
	fmt.Println("node tree string is " + addedTreeNode.TreeString())

	assocAddedTree := ipaddr.AssociativeAddedTree[*ipaddr.IPv4Address, int]{}
	fmt.Println("\nassoc zero tree is " + assocAddedTree.String())
	fmt.Println("root is " + assocAddedTree.GetRoot().String())
	fmt.Println("root key is " + assocAddedTree.GetRoot().GetKey().String())
	fmt.Println("root value is ", assocAddedTree.GetRoot().GetValue())
	fmt.Println("root subnodes are ", assocAddedTree.GetRoot().GetSubNodes())
	fmt.Println("root tree string is " + assocAddedTree.GetRoot().TreeString())

	assocAddedTreeNode := ipaddr.AssociativeAddedTreeNode[*ipaddr.IPAddress, float64]{}
	fmt.Println("assoc node is " + assocAddedTreeNode.String())
	fmt.Println("assoc node key is " + assocAddedTreeNode.GetKey().String())
	fmt.Println("assoc node value is ", assocAddedTreeNode.GetValue())
	fmt.Println("assoc node subnodes are ", assocAddedTreeNode.GetSubNodes())
	fmt.Println("assoc node tree string is " + assocAddedTreeNode.TreeString())

	fmt.Println()
	zeros()
}

func splitIntoBlocks(one, two string) {
	blocks := split(one, two)
	fmt.Printf("%v from splitting %v and %v: %v\n", len(blocks), one, two, blocks)
}

func splitIntoBlocksSeq(one, two string) {
	blocks := splitSeq(one, two)
	fmt.Printf("%v from splitting %v and %v: %v\n", len(blocks), one, two, blocks)
}

func split(oneStr, twoStr string) []*ipaddr.IPv4Address {
	one := ipaddr.NewIPAddressString(oneStr)
	two := ipaddr.NewIPAddressString(twoStr)
	return one.GetAddress().ToIPv4().SpanWithPrefixBlocksTo(two.GetAddress().ToIPv4())
}

func splitSeq(oneStr, twoStr string) []*ipaddr.IPv4Address {
	one := ipaddr.NewIPAddressString(oneStr)
	two := ipaddr.NewIPAddressString(twoStr)
	return one.GetAddress().ToIPv4().SpanWithSequentialBlocksTo(two.GetAddress().ToIPv4())
}

/*
8 from splitting 0.0.0.0 and 0.0.0.254: [0.0.0.0/25, 0.0.0.128/26, 0.0.0.192/27, 0.0.0.224/28, 0.0.0.240/29, 0.0.0.248/30, 0.0.0.252/31, 0.0.0.254/32]
14 from splitting 0.0.0.1 and 0.0.0.254: [0.0.0.1/32, 0.0.0.2/31, 0.0.0.4/30, 0.0.0.8/29, 0.0.0.16/28, 0.0.0.32/27, 0.0.0.64/26, 0.0.0.128/26, 0.0.0.192/27, 0.0.0.224/28, 0.0.0.240/29, 0.0.0.248/30, 0.0.0.252/31, 0.0.0.254/32]
8 from splitting 0.0.0.0 and 0.0.0.254: [0.0.0.0/25, 0.0.0.128/26, 0.0.0.192/27, 0.0.0.224/28, 0.0.0.240/29, 0.0.0.248/30, 0.0.0.252/31, 0.0.0.254/32]
4 from splitting 0.0.0.10 and 0.0.0.21: [0.0.0.10/31, 0.0.0.12/30, 0.0.0.16/30, 0.0.0.20/31]
1 from splitting 1.2.3.4 and 1.2.3.3-5: [1.2.3.3-5]
4 from splitting 1.2-3.4.5-6 and 2.0.0.0: [1.2.4.5-255, 1.2.5-255.*, 1.3-255.*.*, 2.0.0.0]
2 from splitting 1.2.3.4 and 1.2.4.4: [1.2.3.4-255, 1.2.4.0-4]
2 from splitting 0.0.0.0 and 255.0.0.0: [0-254.*.*.*, 255.0.0.0]
*/

func merge(strs ...string) []*ipaddr.IPAddress {
	first := ipaddr.NewIPAddressString(strs[0]).GetAddress()
	var remaining = make([]*ipaddr.IPAddress, len(strs))
	for i := range strs {
		remaining[i] = ipaddr.NewIPAddressString(strs[i]).GetAddress()
	}
	return first.MergeToPrefixBlocks(remaining...)
}

//func foo(bytes []byte) {
//	fmt.Printf("%v\n", bytes)
//}
//func foo2(bytes net.IP) {
//	fmt.Printf("%v\n", bytes)
//}
//func foo3(bytes net.IPAddr) {
//	fmt.Printf("%v\n", bytes)
//}

// go install golang.org/x/tools/cmd/godoc
// cd /Users/scfoley@us.ibm.com/goworkspace/bin
// ./godoc -http=localhost:6060
// http://localhost:6060/pkg/github.com/seancfoley/ipaddress/ipaddress-go/ipaddr/

// src/golang.org/x/tools/godoc/static/ has the templates, specifically godoc.html

// godoc cheat sheet
//https://godoc.org/github.com/fluhus/godoc-tricks#Links

// gdb tips https://gist.github.com/danisfermi/17d6c0078a2fd4c6ee818c954d2de13c
func getDoc() error {
	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	pkgs, err := parser.ParseDir(
		fset,
		//"/Users/scfoley@us.ibm.com/goworkspace/src/github.com/seancfoley/ipaddress/ipaddress-go/ipaddr",
		"/Users/scfoley/go/src/github.com/seancfoley/ipaddress/ipaddress-go/ipaddr",
		func(f os.FileInfo) bool { return true },
		parser.ParseComments)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return err
		//panic(err)
	}
	for keystr, valuePkg := range pkgs {
		pkage := doc.New(valuePkg, keystr, 0)
		//pkage := doc.New(valuePkg, keystr, doc.AllMethods)
		//pkage := doc.New(valuePkg, keystr, doc.AllDecls)
		//fmt.Printf("\n%+v", pkage)
		// Print the AST.
		//		ast.Print(fset, pkage)

		for _, t := range pkage.Types {
			fmt.Printf("\n%s", t.Name)
			for _, m := range t.Methods {
				//fmt.Printf("bool %v", doc.AllMethods&doc.AllMethods != 0)
				//https: //golang.org/src/go/doc/doc.go
				//https://golang.org/src/go/doc/reader.go sortedTypes sortedFuncs show how they are filtered
				fmt.Printf("\n%+v", m)
			}
		}
	}
	return nil
}

var faillog2, failceillog2, faillogbx, faililogbx, failBitsFor, ilogbShift, total int

func log2() {

	bitsFor := func(x uint64, expected uint64) {
		total++
		fmt.Printf("trying %x, want %d\n", x, expected)
		res := math.Log2(float64(x))
		if uint64(res) != expected {
			faillog2++
		}
		fmt.Println("log2", res)
		res = math.Ceil(math.Log2(float64(x)))
		if uint64(res) != expected {
			failceillog2++
		}
		fmt.Println("ceil log2", res)
		//fmt.Println("logb", math.Logb(float64(x)))
		//fmt.Println("ilogb", math.Ilogb(float64(x)))
		res = math.Logb(float64(2*x - 1))
		if uint64(res) != expected {
			faillogbx++
		}
		fmt.Println("logb x * 2 - 1", res)
		resi := math.Ilogb(float64(2*x - 1))
		if uint64(resi) != expected {
			faililogbx++
		}
		fmt.Println("ilogb x * 2 - 1", resi)
		//fmt.Println("ceil logb x * 2 - 1", math.Ceil(math.Logb(float64(2*x-1))))

		//  https://janmr.com/blog/2010/09/computing-the-integer-binary-logarithm/
		// OR combo of that with ILogb

		//	subtract 1 then shift, then we have 1 which needs to add 1 to result
		//I think I need to add 1 back

		limit := uint(53)
		const mask = 0xfff0000000000000

		BitsFor := func(x uint64) (result int) {
			if ((x - 1) & mask) != 0 { // conversion to float64 will fail
				x = ((x - 1) >> limit) + 1
				result = int(limit)
			}
			result += math.Ilogb(float64((x << 1) - 1))
			return
		}
		resi = BitsFor(x)
		if uint64(resi) != expected {
			failBitsFor++
		}
		fmt.Println("BitsFor", resi)

		//maintissa is 52 bits I think
		//var extra int
		//y := (x - 1) >> 52
		//if y != 0 { // equivalent to x > (1 << 52) or (x - 1) & 0x fff0000000000000  != 0
		//	fmt.Println("in extra block")
		//	x = ((x - 1) >> 52) + 1
		//	extra += 52
		//}

		var extra int
		if ((x - 1) & mask) != 0 { // equivalent to x > (1 << 52) or (x - 1) & 0xfffffffffffff != 0
			//fmt.Println("in extra block")
			x = ((x - 1) >> limit) + 1
			extra += int(limit)
		}

		resi = extra + math.Ilogb(float64((x<<1)-1))
		if uint64(resi) != expected {
			ilogbShift++
		}
		//fmt.Println("ilogb with shift", result+math.Ilogb(float64(2*x-1)))
		fmt.Println("ilogb with shift", resi)

		fmt.Println()

	}

	// x bits holds 2 power x values, the largest being 2 power x - 1
	bitsFor(1, 0)
	bitsFor(2, 1)
	bitsFor(4, 2)
	bitsFor(5, 3)
	bitsFor(6, 3)
	bitsFor(7, 3)
	bitsFor(8, 3)
	bitsFor(9, 4)

	bitsFor(0x4, 2)
	bitsFor(0x5, 3)

	bitsFor(0x8, 3)
	bitsFor(0x9, 4)

	bitsFor(0x10, 4)
	bitsFor(0x10+1, 5)

	bitsFor(0x100, 8)
	bitsFor(0x100+1, 9)

	bitsFor(0x1000000000000, 48)
	bitsFor(0x1000000000000+1, 49)

	bitsFor(0x4000000000000, 50)
	bitsFor(0x4000000000000+1, 51)

	bitsFor(0x8000000000000-1, 51)
	bitsFor(0x8000000000000, 51)
	bitsFor(0x8000000000000+1, 52)

	bitsFor(0x10000000000000-1, 52)
	bitsFor(0x10000000000000, 52)
	bitsFor(0x10000000000000+1, 53)

	bitsFor(0x20000000000000-1, 53)
	bitsFor(0x20000000000000, 53)
	bitsFor(0x20000000000000+1, 54)

	bitsFor(0x40000000000000-1, 54)
	bitsFor(0x40000000000000, 54)
	bitsFor(0x40000000000000+1, 55)

	bitsFor(0x100000000000000, 56)
	bitsFor(0x100000000000000+1, 57)

	bitsFor(0x1000000000000000, 60)
	bitsFor(0x1000000000000000+1, 61)

	bitsFor(0x8000000000000000, 63)
	bitsFor(0x8000000000000000+1, 64)
	bitsFor(0x8000000000000000+2, 64)
	bitsFor(0x10000000000000000-1, 64)

	fmt.Printf("fail counts %d %d %d %d %d %d total:%d\n", faillog2, failceillog2, faillogbx, faililogbx, failBitsFor, ilogbShift, total)

	//fmt.Printf("%x\n", uint64(float64(0x10000000000000)))
	//fmt.Printf("%x\n", uint64(float64(0x10000000000001)))
	//fmt.Printf("10000000000000\n")
	//fmt.Printf("10000000000001\n\n")
	//fmt.Printf("%x\n", uint64(float64(0x100000000000000)))
	//fmt.Printf("%x\n", uint64(float64(0x100000000000001)))
	//fmt.Printf("100000000000000\n")
	//fmt.Printf("100000000000001\n\n")
	//fmt.Printf("%x\n", uint64(float64(0x1000000000000000)))
	//fmt.Printf("%x\n", uint64(float64(0x1000000000000001)))
	//fmt.Printf("1000000000000000\n")
	//fmt.Printf("1000000000000001\n\n")
	//fmt.Printf("%x\n", uint64(float64(0x800000000000000)))
	//fmt.Printf("%x\n", uint64(float64(0x800000000000001)))
	//fmt.Printf("800000000000000\n")
	//fmt.Printf("800000000000001\n\n")
	//fmt.Printf("%x\n", uint64(float64(0x400000000000000)))
	//fmt.Printf("%x\n", uint64(float64(0x400000000000001)))
	//fmt.Printf("400000000000000\n")
	//fmt.Printf("400000000000001\n\n")
	//fmt.Printf("%x\n", uint64(float64(0x200000000000000)))
	//fmt.Printf("%x\n", uint64(float64(0x200000000000001)))
	//fmt.Printf("200000000000000\n")
	//fmt.Printf("200000000000001\n\n")

	x := -1
	fmt.Println(uint64(x))
	fmt.Println(uint64(x - 1))

	var y uint64 = 0xffffffffffffffff
	var z uint = 2
	fmt.Println(y + uint64(z))
}

func NewIPv4AddressTrie() ipaddr.IPv4AddressTrie {
	return ipaddr.IPv4AddressTrie{}
}

func NewAddressTrieNode() ipaddr.TrieNode[*ipaddr.Address] {
	return ipaddr.TrieNode[*ipaddr.Address]{}
}

func zeros() {

	strip := func(s string) string {
		return strings.ReplaceAll(strings.ReplaceAll(s, "ipaddr.", ""),
			"github.com/seancfoley/ipaddress-go/", "")
	}

	typeName := func(i any) string {
		return strip(reflect.ValueOf(i).Elem().Type().Name())
	}

	interfaceTypeName := func(i any) string {
		return strip(reflect.TypeOf(i).String())
	}

	truncateIndent := func(s, indent string) string {
		if boundary := len(indent) - (len(s) >> 3); boundary >= 0 {
			return indent[:boundary] + "\t" // every 8 chars eliminates a tab
		}
		return ""
	}

	baseIndent := "\t\t\t"
	title := "Address item zero values"
	fmt.Printf("%s%sint\tbits\tcount\tstring\n", title, truncateIndent(title, baseIndent))
	vars := []ipaddr.AddressItem{
		&ipaddr.Address{}, &ipaddr.IPAddress{},
		&ipaddr.IPv4Address{}, &ipaddr.IPv6Address{}, &ipaddr.MACAddress{},

		&ipaddr.AddressSection{}, &ipaddr.IPAddressSection{},
		&ipaddr.IPv4AddressSection{}, &ipaddr.IPv6AddressSection{}, &ipaddr.MACAddressSection{},
		&ipaddr.EmbeddedIPv6AddressSection{},
		&ipaddr.AddressDivisionGrouping{}, &ipaddr.IPAddressLargeDivisionGrouping{},
		&ipaddr.IPv6v4MixedAddressGrouping{},

		&ipaddr.AddressSegment{}, &ipaddr.IPAddressSegment{},
		&ipaddr.IPv4AddressSegment{}, &ipaddr.IPv6AddressSegment{}, &ipaddr.MACAddressSegment{},
		&ipaddr.AddressDivision{}, &ipaddr.IPAddressLargeDivision{},

		&ipaddr.IPAddressSeqRange{}, &ipaddr.IPv4AddressSeqRange{}, &ipaddr.IPv6AddressSeqRange{},
	}
	for _, v := range vars {
		name := typeName(v) + "{}"
		indent := truncateIndent(name, baseIndent)
		fmt.Printf("%s%s%v\t%v\t%v\t\"%v\"\n", name, indent, v.GetValue(), v.GetBitCount(), v.GetCount(), v)
	}

	title = "Address item nil pointers"
	fmt.Printf("\n%s%scount\tstring\n", title, truncateIndent(title, baseIndent+"\t\t"))
	nilPtrItems := []ipaddr.AddressItem{
		(*ipaddr.Address)(nil), (*ipaddr.IPAddress)(nil),
		(*ipaddr.IPv4Address)(nil), (*ipaddr.IPv6Address)(nil), (*ipaddr.MACAddress)(nil),

		(*ipaddr.AddressSection)(nil), (*ipaddr.IPAddressSection)(nil),
		(*ipaddr.IPv4AddressSection)(nil), (*ipaddr.IPv6AddressSection)(nil), (*ipaddr.MACAddressSection)(nil),

		(*ipaddr.AddressSegment)(nil), (*ipaddr.IPAddressSegment)(nil),
		(*ipaddr.IPv4AddressSegment)(nil), (*ipaddr.IPv6AddressSegment)(nil), (*ipaddr.MACAddressSegment)(nil),

		(*ipaddr.IPAddressSeqRange)(nil), (*ipaddr.IPv4AddressSeqRange)(nil), (*ipaddr.IPv6AddressSeqRange)(nil),
	}
	for _, v := range nilPtrItems {
		name := "(" + interfaceTypeName(v) + ")(nil)"
		indent := truncateIndent(name, baseIndent+"\t\t")
		fmt.Printf("%s%s%v\t\"%v\"\n", name, indent, v.GetCount(), v)
	}

	title = "Address key zero values"
	fmt.Printf("\n%s%sstring\n", title, truncateIndent(title, baseIndent+"\t\t\t"))
	keys := []fmt.Stringer{
		&ipaddr.AddressKey{}, &ipaddr.IPAddressKey{},
		&ipaddr.IPv4AddressKey{}, &ipaddr.IPv6AddressKey{}, &ipaddr.MACAddressKey{},
		&ipaddr.IPAddressSeqRangeKey{}, &ipaddr.IPv4AddressSeqRangeKey{}, &ipaddr.IPv6AddressSeqRangeKey{},
	}
	for _, k := range keys {
		name := typeName(k) + "{}"
		indent := truncateIndent(name, baseIndent+"\t\t\t")
		fmt.Printf("%s%s\"%v\"\n", name, indent, k)
	}

	title = "Host id zero values"
	fmt.Printf("\n%s%sstring\n", title, truncateIndent(title, baseIndent+"\t\t\t"))
	hostids := []ipaddr.HostIdentifierString{
		&ipaddr.HostName{}, &ipaddr.IPAddressString{}, &ipaddr.MACAddressString{},
	}
	for _, k := range hostids {
		name := typeName(k) + "{}"
		indent := truncateIndent(name, baseIndent+"\t\t\t")
		fmt.Printf("%s%s\"%v\"\n", name, indent, k)
	}

	title = "Host id nil pointers"
	fmt.Printf("\n%s%sstring\n", title, truncateIndent(title, baseIndent+"\t\t\t"))
	nilPtrIds := []ipaddr.HostIdentifierString{
		(*ipaddr.HostName)(nil), (*ipaddr.IPAddressString)(nil), (*ipaddr.MACAddressString)(nil),
	}
	for _, v := range nilPtrIds {
		name := "(" + interfaceTypeName(v) + ")(nil)"
		indent := truncateIndent(name, baseIndent+"\t\t\t")
		fmt.Printf("%s%s\"%v\"\n", name, indent, v)
	}
}
