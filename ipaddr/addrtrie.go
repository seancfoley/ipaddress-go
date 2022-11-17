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

package ipaddr

import (
	"fmt"
	"github.com/seancfoley/bintree/tree"
	"github.com/seancfoley/ipaddress-go/ipaddr/addrerr"
	"unsafe"
)

type trieBase[T TrieKeyConstraint[T], V any] struct {
	trie tree.BinTrie
}

// Clear removes all added nodes from the tree, after which IsEmpty() will return true
func (trie *trieBase[T, V]) Clear() {
	trie.trie.Clear()
}

// getRoot returns the root node of this trie, which can be nil for an implicitly zero-valued uninitialized trie, but not for any other trie
func (trie *trieBase[T, V]) getRoot() *tree.BinTrieNode {
	return trie.trie.GetRoot()
}

func (trie *trieBase[T, V]) add(addr T) bool {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.Add(trieKey[T]{addr})
}

func (trie *trieBase[T, V]) addNode(addr T) *tree.BinTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.AddNode(trieKey[T]{addr})
}

// constructAddedNodesTree provides an associative trie in which the root and each added node are mapped to a list of their respective direct added sub-nodes.
// This trie provides an alternative non-binary tree structure of the added nodes.
// It is used by ToAddedNodesTreeString to produce a string showing the alternative structure.
// If there are no non-added nodes in this trie, then the alternative tree structure provided by this method is the same as the original trie.
func (trie *trieBase[T, V]) constructAddedNodesTree() trieBase[T, V] {
	return trieBase[T, V]{trie.trie.ConstructAddedNodesTree()}
}

func (trie *trieBase[T, V]) addTrie(added *trieNode[T, V]) *tree.BinTrieNode {
	return trie.trie.AddTrie(added.toBinTrieNode())
}

func (trie *trieBase[T, V]) contains(addr T) bool {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.Contains(trieKey[T]{addr})
}

func (trie *trieBase[T, V]) remove(addr T) bool {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.Remove(trieKey[T]{addr})
}

func (trie *trieBase[T, V]) removeElementsContainedBy(addr T) *tree.BinTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.RemoveElementsContainedBy(trieKey[T]{addr})
}

func (trie *trieBase[T, V]) elementsContainedBy(addr T) *tree.BinTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.ElementsContainedBy(trieKey[T]{addr})
}

func (trie *trieBase[T, V]) elementsContaining(addr T) *containmentPath[T, V] {
	addr = mustBeBlockOrAddress(addr)
	return toContainmentPathX[T, V](trie.trie.ElementsContaining(trieKey[T]{addr}))
}

func (trie *trieBase[T, V]) longestPrefixMatch(addr T) (t T) {
	addr = mustBeBlockOrAddress(addr)
	res := trie.trie.LongestPrefixMatch(trieKey[T]{addr})
	if res == nil {
		return
	}
	return res.(trieKey[T]).address
}

// only added nodes are added to the linked list
func (trie *trieBase[T, V]) longestPrefixMatchNode(addr T) *tree.BinTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.LongestPrefixMatchNode(trieKey[T]{addr})
}

func (trie *trieBase[T, V]) elementContains(addr T) bool {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.ElementContains(trieKey[T]{addr})
}

func (trie *trieBase[T, V]) getNode(addr T) *tree.BinTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.GetNode(trieKey[T]{addr})
}

func (trie *trieBase[T, V]) getAddedNode(addr T) *tree.BinTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.GetAddedNode(trieKey[T]{addr})
}

func (trie *trieBase[T, V]) iterator() Iterator[T] {
	if trie == nil {
		return nilAddressIterator[T]()
	}
	return addressKeyIterator[T]{trie.trie.Iterator()}
}

func (trie *trieBase[T, V]) descendingIterator() Iterator[T] {
	if trie == nil {
		return nilAddressIterator[T]()
	}
	return addressKeyIterator[T]{trie.trie.DescendingIterator()}
}

func (trie *trieBase[T, V]) nodeIterator(forward bool) tree.TrieNodeIteratorRem {
	return trie.toTrie().NodeIterator(forward)
}

func (trie *trieBase[T, V]) allNodeIterator(forward bool) tree.TrieNodeIteratorRem {
	return trie.toTrie().AllNodeIterator(forward)
}

// Iterates the added nodes in the trie, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *trieBase[T, V]) blockSizeNodeIterator(lowerSubNodeFirst bool) tree.TrieNodeIteratorRem {
	return trie.toTrie().BlockSizeNodeIterator(lowerSubNodeFirst)
}

// Iterates all nodes in the trie, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *trieBase[T, V]) blockSizeAllNodeIterator(lowerSubNodeFirst bool) tree.TrieNodeIteratorRem {
	return trie.toTrie().BlockSizeAllNodeIterator(lowerSubNodeFirst)
}

// Iterates all nodes, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
func (trie *trieBase[T, V]) blockSizeCachingAllNodeIterator() tree.CachingTrieNodeIterator {
	return trie.toTrie().BlockSizeCachingAllNodeIterator()
}

// Iterates all nodes, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
func (trie *trieBase[T, V]) containingFirstIterator(forwardSubNodeOrder bool) tree.CachingTrieNodeIterator {
	return trie.toTrie().ContainingFirstIterator(forwardSubNodeOrder)
}

func (trie *trieBase[T, V]) containingFirstAllNodeIterator(forwardSubNodeOrder bool) tree.CachingTrieNodeIterator {
	return trie.toTrie().ContainingFirstAllNodeIterator(forwardSubNodeOrder)
}

func (trie *trieBase[T, V]) containedFirstIterator(forwardSubNodeOrder bool) tree.TrieNodeIteratorRem {
	return trie.toTrie().ContainedFirstIterator(forwardSubNodeOrder)
}

func (trie *trieBase[T, V]) containedFirstAllNodeIterator(forwardSubNodeOrder bool) tree.TrieNodeIterator {
	return trie.toTrie().ContainedFirstAllNodeIterator(forwardSubNodeOrder)
}

func (trie *trieBase[T, V]) lowerAddedNode(addr T) *tree.BinTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.LowerAddedNode(trieKey[T]{addr})
}

func (trie *trieBase[T, V]) floorAddedNode(addr T) *tree.BinTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.FloorAddedNode(trieKey[T]{addr})
}

func (trie *trieBase[T, V]) higherAddedNode(addr T) *tree.BinTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.HigherAddedNode(trieKey[T]{addr})
}

func (trie *trieBase[T, V]) ceilingAddedNode(addr T) *tree.BinTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.CeilingAddedNode(trieKey[T]{addr})
}

func (trie *trieBase[T, V]) clone() *tree.BinTrie {
	return trie.toTrie().Clone()
}

func (trie *trieBase[T, V]) toTrie() *tree.BinTrie {
	return (*tree.BinTrie)(unsafe.Pointer(trie))
}

