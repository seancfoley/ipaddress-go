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

import "github.com/seancfoley/ipaddress-go/ipaddr/tree"

// IPv6TrieNodeIteratorRem iterates through an IPv6 address trie, until both Next() returns nil and HasNext() returns false.
// The iterator also allows you to remove the last visited node.
type IPv6TrieNodeIteratorRem interface {
	IPv6TrieNodeIterator

	// Remove removes the last iterated element from the underlying trie, and returns that element.
	// If there is no such element, it returns nil.
	Remove() *IPv6AddressTrieNode
}

// IPv6TrieNodeIterator iterates through an IPv6 address trie, until both Next() returns nil and HasNext() returns false
type IPv6TrieNodeIterator interface {
	HasNext

	Next() *IPv6AddressTrieNode
}

type ipv6TrieNodeIteratorRem struct {
	AddressTrieNodeIteratorRem
}

func (iter ipv6TrieNodeIteratorRem) Next() *IPv6AddressTrieNode {
	return iter.AddressTrieNodeIteratorRem.Next().ToIPv6()
}

func (iter ipv6TrieNodeIteratorRem) Remove() *IPv6AddressTrieNode {
	return iter.AddressTrieNodeIteratorRem.Remove().ToIPv6()
}

type ipv6TrieNodeIterator struct {
	AddressTrieNodeIterator
}

func (iter ipv6TrieNodeIterator) Next() *IPv6AddressTrieNode {
	return iter.AddressTrieNodeIterator.Next().ToIPv6()
}

type CachingIPv6TrieNodeIterator interface {
	IPv6TrieNodeIteratorRem
	tree.CachingIterator
}

type cachingIPv6TrieNodeIterator struct {
	CachingAddressTrieNodeIterator
}

func (iter cachingIPv6TrieNodeIterator) Next() *IPv6AddressTrieNode {
	return iter.CachingAddressTrieNodeIterator.Next().ToIPv6()
}

func (iter cachingIPv6TrieNodeIterator) Remove() *IPv6AddressTrieNode {
	return iter.CachingAddressTrieNodeIterator.Remove().ToIPv6()
}

//////////////////////////////////////////////////////////////////
//////

// IPv6AssociativeTrieNodeIteratorRem iterates through an IPv6 associative address trie, until both Next() returns nil and HasNext() returns false.
// The iterator also allows you to remove the last visited node.
type IPv6AssociativeTrieNodeIteratorRem interface {
	IPv6AssociativeTrieNodeIterator

	// Remove removes the last iterated element from the underlying trie, and returns that element.
	// If there is no such element, it returns nil.
	Remove() *IPv6AddressAssociativeTrieNode
}

// IPv6AssociativeTrieNodeIterator iterates through an IPv6 associative address trie, until both Next() returns nil and HasNext() returns false
type IPv6AssociativeTrieNodeIterator interface {
	HasNext

	Next() *IPv6AddressAssociativeTrieNode
}

type ipv6AssociativeTrieNodeIteratorRem struct {
	AddressTrieNodeIteratorRem
}

func (iter ipv6AssociativeTrieNodeIteratorRem) Next() *IPv6AddressAssociativeTrieNode {
	return iter.AddressTrieNodeIteratorRem.Next().ToIPv6Associative()
}

func (iter ipv6AssociativeTrieNodeIteratorRem) Remove() *IPv6AddressAssociativeTrieNode {
	return iter.AddressTrieNodeIteratorRem.Remove().ToIPv6Associative()
}

type ipv6AssociativeTrieNodeIterator struct {
	AddressTrieNodeIterator
}

func (iter ipv6AssociativeTrieNodeIterator) Next() *IPv6AddressAssociativeTrieNode {
	return iter.AddressTrieNodeIterator.Next().ToIPv6Associative()
}

type CachingIPv6AssociativeTrieNodeIterator interface {
	IPv6AssociativeTrieNodeIteratorRem
	tree.CachingIterator
}

type cachingIPv6AssociativeTrieNodeIterator struct {
	CachingAddressTrieNodeIterator
}

func (iter cachingIPv6AssociativeTrieNodeIterator) Next() *IPv6AddressAssociativeTrieNode {
	return iter.CachingAddressTrieNodeIterator.Next().ToIPv6Associative()
}

func (iter cachingIPv6AssociativeTrieNodeIterator) Remove() *IPv6AddressAssociativeTrieNode {
	return iter.CachingAddressTrieNodeIterator.Remove().ToIPv6Associative()
}

//////////////////////////////////////////////////////////////////
//////
