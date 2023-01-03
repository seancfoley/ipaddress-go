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
	"sync/atomic"
)

type trieTesterGeneric struct { //TODO to truly test the generics, need to create trieTesterGeneric[T][V] and then filter out the code for each in run()
	testBase
}

var didOneMegaTree int32

type AddressTrie = ipaddr.AddressTrie
type AddressTrieNode = ipaddr.TrieNode[*ipaddr.Address]

func NewIPv4AddressGenericTrie() *AddressTrie {
	return &AddressTrie{}
}

func NewIPv4AddressAssociativeGenericTrie[V any]() *ipaddr.AssociativeTrie[*ipaddr.Address, V] {
	return &ipaddr.AssociativeTrie[*ipaddr.Address, V]{}
}

func NewIPv6AddressGenericTrie() *AddressTrie {
	return &AddressTrie{}
}

func NewIPv6AddressAssociativeGenericTrie[V any]() *ipaddr.AssociativeTrie[*ipaddr.Address, V] {
	return &ipaddr.AssociativeTrie[*ipaddr.Address, V]{}
}

func NewAddressGenericTrie() *AddressTrie {
	return &AddressTrie{}
}

func NewAssociativeAddressGenericTrie[V any]() *ipaddr.AssociativeTrie[*ipaddr.Address, V] {
	return &ipaddr.AssociativeTrie[*ipaddr.Address, V]{}
}

