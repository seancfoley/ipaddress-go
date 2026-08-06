//
// Copyright 2026 Sean C Foley
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
	"math/big"

	"github.com/seancfoley/bintree/tree"
)

type collectionTrie[T ipAddressTypeConstraint[T]] struct {
	Trie[T]
}

func (trie *collectionTrie[T]) addIfNoElementsContaining(block T) *TrieNode[T] {
	return toAddressTrieNode(trie.addIfNoElementsContainingNoCheck(block))
}

func (trie *collectionTrie[T]) elementsIntersected(block T) *TrieNode[T] {
	return trie.elementsIntersectedByNoCheck(block)
}

func (trie *collectionTrie[T]) add(parentNode *TrieNode[T], addr T) bool {
	if parentNode == nil {
		return trie.Trie.addNoCheck(addr)
	}
	return parentNode.containmentReplace(addr)
}

func (trie *collectionTrie[T]) containingLowerAddedNode(addr T) *TrieNode[T] {
	return toAddressTrieNode(trie.containingLowerAddedNodeNoCheck(addr))
}

func (trie *collectionTrie[T]) containingHigherAddedNode(addr T) *TrieNode[T] {
	return toAddressTrieNode(trie.containingHigherAddedNodeNoCheck(addr))
}

func (trie *collectionTrie[T]) containingFloorAddedNode(addr T) *TrieNode[T] {
	return toAddressTrieNode(trie.containingFloorAddedNodeNoCheck(addr))
}

func (trie *collectionTrie[T]) containingCeilingAddedNode(addr T) *TrieNode[T] {
	return toAddressTrieNode(trie.containingCeilingAddedNodeNoCheck(addr))
}

// ContainmentTrieBase is an IP address collection backed by an IP address trie.
//
// Sequential ranges and subnets are converted to prefix blocks in order to be inserted into the trie.
//
// It is one of the efficient options provided by this library to maintain sets of individual IP addresses,
// the other being [SequentialRangeList].
//
// Lookups of addresses and subnets are performed by binary search on the backing trie.
//
// Finding the address at a specific index in the list is performed by binary search on the sizes of the individual prefix blocks in the trie.
//
// The elements of this collection are individual IP addresses, unlike the Trie,
// in which the elements are individual addresses or CIDR prefix blocks,
// and an address can co-exist in the trie with CIDR prefix blocks that contain the address, as separate and distinct elements in the trie.
// This trie will change shape as addresses are added and removed to contain the minimal number of nodes to represent the addresses in the collection.
//
// IP address collection equality with another IP address collection is determined by the contents of the collections.
// A SequentialRangeList is equal to a ContainmentTrieBase if the collections contain the same set of individual addresses.
// The same is true for IP address aggregation equality.
//
// A ContainmentTrieBase may contain either IPv6 addresses, or IPv4 addresses, but not both at the same time.
// An attempt to add an address when the collection already contains an address of a different version will panic.
// However, once such a collection becomes empty again, it can accept either an IPv6 address or IPv4 address once more.
type ContainmentTrieBase[T ipAddressTypeConstraint[T]] struct {
	trie collectionTrie[T]
}

func (coll *ContainmentTrieBase[T]) addBlock(block T) bool {
	// The node is added to the trie if no existing elements in trie contain the new key.
	// The added node is returned.
	// Go up the chain of parents of the returned node.  If any such parent (which are all non-added) is full according to the contained count, we record it and keep going up.
	// After, the highest such node is set to added, and its two subnodes are removed.  Otherwise, the subnodes of the added node are removed.
	node := coll.trie.addIfNoElementsContaining(block)
	if node != nil {
		var largestFull *TrieNode[T]
		parent := node.GetParent()
		for parent != nil {
			if parent.ContainingMaxElements() {
				largestFull = parent
			}
			parent = parent.GetParent()
		}
		if largestFull == nil {
			node.RemoveChildren()
		} else {
			largestFull.SetAdded() // set to "added" first so the containment count roll-up is simpler for the children removal
			largestFull.RemoveChildren()
		}
		return true
	}
	return false
}

func (coll *ContainmentTrieBase[T]) removeBlock(block T) bool {
	deletedNode := coll.trie.elementsIntersected(block)
	if deletedNode != nil {
		parentNode := deletedNode.GetParent()
		if parentNode != nil && !parentNode.IsAdded() {
			parentNode = parentNode.GetParent()
		}
		deletedNode.Clear()
		deletedNodeKey := deletedNode.GetKey()
		if block.GetPrefixLen().Compare(deletedNodeKey.GetPrefixLen()) > 0 {
			remainder := deletedNodeKey.Subtract(block)
			for _, remainderAddr := range remainder {
				newBlocks := remainderAddr.SpanWithPrefixBlocks()
				for _, newBlock := range newBlocks {
					coll.trie.add(parentNode, newBlock.RemoveBitCountPrefixLen())
				}
			}
		}
		return true
	}
	return false
}

func (coll *ContainmentTrieBase[T]) addressPredicateOp(addr T, op func(T) bool, all, breakEarly, stripSingleAddressPrefLen bool) bool {
	if !addr.IsMultiple() {
		if !addr.IsPrefixed() {
			return op(addr)
		}
		return op(addr.WithoutPrefixLen())
	} else if addr.IsSinglePrefixBlock() { // fast track for prefix blocks
		return op(addr)
	}
	return coll.blocksPredicateOp(addr.SpanWithPrefixBlocks(), op, all, breakEarly, stripSingleAddressPrefLen)
}

func (coll *ContainmentTrieBase[T]) blocksPredicateOp(blocks []T, op func(T) bool, all, breakEarly, stripSingleAddressPrefLen bool) bool {
	result := all
	for _, block := range blocks {
		if stripSingleAddressPrefLen {
			block = block.RemoveBitCountPrefixLen()
		}
		res := op(block)
		if all { // all must return true for the full operation to be true
			if !res {
				result = false
				if breakEarly { // exit once the return value is finalized
					break
				}
			}
		} else { // any can return true for the full operation to be true
			if res {
				result = true
				if breakEarly { // exit once the return value is finalized
					break
				}
			}
		}
	}
	return result
}

