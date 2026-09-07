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

import "math/big"

// AddressAggregation represents any type that can represent multiple individual addresses.
// This includes all address types, since they can represent subnets by making use of ranges within each segment.
// This includes sequential ranges of addresses.
// This includes collections of addresses.
// The interface is satisfied by all these types: *Address, *MACAddress, *IPAddress,
// *IPAddressSeqRange, *IPAddressSeqRangeList, *IPAddressContainmentTrie,
// *IPv4Address, *IPv4AddressSeqRange, *IPv4AddressSeqRangeList, *IPv4AddressContainmentTrie,
// *IPv6Address, *IPv6AddressSeqRange, *IPv6AddressSeqRangeList, *IPv6AddressContainmentTrie
type AddressAggregation interface { // this is the type used by equality in containers (and maybe others) - cannot combine with AddressItemAggregation
	AddressItemAggregation

	AddressIterator() Iterator[AddressType]

	// Contains returns whether this aggregation contains all individual addresses in the given address or subnet.
	Contains(AddressType) bool

	// Enumerate indicates where an address sits relative to the address ordering.
	//
	// It determines how many individual address elements precede the given address element, if the address is in the aggregation.
	// If above all addresses in the aggregation, it is the distance to the upper boundary added to the aggregation count less one, and if below the aggregation, the distance to the lower boundary.
	//
	// In other words, if the given address is not in the aggregation but above it, returns the number of addresses preceding the address from the upper aggregation boundary,
	// added to one less than the total number of aggregation addresses.  If the given address is not in the aggregation but below it, returns the number of addresses following the address to the lower aggregation boundary.
	//
	// If the argument is not in the aggregation, but neither above nor below it, then nil is returned.
	//
	// Enumerate returns nil when the argument is multi-valued. The argument must be an individual address.
	//
	// When this aggregation happens to be an individual address, the returned value is the distance (difference) between the two addresses.
	//
	// If the given address does not have the same version or type as the addresses in this aggregation, then nil is returned.
	Enumerate(AddressType) *big.Int

	// OverlapsAddr returns whether this aggregation contains any individual addresses in the given address or subnet.
	OverlapsAddr(AddressType) bool

	// EqualAggregation returns true if and only if this aggregation of addresses are the same as the=ose in the given aggregation.
	EqualAggregation(AddressAggregation) bool
}

// IsEmpty returns true if the aggregation has no elements.
// Much like the len function, it handles nil, returning true for nil interfaces and nil pointer types.
func IsEmpty(aggregation AddressAggregation) bool {
	switch other := aggregation.(type) {
	case nil:
		return true
	case *IPAddress:
		return other == nil
	case *IPv4Address:
		return other == nil
	case *IPv6Address:
		return other == nil
	case *MACAddress:
		return other == nil
	case *IPAddressSeqRange:
		return other == nil
	case *IPv4AddressSeqRange:
		return other == nil
	case *IPv6AddressSeqRange:
		return other == nil
	case *IPAddressContainmentTrie:
		return other == nil || other.IsEmpty()
	case *IPv4AddressContainmentTrie:
		return other == nil || other.IsEmpty()
	case *IPv6AddressContainmentTrie:
		return other == nil || other.IsEmpty()
	case *IPAddressSeqRangeList:
		return other == nil || other.IsEmpty()
	case *IPv4AddressSeqRangeList:
		return other == nil || other.IsEmpty()
	case *IPv6AddressSeqRangeList:
		return other == nil || other.IsEmpty()
	case AddressType, IPAddressSeqRangeType:
		return false
	case IPAddressCollection:
		return other.IsEmpty()
	default:
		return other.GetCount().Sign() == 0
	}
}

var _, _, _, _, _, _, _, _, _, _, _, _, _,
	_ AddressAggregation = &Address{},
	&MACAddress{},
	&IPAddress{}, &IPAddressSeqRange{}, &IPAddressSeqRangeList{}, &IPAddressContainmentTrie{},
	&IPv4Address{}, &IPv4AddressSeqRange{}, &IPv4AddressSeqRangeList{}, &IPv4AddressContainmentTrie{},
	&IPv6Address{}, &IPv6AddressSeqRange{}, &IPv6AddressSeqRangeList{}, &IPv6AddressContainmentTrie{}

