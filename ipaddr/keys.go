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

func (pref *PrefixKey) ToPrefixLen() PrefixLen {
	if pref.IsPrefixed {
		return &pref.PrefixLen
	}
	return nil
}

// IPv4AddressKey is a representation of IPv4Address that is comparable as defined by the language specification.
// See https://go.dev/ref/spec#Comparison_operators
// It can be used as a map key.
// The zero value is the address 0.0.0.0
type IPv4AddressKey struct {
	Values [IPv4SegmentCount]struct {
		Value      IPv4SegInt
		UpperValue IPv4SegInt
	}
	Prefix PrefixKey
}

// Normalize normalizes the given key.
// Normalizing a key ensures it is the single unique key for any given address or subnet.
func (key *IPv4AddressKey) Normalize() {
	for i := range key.Values {
		seg := &key.Values[i]
		if seg.Value > seg.UpperValue {
			seg.Value, seg.UpperValue = seg.UpperValue, seg.Value
		}
	}
	key.Prefix.Normalize()
}

// Val provides the lower value for a given segment.
// For a given key, the val method provides a function of type IPv4SegmentProvider
func (key *IPv4AddressKey) Val(segmentIndex int) IPv4SegInt {
	return key.Values[segmentIndex].Value
}

// UpperVal provides the upper value for a given segment.
// For a given key, the val method provides a function of type IPv4SegmentProvider
func (key *IPv4AddressKey) UpperVal(segmentIndex int) IPv4SegInt {
	return key.Values[segmentIndex].UpperValue
}

// ToAddress converts to an address instance
func (key *IPv4AddressKey) ToAddress() *IPv4Address {
	return newIPv4AddressFromPrefixedSingle(key.Val, key.UpperVal, key.Prefix.ToPrefixLen())
}

func (key *IPv4AddressKey) ToBaseKey() *AddressKey {
	baseKey := &AddressKey{Prefix: key.Prefix, SegmentCount: IPv4SegmentCount}
	for i, val := range key.Values {
		baseVals := &baseKey.Values[i]
		baseVals.Value, baseVals.UpperValue = SegInt(val.Value), SegInt(val.UpperValue)
	}
	return baseKey
}

// IPv6AddressKey is a representation of IPv6Address that is comparable as defined by the language specification.
// See https://go.dev/ref/spec#Comparison_operators
// It can be used as a map key.
// The zero value is the address ::
type IPv6AddressKey struct {
	Values [IPv6SegmentCount]struct {
		Value      IPv6SegInt
		UpperValue IPv6SegInt
	}
	Prefix PrefixKey
	Zone   Zone
}

// Normalize normalizes the given key.
// Normalizing a key ensures it is the single unique key for any given address or subnet.
func (key *IPv6AddressKey) Normalize() {
	for i := range key.Values {
		seg := &key.Values[i]
		if seg.Value > seg.UpperValue {
			seg.Value, seg.UpperValue = seg.UpperValue, seg.Value
		}
	}
	key.Prefix.Normalize()
}

// Val provides the lower value for a given segment.
// For a given key, the val method provides a function of type IPv4SegmentProvider
func (key *IPv6AddressKey) Val(segmentIndex int) IPv6SegInt {
	return key.Values[segmentIndex].Value
}

// UpperVal provides the upper value for a given segment.
// For a given key, the val method provides a function of type IPv4SegmentProvider
func (key *IPv6AddressKey) UpperVal(segmentIndex int) IPv6SegInt {
	return key.Values[segmentIndex].UpperValue
}

// ToAddress converts to an address instance
func (key *IPv6AddressKey) ToAddress() *IPv6Address {
	return newIPv6AddressFromPrefixedSingle(key.Val, key.UpperVal, key.Prefix.ToPrefixLen(), key.Zone.String())
}

