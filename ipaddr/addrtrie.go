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

package ipaddr

import (
	"fmt"
	"github.com/seancfoley/bintree/tree"
	"github.com/seancfoley/ipaddress-go/ipaddr/addrerr"
	"unsafe"
)

//TODO LATER with generics I will do:
//
//type AddressTrieGeneric[K] struct {
//
//}
//
//type AddressTrie struct {
//	AddressTrieGeneric[interface{}]
//}
// But I cannot share the same name AddressTrie for AddressTrieGeneric, so what naming scheme can I use?
// AddressTrieBase[K] seems like a good name.
// AssociativeAddressTrieBase[K,V] seems like a good name.
// IPv4AssociativeAddressTrie can use IPv4AssociativeTrie[V] ? hmmmmmmm
// IPv4AddressTrie can stay the same.
// Or, you can shorten Address to Addr to get new name

// TODO LATER PLAN FOR GENERICS?  You can either shorten "Address" to "Addr" or drop "Address" entirely
// the name of the key type (which you can extend of TrieKey can help identify as "Address" instead
//
//AddressTrieNode - TrieNode in java
//AssociativeAddressTrieNode - AssociativeTrieNode in java
//IPv4AddressTrieNode
//IPv4AddressAssociativeTrieNode
//
//IPv4AddressAssociativeTrie
//AssociativeAddressTrie
//IPv4AddressTrie
//AddressTrie

type addressTrie struct {
	trie tree.BinTrie
}

// Clear removes all added nodes from the tree, after which IsEmpty() will return true
func (trie *addressTrie) Clear() {
	trie.trie.Clear()
}

// GetRoot returns the root node of this trie, which can be nil for an implicitly zero-valued uninitialized trie, but not for any other trie
func (trie *addressTrie) getRoot() *AddressTrieNode {
	return toAddressTrieNode(trie.trie.GetRoot())
}

func (trie *addressTrie) add(addr *Address) bool {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.Add(&addressTrieKey{addr})
}

func (trie *addressTrie) addNode(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(trie.trie.AddNode(&addressTrieKey{addr}))
}

// ConstructAddedNodesTree provides an associative trie in which the root and each added node are mapped to a list of their respective direct added sub-nodes.
// This trie provides an alternative non-binary tree structure of the added nodes.
// It is used by ToAddedNodesTreeString to produce a string showing the alternative structure.
// If there are no non-added nodes in this trie, then the alternative tree structure provided by this method is the same as the original trie.
func (trie *addressTrie) constructAddedNodesTree() addressTrie {
	return addressTrie{trie.trie.ConstructAddedNodesTree()}
}

func (trie *addressTrie) addTrie(added *addressTrieNode) *AddressTrieNode {
	return toAddressTrieNode(trie.trie.AddTrie(added.toTrieNode()))
}

func (trie *addressTrie) contains(addr *Address) bool {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.Contains(&addressTrieKey{addr})
}

func (trie *addressTrie) remove(addr *Address) bool {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.Remove(&addressTrieKey{addr})
}

func (trie *addressTrie) removeElementsContainedBy(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(trie.trie.RemoveElementsContainedBy(&addressTrieKey{addr}))
}

func (trie *addressTrie) elementsContainedBy(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(trie.trie.ElementsContainedBy(&addressTrieKey{addr}))
}

func (trie *addressTrie) elementsContaining(addr *Address) *ContainmentPath {
	addr = mustBeBlockOrAddress(addr)
	return toContainmentPath(trie.trie.ElementsContaining(&addressTrieKey{addr}))
}

func (trie *addressTrie) longestPrefixMatch(addr *Address) *Address {
	addr = mustBeBlockOrAddress(addr)
	res := trie.trie.LongestPrefixMatch(&addressTrieKey{addr})
	if res == nil {
		return nil
	}
	return res.(*addressTrieKey).Address
}

// only added nodes are added to the linked list
func (trie *addressTrie) longestPrefixMatchNode(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(trie.trie.LongestPrefixMatchNode(&addressTrieKey{addr}))
}

func (trie *addressTrie) elementContains(addr *Address) bool {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.ElementContains(&addressTrieKey{addr})
}

func (trie *addressTrie) getNode(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(trie.trie.GetNode(&addressTrieKey{addr}))
}

func (trie *addressTrie) getAddedNode(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(trie.trie.GetAddedNode(&addressTrieKey{addr}))
}

func (trie *addressTrie) iterator() AddressIterator {
	if trie == nil {
		return nilAddrIterator()
	}
	return addressKeyIterator{trie.trie.Iterator()}
}

func (trie *addressTrie) descendingIterator() AddressIterator {
	if trie == nil {
		return nilAddrIterator()
	}
	return addressKeyIterator{trie.trie.DescendingIterator()}
}

func (trie *addressTrie) nodeIterator(forward bool) AddressTrieNodeIteratorRem {
	return addrTrieNodeIteratorRem{trie.toTrie().NodeIterator(forward)}
}

func (trie *addressTrie) allNodeIterator(forward bool) AddressTrieNodeIteratorRem {
	return addrTrieNodeIteratorRem{trie.toTrie().AllNodeIterator(forward)}
}

// Iterates the added nodes in the trie, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *addressTrie) blockSizeNodeIterator(lowerSubNodeFirst bool) AddressTrieNodeIteratorRem {
	return addrTrieNodeIteratorRem{trie.toTrie().BlockSizeNodeIterator(lowerSubNodeFirst)}
}

// Iterates all nodes in the trie, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *addressTrie) blockSizeAllNodeIterator(lowerSubNodeFirst bool) AddressTrieNodeIteratorRem {
	return addrTrieNodeIteratorRem{trie.toTrie().BlockSizeAllNodeIterator(lowerSubNodeFirst)}
}

