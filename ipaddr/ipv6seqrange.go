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

import (
	"fmt"
	"math/big"
	"net"
)

func NewIPv6SeqRange(one, two *IPv6Address) *IPv6AddressSeqRange {
	if one == nil && two == nil {
		one = zeroIPv6
	}
	one = one.WithoutZone()
	two = two.WithoutZone()
	return newSeqRange(one.ToIP(), two.ToIP()).ToIPv6()
}

var zeroIPv6Range = NewIPv6SeqRange(zeroIPv6, zeroIPv6)

// IPv6AddressSeqRange represents an arbitrary range of consecutive IPv6 addresses.
//
// The zero value is a range from :: to ::.
//
// See IPAddressSeqRange for more details.
type IPv6AddressSeqRange struct {
	ipAddressSeqRangeInternal
}

func (rng *IPv6AddressSeqRange) init() *IPv6AddressSeqRange {
	if rng.lower == nil {
		return zeroIPv6Range
	}
	return rng
}

// GetCount returns the count of addresses that this sequential range spans.
//
// Use IsMultiple if you simply want to know if the count is greater than 1.
func (rng *IPv6AddressSeqRange) GetCount() *big.Int {
	if rng == nil {
		return bigZero()
	}
	return rng.init().getCount()
}

// IsMultiple returns whether this range represents a range of multiple addresses
func (rng *IPv6AddressSeqRange) IsMultiple() bool {
	return rng != nil && rng.isMultiple()
}

// String implements the fmt.Stringer interface,
// returning the lower address canonical string, followed by the default separator " -> ",
// followed by the upper address canonical string.
// It returns "<nil>" if the receiver is a nil pointer.
func (rng *IPv6AddressSeqRange) String() string {
	if rng == nil {
		return nilString()
	}
	return rng.ToString((*IPv6Address).String, DefaultSeqRangeSeparator, (*IPv6Address).String)
}

// Format implements fmt.Formatter interface.
//
// It prints the string as "lower -> upper" where lower and upper are the formatted strings for the lowest and highest addresses in the range, given by GetLower and GetUpper.
// The formats, flags, and other specifications supported are those supported by Format in IPAddress.
func (rng IPv6AddressSeqRange) Format(state fmt.State, verb rune) {
	rng.init().format(state, verb)
}

func (rng *IPv6AddressSeqRange) ToString(lowerStringer func(*IPv6Address) string, separator string, upperStringer func(*IPv6Address) string) string {
	if rng == nil {
		return nilString()
	}
	return rng.init().toString(
		func(addr *IPAddress) string {
			return lowerStringer(addr.ToIPv6())
		},
		separator,
		func(addr *IPAddress) string {
			return upperStringer(addr.ToIPv6())
		},
	)
}

// ToNormalizedString produces a normalized string for the address range.
// It has the format "lower -> upper" where lower and upper are the normalized strings for the lowest and highest addresses in the range, given by GetLower and GetUpper.
func (rng *IPv6AddressSeqRange) ToNormalizedString() string {
	return rng.ToString((*IPv6Address).ToNormalizedString, DefaultSeqRangeSeparator, (*IPv6Address).ToNormalizedString)
}

// ToCanonicalString produces a canonical string for the address range.
// It has the format "lower -> upper" where lower and upper are the canonical strings for the lowest and highest addresses in the range, given by GetLower and GetUpper.
func (rng *IPv6AddressSeqRange) ToCanonicalString() string {
	return rng.ToString((*IPv6Address).ToCanonicalString, DefaultSeqRangeSeparator, (*IPv6Address).ToNormalizedString)
}

// GetBitCount returns the number of bits in each address in the range, which is 16
func (rng *IPv6AddressSeqRange) GetBitCount() BitCount {
	return rng.GetLower().GetBitCount()
}

// GetByteCount returns the number of bytes in each address in the range, which is 2
func (rng *IPv6AddressSeqRange) GetByteCount() int {
	return rng.GetLower().GetByteCount()
}

// GetLowerIPAddress satisfies the IPAddressRange interface, returning the lower address in the range, same as GetLower()
func (rng *IPv6AddressSeqRange) GetLowerIPAddress() *IPAddress {
	return rng.init().lower
}

// GetUpperIPAddress satisfies the IPAddressRange interface, returning the upper address in the range, same as GetUpper()
func (rng *IPv6AddressSeqRange) GetUpperIPAddress() *IPAddress {
	return rng.init().upper
}

// GetLower returns the lowest address of the sequential range, the one with the lowest numeric value
func (rng *IPv6AddressSeqRange) GetLower() *IPv6Address {
	return rng.init().lower.ToIPv6()
}

