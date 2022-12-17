//
// Copyright 2022 Sean C Foley
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

package test

import "github.com/seancfoley/ipaddress-go/ipaddr"

type trieStrings struct {
	addrs []string
	treeString, addedNodeString,
	treeToIndexString, addedNodeToIndexString string
}

var treeOne = trieStrings{
	addrs: []string{
		"1::ffff:2:3:5",
		"1::ffff:2:3:4",
		"1::ffff:2:3:6",
		"1::ffff:2:3:12",
		"1::ffff:aa:3:4",
		"1::ff:aa:3:4",
		"1::ff:aa:3:12",
		"bb::ffff:2:3:6",
		"bb::ffff:2:3:12",
		"bb::ffff:2:3:22",
		"bb::ffff:2:3:32",
		"bb::ffff:2:3:42",
		"bb::ffff:2:3:43",
	},

	treeString: "\n" +
		"○ ::/0 (13)\n" +
		"└─○ ::/8 (13)\n" +
		"  ├─○ 1::/64 (7)\n" +
		"  │ ├─○ 1::ff:aa:3:0/123 (2)\n" +
		"  │ │ ├─● 1::ff:aa:3:4 (1)\n" +
		"  │ │ └─● 1::ff:aa:3:12 (1)\n" +
		"  │ └─○ 1::ffff:0:0:0/88 (5)\n" +
		"  │   ├─○ 1::ffff:2:3:0/123 (4)\n" +
		"  │   │ ├─○ 1::ffff:2:3:4/126 (3)\n" +
		"  │   │ │ ├─○ 1::ffff:2:3:4/127 (2)\n" +
		"  │   │ │ │ ├─● 1::ffff:2:3:4 (1)\n" +
		"  │   │ │ │ └─● 1::ffff:2:3:5 (1)\n" +
		"  │   │ │ └─● 1::ffff:2:3:6 (1)\n" +
		"  │   │ └─● 1::ffff:2:3:12 (1)\n" +
		"  │   └─● 1::ffff:aa:3:4 (1)\n" +
		"  └─○ bb::ffff:2:3:0/121 (6)\n" +
		"    ├─○ bb::ffff:2:3:0/122 (4)\n" +
		"    │ ├─○ bb::ffff:2:3:0/123 (2)\n" +
		"    │ │ ├─● bb::ffff:2:3:6 (1)\n" +
		"    │ │ └─● bb::ffff:2:3:12 (1)\n" +
		"    │ └─○ bb::ffff:2:3:20/123 (2)\n" +
		"    │   ├─● bb::ffff:2:3:22 (1)\n" +
		"    │   └─● bb::ffff:2:3:32 (1)\n" +
		"    └─○ bb::ffff:2:3:42/127 (2)\n" +
		"      ├─● bb::ffff:2:3:42 (1)\n" +
		"      └─● bb::ffff:2:3:43 (1)\n",

	addedNodeString: "\n" +
		"○ ::/0\n" +
		"├─● 1::ff:aa:3:4\n" +
		"├─● 1::ff:aa:3:12\n" +
		"├─● 1::ffff:2:3:4\n" +
		"├─● 1::ffff:2:3:5\n" +
		"├─● 1::ffff:2:3:6\n" +
		"├─● 1::ffff:2:3:12\n" +
		"├─● 1::ffff:aa:3:4\n" +
		"├─● bb::ffff:2:3:6\n" +
		"├─● bb::ffff:2:3:12\n" +
		"├─● bb::ffff:2:3:22\n" +
		"├─● bb::ffff:2:3:32\n" +
		"├─● bb::ffff:2:3:42\n" +
		"└─● bb::ffff:2:3:43\n",

	treeToIndexString: "\n" +
		"○ ::/0 = 0 (13)\n" +
		"└─○ ::/8 = 0 (13)\n" +
		"  ├─○ 1::/64 = 0 (7)\n" +
		"  │ ├─○ 1::ff:aa:3:0/123 = 0 (2)\n" +
		"  │ │ ├─● 1::ff:aa:3:4 = 5 (1)\n" +
		"  │ │ └─● 1::ff:aa:3:12 = 6 (1)\n" +
		"  │ └─○ 1::ffff:0:0:0/88 = 0 (5)\n" +
		"  │   ├─○ 1::ffff:2:3:0/123 = 0 (4)\n" +
		"  │   │ ├─○ 1::ffff:2:3:4/126 = 0 (3)\n" +
		"  │   │ │ ├─○ 1::ffff:2:3:4/127 = 0 (2)\n" +
		"  │   │ │ │ ├─● 1::ffff:2:3:4 = 1 (1)\n" +
		"  │   │ │ │ └─● 1::ffff:2:3:5 = 0 (1)\n" +
		"  │   │ │ └─● 1::ffff:2:3:6 = 2 (1)\n" +
		"  │   │ └─● 1::ffff:2:3:12 = 3 (1)\n" +
		"  │   └─● 1::ffff:aa:3:4 = 4 (1)\n" +
		"  └─○ bb::ffff:2:3:0/121 = 0 (6)\n" +
		"    ├─○ bb::ffff:2:3:0/122 = 0 (4)\n" +
		"    │ ├─○ bb::ffff:2:3:0/123 = 0 (2)\n" +
		"    │ │ ├─● bb::ffff:2:3:6 = 7 (1)\n" +
		"    │ │ └─● bb::ffff:2:3:12 = 8 (1)\n" +
		"    │ └─○ bb::ffff:2:3:20/123 = 0 (2)\n" +
		"    │   ├─● bb::ffff:2:3:22 = 9 (1)\n" +
		"    │   └─● bb::ffff:2:3:32 = 10 (1)\n" +
		"    └─○ bb::ffff:2:3:42/127 = 0 (2)\n" +
		"      ├─● bb::ffff:2:3:42 = 11 (1)\n" +
		"      └─● bb::ffff:2:3:43 = 12 (1)\n",

	addedNodeToIndexString: "\n" +
		"○ ::/0 = 0\n" +
		"├─● 1::ff:aa:3:4 = 5\n" +
		"├─● 1::ff:aa:3:12 = 6\n" +
		"├─● 1::ffff:2:3:4 = 1\n" +
		"├─● 1::ffff:2:3:5 = 0\n" +
		"├─● 1::ffff:2:3:6 = 2\n" +
		"├─● 1::ffff:2:3:12 = 3\n" +
		"├─● 1::ffff:aa:3:4 = 4\n" +
		"├─● bb::ffff:2:3:6 = 7\n" +
		"├─● bb::ffff:2:3:12 = 8\n" +
		"├─● bb::ffff:2:3:22 = 9\n" +
		"├─● bb::ffff:2:3:32 = 10\n" +
		"├─● bb::ffff:2:3:42 = 11\n" +
		"└─● bb::ffff:2:3:43 = 12\n",
}