//
// Trie is a compact binary trie (aka compact binary prefix tree, or binary radix trie), for addresses and/or CIDR prefix block subnets.
// The prefixes in used by the prefix trie are the CIDR prefixes, or the full address in the case of individual addresses with no prefix length.
// The elements of the trie are CIDR prefix blocks or addresses.
//
// The zero-value of an AddressTrie is a trie ready for use.  Its root will be nil until an element is added to it.
// Once any subnet or address is added to the trie, it will have an assigned root, and any further addition to the trie must match the type and version of the root,
// in addition to the generic type of the trie's keys.
// Once there is a root, the root cannot be removed.
//
// So, for instance, an instance of ipaddr.Trie[*ipaddr.IPAddress] can contain either IPv4 or IPv6 keys, but not both.
// Once it has been populated with the first key, all remaining additions must have the same IP version, even if the trie is cleared.
//
// Any trie can be copied. If a trie has no root, a copy produces a new zero-valued trie with no root.
// If a trie has a root, a copy produces a reference to the same trie, much like copying a map or slice.
//
// The trie data structure allows you to check an address for containment in many subnets at once, in constant time.
// The trie allows you to check a subnet for containment of many smaller subnets or addresses at once, in constant time.
// The trie allows you to check for equality of a subnet or address with a large number of subnets or addresses at once.
//
// There is only a single possible trie for any given set of address and subnets.  For one thing, this means they are automatically balanced.
// Also, this makes access to subtries and to the nodes themselves more useful, allowing for many of the same operations performed on the original trie.
//
// Each node has either a prefix block or a single address as its key.
// Each prefix block node can have two sub-nodes, each sub-node a prefix block or address contained by the node.
//
// There are more nodes in the trie than elements added to the trie.
// A node is considered "added" if it was explicitly added to the trie and is included as an element when viewed as a set.
// There are non-added prefix block nodes that are generated in the trie as well.
// When two or more added addresses share the same prefix up until they differ with the bit at index x,
// then a prefix block node is generated (if not already added to the trie) for the common prefix of length x,
// with the nodes for those addresses to be found following the lower
// or upper sub-nodes according to the bit at index x + 1 in each address.
// If that bit is 1, the node can be found by following the upper sub-node,
// and when it is 0, the lower sub-node.
//
// Nodes that were generated as part of the trie structure only
// because of other added elements are not elements of the represented set of addresses and subnets.
// The set elements are the elements that were explicitly added.
//
// You can work with parts of the trie, starting from any node in the trie,
// calling methods that start with any given node, such as iterating the subtrie,
// finding the first or last in the subtrie, doing containment checks with the subtrie, and so on.
//
// The binary trie structure defines a natural ordering of the trie elements.
// Addresses of equal prefix length are sorted by prefix value.  Addresses with no prefix length are sorted by address value.
// Addresses of differing prefix length are sorted according to the bit that follows the shorter prefix length in the address with the longer prefix length,
// whether that bit is 0 or 1 determines if that address is ordered before or after the address of shorter prefix length.
//
// The unique and pre-defined structure for a trie means that different means of traversing the trie can be more meaningful.
// This trie implementation provides 8 different ways of iterating through the trie:
//  1, 2: the natural sorted trie order, forward and reverse (spliterating is also an option for these two orders).  Use the methods NodeIterator, Iterator or DescendingIterator.  Functions for incrementing and decrementing keys, or comparing keys, is also provided for this order.
//  3, 4: pre-order tree traversal, in which parent node is visited before sub-nodes, with sub-nodes visited in forward or reverse order
//  5, 6: post-order tree traversal, in which sub-nodes are visited before parent nodes, with sub-nodes visited in forward or reverse order
//  7, 8: prefix-block order, in which larger prefix blocks are visited before smaller, and blocks of equal size are visited in forward or reverse sorted order
//
// All of these orderings are useful in specific contexts.
//
// If you create an iterator, then that iterator can no longer be advanced following any further modification to the trie.
// Any call to Next or Remove will panic if the trie was changed following creation of the iterator.
//
// You can do lookup and containment checks on all the subnets and addresses in the trie at once, in constant time.
// A generic trie data structure lookup is O(m) where m is the entry length.
// For this trie, which operates on address bits, entry length is capped at 128 bits for IPv6 and 32 bits for IPv4.
// That makes lookup a constant time operation.
// Subnet containment or equality checks are also constant time since they work the same way as lookup, by comparing prefix bits.
//
// For a generic trie data structure, construction is O(m * n) where m is entry length and n is the number of addresses,
// but for this trie, since entry length is capped at 128 bits for IPv6 and 32 bits for IPv4, construction is O(n),
// in linear proportion to the number of added elements.
//
// This trie also allows for constant time size queries (count of added elements, not node count), by storing sub-trie size in each node.
// It works by updating the size of every node in the path to any added or removed node.
// This does not change insertion or deletion operations from being constant time (because tree-depth is limited to address bit count).
// At the same this makes size queries constant time, rather than being O(n) time.
//
// A single trie can use just a single address type or version, since it works with bits alone,
// and cannot distinguish between different versions and types in the trie structure.
//
// Instead, you could aggregate multiple subtries to create a collection of multiple address types or versions.
// You can use the method ToString for a String that represents multiple tries as a single tree.
//
// Tries are concurrency-safe when not being modified (elements added or removed), but are not concurrency-safe when any goroutine is modifying the trie.
type Trie[T TrieKeyConstraint[T]] struct {
	trieBase[T, emptyValue]
}

func (trie *Trie[T]) tobase() *trieBase[T, emptyValue] {
	return (*trieBase[T, emptyValue])(unsafe.Pointer(trie))
}

// GetRoot returns the root node of this trie, which can be nil for an implicitly zero-valued uninitialized trie, but not for any other trie
func (trie *Trie[T]) GetRoot() *TrieNode[T] {
	return toAddressTrieNode[T](trie.getRoot())
}

// Size returns the number of elements in the tree.
// It does not return the number of nodes.
// Only nodes for which IsAdded() returns true are counted (those nodes corresponding to added addresses and prefix blocks).
// When zero is returned, IsEmpty() returns true.
func (trie *Trie[T]) Size() int {
	return trie.toTrie().Size()
}

// NodeSize returns the number of nodes in the tree, which is always more than the number of elements.
func (trie *Trie[T]) NodeSize() int {
	return trie.toTrie().NodeSize()
}

// IsEmpty returns true if there are not any added nodes within this tree
func (trie *Trie[T]) IsEmpty() bool {
	return trie.Size() == 0
}

// TreeString returns a visual representation of the tree with one node per line, with or without the non-added keys.
func (trie *Trie[T]) TreeString(withNonAddedKeys bool) string {
	return trie.toTrie().TreeString(withNonAddedKeys)
}

// String returns a visual representation of the tree with one node per line.
func (trie *Trie[T]) String() string {
	return trie.toTrie().String()
}

// AddedNodesTreeString provides a flattened version of the trie showing only the contained added nodes and their containment structure, which is non-binary.
// The root node is included, which may or may not be added.
func (trie *Trie[T]) AddedNodesTreeString() string {
	return trie.toTrie().AddedNodesTreeString()
}

