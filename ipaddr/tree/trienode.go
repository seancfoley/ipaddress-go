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

package tree

import (
	"fmt"
	"reflect"
	"unsafe"
)

type operation int

const (
	// Given a key E
	insert         operation = iota // add node for E if not already there
	remap                           // alters nodes based on the existing nodes and their values
	lookup                          // find node for E, traversing all containing elements along the way
	near                            // closest match, going down trie to get element considered closest. Whether one thing is closer than another is determined by the sorted order.
	containing                      // list the nodes whose keys contain E
	insertedDelete                  // Remove node for E
	subtreeDelete                   // Remove nodes whose keys are contained by E
)

type opResult struct {
	key TrieKey

	// whether near is searching for a floor or ceiling
	// a floor is greatest element below addr
	// a ceiling is lowest element above addr
	nearestFloor,

	// whether near cannot be an exact match
	nearExclusive bool

	op operation

	// lookups:

	// an inserted tree element matches the supplied argument
	// exists is set to true only for "added" nodes
	exists bool

	// the matching tree element, when doing a lookup operation, or the pre-existing node for an insert operation
	// existingNode is set for both added and not added nodes
	existingNode,

	// the closest tree element, when doing a near operation
	nearestNode,

	// if searching for a floor/lower, and the nearest node is above addr, then we must backtrack to get below
	// if searching for a ceiling/higher, and the nearest node is below addr, then we must backtrack to get above
	backtrackNode,

	// contains:

	// A linked list of the tree elements, from largest to smallest,
	// that contain the supplied argument, and the end of the list
	containing, containingEnd,

	// The tree node with the smallest subnet or address containing the supplied argument
	smallestContaining,

	// contained by:

	// this tree is contained by the supplied argument
	containedBy,

	// deletions:

	// this tree was deleted
	deleted *BinTrieNode

	// adds and puts:

	// new and existing values for add, put and remap operations
	newValue, existingValue interface{}

	// this added tree node was newly created for an add
	inserted,

	// this added tree node previously existed but had not been added yet
	added,

	// this added tree node was already added to the trie
	addedAlready *BinTrieNode

	// remaps:

	remapper func(V) (V, remapAction)
}

func getNextAdded(node *BinTrieNode) *BinTrieNode {
	for node != nil && !node.IsAdded() {
		// Since only one of upper and lower can be populated, whether we start with upper or lower does not matter
		next := node.GetUpperSubNode()
		if next == nil {
			node = node.GetLowerSubNode()
		} else {
			node = next
		}
	}
	return node
}

func (result *opResult) getContaining() *BinTrieNode {
	containing := getNextAdded(result.containing)
	result.containing = containing
	if containing != nil {
		current := containing
		for {
			next := current.GetUpperSubNode()
			var nextAdded *BinTrieNode
			if next == nil {
				next = current.GetLowerSubNode()
				nextAdded = getNextAdded(next)
				if next != nextAdded {
					current.setLower(nextAdded)
				}
			} else {
				nextAdded = getNextAdded(next)
				if next != nextAdded {
					current.setUpper(nextAdded)
				}
			}
			current = nextAdded
			if current == nil {
				break
			}
		}
	}
	return containing
}

// add to the list of tree elements that contain the supplied argument
func (result *opResult) addContaining(containingSub *BinTrieNode) {
	cloned := containingSub.Clone()
	if result.containing == nil {
		result.containing = cloned
	} else {
		if result.containingEnd.Compare(cloned) > 0 {
			result.containingEnd.setLower(cloned)
		} else {
			result.containingEnd.setUpper(cloned)
		}
		result.containingEnd.adjustCount(1)
	}
	result.containingEnd = cloned
}

type TrieKeyIterator interface {
	HasNext

	Next() TrieKey

	// Remove removes the last iterated element from the underlying trie, and returns that element.
	// If there is no such element, it returns nil.
	Remove() TrieKey
}

type trieKeyIterator struct {
	keyIterator
}

func (iter trieKeyIterator) Next() TrieKey {
	next := iter.keyIterator.Next()
	if next == nil {
		return nil
	}
	return next.(TrieKey)
}

func (iter trieKeyIterator) Remove() TrieKey {
	removed := iter.keyIterator.Remove()
	if removed == nil {
		return nil
	}
	return removed.(TrieKey)
}

type TrieNodeIterator interface {
	HasNext

	Next() *BinTrieNode
}

type TrieNodeIteratorRem interface {
	TrieNodeIterator

	// Remove removes the last iterated element from the underlying trie, and returns that element.
	// If there is no such element, it returns nil.
	Remove() *BinTrieNode
}

type trieNodeIteratorRem struct {
	nodeIteratorRem
}

func (iter trieNodeIteratorRem) Next() *BinTrieNode {
	return iter.nodeIteratorRem.Next().toTrieNode()
}

func (iter trieNodeIteratorRem) Remove() *BinTrieNode {
	return iter.nodeIteratorRem.Remove().toTrieNode()
}

type trieNodeIterator struct {
	nodeIterator
}

func (iter trieNodeIterator) Next() *BinTrieNode {
	return iter.nodeIterator.Next().toTrieNode()
}

type CachingTrieNodeIterator interface {
	TrieNodeIteratorRem
	CachingIterator
}

type cachingTrieNodeIterator struct {
	cachingNodeIterator // an interface
}

func (iter *cachingTrieNodeIterator) Next() *BinTrieNode {
	return iter.cachingNodeIterator.Next().toTrieNode()
}

func (iter *cachingTrieNodeIterator) Remove() *BinTrieNode {
	return iter.cachingNodeIterator.Remove().toTrieNode()
}

// KeyCompareResult has callbacks for a key comparison of a new key with a key pre-existing in the trie.
// At most one of the two methods should be called when comparing keys.
// If existing key is shorter, and the new key matches all bits in the existing key, then neither method should be called.
type KeyCompareResult interface {
	// BitsMatch should be called when the existing key is the same size or large as the new key and the new key bits match the exiting key bits.
	BitsMatch()

	// BitsDoNotMatch should be called when at least one bit in the new key does not match the same bit in the existing key.
	BitsDoNotMatch(matchedBits BitCount)
}

