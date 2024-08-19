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
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/seancfoley/ipaddress-go/ipaddr"
)

type ipAddressAllTester struct {
	ipAddressRangeTester
}

func (t ipAddressAllTester) run() {
	t.testSliceConversions()

	t.testMatches(true, "-", "*.*")
	t.testMatches(true, "-", "*.*.*.*")

	t.testMatches(true, "-0000000000000000efabffffffffffff", "00000000000000000000000000000000-0000000000000000efabffffffffffff")
	t.testMatches(true, "00000000000000000000000000000000-", "00000000000000000000000000000000-ffffffffffffffffffffffffffffffff")
	t.testMatches(true, "abfe0000000000000000000000000000-", "abfe0000000000000000000000000000-ffffffffffffffffffffffffffffffff")

	t.testMatches(true, "-0x0000000000000000efabffffffffffff", "00000000000000000000000000000000-0000000000000000efabffffffffffff")
	t.testMatches(true, "-0X0000000000000000efabffffffffffff", "00000000000000000000000000000000-0000000000000000efabffffffffffff")
	t.testMatches(true, "0x00000000000000000000000000000000-", "00000000000000000000000000000000-ffffffffffffffffffffffffffffffff")
	t.testMatches(true, "0xabcd0000000000000000000000000000-", "abcd0000000000000000000000000000-ffffffffffffffffffffffffffffffff")

	// these are the same addresses as the above tests in hex, but here in base 85

	t.testMatches(true, ipaddr.AlternativeRangeSeparatorStr+"0000000000=l?k|EPzi+", "00000000000000000000"+ipaddr.AlternativeRangeSeparatorStr+"0000000000=l?k|EPzi+")
	t.testMatches(true, "00000000000000000000"+ipaddr.AlternativeRangeSeparatorStr, "00000000000000000000"+ipaddr.AlternativeRangeSeparatorStr+"=r54lj&NUUO~Hi%c2ym0")
	t.testMatches(true, "oBky9Vh_d)e!eUd#8280"+ipaddr.AlternativeRangeSeparatorStr, "oBky9Vh_d)e!eUd#8280"+ipaddr.AlternativeRangeSeparatorStr+"=r54lj&NUUO~Hi%c2ym0")

	t.testMatches(true, "*.*.*.*", "-4294967295")   // ok on all tests
	t.testMatches(true, "*.*.*.*", "-0xffffffff")   // ok on all tests
	t.testMatches(true, "*.*.*.*", "-037777777777") // ok on all tests

	t.testMatches(true, "*.*.*.*", "0-")
	t.testMatches(true, "*.*.*.*", "-")

	t.testMatches(true, "0.-", "0.*.*.*")
	t.testMatches(true, "0.-", "0.*")
	t.testMatches(true, "0.0.-", "0.0.*.*")
	t.testMatches(true, "0.0.-", "0.0.*")
	t.testMatches(true, "0.-.0", "0.*.0.0") //ok
	t.testMatches(true, "-.0.-", "*.0.*.*") // more than one inferred range
	t.testMatches(true, "-.0.-", "*.0.*")
	t.testMatches(true, "1-.0.256-", "1-255.0.256-65535")    // 1-.0.256- becomes 1-255.0.*.255 // more than one inferred range
	t.testMatches(true, "0.1-.256-", "0.1-255.256-65535")    // more than one inferred range
	t.testMatches(true, "1-.65536-", "1-255.65536-16777215") // test more than one inferred range

	t.testMatches(true, "0b1.0b01.0b101.0b11111111", "1.1.5.255")
	t.testMatches(true, "0b1.0b01.0b101.0b11111111/16", "1.1.5.255/16")
	t.testMatches(true, "0b1.1.0b101.0b11111111/16", "1.1.5.255/16")

	t.testMatches(true, "aaaabbbbccccddddeeeeffffaaaabbbb", "aaaa:bbbb:cccc:dddd:eeee:ffff:aaaa:bbbb")
	t.testMatches(true, "aaaabbbbcccccdddffffffffffffffff-aaaabbbbccccdddd0000000000000000", "aaaa:bbbb:cccc:cddd-dddd:*:*:*:*")
	t.testMatches(true, "aaaabbbbccccdddd0000000000000000-aaaabbbbcccccdddffffffffffffffff", "aaaa:bbbb:cccc:cddd-dddd:*:*:*:*")

	t.testMatches(true, "4)+k&C#VzJ4br>0wv%Yp", "1080::8:800:200c:417a")
	t.testMatches(true, "=r54lj&NUUO~Hi%c2ym0", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")
	t.testMatches(true, "=r54lj&NUUO~Hi%c2yl0"+ipaddr.AlternativeRangeSeparatorStr+"=r54lj&NUUO~Hi%c2ym0", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffaa-ffff")
	t.testMatches(true, "ef86:1dc3:deba:d48:612d:f19c:de7d:e89c", "********************") // base 85
	t.testMatches(true, "--------------------", "f677:73f6:11b4:5073:4a06:76c2:ceae:1474")

	t.ipv6test(true, "0x00010002000300040000000000000000-0x0001000200030004ffffffffffffffff")
	t.ipv6test(true, "0x0001000200030004ffffffffffffffff-0x00010002000300040000000000000000")
	t.ipv6test(true, "0x00010002000300040000000000000000")

	t.ipv6test(true, "00010002000300040000000000000000-0001000200030004ffffffffffffffff")
	t.ipv6test(true, "0001000200030004ffffffffffffffff-00010002000300040000000000000000")
	t.ipv6test(true, "00010002000300040000000000000000")

	t.ipv6test(true, "00|M>t|ttwH6V6EEzblZ"+ipaddr.AlternativeRangeSeparatorStr+"00|M>t|ttwH6V6EEzkrZ")
	t.ipv6test(true, "00|M>t|ttwH6V6EEzkrZ"+ipaddr.AlternativeRangeSeparatorStr+"00|M>t|ttwH6V6EEzblZ")
	t.ipv6test(false, "00|M>t|ttwH6V6EEzkr"+ipaddr.AlternativeRangeSeparatorStr+"00|M>t|ttwH6V6EEzblZ")
	t.ipv6test(false, "00|M>t|ttwH6V6EEzkrZ"+ipaddr.AlternativeRangeSeparatorStr+"0|M>t|ttwH6V6EEzblZ")
	t.ipv6test(false, "00|M>t|ttwH6V6EEzkrZx"+ipaddr.AlternativeRangeSeparatorStr+"00|M>t|ttwH6V6EEzblZ")
	t.ipv6test(false, "00|M>t|ttwH6V6EEzkrZ"+ipaddr.AlternativeRangeSeparatorStr+"x00|M>t|ttwH6V6EEzblZ")

	t.ipv6test(true, "00000000000000000000000000000000-0001ffffffffffffffffffffffffffff")

	t.ipv6test(true, "=q{+M|w0(OeO5^F85=Cb")
	t.ipv6test(false, "=q{+M|w0.OeO5^F85=Cb")                                              // .
	t.ipv6test(false, "=q{+:|w0(OeO5^F85=Cb")                                              // :
	t.ipv6test(false, "=q{+M|w0(OeO5^F85=C/")                                              // / in middle
	t.ipv6test(false, "=q{+M|w0(OeO5^F85=/b")                                              // / in middle
	t.ipv6test(true, "=q{+M|w0(OeO5^F85=Cb/127")                                           // ok
	t.ipv6test(true, "=q{+-|w0(OeO5^-85=Cb")                                               // two '-'
	t.ipv6test(true, "=q{+M|w0(OeO5^F85=Cb"+ipaddr.IPv6AlternativeZoneSeparatorStr+"eth0") // ok
	t.ipv6test(false, "=q{+M|w0(OeO5^F85=C"+ipaddr.IPv6AlternativeZoneSeparatorStr+"eth0") // too soon

	t.testAllContains("*", "1:2:3:4:1:2:3:4", true)
	t.testAllContains("*", "1.2.3.4.5", false)
	t.testAllContains("*", "1.2.3.4", true)
	t.testAllContains("*/64", "1.2.3.4", false)
	t.testAllContains("*.*", "1::", false)
	t.testAllContains("*:*", "1::", true)
	t.testAllContains("*:*", "1.2.3.4", false)
	t.testAllContains("*.*", "1.2.3.4", true)
	t.testAllContains("*/64", "::", true)

	t.testNormalized("aaaabbbbcccccddd0000000000000000-aaaabbbbccccddddffffffffffffffff", "aaaa:bbbb:cccc:cddd-dddd:*:*:*:*")
	t.testCanonical("aaaabbbbcccccddd0000000000000000-aaaabbbbccccddddffffffffffffffff", "aaaa:bbbb:cccc:cddd-dddd:*:*:*:*")

	p0 := cacheTestBits(0)
	p1 := cacheTestBits(1)
	p15 := cacheTestBits(15)
	p16 := cacheTestBits(16)
	p32 := cacheTestBits(32)
	p64 := cacheTestBits(64)
	p89 := cacheTestBits(89)
	p126 := cacheTestBits(126)
	p128 := cacheTestBits(128)

	t.testSubnetStringRange2("*.0-65535", "0.0.0.0", "255.0.255.255", []interface{}{[]uint{0, 255}, []uint{0, 65535}}) // only valid with inet_aton allowed, and inet_aton takes precedence over wildcard
	t.testSubnetStringRange2("00000000000000000000000000000000-00000000000000000000007fffffffff", "::", "::7f:ffff:ffff",
		[]interface{}{[]*big.Int{bigZeroConst(), setBigString("00000000000000000000007fffffffff", 16)}})
	t.testSubnetStringRange2("00000000000000000000000000000000-00000000007fffffffffffffffffffff", "::", "::7f:ffff:ffff:ffff:ffff:ffff",
		[]interface{}{[]*big.Int{bigZeroConst(), setBigString("00000000007fffffffffffffffffffff", 16)}})
	t.testSubnetStringRange2("00000000000000000000000000000000-7fffffffffffffffffffffffffffffff", "::", "7fff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
		[]interface{}{[]*big.Int{bigZeroConst(), setBigString("7fffffffffffffffffffffffffffffff", 16)}})
	t.testSubnetStringRange2("00000000000000000000000000000000-ffffffffffffffffffffffffffffffff", "::", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
		[]interface{}{[]*big.Int{bigZeroConst(), setBigString("ffffffffffffffffffffffffffffffff", 16)}})
	t.testSubnetStringRange2("0000000000000000000000000000abcd-0000000000000000000000000000bbcd", "::abcd", "::bbcd",
		[]interface{}{[]uint{0xabcd, 0xbbcd}})

	t.testMaskedIncompatibleAddress("*/f0ff::", "::", "f0ff::")
	t.testMaskedIncompatibleAddress("*/129.0.0.0", "0.0.0.0", "129.0.0.0")

	t.testMaskedIncompatibleAddress("*:*/f0ff::", "::", "f0ff::")
	t.testMaskedIncompatibleAddress("*.*/129.0.0.0", "0.0.0.0", "129.0.0.0")

	t.testIncompatibleAddress2("*.257-65535", "0.0.1.1", "255.0.255.255", []interface{}{[2]uint{0, 255}, [2]uint{257, 65535}})                                                                                                                                                                                     //[0-255, 257-65535]
	t.testIncompatibleAddress2("1-1000", "1", "1000", []interface{}{[2]uint{1, 1000}})                                                                                                                                                                                                                             //[1-1000]
	t.testIncompatibleAddress2("50000-60000", "50000", "60000", []interface{}{[2]uint{50000, 60000}})                                                                                                                                                                                                              //[50000-60000]
	t.testIncompatibleAddress2("*.11-16000111", "0.11", "255.16000111", []interface{}{[2]uint{0, 255}, [2]uint{11, 16000111}})                                                                                                                                                                                     //[0-255, 11-16000111]
	t.testIncompatibleAddress2("0-255.11-16000111", "0.11", "255.16000111", []interface{}{[2]uint{0, 255}, [2]uint{11, 16000111}})                                                                                                                                                                                 //[0-255, 11-16000111] // inet_aton
	t.testIncompatibleAddress2("0-254.10101-16000111", "0.10101", "254.16000111", []interface{}{[2]uint{0, 254}, [2]uint{10101, 16000111}})                                                                                                                                                                        // [0-254, 10101-16000111] // inet_aton
	t.testIncompatibleAddress2("1.10101-16000111", "1.10101", "1.16000111", []interface{}{1, [2]uint{10101, 16000111}})                                                                                                                                                                                            //[1, 10101-16000111] // inet_aton
	t.testIncompatibleAddress2("3-1.10101-16000111", "1.10101", "3.16000111", []interface{}{[2]uint{1, 3}, [2]uint{10101, 16000111}})                                                                                                                                                                              //[1-3, 10101-16000111] // inet_aton
	t.testIncompatibleAddress2("00000000000000000000000000000000-abcdefabcdefabcdefabcdefabcdefab", "::", "abcd:efab:cdef:abcd:efab:cdef:abcd:efab", [2]*big.Int{bigZeroConst(), setBigString("abcdefabcdefabcdefabcdefabcdefab", 16)})                                                                            //[0-abcdefabcdefabcdefabcdefabcdefab]
	t.testIncompatibleAddress2("abcdefabcdefabcdefabcdefabcdefab-ffffffffffffffffffffffffffffffff", "abcd:efab:cdef:abcd:efab:cdef:abcd:efab", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", [2]*big.Int{setBigString("abcdefabcdefabcdefabcdefabcdefab", 16), setBigString("ffffffffffffffffffffffffffffffff", 16)}) //[abcdefabcdefabcdefabcdefabcdefab-ffffffffffffffffffffffffffffffff]
	t.testIncompatibleAddress2("abcdefabcdefabcdefabcdefabcdefab-bbcdefabcdefabcdefabcdefabcdefab", "abcd:efab:cdef:abcd:efab:cdef:abcd:efab", "bbcd:efab:cdef:abcd:efab:cdef:abcd:efab", [2]*big.Int{setBigString("abcdefabcdefabcdefabcdefabcdefab", 16), setBigString("bbcdefabcdefabcdefabcdefabcdefab", 16)}) //[abcdefabcdefabcdefabcdefabcdefab-bbcdefabcdefabcdefabcdefabcdefab]
	t.testIncompatibleAddress2("-abcdefabcdefabcdefabcdefabcdefab", "::", "abcd:efab:cdef:abcd:efab:cdef:abcd:efab", [2]*big.Int{bigZeroConst(), setBigString("abcdefabcdefabcdefabcdefabcdefab", 16)})                                                                                                            //[0-abcdefabcdefabcdefabcdefabcdefab]
	t.testIncompatibleAddress2("abcdefabcdefabcdefabcdefabcdefab-", "abcd:efab:cdef:abcd:efab:cdef:abcd:efab", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", [2]*big.Int{setBigString("abcdefabcdefabcdefabcdefabcdefab", 16), setBigString("ffffffffffffffffffffffffffffffff", 16)})                                 //[abcdefabcdefabcdefabcdefabcdefab-ffffffffffffffffffffffffffffffff]

	t.testIncompatibleAddress2("a:bb:c:dd:e:f:1.1-65535", "a:bb:c:dd:e:f:1.1", "a:bb:c:dd:e:f:1.65535", []interface{}{0xa, 0xbb, 0xc, 0xdd, 0xe, 0xf, 1, []uint{1, 0xffff}}) // mixed with inet_aton, mixed is incompatible address //[a, bb, c, dd, e, f, 1, 1-ffff]

	// with prefix lengths

	// inet_aton *.0.*.*/15
	t.testSubnetStringRange("*.0-65535/15", "0.0.0.0", "255.0.255.255", []interface{}{[2]uint{0, 255}, [2]uint{0, 65535}}, p15)   // only valid with inet_aton allowed, and inet_aton takes precedence over wildcard
	t.testSubnetStringRange("*.0-131071/15", "0.0.0.0", "255.1.255.255", []interface{}{[2]uint{0, 255}, [2]uint{0, 131071}}, p15) // only valid with inet_aton allowed, and inet_aton takes precedence over wildcard
	t.testSubnetStringRange("*.0.0-65535/15", "0.0.0.0", "255.0.255.255", []interface{}{[2]uint{0, 255}, 0, [2]uint{0, 65535}}, p15)
	t.testSubnetStringRange("*.0-1.0-65535/15", "0.0.0.0", "255.1.255.255", []interface{}{[2]uint{0, 255}, [2]uint{0, 1}, [2]uint{0, 65535}}, p15)

	t.testSubnetStringRange("00000000000000000000000000000000-00000000000000000000007fffffffff/89", "::", "::7f:ffff:ffff",
		[]interface{}{[2]*big.Int{bigZeroConst(), setBigString("00000000000000000000007fffffffff", 16)}}, p89)

	t.testSubnetStringRange("00000000000000000000000000000000-00000000007fffffffffffffffffffff/89", "::", "::7f:ffff:ffff:ffff:ffff:ffff",
		[]interface{}{[2]*big.Int{bigZeroConst(), setBigString("00000000007fffffffffffffffffffff", 16)}}, p89)

	t.testSubnetStringRange("00000000000000000000000000000000-7fffffffffffffffffffffffffffffff/0", "::", "7fff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
		[]interface{}{[2]*big.Int{bigZeroConst(), setBigString("7fffffffffffffffffffffffffffffff", 16)}}, p0)

	t.testSubnetStringRange("00000000000000000000000000000000-7fffffffffffffffffffffffffffffff/1", "::", "7fff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
		[]interface{}{[2]*big.Int{bigZeroConst(), setBigString("7fffffffffffffffffffffffffffffff", 17)}}, p1)

	t.testSubnetStringRange("00000000000000000000000000000000/1", "::", "7fff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
		[]interface{}{[2]*big.Int{bigZeroConst(), setBigString("7fffffffffffffffffffffffffffffff", 17)}}, p1)

	t.testSubnetStringRange("00000000000000000000000000000000-7fffffffffffffffffffffffffffffff/1", "::", "7fff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
		[]interface{}{[2]*big.Int{bigZeroConst(), setBigString("7fffffffffffffffffffffffffffffff", 16)}}, p1)

	t.testSubnetStringRange("0000000000000000000000000000abcd-0000000000000000000000000000bbcd/126", "::abcd", "::bbcd",
		[]interface{}{[2]uint{0xabcd, 0xbbcd}}, p126)

	t.testSubnetStringRange("00000000000000000000000000000000/89", "::", "::7f:ffff:ffff",
		[]interface{}{[2]*big.Int{bigZeroConst(), setBigString("00000000000000000000007fffffffff", 16)}}, p89)

	t.testIncompatibleAddress("*.11-16000111/32", "0.11", "255.16000111", []interface{}{[2]uint{0, 255}, [2]uint{11, 16000111}}, p32) //[0-255, 11-16000111]

	t.testIncompatibleAddress("*.257-65535/16", "0.0.1.1", "255.0.255.255", []interface{}{[2]uint{0, 255}, [2]uint{257, 65535}}, p16)                                                                                                                                                                                     //[0-255, 257-65535]
	t.testIncompatibleAddress("1-1000/16", "1", "1000", []interface{}{[2]uint{1, 1000}}, p16)                                                                                                                                                                                                                             //[1-1000]
	t.testIncompatibleAddress("50000-60000/16", "50000", "60000", []interface{}{[2]uint{50000, 60000}}, p16)                                                                                                                                                                                                              //[50000-60000]
	t.testIncompatibleAddress("3-1.10101-16000111/16", "1.10101", "3.16000111", []interface{}{[2]uint{1, 3}, [2]uint{10101, 16000111}}, p16)                                                                                                                                                                              //[1-3, 10101-16000111] // inet_aton
	t.testIncompatibleAddress("00000000000000000000000000000000-abcdefabcdefabcdefabcdefabcdefab/64", "::", "abcd:efab:cdef:abcd:efab:cdef:abcd:efab", [2]*big.Int{bigZeroConst(), setBigString("abcdefabcdefabcdefabcdefabcdefab", 16)}, p64)                                                                            //[0-abcdefabcdefabcdefabcdefabcdefab]
	t.testIncompatibleAddress("abcdefabcdefabcdefabcdefabcdefab-ffffffffffffffffffffffffffffffff/64", "abcd:efab:cdef:abcd:efab:cdef:abcd:efab", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", [2]*big.Int{setBigString("abcdefabcdefabcdefabcdefabcdefab", 16), setBigString("ffffffffffffffffffffffffffffffff", 16)}, p64) //[abcdefabcdefabcdefabcdefabcdefab-ffffffffffffffffffffffffffffffff]
	t.testIncompatibleAddress("abcdefabcdefabcdefabcdefabcdefab-bbcdefabcdefabcdefabcdefabcdefab/64", "abcd:efab:cdef:abcd:efab:cdef:abcd:efab", "bbcd:efab:cdef:abcd:efab:cdef:abcd:efab", [2]*big.Int{setBigString("abcdefabcdefabcdefabcdefabcdefab", 16), setBigString("bbcdefabcdefabcdefabcdefabcdefab", 16)}, p64) //[abcdefabcdefabcdefabcdefabcdefab-bbcdefabcdefabcdefabcdefabcdefab]
	t.testIncompatibleAddress("-abcdefabcdefabcdefabcdefabcdefab/64", "::", "abcd:efab:cdef:abcd:efab:cdef:abcd:efab", [2]*big.Int{bigZeroConst(), setBigString("abcdefabcdefabcdefabcdefabcdefab", 16)}, p64)                                                                                                            //[0-abcdefabcdefabcdefabcdefabcdefab]
	t.testIncompatibleAddress("abcdefabcdefabcdefabcdefabcdefab-/64", "abcd:efab:cdef:abcd:efab:cdef:abcd:efab", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", [2]*big.Int{setBigString("abcdefabcdefabcdefabcdefabcdefab", 16), setBigString("ffffffffffffffffffffffffffffffff", 16)}, p64)                                 //[abcdefabcdefabcdefabcdefabcdefab-ffffffffffffffffffffffffffffffff]

	t.testIncompatibleAddress2("a:bb:c:dd:e:f:1.1-65535", "a:bb:c:dd:e:f:1.1", "a:bb:c:dd:e:f:1.65535", []interface{}{0xa, 0xbb, 0xc, 0xdd, 0xe, 0xf, 1, []uint{1, 0xffff}}) // mixed with inet_aton, mixed is incompatible address //[a, bb, c, dd, e, f, 1, 1-ffff]

	t.testMaskedIncompatibleAddress("1234567890abcdef1234567890abcdef-2234567890abcdef1234567890abcdef/ffff:0:ffff:0:ffff:0:ffff:0",
		"1234::", "2234:0:ffff:0:ffff:0:ffff:0")

	t.testSubnetStringRange1("1234567890abcdef1234567890abcdef-2234567890abcdef1234567890abcdef/::ffff:ffff:FFFF:ffff:FFFF",
		"00000000000000000000000000000000", "000000000000ffffffffffffffffffff",
		[]interface{}{[]*big.Int{bigZeroConst(), setBigString("000000000000ffffffffffffffffffff", 16)}},
		nil, true,
	)
	t.testIncompatibleAddress1("1234567890abcdef1234567890abcdef-2234567890abcdef1234567890abcdef/ffff:ffff:ffff:ffff:ffff:FFFF:ffff:FFFF",
		"1234567890abcdef1234567890abcdef", "2234567890abcdef1234567890abcdef",
		[]interface{}{[]*big.Int{setBigString("1234567890abcdef1234567890abcdef", 16), setBigString("2234567890abcdef1234567890abcdef", 16)}},
		p128, true,
	)
	t.testSubnetStringRange1("1234567890abcdef1234567890abcdef-2234567890abcdef1234567890abcdef/fff:ffff:ffff:ffff:ffff:FFFF:ffff:FFFF",
		"00000000000000000000000000000000", "0fffffffffffffffffffffffffffffff",
		[]interface{}{[]*big.Int{bigZeroConst(), setBigString("0fffffffffffffffffffffffffffffff", 16)}},
		nil, true,
	)
	t.testMaskedIncompatibleAddress("1234567890abcdef1234567890abcdef-2234567890abcdef1234567890abcded/fff:ffff:ffff:ffff:ffff:FFFF:ffff:FFFF",
		"00000000000000000000000000000000", "0fffffffffffffffffffffffffffffff",
	)
	t.testSubnetStringRange1("1234567890abcdef1234567890abcdef-2234567890abcdef2234567890abcdef/::ffff:ffff:FFFF:ffff:FFFF",
		"00000000000000000000000000000000", "000000000000ffffffffffffffffffff",
		[]interface{}{[]*big.Int{bigZeroConst(), setBigString("000000000000ffffffffffffffffffff", 16)}},
		nil, true)
	t.testSubnetStringRange1("1234567890abcdef1234567890abcdef-2234567890abcdef2234567890abcdef/::FFFF:ffff:FFFF",
		"00000000000000000000000000000000", "00000000000000000000ffffffffffff",
		[]interface{}{[]*big.Int{bigZeroConst(), setBigString("00000000000000000000ffffffffffff", 16)}},
		nil, true)
	t.testMaskedIncompatibleAddress("1234567890abcdef1234567890abcdef-2234567890abcdef2234567890abcdef/::FFFF:ffff:0000",
		"00000000000000000000000000000000", "00000000000000000000ffffffff0000")

	t.testIncompatibleAddress1("1234567890abcdef1234567890abcdef-2234567890abcdef1234567890abcdef/ffff:FFFF:ffff:FFFF::",
		"1234567890abcdef1234567890abcdef", "2234567890abcdef1234567890abcdef",
		[]interface{}{[]*big.Int{setBigString("1234567890abcdef1234567890abcdef", 16), setBigString("2234567890abcdef1234567890abcdef", 16)}},
		p64, true)

	//void testMaskedRange(long value, long upperValue, long maskValue, boolean expectedIsSequential, long expectedLower, long expectedUpper) {
	t.testMaskedRange(2, 5, 2, false, 0, 2) // for range 2 to 5, masking with 2 gives range 2 to 0, ie reverse the range,
	t.testMaskedRange(2, 5, 6, false, 2, 4)
	t.testMaskedRange(2, 5, 7, true, 2, 5)
	t.testMaskedRange(2, 5, 1, true, 0, 1)
	t.testMaskedRange(1, 3, 1, true, 0, 1)
	t.testMaskedRange(2, 5, 0, true, 0, 0)
	t.testMaskedRange(1, 3, 0, true, 0, 0)

	t.testMaskedRange(1, 511, 511, true, 1, 511)
	t.testMaskedRange(101, 612, 511, true, 0, 511)
	t.testMaskedRange(102, 612, 511, false, 0, 511)
	t.testMaskedRange(102, 611, 511, false, 0, 511)

	t.testMaskedRange(1024, 1535, 511, true, 0, 511) //0x400 to 0x5ff with mask
	t.testMaskedRange(1024, 1534, 511, true, 0, 510)
	t.testMaskedRange(1026, 1536, 511, false, 0, 511)
	t.testMaskedRange(1025, 1536, 511, true, 0, 511)
	t.testMaskedRange(1025, 1535, 511, true, 1, 511)

	t.testMaskedRange(0x400, 0x5ff, 0x1ff, true, 0, 0x1ff) //0x400 to 0x5ff with mask
	t.testMaskedRange(0x400, 0x5fe, 0x1ff, true, 0, 0x1fe)
	t.testMaskedRange(0x402, 0x600, 0x1ff, false, 0, 0x1ff)
	t.testMaskedRange(0x401, 0x600, 0x1ff, true, 0, 0x1ff)
	t.testMaskedRange(0x401, 0x5ff, 0x1ff, true, 1, 0x1ff)
	t.testMaskedRange(0x401, 0x5ff, 0, true, 0, 0)
	t.testMaskedRange(0x401, 0x5ff, 1, true, 0, 1)

	// these 5 essentially the same as above 5 but in the extended 8 bytes
	t.testMaskedRange(0x40000000000, 0x5ffffffffff, 0x1ffffffffff, true, 0, 0x1ffffffffff)
	t.testMaskedRange(0x40000000000, 0x5fffffffffe, 0x1ffffffffff, true, 0, 0x1fffffffffe)
	t.testMaskedRange(0x40000000002, 0x60000000000, 0x1ffffffffff, false, 0, 0x1ffffffffff)
	t.testMaskedRange(0x40000000001, 0x60000000000, 0x1ffffffffff, true, 0, 0x1ffffffffff)
	t.testMaskedRange(0x40000000001, 0x5ffffffffff, 0x1ffffffffff, true, 1, 0x1ffffffffff)

	// mask 0x1ff is 9 ones, 5ff is 10 followed by 9 ones, 0x400 is 10 followed by 9 zeros
	// ignoring the last 7 zeros,
	// this is equivalent to 1000 to 1010 masked by 11, so we clearly must use the highest value to get the masked highest value
	t.testMaskedRange(0x40000000000, 0x5ff00000000, 0x1ffffffffff, true, 0, 0x1ff00000000)
	t.testMaskedRange(0x40000000000, 0x5fe00000000, 0x1ffffffffff, true, 0, 0x1fe00000000)
	// now this is equivalent to 1000 to 10000 masked by 11, so we've now include the mask value in the range
	// 0x600 is 110 followed by 8 zeros
	// 0x400 is 100 followed by 8 zeros
	// 0x401 is 100 followed by 7 zeros and a 1
	// 0x402 is 100 followed by 7 zeros and a 2
	// 0x1ff is 001 followed by 8 ones
	// so we can get the lowest value by masking the top value 0x600
	// and we need all values in between 0x600 and 0x601 to fill in the gap to 0x401 and make it sequential again
	t.testMaskedRange(0x40000000000, 0x60000000000, 0x1ffffffffff, true, 0, 0x1ffffffffff)
	t.testMaskedRange(0x40200000000, 0x60000000000, 0x1ffffffffff, false, 0, 0x1ffffffffff)
	t.testMaskedRange(0x40100000000, 0x60000000000, 0x1ffffffffff, false, 0, 0x1ffffffffff)
	t.testMaskedRange(0x40100000000, 0x600ffffffff, 0x1ffffffffff, true, 0, 0x1ffffffffff)

	t.testMaskedRange(0x40100000000, 0x5ff00000000, 0x1ffffffffff, true, 0x100000000, 0x1ff00000000)
	t.testMaskedRange(0x40100000000, 0x5ffffffffff, 0x1ffffffffff, true, 0x100000000, 0x1ffffffffff)
	t.testMaskedRange(0x400ffffffff, 0x5ffffffffff, 0x1ffffffffff, true, 0xffffffff, 0x1ffffffffff)

	t.testMaskedRangeExtended(
		1, 0xcafe, // lower
		1, 0xbadcafe, // upper
		0x1ff, 0x10000000, // mask
		math.MaxUint64, 0x10000000000-1, // max
		true, //sequential
		0, 0, // lower result
		0x1ff, 0) // upper result
	t.testMaskedRangeExtended(1, 0xcafe,
		1, 0xbadcafe,
		0x1fe, 0x10000000, // mask
		math.MaxUint64, 0x10000000000-1,
		false,
		0, 0,
		0x1fe, 0)
	t.testMaskedRangeExtended(1, 0xcafe,
		1, 0xbadcafe,
		math.MaxUint64, 0x10000000, // mask
		math.MaxUint64, 0x10000000000-1,
		true,
		0, 0,
		math.MaxUint64, 0)
	t.testMaskedRangeExtended(1, 0xcafe,
		1, 0xbadcafe,
		math.MaxUint64>>1, 0x10000000, // mask
		math.MaxUint64, 0x10000000000-1,
		true,
		0, 0,
		math.MaxUint64>>1, 0)
	t.testMaskedRangeExtended(1, 0xcafe,
		1, 0xbadcafe,
		1, 0x10000000, // mask
		math.MaxUint64, 0x10000000000-1,
		true,
		0, 0,
		1, 0)
	t.testMaskedRangeExtended(1, 0xcafe,
		1, 0xbadcafe,
		0, 0x10000000,
		math.MaxUint64, 0x10000000000-1,
		true,
		0, 0,
		0, 0)

	t.testStrings()

	t.testBackAndForth()

	t.ipAddressRangeTester.run()
}

func (t ipAddressAllTester) testBackAndForth() {
	t.testBackAndForthIPv4("127.0.0.1")
	t.testBackAndForthIPv4("128.0.0.1")
	t.testBackAndForthIPv4("255.255.255.255")
	t.testBackAndForthIPv4("128.255.255.255")
	t.testBackAndForthIPv6("::1")
	t.testBackAndForthIPv6("8000::1")
	t.testBackAndForthIPv6("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")
	t.testBackAndForthIPv6("ffff:a:b:c:d:e:f:cccc")
	t.testBackAndForthIPv6("cfff:a:b:c:d:e:f:cccc")
	t.testBackAndForthIPv6("7fff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")
}

func (t ipAddressAllTester) testBackAndForthIPv4(addrStr string) {
	// agnostic BigInteger and back
	addr := ipaddr.NewIPAddressString(addrStr).GetAddress()
	value := addr.GetValue()
	bigIntBytes := value.Bytes()
	byteCount := addr.GetByteCount()
	if len(bigIntBytes) < byteCount { // want correct byte length
		bytes := make([]byte, byteCount)
		copy(bytes[len(bytes)-len(bigIntBytes):], bigIntBytes)
		bigIntBytes = bytes
	}
	andAgain, _ := ipaddr.NewIPAddressFromNetIP(bigIntBytes)
	if !andAgain.Equal(addr) {
		t.addFailure(newIPAddrFailure("BigInteger result was "+andAgain.String()+" original was "+addr.String(), addr))
	}

	// byte[] and back
	bytes := addr.Bytes()
	backAgain, _ := ipaddr.NewIPAddressFromNetIP(bytes)
	if !backAgain.Equal(addr) {
		t.addFailure(newIPAddrFailure("bytes result was "+backAgain.String()+" original was "+addr.String(), addr))
	}

	// IPv4 int and back
	addrv4 := addr.ToIPv4()
	val := addrv4.Uint32Value()
	backAgainv4 := ipaddr.NewIPv4AddressFromUint32(val)
	if !backAgainv4.Equal(addrv4) {
		t.addFailure(newIPAddrFailure("int result was "+backAgainv4.String()+" original was "+addrv4.String(), addrv4.ToIP()))
	}
}

func (t ipAddressAllTester) testBackAndForthIPv6(addrStr string) {
	// agnostic BigInteger and back
	addr := ipaddr.NewIPAddressString(addrStr).GetAddress()
	value := addr.GetValue()
	bigIntBytes := value.Bytes()
	byteCount := addr.GetByteCount()
	if len(bigIntBytes) < byteCount { // want correct byte length
		bytes := make([]byte, byteCount)
		copy(bytes[len(bytes)-len(bigIntBytes):], bigIntBytes)
		bigIntBytes = bytes
	}
	andAgain, _ := ipaddr.NewIPAddressFromNetIP(bigIntBytes)
	if !andAgain.Equal(addr) {
		t.addFailure(newIPAddrFailure("BigInteger result was "+andAgain.String()+" original was "+addr.String(), addr))
	}

	// byte[] and back
	bytes := addr.Bytes()
	backAgain, _ := ipaddr.NewIPAddressFromNetIP(bytes)
	if !backAgain.Equal(addr) {
		t.addFailure(newIPAddrFailure("bytes result was "+backAgain.String()+" original was "+addr.String(), addr))
	}

	// IPv6 BigInteger and back
	addrv6 := addr.ToIPv6()
	value = addrv6.GetValue()
	backAgainv6, err := ipaddr.NewIPv6AddressFromInt(value)
	if err != nil {
		t.addFailure(newIPAddrFailure("got error creating from bytes "+value.String()+" err: "+err.Error(), addr))
	} else if !backAgainv6.Equal(addrv6) {
		t.addFailure(newIPAddrFailure("int result was "+backAgainv6.String()+" original was "+addrv6.String(), addrv6.ToIP()))
	}
}

// tests the maskRange method and its counterpart that works with divs > 64 bits, maskExtendedRange
func (t ipAddressAllTester) testMaskedRange(value, upperValue, maskValue uint64, expectedIsSequential bool, expectedLower, expectedUpper uint64) {
	masker := ipaddr.MaskRange(value, upperValue, maskValue, math.MaxUint64)
	lowerResult := masker.GetMaskedLower(value, maskValue)
	upperResult := masker.GetMaskedUpper(upperValue, maskValue)
	isSequential := masker.IsSequential()
	if isSequential != expectedIsSequential || lowerResult != expectedLower || upperResult != expectedUpper {
		reason := ""
		if lowerResult != expectedLower {
			reason += "lower mismatch " + strconv.FormatUint(lowerResult, 10) + "(" + strconv.FormatUint(lowerResult, 2) + ") with expected " +
				strconv.FormatUint(expectedLower, 10) + "(" + strconv.FormatUint(expectedLower, 2) + ") "
		}
		if upperResult != expectedUpper {
			reason += "upper mismatch " + strconv.FormatUint(upperResult, 10) + "(" + strconv.FormatUint(upperResult, 2) + ") with expected " +
				strconv.FormatUint(expectedUpper, 10) + "(" + strconv.FormatUint(expectedUpper, 2) + ") "
		}
		if isSequential != expectedIsSequential {
			reason += "sequential mismatch "
		}
		t.addFailure(newFailure("invalid masking, "+reason+
			strconv.FormatUint(value, 10)+"("+strconv.FormatUint(value, 2)+")"+" to "+
			strconv.FormatUint(upperValue, 10)+"("+strconv.FormatUint(upperValue, 2)+")"+" masked with "+
			strconv.FormatUint(maskValue, 10)+"("+strconv.FormatUint(maskValue, 2)+")"+" results in "+
			strconv.FormatUint(lowerResult, 10)+"("+strconv.FormatUint(lowerResult, 2)+")"+" lower and "+
			strconv.FormatUint(upperResult, 10)+"("+strconv.FormatUint(upperResult, 2)+")"+" upper and sequential "+
			strconv.FormatBool(isSequential)+" instead of expected "+
			strconv.FormatUint(expectedLower, 10)+"("+strconv.FormatUint(expectedLower, 2)+")"+" lower and "+
			strconv.FormatUint(expectedUpper, 10)+"("+strconv.FormatUint(expectedUpper, 2)+")"+" upper and sequential "+
			strconv.FormatBool(expectedIsSequential), nil))
	}
	t.incrementTestCount()
	t.testMaskedRangeExtended(value, 0, upperValue, 0, maskValue, 0, math.MaxUint64, math.MaxUint64,
		expectedIsSequential, expectedLower, 0, expectedUpper, 0)
	t.testMaskedRangeExtended(0, value, math.MaxUint64, upperValue, math.MaxUint64, maskValue, math.MaxUint64, math.MaxUint64,
		expectedIsSequential, 0, expectedLower, math.MaxUint64, expectedUpper)
}

func (t ipAddressAllTester) testMaskedRangeExtended(value, extendedValue,
	upperValue, extendedUpperValue,
	maskValue, extendedMaskValue,
	maxValue, extendedMaxValue uint64,
	expectedIsSequential bool,
	expectedLower, expectedExtendedLower,
	expectedUpper, expectedExtendedUpper uint64) {
	masker := ipaddr.MaskExtendedRange(
		value, extendedValue,
		upperValue, extendedUpperValue,
		maskValue, extendedMaskValue,
		maxValue, extendedMaxValue)
	lowerResult := masker.GetMaskedLower(value, maskValue)
	upperResult := masker.GetMaskedUpper(upperValue, maskValue)
	extendedLowerResult := masker.GetExtendedMaskedLower(extendedValue, extendedMaskValue)
	extendedUpperResult := masker.GetExtendedMaskedUpper(extendedUpperValue, extendedMaskValue)
	isSequential := masker.IsSequential()
	if masker.IsSequential() != expectedIsSequential ||
		lowerResult != expectedLower || upperResult != expectedUpper ||
		extendedLowerResult != expectedExtendedLower || extendedUpperResult != expectedExtendedUpper {
		reason := ""
		if lowerResult != expectedLower || extendedLowerResult != expectedExtendedLower {
			reason += "lower mismatch "
		}
		if upperResult != expectedUpper || extendedUpperResult != expectedExtendedUpper {
			reason += "upper mismatch "
		}
		if isSequential != expectedIsSequential {
			reason += "sequential mismatch "
		}
		t.addFailure(newFailure("invalid masking, "+reason, nil))

	}
	t.incrementTestCount()
}

func (t ipAddressAllTester) testAllContains(cidr1, cidr2 string, result bool) {
	wstr := t.createAddress(cidr1)
	w2str := t.createAddress(cidr2)

	t.testStringContains(result, false, wstr, w2str)

	t.incrementTestCount()
}

func (t ipAddressAllTester) testStrings() {
	//It is good to have at least one base 85 input test, since we have code that caches base 85 input strings for output
	t.testIPv6Strings("4)+k&C#VzJ4br>0wv%Yp",
		"1080:0:0:0:8:800:200c:417a", //normalized
		"1080:0:0:0:8:800:200c:417a", //normalizedWildcards
		"1080::8:800:200c:417a",      //canonicalWildcards
		"1080:0:0:0:8:800:200c:417a", //sql
		"1080:0000:0000:0000:0008:0800:200c:417a",
		"1080::8:800:200c:417a", //compressed
		"1080::8:800:200c:417a",
		"1080::8:800:200c:417a",    //subnet
		"1080::8:800:200c:417a",    //compressedWildcard
		"1080::8:800:32.12.65.122", //mixed no compress
		"1080::8:800:32.12.65.122", //mixedNoCompressHost
		"1080::8:800:32.12.65.122",
		"1080::8:800:32.12.65.122",
		"a.7.1.4.c.0.0.2.0.0.8.0.8.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.0.1.ip6.arpa",
		"1080-0-0-0-8-800-200c-417a.ipv6-literal.net",
		"4)+k&C#VzJ4br>0wv%Yp",
		"0x108000000000000000080800200c417a",
		"00204000000000000000000000100200004003040572")

	t.testIPv6Strings("008JOm8Mm5*yBppL!sg0",
		"0:ffff:ffff:ffff:ffff:ffff:ffff:ffff", //normalized
		"0:ffff:ffff:ffff:ffff:ffff:ffff:ffff", //normalizedWildcards
		"0:ffff:ffff:ffff:ffff:ffff:ffff:ffff", //canonicalWildcards
		"0:ffff:ffff:ffff:ffff:ffff:ffff:ffff", //sql
		"0000:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
		"::ffff:ffff:ffff:ffff:ffff:ffff:ffff", //compressed
		"0:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
		"::ffff:ffff:ffff:ffff:ffff:ffff:ffff",       //subnet
		"::ffff:ffff:ffff:ffff:ffff:ffff:ffff",       //compressedWildcard
		"::ffff:ffff:ffff:ffff:ffff:255.255.255.255", //mixed no compress
		"::ffff:ffff:ffff:ffff:ffff:255.255.255.255", //mixedNoCompressHost
		"::ffff:ffff:ffff:ffff:ffff:255.255.255.255",
		"::ffff:ffff:ffff:ffff:ffff:255.255.255.255",
		"f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.0.0.0.0.ip6.arpa",
		"0-ffff-ffff-ffff-ffff-ffff-ffff-ffff.ipv6-literal.net",
		"008JOm8Mm5*yBppL!sg0",
		"0x0000ffffffffffffffffffffffffffff",
		"00000017777777777777777777777777777777777777")

	t.testIPv6Strings("=r54lj&NUUO~Hi%c2ym0",
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", //normalized
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", //normalizedWildcards
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", //canonicalWildcards
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", //sql
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", //compressed
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",       //subnet
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",       //compressedWildcard
		"ffff:ffff:ffff:ffff:ffff:ffff:255.255.255.255", //mixed no compress
		"ffff:ffff:ffff:ffff:ffff:ffff:255.255.255.255", //mixedNoCompressHost
		"ffff:ffff:ffff:ffff:ffff:ffff:255.255.255.255",
		"ffff:ffff:ffff:ffff:ffff:ffff:255.255.255.255",
		"f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.f.ip6.arpa",
		"ffff-ffff-ffff-ffff-ffff-ffff-ffff-ffff.ipv6-literal.net",
		"=r54lj&NUUO~Hi%c2ym0",
		"0xffffffffffffffffffffffffffffffff",
		"03777777777777777777777777777777777777777777")
}

func (t ipAddressAllTester) testSliceConversions() {
	t.testIPAddrSliceConversion("1.2.3.4")
	t.testIPAddrSliceConversion("aa:bb:cccc:dd:eeee:ffff:a:b")
	t.testMACAddrSliceConversion("aa:bb:cc:dd:ee:ff")
}

func (t ipAddressAllTester) testIPAddrSliceConversion(addrString string) {
	addr := t.createAddress(addrString).GetAddress()
	// start with ip addresses.
	ipAddrSlice := []*ipaddr.IPAddress{addr, addr, addr}
	if addr.IsIPv4() {
		ipv4AddrSlice := ipaddr.ToIPv4Slice[*ipaddr.IPAddress, *ipaddr.IPv4Address](ipAddrSlice)
		if !ipaddr.AddrsMatchOrdered(ipAddrSlice, ipv4AddrSlice) {
			t.addFailure(newIPAddrFailure(fmt.Sprintf("failed to match %v with %v", ipAddrSlice, ipv4AddrSlice), addr))
		}
		ipAddrSlice = ipaddr.ToIPSlice[*ipaddr.IPv4Address, *ipaddr.IPAddress](ipv4AddrSlice)
	} else if addr.IsIPv6() {
		ipv6AddrSlice := ipaddr.ToIPv6Slice[*ipaddr.IPAddress, *ipaddr.IPv6Address](ipAddrSlice)
		if !ipaddr.AddrsMatchOrdered(ipAddrSlice, ipv6AddrSlice) {
			t.addFailure(newIPAddrFailure(fmt.Sprintf("failed to match %v with %v", ipAddrSlice, ipv6AddrSlice), addr))
		}
		ipAddrSlice = ipaddr.ToIPSlice[*ipaddr.IPv6Address, *ipaddr.IPAddress](ipv6AddrSlice)
	}
	addrSlice := ipaddr.ToAddressBaseSlice[*ipaddr.IPAddress, *ipaddr.Address](ipAddrSlice)
	if !ipaddr.AddrsMatchOrdered(ipAddrSlice, addrSlice) {
		t.addFailure(newIPAddrFailure(fmt.Sprintf("failed to match %v with %v", addrSlice, ipAddrSlice), addr))
	}
	if addr.IsIPv4() {
		ipv4AddrSlice := ipaddr.ToIPv4Slice[*ipaddr.Address, *ipaddr.IPv4Address](addrSlice)
		if !ipaddr.AddrsMatchOrdered(addrSlice, ipv4AddrSlice) {
			t.addFailure(newIPAddrFailure(fmt.Sprintf("failed to match %v with %v", ipAddrSlice, ipv4AddrSlice), addr))
		}
		addrSlice = ipaddr.ToAddressBaseSlice[*ipaddr.IPv4Address, *ipaddr.Address](ipv4AddrSlice)
	} else if addr.IsIPv6() {
		ipv6AddrSlice := ipaddr.ToIPv6Slice[*ipaddr.Address, *ipaddr.IPv6Address](addrSlice)
		if !ipaddr.AddrsMatchOrdered(addrSlice, ipv6AddrSlice) {
			t.addFailure(newIPAddrFailure(fmt.Sprintf("failed to match %v with %v", ipAddrSlice, ipv6AddrSlice), addr))
		}
		addrSlice = ipaddr.ToAddressBaseSlice[*ipaddr.IPv6Address, *ipaddr.Address](ipv6AddrSlice)
	}
	t.testSectionSliceConversion(addr.ToAddressBase().GetSection())
	t.testSegmentSliceConversion(addr.ToAddressBase().GetSegment(0))

	ipAddrTypeSlice := []ipaddr.IPAddressType{addr, addr, addr}
	ipAddrSlice = ipaddr.ToIPSlice[ipaddr.IPAddressType, *ipaddr.IPAddress](ipAddrTypeSlice)
	if !ipaddr.AddrsMatchOrdered(ipAddrTypeSlice, ipAddrSlice) {
		t.addFailure(newIPAddrFailure(fmt.Sprintf("failed to match %v with %v", ipAddrTypeSlice, ipAddrSlice), addr))
	}

	addrTypeSlice := []ipaddr.AddressType{addr, addr, addr}
	addrSlice = ipaddr.ToAddressBaseSlice[ipaddr.AddressType, *ipaddr.Address](addrTypeSlice)
	if !ipaddr.AddrsMatchOrdered(addrTypeSlice, addrSlice) {
		t.addFailure(newIPAddrFailure(fmt.Sprintf("failed to match %v with %v", addrTypeSlice, addrSlice), addr))
	}
	if !ipaddr.AddrsMatchUnordered(addrTypeSlice, addrSlice) {
		t.addFailure(newIPAddrFailure(fmt.Sprintf("failed to match %v with %v", addrTypeSlice, addrSlice), addr))
	}

	//ranges
	rng := addr.SpanWithRange(addr)
	ipRangeSlice := []*ipaddr.IPAddressSeqRange{rng, rng, rng}
	ipRangeSlice2 := ipRangeSlice
	if addr.IsIPv4() {
		ipv4AddrSlice := ipaddr.ToIPv4Slice[*ipaddr.IPAddressSeqRange, *ipaddr.IPv4AddressSeqRange](ipRangeSlice)
		if !ipv4AddrSlice[0].Equal(ipRangeSlice[0]) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipv4AddrSlice[0], ipRangeSlice[0]), ipRangeSlice[0]))
		}
		ipRangeSlice2 = ipaddr.ToIPSlice[*ipaddr.IPv4AddressSeqRange, *ipaddr.IPAddressSeqRange](ipv4AddrSlice)
	} else if addr.IsIPv6() {
		ipv6AddrSlice := ipaddr.ToIPv6Slice[*ipaddr.IPAddressSeqRange, *ipaddr.IPv6AddressSeqRange](ipRangeSlice)
		if !ipv6AddrSlice[0].Equal(ipRangeSlice[0]) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipv6AddrSlice[0], ipRangeSlice[0]), ipRangeSlice[0]))
		}
		ipRangeSlice2 = ipaddr.ToIPSlice[*ipaddr.IPv6AddressSeqRange, *ipaddr.IPAddressSeqRange](ipv6AddrSlice)
	}
	if !ipRangeSlice2[0].Equal(ipRangeSlice[0]) {
		t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipRangeSlice2[0], ipRangeSlice[0]), ipRangeSlice[0]))
	}
	ipRangeTypeSlice := []ipaddr.IPAddressSeqRangeType{rng, rng, rng}
	ipRangeSlice2 = ipaddr.ToIPSlice[ipaddr.IPAddressSeqRangeType, *ipaddr.IPAddressSeqRange](ipRangeTypeSlice)
	if !ipRangeSlice2[0].Equal(ipRangeSlice[0]) {
		t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipRangeSlice2[0], ipRangeSlice[0]), ipRangeSlice[0]))
	}

	// from there you can get ranges, wrapped addresses and wrapped sections.
	// ExtendedIPSegmentSeries, WrappedIPAddress, WrappedIPAddressSection, ToIPv6 ToIPv4
	// ExtendedSegmentSeries, WrappedAddress, or WrappedAddressSection ToIPv6 ToIPv4 ToIP ToMAC

	ipWrappedSlice := []ipaddr.WrappedIPAddress{addr.Wrap(), addr.Wrap(), addr.Wrap()}
	ipAddrSlice = ipaddr.ToIPSlice[ipaddr.WrappedIPAddress, *ipaddr.IPAddress](ipWrappedSlice)
	if !ipaddr.AddrsMatchOrdered(ipAddrTypeSlice, ipAddrSlice) {
		t.addFailure(newIPAddrFailure(fmt.Sprintf("failed to match %v with %v", ipAddrTypeSlice, ipAddrSlice), addr))
	}
	if addr.IsIPv4() {
		ipWrappedExtSlice := []ipaddr.ExtendedIPSegmentSeries{addr.Wrap(), addr.Wrap(), addr.Wrap()}
		ipv4SeriesSlice := ipaddr.ToIPv4Slice[ipaddr.ExtendedIPSegmentSeries, ipaddr.IPv4AddressSegmentSeries](ipWrappedExtSlice)
		if ipv4SeriesSlice[0].Compare(addr) != 0 {
			t.addFailure(newIPAddrFailure(fmt.Sprintf("failed to match %v with %v", ipv4SeriesSlice[0], addr), addr))
		}

		wrappedExtSlice := []ipaddr.ExtendedSegmentSeries{addr.ToAddressBase().Wrap(), addr.ToAddressBase().Wrap(), addr.ToAddressBase().Wrap()}
		ipv4SerSlice := ipaddr.ToIPv4Slice[ipaddr.ExtendedSegmentSeries, ipaddr.IPv4AddressSegmentSeries](wrappedExtSlice)
		if ipv4SerSlice[0].Compare(addr) != 0 {
			t.addFailure(newIPAddrFailure(fmt.Sprintf("failed to match %v with %v", ipv4SerSlice[0], addr), addr))
		}
	} else if addr.IsIPv6() {
		ipWrappedExtSlice := []ipaddr.ExtendedIPSegmentSeries{addr.Wrap(), addr.Wrap(), addr.Wrap()}
		ipv6SeriesSlice := ipaddr.ToIPv6Slice[ipaddr.ExtendedIPSegmentSeries, ipaddr.IPv6AddressSegmentSeries](ipWrappedExtSlice)
		if ipv6SeriesSlice[0].Compare(addr) != 0 {
			t.addFailure(newIPAddrFailure(fmt.Sprintf("failed to match %v with %v", ipv6SeriesSlice[0], addr), addr))
		}

		wrappedExtSlice := []ipaddr.ExtendedSegmentSeries{addr.ToAddressBase().Wrap(), addr.ToAddressBase().Wrap(), addr.ToAddressBase().Wrap()}
		ipv6SerSlice := ipaddr.ToIPv6Slice[ipaddr.ExtendedSegmentSeries, ipaddr.IPv6AddressSegmentSeries](wrappedExtSlice)
		if ipv6SerSlice[0].Compare(addr) != 0 {
			t.addFailure(newIPAddrFailure(fmt.Sprintf("failed to match %v with %v", ipv6SerSlice[0], addr), addr))
		}
	}
}

