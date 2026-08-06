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
	"strings"

	"github.com/seancfoley/bintree/tree"
)

// SequentialRangeList is an IP address collection backed by a sorted list of sequential IP address ranges.
// It consists of a series of IPAddressSeqRange instances to describe a range of addresses that is non-sequential, if more than one SequentialRange is in the list.
// As addresses are added and removed, the list of SequentialRange instances are adjusted,
// so that it is always the minimal list of sequential ranges that includes all the specified addresses.
//
// It is one of the efficient options provided by this library to maintain sets of individual IP addresses,
// the other option being [ContainmentTrieBase], representing collections backed by IP address tries.
//
// Lookups of addresses, subnets, or sequential ranges the list are performed by binary search on the list of sequential address ranges.
//
// Finding the address at a specific index in the list is performed by binary search on the sizes of the individual sequential ranges in the list.
//
// The maximum number of addresses in the sequential range list is unlimited.
// However, the maximum number of disconnected sequential ranges is limited to the maximum size of an array, which is limited by the max value of an integer.
//
// With some data-sets, this collection type will have better search performance than a trie or containment trie due to improved cache coherency with many CPU processors.
//
// IP address collection equality with another IP address collection is determined by the contents of the collections.
// A SequentialRangeList is equal to a ContainmentTrieBase if the collections contain the same set of individual addresses.
// The same is true for IP address aggregation equality.
//
// A SequentialRangeList may contain either IPv6 addresses, or IPv4 addresses, but not both at the same time.
// An attempt to add an address when the collection already contains an address of a different version will panic.
// However, once such a collection becomes empty again, it can accept either an IPv6 address or IPv4 address once more.
type SequentialRangeList[T ipAddressTypeConstraint[T]] struct {
	ranges []SequentialRange[T]

	rangeSizes []*big.Int

	changeTracker tree.ChangeTracker
}

// NewSequentialRangeList creates a new sequential range list preallocated to hold the given number of elements.
func NewSequentialRangeList[T ipAddressTypeConstraint[T]](initialCapacity int) *SequentialRangeList[T] {
	return &SequentialRangeList[T]{
		ranges:     make([]SequentialRange[T], 0, initialCapacity),
		rangeSizes: make([]*big.Int, 0, initialCapacity),
	}
}

// Contains returns true if and only if this list contains all the individual addresses in the given address or subnet
func (list *SequentialRangeList[T]) Contains(address AddressType) bool {
	addr, isNil, ok := ConvertAddressTypeCheckNil[T](address)
	return ok && !isNil && list != nil && list.IndexOfSeqRangeContainingAddress(addr) >= 0
}

// ContainsAddress returns true if and only if this list contains all the individual addresses in the given address or subnet
func (list *SequentialRangeList[T]) ContainsAddress(address T) bool {
	return list != nil && !isNilPtr(address) && list.IndexOfSeqRangeContainingAddress(address) >= 0
}

// IndexOfSeqRangeContainingAddress returns the index of the lowest sequential range in this list containing some elements of the given address or subnet,
// if this list contains all the addresses in the given list,
//
// If this list does not contain all the addresses in the given subnet, a negative number is returned.
// If the given address or subnet contains addresses of a different version, then -1 is returned.
// Otherwise, the returned negative number is -(insertion index) - 1, where the insertion index is the lowest index in this list where addresses would have to be inserted in order to contain the given address or subnet.
//
// This means that the returned value will be >= 0 if and only if the ranges in the list contain the given address or subnet.
func (list *SequentialRangeList[T]) IndexOfSeqRangeContainingAddress(address T) int {
	if len(list.ranges) == 0 || !versionsMatch(list.ranges[0].GetLower(), address) {
		return -1
	} else if address.IsSequential() {
		return list.containsSequentialAddr(address, 0)
	}
	// an unusual case is non-sequential addresses fitting into separate ranges
	// eg 1.2.3-5.0 and the disjunct range being [1.2.3.0 -> 1.2.4.0], [1.2.5.0 -> 1.2.6.0]
	iterator := address.SequentialBlockIterator()
	next := iterator.Next()
	result := list.containsSequentialAddr(next, 0)
	if result >= 0 {
		for index := result; iterator.HasNext(); {
			next = iterator.Next()
			index = list.containsSequentialAddr(next, index)
			if index < 0 {
				return index
			}
		}
	}
	return result
}

// Contains returns whether this list contains all addresses in the given containment trie.
func (list *SequentialRangeList[T]) ContainsContainmentTrie(collection *ContainmentTrieBase[T]) bool {
	return list != nil && collection != nil && list.IndexOfSeqRangeContainingContainmentTrie(collection) >= 0
}

// IndexOfSeqRangeContainingContainmentTrie returns the index of the lowest sequential range in this list containing some elements of the containment trie,
// if this list contains all the addresses in the given containment trie, .
//
// If this list does not contain all the addresses in the given containment trie, a negative number is returned.
// If the given trie contains addresses of a different version, then -1 is returned.
// Otherwise, the returned negative number is -(insertion index) - 1, where the insertion index is the lowest index in this list where addresses would have to be inserted in order to contain the given containment trie.
//
// It returns 0 if both this list and the given containment trie are empty.
//
// This means that the returned value will be >= 0 if and only if the ranges in the list contain the addresses in the given containment trie.
func (list *SequentialRangeList[T]) IndexOfSeqRangeContainingContainmentTrie(collection *ContainmentTrieBase[T]) int {
	return list.IndexOfSeqRangeContainingTrie(&collection.trie.Trie)
}

// ContainsTrie returns whether this list contains all added individual addresses and added prefix blocks in the given trie.
func (list *SequentialRangeList[T]) ContainsTrie(trie *Trie[T]) bool {
	return list != nil && trie != nil && list.IndexOfSeqRangeContainingTrie(trie) >= 0
}

// IndexOfSeqRangeContainingTrie returns the index of the lowest sequential range in this list containing some elements of the trie,
// if this list contains all the added individual ddresses and the added prefix blocks in the given trie,
//
// If this list does not contain all the addresses in the given trie, a negative number is returned.
// If the given trie contains addresses of a different version, then -1 is returned.
// Otherwise, the returned negative number is -(insertion index) - 1, where the insertion index is the lowest index in this list where addresses would have to be inserted in order to contain the given trie.
//
// It returns 0 if both this list and the given trie are empty.
//
// This means that the returned value will be >= 0 if and only if the ranges in the list contain the addresses in the given trie.
func (list *SequentialRangeList[T]) IndexOfSeqRangeContainingTrie(trie *Trie[T]) int {
	ranges := list.ranges
	if len(ranges) == 0 {
		if trie.isEmpty() {
			return 0
		}
		return -1
	} else if list.getCount(len(ranges)).CmpAbs(trie.GetMatchingAddressCount()) < 0 {
		return -1
	}
	// this iterator cannot be the collection iterator which goes by individual address, it must be the enclosed trie's iterator
	iterator := trie.iterator()
	next := iterator.Next()
	if !versionsMatch(ranges[0].GetLower(), next) {
		return -1
	}
	result := list.containsSequentialAddr(next, 0)
	if result >= 0 {
		for index := result; iterator.HasNext(); {
			next = iterator.Next()
			index = list.containsSequentialAddr(next, index)
			if index < 0 {
				return index
			}
		}
	}
	return result
}

// ContainsOther returns whether this sequential range list contains all addresses in the given sequential range list
func (list *SequentialRangeList[T]) ContainsOther(otherList *SequentialRangeList[T]) bool {
	return otherList.IsEmpty() || (list != nil && list.IndexOfSeqRangeContainingSeqRangeList(otherList) >= 0)
}

// IndexOfSeqRangeContainingSeqRangeList returns the index of the lowest sequential range in this list containing some elements of the given sequential range list,
// if this list contains all the addresses in the given list.
//
// If this list does not contain all the addresses in the given list, a negative number is returned.
// If the given list contains addresses of a different version, then -1 is returned.
// Otherwise, the returned negative number is -(insertion index) - 1, where the insertion index is the lowest index in this list where addresses would have to be inserted in order to contain the given list.
//
// If the given list is empty, then 0 is returned.
//
// This means that the returned value will be >= 0 if and only if the ranges in the list contain the given list.
func (list *SequentialRangeList[T]) IndexOfSeqRangeContainingSeqRangeList(otherList *SequentialRangeList[T]) int {
	otherRangeSize := len(otherList.ranges)
	if otherRangeSize == 0 {
		return 0
	}
	ranges := list.ranges
	if len(ranges) == 0 || !versionsMatch(ranges[0].GetLower(), otherList.GetSeqRange(0).GetLower()) || list.getCount(len(ranges)).CmpAbs(otherList.GetCount()) < 0 {
		return -1
	}
	other := otherList.GetSeqRange(0)
	result := list.containsSequential(other, 0)
	if result >= 0 {
		for i, lowerIndex := 1, result; i < otherRangeSize; i++ {
			other = otherList.GetSeqRange(i)
			lowerIndex = list.containsSequential(other, lowerIndex)
			if lowerIndex < 0 {
				return lowerIndex
			}
		}
	}
	return result
}

// ContainsRange returns true if and only if all the addresses in the given sequential range are also in this sequential range list.
// Implements the IPAddressAggregation interface.
func (list *SequentialRangeList[T]) ContainsRange(rng IPAddressSeqRangeType) bool {
	r, isNil, ok := ConvertRangeTypeCheckNil[T](rng)
	return ok && !isNil && list != nil && list.IndexOfSeqRangeContainingSeqRange(r) >= 0
}

// ContainsSeqRange returns true if and only if all the addresses in the given sequential range are also in this sequential range list.
// Implements the IPAddressCollAddrConstraint interface.
func (list *SequentialRangeList[T]) ContainsSeqRange(seqRange *SequentialRange[T]) bool {
	return list != nil && seqRange != nil && list.IndexOfSeqRangeContainingSeqRange(seqRange) >= 0
}

// IndexOfSeqRangeContainingSeqRange eturns the index of the sequential range in this list containing the given sequential range,
// if this list contains the given sequential range.
//
// If this list does not contain the given sequential range, a negative number is returned.
// If the given list contains addresses of a different version, then -1 is returned.
// Otherwise, the returned negative number is -(insertion index) - 1, where the insertion index is the lowest index in this list where addresses from the given sequential range would have to be inserted in order to contain the given sequential range.
//
// This means that the returned value will be >= 0 if and only if a range in the list contain the given sequential range.
func (list *SequentialRangeList[T]) IndexOfSeqRangeContainingSeqRange(seqRange *SequentialRange[T]) int {
	ranges := list.ranges
	if len(ranges) == 0 || !versionsMatch(ranges[0].GetLower(), seqRange.GetLower()) {
		return -1
	}
	return list.containsSequential(seqRange, 0)
}

// negative value means not contains, positive value gives the index of the containing sequential range
func (list *SequentialRangeList[T]) containsSequential(seqRange *SequentialRange[T], startIndex int) int {
	lowerIndex := list.binarySearchLower(startIndex, seqRange.GetLower())
	if lowerIndex < 0 || compareUpperValues(seqRange.GetUpper(), list.ranges[lowerIndex].GetUpper()) <= 0 {
		return lowerIndex
	}
	return -(lowerIndex + 1)
}

// negative value means not contains, positive value gives the index of the containing sequential range
func (list *SequentialRangeList[T]) containsSequentialAddr(addr T, startIndex int) int {
	lowerIndex := list.binarySearchLower(startIndex, addr)
	if lowerIndex < 0 || !addr.IsMultiple() || compareUpperValues(addr, list.ranges[lowerIndex].GetUpper()) <= 0 {
		return lowerIndex
	}
	return -(lowerIndex + 1)
}

// OverlapsRange returns true if and only the given sequential range contains at least one address that is also in this sequential range list.
// Implements the IPAddressAggregation interface.
func (list *SequentialRangeList[T]) OverlapsRange(rng IPAddressSeqRangeType) bool {
	r, isNil, ok := ConvertRangeTypeCheckNil[T](rng)
	return ok && !isNil && list != nil && list.IndexOfSeqRangeOverlappingSeqRange(r) >= 0
}

// OverlapsSeqRange returns true if and only if the given sequential range contains at least one address that is also in this sequential range list.
// Implements the IPAddressCollAddrConstraint interface.
func (list *SequentialRangeList[T]) OverlapsSeqRange(seqRange *SequentialRange[T]) bool {
	return list != nil && seqRange != nil && list.IndexOfSeqRangeOverlappingSeqRange(seqRange) >= 0
}

// IndexOfSeqRangeOverlappingSeqRange returns the index of the lowest sequential range in this list overlapping the given sequential range, if this list overlaps with the given sequential range.
//
// If this list does not overlap with the given sequential range, a negative number is returned.
// If the given sequential range contains addresses that are a different version from those in this list, then -1 is returned.
// Otherwise, the returned negative number is -(insertion index) - 1, where the insertion index is the lowest index in this list where addresses could be inserted to result in overlap.
//
// This means that the returned value will be >= 0 if and only if the ranges in this list overlap with the given sequential range.
func (list *SequentialRangeList[T]) IndexOfSeqRangeOverlappingSeqRange(seqRange *SequentialRange[T]) int {
	lower, upper := seqRange.GetLowerAndUpper()
	return list.doOverlapsSequential(seqRange, 0, lower, upper)
}

// OverlapsAddr returns true if and only the given individual address or subnet contains at least one individual address that is also in this sequential range list.
// Implements the IPAddressAggregation interface.
func (list *SequentialRangeList[T]) OverlapsAddr(address AddressType) bool {
	addr, isNil, ok := ConvertAddressTypeCheckNil[T](address)
	return ok && !isNil && list != nil && list.IndexOfSeqRangeOverlappingAddress(addr) >= 0
}

