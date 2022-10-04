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

// IPv4TrieNodeIteratorRem iterates through an IPv4 address trie, until both Next returns nil and HasNext returns false
// The iterator also allows you to remove the last visited node.
type IPv4TrieNodeIteratorRem interface {
	IPv4TrieNodeIterator

	// Remove removes the last iterated element from the underlying trie, and returns that element.
	// If there is no such element, it returns nil.
	Remove() *IPv4AddressTrieNode
}

// IPv4TrieNodeIterator iterates through an IPv4 address trie, until both Next returns nil and HasNext returns false
type IPv4TrieNodeIterator interface {
	hasNext

	Next() *IPv4AddressTrieNode
}

type ipv4TrieNodeIteratorRem struct {
	AddressTrieNodeIteratorRem
}

func (iter ipv4TrieNodeIteratorRem) Next() *IPv4AddressTrieNode {
	return iter.AddressTrieNodeIteratorRem.Next().ToIPv4()
}

func (iter ipv4TrieNodeIteratorRem) Remove() *IPv4AddressTrieNode {
	return iter.AddressTrieNodeIteratorRem.Remove().ToIPv4()
}

type ipv4TrieNodeIterator struct {
	AddressTrieNodeIterator
}

func (iter ipv4TrieNodeIterator) Next() *IPv4AddressTrieNode {
	return iter.AddressTrieNodeIterator.Next().ToIPv4()
}

type CachingIPv4TrieNodeIterator interface {
	IPv4TrieNodeIteratorRem
	tree.CachingIterator
}

type cachingIPv4TrieNodeIterator struct {
	CachingAddressTrieNodeIterator
}

func (iter cachingIPv4TrieNodeIterator) Next() *IPv4AddressTrieNode {
	return iter.CachingAddressTrieNodeIterator.Next().ToIPv4()
}

func (iter cachingIPv4TrieNodeIterator) Remove() *IPv4AddressTrieNode {
	return iter.CachingAddressTrieNodeIterator.Remove().ToIPv4()
}

//////////////////////////////////////////////////////////////////
//////

// IPv4AssociativeTrieNodeIteratorRem iterates through an IPv4 associative address trie, until both Next returns nil and HasNext returns false.
// The iterator also allows you to remove the last added node.
type IPv4AssociativeTrieNodeIteratorRem interface {
	IPv4AssociativeTrieNodeIterator

	// Remove removes the last iterated element from the underlying trie, and returns that element.
	// If there is no such element, it returns nil.
	Remove() *IPv4AddressAssociativeTrieNode
}

// IPv4AssociativeTrieNodeIterator iterates through an IPv4 associative address trie, until both Next returns nil and HasNext returns false
type IPv4AssociativeTrieNodeIterator interface {
	hasNext

	Next() *IPv4AddressAssociativeTrieNode
}

type ipv4AssociativeTrieNodeIteratorRem struct {
	AddressTrieNodeIteratorRem
}

func (iter ipv4AssociativeTrieNodeIteratorRem) Next() *IPv4AddressAssociativeTrieNode {
	return iter.AddressTrieNodeIteratorRem.Next().ToIPv4Associative()
}

func (iter ipv4AssociativeTrieNodeIteratorRem) Remove() *IPv4AddressAssociativeTrieNode {
	return iter.AddressTrieNodeIteratorRem.Remove().ToIPv4Associative()
}

type ipv4AssociativeTrieNodeIterator struct {
	AddressTrieNodeIterator
}

func (iter ipv4AssociativeTrieNodeIterator) Next() *IPv4AddressAssociativeTrieNode {
	return iter.AddressTrieNodeIterator.Next().ToIPv4Associative()
}

type CachingIPv4AssociativeTrieNodeIterator interface {
	IPv4AssociativeTrieNodeIteratorRem
	tree.CachingIterator
}

type cachingIPv4AssociativeTrieNodeIterator struct {
	CachingAddressTrieNodeIterator
}

func (iter cachingIPv4AssociativeTrieNodeIterator) Next() *IPv4AddressAssociativeTrieNode {
	return iter.CachingAddressTrieNodeIterator.Next().ToIPv4Associative()
}

func (iter cachingIPv4AssociativeTrieNodeIterator) Remove() *IPv4AddressAssociativeTrieNode {
	return iter.CachingAddressTrieNodeIterator.Remove().ToIPv4Associative()
}

//////////////////////////////////////////////////////////////////
//////