// GetUpper returns the highest address of the sequential range, the one with the highest numeric value
func (rng *IPv6AddressSeqRange) GetUpper() *IPv6Address {
	return rng.init().upper.ToIPv6()
}

// GetNetIP returns the lower IP address in the range as a net.IP
func (rng *IPv6AddressSeqRange) GetNetIP() net.IP {
	return rng.GetLower().GetNetIP()
}

// CopyNetIP copies the value of the lower IP address in the range into a net.IP.
//
// If the value can fit in the given net.IP slice, the value is copied into that slice and a length-adjusted sub-slice is returned.
// Otherwise, a new slice is created and returned with the value.
func (rng *IPv6AddressSeqRange) CopyNetIP(bytes net.IP) net.IP {
	return rng.GetLower().CopyNetIP(bytes)
}

// GetUpperNetIP returns the upper IP address in the range as a net.IP
func (rng *IPv6AddressSeqRange) GetUpperNetIP() net.IP {
	return rng.GetUpper().GetUpperNetIP()
}

// CopyUpperNetIP copies the upper IP address in the range into a net.IP.
//
// If the value can fit in the given net.IP slice, the value is copied into that slice and a length-adjusted sub-slice is returned.
// Otherwise, a new slice is created and returned with the value.
func (rng *IPv6AddressSeqRange) CopyUpperNetIP(bytes net.IP) net.IP {
	return rng.GetUpper().CopyUpperNetIP(bytes)
}

// Bytes returns the lowest address in the range, the one with the lowest numeric value, as a byte slice
func (rng *IPv6AddressSeqRange) Bytes() []byte {
	return rng.GetLower().Bytes()
}

// CopyBytes copies the value of the lowest address in the range into a byte slice.
//
// If the value can fit in the given slice, the value is copied into that slice and a length-adjusted sub-slice is returned.
// Otherwise, a new slice is created and returned with the value.
func (rng *IPv6AddressSeqRange) CopyBytes(bytes []byte) []byte {
	return rng.GetLower().CopyBytes(bytes)
}

// UpperBytes returns the highest address in the range, the one with the highest numeric value, as a byte slice
func (rng *IPv6AddressSeqRange) UpperBytes() []byte {
	return rng.GetUpper().UpperBytes()
}

// CopyUpperBytes copies the value of the highest address in the range into a byte slice.
//
// If the value can fit in the given slice, the value is copied into that slice and a length-adjusted sub-slice is returned.
// Otherwise, a new slice is created and returned with the value.
func (rng *IPv6AddressSeqRange) CopyUpperBytes(bytes []byte) []byte {
	return rng.GetUpper().CopyUpperBytes(bytes)
}

// GetValue returns the lowest address in the range, the one with the lowest numeric value, as an integer
func (rng *IPv6AddressSeqRange) GetValue() *big.Int {
	return rng.GetLower().GetValue()
}

// GetUpperValue returns the highest address in the range, the one with the highest numeric value, as an integer
func (rng *IPv6AddressSeqRange) GetUpperValue() *big.Int {
	return rng.GetUpper().GetValue()
}

// Contains returns whether this range contains all addresses in the given address or subnet.
func (rng *IPv6AddressSeqRange) Contains(other IPAddressType) bool {
	if rng == nil {
		return other == nil || other.ToAddressBase() == nil
	}
	return rng.init().contains(other)
}

// ContainsRange returns whether all the addresses in the given sequential range are also contained in this sequential range.
func (rng *IPv6AddressSeqRange) ContainsRange(other IPAddressSeqRangeType) bool {
	if rng == nil {
		return other == nil || other.ToIP() == nil
	}
	return rng.init().containsRange(other)
}

// Equal returns whether the given sequential address range is equal to this sequential address range.
// Two sequential address ranges are equal if their lower and upper range boundaries are equal.
func (rng *IPv6AddressSeqRange) Equal(other IPAddressSeqRangeType) bool {
	if rng == nil {
		return other == nil || other.ToIP() == nil
	}
	return rng.init().equals(other)
}

// Compare returns a negative integer, zero, or a positive integer if this sequential address range is less than, equal, or greater than the given item.
// Any address item is comparable to any other.  All address items use CountComparator to compare.
func (rng *IPv6AddressSeqRange) Compare(item AddressItem) int {
	if rng != nil {
		rng = rng.init()
	}
	return CountComparator.Compare(rng, item)
}

