//
// Copyright 2024 Sean C Foley
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

package ipaddr

import (
	"fmt"

	"github.com/seancfoley/bintree/tree"
)

type baseDualIPv4v6Tries[V any] struct {
	ipv4Trie, ipv6Trie trieBase[*IPAddress, V]
}

// ensureRoots ensures both the IPv4 and IPv6 tries each have a root.
// The roots will be 0.0.0.0/0 and ::/0 for the IPv4 and IPv6 tries, respectively.
// Once the roots are created, any copy of the this instance will reference the same underlying tries.
// Any copy of either trie, such as through GetIPv4Trie or GetIPv6Trie, will reference the same underlying trie.
// Calling this method will have no effect if both trie roots already exist.
// Calling this method is not necessary, adding any key to a trie will also cause its trie root to be created.
// It only makes sense to call this method if at least one trie is empty, and you wish to ensure that copies of this instance or either of its tries will reference the same underlying trie structures.
func (tries *baseDualIPv4v6Tries[V]) ensureRoots() {
	if tries.ipv4Trie.getRoot() == nil {
		key := IPv4Address{}
		tries.ipv4Trie.trie.EnsureRoot(trieKey[*IPAddress]{key.ToIP()})
	}
	if tries.ipv6Trie.getRoot() == nil {
		key := IPv6Address{}
		tries.ipv6Trie.trie.EnsureRoot(trieKey[*IPAddress]{key.ToIP()})
	}
}

// equal returns whether the two given pair of tries is equal to this pair of tries
func (tries *baseDualIPv4v6Tries[V]) equal(other *baseDualIPv4v6Tries[V]) bool {
	return tries.ipv4Trie.equal(&other.ipv4Trie) && tries.ipv6Trie.equal(&other.ipv6Trie)
}

// Clone clones this pair of tries.
func (tries *baseDualIPv4v6Tries[V]) clone() baseDualIPv4v6Tries[V] {
	return baseDualIPv4v6Tries[V]{
		ipv4Trie: *toBase(tries.ipv4Trie.clone()),
		ipv6Trie: *toBase(tries.ipv6Trie.clone()),
	}
}

// Clear removes all added nodes from the tries, after which IsEmpty will return true.
func (tries *baseDualIPv4v6Tries[V]) Clear() {
	tries.ipv4Trie.clear()
	tries.ipv6Trie.clear()
}

// Size returns the number of elements in the tries.
// Only added nodes are counted.
// When zero is returned, IsEmpty() returns true.
func (tries *baseDualIPv4v6Tries[V]) Size() int {
	return tries.ipv4Trie.size() + tries.ipv6Trie.size()
}

// iIsEmpty returns true if there are no added nodes within the two tries
func (tries *baseDualIPv4v6Tries[V]) IsEmpty() bool {
	return tries.ipv4Trie.size() == 0 && tries.ipv6Trie.size() == 0
}

// Add adds the given single address or prefix block subnet to one of the two tries.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The [Partition] type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Add returns true if the prefix block or address was inserted, false if already in one of the two tries.
func (tries *baseDualIPv4v6Tries[V]) Add(addr *IPAddress) bool {
	return addressFuncOp(addr, tries.ipv4Trie.add, tries.ipv6Trie.add)
}

// Contains returns whether the given address or prefix block subnet is in one of the two tries (as an added element).
//
// If the argument is not a single address nor prefix block, this method will panic.
// The [Partition] type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Contains returns true if the prefix block or address address exists already in one the two tries, false otherwise.
//
// Use GetAddedNode to get the node for the address rather than just checking for its existence.
func (tries *baseDualIPv4v6Tries[V]) Contains(addr *IPAddress) bool {
	return addressFuncOp(addr, tries.ipv4Trie.contains, tries.ipv6Trie.contains)
}

// Remove Removes the given single address or prefix block subnet from the tries.
//
// Removing an element will not remove contained elements (nodes for contained blocks and addresses).
//
// If the argument is not a single address nor prefix block, this method will panic.
// The [Partition] type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the prefix block or address was removed, false if not already in one of the two tries.
//
// You can also remove by calling GetAddedNode to get the node and then calling Remove on the node.
//
// When an address is removed, the corresponding node may remain in the trie if it remains a subnet block for two sub-nodes.
// If the corresponding node can be removed from the trie, it will be.
func (tries *baseDualIPv4v6Tries[V]) Remove(addr *IPAddress) bool {
	return addressFuncOp(addr, tries.ipv4Trie.remove, tries.ipv6Trie.remove)
}