// TrieKey represents a key for a trie.
//
// All trie keys represent a sequence of bits.
// The bit count, which is the same for all keys, is the total number of bits in the key.
//
// Some trie keys represent a fixed sequence of bits.  The bits have a single value.
//
// The remaining trie keys have an initial sequence of bits, the prefix, within which the bits are fixed,
// and the remaining bits beyond the prefix are not fixed and represent all potential bit values.
// Such keys represent all values with the same prefix.
//
// When all bits in a given key are fixed, the key has no prefix or prefix length.
//
// When not all bits are fixed, the prefix length is the number of bits in the initial fixed sequence.
// A key with a prefix length is also known as a prefix block.
//
// A key should never change.
// For keys with a prefix length, the prefix length must remain constance, and the prefix bits must remain constant.
// For keys with no prefix length, all the key bits must remain constant.
type TrieKey interface {

	// MatchBits matches the bits in this key to the bits in the given key, starting from the given bit index.
	// Only the remaining bits in the prefix can be compared for either key.
	// If the prefix length of a key is nil, all the remaining bits can be compared.
	//
	// MatchBits returns true on a successful match or mismatch, and false if only a partial match.
	//
	// MatchBits calls BitsMatch in handleMatch when the given key matches all the bits in this key (even if this key has a shorter prefix),
	// or calls BitsDoNotMatch in handleMatch when there is a mismatch of bits, returning true in both cases.
	//
	// If the given key has a shorter prefix length, so not all bits in this key can be compared to the given key,
	// but the bits that can be compared are a match, then that is a partial match.
	// MatchBits calls neither method in handleMatch and returns false in that case.
	MatchBits(key TrieKey, bitIndex BitCount, handleMatch KeyCompareResult) bool

	// Compare returns a negative integer, zero, or a positive integer if this instance is less than, equal, or greater than the give item.
	// When comparing, the first mismatched bit determines the result.
	// If either key is prefixed, you compare only the bits up until the minumum prefix length.
	// If those bits are equal, and both have the same prefix length, they are equal.
	// Otherwise, the next bit in the key with the longer prefix (or no prefix at all) determines the result.
	// If that bit is 1, that key is larger, if it is 0, then smaller.
	Compare(TrieKey) int

	// GetBitCount returns the bit count for the key, which is a fixed value for any and all keys in the trie.
	GetBitCount() BitCount

	// GetPrefixLen returns the prefix length if this key has a prefix length (ie it is a prefix block).
	// It returns nil if not a prefix block.
	GetPrefixLen() PrefixLen

	// IsOneBit returns whether a given bit in the prefix is 1.
	// If the key is a prefix block, the operation is undefined if the bit index falls outside the prefix.
	// This method will never be called with a bit index that exceeds the prefix.
	IsOneBit(bitIndex BitCount) bool

	// ToPrefixBlockLen creates a new key with a prefix of the given length
	ToPrefixBlockLen(prefixLen BitCount) TrieKey

	// GetTrailingBitCount returns the number of trailing ones or zeros in the key.
	// If the key has a prefix length, GetTrailingBitCount is undefined.
	// This method will never be called on a key with a prefix length.
	GetTrailingBitCount(ones bool) BitCount

	// ToMaxLower returns a new key. If this key has a prefix length, it is converted to a key with a 0 as the first bit following the prefix, followed by all ones to the end, and with the prefix length then removed.
	// It returns this same key if it has no prefix length.
	// For instance, if this key is 1010**** with a prefix length of 4, the returned key is 10100111 with no prefix length.
	ToMaxLower() TrieKey

	// ToMinUpper returns a new key. If this key has a prefix length, it is converted to a key with a 1 as the first bit following the prefix, followed by all zeros to the end, and with the prefix length then removed.
	// It returns this same key if it has no prefix length.
	// For instance, if this key is 1010**** with a prefix length of 4, the returned key is 10101000 with no prefix length.
	ToMinUpper() TrieKey
}

type BinTrieNode struct {
	binTreeNode
}

// works with nil
func (node *BinTrieNode) toBinTreeNode() *binTreeNode {
	return (*binTreeNode)(unsafe.Pointer(node))
}

// setKey sets the key used for placing the node in the tree.
// when freezeRoot is true, this is never called (and freezeRoot is always true)
func (node *BinTrieNode) setKey(item TrieKey) {
	node.binTreeNode.setKey(item)
}

// GetKey gets the key used for placing the node in the tree.
func (node *BinTrieNode) GetKey() TrieKey {
	val := node.toBinTreeNode().getKey()
	if val == nil {
		return nil
	}
	return val.(TrieKey)
}

// IsRoot returns whether this is the root of the backing tree.
func (node *BinTrieNode) IsRoot() bool {
	return node.toBinTreeNode().IsRoot()
}

// IsAdded returns whether the node was "added".
// Some binary tree nodes are considered "added" and others are not.
// Those nodes created for key elements added to the tree are "added" nodes.
// Those that are not added are those nodes created to serve as junctions for the added nodes.
// Only added elements contribute to the size of a tree.
// When removing nodes, non-added nodes are removed automatically whenever they are no longer needed,
// which is when an added node has less than two added sub-nodes.
func (node *BinTrieNode) IsAdded() bool {
	return node.toBinTreeNode().IsAdded()
}

// Clear removes this node and all sub-nodes from the tree, after which isEmpty() will return true.
func (node *BinTrieNode) Clear() {
	node.toBinTreeNode().Clear()
}

// IsEmpty returns where there are not any elements in the sub-tree with this node as the root.
func (node *BinTrieNode) IsEmpty() bool {
	return node.toBinTreeNode().IsEmpty()
}

