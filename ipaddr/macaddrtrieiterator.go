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

// MACTrieNodeIteratorRem iterates through a MAC address trie, until both Next returns nil and HasNext returns false.
// The iterator also allows you to remove the last visited node.
type MACTrieNodeIteratorRem interface {
	MACTrieNodeIterator

	// Remove removes the last iterated element from the underlying trie, and returns that element.
	// If there is no such element, it returns nil.
	Remove() *MACAddressTrieNode
}

// MACTrieNodeIterator iterates through a MAC address trie, until both Next returns nil and HasNext returns false
type MACTrieNodeIterator interface {
	hasNext

	Next() *MACAddressTrieNode
}

type macTrieNodeIteratorRem struct {
	AddressTrieNodeIteratorRem
}

func (iter macTrieNodeIteratorRem) Next() *MACAddressTrieNode {
	return iter.AddressTrieNodeIteratorRem.Next().ToMAC()
}

func (iter macTrieNodeIteratorRem) Remove() *MACAddressTrieNode {
	return iter.AddressTrieNodeIteratorRem.Remove().ToMAC()
}

type macTrieNodeIterator struct {
	AddressTrieNodeIterator
}

func (iter macTrieNodeIterator) Next() *MACAddressTrieNode {
	return iter.AddressTrieNodeIterator.Next().ToMAC()
}

type CachingMACTrieNodeIterator interface {
	MACTrieNodeIteratorRem
	tree.CachingIterator
}

type cachingMACTrieNodeIterator struct {
	CachingAddressTrieNodeIterator
}

func (iter cachingMACTrieNodeIterator) Next() *MACAddressTrieNode {
	return iter.CachingAddressTrieNodeIterator.Next().ToMAC()
}

func (iter cachingMACTrieNodeIterator) Remove() *MACAddressTrieNode {
	return iter.CachingAddressTrieNodeIterator.Remove().ToMAC()
}

//////////////////////////////////////////////////////////////////
//////

// MACAssociativeTrieNodeIteratorRem iterates through a MAC associative address trie, until both Next returns nil and HasNext returns false
type MACAssociativeTrieNodeIteratorRem interface {
	MACAssociativeTrieNodeIterator

	// Remove removes the last iterated element from the underlying trie, and returns that element.
	// If there is no such element, it returns nil.
	Remove() *MACAddressAssociativeTrieNode
}

// MACAssociativeTrieNodeIterator iterates through a MAC associative address trie, until both Next returns nil and HasNext returns false
type MACAssociativeTrieNodeIterator interface {
	hasNext

	Next() *MACAddressAssociativeTrieNode
}

type macAssociativeTrieNodeIteratorRem struct {
	AddressTrieNodeIteratorRem
}

func (iter macAssociativeTrieNodeIteratorRem) Next() *MACAddressAssociativeTrieNode {
	return iter.AddressTrieNodeIteratorRem.Next().ToMACAssociative()
}

func (iter macAssociativeTrieNodeIteratorRem) Remove() *MACAddressAssociativeTrieNode {
	return iter.AddressTrieNodeIteratorRem.Remove().ToMACAssociative()
}

type macAssociativeTrieNodeIterator struct {
	AddressTrieNodeIterator
}

func (iter macAssociativeTrieNodeIterator) Next() *MACAddressAssociativeTrieNode {
	return iter.AddressTrieNodeIterator.Next().ToMACAssociative()
}

type CachingMACAssociativeTrieNodeIterator interface {
	MACAssociativeTrieNodeIteratorRem
	tree.CachingIterator
}

type cachingMACAssociativeTrieNodeIterator struct {
	CachingAddressTrieNodeIterator
}

func (iter cachingMACAssociativeTrieNodeIterator) Next() *MACAddressAssociativeTrieNode {
	return iter.CachingAddressTrieNodeIterator.Next().ToMACAssociative()
}

func (iter cachingMACAssociativeTrieNodeIterator) Remove() *MACAddressAssociativeTrieNode {
	return iter.CachingAddressTrieNodeIterator.Remove().ToMACAssociative()
}

//////////////////////////////////////////////////////////////////
//////