// OverlapsAddress returns true if and only the given individual address or subnet contains at least one individual address that is also in this sequential range list.
// Implements the IPAddressCollAddrConstraint interface.
func (list *SequentialRangeList[T]) OverlapsAddress(address T) bool {
	return list != nil && !isNilPtr(address) && list.IndexOfSeqRangeOverlappingAddress(address) >= 0
}

// IndexOfSeqRangeOverlappingAddress returns the index of the lowest sequential range in this list overlapping addresses from the given address or subnet,
// if this list overlaps with the given address or subnet.
//
// If this list does not overlap with the given address or subnet, a negative number is returned.
// If the given address or subnet is a different version, then -1 is returned.
// Otherwise, the returned negative number is -(insertion index) - 1, where the insertion index is the lowest index in this list where addresses could be inserted to result in overlap.
//
// This means that the returned value will be >= 0 if and only if the ranges in this list overlap with the given address or subnet.
func (list *SequentialRangeList[T]) IndexOfSeqRangeOverlappingAddress(address T) int {
	ranges := list.ranges
	if len(ranges) == 0 || !versionsMatch(ranges[0].GetLower(), address) {
		return -1
	} else if address.IsSequential() {
		return list.overlapsSequential(address)
	}
	// an unusual case is non-sequential addresses fitting into separate ranges
	// eg 1.2.3-5.0 and the disjunct range being [1.2.3.0 -> 1.2.4.0], [1.2.5.0 -> 1.2.6.0]
	iterator := address.SequentialBlockIterator()
	next := iterator.Next()
	result := list.overlapsSequential(next)
	if result < 0 {
		for index := result; iterator.HasNext(); {
			index = -(index + 1)
			next = iterator.Next()
			index = list.doOverlapsSequential(next, index, next, next)
			if index >= 0 {
				return index
			} else if -(index + 1) >= len(ranges) {
				break
			}
		}
	}
	return result
}

func (list *SequentialRangeList[T]) overlapsSequential(address T) int {
	return list.doOverlapsSequential(address, 0, address, address)
}

// OverlapsOther returns whether there is any overlap with the given sequential range list
func (list *SequentialRangeList[T]) OverlapsOther(otherList *SequentialRangeList[T]) bool {
	if list == nil || otherList == nil {
		return false
	}
	ranges := list.ranges
	thisCount := len(ranges)
	otherCount := len(otherList.ranges)
	if thisCount == 0 || otherCount == 0 || !versionsMatch(ranges[0].GetLower(), otherList.GetSeqRange(0).GetLower()) {
		return false
	}
	// reduce the number of binary searches by iterating over the smaller range list
	if thisCount < otherCount {
		return otherList.doOverlaps(list) >= 0
	}
	return list.doOverlaps(otherList) >= 0
}

// IndexOfSeqRangeOverlappingSeqRangeList returns the index of the lowest sequential range in this list overlapping some elements of the given sequential range list,
// if this list overlaps with addresses in the given list.
//
// If this list does not overlap with the given list, a negative number is returned.
// If the given list is empty or contains addresses of a different version, then -1 is returned.
// Otherwise, the returned negative number is -(insertion index) - 1, where the insertion index is the lowest index in this list where addresses could be inserted to result in overlap.
//
// This means that the returned value will be >= 0 if and only if the ranges in this list overlap with the given list.
func (list *SequentialRangeList[T]) IndexOfSeqRangeOverlappingSeqRangeList(otherList *SequentialRangeList[T]) int {
	ranges := list.ranges
	if len(ranges) == 0 || len(otherList.ranges) == 0 || !versionsMatch(ranges[0].GetLower(), otherList.GetSeqRange(0).GetLower()) {
		return -1
	}
	return list.doOverlaps(otherList)
}

// smallerList does not need to be smaller, it's just more efficient if that's the case
func (list *SequentialRangeList[T]) doOverlaps(smallerList *SequentialRangeList[T]) int {
	// Check each IPAddressSeqRange in smallerList for overlap.
	// We take advantage of the ordering of the list of IPAddressSeqRanges,
	// for each new binary search, the lower end can be the index returned by the previous search
	other := smallerList.GetSeqRange(0)
	lower, upper := other.GetLowerAndUpper()
	result := list.doOverlapsSequential(other, 0, lower, upper)
	ranges := list.ranges
	if result < 0 {
		for i, lowerIndex := 0, result; i < smallerList.GetSeqRangeCount(); i++ {
			lowerIndex = -(lowerIndex + 1)
			other = smallerList.GetSeqRange(i)
			lower, upper = other.GetLowerAndUpper()
			lowerIndex = list.doOverlapsSequential(other, lowerIndex, lower, upper)
			if lowerIndex >= 0 {
				return lowerIndex
			} else if lowerIndex >= len(ranges) {
				break
			}
		}
	}
	return result
}

// lowerCompare is an address whose lower values can be used to represent the lower values of the range
// upperCompare is an address whose upper values can be used to represent the upper values of the range
// In practice, this means that for a subnet, the subnet itself can be used,
// but for a sequential range, the lower or upper range boundary must be used.
//
// If there is overlap, returns the non-negative index where the lowest overlap occurs
// If no overlap, returns the index where the given range would be inserted
func (list *SequentialRangeList[T]) doOverlapsSequential(rng IPAddressRange, lowerBound int, lowerCompare, upperCompare T) int {
	lowerIndex := list.binarySearchLower(lowerBound, lowerCompare)
	if lowerIndex < 0 {
		if rng.IsMultiple() {
			lowerIndexAdjusted := -(lowerIndex + 1)
			upperIndex := list.binarySearchUpper(lowerIndexAdjusted, upperCompare)
			if upperIndex >= 0 || upperIndex != lowerIndex {
				lowerIndex = lowerIndexAdjusted
			}
		}
	}
	return lowerIndex
}

func nilPtr[T any]() (t T) {
	return
}

func isNilPtr[T AddressType](addr T) bool {
	return any(addr) == any(nilPtr[T]())
}

// Lower returns the highest address in the collection strictly less than all addresses in the given address or subnet.
func (list *SequentialRangeList[T]) Lower(addr T) T {
	ranges := list.ranges
	if len(ranges) == 0 || !versionsMatch(ranges[0].GetLower(), addr) {
		return nilPtr[T]()
	}
	index, onLowerRangeBoundary, _ := list.binarySearchLowerAddrWithBoundaries(addr)
	if index < 0 {
		// not in the list
		indexAdjusted := -(index + 1)
		if indexAdjusted == 0 {
			return nilPtr[T]()
		}
		return ranges[indexAdjusted-1].GetUpper()
	}
	// is in the range list
	if onLowerRangeBoundary {
		// lowest value in the range
		if index == 0 {
			return nilPtr[T]()
		}
		return ranges[index-1].GetUpper()
	}
	return addr.DecrementSingle()
}

// Floor returns the highest address in the collection less than or equal to the lowest address in the given address or subnet.
func (list *SequentialRangeList[T]) Floor(addr T) T {
	ranges := list.ranges
	if len(ranges) == 0 || !versionsMatch(ranges[0].GetLower(), addr) {
		return nilPtr[T]()
	}
	index := list.binarySearchLowerAddr(addr)
	if index < 0 {
		// not in the list
		indexAdjusted := -(index + 1)
		if indexAdjusted == 0 {
			return nilPtr[T]()
		}
		return ranges[indexAdjusted-1].GetUpper()
	}
	// is in the range list, just return it
	return addr.WithoutPrefixLen().GetLower()
}

// Ceiling returns the lowest address in the collection greater than or equal to the highest address in the given address or subnet.
func (list *SequentialRangeList[T]) Ceiling(addr T) T {
	ranges := list.ranges
	if len(ranges) == 0 || !versionsMatch(ranges[0].GetLower(), addr) {
		return nilPtr[T]()
	}
	index := list.binarySearchUpperAddr(addr)
	if index < 0 {
		// not in the list
		indexAdjusted := -(index + 1)
		if indexAdjusted == len(ranges) {
			return nilPtr[T]()
		}
		return ranges[indexAdjusted].GetLower()
	}
	// is in the range list, just return it
	return addr.WithoutPrefixLen().GetUpper()
}

// Higher returns the lowest address in the collection strictly greater than all addresses in the given address or subnet.
func (list *SequentialRangeList[T]) Higher(addr T) T {
	ranges := list.ranges
	if len(ranges) == 0 || !versionsMatch(ranges[0].GetLower(), addr) {
		return nilPtr[T]()
	}
	index, _, onUpperRangeBoundary := list.binarySearchUpperAddrWithBoundaries(addr)
	if index < 0 {
		// not in the list
		indexAdjusted := -(index + 1)
		if indexAdjusted == len(ranges) {
			return nilPtr[T]()
		}
		return ranges[indexAdjusted].GetLower()
	}
	// is in the range list
	if onUpperRangeBoundary {
		// highest value in the range
		nextRangeIndex := index + 1
		if nextRangeIndex == len(ranges) {
			return nilPtr[T]()
		}
		return ranges[nextRangeIndex].GetLower()
	}
	// somewhere in the middle
	return addr.IncrementBoundarySingle()
}

// Returns a new IPAddressSeqRangeList comprising all the addresses not contained in this list.
//
// If this list is empty and is not restricted to a single IP version of IPv4 or IPv6,
// then the IP version is ambiguous, so the complement is indeterminate, in which case nil is returned.
func (list *SequentialRangeList[T]) ComplementIntoNew() *SequentialRangeList[T] {
	ranges := list.ranges
	if len(ranges) == 0 {
		var t T
		network := t.GetIPNetwork()
		if network == nil {
			return nil
		}
		newList := NewSequentialRangeList[T](1)
		zero, maxAddr := network.GetBoundaryAddresses()
		newList.ranges = append(newList.ranges, *newSequRangeUnchecked(zero, maxAddr, true))
		return newList
	}
	var newRanges []SequentialRange[T]
	firstRng := ranges[0]
	last := ranges[len(ranges)-1]
	hasMax := !last.IncludesMax()
	firstUpper, previousUpper := firstRng.GetLowerAndUpper()
	network := firstRng.GetLower().GetIPNetwork()
	zero, maxAddr := network.GetBoundaryAddresses()
	var newList *SequentialRangeList[T]
	if !firstRng.IncludesZero() {
		if hasMax {
			newList = NewSequentialRangeList[T](len(ranges) + 1)
		} else {
			newList = NewSequentialRangeList[T](len(ranges))
		}
		newRanges = append(newList.ranges, *newSequRangeCheckSize(zero, firstUpper.DecrementSingle()))
	} else {
		if hasMax {
			newList = NewSequentialRangeList[T](len(ranges))
		} else {
			newList = NewSequentialRangeList[T](len(ranges) - 1)
		}
		newRanges = newList.ranges
	}
	for i := 1; i < len(ranges); i++ {
		lower, upper := ranges[i].GetLowerAndUpper()
		newRanges = append(newRanges, *newSequRangeCheckSize(previousUpper.IncrementSingle(), lower.DecrementSingle()))
		previousUpper = upper
	}
	if hasMax {
		newRanges = append(newRanges, *newSequRangeCheckSize(previousUpper.IncrementSingle(), maxAddr))
	}
	newList.ranges = newRanges
	return newList
}

// RemoveIntoNew produces a new IPAddressSeqRangeList that has the addresses in this list that are not in the given list.
// Neither this list nor the given list are altered, instead a new list is created and returned.
func (list *SequentialRangeList[T]) RemoveIntoNew(otherList *SequentialRangeList[T]) (result *SequentialRangeList[T]) {
	result = NewSequentialRangeList[T](len(list.ranges) + len(otherList.ranges))
	list.remove(otherList, result)
	return
}

func (list *SequentialRangeList[T]) remove(otherList, result *SequentialRangeList[T]) {
	ranges := list.ranges
	if len(ranges) > 0 { // something to remove
		if len(otherList.ranges) == 0 || !versionsMatch(ranges[0].GetLower(), otherList.ranges[0].GetLower()) { // not removing anything
			result.ranges = append(result.ranges, ranges...)
			result.rangeSizes = append(result.rangeSizes, list.rangeSizes...)
		} else {
			list.removeRanges(otherList, result)
		}
	}
}

func (list *SequentialRangeList[T]) removeRanges(otherList, result *SequentialRangeList[T]) {
	ranges := list.ranges
	resultRanges := result.ranges
	otherRanges := otherList.ranges
	currentIndex := 0
	var pending pendingRange[T]
	for i := 0; i < len(otherRanges); i++ {
		otherRange := &otherRanges[i]
		currentIndex, resultRanges = list.removeRange(otherRange, resultRanges, &pending, currentIndex)
		if currentIndex >= len(ranges) {
			break
		}
	}
	// If there is a pending range, then the range at upper index is in the pending range,
	// so the ranges ar upperIndex + 1 must be added after the pending range.
	// Otherwise the ranges at upperIndex must be added.
	if !pending.isEmpty() {
		currentIndex++
		resultRanges = append(resultRanges, *newSequRangeCheckSize(pending.lower, pending.upper))
		pending.clear()
	}
	if currentIndex < len(ranges) {
		resultRanges = append(resultRanges, ranges[currentIndex:]...)
	}
	result.ranges = resultRanges
}