// IsLeaf returns whether this node is in the tree (a node for which IsAdded() is true)
// and there are no elements in the sub-tree with this node as the root.
func (node *BinTrieNode) IsLeaf() bool {
	return node.toBinTreeNode().IsLeaf()
}

func (node *BinTrieNode) GetValue() (val V) {
	return node.toBinTreeNode().GetValue()
}

func (node *BinTrieNode) ClearValue() {
	node.toBinTreeNode().ClearValue()
}

// Remove removes this node from the collection of added nodes,
// and also removes from the tree if possible.
// If it has two sub-nodes, it cannot be removed from the tree, in which case it is marked as not "added",
// nor is it counted in the tree size.
// Only added nodes can be removed from the tree.  If this node is not added, this method does nothing.
func (node *BinTrieNode) Remove() {
	node.toBinTreeNode().Remove()
}

// NodeSize returns the count of all nodes in the tree starting from this node and extending to all sub-nodes.
// Unlike for the Size method, this is not a constant-time operation and must visit all sub-nodes of this node.
func (node *BinTrieNode) NodeSize() int {
	return node.toBinTreeNode().NodeSize()
}

// Size returns the count of nodes added to the sub-tree starting from this node as root and moving downwards to sub-nodes.
// This is a constant-time operation since the size is maintained in each node and adjusted with each add and Remove operation in the sub-tree.
func (node *BinTrieNode) Size() int {
	return node.toBinTreeNode().Size()
}

// TreeString returns a visual representation of the sub-tree with this node as root, with one node per line.
//
// withNonAddedKeys: whether to show nodes that are not added nodes
// withSizes: whether to include the counts of added nodes in each sub-tree
func (node *BinTrieNode) TreeString(withNonAddedKeys, withSizes bool) string {
	return node.toBinTreeNode().TreeString(withNonAddedKeys, withSizes)
}

// Returns a visual representation of this node including the key, with an open circle indicating this node is not an added node,
// a closed circle indicating this node is an added node.
func (node *BinTrieNode) String() string {
	return node.toBinTreeNode().String()
}

func (node *BinTrieNode) setUpper(upper *BinTrieNode) {
	node.binTreeNode.setUpper(&upper.binTreeNode)
}

func (node *BinTrieNode) setLower(lower *BinTrieNode) {
	node.binTreeNode.setLower(&lower.binTreeNode)
}

// GetUpperSubNode gets the direct child node whose key is largest in value
func (node *BinTrieNode) GetUpperSubNode() *BinTrieNode {
	return node.toBinTreeNode().getUpperSubNode().toTrieNode()
}

// GetLowerSubNode gets the direct child node whose key is smallest in value
func (node *BinTrieNode) GetLowerSubNode() *BinTrieNode {
	return node.toBinTreeNode().getLowerSubNode().toTrieNode()
}

// GetParent gets the node from which this node is a direct child node, or nil if this is the root.
func (node *BinTrieNode) GetParent() *BinTrieNode {
	return node.toBinTreeNode().getParent().toTrieNode()
}

func (node *BinTrieNode) Contains(addr TrieKey) bool {
	result := node.doLookup(addr)
	return result.exists
}

func (node *BinTrieNode) RemoveNode(key TrieKey) bool {
	result := &opResult{
		key: key,
		op:  insertedDelete,
	}
	if node != nil {
		node.matchBits(result)
	}
	return result.exists
}

// GetNode gets the node in the trie corresponding to the given address,
// or returns nil if not such element exists.
//
// It returns any node, whether added or not,
// including any prefix block node that was not added.
func (node *BinTrieNode) GetNode(key TrieKey) *BinTrieNode {
	result := node.doLookup(key)
	return result.existingNode
}

// GetAddedNode gets trie nodes representing added elements.
//
// Use Contains to check for the existence of a given address in the trie,
// as well as GetNode to search for all nodes including those not-added but also auto-generated nodes for subnet blocks.
func (node *BinTrieNode) GetAddedNode(key TrieKey) *BinTrieNode {
	if res := node.GetNode(key); res == nil || res.IsAdded() {
		return res
	}
	return nil
}

func (node *BinTrieNode) Get(key TrieKey) V {
	result := &opResult{
		key: key,
		op:  lookup,
	}
	if node != nil {
		node.matchBits(result)
	}
	resultNode := result.existingNode
	if resultNode == nil {
		return nil
	}
	return resultNode.GetValue()
}

func (node *BinTrieNode) RemoveElementsContainedBy(key TrieKey) *BinTrieNode {
	result := &opResult{
		key: key,
		op:  subtreeDelete,
	}
	if node != nil {
		node.matchBits(result)
	}
	return result.deleted
}

func (node *BinTrieNode) ElementsContainedBy(key TrieKey) *BinTrieNode {
	result := node.doLookup(key)
	return result.containedBy
}

// ElementsContaining finds the trie nodes containing the given key and returns them as a linked list
// only added nodes are added to the linked list
func (node *BinTrieNode) ElementsContaining(key TrieKey) *BinTrieNode {
	result := &opResult{
		key: key,
		op:  containing,
	}
	if node != nil {
		node.matchBits(result)
	}
	return result.getContaining()
}

// LongestPrefixMatch finds the longest matching prefix amongst keys added to the trie
func (node *BinTrieNode) LongestPrefixMatch(key TrieKey) TrieKey {
	res := node.LongestPrefixMatchNode(key)
	if res == nil {
		return nil
	}
	return res.GetKey()
}

// LongestPrefixMatchNode finds the node with the longest matching prefix
// only added nodes are added to the linked list
func (node *BinTrieNode) LongestPrefixMatchNode(key TrieKey) *BinTrieNode {
	return node.doLookup(key).smallestContaining
}

func (node *BinTrieNode) ElementContains(key TrieKey) bool {
	return node.LongestPrefixMatch(key) != nil
}

func (node *BinTrieNode) doLookup(key TrieKey) *opResult {
	result := &opResult{
		key: key,
		op:  lookup,
	}
	if node != nil {
		node.matchBits(result)
	}
	return result
}

