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

func toMACAddressTrieNode(node *tree.BinTrieNode) *MACAddressTrieNode {
	return (*MACAddressTrieNode)(unsafe.Pointer(node))
}

func toMACAAssociativeAddressTrieNode(node *tree.BinTrieNode) *MACAddressAssociativeTrieNode {
	return (*MACAddressAssociativeTrieNode)(unsafe.Pointer(node))
}

// MACAddressTrieNode represents a node in a MACAddressTrie.
//
// Trie nodes are created by tries during add operations.
//
// If a trie node is copied, a panic will result when methods that alter the trie are called on the copied node.
//
// Iterator methods allow for traversal of the sub-trie with this node as the root.
//
// If an iterator is advanced following a trie modification that followed the creation of the iterator, the iterator will panic.
type MACAddressTrieNode struct {
	addressTrieNode
}

func (node *MACAddressTrieNode) toTrieNode() *tree.BinTrieNode {
	return (*tree.BinTrieNode)(unsafe.Pointer(node))
}

// toBase is used to convert the pointer rather than doing a field dereference, so that nil pointer handling can be done in *addressTrieNode
func (node *MACAddressTrieNode) toBase() *addressTrieNode {
	return (*addressTrieNode)(unsafe.Pointer(node))
}

// ToBase converts to the polymorphic base representation of this MAC trie node.
// The node is unchanged, the returned node is the same underlying node.
func (node *MACAddressTrieNode) ToBase() *AddressTrieNode {
	return (*AddressTrieNode)(node)
}

// ToAssociative converts to the associative trie node representation of this MAC trie node.
// The node is unchanged, the returned node is the same underlying node.
func (node *MACAddressTrieNode) ToAssociative() *MACAddressAssociativeTrieNode {
	return (*MACAddressAssociativeTrieNode)(node)
}

// GetKey gets the key used for placing the node in the tree.
func (node *MACAddressTrieNode) GetKey() *MACAddress {
	return node.toBase().getKey().ToMAC()
}

// IsRoot returns whether this is the root of the backing tree.
func (node *MACAddressTrieNode) IsRoot() bool {
	return node.toTrieNode().IsRoot()
}

// IsAdded returns whether the node was "added".
// Some binary tree nodes are considered "added" and others are not.
// Those nodes created for key elements added to the tree are "added" nodes.
// Those that are not added are those nodes created to serve as junctions for the added nodes.
// Only added elements contribute to the size of a tree.
// When removing nodes, non-added nodes are removed automatically whenever they are no longer needed,
// which is when an added node has less than two added sub-nodes.
func (node *MACAddressTrieNode) IsAdded() bool {
	return node.toTrieNode().IsAdded()
}

// SetAdded makes this node an added node, which is equivalent to adding the corresponding key to the tree.
// If the node is already an added node, this method has no effect.
// You cannot set an added node to non-added, for that you should Remove the node from the tree by calling Remove.
// A non-added node will only remain in the tree if it needs to in the tree.
func (node *MACAddressTrieNode) SetAdded() {
	node.toTrieNode().SetAdded()
}

// Clear removes this node and all sub-nodes from the tree, after which isEmpty() will return true.
func (node *MACAddressTrieNode) Clear() {
	node.toTrieNode().Clear()
}

// IsLeaf returns whether this node is in the tree (a node for which IsAdded() is true)
// and there are no elements in the sub-trie with this node as the root.
func (node *MACAddressTrieNode) IsLeaf() bool {
	return node.toTrieNode().IsLeaf()
}

// GetUpperSubNode gets the direct child node whose key is largest in value
func (node *MACAddressTrieNode) GetUpperSubNode() *MACAddressTrieNode {
	return toMACAddressTrieNode(node.toTrieNode().GetUpperSubNode())
}

// GetLowerSubNode gets the direct child node whose key is smallest in value
func (node *MACAddressTrieNode) GetLowerSubNode() *MACAddressTrieNode {
	return toMACAddressTrieNode(node.toTrieNode().GetLowerSubNode())
}

// GetParent gets the node from which this node is a direct child node, or nil if this is the root.
func (node *MACAddressTrieNode) GetParent() *MACAddressTrieNode {
	return toMACAddressTrieNode(node.toTrieNode().GetParent())
}

// PreviousAddedNode returns the first added node that precedes this node following the tree order
func (node *MACAddressTrieNode) PreviousAddedNode() *MACAddressTrieNode {
	return toMACAddressTrieNode(node.toTrieNode().PreviousAddedNode())
}

// NextAddedNode returns the first added node that follows this node following the tree order
func (node *MACAddressTrieNode) NextAddedNode() *MACAddressTrieNode {
	return toMACAddressTrieNode(node.toTrieNode().NextAddedNode())
}

// NextNode returns the node that follows this node following the tree order
func (node *MACAddressTrieNode) NextNode() *MACAddressTrieNode {
	return toMACAddressTrieNode(node.toTrieNode().NextNode())
}

// PreviousNode returns the node that precedes this node following the tree order
func (node *MACAddressTrieNode) PreviousNode() *MACAddressTrieNode {
	return toMACAddressTrieNode(node.toTrieNode().PreviousNode())
}

// FirstNode returns the first (the lowest valued) node in the sub-trie originating from this node
func (node *MACAddressTrieNode) FirstNode() *MACAddressTrieNode {
	return toMACAddressTrieNode(node.toTrieNode().FirstNode())
}