// Iterates all nodes, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
func (trie *addressTrie) blockSizeCachingAllNodeIterator() CachingAddressTrieNodeIterator {
	return cachingAddressTrieNodeIterator{trie.toTrie().BlockSizeCachingAllNodeIterator()}
}

// Iterates all nodes, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
func (trie *addressTrie) containingFirstIterator(forwardSubNodeOrder bool) CachingAddressTrieNodeIterator {
	return cachingAddressTrieNodeIterator{trie.toTrie().ContainingFirstIterator(forwardSubNodeOrder)}
}

func (trie *addressTrie) containingFirstAllNodeIterator(forwardSubNodeOrder bool) CachingAddressTrieNodeIterator {
	return cachingAddressTrieNodeIterator{trie.toTrie().ContainingFirstAllNodeIterator(forwardSubNodeOrder)}
}

func (trie *addressTrie) containedFirstIterator(forwardSubNodeOrder bool) AddressTrieNodeIteratorRem {
	return addrTrieNodeIteratorRem{trie.toTrie().ContainedFirstIterator(forwardSubNodeOrder)}
}

func (trie *addressTrie) containedFirstAllNodeIterator(forwardSubNodeOrder bool) AddressTrieNodeIterator {
	return addrTrieNodeIterator{trie.toTrie().ContainedFirstAllNodeIterator(forwardSubNodeOrder)}
}

func (trie *addressTrie) lowerAddedNode(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(trie.trie.LowerAddedNode(&addressTrieKey{addr}))
}

func (trie *addressTrie) floorAddedNode(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(trie.trie.FloorAddedNode(&addressTrieKey{addr}))
}

func (trie *addressTrie) higherAddedNode(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(trie.trie.HigherAddedNode(&addressTrieKey{addr}))
}

func (trie *addressTrie) ceilingAddedNode(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(trie.trie.CeilingAddedNode(&addressTrieKey{addr}))
}

func (trie *addressTrie) clone() *AddressTrie {
	return toAddressTrie(trie.toTrie().Clone())
}

func toAddressTrie(trie *tree.BinTrie) *AddressTrie {
	return (*AddressTrie)(unsafe.Pointer(trie))
}

func (trie *addressTrie) toTrie() *tree.BinTrie {
	return (*tree.BinTrie)(unsafe.Pointer(trie))
}

//
// AddressTrie is a compact binary trie (aka compact binary prefix tree, or binary radix trie), for addresses and/or CIDR prefix block subnets.
// The prefixes in used by the prefix trie are the CIDR prefixes, or the full address in the case of individual addresses with no prefix length.
// The elements of the trie are CIDR prefix blocks or addresses.
//
// The zero-value of an AddressTrie is a trie ready for use.  Its root will be nil until an element is added to it.
// Any trie without a root can be converted to a trie of any address type of version.
// However, once any subnet or address is added to the trie, it will have an assigned root, and any further addition to the trie must match the type and version of the root,
// and the trie can no longer be converted a trie for any other address version or type.  Once there is a root, the root cannot be removed.
//
// Any trie created by a creation function will start with an assigned root.
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
type AddressTrie struct {
	addressTrie
}

func (trie *AddressTrie) toTrie() *tree.BinTrie {
	return (*tree.BinTrie)(unsafe.Pointer(trie))
}

func (trie *AddressTrie) tobase() *addressTrie {
	return (*addressTrie)(unsafe.Pointer(trie))
}

// ToIPv4 converts this trie to an IPv4 trie.  If this trie has no root, or the trie has an IPv4 root, the trie can be converted, otherwise, this method returns nil.
// The underlying trie does not change.  The IPv4 type simply provides type safety, because you cannot mix different address versions or types in the same trie.
// Mixing versions and or types will cause a panic.
func (trie *AddressTrie) ToIPv4() *IPv4AddressTrie {
	if trie != nil {
		if root := trie.GetRoot(); root == nil || root.GetKey().IsIPv4() {
			return &IPv4AddressTrie{trie.addressTrie}
		}
	}
	return nil
}

// ToIPv6 converts this trie to an IPv6 trie.  If this trie has no root, or the trie has an IPv6 root, the trie can be converted, otherwise, this method returns nil.
// The underlying trie does not change.  The IPv6 type simply provides type safety, because you cannot mix different address versions or types in the same trie.
// Mixing versions and or types will cause a panic.
func (trie *AddressTrie) ToIPv6() *IPv6AddressTrie {
	if trie != nil {
		if root := trie.GetRoot(); root == nil || root.GetKey().IsIPv6() {
			return &IPv6AddressTrie{trie.addressTrie}
		}
	}
	return nil
}

// ToMAC converts this trie to a MAC trie.  If this trie has no root, or the trie has a MAC root, the trie can be converted, otherwise, this method returns nil.
// The underlying trie does not change.  The MAC type simply provides type safety, because you cannot mix different address versions or types in the same trie.
// Mixing versions and or types will cause a panic.
func (trie *AddressTrie) ToMAC() *MACAddressTrie {
	if trie != nil {
		if root := trie.GetRoot(); root == nil || root.GetKey().IsMAC() {
			return &MACAddressTrie{trie.addressTrie}
		}
	}
	return nil
}

// ToAssociative converts this trie to an associative trie.  The underlying trie does not change.
// Associative tries provide additional API to associate each node with a mapped value.
func (trie *AddressTrie) ToAssociative() *AssociativeAddressTrie {
	return (*AssociativeAddressTrie)(unsafe.Pointer(trie))
}

