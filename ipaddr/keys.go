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

import "fmt"

// RangeBoundaryKey represents an address key used within a sequential range key.
type RangeBoundaryKey[T any] interface {
	ToAddress() T
}

// SequentialRangeKey is a representation of SequentialRange that is comparable as defined by the language specification.
// See https://go.dev/ref/spec#Comparison_operators
//
// It can be used as a map key.
// The zero value is a range from a zero-length address to itself.
type SequentialRangeKey[T SequentialRangeConstraint[T]] struct {
	lowerKey, upperKey RangeBoundaryKey[T]
}

// ToSeqRange converts back to a sequential range instance
func (key SequentialRangeKey[T]) ToSeqRange() *SequentialRange[T] {
	lower := key.lowerKey.ToAddress()
	upper := key.upperKey.ToAddress()
	return newSequRangeUnchecked(lower, upper, lower != upper)
}

// String calls the String method in the corresponding sequential range
func (key SequentialRangeKey[T]) String() string {
	return key.ToSeqRange().String()
}

// GetLowerKey returns the lower key of the pair of address keys comprising this sequential range key.
func (key SequentialRangeKey[T]) GetLowerKey() RangeBoundaryKey[T] {
	return key.lowerKey
}

// GetUpperKey returns the upper key of the pair of address keys comprising this sequential range key.
func (key SequentialRangeKey[T]) GetUpperKey() RangeBoundaryKey[T] {
	return key.upperKey
}

// IPv4AddressKey is a representation of an IPv4 address that is comparable as defined by the language specification.
// See https://go.dev/ref/spec#Comparison_operators
//
// It can be used as a map key.  It can be obtained from its originating address instances.
// The zero value corresponds to the zero-value for IPv4Address.
// Keys do not incorporate prefix length to ensure that all equal addresses have equal keys.
// To create a key that has prefix length, combine into a struct with the PrefixKey obtained by passing the address into PrefixKeyFrom.
// IPv4Address can be compared using the Compare or Equal methods, or using an AddressComparator.
type IPv4AddressKey struct {
	vals uint64 // upper and lower combined into one uint64
}

// ToAddress converts back to an address instance
func (key IPv4AddressKey) ToAddress() *IPv4Address {
	return fromIPv4Key(key)
}

// String calls the String method in the corresponding address
func (key IPv4AddressKey) String() string {
	return key.ToAddress().String()
}

// IPv4KeyFromRangeKey converts the IPv4 range key to an IPv4AddressKey.
// The IPv4AddressKey satisfies the "comparable" generic constraint, unlike RangeBoundaryKey.
// Both key types are "comparable" with respect to comparison operators and map keys.
func IPv4KeyFromRangeKey(key RangeBoundaryKey[*IPv4Address]) IPv4AddressKey {
	if conv, ok := key.(IPv4AddressKey); ok {
		return conv
	}
	var addr *IPv4Address = key.ToAddress()
	return addr.ToKey()
}

type testComparableConstraint[T comparable] struct{}

var (
	// ensure our 5 key types are indeed comparable
	_ testComparableConstraint[IPv4AddressKey]
	_ testComparableConstraint[IPv6AddressKey]
	_ testComparableConstraint[MACAddressKey]
	_ testComparableConstraint[Key[*IPAddress]]
	_ testComparableConstraint[Key[*Address]]
	//_ testComparableConstraint[RangeBoundaryKey[*IPv4Address]] // does not compile, as expected, because it has an interface field.  But it is still go-comparable.
)

func testCompare() {
	var one1, two1 IPv4AddressKey
	_ = one1 == two1 // comparable
	var one2, two2 RangeBoundaryKey[*IPv4Address]
	_ = one2 == two2 // comparable
}

// IPv6AddressKey is a representation of an IPv6 address that is comparable as defined by the language specification.
// See https://go.dev/ref/spec#Comparison_operators
//
// It can be used as a map key.  It can be obtained from its originating address instances.
// The zero value corresponds to the zero-value for IPv6Address.
// Keys do not incorporate prefix length to ensure that all equal addresses have equal keys.
// To create a key that has prefix length, combine into a struct with the PrefixKey obtained by passing the address into PrefixKeyFrom.
// IPv6Address can be compared using the Compare or Equal methods, or using an AddressComparator.
type IPv6AddressKey struct {
	keyContents
}

// ToAddress converts back to an address instance
func (key IPv6AddressKey) ToAddress() *IPv6Address {
	return fromIPv6Key(key)
}

// String calls the String method in the corresponding address
func (key IPv6AddressKey) String() string {
	return key.ToAddress().String()
}

// IPv6KeyFromRangeKey converts the IPv6 range key to an IPv6AddressKey.
// The IPv6AddressKey satisfies the "comparable" generic constraint, unlike RangeBoundaryKey.
// Both key types are "comparable" with respect to comparison operators and map keys.
func IPv6KeyFromRangeKey(key RangeBoundaryKey[*IPv6Address]) IPv6AddressKey {
	if conv, ok := key.(IPv6AddressKey); ok {
		return conv
	}
	var addr *IPv6Address = key.ToAddress()
	return addr.ToKey()
}