func (t trieTesterGeneric) run() {

	t.testAddressCheck()
	t.partitionTest()

	sampleIPAddressTries := t.getSampleIPAddressTries()
	for _, treeAddrs := range sampleIPAddressTries {
		t.testRemove(treeAddrs)
	}
	notDoneEmptyIPv6 := true
	notDoneEmptyIPv4 := true
	for _, treeAddrs := range sampleIPAddressTries {
		ipv6Tree := NewIPv6AddressGenericTrie()
		t.createIPv6SampleTree(ipv6Tree, treeAddrs)
		size := ipv6Tree.Size()
		if size > 0 || notDoneEmptyIPv6 {
			if notDoneEmptyIPv6 {
				notDoneEmptyIPv6 = size != 0
			}
			t.testIterate(ipv6Tree)
			t.testContains(ipv6Tree)
		}

		ipv4Tree := NewIPv4AddressGenericTrie()
		t.createIPv4SampleTree(ipv4Tree, treeAddrs)
		size = ipv4Tree.Size()
		if size > 0 || notDoneEmptyIPv4 {
			if notDoneEmptyIPv4 {
				notDoneEmptyIPv4 = size != 0
			}
			t.testIterate(ipv4Tree)
			t.testContains(ipv4Tree)
		}
	}
	notDoneEmptyIPv6 = true
	notDoneEmptyIPv4 = true
	for i, treeAddrs := range sampleIPAddressTries {
		_ = i
		addrs := collect(treeAddrs, func(addrStr string) *ipaddr.Address {
			return t.createAddress(addrStr).GetAddress().ToIPv4().ToAddressBase()
		})
		size := len(addrs)
		if size > 0 || notDoneEmptyIPv4 {
			if notDoneEmptyIPv4 {
				notDoneEmptyIPv4 = size != 0
			}

			t.testAdd(NewIPv4AddressGenericTrie(), addrs)
			t.testEdges(NewIPv4AddressGenericTrie(), addrs)
			t.testMap(NewIPv4AddressAssociativeGenericTrie[any](), addrs, func(i int) any { return i }, func(v any) any { return 2 * 1 })
		}
		addrsv6 := collect(treeAddrs, func(addrStr string) *ipaddr.Address {
			return t.createAddress(addrStr).GetAddress().ToIPv6().ToAddressBase()
		})
		size = len(addrsv6)
		if size > 0 || notDoneEmptyIPv6 {
			if notDoneEmptyIPv6 {
				notDoneEmptyIPv6 = size != 0
			}
			t.testAdd(NewIPv6AddressGenericTrie(), addrsv6)
			t.testEdges(NewIPv6AddressGenericTrie(), addrsv6)
			t.testMap(NewIPv6AddressAssociativeGenericTrie[any](), addrsv6,
				func(i int) any { return "bla" + strconv.Itoa(i) },
				func(str any) any { return str.(string) + "foo" })
		}
	}

	notDoneEmptyMAC := true
	for _, treeAddrs := range testMACTries {
		tree := NewAddressGenericTrie()
		macTree := tree
		t.createMACSampleTree(macTree, treeAddrs)
		size := macTree.Size()
		if size > 0 || notDoneEmptyMAC {
			if notDoneEmptyMAC {
				notDoneEmptyMAC = size != 0
			}
			t.testIterate(macTree)
			t.testContains(macTree)
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
			tree := NewAddressGenericTrie()
			t.testAdd(tree, addrs)
			tree2 := NewAddressGenericTrie()
			t.testEdges(tree2, addrs)
			tree3 := NewAssociativeAddressGenericTrie[any]()
			t.testMap(tree3, addrs,
				func(i int) any { return i },
				func(i any) any {
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

	doMegaTreeInt := atomic.LoadInt32(&didOneMegaTree)
	if doMegaTreeInt == 0 {
		cached := t.getAllCached()
		if len(cached) > 0 {
			doMegaTree := atomic.CompareAndSwapInt32(&didOneMegaTree, 0, 1)
			if doMegaTree {
				//fmt.Println("doing the mega")
				ipv6Tree1 := NewIPv6AddressGenericTrie()
				t.createIPv6SampleTreeAddrs(ipv6Tree1, cached)
				//fmt.Println(ipv6Tree1)
				//fmt.Printf("ipv6 mega tree has %v elements", ipv6Tree1.Size())
				t.testIterate(ipv6Tree1)
				t.testContains(ipv6Tree1)

				ipv4Tree1 := NewIPv4AddressGenericTrie()
				t.createIPv4SampleTreeAddrs(ipv4Tree1, cached)
				//fmt.Println(ipv4Tree1)
				//fmt.Printf("ipv4 mega tree has %v elements", ipv4Tree1.Size())
				t.testIterate(ipv4Tree1)
				t.testContains(ipv4Tree1)
			}
		}
	}

	t.testString(treeOne)
	t.testString(treeTwo)
	t.testString(treeThree)
	t.testString(treeFour)

	t.testZeroValuedAddedTrees()

	// try deleting the root
	trieb := NewAddressGenericTrie()
	trie := trieb
	if trie.Size() != 0 {
		t.addFailure(newTrieFailure("unexpected size "+strconv.Itoa(trie.Size()), trie))
	}
	if trie.NodeSize() != 0 {
		t.addFailure(newTrieFailure("unexpected size "+strconv.Itoa(trie.NodeSize()), trie))
	}
	trie.Add(ipaddr.NewIPAddressString("0.0.0.0/0").GetAddress().ToAddressBase())
	if trie.Size() != 1 {
		t.addFailure(newTrieFailure("unexpected size "+strconv.Itoa(trie.Size()), trie))
	}
	if trie.NodeSize() != 1 {
		t.addFailure(newTrieFailure("unexpected size "+strconv.Itoa(trie.Size()), trie))
	}
	trie.GetRoot().Remove()
	if trie.Size() != 0 {
		t.addFailure(newTrieFailure("unexpected size "+strconv.Itoa(trie.Size()), trie))
	}
	if trie.NodeSize() != 1 {
		t.addFailure(newTrieFailure("unexpected node size "+strconv.Itoa(trie.NodeSize()), trie))
	}

	trie = NewIPv4AddressGenericTrie()
	trie.Add(ipaddr.NewIPAddressString("1.2.3.4").GetAddress().ToAddressBase())
	trie.GetRoot().SetAdded()
	if trie.Size() != 2 {
		t.addFailure(newTrieFailure("unexpected size "+strconv.Itoa(trie.Size()), trie))
	}
	trie.GetRoot().Remove()
	if trie.Size() != 1 {
		t.addFailure(newTrieFailure("unexpected size "+strconv.Itoa(trie.Size()), trie))
	}
	if trie.NodeSize() != 2 {
		t.addFailure(newTrieFailure("unexpected node size "+strconv.Itoa(trie.NodeSize()), trie))
	}
	trie.Clear()
	if trie.NodeSize() != 1 {
		t.addFailure(newTrieFailure("unexpected node size "+strconv.Itoa(trie.NodeSize()), trie))
	}
	if trie.Size() != 0 {
		t.addFailure(newTrieFailure("unexpected size "+strconv.Itoa(trie.Size()), trie))
	}
	trie.GetRoot().Remove()
	if trie.NodeSize() != 1 {
		t.addFailure(newTrieFailure("unexpected node size "+strconv.Itoa(trie.NodeSize()), trie))
	}
	if trie.Size() != 0 {
		t.addFailure(newTrieFailure("unexpected size "+strconv.Itoa(trie.Size()), trie))
	}
	t.incrementTestCount()
}

func (t trieTesterGeneric) testString(strs trieStrings) {

	addrTree := &AddressTrie{}
	t.createIPSampleTree(addrTree, strs.addrs)
	treeStr := addrTree.String()
	if treeStr != strs.treeString {
		t.addFailure(newTrieFailure("trie string not right, got "+treeStr+" instead of expected "+strs.treeString, addrTree))
	}

	addedString := addrTree.AddedNodesTreeString()
	if addedString != strs.addedNodeString {
		t.addFailure(newTrieFailure("trie string not right, got "+addedString+" instead of expected "+strs.addedNodeString, addrTree))
	}

	tree := addrTree.ConstructAddedNodesTree()
	addedNodeTreeStr := tree.String()
	if addedNodeTreeStr != strs.addedNodeString {
		t.addFailure(newTrieFailure("assoc trie string not right, got "+addedNodeTreeStr+" instead of expected "+strs.addedNodeString, addrTree))
	}

	troot := tree.GetRoot()
	tcount := tcountNodes(troot)
	if !troot.IsAdded() {
		tcount--
	}
	if tcount != addrTree.Size() {
		t.addFailure(newTrieFailure("size not right, got "+strconv.Itoa(tcount)+" instead of expected "+strconv.Itoa(addrTree.Size()), addrTree))
	}

	assocTrie := &ipaddr.AssociativeTrie[*ipaddr.Address, int]{}
	for i, addr := range strs.addrs {
		addressStr := t.createAddress(addr)
		address := addressStr.GetAddress()
		assocTrie.Put(address.ToAddressBase(), i)
	}

	treeStr = assocTrie.String()
	if treeStr != strs.treeToIndexString {
		t.addFailure(newTrieFailure("trie string not right, got "+treeStr+" instead of expected "+strs.treeToIndexString, addrTree))
	}

	addedNodeTreeStr = assocTrie.AddedNodesTreeString()
	associatedTreeStr := strs.addedNodeToIndexString
	if addedNodeTreeStr != associatedTreeStr {
		t.addFailure(newTrieFailure("assoc trie string not right, got "+treeStr+" instead of expected "+strs.addedNodeToIndexString, addrTree))
	}

	contree := assocTrie.ConstructAddedNodesTree()
	addedNodeTreeStr = contree.String()
	if addedNodeTreeStr != associatedTreeStr {
		t.addFailure(newTrieFailure("assoc trie string not right, got "+addedNodeTreeStr+" instead of expected "+strs.addedNodeToIndexString, addrTree))
	}

	root := contree.GetRoot()
	count := countNodes(root)
	if !root.IsAdded() {
		count--
	}
	if count != assocTrie.Size() {
		t.addFailure(newTrieFailure("size not right, got "+strconv.Itoa(count)+" instead of expected "+strconv.Itoa(assocTrie.Size()), addrTree))
	}

	t.incrementTestCount()
}

func countNodes(node ipaddr.AssociativeAddedTreeNode[*ipaddr.Address, int]) int {
	count := 1
	for _, n := range node.GetSubNodes() {
		count += countNodes(n)
	}
	return count
}

func tcountNodes(node ipaddr.AddedTreeNode[*ipaddr.Address]) int {
	count := 1
	for _, n := range node.GetSubNodes() {
		count += tcountNodes(n)
	}
	return count
}

func (t trieTesterGeneric) testZeroValuedAddedTrees() {
	addedTree := ipaddr.AddedTree[*ipaddr.IPv4Address]{}
	t.checkString(addedTree.String(), "\n○ <nil>\n")
	t.checkString(addedTree.GetRoot().String(), "○ <nil>")
	t.checkString(addedTree.GetRoot().GetKey().String(), "<nil>")
	t.checkString(fmt.Sprint(addedTree.GetRoot().GetSubNodes()), "[]")
	t.checkString(addedTree.GetRoot().TreeString(), "\n○ <nil>\n")

	addedTreeNode := ipaddr.AddedTreeNode[*ipaddr.IPv4Address]{}
	t.checkString(addedTreeNode.String(), "○ <nil>")
	t.checkString(addedTreeNode.GetKey().String(), "<nil>")
	t.checkString(fmt.Sprint(addedTreeNode.GetSubNodes()), "[]")
	t.checkString(addedTreeNode.TreeString(), "\n○ <nil>\n")

	assocAddedTree := ipaddr.AssociativeAddedTree[*ipaddr.IPv4Address, int]{}
	t.checkString(assocAddedTree.String(), "\n○ <nil> = 0\n")
	t.checkString(assocAddedTree.GetRoot().String(), "○ <nil> = 0")
	t.checkString(assocAddedTree.GetRoot().GetKey().String(), "<nil>")
	t.checkString(fmt.Sprint(assocAddedTree.GetRoot().GetValue()), "0")
	t.checkString(fmt.Sprint(assocAddedTree.GetRoot().GetSubNodes()), "[]")
	t.checkString(assocAddedTree.GetRoot().TreeString(), "\n○ <nil> = 0\n")

	assocAddedTreeNode := ipaddr.AssociativeAddedTreeNode[*ipaddr.IPAddress, float64]{}
	t.checkString(assocAddedTreeNode.String(), "○ <nil> = 0")
	t.checkString(assocAddedTreeNode.GetKey().String(), "<nil>")
	t.checkString(fmt.Sprint(assocAddedTreeNode.GetValue()), "0")
	t.checkString(fmt.Sprint(assocAddedTreeNode.GetSubNodes()), "[]")
	t.checkString(assocAddedTreeNode.TreeString(), "\n○ <nil> = 0\n")

	t.incrementTestCount()
}

func (t trieTesterGeneric) checkString(actual, expected string) {
	if actual != expected {
		t.addFailure(newAddressItemFailure(" mismatched strings, expected "+expected+" got "+actual, nil))
	}
}

func (t trieTesterGeneric) testAddressCheck() {
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
	t.testNonBlock("a:b:c:*:2:*") // passes nil into checkBlockOrAddress
	t.testMACAddrBlock("a:b:c:1:2:3")
}

func (t trieTesterGeneric) testConvertedBlock(str string, expectedPrefLen ipaddr.PrefixLen) {
	addr := t.createAddress(str).GetAddress()
	t.testConvertedAddrBlock(addr.ToAddressBase(), expectedPrefLen)
}

func (t trieTesterGeneric) testConvertedAddrBlock(addr *ipaddr.Address, expectedPrefLen ipaddr.PrefixLen) {
	result := addr.ToSinglePrefixBlockOrAddress()
	if result == nil {
		t.addFailure(newAddrFailure("unexpectedly got no single block or address for "+addr.String(), addr))
	}
	if !addr.Equal(result) && !result.GetPrefixLen().Equal(expectedPrefLen) {
		t.addFailure(newAddrFailure("unexpectedly got wrong pref len "+result.GetPrefixLen().String()+" not "+expectedPrefLen.String(), addr))
	}
}

func (t trieTesterGeneric) testNonBlock(str string) {
	addr := t.createAddress(str).GetAddress()
	result := addr.ToSinglePrefixBlockOrAddress()
	if result != nil {
		t.addFailure(newIPAddrFailure("unexpectedly got a single block or address for "+addr.String(), addr))
	}
}

func (t trieTesterGeneric) testNonMACBlock(str string) {
	addr := t.createMACAddress(str).GetAddress()
	result := addr.ToSinglePrefixBlockOrAddress()
	if result != nil {
		t.addFailure(newMACAddrFailure("unexpectedly got a single block or address for "+addr.String(), addr))
	}
}

func (t trieTesterGeneric) testIPAddrBlock(str string) {
	addr := t.createAddress(str).GetAddress()
	result := addr.ToSinglePrefixBlockOrAddress()
	if result != addr {
		t.addFailure(newIPAddrFailure("unexpectedly got different address "+result.String()+" for "+addr.String(), addr))
	}
}

func (t trieTesterGeneric) testMACAddrBlock(str string) {
	addr := t.createMACAddress(str).GetAddress()
	result := addr.ToSinglePrefixBlockOrAddress()
	if result != addr {
		t.addFailure(newMACAddrFailure("unexpectedly got different address "+result.String()+" for "+addr.String(), addr))
	}
}

func (t trieTesterGeneric) partitionTest() {
	addrs := "1.2.1-15.*"
	trie := NewIPv4AddressGenericTrie()
	addr := t.createAddress(addrs).GetAddress()
	t.partitionForTrie(trie, addr)
}

func (t trieTesterGeneric) partitionForTrie(trie *AddressTrie, subnet *ipaddr.IPAddress) {
	ipaddr.PartitionWithSingleBlockSize(subnet).PredicateForEach(IPAddressPredicateAdapter{trie.Add}.IPPredicate)
	if trie.Size() != 15 {
		t.addFailure(newTrieFailure("partition size unexpected "+strconv.Itoa(trie.Size())+", expected 15", trie.Clone()))
	}

	partition := ipaddr.PartitionWithSingleBlockSize(subnet.ToAddressBase())
	all := ipaddr.ApplyForEach[*ipaddr.Address, *AddressTrieNode](partition, trie.GetAddedNode)

	if len(all) != 15 {
		t.addFailure(newTrieFailure("map size unexpected "+strconv.Itoa(trie.Size())+", expected 15", trie))
	}
	partition = ipaddr.PartitionWithSingleBlockSize(subnet.ToAddressBase())
	keyAll := ipaddr.ApplyForEach[*ipaddr.Address, struct{}](partition, func(*ipaddr.Address) struct{} { return struct{}{} })
	if len(keyAll) != 15 {
		t.addFailure(newTrieFailure("map size unexpected "+strconv.Itoa(trie.Size())+", expected 15", trie))
	}
	keyAll1 := make(map[ipaddr.Key[*ipaddr.Address]]struct{})
	for k, v := range keyAll {
		keyAll1[k.ToAddress().ToKey()] = v
	}

	all2 := make(ipaddr.MappedPartition[*ipaddr.Address, *AddressTrieNode])
	keyAll2 := make(map[ipaddr.Key[*ipaddr.Address]]struct{})
	ipaddr.PartitionWithSingleBlockSize(subnet).ForEach(func(addr *ipaddr.IPAddress) {
		node := trie.GetAddedNode(addr.ToAddressBase())
		all2[addr.ToAddressBase().ToKey()] = node
		keyAll2[addr.ToAddressBase().ToKey()] = struct{}{}
	})

	// using keys and struct{} allow for deep-equal comparison
	if !reflect.DeepEqual(keyAll1, keyAll2) {
		str := fmt.Sprintf("maps not equal %v and %v", all, all2)
		reflect.DeepEqual(keyAll1, keyAll2)
		t.addFailure(newTrieFailure(str, trie))
	}

	// the maps using *Address keys and *Address nodes are not expected to be equal since they use pointers
	for k, v := range all {
		if !k.ToAddress().Equal(v.GetKey()) {
			t.addFailure(newTrieFailure("node key wrong for "+k.String(), trie))
		}
		found := false
		for k2, v2 := range all2 {
			if !k2.ToAddress().Equal(v2.GetKey()) {
				t.addFailure(newTrieFailure("node key wrong for "+k2.String(), trie))
			}
			if k.ToAddress().Equal(k2.ToAddress()) {
				found = true
				break
			}
		}
		if !found {
			t.addFailure(newTrieFailure("could not find "+k.String(), trie))
		}
	}

	trie.Clear()
	ipaddr.PartitionWithSpanningBlocks(subnet).PredicateForEach(IPAddressPredicateAdapter{trie.Add}.IPPredicate)
	if trie.Size() != 4 {
		t.addFailure(newTrieFailure("partition size unexpected "+strconv.Itoa(trie.Size())+", expected 4", trie.Clone()))
	}
	trie.Clear()
	ipaddr.PartitionWithSingleBlockSize(subnet).PredicateForEach(IPAddressPredicateAdapter{trie.Add}.IPPredicate)
	ipaddr.PartitionWithSpanningBlocks(subnet).PredicateForEach(IPAddressPredicateAdapter{trie.Add}.IPPredicate)
	if trie.Size() != 18 {
		t.addFailure(newTrieFailure("partition size unexpected "+strconv.Itoa(trie.Size())+", expected 18", trie.Clone()))
	}
	allAreThere := ipaddr.PartitionWithSingleBlockSize(subnet).PredicateForEach(IPAddressPredicateAdapter{trie.Contains}.IPPredicate)
	allAreThere2 := ipaddr.PartitionWithSpanningBlocks(subnet).PredicateForEach(IPAddressPredicateAdapter{trie.Contains}.IPPredicate)
	if !(allAreThere && allAreThere2) {
		t.addFailure(newTrieFailure("partition contains check failing", trie))
	}
	t.incrementTestCount()
}

func (t trieTesterGeneric) getSampleIPAddressTries() [][]string {
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

func (t trieTesterGeneric) testRemove(addrs []string) {
	ipv6Tree := NewIPv6AddressGenericTrie()
	ipv4Tree := NewIPv4AddressGenericTrie()

	t.testRemoveAddrs(ipv6Tree, addrs, func(addrStr string) *ipaddr.Address {
		return t.createAddress(addrStr).GetAddress().ToIPv6().ToAddressBase()
	})
	t.testRemoveAddrs(ipv4Tree, addrs, func(addrStr string) *ipaddr.Address {
		return t.createAddress(addrStr).GetAddress().ToIPv4().ToAddressBase()
	})

	// reverse the address order
	var addrs2 = make([]string, len(addrs))
	for i := range addrs {
		addrs2[len(addrs2)-i-1] = addrs[i]
	}

	// both trees should be empty now
	t.testRemoveAddrs(ipv6Tree, addrs, func(addrStr string) *ipaddr.Address {
		return t.createAddress(addrStr).GetAddress().ToIPv6().ToAddressBase()
	})
	t.testRemoveAddrs(ipv4Tree, addrs, func(addrStr string) *ipaddr.Address {
		return t.createAddress(addrStr).GetAddress().ToIPv4().ToAddressBase()
	})
}

func (t trieTesterGeneric) testRemoveMAC(addrs []string) {
	tree := NewAddressGenericTrie()
	t.testRemoveAddrs(tree, addrs, func(addrStr string) *ipaddr.Address {
		return t.createMACAddress(addrStr).GetAddress().ToAddressBase()
	})

	// reverse the address order
	var addrs2 = make([]string, len(addrs))
	for i := range addrs {
		addrs2[len(addrs2)-i-1] = addrs[i]
	}

	// tree should be empty now
	t.testRemoveAddrs(tree, addrs, func(addrStr string) *ipaddr.Address {
		return t.createMACAddress(addrStr).GetAddress().ToAddressBase()
	})
	t.incrementTestCount()
}

func (t trieTesterGeneric) testRemoveAddrs(tree *AddressTrie, addrs []string, converter func(addrStr string) *ipaddr.Address) {
	count := 0
	var list []*ipaddr.Address
	dupChecker := make(map[AddressKey]struct{})
	for _, str := range addrs {
		addr := converter(str)
		if addr != nil {
			key := addr.ToKey()
			if _, exists := dupChecker[key]; !exists {
				dupChecker[key] = struct{}{}
				list = append(list, addr)
				count++
				tree.Add(addr)
			}
		}
	}
	t.testRemoveAddrsConverted(tree, count, list)
}

func (t trieTesterGeneric) testRemoveAddrsConverted(tree *AddressTrie, count int, addrs []*ipaddr.Address) {
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

func (t trieTesterGeneric) createMACSampleTree(tree *AddressTrie, addrs []string) {
	for _, addr := range addrs {
		addressStr := t.createMACAddress(addr)
		address := addressStr.GetAddress()
		tree.Add(address.ToAddressBase())
	}
}

func (t trieTesterGeneric) createIPv6SampleTree(tree *AddressTrie, addrs []string) {
	for _, addr := range addrs {
		addressStr := t.createAddress(addr)
		if addressStr.IsIPv6() {
			address := addressStr.GetAddress()
			tree.Add(address.ToAddressBase())
		}
	}
}

func (t trieTesterGeneric) createIPv6SampleTreeAddrs(tree *AddressTrie, addrs []*ipaddr.IPAddress) {
	for _, addr := range addrs {
		if addr.IsIPv6() {
			addr = addr.ToSinglePrefixBlockOrAddress()
			if addr != nil {
				address := addr.ToIPv6()
				tree.Add(address.ToAddressBase())
			}
		}
	}
}

func (t trieTesterGeneric) createIPv4SampleTreeAddrs(tree *AddressTrie, addrs []*ipaddr.IPAddress) {
	for _, addr := range addrs {
		if addr.IsIPv4() {
			addr = addr.ToSinglePrefixBlockOrAddress()
			if addr != nil {
				address := addr.ToIPv4()
				tree.Add(address.ToAddressBase())
			}
		}
	}
}

func (t trieTesterGeneric) createIPv4SampleTree(tree *AddressTrie, addrs []string) {
	for _, addr := range addrs {
		addressStr := t.createAddress(addr)
		if addressStr.IsIPv4() {
			address := addressStr.GetAddress().ToIPv4()
			tree.Add(address.ToAddressBase())
		}
	}
}

func (t trieTesterGeneric) createIPSampleTree(tree *AddressTrie, addrs []string) {
	for _, addr := range addrs {
		addressStr := t.createAddress(addr)
		address := addressStr.GetAddress()
		tree.Add(address.ToAddressBase())
	}
}

//<R extends AddressTrie<T>, T extends Address>
func (t trieTesterGeneric) testIterationContainment(tree *AddressTrie) {
	t.testIterationContainmentTree(tree, func(trie *AddressTrie) ipaddr.CachingTrieIterator[*AddressTrieNode] {
		return trie.BlockSizeCachingAllNodeIterator()
	}, false)
	t.testIterationContainmentTree(tree, func(trie *AddressTrie) ipaddr.CachingTrieIterator[*AddressTrieNode] {
		return trie.ContainingFirstAllNodeIterator(true)
	}, false /* added only */)
	t.testIterationContainmentTree(tree, func(trie *AddressTrie) ipaddr.CachingTrieIterator[*AddressTrieNode] {
		return trie.ContainingFirstAllNodeIterator(false)
	}, false /* added only */)
	t.testIterationContainmentTree(tree, func(trie *AddressTrie) ipaddr.CachingTrieIterator[*AddressTrieNode] {
		return trie.ContainingFirstIterator(true)
	}, true /* added only */)
	t.testIterationContainmentTree(tree, func(trie *AddressTrie) ipaddr.CachingTrieIterator[*AddressTrieNode] {
		return trie.ContainingFirstIterator(false)
	}, true /* added only */)
}

func (t trieTesterGeneric) testIterationContainmentTree(
	trie *ipaddr.Trie[*ipaddr.Address],
	iteratorFunc func(addressTrie *AddressTrie) ipaddr.CachingTrieIterator[*AddressTrieNode],
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

func (t trieTesterGeneric) testIterate(tree *AddressTrie) {

	type triePtr = *AddressTrie
	t.testIteratorRem(tree, func(trie *AddressTrie) ipaddr.IteratorWithRemove[*AddressTrieNode] {
		return trie.BlockSizeNodeIterator(true)
	}, triePtr.Size)
	t.testIteratorRem(tree, func(trie *AddressTrie) ipaddr.IteratorWithRemove[*AddressTrieNode] {
		return trie.BlockSizeAllNodeIterator(true)
	}, triePtr.NodeSize)
	t.testIteratorRem(tree, func(trie *AddressTrie) ipaddr.IteratorWithRemove[*AddressTrieNode] {
		return trie.BlockSizeNodeIterator(false)
	}, triePtr.Size)
	t.testIteratorRem(tree, func(trie *AddressTrie) ipaddr.IteratorWithRemove[*AddressTrieNode] {
		return trie.BlockSizeAllNodeIterator(false)
	}, triePtr.NodeSize)

	t.testIteratorRem(tree, func(trie *AddressTrie) ipaddr.IteratorWithRemove[*AddressTrieNode] {
		return trie.BlockSizeCachingAllNodeIterator()
	}, triePtr.NodeSize)

	t.testIteratorRem(tree, func(trie *AddressTrie) ipaddr.IteratorWithRemove[*AddressTrieNode] {
		return trie.NodeIterator(true)
	}, triePtr.Size)
	t.testIteratorRem(tree, func(trie *AddressTrie) ipaddr.IteratorWithRemove[*AddressTrieNode] {
		return trie.AllNodeIterator(true)
	}, triePtr.NodeSize)
	t.testIteratorRem(tree, func(trie *AddressTrie) ipaddr.IteratorWithRemove[*AddressTrieNode] {
		return trie.NodeIterator(false)
	}, triePtr.Size)
	t.testIteratorRem(tree, func(trie *AddressTrie) ipaddr.IteratorWithRemove[*AddressTrieNode] {
		return trie.AllNodeIterator(false)
	}, triePtr.NodeSize)

	t.testIteratorRem(tree, func(trie *AddressTrie) ipaddr.IteratorWithRemove[*AddressTrieNode] {
		return trie.ContainedFirstIterator(true)
	}, triePtr.Size)
	t.testIterator(tree, func(trie *AddressTrie) ipaddr.Iterator[*AddressTrieNode] {
		return trie.ContainedFirstAllNodeIterator(true)
	}, triePtr.NodeSize)
	t.testIteratorRem(tree, func(trie *AddressTrie) ipaddr.IteratorWithRemove[*AddressTrieNode] {
		return trie.ContainedFirstIterator(false)
	}, triePtr.Size)
	t.testIterator(tree, func(trie *AddressTrie) ipaddr.Iterator[*AddressTrieNode] {
		return trie.ContainedFirstAllNodeIterator(false)
	}, triePtr.NodeSize)

	t.testIteratorRem(tree, func(trie *AddressTrie) ipaddr.IteratorWithRemove[*AddressTrieNode] {
		return trie.ContainingFirstIterator(true)
	}, triePtr.Size)
	t.testIteratorRem(tree, func(trie *AddressTrie) ipaddr.IteratorWithRemove[*AddressTrieNode] {
		return trie.ContainingFirstAllNodeIterator(true)
	}, triePtr.NodeSize)
	t.testIteratorRem(tree, func(trie *AddressTrie) ipaddr.IteratorWithRemove[*AddressTrieNode] {
		return trie.ContainingFirstIterator(false)
	}, triePtr.Size)
	t.testIteratorRem(tree, func(trie *AddressTrie) ipaddr.IteratorWithRemove[*AddressTrieNode] {
		return trie.ContainingFirstAllNodeIterator(false)
	}, triePtr.NodeSize)

	t.testIterationContainment(tree)

	t.incrementTestCount()
}

func (t trieTesterGeneric) testIteratorRem(
	trie *AddressTrie,
	iteratorFunc func(*AddressTrie) ipaddr.IteratorWithRemove[*AddressTrieNode],
	countFunc func(*AddressTrie) int) {
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
		set := make(map[AddressKey]struct{})
		iterator := iteratorFunc(trie)
		for iterator.HasNext() {
			next := iterator.Next()
			nextAddr := next.GetKey()
			set[nextAddr.ToKey()] = struct{}{}
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

func (t trieTesterGeneric) testIterator(
	trie *AddressTrie,
	iteratorFunc func(*AddressTrie) ipaddr.Iterator[*AddressTrieNode],
	countFunc func(*AddressTrie) int) {
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
	set := make(map[AddressKey]struct{})
	iterator := iteratorFunc(trie)
	for iterator.HasNext() {
		next := iterator.Next()
		nextAddr := next.GetKey()
		set[nextAddr.ToKey()] = struct{}{}
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

func (t trieTesterGeneric) testContains(trie *AddressTrie) {
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
		//if ok != (lpm != nil) {
		//	t.addFailure(newTrieFailure("expecting a match for "+halfwayAddr.String(), trie))
		//}
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
			//fmt.Printf("empty containing is %s\n", containing)
			if (containing != nil && containing.Count() > 0) || lpm != nil {
				t.addFailure(newTrieFailure("containing is "+containing.String()+" for address "+halfwayAddr.String()+" instead of expected nil", trie))
			} else if elementsContains {
				t.addFailure(newTrieFailure("containing is true for address "+halfwayAddr.String()+" instead of expected false", trie))
			}
		} else {
			var lastContaining *ipaddr.ContainmentPathNode[*ipaddr.Address]
			//fmt.Printf("containing is %s\n", containing)
			if containing.Count() > 0 {
				lastContaining = containing.ShortestPrefixMatch()
				for {
					if next := lastContaining.Next(); next == nil {
						break
					} else {
						lastContaining = next
					}
				}
			}
			if lastContaining == nil || !lastContaining.GetKey().Equal(addedParent.GetKey()) {
				t.addFailure(newTrieFailure("containing ends with "+lastContaining.String()+" for address "+halfwayAddr.String()+" instead of expected "+addedParent.String(), trie))
			} else if !lastContaining.GetKey().Equal(smallestContaining.GetKey()) {
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

func (t trieTesterGeneric) testEdges(trie *AddressTrie, addrs []*ipaddr.Address) {
	trie2 := trie.Clone()
	for _, addr := range addrs {
		trie.Add(addr)
	}
	i := 0
	ordered := make([]*AddressTrieNode, 0, len(addrs))
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
func (t trieTesterGeneric) testAdd(trie *AddressTrie, addrs []*ipaddr.Address) {
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

func (t trieTesterGeneric) testMap(trie *ipaddr.AssociativeTrie[*ipaddr.Address, any], addrs []*ipaddr.Address,
	valueProducer func(int) any, mapper func(any) any) { // seems we used to use the mapper but no now
	// put tests
	trie2 := trie.Clone()
	trie4 := trie.Clone()
	for i, addr := range addrs {
		trie.Put(addr, valueProducer(i))
	}
	for i, addr := range addrs {
		v, _ := trie.Get(addr)
		expected := valueProducer(i)
		if !reflect.DeepEqual(v, expected) { //reflect deep equal
			//fmt.Println(trie)
			t.addFailure(newAssocTrieFailure(fmt.Sprintf("got mismatch, got %v, not %v for %v", v, expected, addr), trie))
			//v, _ = trie.Get(addr)
		}
	}
	// all trie2 from now on
	trie2.PutTrie(trie.GetRoot())
	for i, addr := range addrs {
		v, _ := trie2.Get(addr)
		expected := valueProducer(i)
		if !reflect.DeepEqual(v, expected) {
			t.addFailure(newAssocTrieFailure(fmt.Sprintf("got mismatch, got %v, not %v", v, expected), trie2))
		}
		if i%2 == 0 {
			trie2.Remove(addr)
		}
	}
	if trie2.Size() != (len(addrs) >> 1) {
		t.addFailure(newAssocTrieFailure("got size mismatch, got "+strconv.Itoa(trie2.Size())+" not "+strconv.Itoa(len(addrs)>>1), trie2))
	}
	trie2.PutTrie(trie.GetRoot())
	for i, addr := range addrs {
		v, _ := trie2.Get(addr)
		expected := valueProducer(i)
		if !reflect.DeepEqual(v, expected) {
			t.addFailure(newAssocTrieFailure(fmt.Sprintf("get mismatch, got %v, not %v", v, expected), trie2))
		}
	}
	if trie2.Size() != len(addrs) {
		t.addFailure(newAssocTrieFailure("got size mismatch, got "+strconv.Itoa(trie2.Size())+" not "+strconv.Itoa(len(addrs)), trie2))
	}
	for i, addr := range addrs {
		if i%2 == 0 {
			b := trie2.Remove(addr)
			if !b {
				t.addFailure(newAssocTrieFailure("remove should have succeeded", trie2))
			}
			b = trie2.Remove(addr)
			if b {
				t.addFailure(newAssocTrieFailure("remove should not have succeeded", trie2))
			}
		}

	}
	for i, addr := range addrs {
		_, exists := trie2.Put(addr, valueProducer(i))
		if exists != (i%2 == 0) {
			t.addFailure(newAssocTrieFailure("putNew mismatch", trie2))
		}
	}
	if trie2.Size() != len(addrs) {
		t.addFailure(newAssocTrieFailure("got size mismatch, got "+strconv.Itoa(trie.Size())+" not "+strconv.Itoa(len(addrs)), trie2))
	}
	for i, addr := range addrs {
		_, res := trie2.Put(addr, valueProducer(i+1))
		if res {
			t.addFailure(newAssocTrieFailure("putNew mismatch", trie2))
		}
	}
	if trie2.Size() != len(addrs) {
		t.addFailure(newAssocTrieFailure("got size mismatch, got "+strconv.Itoa(trie.Size())+" not "+strconv.Itoa(len(addrs)), trie2))
	}
	for i, addr := range addrs {
		v, _ := trie2.Get(addr)
		expected := valueProducer(i + 1)
		if !reflect.DeepEqual(v, expected) {
			t.addFailure(newAssocTrieFailure(fmt.Sprintf("get mismatch, got %v, not %v", v, expected), trie))
		}
	}
	for i, addr := range addrs {
		v, _ := trie2.Put(addr, valueProducer(i))
		expected := valueProducer(i + 1)
		if !reflect.DeepEqual(v, expected) {
			t.addFailure(newAssocTrieFailure(fmt.Sprintf("get mismatch, got %v, not %v", v, expected), trie))
		}
		v, _ = trie2.Get(addr)
		expected = valueProducer(i)
		if !reflect.DeepEqual(v, expected) {
			t.addFailure(newAssocTrieFailure(fmt.Sprintf("get mismatch, got %v, not %v", v, expected), trie))
		}
	}

	k := 0
	for i, addr := range addrs {
		if i%2 == 0 {
			b := trie2.Remove(addr)
			if !b {
				t.addFailure(newAssocTrieFailure("remove should have succeeded", trie2))
			}
		}
		// the reason for the (i % 8 == 1) is that the existing value is already valueProducer.apply(i),
		// so half the time we are re-adding the existing value,
		// half the time we are changing to a new value
		var value any
		if i%4 == 1 {
			if i%8 == 1 {
				value = valueProducer(i + 1)
			} else {
				value = valueProducer(i)
			}
		}
		node := trie2.Remap(addr, func(val any, found bool) (any, bool) {
			if val == nil {
				if found {
					t.addFailure(newAssocTrieFailure("got unexpected vals", trie2))
				}
				return valueProducer(0), true
			} else {
				if !found {
					t.addFailure(newAssocTrieFailure("got unexpected vals", trie2))
				}
				return value, value != nil
			}
		})
		if node == nil || !node.GetKey().Equal(addr) {
			t.addFailure(newAssocTrieFailure("got unexpected return, got "+node.String(), trie2))
		}
		if i%2 != 0 && value == nil {
			k++
		}
	}
	if trie2.Size()+k != len(addrs) {
		t.addFailure(newAssocTrieFailure("got size mismatch, got "+strconv.Itoa(trie2.Size())+" not "+strconv.Itoa(len(addrs)-k), trie2))
	}
	for i, addr := range addrs {
		v, _ := trie2.Get(addr)
		var expected any
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
			t.addFailure(newAssocTrieFailure(fmt.Sprintf("got mismatch, got %v, not %v", v, expected), trie))
		}
	}
	for _, addr := range addrs {
		trie2.RemapIfAbsent(addr, func() any {
			return valueProducer(1)
		}) // false
	}
	if trie2.Size() != len(addrs) {
		t.addFailure(newAssocTrieFailure("got size mismatch, got "+strconv.Itoa(trie2.Size())+" not "+strconv.Itoa(len(addrs)), trie2))
	}
	for i, addr := range addrs {
		v, _ := trie2.Get(addr)
		var expected any
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
			t.addFailure(newAssocTrieFailure(fmt.Sprintf("get mismatch, got %v, not %v", v, expected), trie))
		}
	}
	if trie2.Size() != len(addrs) {
		t.addFailure(newAssocTrieFailure("got size mismatch, got "+strconv.Itoa(trie2.Size())+" not "+strconv.Itoa(len(addrs)), trie2))
	}

	for i, addr := range addrs {
		if i%2 == 0 {
			trie2.GetNode(addr).Remove()
		}
	}
	if trie2.Size() != (len(addrs) >> 1) {
		t.addFailure(newAssocTrieFailure("got size mismatch, got "+strconv.Itoa(trie2.Size())+" not "+strconv.Itoa(len(addrs)>>1), trie2))
	}

	//for i, addr := range addrs {
	//	node := trie2.RemapIfAbsent(addr, func() (any, bool) {
	//		return nil, false
	//	}) // false
	//	if (node == nil) != (i%2 == 0) {
	//		t.addFailure(newAssocTrieFailure("got unexpected return, got "+node.String(), trie2))
	//	}
	//}
	//if trie2.Size() != (len(addrs) >> 1) {
	//	t.addFailure(newAssocTrieFailure("got size mismatch, got "+strconv.Itoa(trie2.Size())+" not "+strconv.Itoa(len(addrs)>>1), trie2))
	//}

	for _, addr := range addrs {
		node := trie2.RemapIfAbsent(addr, func() any {
			return nil
		}) // true
		if node != nil && !node.GetKey().Equal(addr) {
			t.addFailure(newAssocTrieFailure("got unexpected return, got "+node.String(), trie2))
		}
		//if node == nil {
		//	trie2.Put(addr, nil)
		//}
	}
	if trie2.Size() != len(addrs) {
		t.addFailure(newAssocTrieFailure("got size mismatch, got "+strconv.Itoa(trie2.Size())+" not "+strconv.Itoa(len(addrs)), trie2))
	}
	for i, addr := range addrs {
		v, _ := trie2.Get(addr)
		var expected any
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
			t.addFailure(newAssocTrieFailure(fmt.Sprintf("get mismatch, got %v, not %v", v, expected), trie))
		}
	}
	var firstNode *ipaddr.AssociativeTrieNode[*ipaddr.Address, any]
	func() {
		defer func() {
			if r := recover(); r != nil { // r is what was passed to panic
				v := firstNode.GetValue() // firstNode is the node with the value we were remapping
				_, b := trie2.Put(firstNode.GetKey(), v)
				if b {
					t.addFailure(newAssocTrieFailure("should have added", trie2))
				}
			}
		}()
		// remove the first addr so we should panic

		for _, addr := range addrs {
			node := trie2.GetAddedNode(addr)
			firstNode = node
			// remove the node so we should panic on the remap
			trie2.Remove(addr)
			trie2.RemapIfAbsent(addr, func() any {
				trie2.Add(addr)
				return valueProducer(1)
			}) // false
			t.addFailure(newAssocTrieFailure("should have paniced", trie2))
		}
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				v := firstNode.GetValue()
				_, b := trie2.Put(firstNode.GetKey(), v)
				if !b {
					t.addFailure(newAssocTrieFailure("should have added", trie2))
				}
			}
		}()
		for _, addr := range addrs {
			node := trie2.GetAddedNode(addr)
			firstNode = node
			trie2.Remap(addr, func(any, bool) (any, bool) {
				node.Remove()
				return valueProducer(1), true
			})
			t.addFailure(newAssocTrieFailure("should have paniced", trie2))
		}
	}()

	// all trie4 from now on
	for i, addr := range addrs {
		node := trie4.PutNode(addr, valueProducer(i))
		v := node.GetValue()
		if !reflect.DeepEqual(v, valueProducer(i)) {
			t.addFailure(newAssocTrieFailure(fmt.Sprintf("got putNode mismatch, got %v not %v", node.GetValue(), valueProducer(i)), trie))
		}
	}
	if trie4.Size() != len(addrs) {
		t.addFailure(newAssocTrieFailure("got size mismatch, got "+strconv.Itoa(trie4.Size())+" not "+strconv.Itoa(len(addrs)), trie4))
	}
	// end put tests
}
