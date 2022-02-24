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

import (
	"fmt"
	"github.com/seancfoley/ipaddress-go/ipaddr"
	"reflect"
	"strconv"
)

type trieTester struct {
	testBase
}

var didOneMegaTree bool

func (t trieTester) run() {

	t.testAddressCheck()
	t.partitionTest()

	sampleIPAddressTries := t.getSampleIPAddressTries()
	for _, treeAddrs := range sampleIPAddressTries {
		t.testRemove(treeAddrs)
	}
	notDoneEmptyIPv6 := true
	notDoneEmptyIPv4 := true
	for _, treeAddrs := range sampleIPAddressTries {
		ipv6Tree := ipaddr.NewIPv6AddressTrie()
		t.createIPv6SampleTree(ipv6Tree, treeAddrs)
		size := ipv6Tree.Size()
		if size > 0 || notDoneEmptyIPv6 {
			if notDoneEmptyIPv6 {
				notDoneEmptyIPv6 = size != 0
			}
			t.testIterate(ipv6Tree.ToBase())
			t.testContains(ipv6Tree.ToBase())
		}

		ipv4Tree := ipaddr.NewIPv4AddressTrie()
		t.createIPv4SampleTree(ipv4Tree, treeAddrs)
		size = ipv4Tree.Size()
		if size > 0 || notDoneEmptyIPv4 {
			if notDoneEmptyIPv4 {
				notDoneEmptyIPv4 = size != 0
			}
			t.testIterate(ipv4Tree.ToBase())
			t.testContains(ipv4Tree.ToBase())
		}
	}
	notDoneEmptyIPv6 = true
	notDoneEmptyIPv4 = true
	for _, treeAddrs := range sampleIPAddressTries {
		addrs := collect(treeAddrs, func(addrStr string) *ipaddr.Address {
			return t.createAddress(addrStr).GetAddress().ToIPv4().ToAddressBase()
		})
		size := len(addrs)
		if size > 0 || notDoneEmptyIPv4 {
			if notDoneEmptyIPv4 {
				notDoneEmptyIPv4 = size != 0
			}
			t.testAdd(ipaddr.NewIPv4AddressTrie().ToBase(), addrs)
			t.testEdges(ipaddr.NewIPv4AddressTrie().ToBase(), addrs)
			t.testMap(ipaddr.NewIPv4AddressAssociativeTrie().ToAssociativeBase(), addrs, func(i int) ipaddr.NodeValue { return i }, func(v ipaddr.NodeValue) ipaddr.NodeValue { return 2 * 1 })
		}
		addrsv6 := collect(treeAddrs, func(addrStr string) *ipaddr.Address {
			return t.createAddress(addrStr).GetAddress().ToIPv6().ToAddressBase()
		})
		size = len(addrsv6)
		if size > 0 || notDoneEmptyIPv6 {
			if notDoneEmptyIPv6 {
				notDoneEmptyIPv6 = size != 0
			}
			t.testAdd(ipaddr.NewIPv6AddressTrie().ToBase(), addrsv6)
			t.testEdges(ipaddr.NewIPv6AddressTrie().ToBase(), addrsv6)
			t.testMap(ipaddr.NewIPv6AddressAssociativeTrie().ToAssociativeBase(), addrsv6,
				func(i int) ipaddr.NodeValue { return "bla" + strconv.Itoa(i) },
				func(str ipaddr.NodeValue) ipaddr.NodeValue { return str.(string) + "foo" })
		}
	}

	notDoneEmptyMAC := true
	for _, treeAddrs := range testMACTries {
		tree := ipaddr.AddressTrie{}
		macTree := tree.ToMAC()
		t.createMACSampleTree(macTree, treeAddrs)
		size := macTree.Size()
		if size > 0 || notDoneEmptyMAC {
			if notDoneEmptyMAC {
				notDoneEmptyMAC = size != 0
			}
			t.testIterate(macTree.ToBase())
			t.testContains(macTree.ToBase())
		}
	}
	notDoneEmptyMAC = true
	for _, treeAddrs := range testMACTries {
		addrs := collect(treeAddrs, func(addrStr string) *ipaddr.Address {
			return t.createMACAddress(addrStr).GetAddress().ToAddressBase()
		})
		size := len(addrs)
		if size > 0 || notDoneEmptyIPv4 {
			if notDoneEmptyMAC {
				notDoneEmptyMAC = size != 0
			}
			tree := ipaddr.AddressTrie{}
			t.testAdd(&tree, addrs)
			tree2 := ipaddr.AddressTrie{}
			t.testEdges(&tree2, addrs)
			tree3 := ipaddr.AssociativeAddressTrie{}
			t.testMap(&tree3, addrs,
				func(i int) ipaddr.NodeValue { return i },
				func(i ipaddr.NodeValue) ipaddr.NodeValue {
					if i == nil {
						return 3
					}
					return 3 * i.(int)
				})
		}
	}
	for _, treeAddrs := range testMACTries {
		t.testRemoveMAC(treeAddrs)
	}

	cached := t.getAllCached()
	if len(cached) > 0 && !didOneMegaTree {
		didOneMegaTree = true
		ipv6Tree1 := ipaddr.NewIPv6AddressTrie()
		t.createIPv6SampleTreeAddrs(ipv6Tree1, cached)
		//fmt.Println(ipv6Tree1)
		//fmt.Printf("ipv6 mega tree has %v elements", ipv6Tree1.Size())
		t.testIterate(ipv6Tree1.ToBase())
		t.testContains(ipv6Tree1.ToBase())

		ipv4Tree1 := ipaddr.NewIPv4AddressTrie()
		t.createIPv4SampleTreeAddrs(ipv4Tree1, cached)
		//fmt.Println(ipv4Tree1)
		//fmt.Printf("ipv4 mega tree has %v elements", ipv4Tree1.Size())
		t.testIterate(ipv4Tree1.ToBase())
		t.testContains(ipv4Tree1.ToBase())
	}

	t.testString(treeOne)
	t.testString(treeTwo)

	// try deleting the root
	trieb := ipaddr.AddressTrie{}
	trie := trieb.ToIPv4()
	if trie.Size() != 0 {
		t.addFailure(newTrieFailure("unexpected size "+strconv.Itoa(trie.Size()), trie.ToBase()))
	}
	if trie.NodeSize() != 0 {
		t.addFailure(newTrieFailure("unexpected size "+strconv.Itoa(trie.NodeSize()), trie.ToBase()))
	}
	trie.Add(ipaddr.NewIPAddressString("0.0.0.0/0").GetAddress().ToIPv4())
	if trie.Size() != 1 {
		t.addFailure(newTrieFailure("unexpected size "+strconv.Itoa(trie.Size()), trie.ToBase()))
	}
	if trie.NodeSize() != 1 {
		t.addFailure(newTrieFailure("unexpected size "+strconv.Itoa(trie.Size()), trie.ToBase()))
	}
	trie.GetRoot().Remove()
	if trie.Size() != 0 {
		t.addFailure(newTrieFailure("unexpected size "+strconv.Itoa(trie.Size()), trie.ToBase()))
	}
	if trie.NodeSize() != 1 {
		t.addFailure(newTrieFailure("unexpected node size "+strconv.Itoa(trie.NodeSize()), trie.ToBase()))
	}

	trie = ipaddr.NewIPv4AddressTrie()
	trie.Add(ipaddr.NewIPAddressString("1.2.3.4").GetAddress().ToIPv4())
	trie.GetRoot().SetAdded()
	if trie.Size() != 2 {
		t.addFailure(newTrieFailure("unexpected size "+strconv.Itoa(trie.Size()), trie.ToBase()))
	}
	trie.GetRoot().Remove()
	if trie.Size() != 1 {
		t.addFailure(newTrieFailure("unexpected size "+strconv.Itoa(trie.Size()), trie.ToBase()))
	}
	if trie.NodeSize() != 2 {
		t.addFailure(newTrieFailure("unexpected node size "+strconv.Itoa(trie.NodeSize()), trie.ToBase()))
	}
	trie.Clear()
	if trie.NodeSize() != 1 {
		t.addFailure(newTrieFailure("unexpected node size "+strconv.Itoa(trie.NodeSize()), trie.ToBase()))
	}
	if trie.Size() != 0 {
		t.addFailure(newTrieFailure("unexpected size "+strconv.Itoa(trie.Size()), trie.ToBase()))
	}
	trie.GetRoot().Remove()
	if trie.NodeSize() != 1 {
		t.addFailure(newTrieFailure("unexpected node size "+strconv.Itoa(trie.NodeSize()), trie.ToBase()))
	}
	if trie.Size() != 0 {
		t.addFailure(newTrieFailure("unexpected size "+strconv.Itoa(trie.Size()), trie.ToBase()))
	}
	t.incrementTestCount()
}