func (list *SequentialRangeList[T]) removeRange(seqRange *SequentialRange[T], result []SequentialRange[T], pending *pendingRange[T], index int) (int, []SequentialRange[T]) {
	lower, upper := seqRange.GetLower(), seqRange.GetUpper()
	lowerIndex, onLowerRangeBoundary, onUpperRangeBoundary := list.binarySearchLowerWithBoundaries(index, lower)
	splitLower := lowerIndex >= 0
	if !splitLower {
		lowerIndex = -(lowerIndex + 1)
	}
	var (
		upperIndex                int
		splitUpper                bool
		upperOnUpperRangeBoundary bool
	)
	ranges := list.ranges
	if seqRange.IsMultiple() && lowerIndex != len(ranges) {
		upperIndex, _, upperOnUpperRangeBoundary = list.binarySearchLowerWithBoundaries(lowerIndex, upper)
		splitUpper = upperIndex >= 0
		if !splitUpper {
			upperIndex = -(upperIndex + 1)
		}
	} else {
		upperOnUpperRangeBoundary = onUpperRangeBoundary
		upperIndex = lowerIndex
		splitUpper = splitLower
	}
	if pending.isEmpty() {
		// add ranges following the last range and preceding this one
		if lowerIndex > index {
			result = append(result, ranges[index:lowerIndex]...)
		}
		if splitLower {
			existingRange := &ranges[lowerIndex]
			splitLower = compareLowerValues(existingRange.GetLower(), lower) != 0
		}
	} else {
		// check if the pending range overlaps with this one, creating a large unified pending range
		if lowerIndex > pending.existingRangeUpperIndex {
			// there is no overlap, add the pending range, and then add the succeeding ranges that precede this one
			result = append(result, *newSequRangeCheckSize(pending.lower, pending.upper))
			pending.clear()
			// at this time, index is the pending range upper index
			index++
			if index < lowerIndex {
				result = append(result, ranges[index:lowerIndex]...)
			}
			if splitLower {
				splitLower = !onLowerRangeBoundary
			}
		}
	}
	if splitUpper {
		if upperOnUpperRangeBoundary {
			splitUpper = false
			upperIndex++
		}
	}
	if lowerIndex < upperIndex { // spans at least one existing range
		if !pending.isEmpty() {
			// came in with: y1     x1  removed x2  pending  y2
			// here we have: y1     x1  removed x2  pending lower  y2  upper
			// we want to add x2 to lower to result
			result = append(result, *newSequRangeCheckSize(pending.lower, lower.DecrementSingle()))
			pending.clear()
		} else if splitLower {
			result = append(result, *ranges[lowerIndex].lowerSplit(lower))
		}
		if splitUpper {
			// new pending
			pending.from = seqRange
			pending.lower = upper.IncrementSingle()
			pending.existingRangeUpperIndex, pending.lowerIndex = upperIndex, upperIndex
			pending.upper = ranges[upperIndex].GetUpper()
		}
	} else { // spans 0 or 1 existing range
		if !pending.isEmpty() {
			// came in with: y1     x1  removed x2  pending  y2
			// here we have: y1     x1  removed x2  pending lower upper y2
			// we want to add x2 to lower to result
			result = append(result, *newSequRangeCheckSize(pending.lower, lower.DecrementSingle()))
			// still pending in the same range
			pending.lower = seqRange.GetUpper().IncrementSingle()
		} else {
			if splitLower {
				result = append(result, *ranges[lowerIndex].lowerSplit(lower))
			}
			if splitUpper {
				// we want to set pending to upper to y2
				pending.from = seqRange
				pending.lower = seqRange.GetUpper().IncrementSingle()
				pending.existingRangeUpperIndex, pending.lowerIndex = lowerIndex, lowerIndex
				pending.upper = ranges[upperIndex].GetUpper()
			} // else range does not intersect with anything
		}
	}
	return upperIndex, result
}

// IntersectIntoNew produces a new IPAddressSeqRangeList that is the intersection of this list with the given list.
// Neither this list nor the given list are altered, instead a new intersection list is created and returned.
func (list *SequentialRangeList[T]) IntersectIntoNew(otherList *SequentialRangeList[T]) (result *SequentialRangeList[T]) {
	result = NewSequentialRangeList[T](len(list.ranges) + len(otherList.ranges))
	list.intersect(otherList, result)
	return result
}

func (list *SequentialRangeList[T]) intersect(otherList, result *SequentialRangeList[T]) {
	ranges := list.ranges
	otherRanges := otherList.ranges
	if len(ranges) == 0 || len(otherRanges) == 0 {
		return
	} else if !versionsMatch(ranges[0].GetLower(), otherRanges[0].GetLower()) {
		return
	}
	thisCount := len(ranges)
	otherCount := len(otherRanges)
	if thisCount < otherCount {
		otherList.intersectSmaller(list, result)
	} else {
		list.intersectSmaller(otherList, result)
	}
}

func (list *SequentialRangeList[T]) intersectSmaller(otherList, result *SequentialRangeList[T]) {
	ranges := list.ranges
	resultList := result.ranges
	otherRanges := otherList.ranges
	currentIndex := 0
	for i := 0; i < len(otherRanges); i++ {
		otherRange := &otherRanges[i]
		currentIndex, resultList = list.intersectRange(otherRange, resultList, currentIndex)
		if currentIndex >= len(ranges) {
			break
		}
	}
	result.ranges = resultList
}

func (list *SequentialRangeList[T]) intersectRange(seqRange *SequentialRange[T], result []SequentialRange[T], index int) (int, []SequentialRange[T]) {
	lowerIndex := list.binarySearchLower(index, seqRange.GetLower())
	lowerIntersects := lowerIndex >= 0
	if !lowerIntersects {
		lowerIndex = -(lowerIndex + 1)
	}
	var (
		upperIndex      int
		upperIntersects bool
	)
	ranges := list.ranges
	if seqRange.IsMultiple() && lowerIndex != len(ranges) {
		upper := seqRange.GetUpper()
		upperIndex = list.binarySearchLower(lowerIndex, upper)
		upperIntersects = upperIndex >= 0
		if !upperIntersects {
			upperIndex = -(upperIndex + 1)
		}
	} else {
		upperIndex = lowerIndex
		upperIntersects = lowerIntersects
	}

	if lowerIndex < upperIndex { // spans at least one existing range
		if lowerIntersects {
			// add part of the lower intersecting range
			result = append(result, *ranges[lowerIndex].upperSplit(seqRange.GetLower()))
			lowerIndex++
			if lowerIndex < upperIndex {
				result = append(result, ranges[lowerIndex:upperIndex]...)
			}
		} else {
			result = append(result, ranges[lowerIndex:upperIndex]...)
		}
		if upperIntersects {
			result = append(result, *seqRange.upperSplit(ranges[upperIndex].GetLower()))
		}
	} else { // spans 0 or 1 existing range
		if lowerIntersects { // intersects with the one range
			// we know upperIntersects is true because upperIndex points to the same range,
			// so the both lower and upper is contained in the range at lowerIndex,
			// so the intersection is the exact same range
			result = append(result, *seqRange)
		} else if upperIntersects { // intersects partially
			result = append(result, *seqRange.upperSplit(ranges[upperIndex].GetLower()))
		} // else no intersection
	}
	return upperIndex, result
}

type pendingRange[T ipAddressTypeConstraint[T]] struct {
	from                                *SequentialRange[T]
	lower, upper                        T
	lowerIndex, existingRangeUpperIndex int
	isMult                              bool // used by Join but not used by Remove
}

func (pending *pendingRange[T]) clear() {
	pending.from = nil
}

func (pending *pendingRange[T]) isEmpty() bool {
	return pending.from == nil
}

func (pending *pendingRange[T]) String() string {
	if pending.isEmpty() {
		return "<empty>"
	}
	return spanWithRange(pending.lower, pending.upper).ToString(T.ToCanonicalString, DefaultSeqRangeSeparator, T.ToCanonicalString)
}

// JoinIntoNew creates a new list that has all addresses in this list and the provided list.
func (list *SequentialRangeList[T]) JoinIntoNew(otherList *SequentialRangeList[T]) *SequentialRangeList[T] {
	ranges := list.ranges
	otherRanges := otherList.ranges
	if len(otherRanges) == 0 {
		return list.Clone()
	} else if len(ranges) != 0 && !versionsMatch(ranges[0].GetLower(), otherRanges[0].GetLower()) {
		// panic, return nil, return error, or return a copy of this one?
		// All of them have pros and cons.  In fact, if I change the workding to : create a new list with all the addresses in the given list added to this list.
		// Returning an error, I don't like that because with a single arg you can have the nice arithmetic expressions.
		// Panic is a little heavy handed for valid arguments.
		// On the other hand, retruning htis list suggests everything is fine.  Not a good option.
		// So either panic or return nil.
		// I am leaning towards returning nil because it seems to have none of the cons.  Not too heavy, but does not give the illusion of working.  Just an alternative return value.
		// And they may end up with the panic anyway.  Their option.
		return nil
	}
	result := NewSequentialRangeList[T](len(ranges) + len(otherRanges))
	list.join(otherList, result)
	return result
}

func (list *SequentialRangeList[T]) join(otherList, result *SequentialRangeList[T]) {
	ranges := list.ranges
	if len(ranges) == 0 {
		result.ranges = append(result.ranges, otherList.ranges...)
		result.rangeSizes = append(result.rangeSizes, otherList.rangeSizes...)
	} else {
		thisCount := len(ranges)
		otherCount := len(otherList.ranges)
		if thisCount < otherCount {
			otherList.joinSmaller(list, result)
		} else {
			list.joinSmaller(otherList, result)
		}
	}
}

func (list *SequentialRangeList[T]) joinSmaller(otherList, result *SequentialRangeList[T]) {
	resultRanges := result.ranges
	ranges := list.ranges
	otherRanges := otherList.ranges
	currentIndex := 0
	var pending pendingRange[T]
	for i := 0; i < len(otherRanges); i++ {
		seqRange := &otherRanges[i]
		currentIndex, resultRanges = list.joinRange(seqRange, resultRanges, &pending, currentIndex)
	}
	// If there is a pending range, then the range at upper index is in the pending range,
	// so the ranges at upperIndex + 1 must be added after the pending range.
	// Otherwise the ranges at upperIndex must be added.
	if !pending.isEmpty() {
		currentIndex++
		resultRanges = append(resultRanges, *newSequRangeUnchecked(pending.lower, pending.upper, pending.isMult))
		pending.clear()
	}
	if currentIndex < len(ranges) {
		resultRanges = append(resultRanges, ranges[currentIndex:]...)
	}
	result.ranges = resultRanges
}

func (list *SequentialRangeList[T]) joinRange(seqRange *SequentialRange[T], result []SequentialRange[T], pending *pendingRange[T], index int) (int, []SequentialRange[T]) {
	lowerIndex := list.binarySearchLower(index, seqRange.GetLower())
	extendLower := lowerIndex < 0
	if extendLower {
		lowerIndex = -(lowerIndex + 1)
	}
	var (
		upperIndex  int
		extendUpper bool
	)
	ranges := list.ranges
	if seqRange.IsMultiple() && lowerIndex != len(ranges) {
		upperIndex = list.binarySearchLower(lowerIndex, seqRange.GetUpper())
		extendUpper = upperIndex < 0
		if extendUpper {
			upperIndex = -(upperIndex + 1)
		}
	} else {
		upperIndex = lowerIndex
		extendUpper = extendLower
	}
	// check if the lower address is 1 above the upper address of the previous range
	if extendLower && lowerIndex > 0 && ranges[lowerIndex-1].GetUpper().upperIsAdjacentTo(seqRange.GetLower()) {
		lowerIndex--
		extendLower = false
	}
	// check if the upper address is 1 below the lower address of the next range
	if extendUpper && upperIndex < len(ranges) {
		extendUpper = !seqRange.GetUpper().upperIsAdjacentTo(ranges[upperIndex].GetLower())
	}

	if pending.isEmpty() {
		// add ranges following the last range and preceding this one
		if lowerIndex > index {
			result = append(result, ranges[index:lowerIndex]...)
		}
	} else {
		// check if the pending range overlaps with this one, creating a large unified range
		if lowerIndex == pending.existingRangeUpperIndex {
			lowerIndex = pending.lowerIndex
			pending.isMult = true
		} else {
			// there is no overlap, add the pending range, and then add the succeeding ranges that precede this one

			// The only way a pending range can be single, with isMult false, is when the previous input range was single,
			// it intersected with the exact same single range, and then the next time back in we ended up right here,
			// not intersecting with the next input range
			result = append(result, *newSequRangeUnchecked(pending.lower, pending.upper, pending.isMult))
			pending.clear()
			// at this time, index is the pending range upper index
			index++
			if index < lowerIndex {
				result = append(result, ranges[index:lowerIndex]...)
			}
		}
	}

	noPending := pending.isEmpty()
	if lowerIndex < upperIndex { // spans at least one existing range
		var newLower T
		existingRange := &ranges[lowerIndex]
		if !noPending {
			newLower = pending.lower
		} else if extendLower {
			newLower = seqRange.GetLower()
		} else {
			newLower = existingRange.GetLower()
		}
		if extendUpper {
			result = append(result, *newSequRangeUnchecked(newLower, seqRange.GetUpper(), true))
			if !noPending {
				pending.clear()
			}
		} else {
			// the range ends with the existing range,
			// which may overlap the next range to check,
			// so we create a pending range to see if it does
			//
			// note: the pending range is unnecessary if the upper address of the existing range does not exceed seqRange.getUpper(), but checking that is not worth the bother
			if noPending {
				pending.from = seqRange
				pending.lower = newLower
				pending.lowerIndex = lowerIndex
				pending.isMult = true
			}
			pending.existingRangeUpperIndex = upperIndex
			pending.upper = ranges[upperIndex].GetUpper()
		}
	} else { // spans 0 or 1 existing range
		if noPending {
			if extendLower {
				if extendUpper { // spans no existing range, just add it
					result = append(result, *seqRange)
				} else { // spans the single range at lowerIndex
					// the range ends with the existing range,
					// which may overlap the next range to check,
					// so we create a pending range to see if it does
					pending.from = seqRange
					pending.lower = seqRange.GetLower()
					pending.lowerIndex, pending.existingRangeUpperIndex = lowerIndex, lowerIndex
					pending.upper = ranges[lowerIndex].GetUpper()
					pending.isMult = true
				}
			} else {
				// the range is contained in the range at lowerIndex
				// the range ends with the existing range,
				// which may overlap the next range to check,
				// so we create a pending range to see if it does
				//
				// note: the pending range is unnecessary if the upper address of the existing range does not exceed seqRange.getUpper(), but checking that is not worth the bother
				existingRange := &ranges[lowerIndex]
				pending.from = seqRange
				pending.lower = existingRange.GetLower()
				pending.lowerIndex, pending.existingRangeUpperIndex = lowerIndex, lowerIndex
				pending.upper = existingRange.GetUpper()
				pending.isMult = existingRange.IsMultiple()
			}
		} //else we are contained in the same pending range
	}
	return upperIndex, result
}