var treeTwo = trieStrings{
	addrs: []string{
		"ff80::/8",
		"ff80:8000::/16",
		"ff80:8000::/24",
		"ff80:8000::/32",
		"ff80:8000:c000::/34",
		"ff80:8000:c800::/36",
		"ff80:8000:cc00::/38",
		"ff80:8000:cc00::/40",
	},

	treeString: "\n" +
		"○ ::/0 (8)\n" +
		"└─○ ff80::/16 (8)\n" +
		"  ├─● ff80:: (1)\n" +
		"  └─● ff80:8000::/24 (7)\n" +
		"    └─● ff80:8000::/32 (6)\n" +
		"      ├─● ff80:8000:: (1)\n" +
		"      └─● ff80:8000:c000::/34 (4)\n" +
		"        └─○ ff80:8000:c800::/37 (3)\n" +
		"          ├─● ff80:8000:c800:: (1)\n" +
		"          └─● ff80:8000:cc00::/38 (2)\n" +
		"            └─● ff80:8000:cc00::/40 (1)\n",

	addedNodeString: "\n" +
		"○ ::/0\n" +
		"├─● ff80::\n" +
		"└─● ff80:8000::/24\n" +
		"  └─● ff80:8000::/32\n" +
		"    ├─● ff80:8000::\n" +
		"    └─● ff80:8000:c000::/34\n" +
		"      ├─● ff80:8000:c800::\n" +
		"      └─● ff80:8000:cc00::/38\n" +
		"        └─● ff80:8000:cc00::/40\n",

	treeToIndexString: "\n" +
		"○ ::/0 = 0 (8)\n" +
		"└─○ ff80::/16 = 0 (8)\n" +
		"  ├─● ff80:: = 0 (1)\n" +
		"  └─● ff80:8000::/24 = 2 (7)\n" +
		"    └─● ff80:8000::/32 = 3 (6)\n" +
		"      ├─● ff80:8000:: = 1 (1)\n" +
		"      └─● ff80:8000:c000::/34 = 4 (4)\n" +
		"        └─○ ff80:8000:c800::/37 = 0 (3)\n" +
		"          ├─● ff80:8000:c800:: = 5 (1)\n" +
		"          └─● ff80:8000:cc00::/38 = 6 (2)\n" +
		"            └─● ff80:8000:cc00::/40 = 7 (1)\n",

	addedNodeToIndexString: "\n" +
		"○ ::/0 = 0\n" +
		"├─● ff80:: = 0\n" +
		"└─● ff80:8000::/24 = 2\n" +
		"  └─● ff80:8000::/32 = 3\n" +
		"    ├─● ff80:8000:: = 1\n" +
		"    └─● ff80:8000:c000::/34 = 4\n" +
		"      ├─● ff80:8000:c800:: = 5\n" +
		"      └─● ff80:8000:cc00::/38 = 6\n" +
		"        └─● ff80:8000:cc00::/40 = 7\n",
}

