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

// MACAddressTrie represents a MAC address binary trie.
//
// The keys are MAC addresses or prefix blocks.
//
// The zero value for MACAddressTrie is a binary trie ready for use.
type MACAddressTrie struct {
	addressTrie
}

func (trie *MACAddressTrie) toTrie() *tree.BinTrie {
	return (*tree.BinTrie)(unsafe.Pointer(trie))
}

// toBase converts to addressTrie by pointer conversion, avoiding dereferencing, which works with nil pointers
func (trie *MACAddressTrie) toBase() *addressTrie {
	return (*addressTrie)(unsafe.Pointer(trie))
}

// ToBase converts to the polymorphic representation of this trie
func (trie *MACAddressTrie) ToBase() *AddressTrie {
	return (*AddressTrie)(trie)
}

// ToAssociative converts to the associative representation of this trie
func (trie *MACAddressTrie) ToAssociative() *MACAddressAssociativeTrie {
	return (*MACAddressAssociativeTrie)(unsafe.Pointer(trie))
}

// GetRoot returns the root node of this trie, which can be nil for an implicitly zero-valued uninitialized trie, but not for any other trie
func (trie *MACAddressTrie) GetRoot() *MACAddressTrieNode {
	return trie.getRoot().ToMAC()
}

// Size returns the number of elements in the tree.
// It does not return the number of nodes.
// Only nodes for which IsAdded() returns true are counted (those nodes corresponding to added addresses and prefix blocks).
// When zero is returned, IsEmpty() returns true.
func (trie *MACAddressTrie) Size() int {
	return trie.toTrie().Size()
}

// NodeSize returns the number of nodes in the tree, which is always more than the number of elements.
func (trie *MACAddressTrie) NodeSize() int {
	return trie.toTrie().NodeSize()
}

// IsEmpty returns true if there are not any added nodes within this tree
func (trie *MACAddressTrie) IsEmpty() bool {
	return trie.Size() == 0
}

// TreeString returns a visual representation of the tree with one node per line, with or without the non-added keys.
func (trie *MACAddressTrie) TreeString(withNonAddedKeys bool) string {
	return trie.toTrie().TreeString(withNonAddedKeys)
}

// String returns a visual representation of the tree with one node per line.
func (trie *MACAddressTrie) String() string {
	return trie.toTrie().String()
}

// AddedNodesTreeString provides a flattened version of the trie showing only the contained added nodes and their containment structure, which is non-binary.
// The root node is included, which may or may not be added.
func (trie *MACAddressTrie) AddedNodesTreeString() string {
	return trie.toTrie().AddedNodesTreeString()
}

// Iterator returns an iterator that iterates through the added prefix blocks and addresses in the trie.
// The iteration is in sorted element order.
func (trie *MACAddressTrie) Iterator() MACAddressIterator {
	return macAddressIterator{trie.toBase().iterator()}
}

// DescendingIterator returns an iterator that iterates through the added prefix blocks and addresses in the trie.
// The iteration is in reverse sorted element order.
func (trie *MACAddressTrie) DescendingIterator() MACAddressIterator {
	return macAddressIterator{trie.toBase().descendingIterator()}
}

// Add adds the address to this trie.
// Returns true if the address did not already exist in the trie.
func (trie *MACAddressTrie) Add(addr *MACAddress) bool {
	return trie.add(addr.ToAddressBase())
}

// AddNode adds the address to this trie.
// The new or existing node for the address is returned.
func (trie *MACAddressTrie) AddNode(addr *MACAddress) *MACAddressTrieNode {
	return trie.addNode(addr.ToAddressBase()).ToMAC()
}

// ConstructAddedNodesTree provides an associative trie in which the root and each added node are mapped to a list of their respective direct added sub-nodes.
// This trie provides an alternative non-binary tree structure of the added nodes.
// It is used by ToAddedNodesTreeString to produce a string showing the alternative structure.
// If there are no non-added nodes in this trie, then the alternative tree structure provided by this method is the same as the original trie.
func (trie *MACAddressTrie) ConstructAddedNodesTree() *MACAddressTrie {
	return &MACAddressTrie{trie.constructAddedNodesTree()}
}

