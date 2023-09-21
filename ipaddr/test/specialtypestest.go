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

package test

import (
	"math/big"
	"net"
	"strconv"

	"github.com/seancfoley/ipaddress-go/ipaddr"
	"github.com/seancfoley/ipaddress-go/ipaddr/addrstrparam"
)

type specialTypesTester struct {
	testBase
}

var (
	hostOptionsSpecial            = new(addrstrparam.HostNameParamsBuilder).AllowEmpty(true).GetIPAddressParamsBuilder().AllowEmpty(true).ParseEmptyStrAs(addrstrparam.LoopbackOption).SetRangeParams(addrstrparam.WildcardOnly).AllowAll(true).GetParentBuilder().ToParams()
	addressOptionsSpecial         = new(addrstrparam.IPAddressStringParamsBuilder).Set(hostOptionsSpecial.GetIPAddressParams()).AllowEmpty(true).ParseEmptyStrAs(addrstrparam.LoopbackOption).ToParams()
	macOptionsSpecial             = new(addrstrparam.MACAddressStringParamsBuilder).Set(macAddressOptions).AllowEmpty(true).SetRangeParams(addrstrparam.WildcardOnly).AllowAll(true).ToParams()
	emptyAddressOptions           = new(addrstrparam.HostNameParamsBuilder).Set(hostOptions).GetIPAddressParamsBuilder().AllowEmpty(true).ParseEmptyStrAs(addrstrparam.LoopbackOption).GetParentBuilder().ToParams()
	emptyAddressNoLoopbackOptions = new(addrstrparam.HostNameParamsBuilder).Set(emptyAddressOptions).GetIPAddressParamsBuilder().ParseEmptyStrAs(addrstrparam.NoAddressOption).GetParentBuilder().ToParams()
)

func (t specialTypesTester) run() {
	allSingleHex := "0x00000000-0xffffffff"
	allSingleOctal := "000000000000-037777777777"

	t.testIPv4Strings("*", true, "*.*.*.*", "*.*.*.*", "%.%.%.%", "000-255.000-255.000-255.000-255", "*.*.*.*.in-addr.arpa", allSingleHex, allSingleOctal)
	t.testIPv4Strings("***.***.***.***", true, "*.*.*.*", "*.*.*.*", "%.%.%.%", "000-255.000-255.000-255.000-255", "*.*.*.*.in-addr.arpa", allSingleHex, allSingleOctal)
	t.testIPv4Strings("*.*", false, "*.*.*.*", "*.*.*.*", "%.%.%.%", "000-255.000-255.000-255.000-255", "*.*.*.*.in-addr.arpa", allSingleHex, allSingleOctal)
	t.testIPv4Strings("*.*/16", false, "*.*.0.0/16", "*.*.*.*", "%.%.%.%", "000-255.000-255.000.000/16", "*.*.*.*.in-addr.arpa", allSingleHex, allSingleOctal)
	t.testIPv4Strings("*.*/16", true, "*.*.0.0/16", "*.*.*.*", "%.%.%.%", "000-255.000-255.000.000/16", "*.*.*.*.in-addr.arpa", allSingleHex, allSingleOctal)
	t.testIPv4Strings("*/16", true, "*.*.0.0/16", "*.*.*.*", "%.%.%.%", "000-255.000-255.000.000/16", "*.*.*.*.in-addr.arpa", allSingleHex, allSingleOctal)
	t.testIPv4Strings("*/255.255.0.0", false, "*.*.0.0/16", "*.*.*.*", "%.%.%.%", "000-255.000-255.000.000/16", "*.*.*.*.in-addr.arpa", allSingleHex, allSingleOctal)
	t.testIPv4Strings("*/255.255.0.0", true, "*.*.0.0/16", "*.*.*.*", "%.%.%.%", "000-255.000-255.000.000/16", "*.*.*.*.in-addr.arpa", allSingleHex, allSingleOctal)
	t.testIPv4Strings("", false, "127.0.0.1", "127.0.0.1", "127.0.0.1", "127.000.000.001", "1.0.0.127.in-addr.arpa", "0x7f000001", "017700000001")
	t.testIPv4Strings("", true, "127.0.0.1", "127.0.0.1", "127.0.0.1", "127.000.000.001", "1.0.0.127.in-addr.arpa", "0x7f000001", "017700000001")

	base85All := "00000000000000000000" + ipaddr.ExtendedDigitsRangeSeparatorStr + "=r54lj&NUUO~Hi%c2ym0"
	//base85AllPrefixed := base85All + "/16"
	//base85AllPrefixed64 := base85All + "/64"
	base8516 := "00000000000000000000" + ipaddr.ExtendedDigitsRangeSeparatorStr + "=q{+M|w0(OeO5^EGP660" + "/16"
	base8564 := "00000000000000000000" + ipaddr.ExtendedDigitsRangeSeparatorStr + "=r54lj&NUTUTif>jH#O0" + "/64"
	allSingleHexIPv6 := "0x00000000000000000000000000000000-0xffffffffffffffffffffffffffffffff"
	allSingleOctalIPv6 := "00000000000000000000000000000000000000000000-03777777777777777777777777777777777777777777"

	t.testIPv6Strings("*", true, "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "%:%:%:%:%:%:%:%", "0000-ffff:0000-ffff:0000-ffff:0000-ffff:0000-ffff:0000-ffff:0000-ffff:0000-ffff", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*.*.*.*", "*:*:*:*:*:*:*.*.*.*", "*:*:*:*:*:*:*.*.*.*", "*:*:*:*:*:*:*.*.*.*", "*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.ip6.arpa", "*-*-*-*-*-*-*-*.ipv6-literal.net", base85All, allSingleHexIPv6, allSingleOctalIPv6)
	t.testIPv6Strings("*:*", false, "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "%:%:%:%:%:%:%:%", "0000-ffff:0000-ffff:0000-ffff:0000-ffff:0000-ffff:0000-ffff:0000-ffff:0000-ffff", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*.*.*.*", "*:*:*:*:*:*:*.*.*.*", "*:*:*:*:*:*:*.*.*.*", "*:*:*:*:*:*:*.*.*.*", "*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.ip6.arpa", "*-*-*-*-*-*-*-*.ipv6-literal.net", base85All, allSingleHexIPv6, allSingleOctalIPv6)
	t.testIPv6Strings("*:*", true, "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "%:%:%:%:%:%:%:%", "0000-ffff:0000-ffff:0000-ffff:0000-ffff:0000-ffff:0000-ffff:0000-ffff:0000-ffff", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*.*.*.*", "*:*:*:*:*:*:*.*.*.*", "*:*:*:*:*:*:*.*.*.*", "*:*:*:*:*:*:*.*.*.*", "*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.ip6.arpa", "*-*-*-*-*-*-*-*.ipv6-literal.net", base85All, allSingleHexIPv6, allSingleOctalIPv6)

	t.testIPv6Strings("*/16", true,
		"*:0:0:0:0:0:0:0/16",
		"*:*:*:*:*:*:*:*",
		"*:*:*:*:*:*:*:*",
		"%:%:%:%:%:%:%:%",
		"0000-ffff:0000:0000:0000:0000:0000:0000:0000/16",
		"*::/16",
		"*::/16",
		"*::/16",
		"*:*:*:*:*:*:*:*",
		"*::0.0.0.0/16",
		"*::0.0.0.0/16",
		"*::/16",
		"*::/16",
		"*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.ip6.arpa",
		"*-0-0-0-0-0-0-0.ipv6-literal.net/16",
		base8516, allSingleHexIPv6, allSingleOctalIPv6)
	t.testIPv6Strings("*:*/16", false,
		"*:0:0:0:0:0:0:0/16", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "%:%:%:%:%:%:%:%", "0000-ffff:0000:0000:0000:0000:0000:0000:0000/16", "*::/16", "*::/16", "*::/16", "*:*:*:*:*:*:*:*", "*::0.0.0.0/16", "*::0.0.0.0/16", "*::/16", "*::/16", "*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.ip6.arpa", "*-0-0-0-0-0-0-0.ipv6-literal.net/16", base8516, allSingleHexIPv6, allSingleOctalIPv6)
	t.testIPv6Strings("*:*/16", true, "*:0:0:0:0:0:0:0/16", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "%:%:%:%:%:%:%:%", "0000-ffff:0000:0000:0000:0000:0000:0000:0000/16", "*::/16", "*::/16", "*::/16", "*:*:*:*:*:*:*:*", "*::0.0.0.0/16", "*::0.0.0.0/16", "*::/16", "*::/16", "*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.ip6.arpa", "*-0-0-0-0-0-0-0.ipv6-literal.net/16", base8516, allSingleHexIPv6, allSingleOctalIPv6)
	t.testIPv6Strings("*/64", false, "*:*:*:*:0:0:0:0/64", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "%:%:%:%:%:%:%:%", "0000-ffff:0000-ffff:0000-ffff:0000-ffff:0000:0000:0000:0000/64", "*:*:*:*::/64", "*:*:*:*::/64", "*:*:*:*::/64", "*:*:*:*:*:*:*:*", "*:*:*:*::0.0.0.0/64", "*:*:*:*::0.0.0.0/64", "*:*:*:*::/64", "*:*:*:*::/64", "*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.ip6.arpa", "*-*-*-*-0-0-0-0.ipv6-literal.net/64", base8564, allSingleHexIPv6, allSingleOctalIPv6)
	t.testIPv6Strings("*/64", true, "*:*:*:*:0:0:0:0/64", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "%:%:%:%:%:%:%:%", "0000-ffff:0000-ffff:0000-ffff:0000-ffff:0000:0000:0000:0000/64", "*:*:*:*::/64", "*:*:*:*::/64", "*:*:*:*::/64", "*:*:*:*:*:*:*:*", "*:*:*:*::0.0.0.0/64", "*:*:*:*::0.0.0.0/64", "*:*:*:*::/64", "*:*:*:*::/64", "*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.ip6.arpa", "*-*-*-*-0-0-0-0.ipv6-literal.net/64", base8564, allSingleHexIPv6, allSingleOctalIPv6)
	t.testIPv6Strings("*:*/64", false, "*:*:*:*:0:0:0:0/64", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "%:%:%:%:%:%:%:%", "0000-ffff:0000-ffff:0000-ffff:0000-ffff:0000:0000:0000:0000/64", "*:*:*:*::/64", "*:*:*:*::/64", "*:*:*:*::/64", "*:*:*:*:*:*:*:*", "*:*:*:*::0.0.0.0/64", "*:*:*:*::0.0.0.0/64", "*:*:*:*::/64", "*:*:*:*::/64", "*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.ip6.arpa", "*-*-*-*-0-0-0-0.ipv6-literal.net/64", base8564, allSingleHexIPv6, allSingleOctalIPv6)
	t.testIPv6Strings("*:*/64", true, "*:*:*:*:0:0:0:0/64", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "%:%:%:%:%:%:%:%", "0000-ffff:0000-ffff:0000-ffff:0000-ffff:0000:0000:0000:0000/64", "*:*:*:*::/64", "*:*:*:*::/64", "*:*:*:*::/64", "*:*:*:*:*:*:*:*", "*:*:*:*::0.0.0.0/64", "*:*:*:*::0.0.0.0/64", "*:*:*:*::/64", "*:*:*:*::/64", "*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.ip6.arpa", "*-*-*-*-0-0-0-0.ipv6-literal.net/64", base8564, allSingleHexIPv6, allSingleOctalIPv6)
	t.testIPv6Strings("*/ffff::", false, "*:0:0:0:0:0:0:0/16", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "%:%:%:%:%:%:%:%", "0000-ffff:0000:0000:0000:0000:0000:0000:0000/16", "*::/16", "*::/16", "*::/16", "*:*:*:*:*:*:*:*", "*::0.0.0.0/16", "*::0.0.0.0/16", "*::/16", "*::/16", "*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.ip6.arpa", "*-0-0-0-0-0-0-0.ipv6-literal.net/16", base8516, allSingleHexIPv6, allSingleOctalIPv6)
	t.testIPv6Strings("*/ffff::", true, "*:0:0:0:0:0:0:0/16", "*:*:*:*:*:*:*:*", "*:*:*:*:*:*:*:*", "%:%:%:%:%:%:%:%", "0000-ffff:0000:0000:0000:0000:0000:0000:0000/16", "*::/16", "*::/16", "*::/16", "*:*:*:*:*:*:*:*", "*::0.0.0.0/16", "*::0.0.0.0/16", "*::/16", "*::/16", "*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.*.ip6.arpa", "*-0-0-0-0-0-0-0.ipv6-literal.net/16", base8516, allSingleHexIPv6, allSingleOctalIPv6)

	t.testIPv6Strings("", true, "0:0:0:0:0:0:0:1", "0:0:0:0:0:0:0:1", "::1", "0:0:0:0:0:0:0:1", "0000:0000:0000:0000:0000:0000:0000:0001", "::1", "::1", "::1", "::1", "::0.0.0.1", "::0.0.0.1", "::0.0.0.1", "::0.0.0.1", "1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.ip6.arpa", "0-0-0-0-0-0-0-1.ipv6-literal.net", "00000000000000000001", "0x00000000000000000000000000000001", "00000000000000000000000000000000000000000001")

	nilStr := `<nil>`
	t.testBase.testIPv6Strings(nil, nil,
		nilStr,
		nilStr,
		nilStr,
		nilStr,
		nilStr,
		nilStr,
		nilStr,
		nilStr,
		nilStr,
		nilStr,
		nilStr,
		nilStr,
		nilStr,
		nilStr,
		nilStr,
		nilStr,
		nilStr,
		nilStr)

	t.testInvalidValues()

	t.testValidity()

	t.testEmptyValues()

	t.testAllValues()
	t.testAllValuesVersioned(ipaddr.IPv4, getCount(255, 4))
	t.testAllValuesVersioned(ipaddr.IPv6, getCount(0xffff, 8))
	t.testAllMACValues(getCount(0xff, 6), getCount(0xff, 8))

	addressEmpty := t.createParamsHost("", emptyAddressOptions)
	t.hostLabelsHostTest(addressEmpty, []string{"127", "0", "0", "1"})
	addressEmpty2 := t.createParamsHost("", emptyAddressNoLoopbackOptions)
	t.hostLabelsHostTest(addressEmpty2, []string{})
	hostEmpty := t.createParamsHost("", hostOptionsSpecial)
	t.hostLabelsHostTest(hostEmpty, []string{"127", "0", "0", "1"})

	t.testEmptyIsSelf()
	t.testSelf("localhost", true)
	t.testSelf("127.0.0.1", true)
	t.testSelf("::1", true)
	t.testSelf("[::1]", true)
	t.testSelf("*", false)
	t.testSelf("sean.com", false)
	t.testSelf("1.2.3.4", false)
	t.testSelf("::", false)
	t.testSelf("[::]", false)
	t.testSelf("[1:2:3:4:1:2:3:4]", false)
	t.testSelf("1:2:3:4:1:2:3:4", false)

	t.testEmptyLoopback()
	t.testLoopback("127.0.0.1", true)
	t.testLoopback("::1", true)
	t.testLoopback("*", false)
	t.testLoopback("1.2.3.4", false)
	t.testLoopback("::", false)
	t.testLoopback("1:2:3:4:1:2:3:4", false)

	t.testNils()
	t.testZeros()
}

