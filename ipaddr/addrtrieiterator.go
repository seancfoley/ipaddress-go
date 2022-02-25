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

import "github.com/seancfoley/bintree/tree"

// addressKeyIterator implements AddressIterator for tries
type addressKeyIterator struct {
	tree.TrieKeyIterator
}

func (iter addressKeyIterator) Next() *Address {
	key := iter.TrieKeyIterator.Next()
	if key != nil {
		return key.(*addressTrieKey).Address
	}
	return nil
}

func (iter addressKeyIterator) Remove() *Address {
	key := iter.TrieKeyIterator.Remove()
	if key != nil {
		return key.(*addressTrieKey).Address
	}
	return nil
}

// AddressTrieNodeIteratorRem iterates through an address trie, until both Next() returns nil and HasNext() returns false
type AddressTrieNodeIterator interface {
	HasNext

	Next() *AddressTrieNode
}

// AddressTrieNodeIteratorRem iterates through an address trie, until both Next() returns nil and HasNext() returns false,
// and also allows you to remove the node just visited
type AddressTrieNodeIteratorRem interface {
	AddressTrieNodeIterator

	// Remove removes the last iterated element from the underlying trie, and returns that element.
	// If there is no such element, it returns nil.
	Remove() *AddressTrieNode
}

type addrTrieNodeIteratorRem struct {
	tree.TrieNodeIteratorRem
}

func (iter addrTrieNodeIteratorRem) Next() *AddressTrieNode {
	return toAddressTrieNode(iter.TrieNodeIteratorRem.Next())
}

func (iter addrTrieNodeIteratorRem) Remove() *AddressTrieNode {
	return toAddressTrieNode(iter.TrieNodeIteratorRem.Remove())
}

type addrTrieNodeIterator struct {
	tree.TrieNodeIterator
}

func (iter addrTrieNodeIterator) Next() *AddressTrieNode {
	return toAddressTrieNode(iter.TrieNodeIterator.Next())
}

type CachingAddressTrieNodeIterator interface {
	AddressTrieNodeIteratorRem
	tree.CachingIterator
}

type cachingAddressTrieNodeIterator struct {
	tree.CachingTrieNodeIterator
}

func (iter cachingAddressTrieNodeIterator) Next() *AddressTrieNode {
	return toAddressTrieNode(iter.CachingTrieNodeIterator.Next())
}

func (iter cachingAddressTrieNodeIterator) Remove() *AddressTrieNode {
	return toAddressTrieNode(iter.CachingTrieNodeIterator.Remove())
}

//////////////////////////////////////////////////////////////////
//////

// AssociativeAddressTrieNodeIteratorRem iterates through an associative address trie, until both Next() returns nil and HasNext() returns false.
// It also allows you to remove the last visited node.
type AssociativeAddressTrieNodeIteratorRem interface {
	AssociativeAddressTrieNodeIterator

	// Remove removes the last iterated element from the underlying trie, and returns that element.
	// If there is no such element, it returns nil.
	Remove() *AssociativeAddressTrieNode
}

// AssociativeAddressTrieNodeIterator iterates through an associative address trie, until both Next() returns nil and HasNext() returns false
type AssociativeAddressTrieNodeIterator interface {
	HasNext

	Next() *AssociativeAddressTrieNode
}

type associativeAddressTrieNodeIteratorRem struct {
	AddressTrieNodeIteratorRem
}

func (iter associativeAddressTrieNodeIteratorRem) Next() *AssociativeAddressTrieNode {
	return iter.AddressTrieNodeIteratorRem.Next().ToAssociative()
}

func (iter associativeAddressTrieNodeIteratorRem) Remove() *AssociativeAddressTrieNode {
	return iter.AddressTrieNodeIteratorRem.Remove().ToAssociative()
}

type associativeAddressTrieNodeIterator struct {
	AddressTrieNodeIterator
}

func (iter associativeAddressTrieNodeIterator) Next() *AssociativeAddressTrieNode {
	return iter.AddressTrieNodeIterator.Next().ToAssociative()
}

type CachingAssociativeAddressTrieNodeIterator interface {
	AssociativeAddressTrieNodeIteratorRem
	tree.CachingIterator
}

type cachingAssociativeAddressTrieNodeIterator struct {
	CachingAddressTrieNodeIterator
}

func (iter cachingAssociativeAddressTrieNodeIterator) Next() *AssociativeAddressTrieNode {
	return iter.CachingAddressTrieNodeIterator.Next().ToAssociative()
}

func (iter cachingAssociativeAddressTrieNodeIterator) Remove() *AssociativeAddressTrieNode {
	return iter.CachingAddressTrieNodeIterator.Remove().ToAssociative()
}