// ToIPv4Associative converts this trie to an IPv4 associative trie.  If this trie has no root, or the trie has an IPv4 root, the trie can be converted, otherwise, this method returns nil.
// The underlying trie does not change.  The IPv4 type simply provides type safety, because you cannot mix different address versions or types in the same trie.
// Mixing versions and or types will cause a panic.
func (trie *AddressTrie) ToIPv4Associative() *IPv4AddressAssociativeTrie {
	if trie != nil {
		if root := trie.GetRoot(); root == nil || root.GetKey().IsIPv4() {
			return (*IPv4AddressAssociativeTrie)(unsafe.Pointer(trie))
		}
	}
	return nil
}

// ToIPv6Associative converts this trie to an IPv6 associative trie.  If this trie has no root, or the trie has an IPv6 root, the trie can be converted, otherwise, this method returns nil.
// The underlying trie does not change.  The IPv6 type simply provides type safety, because you cannot mix different address versions or types in the same trie.
// Mixing versions and or types will cause a panic.
func (trie *AddressTrie) ToIPv6Associative() *IPv6AddressAssociativeTrie {
	if trie != nil {
		if root := trie.GetRoot(); root == nil || root.GetKey().IsIPv6() {
			return (*IPv6AddressAssociativeTrie)(unsafe.Pointer(trie))
		}
	}
	return nil
}

// ToMACAssociative converts this trie to a MAC associative trie.  If this trie has no root, or the trie has a MAC root, the trie can be converted, otherwise, this method returns nil.
// The underlying trie does not change.  The MAC type simply provides type safety, because you cannot mix different address versions or types in the same trie.
// Mixing versions and or types will cause a panic.
func (trie *AddressTrie) ToMACAssociative() *MACAddressAssociativeTrie {
	if trie != nil {
		if root := trie.GetRoot(); root == nil || root.GetKey().IsMAC() {
			return (*MACAddressAssociativeTrie)(unsafe.Pointer(trie))
		}
	}
	return nil
}

// GetRoot returns the root node of this trie, which can be nil for an implicitly zero-valued uninitialized trie, but not for any other trie
func (trie *AddressTrie) GetRoot() *AddressTrieNode {
	return trie.getRoot()
}

// Size returns the number of elements in the tree.
// It does not return the number of nodes.
// Only nodes for which IsAdded() returns true are counted (those nodes corresponding to added addresses and prefix blocks).
// When zero is returned, IsEmpty() returns true.
func (trie *AddressTrie) Size() int {
	return trie.toTrie().Size()
}

// NodeSize returns the number of nodes in the tree, which is always more than the number of elements.
func (trie *AddressTrie) NodeSize() int {
	return trie.toTrie().NodeSize()
}

// IsEmpty returns true if there are not any added nodes within this tree
func (trie *AddressTrie) IsEmpty() bool {
	return trie.Size() == 0
}

// TreeString returns a visual representation of the tree with one node per line, with or without the non-added keys.
func (trie *AddressTrie) TreeString(withNonAddedKeys bool) string {
	return trie.toTrie().TreeString(withNonAddedKeys)
}

// String returns a visual representation of the tree with one node per line.
func (trie *AddressTrie) String() string {
	return trie.toTrie().String()
}

// AddedNodesTreeString provides a flattened version of the trie showing only the contained added nodes and their containment structure, which is non-binary.
// The root node is included, which may or may not be added.
func (trie *AddressTrie) AddedNodesTreeString() string {
	return trie.toTrie().AddedNodesTreeString()
}

// Add adds the address to this trie.
// The address must match the same type and version of any existing addresses already in the trie.
// Returns true if the address did not already exist in the trie.
func (trie *AddressTrie) Add(addr *Address) bool {
	return trie.add(addr)
}

// AddNode adds the address to this trie.
// The address must match the same type and version of any existing addresses already in the trie.
// The new or existing node for the address is returned.
func (trie *AddressTrie) AddNode(addr *Address) *AddressTrieNode {
	return trie.addNode(addr)
}

// AddTrie adds nodes for the keys in the trie with the root node as the passed in node.  To add both keys and values, use PutTrie.
func (trie *AddressTrie) AddTrie(added *AddressTrieNode) *AddressTrieNode {
	return trie.addTrie(added.tobase())
}

// ConstructAddedNodesTree provides an associative trie in which the root and each added node are mapped to a list of their respective direct added sub-nodes.
// This trie provides an alternative non-binary tree structure of the added nodes.
// It is used by ToAddedNodesTreeString to produce a string showing the alternative structure.
// If there are no non-added nodes in this trie, then the alternative tree structure provided by this method is the same as the original trie.
func (trie *AddressTrie) ConstructAddedNodesTree() *AddressTrie {
	return &AddressTrie{trie.constructAddedNodesTree()}
}