func (t specialTypesTester) testIPv4Strings(addr string, explicit bool, normalizedString, normalizedWildcardString, sqlString, fullString, reverseDNSString, singleHex, singleOctal string) {
	w := t.createParamsAddress(addr, addressOptionsSpecial)
	var ipAddr *ipaddr.IPAddress
	if explicit {
		ipAddr = w.GetVersionedAddress(ipaddr.IPv4)
	} else {
		ipAddr = w.GetAddress()
	}
	t.testStrings(w, ipAddr, normalizedString, normalizedWildcardString, normalizedWildcardString, sqlString, fullString,
		normalizedString, normalizedString, normalizedWildcardString, normalizedString, normalizedWildcardString, reverseDNSString, normalizedString,
		singleHex, singleOctal)
}

func (t specialTypesTester) testIPv6Strings(addr string,
	explicit bool,
	normalizedString,
	normalizedWildcardString,
	canonicalWildcardString,
	sqlString,
	fullString,
	compressedString,
	canonicalString,
	subnetString,
	compressedWildcardString,
	mixedStringNoCompressMixed,
	mixedStringNoCompressHost,
	mixedStringCompressCoveredHost,
	mixedString,
	reverseDNSString,
	uncHostString,
	base85String,
	singleHex,
	singleOctal string) {
	w := t.createParamsAddress(addr, addressOptionsSpecial)
	var ipAddr *ipaddr.IPAddress
	if explicit {
		ipAddr = w.GetVersionedAddress(ipaddr.IPv6)
	} else {
		ipAddr = w.GetAddress()
	}
	t.testBase.testIPv6Strings(w,
		ipAddr,
		normalizedString,
		normalizedWildcardString,
		canonicalWildcardString,
		sqlString,
		fullString,
		compressedString,
		canonicalString,
		subnetString,
		compressedWildcardString,
		mixedStringNoCompressMixed,
		mixedStringNoCompressHost,
		mixedStringCompressCoveredHost,
		mixedString,
		reverseDNSString,
		uncHostString,
		base85String,
		singleHex,
		singleOctal)
}

func (t specialTypesTester) testEmptyValues() {
	t.testEmptyValuesOpts(hostOptionsSpecial, addressOptionsSpecial)

	zeroHostOptions := new(addrstrparam.HostNameParamsBuilder).GetIPAddressParamsBuilder().ParseEmptyStrAs(addrstrparam.LoopbackOption).GetParentBuilder().ToParams()
	zeroAddrOptions := new(addrstrparam.IPAddressStringParamsBuilder).ParseEmptyStrAs(addrstrparam.LoopbackOption).ToParams()
	t.testEmptyValuesOpts(zeroHostOptions, zeroAddrOptions)

	zeroHostOptions = new(addrstrparam.HostNameParamsBuilder).GetIPAddressParamsBuilder().ParseEmptyStrAs(addrstrparam.ZeroAddressOption).GetParentBuilder().ToParams()
	zeroAddrOptions = new(addrstrparam.IPAddressStringParamsBuilder).ParseEmptyStrAs(addrstrparam.ZeroAddressOption).ToParams()
	t.testEmptyValuesOpts(zeroHostOptions, zeroAddrOptions)

	zeroHostOptions = new(addrstrparam.HostNameParamsBuilder).GetIPAddressParamsBuilder().ParseEmptyStrAs(addrstrparam.NoAddressOption).GetParentBuilder().ToParams()
	zeroAddrOptions = new(addrstrparam.IPAddressStringParamsBuilder).ParseEmptyStrAs(addrstrparam.NoAddressOption).ToParams()
	t.testEmptyValuesOpts(zeroHostOptions, zeroAddrOptions)
}

func (t specialTypesTester) testEmptyValuesOpts(hp addrstrparam.HostNameParams, sp addrstrparam.IPAddressStringParams) {
	hostEmpty := t.createParamsHost("", hp)
	addressEmpty := t.createParamsAddress("", sp)

	// preferredVersion := new(ipaddr.IPAddressStringParamsBuilder).ToParams().GetPreferredVersion()
	preferredAddressVersion := ipaddr.IPVersion(sp.GetPreferredVersion())
	preferredHostVersion := ipaddr.IPVersion(hp.GetPreferredVersion())

	//var addr, addr2 net.IP
	var addr net.IP
	if preferredAddressVersion.IsIPv6() {
		if sp.EmptyStrParsedAs() == addrstrparam.LoopbackOption {
			addr = net.IPv6loopback
		} else if sp.EmptyStrParsedAs() == addrstrparam.ZeroAddressOption {
			addr = net.IPv6zero
		}
	} else {
		if sp.EmptyStrParsedAs() == addrstrparam.LoopbackOption {
			addr = net.IPv4(127, 0, 0, 1)
		} else if sp.EmptyStrParsedAs() == addrstrparam.ZeroAddressOption {
			addr = net.IPv4(0, 0, 0, 0)
		}
	}

	if preferredAddressVersion != preferredHostVersion {
		t.addFailure(newFailure("failure: precondition to test is that options have same preferred version", addressEmpty))
	}
	if addr == nil {
		// empty string not parsed as an address
		if addressEmpty.GetAddress() != nil {
			t.addFailure(newFailure("no match "+addressEmpty.GetAddress().String(), addressEmpty))
		}
		addr, err := addressEmpty.ToAddress()
		if addr != nil {
			t.addFailure(newFailure("no match "+addr.String(), addressEmpty))
		}
		if err != nil {
			t.addFailure(newFailure("no match "+err.Error(), addressEmpty))
		}
		if hostEmpty.AsAddress() != nil {
			t.addFailure(newHostFailure("host "+hostEmpty.String()+" treated as address "+hostEmpty.AsAddress().String(), hostEmpty))
			//t.addFailure(newHostFailure("no match "+hostEmpty.AsAddress().String(), hostEmpty))
		}
		return
	}
	address, _ := ipaddr.NewIPAddressFromNetIP(addr)

	if !addressEmpty.GetAddress().Equal(address) {
		t.addFailure(newFailure("no match "+addr.String(), addressEmpty))
	} else if addressEmpty.GetAddress().Compare(address) != 0 {
		t.addFailure(newFailure("no match "+addr.String(), addressEmpty))
	} else if addressEmpty.GetAddress().GetCount().Cmp(bigOneConst()) != 0 {
		t.addFailure(newFailure("no count match "+addr.String(), addressEmpty))
	} else {
		addressEmpty = hostEmpty.AsAddressString() //note that hostEmpty allows empty strings and they resolve to loopbacks, but they are not treated as addresses
		if addressEmpty == nil {
			t.addFailure(newFailure("host "+hostEmpty.String()+" treated as address "+addressEmpty.String(), addressEmpty))
		} else if !addressEmpty.GetAddress().Equal(address) {
			t.addFailure(newFailure("no match "+addressEmpty.GetAddress().String()+" with "+address.String(), addressEmpty))
		} else if addressEmpty.GetAddress().Compare(address) != 0 {
			t.addFailure(newFailure("no match "+addr.String(), addressEmpty))
		} else if addressEmpty.GetAddress().GetCount().Cmp(bigOneConst()) != 0 {
			t.addFailure(newFailure("no count match "+addr.String(), addressEmpty))
		} else {
			addressEmptyValue := hostEmpty.GetAddress()
			if !addressEmptyValue.Equal(address) {
				t.addFailure(newFailure("no match "+addr.String(), addressEmpty))
			} else if addressEmptyValue.Compare(address) != 0 {
				t.addFailure(newFailure("no match "+addr.String(), addressEmpty))
			} else if addressEmptyValue.GetCount().Cmp(bigOneConst()) != 0 {
				t.addFailure(newFailure("no count match "+addr.String(), addressEmpty))
			}
		}
	}
	t.incrementTestCount()
}

func (t specialTypesTester) testInvalidValues() {
	// invalid mask
	addressAll := t.createParamsAddress("*/f0ff::", addressOptionsSpecial)
	_, err := addressAll.ToAddress()
	if err == nil {
		t.addFailure(newFailure("unexpectedly valid", addressAll))
	} else {
		// valid mask
		addressAll = t.createParamsAddress("*/fff0::", addressOptionsSpecial)
		//try {
		if addressAll.GetAddress() == nil {
			t.addFailure(newFailure("unexpectedly invalid", addressAll))
		} else {
			//ambiguous
			addressAll = t.createParamsAddress("*", addressOptionsSpecial)
			if addressAll.GetAddress() != nil {
				t.addFailure(newFailure("unexpectedly invalid", addressAll))
			} else {
				//ambiguous
				addressAll = t.createParamsAddress("*/16", addressOptionsSpecial)
				if addressAll.GetAddress() != nil {
					t.addFailure(newFailure("unexpectedly invalid", addressAll))
				}
				//unambiguous similar addresses tested with testStrings()
			}
		}
	}
}