// IPAddressAggregation represents any type that can represent multiple individual IP addresses.
// This includes all IP address types, since they can represent IP address subnets by making use of ranges within each segment,
// which also includes the representation of CIDR prefix block subnets.
// This includes sequential ranges of IP addresses.
// This includes collections of IP addresses.
// The interface is satisfied by all these types: *IPAddress,
// *IPAddressSeqRange, *IPAddressSeqRangeList, *IPAddressContainmentTrie,
// *IPv4Address, *IPv4AddressSeqRange, *IPv4AddressSeqRangeList, *IPv4AddressContainmentTrie,
// *IPv6Address, *IPv6AddressSeqRange, *IPv6AddressSeqRangeList, *IPv6AddressContainmentTrie
type IPAddressAggregation interface {
	AddressAggregation

	// ContainsRange returns whether all the addresses in the given sequential range are also contained in this aggregation of addresses.
	ContainsRange(IPAddressSeqRangeType) bool

	// OverlapsRange returns whether this aggregation includes any of the addresses in the given sequential range, if there is at least one individual address common to both.
	OverlapsRange(IPAddressSeqRangeType) bool
}

var _, _, _, _, _, _, _, _, _, _, _,
	_ IPAddressAggregation = &IPAddress{}, &IPAddressSeqRange{}, &IPAddressSeqRangeList{}, &IPAddressContainmentTrie{},
	&IPv4Address{}, &IPv4AddressSeqRange{}, &IPv4AddressSeqRangeList{}, &IPv4AddressContainmentTrie{},
	&IPv6Address{}, &IPv6AddressSeqRange{}, &IPv6AddressSeqRangeList{}, &IPv6AddressContainmentTrie{}

// IPAddressAggregationConstraint constrains IPAddressAggregation, restricting it to a single generic address type,
// rather than representing any one of multiple IP address types.
// At the same time, IPAddressAggregationConstraint expands the available methods beyond those offered by IPAddressAggregation.
// It is particularly useful to provide full functionality in methods using generic address types.
// Use this type as a generic type constraint to retain full access to all IP address aggregation functionality in your generic function or method.
// The type T can be any one of *IPAddress, *IPv4Address, or *IPv6Address
// The interface is satisfied by all these types: *IPAddress,
// *IPAddressSeqRange, *IPAddressSeqRangeList, *IPAddressContainmentTrie,
// *IPv4Address, *IPv4AddressSeqRange, *IPv4AddressSeqRangeList, *IPv4AddressContainmentTrie,
// *IPv6Address, *IPv6AddressSeqRange, *IPv6AddressSeqRangeList, *IPv6AddressContainmentTrie
type IPAddressAggregationConstraint[T IPAddressTypeConstraint[T]] interface {
	IPAddressAggregation

	// Get returns the address at the given index into the sorted collection.
	// The index of zero returns the first address.
	//
	// If the index is negative, or the index exceeds GetCount() - 1, Get will panic.  It is much like indexing a slice or array.
	Get(int64) T

	// GetBig returns the address at the given index into the sorted collection.
	// The index of zero returns the first address.
	//
	// If the index is negative, or the index exceeds GetCount() - 1, Get will panic.  It is much like indexing a slice or array.
	GetBig(*big.Int) T

	// Iterator returns an iterator to iterate through the individual addressesin the collection in order.
	//
	// Use the function ipaddr.StdPushIterator to convert the returned iterator to a standard library iter.Seq
	Iterator() Iterator[T]

	// SpanningPrefixBlockIterator returns an iterator to iterate over the minimal set of prefix blocks that spans the aggregation of addresses, no less and no more, in order
	SpanningPrefixBlockIterator() Iterator[T]

	// SpanningPrefixBlockIterator returns an iterator to iterate over the minimal set of sequential blocks that spans the aggregation of addresses, no less and no more, in order
	SpanningSeqBlockIterator() Iterator[T]

	// GetLower returns the individual address with the lowest numeric value in the collection
	GetLower() T

	// GetLower returns the individual address with the highest numeric value in the collection
	GetUpper() T

	// GetLowerAndUpper returns the individual addresses with the lowest and highest numeric values in the collection
	GetLowerAndUpper() (lower, upper T)

	// CoverWithSequentialRange returns the unique sequential range of minimal size that includes all the addresses in this collection.
	// If there are no addresses in this collection, then nil is returned.
	//
	// The result will represent the same set of addresses if and only if the set of addresses in this collection are sequential, in which case IsSequential returns true.
	CoverWithSequentialRange() *SequentialRange[T]

	// CoverWithPrefixBlock returns the unique CIDR prefix block subnet or individual address of minimal size that includes all the addresses in this collection.
	// If there are no addresses in this collection, then nil is returned.
	CoverWithPrefixBlock() T
}

