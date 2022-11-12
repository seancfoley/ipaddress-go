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

// KeyConstraint is the generic type constraint for an address key
type KeyConstraint[T any] interface {
	AddressType
	fromKey(*keyContents) T
}

type keyContents struct {
	scheme addressScheme
	vals   [2]struct {
		lower,
		upper uint64
	}
	zone Zone
}

// Key is a representation of an address that is comparable as defined by the language specification.
// See https://go.dev/ref/spec#Comparison_operators
// It can be used as a map key.  They can be obtained from their originating address instances.
// The zero value is a zero-length address.
type Key[T KeyConstraint[T]] struct {
	keyContents
}

// ToAddress converts back to an address instance
func (key Key[T]) ToAddress() T {
	var t T
	return t.fromKey(&key.keyContents)
}

// String calls the String method in the corresponding address
func (key Key[T]) String() string {
	return key.ToAddress().String()
}

type addressScheme byte

const (
	anyScheme   addressScheme = 0
	ipv4Scheme                = 1
	ipv6Scheme                = 2
	mac48Scheme addressScheme = 3
	eui64Scheme addressScheme = 4
)

// RangeKey is a representation of SequentialRange that is comparable as defined by the language specification.
// See https://go.dev/ref/spec#Comparison_operators
// It can be used as a map key.
// The zero value is a range from a zero-length address to itself.
type RangeKey[T SequentialRangeConstraint[T]] struct {
	lowerKey, upperKey Key[T]
}

// ToSeqRange converts back to a sequential range instance
func (key RangeKey[T]) ToSeqRange() *SequentialRange[T] {
	lower := key.lowerKey.ToAddress()
	upper := key.upperKey.ToAddress()
	return newSequRangeUnchecked(lower, upper, lower != upper)
}

// String calls the String method in the corresponding sequential range
func (key RangeKey[T]) String() string {
	return key.ToSeqRange().String()
}

// PrefixKey is a representation of a prefix length that is comparable as defined by the language specification.
// See https://go.dev/ref/spec#Comparison_operators
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

// Normalize normalizes the prefix length.
// Normalizing a prefix length ensures it is unique for the represented prefix length.
func (pref *PrefixKey) Normalize() {
	if pref.IsPrefixed {
		if pref.PrefixLen > IPv4BitCount {
			pref.PrefixLen = IPv4BitCount
		}
	} else {
		pref.PrefixLen = 0
	}
}

// ToPrefixLen converts this key to its corresponding prefix length.
func (pref *PrefixKey) ToPrefixLen() PrefixLen {
	if pref.IsPrefixed {
		return &pref.PrefixLen
	}
	return nil
}