// FirstAddedNode returns the first (the lowest valued) added node in the sub-trie originating from this node
// or nil if there are no added entries in this tree or sub-trie
func (node *MACAddressTrieNode) FirstAddedNode() *MACAddressTrieNode {
	return toMACAddressTrieNode(node.toTrieNode().FirstAddedNode())
}

// LastNode returns the last (the highest valued) node in the sub-trie originating from this node
func (node *MACAddressTrieNode) LastNode() *MACAddressTrieNode {
	return toMACAddressTrieNode(node.toTrieNode().LastNode())
}

// LastAddedNode returns the last (the highest valued) added node in the sub-trie originating from this node,
// or nil if there are no added entries in this tree or sub-trie
func (node *MACAddressTrieNode) LastAddedNode() *MACAddressTrieNode {
	return toMACAddressTrieNode(node.toTrieNode().LastAddedNode())
}

// LowerAddedNode returns the added node, in this sub-trie with this node as the root, whose address is the highest address strictly less than the given address.
func (node *MACAddressTrieNode) LowerAddedNode(addr *Address) *MACAddressTrieNode {
	return node.toBase().lowerAddedNode(addr).ToMAC()
}

// FloorAddedNode returns the added node, in this sub-trie with this node as the root, whose address is the highest address less than or equal to the given address.
func (node *MACAddressTrieNode) FloorAddedNode(addr *Address) *MACAddressTrieNode {
	return node.toBase().floorAddedNode(addr).ToMAC()
}

// HigherAddedNode returns the added node, in this sub-trie with this node as the root, whose address is the lowest address strictly greater than the given address.
func (node *MACAddressTrieNode) HigherAddedNode(addr *Address) *MACAddressTrieNode {
	return node.toBase().higherAddedNode(addr).ToMAC()
}

// CeilingAddedNode returns the added node, in this sub-trie with this node as the root, whose address is the lowest address greater than or equal to the given address.
func (node *MACAddressTrieNode) CeilingAddedNode(addr *Address) *MACAddressTrieNode {
	return node.toBase().ceilingAddedNode(addr).ToMAC()
}

// Iterator returns an iterator that iterates through the elements of the sub-trie with this node as the root.
// The iteration is in sorted element order.
func (node *MACAddressTrieNode) Iterator() MACAddressIterator {
	return macAddressIterator{node.toBase().iterator()}
}

// DescendingIterator returns an iterator that iterates through the elements of the sub-trie with this node as the root.
// The iteration is in reverse sorted element order.
func (node *MACAddressTrieNode) DescendingIterator() MACAddressIterator {
	return macAddressIterator{node.toBase().descendingIterator()}
}

// NodeIterator returns an iterator that iterates through the added nodes of the sub-trie with this node as the root, in forward or reverse tree order.
func (node *MACAddressTrieNode) NodeIterator(forward bool) MACTrieNodeIteratorRem {
	return macTrieNodeIteratorRem{node.toBase().nodeIterator(forward)}
}

// AllNodeIterator returns an iterator that iterates through all the nodes of the sub-trie with this node as the root, in forward or reverse tree order.
func (node *MACAddressTrieNode) AllNodeIterator(forward bool) MACTrieNodeIteratorRem {
	return macTrieNodeIteratorRem{node.toBase().allNodeIterator(forward)}
}