// verify we can assign constraints for IP addresses, IP sequential ranges, and IP address collections, to IPAddressAggregationConstraint
func f[T IPAddressTypeConstraint[T]]() (x IPAddressAggregationConstraint[T]) {
	var s *SequentialRange[T]
	var i IPAddressTypeConstraint[T]
	var c IPAddressCollAddrConstraint[T]
	//var r IPAddressRange
	//x = r
	x = s
	x = i
	x = c
	return x
}

var (
	_, _, _, _ IPAddressAggregationConstraint[*IPAddress]   = &IPAddress{}, &IPAddressSeqRange{}, &IPAddressSeqRangeList{}, &IPAddressContainmentTrie{}
	_, _, _, _ IPAddressAggregationConstraint[*IPv4Address] = &IPv4Address{}, &IPv4AddressSeqRange{}, &IPv4AddressSeqRangeList{}, &IPv4AddressContainmentTrie{}
	_, _, _, _ IPAddressAggregationConstraint[*IPv6Address] = &IPv6Address{}, &IPv6AddressSeqRange{}, &IPv6AddressSeqRangeList{}, &IPv6AddressContainmentTrie{}

	_ = f[*IPAddress]()
)

// IPAddressCollection represents an arbitrary collection of IP addresses.
// Unlike IPAddressAggregation, the collection need not follow any pattern or limitation,
// such as being sequential like IP address sequential ranges,
// or being representable by ranges within each segment, like the address types.
// A collection may contain any arbitrary set of IP addresses.
// The difference between collections is the underlying data structures used to accomplish that objective.
// This interface is satisfied by these types: *IPAddressSeqRangeList, *IPAddressContainmentTrie,
// *IPv4AddressSeqRangeList, *IPv4AddressContainmentTrie,
// *IPv6AddressSeqRangeList, *IPv6AddressContainmentTrie
type IPAddressCollection interface {
	IPAddressAggregation

	// IsEmpty returns true if the collection is empty
	IsEmpty() bool

	// Clear empties the collection
	Clear()
}

var _, _, _, _, _, _ IPAddressCollection = &IPAddressSeqRangeList{}, &IPAddressContainmentTrie{},
	&IPv4AddressSeqRangeList{}, &IPv4AddressContainmentTrie{},
	&IPv6AddressSeqRangeList{}, &IPv6AddressContainmentTrie{}