func (t specialTypesTester) testValidity() {
	hostEmpty := t.createHost("")
	hostAll := t.createHost("*")
	hostAllIPv4 := t.createHost("*.*.*.*")
	hostAllIPv6 := t.createHost("*:*:*:*:*:*:*:*")
	addressEmpty := t.createAddress("")
	addressAll := t.createAddress("*")
	macEmpty := t.createMACAddress("")
	macAll := t.createMACAddress("*")

	if hostEmpty.IsValid() {
		t.addFailure(newHostFailure("unexpectedly valid", hostEmpty))
	} else if hostAll.IsValid() {
		t.addFailure(newHostFailure("unexpectedly valid", hostAll))
	} else if hostAllIPv4.IsValid() {
		t.addFailure(newHostFailure("unexpectedly valid", hostAllIPv4))
	} else if hostAllIPv6.IsValid() {
		t.addFailure(newHostFailure("unexpectedly valid", hostAllIPv6))
	} else if addressEmpty.IsValid() {
		t.addFailure(newFailure("unexpectedly valid", addressEmpty))
	} else if addressAll.IsValid() {
		t.addFailure(newFailure("unexpectedly valid", addressAll))
	} else if macEmpty.IsValid() {
		t.addFailure(newMACFailure("unexpectedly valid", macEmpty))
	} else if macAll.IsValid() {
		t.addFailure(newMACFailure("unexpectedly valid", macAll))
	} else if hostAll.GetAddress() != nil {
		t.addFailure(newHostFailure("unexpectedly valid", hostAll))
	} else if hostEmpty.GetAddress() != nil {
		t.addFailure(newHostFailure("unexpectedly valid", hostEmpty))
	} else {
		hostEmpty = t.createParamsHost("", hostOptionsSpecial)
		hostAll = t.createParamsHost("*", hostOptionsSpecial)
		hostAllIPv4 = t.createParamsHost("*.*.*.*", hostOptionsSpecial)
		hostAllIPv6 = t.createParamsHost("*:*:*:*:*:*:*:*", hostOptionsSpecial)
		addressEmpty = t.createParamsAddress("", addressOptionsSpecial)
		addressAll = t.createParamsAddress("*", addressOptionsSpecial)
		macEmpty = t.createMACParamsAddress("", macOptionsSpecial)
		macAll = t.createMACParamsAddress("*", macOptionsSpecial)
		if !hostEmpty.IsValid() {
			t.addFailure(newHostFailure("unexpectedly invalid", hostEmpty))
		} else if !hostAll.IsValid() {
			t.addFailure(newHostFailure("unexpectedly invalid", hostAll))
		} else if !hostAllIPv4.IsValid() {
			t.addFailure(newHostFailure("unexpectedly invalid", hostAllIPv4))
		} else if !hostAllIPv6.IsValid() {
			t.addFailure(newHostFailure("unexpectedly invalid", hostAllIPv6))
		} else if !addressEmpty.IsValid() {
			t.addFailure(newFailure("unexpectedly invalid", addressEmpty))
		} else if !addressAll.IsValid() {
			t.addFailure(newFailure("unexpectedly invalid", addressAll))
		} else if !macEmpty.IsValid() {
			t.addFailure(newMACFailure("unexpectedly invalid", macEmpty))
		} else if !macAll.IsValid() {
			t.addFailure(newMACFailure("unexpectedly invalid", macAll))
		} else if hostEmpty.GetAddress() == nil { //loopback
			t.addFailure(newHostFailure("unexpectedly invalid", hostEmpty))
		} else if hostAll.GetAddress() != nil {
			t.addFailure(newHostFailure("unexpectedly invalid", hostAll))
		} else {
			//With empty strings, if we wish to allow them, there are two options,
			//we can either treat them as host names and we defer to the validation options for host names, as done above,
			//or we treat than as addresses and use the address options to control behaviour, as we do here.
			hostEmpty = t.createParamsHost("", emptyAddressOptions)
			if !hostEmpty.IsValid() {
				t.addFailure(newHostFailure("unexpectedly invalid", hostEmpty))
			} else if hostEmpty.GetAddress() == nil { //loopback
				t.addFailure(newHostFailure("unexpectedly invalid", hostEmpty))
			} else {
				addressAll = t.createParamsAddress("*.*/64", addressOptionsSpecial) // invalid prefix
				if addressAll.IsValid() {
					t.addFailure(newFailure("unexpectedly valid: "+addressAll.String(), addressAll))
				}
			}
		}
	}
	t.incrementTestCount()
}

func (t specialTypesTester) testAllMACValues(count1, count2 *big.Int) {
	macAll := t.createMACParamsAddress("*", macOptionsSpecial).GetAddress()
	macAll2 := t.createMACParamsAddress("*:*:*:*:*:*:*", macOptionsSpecial).GetAddress()
	address1Str := "*:*:*:*:*:*"
	address2Str := "*:*:*:*:*:*:*:*"
	mac1 := t.createMACParamsAddress(address1Str, macOptionsSpecial).GetAddress()
	mac2 := t.createMACParamsAddress(address2Str, macOptionsSpecial).GetAddress()
	if !macAll.Equal(mac1) {
		t.addFailure(newSegmentSeriesFailure("no match "+macAll.String(), mac1))
	} else if !macAll2.Equal(mac2) {
		t.addFailure(newSegmentSeriesFailure("no match "+macAll2.String(), mac2))
	} else if macAll.Compare(mac1) != 0 {
		t.addFailure(newSegmentSeriesFailure("no match "+macAll.String(), mac1))
	} else if macAll2.Compare(mac2) != 0 {
		t.addFailure(newSegmentSeriesFailure("no match "+macAll2.String(), mac2))
	} else if macAll.GetCount().Cmp(count1) != 0 {
		t.addFailure(newSegmentSeriesFailure("no count match ", macAll))
	} else if macAll2.GetCount().Cmp(count2) != 0 {
		t.addFailure(newSegmentSeriesFailure("no count match ", macAll2))
	}
	t.incrementTestCount()
}

func (t specialTypesTester) testAllValuesVersioned(version ipaddr.IPVersion, count *big.Int) {
	hostAll := t.createParamsHost("*", hostOptionsSpecial)
	addressAllStr := t.createParamsAddress("*", addressOptionsSpecial)
	addressAll := addressAllStr.GetVersionedAddress(version)
	var address2Str = "*.*.*.*"
	if !version.IsIPv4() {
		address2Str = "*:*:*:*:*:*:*:*"
	}
	address := t.createParamsAddress(address2Str, addressOptionsSpecial).GetAddress()
	if !addressAll.Equal(address) {
		t.addFailure(newIPAddrFailure("no match "+address.String(), addressAll))
	} else if addressAll.Compare(address) != 0 {
		t.addFailure(newIPAddrFailure("no match "+address.String(), addressAll))
	} else if addressAll.GetCount().Cmp(count) != 0 {
		t.addFailure(newIPAddrFailure("no count match ", addressAll))
	} else {
		str := hostAll.AsAddressString()
		addressAll = str.GetVersionedAddress(version)
		if !addressAll.Equal(address) {
			t.addFailure(newIPAddrFailure("no match "+address.String(), addressAll))
		} else if addressAll.Compare(address) != 0 {
			t.addFailure(newIPAddrFailure("no match "+address.String(), addressAll))
		} else if addressAll.GetCount().Cmp(count) != 0 {
			t.addFailure(newIPAddrFailure("no count match ", addressAll))
		}
	}
	t.incrementTestCount()
}

func (t specialTypesTester) testAllValues() {
	hostAll := t.createParamsHost("*", hostOptionsSpecial)
	addressAll := t.createParamsAddress("*", addressOptionsSpecial)
	macAll := t.createMACParamsAddress("*", macOptionsSpecial)
	if addressAll.GetAddress() != nil {
		t.addFailure(newFailure("non nil", addressAll))
	} else if hostAll.AsAddress() != nil {
		t.addFailure(newHostFailure("non nil", hostAll))
	} else if hostAll.GetAddress() != nil {
		t.addFailure(newHostFailure("non nil", hostAll))
	} else if macAll.GetAddress() == nil {
		t.addFailure(newMACFailure("nil", macAll))
	}
	t.incrementTestCount()
}

func (t specialTypesTester) testEmptyIsSelf() {
	w := t.createParamsHost("", hostOptionsSpecial)
	if !w.IsSelf() {
		t.addFailure(newHostFailure("failed: isSelf is "+strconv.FormatBool(w.IsSelf()), w))
	}
	w2 := t.createParamsHost("", emptyAddressOptions)
	if !w2.IsSelf() {
		t.addFailure(newHostFailure("failed: isSelf is "+strconv.FormatBool(w2.IsSelf()), w2))
	}
	t.incrementTestCount()
}

func (t specialTypesTester) testSelf(host string, isSelf bool) {
	w := t.createParamsHost(host, hostOptionsSpecial)
	if isSelf != w.IsSelf() {
		t.addFailure(newHostFailure("failed: isSelf is "+strconv.FormatBool(isSelf), w))
	}
	t.incrementTestCount()
}

func (t specialTypesTester) testEmptyLoopback() {
	w := t.createParamsHost("", hostOptionsSpecial)
	if !w.IsLoopback() {
		t.addFailure(newHostFailure("failed: isSelf is "+strconv.FormatBool(w.IsSelf()), w))
	}
	addressEmptyValue := w.GetAddress()
	if !addressEmptyValue.IsLoopback() {
		t.addFailure(newIPAddrFailure("failed: isSelf is "+strconv.FormatBool(addressEmptyValue.IsLoopback()), addressEmptyValue))
	}
	w2 := t.createParamsHost("", emptyAddressOptions)
	if !w2.IsLoopback() {
		t.addFailure(newHostFailure("failed: isSelf is "+strconv.FormatBool(w2.IsSelf()), w2))
	}
	t.incrementTestCount()
}

func (t specialTypesTester) testLoopback(host string, isSelf bool) {
	w := t.createParamsHost(host, hostOptionsSpecial)
	if isSelf != w.IsLoopback() {
		t.addFailure(newHostFailure("failed: isSelf is "+strconv.FormatBool(isSelf), w))
	}
	w2 := t.createParamsAddress(host, addressOptionsSpecial)
	if isSelf != w2.IsLoopback() {
		t.addFailure(newHostFailure("failed: isSelf is "+strconv.FormatBool(isSelf), w))
	}
	t.incrementTestCount()
}