// BlockSizeNodeIterator returns an iterator that iterates the added nodes, ordered by keys from the largest prefix blocks to the smallest and then to individual addresses,
// in the sub-trie with this node as the root.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order is taken.
func (node *MACAddressTrieNode) BlockSizeNodeIterator(lowerSubNodeFirst bool) MACTrieNodeIteratorRem {
	return macTrieNodeIteratorRem{node.toBase().blockSizeNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeAllNodeIterator returns an iterator that iterates all the nodes, ordered by keys from the largest prefix blocks to the smallest and then to individual addresses,
// in the sub-trie with this node as the root.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (node *MACAddressTrieNode) BlockSizeAllNodeIterator(lowerSubNodeFirst bool) MACTrieNodeIteratorRem {
	return macTrieNodeIteratorRem{node.toBase().blockSizeAllNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeCachingAllNodeIterator returns an iterator that iterates all nodes, ordered by keys from the largest prefix blocks to the smallest and then to individual addresses,
// in the sub-trie with this node as the root.
func (node *MACAddressTrieNode) BlockSizeCachingAllNodeIterator() CachingMACTrieNodeIterator {
	return cachingMACTrieNodeIterator{node.toBase().blockSizeCachingAllNodeIterator()}
}

// ContainingFirstIterator returns an iterator that does a pre-order binary tree traversal of the added nodes
// of the sub-trie with this node as the root.
//
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
func (node *MACAddressTrieNode) ContainingFirstIterator(forwardSubNodeOrder bool) CachingMACTrieNodeIterator {
	return cachingMACTrieNodeIterator{node.toBase().containingFirstIterator(forwardSubNodeOrder)}
}

// ContainingFirstAllNodeIterator returns an iterator that does a pre-order binary tree traversal of all the nodes
// of the sub-trie with this node as the root.
//
// All nodes will be visited before their sub-nodes.
// For an address trie this means containing subnet blocks will be visited before their contained addresses and subnet blocks.
//
// Once a given node is visited, the iterator allows you to cache an object corresponding to the
// lower or upper sub-node that can be retrieved when you later visit that sub-node.
// That allows you to provide iteration context from a parent to its sub-nodes when iterating.
// The caching and retrieval is done in constant-time and linear space (proportional to tree size).
func (node *MACAddressTrieNode) ContainingFirstAllNodeIterator(forwardSubNodeOrder bool) CachingMACTrieNodeIterator {
	return cachingMACTrieNodeIterator{node.toBase().containingFirstAllNodeIterator(forwardSubNodeOrder)}
}

// ContainedFirstIterator returns an iterator that does a post-order binary tree traversal of the added nodes
// of the sub-trie with this node as the root.
// All added sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (node *MACAddressTrieNode) ContainedFirstIterator(forwardSubNodeOrder bool) MACTrieNodeIteratorRem {
	return macTrieNodeIteratorRem{node.toBase().containedFirstIterator(forwardSubNodeOrder)}
}

// ContainedFirstAllNodeIterator returns an iterator that does a post-order binary tree traversal of all the nodes
// of the sub-trie with this node as the root.
// All sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (node *MACAddressTrieNode) ContainedFirstAllNodeIterator(forwardSubNodeOrder bool) MACTrieNodeIterator {
	return macTrieNodeIterator{node.toBase().containedFirstAllNodeIterator(forwardSubNodeOrder)}
}

// Clone clones the node.
// Keys remain the same, but the parent node and the lower and upper sub-nodes are all set to nil.
func (node *MACAddressTrieNode) Clone() *MACAddressTrieNode {
	return toMACAddressTrieNode(node.toTrieNode().Clone())
}

// CloneTree clones the sub-trie starting with this node as the root.
// The nodes are cloned, but their keys and values are not cloned.
func (node *MACAddressTrieNode) CloneTree() *MACAddressTrieNode {
	return toMACAddressTrieNode(node.toTrieNode().CloneTree())
}

// AsNewTrie creates a new sub-trie, copying the nodes starting with this node as the root.
// The nodes are copies of the nodes in this sub-trie, but their keys and values are not copies.
func (node *MACAddressTrieNode) AsNewTrie() *MACAddressTrie {
	return toAddressTrie(node.toTrieNode().AsNewTrie()).ToMAC()
}

// Compare returns -1, 0 or 1 if this node is less than, equal, or greater than the other, according to the key and the trie order.
func (node *MACAddressTrieNode) Compare(other *MACAddressTrieNode) int {
	return node.toTrieNode().Compare(other.toTrieNode())
}

// Equal returns whether the key and mapped value match those of the given node
func (node *MACAddressTrieNode) Equal(other *MACAddressTrieNode) bool {
	return node.toTrieNode().Equal(other.toTrieNode())
}

// TreeEqual returns whether the sub-trie represented by this node as the root node matches the given sub-trie
func (node *MACAddressTrieNode) TreeEqual(other *MACAddressTrieNode) bool {
	return node.toTrieNode().TreeEqual(other.toTrieNode())
}

// Remove removes this node from the collection of added nodes, and also from the trie if possible.
// If it has two sub-nodes, it cannot be removed from the trie, in which case it is marked as not "added",
// nor is it counted in the trie size.
// Only added nodes can be removed from the trie.  If this node is not added, this method does nothing.
func (node *MACAddressTrieNode) Remove() {
	node.toTrieNode().Remove()
}

// Contains returns whether the given address or prefix block subnet is in the sub-trie, as an added element, with this node as the root.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the prefix block or address exists already in the trie, false otherwise.
//
// Use GetAddedNode  to get the node for the address rather than just checking for its existence.
func (node *MACAddressTrieNode) Contains(addr *MACAddress) bool {
	return node.toBase().contains(addr.ToAddressBase())
}

// RemoveNode removes the given single address or prefix block subnet from the trie with this node as the root.
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
func (node *MACAddressTrieNode) RemoveNode(addr *MACAddress) bool {
	return node.toBase().removeNode(addr.ToAddressBase())
}

// RemoveElementsContainedBy removes any single address or prefix block subnet from the trie, with this node as the root, that is contained in the given individual address or prefix block subnet.
//
// Goes further than Remove, not requiring a match to an inserted node, and also removing all the sub-nodes of any removed node or sub-node.
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
func (node *MACAddressTrieNode) RemoveElementsContainedBy(addr *MACAddress) *MACAddressTrieNode {
	return node.toBase().removeElementsContainedBy(addr.ToAddressBase()).ToMAC()
}

// ElementsContainedBy checks if a part of this trie, with this node as the root, is contained by the given prefix block subnet or individual address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the root node of the contained sub-trie, or nil if no sub-trie is contained.
// The node returned need not be an "added" node, see IsAdded for more details on added nodes.
// The returned sub-trie is backed by this trie, so changes in this trie are reflected in those nodes and vice-versa.
func (node *MACAddressTrieNode) ElementsContainedBy(addr *MACAddress) *MACAddressTrieNode {
	return node.toBase().elementsContainedBy(addr.ToAddressBase()).ToMAC()
}

// ElementsContaining finds the trie nodes in the trie, with this sub-node as the root,
// containing the given key and returns them as a linked list.
// Only added nodes are added to the linked list
//
// If the argument is not a single address nor prefix block, this method will panic.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (node *MACAddressTrieNode) ElementsContaining(addr *MACAddress) *ContainmentPath {
	return node.toBase().elementsContaining(addr.ToAddressBase())
}

// LongestPrefixMatch returns the address or subnet with the longest prefix of all the added subnets or the address whose prefix matches the given address.
// This is equivalent to finding the containing subnet or address with the smallest subnet size.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns nil if no added subnet or address contains the given argument.
//
// Use ElementContains to check for the existence of a containing address.
// To get all the containing addresses (subnets with matching prefix), use ElementsContaining.
// To get the node corresponding to the result of this method, use LongestPrefixMatchNode.
func (node *MACAddressTrieNode) LongestPrefixMatch(addr *MACAddress) *Address {
	return node.toBase().longestPrefixMatch(addr.ToAddressBase())
}