// Add adds the address, if not already in the list.
//
// If the address version does match existing addresses in the list, the address is not added.
//
// Returns whether addresses were added, whether the list was changed.
func (list *SequentialRangeList[T]) Add(address T) bool {
	ranges := list.ranges
	if len(ranges) == 0 {
		list.addAddressToEmptyList(address)
		return true
	} else if !versionsMatch(ranges[0].GetLower(), address) {
		panic(lookupStr("ipaddress.error.ipVersionMismatch"))
	}
	return list.doAdd(address)
}

func (list *SequentialRangeList[T]) addAddressToEmptyList(address T) {
	if address.IsSequential() {
		rng := coverWithSequentialRange(address)
		list.ranges = append(list.ranges, *rng)
		list.rangeSizes = append(list.rangeSizes, rng.GetCount())
	} else {
		iterator := address.SequentialBlockIterator()
		count := bigZero()
		for {
			rng := coverWithSequentialRange(iterator.Next())
			list.ranges = append(list.ranges, *rng)
			count.Add(count, rng.GetCount())
			list.rangeSizes = append(list.rangeSizes, bigZero().Set(count))
			if !iterator.HasNext() {
				break
			}
		}
	}
	list.changeTracker.Changed()
}

func (list *SequentialRangeList[T]) doAdd(address T) bool {
	if address.IsSequential() {
		return list.addSequentialAddr(address, 0) >= 0
	}
	isChanged := false
	iterator := address.SequentialBlockIterator()
	startIndex := 0
	for {
		startIndex := list.addSequentialAddr(iterator.Next(), startIndex)
		if startIndex >= 0 {
			isChanged = true
		} else {
			startIndex = -(startIndex + 1)
		}
		if !iterator.HasNext() {
			break
		}
	}
	return isChanged
}

func (list *SequentialRangeList[T]) addSequentialAddr(address T, startIndex int) int {
	return list.addSequential(address.IsMultiple(), func() *SequentialRange[T] { return coverWithSequentialRange(address) }, address, address, startIndex)
}

// AddSeqRange sdds the address in the sequential range to the list, if not already in the list.
//
// If the address version of the addresses in the range does match the version of existing addresses in the list, this method panics.
//
// Returns whether addresses in the range were added, whether the list was changed.
func (list *SequentialRangeList[T]) AddSeqRange(seqRange *SequentialRange[T]) bool {
	ranges := list.ranges
	if len(ranges) == 0 {
		list.addRangeToEmptyList(seqRange)
		return true
	} else if !versionsMatch(ranges[0].GetLower(), seqRange.GetLower()) {
		panic(lookupStr("ipaddress.error.ipVersionMismatch"))
	}
	return list.doAddRange(seqRange)
}

func (list *SequentialRangeList[T]) addRangeToEmptyList(seqRange *SequentialRange[T]) {
	list.ranges = append(list.ranges, *seqRange)
	list.rangeSizes = append(list.rangeSizes, seqRange.GetCount())
	list.changeTracker.Changed()
}

func (list *SequentialRangeList[T]) doAddRange(seqRange *SequentialRange[T]) bool {
	return list.addSequential(seqRange.IsMultiple(), func() *SequentialRange[T] { return seqRange }, seqRange.GetLower(), seqRange.GetUpper(), 0) >= 0
}

// addSequential returns true if the collection was changed
func (list *SequentialRangeList[T]) addSequential(isMultiple bool, coverWithSeqRange func() *SequentialRange[T], lowerCompare, upperCompare T, startIndex int) int {
	lowerIndex := list.binarySearchLower(startIndex, lowerCompare)
	extendLower := lowerIndex < 0
	if extendLower {
		lowerIndex = -(lowerIndex + 1)
	}
	var (
		upperIndex  int
		extendUpper bool
	)
	ranges := list.ranges
	if isMultiple && lowerIndex != len(ranges) {
		upperIndex = list.binarySearchUpper(lowerIndex, upperCompare)
		extendUpper = upperIndex < 0
		if extendUpper {
			upperIndex = -(upperIndex + 1)
		}
	} else {
		upperIndex = lowerIndex
		extendUpper = extendLower
	}
	// check if the lower address is 1 above the upper address of the previous range
	if extendLower && lowerIndex > 0 && ranges[lowerIndex-1].GetUpper().upperIsAdjacentTo(lowerCompare) {
		lowerIndex--
		extendLower = false
	}
	// check if the upper address is 1 below the lower address of the next range
	if extendUpper && upperIndex < len(ranges) {
		extendUpper = !upperCompare.upperIsAdjacentTo(ranges[upperIndex].GetLower())
	}
	if lowerIndex < upperIndex { // spans at least one existing range
		var newLower, newUpper T
		existingRange := &ranges[lowerIndex]
		if extendLower {
			newLower = lowerCompare.WithoutPrefixLen().GetLower()
		} else {
			newLower = existingRange.GetLower()
		}
		var nextUpperIndex int
		if extendUpper {
			newUpper = upperCompare.WithoutPrefixLen().GetUpper()
			nextUpperIndex = upperIndex
		} else {
			newUpper = ranges[upperIndex].GetUpper()
			nextUpperIndex = upperIndex + 1
		}
		if extendUpper || extendLower || existingRange.IsMultiple() || lowerIndex+1 < upperIndex {
			ranges[lowerIndex] = *newSequRangeUnchecked(newLower, newUpper, true)
		} else {
			ranges[lowerIndex] = *newSequRangeCheckSize(newLower, newUpper)
		}
		nextLowerIndex := lowerIndex + 1
		if nextLowerIndex < nextUpperIndex {
			// remove the ranges from lower index inclusive to upper index exclusive
			ranges = removeElements(ranges, nextLowerIndex, nextUpperIndex)
			upperIndex -= nextUpperIndex - nextLowerIndex
		}
	} else { // spans 0 or 1 existing range
		if extendLower {
			if extendUpper { // spans no existing range, insert the range
				ranges = insertElementAt(ranges, lowerIndex, *coverWithSeqRange())
			} else { // spans the single range at lowerIndex (which matches upperIndex)
				newLower := lowerCompare.WithoutPrefixLen().GetLower()
				ranges[lowerIndex] = *newSequRangeUnchecked(newLower, ranges[upperIndex].GetUpper(), true)
			}
		} else {
			// nothing to do, the address is contained in the range at lowerIndex
			upperIndex = -(upperIndex + 1) // we've added something, make return value negative to indicate that
			return upperIndex
		}
	}
	list.clearRangeSizesFrom(lowerIndex)
	list.changeTracker.Changed()
	list.ranges = ranges
	return upperIndex
}

func (list *SequentialRangeList[T]) clearRangeSizesFrom(index int) {
	if index < len(list.rangeSizes) {
		list.rangeSizes = list.rangeSizes[:index]
	}
}

// Intersect intersects this list with the given individual address or subnet.
// Afterwards, this list will include only those addresses in both.
func (list *SequentialRangeList[T]) Intersect(address T) bool {
	ranges := list.ranges
	if len(ranges) == 0 {
		return false
	} else if !versionsMatch(ranges[0].GetLower(), address) {
		return false
	}
	startIndex := 0
	isChanged := false
	if address.IsSequential() {
		startIndex = list.intersectSequentialAddr(address, startIndex, true)
		isChanged = startIndex >= 0
	} else {
		iterator := address.SequentialBlockIterator()
		for {
			next := iterator.Next()
			hasNext := iterator.HasNext()
			startIndex = list.intersectSequentialAddr(next, startIndex, !hasNext)
			if isChanged = (startIndex >= 0); !isChanged {
				startIndex = -(startIndex + 1)
			}
			if startIndex >= len(list.ranges) {
				break
			}
			if !hasNext {
				break
			}
		}
	}
	return isChanged
}

func (list *SequentialRangeList[T]) intersectSequentialAddr(address T, startIndex int, isLast bool) int {
	return list.intersectSequential(address.IsMultiple(), nil, address, address, startIndex, isLast)
}

// IntersectSeqRange intersects this list with the given sequential range.
// Afterwards, this list will include only those addresses in both.
func (list *SequentialRangeList[T]) IntersectSeqRange(seqRange *SequentialRange[T]) bool {
	ranges := list.ranges
	if len(ranges) == 0 {
		return false
	} else if !versionsMatch(ranges[0].GetLower(), seqRange.GetLower()) { // } else if !versionsMatch(ranges[0].GetLower(), address) {
		return false
	}
	startIndex := list.intersectSequential(seqRange.IsMultiple(), seqRange, seqRange.GetLower(), seqRange.GetUpper(), 0, true)
	return startIndex >= 0
}

func (list *SequentialRangeList[T]) intersectSequential(isMultiple bool, seqRange *SequentialRange[T], lowerCompare, upperCompare T, startIndex int, isLast bool) int {
	lowerIndex, onLowerRangeBoundary, onUpperRangeBoundary := list.binarySearchLowerWithBoundaries(startIndex, lowerCompare)
	lowerIntersects := lowerIndex >= 0
	if !lowerIntersects {
		lowerIndex = -(lowerIndex + 1)
	}
	var (
		upperIndex                int
		upperIntersects           bool
		upperOnUpperRangeBoundary bool
	)
	ranges := list.ranges
	if isMultiple && lowerIndex != len(ranges) {
		upperIndex, _, upperOnUpperRangeBoundary = list.binarySearchUpperWithBoundaries(lowerIndex, upperCompare)
		upperIndex = list.binarySearchUpper(lowerIndex, upperCompare)
		upperIntersects = upperIndex >= 0
		if !upperIntersects {
			upperIndex = -(upperIndex + 1)
		}
	} else {
		upperIndex = lowerIndex
		upperIntersects = lowerIntersects
		upperOnUpperRangeBoundary = onUpperRangeBoundary
	}
	if lowerIntersects && onLowerRangeBoundary {
		lowerIntersects = false
	}
	if upperIntersects && upperOnUpperRangeBoundary {
		upperIntersects = false
		upperIndex++
	}
	lowestChangedIndex := 0
	if lowerIndex < upperIndex { // spans at least one existing range
		if upperIntersects {
			// range at upper index gets chopped
			existingRange := ranges[upperIndex] // need to copy the range before overwriting
			upper := upperCompare.WithoutPrefixLen().GetUpper()
			ranges[upperIndex] = *newSequRangeCheckSize(existingRange.GetLower(), upper)
			lowestChangedIndex = upperIndex
			upperIndex++
			if !isLast {
				// we need to put back in the remaining in case it might intersect with the next range
				ranges = insertElementAt(ranges, upperIndex, *existingRange.upperSplit(upper.IncrementSingle()))
			}
		}
		if lowerIntersects {
			// range at lower index gets chopped
			if !onLowerRangeBoundary {
				newLower := lowerCompare.WithoutPrefixLen().GetLower()
				ranges[lowerIndex] = *ranges[lowerIndex].upperSplit(newLower)
				lowestChangedIndex = lowerIndex
			} //else the whole lower range intersects
		}
	} else { // spans 0 or 1 existing range
		if upperIntersects {
			existingRange := ranges[upperIndex] // need to copy the range before overwriting
			// range at upper index gets chopped
			var upper T
			if lowerIntersects {
				if seqRange == nil {
					newLower := lowerCompare.WithoutPrefixLen().GetLower()
					upper = upperCompare.WithoutPrefixLen().GetUpper()
					ranges[upperIndex] = *newSequRangeCheckSize(newLower, upper)
				} else {
					upper = seqRange.GetUpper()
					ranges[upperIndex] = *seqRange
				}
			} else {
				newLower := existingRange.GetLower()
				upper = upperCompare.WithoutPrefixLen().GetUpper()
				ranges[upperIndex] = *newSequRangeCheckSize(newLower, upper)
			}
			lowestChangedIndex = upperIndex
			upperIndex++
			if !isLast { // need to put back the remaining in case it intersects with ranges to come
				ranges = insertElementAt(ranges, upperIndex, *existingRange.upperSplit(upper.IncrementSingle()))
			}
		} // else intersects with nothing
	}
	isChanged := lowerIntersects || upperIntersects
	if startIndex < lowerIndex {
		upperIndex -= lowerIndex - startIndex
		ranges = removeElements(ranges, startIndex, lowerIndex)
		isChanged = true
		lowestChangedIndex = startIndex
	}
	if isLast && upperIndex < len(ranges) {
		ranges = removeEndElements(ranges, upperIndex)
		if !isChanged {
			lowestChangedIndex = upperIndex
		}
		isChanged = true
	}
	if isChanged {
		list.ranges = ranges
		list.clearRangeSizesFrom(lowestChangedIndex)
		list.changeTracker.Changed()
	} else {
		upperIndex = -(upperIndex + 1) // we've not changed anything, make return value negative to indicate that
	}
	return upperIndex
}