// AddTrie adds nodes for the keys in the trie with the root node as the passed in node.  To add both keys and values, use PutTrie.
func (trie *MACAddressTrie) AddTrie(added *MACAddressTrieNode) *MACAddressTrieNode {
	return trie.addTrie(added.toBase()).ToMAC()
}

// Contains returns whether the given address or prefix block subnet is in the trie as an added element.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the prefix block or address exists already in the trie, false otherwise.
//
// Use GetAddedNode  to get the node for the address rather than just checking for its existence.
func (trie *MACAddressTrie) Contains(addr *MACAddress) bool {
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
func (trie *MACAddressTrie) Remove(addr *MACAddress) bool {
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
func (trie *MACAddressTrie) RemoveElementsContainedBy(addr *MACAddress) *MACAddressTrieNode {
	return trie.removeElementsContainedBy(addr.ToAddressBase()).ToMAC()
}

// ElementsContainedBy checks if a part of this trie is contained by the given prefix block subnet or individual address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the root node of the contained sub-trie, or nil if no sub-trie is contained.
// The node returned need not be an "added" node, see IsAdded for more details on added nodes.
// The returned sub-trie is backed by this trie, so changes in this trie are reflected in those nodes and vice-versa.
func (trie *MACAddressTrie) ElementsContainedBy(addr *MACAddress) *MACAddressTrieNode {
	return trie.elementsContainedBy(addr.ToAddressBase()).ToMAC()
}

// ElementsContaining finds the trie nodes in the trie containing the given key and returns them as a linked list.
// Only added nodes are added to the linked list
//
// If the argument is not a single address nor prefix block, this method will panic.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (trie *MACAddressTrie) ElementsContaining(addr *MACAddress) *ContainmentPath {
	return trie.elementsContaining(addr.ToAddressBase())
}

// LongestPrefixMatch returns the address added to the trie with the longest matching prefix compared to the provided address, or nil if no matching address
func (trie *MACAddressTrie) LongestPrefixMatch(addr *MACAddress) *MACAddress {
	return trie.longestPrefixMatch(addr.ToAddressBase()).ToMAC()
}

// only added nodes are added to the linked list

// LongestPrefixMatchNode returns the node of address added to the trie with the longest matching prefix compared to the provided address, or nil if no matching address
func (trie *MACAddressTrie) LongestPrefixMatchNode(addr *MACAddress) *MACAddressTrieNode {
	return trie.longestPrefixMatchNode(addr.ToAddressBase()).ToMAC()
}

// ElementContains checks if a prefix block subnet or address in the trie contains the given subnet or address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the subnet or address is contained by a trie element, false otherwise.
//
// To get all the containing addresses, use ElementsContaining
func (trie *MACAddressTrie) ElementContains(addr *MACAddress) bool {
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
func (trie *MACAddressTrie) GetNode(addr *MACAddress) *MACAddressTrieNode {
	return trie.getNode(addr.ToAddressBase()).ToMAC()
}

// GetAddedNode gets trie nodes representing added elements.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Use Contains to check for the existence of a given address in the trie,
// as well as GetNode to search for all nodes including those not-added but also auto-generated nodes for subnet blocks.
func (trie *MACAddressTrie) GetAddedNode(addr *MACAddress) *MACAddressTrieNode {
	return trie.getAddedNode(addr.ToAddressBase()).ToMAC()
}

// AllNodeIterator returns an iterator that iterates through all the nodes in the trie in forward or reverse tree order.
func (trie *MACAddressTrie) AllNodeIterator(forward bool) MACTrieNodeIteratorRem {
	return macTrieNodeIteratorRem{trie.toBase().allNodeIterator(forward)}
}

// NodeIterator returns an iterator that iterates through the added nodes in the trie in forward or reverse tree order.
func (trie *MACAddressTrie) NodeIterator(forward bool) MACTrieNodeIteratorRem {
	return macTrieNodeIteratorRem{trie.toBase().nodeIterator(forward)}
}

// BlockSizeNodeIterator returns an iterator that iterates the added nodes in the trie, ordered by keys from the largest prefix blocks to the smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *MACAddressTrie) BlockSizeNodeIterator(lowerSubNodeFirst bool) MACTrieNodeIteratorRem {
	return macTrieNodeIteratorRem{trie.toBase().blockSizeNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeAllNodeIterator returns an iterator that iterates all nodes in the trie, ordered by keys from the largest prefix blocks to the smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *MACAddressTrie) BlockSizeAllNodeIterator(lowerSubNodeFirst bool) MACTrieNodeIteratorRem {
	return macTrieNodeIteratorRem{trie.toBase().blockSizeAllNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeCachingAllNodeIterator returns an iterator that iterates all nodes, ordered by keys from the largest prefix blocks to the smallest, and then to individual addresses.
func (trie *MACAddressTrie) BlockSizeCachingAllNodeIterator() CachingMACTrieNodeIterator {
	return cachingMACTrieNodeIterator{trie.toBase().blockSizeCachingAllNodeIterator()}
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
func (trie *MACAddressTrie) ContainingFirstIterator(forwardSubNodeOrder bool) CachingMACTrieNodeIterator {
	return cachingMACTrieNodeIterator{trie.toBase().containingFirstIterator(forwardSubNodeOrder)}
}

// ContainingFirstAllNodeIterator returns an iterator that does a pre-order binary tree traversal.
// All nodes will be visited before their sub-nodes.
// For an address trie this means containing subnet blocks will be visited before their contained addresses and subnet blocks.
//
// Once a given node is visited, the iterator allows you to cache an object corresponding to the
// lower or upper sub-node that can be retrieved when you later visit that sub-node.
// That allows you to provide iteration context from a parent to its sub-nodes when iterating.
// The caching and retrieval is done in constant-time and linear space (proportional to tree size).
func (trie *MACAddressTrie) ContainingFirstAllNodeIterator(forwardSubNodeOrder bool) CachingMACTrieNodeIterator {
	return cachingMACTrieNodeIterator{trie.toBase().containingFirstAllNodeIterator(forwardSubNodeOrder)}
}

// ContainedFirstIterator returns an iterator that does a post-order binary tree traversal of the added nodes.
// All added sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (trie *MACAddressTrie) ContainedFirstIterator(forwardSubNodeOrder bool) MACTrieNodeIteratorRem {
	return macTrieNodeIteratorRem{trie.toBase().containedFirstIterator(forwardSubNodeOrder)}
}

// ContainedFirstAllNodeIterator returns an iterator that does a post-order binary tree traversal.
// All sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (trie *MACAddressTrie) ContainedFirstAllNodeIterator(forwardSubNodeOrder bool) MACTrieNodeIterator {
	return macTrieNodeIterator{trie.toBase().containedFirstAllNodeIterator(forwardSubNodeOrder)}
}

// FirstNode returns the first (lowest-valued) node in the trie or nil if the trie has no nodes
func (trie *MACAddressTrie) FirstNode() *MACAddressTrieNode {
	return toMACAddressTrieNode(trie.trie.FirstNode())
}

// FirstAddedNode returns the first (lowest-valued) added node in the trie,
// or nil if there are no added entries in this tree
func (trie *MACAddressTrie) FirstAddedNode() *MACAddressTrieNode {
	return toMACAddressTrieNode(trie.trie.FirstAddedNode())
}

// LastNode returns the last (highest-valued) node in the trie or nil if the trie has no nodes
func (trie *MACAddressTrie) LastNode() *MACAddressTrieNode {
	return toMACAddressTrieNode(trie.trie.LastNode())
}

// LastAddedNode returns the last (highest-valued) added node in the sub-trie originating from this node,
// or nil if there are no added entries in this tree
func (trie *MACAddressTrie) LastAddedNode() *MACAddressTrieNode {
	return toMACAddressTrieNode(trie.trie.LastAddedNode())
}

// LowerAddedNode returns the added node whose address is the highest address strictly less than the given address,
// or nil if there are no added entries in this tree
func (trie *MACAddressTrie) LowerAddedNode(addr *MACAddress) *MACAddressTrieNode {
	return trie.lowerAddedNode(addr.ToAddressBase()).ToMAC()
}

// FloorAddedNode returns the added node whose address is the highest address less than or equal to the given address,
// or nil if there are no added entries in this tree
func (trie *MACAddressTrie) FloorAddedNode(addr *MACAddress) *MACAddressTrieNode {
	return trie.floorAddedNode(addr.ToAddressBase()).ToMAC()
}

// HigherAddedNode returns the added node whose address is the lowest address strictly greater than the given address,
// or nil if there are no added entries in this tree
func (trie *MACAddressTrie) HigherAddedNode(addr *MACAddress) *MACAddressTrieNode {
	return trie.higherAddedNode(addr.ToAddressBase()).ToMAC()
}

// CeilingAddedNode returns the added node whose address is the lowest address greater than or equal to the given address,
// or nil if there are no added entries in this tree
func (trie *MACAddressTrie) CeilingAddedNode(addr *MACAddress) *MACAddressTrieNode {
	return trie.ceilingAddedNode(addr.ToAddressBase()).ToMAC()
}

// Clone clones this trie
func (trie *MACAddressTrie) Clone() *MACAddressTrie {
	return trie.toBase().clone().ToMAC()
	//return &MACAddressTrie{trie.clone()}
}

// Equal returns whether the given argument is a trie with a set of nodes with the same keys as in this trie
func (trie *MACAddressTrie) Equal(other *MACAddressTrie) bool {
	//return trie.equal(other.addressTrie)
	return trie.toTrie().Equal(other.toTrie())
}

// For some reason Format must be here and not in addressTrieNode for nil node.
// It panics in fmt code either way, but if in here then it is handled by a recover() call in fmt properly in the debugger.

// Format implements the fmt.Formatter interface
func (trie MACAddressTrie) Format(state fmt.State, verb rune) {
	trie.trie.Format(state, verb)
}

// NewMACAddressTrie constructs a MAC address trie with the root as the zero-prefix block
// If extended is true, the trie will consist of 64-bit EUI addresses, otherwise the addresses will be 48-bit.
// If you wish to construct a trie in which the address size is determined by the first added address,
// use the zero-value MACAddressTrie{}
func NewMACAddressTrie(extended bool) *MACAddressTrie {
	var rootAddr *Address
	if extended {
		rootAddr = macAllExtended.ToAddressBase()
	} else {
		rootAddr = macAll.ToAddressBase()
	}
	return &MACAddressTrie{addressTrie{
		tree.NewBinTrie(&addressTrieKey{rootAddr})},
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

// MACAddressAssociativeTrie represents a MAC address associative binary trie.
//
// The keys are MAC addresses or prefix blocks.  Each can be mapped to a value.
//
// The zero value for MACAddressAssociativeTrie is a binary trie ready for use.
type MACAddressAssociativeTrie struct {
	associativeAddressTrie
}

func (trie *MACAddressAssociativeTrie) toTrie() *tree.BinTrie {
	return (*tree.BinTrie)(unsafe.Pointer(trie))
}

// toBase is used to convert the pointer rather than doing a field dereference, so that nil pointer handling can be done in *addressTrieNode
func (trie *MACAddressAssociativeTrie) toBase() *addressTrie {
	return (*addressTrie)(unsafe.Pointer(trie))
}

// ToBase converts to the polymorphic non-associative representation of this trie
func (trie *MACAddressAssociativeTrie) ToBase() *AddressTrie {
	return (*AddressTrie)(unsafe.Pointer(trie))
}

// ToMACBase converts to the non-associative representation of this trie
func (trie *MACAddressAssociativeTrie) ToMACBase() *MACAddressTrie {
	return (*MACAddressTrie)(unsafe.Pointer(trie))
}

// ToAssociativeBase converts to the polymorphic associative trie representation of this trie
func (trie *MACAddressAssociativeTrie) ToAssociativeBase() *AssociativeAddressTrie {
	return (*AssociativeAddressTrie)(trie)
}

// GetRoot returns the root node of this trie, which can be nil for an implicitly zero-valued uninitialized trie, but not for any other trie
func (trie *MACAddressAssociativeTrie) GetRoot() *MACAddressAssociativeTrieNode {
	return trie.getRoot().ToMACAssociative()
}

// Size returns the number of elements in the tree.
// It does not return the number of nodes.
// Only nodes for which IsAdded() returns true are counted (those nodes corresponding to added addresses and prefix blocks).
// When zero is returned, IsEmpty() returns true.
func (trie *MACAddressAssociativeTrie) Size() int {
	return trie.toTrie().Size()
}

// NodeSize returns the number of nodes in the tree, which is always more than the number of elements.
func (trie *MACAddressAssociativeTrie) NodeSize() int {
	return trie.toTrie().NodeSize()
}

// IsEmpty returns true if there are not any added nodes within this tree
func (trie *MACAddressAssociativeTrie) IsEmpty() bool {
	return trie.Size() == 0
}

// TreeString returns a visual representation of the tree with one node per line, with or without the non-added keys.
func (trie *MACAddressAssociativeTrie) TreeString(withNonAddedKeys bool) string {
	return trie.toTrie().TreeString(withNonAddedKeys)
}

// String returns a visual representation of the tree with one node per line.
func (trie *MACAddressAssociativeTrie) String() string {
	return trie.toTrie().String()
}

// AddedNodesTreeString provides a flattened version of the trie showing only the contained added nodes and their containment structure, which is non-binary.
// The root node is included, which may or may not be added.
func (trie *MACAddressAssociativeTrie) AddedNodesTreeString() string {
	return trie.toTrie().AddedNodesTreeString()
}

// Iterator returns an iterator that iterates through the added prefix blocks and addresses in the trie.
// The iteration is in sorted element order.
func (trie *MACAddressAssociativeTrie) Iterator() MACAddressIterator {
	return macAddressIterator{trie.toBase().iterator()}
}

// DescendingIterator returns an iterator that iterates through the added prefix blocks and addresses in the trie.
// The iteration is in reverse sorted element order.
func (trie *MACAddressAssociativeTrie) DescendingIterator() MACAddressIterator {
	return macAddressIterator{trie.toBase().descendingIterator()}
}

// Add adds the given address key to the trie, returning true if not there already.
func (trie *MACAddressAssociativeTrie) Add(addr *MACAddress) bool {
	return trie.add(addr.ToAddressBase())
}

// AddNode adds the address key to this trie.
// The new or existing node for the address is returned.
func (trie *MACAddressAssociativeTrie) AddNode(addr *MACAddress) *MACAddressAssociativeTrieNode {
	return trie.addNode(addr.ToAddressBase()).ToMACAssociative()
}

// ConstructAddedNodesTree provides an associative trie in which the root and each added node are mapped to a list of their respective direct added sub-nodes.
// This trie provides an alternative non-binary tree structure of the added nodes.
// It is used by ToAddedNodesTreeString to produce a string showing the alternative structure.
// If there are no non-added nodes in this trie, then the alternative tree structure provided by this method is the same as the original trie.
func (trie *MACAddressAssociativeTrie) ConstructAddedNodesTree() *MACAddressAssociativeTrie {
	return &MACAddressAssociativeTrie{associativeAddressTrie{trie.constructAddedNodesTree()}}
}

// AddTrie adds nodes for the keys in the trie with the root node as the passed in node.  To add both keys and values, use PutTrie.
func (trie *MACAddressAssociativeTrie) AddTrie(added *MACAddressAssociativeTrieNode) *MACAddressAssociativeTrieNode {
	return trie.addTrie(added.toBase()).ToMACAssociative()
}

// Contains returns whether the given address or prefix block subnet is in the trie as an added element.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the prefix block or address exists already in the trie, false otherwise.
//
// Use GetAddedNode  to get the node for the address rather than just checking for its existence.
func (trie *MACAddressAssociativeTrie) Contains(addr *MACAddress) bool {
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
func (trie *MACAddressAssociativeTrie) Remove(addr *MACAddress) bool {
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
func (trie *MACAddressAssociativeTrie) RemoveElementsContainedBy(addr *MACAddress) *MACAddressAssociativeTrieNode {
	return trie.removeElementsContainedBy(addr.ToAddressBase()).ToMACAssociative()
}

// ElementsContainedBy checks if a part of this trie is contained by the given prefix block subnet or individual address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the root node of the contained sub-trie, or nil if no sub-trie is contained.
// The node returned need not be an "added" node, see IsAdded for more details on added nodes.
// The returned sub-trie is backed by this trie, so changes in this trie are reflected in those nodes and vice-versa.
func (trie *MACAddressAssociativeTrie) ElementsContainedBy(addr *MACAddress) *MACAddressAssociativeTrieNode {
	return trie.elementsContainedBy(addr.ToAddressBase()).ToMACAssociative()
}

// ElementsContaining finds the trie nodes in the trie containing the given key and returns them as a linked list.
// Only added nodes are added to the linked list
//
// If the argument is not a single address nor prefix block, this method will panic.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (trie *MACAddressAssociativeTrie) ElementsContaining(addr *MACAddress) *ContainmentPath {
	return trie.elementsContaining(addr.ToAddressBase())
}

// LongestPrefixMatch returns the address added to the trie with the longest matching prefix compared to the provided address, or nil if no matching address
func (trie *MACAddressAssociativeTrie) LongestPrefixMatch(addr *MACAddress) *MACAddress {
	return trie.longestPrefixMatch(addr.ToAddressBase()).ToMAC()
}

// only added nodes are added to the linked list

// LongestPrefixMatchNode returns the node of address added to the trie with the longest matching prefix compared to the provided address, or nil if no matching address
func (trie *MACAddressAssociativeTrie) LongestPrefixMatchNode(addr *MACAddress) *MACAddressAssociativeTrieNode {
	return trie.longestPrefixMatchNode(addr.ToAddressBase()).ToMACAssociative()
}

// ElementContains checks if a prefix block subnet or address in the trie contains the given subnet or address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the subnet or address is contained by a trie element, false otherwise.
//
// To get all the containing addresses, use ElementsContaining
func (trie *MACAddressAssociativeTrie) ElementContains(addr *MACAddress) bool {
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
func (trie *MACAddressAssociativeTrie) GetNode(addr *MACAddress) *MACAddressAssociativeTrieNode {
	return trie.getNode(addr.ToAddressBase()).ToMACAssociative()
}

// GetAddedNode gets trie nodes representing added elements.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Use Contains to check for the existence of a given address in the trie,
// as well as GetNode to search for all nodes including those not added but also auto-generated nodes for subnet blocks.
func (trie *MACAddressAssociativeTrie) GetAddedNode(addr *MACAddress) *MACAddressAssociativeTrieNode {
	return trie.getAddedNode(addr.ToAddressBase()).ToMACAssociative()
}

// NodeIterator returns an iterator that iterates through all the added nodes in the trie in forward or reverse tree order.
func (trie *MACAddressAssociativeTrie) NodeIterator(forward bool) MACAssociativeTrieNodeIteratorRem {
	return macAssociativeTrieNodeIteratorRem{trie.toBase().nodeIterator(forward)}
}

// AllNodeIterator returns an iterator that iterates the added all nodes in the trie following the natural trie order
func (trie *MACAddressAssociativeTrie) AllNodeIterator(forward bool) MACAssociativeTrieNodeIteratorRem {
	return macAssociativeTrieNodeIteratorRem{trie.toBase().allNodeIterator(forward)}
}

// BlockSizeNodeIterator returns an iterator that iterates the added nodes in the trie, ordered by keys from the largest prefix blocks to the smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *MACAddressAssociativeTrie) BlockSizeNodeIterator(lowerSubNodeFirst bool) MACAssociativeTrieNodeIteratorRem {
	return macAssociativeTrieNodeIteratorRem{trie.toBase().blockSizeNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeAllNodeIterator returns an iterator that iterates all nodes in the trie, ordered by keys from the largest prefix blocks to the smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *MACAddressAssociativeTrie) BlockSizeAllNodeIterator(lowerSubNodeFirst bool) MACAssociativeTrieNodeIteratorRem {
	return macAssociativeTrieNodeIteratorRem{trie.toBase().blockSizeAllNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeCachingAllNodeIterator returns an iterator that iterates all nodes, ordered by keys from the largest prefix blocks to the smallest, and then to individual addresses.
func (trie *MACAddressAssociativeTrie) BlockSizeCachingAllNodeIterator() CachingMACAssociativeTrieNodeIterator {
	return cachingMACAssociativeTrieNodeIterator{trie.toBase().blockSizeCachingAllNodeIterator()}
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
func (trie *MACAddressAssociativeTrie) ContainingFirstIterator(forwardSubNodeOrder bool) CachingMACAssociativeTrieNodeIterator {
	return cachingMACAssociativeTrieNodeIterator{trie.toBase().containingFirstIterator(forwardSubNodeOrder)}
}

// ContainingFirstAllNodeIterator returns an iterator that does a pre-order binary tree traversal.
// All nodes will be visited before their sub-nodes.
// For an address trie this means containing subnet blocks will be visited before their contained addresses and subnet blocks.
//
// Once a given node is visited, the iterator allows you to cache an object corresponding to the
// lower or upper sub-node that can be retrieved when you later visit that sub-node.
// That allows you to provide iteration context from a parent to its sub-nodes when iterating.
// The caching and retrieval is done in constant-time and linear space (proportional to tree size).
func (trie *MACAddressAssociativeTrie) ContainingFirstAllNodeIterator(forwardSubNodeOrder bool) CachingMACAssociativeTrieNodeIterator {
	return cachingMACAssociativeTrieNodeIterator{trie.toBase().containingFirstAllNodeIterator(forwardSubNodeOrder)}
}

// ContainedFirstIterator returns an iterator that does a post-order binary tree traversal of the added nodes.
// All added sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (trie *MACAddressAssociativeTrie) ContainedFirstIterator(forwardSubNodeOrder bool) MACAssociativeTrieNodeIteratorRem {
	return macAssociativeTrieNodeIteratorRem{trie.toBase().containedFirstIterator(forwardSubNodeOrder)}
}

// ContainedFirstAllNodeIterator returns an iterator that does a post-order binary tree traversal.
// All sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (trie *MACAddressAssociativeTrie) ContainedFirstAllNodeIterator(forwardSubNodeOrder bool) MACAssociativeTrieNodeIterator {
	return macAssociativeTrieNodeIterator{trie.toBase().containedFirstAllNodeIterator(forwardSubNodeOrder)}
}

// FirstNode returns the first (lowest-valued) node in the trie or nil if the trie is empty
func (trie *MACAddressAssociativeTrie) FirstNode() *MACAddressAssociativeTrieNode {
	return toMACAAssociativeAddressTrieNode(trie.trie.FirstNode())
}

// FirstAddedNode returns the first (lowest-valued) added node in the trie
// or nil if there are no added entries in this tree
func (trie *MACAddressAssociativeTrie) FirstAddedNode() *MACAddressAssociativeTrieNode {
	return toMACAAssociativeAddressTrieNode(trie.trie.FirstAddedNode())
}

// LastNode returns the last (highest-valued) node in the trie or nil if the trie is empty
func (trie *MACAddressAssociativeTrie) LastNode() *MACAddressAssociativeTrieNode {
	return toMACAAssociativeAddressTrieNode(trie.trie.LastNode())
}

// LastAddedNode returns the last (highest-valued) added node in the sub-trie originating from this node,
// or nil if there are no added entries in this tree
func (trie *MACAddressAssociativeTrie) LastAddedNode() *MACAddressAssociativeTrieNode {
	return toMACAAssociativeAddressTrieNode(trie.trie.LastAddedNode())
}

// LowerAddedNode returns the added node whose address is the highest address strictly less than the given address,
// or nil if there are no added entries in this tree
func (trie *MACAddressAssociativeTrie) LowerAddedNode(addr *MACAddress) *MACAddressAssociativeTrieNode {
	return trie.lowerAddedNode(addr.ToAddressBase()).ToMACAssociative()
}

// FloorAddedNode returns the added node whose address is the highest address less than or equal to the given address,
// or nil if there are no added entries in this tree
func (trie *MACAddressAssociativeTrie) FloorAddedNode(addr *MACAddress) *MACAddressAssociativeTrieNode {
	return trie.floorAddedNode(addr.ToAddressBase()).ToMACAssociative()
}

// HigherAddedNode returns the added node whose address is the lowest address strictly greater than the given address,
// or nil if there are no added entries in this tree
func (trie *MACAddressAssociativeTrie) HigherAddedNode(addr *MACAddress) *MACAddressAssociativeTrieNode {
	return trie.higherAddedNode(addr.ToAddressBase()).ToMACAssociative()
}

// CeilingAddedNode returns the added node whose address is the lowest address greater than or equal to the given address,
// or nil if there are no added entries in this tree
func (trie *MACAddressAssociativeTrie) CeilingAddedNode(addr *MACAddress) *MACAddressAssociativeTrieNode {
	return trie.ceilingAddedNode(addr.ToAddressBase()).ToMACAssociative()
}

// Clone clones this trie
func (trie *MACAddressAssociativeTrie) Clone() *MACAddressAssociativeTrie {
	return trie.toBase().clone().ToMACAssociative()
}

// Equal returns whether the given argument is a trie with a set of nodes with the same keys as in this trie
func (trie *MACAddressAssociativeTrie) Equal(other *MACAddressAssociativeTrie) bool {
	return trie.toTrie().Equal(other.toTrie())
}

// DeepEqual returns whether the given argument is a trie with a set of nodes with the same keys and values as in this trie,
// the values being compared with reflect.DeepEqual
func (trie *MACAddressAssociativeTrie) DeepEqual(other *MACAddressAssociativeTrie) bool {
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
func (trie *MACAddressAssociativeTrie) Put(addr *MACAddress, value NodeValue) (bool, NodeValue) {
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
func (trie *MACAddressAssociativeTrie) PutTrie(added *MACAddressAssociativeTrieNode) *MACAddressAssociativeTrieNode {
	return trie.putTrie(added.toBase()).ToMAC()
}

// PutNode associates the specified value with the specified key in this map.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the node for the added address, whether it was already in the tree or not.
//
// If you wish to know whether the node was already there when adding, use PutNew, or before adding you can use GetAddedNode.
func (trie *MACAddressAssociativeTrie) PutNode(addr *MACAddress, value NodeValue) *MACAddressAssociativeTrieNode {
	return trie.putNode(addr.ToAddressBase(), value).ToMAC()
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
func (trie *MACAddressAssociativeTrie) Remap(addr *MACAddress, remapper func(NodeValue) NodeValue) *MACAddressAssociativeTrieNode {
	return trie.remap(addr.ToAddressBase(), remapper).ToMAC()
}

// RemapIfAbsent remaps node values in the trie, but only for nodes that do not exist or are mapped to nil.
//
// This will look up the node corresponding to the given key.
// If the node is not found or mapped to nil, this will call the remapping function.
//
// If the remapping function returns a non-nil value, then it will either set the existing node to have that value,
// or if there was no matched node, it will create a new node with that value.
// If the remapping function returns nil, then it will do the same if insertNull is true, otherwise it will do nothing.
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
func (trie *MACAddressAssociativeTrie) RemapIfAbsent(addr *MACAddress, supplier func() NodeValue, insertNil bool) *MACAddressAssociativeTrieNode {
	return trie.remapIfAbsent(addr.ToAddressBase(), supplier, insertNil).ToMAC()
}

// Get gets the specified value for the specified key in this mapped trie or sub-trie.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the value for the given key.
// Returns nil if the contains no mapping for that key or if the mapped value is nil.
func (trie *MACAddressAssociativeTrie) Get(addr *MACAddress) NodeValue {
	return trie.get(addr.ToAddressBase())
}

// Format implements the fmt.Formatter interface
func (trie MACAddressAssociativeTrie) Format(state fmt.State, verb rune) {
	trie.ToBase().Format(state, verb)
}

// NewMACAddressAssociativeTrie constructs a MAC associative address trie with the root as the zero-prefix prefix block
func NewMACAddressAssociativeTrie(extended bool) *MACAddressAssociativeTrie {
	var rootAddr *Address
	if extended {
		rootAddr = macAllExtended.ToAddressBase()
	} else {
		rootAddr = macAll.ToAddressBase()
	}
	return &MACAddressAssociativeTrie{
		associativeAddressTrie{
			addressTrie{tree.NewBinTrie(&addressTrieKey{rootAddr})},
		},
	}
}