// MACAddressKey is a representation of a MAC address that is comparable as defined by the language specification.
// See https://go.dev/ref/spec#Comparison_operators
//
// It can be used as a map key.  It can be obtained from its originating address instances.
// The zero value corresponds to the zero-value for MACAddress.
// Keys do not incorporate prefix length to ensure that all equal addresses have equal keys.
// To create a key that has prefix length, combine into a struct with the PrefixKey obtained by passing the address into PrefixKeyFrom.
// MACAddress can be compared using the Compare or Equal methods, or using an AddressComparator.
type MACAddressKey struct {
	vals struct {
		lower,
		upper uint64
	}
	additionalByteCount uint8 // 0 for MediaAccessControlSegmentCount or 2 for ExtendedUniqueIdentifier64SegmentCount
}

// ToAddress converts back to an address instance
func (key MACAddressKey) ToAddress() *MACAddress {
	return fromMACKey(key)
}

// String calls the String method in the corresponding address
func (key MACAddressKey) String() string {
	return key.ToAddress().String()
}

// KeyConstraint is the generic type constraint for an address type that can be generated from a generic address key.
type KeyConstraint[T any] interface {
	fmt.Stringer
	fromKey(addressScheme, *keyContents) T // implemented by IPAddress and Address
}

type addressScheme byte

const (
	adaptiveZeroScheme addressScheme = 0 // adaptiveZeroScheme needs to be zero, to coincide with the zero value for Address and IPAddress, which is a zero-length address
	ipv4Scheme         addressScheme = 1
	ipv6Scheme         addressScheme = 2
	mac48Scheme        addressScheme = 3
	eui64Scheme        addressScheme = 4
)

// KeyGeneratorConstraint is the generic type constraint for an address type that can generate a generic address key.
type KeyGeneratorConstraint[T KeyConstraint[T]] interface {
	ToGenericKey() Key[T]
}

// GenericKeyConstraint is the generic type constraint for an address type that can generate and be generated from a generic address key.
type GenericKeyConstraint[T KeyConstraint[T]] interface {
	KeyGeneratorConstraint[T]
	KeyConstraint[T]
}

// Key is a representation of an address that is comparable as defined by the language specification.
// See https://go.dev/ref/spec#Comparison_operators
//
// It can be used as a map key.  It can be obtained from its originating address instances.
// The zero value corresponds to the zero-value for its generic address type.
// Keys do not incorporate prefix length to ensure that all equal addresses have equal keys.
// To create a key that has prefix length, combine into a struct with the PrefixKey obtained by passing the address into PrefixKeyFrom.
type Key[T KeyConstraint[T]] struct {
	scheme addressScheme
	keyContents
}

// IPKeyFromRangeKey converts the IP address range key to a Key[*IPAddress].
// The type Key[*IPAddress] satisfies the "comparable" generic constraint, unlike RangeBoundaryKey.
// Both key types are "comparable" with respect to comparison operators and map keys.
func IPKeyFromRangeKey(key RangeBoundaryKey[*IPAddress]) Key[*IPAddress] {
	if conv, ok := key.(Key[*IPAddress]); ok {
		return conv
	}
	var addr *IPAddress = key.ToAddress()
	return addr.ToKey()
}

// ToAddress converts back to an address instance
func (key Key[T]) ToAddress() T {
	var t T
	return t.fromKey(key.scheme, &key.keyContents)
}

// String calls the String method in the corresponding address
func (key Key[T]) String() string {
	return key.ToAddress().String()
}

type keyContents struct {
	vals [2]struct {
		lower,
		upper uint64
	}
	zone Zone
}

type (
	AddressKey             = Key[*Address]
	IPAddressKey           = Key[*IPAddress]
	IPAddressSeqRangeKey   = SequentialRangeKey[*IPAddress]
	IPv4AddressSeqRangeKey = SequentialRangeKey[*IPv4Address]
	IPv6AddressSeqRangeKey = SequentialRangeKey[*IPv6Address]
)

var (
	_ Key[*IPv4Address]
	_ Key[*IPv6Address]
	_ Key[*MACAddress]

	_ AddressKey
	_ IPAddressKey
	_ IPv4AddressKey
	_ IPv6AddressKey
	_ MACAddressKey

	_ IPAddressSeqRangeKey
	_ IPv4AddressSeqRangeKey
	_ IPv6AddressSeqRangeKey
)

// PrefixKey is a representation of a prefix length that is comparable as defined by the language specification.
// See https://go.dev/ref/spec#Comparison_operators
//
// It can be used as a map key.
// The zero value is the absence of a prefix length.
type PrefixKey struct {
	// If true, the prefix length is indicated by PrefixLen.
	// If false, this indicates no prefix length for the associated address or subnet.
	IsPrefixed bool

	// If IsPrefixed is true, this holds the prefix length.
	// Otherwise, this should be zero if you wish that each address has a unique key.
	PrefixLen PrefixBitCount
}

// ToPrefixLen converts this key to its corresponding prefix length.
func (pref PrefixKey) ToPrefixLen() PrefixLen {
	if pref.IsPrefixed {
		return &pref.PrefixLen
	}
	return nil
}

func PrefixKeyFrom(addr AddressType) PrefixKey {
	if addr.IsPrefixed() {
		return PrefixKey{
			IsPrefixed: true,
			PrefixLen:  *addr.ToAddressBase().getPrefixLen(), // doing this instead of calling GetPrefixLen() on AddressType avoids the prefix len copy
		}
	}
	return PrefixKey{}
}