// Contains returns whether the given address or prefix block subnet is in the trie as an added element.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the prefix block or address exists already in the trie, false otherwise.
//
// Use GetAddedNode to get the node for the address rather than just checking for its existence.
func (trie *AddressTrie) Contains(addr *Address) bool {
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
func (trie *AddressTrie) Remove(addr *Address) bool {
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
func (trie *AddressTrie) RemoveElementsContainedBy(addr *Address) *AddressTrieNode {
	return trie.removeElementsContainedBy(addr)
}

// ElementsContainedBy checks if a part of this trie is contained by the given prefix block subnet or individual address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the root node of the contained sub-trie, or nil if no sub-trie is contained.
// The node returned need not be an "added" node, see IsAdded for more details on added nodes.
// The returned sub-trie is backed by this trie, so changes in this trie are reflected in those nodes and vice-versa.
func (trie *AddressTrie) ElementsContainedBy(addr *Address) *AddressTrieNode {
	return trie.elementsContainedBy(addr)
}

// ElementsContaining finds the trie nodes in the trie containing the given key and returns them as a linked list.
// Only added nodes are added to the linked list
//
// If the argument is not a single address nor prefix block, this method will panic.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (trie *AddressTrie) ElementsContaining(addr *Address) *ContainmentPath {
	return trie.elementsContaining(addr)
}

// LongestPrefixMatch returns the address added to the trie with the longest matching prefix compared to the provided address, or nil if no matching address
func (trie *AddressTrie) LongestPrefixMatch(addr *Address) *Address {
	return trie.longestPrefixMatch(addr)
}

// only added nodes are added to the linked list

// LongestPrefixMatchNode returns the node of address added to the trie with the longest matching prefix compared to the provided address, or nil if no matching address
func (trie *AddressTrie) LongestPrefixMatchNode(addr *Address) *AddressTrieNode {
	return trie.longestPrefixMatchNode(addr)
}

// ElementContains checks if a prefix block subnet or address in the trie contains the given subnet or address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the subnet or address is contained by a trie element, false otherwise.
//
// To get all the containing addresses, use ElementsContaining
func (trie *AddressTrie) ElementContains(addr *Address) bool {
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
func (trie *AddressTrie) GetNode(addr *Address) *AddressTrieNode {
	return trie.getNode(addr)
}

// GetAddedNode gets trie nodes representing added elements.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Use Contains to check for the existence of a given address in the trie,
// as well as GetNode to search for all nodes including those not-added but also auto-generated nodes for subnet blocks.
func (trie *AddressTrie) GetAddedNode(addr *Address) *AddressTrieNode {
	return trie.getAddedNode(addr)
}

// Iterator returns an iterator that iterates through the added addresses and prefix blocks in the trie.
// The iteration is in sorted element order.
func (trie *AddressTrie) Iterator() AddressIterator {
	return trie.tobase().iterator()
}

// DescendingIterator returns an iterator that iterates through the added addresses and prefix blocks in the trie.
// The iteration is in reverse sorted element order.
func (trie *AddressTrie) DescendingIterator() AddressIterator {
	return trie.tobase().descendingIterator()
}

// NodeIterator returns an iterator that iterates through all the added nodes in the trie in forward or reverse tree order.
func (trie *AddressTrie) NodeIterator(forward bool) AddressTrieNodeIteratorRem {
	return trie.tobase().nodeIterator(forward)
}

// AllNodeIterator returns an iterator that iterates through all the nodes in the trie in forward or reverse tree order.
func (trie *AddressTrie) AllNodeIterator(forward bool) AddressTrieNodeIteratorRem {
	return trie.tobase().allNodeIterator(forward)
}

// BlockSizeNodeIterator returns an iterator that iterates the added nodes in the trie, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *AddressTrie) BlockSizeNodeIterator(lowerSubNodeFirst bool) AddressTrieNodeIteratorRem {
	return trie.tobase().blockSizeNodeIterator(lowerSubNodeFirst)
}

// BlockSizeAllNodeIterator returns an iterator that iterates all nodes in the trie, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *AddressTrie) BlockSizeAllNodeIterator(lowerSubNodeFirst bool) AddressTrieNodeIteratorRem {
	return trie.tobase().blockSizeAllNodeIterator(lowerSubNodeFirst)
}

// BlockSizeCachingAllNodeIterator returns an iterator that iterates all nodes, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
func (trie *AddressTrie) BlockSizeCachingAllNodeIterator() CachingAddressTrieNodeIterator {
	return trie.tobase().blockSizeCachingAllNodeIterator()
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
// The caching and retrieval is done in constant-time and linear space (proportional to tree size).
func (trie *AddressTrie) ContainingFirstIterator(forwardSubNodeOrder bool) CachingAddressTrieNodeIterator {
	return trie.tobase().containingFirstIterator(forwardSubNodeOrder)
}

// ContainingFirstAllNodeIterator returns an iterator that does a pre-order binary tree traversal.
// All nodes will be visited before their sub-nodes.
// For an address trie this means containing subnet blocks will be visited before their contained addresses and subnet blocks.
//
// Once a given node is visited, the iterator allows you to cache an object corresponding to the
// lower or upper sub-node that can be retrieved when you later visit that sub-node.
// That allows you to provide iteration context from a parent to its sub-nodes when iterating.
// The caching and retrieval is done in constant-time and linear space (proportional to tree size).
func (trie *AddressTrie) ContainingFirstAllNodeIterator(forwardSubNodeOrder bool) CachingAddressTrieNodeIterator {
	return trie.tobase().containingFirstAllNodeIterator(forwardSubNodeOrder)
}

// ContainedFirstIterator returns an iterator that does a post-order binary tree traversal of the added nodes.
// All added sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (trie *AddressTrie) ContainedFirstIterator(forwardSubNodeOrder bool) AddressTrieNodeIteratorRem {
	return trie.tobase().containedFirstIterator(forwardSubNodeOrder)
}

// ContainedFirstAllNodeIterator returns an iterator that does a post-order binary tree traversal.
// All sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (trie *AddressTrie) ContainedFirstAllNodeIterator(forwardSubNodeOrder bool) AddressTrieNodeIterator {
	return trie.tobase().containedFirstAllNodeIterator(forwardSubNodeOrder)
}

// FirstNode returns the first (lowest valued) node in the trie
func (trie *AddressTrie) FirstNode() *AddressTrieNode {
	return toAddressTrieNode(trie.trie.FirstNode())
}

// FirstAddedNode returns the first (lowest valued) added node in the trie,
// or nil if there are no added entries in this tree or sub-tree
func (trie *AddressTrie) FirstAddedNode() *AddressTrieNode {
	return toAddressTrieNode(trie.trie.FirstAddedNode())
}