func (coll *ContainmentTrieBase[T]) rangePredicateOp(rng *SequentialRange[T], op func(T) bool, all, breakEarly, stripSingleAddressPrefLen bool) bool {
	if !rng.IsMultiple() {
		addr := rng.GetLower()
		if stripSingleAddressPrefLen {
			addr = addr.RemoveBitCountPrefixLen()
		}
		return op(addr)
	}
	return coll.blocksPredicateOp(rng.SpanWithPrefixBlocks(), op, all, breakEarly, stripSingleAddressPrefLen)
}

// Clear empties the collection
func (coll *ContainmentTrieBase[T]) Clear() {
	coll.trie.clear()
}

// Add adds the address to the collection, if not already in the collection.
//
// If the address version does match existing addresses in the collection, the address is not added.
//
// Returns whether addresses were added, whether the collection was changed.
func (coll *ContainmentTrieBase[T]) Add(addr T) bool {
	return coll.addressPredicateOp(addr, coll.addBlock, false, false, true)
}

// AddSeqRange sdds the addresses in the sequential range to the collection, if not already in the collection.
//
// If the address version of the addresses in the collection does match the version of addresses in the given range, this method panics.
//
// Returns whether at least one address in the given sequential range was added, whether the collection was changed.
func (coll *ContainmentTrieBase[T]) AddSeqRange(rng *SequentialRange[T]) bool {
	return coll.rangePredicateOp(rng, coll.addBlock, false, false, true)
}

// Remove removes the given address from the collection.  It returns true if the collection was changed.
// It returns false if the address was not in the collection.
func (coll *ContainmentTrieBase[T]) Remove(addr T) bool {
	return coll.addressPredicateOp(addr, coll.removeBlock, false, false, false)
}

// RemoveSeqRange removes all the addresses in the sequential range from the collection.
// Returns true if the collection was changed.
func (coll *ContainmentTrieBase[T]) RemoveSeqRange(rng *SequentialRange[T]) bool {
	return coll.rangePredicateOp(rng, coll.removeBlock, false, false, false)
}

// Contains returns true if and only if this collection contains all the individual addresses in the given address or subnet
func (coll *ContainmentTrieBase[T]) Contains(address AddressType) bool {
	addr, isNil, ok := ConvertAddressTypeCheckNil[T](address)
	return ok && !isNil && coll != nil && coll.addressPredicateOp(addr, coll.trie.elementContainsNoCheck, true, true, false)
}

// ContainsAddress returns true if and only if this collection contains all the individual addresses in the given address or subnet
func (coll *ContainmentTrieBase[T]) ContainsAddress(addr T) bool {
	return coll != nil && !isNilPtr(addr) && coll.addressPredicateOp(addr, coll.trie.elementContainsNoCheck, true, true, false)
}

// ContainsRange returns true if and only if all individual addresses in the given sequential range are also in this collection
// Implements the IPAddressAggregation interface.
func (coll *ContainmentTrieBase[T]) ContainsRange(rng IPAddressSeqRangeType) bool {
	r, isNil, ok := ConvertRangeTypeCheckNil[T](rng)
	return ok && !isNil && coll != nil && coll.rangePredicateOp(r, coll.trie.elementContainsNoCheck, true, true, false)
}

// ContainsSeqRange returns true if and only if all individual addresses in the given sequential range are also in this collection
// Implements the IPAddressCollAddrConstraint interface.
func (coll *ContainmentTrieBase[T]) ContainsSeqRange(rng *SequentialRange[T]) bool {
	return coll != nil && rng != nil && coll.rangePredicateOp(rng, coll.trie.elementContainsNoCheck, true, true, false)
}

// OverlapsAddr returns true if and only the given individual address or subnet contains at least one individual address that is also in this collection.
// Implements the IPAddressAggregation interface.
func (coll *ContainmentTrieBase[T]) OverlapsAddr(address AddressType) bool {
	addr, isNil, ok := ConvertAddressTypeCheckNil[T](address)
	return ok && !isNil && coll != nil && coll.addressPredicateOp(addr, coll.trie.elementOverlapsNoCheck, false, true, false)
}

// OverlapsAddress returns true if and only the given individual address or subnet contains at least one individual address that is also in this collection.
// In a trie of prefix blocks, for a block to overlap with another block means that one of the two blocks contains the other, or they are equal.
// Implements the IPAddressCollAddrConstraint interface.
func (coll *ContainmentTrieBase[T]) OverlapsAddress(addr T) bool {
	return coll != nil && !isNilPtr(addr) && coll.addressPredicateOp(addr, coll.trie.elementOverlapsNoCheck, false, true, false)
}

func (coll *ContainmentTrieBase[T]) OverlapsRange(rng IPAddressSeqRangeType) bool {
	r, isNil, ok := ConvertRangeTypeCheckNil[T](rng)
	return ok && !isNil && coll != nil && coll.rangePredicateOp(r, coll.trie.elementOverlapsNoCheck, false, true, false)
}

// OverlapsRange returns true if and only if the given sequential range overlaps with blocks or addresses in the trie.
// In a trie of prefix blocks, for a block to overlap with another block means that one of the two blocks contains the other, or they are equal.
func (coll *ContainmentTrieBase[T]) OverlapsSeqRange(rng *SequentialRange[T]) bool {
	return coll != nil && rng != nil && coll.rangePredicateOp(rng, coll.trie.elementOverlapsNoCheck, false, true, false)
}

// Lower returns the highest address in the collection strictly less than all addresses in the given address or subnet.
func (coll *ContainmentTrieBase[T]) Lower(addr T) T {
	addr = addr.GetLower()
	node := coll.trie.containingLowerAddedNode(addr)
	if node == nil {
		return trieKeyZero[T]()
	}
	key := node.GetKey()
	if key.Contains(addr) {
		return addr.WithoutPrefixLen().DecrementSingle()
	}
	return key.WithoutPrefixLen().GetUpper()
}

// Floor returns the highest address in the collection less than or equal to the lowest address in the given address or subnet.
func (coll *ContainmentTrieBase[T]) Floor(addr T) T {
	addr = addr.GetLower()
	node := coll.trie.containingFloorAddedNode(addr)
	if node == nil {
		return trieKeyZero[T]()
	}
	key := node.GetKey()
	if key.Contains(addr) {
		return addr.WithoutPrefixLen()
	}
	return key.WithoutPrefixLen().GetUpper()
}