func (t specialTypesTester) testZeros() {
	addrZero := ipaddr.Address{}
	ipZero := ipaddr.IPAddress{}
	macZero := ipaddr.MACAddress{}
	ipv4Zero := ipaddr.IPv4Address{}
	ipv6Zero := ipaddr.IPv6Address{}

	if addrZero.ToIP() == nil || addrZero.ToIPv4() != nil || addrZero.ToIPv6() != nil || addrZero.ToMAC() != nil {
		t.addFailure(newAddrFailure("zero of "+addrZero.String(), &addrZero))
	} else if !ipZero.ToAddressBase().Equal(&ipZero) || !ipZero.ToIP().Equal(&ipZero) || ipZero.ToIPv4() != nil || ipZero.ToIPv6() != nil {
		t.addFailure(newIPAddrFailure("zero of "+ipZero.String(), &ipZero))
	} else if !macZero.ToAddressBase().Equal(&macZero) {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero.String(), &ipv4Zero))
	} else if !ipv4Zero.ToAddressBase().Equal(&ipv4Zero) || !ipv4Zero.ToIP().Equal(&ipv4Zero) {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero.String(), &ipv4Zero))
	} else if !ipv6Zero.ToAddressBase().Equal(&ipv6Zero) || !ipv6Zero.ToIP().Equal(&ipv6Zero) {
		t.addFailure(newAddressItemFailure("zero of "+ipv6Zero.String(), &ipv6Zero))
	}

	addrZeroKey := ipaddr.Key[*ipaddr.Address]{}
	ipZeroKey := ipaddr.Key[*ipaddr.IPAddress]{}
	ipv4ZeroKey := ipaddr.IPv4AddressKey{}
	ipv6ZeroKey := ipaddr.IPv6AddressKey{}
	macZeroKey := ipaddr.MACAddressKey{}

	addrZero2 := addrZeroKey.ToAddress()
	ipZero2 := ipZeroKey.ToAddress()
	ipv4Zero2 := ipv4ZeroKey.ToAddress()
	ipv6Zero2 := ipv6ZeroKey.ToAddress()
	macZero2 := macZeroKey.ToAddress()

	//fmt.Println(addrZeroKey, ipZeroKey, ipv4ZeroKey, ipv6ZeroKey, macZeroKey)

	// check that zero values from addresses coming from zero value keys match zero values from addresses
	if !addrZero.Equal(addrZero2) || !addrZero2.Equal(&addrZero) || !addrZero2.Equal(addrZero2) || !addrZero.Equal(&addrZero) {
		t.addFailure(newAddrFailure("zero of "+addrZero.String(), addrZero2))
	} else if addrZero.Compare(addrZero2) != 0 || addrZero2.Compare(&addrZero) != 0 || addrZero.Compare(&addrZero) != 0 || addrZero2.Compare(addrZero2) != 0 {
		t.addFailure(newAddrFailure("zero of "+addrZero.String(), addrZero2))
	} else if !ipZero.Equal(ipZero2) || !ipZero2.Equal(&ipZero) || !ipZero2.Equal(ipZero2) || !ipZero.Equal(&ipZero) {
		t.addFailure(newIPAddrFailure("zero of "+ipZero.String(), ipZero2))
	} else if ipZero.Compare(ipZero2) != 0 || ipZero2.Compare(&ipZero) != 0 || ipZero.Compare(&ipZero) != 0 || ipZero2.Compare(ipZero2) != 0 {
		t.addFailure(newIPAddrFailure("zero of "+ipZero.String(), ipZero2))
	} else if !ipv4Zero.Equal(ipv4Zero2) || !ipv4Zero2.Equal(&ipv4Zero) || !ipv4Zero2.Equal(ipv4Zero2) || !ipv4Zero.Equal(&ipv4Zero) {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero.String(), ipv4Zero2))
	} else if ipv4Zero.Compare(ipv4Zero2) != 0 || ipv4Zero2.Compare(&ipv4Zero) != 0 || ipv4Zero.Compare(&ipv4Zero) != 0 || ipv4Zero2.Compare(ipv4Zero2) != 0 {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero.String(), ipv4Zero2))
	} else if !ipv6Zero.Equal(ipv6Zero2) || !ipv6Zero2.Equal(&ipv6Zero) || !ipv6Zero2.Equal(ipv6Zero2) || !ipv6Zero.Equal(&ipv6Zero) {
		t.addFailure(newAddressItemFailure("zero of "+ipv6Zero.String(), ipv6Zero2))
	} else if ipv6Zero.Compare(ipv6Zero2) != 0 || ipv6Zero2.Compare(&ipv6Zero) != 0 || ipv6Zero.Compare(&ipv6Zero) != 0 || ipv6Zero2.Compare(ipv6Zero2) != 0 {
		t.addFailure(newAddressItemFailure("zero of "+ipv6Zero.String(), ipv6Zero2))
	} else if !macZero.Equal(macZero2) || !macZero2.Equal(&macZero) || !macZero2.Equal(macZero2) || !macZero.Equal(&macZero) {
		t.addFailure(newAddressItemFailure("zero of "+macZero.String(), macZero2))
	} else if macZero.Compare(macZero2) != 0 || macZero2.Compare(&macZero) != 0 || macZero.Compare(&macZero) != 0 || macZero2.Compare(macZero2) != 0 {
		t.addFailure(newAddressItemFailure("zero of "+macZero.String(), macZero2))
	}

	// check that the various addresses from zero value keys are all different
	if !addrZero2.Equal(ipZero2) || addrZero2.Equal(ipv4Zero2) || addrZero2.Equal(ipv6Zero2) || addrZero2.Equal(macZero2) {
		t.addFailure(newAddrFailure("zero of "+addrZero2.String(), addrZero2))
	} else if addrZero2.Compare(ipZero2) != 0 || addrZero2.Compare(ipv4Zero2) == 0 || addrZero2.Compare(ipv6Zero2) == 0 || addrZero2.Compare(macZero2) == 0 {
		t.addFailure(newAddrFailure("zero of "+addrZero2.String(), addrZero2))
	} else if !ipZero2.Equal(addrZero2) || ipZero2.Equal(ipv4Zero2) || ipZero2.Equal(ipv6Zero2) || ipZero2.Equal(macZero2) {
		t.addFailure(newIPAddrFailure("zero of "+ipZero2.String(), ipZero2))
	} else if ipZero2.Compare(addrZero2) != 0 || ipZero2.Compare(ipv4Zero2) == 0 || ipZero2.Compare(ipv6Zero2) == 0 || ipZero2.Compare(macZero2) == 0 {
		t.addFailure(newIPAddrFailure("zero of "+ipZero2.String(), ipZero2))
	} else if ipv4Zero2.Equal(addrZero2) || ipv4Zero2.Equal(ipZero2) || ipv4Zero2.Equal(ipv6Zero2) || ipv4Zero2.Equal(macZero2) {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero2.String(), ipv4Zero2))
	} else if ipv4Zero2.Compare(addrZero2) == 0 || ipv4Zero2.Compare(ipZero2) == 0 || ipv4Zero2.Compare(ipv6Zero2) == 0 || ipv4Zero2.Compare(macZero2) == 0 {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero2.String(), ipv4Zero2))
	} else if ipv6Zero2.Equal(addrZero2) || ipv6Zero2.Equal(ipZero2) || ipv6Zero2.Equal(ipv4Zero2) || ipv6Zero2.Equal(macZero2) {
		t.addFailure(newAddressItemFailure("zero of "+ipv6Zero2.String(), ipv6Zero2))
	} else if ipv6Zero2.Compare(addrZero2) == 0 || ipv6Zero2.Compare(ipZero2) == 0 || ipv6Zero2.Compare(ipv4Zero2) == 0 || ipv6Zero2.Compare(macZero2) == 0 {
		t.addFailure(newAddressItemFailure("zero of "+ipv6Zero2.String(), ipv6Zero2))
	} else if macZero2.Equal(addrZero2) || macZero2.Equal(ipZero2) || macZero2.Equal(ipv4Zero2) || macZero2.Equal(ipv6Zero2) {
		t.addFailure(newAddressItemFailure("zero of "+macZero2.String(), macZero2))
	} else if macZero2.Compare(addrZero2) == 0 || macZero2.Compare(ipZero2) == 0 || macZero2.Compare(ipv4Zero2) == 0 || macZero2.Compare(ipv6Zero2) == 0 {
		t.addFailure(newAddressItemFailure("zero of "+macZero2.String(), macZero2))
	}

	// check that the zero addresses are all different
	if !addrZero.Equal(&ipZero) || addrZero.Equal(&ipv4Zero) || addrZero.Equal(&ipv6Zero) || addrZero.Equal(&macZero) {
		t.addFailure(newAddrFailure("zero of "+addrZero.String(), &addrZero))
	} else if addrZero.Compare(&ipZero) != 0 || addrZero.Compare(&ipv4Zero) == 0 || addrZero.Compare(&ipv6Zero) == 0 || addrZero.Compare(&macZero) == 0 {
		t.addFailure(newAddrFailure("zero of "+addrZero.String(), &addrZero))
	} else if !ipZero.Equal(&addrZero) || ipZero.Equal(&ipv4Zero) || ipZero.Equal(&ipv6Zero) || ipZero.Equal(&macZero) {
		t.addFailure(newIPAddrFailure("zero of "+ipZero.String(), &ipZero))
	} else if ipZero.Compare(&addrZero) != 0 || ipZero.Compare(&ipv4Zero) == 0 || ipZero.Compare(&ipv6Zero) == 0 || ipZero.Compare(&macZero) == 0 {
		t.addFailure(newIPAddrFailure("zero of "+ipZero.String(), &ipZero))
	} else if ipv4Zero.Equal(&addrZero) || ipv4Zero.Equal(&ipZero) || ipv4Zero.Equal(&ipv6Zero) || ipv4Zero.Equal(&macZero) {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero.String(), &ipv4Zero))
	} else if ipv4Zero.Compare(&addrZero) == 0 || ipv4Zero.Compare(&ipZero) == 0 || ipv4Zero.Compare(&ipv6Zero) == 0 || ipv4Zero.Compare(&macZero) == 0 {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero.String(), &ipv4Zero))
	} else if ipv6Zero.Equal(&addrZero) || ipv6Zero.Equal(&ipZero) || ipv6Zero.Equal(&ipv4Zero) || ipv6Zero.Equal(&macZero) {
		t.addFailure(newAddressItemFailure("zero of "+ipv6Zero.String(), &ipv6Zero))
	} else if ipv6Zero.Compare(&addrZero) == 0 || ipv6Zero.Compare(&ipZero) == 0 || ipv6Zero.Compare(&ipv4Zero) == 0 || ipv6Zero.Compare(&macZero) == 0 {
		t.addFailure(newAddressItemFailure("zero of "+ipv6Zero.String(), &ipv6Zero))
	} else if macZero.Equal(&addrZero) || macZero.Equal(&ipZero) || macZero.Equal(&ipv4Zero) || macZero.Equal(&ipv6Zero) {
		t.addFailure(newAddressItemFailure("zero of "+macZero.String(), &macZero))
	} else if macZero.Compare(&addrZero) == 0 || macZero.Compare(&ipZero) == 0 || macZero.Compare(&ipv4Zero) == 0 || macZero.Compare(&ipv6Zero) == 0 {
		t.addFailure(newAddressItemFailure("zero of "+macZero.String(), &macZero))
	}

	ipRangeZero := ipaddr.IPAddressSeqRange{}
	ipv4RangeZero := ipaddr.IPv4AddressSeqRange{}
	ipv6RangeZero := ipaddr.IPv6AddressSeqRange{}

	ipZero3 := ipRangeZero.GetLower()
	ipv4Zero3 := ipv4RangeZero.GetLower()
	ipv6Zero3 := ipv6RangeZero.GetLower()

	// check that zero values from address ranges match zero values from addresses
	if !ipZero.Equal(ipZero3) || !ipZero3.Equal(&ipZero) || !ipZero3.Equal(ipZero3) || !ipZero.Equal(&ipZero) {
		t.addFailure(newIPAddrFailure("zero of "+ipZero.String(), ipZero3))
	} else if ipZero.Compare(ipZero3) != 0 || ipZero3.Compare(&ipZero) != 0 || ipZero.Compare(&ipZero) != 0 || ipZero3.Compare(ipZero3) != 0 {
		t.addFailure(newIPAddrFailure("zero of "+ipZero.String(), ipZero3))
	} else if !ipv4Zero.Equal(ipv4Zero3) || !ipv4Zero3.Equal(&ipv4Zero) || !ipv4Zero3.Equal(ipv4Zero3) || !ipv4Zero.Equal(&ipv4Zero) {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero.String(), ipv4Zero3))
	} else if ipv4Zero.Compare(ipv4Zero3) != 0 || ipv4Zero3.Compare(&ipv4Zero) != 0 || ipv4Zero.Compare(&ipv4Zero) != 0 || ipv4Zero3.Compare(ipv4Zero3) != 0 {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero.String(), ipv4Zero3))
	} else if !ipv6Zero.Equal(ipv6Zero3) || !ipv6Zero3.Equal(&ipv6Zero) || !ipv6Zero3.Equal(ipv6Zero3) || !ipv6Zero.Equal(&ipv6Zero) {
		t.addFailure(newAddressItemFailure("zero of "+ipv6Zero.String(), ipv6Zero3))
	} else if ipv6Zero.Compare(ipv6Zero3) != 0 || ipv6Zero3.Compare(&ipv6Zero) != 0 || ipv6Zero.Compare(&ipv6Zero) != 0 || ipv6Zero3.Compare(ipv6Zero3) != 0 {
		t.addFailure(newAddressItemFailure("zero of "+ipv6Zero.String(), ipv6Zero3))
	}

	ipZero3 = ipRangeZero.GetUpper()
	ipv4Zero3 = ipv4RangeZero.GetUpper()
	ipv6Zero3 = ipv6RangeZero.GetUpper()

	// check that zero values from address ranges match zero values from addresses
	if !ipZero.Equal(ipZero3) || !ipZero3.Equal(&ipZero) || !ipZero3.Equal(ipZero3) || !ipZero.Equal(&ipZero) {
		t.addFailure(newIPAddrFailure("zero of "+ipZero.String(), ipZero3))
	} else if ipZero.Compare(ipZero3) != 0 || ipZero3.Compare(&ipZero) != 0 || ipZero.Compare(&ipZero) != 0 || ipZero3.Compare(ipZero3) != 0 {
		t.addFailure(newIPAddrFailure("zero of "+ipZero.String(), ipZero3))
	} else if !ipv4Zero.Equal(ipv4Zero3) || !ipv4Zero3.Equal(&ipv4Zero) || !ipv4Zero3.Equal(ipv4Zero3) || !ipv4Zero.Equal(&ipv4Zero) {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero.String(), ipv4Zero3))
	} else if ipv4Zero.Compare(ipv4Zero3) != 0 || ipv4Zero3.Compare(&ipv4Zero) != 0 || ipv4Zero.Compare(&ipv4Zero) != 0 || ipv4Zero3.Compare(ipv4Zero3) != 0 {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero.String(), ipv4Zero3))
	} else if !ipv6Zero.Equal(ipv6Zero3) || !ipv6Zero3.Equal(&ipv6Zero) || !ipv6Zero3.Equal(ipv6Zero3) || !ipv6Zero.Equal(&ipv6Zero) {
		t.addFailure(newAddressItemFailure("zero of "+ipv6Zero.String(), ipv6Zero3))
	} else if ipv6Zero.Compare(ipv6Zero3) != 0 || ipv6Zero3.Compare(&ipv6Zero) != 0 || ipv6Zero.Compare(&ipv6Zero) != 0 || ipv6Zero3.Compare(ipv6Zero3) != 0 {
		t.addFailure(newAddressItemFailure("zero of "+ipv6Zero.String(), ipv6Zero3))
	}

	ipZero3 = ipRangeZero.ToKey().ToSeqRange().GetLower().ToKey().ToAddress()
	ipv4Zero3 = ipv4RangeZero.ToKey().ToSeqRange().GetLower().ToKey().ToAddress() // Goland parser mistakenly flags this, resolving ToKey incorrectly
	ipv6Zero3 = ipv6RangeZero.ToKey().ToSeqRange().GetLower().ToKey().ToAddress() // Goland parser mistakenly flags this, resolving ToKey incorrectly

	// check that zero values from address ranges match zero values from addresses
	if !ipZero.Equal(ipZero3) || !ipZero3.Equal(&ipZero) || !ipZero3.Equal(ipZero3) || !ipZero.Equal(&ipZero) {
		t.addFailure(newIPAddrFailure("zero of "+ipZero.String(), ipZero3))
	} else if ipZero.Compare(ipZero3) != 0 || ipZero3.Compare(&ipZero) != 0 || ipZero.Compare(&ipZero) != 0 || ipZero3.Compare(ipZero3) != 0 {
		t.addFailure(newIPAddrFailure("zero of "+ipZero.String(), ipZero3))
	} else if !ipv4Zero.Equal(ipv4Zero3) || !ipv4Zero3.Equal(&ipv4Zero) || !ipv4Zero3.Equal(ipv4Zero3) || !ipv4Zero.Equal(&ipv4Zero) {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero.String(), ipv4Zero3))
	} else if ipv4Zero.Compare(ipv4Zero3) != 0 || ipv4Zero3.Compare(&ipv4Zero) != 0 || ipv4Zero.Compare(&ipv4Zero) != 0 || ipv4Zero3.Compare(ipv4Zero3) != 0 {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero.String(), ipv4Zero3))
	} else if !ipv6Zero.Equal(ipv6Zero3) || !ipv6Zero3.Equal(&ipv6Zero) || !ipv6Zero3.Equal(ipv6Zero3) || !ipv6Zero.Equal(&ipv6Zero) {
		t.addFailure(newAddressItemFailure("zero of "+ipv6Zero.String(), ipv6Zero3))
	} else if ipv6Zero.Compare(ipv6Zero3) != 0 || ipv6Zero3.Compare(&ipv6Zero) != 0 || ipv6Zero.Compare(&ipv6Zero) != 0 || ipv6Zero3.Compare(ipv6Zero3) != 0 {
		t.addFailure(newAddressItemFailure("zero of "+ipv6Zero.String(), ipv6Zero3))
	}

	ipZero3 = ipRangeZero.ToKey().ToSeqRange().GetUpper().ToKey().ToAddress()
	ipv4Zero3 = ipv4RangeZero.ToKey().ToSeqRange().GetUpper().ToKey().ToAddress() // Goland parser mistakenly flags this, resolving ToKey incorrectly
	ipv6Zero3 = ipv6RangeZero.ToKey().ToSeqRange().GetUpper().ToKey().ToAddress() // Goland parser mistakenly flags this, resolving ToKey incorrectly

	// check that zero values from address ranges match zero values from addresses
	if !ipZero.Equal(ipZero3) || !ipZero3.Equal(&ipZero) || !ipZero3.Equal(ipZero3) || !ipZero.Equal(&ipZero) {
		t.addFailure(newIPAddrFailure("zero of "+ipZero.String(), ipZero3))
	} else if ipZero.Compare(ipZero3) != 0 || ipZero3.Compare(&ipZero) != 0 || ipZero.Compare(&ipZero) != 0 || ipZero3.Compare(ipZero3) != 0 {
		t.addFailure(newIPAddrFailure("zero of "+ipZero.String(), ipZero3))
	} else if !ipv4Zero.Equal(ipv4Zero3) || !ipv4Zero3.Equal(&ipv4Zero) || !ipv4Zero3.Equal(ipv4Zero3) || !ipv4Zero.Equal(&ipv4Zero) {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero.String(), ipv4Zero3))
	} else if ipv4Zero.Compare(ipv4Zero3) != 0 || ipv4Zero3.Compare(&ipv4Zero) != 0 || ipv4Zero.Compare(&ipv4Zero) != 0 || ipv4Zero3.Compare(ipv4Zero3) != 0 {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero.String(), ipv4Zero3))
	} else if !ipv6Zero.Equal(ipv6Zero3) || !ipv6Zero3.Equal(&ipv6Zero) || !ipv6Zero3.Equal(ipv6Zero3) || !ipv6Zero.Equal(&ipv6Zero) {
		t.addFailure(newAddressItemFailure("zero of "+ipv6Zero.String(), ipv6Zero3))
	} else if ipv6Zero.Compare(ipv6Zero3) != 0 || ipv6Zero3.Compare(&ipv6Zero) != 0 || ipv6Zero.Compare(&ipv6Zero) != 0 || ipv6Zero3.Compare(ipv6Zero3) != 0 {
		t.addFailure(newAddressItemFailure("zero of "+ipv6Zero.String(), ipv6Zero3))
	}

	ipZero3 = ipRangeZero.ToKey().ToSeqRange().GetLower()
	ipv4Zero3 = ipv4RangeZero.ToKey().ToSeqRange().GetLower()
	ipv6Zero3 = ipv6RangeZero.ToKey().ToSeqRange().GetLower()

	// check that zero values from address ranges match zero values from addresses
	if !ipZero.Equal(ipZero3) || !ipZero3.Equal(&ipZero) || !ipZero3.Equal(ipZero3) || !ipZero.Equal(&ipZero) {
		t.addFailure(newIPAddrFailure("zero of "+ipZero.String(), ipZero3))
	} else if ipZero.Compare(ipZero3) != 0 || ipZero3.Compare(&ipZero) != 0 || ipZero.Compare(&ipZero) != 0 || ipZero3.Compare(ipZero3) != 0 {
		t.addFailure(newIPAddrFailure("zero of "+ipZero.String(), ipZero3))
	} else if !ipv4Zero.Equal(ipv4Zero3) || !ipv4Zero3.Equal(&ipv4Zero) || !ipv4Zero3.Equal(ipv4Zero3) || !ipv4Zero.Equal(&ipv4Zero) {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero.String(), ipv4Zero3))
	} else if ipv4Zero.Compare(ipv4Zero3) != 0 || ipv4Zero3.Compare(&ipv4Zero) != 0 || ipv4Zero.Compare(&ipv4Zero) != 0 || ipv4Zero3.Compare(ipv4Zero3) != 0 {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero.String(), ipv4Zero3))
	} else if !ipv6Zero.Equal(ipv6Zero3) || !ipv6Zero3.Equal(&ipv6Zero) || !ipv6Zero3.Equal(ipv6Zero3) || !ipv6Zero.Equal(&ipv6Zero) {
		t.addFailure(newAddressItemFailure("zero of "+ipv6Zero.String(), ipv6Zero3))
	} else if ipv6Zero.Compare(ipv6Zero3) != 0 || ipv6Zero3.Compare(&ipv6Zero) != 0 || ipv6Zero.Compare(&ipv6Zero) != 0 || ipv6Zero3.Compare(ipv6Zero3) != 0 {
		t.addFailure(newAddressItemFailure("zero of "+ipv6Zero.String(), ipv6Zero3))
	}

	ipZero3 = ipRangeZero.ToKey().ToSeqRange().GetUpper()
	ipv4Zero3 = ipv4RangeZero.ToKey().ToSeqRange().GetUpper()
	ipv6Zero3 = ipv6RangeZero.ToKey().ToSeqRange().GetUpper()

	// check that zero values from address ranges match zero values from addresses
	if !ipZero.Equal(ipZero3) || !ipZero3.Equal(&ipZero) || !ipZero3.Equal(ipZero3) || !ipZero.Equal(&ipZero) {
		t.addFailure(newIPAddrFailure("zero of "+ipZero.String(), ipZero3))
	} else if ipZero.Compare(ipZero3) != 0 || ipZero3.Compare(&ipZero) != 0 || ipZero.Compare(&ipZero) != 0 || ipZero3.Compare(ipZero3) != 0 {
		t.addFailure(newIPAddrFailure("zero of "+ipZero.String(), ipZero3))
	} else if !ipv4Zero.Equal(ipv4Zero3) || !ipv4Zero3.Equal(&ipv4Zero) || !ipv4Zero3.Equal(ipv4Zero3) || !ipv4Zero.Equal(&ipv4Zero) {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero.String(), ipv4Zero3))
	} else if ipv4Zero.Compare(ipv4Zero3) != 0 || ipv4Zero3.Compare(&ipv4Zero) != 0 || ipv4Zero.Compare(&ipv4Zero) != 0 || ipv4Zero3.Compare(ipv4Zero3) != 0 {
		t.addFailure(newAddressItemFailure("zero of "+ipv4Zero.String(), ipv4Zero3))
	} else if !ipv6Zero.Equal(ipv6Zero3) || !ipv6Zero3.Equal(&ipv6Zero) || !ipv6Zero3.Equal(ipv6Zero3) || !ipv6Zero.Equal(&ipv6Zero) {
		t.addFailure(newAddressItemFailure("zero of "+ipv6Zero.String(), ipv6Zero3))
	} else if ipv6Zero.Compare(ipv6Zero3) != 0 || ipv6Zero3.Compare(&ipv6Zero) != 0 || ipv6Zero.Compare(&ipv6Zero) != 0 || ipv6Zero3.Compare(ipv6Zero3) != 0 {
		t.addFailure(newAddressItemFailure("zero of "+ipv6Zero.String(), ipv6Zero3))
	}

	ipRangeZeroKey := ipRangeZero.ToKey()
	ipv4ZeroRangeKey := ipv4RangeZero.ToIP().ToKey()
	ipv6ZeroRangeKey := ipv6RangeZero.ToIP().ToKey()

	if ipRangeZeroKey != ipRangeZeroKey || ipRangeZeroKey == ipv4ZeroRangeKey || ipRangeZeroKey == ipv6ZeroRangeKey {
		t.addFailure(newAddressItemFailure("zero of "+ipRangeZeroKey.String(), &ipRangeZero))
	}

	iprange := ipaddr.NewIPSeqRange(nil, &ipZero)
	ipv4range := ipaddr.NewIPv4SeqRange(nil, &ipv4Zero).ToIP()
	ipv6range := ipaddr.NewIPv6SeqRange(nil, &ipv6Zero).ToIP()

	if iprange.ToKey() != iprange.ToKey() || iprange.ToKey() == ipv4range.ToKey() || iprange.ToKey() == ipv6range.ToKey() {
		t.addFailure(newAddressItemFailure("range from nil "+iprange.String(), iprange))
	}

	iprange = ipaddr.NewSequentialRange(nil, &ipZero)
	ipv4range = ipaddr.NewSequentialRange(nil, &ipv4Zero).ToIP()
	ipv6range = ipaddr.NewSequentialRange(nil, &ipv6Zero).ToIP()

	if iprange.ToKey() != iprange.ToKey() || iprange.ToKey() == ipv4range.ToKey() || iprange.ToKey() == ipv6range.ToKey() {
		t.addFailure(newAddressItemFailure("range from nil "+iprange.String(), iprange))
	}

	iprange = ipaddr.NewIPSeqRange(&ipZero, nil)
	ipv4range = ipaddr.NewIPv4SeqRange(&ipv4Zero, nil).ToIP()
	ipv6range = ipaddr.NewIPv6SeqRange(&ipv6Zero, nil).ToIP()

	if iprange.ToKey() != iprange.ToKey() || iprange.ToKey() == ipv4range.ToKey() || iprange.ToKey() == ipv6range.ToKey() {
		t.addFailure(newAddressItemFailure("range from nil "+iprange.String(), iprange))
	}

	iprange = ipaddr.NewSequentialRange(&ipZero, nil)
	ipv4range = ipaddr.NewSequentialRange(&ipv4Zero, nil).ToIP()
	ipv6range = ipaddr.NewSequentialRange(&ipv6Zero, nil).ToIP()

	if iprange.ToKey() != iprange.ToKey() || iprange.ToKey() == ipv4range.ToKey() || iprange.ToKey() == ipv6range.ToKey() {
		t.addFailure(newAddressItemFailure("range from nil "+iprange.String(), iprange))
	}

	rng1 := ipaddr.NewSequentialRange[*ipaddr.IPAddress](nil, nil)
	rng2 := ipaddr.NewSequentialRange[*ipaddr.IPv4Address](nil, nil)
	rng3 := ipaddr.NewSequentialRange[*ipaddr.IPv6Address](nil, nil)

	if !rng1.Equal(iprange) || !iprange.Equal(rng1) {
		t.addFailure(newAddressItemFailure("range from double nil "+rng1.String()+" not same as zero range "+ipRangeZero.String(), iprange))
	}
	if !rng2.Equal(ipv4range) || !ipv4range.Equal(rng2) {
		t.addFailure(newAddressItemFailure("range from double nil "+rng2.String()+" not same as zero range "+ipv4range.String(), ipv4range))
	}
	if !rng3.Equal(ipv6range) || !ipv6range.Equal(rng3) {
		t.addFailure(newAddressItemFailure("range from double nil "+rng3.String()+" not same as zero range "+ipv6range.String(), ipv6range))
	}

	if !rng1.Equal(&ipRangeZero) || !ipRangeZero.Equal(rng1) {
		t.addFailure(newAddressItemFailure("range from double nil "+rng1.String()+" not same as zero range "+ipRangeZero.String(), iprange))
	}
	if !rng2.Equal(&ipv4RangeZero) || !ipv4RangeZero.Equal(rng2) {
		t.addFailure(newAddressItemFailure("range from double nil "+rng2.String()+" not same as zero range "+ipv4range.String(), ipv4range))
	}
	if !rng3.Equal(&ipv6RangeZero) || !ipv6RangeZero.Equal(rng3) {
		t.addFailure(newAddressItemFailure("range from double nil "+rng3.String()+" not same as zero range "+ipv6range.String(), ipv6range))
	}

	iprange = ipaddr.NewIPSeqRange(&ipZero, &ipZero)
	ipv4range = ipaddr.NewIPv4SeqRange(&ipv4Zero, &ipv4Zero).ToIP()
	ipv6range = ipaddr.NewIPv6SeqRange(&ipv6Zero, &ipv6Zero).ToIP()
	if iprange.ToKey() != iprange.ToKey() || iprange.ToKey() == ipv4range.ToKey() || iprange.ToKey() == ipv6range.ToKey() {
		t.addFailure(newAddressItemFailure("range from zeros "+iprange.String(), iprange))
	}

	if !rng1.Equal(iprange) || !iprange.Equal(rng1) {
		t.addFailure(newAddressItemFailure("range from double nil "+rng1.String()+" not same as zero range "+ipRangeZero.String(), iprange))
	}
	if !rng2.Equal(ipv4range) || !ipv4range.Equal(rng2) {
		t.addFailure(newAddressItemFailure("range from double nil "+rng2.String()+" not same as zero range "+ipv4range.String(), ipv4range))
	}
	if !rng3.Equal(ipv6range) || !ipv6range.Equal(rng3) {
		t.addFailure(newAddressItemFailure("range from double nil "+rng3.String()+" not same as zero range "+ipv6range.String(), ipv6range))
	}

	t.incrementTestCount()
}