// Add adds the address to this trie.
// The address must match the same type and version of any existing addresses already in the trie.
// Returns true if the address did not already exist in the trie.
func (trie *Trie[T]) Add(addr T) bool {
	return trie.add(addr)
}

// AddNode adds the address to this trie.
// The address must match the same type and version of any existing addresses already in the trie.
// The new or existing node for the address is returned.
func (trie *Trie[T]) AddNode(addr T) *TrieNode[T] {
	return toAddressTrieNode[T](trie.addNode(addr))
}

// AddTrie adds nodes for the keys in the trie with the root node as the passed in node.  To add both keys and values, use PutTrie.
func (trie *Trie[T]) AddTrie(added *TrieNode[T]) *TrieNode[T] {
	return toAddressTrieNode[T](trie.addTrie(added.tobase()))
}

// ConstructAddedNodesTree provides an associative trie in which the root and each added node are mapped to a list of their respective direct added sub-nodes.
// This trie provides an alternative non-binary tree structure of the added nodes.
// It is used by ToAddedNodesTreeString to produce a string showing the alternative structure.
// If there are no non-added nodes in this trie, then the alternative tree structure provided by this method is the same as the original trie.
func (trie *Trie[T]) ConstructAddedNodesTree() *Trie[T] {
	return &Trie[T]{trie.constructAddedNodesTree()}
}

// Contains returns whether the given address or prefix block subnet is in the trie as an added element.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the prefix block or address exists already in the trie, false otherwise.
//
// Use GetAddedNode to get the node for the address rather than just checking for its existence.
func (trie *Trie[T]) Contains(addr T) bool {
	return trie.contains(addr)
}

// Remove removes the given single address or prefix block subnet from the trie.
//
// Removing an element will not remove contained elements (nodes for contained blocks and addresses).
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the prefix block or address was removed, false if not already in the trie.
//
// You can also remove by calling GetAddedNode to get the node and then calling Remove on the node.
//
// When an address is removed, the corresponding node may remain in the trie if it remains a subnet block for two sub-nodes.
// If the corresponding node can be removed from the trie, it will be.
func (trie *Trie[T]) Remove(addr T) bool {
	return trie.remove(addr)
}

// RemoveElementsContainedBy removes any single address or prefix block subnet from the trie that is contained in the given individual address or prefix block subnet.
//
// This goes further than Remove, not requiring a match to an inserted node, and also removing all the sub-nodes of any removed node or sub-node.
//
// For example, after inserting 1.2.3.0 and 1.2.3.1, passing 1.2.3.0/31 to RemoveElementsContainedBy will remove them both,
// while the Remove method will remove nothing.
// After inserting 1.2.3.0/31, then Remove will remove 1.2.3.0/31, but will leave 1.2.3.0 and 1.2.3.1 in the trie.
//
// It cannot partially delete a node, such as deleting a single address from a prefix block represented by a node.
// It can only delete the whole node if the whole address or block represented by that node is contained in the given address or block.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
//Returns the root node of the sub-trie that was removed from the trie, or nil if nothing was removed.
func (trie *Trie[T]) RemoveElementsContainedBy(addr T) *TrieNode[T] {
	return toAddressTrieNode[T](trie.removeElementsContainedBy(addr))
}

// ElementsContainedBy checks if a part of this trie is contained by the given prefix block subnet or individual address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the root node of the contained sub-trie, or nil if no sub-trie is contained.
// The node returned need not be an "added" node, see IsAdded for more details on added nodes.
// The returned sub-trie is backed by this trie, so changes in this trie are reflected in those nodes and vice-versa.
func (trie *Trie[T]) ElementsContainedBy(addr T) *TrieNode[T] {
	return toAddressTrieNode[T](trie.elementsContainedBy(addr))
}

// ElementsContaining finds the trie nodes in the trie containing the given key and returns them as a linked list.
// Only added nodes are added to the linked list
//
// If the argument is not a single address nor prefix block, this method will panic.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (trie *Trie[T]) ElementsContaining(addr T) *ContainmentPath[T] {
	return &ContainmentPath[T]{*trie.elementsContaining(addr)}
}

// LongestPrefixMatch returns the address added to the trie with the longest matching prefix compared to the provided address, or nil if no matching address
func (trie *Trie[T]) LongestPrefixMatch(addr T) T {
	return trie.longestPrefixMatch(addr)
}

// only added nodes are added to the linked list

// LongestPrefixMatchNode returns the node of address added to the trie with the longest matching prefix compared to the provided address, or nil if no matching address
func (trie *Trie[T]) LongestPrefixMatchNode(addr T) *TrieNode[T] {
	return toAddressTrieNode[T](trie.longestPrefixMatchNode(addr))
}

// ElementContains checks if a prefix block subnet or address in the trie contains the given subnet or address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the subnet or address is contained by a trie element, false otherwise.
//
// To get all the containing addresses, use ElementsContaining
func (trie *Trie[T]) ElementContains(addr T) bool {
	return trie.elementContains(addr)
}

// GetNode gets the node in the trie corresponding to the given address,
// or returns nil if not such element exists.
//
// It returns any node, whether added or not,
// including any prefix block node that was not added.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (trie *Trie[T]) GetNode(addr T) *TrieNode[T] {
	return toAddressTrieNode[T](trie.getNode(addr))
}

// GetAddedNode gets trie nodes representing added elements.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Use Contains to check for the existence of a given address in the trie,
// as well as GetNode to search for all nodes including those not-added but also auto-generated nodes for subnet blocks.
func (trie *Trie[T]) GetAddedNode(addr T) *TrieNode[T] {
	return toAddressTrieNode[T](trie.getAddedNode(addr))
}

// Iterator returns an iterator that iterates through the added addresses and prefix blocks in the trie.
// The iteration is in sorted element order.
func (trie *Trie[T]) Iterator() Iterator[T] {
	return trie.tobase().iterator()
}

// DescendingIterator returns an iterator that iterates through the added addresses and prefix blocks in the trie.
// The iteration is in reverse sorted element order.
func (trie *Trie[T]) DescendingIterator() Iterator[T] {
	return trie.tobase().descendingIterator()
}

// NodeIterator returns an iterator that iterates through all the added nodes in the trie in forward or reverse tree order.
func (trie *Trie[T]) NodeIterator(forward bool) IteratorWithRemove[*TrieNode[T]] {
	return addrTrieNodeIteratorRem[T]{trie.tobase().nodeIterator(forward)}
}

// AllNodeIterator returns an iterator that iterates through all the nodes in the trie in forward or reverse tree order.
func (trie *Trie[T]) AllNodeIterator(forward bool) IteratorWithRemove[*TrieNode[T]] {
	return addrTrieNodeIteratorRem[T]{trie.tobase().allNodeIterator(forward)}
}