// Ceiling returns the lowest address in the collection greater than or equal to the highest address in the given address or subnet.
func (coll *ContainmentTrieBase[T]) Ceiling(addr T) T {
	addr = addr.GetUpper()
	node := coll.trie.containingCeilingAddedNode(addr)
	if node == nil {
		return trieKeyZero[T]()
	}
	key := node.GetKey()
	if key.Contains(addr) {
		return addr.WithoutPrefixLen()
	}
	return key.WithoutPrefixLen().GetLower()
}

// Higher returns the lowest address in the collection strictly greater than all addresses in the given address or subnet.
func (coll *ContainmentTrieBase[T]) Higher(addr T) T {
	addr = addr.GetUpper()
	node := coll.trie.containingHigherAddedNode(addr)
	if node == nil {
		return trieKeyZero[T]()
	}
	key := node.GetKey()
	if key.Contains(addr) {
		return addr.WithoutPrefixLen().IncrementSingle()
	}
	return key.WithoutPrefixLen().GetLower()
}

// CoverWithPrefixBlock returns the unique CIDR prefix block subnet or individual address of minimal size that includes all the addresses in this collection.
// If there are no addresses in this collection, then nil is returned.
func (coll *ContainmentTrieBase[T]) CoverWithPrefixBlock() T {
	root := coll.trie.GetRoot()
	var coveringNode *TrieNode[T]
	if root.IsAdded() {
		coveringNode = root
	} else {
		lower, upper := root.GetLowerSubNode(), root.GetUpperSubNode()
		if lower == nil {
			if upper == nil {
				return trieKeyZero[T]()
			} else {
				coveringNode = upper
			}
		} else if upper == nil {
			coveringNode = lower
		} else {
			coveringNode = root
		}
	}
	return coveringNode.GetKey()
}

func trieKeyZero[T ipAddressTypeConstraint[T]]() (t T) {
	return
}

// Get returns the address at the given index into the sorted collection, with the index of zero returning the first address.
// It is much like indexing a slice or array.
//
// If the increment is negative, or the increment exceeds GetCount() - 1, this method panics.
func (coll *ContainmentTrieBase[T]) Get(addressIndex int64) T {
	node, index := coll.trie.GetElementAddress(addressIndex)
	return node.getKey().WithoutPrefixLen().Increment(index)
}

// GetBig returns the address at the given index into the sprted collection, with the index of zero returning the first address.
// It is much like indexing a slice or array.
//
// If the increment is negative, or the increment exceeds GetCount() - 1, this method panics.
func (coll *ContainmentTrieBase[T]) GetBig(addressIndex *big.Int) T {
	node, index := coll.trie.GetElementAddressBig(addressIndex)
	return node.getKey().WithoutPrefixLen().IncrementBig(index)
}

// RemoveAt removes the individual address at the given index into the lists of addresses.  Returns that address.
// Similar to Get but also removes the address found.
//
// If the index is negative or larger than GetCount() - 1, this method panics.
func (coll *ContainmentTrieBase[T]) RemoveAt(addressIndex int64) T {
	deletedNode, index := coll.trie.GetElementAddress(addressIndex)
	address := deletedNode.getKey().WithoutPrefixLen().Increment(index)
	return coll.removeFromNode(deletedNode, address)
}

// RemoveAt removes the individual address at the given index into the lists of addresses.  Returns that address.
// Similar to GetBig but also removes the address found.
//
// If the index is negative or larger than GetCount() - 1, this method panics.
func (coll *ContainmentTrieBase[T]) RemoveAtBig(addressIndex *big.Int) T {
	deletedNode, index := coll.trie.GetElementAddressBig(addressIndex)
	address := deletedNode.getKey().WithoutPrefixLen().IncrementBig(index)
	return coll.removeFromNode(deletedNode, address)
}

func (coll *ContainmentTrieBase[T]) removeFromNode(deletedNode *TrieNode[T], address T) T {
	parentNode := deletedNode.GetParent()
	if parentNode != nil && !parentNode.IsAdded() {
		parentNode = parentNode.GetParent()
	}
	deletedNode.Clear()
	deletedNodeKey := deletedNode.GetKey()
	if deletedNodeKey.IsMultiple() {
		remainder := deletedNodeKey.Subtract(address)
		for _, remainderAddr := range remainder {
			newBlocks := remainderAddr.SpanWithPrefixBlocks()
			for _, newBlock := range newBlocks {
				coll.trie.add(parentNode, newBlock.RemoveBitCountPrefixLen())
			}
		}
	}
	return address
}

// CoverWithSequentialRange returns the unique sequential range of minimal size that includes all the addresses in this collection.
// If there are no addresses in this collection, then nil is returned.
//
// The result will represent the same set of addresses if and only if the set of addresses in this collection are sequential, in which case IsSequential is true.
func (coll *ContainmentTrieBase[T]) CoverWithSequentialRange() *SequentialRange[T] {
	if coll.IsEmpty() {
		return nil
	}
	return spanWithRange(coll.GetLowerAndUpper())
}

// GetCount returns the number of individual addresses in this containment trie, the number of elements in this collection.
func (coll *ContainmentTrieBase[T]) GetCount() *big.Int {
	if coll == nil {
		return bigZero()
	}
	return coll.trie.GetMatchingAddressCount()
}

// GetLower returns the individual address with the lowest numeric value in this collection.
func (coll *ContainmentTrieBase[T]) GetLower() T {
	firstNode := coll.trie.FirstAddedNode()
	if firstNode == nil {
		return trieKeyZero[T]()
	}
	return firstNode.GetKey().GetLower()
}

// GetUpper returns the individual addresses with the highest numeric value in this collection.
func (coll *ContainmentTrieBase[T]) GetUpper() T {
	lastNode := coll.trie.LastAddedNode()
	if lastNode == nil {
		return trieKeyZero[T]()
	}
	return lastNode.GetKey().GetUpper()
}

// GetLowerAndUpper returns the individual addresses with the lowest and highest numeric values in this collection.
func (coll *ContainmentTrieBase[T]) GetLowerAndUpper() (lower, upper T) {
	return coll.GetLower(), coll.GetUpper()
}