// CompareSize compares the counts of two address ranges, the number of individual addresses within.
//
// Rather than calculating counts with GetCount, there can be more efficient ways of comparing whether one range spans more individual addresses than another.
//
// CompareSize returns a positive integer if this address range has a larger count than the one given, 0 if they are the same, or a negative integer if the other has a larger count.
func (rng *IPv6AddressSeqRange) CompareSize(other IPAddressSeqRangeType) int {
	if rng == nil {
		if other != nil && other.ToIP() != nil {
			// we have size 0, other has size >= 1
			return -1
		}
		return 0
	}
	return rng.compareSize(other)
}

// ContainsPrefixBlock returns whether the range contains the block of addresses for the given prefix length.
//
// Unlike ContainsSinglePrefixBlock, whether there are multiple prefix values for the given prefix length makes no difference.
//
// Use GetMinPrefixLenForBlock to determine whether there is a prefix length for which this method returns true.
func (rng *IPv6AddressSeqRange) ContainsPrefixBlock(prefixLen BitCount) bool {
	return rng.init().ipAddressSeqRangeInternal.ContainsPrefixBlock(prefixLen)
}

// ContainsSinglePrefixBlock returns whether this address range contains a single prefix block for the given prefix length.
//
// This means there is only one prefix value for the given prefix length, and it also contains the full prefix block for that prefix, all addresses with that prefix.
//
// Use GetPrefixLenForSingleBlock to determine whether there is a prefix length for which this method returns true.
func (rng *IPv6AddressSeqRange) ContainsSinglePrefixBlock(prefixLen BitCount) bool {
	return rng.init().ipAddressSeqRangeInternal.ContainsSinglePrefixBlock(prefixLen)
}

// GetPrefixLenForSingleBlock returns a prefix length for which there is only one prefix in this range,
// and the range of values in this range matches the block of all values for that prefix.
//
// If the range can be described this way, then this method returns the same value as GetMinPrefixLenForBlock.
//
// If no such prefix length exists, returns nil.
//
// If this item represents a single value, this returns the bit count.
func (rng *IPv6AddressSeqRange) GetPrefixLenForSingleBlock() PrefixLen {
	return rng.init().ipAddressSeqRangeInternal.GetPrefixLenForSingleBlock()
}

// GetMinPrefixLenForBlock returns the smallest prefix length such that this includes the block of addresses for that prefix length.
//
// If the entire range can be described this way, then this method returns the same value as GetPrefixLenForSingleBlock.
//
// There may be a single prefix, or multiple possible prefix values in this item for the returned prefix length.
// Use GetPrefixLenForSingleBlock to avoid the case of multiple prefix values.
func (rng *IPv6AddressSeqRange) GetMinPrefixLenForBlock() BitCount {
	return rng.init().ipAddressSeqRangeInternal.GetMinPrefixLenForBlock()
}

// IsFullRange returns whether this address range covers the entire IPv6 address space.
//
// This is true if and only if both IncludesZero and IncludesMax return true.
func (rng *IPv6AddressSeqRange) IsFullRange() bool {
	return rng.init().ipAddressSeqRangeInternal.IsFullRange()
}

// IsMax returns whether this sequential range spans from the max address, the address whose bits are all ones, to itself.
func (rng *IPv6AddressSeqRange) IsMax() bool {
	return rng.init().ipAddressSeqRangeInternal.IsMax()
}

// IncludesMax returns whether this sequential range's upper value is the max value, the value whose bits are all ones.
func (rng *IPv6AddressSeqRange) IncludesMax() bool {
	return rng.init().ipAddressSeqRangeInternal.IncludesMax()
}

// Iterator provides an iterator to iterate through the individual addresses of this address range.
//
// Call GetCount for the count.
func (rng *IPv6AddressSeqRange) Iterator() IPv6AddressIterator {
	if rng == nil {
		return ipv6AddressIterator{nilAddrIterator()}
	}
	return ipv6AddressIterator{rng.init().iterator()}
}

// PrefixBlockIterator provides an iterator to iterate through the individual prefix blocks of the given prefix length,
// one for each prefix of that length in the address range.
func (rng *IPv6AddressSeqRange) PrefixBlockIterator(prefLength BitCount) IPv6AddressIterator {
	return &ipv6AddressIterator{rng.init().prefixBlockIterator(prefLength)}
}

// PrefixIterator provides an iterator to iterate through the individual prefixes of the given prefix length in this address range,
// each iterated element spanning the range of values for its prefix.
//
// It is similar to the prefix block iterator, except for possibly the first and last iterated elements, which might not be prefix blocks,
// instead constraining themselves to values from this range.
func (rng *IPv6AddressSeqRange) PrefixIterator(prefLength BitCount) IPv6AddressSeqRangeIterator {
	return &ipv6RangeIterator{rng.init().prefixIterator(prefLength)}
}