func (t specialTypesTester) testNils() {
	var ipRangesIPv4 []*ipaddr.IPAddressSeqRange
	ipv4Addr1 := ipaddr.NewIPAddressString("1.2.3.3").GetAddress().ToIPv4()
	ipv4Addr2 := ipaddr.NewIPAddressString("2.2.3.4-5").GetAddress().ToIPv4()

	ipRangesIPv4 = append(ipRangesIPv4, nil)
	ipRangesIPv4 = append(ipRangesIPv4, &ipaddr.IPAddressSeqRange{})
	ipRangesIPv4 = append(ipRangesIPv4, ipaddr.NewIPv4SeqRange(nil, nil).ToIP())
	ipRangesIPv4 = append(ipRangesIPv4, (&ipaddr.IPv4AddressSeqRange{}).ToIP())
	ipRangesIPv4 = append(ipRangesIPv4, ipaddr.NewIPv4SeqRange(&ipaddr.IPv4Address{}, nil).ToIP())
	ipRangesIPv4 = append(ipRangesIPv4, ipaddr.NewIPv4SeqRange(ipv4Addr1, nil).ToIP())
	ipRangesIPv4 = append(ipRangesIPv4, ipaddr.NewIPv4SeqRange(nil, ipv4Addr2).ToIP())
	ipRangesIPv4 = append(ipRangesIPv4, ipaddr.NewIPv4SeqRange(ipv4Addr1, ipv4Addr2).ToIP())

	for i := range ipRangesIPv4 {
		range1 := ipRangesIPv4[i]
		for j := i; j < len(ipRangesIPv4); j++ {
			range2 := ipRangesIPv4[j]
			if i == j {
				if range1.Compare(range2) != 0 {
					t.addFailure(newSeqRangeFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
				} else if range2.Compare(range1) != 0 {
					t.addFailure(newSeqRangeFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
				} else if !range1.Equal(range2) {
					t.addFailure(newSeqRangeFailure(range1.String()+" and "+range2.String()+" not equal", range1))
				} else if !range2.Equal(range1) {
					t.addFailure(newSeqRangeFailure(range2.String()+" and "+range1.String()+" not equal", range1))
				}
			} else {
				if c := range1.Compare(range2); c > 0 {
					t.addFailure(newSeqRangeFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
				} else if c == 0 && !range1.Equal(range2) {
					t.addFailure(newSeqRangeFailure(range1.String()+" and "+range2.String()+" not equal", range1))
				} else if c2 := range2.Compare(range1); c2 < 0 {
					t.addFailure(newSeqRangeFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
				} else if c2 == 0 && (!range2.Equal(range1) || c != 0) {
					t.addFailure(newSeqRangeFailure(range2.String()+" and "+range1.String()+" not equal", range1))
				}
			}
		}
	}

	ipv6Addr1 := ipaddr.NewIPAddressString("1:2:3:3::").GetAddress().ToIPv6()
	ipv6Addr2 := ipaddr.NewIPAddressString("2:2:3:4-5::").GetAddress().ToIPv6()

	var ipRangesIPv6 []*ipaddr.IPAddressSeqRange

	ipRangesIPv6 = append(ipRangesIPv6, nil)
	ipRangesIPv6 = append(ipRangesIPv6, &ipaddr.IPAddressSeqRange{})
	ipRangesIPv6 = append(ipRangesIPv6, ipaddr.NewIPv6SeqRange(nil, nil).ToIP())
	ipRangesIPv6 = append(ipRangesIPv6, (&ipaddr.IPv6AddressSeqRange{}).ToIP())
	ipRangesIPv6 = append(ipRangesIPv6, ipaddr.NewIPv6SeqRange(ipv6Addr1, nil).ToIP())
	ipRangesIPv6 = append(ipRangesIPv6, ipaddr.NewIPv6SeqRange(nil, ipv6Addr2).ToIP())
	ipRangesIPv6 = append(ipRangesIPv6, ipaddr.NewIPv6SeqRange(ipv6Addr1, ipv6Addr2).ToIP())

	for i := range ipRangesIPv6 {
		range1 := ipRangesIPv6[i]
		for j := i; j < len(ipRangesIPv6); j++ {
			range2 := ipRangesIPv6[j]
			if i == j {
				if range1.Compare(range2) != 0 {
					t.addFailure(newSeqRangeFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
				} else if range2.Compare(range1) != 0 {
					t.addFailure(newSeqRangeFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
				} else if !range1.Equal(range2) {
					t.addFailure(newSeqRangeFailure(range1.String()+" and "+range2.String()+" not equal", range1))
				} else if !range2.Equal(range1) {
					t.addFailure(newSeqRangeFailure(range2.String()+" and "+range1.String()+" not equal", range1))
				}
			} else {
				if c := range1.Compare(range2); c > 0 {
					t.addFailure(newSeqRangeFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
				} else if c == 0 && !range1.Equal(range2) {
					t.addFailure(newSeqRangeFailure(range1.String()+" and "+range2.String()+" not equal", range1))
				} else if c2 := range2.Compare(range1); c2 < 0 {
					t.addFailure(newSeqRangeFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
				} else if c2 == 0 && (!range2.Equal(range1) || c != 0) {
					t.addFailure(newSeqRangeFailure(range2.String()+" and "+range1.String()+" not equal", range1))
				}
			}
		}
	}

	for _, range1 := range ipRangesIPv4 {
		for _, range2 := range ipRangesIPv6 {
			// the nils and the blank ranges
			c1 := range1.Compare(range2)
			c2 := range2.Compare(range1)
			if range1 == nil {
				if range2 == nil {
					if c1 != 0 || c2 != 0 {
						t.addFailure(newSeqRangeFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
					}
				} else if c1 >= 0 {
					t.addFailure(newSeqRangeFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
				}
			} else if range2 == nil {
				if c1 <= 0 || c2 >= 0 {
					t.addFailure(newSeqRangeFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
				}
			} else if range1.GetByteCount() == 0 {
				if range2.GetByteCount() == 0 {
					if c1 != 0 || c2 != 0 {
						t.addFailure(newSeqRangeFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
					}
				} else {
					if c1 >= 0 || c2 <= 0 {
						t.addFailure(newSeqRangeFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
					}
				}
			} else if range2.GetByteCount() == 0 {
				if c1 <= 0 || c2 >= 0 {
					t.addFailure(newSeqRangeFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
				}
			} else if c1 >= 0 {
				t.addFailure(newSeqRangeFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
			} else if c2 <= 0 {
				t.addFailure(newSeqRangeFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
			}
		}
	}

	nil1 := ipaddr.CountComparator.CompareSeries(nil, nil)
	nil2 := ipaddr.CountComparator.CompareRanges(nil, nil)
	nil3 := ipaddr.CountComparator.CompareAddresses(nil, nil)
	nil4 := ipaddr.CountComparator.CompareDivisions(nil, nil)
	nil5 := ipaddr.CountComparator.CompareAddressSections(nil, nil)
	nil6 := ipaddr.CountComparator.CompareSegments(nil, nil)
	nil7 := ipaddr.CountComparator.Compare(nil, nil)
	if nil1 != 0 || nil2 != 0 || nil3 != 0 || nil4 != 0 || nil5 != 0 || nil6 != 0 || nil7 != 0 {
		t.addFailure(newSegmentSeriesFailure("comparison of nils yields non-zero", nil))
	}

	ipv4Section1 := ipv4Addr1.GetSection()
	ipv6Section1 := ipv6Addr1.GetSection()

	ipv4Range1 := ipRangesIPv4[len(ipRangesIPv4)-1]
	ipv6Range1 := ipRangesIPv6[len(ipRangesIPv6)-1]

	ipv4Segment1 := ipv4Section1.GetSegment(0)
	ipv6Segment1 := ipv6Section1.GetSegment(0)
	ipDivision := ipaddr.NewDivision(11, 8)

	nil1 = ipaddr.CountComparator.CompareSeries(ipv4Addr1, nil)
	nil11 := ipaddr.CountComparator.CompareSeries(ipv6Addr1, nil)
	nil2 = ipaddr.CountComparator.CompareRanges(ipv4Range1, nil)
	nil21 := ipaddr.CountComparator.CompareRanges(ipv6Range1, nil)
	nil3 = ipaddr.CountComparator.CompareAddresses(ipv4Addr1, nil)
	nil4 = ipaddr.CountComparator.CompareDivisions(ipv4Segment1, nil)
	nil400 := ipaddr.CountComparator.CompareDivisions(ipv6Segment1, nil)
	nil40 := ipaddr.CountComparator.CompareDivisions(ipDivision, nil)
	nil41 := ipaddr.CountComparator.CompareSeries(ipv4Section1, nil)
	nil42 := ipaddr.CountComparator.CompareSeries(ipv6Section1, nil)
	nil5 = ipaddr.CountComparator.CompareAddressSections(ipv4Section1, nil)
	nil51 := ipaddr.CountComparator.CompareAddressSections(ipv6Section1, nil)
	nil6 = ipaddr.CountComparator.CompareSegments(ipv4Segment1, nil)
	nil60 := ipaddr.CountComparator.CompareSegments(ipv6Segment1, nil)
	nil7 = ipaddr.CountComparator.Compare(ipv4Addr1, nil)
	if nil1 <= 0 || nil11 <= 0 || nil2 <= 0 || nil21 <= 0 || nil3 <= 0 || nil4 <= 0 || nil400 <= 0 || nil40 <= 0 || nil41 <= 0 || nil42 <= 0 || nil5 <= 0 || nil51 <= 0 || nil6 <= 0 || nil60 <= 0 || nil7 <= 0 {
		t.addFailure(newSegmentSeriesFailure("comparison of nils yields negative", nil))
	}

	nil1 = ipaddr.CountComparator.CompareSeries(nil, ipv4Addr1)
	nil11 = ipaddr.CountComparator.CompareSeries(nil, ipv6Addr1)
	nil2 = ipaddr.CountComparator.CompareRanges(nil, ipv4Range1)
	nil21 = ipaddr.CountComparator.CompareRanges(nil, ipv6Range1)
	nil3 = ipaddr.CountComparator.CompareAddresses(nil, ipv4Addr1)
	nil4 = ipaddr.CountComparator.CompareDivisions(nil, ipv4Segment1)
	nil400 = ipaddr.CountComparator.CompareDivisions(nil, ipv6Segment1)
	nil40 = ipaddr.CountComparator.CompareDivisions(nil, ipDivision)
	nil41 = ipaddr.CountComparator.CompareSeries(nil, ipv4Section1)
	nil42 = ipaddr.CountComparator.CompareSeries(nil, ipv6Section1)
	nil5 = ipaddr.CountComparator.CompareAddressSections(nil, ipv4Section1)
	nil51 = ipaddr.CountComparator.CompareAddressSections(nil, ipv6Section1)
	nil6 = ipaddr.CountComparator.CompareSegments(nil, ipv4Segment1)
	nil60 = ipaddr.CountComparator.CompareSegments(nil, ipv6Segment1)
	nil7 = ipaddr.CountComparator.Compare(nil, ipv4Addr1)
	if nil1 >= 0 || nil11 >= 0 || nil2 >= 0 || nil21 >= 0 || nil3 >= 0 || nil4 >= 0 || nil400 >= 0 || nil40 >= 0 || nil41 >= 0 || nil42 >= 0 || nil5 >= 0 || nil51 >= 0 || nil6 >= 0 || nil60 >= 0 || nil7 >= 0 {
		t.addFailure(newSegmentSeriesFailure("comparison of nils yields positive", nil))
	}

	noIPV6Error := func(sect *ipaddr.IPv6AddressSection) *ipaddr.IPAddress {
		ipv6addrx, _ := ipaddr.NewIPv6Address(sect)
		return ipv6addrx.ToIP()
	}

	var ipAddressesIPv6 []*ipaddr.IPAddress

	ipAddressesIPv6 = append(ipAddressesIPv6, nil)
	ipAddressesIPv6 = append(ipAddressesIPv6, &ipaddr.IPAddress{})
	ipAddressesIPv6 = append(ipAddressesIPv6, (&ipaddr.IPv6Address{}).ToIP())
	ipAddressesIPv6 = append(ipAddressesIPv6, (&ipaddr.IPv6AddressSeqRange{}).GetLower().ToIP())
	ipAddressesIPv6 = append(ipAddressesIPv6, noIPV6Error(nil))
	ipAddressesIPv6 = append(ipAddressesIPv6, noIPV6Error(ipv6Section1))

	for i := range ipAddressesIPv6 {
		range1 := ipAddressesIPv6[i]
		//fmt.Printf("range %d using fmt is %v\n", i+1, range1)
		//fmt.Printf("range %d using Stringer is "+range1.String()+"\n\n", i+1)
		for j := i; j < len(ipAddressesIPv6); j++ {
			range2 := ipAddressesIPv6[j]
			if i == j {
				if range1.Compare(range2) != 0 {
					t.addFailure(newSegmentSeriesFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
				} else if range2.Compare(range1) != 0 {
					t.addFailure(newSegmentSeriesFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
				} else if !range1.Equal(range2) {
					t.addFailure(newSegmentSeriesFailure(range1.String()+" and "+range2.String()+" not equal", range1))
				} else if !range2.Equal(range1) {
					t.addFailure(newSegmentSeriesFailure(range2.String()+" and "+range1.String()+" not equal", range1))
				}
			} else {
				if c := range1.Compare(range2); c > 0 {
					t.addFailure(newSegmentSeriesFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
				} else if c == 0 && !range1.Equal(range2) {
					t.addFailure(newSegmentSeriesFailure(range1.String()+" and "+range2.String()+" not equal", range1))
				} else if c2 := range2.Compare(range1); c2 < 0 {
					t.addFailure(newSegmentSeriesFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
				} else if c2 == 0 && (!range2.Equal(range1) || c != 0) {
					t.addFailure(newSegmentSeriesFailure(range2.String()+" and "+range1.String()+" not equal", range1))
				}
			}
		}
	}

	noIPV4Error := func(sect *ipaddr.IPv4AddressSection) *ipaddr.IPAddress {
		ipv4addrx, _ := ipaddr.NewIPv4Address(sect)
		return ipv4addrx.ToIP()
	}

	var ipAddressesIPv4 []*ipaddr.IPAddress

	ipAddressesIPv4 = append(ipAddressesIPv4, nil)
	ipAddressesIPv4 = append(ipAddressesIPv4, &ipaddr.IPAddress{})
	ipAddressesIPv4 = append(ipAddressesIPv4, (&ipaddr.IPv4Address{}).ToIP())
	ipAddressesIPv4 = append(ipAddressesIPv4, (&ipaddr.IPv4AddressSeqRange{}).GetLower().ToIP())
	ipAddressesIPv4 = append(ipAddressesIPv4, noIPV4Error(nil))
	ipAddressesIPv4 = append(ipAddressesIPv4, noIPV4Error(ipv4Section1))

	for i := range ipAddressesIPv4 {
		range1 := ipAddressesIPv4[i]
		for j := i; j < len(ipAddressesIPv4); j++ {
			range2 := ipAddressesIPv4[j]
			if i == j {
				if range1.Compare(range2) != 0 {
					t.addFailure(newSegmentSeriesFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
				} else if range2.Compare(range1) != 0 {
					t.addFailure(newSegmentSeriesFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
				} else if !range1.Equal(range2) {
					t.addFailure(newSegmentSeriesFailure(range1.String()+" and "+range2.String()+" not equal", range1))
				} else if !range2.Equal(range1) {
					t.addFailure(newSegmentSeriesFailure(range2.String()+" and "+range1.String()+" not equal", range1))
				}
			} else {
				if c := range1.Compare(range2); c > 0 {
					t.addFailure(newSegmentSeriesFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
				} else if c == 0 && !range1.Equal(range2) {
					t.addFailure(newSegmentSeriesFailure(range1.String()+" and "+range2.String()+" not equal", range1))
				} else if c2 := range2.Compare(range1); c2 < 0 {
					t.addFailure(newSegmentSeriesFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
				} else if c2 == 0 && (!range2.Equal(range1) || c != 0) {
					t.addFailure(newSegmentSeriesFailure(range2.String()+" and "+range1.String()+" not equal", range1))
				}
			}
		}
	}

	var ipSectionsIPv6 []*ipaddr.IPAddressSection

	ipSectionsIPv6 = append(ipSectionsIPv6, nil)
	ipSectionsIPv6 = append(ipSectionsIPv6, &ipaddr.IPAddressSection{})
	ipSectionsIPv6 = append(ipSectionsIPv6, (&ipaddr.IPv6AddressSection{}).ToIP()) // note that this IP section can be any section type
	ipSectionsIPv6 = append(ipSectionsIPv6, ipv6Section1.ToIP())
	ipSectionsIPv6 = append(ipSectionsIPv6, ipv6Addr2.GetSection().ToIP())

	for i := range ipSectionsIPv6 {
		range1 := ipSectionsIPv6[i]
		for j := i; j < len(ipSectionsIPv6); j++ {
			range2 := ipSectionsIPv6[j]
			if i == j {
				if range1.Compare(range2) != 0 {
					t.addFailure(newSegmentSeriesFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
				} else if range2.Compare(range1) != 0 {
					t.addFailure(newSegmentSeriesFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
				} else if !range1.Equal(range2) {
					t.addFailure(newSegmentSeriesFailure(range1.String()+" and "+range2.String()+" not equal", range1))
				} else if !range2.Equal(range1) {
					t.addFailure(newSegmentSeriesFailure(range2.String()+" and "+range1.String()+" not equal", range1))
				}
			} else {
				if c := range1.Compare(range2); c > 0 {
					t.addFailure(newSegmentSeriesFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
				} else if c == 0 && !range1.Equal(range2) {
					t.addFailure(newSegmentSeriesFailure(range1.String()+" and "+range2.String()+" not equal", range1))
				} else if c2 := range2.Compare(range1); c2 < 0 {
					t.addFailure(newSegmentSeriesFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
				} else if c2 == 0 && (!range2.Equal(range1) || c != 0) {
					t.addFailure(newSegmentSeriesFailure(range2.String()+" and "+range1.String()+" not equal", range1))
				}
			}
		}
	}

	var ipSectionsIPv4 []*ipaddr.IPAddressSection

	ipSectionsIPv4 = append(ipSectionsIPv4, nil)
	ipSectionsIPv4 = append(ipSectionsIPv4, &ipaddr.IPAddressSection{})
	ipSectionsIPv4 = append(ipSectionsIPv4, (&ipaddr.IPv4AddressSection{}).ToIP())
	ipSectionsIPv4 = append(ipSectionsIPv4, ipv4Section1.ToIP())
	ipSectionsIPv4 = append(ipSectionsIPv4, ipv4Addr2.GetSection().ToIP())

	for i := range ipSectionsIPv4 {
		range1 := ipSectionsIPv4[i]
		for j := i; j < len(ipSectionsIPv4); j++ {
			range2 := ipSectionsIPv4[j]
			if i == j {
				if range1.Compare(range2) != 0 {
					t.addFailure(newSegmentSeriesFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
				} else if range2.Compare(range1) != 0 {
					t.addFailure(newSegmentSeriesFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
				} else if !range1.Equal(range2) {
					t.addFailure(newSegmentSeriesFailure(range1.String()+" and "+range2.String()+" not equal", range1))
				} else if !range2.Equal(range1) {
					t.addFailure(newSegmentSeriesFailure(range2.String()+" and "+range1.String()+" not equal", range1))
				}
			} else {
				if c := range1.Compare(range2); c > 0 {
					t.addFailure(newSegmentSeriesFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
				} else if c == 0 && !range1.Equal(range2) {
					t.addFailure(newSegmentSeriesFailure(range1.String()+" and "+range2.String()+" not equal", range1))
				} else if c2 := range2.Compare(range1); c2 < 0 {
					t.addFailure(newSegmentSeriesFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
				} else if c2 == 0 && (!range2.Equal(range1) || c != 0) {
					t.addFailure(newSegmentSeriesFailure(range2.String()+" and "+range1.String()+" not equal", range1))
				}
			}
		}
	}

	var ipSegmentsIPv6 []*ipaddr.AddressSegment

	ipv6SegMult := ipv6Addr2.GetSegment(3)

	ipSegmentsIPv6 = append(ipSegmentsIPv6, nil)
	ipSegmentsIPv6 = append(ipSegmentsIPv6, &ipaddr.AddressSegment{})
	ipSegmentsIPv6 = append(ipSegmentsIPv6, (&ipaddr.IPAddressSegment{}).ToSegmentBase())
	ipSegmentsIPv6 = append(ipSegmentsIPv6, (&ipaddr.IPv6AddressSegment{}).ToSegmentBase())
	ipSegmentsIPv6 = append(ipSegmentsIPv6, ipv6Segment1.ToSegmentBase())
	ipSegmentsIPv6 = append(ipSegmentsIPv6, ipv6SegMult.ToSegmentBase())

	for i := range ipSegmentsIPv6 {
		range1 := ipSegmentsIPv6[i]
		for j := i; j < len(ipSegmentsIPv6); j++ {
			range2 := ipSegmentsIPv6[j]
			if i == j {
				if range1.Compare(range2) != 0 {
					t.addFailure(newDivisionFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
				} else if range2.Compare(range1) != 0 {
					t.addFailure(newDivisionFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
				} else if !range1.Equal(range2) {
					t.addFailure(newDivisionFailure(range1.String()+" and "+range2.String()+" not equal", range1))
				} else if !range2.Equal(range1) {
					t.addFailure(newDivisionFailure(range2.String()+" and "+range1.String()+" not equal", range1))
				}
			} else {
				if c := range1.Compare(range2); c > 0 {
					t.addFailure(newDivisionFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
				} else if c == 0 && !range1.Equal(range2) {
					t.addFailure(newDivisionFailure(range1.String()+" and "+range2.String()+" not equal", range1))
				} else if c2 := range2.Compare(range1); c2 < 0 {
					t.addFailure(newDivisionFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
				} else if c2 == 0 && (!range2.Equal(range1) || c != 0) {
					t.addFailure(newDivisionFailure(range2.String()+" and "+range1.String()+" not equal", range1))
				}
			}
		}
	}

	var ipSegmentsIPv4 []*ipaddr.AddressSegment

	ipv4SegMult := ipv4Addr2.GetSegment(3)

	ipSegmentsIPv4 = append(ipSegmentsIPv4, nil)
	ipSegmentsIPv4 = append(ipSegmentsIPv4, &ipaddr.AddressSegment{})
	ipSegmentsIPv4 = append(ipSegmentsIPv4, (&ipaddr.IPAddressSegment{}).ToSegmentBase())
	ipSegmentsIPv4 = append(ipSegmentsIPv4, (&ipaddr.IPv4AddressSegment{}).ToSegmentBase())
	ipSegmentsIPv4 = append(ipSegmentsIPv4, ipv4Segment1.ToSegmentBase())
	ipSegmentsIPv4 = append(ipSegmentsIPv4, ipv4SegMult.ToSegmentBase())

	for i := range ipSegmentsIPv4 {
		range1 := ipSegmentsIPv4[i]
		for j := i; j < len(ipSegmentsIPv4); j++ {
			range2 := ipSegmentsIPv4[j]
			if i == j {
				if range1.Compare(range2) != 0 {
					t.addFailure(newDivisionFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
				} else if range2.Compare(range1) != 0 {
					t.addFailure(newDivisionFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
				} else if !range1.Equal(range2) {
					t.addFailure(newDivisionFailure(range1.String()+" and "+range2.String()+" not equal", range1))
				} else if !range2.Equal(range1) {
					t.addFailure(newDivisionFailure(range2.String()+" and "+range1.String()+" not equal", range1))
				}
			} else {
				if c := range1.Compare(range2); c > 0 {
					t.addFailure(newDivisionFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
				} else if c == 0 && !range1.Equal(range2) {
					t.addFailure(newDivisionFailure(range1.String()+" and "+range2.String()+" not equal", range1))
				} else if c2 := range2.Compare(range1); c2 < 0 {
					t.addFailure(newDivisionFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
				} else if c2 == 0 && (!range2.Equal(range1) || c != 0) {
					t.addFailure(newDivisionFailure(range2.String()+" and "+range1.String()+" not equal", range1))
				}
			}
		}
	}

	var ipv4AddressItems, ipv6AddressItems, ipv4RangeItems, ipv6RangeItems, ipv4SectionItems, ipv6SectionItems,
		ipv4SegmentItems, ipv6SegmentItems []ipaddr.AddressItem

	for _, item := range ipAddressesIPv4 {
		ipv4AddressItems = append(ipv4AddressItems, item)
	}
	for _, item := range ipAddressesIPv6 {
		// items in ipv6 list that are not specifically ipv6 are not necessarily bigger than similar items in ipv4 list
		if item.IsIPv6() {
			ipv6AddressItems = append(ipv6AddressItems, item)
		}
	}
	for _, item := range ipRangesIPv4 {
		ipv4RangeItems = append(ipv4RangeItems, item)
	}
	for _, item := range ipRangesIPv6 {
		// items in ipv6 list that are not specifically ipv6 are not necessarily bigger than similar items in ipv4 list
		if item.IsIPv6() {
			ipv6RangeItems = append(ipv6RangeItems, item)
		}
	}
	for _, item := range ipSectionsIPv4 {
		ipv4SectionItems = append(ipv4SectionItems, item)
	}
	for _, item := range ipSectionsIPv6 {
		// items in ipv6 list that are not specifically ipv6 are not necessarily bigger than similar items in ipv4 list

		if item.IsIPv6() && !item.IsAdaptiveZero() {
			ipv6SectionItems = append(ipv6SectionItems, item)
		}
	}
	for _, item := range ipSegmentsIPv4 {
		ipv4SegmentItems = append(ipv4SegmentItems, item)
	}
	for _, item := range ipSegmentsIPv6 {
		// items in ipv6 list that are not specifically ipv6 are not necessarily bigger than similar items in ipv4 list
		if item.IsIPv6() {
			ipv6SegmentItems = append(ipv6SegmentItems, item)
		}
	}

	// addresses > sections/groupings > seq ranges > divisions
	// ipv6 > ipv4s

	var allLists [][]ipaddr.AddressItem

	allLists = append(allLists, []ipaddr.AddressItem{nil})

	allLists = append(allLists, ipv4SegmentItems)
	allLists = append(allLists, ipv6SegmentItems)

	allLists = append(allLists, ipv4RangeItems)
	allLists = append(allLists, ipv6RangeItems)

	allLists = append(allLists, ipv4SectionItems)
	allLists = append(allLists, ipv6SectionItems)

	allLists = append(allLists, ipv4AddressItems)
	allLists = append(allLists, ipv6AddressItems)

	for i, list1 := range allLists {
		for j := i + 1; j < len(allLists); j++ {
			t.compareLists(list1, allLists[j])
		}
	}
}

func (t specialTypesTester) compareLists(items1, items2 []ipaddr.AddressItem) {
	for _, range1 := range items1 {
		for _, range2 := range items2 {
			// the nils and the blank ranges
			var c1, c2 int
			if range1 == nil || range2 == nil {
				c1 = ipaddr.CountComparator.Compare(range1, range2)
				c2 = ipaddr.CountComparator.Compare(range2, range1)
			} else {
				c1 = range1.Compare(range2)
				c2 = range2.Compare(range1)
			}
			if range1 == nil {
				if range2 == nil {
					if c1 != 0 || c2 != 0 {
						t.addFailure(newAddressItemFailure("comparison of nil with nil", range1))
					}
				} else if c1 >= 0 {
					ipaddr.CountComparator.Compare(range1, range2)
					t.addFailure(newAddressItemFailure("comparison of nil with "+range2.String(), range1))
				}
			} else if range2 == nil {
				if c1 <= 0 || c2 >= 0 {
					t.addFailure(newAddressItemFailure("comparison of "+range1.String()+" with nil", range1))
				}
			} else if c1 > 0 {
				range1.Compare(range2)
				t.addFailure(newAddressItemFailure("comparison of "+range1.String()+" with "+range2.String()+" yields "+strconv.Itoa(range1.Compare(range2)), range1))
			} else if c2 < 0 {
				t.addFailure(newAddressItemFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
			} else if c1 == 0 {
				if c2 != 0 {
					t.addFailure(newAddressItemFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
				} else if range1.GetCount().BitLen() != 0 && !range1.IsZero() {
					t.addFailure(newAddressItemFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
				} else if range2.GetCount().BitLen() != 0 && !range2.IsZero() {
					t.addFailure(newAddressItemFailure("comparison of "+range2.String()+" with "+range1.String()+" yields "+strconv.Itoa(range2.Compare(range1)), range1))
				}
			}
		}
	}
}

func getCount(segmentMax, segmentCount uint64) *big.Int {
	segMax := new(big.Int).SetUint64(segmentMax + 1)
	return segMax.Exp(segMax, new(big.Int).SetUint64(segmentCount), nil)
}