var treeThree = trieStrings{
	addrs: []string{
		"192.168.10.0/24",
		"192.168.10.0/26",
		"192.168.10.64/27",
		"192.168.10.96/27",
		"192.168.10.128/30",
		"192.168.10.132/30",
		"192.168.10.136/30",
	},

	treeString: "\n" +
		"○ 0.0.0.0/0 (7)\n" +
		"└─● 192.168.10.0/24 (7)\n" +
		"  ├─○ 192.168.10.0/25 (3)\n" +
		"  │ ├─● 192.168.10.0/26 (1)\n" +
		"  │ └─○ 192.168.10.64/26 (2)\n" +
		"  │   ├─● 192.168.10.64/27 (1)\n" +
		"  │   └─● 192.168.10.96/27 (1)\n" +
		"  └─○ 192.168.10.128/28 (3)\n" +
		"    ├─○ 192.168.10.128/29 (2)\n" +
		"    │ ├─● 192.168.10.128/30 (1)\n" +
		"    │ └─● 192.168.10.132/30 (1)\n" +
		"    └─● 192.168.10.136/30 (1)\n",

	addedNodeString: "\n" +
		"○ 0.0.0.0/0\n" +
		"└─● 192.168.10.0/24\n" +
		"  ├─● 192.168.10.0/26\n" +
		"  ├─● 192.168.10.64/27\n" +
		"  ├─● 192.168.10.96/27\n" +
		"  ├─● 192.168.10.128/30\n" +
		"  ├─● 192.168.10.132/30\n" +
		"  └─● 192.168.10.136/30\n",

	treeToIndexString: "\n" +
		"○ 0.0.0.0/0 = 0 (7)\n" +
		"└─● 192.168.10.0/24 = 0 (7)\n" +
		"  ├─○ 192.168.10.0/25 = 0 (3)\n" +
		"  │ ├─● 192.168.10.0/26 = 1 (1)\n" +
		"  │ └─○ 192.168.10.64/26 = 0 (2)\n" +
		"  │   ├─● 192.168.10.64/27 = 2 (1)\n" +
		"  │   └─● 192.168.10.96/27 = 3 (1)\n" +
		"  └─○ 192.168.10.128/28 = 0 (3)\n" +
		"    ├─○ 192.168.10.128/29 = 0 (2)\n" +
		"    │ ├─● 192.168.10.128/30 = 4 (1)\n" +
		"    │ └─● 192.168.10.132/30 = 5 (1)\n" +
		"    └─● 192.168.10.136/30 = 6 (1)\n",

	addedNodeToIndexString: "\n" +
		"○ 0.0.0.0/0 = 0\n" +
		"└─● 192.168.10.0/24 = 0\n" +
		"  ├─● 192.168.10.0/26 = 1\n" +
		"  ├─● 192.168.10.64/27 = 2\n" +
		"  ├─● 192.168.10.96/27 = 3\n" +
		"  ├─● 192.168.10.128/30 = 4\n" +
		"  ├─● 192.168.10.132/30 = 5\n" +
		"  └─● 192.168.10.136/30 = 6\n",
}

var treeFour = trieStrings{
	addrs: []string{},
	treeString: "\n" +
		"<nil>\n",
	addedNodeString: "\n" +
		"○ <nil>\n",
	treeToIndexString: "\n" +
		"<nil>\n",
	addedNodeToIndexString: "\n" +
		"○ <nil> = 0\n",
}