// Iterator provides an iterator to iterate through the individual IP addresses in this collection in order.
//
// Use the function ipaddr.StdPushIterator to convert the returned iterator to a standard library iter.Seq
func (coll *ContainmentTrieBase[T]) Iterator() Iterator[T] {
	if coll == nil {
		return nilIterator[T]()
	}
	trie := &coll.trie.Trie
	changeTracker := trie.changeTracker()
	var currentChange tree.Change
	if changeTracker != nil { // can be nil with empty trie
		currentChange = changeTracker.GetCurrent()
	}
	return &containmentTrieIterator[T]{
		trie:          trie,
		changeTracker: changeTracker,
		currentChange: currentChange,
		trieIterator:  trie.Iterator(),
	}
}

// AddressIterator is the same as Iterator while satisying the AddressAggregation interface
func (coll *ContainmentTrieBase[T]) AddressIterator() Iterator[AddressType] {
	return addrTypeIterator[T]{coll.Iterator()}
}

// PrefixBlockIterator returns an iterator for iterating through the prefix blocks in the backing trie, in sorted order.
//
// These prefix blocks are the minimal set of disjoint prefix blocks for containing the addresses in this collection of addresses.
func (coll *ContainmentTrieBase[T]) PrefixBlockIterator() IteratorWithRemove[T] {
	if coll == nil {
		return nilIteratorWithRemove[T]()
	}
	return coll.trie.iterator()
}

// SpanningPrefixBlockIterator returns an iterator for iterating through the minimal set of disjoint prefix blocks containing the addresses in this collection of addresses.
//
// It returns the same iterator as PrefixBlockIterator, while also satisifying the IPAddressAggregationConstraint interface.
func (coll *ContainmentTrieBase[T]) SpanningPrefixBlockIterator() Iterator[T] {
	return coll.PrefixBlockIterator()
}

// GetPrefixBlockCount returns the number of prefix blocks in the backing trie.
//
// The prefix blocks are the minimal set of disjoint prefix blocks for containing the addresses in this  collection of addresses.
func (coll *ContainmentTrieBase[T]) GetPrefixBlockCount() int {
	return coll.trie.size()
}

// GetLowerPrefixBlock returns the prefix block in the backing trie containing the lowest numeric value in this collection, or nil if the trie is empty
func (coll *ContainmentTrieBase[T]) GetLowerPrefixBlock() T {
	firstNode := coll.trie.FirstAddedNode()
	if firstNode == nil {
		return trieKeyZero[T]()
	}
	return firstNode.GetKey()
}

// GetUpperPrefixBlock returns the prefix block in the backing trie containing the highest numeric value in this collection, or nil if the trie is empty
func (coll *ContainmentTrieBase[T]) GetUpperPrefixBlock() T {
	lastNode := coll.trie.LastAddedNode()
	if lastNode == nil {
		return trieKeyZero[T]()
	}
	return lastNode.GetKey()
}

// IsEmpty returns true if and only if there are no elements in this collection.
func (coll *ContainmentTrieBase[T]) IsEmpty() bool {
	if coll == nil {
		return true
	}
	return coll.trie.IsEmpty()
}

// IsMultiple returns true if this collection contains more than 1 element.
func (coll *ContainmentTrieBase[T]) IsMultiple() bool {
	if coll == nil {
		return false
	}
	size := coll.trie.Size()
	if size > 1 {
		return true
	}
	firstNode := coll.trie.FirstAddedNode()
	if firstNode != nil {
		return firstNode.GetKey().IsMultiple()
	}
	return false
}

// IncludesZero Returns whether this collection contains the address matching the version of addresses in this list and having the value of zero.
func (coll *ContainmentTrieBase[T]) IncludesZero() bool {
	firstNode := coll.trie.FirstAddedNode()
	return firstNode != nil && firstNode.GetKey().IncludesZero()
}

// IncludesMax Returns whether this list contains the address matching the version of the addresses in this list and has the maximum value for addresses of that address version.
func (coll *ContainmentTrieBase[T]) IncludesMax() bool {
	lastNode := coll.trie.LastAddedNode()
	return lastNode != nil && lastNode.GetKey().IncludesMax()
}

// IsSequential returns whether the collection represents a range of addresses that are sequential.
//
// Generally, this means that given any two addresses in the collection, all addresses between are also in the collection.
func (coll *ContainmentTrieBase[T]) IsSequential() bool {
	if coll.IsEmpty() {
		return true
	}
	lower, upper := coll.GetLowerAndUpper()
	count := lower.Enumerate(upper)
	count.Add(count, bigOneConst())
	return count.Cmp(coll.GetCount()) == 0
}

// Format implements the [fmt.Formatter] interface.
func (coll ContainmentTrieBase[T]) Format(state fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		_, _ = state.Write([]byte(coll.String()))
		return
	}
	coll.trie.Format(state, verb)
}

// String provides a string representation of the collection which shows the underlying trie structure.
func (coll *ContainmentTrieBase[T]) String() string {
	if coll == nil {
		return "\n" + nilString()
	}
	return coll.trie.TreeStringWithCounts(true, false, true)
}

// EnumerateAddress returns the distance of the given address from the initial value of this containment trie.  It indicates where an address sits relative to the collection ordering.
//
//	If within or above the addresses in containment trie, it is the distance to the lower boundary of the collection.  If below the containment trie, it returns the number of addresses following the address to the initial address in the collection.
//
// You can call Contains or you can compare with GetCount to check for containment.
// An IP address is in the collection if 0 <= Enumerate(IP Address) < GetCount.
//
// If the address is above the lower boundary and below the upper boundary of the containment trie, but is not within a prefix block in the containment trie, then this method returns nil.
//
// Returns nil when the argument is a multi-valued subnet. The argument must be an individual address.
//
// Returns nil when the containment trie is empty.
//
// Returns nil when the address version does not match the addresses in this containment trie.
func (coll *ContainmentTrieBase[T]) EnumerateAddress(address T) *big.Int {
	if isNilPtr(address) || address.IsMultiple() || coll.IsEmpty() {
		return nil
	}
	return coll.enumerateAddress(address)
}

