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
	"unsafe"
)

// IPv6AddressTrie represents an IPv6 address binary trie.
//
// The keys are IPv6 addresses or prefix blocks.
//
// The zero value for IPv6AddressTrie is a binary trie ready for use.
type IPv6AddressTrie struct {
	addressTrie
}

func (trie *IPv6AddressTrie) toTrie() *tree.BinTrie {
	return (*tree.BinTrie)(unsafe.Pointer(trie))
}

// toBase converts to addressTrie by pointer conversion, avoiding dereferencing, which works with nil pointers
func (trie *IPv6AddressTrie) toBase() *addressTrie {
	return (*addressTrie)(unsafe.Pointer(trie))
}

// ToBase converts to the polymorphic representation of this trie
func (trie *IPv6AddressTrie) ToBase() *AddressTrie {
	return (*AddressTrie)(trie)
}

// ToAssociative converts to the associative representation of this trie
func (trie *IPv6AddressTrie) ToAssociative() *IPv6AddressAssociativeTrie {
	return (*IPv6AddressAssociativeTrie)(unsafe.Pointer(trie))
}

// GetRoot returns the root node of this trie, which can be nil for an implicitly zero-valued uninitialized trie, but not for any other trie
func (trie *IPv6AddressTrie) GetRoot() *IPv6AddressTrieNode {
	return trie.getRoot().ToIPv6()
}

// Size returns the number of elements in the tree.
// It does not return the number of nodes.
// Only nodes for which IsAdded() returns true are counted (those nodes corresponding to added addresses and prefix blocks).
// When zero is returned, IsEmpty() returns true.
func (trie *IPv6AddressTrie) Size() int {
	return trie.toTrie().Size()
}

// NodeSize returns the number of nodes in the tree, which is always more than the number of elements.
func (trie *IPv6AddressTrie) NodeSize() int {
	return trie.toTrie().NodeSize()
}

// IsEmpty returns true if there are not any added nodes within this tree
func (trie *IPv6AddressTrie) IsEmpty() bool {
	return trie.Size() == 0
}

// TreeString returns a visual representation of the tree with one node per line, with or without the non-added keys.
func (trie *IPv6AddressTrie) TreeString(withNonAddedKeys bool) string {
	return trie.toTrie().TreeString(withNonAddedKeys)
}

// String returns a visual representation of the tree with one node per line.
func (trie *IPv6AddressTrie) String() string {
	return trie.toTrie().String()
}

// AddedNodesTreeString provides a flattened version of the trie showing only the contained added nodes and their containment structure, which is non-binary.
// The root node is included, which may or may not be added.
func (trie *IPv6AddressTrie) AddedNodesTreeString() string {
	return trie.toTrie().AddedNodesTreeString()
}

// Iterator returns an iterator that iterates through the elements of the sub-trie with this node as the root.
// The iteration is in sorted element order.
func (trie *IPv6AddressTrie) Iterator() IPv6AddressIterator {
	return ipv6AddressIterator{trie.toBase().iterator()}
}

// DescendingIterator returns an iterator that iterates through the elements of the sub-trie with this node as the root.
// The iteration is in reverse sorted element order.
func (trie *IPv6AddressTrie) DescendingIterator() IPv6AddressIterator {
	return ipv6AddressIterator{trie.toBase().descendingIterator()}
}

// Add adds the address to this trie.
// Returns true if the address did not already exist in the trie.
func (trie *IPv6AddressTrie) Add(addr *IPv6Address) bool {
	return trie.add(addr.ToAddressBase())
}

// AddNode adds the address to this trie.
// The new or existing node for the address is returned.
func (trie *IPv6AddressTrie) AddNode(addr *IPv6Address) *IPv6AddressTrieNode {
	return trie.addNode(addr.ToAddressBase()).ToIPv6()
}

// ConstructAddedNodesTree provides an associative trie in which the root and each added node are mapped to a list of their respective direct added sub-nodes.
// This trie provides an alternative non-binary tree structure of the added nodes.
// It is used by ToAddedNodesTreeString to produce a string showing the alternative structure.
// If there are no non-added nodes in this trie, then the alternative tree structure provided by this method is the same as the original trie.
func (trie *IPv6AddressTrie) ConstructAddedNodesTree() *IPv6AddressTrie {
	return &IPv6AddressTrie{trie.constructAddedNodesTree()}
}