// BlockSizeNodeIterator returns an iterator that iterates the added nodes in the trie, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *Trie[T]) BlockSizeNodeIterator(lowerSubNodeFirst bool) IteratorWithRemove[*TrieNode[T]] {
	return addrTrieNodeIteratorRem[T]{trie.tobase().blockSizeNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeAllNodeIterator returns an iterator that iterates all nodes in the trie, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *Trie[T]) BlockSizeAllNodeIterator(lowerSubNodeFirst bool) IteratorWithRemove[*TrieNode[T]] {
	return addrTrieNodeIteratorRem[T]{trie.tobase().blockSizeAllNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeCachingAllNodeIterator returns an iterator that iterates all nodes, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
func (trie *Trie[T]) BlockSizeCachingAllNodeIterator() CachingTrieIterator[*TrieNode[T]] {
	return cachingAddressTrieNodeIterator[T]{trie.tobase().blockSizeCachingAllNodeIterator()}
}

// ContainingFirstIterator returns an iterator that does a pre-order binary tree traversal of the added nodes.
// All added nodes will be visited before their added sub-nodes.
// For an address trie this means added containing subnet blocks will be visited before their added contained addresses and subnet blocks.
//
// Once a given node is visited, the iterator allows you to cache an object corresponding to the
// lower or upper sub-node that can be retrieved when you later visit that sub-node.
//
// Objects are cached only with nodes to be visited.
// So for this iterator that means an object will be cached with the first added lower or upper sub-node,
// the next lower or upper sub-node to be visited,
// which is not necessarily the direct lower or upper sub-node of a given node.
//
// The caching allows you to provide iteration context from a parent to its sub-nodes when iterating.
// The caching and retrieval is done in constant-time.
func (trie *Trie[T]) ContainingFirstIterator(forwardSubNodeOrder bool) CachingTrieIterator[*TrieNode[T]] {
	return cachingAddressTrieNodeIterator[T]{trie.tobase().containingFirstIterator(forwardSubNodeOrder)}
}

// ContainingFirstAllNodeIterator returns an iterator that does a pre-order binary tree traversal.
// All nodes will be visited before their sub-nodes.
// For an address trie this means containing subnet blocks will be visited before their contained addresses and subnet blocks.
//
// Once a given node is visited, the iterator allows you to cache an object corresponding to the
// lower or upper sub-node that can be retrieved when you later visit that sub-node.
// That allows you to provide iteration context from a parent to its sub-nodes when iterating.
// The caching and retrieval is done in constant time.
func (trie *Trie[T]) ContainingFirstAllNodeIterator(forwardSubNodeOrder bool) CachingTrieIterator[*TrieNode[T]] {
	return cachingAddressTrieNodeIterator[T]{trie.tobase().containingFirstAllNodeIterator(forwardSubNodeOrder)}
}

// ContainedFirstIterator returns an iterator that does a post-order binary tree traversal of the added nodes.
// All added sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (trie *Trie[T]) ContainedFirstIterator(forwardSubNodeOrder bool) IteratorWithRemove[*TrieNode[T]] {
	return addrTrieNodeIteratorRem[T]{trie.tobase().containedFirstIterator(forwardSubNodeOrder)}
}

// ContainedFirstAllNodeIterator returns an iterator that does a post-order binary tree traversal.
// All sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (trie *Trie[T]) ContainedFirstAllNodeIterator(forwardSubNodeOrder bool) Iterator[*TrieNode[T]] {
	return addrTrieNodeIterator[T]{trie.tobase().containedFirstAllNodeIterator(forwardSubNodeOrder)}
}

// FirstNode returns the first (lowest valued) node in the trie
func (trie *Trie[T]) FirstNode() *TrieNode[T] {
	return toAddressTrieNode[T](trie.trieBase.trie.FirstNode())
}

// FirstAddedNode returns the first (lowest valued) added node in the trie,
// or nil if there are no added entries in this tree or sub-tree
func (trie *Trie[T]) FirstAddedNode() *TrieNode[T] {
	return toAddressTrieNode[T](trie.trieBase.trie.FirstAddedNode())
}

// LastNode returns the last (highest valued) node in the trie
func (trie *Trie[T]) LastNode() *TrieNode[T] {
	return toAddressTrieNode[T](trie.trieBase.trie.LastNode())
}

// LastAddedNode returns the last (highest valued) added node in the trie,
// or nil if there are no added entries in this tree or sub-tree
func (trie *Trie[T]) LastAddedNode() *TrieNode[T] {
	return toAddressTrieNode[T](trie.trieBase.trie.LastAddedNode())
}

// LowerAddedNode returns the added node whose address is the highest address strictly less than the given address.
func (trie *Trie[T]) LowerAddedNode(addr T) *TrieNode[T] {
	return toAddressTrieNode[T](trie.lowerAddedNode(addr))
}

// FloorAddedNode returns the added node whose address is the highest address less than or equal to the given address.
func (trie *Trie[T]) FloorAddedNode(addr T) *TrieNode[T] {
	return toAddressTrieNode[T](trie.floorAddedNode(addr))
}

// HigherAddedNode returns the added node whose address is the lowest address strictly greater than the given address.
func (trie *Trie[T]) HigherAddedNode(addr T) *TrieNode[T] {
	return toAddressTrieNode[T](trie.higherAddedNode(addr))
}

// CeilingAddedNode returns the added node whose address is the lowest address greater than or equal to the given address.
func (trie *Trie[T]) CeilingAddedNode(addr T) *TrieNode[T] {
	return toAddressTrieNode[T](trie.ceilingAddedNode(addr))
}

// Clone clones this trie
func (trie *Trie[T]) Clone() *Trie[T] {
	return toAddressTrie[T](trie.tobase().clone())
}

// Equal returns whether the given argument is a trie with a set of nodes with the same keys as in this trie
func (trie *Trie[T]) Equal(other *Trie[T]) bool {
	return trie.toTrie().Equal(other.toTrie())
}

// For some reason Format must be here and not in addressTrieNode for nil node.
// It panics in fmt code either way, but if in here then it is handled by a recover() call in fmt properly.
// Seems to be a problem only in the debugger.

// Format implements the fmt.Formatter interface.
func (trie Trie[T]) Format(state fmt.State, verb rune) {
	// without this, prints like {{{{<nil>}}}} or {{{{0xc00014ca50}}}}
	// which is done by printValue in print.go of fmt
	trie.trieBase.trie.Format(state, verb)
}

func TreesString[T TrieKeyConstraint[T]](withNonAddedKeys bool, tries ...*Trie[T]) string {
	binTries := make([]*tree.BinTrie, 0, len(tries))
	for _, trie := range tries {
		binTries = append(binTries, toBinTrie[T](trie))
	}
	return tree.TreesString(withNonAddedKeys, binTries...)
}

func toBinTrie[T TrieKeyConstraint[T]](trie *Trie[T]) *tree.BinTrie {
	return (*tree.BinTrie)(unsafe.Pointer(trie))
}