// ElementContains checks if a prefix block subnet or address in ones of the two tries contains the given subnet or address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The [Partition] type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// ElementContains returns true if the subnet or address is contained by a trie element, false otherwise.
//
// To get all the containing addresses, use ElementsContaining.
func (tries *baseDualIPv4v6Tries[V]) ElementContains(addr *IPAddress) bool {
	return addressFuncOp(addr, tries.ipv4Trie.elementContains, tries.ipv6Trie.elementContains)
}

func (tries *baseDualIPv4v6Tries[V]) elementsContaining(addr *IPAddress) *containmentPath[*IPAddress, V] {
	return addressFuncOp(addr, tries.ipv4Trie.elementsContaining, tries.ipv6Trie.elementsContaining)
}

func (tries *baseDualIPv4v6Tries[V]) elementsContainedBy(addr *IPAddress) *tree.BinTrieNode[trieKey[*IPAddress], V] {
	return addressFuncOp(addr, tries.ipv4Trie.elementsContainedBy, tries.ipv6Trie.elementsContainedBy)
}

func (tries *baseDualIPv4v6Tries[V]) removeElementsContainedBy(addr *IPAddress) *tree.BinTrieNode[trieKey[*IPAddress], V] {
	return addressFuncOp(addr, tries.ipv4Trie.removeElementsContainedBy, tries.ipv6Trie.removeElementsContainedBy)
}

func (tries *baseDualIPv4v6Tries[V]) getAddedNode(addr *IPAddress) *tree.BinTrieNode[trieKey[*IPAddress], V] {
	return addressFuncOp(addr, tries.ipv4Trie.getAddedNode, tries.ipv6Trie.getAddedNode)
}

func (tries *baseDualIPv4v6Tries[V]) longestPrefixMatchNode(addr *IPAddress) *tree.BinTrieNode[trieKey[*IPAddress], V] {
	return addressFuncOp(addr, tries.ipv4Trie.longestPrefixMatchNode, tries.ipv6Trie.longestPrefixMatchNode)
}

func (tries *baseDualIPv4v6Tries[V]) LongestPrefixMatch(addr *IPAddress) *IPAddress {
	return addressFuncOp(addr, tries.ipv4Trie.longestPrefixMatch, tries.ipv6Trie.longestPrefixMatch)
}

func (tries *baseDualIPv4v6Tries[V]) shortestPrefixMatchNode(addr *IPAddress) *tree.BinTrieNode[trieKey[*IPAddress], V] {
	return addressFuncOp(addr, tries.ipv4Trie.shortestPrefixMatchNode, tries.ipv6Trie.shortestPrefixMatchNode)
}

func (tries *baseDualIPv4v6Tries[V]) ShortestPrefixMatch(addr *IPAddress) *IPAddress {
	return addressFuncOp(addr, tries.ipv4Trie.shortestPrefixMatch, tries.ipv6Trie.shortestPrefixMatch)
}

func (tries *baseDualIPv4v6Tries[V]) addNode(addr *IPAddress) *tree.BinTrieNode[trieKey[*IPAddress], V] {
	return addressFuncOp(addr, tries.ipv4Trie.addNode, tries.ipv6Trie.addNode)
}

func (tries *baseDualIPv4v6Tries[V]) addTrie(trie *trieNode[*IPAddress, V]) *tree.BinTrieNode[trieKey[*IPAddress], V] {
	return switchOp(trie.getKey(), trie, tries.ipv4Trie.addTrie, tries.ipv6Trie.addTrie)
}

func (tries *baseDualIPv4v6Tries[V]) floorAddedNode(addr *IPAddress) *tree.BinTrieNode[trieKey[*IPAddress], V] {
	return addressFuncOp(addr, tries.ipv4Trie.floorAddedNode, tries.ipv6Trie.floorAddedNode)
}

func (tries *baseDualIPv4v6Tries[V]) lowerAddedNode(addr *IPAddress) *tree.BinTrieNode[trieKey[*IPAddress], V] {
	return addressFuncOp(addr, tries.ipv4Trie.lowerAddedNode, tries.ipv6Trie.lowerAddedNode)
}

func (tries *baseDualIPv4v6Tries[V]) ceilingAddedNode(addr *IPAddress) *tree.BinTrieNode[trieKey[*IPAddress], V] {
	return addressFuncOp(addr, tries.ipv4Trie.ceilingAddedNode, tries.ipv6Trie.ceilingAddedNode)
}