// Remove removes the given address from the list.  It returns true if the list was changed.
// It returns false if the address was not in the list.
func (list *SequentialRangeList[T]) Remove(address T) bool {
	ranges := list.ranges
	if len(ranges) == 0 {
		return false
	} else if !versionsMatch(ranges[0].GetLower(), address) {
		return false
	}
	if address.IsSequential() {
		return list.removeSequentialAddr(address, 0) >= 0
	}
	result := false
	iterator := address.SequentialBlockIterator()
	startIndex := 0
	for {
		startIndex = list.removeSequentialAddr(iterator.Next(), startIndex)
		if startIndex >= 0 {
			result = true
		} else {
			startIndex = -(startIndex + 1)
		}
		if startIndex >= len(list.ranges) {
			break
		}
		if !iterator.HasNext() {
			break
		}
	}
	return result
}

func (list *SequentialRangeList[T]) removeSequentialAddr(address T, startIndex int) int {
	return list.removeSequential(address.IsMultiple(), address, address, startIndex)
}

// RemoveSeqRange removes all the addresses in the given sequential range from the collection.
// Returns true if the collection was changed.
func (list *SequentialRangeList[T]) RemoveSeqRange(seqRange *SequentialRange[T]) bool {
	ranges := list.ranges
	if len(ranges) == 0 {
		return false
	} else if !versionsMatch(ranges[0].GetLower(), seqRange.GetLower()) {
		return false
	}
	return list.removeSequential(seqRange.IsMultiple(), seqRange.GetLower(), seqRange.GetUpper(), 0) >= 0
}

func (list *SequentialRangeList[T]) removeSequential(isMultiple bool, lowerCompare, upperCompare T, startIndex int) int {
	lowerIndex, onLowerRangeBoundary, onUpperRangeBoundary := list.binarySearchLowerWithBoundaries(startIndex, lowerCompare)
	splitLower := lowerIndex >= 0
	if !splitLower {
		lowerIndex = -(lowerIndex + 1)
	}
	var (
		upperIndex                int
		splitUpper                bool
		upperOnUpperRangeBoundary bool
	)
	ranges := list.ranges
	if isMultiple && lowerIndex != len(ranges) {
		upperIndex, _, upperOnUpperRangeBoundary = list.binarySearchUpperWithBoundaries(lowerIndex, upperCompare)
		splitUpper = upperIndex >= 0
		if !splitUpper {
			upperIndex = -(upperIndex + 1)
		}
	} else {
		upperIndex = lowerIndex
		splitUpper = splitLower
		upperOnUpperRangeBoundary = onUpperRangeBoundary
	}
	if splitLower {
		if onLowerRangeBoundary {
			splitLower = false
		}
	}
	if splitUpper {
		if upperOnUpperRangeBoundary {
			splitUpper = false
			upperIndex++
		}
	}
	if lowerIndex < upperIndex { // spans at least one existing range
		if splitUpper {
			ranges[upperIndex] = *ranges[upperIndex].upperSplit(upperCompare.IncrementBoundarySingle())
		}
		if splitLower {
			ranges[lowerIndex] = *ranges[lowerIndex].lowerSplit(lowerCompare)
			nextIndex := lowerIndex + 1
			if nextIndex < upperIndex {
				ranges = removeElements(ranges, nextIndex, upperIndex)
				upperIndex -= upperIndex - nextIndex
			}
		} else {
			ranges = removeElements(ranges, lowerIndex, upperIndex)
			upperIndex -= upperIndex - lowerIndex
		}
	} else { // spans 0 or 1 existing range
		if splitLower {
			// splitUpper must also be true
			// a slab in the middle is removed
			existingRange := ranges[lowerIndex] // need to copy the range before overwriting
			ranges[lowerIndex] = *ranges[lowerIndex].lowerSplit(lowerCompare)
			upperIndex++
			ranges = insertElementAt(ranges, upperIndex, *existingRange.upperSplit(upperCompare.IncrementBoundarySingle()))
		} else if splitUpper { // spans the single range at lowerIndex
			// range gets chopped
			ranges[lowerIndex] = *ranges[upperIndex].upperSplit(upperCompare.IncrementBoundarySingle())
		} else { // spans no existing range, nothing to do
			upperIndex = -(upperIndex + 1)
			return upperIndex
		}
	}
	list.clearRangeSizesFrom(lowerIndex)
	list.changeTracker.Changed()
	list.ranges = ranges
	return upperIndex
}

// RemoveSeqRangeAt removes the range at the given index.
// An index out of bounds is simply ignored, and nothing is removed.
// Returns true if the list was changed, which is true if and only if the index was not out of bounds.
func (list *SequentialRangeList[T]) RemoveSeqRangeAt(index int) bool {
	ranges := list.ranges
	if index < 0 && index >= len(ranges) {
		return false
	}
	list.ranges = removeElements(list.ranges, index, index+1)
	list.changeTracker.Changed()
	list.clearRangeSizesFrom(index)
	return true
}

// RemoveSeqRanges removes the ranges from fromIndex inclusive to toIndex exclusive.
// Any index out of bounds is simply ignored.
// If fromIndex is greater or equal to toIndex, nothing happens and false is returned.
// Returns true if the list was changed, meaning an index in the range was not out of bounds.
func (list *SequentialRangeList[T]) RemoveSeqRanges(fromIndex, toIndex int) bool {
	ranges := list.ranges
	if fromIndex < 0 {
		fromIndex = 0
	}
	if toIndex > len(ranges) {
		toIndex = len(ranges)
	}
	if fromIndex >= toIndex {
		return false
	}
	list.ranges = removeElements(list.ranges, fromIndex, toIndex)
	list.changeTracker.Changed()
	list.clearRangeSizesFrom(fromIndex)
	return true
}

func (list *SequentialRangeList[T]) binarySearchLowerAddr(key T) int {
	return list.binarySearchForRangeIndex(0, true, key)
}

func (list *SequentialRangeList[T]) binarySearchLowerAddrWithBoundaries(key T) (index int, onLowerRangeBoundary, onUpperRangeBoundary bool) {
	return list.binarySearchForRangeIndexWithBoundaries(0, true, key)
}

func (list *SequentialRangeList[T]) binarySearchUpperAddr(key T) int {
	return list.binarySearchForRangeIndex(0, false, key)
}

func (list *SequentialRangeList[T]) binarySearchUpperAddrWithBoundaries(key T) (index int, onLowerRangeBoundary, onUpperRangeBoundary bool) {
	return list.binarySearchForRangeIndexWithBoundaries(0, false, key)
}

func (list *SequentialRangeList[T]) binarySearchLower(fromIndex int, key T) int {
	return list.binarySearchForRangeIndex(fromIndex, true, key)
}

func (list *SequentialRangeList[T]) binarySearchLowerWithBoundaries(fromIndex int, key T) (index int, onLowerRangeBoundary, onUpperRangeBoundary bool) {
	return list.binarySearchForRangeIndexWithBoundaries(fromIndex, true, key)
}

func (list *SequentialRangeList[T]) binarySearchUpper(fromIndex int, key T) int {
	return list.binarySearchForRangeIndex(fromIndex, false, key)
}

func (list *SequentialRangeList[T]) binarySearchUpperWithBoundaries(fromIndex int, key T) (index int, onLowerRangeBoundary, onUpperRangeBoundary bool) {
	return list.binarySearchForRangeIndexWithBoundaries(fromIndex, false, key)
}

// Returns the index of the range containing the address.
// Otherwise, returns -(insertion index) - 1 where insertion index is the index at which the address would fit into the list.
func (list *SequentialRangeList[T]) binarySearchForRangeIndexWithBoundaries(lowIndex int, lower bool, key T) (rngIndex int, onLowerRangeBoundary, onUpperRangeBoundary bool) {
	ranges := list.ranges
	highIndex := len(ranges) - 1

	if lowIndex <= highIndex {
		var cmp int

		// optimization:
		// in cases when adding a list of sorted and disjoint addresses or ranges,
		// from lowest to highest in order, the newest key will always be above the highest range, so we check that first,
		// checking the entire address space above all the existing ranges
		rng := &ranges[highIndex]
		seqAddr := rng.GetUpper()
		if lower {
			cmp = compareLowerValues(seqAddr, key)
		} else {
			cmp = compareUpperValues(seqAddr, key)
		}
		if cmp < 0 {
			rngIndex = -(len(ranges) + 1)
			return
		} else if cmp == 0 {
			rngIndex = highIndex
			onLowerRangeBoundary = !rng.IsMultiple()
			onUpperRangeBoundary = true
			return
		}

		// optimization:
		// now we do the same for the lowest range, checking the entire address space below all the existing ranges
		rng = &ranges[lowIndex]
		seqAddr = rng.GetLower()
		if lower {
			cmp = compareLowerValues(seqAddr, key)
		} else {
			cmp = compareUpperValues(seqAddr, key)
		}
		if cmp > 0 {
			rngIndex = -(lowIndex + 1)
			return
		} else if cmp == 0 {
			onLowerRangeBoundary = true
			onUpperRangeBoundary = !rng.IsMultiple()
			rngIndex = lowIndex
			return
		} else if lowIndex == highIndex { // only one range and we already determined we are below the upper in the range
			rngIndex = lowIndex
			return
		}

		// now we do the binary search
		for {
			rngIndex = (lowIndex + highIndex) >> 1
			rng = &ranges[rngIndex]
			seqAddr := rng.GetLower()
			if lower {
				cmp = compareLowerValues(seqAddr, key)
			} else {
				cmp = compareUpperValues(seqAddr, key)
			}
			if cmp > 0 {
				highIndex = rngIndex - 1
			} else if cmp < 0 {
				seqAddr = rng.GetUpper()
				if lower {
					cmp = compareLowerValues(seqAddr, key)
				} else {
					cmp = compareUpperValues(seqAddr, key)
				}
				if cmp < 0 {
					lowIndex = rngIndex + 1
				} else if cmp > 0 {
					return
				} else { // cmp == 0
					onUpperRangeBoundary = true
					return
				}
			} else { // cmp == 0
				onLowerRangeBoundary = true
				onUpperRangeBoundary = !rng.IsMultiple()
				return
			}
			if lowIndex > highIndex {
				break
			}
		}
	}
	rngIndex = -(lowIndex + 1)
	return
}

// Returns the index of the range containing the address.
// Otherwise, returns -(insertion index) - 1 where insertion index is the index at which the address would fit into the list.
func (list *SequentialRangeList[T]) binarySearchForRangeIndex(lowIndex int, lower bool, key T) (rngIndex int) {
	ranges := list.ranges
	highIndex := len(ranges) - 1

	if lowIndex <= highIndex {
		var cmp int

		// optimization:
		// in cases when adding a list of sorted and disjoint addresses or ranges,
		// from lowest to highest in order, the newest key will always be above the highest range, so we check that first,
		// checking the entire address space above all the existing ranges
		seqAddr := ranges[highIndex].GetUpper()
		if lower {
			cmp = compareLowerValues(seqAddr, key)
		} else {
			cmp = compareUpperValues(seqAddr, key)
		}
		if cmp < 0 {
			return -(len(ranges) + 1)
		} else if cmp == 0 {
			return highIndex
		}

		// optimization:
		// now we do the same for the lowest range, checking the entire address space below all the existing ranges
		seqAddr = ranges[lowIndex].GetLower()
		if lower {
			cmp = compareLowerValues(seqAddr, key)
		} else {
			cmp = compareUpperValues(seqAddr, key)
		}
		if cmp > 0 {
			return -(lowIndex + 1)
		} else if cmp == 0 || lowIndex == highIndex {
			return lowIndex
		}

		// now we do the binary search
		for {
			rngIndex = (lowIndex + highIndex) >> 1
			mid := &ranges[rngIndex]
			seqAddr := mid.GetLower()
			if lower {
				cmp = compareLowerValues(seqAddr, key)
			} else {
				cmp = compareUpperValues(seqAddr, key)
			}
			if cmp > 0 {
				highIndex = rngIndex - 1
			} else if cmp < 0 {
				seqAddr = mid.GetUpper()
				if lower {
					cmp = compareLowerValues(seqAddr, key)
				} else {
					cmp = compareUpperValues(seqAddr, key)
				}
				if cmp >= 0 {
					return
				}
				lowIndex = rngIndex + 1
			} else {
				return
			}
			if lowIndex > highIndex {
				break
			}
		}
	}
	return -(lowIndex + 1)
}

// IsMultiple returns true if and only if this range list has at least two addresses in it.
func (list *SequentialRangeList[T]) IsMultiple() bool {
	if list == nil {
		return false
	}
	ranges := list.ranges
	return len(ranges) > 1 || (len(ranges) == 1 && ranges[0].IsMultiple())
}

// IsEmpty returns true if and only if this range list has no elements.
func (list *SequentialRangeList[T]) IsEmpty() bool {
	return list == nil || len(list.ranges) == 0
}