func (key *IPv6AddressKey) ToBaseKey() *AddressKey {
	baseKey := &AddressKey{Prefix: key.Prefix, SegmentCount: IPv6SegmentCount, Zone: key.Zone}
	for i, val := range key.Values {
		baseVals := &baseKey.Values[i]
		baseVals.Value, baseVals.UpperValue = SegInt(val.Value), SegInt(val.UpperValue)
	}
	return baseKey
}

// MACAddressKey is a representation of MACAddress that is comparable as defined by the language specification.
// See https://go.dev/ref/spec#Comparison_operators
// It can be used as a map key.
// The zero value is a zero-length MAC address.
type MACAddressKey struct {
	Values [ExtendedUniqueIdentifier64SegmentCount]struct {
		Value      MACSegInt
		UpperValue MACSegInt
	}
	Prefix       PrefixKey
	SegmentCount uint8
}

// Val provides the lower value for a given segment.
// For a given key, the val method provides a function of type MACSegmentProvider
func (key *MACAddressKey) Val(segmentIndex int) IPv4SegInt {
	return key.Values[segmentIndex].Value
}

// UpperVal provides the upper value for a given segment.
// For a given key, the val method provides a function of type MACSegmentProvider
func (key *MACAddressKey) UpperVal(segmentIndex int) IPv4SegInt {
	return key.Values[segmentIndex].UpperValue
}

// ToAddress converts to an address instance
func (key *MACAddressKey) ToAddress() *MACAddress {
	res := NewMACAddressFromRangeExt(key.Val, key.UpperVal, key.SegmentCount > MediaAccessControlSegmentCount)
	if key.Prefix.IsPrefixed {
		res.SetPrefixLen(key.Prefix.ToPrefixLen().Len())
	}
	return res
}

func (key *MACAddressKey) ToBaseKey() *AddressKey {
	baseKey := &AddressKey{Prefix: key.Prefix, SegmentCount: key.SegmentCount}
	for i, val := range key.Values {
		baseVals := &baseKey.Values[i]
		baseVals.Value, baseVals.UpperValue = SegInt(val.Value), SegInt(val.UpperValue)
	}
	return baseKey
}

const MaxSegmentCount = IPv6SegmentCount

// AddressKey is a representation of Address that is comparable as defined by the language specification.
// See https://go.dev/ref/spec#Comparison_operators
// It can be used as a map key.
// The zero value is a zero-length address.
type AddressKey struct {
	Values [MaxSegmentCount]struct {
		Value      SegInt
		UpperValue SegInt
	}
	SegmentCount uint8
	Prefix       PrefixKey
	Zone         Zone
}

// IPv4AddressSeqRangeKey is a representation of IPv4AddressSeqRange that is comparable as defined by the language specification.
// See https://go.dev/ref/spec#Comparison_operators
// It can be used as a map key.
// The zero value is a range from 0.0.0.0 to itself
type IPv4AddressSeqRangeKey struct {
	lower, upper IPv4AddressKey
}

// ToSeqRange converts to the associated sequential range
func (key *IPv4AddressSeqRangeKey) ToSeqRange() *IPv4AddressSeqRange {
	return NewIPv4SeqRange(key.lower.ToAddress(), key.upper.ToAddress())
}

// IPv6AddressSeqRangeKey is a representation of IPv6AddressSeqRange that is comparable as defined by the language specification.
// See https://go.dev/ref/spec#Comparison_operators
// It can be used as a map key.
// The zero value is a range from :: to itself
type IPv6AddressSeqRangeKey struct {
	lower, upper IPv6AddressKey
}

// ToSeqRange converts to the associated sequential range
func (key *IPv6AddressSeqRangeKey) ToSeqRange() *IPv6AddressSeqRange {
	return NewIPv6SeqRange(key.lower.ToAddress(), key.upper.ToAddress())
}

// IPAddressSeqRangeKey is a representation of IPAddressSeqRange that is comparable as defined by the language specification.
// See https://go.dev/ref/spec#Comparison_operators
// It can be used as a map key.
// The zero value is a range from a zero-length address to itself.
type IPAddressSeqRangeKey struct {
	lower, upper AddressKey
}