func (tries *baseDualIPv4v6Tries[V]) higherAddedNode(addr *IPAddress) *tree.BinTrieNode[trieKey[*IPAddress], V] {
	return addressFuncOp(addr, tries.ipv4Trie.higherAddedNode, tries.ipv6Trie.higherAddedNode)
}

func (tries *baseDualIPv4v6Tries[V]) Floor(addr *IPAddress) *IPAddress {
	return addressFuncOp(addr, tries.ipv4Trie.floor, tries.ipv6Trie.floor)
}

func (tries *baseDualIPv4v6Tries[V]) Lower(addr *IPAddress) *IPAddress {
	return addressFuncOp(addr, tries.ipv4Trie.lower, tries.ipv6Trie.lower)
}

func (tries *baseDualIPv4v6Tries[V]) Ceiling(addr *IPAddress) *IPAddress {
	return addressFuncOp(addr, tries.ipv4Trie.ceiling, tries.ipv6Trie.ceiling)
}

func (tries *baseDualIPv4v6Tries[V]) Higher(addr *IPAddress) *IPAddress {
	return addressFuncOp(addr, tries.ipv4Trie.higher, tries.ipv6Trie.higher)
}

func (tries *baseDualIPv4v6Tries[V]) Iterator() IteratorWithRemove[*IPAddress] {
	return addrTrieIteratorRem[*IPAddress, V]{tries.nodeIterator(true)}
}

func (tries *baseDualIPv4v6Tries[V]) DescendingIterator() IteratorWithRemove[*IPAddress] {
	return addrTrieIteratorRem[*IPAddress, V]{tries.nodeIterator(false)}
}

func (tries *baseDualIPv4v6Tries[V]) nodeIterator(forward bool) tree.TrieNodeIteratorRem[trieKey[*IPAddress], V] {
	tries.ensureRoots() // we need a change tracker from each trie to monitor for changes, and each change tracker is created by the trie root, so we need the roots
	ipv4Iterator := tries.ipv4Trie.nodeIterator(forward)
	ipv6Iterator := tries.ipv6Trie.nodeIterator(forward)
	return tree.CombineSequentiallyRem(tries.ipv4Trie.getRoot(), tries.ipv6Trie.getRoot(), ipv4Iterator, ipv6Iterator, forward)
}

func (tries *baseDualIPv4v6Tries[V]) containingFirstIterator(forwardSubNodeOrder bool) tree.TrieNodeIteratorRem[trieKey[*IPAddress], V] {
	tries.ensureRoots() // we need a change tracker from each trie to monitor for changes, and each change tracker is created by the trie root, so we need the roots
	ipv4Iterator := tries.ipv4Trie.containingFirstIterator(forwardSubNodeOrder)
	ipv6Iterator := tries.ipv6Trie.containingFirstIterator(forwardSubNodeOrder)
	return tree.CombineSequentiallyRem[trieKey[*IPAddress], V](tries.ipv4Trie.getRoot(), tries.ipv6Trie.getRoot(), ipv4Iterator, ipv6Iterator, forwardSubNodeOrder)
}

func (tries *baseDualIPv4v6Tries[V]) containedFirstIterator(forwardSubNodeOrder bool) tree.TrieNodeIteratorRem[trieKey[*IPAddress], V] {
	tries.ensureRoots()
	ipv4Iterator := tries.ipv4Trie.containedFirstIterator(forwardSubNodeOrder)
	ipv6Iterator := tries.ipv6Trie.containedFirstIterator(forwardSubNodeOrder)
	return tree.CombineSequentiallyRem(tries.ipv4Trie.getRoot(), tries.ipv6Trie.getRoot(), ipv4Iterator, ipv6Iterator, forwardSubNodeOrder)
}

func (tries *baseDualIPv4v6Tries[V]) blockSizeNodeIterator(lowerSubNodeFirst bool) tree.TrieNodeIteratorRem[trieKey[*IPAddress], V] {
	tries.ensureRoots()
	ipv4Iterator := tries.ipv4Trie.blockSizeNodeIterator(lowerSubNodeFirst)
	ipv6Iterator := tries.ipv6Trie.blockSizeNodeIterator(lowerSubNodeFirst)
	return tree.CombineByBlockSize(tries.ipv4Trie.getRoot(), tries.ipv6Trie.getRoot(), ipv4Iterator, ipv6Iterator, lowerSubNodeFirst)
}

