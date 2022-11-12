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

import "github.com/seancfoley/bintree/tree"

// CachingTrieIterator is an iterator of a tree that allows you to cache an object with the
// lower or upper sub-node of the currently visited node.
// The cached object can be retrieved later when iterating the sub-node.
// That allows you to provide iteration context from a parent to its sub-nodes when iterating,
// but can only be provided with iterators in which parent nodes are visited before their sub-nodes.
// The caching and retrieval is done in constant-time.
type CachingTrieIterator[T any] interface {
	IteratorRem[T]

	// Note: We could theoretically try to make the cached type generic.
	// But the problem with that is that the iterator methods that return them cannot be generic on their own.
	// The other problem is that even if we could, some callers would not care about the caching behaviour and thus would not want to have to specify a cache type.

	tree.CachingIterator
}

// addressKeyIterator implements the address key iterator for tries
type addressKeyIterator[T TrieKeyConstraint[T]] struct {
	tree.TrieKeyIterator
}

func (iter addressKeyIterator[T]) Next() (t T) {
	key := iter.TrieKeyIterator.Next()
	if key != nil {
		return key.(trieKey[T]).address
	}
	return
}

func (iter addressKeyIterator[T]) Remove() (t T) {
	key := iter.TrieKeyIterator.Remove()
	if key != nil {
		return key.(trieKey[T]).address
	}
	return
}

//
type addrTrieNodeIteratorRem[T TrieKeyConstraint[T]] struct {
	tree.TrieNodeIteratorRem
}

func (iter addrTrieNodeIteratorRem[T]) Next() *TrieNode[T] {
	return toAddressTrieNodeX[T](iter.TrieNodeIteratorRem.Next())
}

func (iter addrTrieNodeIteratorRem[T]) Remove() *TrieNode[T] {
	return toAddressTrieNodeX[T](iter.TrieNodeIteratorRem.Remove())
}

//
type addrTrieNodeIterator[T TrieKeyConstraint[T]] struct {
	tree.TrieNodeIterator
}

func (iter addrTrieNodeIterator[T]) Next() *TrieNode[T] {
	return toAddressTrieNodeX[T](iter.TrieNodeIterator.Next())
}

//
type cachingAddressTrieNodeIterator[T TrieKeyConstraint[T]] struct {
	tree.CachingTrieNodeIterator
}

func (iter cachingAddressTrieNodeIterator[T]) Next() *TrieNode[T] {
	return toAddressTrieNodeX[T](iter.CachingTrieNodeIterator.Next())
}

func (iter cachingAddressTrieNodeIterator[T]) Remove() *TrieNode[T] {
	return toAddressTrieNodeX[T](iter.CachingTrieNodeIterator.Remove())
}

//////////////////////////////////////////////////////////////////
//////

type associativeAddressTrieNodeIteratorRem[T TrieKeyConstraint[T], V any] struct {
	tree.TrieNodeIteratorRem
}

func (iter associativeAddressTrieNodeIteratorRem[T, V]) Next() *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](iter.TrieNodeIteratorRem.Next())
}

func (iter associativeAddressTrieNodeIteratorRem[T, V]) Remove() *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](iter.TrieNodeIteratorRem.Remove())
}

//
type associativeAddressTrieNodeIterator[T TrieKeyConstraint[T], V any] struct {
	tree.TrieNodeIterator
}

func (iter associativeAddressTrieNodeIterator[T, V]) Next() *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](iter.TrieNodeIterator.Next())
}

//
type cachingAssociativeAddressTrieNodeIteratorX[T TrieKeyConstraint[T], V any] struct {
	tree.CachingTrieNodeIterator
}

func (iter cachingAssociativeAddressTrieNodeIteratorX[T, V]) Next() *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](iter.CachingTrieNodeIterator.Next())
}

func (iter cachingAssociativeAddressTrieNodeIteratorX[T, V]) Remove() *AssociativeTrieNode[T, V] {
	return toAssociativeTrieNode[T, V](iter.CachingTrieNodeIterator.Remove())
}

//
type emptyIterator[T any] struct{}

func (it emptyIterator[T]) HasNext() bool {
	return false
}

func (it emptyIterator[T]) Next() (t T) {
	return
}

func nilAddressIterator[T any]() Iterator[T] {
	return emptyIterator[T]{}
}