// ToIP converts to an IPAddressSeqRange, a polymorphic type usable with all IP address sequential ranges.
//
// ToIP can be called with a nil receiver, enabling you to chain this method with methods that might return a nil pointer.
func (rng *IPv6AddressSeqRange) ToIP() *IPAddressSeqRange {
	if rng != nil {
		rng = rng.init()
	}
	return (*IPAddressSeqRange)(rng)
}

// Overlaps returns true if this sequential range overlaps with the given sequential range.
func (rng *IPv6AddressSeqRange) Overlaps(other *IPv6AddressSeqRange) bool {
	return rng.init().overlaps(other.ToIP())
}

// Intersect returns the intersection of this range with the given range, a range which includes those addresses found in both.
func (rng *IPv6AddressSeqRange) Intersect(other *IPv6AddressSeqRange) *IPAddressSeqRange {
	return rng.init().intersect(other.toIPSequentialRange())
}

// CoverWithPrefixBlock returns the minimal-size prefix block that covers all the addresses in this range.
// The resulting block will have a larger count than this, unless this range already directly corresponds to a prefix block.
func (rng *IPv6AddressSeqRange) CoverWithPrefixBlock() *IPv6Address {
	return rng.GetLower().CoverWithPrefixBlockTo(rng.GetUpper())
}

// SpanWithPrefixBlocks returns an array of prefix blocks that spans the same set of addresses as this range.
func (rng *IPv6AddressSeqRange) SpanWithPrefixBlocks() []*IPv6Address {
	return rng.GetLower().SpanWithPrefixBlocksTo(rng.GetUpper())
}

// SpanWithSequentialBlocks produces the smallest slice of sequential blocks that cover the same set of addresses as this range.
// This slice can be shorter than that produced by SpanWithPrefixBlocks and is never longer.
func (rng *IPv6AddressSeqRange) SpanWithSequentialBlocks() []*IPv6Address {
	return rng.GetLower().SpanWithSequentialBlocksTo(rng.GetUpper())
}

// Join joins the given ranges into the fewest number of ranges.
// The returned array will be sorted by ascending lowest range value.
func (rng *IPv6AddressSeqRange) Join(ranges ...*IPv6AddressSeqRange) []*IPv6AddressSeqRange {
	origLen := len(ranges)
	ranges2 := make([]*IPAddressSeqRange, 0, origLen+1)
	for _, rng := range ranges {
		ranges2 = append(ranges2, rng.ToIP())
	}
	ranges2 = append(ranges2, rng.ToIP())
	return cloneToIPv6SeqRange(join(ranges2))
}

// JoinTo joins this range to the other if they are contiguous.  If this range overlaps with the given range,
// or if the highest value of the lower range is one below the lowest value of the higher range,
// then the two are joined into a new larger range that is returned.
// Otherwise nil is returned.
func (rng *IPv6AddressSeqRange) JoinTo(other *IPv6AddressSeqRange) *IPv6AddressSeqRange {
	return rng.init().joinTo(other.init().ToIP()).ToIPv6()
}

// Extend extends this sequential range to include all address in the given range.
// If the argument has a different IP version than this, nil is returned.
// Otherwise, this method returns the range that includes this range, the given range, and all addresses in-between.
func (rng *IPv6AddressSeqRange) Extend(other *IPv6AddressSeqRange) *IPv6AddressSeqRange {
	return rng.init().extend(other.init().ToIP()).ToIPv6()
}

// Subtract Subtracts the given range from this range, to produce either zero, one, or two address ranges that contain the addresses in this range and not in the given range.
// If the result has length 2, the two ranges are ordered by ascending lowest range value.
func (rng *IPv6AddressSeqRange) Subtract(other *IPv6AddressSeqRange) []*IPv6AddressSeqRange {
	return cloneToIPv6SeqRange(rng.init().subtract(other.init().ToIP()))
}

// ToKey creates the associated address range key.
// While address ranges can be compared with the Compare or Equal methods as well as various provided instances of AddressComparator,
// they are not comparable with go operators.
// However, IPv6AddressSeqRangeKey instances are comparable with go operators, and thus can be used as map keys.
func (rng *IPv6AddressSeqRange) ToKey() *IPv6AddressSeqRangeKey {
	return &IPv6AddressSeqRangeKey{
		lower: *rng.GetLower().ToKey(),
		upper: *rng.GetUpper().ToKey(),
	}
}