func (node *BinTrieNode) removeSubtree(result *opResult) {
	result.deleted = node.toTrieNode()
	node.Clear()
}

func (node *BinTrieNode) removeOp(result *opResult) {
	result.deleted = node.toTrieNode()
	node.binTreeNode.Remove()
}

func (node *BinTrieNode) matchBits(result *opResult) {
	node.matchBitsFromIndex(0, result)
}

// traverses the tree, matching bits with prefix block nodes, until we can match no longer,
// at which point it completes the operation, whatever that operation is
func (node *BinTrieNode) matchBitsFromIndex(bitIndex int, result *opResult) {
	matchNode := node
	for {
		//fmt.Printf("matching %v to %v\n", result.key, node)
		bits := matchNode.matchNodeBits(bitIndex, result)
		if bits >= 0 {
			// matched all node bits up the given count, so move into sub-nodes
			matchNode = matchNode.matchSubNode(bits, result)
			if matchNode == nil {
				// reached the end of the line
				break
			}
			// Matched a sub-node.
			// The sub-node was chosen according to the next bit.
			// That bit is therefore now a match,
			// so increment the matched bits by 1, and keep going.
			bitIndex = bits + 1
		} else {
			// reached the end of the line
			break
		}
	}
}

func (node *BinTrieNode) matchNodeBits(bitIndex int, result *opResult) BitCount {
	existingKey := node.GetKey()
	newKey := result.key
	if newKey.GetBitCount() != existingKey.GetBitCount() {
		panic("mismatched bit length between trie keys")
	} else if !newKey.MatchBits(existingKey, bitIndex, nodeCompare{node: node, result: result}) {
		if node.IsAdded() {
			node.handleContains(result)
		}
		return existingKey.GetPrefixLen().Len()
	}
	return -1
}

type nodeCompare struct {
	result *opResult
	node   *BinTrieNode
}

func (comp nodeCompare) BitsMatch() {
	node := comp.node
	result := comp.result
	result.containedBy = node.toTrieNode()
	existingKey := node.GetKey()
	existingPref := existingKey.GetPrefixLen()
	newKey := result.key
	newPrefixLen := newKey.GetPrefixLen()
	if existingPref == nil {
		if newPrefixLen == nil {
			// note that "added" is already true here,
			// we can only be here if explicitly inserted already
			// since it is a non-prefixed full address
			node.handleMatch(result)
		} else if newPrefixLen.Len() == newKey.GetBitCount() {
			node.handleMatch(result)
		} else {
			node.handleContained(result, newPrefixLen.Len())
		}
	} else {
		// we know newPrefixLen != nil since we know all of the bits of newAddr match,
		// which is impossible if newPrefixLen is nil and existingPref not nil
		if newPrefixLen.Len() == existingPref.Len() {
			if node.IsAdded() {
				node.handleMatch(result)
			} else {
				node.handleNodeMatch(result)
			}
		} else if existingPref.Len() == existingKey.GetBitCount() {
			node.handleMatch(result)
		} else { // existing prefix > newPrefixLen
			node.handleContained(result, newPrefixLen.Len())
		}
	}
}

func (comp nodeCompare) BitsDoNotMatch(matchedBits BitCount) {
	comp.node.handleSplitNode(comp.result, matchedBits)
}

func (node *BinTrieNode) handleContained(result *opResult, newPref int) {
	op := result.op
	if op == insert {
		// if we have 1.2.3.4 and 1.2.3.4/32, and we are looking at the last segment,
		// then there are no more bits to look at, and this makes the former a sub-node of the latter.
		// In most cases, however, there are more bits in existingAddr, the latter, to look at.
		node.replace(result, newPref)
	} else if op == subtreeDelete {
		node.removeSubtree(result)
	} else if op == near {
		node.findNearest(result, newPref)
	} else if op == remap {
		node.remapNonExistingReplace(result, newPref)
	}
}

func (node *BinTrieNode) handleContains(result *opResult) bool {
	result.smallestContaining = node.toTrieNode()
	if result.op == containing {
		result.addContaining(node.toTrieNode())
		return true
	}
	return false
}

func (node *BinTrieNode) handleSplitNode(result *opResult, totalMatchingBits BitCount) {
	op := result.op
	if op == insert {
		node.split(result, totalMatchingBits, node.createNew(result.key))
	} else if op == near {
		node.findNearest(result, totalMatchingBits)
	} else if op == remap {
		node.remapNonExistingSplit(result, totalMatchingBits)
	}
}

// a node exists for the given key but the node is not added,
// so not a match, but a split not required
func (node *BinTrieNode) handleNodeMatch(result *opResult) {
	op := result.op
	if op == lookup {
		result.existingNode = node.toTrieNode()
	} else if op == insert {
		node.existingAdded(result)
	} else if op == subtreeDelete {
		node.removeSubtree(result)
	} else if op == near {
		node.findNearestFromMatch(result)
	} else if op == remap {
		node.remapNonAdded(result)
	}
}

func (node *BinTrieNode) handleMatch(result *opResult) {
	result.exists = true
	if !node.handleContains(result) {
		op := result.op
		if op == lookup {
			node.matched(result)
		} else if op == insert {
			node.matchedInserted(result)
		} else if op == insertedDelete {
			node.removeOp(result)
		} else if op == subtreeDelete {
			node.removeSubtree(result)
		} else if op == near {
			if result.nearExclusive {
				node.findNearestFromMatch(result)
			} else {
				node.matched(result)
			}
		} else if op == remap {
			node.remapMatch(result)
		}
	}
}

func (node *BinTrieNode) remapNonExistingReplace(result *opResult, totalMatchingBits BitCount) {
	if node.remap(result, false) {
		node.replace(result, totalMatchingBits)
	}
}