// LongestPrefixMatchNode finds the containing subnet or address in the trie with the smallest subnet size,
// which is equivalent to finding the subnet or address with the longest matching prefix.
// Returns the node corresponding to that subnet.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns nil if no added subnet or address contains the given argument.
//
// Use ElementContains to check for the existence of a containing address.
// To get all the containing addresses, use ElementsContaining.
// Use LongestPrefixMatch to get only the address corresponding to the result of this method.
func (node *MACAddressTrieNode) LongestPrefixMatchNode(addr *MACAddress) *MACAddressTrieNode {
	return node.toBase().longestPrefixMatchNode(addr.ToAddressBase()).ToMAC()
}

// ElementContains checks if a prefix block subnet or address in the trie, with this node as the root, contains the given subnet or address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the subnet or address is contained by a trie element, false otherwise.
//
// To get all the containing addresses, use ElementsContaining.
func (node *MACAddressTrieNode) ElementContains(addr *MACAddress) bool {
	return node.toBase().elementContains(addr.ToAddressBase())
}

// GetNode gets the node in the trie, with this sub-node as the root, corresponding to the given address,
// or returns nil if not such element exists.
//
// It returns any node, whether added or not,
// including any prefix block node that was not added.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (node *MACAddressTrieNode) GetNode(addr *MACAddress) *MACAddressTrieNode {
	return node.toBase().getNode(addr.ToAddressBase()).ToMAC()
}

// GetAddedNode gets trie nodes representing added elements.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Use Contains to check for the existence of a given address in the trie,
// as well as GetNode to search for all nodes including those not-added but also auto-generated nodes for subnet blocks.
func (node *MACAddressTrieNode) GetAddedNode(addr *MACAddress) *MACAddressTrieNode {
	return node.toBase().getAddedNode(addr.ToAddressBase()).ToMAC()
}

// NodeSize returns the number of nodes in the trie with this node as the root, which is more than the number of added addresses or blocks.
func (node *MACAddressTrieNode) NodeSize() int {
	return node.toTrieNode().NodeSize()
}

// Size returns the number of elements in the tree.
// Only nodes for which IsAdded returns true are counted.
// When zero is returned, IsEmpty returns true.
func (node *MACAddressTrieNode) Size() int {
	return node.toTrieNode().Size()
}

// IsEmpty returns whether the size is 0
func (node *MACAddressTrieNode) IsEmpty() bool {
	return node.Size() == 0
}

// TreeString returns a visual representation of the sub-trie with this node as the root, with one node per line.
//
// withNonAddedKeys: whether to show nodes that are not added nodes
// withSizes: whether to include the counts of added nodes in each sub-trie
func (node *MACAddressTrieNode) TreeString(withNonAddedKeys, withSizes bool) string {
	return node.toTrieNode().TreeString(withNonAddedKeys, withSizes)
}

// String returns a visual representation of this node including the key, with an open circle indicating this node is not an added node,
// a closed circle indicating this node is an added node.
func (node *MACAddressTrieNode) String() string {
	return node.toTrieNode().String()
}

// For some reason Format must be here and not in addressTrieNode for nil node.
// It panics in fmt code either way, but if in here then it is handled by a recover() call in fmt properly in the debugger.

// Format implements the fmt.Formatter interface
func (node MACAddressTrieNode) Format(state fmt.State, verb rune) {
	node.toTrieNode().Format(state, verb)
}

//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//

// MACAddressAssociativeTrieNode is a node in a MACAddressAssociativeTrie.
//
// In an associative trie, each key or node can be associated with a value.
//
// Trie nodes are created by tries during add or put operations.
//
// If a trie node is copied, a panic will result when methods that alter the trie are called on the copied node.
//
// Iterator methods allow for traversal of the sub-trie with this node as the root.
//
// If an iterator is advanced following a trie modification that followed the creation of the iterator, the iterator will panic.
type MACAddressAssociativeTrieNode struct {
	addressTrieNode
}

func (node *MACAddressAssociativeTrieNode) toTrieNode() *tree.BinTrieNode {
	return (*tree.BinTrieNode)(unsafe.Pointer(node))
}

// toBase is used to convert the pointer rather than doing a field dereference, so that nil pointer handling can be done in *addressTrieNode
func (node *MACAddressAssociativeTrieNode) toBase() *addressTrieNode {
	return (*addressTrieNode)(unsafe.Pointer(node))
}

// ToBase converts to the polymorphic non-associative representation of this trie node
// The node is unchanged, the returned node is the same underlying node.
func (node *MACAddressAssociativeTrieNode) ToBase() *AddressTrieNode {
	return (*AddressTrieNode)(node)
}

// ToMACBase converts to the non-associative representation of this MAC trie node.
// The node is unchanged, the returned node is the same underlying node.
func (node *MACAddressAssociativeTrieNode) ToMACBase() *MACAddressTrieNode {
	return (*MACAddressTrieNode)(node)
}

// ToAssociativeBase converts to the polymorphic associative representation of this trie node
// The node is unchanged, the returned node is the same underlying node.
func (node *MACAddressAssociativeTrieNode) ToAssociativeBase() *AssociativeAddressTrieNode {
	return (*AssociativeAddressTrieNode)(node)
}

// GetKey gets the key used for placing the node in the tree.
func (node *MACAddressAssociativeTrieNode) GetKey() *MACAddress {
	return node.toBase().getKey().ToMAC()
}