// AddTrie adds nodes for the keys in the trie with the root node as the passed in node.  To add both keys and values, use PutTrie.
func (trie *IPv6AddressTrie) AddTrie(added *IPv6AddressTrieNode) *IPv6AddressTrieNode {
	return trie.addTrie(added.toBase()).ToIPv6()
}

// Contains returns whether the given address or prefix block subnet is in the trie as an added element.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the prefix block or address exists already in the trie, false otherwise.
//
// Use GetAddedNode  to get the node for the address rather than just checking for its existence.
func (trie *IPv6AddressTrie) Contains(addr *IPv6Address) bool {
	return trie.contains(addr.ToAddressBase())
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
func (trie *IPv6AddressTrie) Remove(addr *IPv6Address) bool {
	return trie.remove(addr.ToAddressBase())
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
func (trie *IPv6AddressTrie) RemoveElementsContainedBy(addr *IPv6Address) *IPv6AddressTrieNode {
	return trie.removeElementsContainedBy(addr.ToAddressBase()).ToIPv6()
}

// ElementsContainedBy checks if a part of this trie is contained by the given prefix block subnet or individual address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the root node of the contained sub-trie, or nil if no sub-trie is contained.
// The node returned need not be an "added" node, see IsAdded for more details on added nodes.
// The returned sub-trie is backed by this trie, so changes in this trie are reflected in those nodes and vice-versa.
func (trie *IPv6AddressTrie) ElementsContainedBy(addr *IPv6Address) *IPv6AddressTrieNode {
	return trie.elementsContainedBy(addr.ToAddressBase()).ToIPv6()
}

// ElementsContaining finds the trie nodes in the trie containing the given key and returns them as a linked list.
// Only added nodes are added to the linked list
//
// If the argument is not a single address nor prefix block, this method will panic.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (trie *IPv6AddressTrie) ElementsContaining(addr *IPv6Address) *ContainmentPath {
	return trie.elementsContaining(addr.ToAddressBase())
}

// LongestPrefixMatch returns the address added to the trie with the longest matching prefix compared to the provided address, or nil if no matching address
func (trie *IPv6AddressTrie) LongestPrefixMatch(addr *IPv6Address) *IPv6Address {
	return trie.longestPrefixMatch(addr.ToAddressBase()).ToIPv6()
}

// only added nodes are added to the linked list

// LongestPrefixMatchNode returns the node of address added to the trie with the longest matching prefix compared to the provided address, or nil if no matching address
func (trie *IPv6AddressTrie) LongestPrefixMatchNode(addr *IPv6Address) *IPv6AddressTrieNode {
	return trie.longestPrefixMatchNode(addr.ToAddressBase()).ToIPv6()
}

// ElementContains checks if a prefix block subnet or address in the trie contains the given subnet or address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the subnet or address is contained by a trie element, false otherwise.
//
// To get all the containing addresses, use ElementsContaining
func (trie *IPv6AddressTrie) ElementContains(addr *IPv6Address) bool {
	return trie.elementContains(addr.ToAddressBase())
}

// GetNode gets the node in the trie corresponding to the given address,
// or returns nil if not such element exists.
//
// It returns any node, whether added or not,
// including any prefix block node that was not added.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (trie *IPv6AddressTrie) GetNode(addr *IPv6Address) *IPv6AddressTrieNode {
	return trie.getNode(addr.ToAddressBase()).ToIPv6()
}

// GetAddedNode gets trie nodes representing added elements.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Use Contains to check for the existence of a given address in the trie,
// as well as GetNode to search for all nodes including those not-added but also auto-generated nodes for subnet blocks.
func (trie *IPv6AddressTrie) GetAddedNode(addr *IPv6Address) *IPv6AddressTrieNode {
	return trie.getAddedNode(addr.ToAddressBase()).ToIPv6()
}

// AllNodeIterator returns an iterator that iterates through all the nodes of the trie in forward or reverse tree order.
func (trie *IPv6AddressTrie) AllNodeIterator(forward bool) IPv6TrieNodeIteratorRem {
	return ipv6TrieNodeIteratorRem{trie.toBase().allNodeIterator(forward)}
}