func (coll *ContainmentTrieBase[T]) enumerateAddress(address T) *big.Int {
	if coll.IsEmpty() {
		return nil
	}
	lower := coll.GetLowerPrefixBlock()
	if !versionsMatch(lower, address) {
		return nil
	}
	node, index := coll.trie.enumerateNoCheck(address)
	if node == nil {
		if compareLowerValues(lower, address) > 0 {
			return lower.Enumerate(address)
		}
		upper := coll.GetUpperPrefixBlock()
		if compareUpperValues(upper, address) < 0 {
			res := upper.Enumerate(address)
			if res != nil {
				return res.Add(res, coll.GetCount())
			}
		}
		return nil
	}
	return index.Add(index, node.GetKey().Enumerate(address))
}

// Enumerate returns the distance of the given address from the initial value of this containment trie.  It indicates where an address sits relative to the collection ordering.
//
// If within or above thecontainment trie, it is the distance to the lower boundary of the collection.  If below the containment trie, it returns the number of addresses following the address to the collection boundary.
//
// You can call Contains or you can compare with GetCount to check for containment.
// An IP address is in the collection if 0 <= Enumerate(IP Address) < GetCount.
//
// If the address is above the lower boundary and below the upper boundary of the containment trie, but is not within a prefix block in the containment trie, then this method returns nil.
//
// Returns nil when the argument is a multi-valued subnet. The argument must be an individual address.
//
// Returns nil when the containment trie is empty.
//
// Returns nil when the address version does not match the addresses in this containment trie.
func (coll *ContainmentTrieBase[T]) Enumerate(address AddressType) *big.Int {
	addr, isNil, ok := ConvertAddressTypeCheckNil[T](address)
	if !ok || isNil || coll == nil {
		return nil
	}
	return coll.enumerateAddress(addr)
}

// Equal returns true if and only if this collection has the same set of individual addresses as the given collectioj
func (coll *ContainmentTrieBase[T]) Equal(other *ContainmentTrieBase[T]) bool {
	if coll == nil {
		return other == nil || other.IsEmpty()
	} else if other == nil {
		return coll.IsEmpty()
	}
	return coll.trie.Equal(&other.trie.Trie) //   nil aggregation contains
}

// EqualAggregation returns true if and only if this collection has the same set of individual addresses as the given aggregation of addresses
func (coll *ContainmentTrieBase[T]) EqualAggregation(otherAggregation AddressAggregation) bool {
	if coll == nil {
		return IsEmpty(otherAggregation)
	}
	switch other := otherAggregation.(type) {
	case nil:
		return false
	case AddressType:
		return coll.equalAddr(other)
	case IPAddressSeqRangeType:
		return coll.equalRange(other)
	case *ContainmentTrieBase[T]:
		return coll.Equal(other)
	case *IPAddressContainmentTrie:
		return equalContainmentTries(coll, other)
	case *IPv4AddressContainmentTrie:
		return equalContainmentTries(coll, other)
	case *IPv6AddressContainmentTrie:
		return equalContainmentTries(coll, other)
	case *IPAddressSeqRangeList:
		return equalListAndContainmentTrie(other, coll)
	case *IPv4AddressSeqRangeList:
		return equalListAndContainmentTrie(other, coll)
	case *IPv6AddressSeqRangeList:
		return equalListAndContainmentTrie(other, coll)
	default:
		return equalAggregation(coll, other)
	}
}

func equalAggregation(one, two AddressAggregation) bool {
	if one.GetCount().Cmp(two.GetCount()) != 0 {
		return false
	}
	iter, otherIter := one.AddressIterator(), two.AddressIterator()
	for iter.HasNext() {
		if !iter.Next().Equal(otherIter.Next()) {
			return false
		}
	}
	return true
}

func (coll *ContainmentTrieBase[T]) equalAddr(other AddressType) bool {
	if coll == nil {
		return isEmptyAddr(other) // addresses are never empty, unless they are nil pointers
	} else if other == nil {
		return IsEmpty(coll)
	}
	addr := other.ToAddressBase()
	if addr == nil {
		return IsEmpty(coll)
	}
	ipAddr := addr.ToIP()
	if ipAddr == nil {
		return false // not an IP address
	}
	addrIsSequential := ipAddr.IsSequential()
	trieIsSequential := coll.IsSequential()
	if addrIsSequential != trieIsSequential {
		return false
	} else if coll.GetCount().Cmp(other.GetCount()) != 0 {
		return false
	}
	if addrIsSequential {
		// when both are sequential just need to check the first and last, in fact really only the first since we already checked the counts
		return coll.GetLower().Equal(ipAddr.GetLower())
	}
	// neither address nor trie is sequential
	addrSeqBlocks := ipAddr.SequentialBlockIterator()
	prefBlocks := coll.PrefixBlockIterator()
	if addrSeqBlocks.HasNext() {
		for seqBlock := addrSeqBlocks.Next(); ; seqBlock = addrSeqBlocks.Next() {
			addrPrefBlocks := seqBlock.SpanWithPrefixBlocks()
			for _, addrPrefBlock := range addrPrefBlocks {
				// each should match the next one to be found in the collection
				if !prefBlocks.HasNext() {
					return false
				} else if !addrPrefBlock.Equal(prefBlocks.Next()) {
					return false
				}
			}
			if !addrSeqBlocks.HasNext() {
				break
			}
		}
	}
	return true
}

func (coll *ContainmentTrieBase[T]) equalRange(other IPAddressSeqRangeType) bool {
	if coll == nil {
		return isEmptyRange(other) // addresses are never empty, unless they are nil pointers
	} else if other == nil || other.ToIP() == nil {
		return IsEmpty(coll)
	}
	trieIsSequential := coll.IsSequential()
	if !trieIsSequential {
		return false
	} else if coll.GetCount().Cmp(other.GetCount()) != 0 {
		return false
	}
	// when both are sequential just need to check the first and last, in fact really only the first since we already checked the counts
	return coll.GetLower().Equal(other.ToIP().GetLower())
}