// IsRoot returns whether this is the root of the backing tree.
func (node *MACAddressAssociativeTrieNode) IsRoot() bool {
	return node.toTrieNode().IsRoot()
}

// IsAdded returns whether the node was "added".
// Some binary tree nodes are considered "added" and others are not.
// Those nodes created for key elements added to the tree are "added" nodes.
// Those that are not added are those nodes created to serve as junctions for the added nodes.
// Only added elements contribute to the size of a tree.
// When removing nodes, non-added nodes are removed automatically whenever they are no longer needed,
// which is when an added node has less than two added sub-nodes.
func (node *MACAddressAssociativeTrieNode) IsAdded() bool {
	return node.toTrieNode().IsAdded()
}

// SetAdded makes this node an added node, which is equivalent to adding the corresponding key to the tree.
// If the node is already an added node, this method has no effect.
// You cannot set an added node to non-added, for that you should Remove the node from the tree by calling Remove.
// A non-added node will only remain in the tree if it needs to in the tree.
func (node *MACAddressAssociativeTrieNode) SetAdded() {
	node.toTrieNode().SetAdded()
}

// Clear removes this node and all sub-nodes from the tree, after which isEmpty() will return true.
func (node *MACAddressAssociativeTrieNode) Clear() {
	node.toTrieNode().Clear()
}

// IsLeaf returns whether this node is in the tree (a node for which IsAdded() is true)
// and there are no elements in the sub-trie with this node as the root.
func (node *MACAddressAssociativeTrieNode) IsLeaf() bool {
	return node.toTrieNode().IsLeaf()
}

// ClearValue makes the value associated with this node the nil value
func (node *MACAddressAssociativeTrieNode) ClearValue() {
	node.toTrieNode().ClearValue()
}

// SetValue sets the value associated with this node
func (node *MACAddressAssociativeTrieNode) SetValue(val NodeValue) {
	node.toTrieNode().SetValue(val)
}

// GetValue sets the value associated with this node
func (node *MACAddressAssociativeTrieNode) GetValue() NodeValue {
	return node.toTrieNode().GetValue()
}

// GetUpperSubNode gets the direct child node whose key is largest in value
func (node *MACAddressAssociativeTrieNode) GetUpperSubNode() *MACAddressAssociativeTrieNode {
	return toMACAAssociativeAddressTrieNode(node.toTrieNode().GetUpperSubNode())
}

// GetLowerSubNode gets the direct child node whose key is smallest in value
func (node *MACAddressAssociativeTrieNode) GetLowerSubNode() *MACAddressAssociativeTrieNode {
	return toMACAAssociativeAddressTrieNode(node.toTrieNode().GetLowerSubNode())
}

// GetParent gets the node from which this node is a direct child node, or nil if this is the root.
func (node *MACAddressAssociativeTrieNode) GetParent() *MACAddressAssociativeTrieNode {
	return toMACAAssociativeAddressTrieNode(node.toTrieNode().GetParent())
}

// PreviousAddedNode returns the previous node in the tree that is an added node, following the tree order in reverse,
// or nil if there is no such node.
func (node *MACAddressAssociativeTrieNode) PreviousAddedNode() *MACAddressAssociativeTrieNode {
	return toMACAAssociativeAddressTrieNode(node.toTrieNode().PreviousAddedNode())
}

// NextAddedNode returns the next node in the tree that is an added node, following the tree order,
// or nil if there is no such node.
func (node *MACAddressAssociativeTrieNode) NextAddedNode() *MACAddressAssociativeTrieNode {
	return toMACAAssociativeAddressTrieNode(node.toTrieNode().NextAddedNode())
}

// NextNode returns the node that follows this node following the tree order
func (node *MACAddressAssociativeTrieNode) NextNode() *MACAddressAssociativeTrieNode {
	return toMACAAssociativeAddressTrieNode(node.toTrieNode().NextNode())
}

// PreviousNode returns the node that precedes this node following the tree order.
func (node *MACAddressAssociativeTrieNode) PreviousNode() *MACAddressAssociativeTrieNode {
	return toMACAAssociativeAddressTrieNode(node.toTrieNode().PreviousNode())
}

// FirstNode returns the first (the lowest valued) node in the sub-trie originating from this node.
func (node *MACAddressAssociativeTrieNode) FirstNode() *MACAddressAssociativeTrieNode {
	return toMACAAssociativeAddressTrieNode(node.toTrieNode().FirstNode())
}

// FirstAddedNode returns the first (the lowest valued) added node in the sub-trie originating from this node,
// or nil if there are no added entries in this tree or sub-trie
func (node *MACAddressAssociativeTrieNode) FirstAddedNode() *MACAddressAssociativeTrieNode {
	return toMACAAssociativeAddressTrieNode(node.toTrieNode().FirstAddedNode())
}

// LastNode returns the last (the highest valued) node in the sub-trie originating from this node.
func (node *MACAddressAssociativeTrieNode) LastNode() *MACAddressAssociativeTrieNode {
	return toMACAAssociativeAddressTrieNode(node.toTrieNode().LastNode())
}

// LastAddedNode returns the last (the highest valued) added node in the sub-trie originating from this node,
// or nil if there are no added entries in this tree or sub-trie
func (node *MACAddressAssociativeTrieNode) LastAddedNode() *MACAddressAssociativeTrieNode {
	return toMACAAssociativeAddressTrieNode(node.toTrieNode().LastAddedNode())
}