type trieStrings struct {
	addrs                       []string
	treeString, addedNodeString string
}

var treeOne = trieStrings{
	addrs: []string{"1::ffff:2:3:5",
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
}

var treeTwo = trieStrings{
	addrs: []string{"ff80::/8",
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
}

func (t trieTester) testString(strs trieStrings) {
	ipv6Tree := ipaddr.NewIPv6AddressTrie()
	t.createIPv6SampleTree(ipv6Tree, strs.addrs)
	treeStr := ipv6Tree.String()
	if treeStr != strs.treeString {
		t.addFailure(newTrieFailure("trie string not right, got "+treeStr+" instead of expected "+strs.treeString, ipv6Tree.ToBase()))
	}
	addedString := ipv6Tree.AddedNodesTreeString()
	if addedString != strs.addedNodeString {
		t.addFailure(newTrieFailure("trie string not right, got "+addedString+" instead of expected "+strs.addedNodeString, ipv6Tree.ToBase()))
	}
}

func collect(addrs []string, converter func(string) *ipaddr.Address) []*ipaddr.Address {
	list := make([]*ipaddr.Address, 0, len(addrs))
	dupChecker := make(map[ipaddr.AddressKey]struct{})
	for _, str := range addrs {
		addr := converter(str)
		if addr != nil {
			key := addr.ToKey()
			if _, exists := dupChecker[*key]; !exists {
				dupChecker[*key] = struct{}{}
				list = append(list, addr)
			}
		}
	}
	return list
}

func (t trieTester) testAddressCheck() {
	addr := t.createAddress("1.2.3.4/16").GetAddress()

	t.testConvertedAddrBlock(addr.ToAddressBase(), nil)

	t.testIPAddrBlock("1.2.3.4")
	t.testIPAddrBlock("::")
	t.testNonBlock("1-3.2.3.4")

	t.testConvertedBlock("1.2.3.4-5", p31)
	t.testNonBlock("1.2.3.5-6")
	t.testConvertedBlock("1.2.3.4-7", p30)
	t.testNonBlock("::1-2:0")
	t.testNonBlock("::1-2:0/112")

	t.testConvertedBlock("::0-3:0/112", p110)
	t.testIPAddrBlock("::/64")
	t.testIPAddrBlock("1.2.0.0/16")

	mac := t.createMACAddress("a:b:c:*:*:*").GetAddress()
	mac = mac.SetPrefixLen(48)
	t.testConvertedAddrBlock(mac.ToAddressBase(), p24)
	t.testMACAddrBlock("a:b:c:*:*:*")
	t.testNonMACBlock("a:b:c:*:2:*")
	t.testNonBlock("a:b:c:*:2:*") // passes null into checkBlockOrAddress
	t.testMACAddrBlock("a:b:c:1:2:3")
}

func (t trieTester) testConvertedBlock(str string, expectedPrefLen ipaddr.PrefixLen) {
	addr := t.createAddress(str).GetAddress()
	t.testConvertedAddrBlock(addr.ToAddressBase(), expectedPrefLen)
}

func (t trieTester) testConvertedAddrBlock(addr *ipaddr.Address, expectedPrefLen ipaddr.PrefixLen) {
	result := addr.ToSinglePrefixBlockOrAddress()
	if result == nil {
		t.addFailure(newAddrFailure("unexpectedly got no single block or address for "+addr.String(), addr))
	}
	if !addr.Equal(result) && !result.GetPrefixLen().Equal(expectedPrefLen) {
		t.addFailure(newAddrFailure("unexpectedly got wrong pref len "+result.GetPrefixLen().String()+" not "+expectedPrefLen.String(), addr))
	}
}

func (t trieTester) testNonBlock(str string) {
	addr := t.createAddress(str).GetAddress()
	result := addr.ToSinglePrefixBlockOrAddress()
	if result != nil {
		t.addFailure(newIPAddrFailure("unexpectedly got a single block or address for "+addr.String(), addr))
	}
}

func (t trieTester) testNonMACBlock(str string) {
	addr := t.createMACAddress(str).GetAddress()
	result := addr.ToSinglePrefixBlockOrAddress()
	if result != nil {
		t.addFailure(newMACAddrFailure("unexpectedly got a single block or address for "+addr.String(), addr))
	}
}

func (t trieTester) testIPAddrBlock(str string) {
	addr := t.createAddress(str).GetAddress()
	result := addr.ToSinglePrefixBlockOrAddress()
	if result != addr {
		t.addFailure(newIPAddrFailure("unexpectedly got different address "+result.String()+" for "+addr.String(), addr))
	}
}

func (t trieTester) testMACAddrBlock(str string) {
	addr := t.createMACAddress(str).GetAddress()
	result := addr.ToSinglePrefixBlockOrAddress()
	if result != addr {
		t.addFailure(newMACAddrFailure("unexpectedly got different address "+result.String()+" for "+addr.String(), addr))
	}
}

func (t trieTester) partitionTest() {
	addrs := "1.2.1-15.*"
	trie := ipaddr.NewIPv4AddressTrie()
	addr := t.createAddress(addrs).GetAddress()
	t.partitionForTrie(trie.ToBase(), addr)
}

func (t trieTester) partitionForTrie(trie *ipaddr.AddressTrie, subnet *ipaddr.IPAddress) {
	ipaddr.PartitionIPWithSingleBlockSize(subnet).PredicateForEach(ipaddr.IPAddressPredicateAdapter{trie.Add}.IPPredicate)
	if trie.Size() != 15 {
		t.addFailure(newTrieFailure("partition size unexpected "+strconv.Itoa(trie.Size())+", expected 15", trie.Clone()))
	}
	// TODO LATER this one needs generics
	//Map<T, TrieNode<T>> all = ipaddr.PartitionIPWithSingleBlockSize(subnet).ApplyForEach(trie::getAddedNode);
	//if(all.size() != 15) {
	//	addFailure("map size unexpected " + trie.size() + ", expected " + 15, trie);
	//}
	//HashMap<T, TrieNode<T>> all2 = new HashMap<>();
	all2 := make(map[*ipaddr.IPAddress]*ipaddr.AddressTrieNode)
	ipaddr.PartitionIPWithSingleBlockSize(subnet).ForEach(func(addr *ipaddr.IPAddress) {
		node := trie.GetAddedNode(addr.ToAddressBase())
		all2[addr] = node
	})
	//TODO LATER needs generics
	//if(!reflect.DeepEqual(all,all2) {
	//	t.addFailure("maps not equal " + all + " and " + all2, trie);
	//}
	trie.Clear()
	ipaddr.PartitionIPWithSpanningBlocks(subnet).PredicateForEach(ipaddr.IPAddressPredicateAdapter{trie.Add}.IPPredicate)
	if trie.Size() != 4 {
		t.addFailure(newTrieFailure("partition size unexpected "+strconv.Itoa(trie.Size())+", expected 4", trie.Clone()))
	}
	trie.Clear()
	ipaddr.PartitionIPWithSingleBlockSize(subnet).PredicateForEach(ipaddr.IPAddressPredicateAdapter{trie.Add}.IPPredicate)
	ipaddr.PartitionIPWithSpanningBlocks(subnet).PredicateForEach(ipaddr.IPAddressPredicateAdapter{trie.Add}.IPPredicate)
	if trie.Size() != 18 {
		t.addFailure(newTrieFailure("partition size unexpected "+strconv.Itoa(trie.Size())+", expected 18", trie.Clone()))
	}
	allAreThere := ipaddr.PartitionIPWithSingleBlockSize(subnet).PredicateForEach(ipaddr.IPAddressPredicateAdapter{trie.Contains}.IPPredicate)
	allAreThere2 := ipaddr.PartitionIPWithSpanningBlocks(subnet).PredicateForEach(ipaddr.IPAddressPredicateAdapter{trie.Contains}.IPPredicate)
	if !(allAreThere && allAreThere2) {
		t.addFailure(newTrieFailure("partition contains check failing", trie))
	}
	t.incrementTestCount()
}

func (t trieTester) getSampleIPAddressTries() [][]string {
	if !t.fullTest {
		return testIPAddressTries
	}
	var oneMore []string
	for _, tree := range testIPAddressTries {
		for _, addr := range tree {
			oneMore = append(oneMore, addr)
		}
	}
	return append(testIPAddressTries, oneMore)
}

func (t trieTester) testRemove(addrs []string) {
	ipv6Tree := ipaddr.NewIPv6AddressTrie()
	ipv4Tree := ipaddr.NewIPv4AddressTrie()

	t.testRemoveAddrs(ipv6Tree.ToBase(), addrs, func(addrStr string) *ipaddr.Address {
		return t.createAddress(addrStr).GetAddress().ToIPv6().ToAddressBase()
	})
	t.testRemoveAddrs(ipv4Tree.ToBase(), addrs, func(addrStr string) *ipaddr.Address {
		return t.createAddress(addrStr).GetAddress().ToIPv4().ToAddressBase()
	})

	// reverse the address order
	var addrs2 = make([]string, len(addrs))
	for i := range addrs {
		addrs2[len(addrs2)-i-1] = addrs[i]
	}

	// both trees should be empty now
	t.testRemoveAddrs(ipv6Tree.ToBase(), addrs, func(addrStr string) *ipaddr.Address {
		return t.createAddress(addrStr).GetAddress().ToIPv6().ToAddressBase()
	})
	t.testRemoveAddrs(ipv4Tree.ToBase(), addrs, func(addrStr string) *ipaddr.Address {
		return t.createAddress(addrStr).GetAddress().ToIPv4().ToAddressBase()
	})
}

func (t trieTester) testRemoveMAC(addrs []string) {
	tree := ipaddr.AddressTrie{}
	t.testRemoveAddrs(&tree, addrs, func(addrStr string) *ipaddr.Address {
		return t.createMACAddress(addrStr).GetAddress().ToAddressBase()
	})

	// reverse the address order
	var addrs2 = make([]string, len(addrs))
	for i := range addrs {
		addrs2[len(addrs2)-i-1] = addrs[i]
	}

	// tree should be empty now
	t.testRemoveAddrs(&tree, addrs, func(addrStr string) *ipaddr.Address {
		return t.createMACAddress(addrStr).GetAddress().ToAddressBase()
	})
	t.incrementTestCount()
}

func (t trieTester) testRemoveAddrs(tree *ipaddr.AddressTrie, addrs []string, converter func(addrStr string) *ipaddr.Address) {
	count := 0
	var list []*ipaddr.Address
	dupChecker := make(map[ipaddr.AddressKey]struct{})
	for _, str := range addrs {
		addr := converter(str)
		if addr != nil {
			key := addr.ToKey()
			if _, exists := dupChecker[*key]; !exists {
				dupChecker[*key] = struct{}{}
				list = append(list, addr)
				count++
				tree.Add(addr)
			}
		}
	}
	t.testRemoveAddrsConverted(tree, count, list)
}

func (t trieTester) testRemoveAddrsConverted(tree *ipaddr.AddressTrie, count int, addrs []*ipaddr.Address) {
	tree2 := tree.Clone()
	tree3 := tree2.Clone()
	tree4 := tree2.Clone()
	tree5 := tree4.Clone()
	tree5.Clear()
	tree5.AddTrie(tree4.GetRoot())
	nodeSize4 := tree4.NodeSize()
	if tree4.Size() != count {
		t.addFailure(newTrieFailure("trie size not right, got "+strconv.Itoa(tree4.Size())+" instead of expected "+strconv.Itoa(count), tree4))
	}
	tree4.Clear()
	if tree4.Size() != 0 {
		t.addFailure(newTrieFailure("trie size not zero, got "+strconv.Itoa(tree4.Size())+" after clearing trie", tree4))
	}
	if tree4.NodeSize() != 1 {
		if tree4.NodeSize() != 0 || tree4.GetRoot() != nil {
			t.addFailure(newTrieFailure("node size not 1, got "+strconv.Itoa(tree4.NodeSize())+" after clearing trie", tree4))
		}
	}
	if tree5.Size() != count {
		t.addFailure(newTrieFailure("trie size not right, got "+strconv.Itoa(tree5.Size())+" instead of expected "+strconv.Itoa(count), tree5))
	}
	if tree5.NodeSize() != nodeSize4 {
		t.addFailure(newTrieFailure("trie size not right, got "+strconv.Itoa(tree5.Size())+" instead of expected "+strconv.Itoa(nodeSize4), tree5))
	}
	origSize := tree.Size()
	origNodeSize := tree.NodeSize()
	size := origSize
	nodeSize := origNodeSize
	iterator := tree.NodeIterator(true)
	for iterator.HasNext() {
		node := iterator.Next()
		iterator.Remove()
		newSize := tree.Size()
		if size-1 != newSize {
			t.addFailure(newTrieFailure("trie size mismatch, expected "+strconv.Itoa(size-1)+" got "+strconv.Itoa(newSize)+" when removing node "+node.String(), tree))
		}
		size = newSize
		newSize = tree.NodeSize()
		if newSize > nodeSize {
			t.addFailure(newTrieFailure("node size mismatch, expected smaller than "+strconv.Itoa(nodeSize)+" got "+strconv.Itoa(newSize)+" when removing node "+node.String(), tree))
		}
		nodeSize = newSize
	}

	if tree.Size() != 0 || !tree.IsEmpty() {
		t.addFailure(newTrieFailure("trie size not zero, got "+strconv.Itoa(tree.Size())+" after clearing trie", tree))
	}
	if tree.NodeSize() != 1 {
		if tree.NodeSize() != 0 || tree.GetRoot() != nil {
			t.addFailure(newTrieFailure("node size not 1, got "+strconv.Itoa(tree.NodeSize())+" after clearing trie", tree))
		}
	}

	size = origSize
	nodeSize = origNodeSize

	// now remove by order from array addrs[]
	for _, addr := range addrs {
		if addr != nil {
			tree2.Remove(addr)
			newSize := tree2.Size()
			if size-1 != newSize {
				t.addFailure(newTrieFailure("trie size mismatch, expected "+strconv.Itoa(size-1)+" got "+strconv.Itoa(newSize), tree2))
			}
			size = newSize
			newSize = tree2.NodeSize()
			if newSize > nodeSize {
				t.addFailure(newTrieFailure("node size mismatch, expected smaller than "+strconv.Itoa(nodeSize)+" got "+strconv.Itoa(newSize), tree2))
			}
			nodeSize = newSize
		}
	}

	if tree2.Size() != 0 || !tree2.IsEmpty() {
		t.addFailure(newTrieFailure("trie size not zero, got "+strconv.Itoa(tree2.Size())+" after clearing trie", tree2))
	}
	if tree2.NodeSize() != 1 {
		if tree2.NodeSize() != 0 || tree2.GetRoot() != nil {
			t.addFailure(newTrieFailure("node size not 1, got "+strconv.Itoa(tree2.NodeSize())+" after clearing trie", tree2))
		}
	}
	// now remove full subtrees at once
	addressesRemoved := 0

	for _, addr := range addrs {
		if addr != nil {
			node := tree3.GetAddedNode(addr)
			nodeCountToBeRemoved := 0
			if node != nil {
				nodeCountToBeRemoved = 1
				lowerNode := node.GetLowerSubNode()
				if lowerNode != nil {
					nodeCountToBeRemoved += lowerNode.Size()
				}
				upperNode := node.GetUpperSubNode()
				if upperNode != nil {
					nodeCountToBeRemoved += upperNode.Size()
				}

			}
			preRemovalSize := tree3.Size()

			tree3.RemoveElementsContainedBy(addr)
			addressesRemoved++

			// we cannot check for smaller tree or node size because many elements might have been already erased
			newSize := tree3.Size()
			if newSize != preRemovalSize-nodeCountToBeRemoved {
				t.addFailure(newTrieFailure("removal size mismatch, expected to remove "+strconv.Itoa(nodeCountToBeRemoved)+" but removed "+strconv.Itoa(preRemovalSize-newSize), tree3))
			}
			if newSize > origSize-addressesRemoved {
				t.addFailure(newTrieFailure("trie size mismatch, expected smaller than "+strconv.Itoa(origSize-addressesRemoved)+" got "+strconv.Itoa(newSize), tree3))
			}
			newSize = tree3.NodeSize()
			if newSize > origNodeSize-addressesRemoved && newSize > 1 {
				t.addFailure(newTrieFailure("node size mismatch, expected smaller than "+strconv.Itoa(origSize-addressesRemoved)+" got "+strconv.Itoa(newSize), tree3))
			}
		}
	}
	if tree3.Size() != 0 || !tree3.IsEmpty() {
		t.addFailure(newTrieFailure("trie size not zero, got "+strconv.Itoa(tree3.Size())+" after clearing trie", tree3))
	}
	if tree3.NodeSize() != 1 {
		if tree3.NodeSize() != 0 || tree3.GetRoot() != nil {
			t.addFailure(newTrieFailure("node size not 1, got "+strconv.Itoa(tree3.NodeSize())+" after clearing trie", tree3))
		}
	}

	t.incrementTestCount()
}

func (t trieTester) createMACSampleTree(tree *ipaddr.MACAddressTrie, addrs []string) {
	for _, addr := range addrs {
		addressStr := t.createMACAddress(addr)
		address := addressStr.GetAddress()
		tree.Add(address)
	}
}

func (t trieTester) createIPv6SampleTree(tree *ipaddr.IPv6AddressTrie, addrs []string) {
	for _, addr := range addrs {
		addressStr := t.createAddress(addr)
		if addressStr.IsIPv6() {
			address := addressStr.GetAddress().ToIPv6()
			tree.Add(address)
		}
	}
}

func (t trieTester) createIPv6SampleTreeAddrs(tree *ipaddr.IPv6AddressTrie, addrs []*ipaddr.IPAddress) {
	for _, addr := range addrs {
		if addr.IsIPv6() {
			addr = addr.ToSinglePrefixBlockOrAddress()
			if addr != nil {
				address := addr.ToIPv6()
				tree.Add(address)
			}
		}
	}
}

func (t trieTester) createIPv4SampleTreeAddrs(tree *ipaddr.IPv4AddressTrie, addrs []*ipaddr.IPAddress) {
	for _, addr := range addrs {
		if addr.IsIPv4() {
			addr = addr.ToSinglePrefixBlockOrAddress()
			if addr != nil {
				address := addr.ToIPv4()
				tree.Add(address)
			}
		}
	}
}

func (t trieTester) createIPv4SampleTree(tree *ipaddr.IPv4AddressTrie, addrs []string) {
	for _, addr := range addrs {
		addressStr := t.createAddress(addr)
		if addressStr.IsIPv4() {
			address := addressStr.GetAddress().ToIPv4()
			tree.Add(address)
		}
	}
}

//<R extends AddressTrie<T>, T extends Address>
func (t trieTester) testIterationContainment(tree *ipaddr.AddressTrie) {
	t.testIterationContainmentTree(tree, func(trie *ipaddr.AddressTrie) ipaddr.CachingAddressTrieNodeIterator {
		return trie.BlockSizeCachingAllNodeIterator()
	}, false)
	t.testIterationContainmentTree(tree, func(trie *ipaddr.AddressTrie) ipaddr.CachingAddressTrieNodeIterator {
		return trie.ContainingFirstAllNodeIterator(true)
	}, false /* added only */)
	t.testIterationContainmentTree(tree, func(trie *ipaddr.AddressTrie) ipaddr.CachingAddressTrieNodeIterator {
		return trie.ContainingFirstAllNodeIterator(false)
	}, false /* added only */)
	t.testIterationContainmentTree(tree, func(trie *ipaddr.AddressTrie) ipaddr.CachingAddressTrieNodeIterator {
		return trie.ContainingFirstIterator(true)
	}, true /* added only */)
	t.testIterationContainmentTree(tree, func(trie *ipaddr.AddressTrie) ipaddr.CachingAddressTrieNodeIterator {
		return trie.ContainingFirstIterator(false)
	}, true /* added only */)
}

func (t trieTester) testIterationContainmentTree(
	trie *ipaddr.AddressTrie,
	iteratorFunc func(addressTrie *ipaddr.AddressTrie) ipaddr.CachingAddressTrieNodeIterator,
	addedNodesOnly bool) {
	iterator := iteratorFunc(trie)
	for iterator.HasNext() {
		next := iterator.Next()
		nextAddr := next.GetKey()
		var parentPrefix ipaddr.PrefixLen
		parent := next.GetParent()
		skipCheck := false
		if parent != nil {
			parentPrefix = parent.GetKey().GetPrefixLen()
			if addedNodesOnly {
				if !parent.IsAdded() {
					skipCheck = true
				} else {
					parentPrefix = parent.GetKey().GetPrefixLen()
				}
			}
		}
		cached := iterator.GetCached()
		if !skipCheck { //panic: interface conversion: tree.C is nil, not *ipaddr.PrefixBitCount
			if cached == nil {
				if parentPrefix != nil {
					t.addFailure(newTrieFailure("mismatched prefix for "+next.String()+", cached is "+fmt.Sprint(iterator.GetCached())+" and expected value is "+parentPrefix.String(), trie))
				}
			} else if !cached.(ipaddr.PrefixLen).Equal(parentPrefix) {
				t.addFailure(newTrieFailure("mismatched prefix for "+next.String()+", cached is "+fmt.Sprint(iterator.GetCached())+" and expected value is "+parentPrefix.String(), trie))
			}
		}
		prefLen := nextAddr.GetPrefixLen()
		iterator.CacheWithLowerSubNode(prefLen)
		iterator.CacheWithUpperSubNode(prefLen)

	}
	t.incrementTestCount()
}

func (t trieTester) testIterate(tree *ipaddr.AddressTrie) {

	t.testIteratorRem(tree, func(trie *ipaddr.AddressTrie) ipaddr.AddressTrieNodeIteratorRem {
		return trie.BlockSizeNodeIterator(true)
	}, (*ipaddr.AddressTrie).Size)
	t.testIteratorRem(tree, func(trie *ipaddr.AddressTrie) ipaddr.AddressTrieNodeIteratorRem {
		return trie.BlockSizeAllNodeIterator(true)
	}, (*ipaddr.AddressTrie).NodeSize)
	t.testIteratorRem(tree, func(trie *ipaddr.AddressTrie) ipaddr.AddressTrieNodeIteratorRem {
		return trie.BlockSizeNodeIterator(false)
	}, (*ipaddr.AddressTrie).Size)
	t.testIteratorRem(tree, func(trie *ipaddr.AddressTrie) ipaddr.AddressTrieNodeIteratorRem {
		return trie.BlockSizeAllNodeIterator(false)
	}, (*ipaddr.AddressTrie).NodeSize)

	t.testIteratorRem(tree, func(trie *ipaddr.AddressTrie) ipaddr.AddressTrieNodeIteratorRem {
		return trie.BlockSizeCachingAllNodeIterator()
	}, (*ipaddr.AddressTrie).NodeSize)

	t.testIteratorRem(tree, func(trie *ipaddr.AddressTrie) ipaddr.AddressTrieNodeIteratorRem { return trie.NodeIterator(true) }, (*ipaddr.AddressTrie).Size)
	t.testIteratorRem(tree, func(trie *ipaddr.AddressTrie) ipaddr.AddressTrieNodeIteratorRem { return trie.AllNodeIterator(true) }, (*ipaddr.AddressTrie).NodeSize)
	t.testIteratorRem(tree, func(trie *ipaddr.AddressTrie) ipaddr.AddressTrieNodeIteratorRem { return trie.NodeIterator(false) }, (*ipaddr.AddressTrie).Size)
	t.testIteratorRem(tree, func(trie *ipaddr.AddressTrie) ipaddr.AddressTrieNodeIteratorRem { return trie.AllNodeIterator(false) }, (*ipaddr.AddressTrie).NodeSize)

	t.testIteratorRem(tree, func(trie *ipaddr.AddressTrie) ipaddr.AddressTrieNodeIteratorRem {
		return trie.ContainedFirstIterator(true)
	}, (*ipaddr.AddressTrie).Size)
	t.testIterator(tree, func(trie *ipaddr.AddressTrie) ipaddr.AddressTrieNodeIterator {
		return trie.ContainedFirstAllNodeIterator(true)
	}, (*ipaddr.AddressTrie).NodeSize)
	t.testIteratorRem(tree, func(trie *ipaddr.AddressTrie) ipaddr.AddressTrieNodeIteratorRem {
		return trie.ContainedFirstIterator(false)
	}, (*ipaddr.AddressTrie).Size)
	t.testIterator(tree, func(trie *ipaddr.AddressTrie) ipaddr.AddressTrieNodeIterator {
		return trie.ContainedFirstAllNodeIterator(false)
	}, (*ipaddr.AddressTrie).NodeSize)

	t.testIteratorRem(tree, func(trie *ipaddr.AddressTrie) ipaddr.AddressTrieNodeIteratorRem {
		return trie.ContainingFirstIterator(true)
	}, (*ipaddr.AddressTrie).Size)
	t.testIteratorRem(tree, func(trie *ipaddr.AddressTrie) ipaddr.AddressTrieNodeIteratorRem {
		return trie.ContainingFirstAllNodeIterator(true)
	}, (*ipaddr.AddressTrie).NodeSize)
	t.testIteratorRem(tree, func(trie *ipaddr.AddressTrie) ipaddr.AddressTrieNodeIteratorRem {
		return trie.ContainingFirstIterator(false)
	}, (*ipaddr.AddressTrie).Size)
	t.testIteratorRem(tree, func(trie *ipaddr.AddressTrie) ipaddr.AddressTrieNodeIteratorRem {
		return trie.ContainingFirstAllNodeIterator(false)
	}, (*ipaddr.AddressTrie).NodeSize)

	t.testIterationContainment(tree)

	t.incrementTestCount()
}

func (t trieTester) testIteratorRem(
	trie *ipaddr.AddressTrie,
	iteratorFunc func(*ipaddr.AddressTrie) ipaddr.AddressTrieNodeIteratorRem,
	countFunc func(*ipaddr.AddressTrie) int) {
	// iterate the tree, confirm the size by counting
	// clone the trie, iterate again, but remove each time, confirm the size
	// confirm trie is empty at the end

	if trie.Size() > 0 {
		clonedTrie := trie.Clone()
		node := clonedTrie.FirstNode()
		toAdd := node.GetKey()
		node.Remove()
		modIterator := iteratorFunc(clonedTrie)
		mod := clonedTrie.Size() / 2
		i := 0
		shouldThrow := false

		func() {
			defer func() {
				if r := recover(); r != nil {
					if !shouldThrow {
						t.addFailure(newTrieFailure("unexpected throw ", clonedTrie))
					}
				}
			}()

			for modIterator.HasNext() {
				i++
				if i == mod {
					shouldThrow = true
					clonedTrie.Add(toAdd)
				}
				modIterator.Next()
				if shouldThrow {
					t.addFailure(newTrieFailure("expected panic ", clonedTrie))
					break
				}
			}
		}()
	}

	firstTime := true
	for {
		expectedSize := countFunc(trie)
		actualSize := 0
		set := make(map[ipaddr.AddressKey]struct{})
		iterator := iteratorFunc(trie)
		for iterator.HasNext() {
			next := iterator.Next()
			nextAddr := next.GetKey()
			set[*nextAddr.ToKey()] = struct{}{}
			actualSize++
			if !firstTime {
				func() {
					defer func() {
						if r := recover(); r != nil {
							t.addFailure(newTrieFailure("removal "+next.String()+" should be supported", trie))
						}
					}()
					iterator.Remove()
					if trie.Contains(nextAddr) {
						t.addFailure(newTrieFailure("after removal "+next.String()+" still in trie ", trie))
					}
				}()
			} else {
				if next.IsAdded() {
					if !trie.Contains(nextAddr) {
						t.addFailure(newTrieFailure("after iteration "+next.String()+" not in trie ", trie))
					} else if trie.GetAddedNode(nextAddr) == nil {
						t.addFailure(newTrieFailure("after iteration address node for "+nextAddr.String()+" not in trie ", trie))
					}
				} else {
					if trie.Contains(nextAddr) {
						t.addFailure(newTrieFailure("non-added node "+next.String()+" in trie ", trie))
					} else if trie.GetNode(nextAddr) == nil {
						t.addFailure(newTrieFailure("after iteration address node for "+nextAddr.String()+" not in trie ", trie))
					} else if trie.GetAddedNode(nextAddr) != nil {
						t.addFailure(newTrieFailure("after iteration non-added node for "+nextAddr.String()+" added in trie ", trie))
					}
				}
			}
		}
		if len(set) != expectedSize {
			t.addFailure(newTrieFailure("set count was "+strconv.Itoa(len(set))+" instead of expected "+strconv.Itoa(expectedSize), trie))
		} else if actualSize != expectedSize {
			t.addFailure(newTrieFailure("count was "+strconv.Itoa(actualSize)+" instead of expected "+strconv.Itoa(expectedSize), trie))
		}
		trie = trie.Clone()
		if !firstTime {
			break
		}
		firstTime = false
	}
	if !trie.IsEmpty() {
		t.addFailure(newTrieFailure("trie not empty, size "+strconv.Itoa(trie.Size())+" after removing everything", trie))
	} else if trie.NodeSize() > 1 {
		t.addFailure(newTrieFailure("trie node size not 1, "+strconv.Itoa(trie.NodeSize())+" after removing everything", trie))
	} else if trie.Size() > 0 {
		t.addFailure(newTrieFailure("trie size not 0, "+strconv.Itoa(trie.Size())+" after removing everything", trie))
	}
	t.incrementTestCount()
}

func (t trieTester) testIterator(
	trie *ipaddr.AddressTrie,
	iteratorFunc func(*ipaddr.AddressTrie) ipaddr.AddressTrieNodeIterator,
	countFunc func(*ipaddr.AddressTrie) int) {
	// iterate the tree, confirm the size by counting
	// clone the trie, iterate again, but remove each time, confirm the size
	// confirm trie is empty at the end

	if trie.Size() > 0 {
		clonedTrie := trie.Clone()
		node := clonedTrie.FirstNode()
		toAdd := node.GetKey()
		node.Remove()
		modIterator := iteratorFunc(clonedTrie)
		mod := clonedTrie.Size() / 2
		i := 0
		shouldThrow := false

		func() {
			defer func() {
				if r := recover(); r != nil {
					if !shouldThrow {
						t.addFailure(newTrieFailure("unexpected throw ", clonedTrie))
					}
				}
			}()
			for modIterator.HasNext() {
				i++
				if i == mod {
					shouldThrow = true
					clonedTrie.Add(toAdd)
				}
				modIterator.Next()
				if shouldThrow {
					t.addFailure(newTrieFailure("expected panic ", clonedTrie))
					break
				}
			}
		}()
	}

	expectedSize := countFunc(trie)
	actualSize := 0
	set := make(map[ipaddr.AddressKey]struct{})
	iterator := iteratorFunc(trie)
	for iterator.HasNext() {
		next := iterator.Next()
		nextAddr := next.GetKey()
		set[*nextAddr.ToKey()] = struct{}{}
		actualSize++

		if next.IsAdded() {
			if !trie.Contains(nextAddr) {
				t.addFailure(newTrieFailure("after iteration "+next.String()+" not in trie ", trie))
			} else if trie.GetAddedNode(nextAddr) == nil {
				t.addFailure(newTrieFailure("after iteration address node for "+nextAddr.String()+" not in trie ", trie))
			}
		} else {
			if trie.Contains(nextAddr) {
				t.addFailure(newTrieFailure("non-added node "+next.String()+" in trie ", trie))
			} else if trie.GetNode(nextAddr) == nil {
				t.addFailure(newTrieFailure("after iteration address node for "+nextAddr.String()+" not in trie ", trie))
			} else if trie.GetAddedNode(nextAddr) != nil {
				t.addFailure(newTrieFailure("after iteration non-added node for "+nextAddr.String()+" added in trie ", trie))
			}
		}
	}
	if len(set) != expectedSize {
		t.addFailure(newTrieFailure("set count was "+strconv.Itoa(len(set))+" instead of expected "+strconv.Itoa(expectedSize), trie))
	} else if actualSize != expectedSize {
		t.addFailure(newTrieFailure("count was "+strconv.Itoa(actualSize)+" instead of expected "+strconv.Itoa(expectedSize), trie))
	}
	t.incrementTestCount()
}

func (t trieTester) testContains(trie *ipaddr.AddressTrie) {
	if trie.Size() > 0 {
		last := trie.GetAddedNode(trie.LastAddedNode().GetKey())
		if !trie.Contains(last.GetKey()) {
			t.addFailure(newTrieFailure("failure "+last.String()+" not in trie ", trie))
		}
		last.Remove()
		if trie.Contains(last.GetKey()) {
			t.addFailure(newTrieFailure("failure "+last.String()+" is in trie ", trie))
		}
		trie.Add(last.GetKey())
		if !trie.Contains(last.GetKey()) {
			t.addFailure(newTrieFailure("failure "+last.String()+" not in trie ", trie))
		}
	}
	iterator := trie.AllNodeIterator(true)
	for iterator.HasNext() {
		next := iterator.Next()
		nextAddr := next.GetKey()
		if next.IsAdded() {
			if !trie.Contains(nextAddr) {
				t.addFailure(newTrieFailure("after iteration "+next.String()+" not in trie ", trie))
			} else if trie.GetAddedNode(nextAddr) == nil {
				t.addFailure(newTrieFailure("after iteration address node for "+nextAddr.String()+" not in trie ", trie))
			}
		} else {
			if trie.Contains(nextAddr) {
				t.addFailure(newTrieFailure("non-added node "+next.String()+" in trie ", trie))
			} else if trie.GetNode(nextAddr) == nil {
				t.addFailure(newTrieFailure("after iteration address node for "+nextAddr.String()+" not in trie ", trie))
			} else if trie.GetAddedNode(nextAddr) != nil {
				t.addFailure(newTrieFailure("after iteration non-added node for "+nextAddr.String()+" added in trie ", trie))
			}
		}
		parent := next.GetParent()
		var parentPrefLen ipaddr.PrefixLen
		if parent != nil {
			parentKey := parent.GetKey()
			parentPrefLen = parentKey.GetPrefixLen()
		} else {
			parentPrefLen = ipaddr.ToPrefixLen(0)
		}
		prefLen := nextAddr.GetPrefixLen()
		var halfwayAddr *ipaddr.Address
		var halfway ipaddr.BitCount
		if prefLen == nil {
			prefLen = ipaddr.ToPrefixLen(nextAddr.GetBitCount())
		}
		halfway = parentPrefLen.Len() + ((prefLen.Len() - parentPrefLen.Len()) >> 1)
		halfwayAddr = nextAddr.SetPrefixLen(halfway).ToPrefixBlock()

		halfwayIsParent := parent != nil && parentPrefLen.Len() == halfway
		containedBy := trie.ElementsContainedBy(halfwayAddr)
		if halfwayIsParent {
			if containedBy != parent {
				t.addFailure(newTrieFailure("containedBy is "+containedBy.String()+" for address "+halfwayAddr.String()+" instead of expected "+parent.String(), trie))
			}
		} else {
			if containedBy != next {
				t.addFailure(newTrieFailure("containedBy is "+containedBy.String()+" for address "+halfwayAddr.String()+" instead of expected "+next.String(), trie))
			}
		}
		lpm := trie.LongestPrefixMatch(halfwayAddr)
		smallestContaining := trie.LongestPrefixMatchNode(halfwayAddr)
		containing := trie.ElementsContaining(halfwayAddr)
		elementsContains := trie.ElementContains(halfwayAddr)
		addedParent := parent
		for addedParent != nil && !addedParent.IsAdded() {
			addedParent = addedParent.GetParent()
		}
		if addedParent == nil && prefLen.Len() == 0 && next.IsAdded() {
			addedParent = next
		}
		if addedParent == nil {
			if containing != nil || lpm != nil {
				t.addFailure(newTrieFailure("containing is "+containing.String()+" for address "+halfwayAddr.String()+" instead of expected nil", trie))
			} else if elementsContains {
				t.addFailure(newTrieFailure("containing is true for address "+halfwayAddr.String()+" instead of expected false", trie))
			}
		} else {
			lastContaining := containing
			for lastContaining != nil {
				lower := lastContaining.GetLowerSubNode()
				if lower != nil {
					lastContaining = lower
				} else {
					upper := lastContaining.GetUpperSubNode()
					if upper != nil {
						lastContaining = upper
					} else {
						break
					}
				}
			}
			if lastContaining == nil || !lastContaining.Equal(addedParent) {
				t.addFailure(newTrieFailure("containing ends with "+lastContaining.String()+" for address "+halfwayAddr.String()+" instead of expected "+addedParent.String(), trie))
			} else if !lastContaining.Equal(smallestContaining) {
				t.addFailure(newTrieFailure("containing ends with "+lastContaining.String()+" for address "+halfwayAddr.String()+" instead of expected smallest containing "+smallestContaining.String(), trie))
			} else if lastContaining.GetKey() != lpm {
				t.addFailure(newTrieFailure("containing ends with addr "+lastContaining.GetKey().String()+" for address "+halfwayAddr.String()+" instead of expected "+lpm.String(), trie))
			}
			if !elementsContains {
				t.addFailure(newTrieFailure("containing is false for address "+halfwayAddr.String()+" instead of expected true", trie))
			}
		}
	}
	t.incrementTestCount()
}

func (t trieTester) testEdges(trie *ipaddr.AddressTrie, addrs []*ipaddr.Address) {
	trie2 := trie.Clone()
	for _, addr := range addrs {
		trie.Add(addr)
	}
	i := 0
	ordered := make([]*ipaddr.AddressTrieNode, 0, len(addrs))
	iter := trie.Iterator()
	for addr := iter.Next(); addr != nil; addr = iter.Next() {
		if i%2 == 0 {
			trie2.Add(addr)
		}
		i++
		ordered = append(ordered, trie.GetAddedNode(addr))
	}
	i = 0
	nodeIter := trie.NodeIterator(true)
	treeSize := trie.Size()
	iter = trie.Iterator()
	for addr := iter.Next(); addr != nil; addr = iter.Next() {
		node := nodeIter.Next()
		floor := trie2.FloorAddedNode(addr)
		lower := trie2.LowerAddedNode(addr)
		ceiling := trie2.CeilingAddedNode(addr)
		higher := trie2.HigherAddedNode(addr)
		if i == 0 {
			if node != trie.FirstAddedNode() {
				t.addFailure(newTrieFailure("wrong first, got "+trie.FirstAddedNode().String()+" not "+node.String(), trie))
			}
		} else if i == treeSize-1 {
			if node != trie.LastAddedNode() {
				t.addFailure(newTrieFailure("wrong last, got "+trie.LastAddedNode().String()+" not "+node.String(), trie))
			}
		}
		if i%2 == 0 {
			// in the second trie
			if !floor.Equal(node) {
				t.addFailure(newTrieFailure("wrong floor, got "+floor.String()+" not "+node.String(), trie))
			} else if !ceiling.Equal(node) {
				t.addFailure(newTrieFailure("wrong ceiling, got "+ceiling.String()+" not "+node.String(), trie))
			} else {
				if i > 0 {
					expected := ordered[i-2]
					if !lower.Equal(expected) {
						t.addFailure(newTrieFailure("wrong lower, got "+lower.String()+" not "+expected.String(), trie))
					}
				} else {
					if lower != nil {
						t.addFailure(newTrieFailure("wrong lower, got "+lower.String()+" not nil", trie))
					}
				}
				if i < len(ordered)-2 {
					expected := ordered[i+2]
					if !higher.Equal(expected) {
						t.addFailure(newTrieFailure("wrong higher, got "+higher.String()+" not "+expected.String(), trie))
					}
				} else {
					if higher != nil {
						t.addFailure(newTrieFailure("wrong higher, got "+higher.String()+" not nil", trie))
					}
				}
			}
		} else {
			// not in the second trie
			if i > 0 {
				expected := ordered[i-1]
				if !lower.Equal(expected) {
					t.addFailure(newTrieFailure("wrong lower, got "+lower.String()+" not "+expected.String(), trie))
				} else if !lower.Equal(floor) {
					t.addFailure(newTrieFailure("wrong floor, got "+floor.String()+" not "+expected.String(), trie))
				}
			} else {
				if lower != nil {
					t.addFailure(newTrieFailure("wrong lower, got "+lower.String()+" not nil", trie))
				} else if floor != nil {
					t.addFailure(newTrieFailure("wrong floor, got "+floor.String()+" not nil", trie))
				}
			}
			if i < len(ordered)-1 {
				expected := ordered[i+1]
				if !higher.Equal(expected) {
					t.addFailure(newTrieFailure("wrong higher, got "+higher.String()+" not "+expected.String(), trie))
				} else if !higher.Equal(ceiling) {
					t.addFailure(newTrieFailure("wrong ceiling, got "+ceiling.String()+" not "+expected.String(), trie))
				}
			} else {
				if higher != nil {
					t.addFailure(newTrieFailure("wrong higher, got "+higher.String()+" not nil", trie))
				} else if ceiling != nil {
					t.addFailure(newTrieFailure("wrong ceiling, got "+ceiling.String()+" not nil", trie))
				}
			}
		}
		i++
	}
	t.incrementTestCount()
}

// pass in an empty trie
func (t trieTester) testAdd(trie *ipaddr.AddressTrie, addrs []*ipaddr.Address) {
	trie2 := trie.Clone()
	trie3 := trie.Clone()
	trie4 := trie.Clone()
	k := 0
	for _, addr := range addrs {
		k++
		if k%2 == 0 {
			added := trie.Add(addr)
			if !added {
				t.addFailure(newTrieFailure("trie empty, adding "+addr.String()+" should succeed ", trie))
			}
		} else {
			node := trie.AddNode(addr)
			if node == nil || !node.GetKey().Equal(addr) {
				t.addFailure(newTrieFailure("trie empty, adding "+addr.String()+" should succeed ", trie))
			}
		}
	}
	if trie.Size() != len(addrs) {
		t.addFailure(newTrieFailure("trie size incorrect: "+strconv.Itoa(trie.Size())+", not "+strconv.Itoa(len(addrs)), trie))
	}
	node := trie.GetRoot()
	i := 0
	for ; i < len(addrs)/2; i++ {
		trie2.Add(addrs[i])
	}
	for ; i < len(addrs); i++ {
		trie3.Add(addrs[i])
	}
	trie2.AddTrie(node)
	trie3.AddTrie(node)
	trie4.AddTrie(node)
	if !trie.Equal(trie2) {
		t.addFailure(newTrieFailure("tries not equal: "+trie.String()+" and "+trie2.String(), trie))
	}
	if !trie3.Equal(trie2) {
		t.addFailure(newTrieFailure("tries not equal: "+trie3.String()+" and "+trie2.String(), trie))
	}
	if !trie3.Equal(trie4) {
		t.addFailure(newTrieFailure("tries not equal: "+trie3.String()+" and "+trie4.String(), trie))
	}
	t.incrementTestCount()
}

func (t trieTester) testMap(trie *ipaddr.AssociativeAddressTrie, addrs []*ipaddr.Address,
	valueProducer func(int) ipaddr.NodeValue, mapper func(ipaddr.NodeValue) ipaddr.NodeValue) {
	// put tests
	trie2 := trie.Clone()
	trie4 := trie.Clone()
	for i, addr := range addrs {
		trie.Put(addr, valueProducer(i))
	}
	for i, addr := range addrs {
		v := trie.Get(addr)
		expected := valueProducer(i)
		if !reflect.DeepEqual(v, expected) { //reflect deep equal
			fmt.Println(trie)
			t.addFailure(newTrieFailure(fmt.Sprintf("got mismatch, got %v, not %v for %v", v, expected, addr), trie.ToBase()))
			v = trie.Get(addr)
		}
	}
	// all trie2 from now on
	trie2.PutTrie(trie.GetRoot())
	for i, addr := range addrs {
		v := trie2.Get(addr)
		expected := valueProducer(i)
		if !reflect.DeepEqual(v, expected) {
			t.addFailure(newTrieFailure(fmt.Sprintf("got mismatch, got %v, not %v", v, expected), trie2.ToBase()))
		}
		if i%2 == 0 {
			trie2.Remove(addr)
		}
	}
	if trie2.Size() != (len(addrs) >> 1) {
		t.addFailure(newTrieFailure("got size mismatch, got "+strconv.Itoa(trie2.Size())+" not "+strconv.Itoa(len(addrs)>>1), trie2.ToBase()))
	}
	trie2.PutTrie(trie.GetRoot())
	for i, addr := range addrs {
		v := trie2.Get(addr)
		expected := valueProducer(i)
		if !reflect.DeepEqual(v, expected) {
			t.addFailure(newTrieFailure(fmt.Sprintf("get mismatch, got %v, not %v", v, expected), trie2.ToBase()))
		}
	}
	if trie2.Size() != len(addrs) {
		t.addFailure(newTrieFailure("got size mismatch, got "+strconv.Itoa(trie2.Size())+" not "+strconv.Itoa(len(addrs)), trie2.ToBase()))
	}
	for i, addr := range addrs {
		if i%2 == 0 {
			b := trie2.Remove(addr)
			if !b {
				t.addFailure(newTrieFailure("remove should have succeeded", trie2.ToBase()))
			}
			b = trie2.Remove(addr)
			if b {
				t.addFailure(newTrieFailure("remove should not have succeeded", trie2.ToBase()))
			}
		}

	}
	for i, addr := range addrs {
		exists, _ := trie2.Put(addr, valueProducer(i))
		if exists != (i%2 == 0) {
			t.addFailure(newTrieFailure("putNew mismatch", trie2.ToBase()))
		}
	}
	if trie2.Size() != len(addrs) {
		t.addFailure(newTrieFailure("got size mismatch, got "+strconv.Itoa(trie.Size())+" not "+strconv.Itoa(len(addrs)), trie2.ToBase()))
	}
	for i, addr := range addrs {
		res, _ := trie2.Put(addr, valueProducer(i+1))
		if res {
			t.addFailure(newTrieFailure("putNew mismatch", trie2.ToBase()))
		}
	}
	if trie2.Size() != len(addrs) {
		t.addFailure(newTrieFailure("got size mismatch, got "+strconv.Itoa(trie.Size())+" not "+strconv.Itoa(len(addrs)), trie2.ToBase()))
	}
	i = 0
	for i, addr := range addrs {
		v := trie2.Get(addr)
		expected := valueProducer(i + 1)
		if !reflect.DeepEqual(v, expected) {
			t.addFailure(newTrieFailure(fmt.Sprintf("get mismatch, got %v, not %v", v, expected), trie.ToBase()))
		}
	}
	i = 0
	for i, addr := range addrs {
		_, v := trie2.Put(addr, valueProducer(i))
		expected := valueProducer(i + 1)
		if !reflect.DeepEqual(v, expected) {
			t.addFailure(newTrieFailure(fmt.Sprintf("get mismatch, got %v, not %v", v, expected), trie.ToBase()))
		}
		v = trie2.Get(addr)
		expected = valueProducer(i)
		if !reflect.DeepEqual(v, expected) {
			t.addFailure(newTrieFailure(fmt.Sprintf("get mismatch, got %v, not %v", v, expected), trie.ToBase()))
		}
	}

	i = 0
	k := 0
	for i, addr := range addrs {
		if i%2 == 0 {
			b := trie2.Remove(addr)
			if !b {
				t.addFailure(newTrieFailure("remove should have succeeded", trie2.ToBase()))
			}
		}
		// the reason for the (i % 8 == 1) is that the existing value is already valueProducer.apply(i),
		// so half the time we are re-adding the existing value,
		// half the time we are changing to a new value
		var value ipaddr.NodeValue
		if i%4 == 1 {
			if i%8 == 1 {
				value = valueProducer(i + 1)
			} else {
				value = valueProducer(i)
			}
		}
		node := trie2.Remap(addr, func(val ipaddr.NodeValue) ipaddr.NodeValue {
			if val == nil {
				return valueProducer(0)
			} else {
				return value
			}
		})
		if node == nil || !node.GetKey().Equal(addr) {
			t.addFailure(newTrieFailure("got unexpected return, got "+node.String(), trie2.ToBase()))
		}
		if i%2 != 0 && value == nil {
			k++
		}
	}
	if trie2.Size()+k != len(addrs) {
		t.addFailure(newTrieFailure("got size mismatch, got "+strconv.Itoa(trie2.Size())+" not "+strconv.Itoa(len(addrs)-k), trie2.ToBase()))
	}
	for i, addr := range addrs {
		v := trie2.Get(addr)
		var expected ipaddr.NodeValue
		if i%2 == 0 {
			expected = valueProducer(0)
		} else if i%4 == 1 {
			if i%8 == 1 {
				expected = valueProducer(i + 1)
			} else {
				expected = valueProducer(i)
			}
		}
		if !reflect.DeepEqual(v, expected) {
			t.addFailure(newTrieFailure(fmt.Sprintf("got mismatch, got %v, not %v", v, expected), trie.ToBase()))
		}
	}
	for _, addr := range addrs {
		trie2.RemapIfAbsent(addr, func() ipaddr.NodeValue {
			return valueProducer(1)
		}, false)
	}
	if trie2.Size() != len(addrs) {
		t.addFailure(newTrieFailure("got size mismatch, got "+strconv.Itoa(trie2.Size())+" not "+strconv.Itoa(len(addrs)), trie2.ToBase()))
	}
	for i, addr := range addrs {
		v := trie2.Get(addr)
		var expected ipaddr.NodeValue
		if i%2 == 0 {
			expected = valueProducer(0)
		} else if i%4 == 1 {
			if i%8 == 1 {
				expected = valueProducer(i + 1)
			} else {
				expected = valueProducer(i)
			}
		} else {
			// remapped
			expected = valueProducer(1)
		}
		if !reflect.DeepEqual(v, expected) {
			t.addFailure(newTrieFailure(fmt.Sprintf("get mismatch, got %v, not %v", v, expected), trie.ToBase()))
		}
	}
	if trie2.Size() != len(addrs) {
		t.addFailure(newTrieFailure("got size mismatch, got "+strconv.Itoa(trie2.Size())+" not "+strconv.Itoa(len(addrs)), trie2.ToBase()))
	}

	for i, addr := range addrs {
		if i%2 == 0 {
			trie2.GetNode(addr).Remove()
		}
	}
	if trie2.Size() != (len(addrs) >> 1) {
		t.addFailure(newTrieFailure("got size mismatch, got "+strconv.Itoa(trie2.Size())+" not "+strconv.Itoa(len(addrs)>>1), trie2.ToBase()))
	}

	for i, addr := range addrs {
		node := trie2.RemapIfAbsent(addr, func() ipaddr.NodeValue {
			return nil
		}, false)
		if (node == nil) != (i%2 == 0) {
			t.addFailure(newTrieFailure("got unexpected return, got "+node.String(), trie2.ToBase()))
		}
	}
	if trie2.Size() != (len(addrs) >> 1) {
		t.addFailure(newTrieFailure("got size mismatch, got "+strconv.Itoa(trie2.Size())+" not "+strconv.Itoa(len(addrs)>>1), trie2.ToBase()))
	}

	for _, addr := range addrs {
		node := trie2.RemapIfAbsent(addr, func() ipaddr.NodeValue {
			return nil
		}, true)
		if node == nil || !node.GetKey().Equal(addr) {
			t.addFailure(newTrieFailure("got unexpected return, got "+node.String(), trie2.ToBase()))
		}
	}
	if trie2.Size() != len(addrs) {
		t.addFailure(newTrieFailure("got size mismatch, got "+strconv.Itoa(trie2.Size())+" not "+strconv.Itoa(len(addrs)), trie2.ToBase()))
	}

	for i, addr := range addrs {
		v := trie2.Get(addr)
		var expected ipaddr.NodeValue
		if i%2 == 0 {
		} else if i%4 == 1 {
			if i%8 == 1 {
				expected = valueProducer(i + 1)
			} else {
				expected = valueProducer(i)
			}
		} else {
			expected = valueProducer(1)
		}
		if !reflect.DeepEqual(v, expected) {
			t.addFailure(newTrieFailure(fmt.Sprintf("get mismatch, got %v, not %v", v, expected), trie.ToBase()))
		}
	}
	var firstNode *ipaddr.AssociativeAddressTrieNode
	func() {
		defer func() {
			if r := recover(); r != nil {
				b, _ := trie2.Put(firstNode.GetKey(), firstNode.GetValue())
				if !b {
					t.addFailure(newTrieFailure("should have added", trie2.ToBase()))
				}
			}
		}()
		for _, addr := range addrs {
			node := trie2.GetAddedNode(addr)
			firstNode = node
			trie2.RemapIfAbsent(addr, func() ipaddr.NodeValue {
				node.Remove()
				return valueProducer(1)
			}, false)
			t.addFailure(newTrieFailure("should have paniced", trie2.ToBase()))
		}
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				b, _ := trie2.Put(firstNode.GetKey(), firstNode.GetValue())
				if !b {
					t.addFailure(newTrieFailure("should have added", trie2.ToBase()))
				}
			}
		}()
		for _, addr := range addrs {
			node := trie2.GetAddedNode(addr)
			firstNode = node
			trie2.Remap(addr, func(ipaddr.NodeValue) ipaddr.NodeValue {
				node.Remove()
				return valueProducer(1)
			})
			t.addFailure(newTrieFailure("should have paniced", trie2.ToBase()))
		}
	}()

	// all trie4 from now on
	for i, addr := range addrs {
		node := trie4.PutNode(addr, valueProducer(i))
		if !reflect.DeepEqual(node.GetValue(), valueProducer(i)) {
			t.addFailure(newTrieFailure(fmt.Sprintf("got putNode mismatch, got %v not %v", node.GetValue(), valueProducer(i)), trie.ToBase()))
		}
	}
	if trie4.Size() != len(addrs) {
		t.addFailure(newTrieFailure("got size mismatch, got "+strconv.Itoa(trie4.Size())+" not "+strconv.Itoa(len(addrs)), trie4.ToBase()))
	}
	// end put tests
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