// IPAddressCollAddrConstraint constrains IPAddressCollection, restricting it to a single generic IP address type,
// rather than representing any one of multiple IP address types.
// At the same time, IPAddressCollAddrConstraint expands the available methods beyond those offered by IPAddressCollection.
// It is particularly useful to provide additioal functionality over IPAddressCollection and IPAddressAggregationConstraint in methods using generic address types.
// Use this type as a generic type constraint to retain full access to all IP address collection functionality in your generic function or method.
// The type T can be any one of *IPAddress, *IPv4Address, or *IPv6Address
// This interface is satisfied by these types: *IPAddressSeqRangeList, *IPAddressContainmentTrie,
// *IPv4AddressSeqRangeList, *IPv4AddressContainmentTrie,
// *IPv6AddressSeqRangeList, *IPv6AddressContainmentTrie
type IPAddressCollAddrConstraint[T IPAddressTypeConstraint[T]] interface {
	IPAddressAggregationConstraint[T]

	IPAddressCollection

	// OverlapsAddress returns true if and only the given individual address or subnet contains at least one individual address that is also in the collection.
	OverlapsAddress(T) bool

	// OverlapsSeqRange returns true if and only if the given sequential range contains at least one address that is also in the collection.
	OverlapsSeqRange(*SequentialRange[T]) bool

	// EnumerateAddress returns the distance of the given address from the initial and lowest address in the collection.  It indicates where an address sits relative to the collection ordering.
	//
	//	If within or above the addresses in collection, it is the distance to the lower boundary of the collection.  If below the collection, it returns the number of addresses following the address to the initial address in the collection, as a negative number.
	//
	// You can call Contains or you can compare with GetCount to check for containment.
	// An IP address is in the collection if 0 <= Enumerate(IP Address) < GetCount.
	//
	// If the address is above the lower boundary and below the upper boundary of the collection, but is not within a prefix block in the collection, then this method returns nil.
	//
	// Returns nil when the argument is a multi-valued subnet. The argument must be an individual address.
	//
	// Returns nil when the collection is empty.
	//
	// Returns nil when the address version of the given address does not match the addresses in this collection.
	EnumerateAddress(T) *big.Int

	// ContainsAddress returns true if and only if this collection contains all the individual addresses in the given address or subnet.
	ContainsAddress(T) bool

	// ContainsSeqRange returns true if and only if this collection contains all the individual addresses in the given sequential range.
	ContainsSeqRange(*SequentialRange[T]) bool

	// Add adds the address to the collection, if not already in the collection.
	//
	// If the address version does match existing addresses in the collection, the address is not added.
	//
	// Returns whether addresses were added, whether the collection was changed.
	Add(T) bool

	// Remove removes the given address from the collection.  It returns true if the collection was changed.
	// It returns false if the address was not in the collection.
	Remove(T) bool

	// AddSeqRange sdds the addresses in the sequential range to the collection, if not already in the collection.
	//
	// If the address version of the addresses in the collection does match the version of addresses in the given range, this method panics.
	//
	// Returns whether at least one address in the given sequential range was added, whether the collection was changed.
	AddSeqRange(*SequentialRange[T]) bool

	// RemoveSeqRange removes all the addresses in the sequential range from the collection.
	// Returns true if the collection was changed.
	RemoveSeqRange(*SequentialRange[T]) bool

	// RemoveAt removes the individual address at the given index into the lists of addresses.  Returns that address.
	// Similar to Get but also removes the address found.
	//
	// If the index is negative or larger than GetCount() - 1, this method panics.
	RemoveAt(int64) T

	// RemoveAt removes the individual address at the given index into the lists of addresses.  Returns that address.
	// Similar to GetBig but also removes the address found.
	//
	// If the index is negative or larger than GetCount() - 1, this method panics.
	RemoveAtBig(*big.Int) T

	// Lower returns the highest address in the collection strictly less than the lowest address in the given address or subnet.
	Lower(T) T

	// Floor returns the highest address in the collection less than or equal to the lowest address in the given address or subnet.
	Floor(T) T

	// Higher returns the lowest address in the collection strictly greater than the highest address in the given address or subnet.
	Higher(T) T

	// Ceiling returns the lowest address in the collection greater than or equal to the highest address in the given address or subnet.
	Ceiling(T) T

	// SpanningSeqRangeIterator returns an iterator for iterating through the minimal set of disjoint sequential ranges containing the addresses in this collection of addresses.
	SpanningSeqRangeIterator() Iterator[*SequentialRange[T]]
}

var (
	_, _ IPAddressCollAddrConstraint[*IPAddress]   = &IPAddressSeqRangeList{}, &IPAddressContainmentTrie{}
	_, _ IPAddressCollAddrConstraint[*IPv4Address] = &IPv4AddressSeqRangeList{}, &IPv4AddressContainmentTrie{}
	_, _ IPAddressCollAddrConstraint[*IPv6Address] = &IPv6AddressSeqRangeList{}, &IPv6AddressContainmentTrie{}
)

//TODO LATER consider possibly adding ContainsCollection and OverlapsCollection, much like you have EqualAggregation which works with all aggregation types.
// I think ContainsAggregation and OverlapsAggregation is likely going too far, you don't want to search for collections inside seq ranges or subnets
// But allowing for the checking of seq range list inside containment trie or vice versa, perhaps that is worthwhile