func equalContainmentTries[T ipAddressTypeConstraint[T], R ipAddressTypeConstraint[R]](coll *ContainmentTrieBase[T], other *ContainmentTrieBase[R]) bool {
	if coll == nil {
		return IsEmpty(other)
	} else if other == nil {
		return IsEmpty(coll)
	}
	if coll.GetCount().Cmp(other.GetCount()) != 0 {
		return false
	} else if coll.trie.Size() != other.trie.Size() {
		return false
	}
	these := coll.PrefixBlockIterator()
	others := other.PrefixBlockIterator()
	if these.HasNext() {
		for thisKey := these.Next(); ; thisKey = these.Next() {
			if !thisKey.Equal(others.Next()) {
				return false
			}
			if !these.HasNext() {
				break
			}
		}
	}
	return true
}

// Clone makes a copy of the collection
func (coll *ContainmentTrieBase[T]) Clone() *ContainmentTrieBase[T] {
	if coll == nil {
		return nil
	}
	return &ContainmentTrieBase[T]{
		trie: collectionTrie[T]{
			Trie: *coll.trie.Clone(),
		},
	}
}

// NewEmpty creates a new ContainmentTrieBase using the same element type T.
// Satisfies the IPAddressCollConstraint[S IPAddressCollAddrConstraint[T], T IPAddressTypeConstraint[T]] interface,
// allowing for ogeneric code that can create new collections generically, with generic code.
// For code that is using a generic collection type, you can simply use &ContainmentTrieBase[T]{}
func (list *ContainmentTrieBase[T]) NewEmpty() *ContainmentTrieBase[T] {
	return &ContainmentTrieBase[T]{}
}

// ComplementIntoNew returns a new collection comprising all the addresses not contained in this collection.
//
// If this list is empty and is not restricted to a single IP version of IPv4 or IPv6,
// then the IP version is ambiguous, so the complement is indeterminate, in which case nil is returned.
func (coll *ContainmentTrieBase[T]) ComplementIntoNew() *ContainmentTrieBase[T] {
	//fmt.Println("getting complement of", coll)
	newColl := &ContainmentTrieBase[T]{}
	if coll.IsEmpty() {
		var t T
		network := t.GetIPNetwork()
		// network can be nil if T is *IPAddress, in which case the IP version is indeterminate and the network is nil
		// network can be nil if the list contains the IPAddress zero address which is neither IPv4 or IPv6
		if network == nil {
			return nil
		}
		newColl.addBlock(network.GetAddressSpace())
		return newColl
	}
	iter := coll.PrefixBlockIterator()
	firstBlock := iter.Next()
	network := firstBlock.GetIPNetwork()
	previousUpper := firstBlock.GetUpper()
	zero, maxAddr := network.GetBoundaryAddresses()
	if !firstBlock.IncludesZero() {
		newColl.AddSeqRange(newSequRangeCheckSize(zero, firstBlock.DecrementSingle().WithoutPrefixLen()))
	}
	for iter.HasNext() {
		lower, upper := iter.Next().GetLowerAndUpper()
		afterPrev := previousUpper.IncrementSingle()
		belowThis := lower.DecrementSingle()
		cmpLower := compareLowerValues(afterPrev, belowThis)
		if cmpLower <= 0 {
			newColl.AddSeqRange(newSequRangeUnchecked(afterPrev.WithoutPrefixLen(), belowThis.WithoutPrefixLen(), cmpLower < 0))
		}
		previousUpper = upper
	}
	if !previousUpper.IncludesMax() {
		newColl.AddSeqRange(newSequRangeCheckSize(previousUpper.IncrementSingle().WithoutPrefixLen(), maxAddr))
	}
	return newColl
}

// JoinIntoNew creates a new containment trie that has all addresses in this containment trie and the provided containment trie.
func (coll *ContainmentTrieBase[T]) JoinIntoNew(other *ContainmentTrieBase[T]) *ContainmentTrieBase[T] {
	if coll.IsEmpty() {
		return other.Clone()
	} else if other.IsEmpty() {
		return coll.Clone()
	}
	result := &ContainmentTrieBase[T]{}
	thisIterator := coll.PrefixBlockIterator()
	otherIterator := other.PrefixBlockIterator()
	thisBlock := thisIterator.Next()
	otherBlock := otherIterator.Next()
	cmpLower := compareLowerValues(thisBlock, otherBlock)
	for {
		if cmpLower < 0 {
			if compareUpperValues(thisBlock, otherBlock) >= 0 {
				// otherBlock is contained within thisBlock
				if otherIterator.HasNext() {
					otherBlock = otherIterator.Next()
				} else {
					result.addBlock(thisBlock)
					for thisIterator.HasNext() {
						result.addBlock(thisIterator.Next())
					}
					break
				}
			} else {
				// otherBlock is above thisBlock
				result.addBlock(thisBlock)
				if thisIterator.HasNext() {
					thisBlock = thisIterator.Next()
					cmpLower = compareLowerValues(thisBlock, otherBlock)
				} else {
					result.addBlock(otherBlock)
					for otherIterator.HasNext() {
						result.addBlock(otherIterator.Next())
					}
					break
				}
			}
		} else if cmpLower > 0 {
			if compareUpperValues(thisBlock, otherBlock) <= 0 {
				// thisBlock is contained within otherBlock
				if thisIterator.HasNext() {
					thisBlock = thisIterator.Next()
				} else {
					result.addBlock(otherBlock)
					for otherIterator.HasNext() {
						result.addBlock(otherIterator.Next())
					}
					break
				}
			} else {
				// thisBlock is above otherBlock
				result.addBlock(otherBlock)
				if otherIterator.HasNext() {
					otherBlock = otherIterator.Next()
					cmpLower = compareLowerValues(thisBlock, otherBlock)
				} else {
					result.addBlock(thisBlock)
					for thisIterator.HasNext() {
						result.addBlock(thisIterator.Next())
					}
					break
				}
			}
		} else {
			cmpUpper := compareUpperValues(thisBlock, otherBlock)
			if cmpUpper > 0 {
				// otherBlock is contained within thisBlock
				if otherIterator.HasNext() {
					otherBlock = otherIterator.Next()
					cmpLower = -1
				} else {
					result.addBlock(thisBlock)
					for thisIterator.HasNext() {
						result.addBlock(thisIterator.Next())
					}
					break
				}
			} else if cmpUpper < 0 {
				// thisBlock is contained within otherBlock
				if thisIterator.HasNext() {
					thisBlock = thisIterator.Next()
					cmpLower = 1
				} else {
					result.addBlock(otherBlock)
					for otherIterator.HasNext() {
						result.addBlock(otherIterator.Next())
					}
					break
				}
			} else {
				// both blocks are the same
				result.addBlock(thisBlock)
				if thisIterator.HasNext() {
					if otherIterator.HasNext() {
						thisBlock = thisIterator.Next()
						otherBlock = otherIterator.Next()
						cmpLower = compareLowerValues(thisBlock, otherBlock)
					} else {
						for thisIterator.HasNext() {
							result.addBlock(thisIterator.Next())
						}
						break
					}
				} else {
					for otherIterator.HasNext() {
						result.addBlock(otherIterator.Next())
					}
					break
				}
			}
		}
	}
	return result
}