func toAddressTrie[T TrieKeyConstraint[T]](trie *tree.BinTrie) *Trie[T] {
	return (*Trie[T])(unsafe.Pointer(trie))
}

////////
////////
////////
////////
////////
////////
////////
////////

// AssociativeTrie represents a binary address trie in which each added node can be associated with a value.
// It is an instance of Trie that can also function as a key-value map. The keys are addresses or prefix blocks.
// Each can be mapped to a value with type specified by the generic type V.
//
// All the characteristics of Trie are common to AssociativeTrie.
//
// The zero value is a binary trie ready for use.
type AssociativeTrie[T TrieKeyConstraint[T], V any] struct {
	trieBase[T, V]
}

//func (trie *AssociativeAddressTrie) toTrie() *tree.BinTrie {
//	return (*tree.BinTrie)(unsafe.Pointer(trie))
//}

func (trie *AssociativeTrie[T, V]) tobase() *trieBase[T, V] {
	return (*trieBase[T, V])(unsafe.Pointer(trie))
}

// GetRoot returns the root node of this trie, which can be nil for an implicitly zero-valued uninitialized trie, but not for any other trie
func (trie *AssociativeTrie[T, V]) GetRoot() *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](trie.getRoot())
}

// Size returns the number of elements in the tree.
// It does not return the number of nodes.
// Only nodes for which IsAdded() returns true are counted (those nodes corresponding to added addresses and prefix blocks).
// When zero is returned, IsEmpty() returns true.
func (trie *AssociativeTrie[T, V]) Size() int {
	return trie.toTrie().Size()
}

// NodeSize returns the number of nodes in the tree, which is always more than the number of elements.
func (trie *AssociativeTrie[T, V]) NodeSize() int {
	return trie.toTrie().NodeSize()
}

// IsEmpty returns true if there are not any added nodes within this tree
func (trie *AssociativeTrie[T, V]) IsEmpty() bool {
	return trie.Size() == 0
}

// TreeString returns a visual representation of the tree with one node per line, with or without the non-added keys.
func (trie *AssociativeTrie[T, V]) TreeString(withNonAddedKeys bool) string {
	return trie.toTrie().TreeString(withNonAddedKeys)
}

// String returns a visual representation of the tree with one node per line.
func (trie *AssociativeTrie[T, V]) String() string {
	return trie.toTrie().String()
}

// AddedNodesTreeString provides a flattened version of the trie showing only the contained added nodes and their containment structure, which is non-binary.
// The root node is included, which may or may not be added.
func (trie *AssociativeTrie[T, V]) AddedNodesTreeString() string {
	return trie.toTrie().AddedNodesTreeString()
}

// Add adds the address to this trie.
// Returns true if the address did not already exist in the trie.
func (trie *AssociativeTrie[T, V]) Add(addr T) bool {
	return trie.add(addr)
}

// AddNode adds the address key to this trie.
// The new or existing node for the address is returned.
func (trie *AssociativeTrie[T, V]) AddNode(addr T) *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](trie.addNode(addr))
}

// AddTrie adds nodes for the keys in the trie with the root node as the passed in node.  To add both keys and values, use PutTrie.
func (trie *AssociativeTrie[T, V]) AddTrie(added *AssociativeTrieNode[T, V]) *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](trie.addTrie(added.toBase()))
}

// ConstructAddedNodesTree provides an associative trie in which the root and each added node are mapped to a list of their respective direct added sub-nodes.
// This trie provides an alternative non-binary tree structure of the added nodes.
// It is used by ToAddedNodesTreeString to produce a string showing the alternative structure.
// If there are no non-added nodes in this trie, then the alternative tree structure provided by this method is the same as the original trie.
func (trie *AssociativeTrie[T, V]) ConstructAddedNodesTree() *AssociativeTrie[T, V] {
	return &AssociativeTrie[T, V]{trie.constructAddedNodesTree()}
}

// Contains returns whether the given address or prefix block subnet is in the trie as an added element.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the prefix block or address exists already in the trie, false otherwise.
//
// Use GetAddedNode  to get the node for the address rather than just checking for its existence.
func (trie *AssociativeTrie[T, V]) Contains(addr T) bool {
	return trie.contains(addr)
}

// Remove removes the given single address or prefix block subnet from the trie.
//
// Removing an element will not remove contained elements (nodes for contained blocks and addresses).
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the prefix block or address was removed, false if not already in the trie.
//
// You can also remove by calling GetAddedNode to get the node and then calling Remove on the node.
//
// When an address is removed, the corresponding node may remain in the trie if it remains a subnet block for two sub-nodes.
// If the corresponding node can be removed from the trie, it will be.
func (trie *AssociativeTrie[T, V]) Remove(addr T) bool {
	return trie.remove(addr)
}

// RemoveElementsContainedBy removes any single address or prefix block subnet from the trie that is contained in the given individual address or prefix block subnet.
//
// This goes further than Remove, not requiring a match to an inserted node, and also removing all the sub-nodes of any removed node or sub-node.
//
// For example, after inserting 1.2.3.0 and 1.2.3.1, passing 1.2.3.0/31 to RemoveElementsContainedBy will remove them both,
// while the Remove method will remove nothing.
// After inserting 1.2.3.0/31, then Remove will remove 1.2.3.0/31, but will leave 1.2.3.0 and 1.2.3.1 in the trie.
//
// It cannot partially delete a node, such as deleting a single address from a prefix block represented by a node.
// It can only delete the whole node if the whole address or block represented by that node is contained in the given address or block.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
//Returns the root node of the sub-trie that was removed from the trie, or nil if nothing was removed.
func (trie *AssociativeTrie[T, V]) RemoveElementsContainedBy(addr T) *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](trie.removeElementsContainedBy(addr))
}

// ElementsContainedBy checks if a part of this trie is contained by the given prefix block subnet or individual address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the root node of the contained sub-trie, or nil if no sub-trie is contained.
// The node returned need not be an "added" node, see IsAdded for more details on added nodes.
// The returned sub-trie is backed by this trie, so changes in this trie are reflected in those nodes and vice-versa.
func (trie *AssociativeTrie[T, V]) ElementsContainedBy(addr T) *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](trie.elementsContainedBy(addr))
}

// ElementsContaining finds the trie nodes in the trie containing the given key and returns them as a linked list.
// Only added nodes are added to the linked list
//
// If the argument is not a single address nor prefix block, this method will panic.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (trie *AssociativeTrie[T, V]) ElementsContaining(addr T) *ContainmentValuesPath[T, V] {
	return &ContainmentValuesPath[T, V]{*trie.elementsContaining(addr)}
}

// LongestPrefixMatch returns the address with the longest matching prefix compared to the provided address
func (trie *AssociativeTrie[T, V]) LongestPrefixMatch(addr T) T {
	return trie.longestPrefixMatch(addr)
}

// only added nodes are added to the linked list

// LongestPrefixMatchNode returns the node of the address with the longest matching prefix compared to the provided address
func (trie *AssociativeTrie[T, V]) LongestPrefixMatchNode(addr T) *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](trie.longestPrefixMatchNode(addr))
}