// LowerAddedNode returns the added node, in this sub-trie with this node as the root, whose address is the highest address strictly less than the given address.
func (node *MACAddressAssociativeTrieNode) LowerAddedNode(addr *Address) *MACAddressAssociativeTrieNode {
	return node.toBase().lowerAddedNode(addr).ToMACAssociative()
}

// FloorAddedNode returns the added node, in this sub-trie with this node as the root, whose address is the highest address less than or equal to the given address.
func (node *MACAddressAssociativeTrieNode) FloorAddedNode(addr *Address) *MACAddressAssociativeTrieNode {
	return node.toBase().floorAddedNode(addr).ToMACAssociative()
}

// HigherAddedNode returns the added node, in this sub-trie with this node as the root, whose address is the lowest address strictly greater than the given address.
func (node *MACAddressAssociativeTrieNode) HigherAddedNode(addr *Address) *MACAddressAssociativeTrieNode {
	return node.toBase().higherAddedNode(addr).ToMACAssociative()
}

// CeilingAddedNode returns the added node, in this sub-trie with this node as the root, whose address is the lowest address greater than or equal to the given address.
func (node *MACAddressAssociativeTrieNode) CeilingAddedNode(addr *Address) *MACAddressAssociativeTrieNode {
	return node.toBase().ceilingAddedNode(addr).ToMACAssociative()
}

// Iterator returns an iterator that iterates through the elements of the sub-trie with this node as the root.
// The iteration is in sorted element order.
func (node *MACAddressAssociativeTrieNode) Iterator() MACAddressIterator {
	return macAddressIterator{node.toBase().iterator()}
}

// DescendingIterator returns an iterator that iterates through the elements of the sub-trie with this node as the root.
// The iteration is in reverse sorted element order.
func (node *MACAddressAssociativeTrieNode) DescendingIterator() MACAddressIterator {
	return macAddressIterator{node.toBase().descendingIterator()}
}

// NodeIterator returns an iterator that iterates through the added nodes of the sub-trie with this node as the root, in forward or reverse tree order.
func (node *MACAddressAssociativeTrieNode) NodeIterator(forward bool) MACAssociativeTrieNodeIteratorRem {
	return macAssociativeTrieNodeIteratorRem{node.toBase().nodeIterator(forward)}
}

// AllNodeIterator returns an iterator that iterates through all the nodes of the sub-trie with this node as the root, in forward or reverse tree order.
func (node *MACAddressAssociativeTrieNode) AllNodeIterator(forward bool) MACAssociativeTrieNodeIteratorRem {
	return macAssociativeTrieNodeIteratorRem{node.toBase().allNodeIterator(forward)}
}