// NodeIterator returns an iterator that iterates through the added nodes of the trie in forward or reverse tree order.
func (trie *IPv6AddressTrie) NodeIterator(forward bool) IPv6TrieNodeIteratorRem {
	return ipv6TrieNodeIteratorRem{trie.toBase().nodeIterator(forward)}
}

// BlockSizeNodeIterator returns an iterator that iterates the added nodes in the trie, ordered by keys from the largest prefix blocks to the smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *IPv6AddressTrie) BlockSizeNodeIterator(lowerSubNodeFirst bool) IPv6TrieNodeIteratorRem {
	return ipv6TrieNodeIteratorRem{trie.toBase().blockSizeNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeAllNodeIterator returns an iterator that iterates all nodes in the trie, ordered by keys from the largest prefix blocks to the smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *IPv6AddressTrie) BlockSizeAllNodeIterator(lowerSubNodeFirst bool) IPv6TrieNodeIteratorRem {
	return ipv6TrieNodeIteratorRem{trie.toBase().blockSizeAllNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeCachingAllNodeIterator returns an iterator that iterates all nodes, ordered by keys from the largest prefix blocks to the smallest, and then to individual addresses.
func (trie *IPv6AddressTrie) BlockSizeCachingAllNodeIterator() CachingIPv6TrieNodeIterator {
	return cachingIPv6TrieNodeIterator{trie.toBase().blockSizeCachingAllNodeIterator()}
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
func (trie *IPv6AddressTrie) ContainingFirstIterator(forwardSubNodeOrder bool) CachingIPv6TrieNodeIterator {
	return cachingIPv6TrieNodeIterator{trie.toBase().containingFirstIterator(forwardSubNodeOrder)}
}

// ContainingFirstAllNodeIterator returns an iterator that does a pre-order binary tree traversal.
// All nodes will be visited before their sub-nodes.
// For an address trie this means containing subnet blocks will be visited before their contained addresses and subnet blocks.
//
// Once a given node is visited, the iterator allows you to cache an object corresponding to the
// lower or upper sub-node that can be retrieved when you later visit that sub-node.
// That allows you to provide iteration context from a parent to its sub-nodes when iterating.
// The caching and retrieval is done in constant-time and linear space (proportional to tree size).
func (trie *IPv6AddressTrie) ContainingFirstAllNodeIterator(forwardSubNodeOrder bool) CachingIPv6TrieNodeIterator {
	return cachingIPv6TrieNodeIterator{trie.toBase().containingFirstAllNodeIterator(forwardSubNodeOrder)}
}

// ContainedFirstIterator returns an iterator that does a post-order binary tree traversal of the added nodes.
// All added sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (trie *IPv6AddressTrie) ContainedFirstIterator(forwardSubNodeOrder bool) IPv6TrieNodeIteratorRem {
	return ipv6TrieNodeIteratorRem{trie.toBase().containedFirstIterator(forwardSubNodeOrder)}
}

// ContainedFirstAllNodeIterator returns an iterator that does a post-order binary tree traversal.
// All sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (trie *IPv6AddressTrie) ContainedFirstAllNodeIterator(forwardSubNodeOrder bool) IPv6TrieNodeIterator {
	return ipv6TrieNodeIterator{trie.toBase().containedFirstAllNodeIterator(forwardSubNodeOrder)}
}

// FirstNode returns the first (lowest-valued) node in the trie or nil if the trie has no nodes
func (trie *IPv6AddressTrie) FirstNode() *IPv6AddressTrieNode {
	return toIPv6AddressTrieNode(trie.trie.FirstNode())
}

// FirstAddedNode returns the first (lowest-valued) added node in the trie,
// or nil if there are no added entries in this tree
func (trie *IPv6AddressTrie) FirstAddedNode() *IPv6AddressTrieNode {
	return toIPv6AddressTrieNode(trie.trie.FirstAddedNode())
}

// LastNode returns the last (highest-valued) node in the trie or nil if the trie has no nodes
func (trie *IPv6AddressTrie) LastNode() *IPv6AddressTrieNode {
	return toIPv6AddressTrieNode(trie.trie.LastNode())
}

// LastAddedNode returns the last (highest-valued) added node in the sub-trie originating from this node,
// or nil if there are no added entries in this tree
func (trie *IPv6AddressTrie) LastAddedNode() *IPv6AddressTrieNode {
	return toIPv6AddressTrieNode(trie.trie.LastAddedNode())
}

// LowerAddedNode returns the added node whose address is the highest address strictly less than the given address,
// or nil if there are no added entries in this tree
func (trie *IPv6AddressTrie) LowerAddedNode(addr *IPv6Address) *IPv6AddressTrieNode {
	return trie.lowerAddedNode(addr.ToAddressBase()).ToIPv6()
}

// FloorAddedNode returns the added node whose address is the highest address less than or equal to the given address,
// or nil if there are no added entries in this tree
func (trie *IPv6AddressTrie) FloorAddedNode(addr *IPv6Address) *IPv6AddressTrieNode {
	return trie.floorAddedNode(addr.ToAddressBase()).ToIPv6()
}

// HigherAddedNode returns the added node whose address is the lowest address strictly greater than the given address,
// or nil if there are no added entries in this tree
func (trie *IPv6AddressTrie) HigherAddedNode(addr *IPv6Address) *IPv6AddressTrieNode {
	return trie.higherAddedNode(addr.ToAddressBase()).ToIPv6()
}

// CeilingAddedNode returns the added node whose address is the lowest address greater than or equal to the given address,
// or nil if there are no added entries in this tree
func (trie *IPv6AddressTrie) CeilingAddedNode(addr *IPv6Address) *IPv6AddressTrieNode {
	return trie.ceilingAddedNode(addr.ToAddressBase()).ToIPv6()
}

// Clone clones this trie
func (trie *IPv6AddressTrie) Clone() *IPv6AddressTrie {
	return trie.toBase().clone().ToIPv6()
	//return &IPv6AddressTrie{trie.clone()}
}

// Equal returns whether the given argument is a trie with a set of nodes with the same keys as in this trie
func (trie *IPv6AddressTrie) Equal(other *IPv6AddressTrie) bool {
	//return trie.equal(other.addressTrie)
	return trie.toTrie().Equal(other.toTrie())
}

// For some reason Format must be here and not in addressTrieNode for nil node.
// It panics in fmt code either way, but if in here then it is handled by a recover() call in fmt properly in the debugger.

// Format implements the fmt.Formatter interface
func (trie IPv6AddressTrie) Format(state fmt.State, verb rune) {
	trie.trie.Format(state, verb)
}

// NewIPv6AddressTrie constructs an IPv6 address trie with the root as the ::/0 prefix block
func NewIPv6AddressTrie() *IPv6AddressTrie {
	return &IPv6AddressTrie{addressTrie{
		tree.NewBinTrie(&addressTrieKey{ipv6All.ToAddressBase()})},
	}
}

////////
////////
////////
////////
////////
////////
////////
////////
////////
////////
////////
////////
////////
////////
////////
////////

// IPv6AddressAssociativeTrie represents an IPv6 address associative binary trie.
//
// The keys are IPv6 addresses or prefix blocks.  Each can be mapped to a value.
//
// The zero value for IPv6AddressAssociativeTrie is a binary trie ready for use.
type IPv6AddressAssociativeTrie struct {
	associativeAddressTrie
}

func (trie *IPv6AddressAssociativeTrie) toTrie() *tree.BinTrie {
	return (*tree.BinTrie)(unsafe.Pointer(trie))
}

// toBase is used to convert the pointer rather than doing a field dereference, so that nil pointer handling can be done in *addressTrieNode
func (trie *IPv6AddressAssociativeTrie) toBase() *addressTrie {
	return (*addressTrie)(unsafe.Pointer(trie))
}

// ToBase converts to the polymorphic non-associative representation of this trie
func (trie *IPv6AddressAssociativeTrie) ToBase() *AddressTrie {
	return (*AddressTrie)(unsafe.Pointer(trie))
}

// ToIPv6Base converts to the non-associative representation of this trie
func (trie *IPv6AddressAssociativeTrie) ToIPv6Base() *IPv6AddressTrie {
	return (*IPv6AddressTrie)(unsafe.Pointer(trie))
}

// ToAssociativeBase converts to the polymorphic associative trie representation of this trie
func (trie *IPv6AddressAssociativeTrie) ToAssociativeBase() *AssociativeAddressTrie {
	return (*AssociativeAddressTrie)(trie)
}

// GetRoot returns the root node of this trie, which can be nil for an implicitly zero-valued uninitialized trie, but not for any other trie
func (trie *IPv6AddressAssociativeTrie) GetRoot() *IPv6AddressAssociativeTrieNode {
	return trie.getRoot().ToIPv6Associative()
}

// Size returns the number of elements in the tree.
// It does not return the number of nodes.
// Only nodes for which IsAdded() returns true are counted (those nodes corresponding to added addresses and prefix blocks).
// When zero is returned, IsEmpty() returns true.
func (trie *IPv6AddressAssociativeTrie) Size() int {
	return trie.toTrie().Size()
}

// NodeSize returns the number of nodes in the tree, which is always more than the number of elements.
func (trie *IPv6AddressAssociativeTrie) NodeSize() int {
	return trie.toTrie().NodeSize()
}

// IsEmpty returns true if there are not any added nodes within this tree
func (trie *IPv6AddressAssociativeTrie) IsEmpty() bool {
	return trie.Size() == 0
}

// TreeString returns a visual representation of the tree with one node per line, with or without the non-added keys.
func (trie *IPv6AddressAssociativeTrie) TreeString(withNonAddedKeys bool) string {
	return trie.toTrie().TreeString(withNonAddedKeys)
}

// String returns a visual representation of the tree with one node per line.
func (trie *IPv6AddressAssociativeTrie) String() string {
	return trie.toTrie().String()
}

// AddedNodesTreeString provides a flattened version of the trie showing only the contained added nodes and their containment structure, which is non-binary.
// The root node is included, which may or may not be added.
func (trie *IPv6AddressAssociativeTrie) AddedNodesTreeString() string {
	return trie.toTrie().AddedNodesTreeString()
}

// Iterator returns an iterator that iterates through the elements of the sub-trie with this node as the root.
// The iteration is in sorted element order.
func (trie *IPv6AddressAssociativeTrie) Iterator() IPv6AddressIterator {
	return ipv6AddressIterator{trie.toBase().iterator()}
}

// DescendingIterator returns an iterator that iterates through the elements of the sub-trie with this node as the root.
// The iteration is in reverse sorted element order.
func (trie *IPv6AddressAssociativeTrie) DescendingIterator() IPv6AddressIterator {
	return ipv6AddressIterator{trie.toBase().descendingIterator()}
}

// Add adds the given address key to the trie, returning true if not there already.
func (trie *IPv6AddressAssociativeTrie) Add(addr *IPv6Address) bool {
	return trie.add(addr.ToAddressBase())
}

// AddNode adds the address key to this trie.
// The new or existing node for the address is returned.
func (trie *IPv6AddressAssociativeTrie) AddNode(addr *IPv6Address) *IPv6AddressAssociativeTrieNode {
	return trie.addNode(addr.ToAddressBase()).ToIPv6Associative()
}

// ConstructAddedNodesTree provides an associative trie in which the root and each added node are mapped to a list of their respective direct added sub-nodes.
// This trie provides an alternative non-binary tree structure of the added nodes.
// It is used by ToAddedNodesTreeString to produce a string showing the alternative structure.
// If there are no non-added nodes in this trie, then the alternative tree structure provided by this method is the same as the original trie.
func (trie *IPv6AddressAssociativeTrie) ConstructAddedNodesTree() *IPv6AddressAssociativeTrie {
	return &IPv6AddressAssociativeTrie{associativeAddressTrie{trie.constructAddedNodesTree()}}
}

// AddTrie adds nodes for the keys in the trie with the root node as the passed in node.  To add both keys and values, use PutTrie.
func (trie *IPv6AddressAssociativeTrie) AddTrie(added *IPv6AddressAssociativeTrieNode) *IPv6AddressAssociativeTrieNode {
	return trie.addTrie(added.toBase()).ToIPv6Associative()
}

// Contains returns whether the given address or prefix block subnet is in the trie as an added element.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the prefix block or address exists already in the trie, false otherwise.
//
// Use GetAddedNode  to get the node for the address rather than just checking for its existence.
func (trie *IPv6AddressAssociativeTrie) Contains(addr *IPv6Address) bool {
	return trie.contains(addr.ToAddressBase())
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
func (trie *IPv6AddressAssociativeTrie) Remove(addr *IPv6Address) bool {
	return trie.remove(addr.ToAddressBase())
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
func (trie *IPv6AddressAssociativeTrie) RemoveElementsContainedBy(addr *IPv6Address) *IPv6AddressAssociativeTrieNode {
	return trie.removeElementsContainedBy(addr.ToAddressBase()).ToIPv6Associative()
}

// ElementsContainedBy checks if a part of this trie is contained by the given prefix block subnet or individual address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the root node of the contained sub-trie, or nil if no sub-trie is contained.
// The node returned need not be an "added" node, see IsAdded for more details on added nodes.
// The returned sub-trie is backed by this trie, so changes in this trie are reflected in those nodes and vice-versa.
func (trie *IPv6AddressAssociativeTrie) ElementsContainedBy(addr *IPv6Address) *IPv6AddressAssociativeTrieNode {
	return trie.elementsContainedBy(addr.ToAddressBase()).ToIPv6Associative()
}

// ElementsContaining finds the trie nodes in the trie containing the given key and returns them as a linked list.
// Only added nodes are added to the linked list
//
// If the argument is not a single address nor prefix block, this method will panic.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (trie *IPv6AddressAssociativeTrie) ElementsContaining(addr *IPv6Address) *ContainmentPath {
	return trie.elementsContaining(addr.ToAddressBase())
}

// LongestPrefixMatch returns the address added to the trie with the longest matching prefix compared to the provided address, or nil if no matching address
func (trie *IPv6AddressAssociativeTrie) LongestPrefixMatch(addr *IPv6Address) *IPv6Address {
	return trie.longestPrefixMatch(addr.ToAddressBase()).ToIPv6()
}

// only added nodes are added to the linked list

// LongestPrefixMatchNode returns the node of address added to the trie with the longest matching prefix compared to the provided address, or nil if no matching address
func (trie *IPv6AddressAssociativeTrie) LongestPrefixMatchNode(addr *IPv6Address) *IPv6AddressAssociativeTrieNode {
	return trie.longestPrefixMatchNode(addr.ToAddressBase()).ToIPv6Associative()
}

// ElementContains checks if a prefix block subnet or address in the trie contains the given subnet or address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the subnet or address is contained by a trie element, false otherwise.
//
// To get all the containing addresses, use ElementsContaining
func (trie *IPv6AddressAssociativeTrie) ElementContains(addr *IPv6Address) bool {
	return trie.elementContains(addr.ToAddressBase())
}

// GetNode gets the node in the trie corresponding to the given address,
// or returns nil if not such element exists.
//
// It returns any node, whether added or not,
// including any prefix block node that was not added.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (trie *IPv6AddressAssociativeTrie) GetNode(addr *IPv6Address) *IPv6AddressAssociativeTrieNode {
	return trie.getNode(addr.ToAddressBase()).ToIPv6Associative()
}

// GetAddedNode gets trie nodes representing added elements.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Use Contains to check for the existence of a given address in the trie,
// as well as GetNode to search for all nodes including those not added but also auto-generated nodes for subnet blocks.
func (trie *IPv6AddressAssociativeTrie) GetAddedNode(addr *IPv6Address) *IPv6AddressAssociativeTrieNode {
	return trie.getAddedNode(addr.ToAddressBase()).ToIPv6Associative()
}

// NodeIterator iterates through the added nodes of the sub-trie with this node as the root, in forward or reverse tree order.
func (trie *IPv6AddressAssociativeTrie) NodeIterator(forward bool) IPv6AssociativeTrieNodeIteratorRem {
	return ipv6AssociativeTrieNodeIteratorRem{trie.toBase().nodeIterator(forward)}
}

// AllNodeIterator returns an iterator that iterates the added all nodes in the trie following the natural trie order
func (trie *IPv6AddressAssociativeTrie) AllNodeIterator(forward bool) IPv6AssociativeTrieNodeIteratorRem {
	return ipv6AssociativeTrieNodeIteratorRem{trie.toBase().allNodeIterator(forward)}
}

// BlockSizeNodeIterator returns an iterator that iterates the added nodes in the trie, ordered by keys from the largest prefix blocks to the smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *IPv6AddressAssociativeTrie) BlockSizeNodeIterator(lowerSubNodeFirst bool) IPv6AssociativeTrieNodeIteratorRem {
	return ipv6AssociativeTrieNodeIteratorRem{trie.toBase().blockSizeNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeAllNodeIterator returns an iterator that iterates all nodes in the trie, ordered by keys from the largest prefix blocks to the smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *IPv6AddressAssociativeTrie) BlockSizeAllNodeIterator(lowerSubNodeFirst bool) IPv6AssociativeTrieNodeIteratorRem {
	return ipv6AssociativeTrieNodeIteratorRem{trie.toBase().blockSizeAllNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeCachingAllNodeIterator returns an iterator that iterates all nodes, ordered by keys from the largest prefix blocks to the smallest, and then to individual addresses.
func (trie *IPv6AddressAssociativeTrie) BlockSizeCachingAllNodeIterator() CachingIPv6AssociativeTrieNodeIterator {
	return cachingIPv6AssociativeTrieNodeIterator{trie.toBase().blockSizeCachingAllNodeIterator()}
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
func (trie *IPv6AddressAssociativeTrie) ContainingFirstIterator(forwardSubNodeOrder bool) CachingIPv6AssociativeTrieNodeIterator {
	return cachingIPv6AssociativeTrieNodeIterator{trie.toBase().containingFirstIterator(forwardSubNodeOrder)}
}

// ContainingFirstAllNodeIterator returns an iterator that does a pre-order binary tree traversal.
// All nodes will be visited before their sub-nodes.
// For an address trie this means containing subnet blocks will be visited before their contained addresses and subnet blocks.
//
// Once a given node is visited, the iterator allows you to cache an object corresponding to the
// lower or upper sub-node that can be retrieved when you later visit that sub-node.
// That allows you to provide iteration context from a parent to its sub-nodes when iterating.
// The caching and retrieval is done in constant-time and linear space (proportional to tree size).
func (trie *IPv6AddressAssociativeTrie) ContainingFirstAllNodeIterator(forwardSubNodeOrder bool) CachingIPv6AssociativeTrieNodeIterator {
	return cachingIPv6AssociativeTrieNodeIterator{trie.toBase().containingFirstAllNodeIterator(forwardSubNodeOrder)}
}

// ContainedFirstIterator returns an iterator that does a post-order binary tree traversal of the added nodes.
// All added sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (trie *IPv6AddressAssociativeTrie) ContainedFirstIterator(forwardSubNodeOrder bool) IPv6AssociativeTrieNodeIteratorRem {
	return ipv6AssociativeTrieNodeIteratorRem{trie.toBase().containedFirstIterator(forwardSubNodeOrder)}
}

// ContainedFirstAllNodeIterator returns an iterator that does a post-order binary tree traversal.
// All sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (trie *IPv6AddressAssociativeTrie) ContainedFirstAllNodeIterator(forwardSubNodeOrder bool) IPv6AssociativeTrieNodeIterator {
	return ipv6AssociativeTrieNodeIterator{trie.toBase().containedFirstAllNodeIterator(forwardSubNodeOrder)}
}

// FirstNode returns the first (lowest-valued) node in the trie or nil if the trie is empty
func (trie *IPv6AddressAssociativeTrie) FirstNode() *IPv6AddressAssociativeTrieNode {
	return toIPv6AAssociativeAddressTrieNode(trie.trie.FirstNode())
}

// FirstAddedNode returns the first (lowest-valued) added node in the trie
// or nil if there are no added entries in this tree
func (trie *IPv6AddressAssociativeTrie) FirstAddedNode() *IPv6AddressAssociativeTrieNode {
	return toIPv6AAssociativeAddressTrieNode(trie.trie.FirstAddedNode())
}

// LastNode returns the last (highest-valued) node in the trie or nil if the trie is empty
func (trie *IPv6AddressAssociativeTrie) LastNode() *IPv6AddressAssociativeTrieNode {
	return toIPv6AAssociativeAddressTrieNode(trie.trie.LastNode())
}

// LastAddedNode returns the last (highest-valued) added node in the sub-trie originating from this node,
// or nil if there are no added entries in this tree
func (trie *IPv6AddressAssociativeTrie) LastAddedNode() *IPv6AddressAssociativeTrieNode {
	return toIPv6AAssociativeAddressTrieNode(trie.trie.LastAddedNode())
}

// LowerAddedNode returns the added node whose address is the highest address strictly less than the given address,
// or nil if there are no added entries in this tree
func (trie *IPv6AddressAssociativeTrie) LowerAddedNode(addr *IPv6Address) *IPv6AddressAssociativeTrieNode {
	return trie.lowerAddedNode(addr.ToAddressBase()).ToIPv6Associative()
}

// FloorAddedNode returns the added node whose address is the highest address less than or equal to the given address,
// or nil if there are no added entries in this tree
func (trie *IPv6AddressAssociativeTrie) FloorAddedNode(addr *IPv6Address) *IPv6AddressAssociativeTrieNode {
	return trie.floorAddedNode(addr.ToAddressBase()).ToIPv6Associative()
}

// HigherAddedNode returns the added node whose address is the lowest address strictly greater than the given address,
// or nil if there are no added entries in this tree
func (trie *IPv6AddressAssociativeTrie) HigherAddedNode(addr *IPv6Address) *IPv6AddressAssociativeTrieNode {
	return trie.higherAddedNode(addr.ToAddressBase()).ToIPv6Associative()
}

// CeilingAddedNode returns the added node whose address is the lowest address greater than or equal to the given address,
// or nil if there are no added entries in this tree
func (trie *IPv6AddressAssociativeTrie) CeilingAddedNode(addr *IPv6Address) *IPv6AddressAssociativeTrieNode {
	return trie.ceilingAddedNode(addr.ToAddressBase()).ToIPv6Associative()
}

// Clone clones this trie
func (trie *IPv6AddressAssociativeTrie) Clone() *IPv6AddressAssociativeTrie {
	return trie.toBase().clone().ToIPv6Associative()
}

// Equal returns whether the given argument is a trie with a set of nodes with the same keys as in this trie
func (trie *IPv6AddressAssociativeTrie) Equal(other *IPv6AddressAssociativeTrie) bool {
	return trie.toTrie().Equal(other.toTrie())
}

// DeepEqual returns whether the given argument is a trie with a set of nodes with the same keys and values as in this trie,
// the values being compared with reflect.DeepEqual
func (trie *IPv6AddressAssociativeTrie) DeepEqual(other *IPv6AddressAssociativeTrie) bool {
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
func (trie *IPv6AddressAssociativeTrie) Put(addr *IPv6Address, value NodeValue) (bool, NodeValue) {
	return trie.put(addr.ToAddressBase(), value)
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
func (trie *IPv6AddressAssociativeTrie) PutTrie(added *IPv6AddressAssociativeTrieNode) *IPv6AddressAssociativeTrieNode {
	return trie.putTrie(added.toBase()).ToIPv6()
}

// PutNode associates the specified value with the specified key in this map.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the node for the added address, whether it was already in the tree or not.
//
// If you wish to know whether the node was already there when adding, use PutNew, or before adding you can use GetAddedNode.
func (trie *IPv6AddressAssociativeTrie) PutNode(addr *IPv6Address, value NodeValue) *IPv6AddressAssociativeTrieNode {
	return trie.putNode(addr.ToAddressBase(), value).ToIPv6()
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
func (trie *IPv6AddressAssociativeTrie) Remap(addr *IPv6Address, remapper func(NodeValue) NodeValue) *IPv6AddressAssociativeTrieNode {
	return trie.remap(addr.ToAddressBase(), remapper).ToIPv6()
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
func (trie *IPv6AddressAssociativeTrie) RemapIfAbsent(addr *IPv6Address, supplier func() NodeValue, insertNil bool) *IPv6AddressAssociativeTrieNode {
	return trie.remapIfAbsent(addr.ToAddressBase(), supplier, insertNil).ToIPv6()
}

// Get gets the specified value for the specified key in this mapped trie or sub-trie.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the value for the given key.
// Returns nil if the contains no mapping for that key or if the mapped value is nil.
func (trie *IPv6AddressAssociativeTrie) Get(addr *IPv6Address) NodeValue {
	return trie.get(addr.ToAddressBase())
}

// Format implements the fmt.Formatter interface
func (trie IPv6AddressAssociativeTrie) Format(state fmt.State, verb rune) {
	trie.ToBase().Format(state, verb)
}

// NewIPv6AddressAssociativeTrie constructs an IPv6 associative address trie with the root as the ::/0 prefix block
func NewIPv6AddressAssociativeTrie() *IPv6AddressAssociativeTrie {
	return &IPv6AddressAssociativeTrie{
		associativeAddressTrie{
			addressTrie{tree.NewBinTrie(&addressTrieKey{ipv6All.ToAddressBase()})},
		},
	}
}