// RemoveIntoNew produces a new containment trie that has the addresses in this collection that are not in the given collection.
func (coll *ContainmentTrieBase[T]) RemoveIntoNew(other *ContainmentTrieBase[T]) *ContainmentTrieBase[T] {
	if coll.IsEmpty() {
		return &ContainmentTrieBase[T]{}
	} else if other.IsEmpty() {
		return coll.Clone()
	}
	result := &ContainmentTrieBase[T]{}
	thisIterator := coll.PrefixBlockIterator()
	otherIterator := other.PrefixBlockIterator()
	thisBlock := thisIterator.Next()
	otherBlock := otherIterator.Next()
	cmpLower := compareLowerValues(thisBlock, otherBlock)
	var thisSeqRange *SequentialRange[T]
	for {
		if cmpLower < 0 {
			var upperComp int
			if thisSeqRange == nil {
				upperComp = compareUpperValues(thisBlock, otherBlock)
			} else {
				upperComp = compareUpperValues(thisSeqRange.GetUpper(), otherBlock)
			}
			if upperComp > 0 {
				// otherBlock is contained within thisBlock
				if thisSeqRange == nil {
					thisSeqRange = newSequRangeOrdered(thisBlock.GetLowerAndUpper())
				}
				remainingThis := thisSeqRange.Subtract(newSequRangeOrdered(otherBlock.GetLowerAndUpper()))
				result.AddSeqRange(remainingThis[0])
				if otherIterator.HasNext() {
					otherBlock = otherIterator.Next()
					thisSeqRange = remainingThis[1]
					cmpLower = compareLowerValues(thisSeqRange.GetLower(), otherBlock)
				} else {
					result.AddSeqRange(remainingThis[1])
					for thisIterator.HasNext() {
						result.addBlock(thisIterator.Next())
					}
					break
				}
			} else if upperComp == 0 {
				// otherBlock is contained within thisBlock
				if thisSeqRange == nil {
					thisSeqRange = newSequRangeOrdered(thisBlock.GetLowerAndUpper())
				}
				remainingThis := thisSeqRange.Subtract(newSequRangeOrdered(otherBlock.GetLowerAndUpper()))
				result.AddSeqRange(remainingThis[0])
				if thisIterator.HasNext() {
					thisBlock = thisIterator.Next()
					if otherIterator.HasNext() {
						otherBlock = otherIterator.Next()
						thisSeqRange = nil
						cmpLower = compareLowerValues(thisBlock, otherBlock)
					} else {
						result.addBlock(thisBlock)
						for thisIterator.HasNext() {
							result.addBlock(thisIterator.Next())
						}
						break
					}
				} else {
					break
				}
			} else {
				// otherBlock is above thisBlock
				if thisSeqRange == nil {
					result.addBlock(thisBlock)
				} else {
					result.AddSeqRange(thisSeqRange)
				}
				if thisIterator.HasNext() {
					thisBlock = thisIterator.Next()
					thisSeqRange = nil
					cmpLower = compareLowerValues(thisBlock, otherBlock)
				} else {
					break
				}
			}
		} else if cmpLower > 0 {
			if upperComp := compareUpperValues(thisBlock, otherBlock); upperComp <= 0 {
				// thisBlock is contained within otherBlock
				if thisIterator.HasNext() {
					thisBlock = thisIterator.Next()
				} else {
					break
				}
			} else {
				// this block is above otherBlock
				if otherIterator.HasNext() {
					otherBlock = otherIterator.Next()
					if thisSeqRange == nil {
						cmpLower = compareLowerValues(thisBlock, otherBlock)
					} else {
						cmpLower = compareLowerValues(thisSeqRange.GetLower(), otherBlock)
					}
				} else {
					result.addBlock(thisBlock)
					for thisIterator.HasNext() {
						result.addBlock(thisIterator.Next())
					}
					break
				}
			}
		} else {
			var cmpUpper int
			if thisSeqRange == nil {
				cmpUpper = compareUpperValues(thisBlock, otherBlock)
			} else {
				cmpUpper = compareUpperValues(thisSeqRange.GetUpper(), otherBlock)
			}
			if cmpUpper > 0 {
				if thisSeqRange == nil {
					thisSeqRange = newSequRangeOrdered(thisBlock.GetLowerAndUpper())
				}
				remainingThis := thisSeqRange.Subtract(newSequRangeOrdered(otherBlock.GetLowerAndUpper()))
				if otherIterator.HasNext() {
					otherBlock = otherIterator.Next()
					thisSeqRange = remainingThis[0]
					cmpLower = compareLowerValues(thisSeqRange.GetLower(), otherBlock)
				} else {
					result.AddSeqRange(remainingThis[0])
					for thisIterator.HasNext() {
						result.addBlock(thisIterator.Next())
					}
					break
				}
			} else if cmpUpper < 0 {
				// thisBlock is contained within otherBlock
				if thisIterator.HasNext() {
					thisBlock = thisIterator.Next()
					cmpLower = 1
					thisSeqRange = nil
				} else {
					break
				}
			} else {
				// both blocks are the same
				if thisIterator.HasNext() {
					if otherIterator.HasNext() {
						thisBlock = thisIterator.Next()
						otherBlock = otherIterator.Next()
						thisSeqRange = nil
						cmpLower = compareLowerValues(thisBlock, otherBlock)
					} else {
						result.addBlock(thisIterator.Next())
						for thisIterator.HasNext() {
							result.addBlock(thisIterator.Next())
						}
						break
					}
				} else {
					break
				}
			}
		}
	}
	return result
}