func (node *BinTrieNode) remapNonExistingSplit(result *opResult, totalMatchingBits BitCount) {
	if node.remap(result, false) {
		node.split(result, totalMatchingBits, node.createNew(result.key))
	}
}

func (node *BinTrieNode) remapNonExisting(result *opResult) *BinTrieNode {
	if node.remap(result, false) {
		return node.createNew(result.key)
	}
	return nil
}

func (node *BinTrieNode) remapNonAdded(result *opResult) {
	if node.remap(result, false) {
		node.existingAdded(result)
	}
}

func (node *BinTrieNode) remapMatch(result *opResult) {
	result.existingNode = node.toTrieNode()
	if node.remap(result, true) {
		node.matchedInserted(result)
	}
}

type remapAction int

const (
	doNothing remapAction = iota
	removeNode
	remapValue
)

// Remaps the value for a node to a new value.
// This operation works on mapped values
// It returns true if a new node needs to be created (match is nil) or added (match is non-nil)
func (node *BinTrieNode) remap(result *opResult, isMatch bool) bool {
	remapper := result.remapper
	change := node.cTracker.getCurrent()
	var existingValue V
	if isMatch {
		existingValue = node.GetValue()
	}
	result.existingValue = existingValue
	newValue, action := remapper(existingValue)
	if action == doNothing {
		return false
	} else if action == removeNode {
		if isMatch {
			//node.cTracker.changedSince(change)xxx
			cTracker := node.cTracker
			if cTracker != nil && cTracker.changedSince(change) {
				panic("the tree has been modified by the remap")
			}
			node.ClearValue()
			node.removeOp(result)
		}
		return false
	} else if isMatch {
		// functions, maps and slices are not comparable, so here we do not compare values to determine if new one is different
		//node.cTracker.changedSince(change)xxx
		cTracker := node.cTracker
		if cTracker != nil && cTracker.changedSince(change) {
			panic("the tree has been modified by the remap")
		}
		result.newValue = newValue
		return true
	} else {
		result.newValue = newValue
		return true
	}
}

// this node matched when doing a lookup
func (node *BinTrieNode) matched(result *opResult) {
	thisNode := node.toTrieNode()
	result.existingNode = thisNode
	result.nearestNode = thisNode
}

// similar to matched, but when inserting we see it already there.
// this added node had already been added before
func (node *BinTrieNode) matchedInserted(result *opResult) {
	thisNode := node.toTrieNode()
	result.existingNode = thisNode
	result.addedAlready = thisNode
	result.existingValue = node.GetValue()
	node.SetValue(result.newValue)
}

// this node previously existed but was not added til now
func (node *BinTrieNode) existingAdded(result *opResult) {
	result.existingNode = node.toTrieNode()
	result.added = node.toTrieNode()
	node.added(result)
}

// this node is newly inserted and added
func (node *BinTrieNode) inserted(result *opResult) {
	result.inserted = node.toTrieNode()
	node.added(result)
}

func (node *BinTrieNode) added(result *opResult) {
	node.setNodeAdded(true)
	node.adjustCount(1)
	node.SetValue(result.newValue)
	node.cTracker.changed()
}

// The current node and the new node both become sub-nodes of a new block node taking the position of the current node.
func (node *BinTrieNode) split(result *opResult, totalMatchingBits BitCount, newSubNode *BinTrieNode) {
	newBlock := node.GetKey().ToPrefixBlockLen(totalMatchingBits)
	node.replaceToSub(newBlock, totalMatchingBits, newSubNode)
	newSubNode.inserted(result)
}

// The current node is replaced by the new node and becomes a sub-node of the new node.
func (node *BinTrieNode) replace(result *opResult, totalMatchingBits BitCount) {
	result.containedBy = node.toTrieNode()
	newNode := node.replaceToSub(result.key, totalMatchingBits, nil)
	newNode.inserted(result)
}

// The current node is replaced by a new block of the given key.
// The current node and given node become sub-nodes.
func (node *BinTrieNode) replaceToSub(newAssignedKey TrieKey, totalMatchingBits BitCount, newSubNode *BinTrieNode) *BinTrieNode {
	newNode := node.createNew(newAssignedKey)
	newNode.storedSize = node.storedSize
	parent := node.GetParent()
	if parent.GetUpperSubNode() == node.toTrieNode() {
		parent.setUpper(newNode)
	} else if parent.GetLowerSubNode() == node.toTrieNode() {
		parent.setLower(newNode)
	}
	existingKey := node.GetKey()
	if totalMatchingBits < existingKey.GetBitCount() &&
		existingKey.IsOneBit(totalMatchingBits) {
		if newSubNode != nil {
			newNode.setLower(newSubNode)
		}
		newNode.setUpper(node.toTrieNode())
	} else {
		newNode.setLower(node.toTrieNode())
		if newSubNode != nil {
			newNode.setUpper(newSubNode)
		}
	}
	return newNode
}

// only called when lower/higher and not floor/ceiling since for a match ends things for the latter
func (node *BinTrieNode) findNearestFromMatch(result *opResult) {
	if result.nearestFloor {
		// looking for greatest element < queried address
		// since we have matched the address, we must go lower again,
		// and if we cannot, we must backtrack
		lower := node.GetLowerSubNode()
		if lower == nil {
			// no nearest node yet
			result.backtrackNode = node.toTrieNode()
		} else {
			var last *BinTrieNode
			for {
				last = lower
				lower = lower.GetUpperSubNode()
				if lower == nil {
					break
				}
			}
			result.nearestNode = last
		}
	} else {
		// looking for smallest element > queried address
		upper := node.GetUpperSubNode()
		if upper == nil {
			// no nearest node yet
			result.backtrackNode = node.toTrieNode()
		} else {
			var last *BinTrieNode
			for {
				last = upper
				upper = upper.GetLowerSubNode()
				if upper == nil {
					break
				}
			}
			result.nearestNode = last
		}
	}
}