// ElementContains checks if a prefix block subnet or address in the trie contains the given subnet or address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the subnet or address is contained by a trie element, false otherwise.
//
// To get all the containing addresses, use ElementsContaining
func (trie *AssociativeTrie[T, V]) ElementContains(addr T) bool {
	return trie.elementContains(addr)
}

// GetNode gets the node in the trie corresponding to the given address,
// or returns nil if not such element exists.
//
// It returns any node, whether added or not,
// including any prefix block node that was not added.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (trie *AssociativeTrie[T, V]) GetNode(addr T) *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](trie.getNode(addr))
}

// GetAddedNode gets trie nodes representing added elements.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Use Contains to check for the existence of a given address in the trie,
// as well as GetNode to search for all nodes including those not added but also auto-generated nodes for subnet blocks.
func (trie *AssociativeTrie[T, V]) GetAddedNode(addr T) *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](trie.getAddedNode(addr))
}

// Iterator returns an iterator that iterates through the added addresses and prefix blocks in the trie.
// The iteration is in sorted element order.
func (trie *AssociativeTrie[T, V]) Iterator() Iterator[T] {
	return trie.tobase().iterator()
}

// DescendingIterator returns an iterator that iterates through the added addresses and prefix blocks in the trie.
// The iteration is in reverse sorted element order.
func (trie *AssociativeTrie[T, V]) DescendingIterator() Iterator[T] {
	return trie.tobase().descendingIterator()
}

// NodeIterator returns an iterator that iterates through all the added nodes in the trie in forward or reverse tree order.
func (trie *AssociativeTrie[T, V]) NodeIterator(forward bool) IteratorWithRemove[*AssociativeTrieNode[T, V]] {
	return associativeAddressTrieNodeIteratorRem[T, V]{trie.tobase().nodeIterator(forward)}
}

// AllNodeIterator returns an iterator that iterates through all the nodes in the trie in forward or reverse tree order.
func (trie *AssociativeTrie[T, V]) AllNodeIterator(forward bool) IteratorWithRemove[*AssociativeTrieNode[T, V]] {
	return associativeAddressTrieNodeIteratorRem[T, V]{trie.tobase().allNodeIterator(forward)}
}

// BlockSizeNodeIterator returns an iterator that iterates the added nodes in the trie, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *AssociativeTrie[T, V]) BlockSizeNodeIterator(lowerSubNodeFirst bool) IteratorWithRemove[*AssociativeTrieNode[T, V]] {
	return associativeAddressTrieNodeIteratorRem[T, V]{trie.tobase().blockSizeNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeAllNodeIterator returns an iterator that iterates all nodes in the trie, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *AssociativeTrie[T, V]) BlockSizeAllNodeIterator(lowerSubNodeFirst bool) IteratorWithRemove[*AssociativeTrieNode[T, V]] {
	return associativeAddressTrieNodeIteratorRem[T, V]{trie.tobase().blockSizeAllNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeCachingAllNodeIterator returns an iterator that iterates all nodes, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
func (trie *AssociativeTrie[T, V]) BlockSizeCachingAllNodeIterator() CachingTrieIterator[*AssociativeTrieNode[T, V]] {
	return cachingAssociativeAddressTrieNodeIteratorX[T, V]{trie.tobase().blockSizeCachingAllNodeIterator()}
}

// ContainingFirstIterator returns an iterator that does a pre-order binary tree traversal of the added nodes.
// All added nodes will be visited before their added sub-nodes.
// For an address trie this means added containing subnet blocks will be visited before their added contained addresses and subnet blocks.
//
// Once a given node is visited, the iterator allows you to cache an object corresponding to the
// lower or upper sub-node that can be retrieved when you later visit that sub-node.
//
// Objects are cached only with nodes to be visited.
// So for this iterator that means an object will be cached with the first added lower or upper sub-node,
// the next lower or upper sub-node to be visited,
// which is not necessarily the direct lower or upper sub-node of a given node.
//
// The caching allows you to provide iteration context from a parent to its sub-nodes when iterating.
// The caching and retrieval is done in constant-time.
func (trie *AssociativeTrie[T, V]) ContainingFirstIterator(forwardSubNodeOrder bool) CachingTrieIterator[*AssociativeTrieNode[T, V]] {
	return cachingAssociativeAddressTrieNodeIteratorX[T, V]{trie.tobase().containingFirstIterator(forwardSubNodeOrder)}
}

// ContainingFirstAllNodeIterator returns an iterator that does a pre-order binary tree traversal.
// All nodes will be visited before their sub-nodes.
// For an address trie this means containing subnet blocks will be visited before their contained addresses and subnet blocks.
//
// Once a given node is visited, the iterator allows you to cache an object corresponding to the
// lower or upper sub-node that can be retrieved when you later visit that sub-node.
// That allows you to provide iteration context from a parent to its sub-nodes when iterating.
// The caching and retrieval is done in constant-time.
func (trie *AssociativeTrie[T, V]) ContainingFirstAllNodeIterator(forwardSubNodeOrder bool) CachingTrieIterator[*AssociativeTrieNode[T, V]] {
	return cachingAssociativeAddressTrieNodeIteratorX[T, V]{trie.tobase().containingFirstAllNodeIterator(forwardSubNodeOrder)}
}

// ContainedFirstIterator returns an iterator that does a post-order binary tree traversal of the added nodes.
// All added sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (trie *AssociativeTrie[T, V]) ContainedFirstIterator(forwardSubNodeOrder bool) IteratorWithRemove[*AssociativeTrieNode[T, V]] {
	return associativeAddressTrieNodeIteratorRem[T, V]{trie.tobase().containedFirstIterator(forwardSubNodeOrder)}
}

// ContainedFirstAllNodeIterator returns an iterator that does a post-order binary tree traversal.
// All sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (trie *AssociativeTrie[T, V]) ContainedFirstAllNodeIterator(forwardSubNodeOrder bool) Iterator[*AssociativeTrieNode[T, V]] {
	return associativeAddressTrieNodeIterator[T, V]{trie.tobase().containedFirstAllNodeIterator(forwardSubNodeOrder)}
}

// FirstNode returns the first (lowest-valued) node in the trie or nil if the trie is empty
func (trie *AssociativeTrie[T, V]) FirstNode() *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](trie.trieBase.trie.FirstNode())
}

// FirstAddedNode returns the first (lowest-valued) added node in the trie
// or nil if there are no added entries in this tree
func (trie *AssociativeTrie[T, V]) FirstAddedNode() *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](trie.trieBase.trie.FirstAddedNode())
}

// LastNode returns the last (highest-valued) node in the trie or nil if the trie is empty
func (trie *AssociativeTrie[T, V]) LastNode() *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](trie.trieBase.trie.LastNode())
}