func (t ipAddressAllTester) testMACAddrSliceConversion(addrString string) {
	addr := t.createMACAddress(addrString).GetAddress()
	// start with ip addresses.
	macAddrSlice := []*ipaddr.MACAddress{addr, addr, addr}
	addrSlice := ipaddr.ToAddressBaseSlice[*ipaddr.MACAddress, *ipaddr.Address](macAddrSlice)
	if !ipaddr.AddrsMatchOrdered(macAddrSlice, addrSlice) {
		t.addFailure(newMACAddrFailure(fmt.Sprintf("failed to match %v with %v", addrSlice, macAddrSlice), addr))
	}
	macAddrSlice = ipaddr.ToMACSlice[*ipaddr.Address, *ipaddr.MACAddress](addrSlice)
	if !ipaddr.AddrsMatchOrdered(macAddrSlice, addrSlice) {
		t.addFailure(newMACAddrFailure(fmt.Sprintf("failed to match %v with %v", addrSlice, macAddrSlice), addr))
	}
	t.testSectionSliceConversion(addr.ToAddressBase().GetSection())
	t.testSegmentSliceConversion(addr.ToAddressBase().GetSegment(0))

	wrappedExtSlice := []ipaddr.ExtendedSegmentSeries{addr.ToAddressBase().Wrap(), addr.ToAddressBase().Wrap(), addr.ToAddressBase().Wrap()}
	macSerSlice := ipaddr.ToMACSlice[ipaddr.ExtendedSegmentSeries, ipaddr.MACAddressSegmentSeries](wrappedExtSlice)
	if macSerSlice[0].Compare(addr) != 0 {
		t.addFailure(newMACAddrFailure(fmt.Sprintf("failed to match %v with %v", macSerSlice[0], addr), addr))
	}
}