// BlockSizeNodeIterator returns an iterator that iterates the added nodes, ordered by keys from the largest prefix blocks to the smallest and then to individual addresses,
// in the sub-trie with this node as the root.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order is taken.
func (node *MACAddressAssociativeTrieNode) BlockSizeNodeIterator(lowerSubNodeFirst bool) MACAssociativeTrieNodeIteratorRem {
	return macAssociativeTrieNodeIteratorRem{node.toBase().blockSizeNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeAllNodeIterator returns an iterator that iterates all the nodes, ordered by keys from the largest prefix blocks to the smallest and then to individual addresses,
// in the sub-trie with this node as the root.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (node *MACAddressAssociativeTrieNode) BlockSizeAllNodeIterator(lowerSubNodeFirst bool) MACAssociativeTrieNodeIteratorRem {
	return macAssociativeTrieNodeIteratorRem{node.toBase().blockSizeAllNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeCachingAllNodeIterator returns an iterator that iterates all nodes, ordered by keys from the largest prefix blocks to the smallest and then to individual addresses,
// in the sub-trie with this node as the root.
func (node *MACAddressAssociativeTrieNode) BlockSizeCachingAllNodeIterator() CachingMACAssociativeTrieNodeIterator {
	return cachingMACAssociativeTrieNodeIterator{node.toBase().blockSizeCachingAllNodeIterator()}
}

// ContainingFirstIterator returns an iterator that does a pre-order binary tree traversal of the added nodes
// of the sub-trie with this node as the root.
//
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
func (node *MACAddressAssociativeTrieNode) ContainingFirstIterator(forwardSubNodeOrder bool) CachingMACAssociativeTrieNodeIterator {
	return cachingMACAssociativeTrieNodeIterator{node.toBase().containingFirstIterator(forwardSubNodeOrder)}
}

// ContainingFirstAllNodeIterator returns an iterator that does a pre-order binary tree traversal of all the nodes
// of the sub-trie with this node as the root.
//
// All nodes will be visited before their sub-nodes.
// For an address trie this means containing subnet blocks will be visited before their contained addresses and subnet blocks.
//
// Once a given node is visited, the iterator allows you to cache an object corresponding to the
// lower or upper sub-node that can be retrieved when you later visit that sub-node.
// That allows you to provide iteration context from a parent to its sub-nodes when iterating.
// The caching and retrieval is done in constant-time and linear space (proportional to tree size).
func (node *MACAddressAssociativeTrieNode) ContainingFirstAllNodeIterator(forwardSubNodeOrder bool) CachingMACAssociativeTrieNodeIterator {
	return cachingMACAssociativeTrieNodeIterator{node.toBase().containingFirstAllNodeIterator(forwardSubNodeOrder)}
}

// ContainedFirstIterator returns an iterator that does a post-order binary tree traversal of the added nodes
// of the sub-trie with this node as the root.
// All added sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (node *MACAddressAssociativeTrieNode) ContainedFirstIterator(forwardSubNodeOrder bool) MACAssociativeTrieNodeIteratorRem {
	return macAssociativeTrieNodeIteratorRem{node.toBase().containedFirstIterator(forwardSubNodeOrder)}
}

// ContainedFirstAllNodeIterator returns an iterator that does a post-order binary tree traversal of all the nodes
// of the sub-trie with this node as the root.
// All sub-nodes will be visited before their parent nodes.
// For an address trie this means contained addresses and subnets will be visited before their containing subnet blocks.
func (node *MACAddressAssociativeTrieNode) ContainedFirstAllNodeIterator(forwardSubNodeOrder bool) MACAssociativeTrieNodeIterator {
	return macAssociativeTrieNodeIterator{node.toBase().containedFirstAllNodeIterator(forwardSubNodeOrder)}
}

// Clone clones the node.
// Keys remain the same, but the parent node and the lower and upper sub-nodes are all set to nil.
func (node *MACAddressAssociativeTrieNode) Clone() *MACAddressAssociativeTrieNode {
	return toMACAAssociativeAddressTrieNode(node.toTrieNode().Clone())
}

// CloneTree clones the sub-trie starting with this node as the root.
// The nodes are cloned, but their keys and values are not cloned.
func (node *MACAddressAssociativeTrieNode) CloneTree() *MACAddressAssociativeTrieNode {
	return toMACAAssociativeAddressTrieNode(node.toTrieNode().CloneTree())
}

// AsNewTrie creates a new sub-trie, copying the nodes starting with this node as the root.
// The nodes are copies of the nodes in this sub-trie, but their keys and values are not copies.
func (node *MACAddressAssociativeTrieNode) AsNewTrie() *MACAddressAssociativeTrie {
	return toAddressTrie(node.toTrieNode().AsNewTrie()).ToMACAssociative()
}

// Compare returns -1, 0 or 1 if this node is less than, equal, or greater than the other, according to the key and the trie order.
func (node *MACAddressAssociativeTrieNode) Compare(other *MACAddressAssociativeTrieNode) int {
	return node.toTrieNode().Compare(other.toTrieNode())
}

// Equal returns whether the key and mapped value match those of the given node
func (node *MACAddressAssociativeTrieNode) Equal(other *MACAddressAssociativeTrieNode) bool {
	return node.toTrieNode().Equal(other.toTrieNode())
}

// TreeEqual returns whether the sub-trie represented by this node as the root node matches the given sub-trie
func (node *MACAddressAssociativeTrieNode) TreeEqual(other *MACAddressAssociativeTrieNode) bool {
	return node.toTrieNode().TreeEqual(other.toTrieNode())
}

// DeepEqual returns whether the key is equal to that of the given node and the value is deep equal to that of the given node
func (node *MACAddressAssociativeTrieNode) DeepEqual(other *MACAddressAssociativeTrieNode) bool {
	return node.toTrieNode().DeepEqual(other.toTrieNode())
}

// TreeDeepEqual returns whether the sub-trie represented by this node as the root node matches the given sub-trie, matching with Compare on the keys and reflect.DeepEqual on the values
func (node *MACAddressAssociativeTrieNode) TreeDeepEqual(other *MACAddressAssociativeTrieNode) bool {
	return node.toTrieNode().TreeDeepEqual(other.toTrieNode())
}

// Remove removes this node from the collection of added nodes, and also from the trie if possible.
// If it has two sub-nodes, it cannot be removed from the trie, in which case it is marked as not "added",
// nor is it counted in the trie size.
// Only added nodes can be removed from the trie.  If this node is not added, this method does nothing.
func (node *MACAddressAssociativeTrieNode) Remove() {
	node.toTrieNode().Remove()
}

// Contains returns whether the given address or prefix block subnet is in the sub-trie, as an added element, with this node as the root.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the prefix block or address exists already in the trie, false otherwise.
//
// Use GetAddedNode  to get the node for the address rather than just checking for its existence.
func (node *MACAddressAssociativeTrieNode) Contains(addr *MACAddress) bool {
	return node.toBase().contains(addr.ToAddressBase())
}

// RemoveNode removes the given single address or prefix block subnet from the trie with this node as the root.
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
func (node *MACAddressAssociativeTrieNode) RemoveNode(addr *MACAddress) bool {
	return node.toBase().removeNode(addr.ToAddressBase())
}

// RemoveElementsContainedBy removes any single address or prefix block subnet from the trie, with this node as the root, that is contained in the given individual address or prefix block subnet.
//
// Goes further than Remove, not requiring a match to an inserted node, and also removing all the sub-nodes of any removed node or sub-node.
//
// For example, after inserting 1.2.3.0 and 1.2.3.1, passing 1.2.3.0/31 to RemoveElementsContainedBy will remove them both,
// while the Remove method will remove nothing.
// After inserting 1.2.3.0/31, then #remove(Address) will remove 1.2.3.0/31, but will leave 1.2.3.0 and 1.2.3.1 in the trie.
//
// It cannot partially delete a node, such as deleting a single address from a prefix block represented by a node.
// It can only delete the whole node if the whole address or block represented by that node is contained in the given address or block.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
//Returns the root node of the sub-trie that was removed from the trie, or nil if nothing was removed.
func (node *MACAddressAssociativeTrieNode) RemoveElementsContainedBy(addr *MACAddress) *MACAddressAssociativeTrieNode {
	return node.toBase().removeElementsContainedBy(addr.ToAddressBase()).ToMACAssociative()
}

// ElementsContainedBy checks if a part of this trie, with this node as the root, is contained by the given prefix block subnet or individual address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the root node of the contained sub-trie, or nil if no sub-trie is contained.
// The node returned need not be an "added" node, see IsAdded for more details on added nodes.
// The returned sub-trie is backed by this trie, so changes in this trie are reflected in those nodes and vice-versa.
func (node *MACAddressAssociativeTrieNode) ElementsContainedBy(addr *MACAddress) *MACAddressAssociativeTrieNode {
	return node.toBase().elementsContainedBy(addr.ToAddressBase()).ToMACAssociative()
}

// ElementsContaining finds the trie nodes in the trie, with this sub-node as the root,
// containing the given key and returns them as a linked list.
// Only added nodes are added to the linked list
//
// If the argument is not a single address nor prefix block, this method will panic.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (node *MACAddressAssociativeTrieNode) ElementsContaining(addr *MACAddress) *ContainmentPath {
	return node.toBase().elementsContaining(addr.ToAddressBase())
}

// LongestPrefixMatch returns the address or subnet with the longest prefix of all the added subnets or the address whose prefix matches the given address.
// This is equivalent to finding the containing subnet or address with the smallest subnet size.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns nil if no added subnet or address contains the given argument.
//
// Use ElementContains to check for the existence of a containing address.
// To get all the containing addresses (subnets with matching prefix), use ElementsContaining.
// To get the node corresponding to the result of this method, use LongestPrefixMatchNode.
func (node *MACAddressAssociativeTrieNode) LongestPrefixMatch(addr *MACAddress) *Address {
	return node.toBase().longestPrefixMatch(addr.ToAddressBase())
}

// LongestPrefixMatchNode finds the containing subnet or address in the trie with the smallest subnet size,
// which is equivalent to finding the subnet or address with the longest matching prefix.
// Returns the node corresponding to that subnet.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns nil if no added subnet or address contains the given argument.
//
// Use ElementContains to check for the existence of a containing address.
// To get all the containing addresses, use ElementsContaining.
// Use LongestPrefixMatch to get only the address corresponding to the result of this method.
func (node *MACAddressAssociativeTrieNode) LongestPrefixMatchNode(addr *MACAddress) *MACAddressAssociativeTrieNode {
	return node.toBase().longestPrefixMatchNode(addr.ToAddressBase()).ToMACAssociative()
}

// ElementContains checks if a prefix block subnet or address in the trie, with this node as the root, contains the given subnet or address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the subnet or address is contained by a trie element, false otherwise.
//
// To get all the containing addresses, use ElementsContaining.
func (node *MACAddressAssociativeTrieNode) ElementContains(addr *MACAddress) bool {
	return node.toBase().elementContains(addr.ToAddressBase())
}

// GetNode gets the node in the trie, with this sub-node as the root, corresponding to the given address,
// or returns nil if not such element exists.
//
// It returns any node, whether added or not,
// including any prefix block node that was not added.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (node *MACAddressAssociativeTrieNode) GetNode(addr *MACAddress) *MACAddressAssociativeTrieNode {
	return node.toBase().getNode(addr.ToAddressBase()).ToMACAssociative()
}