func (node *BinTrieNode) findNearest(result *opResult, differingBitIndex BitCount) {
	thisKey := node.GetKey()
	if differingBitIndex < thisKey.GetBitCount() && thisKey.IsOneBit(differingBitIndex) {
		// this element and all below are > than the query address
		if result.nearestFloor {
			// looking for greatest element < or <= queried address, so no need to go further
			// need to backtrack and find the last right turn to find node < than the query address again
			result.backtrackNode = node.toTrieNode()
		} else {
			// looking for smallest element > or >= queried address
			lower := node.toTrieNode()
			var last *BinTrieNode
			for {
				last = lower
				lower = lower.GetLowerSubNode()
				if lower == nil {
					break
				}
			}
			result.nearestNode = last
		}
	} else {
		// this element and all below are < than the query address
		if result.nearestFloor {
			// looking for greatest element < or <= queried address
			upper := node.toTrieNode()
			var last *BinTrieNode
			for {
				last = upper
				upper = upper.GetUpperSubNode()
				if upper == nil {
					break
				}
			}
			result.nearestNode = last
		} else {
			// looking for smallest element > or >= queried address, so no need to go further
			// need to backtrack and find the last left turn to find node > than the query address again
			result.backtrackNode = node.toTrieNode()
		}
	}
}

func (node *BinTrieNode) matchSubNode(bitIndex BitCount, result *opResult) *BinTrieNode {
	newKey := result.key
	if !freezeRoot && node.IsEmpty() {
		if result.op == remap {
			node.remapNonAdded(result)
		} else if result.op == insert {
			node.setKey(newKey)
			node.existingAdded(result)
		}
	} else if bitIndex >= newKey.GetBitCount() {
		// we matched all bits, yet somehow we are still going
		// this can only happen when mishandling a match between 1.2.3.4/32 to 1.2.3.4
		// which should never happen and so we do nothing, no match, no remap, no insert, no near
	} else if newKey.IsOneBit(bitIndex) {
		upper := node.GetUpperSubNode()
		if upper == nil {
			// no match
			op := result.op
			if op == insert {
				upper = node.createNew(newKey)
				node.setUpper(upper)
				upper.inserted(result)
			} else if op == near {
				if result.nearestFloor {
					// With only one sub-node at most, normally that would mean this node must be added.
					// But there is one exception, when we are the non-added root node.
					// So must check for added here.
					if node.IsAdded() {
						result.nearestNode = node.toTrieNode()
					} else {
						// check if our lower sub-node is there and added.  It is underneath addr too.
						// find the highest node in that direction.
						lower := node.GetLowerSubNode()
						if lower != nil {
							res := lower
							next := res.GetUpperSubNode()
							for next != nil {
								res = next
								next = res.GetUpperSubNode()
							}
							result.nearestNode = res
						}
					}
				} else {
					result.backtrackNode = node.toTrieNode()
				}
			} else if op == remap {
				upper = node.remapNonExisting(result)
				if upper != nil {
					node.setUpper(upper)
					upper.inserted(result)
				}
			}
		} else {
			return upper
		}
	} else {
		// In most cases, however, there are more bits in newKey, the former, to look at.
		lower := node.GetLowerSubNode()
		if lower == nil {
			// no match
			op := result.op
			if op == insert {
				lower = node.createNew(newKey)
				node.setLower(lower)
				lower.inserted(result)
			} else if op == near {
				if result.nearestFloor {
					result.backtrackNode = node.toTrieNode()
				} else {
					// With only one sub-node at most, normally that would mean this node must be added.
					// But there is one exception, when we are the non-added root node.
					// So must check for added here.
					if node.IsAdded() {
						result.nearestNode = node.toTrieNode()
					} else {
						// check if our upper sub-node is there and added.  It is above addr too.
						// find the highest node in that direction.
						upper := node.GetUpperSubNode()
						if upper != nil {
							res := upper
							next := res.GetLowerSubNode()
							for next != nil {
								res = next
								next = res.GetLowerSubNode()
							}
							result.nearestNode = res
						}
					}
				}
			} else if op == remap {
				lower = node.remapNonExisting(result)
				if lower != nil {
					node.setLower(lower)
					lower.inserted(result)
				}
			}
		} else {
			return lower
		}
	}
	return nil
}

func (node *BinTrieNode) createNew(newKey TrieKey) *BinTrieNode {
	res := &BinTrieNode{
		binTreeNode{
			item:     newKey,
			cTracker: node.cTracker,
		},
	}
	res.setAddr()
	return res
}

// PreviousAddedNode returns the previous node in the tree that is an added node, following the tree order in reverse,
// or nil if there is no such node.
func (node *BinTrieNode) PreviousAddedNode() *BinTrieNode {
	return node.toBinTreeNode().previousAddedNode().toTrieNode()
}

// NextAddedNode returns the next node in the tree that is an added node, following the tree order,
// or nil if there is no such node.
func (node *BinTrieNode) NextAddedNode() *BinTrieNode {
	return node.toBinTreeNode().nextAddedNode().toTrieNode()
}

// NextNode returns the node that follows this node following the tree order
func (node *BinTrieNode) NextNode() *BinTrieNode {
	return node.toBinTreeNode().nextNode().toTrieNode()
}

// PreviousNode returns the node that precedes this node following the tree order.
func (node *BinTrieNode) PreviousNode() *BinTrieNode {
	return node.toBinTreeNode().previousNode().toTrieNode()
}

func (node *BinTrieNode) FirstNode() *BinTrieNode {
	return node.toBinTreeNode().firstNode().toTrieNode()
}

func (node *BinTrieNode) FirstAddedNode() *BinTrieNode {
	return node.toBinTreeNode().firstAddedNode().toTrieNode()
}

func (node *BinTrieNode) LastNode() *BinTrieNode {
	return node.toBinTreeNode().lastNode().toTrieNode()
}

func (node *BinTrieNode) LastAddedNode() *BinTrieNode {
	return node.toBinTreeNode().lastAddedNode().toTrieNode()
}