// GetSeqRangeCount returns the number of discontinuous sequential ranges of addresses in this list.
func (list *SequentialRangeList[T]) GetSeqRangeCount() int {
	if list == nil {
		return 0
	}
	return len(list.ranges)
}

// IncludesZero Returns whether this list contains the address matching the version of the addresses in this list and having the value of zero.
func (list *SequentialRangeList[T]) IncludesZero() bool {
	size := len(list.ranges)
	return size > 0 && list.ranges[0].IncludesZero()
}

// IncludesMax Returns whether this list contains the address matching the version of the addresses in this list and has the maximum value for addresses of that address version.
func (list *SequentialRangeList[T]) IncludesMax() bool {
	size := len(list.ranges)
	return size > 0 && list.ranges[size-1].IncludesMax()
}

// GetSeqRange returns the sequential range at the given index, the index refers to the sequential ranges in the list, not the contained addresses.
//
// To get the sequential range at a given address index, use GetContainingSeqRange(BigInteger)
//
// GetSeqRange panics if index is outside the bounds or the existing ranges
func (list *SequentialRangeList[T]) GetSeqRange(rangeIndex int) *SequentialRange[T] {
	rng := list.ranges[rangeIndex]
	return &rng
}

// SeqRangeIterator returns an iterator to iterate through the discontinuous sequential ranges of addresses in this list.
func (list *SequentialRangeList[T]) SeqRangeIterator() IteratorWithRemove[*SequentialRange[T]] {
	if list == nil {
		return nilIteratorWithRemove[*SequentialRange[T]]()
	}
	return &rangeListRangesIterator[T]{
		list:          list,
		currentChange: list.changeTracker.GetCurrent(),
	}
}

// Iterator returns an iterator that iterates through all addresses in ascending order.  This iterator supports the remove operation.
func (list *SequentialRangeList[T]) Iterator() Iterator[T] {
	return list.IteratorWithRemove()
}

// AddressIterator is the same as Iterator while satisying the AddressAggregation interface
func (list *SequentialRangeList[T]) AddressIterator() Iterator[AddressType] {
	return addrTypeIterator[T]{list.Iterator()}
}

// Iterator returns an iterator that iterates through all addresses in ascending order.  This iterator supports the remove operation.
func (list *SequentialRangeList[T]) IteratorWithRemove() IteratorWithRemove[T] {
	if list == nil {
		return nilIteratorWithRemove[T]()
	}
	return &rangeListAddrIterator[T]{
		list:            list,
		currentChange:   list.changeTracker.GetCurrent(),
		currentIterator: emptyIterator[T]{},
	}
}

// GetSeqRanges returns the sequential ranges in order.
func (list *SequentialRangeList[T]) GetSeqRanges() []SequentialRange[T] {
	return clone(list.ranges)
}

// GetLower returns the individual address with the lowest numeric value in this sequential range list.
func (list *SequentialRangeList[T]) GetLower() T {
	ranges := list.ranges
	if len(ranges) == 0 {
		return nilPtr[T]()
	}
	return ranges[0].GetLower()
}

// GetUpper returns the individual address with the highest numeric value in this sequential range list.
func (list *SequentialRangeList[T]) GetUpper() T {
	ranges := list.ranges
	rangeLen := len(ranges)
	if rangeLen == 0 {
		return nilPtr[T]()
	}
	return ranges[rangeLen-1].GetUpper()
}

// GetLowerAndUpper returns the individual addresses with the lowest and highest numeric values in this sequential range list.
func (list *SequentialRangeList[T]) GetLowerAndUpper() (lower, upper T) {
	ranges := list.ranges
	rangeLen := len(ranges)
	if rangeLen == 0 {
		res := nilPtr[T]()
		return res, res
	}
	return ranges[0].GetLower(), ranges[rangeLen-1].GetUpper()
}

// GetLowerSeqRange returns the lowest sequential range in the list, or nil if the list is empty
func (list *SequentialRangeList[T]) GetLowerSeqRange() *SequentialRange[T] {
	ranges := list.ranges
	if len(ranges) == 0 {
		return nil
	}
	rng := ranges[0]
	return &rng
}

// GetUpperSeqRange returns the highest sequential range in the list, or nil if the list is empty
func (list *SequentialRangeList[T]) GetUpperSeqRange() *SequentialRange[T] {
	ranges := list.ranges
	if len(ranges) == 0 {
		return nil
	}
	rng := ranges[len(ranges)-1]
	return &rng
}

// Clear empties this list
func (list *SequentialRangeList[T]) Clear() {
	if len(list.ranges) != 0 {
		list.ranges = list.ranges[:0]
		list.clearRangeSizesFrom(0)
		list.changeTracker.Changed()
	}
}

// Removes the individual address at the given index into the lists of addresses.  Returns that address.
// Similar to Get but also removes the address found.
//
// If the index is negative or larger than GetCount() - 1, this method panics
func (list *SequentialRangeList[T]) RemoveAtBig(addressIndex *big.Int) T {
	return list.findAddressBig(addressIndex, true, true)
}

// Increment returns the individual address that is the given increment upwards into the list of sequential ranges, with the increment of 0
// returning the first address.
//
// If there are no addresses in this list, then nil is returned.
//
// If the list of ranges has multiple addresses and the increment exceeds the total number (as returned by GetCount,
// then the final address (last iterator value) is incremented amount by which the increments exceeds the size - 1.
// If that increment exceeds the largest possible address for the version or protocol (eg exceeds IPv4 255.255.255.255), then nil is returned.
//
// If the increment is negative, it is added to the lowest address in the list of sequential ranges (the first iterator value).
// If that increment exceeds the smallest possible address for the version or protocol (eg exceeds IPv4 0.0.0.0), then nil is returned.
//
// A positive increment value is equivalent to the same number of values from the iterator returned by Iterator()
// For instance, a increment of 0 is the first value from the iterator, an increment of 1 is the second value from the iterator, and so on.
// A negative increment added to the total count returned by GetCount is equivalent to the same number of values preceding the upper bound of the iterator.
// For instance, an increment of count - 1 is the last value from the iterator, an increment of count - 2 is the second last value, and so on.
//
// An increment of size matching the count gives you the address just above the highest address in the list of sequential ranges.
// To get the address just below the lowest address in the list of sequential ranges, use the increment -1.
func (list *SequentialRangeList[T]) IncrementBig(addressIndex *big.Int) T {
	return list.findAddressBig(addressIndex, false, false)
}

// GetBig returns the address at the given index into the sorted sequential range list.
//
// GetBig is similar to Increment but does not return any address that is not within this sequential range list.
//
// If the index is negative, or the index exceeds GetCount() - 1, it panics.  It is much like indexing a slice or array.
//
// Otherwise, this returns the address that is the given index upwards into the list of sequential ranges.
// The index of zero returns the first address.
func (list *SequentialRangeList[T]) GetBig(addressIndex *big.Int) T {
	return list.findAddressBig(addressIndex, false, true)
}

// Get returns the address at the given index into the sorted sequential range list.
//
// Get is similar to Increment but does not return any address that is not within this sequential range list.
//
// If the index is negative, or the index exceeds GetCount() - 1, Get will panic.  It is much like indexing a slice or array.
//
// Otherwise, this returns the address that is the given index upwards into the list of sequential ranges.
// The index of zero returns the first address.
func (list *SequentialRangeList[T]) Get(addressIndex int64) T {
	return list.findAddress(addressIndex, false, true)
}

// Gets the sequential range containing the address at the given address index.
//
// To get the sequential range at a sequential range index, use {@link #getSeqRange(int)}
func (list *SequentialRangeList[T]) GetContainingSeqRangeBig(addressIndex *big.Int) *SequentialRange[T] {
	rangeIndex := list.findRangeBig(addressIndex)
	if rangeIndex < 0 || rangeIndex == len(list.ranges) {
		outOfBounds()
	}
	return list.GetSeqRange(rangeIndex)
}

// GetContainingSeqRange gets the sequential range containing the address at the given address index.
//
// To get the sequential range at a sequential range index, use GetSeqRange(int)
func (list *SequentialRangeList[T]) GetContainingSeqRange(addressIndex int64) *SequentialRange[T] {
	rangeIndex := list.findRange(addressIndex)
	if rangeIndex < 0 || rangeIndex == len(list.ranges) {
		outOfBounds()
	}
	return list.GetSeqRange(rangeIndex)
}

// RemoveAt removes the individual address at the given index into the lists of addresses.  Returns that address.
// Similar to Get but also removes the address found.
//
// If the index is negative or larger than GetCount() - 1, this method panics.
func (list *SequentialRangeList[T]) RemoveAt(addressIndex int64) T {
	return list.findAddress(addressIndex, true, true)
}

// Increment returns the individual address that is the given increment upwards into the list of sequential ranges, with the increment of 0
// returning the first address.
//
// If there are no addresses in this list, then null is returned.
//
// If the list of ranges has multiple addresses and the increment exceeds the total number (as returned by GetCount,
// then the final address (last iterator value) is incremented amount by which the increments exceeds the size - 1.
// If that increment exceeds the largest possible address for the version or protocol (eg exceeds IPv4 255.255.255.255), then this returns nil.
//
// If the increment is negative, it is added to the lowest address in the list of sequential ranges (the first iterator value).
// If that increment exceeds the smallest possible address for the version or protocol (eg exceeds IPv4 0.0.0.0), then this returns nil.
//
// A positive increment value is equivalent to the same number of values from the {@link #iterator()}
// For instance, a increment of 0 is the first value from the iterator, an increment of 1 is the second value from the iterator, and so on.
// A negative increment added to the total count returned by GetCount is equivalent to the same number of values preceding the upper bound of the iterator.
// For instance, an increment of count - 1 is the last value from the iterator, an increment of count - 2 is the second last value, and so on.
//
// An increment of size matching the count gives you the address just above the highest address in the list of sequential ranges.
// To get the address just below the lowest address in the list of sequential ranges, use the increment -1.
func (list *SequentialRangeList[T]) Increment(addressIndex int64) T {
	return list.findAddress(addressIndex, false, false)
}

func (list *SequentialRangeList[T]) findAddressBig(index *big.Int, remove, inList bool) T {
	rangeIndex := list.findRangeBig(index)
	ranges := list.ranges
	rangeCount := len(ranges)
	if rangeIndex < 0 {
		if remove || inList {
			outOfBounds()
		} else if rangeCount == 0 {
			return nilPtr[T]()
		}
		return ranges[0].GetLower().IncrementBig(index)
	}
	if rangeIndex == rangeCount {
		if remove || inList {
			outOfBounds()
		} else if rangeCount == 0 {
			return nilPtr[T]()
		}
		lastIndex := rangeCount - 1
		totalRangeSize := list.rangeSizes[lastIndex] // this is the same as getCount()
		one := bigOne()
		return ranges[lastIndex].GetUpper().IncrementBig(one.Sub(one, totalRangeSize).Add(one, index))
	} else if rangeIndex == 0 {
		lower := ranges[0].GetLower()
		if index.Sign() == 0 {
			if remove {
				list.removeFirstAddress(lower)
			}
			return lower
		}
		increment := lower.IncrementBig(index)
		if remove {
			list.removeAddress(increment, 0, index, bigZeroConst())
		}
		return increment
	}
	lower := ranges[rangeIndex].GetLower()
	previousRangesSize := list.rangeSizes[rangeIndex-1]

	var inc big.Int
	inc.Sub(index, previousRangesSize)
	increment := lower.IncrementBig(&inc)
	if remove {
		list.removeAddress(increment, rangeIndex, &inc, previousRangesSize)
	}
	return increment
}

func outOfBounds() {
	panic("out of bounds")
}

func (list *SequentialRangeList[T]) findAddress(index int64, remove, inList bool) T {
	rangeIndex := list.findRange(index)
	ranges := list.ranges
	rangeSizes := list.rangeSizes
	rangeCount := len(ranges)
	if rangeIndex < 0 {
		if remove || inList {
			outOfBounds()
		} else if rangeCount == 0 {
			return nilPtr[T]()
		}
		return ranges[0].GetLower().Increment(index)
	}
	if rangeIndex == rangeCount {
		if remove || inList {
			outOfBounds()
		} else if rangeCount == 0 {
			return nilPtr[T]()
		}
		lastIndex := rangeCount - 1
		totalRangeSize := rangeSizes[lastIndex].Int64()
		return ranges[lastIndex].GetUpper().Increment((index - totalRangeSize) + 1)
	} else if rangeIndex == 0 {
		lower := ranges[0].GetLower()
		if index == 0 {
			if remove {
				list.removeFirstAddress(lower)
			}
			return lower
		}
		increment := lower.Increment(index)
		if remove {
			list.removeAddress(increment, 0, bigZero().SetInt64(index), bigZeroConst())
		}
		return increment
	}
	previousRangesSize := rangeSizes[rangeIndex-1]
	index -= previousRangesSize.Int64()
	lower := ranges[rangeIndex].GetLower()
	increment := lower.Increment(index)
	if remove {
		list.removeAddress(increment, rangeIndex, bigZero().SetInt64(index), previousRangesSize)
	}
	return increment
}

// finds the range containing the address with the given index
func (list *SequentialRangeList[T]) findRangeBig(index *big.Int) int {
	signum := index.Sign()
	if signum <= 0 {
		if signum == 0 {
			return 0
		}
		return -1
	}
	return list.searchForRange(index)
}

// finds the range containing the address with the given index
func (list *SequentialRangeList[T]) findRange(index int64) int {
	if index <= 0 {
		if index == 0 {
			return 0
		}
		return -1
	}
	return list.searchForRange(new(big.Int).SetInt64(index))
}