// GetAddedNode gets trie nodes representing added elements.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Use Contains to check for the existence of a given address in the trie,
// as well as GetNode to search for all nodes including those not-added but also auto-generated nodes for subnet blocks.
func (node *MACAddressAssociativeTrieNode) GetAddedNode(addr *MACAddress) *MACAddressAssociativeTrieNode {
	return node.toBase().getAddedNode(addr.ToAddressBase()).ToMACAssociative()
}

// Get gets the specified value for the specified key in this mapped trie or sub-trie.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the value for the given key.
// Returns nil if the contains no mapping for that key or if the mapped value is nil.
func (node *MACAddressAssociativeTrieNode) Get(addr *MACAddress) NodeValue {
	return node.toBase().get(addr.ToAddressBase())
}

// NodeSize returns the number of nodes in the trie with this node as the root, which is more than the number of added addresses or blocks.
func (node *MACAddressAssociativeTrieNode) NodeSize() int {
	return node.toTrieNode().NodeSize()
}

// Size returns the number of elements in the tree.
// Only nodes for which IsAdded returns true are counted.
// When zero is returned, IsEmpty returns true.
func (node *MACAddressAssociativeTrieNode) Size() int {
	return node.toTrieNode().Size()
}

// IsEmpty returns whether the size is 0
func (node *MACAddressAssociativeTrieNode) IsEmpty() bool {
	return node.Size() == 0
}

// TreeString returns a visual representation of the sub-trie with this node as the root, with one node per line.
//
// withNonAddedKeys: whether to show nodes that are not added nodes
// withSizes: whether to include the counts of added nodes in each sub-trie
func (node *MACAddressAssociativeTrieNode) TreeString(withNonAddedKeys, withSizes bool) string {
	return node.toTrieNode().TreeString(withNonAddedKeys, withSizes)
}

// String returns a visual representation of this node including the key, with an open circle indicating this node is not an added node,
// a closed circle indicating this node is an added node.
func (node *MACAddressAssociativeTrieNode) String() string {
	return node.toTrieNode().String()
}

// For some reason Format must be here and not in addressTrieNode for nil node.
// It panics in fmt code either way, but if in here then it is handled by a recover() call in fmt properly in the debugger.

// Format implements the fmt.Formatter interface
func (node MACAddressAssociativeTrieNode) Format(state fmt.State, verb rune) {
	node.ToBase().Format(state, verb)
}