func (node *BinTrieNode) findNodeNear(key TrieKey, below, exclusive bool) *BinTrieNode {
	result := &opResult{
		key:           key,
		op:            near,
		nearestFloor:  below,
		nearExclusive: exclusive,
	}
	if node != nil {
		node.matchBits(result)
	}
	backtrack := result.backtrackNode
	if backtrack != nil {
		parent := backtrack.GetParent()
		for parent != nil {
			if below {
				if backtrack != parent.GetLowerSubNode() {
					break
				}
			} else {
				if backtrack != parent.GetUpperSubNode() {
					break
				}
			}
			backtrack = parent
			parent = backtrack.GetParent()
		}

		if parent != nil {
			if parent.IsAdded() {
				result.nearestNode = parent
			} else {
				if below {
					result.nearestNode = parent.PreviousAddedNode()
				} else {
					result.nearestNode = parent.NextAddedNode()
				}

			}
		}
	}
	return result.nearestNode
}

func (node *BinTrieNode) LowerAddedNode(key TrieKey) *BinTrieNode {
	return node.findNodeNear(key, true, true)
}

func (node *BinTrieNode) FloorAddedNode(key TrieKey) *BinTrieNode {
	return node.findNodeNear(key, true, false)
}

func (node *BinTrieNode) HigherAddedNode(key TrieKey) *BinTrieNode {
	return node.findNodeNear(key, false, true)
}

func (node *BinTrieNode) CeilingAddedNode(key TrieKey) *BinTrieNode {
	return node.findNodeNear(key, false, false)
}

// Iterator returns an iterator that iterates through the elements of the sub-tree with this node as the root.
// The iteration is in sorted element order.
func (node *BinTrieNode) Iterator() TrieKeyIterator {
	return trieKeyIterator{node.toBinTreeNode().iterator()}
}

// DescendingIterator returns an iterator that iterates through the elements of the subtrie with this node as the root.
// The iteration is in reverse sorted element order.
func (node *BinTrieNode) DescendingIterator() TrieKeyIterator {
	return trieKeyIterator{node.toBinTreeNode().descendingIterator()}
}

// NodeIterator returns an iterator that iterates through the added nodes of the sub-tree with this node as the root, in forward or reverse tree order.
func (node *BinTrieNode) NodeIterator(forward bool) TrieNodeIteratorRem {
	return trieNodeIteratorRem{node.toBinTreeNode().nodeIterator(forward)}
}

// AllNodeIterator returns an iterator that iterates through all the nodes of the sub-tree with this node as the root, in forward or reverse tree order.
func (node *BinTrieNode) AllNodeIterator(forward bool) TrieNodeIteratorRem {
	return trieNodeIteratorRem{node.toBinTreeNode().allNodeIterator(forward)}
}

// BlockSizeNodeIterator returns an iterator that iterates the added nodes, ordered by keys from largest prefix blocks (smallest prefix length) to smallest (largest prefix length) and then to individual addresses,
// in the sub-trie with this node as the root.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order is taken.
func (node *BinTrieNode) BlockSizeNodeIterator(lowerSubNodeFirst bool) TrieNodeIteratorRem {
	return node.blockSizeNodeIterator(lowerSubNodeFirst, true)
}

// BlockSizeAllNodeIterator returns an iterator that iterates all the nodes, ordered by keys from largest prefix blocks to smallest and then to individual addresses,
// in the sub-trie with this node as the root.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (node *BinTrieNode) BlockSizeAllNodeIterator(lowerSubNodeFirst bool) TrieNodeIteratorRem {
	return node.blockSizeNodeIterator(lowerSubNodeFirst, false)
}

// BlockSizeCompare compares keys by block size and then by prefix value if block sizes are equal
func BlockSizeCompare(key1, key2 TrieKey, reverseBlocksEqualSize bool) int {
	if key2 == key1 {
		return 0
	}
	pref2 := key2.GetPrefixLen()
	pref1 := key1.GetPrefixLen()
	if pref2 != nil {
		if pref1 != nil {
			val := pref2.Len() - pref1.Len()
			if val == 0 {
				compVal := key2.Compare(key1)
				if reverseBlocksEqualSize {
					compVal = -compVal
				}
				return compVal
			}
			return val
		}
		return -1
	}
	if pref1 != nil {
		return 1
	}
	compVal := key2.Compare(key1)
	if reverseBlocksEqualSize {
		compVal = -compVal
	}
	return compVal
}

func (node *BinTrieNode) blockSizeNodeIterator(lowerSubNodeFirst, addedNodesOnly bool) TrieNodeIteratorRem {
	reverseBlocksEqualSize := !lowerSubNodeFirst
	var size int
	if addedNodesOnly {
		size = node.Size()
	}
	iter := newPriorityNodeIterator(
		size,
		addedNodesOnly,
		node.toBinTreeNode(),
		func(one, two e) int {
			val := BlockSizeCompare(one.(TrieKey), two.(TrieKey), reverseBlocksEqualSize)
			return -val
		})
	return trieNodeIteratorRem{&iter}
}

// BlockSizeCachingAllNodeIterator returns an iterator of all nodes, ordered by keys from largest prefix blocks to smallest and then to individual addresses,
// in the sub-trie with this node as the root.
//
// This iterator allows you to cache an object with subnodes so that when those nodes are visited the cached object can be retrieved.
func (node *BinTrieNode) BlockSizeCachingAllNodeIterator() CachingTrieNodeIterator {
	iter := newCachingPriorityNodeIterator(
		node.toBinTreeNode(),
		func(one, two e) int {
			val := BlockSizeCompare(one.(TrieKey), two.(TrieKey), false)
			return -val
		})
	return &cachingTrieNodeIterator{&iter}
}

func (node *BinTrieNode) ContainingFirstIterator(forwardSubNodeOrder bool) CachingTrieNodeIterator {
	return &cachingTrieNodeIterator{node.toBinTreeNode().containingFirstIterator(forwardSubNodeOrder)}
}