// LastAddedNode returns the last (highest-valued) added node in the sub-trie originating from this node,
// or nil if there are no added entries in this tree
func (trie *AssociativeTrie[T, V]) LastAddedNode() *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](trie.trieBase.trie.LastAddedNode())
}

// LowerAddedNode returns the added node whose address is the highest address strictly less than the given address,
// or nil if there are no added entries in this tree
func (trie *AssociativeTrie[T, V]) LowerAddedNode(addr T) *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](trie.lowerAddedNode(addr))
}

// FloorAddedNode returns the added node whose address is the highest address less than or equal to the given address,
// or nil if there are no added entries in this tree
func (trie *AssociativeTrie[T, V]) FloorAddedNode(addr T) *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](trie.floorAddedNode(addr))
}

// HigherAddedNode returns the added node whose address is the lowest address strictly greater than the given address,
// or nil if there are no added entries in this tree
func (trie *AssociativeTrie[T, V]) HigherAddedNode(addr T) *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](trie.higherAddedNode(addr))
}

// CeilingAddedNode returns the added node whose address is the lowest address greater than or equal to the given address,
// or nil if there are no added entries in this tree
func (trie *AssociativeTrie[T, V]) CeilingAddedNode(addr T) *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](trie.ceilingAddedNode(addr))
}

// Clone clones this trie
func (trie *AssociativeTrie[T, V]) Clone() *AssociativeTrie[T, V] {
	return toAssociativeTrie[T, V](trie.tobase().clone())
}

// Equal returns whether the given argument is a trie with a set of nodes with the same keys as in this trie
func (trie *AssociativeTrie[T, V]) Equal(other *AssociativeTrie[T, V]) bool {
	return trie.toTrie().Equal(other.toTrie())
}

// DeepEqual returns whether the given argument is a trie with a set of nodes with the same keys and values as in this trie,
// the values being compared with reflect.DeepEqual
func (trie *AssociativeTrie[T, V]) DeepEqual(other *AssociativeTrie[T, V]) bool {
	return trie.toTrie().DeepEqual(other.toTrie())
}

// Put associates the specified value with the specified key in this map.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// If this map previously contained a mapping for a key,
// the old value is replaced by the specified value, and false is returned along with the old value.
// If this map did not previously contain a mapping for the key, true is returned along with a nil value.
// The boolean return value allows you to distinguish whether the address was previously mapped to nil or not mapped at all.
func (trie *AssociativeTrie[T, V]) Put(addr T, value V) (bool, V) {
	addr = mustBeBlockOrAddress(addr)
	res, val := trie.trie.Put(trieKey[T]{addr}, value)
	if val == nil {
		var v V
		return res, v
	}
	return res, val.(V) // TODO this should go away when making bintree generic
	//return trie.put(addr, value)
}

// PutTrie adds nodes for the address keys and values in the trie with the root node as the passed in node.  To add only the keys, use AddTrie.
//
// For each added in the given node that does not exist in the trie, a copy of each node will be made,
// the copy including the associated value, and the copy will be inserted into the trie.
//
// The address type/version of the keys must match.
//
// When adding one trie to another, this method is more efficient than adding each node of the first trie individually.
// When using this method, searching for the location to add sub-nodes starts from the inserted parent node.
//
// Returns the node corresponding to the given sub-root node, whether it was already in the trie or not.
func (trie *AssociativeTrie[T, V]) PutTrie(added *AssociativeTrieNode[T, V]) *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](trie.trie.PutTrie(added.toBinTrieNode()))
	//return trie.putTrie(added.toBase())
}

// PutNode associates the specified value with the specified key in this map.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the node for the added address, whether it was already in the tree or not.
//
// If you wish to know whether the node was already there when adding, use PutNew, or before adding you can use GetAddedNode.
func (trie *AssociativeTrie[T, V]) PutNode(addr T, value V) *AssociativeTrieNode[T, V] {
	addr = mustBeBlockOrAddress(addr)
	return toAssociativeTrieNode[T, V](trie.trie.PutNode(trieKey[T]{addr}, value))
	//return trie.putNode(addr, value)
}

// Remap remaps node values in the trie.
//
// This will look up the node corresponding to the given key.
// It will call the remapping function, regardless of whether the node is found or not.
//
// If the node is not found, the value argument will be the zero value.
// If the node is found, the value argument will be the node's value, which can also be the zero value.
// The boolean argument will be true if the node was found.
// If the node was not found, or if the value type V is an interface type and the node is mapped to a nil interface, then the boolean argument will be false.
//
// If the remapping function returns false as the second argument, then the matched node will be removed, if any.
// If it returns true, then it will either set the existing node to have the value given as the first argument,
// or if there was no matched node, it will create a new node with that value.
//
// The method will return the node involved, which is either the matched node, or the newly created node,
// or nil if there was no matched node nor newly created node.
//
// If the remapping function modifies the trie during its computation,
// and the returned value specifies changes to be made,
// then the trie will not be changed and Remap will panic.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (trie *AssociativeTrie[T, V]) Remap(addr T, remapper func(exsting V, found bool) (mapped V, mapIt bool)) *AssociativeTrieNode[T, V] {
	addr = mustBeBlockOrAddress(addr)
	//return toAssociativeTrieNode[T, V](trie.trieBase.trie.Remap(trieKey[T]{addr}, remapper)) // TODO reinstate once remapper is all we need, which is when bintree goes generic
	return toAssociativeTrieNode[T, V](trie.trieBase.trie.Remap(trieKey[T]{addr}, func(t tree.V) tree.V {
		var res bool
		var val V
		if t == nil {
			val, res = remapper(val, false)
		} else {
			val, res = remapper(t.(V), true)
		}
		if !res {
			return nil
		}
		return val
	}))
}

// RemapIfAbsent remaps node values in the trie, but only for nodes that do not exist, or nodes mapped to a nil interface (when V is an interface type).
//
// This will look up the node corresponding to the given key.
// If the node is not found, this will call the supplier function.
// If the node is found, type V is an interface type, and the node is mapped to nil, this will call the supplier function.
// If the node is found and mapped to anything other than a nil interface, then RemapIfAbsent will not call the supplier function.
//
// If the supplier function returns true as the second returned value, then the first returned value will be
// used for the node value, unless the type V is an interface type and the value is nil.
//
// When the returned value is used for the node value, it will either set the existing node to have that value,
// or if there was no matched node, it will create a new node with that value.
//
// The method will return the node involved, which is either the matched node, or the newly created node,
// or nil if there was no matched node nor newly created node.
//
// If the supplier function modifies the trie during its computation,
// and then returned value inidcates further changes to be made to the trie,
// then the trie will not be changed and RemapIfAbsent will panic.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (trie *AssociativeTrie[T, V]) RemapIfAbsent(addr T, supplier func() (V, bool)) *AssociativeTrieNode[T, V] {
	addr = mustBeBlockOrAddress(addr)
	//return toAssociativeTrieNode[T, V](trie.trieBase.trie.RemapIfAbsent(trieKey[T]{addr}, supplier)) // TODO reinstate once bintree uses generics too,  remapper is all we need
	return toAssociativeTrieNode[T, V](trie.trieBase.trie.RemapIfAbsent(trieKey[T]{addr}, func() tree.V {
		val, res := supplier()
		if !res {
			return nil
		}
		return val
	}, false))
}