// IntersectIntoNew produces a new containment tries that is the intersection of this collection with the given collection.
func (coll *ContainmentTrieBase[T]) IntersectIntoNew(other *ContainmentTrieBase[T]) *ContainmentTrieBase[T] {
	result := &ContainmentTrieBase[T]{}
	if coll.IsEmpty() || other.IsEmpty() {
		return result
	}
	thisIterator := coll.PrefixBlockIterator()
	otherIterator := other.PrefixBlockIterator()
	thisBlock := thisIterator.Next()
	otherBlock := otherIterator.Next()
	cmpLower := compareLowerValues(thisBlock, otherBlock)
	for {
		if cmpLower < 0 {
			if compareUpperValues(thisBlock, otherBlock) >= 0 {
				// otherBlock is contained within thisBlock
				result.addBlock(otherBlock)
				if otherIterator.HasNext() {
					otherBlock = otherIterator.Next()
				} else {
					break
				}
			} else {
				// otherBlock is above thisBlock
				if thisIterator.HasNext() {
					thisBlock = thisIterator.Next()
					cmpLower = compareLowerValues(thisBlock, otherBlock)
				} else {
					break
				}
			}
		} else if cmpLower > 0 {
			if compareUpperValues(thisBlock, otherBlock) <= 0 {
				// thisBlock is contained within otherBlock
				result.addBlock(thisBlock)
				if thisIterator.HasNext() {
					thisBlock = thisIterator.Next()
				} else {
					break
				}
			} else {
				// thisBlock is above otherBlock
				if otherIterator.HasNext() {
					otherBlock = otherIterator.Next()
					cmpLower = compareLowerValues(thisBlock, otherBlock)
				} else {
					break
				}
			}
		} else if cmpUpper := compareUpperValues(thisBlock, otherBlock); cmpUpper > 0 {
			// otherBlock is contained within thisBlock, with lower values equal
			result.addBlock(otherBlock)
			if otherIterator.HasNext() {
				otherBlock = otherIterator.Next()
				cmpLower = -1
			} else {
				break
			}
		} else if cmpUpper < 0 {
			// thisBlock is contained within otherBlock, with lower values equal
			result.addBlock(thisBlock)
			if thisIterator.HasNext() {
				thisBlock = thisIterator.Next()
				cmpLower = 1
			} else {
				break
			}
		} else {
			// both blocks are the same, add either one
			result.addBlock(thisBlock)
			if thisIterator.HasNext() && otherIterator.HasNext() {
				thisBlock = thisIterator.Next()
				otherBlock = otherIterator.Next()
				cmpLower = compareLowerValues(thisBlock, otherBlock)
			} else {
				break
			}
		}
	}
	return result
}

// ContainsOther returns whether this containment trie contains all addresses in the given containment trie
func (coll *ContainmentTrieBase[T]) ContainsOther(other *ContainmentTrieBase[T]) bool {
	if other.IsEmpty() {
		return true
	} else if coll.IsEmpty() || coll.GetCount().Cmp(other.GetCount()) < 0 {
		return false
	}
	thisIterator := coll.PrefixBlockIterator()
	otherIterator := other.PrefixBlockIterator()
	thisBlock := thisIterator.Next()
	otherBlock := otherIterator.Next()
	cmpLower := compareLowerValues(thisBlock, otherBlock)
	for {
		if cmpLower < 0 {
			if compareUpperValues(thisBlock, otherBlock) >= 0 {
				// otherBlock is contained within thisBlock
				if otherIterator.HasNext() {
					otherBlock = otherIterator.Next()
				} else {
					return true
				}
			} else {
				// otherBlock is above thisBlock
				if thisIterator.HasNext() {
					thisBlock = thisIterator.Next()
					cmpLower = compareLowerValues(thisBlock, otherBlock)
				} else {
					return false
				}
			}
		} else if cmpLower > 0 {
			return false
		} else {
			if cmpUpper := compareUpperValues(thisBlock, otherBlock); cmpUpper > 0 {
				// otherBlock is contained within thisBlock
				if otherIterator.HasNext() {
					otherBlock = otherIterator.Next()
					cmpLower = -1
				} else {
					return true
				}
			} else if cmpUpper < 0 {
				// we do not need to go further here
				// this is because, even though it is possible another block from thisIterator might be consecutive to the current one,
				// it is not possible that blocks from thisIterator cover all of otherBlock, otherwise they would be merged together into a single block,
				// just like otherBlock
				return false
			} else {
				// both blocks are the same
				if !thisIterator.HasNext() {
					return !otherIterator.HasNext()
				} else if otherIterator.HasNext() {
					thisBlock = thisIterator.Next()
					otherBlock = otherIterator.Next()
					cmpLower = compareLowerValues(thisBlock, otherBlock)
				} else {
					return true
				}
			}
		}
	}
}

// OverlapsOther returns whether there is any overlap with the given containment trie
func (coll *ContainmentTrieBase[T]) OverlapsOther(other *ContainmentTrieBase[T]) bool {
	if coll.IsEmpty() || other.IsEmpty() {
		return false
	}
	thisIterator := coll.PrefixBlockIterator()
	otherIterator := other.PrefixBlockIterator()
	thisBlock := thisIterator.Next()
	otherBlock := otherIterator.Next()
	for {
		if cmpLower := compareLowerValues(thisBlock, otherBlock); cmpLower < 0 {
			if compareUpperValues(thisBlock, otherBlock) >= 0 {
				return true
			} else if thisIterator.HasNext() { // otherBlock is above thisBlock
				thisBlock = thisIterator.Next()
			} else {
				return false
			}
		} else if cmpLower > 0 {
			if compareUpperValues(thisBlock, otherBlock) <= 0 {
				return true
			} else if otherIterator.HasNext() { // thisBlock is above otherBlock
				otherBlock = otherIterator.Next()
			} else {
				return false
			}
		} else {
			return true
		}
	}
}

type (
	IPAddressContainmentTrie   = ContainmentTrieBase[*IPAddress]
	IPv4AddressContainmentTrie = ContainmentTrieBase[*IPv4Address]
	IPv6AddressContainmentTrie = ContainmentTrieBase[*IPv6Address]
)