// IPAddressCollConstraint further constrains IPAddressCollAddrConstraint.
// It has a constraint for itself, as well as for the IP address type.
// It expands the available methods beyond those offered by IPAddressCollAddrConstraint and IPAddressCollection.
// It is particularly useful to provide full functionality in methods using generic address types.
// Use this type as a generic type constraint to retain full access to all IP address collection functionality in your generic function or method.
// The type T can be any one of *IPAddress, *IPv4Address, or *IPv6Address.
// The type S can be either IPAddressSeqRangeList or IPAddressContainmentTrie.
// This interface is satisfied by these types: *IPAddressSeqRangeList, *IPAddressContainmentTrie,
// *IPv4AddressSeqRangeList, *IPv4AddressContainmentTrie,
// *IPv6AddressSeqRangeList, *IPv6AddressContainmentTrie
type IPAddressCollConstraint[S IPAddressCollAddrConstraint[T], T IPAddressTypeConstraint[T]] interface {
	IPAddressCollAddrConstraint[T]

	// Clone makes a copy of the collection
	Clone() S

	// NewEmpty creates a new ContainmentTrie using the same element type T
	NewEmpty() S

	// Equal returns true if and only if this collection has the same set of individual addresses as the given collectioj
	Equal(S) bool

	// ComplementIntoNew returns a new collection comprising all the addresses not contained in this collection.
	//
	// If this list is empty and is not restricted to a single IP version of IPv4 or IPv6,
	// then the IP version is ambiguous, so the complement is indeterminate, in which case nil is returned.
	ComplementIntoNew() S

	// JoinIntoNew creates a new containment trie that has all addresses in this containment trie and the provided containment trie.
	JoinIntoNew(S) S

	// RemoveIntoNew produces a new containment trie that has the addresses in this collection that are not in the given collection.
	RemoveIntoNew(S) S

	// IntersectIntoNew produces a new containment tries that is the intersection of this collection with the given collection.
	IntersectIntoNew(S) S

	// ContainsOther returns whether this collection contains all addresses in the given collection
	ContainsOther(S) bool

	// OverlapsOther returns whether there is any overlap of this collection with the given collection
	OverlapsOther(S) bool
}

var (
	_ IPAddressCollConstraint[*IPAddressSeqRangeList, *IPAddress]     = &IPAddressSeqRangeList{}
	_ IPAddressCollConstraint[*IPv4AddressSeqRangeList, *IPv4Address] = &IPv4AddressSeqRangeList{}
	_ IPAddressCollConstraint[*IPv6AddressSeqRangeList, *IPv6Address] = &IPv6AddressSeqRangeList{}

	_ IPAddressCollConstraint[*IPAddressContainmentTrie, *IPAddress]     = &IPAddressContainmentTrie{}
	_ IPAddressCollConstraint[*IPv4AddressContainmentTrie, *IPv4Address] = &IPv4AddressContainmentTrie{}
	_ IPAddressCollConstraint[*IPv6AddressContainmentTrie, *IPv6Address] = &IPv6AddressContainmentTrie{}
)

//TODO NEXT, another release with the improvements below, so I can then finish the wiki examples
// - Improved the string produced by ToString() of SequentialRangeList for an improved visual representation of the list
// - small changes to framework interfaces: AddressType implements all of AddressAggregation, IPAddressType implements all of IPAddressAggregation, small correction to generic parameter for ConvertAddressType
// - renamed ContainmentTrieBase to ContainmentTrie, retained an alias for ContainmentTrieBase

//TODO need at least 1.19 if I do the atomicStorePointer change
//1.23 is needed for alias of ContainmentTrie

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

//  // TODO  ContainsProper - even though I don't need it in here anymore :-)
//  /*
//      for each segment
//          if lower >= otherLower && upper <= otherUpper {
//                  if lower > otherLower || upper < otherUpper {
//                      for each segment
//                          if lower < otherLower && upper > otherUpper {
//                              return false
//                      }
//                      return true
//                  }
//          } else {
//              return false
//          }
//      }
//      return false
//  */
// however, contains plus compare counts does the trick as well, also contains and equals