// Get gets the value for the specified key in this mapped trie or sub-trie.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the value for the given key.
// Returns nil if the contains no mapping for that key or if the mapped value is nil.
func (trie *AssociativeTrie[T, V]) Get(addr T) (bool, V) {
	addr = mustBeBlockOrAddress(addr)
	val := trie.trie.Get(trieKey[T]{addr})
	if val == nil {
		var v V
		return false, v
	}
	return true, val.(V) // TODO this should go away when making bintree generic
}

// For some reason Format must be here and not in addressTrieNode for nil node.
// It panics in fmt code either way, but if in here then it is handled by a recover() call in fmt properly in the debugger.

// Format implements the fmt.Formatter interface
func (trie AssociativeTrie[T, V]) Format(state fmt.State, verb rune) {
	trie.trieBase.trie.Format(state, verb)
}

func AssociativeTreesString[T TrieKeyConstraint[T], V any](withNonAddedKeys bool, tries ...*AssociativeTrie[T, V]) string {
	binTries := make([]*tree.BinTrie, 0, len(tries))
	for _, trie := range tries {
		binTries = append(binTries, toAssocBinTrie[T, V](trie))
	}
	return tree.TreesString(withNonAddedKeys, binTries...)
}

func toAssocBinTrie[T TrieKeyConstraint[T], V any](trie *AssociativeTrie[T, V]) *tree.BinTrie {
	return (*tree.BinTrie)(unsafe.Pointer(trie))
}

func toAssociativeTrie[T TrieKeyConstraint[T], V any](trie *tree.BinTrie) *AssociativeTrie[T, V] {
	return (*AssociativeTrie[T, V])(unsafe.Pointer(trie))
}

// Ensures the address is either an individual address or a prefix block subnet.
func mustBeBlockOrAddress[T TrieKeyConstraint[T]](addr T) T {
	res, err := checkBlockOrAddress(addr)
	if err != nil {
		panic(err)
	}
	return res
}

// Ensures the address is either an individual address or a prefix block subnet.
// Returns a normalized address which has no prefix length if it is a single address,
// or has a prefix length matching the prefix block size if it is a prefix block.
func checkBlockOrAddress[T TrieKeyConstraint[T]](addr T) (res T, err addrerr.IncompatibleAddressError) {
	return addr.toSinglePrefixBlockOrAddress()
}

// NewTrie constructs an address trie for the given type, without a root.
func NewTrie[T TrieKeyConstraint[T]]() *Trie[T] {
	return &Trie[T]{}
}

// NewAssociativeTrie constructs an associative address trie for the given types, without a root.
func NewAssociativeTrie[T TrieKeyConstraint[T], V any]() *AssociativeTrie[T, V] {
	return &AssociativeTrie[T, V]{}
}

type (
	AssociativeAddressTrie     = AssociativeTrie[*Address, any]
	AddressTrie                = Trie[*Address]
	IPv4AddressAssociativeTrie = AssociativeTrie[*IPv4Address, any]
	IPv4AddressTrie            = Trie[*IPv4Address]
	IPv6AddressAssociativeTrie = AssociativeTrie[*IPv6Address, any]
	IPv6AddressTrie            = Trie[*IPv6Address]

	// NodeValue is an alias for the generic node value type for AssociativeAddressTrie, IPv4AddressAssociativeTrie, and IPv6AddressAssociativeTrie
	NodeValue = any
)

// NewIPv4AddressTrie constructs an IPv4 address trie with the root as the 0.0.0.0/0 prefix block
// This is here for backwards compatibility.  Using NewTrie is recommended instead.
func NewIPv4AddressTrie() *Trie[*IPv4Address] { // for backwards compatibility
	return &Trie[*IPv4Address]{trieBase[*IPv4Address, emptyValue]{tree.NewBinTrie(trieKey[*IPv4Address]{ipv4All})}}
}

// NewIPv4AddressAssociativeTrie constructs an IPv4 associative address trie with the root as the 0.0.0.0/0 prefix block
// This is here for backwards compatibility.  Using NewAssociativeTrie is recommended instead.
func NewIPv4AddressAssociativeTrie() *AssociativeTrie[*IPv4Address, any] { // for backwards compatibility
	return &AssociativeTrie[*IPv4Address, any]{trieBase[*IPv4Address, any]{tree.NewBinTrie(trieKey[*IPv4Address]{ipv4All})}}
}

// NewIPv6AddressTrie constructs an IPv6 address trie with the root as the ::/0 prefix block
// This is here for backwards compatibility.  Using NewTrie is recommended instead.
func NewIPv6AddressTrie() *Trie[*IPv6Address] { // for backwards compatibility
	return &Trie[*IPv6Address]{trieBase[*IPv6Address, emptyValue]{tree.NewBinTrie(trieKey[*IPv6Address]{ipv6All})}}
}

// NewIPv6AddressAssociativeTrie constructs an IPv6 associative address trie with the root as the ::/0 prefix block
// This is here for backwards compatibility.  Using NewAssociativeTrie is recommended instead.
func NewIPv6AddressAssociativeTrie() *AssociativeTrie[*IPv6Address, any] { // for backwards compatibility
	return &AssociativeTrie[*IPv6Address, any]{trieBase[*IPv6Address, any]{tree.NewBinTrie(trieKey[*IPv6Address]{ipv6All})}}
}

// NewMACAddressTrie constructs a MAC address trie with the root as the zero-prefix block
// If extended is true, the trie will consist of 64-bit EUI addresses, otherwise the addresses will be 48-bit.
// If you wish to construct a trie in which the address size is determined by the first added address,
// use the zero-value MACAddressTrie{}
// This is here for backwards compatibility.  Using NewTrie is recommended instead.
func NewMACAddressTrie(extended bool) *Trie[*MACAddress] {
	var rootAddr *MACAddress
	if extended {
		rootAddr = macAllExtended
	} else {
		rootAddr = macAll
	}
	return &Trie[*MACAddress]{trieBase[*MACAddress, emptyValue]{tree.NewBinTrie(trieKey[*MACAddress]{rootAddr})}}
}

// NewMACAddressAssociativeTrie constructs a MAC associative address trie with the root as the zero-prefix prefix block
// This is here for backwards compatibility.  Using NewAssociativeTrie is recommended instead.
func NewMACAddressAssociativeTrie(extended bool) *AssociativeTrie[*MACAddress, any] {
	var rootAddr *MACAddress
	if extended {
		rootAddr = macAllExtended
	} else {
		rootAddr = macAll
	}
	return &AssociativeTrie[*MACAddress, any]{trieBase[*MACAddress, any]{tree.NewBinTrie(trieKey[*MACAddress]{rootAddr})}}
}