func (t ipAddressAllTester) testSegmentSliceConversion(seg *ipaddr.AddressSegment) {
	segSlice := []*ipaddr.AddressSegment{seg, seg, seg}
	if seg.IsIPv4() {
		ipv4AddrSlice := ipaddr.ToIPv4Slice[*ipaddr.AddressSegment, *ipaddr.IPv4AddressSegment](segSlice)
		if !ipv4AddrSlice[0].Equal(segSlice[0]) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipv4AddrSlice[0], segSlice[0]), segSlice[0]))
		}
		ipSecSlice := ipaddr.ToIPSlice[*ipaddr.IPv4AddressSegment, *ipaddr.IPAddressSegment](ipv4AddrSlice)
		if !ipv4AddrSlice[0].Equal(ipSecSlice[0]) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipv4AddrSlice[0], ipSecSlice[0]), ipSecSlice[0]))
		}
	} else if seg.IsIPv6() {
		ipv6AddrSlice := ipaddr.ToIPv6Slice[*ipaddr.AddressSegment, *ipaddr.IPv6AddressSegment](segSlice)
		if !ipv6AddrSlice[0].Equal(segSlice[0]) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipv6AddrSlice[0], segSlice[0]), segSlice[0]))
		}
		ipSecSlice := ipaddr.ToIPSlice[*ipaddr.IPv6AddressSegment, *ipaddr.IPAddressSegment](ipv6AddrSlice)
		if !ipv6AddrSlice[0].Equal(ipSecSlice[0]) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipv6AddrSlice[0], ipSecSlice[0]), ipSecSlice[0]))
		}
	} else if seg.IsMAC() {
		macAddrSlice := ipaddr.ToMACSlice[*ipaddr.AddressSegment, *ipaddr.MACAddressSegment](segSlice)
		if !macAddrSlice[0].Equal(segSlice[0]) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", macAddrSlice[0], segSlice[0]), segSlice[0]))
		}
	}
	div := seg.ToDiv()
	divSlice := []*ipaddr.AddressDivision{div, div, div}
	if seg.IsIPv4() {
		ipv4AddrSlice := ipaddr.ToIPv4Slice[*ipaddr.AddressDivision, *ipaddr.IPv4AddressSegment](divSlice)
		if !ipv4AddrSlice[0].Equal(divSlice[0].ToSegmentBase()) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipv4AddrSlice[0], divSlice[0]), divSlice[0]))
		}
		ipdivSlice := ipaddr.ToIPSlice[*ipaddr.IPv4AddressSegment, *ipaddr.IPAddressSegment](ipv4AddrSlice)
		if !ipv4AddrSlice[0].Equal(ipdivSlice[0].ToSegmentBase()) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipv4AddrSlice[0], ipdivSlice[0]), ipdivSlice[0]))
		}
	} else if seg.IsIPv6() {
		ipv6AddrSlice := ipaddr.ToIPv6Slice[*ipaddr.AddressDivision, *ipaddr.IPv6AddressSegment](divSlice)
		if !ipv6AddrSlice[0].Equal(divSlice[0].ToSegmentBase()) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipv6AddrSlice[0], divSlice[0]), divSlice[0]))
		}
		ipdivSlice := ipaddr.ToIPSlice[*ipaddr.IPv6AddressSegment, *ipaddr.IPAddressSegment](ipv6AddrSlice)
		if !ipv6AddrSlice[0].Equal(ipdivSlice[0].ToSegmentBase()) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipv6AddrSlice[0], ipdivSlice[0]), ipdivSlice[0]))
		}
	} else if seg.IsMAC() {
		macAddrSlice := ipaddr.ToMACSlice[*ipaddr.AddressDivision, *ipaddr.MACAddressSegment](divSlice)
		if !macAddrSlice[0].Equal(divSlice[0].ToSegmentBase()) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", macAddrSlice[0], divSlice[0]), divSlice[0]))
		}
	}
}