// DONE the wiki examples I did in Java
//
// diff
// https://github.com/seancfoley/IPAddress/wiki/Code-Examples-2:-Subnet-Containment,-Matching,-Comparing/_compare/6f56ff179cd18b8545f19ba913dacc2df5be7ede...f7aa6bdaa2329fb56ae9fc93dc64af4c568bd73d
//
// altered:
// DONE selectarbitraryrange.go - ## Select Addresses and Subnets within Arbitrary Address Range
// DONE near.go - ## Select Address Closest to Arbitrary Address
// NAH - https://github.com/seancfoley/IPAddress/wiki/Code-Examples-2:-Subnet-Containment,-Matching,-Comparing#select-address-ranges-intersecting-with-arbitrary-range
//      This is stupid: put the range in a list, then go through them
//      Is there a way to put them all in the collection?
//      Not really
//      In fact, this one, funny enough, seems to not fit with collections at all
//
// new:
// just create some lists, do the intersection:
// DONE * [Select Addresses Common to Lists of Subnets or Address Ranges](https://github.com/seancfoley/IPAddress/wiki/Code-Examples-2:-Subnet-Containment,-Matching,-Comparing#select-addresses-common-to-lists-of-subnets-or-address-ranges)
// see main.go in here and common.go in samples, I have some extra generic code
// I want to create a new verison with my new iteration options in the collection types, then use them in this example.
// so that is ready to do now, see main.go
// DONE I also need to fix up the printing of sequential range list in there
//
// diff
// https://github.com/seancfoley/IPAddress/wiki/Code-Examples-3:-Subnetting-and-Other-Subnet-Operations/_compare/1bd838508ebbd74dc512c43cfbde05e5975ba996...19e899aa3df9919909257f786821c02aa2493fa1
//
//
// DONE, again written in common.go at the bottom  * [Remove Lists of Subnets or Address Ranges from a Collection of IP Addresses](https://github.com/seancfoley/IPAddress/wiki/Code-Examples-3:-Subnetting-and-Other-Subnet-Operations#remove-lists-of-subnets-or-address-ranges-from-a-collection-of-ip-addresses)
// THis one builds on the intersect one above. see main.go in here and common.go in samples, I have some extra generic code
//
// DONE * [Find the Complement of a Collection of Subnets within a Larger Subnet](https://github.com/seancfoley/IPAddress/wiki/Code-Examples-3:-Subnetting-and-Other-Subnet-Operations#find-the-complement-of-a-collection-of-subnets-within-a-larger-subnet)
// DONE * [De Morgan's Laws of Set Theory](https://github.com/seancfoley/IPAddress/wiki/Code-Examples-3:-Subnetting-and-Other-Subnet-Operations#de-morgans-laws-of-set-theory)
//
// DONE Also, in Java I replaced calls to toSequentialRange with either coverWithSequentialRange or spanWithRange
//
// DONE add to the two first examples in section 3 the use of StdPushIterator to get iter.Seq
//
// DONE MAYBE a wiki example that uses IPAddressCollConstraint for polymorphism of collections - see testCollectionBooleanOpSingleAddress or testCollectionOpSingleAddress
//  "you just need to specify the address type, but you can make that polymporhpic as well"
//  show a func
// I've already altered a couple, and will also do DeMorgan's laws that way. but perhaps you might want to think of something inventive and new?  Not sure

// DONE emulate the Java side, on the seocnd example page there are 3 examples I divide as option 1 / 2.  Do the same with go.  It makes it easier to read.  as far as I can tell, I did not refactor any others (the others with options are the new ones I am adding)

// TODO great ideas that I might want to add to Java:
// - properContains which is like contains and not equal
// - all the new node-based add methods (add, addNode, put, putNode) that check prefix first, allowing us to add to nodes directly
// - the new logic for removeBlock above which finds the interesecting node, gets the parent, removes the intersecting node, adds back the non-intersecting pieces directly to the node
//      replaces the logic of a remove that returns the parent and the removed node at the same time
// - IsUpperAdjacentTo instead of using DecrementSingle/IncrementSingle to compare.
// - Enumerate with tries.  The count prior to a node can be acquired by backtracking in the trie and using MatchingAddressCount.  Then add Enumerate to IPAddressAggregation.
// - Increment with tries.  Uses MatchingAddressCount.
// - the new collection methods in ContainmentTrie, then add them to collection, deprecating old names in IPAddressSeqRangeList and deferring to new names:
//      ComplementIntoNew() S
//      JoinIntoNew(S) S
//      RemoveIntoNew(S) S
//      IntersectIntoNew(S) S
//      ContainsOther(S) bool
//      OverlapsOther(S) bool
// - in fact, all the stuff I added to ContainmentTrie
// - I made an optimization to seg range list binary search in which I return two bools indicating if the searched address landed on a range boundary
//      Not sure I can port it to Java, but who knows, maybe I can somehow, for instance if I passed in the pneding range object as an interface could store it there
//  - If I change the behaviour of increment here for range lists, then maybe do the same, even though not backwards compatible?  Not sure about this one.
//