func (tries baseDualIPv4v6Tries[V]) Format(state fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		// same call as String()
		str := tree.TreesString(true, tries.ipv4Trie.toTrie(), tries.ipv6Trie.toTrie())
		_, _ = state.Write([]byte(str))
		return
	}
	// We follow the same pattern as for single tries
	s := flagsFromState(state, verb)
	ipv4Str := fmt.Sprintf(s, (*trieBase[*IPAddress, V])(&tries.ipv4Trie))
	ipv6Str := fmt.Sprintf(s, (*trieBase[*IPAddress, V])(&tries.ipv6Trie))
	totalLen := len(ipv4Str) + len(ipv6Str) + 1
	bytes := make([]byte, totalLen+2)
	bytes[0] = '{'
	shifted := bytes[1:]
	copy(shifted, ipv4Str)
	shifted[len(ipv4Str)] = ' '
	shifted = shifted[len(ipv4Str)+1:]
	copy(shifted, ipv6Str)
	shifted[len(ipv6Str)] = '}'
	_, _ = state.Write(bytes)
}

// DualIPv4v6Tries maintains a pair of tries to store both IPv4 and IPv6 addresses and subnets.
type DualIPv4v6Tries struct {
	baseDualIPv4v6Tries[emptyValue]
}

func (tries *DualIPv4v6Tries) GetIPv4Trie() *Trie[*IPAddress] {
	tries.ensureRoots() // Since we are making a copy of ipv4Trie, we need to ensure the root exists, otherwise the returned trie will not share the same root
	return &Trie[*IPAddress]{tries.ipv4Trie}
}

func (tries *DualIPv4v6Tries) GetIPv6Trie() *Trie[*IPAddress] {
	tries.ensureRoots() // Since we are making a copy of ipv6Trie, we need to ensure the root exists, otherwise the returned trie will not share the same root
	return &Trie[*IPAddress]{tries.ipv6Trie}
}

// Equal returns whether the two given pair of tries is equal to this pair of tries
func (tries *DualIPv4v6Tries) Equal(other *DualIPv4v6Tries) bool {
	return tries.equal(&other.baseDualIPv4v6Tries)
}

// String returns a string representation of the pair of tries.
func (tries *DualIPv4v6Tries) String() string {
	return tries.TreeString(true)
}

// TreeString merges the trie strings of the two tries into a single merged trie string.
func (tries *DualIPv4v6Tries) TreeString(withNonAddedKeys bool) string {
	//tries.ensureRoots()
	if tries == nil {
		return nilString()
	}
	return tree.TreesString(withNonAddedKeys, tries.ipv4Trie.toTrie(), tries.ipv6Trie.toTrie())
}

func (tries *DualIPv4v6Tries) AddedNodesTreeString() string {
	//tries.ensureRoots()
	ipv4AddedTrie := tries.ipv4Trie.constructAddedNodesTree()
	ipv6AddedTrie := tries.ipv6Trie.constructAddedNodesTree()
	return tree.AddedNodesTreesString[trieKey[*IPAddress], emptyValue](&ipv4AddedTrie.trie, &ipv6AddedTrie.trie)
}

// Clone clones this pair of tries.
func (tries *DualIPv4v6Tries) Clone() *DualIPv4v6Tries {
	return &DualIPv4v6Tries{tries.clone()}
}

func (tries *DualIPv4v6Tries) ElementsContaining(addr *IPAddress) *ContainmentPath[*IPAddress] {
	return &ContainmentPath[*IPAddress]{*tries.elementsContaining(addr)} // *containmentPath[*IPAddress, tree.EmptyValueType]
}

func (tries *DualIPv4v6Tries) ElementsContainedBy(addr *IPAddress) *TrieNode[*IPAddress] {
	return toAddressTrieNode(tries.elementsContainedBy(addr))
}

func (tries *DualIPv4v6Tries) RemoveElementsContainedBy(addr *IPAddress) *TrieNode[*IPAddress] {
	return toAddressTrieNode(tries.removeElementsContainedBy(addr))
}

func (tries *DualIPv4v6Tries) GetAddedNode(addr *IPAddress) *TrieNode[*IPAddress] {
	return toAddressTrieNode(tries.getAddedNode(addr))
}

func (tries *DualIPv4v6Tries) LongestPrefixMatchNode(addr *IPAddress) *TrieNode[*IPAddress] {
	return toAddressTrieNode(tries.longestPrefixMatchNode(addr))
}

func (tries *DualIPv4v6Tries) ShortestPrefixMatchNode(addr *IPAddress) *TrieNode[*IPAddress] {
	return toAddressTrieNode(tries.shortestPrefixMatchNode(addr))
}