// LastNode returns the last (highest valued) node in the trie
func (trie *AddressTrie) LastNode() *AddressTrieNode {
	return toAddressTrieNode(trie.trie.LastNode())
}

// LastAddedNode returns the last (highest valued) added node in the trie,
// or nil if there are no added entries in this tree or sub-tree
func (trie *AddressTrie) LastAddedNode() *AddressTrieNode {
	return toAddressTrieNode(trie.trie.LastAddedNode())
}

// LowerAddedNode returns the added node whose address is the highest address strictly less than the given address.
func (trie *AddressTrie) LowerAddedNode(addr *Address) *AddressTrieNode {
	return trie.lowerAddedNode(addr)
}

// FloorAddedNode returns the added node whose address is the highest address less than or equal to the given address.
func (trie *AddressTrie) FloorAddedNode(addr *Address) *AddressTrieNode {
	return trie.floorAddedNode(addr)
}

// HigherAddedNode returns the added node whose address is the lowest address strictly greater than the given address.
func (trie *AddressTrie) HigherAddedNode(addr *Address) *AddressTrieNode {
	return trie.higherAddedNode(addr)
}

// CeilingAddedNode returns the added node whose address is the lowest address greater than or equal to the given address.
func (trie *AddressTrie) CeilingAddedNode(addr *Address) *AddressTrieNode {
	return trie.ceilingAddedNode(addr)
}

// Clone clones this trie
func (trie *AddressTrie) Clone() *AddressTrie {
	return trie.tobase().clone()
}

// Equal returns whether the given argument is a trie with a set of nodes with the same keys as in this trie
func (trie *AddressTrie) Equal(other *AddressTrie) bool {
	return trie.toTrie().Equal(other.toTrie())
}

// For some reason Format must be here and not in addressTrieNode for nil node.
// It panics in fmt code either way, but if in here then it is handled by a recover() call in fmt properly.
// Seems to be a problem only in the debugger.

// Format implements the fmt.Formatter interface.
func (trie AddressTrie) Format(state fmt.State, verb rune) {
	// without this, prints like {{{{<nil>}}}} or {{{{0xc00014ca50}}}}
	// which is done by printValue in print.go of fmt
	trie.trie.Format(state, verb)
}

func TreesString(withNonAddedKeys bool, tries ...*AddressTrie) string {
	binTries := make([]*tree.BinTrie, 0, len(tries))
	for _, trie := range tries {
		binTries = append(binTries, toBinTrie(trie))
	}
	return tree.TreesString(withNonAddedKeys, binTries...)
}

func toBinTrie(trie *AddressTrie) *tree.BinTrie {
	return (*tree.BinTrie)(unsafe.Pointer(trie))
}

////////
////////
////////
////////
////////
////////
////////

type associativeAddressTrie struct {
	addressTrie
}

func (trie *associativeAddressTrie) put(addr *Address, value NodeValue) (bool, NodeValue) {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.Put(&addressTrieKey{addr}, value)
}

func (trie *associativeAddressTrie) putTrie(added *addressTrieNode) *AssociativeAddressTrieNode {
	return toAddressTrieNode(trie.trie.PutTrie(added.toTrieNode())).ToAssociative()
}

func (trie *associativeAddressTrie) putNode(addr *Address, value NodeValue) *AssociativeAddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(trie.trie.PutNode(&addressTrieKey{addr}, value)).ToAssociative()
}

func (trie *associativeAddressTrie) remap(addr *Address, remapper func(NodeValue) NodeValue) *AssociativeAddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(trie.trie.Remap(&addressTrieKey{addr}, remapper)).ToAssociative()
}

func (trie *associativeAddressTrie) remapIfAbsent(addr *Address, supplier func() NodeValue, insertNil bool) *AssociativeAddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(trie.trie.RemapIfAbsent(&addressTrieKey{addr}, supplier, insertNil)).ToAssociative()
}

func (trie *associativeAddressTrie) get(addr *Address) NodeValue {
	addr = mustBeBlockOrAddress(addr)
	return trie.trie.Get(&addressTrieKey{addr})
}

func (trie *associativeAddressTrie) deepEqual(other *associativeAddressTrie) bool {
	return trie.trie.DeepEqual(&other.trie)
}

////////
////////
////////
////////
////////
////////
////////
////////

type AssociativeAddressTrie struct {
	associativeAddressTrie
}

func (trie *AssociativeAddressTrie) toTrie() *tree.BinTrie {
	return (*tree.BinTrie)(unsafe.Pointer(trie))
}

func (trie *AssociativeAddressTrie) tobase() *addressTrie {
	return (*addressTrie)(unsafe.Pointer(trie))
}

func (trie *AssociativeAddressTrie) ToBase() *AddressTrie {
	return (*AddressTrie)(unsafe.Pointer(trie))
}

// ToIPv4 converts this trie to an IPv4 associative trie.  If this trie has no root, or the trie has an IPv4 root, the trie can be converted, otherwise, this method returns nil.
// The underlying trie does not change.  The IPv4 type simply provides type safety, because you cannot mix different address versions or types in the same trie.
// Mixing versions and or types will cause a panic.
// Also, associative tries provide additional API to associate each node with a mapped value.
func (trie *AssociativeAddressTrie) ToIPv4() *IPv4AddressAssociativeTrie {
	if trie != nil {
		if root := trie.GetRoot(); root == nil || root.GetKey().IsIPv4() {
			return (*IPv4AddressAssociativeTrie)(trie)
		}
	}
	return nil
}