func (list *SequentialRangeList[T]) searchForRange(index *big.Int) int {
	// search using the existing range sizes
	rangeIndex := list.binarySearchForRange(index)
	if rangeIndex >= 0 {
		return rangeIndex
	}
	// create missing range sizes, and see if we fall in one of those ranges
	rangeSizes := list.rangeSizes
	rangeSzs := len(rangeSizes)
	previousRangeSize := bigZero()
	if rangeSzs != 0 {
		previousRangeSize.Set(rangeSizes[rangeSzs-1])
	}
	i := rangeSzs
	ranges := list.ranges
	total := len(ranges)
	for ; i < total; i++ {
		rngCount := ranges[i].GetCount()
		rngCount.Add(rngCount, previousRangeSize)
		list.rangeSizes = append(list.rangeSizes, rngCount)
		if index.CmpAbs(rngCount) < 0 {
			return i
		}
		previousRangeSize.Set(rngCount)
	}
	return total
}

func (list *SequentialRangeList[T]) binarySearchForRange(index *big.Int) int {
	rangeSizes := list.rangeSizes
	highIndex := len(rangeSizes)
	if highIndex == 0 {
		return -1
	}

	// above the highest
	highSize := rangeSizes[highIndex-1]
	if highSize.CmpAbs(index) <= 0 {
		return -1
	}

	lowIndex := 0
	for lowIndex <= highIndex {
		midIndex := (lowIndex + highIndex) >> 1
		midSize := rangeSizes[midIndex]
		if index.CmpAbs(midSize) >= 0 {
			lowIndex = midIndex + 1
		} else if midIndex == 0 || index.CmpAbs(rangeSizes[midIndex-1]) >= 0 {
			return midIndex
		} else {
			highIndex = midIndex - 1
		}
	}
	return -1
}

func (list *SequentialRangeList[T]) removeFirstAddress(address T) {
	ranges := list.ranges
	if ranges[0].IsMultiple() {
		// the lower side is removed
		ranges[0] = *ranges[0].upperSplit(address.IncrementSingle())
	} else {
		// the first range is removed
		ranges[0] = SequentialRange[T]{} // erase it to help garbage collection
		list.ranges = ranges[1:]
	}
	list.clearRangeSizesFrom(0)
	list.changeTracker.Changed()
}

func (list *SequentialRangeList[T]) removeAddress(individualAddress T, rngIndex int, addressIndexInRange, previousRangesSize *big.Int) {
	var rngSize big.Int
	rngSize.Set(list.rangeSizes[rngIndex]).Sub(&rngSize, previousRangesSize) // The range size is populated due to the search that got us here
	if addressIndexInRange.Sign() == 0 {
		// the lower address is removed
		if rngSize.Cmp(bigOneConst()) == 0 {
			// the whole range is just that one address
			list.ranges = removeElements(list.ranges, rngIndex, rngIndex+1)
		} else {
			rng := &list.ranges[rngIndex]
			list.ranges[rngIndex] = *rng.upperSplit(individualAddress.IncrementSingle())
		}
	} else {
		one := bigOne()
		ranges := list.ranges
		if rngSize.Cmp(one.Add(addressIndexInRange, one)) == 0 {
			// the upper address is removed
			ranges[rngIndex] = *ranges[rngIndex].lowerSplit(individualAddress)
		} else {
			// a slab in the middle is removed
			existingRng := ranges[rngIndex] // need to copy the range before overwriting
			ranges[rngIndex] = *existingRng.lowerSplit(individualAddress)
			list.ranges = insertElementAt(ranges, rngIndex+1, *existingRng.upperSplit(individualAddress.IncrementSingle()))
		}
	}
	list.clearRangeSizesFrom(rngIndex)
	list.changeTracker.Changed()
}

// gets the count of addresses in the first rangeCount ranges
// Note: callers must always use a copy of the value returned from this method (unless just reading it, of course)
func (list *SequentialRangeList[T]) getCount(rangeCount int) *big.Int {
	rangeSizes := list.rangeSizes
	ranges := list.ranges
	if rangeCount > len(rangeSizes) {
		// always ensure the capacity of rangeSizes is at least the length of the ranges list
		if cap(rangeSizes) < len(ranges) {
			rangeSizes = append(make([]*big.Int, 0, len(ranges)), rangeSizes...)
		}
		var count *big.Int
		index := len(rangeSizes) - 1
		if index >= 0 {
			count = rangeSizes[index]
		} else {
			count = bigZeroConst()
		}
		index++
		seqRange := &list.ranges[index]
		if seqRange.IsIPv4() {
			ipv4Count := count.Uint64()
			ipv4Range := seqRange.ToIPv4()
			ipv4Count += ipv4Range.getIPv4Count()

			rangeSizes = append(rangeSizes, bigZero().SetUint64(ipv4Count))
			index++
			for index < rangeCount {
				ipv4Range := ranges[index]
				ipv4Count += ipv4Range.getIPv4Count()
				rangeSizes = append(rangeSizes, bigZero().SetUint64(ipv4Count))
				index++
			}
			count = bigZero().SetUint64(ipv4Count)
		} else {
			rngCount := seqRange.GetCount()
			rngCount.Add(rngCount, count)
			rangeSizes = append(rangeSizes, rngCount)
			var cumulativeCount big.Int
			cumulativeCount.Set(rngCount)
			index++
			for index < rangeCount {
				seqRange := ranges[index]
				rngCount = seqRange.GetCount()
				rngCount.Add(rngCount, &cumulativeCount)
				cumulativeCount.Set(rngCount)
				rangeSizes = append(rangeSizes, rngCount)
				index++
			}
			count = &cumulativeCount
		}
		list.rangeSizes = rangeSizes
		return count
	} else if rangeCount > 0 {
		return rangeSizes[rangeCount-1]
	}
	return bigZeroConst()
}

// GetCount returns the number of individual addresses in this range list, the number of elements in this collection.
func (list *SequentialRangeList[T]) GetCount() *big.Int {
	if list == nil {
		return bigZero()
	}
	return bigZero().Set(list.getCount(len(list.ranges)))
}

// Enumerate returns the distance of the given address from the initial value of this range list.  It indicates where an address sits relative to the range ordering.
//
//	If within or above the range list, it is the distance to the lower boundary of the sequential range.  If below the range list, it returns the number of addresses following the address to the lower range boundary.
//
// You can call Contains or you can compare with GetCount to check for containment.
// An IP address is in the range list if 0 <= Enumerate(IP Address) < GetCount.
//
// If the address is above the lower boundary and below the upper boundary of the range list, but is not within a range in the range list, then this method returns nil.
//
// Returns nil when the argument is a multi-valued subnet. The argument must be an individual address.
//
// Returns nil when there are no ranges in this sequential range list.
//
// Returns nil when the address version does not match the addresses in this range list.
func (list *SequentialRangeList[T]) Enumerate(address AddressType) *big.Int {
	addr, isNil, ok := ConvertAddressTypeCheckNil[T](address)
	if ok && !isNil && list != nil {
		return list.enumerateAddress(addr)
	}
	return nil
}

// EnumerateAddress returns the distance of the given address from the initial value of this range list.  It indicates where an address sits relative to the range ordering.
//
// If within or above the range list, it is the distance to the lower boundary of the sequential range list.  If below the range list, returns the number of addresses following the address to the lower range boundary.
//
// You can call Contains or you can compare with GetCount to check for containment.
// An address is in the range list if 0 <= Enumerate(address) < GetCount().
//
// If the address is above the lower boundary and below the upper boundary of the range list, but is not within a range in the range list, then this method returns nil.
//
// Returns nil when the argument is a multi-valued subnet. The argument must be an individual address.
//
// Returns nil when there are no ranges in this sequential range list.
//
// Returns nil when the address version does not match the addresses in this range list.
func (list *SequentialRangeList[T]) EnumerateAddress(address T) *big.Int {
	if list == nil || isNilPtr(address) || address.IsMultiple() {
		return nil
	}
	return list.enumerateAddress(address)
}

func (list *SequentialRangeList[T]) enumerateAddress(address T) *big.Int {
	ranges := list.ranges
	rangeLen := len(ranges)
	if rangeLen == 0 {
		return nil
	} else if !versionsMatch(ranges[0].GetLower(), address) {
		return nil
	}
	lowerIndex, onLowerRangeBoundary, onUpperRangeBoundary := list.binarySearchLowerAddrWithBoundaries(address)
	if lowerIndex < 0 {
		lowerIndex = -(lowerIndex + 1)
		if lowerIndex > 0 && lowerIndex < rangeLen {
			return nil
		} else if lowerIndex == 0 {
			return ranges[0].enumerate(address)
		} // lowerIndex == ranges.size() && ranges.size() > 0
		lastRangeIndex := rangeLen - 1
		res := ranges[lastRangeIndex].enumerate(address)
		return res.Add(res, list.getCount(lastRangeIndex))
	}
	if onLowerRangeBoundary {
		return bigZero().Set(list.getCount(lowerIndex))
	} else if onUpperRangeBoundary {
		count := big.NewInt(-1)
		count.Add(count, list.getCount(lowerIndex+1))
		return count
	}
	res := ranges[lowerIndex].enumerate(address)
	return res.Add(res, list.getCount(lowerIndex))
}

// Equal returns true if and only if this list has the same set of individual addresses as the given list
func (list *SequentialRangeList[T]) Equal(other *SequentialRangeList[T]) bool {
	if list == other {
		return true
	}
	return equalRangeLists(list, other)
}

func equalRangeLists[T ipAddressTypeConstraint[T], R ipAddressTypeConstraint[R]](list *SequentialRangeList[T], other *SequentialRangeList[R]) bool {
	//   nil aggregation contains
	if list == nil {
		return IsEmpty(other)
	} else if other == nil {
		return IsEmpty(list)
	}
	ranges := list.ranges
	rangeCount := len(ranges)
	otherRanges := other.ranges
	otherRangeCount := len(otherRanges)
	if rangeCount != otherRangeCount {
		return false
	}
	sizes := list.rangeSizes
	otherSizes := other.rangeSizes
	sizesLen := len(sizes)
	otherSizesLen := len(otherSizes)
	if sizesLen < otherSizesLen {
		sizesLen--
		if sizesLen >= 0 && sizes[sizesLen].Cmp(otherSizes[sizesLen]) != 0 {
			return false
		}
	} else {
		otherSizesLen--
		if otherSizesLen >= 0 && sizes[otherSizesLen].Cmp(otherSizes[otherSizesLen]) != 0 {
			return false
		}
	}
	for i := 0; i < rangeCount; i++ {
		if !ranges[i].Equal(&otherRanges[i]) {
			return false
		}
	}
	return true
}

// EqualAggregation returns true if and only if this list has the same set of individual addresses as the given aggregation of addresses
func (list *SequentialRangeList[T]) EqualAggregation(otherAggregation AddressAggregation) bool {
	if list == nil {
		return IsEmpty(otherAggregation)
	}
	switch other := otherAggregation.(type) {
	case nil:
		return IsEmpty(list)
	case AddressType:
		return list.equalAddr(other)
	case IPAddressSeqRangeType:
		return list.equalRange(other)
	case *IPAddressContainmentTrie:
		return equalListAndContainmentTrie(list, other)
	case *IPv4AddressContainmentTrie:
		return equalListAndContainmentTrie(list, other)
	case *IPv6AddressContainmentTrie:
		return equalListAndContainmentTrie(list, other)
	case *SequentialRangeList[T]:
		return list.Equal(other)
	case *IPAddressSeqRangeList:
		return equalRangeLists(list, other)
	case *IPv4AddressSeqRangeList:
		return equalRangeLists(list, other)
	case *IPv6AddressSeqRangeList:
		return equalRangeLists(list, other)
	default:
		return equalAggregation(list, other)
	}
}

func (list *SequentialRangeList[T]) equalAddr(other AddressType) bool {
	if list == nil {
		return isEmptyAddr(other) // addresses are never empty, unless they are nil
	} else if other == nil {
		return IsEmpty(list)
	}
	addr := other.ToAddressBase()
	if addr == nil {
		return IsEmpty(list)
	}
	ranges := list.ranges
	if len(ranges) == 0 {
		return false // addresses are never empty, unless they are nil
	}
	// at this point we know the list is not empty
	if other.IsSequential() {
		if len(ranges) > 1 { // not sequential
			return false
		}
		rng := &ranges[0]
		ipaddr := addr.ToIP()
		if ipaddr == nil {
			return false
		}
		if !rng.GetLower().Equal(ipaddr.GetLower()) {
			return false
		}
		return !rng.IsMultiple() || rng.GetUpper().Equal(ipaddr.GetUpper())
	} else if list.GetCount().Cmp(other.GetCount()) != 0 {
		return false
	}
	ipaddr := addr.ToIP()
	if ipaddr == nil {
		return false
	}
	if bigZero().SetUint64(uint64(len(ranges))) != ipaddr.GetSequentialBlockCount() {
		return false
	}
	others := ipaddr.SequentialBlockIterator()
	for i := 0; i < len(ranges); i++ {
		if !others.HasNext() {
			return false
		}
		rng := &ranges[i]
		if !rng.GetLower().Equal(ipaddr.GetLower()) {
			return false
		}
		if rng.IsMultiple() {
			if !rng.GetUpper().Equal(ipaddr.GetUpper()) {
				return false
			}
		} else if ipaddr.isMultiple() {
			return false
		}
	}
	return !others.HasNext()
}