func (tries *DualIPv4v6Tries) AddNode(addr *IPAddress) *TrieNode[*IPAddress] {
	return toAddressTrieNode(tries.addNode(addr))
}

func (tries *DualIPv4v6Tries) AddTrie(trie *TrieNode[*IPAddress]) *TrieNode[*IPAddress] {
	return toAddressTrieNode(tries.addTrie(trie.toBase()))
}

func (tries *DualIPv4v6Tries) AddIPv6Trie(trie *TrieNode[*IPv6Address]) *TrieNode[*IPAddress] {
	return AddTrieToDual(tries, trie)
}

func (tries *DualIPv4v6Tries) AddIPv4Trie(trie *TrieNode[*IPv4Address]) *TrieNode[*IPAddress] {
	return AddTrieToDual(tries, trie)
}

func AddTrieToDual[R interface {
	TrieKeyConstraint[R]
	ToIP() *IPAddress
}](tries *DualIPv4v6Tries, trie *TrieNode[R]) *TrieNode[*IPAddress] {
	return toAddressTrieNode(addAssociativeTrieToDual(&tries.baseDualIPv4v6Tries, trie.toBase(), func(e emptyValue) emptyValue { return e }))
}

func (tries *DualIPv4v6Tries) FloorAddedNode(addr *IPAddress) *TrieNode[*IPAddress] {
	return toAddressTrieNode(tries.floorAddedNode(addr))
}

func (tries *DualIPv4v6Tries) LowerAddedNode(addr *IPAddress) *TrieNode[*IPAddress] {
	return toAddressTrieNode(tries.lowerAddedNode(addr))
}

func (tries *DualIPv4v6Tries) CeilingAddedNode(addr *IPAddress) *TrieNode[*IPAddress] {
	return toAddressTrieNode(tries.ceilingAddedNode(addr))
}

func (tries *DualIPv4v6Tries) HigherAddedNode(addr *IPAddress) *TrieNode[*IPAddress] {
	return toAddressTrieNode(tries.higherAddedNode(addr))
}

func (tries *DualIPv4v6Tries) NodeIterator(forward bool) IteratorWithRemove[*TrieNode[*IPAddress]] {
	return addrTrieNodeIteratorRem[*IPAddress, emptyValue]{tries.nodeIterator(forward)}
}

func (tries *DualIPv4v6Tries) ContainingFirstIterator(forwardSubNodeOrder bool) IteratorWithRemove[*TrieNode[*IPAddress]] {
	return addrTrieNodeIteratorRem[*IPAddress, emptyValue]{tries.containingFirstIterator(forwardSubNodeOrder)}
}

func (tries *DualIPv4v6Tries) ContainedFirstIterator(forwardSubNodeOrder bool) IteratorWithRemove[*TrieNode[*IPAddress]] {
	return addrTrieNodeIteratorRem[*IPAddress, emptyValue]{tries.containedFirstIterator(forwardSubNodeOrder)}
}

func (tries *DualIPv4v6Tries) BlockSizeNodeIterator(lowerSubNodeFirst bool) IteratorWithRemove[*TrieNode[*IPAddress]] {
	return addrTrieNodeIteratorRem[*IPAddress, emptyValue]{tries.blockSizeNodeIterator(lowerSubNodeFirst)}
}

//
//
//
//
//

// DualIPv4v6AssociativeTries maintains a pair of associative tries to map both IPv4 and IPv6 addresses and subnets to values of the value type V.
type DualIPv4v6AssociativeTries[V any] struct {
	baseDualIPv4v6Tries[V]
}

func (tries *DualIPv4v6AssociativeTries[V]) GetIPv4Trie() *AssociativeTrie[*IPAddress, V] {
	tries.ensureRoots() // Since we are making a copy of ipv4Trie, we need to ensure the root exists, otherwise the returned trie will not share the same root
	return &AssociativeTrie[*IPAddress, V]{tries.ipv4Trie}
}

func (tries *DualIPv4v6AssociativeTries[V]) GetIPv6Trie() *AssociativeTrie[*IPAddress, V] {
	tries.ensureRoots() // Since we are making a copy of ipv6Trie, we need to ensure the root exists, otherwise the returned trie will not share the same root
	return &AssociativeTrie[*IPAddress, V]{tries.ipv6Trie}
}