// GetRoot returns the root node of this trie, which can be nil for an implicitly zero-valued uninitialized trie, but not for any other trie
func (trie *AssociativeAddressTrie) GetRoot() *AssociativeAddressTrieNode {
	return trie.getRoot().ToAssociative()
}

// Size returns the number of elements in the tree.
// It does not return the number of nodes.
// Only nodes for which IsAdded() returns true are counted (those nodes corresponding to added addresses and prefix blocks).
// When zero is returned, IsEmpty() returns true.
func (trie *AssociativeAddressTrie) Size() int {
	return trie.toTrie().Size()
}

// NodeSize returns the number of nodes in the tree, which is always more than the number of elements.
func (trie *AssociativeAddressTrie) NodeSize() int {
	return trie.toTrie().NodeSize()
}

// IsEmpty returns true if there are not any added nodes within this tree
func (trie *AssociativeAddressTrie) IsEmpty() bool {
	return trie.Size() == 0
}

// TreeString returns a visual representation of the tree with one node per line, with or without the non-added keys.
func (trie *AssociativeAddressTrie) TreeString(withNonAddedKeys bool) string {
	return trie.toTrie().TreeString(withNonAddedKeys)
}

// String returns a visual representation of the tree with one node per line.
func (trie *AssociativeAddressTrie) String() string {
	return trie.toTrie().String()
}

// AddedNodesTreeString provides a flattened version of the trie showing only the contained added nodes and their containment structure, which is non-binary.
// The root node is included, which may or may not be added.
func (trie *AssociativeAddressTrie) AddedNodesTreeString() string {
	return trie.toTrie().AddedNodesTreeString()
}

// Iterator returns an iterator that iterates through the added addresses and prefix blocks in the trie.
// The iteration is in sorted element order.
func (trie *AssociativeAddressTrie) Iterator() AddressIterator {
	return trie.tobase().iterator()
}

// DescendingIterator returns an iterator that iterates through the added addresses and prefix blocks in the trie.
// The iteration is in reverse sorted element order.
func (trie *AssociativeAddressTrie) DescendingIterator() AddressIterator {
	return trie.tobase().descendingIterator()
}

// Add adds the address to this trie.
// Returns true if the address did not already exist in the trie.
func (trie *AssociativeAddressTrie) Add(addr *Address) bool {
	return trie.add(addr)
}

// AddNode adds the address key to this trie.
// The new or existing node for the address is returned.
func (trie *AssociativeAddressTrie) AddNode(addr *Address) *AssociativeAddressTrieNode {
	return trie.addNode(addr).ToAssociative()
}

// AddTrie adds nodes for the keys in the trie with the root node as the passed in node.  To add both keys and values, use PutTrie.
func (trie *AssociativeAddressTrie) AddTrie(added *AssociativeAddressTrieNode) *AssociativeAddressTrieNode {
	return trie.addTrie(added.toBase()).ToAssociative()
}

// ConstructAddedNodesTree provides an associative trie in which the root and each added node are mapped to a list of their respective direct added sub-nodes.
// This trie provides an alternative non-binary tree structure of the added nodes.
// It is used by ToAddedNodesTreeString to produce a string showing the alternative structure.
// If there are no non-added nodes in this trie, then the alternative tree structure provided by this method is the same as the original trie.
func (trie *AssociativeAddressTrie) ConstructAddedNodesTree() *AssociativeAddressTrie {
	return &AssociativeAddressTrie{associativeAddressTrie{trie.constructAddedNodesTree()}}
}