func (t ipAddressAllTester) testSectionSliceConversion(section *ipaddr.AddressSection) {
	secSlice := []*ipaddr.AddressSection{section, section, section}
	if section.IsIPv4() {
		ipv4AddrSlice := ipaddr.ToIPv4Slice[*ipaddr.AddressSection, *ipaddr.IPv4AddressSection](secSlice)
		if !ipv4AddrSlice[0].Equal(secSlice[0]) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipv4AddrSlice[0], secSlice[0]), secSlice[0]))
		}
		ipSecSlice := ipaddr.ToIPSlice[*ipaddr.IPv4AddressSection, *ipaddr.IPAddressSection](ipv4AddrSlice)
		if !ipv4AddrSlice[0].Equal(ipSecSlice[0]) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipv4AddrSlice[0], ipSecSlice[0]), ipSecSlice[0]))
		}
	} else if section.IsIPv6() {
		ipv6AddrSlice := ipaddr.ToIPv6Slice[*ipaddr.AddressSection, *ipaddr.IPv6AddressSection](secSlice)
		if !ipv6AddrSlice[0].Equal(secSlice[0]) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipv6AddrSlice[0], secSlice[0]), secSlice[0]))
		}
		ipSecSlice := ipaddr.ToIPSlice[*ipaddr.IPv6AddressSection, *ipaddr.IPAddressSection](ipv6AddrSlice)
		if !ipv6AddrSlice[0].Equal(ipSecSlice[0]) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipv6AddrSlice[0], ipSecSlice[0]), ipSecSlice[0]))
		}
	} else if section.IsMAC() {
		macAddrSlice := ipaddr.ToMACSlice[*ipaddr.AddressSection, *ipaddr.MACAddressSection](secSlice)
		if !macAddrSlice[0].Equal(secSlice[0]) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", macAddrSlice[0], secSlice[0]), secSlice[0]))
		}
	}
	div := section.ToDivGrouping()
	divSlice := []*ipaddr.AddressDivisionGrouping{div, div, div}
	if section.IsIPv4() {
		ipv4AddrSlice := ipaddr.ToIPv4Slice[*ipaddr.AddressDivisionGrouping, *ipaddr.IPv4AddressSection](divSlice)
		if !ipv4AddrSlice[0].GetSegment(0).Equal(divSlice[0].GetDivision(0).ToSegmentBase()) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipv4AddrSlice[0], divSlice[0]), divSlice[0]))
		}
		ipdivSlice := ipaddr.ToIPSlice[*ipaddr.IPv4AddressSection, *ipaddr.IPAddressSection](ipv4AddrSlice)
		if !ipv4AddrSlice[0].GetSegment(0).Equal(ipdivSlice[0].GetSegment(0).ToSegmentBase()) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipv4AddrSlice[0], ipdivSlice[0]), ipdivSlice[0]))
		}
	} else if section.IsIPv6() {
		ipv6AddrSlice := ipaddr.ToIPv6Slice[*ipaddr.AddressDivisionGrouping, *ipaddr.IPv6AddressSection](divSlice)
		if !ipv6AddrSlice[0].GetSegment(0).Equal(divSlice[0].GetDivision(0).ToSegmentBase()) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipv6AddrSlice[0], divSlice[0]), divSlice[0]))
		}
		ipdivSlice := ipaddr.ToIPSlice[*ipaddr.IPv6AddressSection, *ipaddr.IPAddressSection](ipv6AddrSlice)
		if !ipv6AddrSlice[0].GetSegment(0).Equal(ipdivSlice[0].GetSegment(0).ToSegmentBase()) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", ipv6AddrSlice[0], ipdivSlice[0]), ipdivSlice[0]))
		}
	} else if section.IsMAC() {
		macAddrSlice := ipaddr.ToMACSlice[*ipaddr.AddressDivisionGrouping, *ipaddr.MACAddressSection](divSlice)
		if !macAddrSlice[0].GetSegment(0).Equal(divSlice[0].GetDivision(0).ToSegmentBase()) {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", macAddrSlice[0], divSlice[0]), divSlice[0]))
		}
	}
	wrappedSlice := []ipaddr.WrappedAddressSection{section.Wrap(), section.Wrap(), section.Wrap()}
	if section.IsIP() {
		ipAddrSlice := ipaddr.ToIPSlice[ipaddr.WrappedAddressSection, ipaddr.IPAddressSegmentSeries](wrappedSlice)
		_ = ipAddrSlice
		if wrappedSlice[0].Compare(ipAddrSlice[0]) != 0 {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", wrappedSlice[0], ipAddrSlice[0]), section))
		}
	} else {
		macAddrSlice := ipaddr.ToMACSlice[ipaddr.WrappedAddressSection, ipaddr.MACAddressSegmentSeries](wrappedSlice)
		_ = macAddrSlice
		if wrappedSlice[0].Compare(macAddrSlice[0]) != 0 {
			t.addFailure(newAddressItemFailure(fmt.Sprintf("failed to match %v with %v", wrappedSlice[0], macAddrSlice[0]), section))
		}
	}
}