// Equal returns whether the two given pair of tries is equal to this pair of tries
func (tries *DualIPv4v6AssociativeTries[V]) Equal(other *DualIPv4v6AssociativeTries[V]) bool {
	return tries.equal(&other.baseDualIPv4v6Tries)
}

// DeepEqual returns whether the given argument is a trie with a set of nodes with the same keys as in this trie according to the Compare method,
// and the same values according to the reflect.DeepEqual method
func (tries *DualIPv4v6AssociativeTries[V]) DeepEqual(other *DualIPv4v6AssociativeTries[V]) bool {
	return tries.ipv4Trie.deepEqual(&other.ipv4Trie) && tries.ipv6Trie.deepEqual(&other.ipv6Trie)
}

// String returns a string representation of the pair of tries.
func (tries *DualIPv4v6AssociativeTries[V]) String() string {
	return tries.TreeString(true)
}

// TreeString merges the trie strings of the two tries into a single merged trie string.
func (tries *DualIPv4v6AssociativeTries[V]) TreeString(withNonAddedKeys bool) string {
	if tries == nil {
		return nilString()
	}
	return tree.TreesString(withNonAddedKeys, tries.ipv4Trie.toTrie(), tries.ipv6Trie.toTrie())
}

func (tries *DualIPv4v6AssociativeTries[V]) AddedNodesTreeString() string {
	ipv4AddedTrie := tries.ipv4Trie.constructAddedNodesTree()
	ipv6AddedTrie := tries.ipv6Trie.constructAddedNodesTree()
	return tree.AddedNodesTreesString[trieKey[*IPAddress], V](&ipv4AddedTrie.trie, &ipv6AddedTrie.trie)
}

// Clone clones this pair of tries.
func (tries *DualIPv4v6AssociativeTries[V]) Clone() *DualIPv4v6AssociativeTries[V] {
	return &DualIPv4v6AssociativeTries[V]{tries.clone()}
}

func (tries *DualIPv4v6AssociativeTries[V]) ElementsContaining(addr *IPAddress) *ContainmentValuesPath[*IPAddress, V] {
	return &ContainmentValuesPath[*IPAddress, V]{*tries.elementsContaining(addr)}
}

func (tries *DualIPv4v6AssociativeTries[V]) ElementsContainedBy(addr *IPAddress) *AssociativeTrieNode[*IPAddress, V] {
	return toAssociativeTrieNode(tries.elementsContainedBy(addr))
}

func (tries *DualIPv4v6AssociativeTries[V]) RemoveElementsContainedBy(addr *IPAddress) *AssociativeTrieNode[*IPAddress, V] {
	return toAssociativeTrieNode(tries.removeElementsContainedBy(addr))
}

func (tries *DualIPv4v6AssociativeTries[V]) GetAddedNode(addr *IPAddress) *AssociativeTrieNode[*IPAddress, V] {
	return toAssociativeTrieNode(tries.getAddedNode(addr))
}

func (tries *DualIPv4v6AssociativeTries[V]) LongestPrefixMatchNode(addr *IPAddress) *AssociativeTrieNode[*IPAddress, V] {
	return toAssociativeTrieNode(tries.longestPrefixMatchNode(addr))
}

func (tries *DualIPv4v6AssociativeTries[V]) ShortestPrefixMatchNode(addr *IPAddress) *AssociativeTrieNode[*IPAddress, V] {
	return toAssociativeTrieNode(tries.shortestPrefixMatchNode(addr))
}

func (tries *DualIPv4v6AssociativeTries[V]) AddNode(addr *IPAddress) *AssociativeTrieNode[*IPAddress, V] {
	return toAssociativeTrieNode(tries.addNode(addr))
}

func (tries *DualIPv4v6AssociativeTries[V]) AddTrie(trie *AssociativeTrieNode[*IPAddress, V]) *AssociativeTrieNode[*IPAddress, V] {
	return toAssociativeTrieNode(tries.addTrie(trie.toBase()))
}

func (tries *DualIPv4v6AssociativeTries[V]) AddIPv6Trie(trie *AssociativeTrieNode[*IPAddress, V]) *AssociativeTrieNode[*IPAddress, V] {
	return AddAssociativeTrieToDual(tries, trie, func(v V) V { return v })
}

func (tries *DualIPv4v6AssociativeTries[V]) AddIPv4Trie(trie *AssociativeTrieNode[*IPAddress, V]) *AssociativeTrieNode[*IPAddress, V] {
	return AddAssociativeTrieToDual(tries, trie, func(v V) V { return v })
}