func (list *SequentialRangeList[T]) equalRange(other IPAddressSeqRangeType) bool {
	if list == nil {
		return isEmptyRange(other) // addresses are never empty, unless they are nil
	} else if other == nil || other.ToIP() == nil {
		return IsEmpty(list)
	}
	ranges := list.ranges
	return len(ranges) == 1 && ranges[0].Equal(other)
}

func equalListAndContainmentTrie[T ipAddressTypeConstraint[T], C ipAddressTypeConstraint[C]](list *SequentialRangeList[T], other *ContainmentTrieBase[C]) bool {
	//   nil aggregation contains
	if list == nil {
		return IsEmpty(other)
	} else if other == nil {
		return IsEmpty(list)
	}
	ranges := list.ranges
	if len(ranges) == 0 {
		return other.IsEmpty()
	} else if list.GetCount().Cmp(other.GetCount()) != 0 {
		return false
	}
	prefBlocks := other.PrefixBlockIterator()
	i := 0
	var rngCount big.Int
	for {
		rng := ranges[i]
		rngCount.Set(rng.getCachedCount(false))
		// we should be able to match the upcoming prefix blocks to the range
		for {
			if !prefBlocks.HasNext() {
				return false
			}
			prefBlock := prefBlocks.Next()
			if !rng.Contains(prefBlock) {
				return false
			}
			rngCount.Sub(&rngCount, prefBlock.GetCount())
			sign := rngCount.Sign()
			if sign < 0 {
				return false
			} else if sign == 0 {
				break
			}
		}
		i++
		if i == len(ranges) {
			break
		}
	}
	return !prefBlocks.HasNext() // since we compared counts, HasNext should always return false here
}

// Clone copies this IPAddressSeqRangeList.
func (list *SequentialRangeList[T]) Clone() *SequentialRangeList[T] {
	return &SequentialRangeList[T]{
		ranges:     clone(list.ranges),
		rangeSizes: clone(list.rangeSizes), // this is safe because a big.Int in rangeSizes is never changed, it can only be replaced
	}
}

// NewEmpty creates a new IPAddressSeqRangeList using the same element type T.
// Satisfies the IPAddressCollConstraint[S IPAddressCollAddrConstraint[T], T IPAddressTypeConstraint[T]] interface,
// allowing for ogeneric code that can create new collections generically, with generic code.
// For code that is using a generic collection type, you can simply use &SequentialRangeList[T]{} to create a new list.
func (list *SequentialRangeList[T]) NewEmpty() *SequentialRangeList[T] {
	return &SequentialRangeList[T]{}
}

// IsSequential returns whether the collection represents a range of addresses that are sequential.
//
// Generally, this means that given any two addresses in the collection, all addresses between are also in the collection.
func (list *SequentialRangeList[T]) IsSequential() bool {
	return len(list.ranges) <= 1
}

// Format implements the [fmt.Formatter] interface.
//
// The formats, flags, and other specifications supported are those supported by Format in IPAddress.
func (list *SequentialRangeList[T]) Format(state fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		_, _ = state.Write([]byte(list.String()))
		return
	}
	list.format(state, verb)
}

func (list *SequentialRangeList[T]) format(state fmt.State, verb rune) {
	_, _ = state.Write([]byte{'['})
	ranges := list.ranges
	if len(ranges) > 0 {
		ranges[0].Format(state, verb)
		remainder := list.ranges[1:]
		for i := range remainder {
			_, _ = state.Write([]byte(", "))
			remainder[i].Format(state, verb)
		}
	}
	_, _ = state.Write([]byte{']'})
}

// String returns the canonical string representing this sequential range list.
func (list *SequentialRangeList[T]) String() string {
	return list.ToCanonicalString()
}

// ToCanonicalString returns the canonical string representing this sequential range list, showing the underlying list of sequential ranges.
func (list *SequentialRangeList[T]) ToCanonicalString() string {
	if list == nil {
		return nilString()
	}
	// SequentialRange.String uses ToCanonicalString
	return fmt.Sprint(list.ranges)
}

// ToCanonicalString returns the normalized string representing this sequential range list.
func (list *SequentialRangeList[T]) ToNormalizedString() string {
	return list.toString((*SequentialRange[T]).ToNormalizedString)
}

func (list *SequentialRangeList[T]) toString(rangeStringer func(*SequentialRange[T]) string) string {
	if list == nil {
		return nilString()
	}
	builder := strings.Builder{}
	builder.WriteByte('[')
	ranges := list.ranges
	if len(ranges) > 0 {
		builder.WriteString(rangeStringer(&ranges[0]))
		remainder := list.ranges[1:]
		for i := range remainder {
			builder.WriteString(", ")
			builder.WriteString(rangeStringer(&remainder[i]))
		}
	}
	builder.WriteByte(']')
	return builder.String()
}

// CoverWithSequentialRange returns the unique sequential range of minimal size that includes all the addresses in this collection.
// If there are no addresses in this collection, then nil is returned.
//
// The result will represent the same set of addresses if and only if the set of addresses in this collection are sequential, in which case IsSequential returns true.
func (list *SequentialRangeList[T]) CoverWithSequentialRange() *SequentialRange[T] {
	ranges := list.ranges
	rangeLen := len(ranges)
	if rangeLen == 0 {
		return nil
	} else if rangeLen == 1 {
		rng := ranges[0]
		return &rng
	}
	return spanWithRange(ranges[0].GetLower(), ranges[rangeLen-1].GetUpper())
}

func spanWithRange[T ipAddressTypeConstraint[T]](one, two T) *SequentialRange[T] {
	return newSequRangeCheckSize(one, two)
}

func coverWithSequentialRange[T ipAddressTypeConstraint[T]](addr T) *SequentialRange[T] {
	addr = addr.WithoutPrefixLen()
	lower, upper := addr.GetLowerAndUpper()
	return newSequRangeUnchecked(
		lower,
		upper,
		addr.IsMultiple())
}

// CoverWithPrefixBlock returns the unique CIDR prefix block subnet or individual address of minimal size that includes all the addresses in this sequential range list.
// If there are no addresses in this list, then nil is returned.
func (list *SequentialRangeList[T]) CoverWithPrefixBlock() T {
	ranges := list.ranges
	rangeLen := len(ranges)
	if rangeLen == 0 {
		return nilPtr[T]()
	}
	return ranges[0].GetLower().CoverWithPrefixBlockTo(ranges[rangeLen-1].GetUpper())
}

// SpanWithPrefixBlocks returns the minimal set of disjoint prefix blocks containing the addresses in this collection.
//
// For large lists, it is better to use SpanningPrefixBlockIterator
func (list *SequentialRangeList[T]) SpanWithPrefixBlocks() []T {
	if len(list.ranges) == 0 {
		return nil
	}
	return list.getSpanningBlocks((*SequentialRange[T]).SpanWithPrefixBlocks)
}

// SpanningPrefixBlockIterator iterates over the minimal set of prefix blocks that spans the addresses in the sequential range list
func (list *SequentialRangeList[T]) SpanningPrefixBlockIterator() Iterator[T] {
	return list.PrefixBlockIterator()
}

// PrefixBlockIterator is the same as SpanningPrefixBlockIterator but also provides the Remove operation.
func (list *SequentialRangeList[T]) PrefixBlockIterator() IteratorWithRemove[T] {
	if list == nil {
		return nilIteratorWithRemove[T]()
	}
	return &rangeListAddrIterator[T]{
		prefBlocks:      true,
		list:            list,
		currentChange:   list.changeTracker.GetCurrent(),
		currentIterator: emptyIterator[T]{},
	}
}

// SpanWithPrefixBlocks returns the minimal set of disjoint sequential blocks containing the addresses in this collection.
//
// For large lists, it may be better to use SpanningSeqBlockIterator
func (list *SequentialRangeList[T]) SpanWithSequentialBlocks() []T {
	if len(list.ranges) == 0 {
		return nil
	}
	return list.getSpanningBlocks((*SequentialRange[T]).SpanWithSequentialBlocks)
}

// SpanningSeqBlockIterator returns an iterator to iterate, in order, the minimal set of disjoint sequential blocks containing the addresses in this collection.
func (list *SequentialRangeList[T]) SpanningSeqBlockIterator() Iterator[T] {
	return list.SpanningSeqIteratorWithRemove()
}

// SpanningSeqIteratorWithRemove returns an iterator to iterate, in order, the minimal set of disjoint sequential blocks containing the addresses in this collection,
// while allowing for element removal after each iteration.
func (list *SequentialRangeList[T]) SpanningSeqIteratorWithRemove() IteratorWithRemove[T] {
	if list == nil {
		return nilIteratorWithRemove[T]()
	}
	return &rangeListAddrIterator[T]{
		seqBlocks:       true,
		list:            list,
		currentChange:   list.changeTracker.GetCurrent(),
		currentIterator: emptyIterator[T]{},
	}
}

func (list *SequentialRangeList[T]) getSpanningBlocks(
	blocksProducer func(*SequentialRange[T]) []T) (result []T) {
	ranges := list.ranges
	rangeLen := len(ranges)
	for i := 0; i < rangeLen; i++ {
		rng := &ranges[i]
		result = append(result, blocksProducer(rng)...)
	}
	return
}

type rangeListRangesIterator[T ipAddressTypeConstraint[T]] struct {
	list          *SequentialRangeList[T]
	index         int
	currentChange tree.Change

	hasCurrent bool
}

func (iter *rangeListRangesIterator[T]) HasNext() bool {
	return iter.index < len(iter.list.ranges)
}

func (iter *rangeListRangesIterator[T]) Next() (res *SequentialRange[T]) {
	list := iter.list
	list.changeTracker.ChangedSince(iter.currentChange)
	if iter.HasNext() {
		next := list.ranges[iter.index]
		iter.index++
		iter.hasCurrent = true
		res = &next
	} else if iter.hasCurrent {
		iter.hasCurrent = false
	}
	return
}

func (iter *rangeListRangesIterator[T]) Remove() (res *SequentialRange[T]) {
	if iter.hasCurrent {
		list := iter.list
		list.changeTracker.ChangedSince(iter.currentChange)
		iter.index--
		last := list.ranges[iter.index]
		res = &last
		list.ranges = removeElement(list.ranges, iter.index)
		iter.hasCurrent = false
		iter.currentChange = list.changeTracker.GetCurrent()
	}
	return
}

type rangeListAddrIterator[T ipAddressTypeConstraint[T]] struct {
	prefBlocks, seqBlocks bool
	list                  *SequentialRangeList[T]
	currentChange         tree.Change
	nextRangeIndex        int
	currentIterator       Iterator[T]

	last T

	firstOfRange /* removedLast, */, hasLast bool
}

// Note: If we used an iterator on the range list,
// it would not be enough to handle the change tracking.
// While it is enough to detect any and all changes,
// the problem is that when the range list is changed,
// we might not actually attempt to use the range iterator again
// until the current iterator on the current sequential range is extinguished.
// The check for modifications would be delayed.

func (iter *rangeListAddrIterator[T]) HasNext() bool {
	return iter.currentIterator.HasNext() || iter.nextRangeIndex < len(iter.list.ranges)
}

func (iter *rangeListAddrIterator[T]) Next() T {
	list := iter.list
	list.changeTracker.ChangedSince(iter.currentChange)
	currentIterator := iter.currentIterator
	if iter.firstOfRange = !currentIterator.HasNext() && iter.nextRangeIndex < len(list.ranges); iter.firstOfRange {
		rng := list.ranges[iter.nextRangeIndex]
		if iter.prefBlocks {
			currentIterator = rng.SpanningPrefixBlockIterator()
		} else if iter.seqBlocks {
			currentIterator = rng.SpanningSeqBlockIterator()
		} else {
			currentIterator = rng.Iterator()
		}
		iter.currentIterator = currentIterator
		iter.nextRangeIndex++
	} else {
		iter.firstOfRange = !iter.hasLast // if we removed the previous, that puts as at the first of the new range
	}
	iter.hasLast = true
	iter.last = currentIterator.Next()
	return iter.last
}

func (iter *rangeListAddrIterator[T]) Remove() T {
	list := iter.list
	list.changeTracker.ChangedSince(iter.currentChange)
	if !iter.hasLast {
		return nilPtr[T]()
	}

	currentRangeIndex := iter.nextRangeIndex - 1
	// note: there is no need to reset the iterator in any of these cases
	if iter.firstOfRange {
		if !iter.currentIterator.HasNext() {
			list.ranges = removeElement(list.ranges, currentRangeIndex)
			iter.nextRangeIndex--
		} else {
			ranges := list.ranges
			ranges[currentRangeIndex] = *ranges[currentRangeIndex].upperSplit(iter.last.IncrementSingle())
		}
	} else if !iter.currentIterator.HasNext() {
		// last of range
		ranges := list.ranges
		rng := ranges[currentRangeIndex]
		ranges[currentRangeIndex] = *rng.lowerSplit(iter.last)
	} else {
		// in the middle of the range
		ranges := list.ranges
		rng := ranges[currentRangeIndex]
		ranges[currentRangeIndex] = *rng.lowerSplit(iter.last)
		list.ranges = insertElementAt(ranges, iter.nextRangeIndex, *rng.upperSplit(iter.last.IncrementBoundarySingle()))
		iter.nextRangeIndex++
	}
	iter.hasLast = false
	list.clearRangeSizesFrom(currentRangeIndex)
	iter.currentChange = list.changeTracker.GetCurrent()
	return iter.last
}

type (
	IPAddressSeqRangeList   = SequentialRangeList[*IPAddress]
	IPv4AddressSeqRangeList = SequentialRangeList[*IPv4Address]
	IPv6AddressSeqRangeList = SequentialRangeList[*IPv6Address]
)