// Contains returns whether the given address or prefix block subnet is in the trie as an added element.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the prefix block or address exists already in the trie, false otherwise.
//
// Use GetAddedNode  to get the node for the address rather than just checking for its existence.
func (trie *AssociativeAddressTrie) Contains(addr *Address) bool {
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
func (trie *AssociativeAddressTrie) Remove(addr *Address) bool {
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
func (trie *AssociativeAddressTrie) RemoveElementsContainedBy(addr *Address) *AssociativeAddressTrieNode {
	return trie.removeElementsContainedBy(addr).ToAssociative()
}

// ElementsContainedBy checks if a part of this trie is contained by the given prefix block subnet or individual address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the root node of the contained sub-trie, or nil if no sub-trie is contained.
// The node returned need not be an "added" node, see IsAdded for more details on added nodes.
// The returned sub-trie is backed by this trie, so changes in this trie are reflected in those nodes and vice-versa.
func (trie *AssociativeAddressTrie) ElementsContainedBy(addr *Address) *AssociativeAddressTrieNode {
	return trie.elementsContainedBy(addr).ToAssociative()
}

// ElementsContaining finds the trie nodes in the trie containing the given key and returns them as a linked list.
// Only added nodes are added to the linked list
//
// If the argument is not a single address nor prefix block, this method will panic.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (trie *AssociativeAddressTrie) ElementsContaining(addr *Address) *ContainmentPath {
	return trie.elementsContaining(addr)
}

// LongestPrefixMatch returns the address with the longest matching prefix compared to the provided address
func (trie *AssociativeAddressTrie) LongestPrefixMatch(addr *Address) *Address {
	return trie.longestPrefixMatch(addr)
}

// only added nodes are added to the linked list

// LongestPrefixMatchNode returns the node of the address with the longest matching prefix compared to the provided address
func (trie *AssociativeAddressTrie) LongestPrefixMatchNode(addr *Address) *AssociativeAddressTrieNode {
	return trie.longestPrefixMatchNode(addr).ToAssociative()
}

// ElementContains checks if a prefix block subnet or address in the trie contains the given subnet or address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the subnet or address is contained by a trie element, false otherwise.
//
// To get all the containing addresses, use ElementsContaining
func (trie *AssociativeAddressTrie) ElementContains(addr *Address) bool {
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
func (trie *AssociativeAddressTrie) GetNode(addr *Address) *AssociativeAddressTrieNode {
	return trie.getNode(addr).ToAssociative()
}

// GetAddedNode gets trie nodes representing added elements.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Use Contains to check for the existence of a given address in the trie,
// as well as GetNode to search for all nodes including those not added but also auto-generated nodes for subnet blocks.
func (trie *AssociativeAddressTrie) GetAddedNode(addr *Address) *AssociativeAddressTrieNode {
	return trie.getAddedNode(addr).ToAssociative()
}

// NodeIterator returns an iterator that iterates through all the added nodes in the trie in forward or reverse tree order.
func (trie *AssociativeAddressTrie) NodeIterator(forward bool) AssociativeAddressTrieNodeIteratorRem {
	return associativeAddressTrieNodeIteratorRem{trie.tobase().nodeIterator(forward)}
}

// AllNodeIterator returns an iterator that iterates through all the nodes in the trie in forward or reverse tree order.
func (trie *AssociativeAddressTrie) AllNodeIterator(forward bool) AssociativeAddressTrieNodeIteratorRem {
	return associativeAddressTrieNodeIteratorRem{trie.tobase().allNodeIterator(forward)}
}

// BlockSizeNodeIterator returns an iterator that iterates the added nodes in the trie, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *AssociativeAddressTrie) BlockSizeNodeIterator(lowerSubNodeFirst bool) AssociativeAddressTrieNodeIteratorRem {
	return associativeAddressTrieNodeIteratorRem{trie.tobase().blockSizeNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeAllNodeIterator returns an iterator that iterates all nodes in the trie, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *AssociativeAddressTrie) BlockSizeAllNodeIterator(lowerSubNodeFirst bool) AssociativeAddressTrieNodeIteratorRem {
	return associativeAddressTrieNodeIteratorRem{trie.tobase().blockSizeAllNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeCachingAllNodeIterator returns an iterator that iterates all nodes, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
func (trie *AssociativeAddressTrie) BlockSizeCachingAllNodeIterator() CachingAssociativeAddressTrieNodeIterator {
	return cachingAssociativeAddressTrieNodeIterator{trie.tobase().blockSizeCachingAllNodeIterator()}
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
// The caching and retrieval is done in constant-time and linear space (proportional to tree size).
func (trie *AssociativeAddressTrie) ContainingFirstIterator(forwardSubNodeOrder bool) CachingAssociativeAddressTrieNodeIterator {
	return cachingAssociativeAddressTrieNodeIterator{trie.tobase().containingFirstIterator(forwardSubNodeOrder)}
}

// ContainingFirstAllNodeIterator returns an iterator that does a pre-order binary tree traversal.
// All nodes will be visited before their sub-nodes.
// For an address trie this means containing subnet blocks will be visited before their contained addresses and subnet blocks.
//
// Once a given node is visited, the iterator allows you to cache an object corresponding to the
// lower or upper sub-node that can be retrieved when you later visit that sub-node.
// That allows you to provide iteration context from a parent to its sub-nodes when iterating.
// The caching and retrieval is done in constant-time and linear space (proportional to tree size).
func (trie *AssociativeAddressTrie) ContainingFirstAllNodeIterator(forwardSubNodeOrder bool) CachingAssociativeAddressTrieNodeIterator {
	return cachingAssociativeAddressTrieNodeIterator{trie.tobase().containingFirstAllNodeIterator(forwardSubNodeOrder)}
}

// ContainedFirstIterator returns an iterator that does a post-order binary tree traversal of the added nodes.
// All added sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (trie *AssociativeAddressTrie) ContainedFirstIterator(forwardSubNodeOrder bool) AssociativeAddressTrieNodeIteratorRem {
	return associativeAddressTrieNodeIteratorRem{trie.tobase().containedFirstIterator(forwardSubNodeOrder)}
}

// ContainedFirstAllNodeIterator returns an iterator that does a post-order binary tree traversal.
// All sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (trie *AssociativeAddressTrie) ContainedFirstAllNodeIterator(forwardSubNodeOrder bool) AssociativeAddressTrieNodeIterator {
	return associativeAddressTrieNodeIterator{trie.tobase().containedFirstAllNodeIterator(forwardSubNodeOrder)}
}

// FirstNode returns the first (lowest-valued) node in the trie or nil if the trie is empty
func (trie *AssociativeAddressTrie) FirstNode() *AssociativeAddressTrieNode {
	return toAssociativeAddressTrieNode(trie.trie.FirstNode())
}

// FirstAddedNode returns the first (lowest-valued) added node in the trie
// or nil if there are no added entries in this tree
func (trie *AssociativeAddressTrie) FirstAddedNode() *AssociativeAddressTrieNode {
	return toAssociativeAddressTrieNode(trie.trie.FirstAddedNode())
}

// LastNode returns the last (highest-valued) node in the trie or nil if the trie is empty
func (trie *AssociativeAddressTrie) LastNode() *AssociativeAddressTrieNode {
	return toAssociativeAddressTrieNode(trie.trie.LastNode())
}

// LastAddedNode returns the last (highest-valued) added node in the sub-trie originating from this node,
// or nil if there are no added entries in this tree
func (trie *AssociativeAddressTrie) LastAddedNode() *AssociativeAddressTrieNode {
	return toAssociativeAddressTrieNode(trie.trie.LastAddedNode())
}

// LowerAddedNode returns the added node whose address is the highest address strictly less than the given address,
// or nil if there are no added entries in this tree
func (trie *AssociativeAddressTrie) LowerAddedNode(addr *Address) *AssociativeAddressTrieNode {
	return trie.lowerAddedNode(addr).ToAssociative()
}

// FloorAddedNode returns the added node whose address is the highest address less than or equal to the given address,
// or nil if there are no added entries in this tree
func (trie *AssociativeAddressTrie) FloorAddedNode(addr *Address) *AssociativeAddressTrieNode {
	return trie.floorAddedNode(addr).ToAssociative()
}

// HigherAddedNode returns the added node whose address is the lowest address strictly greater than the given address,
// or nil if there are no added entries in this tree
func (trie *AssociativeAddressTrie) HigherAddedNode(addr *Address) *AssociativeAddressTrieNode {
	return trie.higherAddedNode(addr).ToAssociative()
}

// CeilingAddedNode returns the added node whose address is the lowest address greater than or equal to the given address,
// or nil if there are no added entries in this tree
func (trie *AssociativeAddressTrie) CeilingAddedNode(addr *Address) *AssociativeAddressTrieNode {
	return trie.ceilingAddedNode(addr).ToAssociative()
}

// Clone clones this trie
func (trie *AssociativeAddressTrie) Clone() *AssociativeAddressTrie {
	return trie.tobase().clone().ToAssociative()
}

// Equal returns whether the given argument is a trie with a set of nodes with the same keys as in this trie
func (trie *AssociativeAddressTrie) Equal(other *AssociativeAddressTrie) bool {
	return trie.toTrie().Equal(other.toTrie())
}

// DeepEqual returns whether the given argument is a trie with a set of nodes with the same keys and values as in this trie,
// the values being compared with reflect.DeepEqual
func (trie *AssociativeAddressTrie) DeepEqual(other *AssociativeAddressTrie) bool {
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
func (trie *AssociativeAddressTrie) Put(addr *Address, value NodeValue) (bool, NodeValue) {
	return trie.put(addr, value)
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
func (trie *AssociativeAddressTrie) PutTrie(added *AssociativeAddressTrieNode) *AssociativeAddressTrieNode {
	return trie.putTrie(added.toBase())
}

// PutNode associates the specified value with the specified key in this map.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the node for the added address, whether it was already in the tree or not.
//
// If you wish to know whether the node was already there when adding, use PutNew, or before adding you can use GetAddedNode.
func (trie *AssociativeAddressTrie) PutNode(addr *Address, value NodeValue) *AssociativeAddressTrieNode {
	return trie.putNode(addr, value)
}

// Remap remaps node values in the trie.
//
// This will lookup the node corresponding to the given key.
// It will call the remapping function with the key as the first argument, regardless of whether the node is found or not.
//
// If the node is not found, the value argument will be nil.
// If the node is found, the value argument will be the node's value, which can also be nil.
//
// If the remapping function returns nil, then the matched node will be removed, if any.
// If it returns a non-nil value, then it will either set the existing node to have that value,
// or if there was no matched node, it will create a new node with that value.
//
// The method will return the node involved, which is either the matched node, or the newly created node,
// or nil if there was no matched node nor newly created node.
//
// If the remapping function modifies the trie during its computation,
// and the returned value specifies changes to be made,
// then the trie will not be changed and a panic will ensue.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (trie *AssociativeAddressTrie) Remap(addr *Address, remapper func(NodeValue) NodeValue) *AssociativeAddressTrieNode {
	return trie.remap(addr, remapper)
}

// RemapIfAbsent remaps node values in the trie, but only for nodes that do not exist or are mapped to nil.
//
// This will look up the node corresponding to the given key.
// If the node is not found or mapped to nil, this will call the remapping function.
//
// If the remapping function returns a non-nil value, then it will either set the existing node to have that value,
// or if there was no matched node, it will create a new node with that value.
// If the remapping function returns nil, then it will do the same if insertNil is true, otherwise it will do nothing.
//
// The method will return the node involved, which is either the matched node, or the newly created node,
// or nil if there was no matched node nor newly created node.
//
// If the remapping function modifies the trie during its computation,
// and the returned value specifies changes to be made,
// then the trie will not be changed and ConcurrentModificationException will be thrown instead.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// insertNull indicates whether nil values returned from remapper should be inserted into the map, or whether nil values indicate no remapping
func (trie *AssociativeAddressTrie) RemapIfAbsent(addr *Address, supplier func() NodeValue, insertNil bool) *AssociativeAddressTrieNode {
	return trie.remapIfAbsent(addr, supplier, insertNil)
}

// Get gets the specified value for the specified key in this mapped trie or sub-trie.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the value for the given key.
// Returns nil if the contains no mapping for that key or if the mapped value is nil.
func (trie *AssociativeAddressTrie) Get(addr *Address) NodeValue {
	return trie.get(addr)
}

// For some reason Format must be here and not in addressTrieNode for nil node.
// It panics in fmt code either way, but if in here then it is handled by a recover() call in fmt properly in the debugger.

// Format implements the fmt.Formatter interface
func (trie AssociativeAddressTrie) Format(state fmt.State, verb rune) {
	trie.ToBase().Format(state, verb)
}

// Ensures the address is either an individual address or a prefix block subnet.
// Returns a normalized address which has no prefix length if it is a single address,
// or has a prefix length matching the prefix block size if it is a prefix block.
func checkBlockOrAddress(addr *Address) (res *Address, err addrerr.IncompatibleAddressError) {
	res = addr.ToSinglePrefixBlockOrAddress()
	if res == nil {
		err = &incompatibleAddressError{addressError{key: "ipaddress.error.address.not.block"}}
	}
	return
}

// Ensures the address is either an individual address or a prefix block subnet.
func mustBeBlockOrAddress(addr *Address) *Address {
	res, err := checkBlockOrAddress(addr)
	if res == nil {
		panic(err)
	}
	return res
}