// AddAssociativeTrie adds the given trie's entries to this trie.  The given trie's keys must have a ToIP() method like *IPV4Address or *IPv6Address.
// The given trie can map to different value types.  You must supply a function to map from the given trie's values to this trie's values.
// If you are using the same value type, then you can use DualIPv4v6AssociativeTries[V].AddIPv4Trie or DualIPv4v6AssociativeTries[V].AddIPv6Trie instead.
func AddAssociativeTrieToDual[R interface {
	TrieKeyConstraint[R]
	ToIP() *IPAddress
}, V, V2 any](tries *DualIPv4v6AssociativeTries[V], trie *AssociativeTrieNode[R, V2], valueMap func(v V2) V) *AssociativeTrieNode[*IPAddress, V] {
	return toAssociativeTrieNode(addAssociativeTrieToDual(&tries.baseDualIPv4v6Tries, trie.toBase(), valueMap))
}

func addAssociativeTrieToDual[R interface {
	TrieKeyConstraint[R]
	ToIP() *IPAddress
}, V, V2 any](tries *baseDualIPv4v6Tries[V], trie *trieNode[R, V2], valueMap func(v V2) V) *tree.BinTrieNode[trieKey[*IPAddress], V] {
	if trie == nil {
		return nil
	}
	var targetTrie *trieBase[*IPAddress, V]
	rootKey := trie.getKey().ToIP()
	if rootKey.IsIPv4() {
		targetTrie = &tries.ipv4Trie
	} else if rootKey.IsIPv6() {
		targetTrie = &tries.ipv6Trie
	}
	if targetTrie != nil {
		return tree.AddConvertibleTrie(
			&targetTrie.trie,
			trie.toBinTrieNode(),
			false,
			func(r trieKey[R]) trieKey[*IPAddress] { return trieKey[*IPAddress]{r.address.ToIP()} },
			valueMap)
	}
	return nil
}

func (tries *DualIPv4v6AssociativeTries[V]) FloorAddedNode(addr *IPAddress) *AssociativeTrieNode[*IPAddress, V] {
	return toAssociativeTrieNode(tries.floorAddedNode(addr))
}

func (tries *DualIPv4v6AssociativeTries[V]) LowerAddedNode(addr *IPAddress) *AssociativeTrieNode[*IPAddress, V] {
	return toAssociativeTrieNode(tries.lowerAddedNode(addr))
}

func (tries *DualIPv4v6AssociativeTries[V]) CeilingAddedNode(addr *IPAddress) *AssociativeTrieNode[*IPAddress, V] {
	return toAssociativeTrieNode(tries.ceilingAddedNode(addr))
}

func (tries *DualIPv4v6AssociativeTries[V]) HigherAddedNode(addr *IPAddress) *AssociativeTrieNode[*IPAddress, V] {
	return toAssociativeTrieNode(tries.higherAddedNode(addr))
}

func (tries *DualIPv4v6AssociativeTries[V]) NodeIterator(forward bool) IteratorWithRemove[*AssociativeTrieNode[*IPAddress, V]] {
	return associativeAddressTrieNodeIteratorRem[*IPAddress, V]{tries.nodeIterator(forward)}
}

func (tries *DualIPv4v6AssociativeTries[V]) ContainingFirstIterator(forwardSubNodeOrder bool) IteratorWithRemove[*AssociativeTrieNode[*IPAddress, V]] {
	return associativeAddressTrieNodeIteratorRem[*IPAddress, V]{tries.containingFirstIterator(forwardSubNodeOrder)}
}

func (tries *DualIPv4v6AssociativeTries[V]) ContainedFirstIterator(forwardSubNodeOrder bool) IteratorWithRemove[*AssociativeTrieNode[*IPAddress, V]] {
	return associativeAddressTrieNodeIteratorRem[*IPAddress, V]{tries.containedFirstIterator(forwardSubNodeOrder)}
}

func (tries *DualIPv4v6AssociativeTries[V]) BlockSizeNodeIterator(lowerSubNodeFirst bool) IteratorWithRemove[*AssociativeTrieNode[*IPAddress, V]] {
	return associativeAddressTrieNodeIteratorRem[*IPAddress, V]{tries.blockSizeNodeIterator(lowerSubNodeFirst)}
}

func (tries *DualIPv4v6AssociativeTries[V]) Get(addr *IPAddress) (V, bool) {
	return addressFuncDoubRetOp(addr, tries.ipv4Trie.get, tries.ipv6Trie.get)
}