func (node *BinTrieNode) ContainingFirstAllNodeIterator(forwardSubNodeOrder bool) CachingTrieNodeIterator {
	return &cachingTrieNodeIterator{node.toBinTreeNode().containingFirstAllNodeIterator(forwardSubNodeOrder)}
}

func (node *BinTrieNode) ContainedFirstIterator(forwardSubNodeOrder bool) TrieNodeIteratorRem {
	return trieNodeIteratorRem{node.toBinTreeNode().containedFirstIterator(forwardSubNodeOrder)}
}

func (node *BinTrieNode) ContainedFirstAllNodeIterator(forwardSubNodeOrder bool) TrieNodeIterator {
	return trieNodeIterator{node.toBinTreeNode().containedFirstAllNodeIterator(forwardSubNodeOrder)}
}

// Clone clones the node.
// Keys remain the same, but the parent node and the lower and upper sub-nodes are all set to nil.
func (node *BinTrieNode) Clone() *BinTrieNode {
	return node.toBinTreeNode().clone().toTrieNode()
}

// CloneTree clones the sub-tree starting with this node as root.
// The nodes are cloned, but their keys and values are not cloned.
func (node *BinTrieNode) CloneTree() *BinTrieNode {
	return node.toBinTreeNode().cloneTree().toTrieNode()
}

// AsNewTrie creates a new sub-trie, copying the nodes starting with this node as root.
// The nodes are copies of the nodes in this sub-trie, but their keys and values are not copies.
func (node *BinTrieNode) AsNewTrie() *BinTrie {
	// I suspect clone is faster - in Java I used AddTrie to add the bounded part of the trie if it was bounded
	// but AddTrie needs to insert nodes amongst existing nodes, clone does not
	// newTrie := NewBinTrie(key)
	// newTrie.AddTrie(node)
	key := node.GetKey()
	trie := &BinTrie{binTree{}}
	rootKey := key.ToPrefixBlockLen(0)
	trie.setRoot(rootKey)
	root := trie.root
	newNode := node.cloneTreeTrackerBounds(root.cTracker, nil)
	if rootKey.Compare(key) == 0 {
		root.setUpper(newNode.upper)
		root.setLower(newNode.lower)
		if node.IsAdded() {
			root.SetAdded()
		}
		root.SetValue(node.GetValue())
	} else if key.IsOneBit(0) {
		root.setUpper(newNode)
	} else {
		root.setLower(newNode)
	}
	root.storedSize = sizeUnknown
	return trie
}

// Equal returns whether the key matches the key of the given node
func (node *BinTrieNode) Equal(other *BinTrieNode) bool {
	if node == nil {
		return other == nil
	} else if other == nil {
		return false
	}
	return node == other || node.GetKey().Compare(other.GetKey()) == 0
}

// DeepEqual returns whether the key matches the key of the given node using Compare,
// and whether the value matches the other value using reflect.DeepEqual
func (node *BinTrieNode) DeepEqual(other *BinTrieNode) bool {
	if node == nil {
		return other == nil
	} else if other == nil {
		return false
	}
	return node.GetKey().Compare(other.GetKey()) == 0 && reflect.DeepEqual(node.GetValue(), other.GetValue())
}

// TreeEqual returns whether the sub-tree represented by this node as the root node matches the given sub-tree, matching the trie keys using the Compare method
func (node *BinTrieNode) TreeEqual(other *BinTrieNode) bool {
	if other == node {
		return true
	} else if other.Size() != node.Size() {
		return false
	}
	these, others := node.Iterator(), other.Iterator()
	thisKey := these.Next()
	for ; thisKey != nil; thisKey = these.Next() {
		if thisKey.Compare(others.Next()) != 0 {
			return false
		}
	}
	return true
}

// TreeDeepEqual returns whether the sub-tree represented by this node as the root node matches the given sub-tree, matching the nodes using DeepEqual
func (node *BinTrieNode) TreeDeepEqual(other *BinTrieNode) bool {
	if other == node {
		return true
	} else if other.Size() != node.Size() {
		return false
	}
	these, others := node.NodeIterator(true), other.NodeIterator(true)
	thisNode := these.Next()
	for ; thisNode != nil; thisNode = these.Next() {
		if thisNode.DeepEqual(others.Next()) {
			return false
		}
	}
	return true
}

// Compare returns -1, 0 or 1 if this node is less than, equal, or greater than the other, according to the key and the trie order.
func (node *BinTrieNode) Compare(other *BinTrieNode) int {
	if node == nil {
		if other == nil {
			return 0
		}
		return -1
	} else if other == nil {
		return 1
	}
	return node.GetKey().Compare(other.GetKey())
}

// For some reason Format must be here and not in addressTrieNode for nil node.
// It panics in fmt code either way, but if in here then it is handled by a recover() call in fmt properly.
// Seems to be a problem only in the debugger.

// Format implements the fmt.Formatter interface
func (node BinTrieNode) Format(state fmt.State, verb rune) {
	node.format(state, verb)
}

// TrieIncrement returns the next key according to the trie ordering
func TrieIncrement(key TrieKey) TrieKey {
	prefLen := key.GetPrefixLen()
	if prefLen != nil {
		return key.ToMinUpper()
	}
	bitCount := key.GetBitCount()
	trailingBits := key.GetTrailingBitCount(false)
	if trailingBits < bitCount {
		return key.ToPrefixBlockLen(bitCount - (trailingBits + 1))
	}
	return nil
}

// TrieDecrement returns the previous key according to the trie ordering
func TrieDecrement(key TrieKey) TrieKey {
	prefLen := key.GetPrefixLen()
	if prefLen != nil {
		return key.ToMaxLower()
	}
	bitCount := key.GetBitCount()
	trailingBits := key.GetTrailingBitCount(true)
	if trailingBits < bitCount {
		return key.ToPrefixBlockLen(bitCount - (trailingBits + 1))
	}
	return nil
}