var testIPAddressTries = [][]string{{
	"1.2.3.4",
	"1.2.3.5",
	"1.2.3.6",
	"1.2.3.3",
	"1.2.3.255",
	"2.2.3.5",
	"2.2.3.128",
	"2.2.3.0/24",
	"2.2.4.0/24",
	"2.2.7.0/24",
	"2.2.4.3",
	"1::ffff:2:3:5",
	"1::ffff:2:3:4",
	"1::ffff:2:3:6",
	"1::ffff:2:3:12",
	"1::ffff:aa:3:4",
	"1::ff:aa:3:4",
	"1::ff:aa:3:12",
	"bb::ffff:2:3:6",
	"bb::ffff:2:3:12",
	"bb::ffff:2:3:22",
	"bb::ffff:2:3:32",
	"bb::ffff:2:3:42",
	"bb::ffff:2:3:43",
}, {
	"0.0.0.0/8",
	"0.0.0.0/16",
	"0.0.0.0/24",
	"0.0.0.0",
}, {
	"1.2.3.4",
}, {}, {
	"128.0.0.0",
}, {
	"0.0.0.0",
}, {
	"0.0.0.0/0",
	"128.0.0.0/8",
	"128.128.0.0/16",
	"128.128.128.0/24",
	"128.128.128.128",
}, {
	"0.0.0.0/0",
	"0.0.0.0/8",
	"0.128.0.0/16",
	"0.128.0.0/24",
	"0.128.0.128",
}, {
	"128.0.0.0/8",
	"128.128.0.0/16",
	"128.128.128.0/24",
	"128.128.128.128",
}, {
	"0.0.0.0/8",
	"0.128.0.0/16",
	"0.128.0.0/24",
	"0.128.0.128",
}, {
	"ff80::/8",
	"ff80:8000::/16",
	"ff80:8000::/24",
	"ff80:8000::/32",
	"ff80:8000:c000::/34",
	"ff80:8000:c800::/36",
	"ff80:8000:cc00::/38",
	"ff80:8000:cc00::/40",
}, {
	"0.0.0.0/0",
	"128.0.0.0/8",
	"128.0.0.0/16",
	"128.0.128.0/24",
	"128.0.128.0",
}, {
	"0.0.0.0/0",
	"0.0.0.0/8",
	"0.0.0.0/16",
	"0.0.0.0/24",
	"0.0.0.0",
},
	{
		"1.2.3.0",
		"1.2.3.0/31", // consecutive
		"1.2.3.1",
		"1.2.3.0/30",
		"1.2.3.2",
	},
}

var testMACTries = [][]string{{
	"a:b:c:d:e:f",
	"f:e:c:d:b:a",
	"a:b:c:*:*:*",
	"a:b:c:d:*:*",
	"a:b:c:e:f:*",
}, {
	"a:b:c:d:e:f",
}, {
	"a:b:c:d:*:*",
}, {}, {
	"a:a:a:b:c:d:e:f",
	"a:a:f:e:c:d:b:a",
	"a:a:a:b:c:*:*:*",
	"a:a:a:b:c:d:*:*",
	"a:a:a:b:c:b:*:*",
	"a:a:a:b:c:e:f:*",
}, {
	"*:*:*:*:*:*",
},
}

type AddressKey = ipaddr.Key[*ipaddr.Address]

func collect(addrs []string, converter func(string) *ipaddr.Address) []*ipaddr.Address {
	list := make([]*ipaddr.Address, 0, len(addrs))
	dupChecker := make(map[AddressKey]struct{})
	for _, str := range addrs {
		addr := converter(str)
		if addr != nil {
			key := addr.ToKey()
			if _, exists := dupChecker[key]; !exists {
				dupChecker[key] = struct{}{}
				list = append(list, addr)
			}
		}
	}
	return list
}

// IPAddressPredicateAdapter has methods to supply IP, IPv4, and IPv6 addresses to a wrapped predicate function that takes Address arguments
type IPAddressPredicateAdapter struct {
	Adapted func(*ipaddr.Address) bool
}

// IPPredicate calls the wrapped predicate function with the given IP address as the argument
func (a IPAddressPredicateAdapter) IPPredicate(addr *ipaddr.IPAddress) bool {
	return a.Adapted(addr.ToAddressBase())
}

// IPv4Predicate calls the wrapped predicate function with the given IPv4 address as the argument
func (a IPAddressPredicateAdapter) IPv4Predicate(addr *ipaddr.IPv4Address) bool {
	return a.Adapted(addr.ToAddressBase())
}

// IPv6Predicate calls the wrapped predicate function with the given IPv6 address as the argument
func (a IPAddressPredicateAdapter) IPv6Predicate(addr *ipaddr.IPv6Address) bool {
	return a.Adapted(addr.ToAddressBase())
}

// IPAddressActionAdapter has methods to supply IP, IPv4, and IPv6 addresses to a wrapped consumer function that takes Address arguments
type IPAddressActionAdapter struct {
	Adapted func(*ipaddr.Address)
}

// IPAction calls the wrapped consumer function with the given IP address as the argument
func (a IPAddressActionAdapter) IPAction(addr *ipaddr.IPAddress) {
	a.Adapted(addr.ToAddressBase())
}

// IPv4Action calls the wrapped consumer function with the given IPv4 address as the argument
func (a IPAddressActionAdapter) IPv4Action(addr *ipaddr.IPv4Address) {
	a.Adapted(addr.ToAddressBase())
}

// IPv6Action calls the wrapped consumer function with the given IPv6 address as the argument
func (a IPAddressActionAdapter) IPv6Action(addr *ipaddr.IPv6Address) {
	a.Adapted(addr.ToAddressBase())
}