func (tries *DualIPv4v6AssociativeTries[V]) Put(addr *IPAddress, value V) (V, bool) {
	return addressFuncDoubArgDoubleRetOp(addr, value, tries.ipv4Trie.put, tries.ipv6Trie.put)
}

func (tries *DualIPv4v6AssociativeTries[V]) PutNode(addr *IPAddress, value V) *AssociativeTrieNode[*IPAddress, V] {
	return toAssociativeTrieNode(addressFuncDoubArgOp(addr, value, tries.ipv4Trie.putNode, tries.ipv6Trie.putNode))
}

func (tries *DualIPv4v6AssociativeTries[V]) PutTrie(added *AssociativeTrieNode[*IPAddress, V]) *AssociativeTrieNode[*IPAddress, V] {
	return toAssociativeTrieNode(switchOp(added.GetKey(), added.toBinTrieNode(), tries.ipv4Trie.putTrie, tries.ipv6Trie.putTrie))
}

func (tries *DualIPv4v6AssociativeTries[V]) Remap(addr *IPAddress, remapper func(existingValue V, found bool) (mapped V, mapIt bool)) *AssociativeTrieNode[*IPAddress, V] {
	return toAssociativeTrieNode(addressFuncDoubArgOp(addr, remapper, tries.ipv4Trie.remap, tries.ipv6Trie.remap))
}

func (tries *DualIPv4v6AssociativeTries[V]) RemapIfAbsent(addr *IPAddress, supplier func() V) *AssociativeTrieNode[*IPAddress, V] {
	return toAssociativeTrieNode(addressFuncDoubArgOp(addr, supplier, tries.ipv4Trie.remapIfAbsent, tries.ipv6Trie.remapIfAbsent))
}

func addressFuncOp[T any](addr *IPAddress, ipv4Op, ipv6Op func(*IPAddress) T) T {
	return switchOp(addr, addr, ipv4Op, ipv6Op)
}

func switchOp[R interface {
	IsIPv4() bool
	IsIPv6() bool
}, S, T any](key R, arg S, ipv4Op, ipv6Op func(S) T) T {
	if key.IsIPv4() {
		return ipv4Op(arg)
	} else if key.IsIPv6() {
		return ipv6Op(arg)
	}
	var t T
	return t
}

// Get
func addressFuncDoubRetOp[T1, T2 any](addr *IPAddress, ipv4Op, ipv6Op func(*IPAddress) (T1, T2)) (T1, T2) {
	return switchDoubleRetOp(addr, addr, ipv4Op, ipv6Op)
}

func switchDoubleRetOp[R interface {
	IsIPv4() bool
	IsIPv6() bool
}, S, T1, T2 any](key R, arg S, ipv4Op, ipv6Op func(S) (T1, T2)) (T1, T2) {
	if key.IsIPv4() {
		return ipv4Op(arg)
	} else if key.IsIPv6() {
		return ipv6Op(arg)
	}
	var t1 T1
	var t2 T2
	return t1, t2
}

// PutNode, Remap, RemapIfAbsent
func addressFuncDoubArgOp[S, T any](addr *IPAddress, arg S, ipv4Op, ipv6Op func(*IPAddress, S) T) T {
	return switchDoubleArgOp(addr, addr, arg, ipv4Op, ipv6Op)
}

func switchDoubleArgOp[R interface {
	IsIPv4() bool
	IsIPv6() bool
}, S1, S2, T any](key R, arg1 S1, arg2 S2, ipv4Op, ipv6Op func(S1, S2) T) T {
	if key.IsIPv4() {
		return ipv4Op(arg1, arg2)
	} else if key.IsIPv6() {
		return ipv6Op(arg1, arg2)
	}
	var t T
	return t
}

// Put
func addressFuncDoubArgDoubleRetOp[S, T1, T2 any](addr *IPAddress, arg S, ipv4Op, ipv6Op func(*IPAddress, S) (T1, T2)) (T1, T2) {
	return switchDoubleArgDoubleRetOp(addr, addr, arg, ipv4Op, ipv6Op)
}

func switchDoubleArgDoubleRetOp[R interface {
	IsIPv4() bool
	IsIPv6() bool
}, S1, S2, T1, T2 any](key R, arg1 S1, arg2 S2, ipv4Op, ipv6Op func(S1, S2) (T1, T2)) (T1, T2) {
	if key.IsIPv4() {
		return ipv4Op(arg1, arg2)
	} else if key.IsIPv6() {
		return ipv6Op(arg1, arg2)
	}
	var t1 T1
	var t2 T2
	return t1, t2
}
