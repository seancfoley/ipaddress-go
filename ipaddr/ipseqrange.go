//
// Copyright 2020-2026 Sean C Foley
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
	"math/bits"
	"net"
	"net/netip"
	"sort"
	"strings"
	"unsafe"
)

// DefaultSeqRangeSeparator is the low to high value separator used when creating strings for IP ranges.
const DefaultSeqRangeSeparator = " -> "

type rangeCache struct {
	cachedCount *big.Int
	isMultiple  bool
}

type SequentialRangeConstraint[T any] ipAddressTypeConstraint[T]

var (
	_ SequentialRange[*IPAddress]
	_ SequentialRange[*IPv4Address]
	_ SequentialRange[*IPv6Address]
)

// SequentialRange represents an arbitrary range of consecutive IP addresses, from a lower address to an upper address, inclusive.
//
// For the generic type T you can choose *IPAddress, *IPv4Address, or *IPv6Address.
//
// This type allows the representation of any sequential address range, including those that cannot be represented by [IPAddress] or [IPAddressString].
//
// [IPAddress] and [IPAddressString] allow you to specify a range of values for each segment, allowing
// for single addresses, any address CIDR prefix subnet (for example, "1.2.0.0/16" or "1:2:3:4::/64") or any subnet that can be represented with segment ranges (for example, "1.2.0-255.*" or "1:2:3:4:*").
// See [IPAddressString] for details.
// [IPAddressString] and [IPAddress] cover all potential subnets and addresses that can be represented by a single address string of 4 or less segments for IPv4, and 8 or less segments for IPv6.
// In contrast, this type covers any sequential address range.
//
// String representations of this type include the full address for both the lower and upper bounds of the range.
//
// The zero value is a range from the zero-valued [IPAddress] to itself.
//
// For a range of type SequentialRange[*IPAddress], the range spans from an IPv4 address to another IPv4 address,
// or from an IPv6 address to another IPv6 address.  A sequential range cannot include both IPv4 and IPv6 addresses.
type SequentialRange[T SequentialRangeConstraint[T]] struct {
	lower,
	upper T
	cache *rangeCache
}

func nilConvert[T SequentialRangeConstraint[T]](t T) (result T, isNil bool) {
	anyt := any(t)
	if val, ok := anyt.(*IPv6Address); ok && val == nil {
		isNil = true
		result = any(zeroIPv6).(T)
	} else if val, ok := anyt.(*IPv4Address); ok && val == nil {
		isNil = true
		result = any(zeroIPv4).(T)
	} else if val, ok := anyt.(*IPAddress); ok && val == nil {
		isNil = true
		result = any(zeroIPAddr).(T)
	}
	return
}

func (rng *SequentialRange[T]) init() *SequentialRange[T] {
	if newVal, isNil := nilConvert(rng.lower); isNil {
		zeroSeqRange := newSequRange(newVal, newVal)
		return zeroSeqRange
	}
	return rng
}

// GetIPVersion returns the IP version of this IP address sequential range
func (rng *SequentialRange[T]) GetIPVersion() IPVersion {
	return rng.init().lower.GetIPVersion()
}

func (rng *SequentialRange[T]) getCachedCount(copy bool) (res *big.Int) {
	cache := rng.cache
	count := (*big.Int)(atomicLoadPointer((*unsafe.Pointer)(unsafe.Pointer(&cache.cachedCount))))
	if count == nil {
		if !rng.IsMultiple() {
			count = bigOne()
		} else {
			lower := rng.lower
			upper := rng.upper
			if ipv4Lower, ok := any(lower).(*IPv4Address); ok {
				ipv4Upper := any(upper).(*IPv4Address)
				val := int64(ipv4Upper.Uint32Value()) - int64(ipv4Lower.Uint32Value()) + 1
				count = bigZero().SetInt64(val)
			} else {
				count = upper.GetValue()
				res = lower.GetValue()
				count.Sub(count, res).Add(count, bigOneConst())
				res.Set(count)
			}
		}
		dataLoc := (*unsafe.Pointer)(unsafe.Pointer(&cache.cachedCount))
		atomicStorePointer(dataLoc, unsafe.Pointer(count))
	}
	if res == nil {
		if copy {
			res = bigZero().Set(count)
		} else {
			res = count
		}
	}
	return
}

// GetPrefixCountLen returns the count of the number of distinct values within the prefix part of the range of addresses.
func (rng *SequentialRange[T]) GetPrefixCountLen(prefixLen BitCount) *big.Int {
	if !rng.IsMultiple() { // also checks for zero-ranges
		return bigOne()
	}
	bitCount := rng.lower.GetBitCount()
	if prefixLen <= 0 {
		return bigOne()
	} else if prefixLen >= bitCount {
		return rng.GetCount()
	}
	shiftAdjustment := bitCount - prefixLen
	lower := rng.lower
	if ipv4Lower, ok := any(lower).(*IPv4Address); ok {
		ipv4Upper := any(rng.upper).(*IPv4Address)
		upperAdjusted := ipv4Upper.Uint32Value() >> uint(shiftAdjustment)
		lowerAdjusted := ipv4Lower.Uint32Value() >> uint(shiftAdjustment)
		result := int64(upperAdjusted) - int64(lowerAdjusted) + 1
		return bigZero().SetInt64(result)
	}
	upperVal := rng.upper.GetValue()
	ushiftAdjustment := uint(shiftAdjustment)
	upperVal.Rsh(upperVal, ushiftAdjustment)
	lowerVal := lower.GetValue()
	lowerVal.Rsh(lowerVal, ushiftAdjustment)
	upperVal.Sub(upperVal, lowerVal).Add(upperVal, bigOneConst())
	return upperVal
}

// IsSequential returns whether the address or subnet represents a range of values that are sequential.
//
// IP address sequential ranges are sequential by definition, so this returns true.
func (rng *SequentialRange[T]) IsSequential() bool {
	return true
}

// ContainsPrefixBlock returns whether the range contains the block of addresses for the given prefix length.
//
// Unlike ContainsSinglePrefixBlock, whether there are multiple prefix values for the given prefix length makes no difference.
//
// Use GetMinPrefixLenForBlock to determine whether there is a prefix length for which this method returns true.
func (rng *SequentialRange[T]) ContainsPrefixBlock(prefixLen BitCount) bool {
	rng = rng.init()
	lower := rng.lower
	upper := rng.upper
	prefixLen = checkSubnet(lower, prefixLen)
	divCount := lower.GetDivisionCount()
	bitsPerSegment := lower.GetBitsPerSegment()
	i := getHostSegmentIndex(prefixLen, lower.GetBytesPerSegment(), bitsPerSegment)
	if i < divCount {
		div := lower.GetGenericSegment(i)
		upperDiv := upper.GetGenericSegment(i)
		segmentPrefixLength := getPrefixedSegmentPrefixLength(bitsPerSegment, prefixLen, i)
		if !isPrefixBlockVals(DivInt(div.GetSegmentValue()), DivInt(upperDiv.GetSegmentValue()), segmentPrefixLength.bitCount(), div.GetBitCount()) {
			return false
		}
		for i++; i < divCount; i++ {
			div = lower.GetGenericSegment(i)
			upperDiv = upper.GetGenericSegment(i)
			//is full range?
			if !div.IncludesZero() || !upperDiv.IncludesMax() {
				return false
			}
		}
	}
	return true
}

// ContainsSinglePrefixBlock returns whether this address range contains a single prefix block for the given prefix length.
//
// This means there is only one prefix value for the given prefix length, and it also contains the full prefix block for that prefix, all addresses with that prefix.
//
// Use GetPrefixLenForSingleBlock to determine whether there is a prefix length for which this method returns true.
func (rng *SequentialRange[T]) ContainsSinglePrefixBlock(prefixLen BitCount) bool {
	rng = rng.init()
	lower := rng.lower
	upper := rng.upper
	prefixLen = checkSubnet(lower, prefixLen)
	var prevBitCount BitCount
	divCount := lower.GetDivisionCount()
	for i := 0; i < divCount; i++ {
		div := lower.GetGenericSegment(i)
		upperDiv := upper.GetGenericSegment(i)
		bitCount := div.GetBitCount()
		totalBitCount := bitCount + prevBitCount
		if prefixLen >= totalBitCount {
			if !segValSame(div.GetSegmentValue(), upperDiv.GetSegmentValue()) {
				return false
			}
		} else {
			divPrefixLen := prefixLen - prevBitCount
			if !isPrefixBlockVals(DivInt(div.GetSegmentValue()), DivInt(upperDiv.GetSegmentValue()), divPrefixLen, div.GetBitCount()) {
				return false
			}
			for i++; i < divCount; i++ {
				div = lower.GetGenericSegment(i)
				upperDiv = upper.GetGenericSegment(i)
				if !div.IncludesZero() || !upperDiv.IncludesMax() {
					return false
				}
			}
			return true
		}
		prevBitCount = totalBitCount
	}
	return true
}

// GetPrefixLenForSingleBlock returns a prefix length for which there is only one prefix in this range,
// and the range of values in this range matches the block of all values for that prefix.
//
// If the range can be described this way, then this method returns the same value as GetMinPrefixLenForBlock.
//
// If no such prefix length exists, returns nil.
//
// If this item represents a single value, this returns the bit count.
func (rng *SequentialRange[T]) GetPrefixLenForSingleBlock() PrefixLen {
	rng = rng.init()
	lower := rng.lower
	upper := rng.upper
	count := lower.GetSegmentCount()
	segBitCount := lower.GetBitsPerSegment()
	maxSegValue := ^(^SegInt(0) << uint(segBitCount))
	totalPrefix := BitCount(0)
	for i := 0; i < count; i++ {
		lowerSeg := lower.GetGenericSegment(i)
		upperSeg := upper.GetGenericSegment(i)
		segPrefix := getPrefixLenForSingleBlock(DivInt(lowerSeg.GetSegmentValue()), DivInt(upperSeg.GetSegmentValue()), segBitCount)
		if segPrefix == nil {
			return nil
		}
		dabits := segPrefix.bitCount()
		totalPrefix += dabits
		if dabits < segBitCount {
			//remaining segments must be full range or we return nil
			for i++; i < count; i++ {
				lowerSeg = lower.GetGenericSegment(i)
				upperSeg = upper.GetGenericSegment(i)
				if lowerSeg.GetSegmentValue() != 0 {
					return nil
				} else if upperSeg.GetSegmentValue() != maxSegValue {
					return nil
				}
			}
		}
	}
	return cacheBitCount(totalPrefix)

}

// GetMinPrefixLenForBlock returns the smallest prefix length such that this includes the block of addresses for that prefix length.
//
// If the entire range can be described this way, then this method returns the same value as GetPrefixLenForSingleBlock.
//
// There may be a single prefix, or multiple possible prefix values in this item for the returned prefix length.
// Use GetPrefixLenForSingleBlock to avoid the case of multiple prefix values.
func (rng *SequentialRange[T]) GetMinPrefixLenForBlock() BitCount {
	rng = rng.init()
	lower := rng.lower
	upper := rng.upper
	count := lower.GetSegmentCount()
	totalPrefix := lower.GetBitCount()
	segBitCount := lower.GetBitsPerSegment()
	for i := count - 1; i >= 0; i-- {
		lowerSeg := lower.GetGenericSegment(i)
		upperSeg := upper.GetGenericSegment(i)
		segPrefix := getMinPrefixLenForBlock(DivInt(lowerSeg.GetSegmentValue()), DivInt(upperSeg.GetSegmentValue()), segBitCount)
		if segPrefix == segBitCount {
			break
		} else {
			totalPrefix -= segBitCount
			if segPrefix != 0 {
				totalPrefix += segPrefix
				break
			}
		}
	}
	return totalPrefix
}

// IsZero returns whether this sequential range spans from the zero address to itself.
func (rng *SequentialRange[T]) IsZero() bool {
	return rng.IncludesZero() && !rng.IsMultiple()
}

// IncludesZero returns whether this sequential range's lower value is the zero address.
func (rng *SequentialRange[T]) IncludesZero() bool {
	return rng.init().lower.IsZero()
}

// IsMax returns whether this sequential range spans from the max address, the address whose bits are all ones, to itself.
func (rng *SequentialRange[T]) IsMax() bool {
	return rng.IncludesMax() && !rng.IsMultiple()
}

// IncludesMax returns whether this sequential range's upper value is the max value, the value whose bits are all ones.
func (rng *SequentialRange[T]) IncludesMax() bool {
	return rng.init().upper.IsMax()
}

// IsFullRange returns whether this address range covers the entire address space of this IP address version.
//
// This is true if and only if both IncludesZero and IncludesMax return true.
func (rng *SequentialRange[T]) IsFullRange() bool {
	return rng.IncludesZero() && rng.IncludesMax()
}

// GetCount returns the count of addresses that this sequential range spans.
//
// Use IsMultiple if you simply want to know if the count is greater than 1.
func (rng *SequentialRange[T]) GetCount() *big.Int {
	if rng == nil {
		return bigZero()
	}
	return rng.init().getCachedCount(true)
}

// IsMultiple returns whether this range represents a range of multiple addresses.
func (rng *SequentialRange[T]) IsMultiple() bool {
	return rng != nil && rng.cache != nil && rng.cache.isMultiple
}

// String implements the [fmt.Stringer] interface,
// returning the lower address canonical string, followed by the default separator " -> ",
// followed by the upper address canonical string.
// It returns "<nil>" if the receiver is a nil pointer.
func (rng *SequentialRange[T]) String() string {
	if rng == nil {
		return nilString()
	}
	return rng.ToString(T.String, DefaultSeqRangeSeparator, T.String)
}

// Format implements [fmt.Formatter] interface.
//
// It prints the string as "lower -> upper" where lower and upper are the formatted strings for the lowest and highest addresses in the range, given by GetLower and GetUpper.
// The formats, flags, and other specifications supported are those supported by Format in IPAddress.
func (rng SequentialRange[T]) Format(state fmt.State, verb rune) {
	rngPtr := rng.init()
	rngPtr.lower.Format(state, verb)
	_, _ = state.Write([]byte(DefaultSeqRangeSeparator))
	rngPtr.upper.Format(state, verb)
}

// ToString produces a customized string for the address range.
func (rng *SequentialRange[T]) ToString(lowerStringer func(T) string, separator string, upperStringer func(T) string) string {
	if rng == nil {
		return nilString()
	}
	rng = rng.init()
	builder := strings.Builder{}
	str1, str2, str3 := lowerStringer(rng.lower), separator, upperStringer(rng.upper)
	builder.Grow(len(str1) + len(str2) + len(str3))
	builder.WriteString(str1)
	builder.WriteString(str2)
	builder.WriteString(str3)
	return builder.String()
}

// ToNormalizedString produces a normalized string for the address range.
// It has the format "lower -> upper" where lower and upper are the normalized strings for the lowest and highest addresses in the range, given by GetLower and GetUpper.
func (rng *SequentialRange[T]) ToNormalizedString() string {
	return rng.ToString(T.ToNormalizedString, DefaultSeqRangeSeparator, T.ToNormalizedString)
}

// ToCanonicalString produces a canonical string for the address range.
// It has the format "lower -> upper" where lower and upper are the canonical strings for the lowest and highest addresses in the range, given by GetLower and GetUpper.
func (rng *SequentialRange[T]) ToCanonicalString() string {
	return rng.ToString(T.ToCanonicalString, DefaultSeqRangeSeparator, T.ToCanonicalString)
}

// GetLowerIPAddress satisfies the IPAddressRange interface, returning the lower address in the range, same as GetLower.
func (rng *SequentialRange[T]) GetLowerIPAddress() *IPAddress {
	return rng.GetLower().ToIP()
}

// GetUpperIPAddress satisfies the IPAddressRange interface, returning the upper address in the range, same as GetUpper.
func (rng *SequentialRange[T]) GetUpperIPAddress() *IPAddress {
	return rng.GetUpper().ToIP()
}

// GetLower returns the lowest address in the range, the one with the lowest numeric value.
func (rng *SequentialRange[T]) GetLower() T {
	return rng.init().lower
}

// GetUpper returns the highest address in the range, the one with the highest numeric value.
func (rng *SequentialRange[T]) GetUpper() T {
	return rng.init().upper
}

// GetLowerAndUpper returns the lowest and highest addresses in the range, the ones with the lowest and highest numeric values.
func (rng *SequentialRange[T]) GetLowerAndUpper() (lower, upper T) {
	rng = rng.init()
	return rng.lower, rng.upper
}

func (rng *SequentialRange[T]) GetBig(index *big.Int) T {
	if index.Sign() < 0 || index.Cmp(rng.getCachedCount(false)) >= 0 {
		outOfBounds()
	}
	return rng.GetLower().IncrementBig(index)
}

func (rng *SequentialRange[T]) Get(index int64) T {
	if index < 0 || uint64(index) >= rng.getIPv4Count() {
		outOfBounds()
	}
	return rng.GetLower().Increment(index)
}

// getIPv4Count is equivalent to GetCount but returns a uint64
func (rng *SequentialRange[T]) getIPv4Count() uint64 {
	lower, upper := rng.GetLowerAndUpper()
	return uint64(upper.ToIP().ToIPv4().Uint32Value()-lower.ToIP().ToIPv4().Uint32Value()) + 1
}

// GetBitCount returns the number of bits in each address in the range.
func (rng *SequentialRange[T]) GetBitCount() BitCount {
	return rng.GetLower().GetBitCount()
}

// GetByteCount returns the number of bytes in each address in the range.
func (rng *SequentialRange[T]) GetByteCount() int {
	return rng.GetLower().GetByteCount()
}

// GetNetIP returns the lower IP address in the range as a net.IP.
func (rng *SequentialRange[T]) GetNetIP() net.IP {
	return rng.GetLower().GetNetIP()
}

// GetUpperNetIP returns the upper IP address in the range as a net.IP.
func (rng *SequentialRange[T]) GetUpperNetIP() net.IP {
	return rng.GetUpper().GetUpperNetIP()
}

// GetNetNetIPAddr returns the lowest address in this address range as a netip.Addr.
func (rng *SequentialRange[T]) GetNetNetIPAddr() netip.Addr {
	return rng.GetLower().GetNetNetIPAddr()
}

// GetUpperNetNetIPAddr returns the highest address in this address range as a netip.Addr.
func (rng *SequentialRange[T]) GetUpperNetNetIPAddr() netip.Addr {
	return rng.GetUpper().GetUpperNetNetIPAddr()
}

// CopyNetIP copies the value of the lower IP address in the range into a net.IP.
//
// If the value can fit in the given net.IP slice, the value is copied into that slice and a length-adjusted sub-slice is returned.
// Otherwise, a new slice is created and returned with the value.
func (rng *SequentialRange[T]) CopyNetIP(bytes net.IP) net.IP {
	return rng.GetLower().CopyNetIP(bytes) // this changes the arg to 4 bytes if 16 bytes and ipv4
}

// CopyUpperNetIP copies the upper IP address in the range into a net.IP.
//
// If the value can fit in the given net.IP slice, the value is copied into that slice and a length-adjusted sub-slice is returned.
// Otherwise, a new slice is created and returned with the value.
func (rng *SequentialRange[T]) CopyUpperNetIP(bytes net.IP) net.IP {
	return rng.GetUpper().CopyUpperNetIP(bytes) // this changes the arg to 4 bytes if 16 bytes and ipv4
}

// Bytes returns the lowest address in the range, the one with the lowest numeric value, as a byte slice.
func (rng *SequentialRange[T]) Bytes() []byte {
	return rng.GetLower().Bytes()
}

// CopyBytes copies the value of the lowest address in the range into a byte slice.
//
// If the value can fit in the given slice, the value is copied into that slice and a length-adjusted sub-slice is returned.
// Otherwise, a new slice is created and returned with the value.
func (rng *SequentialRange[T]) CopyBytes(bytes []byte) []byte {
	return rng.GetLower().CopyBytes(bytes)
}

// UpperBytes returns the highest address in the range, the one with the highest numeric value, as a byte slice.
func (rng *SequentialRange[T]) UpperBytes() []byte {
	return rng.GetUpper().UpperBytes()
}

// CopyUpperBytes copies the value of the highest address in the range into a byte slice.
//
// If the value can fit in the given slice, the value is copied into that slice and a length-adjusted sub-slice is returned.
// Otherwise, a new slice is created and returned with the value.
func (rng *SequentialRange[T]) CopyUpperBytes(bytes []byte) []byte {
	return rng.GetUpper().CopyUpperBytes(bytes)
}

// Contains returns whether this range contains all addresses in the given address or subnet.
func (rng *SequentialRange[T]) Contains(other AddressType) bool {
	if rng == nil || other == nil {
		return false
	}
	otherAddr := other.ToAddressBase().ToIP()
	if otherAddr == nil { // not an IP address
		// three cases:
		// address is not an IP type - above N/A, here false
		// address is not the right IP type but is a nil ptr - above true (incorrectly), here true (incorrectly)
		// address is the right IP type, but is a nil ptr - above true, here true
		return false
	}
	rng = rng.init()
	lower := rng.lower
	if lower.getAddrType() != otherAddr.getAddrType() {
		return false
	}
	return compareLowerValuesDifferentTypes(otherAddr, rng.lower) >= 0 &&
		compareUpperValuesDifferentTypes(otherAddr, rng.upper) <= 0
}

// ContainsRange returns whether all the addresses in the given sequential range are also contained in this sequential range.
func (rng *SequentialRange[T]) ContainsRange(other IPAddressSeqRangeType) bool {
	if rng == nil || other == nil {
		return false
	}
	otherRange := other.ToIP()
	if otherRange == nil {
		return false
	}
	rng = rng.init()
	if rng.lower.getAddrType() != otherRange.lower.getAddrType() {
		return false
	}
	otherLower, otherUpper := otherRange.GetLowerAndUpper()
	return compareLowerValuesDifferentTypes(otherLower, rng.lower) >= 0 &&
		compareLowerValuesDifferentTypes(otherUpper, rng.upper) <= 0
}

// Enumerate indicates where an address sits relative to the range ordering.
//
// Determines how many address elements of a range precede the given address element, if the address is in the range.
// If above the range, it is the distance to the upper boundary added to the range count less one, and if below the range, the distance to the lower boundary.
//
// In other words, if the given address is not in the range but above it, returns the number of addresses preceding the address from the upper range boundary,
// added to one less than the total number of range addresses.  If the given address is not in the subnet but below it, returns the number of addresses following the address to the lower subnet boundary.
//
// Returns nil when the argument is multi-valued. The argument must be an individual address.
//
// Enumerate is the inverse of the increment method:
//   - rng.Enumerate(rng.Increment(inc)) = inc
//   - rng.Increment(rng.Enumerate(newAddr)) = newAddr
//
// If the given address is not the same version as this range, then nil is returned.
func (rng *SequentialRange[T]) Enumerate(address AddressType) *big.Int {
	addr, isNil, ok := ConvertAddressTypeCheckNil[T](address)
	if !ok || isNil || rng == nil {
		return nil
	}
	return rng.enumerate(addr)
}

func (rng *SequentialRange[T]) enumerate(other T) *big.Int {
	lower := rng.GetLower()
	if lower.Equal(other) {
		return bigZero()
	} else if rng.GetUpper().Equal(other) {
		count := rng.GetCount()
		return count.Sub(count, bigOneConst())
	}
	return lower.Enumerate(other)
}

// Equal returns whether the given sequential address range is equal to this sequential address range.
// Two sequential address ranges are equal if their lower and upper range boundaries are equal.
func (rng *SequentialRange[T]) Equal(other IPAddressSeqRangeType) bool {
	if rng == nil {
		return other == nil || other.ToIP() == nil // nil contains
	} else if other == nil {
		return false //  nil contains
	}
	rng = rng.init()
	otherRange := other.ToIP()
	if otherRange == nil {
		return false
	}
	if rng.IsMultiple() {
		if !otherRange.IsMultiple() {
			return false
		}
		otherLower, otherUpper := otherRange.GetLowerAndUpper()
		lower, upper := rng.GetLowerAndUpper()
		if lower.getAddrType() != otherLower.getAddrType() {
			return false
		}
		return compareLowerValuesDifferentTypes(otherLower, lower) == 0 &&
			compareLowerValuesDifferentTypes(otherUpper, upper) == 0
	}
	return !otherRange.IsMultiple() && rng.lower.Equal(otherRange.GetLower())
}

// EqualAggregation returns true if and only if this sequential range of addresses has the same set of individual addresses as the given aggregation of addresses
func (rng *SequentialRange[T]) EqualAggregation(otherAggregation AddressAggregation) bool {
	if rng == nil {
		return IsEmpty(otherAggregation)
	}
	switch other := otherAggregation.(type) {
	case nil:
		return IsEmpty(rng)
	case IPAddressType:
		return rng.equalAddr(other)
	case IPAddressSeqRangeType:
		return rng.Equal(other)
	case *IPAddressContainmentTrie:
		return other.equalRange(rng)
	case *IPv4AddressContainmentTrie:
		return other.equalRange(rng)
	case *IPv6AddressContainmentTrie:
		return other.equalRange(rng)
	case *IPAddressSeqRangeList:
		return other.equalRange(rng)
	case *IPv4AddressSeqRangeList:
		return other.equalRange(rng)
	case *IPv6AddressSeqRangeList:
		return other.equalRange(rng)
	default:
		return equalAggregation(rng, otherAggregation)
	}
}

func (rng *SequentialRange[T]) equalAddr(other IPAddressType) bool {
	if rng == nil {
		return isEmptyAddr(other)
	} else if other == nil {
		return false
	}
	otherAddr := other.ToAddressBase()
	if otherAddr == nil {
		return false
	}
	if otherAddr.IsSequential() {
		if otherAddr.IsMultiple() {
			return rng.IsMultiple() &&
				compareLowerValuesDifferentTypes(otherAddr, rng.lower) == 0 &&
				compareUpperValuesDifferentTypes(otherAddr, rng.upper) == 0
		}
		return !rng.IsMultiple() &&
			compareLowerValuesDifferentTypes(otherAddr, rng.lower) == 0
	}
	return false
}

// Compare returns a negative integer, zero, or a positive integer if this sequential address range is less than, equal, or greater than the given item.
// Any address item is comparable to any other.  All address items use CountComparator to compare.
func (rng *SequentialRange[T]) Compare(item AddressItem) int {
	if rng != nil {
		rng = rng.init()
	}
	return CountComparator.Compare(rng, item)
}

// CompareSize compares the counts of two address ranges or items, the number of individual addresses or items within each.
//
// Rather than calculating counts with GetCount, there can be more efficient ways of determining whether this range spans more individual addresses than another item.
//
// CompareSize returns a positive integer if this range has a larger count than the item given, zero if they are the same, or a negative integer if the other has a larger count.
func (rng *SequentialRange[T]) CompareSize(other AddressItem) int {
	if rng == nil {
		if isNilItem(other) {
			return 0
		}
		// we have size 0, other has size >= 1
		return -1
	}
	return compareCounts(rng, other)
}

// GetValue returns the lowest address in the range, the one with the lowest numeric value, as an integer.
func (rng *SequentialRange[T]) GetValue() *big.Int {
	return rng.GetLower().GetValue()
}

// GetUpperValue returns the highest address in the range, the one with the highest numeric value, as an integer.
func (rng *SequentialRange[T]) GetUpperValue() *big.Int {
	return rng.GetUpper().GetValue()
}

func rangeIterator(
	lower, upper *IPAddress,
	valsAreMultiple bool,
	prefixLen PrefixLen,
	segProducer func(addr *IPAddress, index int) *IPAddressSegment,
	segmentIteratorProducer func(seg *IPAddressSegment, index int) Iterator[*IPAddressSegment],
	segValueComparator func(seg1, seg2 *IPAddress, index int) bool,
	networkSegmentIndex,
	hostSegmentIndex int,
	prefixedSegIteratorProducer func(seg *IPAddressSegment, index int) Iterator[*IPAddressSegment],
) Iterator[*Address] {
	divCount := lower.GetSegmentCount()

	// at any given point in time, this list provides an iterator for the segment at each index
	segIteratorProducerList := make([]func() Iterator[*IPAddressSegment], divCount)

	// at any given point in time, finalValue[i] is true if and only if we have reached the very last value for segment i - 1
	// when that happens, the next iterator for the segment at index i will be the last
	finalValue := make([]bool, divCount+1)

	// here is how the segment iterators will work:
	// the low and high values of the range at each segment are low, high
	// the maximum possible values for any segment are min, max
	// we first find the first k >= 0 such that low != high for the segment at index k

	//	the initial set of iterators at each index are as follows:
	//    for i < k finalValue[i] is set to true right away.
	//		we create an iterator from seg = new Seg(low)
	//    for i == k we create a wrapped iterator from Seg(low, high), wrapper will set finalValue[i] once we reach the final value of the iterator
	//    for i > k we create an iterator from Seg(low, max)
	//
	// after the initial iterator has been supplied, any further iterator supplied for the same segment is as follows:
	//    for i <= k, there was only one iterator, there will be no further iterator
	//    for i > k,
	//	  	if i == 0 or of if flagged[i - 1] is true, we create a wrapped iterator from Seg(low, high), wrapper will set finalValue[i] once we reach the final value of the iterator
	//      otherwise we create an iterator from Seg(min, max)
	//
	// By following these rules, we iterate through all possible addresses

	notDiffering := true
	finalValue[0] = true
	var allSegShared *IPAddressSegment
	for i := 0; i < divCount; i++ {
		var segIteratorProducer func(seg *IPAddressSegment, index int) Iterator[*IPAddressSegment]
		if prefixedSegIteratorProducer != nil && i >= networkSegmentIndex {
			segIteratorProducer = prefixedSegIteratorProducer
		} else {
			segIteratorProducer = segmentIteratorProducer
		}
		lowerSeg := segProducer(lower, i)
		indexi := i
		if notDiffering {
			notDiffering = segValueComparator(lower, upper, i)
			if notDiffering {
				// there is only one iterator and it produces only one value
				finalValue[i+1] = true
				iterator := segIteratorProducer(lowerSeg, i)
				segIteratorProducerList[i] = func() Iterator[*IPAddressSegment] { return iterator }
			} else {
				// in the first differing segment the only iterator will go from segment value of lower address to segment value of upper address
				iterator := segIteratorProducer(
					createAddressDivision(lowerSeg.deriveNewMultiSeg(lowerSeg.getSegmentValue(), upper.GetGenericSegment(i).GetSegmentValue(), nil)).ToIP(),
					i)
				wrappedFinalIterator := &wrappedIterator{
					iterator:   iterator,
					finalValue: finalValue,
					indexi:     indexi,
				}
				segIteratorProducerList[i] = func() Iterator[*IPAddressSegment] { return wrappedFinalIterator }
			}
		} else {
			// in the second and all following differing segments, rather than go from segment value of lower address to segment value of upper address
			// we go from segment value of lower address to the max seg value the first time through
			// then we go from the min value of the seg to the max seg value each time until the final time,
			// the final time we go from the min value to the segment value of upper address
			// we know it is the final time through when the previous iterator has reached its final value, which we track

			// the first iterator goes from the segment value of lower address to the max value of the segment
			firstIterator := segIteratorProducer(
				createAddressDivision(lowerSeg.deriveNewMultiSeg(lowerSeg.getSegmentValue(), lower.GetMaxSegmentValue(), nil)).ToIP(),
				i)

			// the final iterator goes from 0 to the segment value of our upper address
			finalIterator := segIteratorProducer(
				createAddressDivision(lowerSeg.deriveNewMultiSeg(0, upper.GetGenericSegment(i).GetSegmentValue(), nil)).ToIP(),
				i)

			// the wrapper iterator detects when the final iterator has reached its final value
			wrappedFinalIterator := &wrappedIterator{
				iterator:   finalIterator,
				finalValue: finalValue,
				indexi:     indexi,
			}
			if allSegShared == nil {
				allSegShared = createAddressDivision(lowerSeg.deriveNewMultiSeg(0, lower.GetMaxSegmentValue(), nil)).ToIP()
			}
			// all iterators after the first iterator and before the final iterator go from 0 the max segment value,
			// and there will be many such iterators
			finalIteratorProducer := func() Iterator[*IPAddressSegment] {
				if finalValue[indexi] {
					return wrappedFinalIterator
				}
				return segIteratorProducer(allSegShared, indexi)
			}
			segIteratorProducerList[i] = func() Iterator[*IPAddressSegment] {
				//the first time through, we replace the iterator producer so the first iterator used only once (ie we remove this function from the list)
				segIteratorProducerList[indexi] = finalIteratorProducer
				return firstIterator
			}
		}
	}
	iteratorProducer := func(iteratorIndex int) Iterator[*AddressSegment] {
		iter := segIteratorProducerList[iteratorIndex]()
		return wrappedSegmentIterator[*IPAddressSegment]{iter}
	}
	return rangeAddrIterator(
		false,
		lower.ToAddressBase(),
		prefixLen,
		valsAreMultiple,
		rangeSegmentsIterator(
			divCount,
			iteratorProducer,
			networkSegmentIndex,
			hostSegmentIndex,
			iteratorProducer,
		),
	)
}

// Iterator provides an iterator to iterate through the individual addresses of this address range.
//
// Call GetCount for the count.
func (rng *SequentialRange[T]) Iterator() Iterator[T] {
	if rng == nil {
		return nilIterator[T]()
	}
	rng = rng.init()
	lower := rng.lower
	if !rng.IsMultiple() {
		return &singleIterator[T]{original: lower}
	}
	divCount := lower.GetSegmentCount()
	return lower.iteratorWrapper(rangeIterator(
		lower.ToIP(),
		rng.upper.ToIP(),
		false,
		nil,
		(*IPAddress).GetSegment,
		func(seg *IPAddressSegment, index int) Iterator[*IPAddressSegment] {
			return seg.Iterator()
		},
		func(addr1, addr2 *IPAddress, index int) bool {
			return addr1.getSegment(index).getSegmentValue() == addr2.getSegment(index).getSegmentValue()
		},
		divCount-1,
		divCount,
		nil))
}

// AddressIterator is the same as Iterator while satisying the AddressAggregation interface
func (rng *SequentialRange[T]) AddressIterator() Iterator[AddressType] {
	return addrTypeIterator[T]{rng.Iterator()}
}

type segPrefData struct {
	prefLen PrefixLen
	shift   BitCount
}

// PrefixBlockIterator provides an iterator to iterate through the individual prefix blocks of the given prefix length,
// one for each prefix of that length in the address range.
func (rng *SequentialRange[T]) PrefixBlockIterator(prefLength BitCount) Iterator[T] {
	rng = rng.init()
	lower := rng.lower
	if !rng.IsMultiple() {
		return &singleIterator[T]{original: lower.ToPrefixBlockLen(prefLength)}
	}
	prefLength = checkSubnet(lower, prefLength)
	bitsPerSegment := lower.GetBitsPerSegment()
	bytesPerSegment := lower.GetBytesPerSegment()
	segCount := lower.GetSegmentCount()
	segPrefs := make([]segPrefData, segCount)
	networkSegIndex := getNetworkSegmentIndex(prefLength, bytesPerSegment, bitsPerSegment)
	for i := networkSegIndex; i < segCount; i++ {
		segPrefLength := getPrefixedSegmentPrefixLength(bitsPerSegment, prefLength, i)
		segPrefs[i] = segPrefData{segPrefLength, bitsPerSegment - segPrefLength.bitCount()}
	}
	hostSegIndex := getHostSegmentIndex(prefLength, bytesPerSegment, bitsPerSegment)
	return lower.iteratorWrapper(
		rangeIterator(
			lower.ToIP(),
			rng.upper.ToIP(),
			true,
			cacheBitCount(prefLength),
			(*IPAddress).GetSegment,
			func(seg *IPAddressSegment, index int) Iterator[*IPAddressSegment] {
				return seg.Iterator()
			},
			func(addr1, addr2 *IPAddress, index int) bool {
				segPref := segPrefs[index]
				if segPref.prefLen == nil {
					return addr1.GetSegment(index).GetSegmentValue() == addr2.GetSegment(index).GetSegmentValue()
				}
				shift := segPref.shift
				return addr1.GetSegment(index).GetSegmentValue()>>uint(shift) == addr2.GetSegment(index).GetSegmentValue()>>uint(shift)

			},
			networkSegIndex,
			hostSegIndex,
			func(seg *IPAddressSegment, index int) Iterator[*IPAddressSegment] {
				segPref := segPrefs[index]
				segPrefLen := segPref.prefLen
				if segPrefLen == nil {
					return seg.Iterator()
				}
				return seg.PrefixedBlockIterator(segPrefLen.bitCount())
			},
		))
}

// PrefixIterator provides an iterator to iterate through the individual prefixes of the given prefix length in this address range,
// each iterated element spanning the range of values for its prefix.
//
// It is similar to the prefix block iterator, except for possibly the first and last iterated elements, which might not be prefix blocks,
// instead constraining themselves to values from this range.
//
// Since a range between two arbitrary addresses cannot always be represented with a single IPAddress instance,
// the returned iterator iterates through SequentialRange instances.
//
// For instance, if iterating from "1.2.3.4" to "1.2.4.5" with prefix 8, the range shares the same prefix of value 1,
// but the range cannot be represented by the address "1.2.3-4.4-5" which does not include "1.2.3.255" or "1.2.4.0" both of which are in the original range.
// Nor can the range be represented by "1.2.3-4.0-255" which includes "1.2.4.6" and "1.2.3.3", both of which were not in the original range.
// A SequentialRange is thus required to represent that prefixed range.
func (rng *SequentialRange[T]) PrefixIterator(prefLength BitCount) Iterator[*SequentialRange[T]] {
	rng = rng.init()
	lower := rng.lower
	if !rng.IsMultiple() {
		return &singleIterator[*SequentialRange[T]]{original: rng}
	}
	prefLength = checkSubnet(lower, prefLength)
	return &sequRangeIterator[T]{
		rng:                 rng,
		creator:             newSequRange[T],
		prefixBlockIterator: rng.PrefixBlockIterator(prefLength),
		prefixLength:        prefLength,
	}
}

// isContainedBy indicates if the range is contained by the address
func isContainedBy(rng IPAddressSeqRangeType, other *IPAddress) bool {
	if rng == nil {
		return false
	} else if other == nil || other.ToIP() == nil {
		return false
	}
	lower := rng.GetLowerIPAddress()
	upper := rng.GetUpperIPAddress()
	if lower.getAddrType() != other.getAddrType() {
		return false
	}
	segCount := lower.GetSegmentCount()
	for i := 0; i < segCount; i++ {
		lowerSeg := lower.GetSegment(i)
		upperSeg := upper.GetSegment(i)
		lowerSegValue := lowerSeg.GetSegmentValue()
		upperSegValue := upperSeg.GetSegmentValue()
		otherSeg := other.GetSegment(i)
		otherSegLowerValue := otherSeg.GetSegmentValue()
		otherSegUpperValue := otherSeg.GetUpperSegmentValue()
		if lowerSegValue < otherSegLowerValue || upperSegValue > otherSegUpperValue {
			return false
		}
		if lowerSegValue != upperSegValue {
			for j := i + 1; j < segCount; j++ {
				otherSeg = other.GetSegment(j)
				if !otherSeg.IsFullRange() {
					return false
				}
			}
			break
		}
	}
	return true
}

// OverlapsAddr returns true if and only the given individual address or subnet contains at least one individual address that is also in this sequential range of addresses.
// Implements the IPAddressAggregation interface.
func (rng *SequentialRange[T]) OverlapsAddr(other AddressType) bool {
	if a, ok := any(other).(IPAddressType); ok {
		return rng.OverlapsAddress(a)
	}
	return false
}

// OverlapsAddress returns true if and only the given individual address or subnet contains at least one individual address that is also in this sequential range of addresses.
// Implements the IPAddressCollAddrConstraint interface.
func (rng *SequentialRange[T]) OverlapsAddress(other IPAddressType) bool {
	if rng == nil || other == nil {
		return false
	}
	otherAddr := other.ToIP()
	if otherAddr == nil { // not an IP address
		return false
	}

	rng = rng.init()
	lower := rng.lower.ToIP()
	if lower.getAddrType() != otherAddr.getAddrType() {
		return false
	}
	upper := rng.upper.ToIP()
	segCount := lower.GetSegmentCount()
	for i := 0; i < segCount; i++ {
		lowerSeg := lower.GetSegment(i)
		upperSeg := upper.GetSegment(i)
		lowerSegValue := lowerSeg.GetSegmentValue()
		upperSegValue := upperSeg.GetSegmentValue()
		otherSeg := otherAddr.GetSegment(i)
		otherSegLowerValue := otherSeg.GetSegmentValue()
		otherSegUpperValue := otherSeg.GetUpperSegmentValue()
		if lowerSegValue == upperSegValue {
			if lowerSegValue < otherSegLowerValue || lowerSegValue > otherSegUpperValue {
				return false
			}
		} else {
			if otherSegLowerValue < upperSegValue && otherSegUpperValue > lowerSegValue {
				return true
			} else if otherSegLowerValue == upperSegValue {
				for j := i + 1; j < segCount; j++ {
					otherSeg = otherAddr.GetSegment(j)
					upperSeg = upper.GetSegment(j)
					upperSegValue = upperSeg.GetSegmentValue()
					otherSegLowerValue = otherSeg.GetSegmentValue()
					if otherSegLowerValue < upperSegValue {
						return true
					} else if otherSegLowerValue > upperSegValue {
						return false
					}
				}
				break
			} else if otherSegUpperValue == lowerSegValue {
				for j := i + 1; j < segCount; j++ {
					otherSeg = otherAddr.GetSegment(j)
					lowerSeg = lower.GetSegment(j)
					lowerSegValue = lowerSeg.getSegmentValue()
					otherSegUpperValue = otherSeg.getUpperSegmentValue()
					if otherSegUpperValue > lowerSegValue {
						return true
					} else if otherSegUpperValue < lowerSegValue {
						return false
					}
				}
				break
			} else {
				return false
			}
		}
	}
	return true
}

// Overlaps returns true if this sequential range overlaps with the given sequential range.
func (rng *SequentialRange[T]) Overlaps(other *SequentialRange[T]) bool {
	return rng != nil && other != nil && rng.overlaps(other)
}

func (rng *SequentialRange[T]) overlaps(other *SequentialRange[T]) bool {
	lower, upper := rng.GetLowerAndUpper()
	otherLower, otherUpper := other.GetLowerAndUpper()
	if lower.getAddrType() != otherLower.getAddrType() {
		return false
	}
	return overlapsCheck(lower, upper, otherLower, otherUpper)
}

func overlapsCheck[T SequentialRangeConstraint[T]](lower, upper, otherLower, otherUpper T) bool {
	return compareLowerValues(otherLower, upper) <= 0 && compareLowerValues(otherUpper, lower) >= 0
}

// OverlapsRange returns true if this sequential range overlaps with the given sequential range.
func (rng *SequentialRange[T]) OverlapsRange(other IPAddressSeqRangeType) bool {
	r, isNil, ok := ConvertRangeTypeCheckNil[T](other)
	return ok && !isNil && rng != nil && rng.overlaps(r)
}

// Intersect returns the intersection of this range with the given range, a range which includes those addresses found in both.
// It returns nil if there is no common address.
func (rng *SequentialRange[T]) Intersect(other *SequentialRange[T]) *SequentialRange[T] {
	rng = rng.init()
	other = other.init()
	if rng.lower.getAddrType() != other.lower.getAddrType() {
		return nil
	}
	otherLower, otherUpper := other.GetLower(), other.GetUpper()
	lower, upper := rng.lower, rng.upper
	if compareLowerValues(lower, otherLower) <= 0 {
		if compareLowerValues(upper, otherUpper) >= 0 { // l, ol, ou, u
			return other
		}
		comp := compareLowerValues(upper, otherLower)
		if comp < 0 { // l, u, ol, ou
			return nil
		}
		return newSequRangeUnchecked(otherLower, upper, comp != 0) // l, ol, u,  ou
	} else if compareLowerValues(otherUpper, upper) >= 0 {
		return rng
	}
	comp := compareLowerValues(otherUpper, lower)
	if comp < 0 {
		return nil
	}
	return newSequRangeUnchecked(lower, otherUpper, comp != 0)
}

// CoverWithSequentialRange implements the IPAddressAggregationConstraint interface
func (rng *SequentialRange[T]) CoverWithSequentialRange() *SequentialRange[T] {
	return rng
}

// CoverWithPrefixBlock returns the minimal-size prefix block that covers all the addresses in this range.
// The resulting block will have a larger count than this, unless this range already directly corresponds to a prefix block.
func (rng *SequentialRange[T]) CoverWithPrefixBlock() T {
	rng = rng.init()
	return rng.lower.CoverWithPrefixBlockTo(rng.upper)
}

// SpanWithPrefixBlocks returns an array of prefix blocks that spans the same set of addresses as this range.
func (rng *SequentialRange[T]) SpanWithPrefixBlocks() []T {
	rng = rng.init()
	return rng.lower.SpanWithPrefixBlocksTo(rng.upper)
}

// SpanningPrefixBlockIterator returns the result of SpanWithPrefixBlocks as an iterator.
//
// Individual addresses will be shown with a prefix length extending to the end of the address.
// They are represented as 2001:4860:4860::8844/128 or 192.168.10.1/32, instead of 2001:4860:4860::8844 or 192.168.10.1.
// You can esily remove such prefix lengths with calls to RemoveBitcountPrefixLen.
func (rng *SequentialRange[T]) SpanningPrefixBlockIterator() Iterator[T] {
	return &sliceIterator[T]{rng.SpanWithPrefixBlocks()}
}

// SpanWithSequentialBlocks produces the smallest slice of sequential blocks that cover the same set of addresses as this range.
// This slice can be shorter than that produced by SpanWithPrefixBlocks and is never longer.
func (rng *SequentialRange[T]) SpanWithSequentialBlocks() []T {
	rng = rng.init()
	return rng.lower.SpanWithSequentialBlocksTo(rng.upper)
}

// SpanningSeqBlockIterator returns the result of SpanWithSequentialBlocks as an iterator.
func (rng *SequentialRange[T]) SpanningSeqBlockIterator() Iterator[T] {
	return &sliceIterator[T]{rng.SpanWithSequentialBlocks()}
}

// Join joins the receiver with the given ranges into the fewest number of ranges.
// The returned array will be sorted by ascending lowest range value.
// Nil ranges are tolerated, and ignored.
func (rng *SequentialRange[T]) Join(ranges ...*SequentialRange[T]) []*SequentialRange[T] {
	ranges = append(append(make([]*SequentialRange[T], 0, len(ranges)+1), ranges...), rng)
	return joinRanges(ranges, true, true)
}

// JoinTo joins this range to the other if they are contiguous.  If this range overlaps with the given range,
// or if the highest value of the lower range is one below the lowest value of the higher range,
// then the two are joined into a new larger range that is returned.
// Otherwise, nil is returned.
func (rng *SequentialRange[T]) JoinTo(other *SequentialRange[T]) *SequentialRange[T] {
	lower, upper := rng.GetLowerAndUpper()
	otherLower, otherUpper := other.GetLowerAndUpper()
	if lower.getAddrType() != otherLower.getAddrType() {
		return nil
	}
	lowerComp := compareLowerValues(lower, otherLower)
	singleJoin := rng.joinOverlapping(lowerComp, lower, upper, otherLower, otherUpper)
	if singleJoin != nil {
		return singleJoin
	}
	if lowerComp > 0 {
		if otherUpper.upperIsAdjacentTo(lower) {
			return newSequRangeUnchecked[T](otherLower, upper, true)
		}
	} else {
		if upper.upperIsAdjacentTo(otherLower) {
			return newSequRangeUnchecked[T](lower, otherUpper, true)
		}
	}
	return nil
}

func (rng *SequentialRange[T]) joinOverlapping(lowerComp int, lower, upper, otherLower, otherUpper T) *SequentialRange[T] {
	if overlapsCheck(lower, upper, otherLower, otherUpper) {
		upperComp := compareLowerValues(upper, otherUpper)
		var lowestLower, highestUpper T
		if lowerComp >= 0 {
			if lowerComp == 0 && upperComp == 0 {
				return rng
			}
			lowestLower = otherLower
		} else {
			lowestLower = lower
		}
		if upperComp >= 0 {
			highestUpper = upper
		} else {
			highestUpper = otherUpper
		}
		//highestUpper = upperComp >= 0 ? upper : otherUpper;
		return newSequRangeUnchecked(lowestLower, highestUpper, true)
	}
	return nil
}

// JoinIntoList creates the minimal number of range lists from the receiver combined with the given ranges.
// Nil ranges are tolerated, and ignored.
// If the input ranges comprise multiple versions of IP addresses, then multiple lists will be returned, the IPv4 followed by the IPv6 list.
// If there are no non-nil input ranges, then nil is returned.
func (rng *SequentialRange[T]) JoinIntoList(ranges ...*SequentialRange[T]) []*SequentialRangeList[T] {
	res := joinRanges(ranges, false, false)
	resLen := len(res)
	if resLen == 0 {
		return nil
	}
	capacity := resLen << 1
	var (
		hasPrevious     bool
		previousAddress T
		multipleLists   []*SequentialRangeList[T]
		previousList    *SequentialRangeList[T]
		list            *SequentialRangeList[T] = NewSequentialRangeList[T](capacity)
	)
	for _, r := range res {
		if r == nil {
			continue
		}
		next := r.GetLower()
		if hasPrevious && !versionsMatch(previousAddress, next) {
			if previousList != nil {
				// second time we switch versions, which is not possible if just IPv4/v6, we may have the zero-valued IP range as well
				multipleLists = append(multipleLists, previousList, list)
				previousList = nil
			} else if multipleLists != nil {
				// third time we switch versions
				multipleLists = append(multipleLists, list)
			} else {
				// first time we switch versions
				previousList = list
			}
			list = NewSequentialRangeList[T](capacity)
		}
		previousAddress = next
		hasPrevious = true
		list.ranges = append(list.ranges, *r)
	}
	if multipleLists != nil {
		return append(multipleLists, list)
	} else if previousList != nil {
		return []*SequentialRangeList[T]{previousList, list}
	}
	return []*SequentialRangeList[T]{list}
}

// JoinToIntoList joins this range to the other.
//
// Similar to JoinTo, but instead the result includes all the addresses in both ranges, regardless of whether they are contiguous,
// unless the two ranges have different versions, in which case nil is returned.
func (rng *SequentialRange[T]) JoinToIntoList(other *SequentialRange[T]) *SequentialRangeList[T] {
	lower, upper := rng.GetLowerAndUpper()
	otherLower, otherUpper := other.GetLowerAndUpper()
	if !versionsMatch(lower, otherLower) {
		return nil
	}
	lowerComp := compareLowerValues(lower, otherLower)
	singleJoin := rng.joinOverlapping(lowerComp, lower, upper, otherLower, otherUpper)
	if singleJoin == nil {
		if lowerComp > 0 {
			if otherUpper.upperIsAdjacentTo(lower) {
				return createSingleRangeList(newSequRangeUnchecked(otherLower, upper, true))
			}
			return createDoubleRangeList(other, rng)
		}
		fmt.Println("hello")
		if upper.upperIsAdjacentTo(otherLower) {
			return createSingleRangeList(newSequRangeUnchecked(lower, otherUpper, true))
		}
		return createDoubleRangeList(rng, other)
	}
	return createSingleRangeList(singleJoin)
}

// Extend extends this sequential range to include all address in the given range.
// If the argument has a different IP version than this, nil is returned.
// Otherwise, this method returns the range that includes this range, the given range, and all addresses in-between.
func (rng *SequentialRange[T]) Extend(other *SequentialRange[T]) *SequentialRange[T] {
	rng = rng.init()
	other = other.init()
	lower, upper := rng.GetLowerAndUpper()
	otherLower, otherUpper := other.GetLowerAndUpper()
	if lower.getAddrType() != otherLower.getAddrType() {
		return nil
	}
	lowerComp := compareLowerValues(lower, otherLower)
	upperComp := compareLowerValues(upper, otherUpper)
	if lowerComp > 0 { //
		if upperComp <= 0 { // ol l u ou
			return other
		}
		// ol l ou u or ol ou l u
		return newSequRangeUnchecked(otherLower, upper, true)
	}
	// lowerComp <= 0
	if upperComp >= 0 { // l ol ou u
		return rng
	}
	return newSequRangeUnchecked(lower, otherUpper, true) // l ol u ou or l u ol ou
}

// Subtract subtracts the given range from the receiver range, to produce either zero, one, or two address ranges that contain the addresses in the receiver range and not in the given range.
// If the result has length 2, the two ranges are ordered by ascending lowest range value.
func (rng *SequentialRange[T]) Subtract(other *SequentialRange[T]) []*SequentialRange[T] {
	return subtract(rng, other, createEmptyRanges[T], createSingleRange[T], createDoubleRange[T])
}

func (rng *SequentialRange[T]) SubtractIntoList(other *SequentialRange[T]) *SequentialRangeList[T] {
	return subtract(rng, other, createEmptyRangeList[T], createSingleRangeList[T], createDoubleRangeList[T])
}

func subtract[T SequentialRangeConstraint[T], U any](rng, other *SequentialRange[T],
	createEmpty func() U,
	createSingle func(*SequentialRange[T]) U,
	createDouble func(one, two *SequentialRange[T]) U) U {
	rng = rng.init()
	other = other.init()
	if rng.lower.getAddrType() != other.lower.getAddrType() {
		return createSingle(rng)
	}
	otherLower, otherUpper := other.GetLowerAndUpper()
	lower, upper := rng.lower, rng.upper
	if compareLowerValues(lower, otherLower) < 0 {
		if compareLowerValues(upper, otherUpper) > 0 { // l ol ou u
			return createDouble(
				newSequRangeCheckSize(lower, otherLower.DecrementSingle()),
				newSequRangeCheckSize(otherUpper.IncrementSingle(), upper))
		} else {
			comp := compareLowerValues(upper, otherLower)
			if comp < 0 { // l u ol ou
				return createSingle(rng)
			} else if comp == 0 { // l u == ol ou
				return createSingle(newSequRangeCheckSize(lower, upper.DecrementSingle()))
			}
			return createSingle(newSequRangeCheckSize(lower, otherLower.DecrementSingle())) // l ol u ou
		}
	} else if compareLowerValues(otherUpper, upper) >= 0 { // ol l u ou
		return createEmpty()
	} else {
		comp := compareLowerValues(otherUpper, lower)
		if comp < 0 {
			return createSingle(rng) // ol ou l u
		} else if comp == 0 {
			return createSingle(newSequRangeCheckSize(lower.IncrementSingle(), upper)) // ol ou == l u
		}
		return createSingle(newSequRangeCheckSize(otherUpper.IncrementSingle(), upper)) // ol l ou u
	}
}

func (rng *SequentialRange[T]) Complement() []*SequentialRange[T] {
	return complement(rng, createEmptyRanges[T], createSingleRange[T], createDoubleRange[T])
}

func (rng *SequentialRange[T]) ComplementIntoList() *SequentialRangeList[T] {
	return complement(rng, createEmptyRangeList[T], createSingleRangeList[T], createDoubleRangeList[T])
}

func complement[T SequentialRangeConstraint[T], U any](
	rng *SequentialRange[T],
	createEmpty func() U,
	createSingle func(*SequentialRange[T]) U,
	createDouble func(one, two *SequentialRange[T]) U) U {
	lower, upper := rng.GetLowerAndUpper()
	if lower.IncludesZero() {
		if upper.IncludesMax() {
			return createEmpty()
		}
		network := lower.GetIPNetwork()
		_, max := network.GetBoundaryAddresses()
		newRng := newSequRangeCheckSize(upper.IncrementSingle(), max)
		return createSingle(newRng)
	}
	network := lower.GetIPNetwork()
	zero, max := network.GetBoundaryAddresses()
	if upper.IncludesMax() {
		newRng := newSequRangeCheckSize(zero, lower.DecrementSingle())
		return createSingle(newRng)
	}
	first := newSequRangeCheckSize(zero, lower.DecrementSingle())
	second := newSequRangeCheckSize(upper.IncrementSingle(), max)
	return createDouble(first, second)
}

// Split splits this range at the given address into two ranges, one lower and one upper.
// The second range starts with the lower address of the given address or subnet.
// The first range consists of all preceding addresses.
//
// This is similar to subtract, but without removing the given address or subnet from the result.
//
// In some cases, one or both of the two returned ranges is nil.
//
// If the given address or subnet includes the first address in this range,
// or all addresses of the given address or subnet are below the lower value of this range,
// then the first range is nil, and the second range is the same range as this range.
//
// If all addresses of the given address or subnet are above the upper value of this range,
// then the first range is is the same range as this range, and the second range is nil.
//
// If the given address has a different version than this, then both returned ranges are nil.
func (rng *SequentialRange[T]) Split(other T) (lowerFromSplit, upperFromSplit *SequentialRange[T]) {
	lower := rng.GetLower()
	if lower.getAddrType() != other.getAddrType() {
		return
	}
	if compareLowerValues(lower, other) < 0 {
		upper := rng.GetUpper()
		if compareLowerValues(upper, other) >= 0 { // l ol u
			otherLower := other.WithoutPrefixLen().GetLower()
			lowerFromSplit = newSequRangeCheckSize(lower, otherLower.DecrementSingle())
			upperFromSplit = newSequRangeCheckSize(otherLower, upper)
			return
		}
		// l u ol
		lowerFromSplit = rng
		return
	}
	// ol l u
	upperFromSplit = rng
	return
}

// LowerFromSplit is the same as split, but returns only the lower range.
func (rng *SequentialRange[T]) LowerFromSplit(other T) (lowerFromSplit *SequentialRange[T]) {
	lower := rng.GetLower()
	if lower.getAddrType() != other.getAddrType() {
		return
	}
	if compareLowerValues(lower, other) < 0 {
		upper := rng.GetUpper()
		if compareLowerValues(upper, other) >= 0 { // l ol u
			return newSequRangeCheckSize(lower, other.WithoutPrefixLen().DecrementSingle())
		}
		// l u ol
		return rng
	}
	// ol l u
	return
}

func (rng *SequentialRange[T]) lowerSplit(other T) (lowerFromSplit *SequentialRange[T]) {
	return newSequRangeCheckSize(rng.GetLower(), other.WithoutPrefixLen().DecrementSingle())
}

// UpperFromSplit is the same as split, but returns only the upper range.
func (rng *SequentialRange[T]) UpperFromSplit(other T) (upperFromSplit *SequentialRange[T]) {
	lower := rng.GetLower()
	if !versionsMatch(lower, other) {
		return
	}
	if compareLowerValues(lower, other) < 0 {
		upper := rng.GetUpper()
		if compareLowerValues(upper, other) >= 0 { // l ol u
			return newSequRangeCheckSize(other.WithoutPrefixLen().GetLower(), rng.GetUpper())
		}
		// l u ol
		return
	}
	// ol l u
	return rng
}

func (rng *SequentialRange[T]) upperSplit(other T) (upperFromSplit *SequentialRange[T]) {
	return newSequRangeCheckSize(other.WithoutPrefixLen().GetLower(), rng.GetUpper())
}

// IntoSequentialRangeList creates a new sequential range list collection containing all the individual addresses in this sequential range list.
func (rng *SequentialRange[T]) IntoSequentialRangeList() *SequentialRangeList[T] {
	return &SequentialRangeList[T]{
		ranges: []SequentialRange[T]{*rng.init()},
	}
}

// IntoContainmentTrie creates a new containement trie collection containing all the individual addresses in this sequential range list.
func (rng *SequentialRange[T]) IntoContainmentTrie() *ContainmentTrie[T] {
	trie := &ContainmentTrie[T]{}
	trie.AddSeqRange(rng)
	return trie
}

// ToKey creates the associated address range key.
// While address ranges can be compared with the Compare or Equal methods as well as various provided instances of AddressComparator,
// they are not comparable with Go operators.
// However, SequentialRangeKey instances are comparable with Go operators, and thus can be used as map keys.
func (rng *SequentialRange[T]) ToKey() SequentialRangeKey[T] {
	return newSequentialRangeKey(rng.init())
}

// IsIPv4 returns true if this sequential address range is an IPv4 sequential address range.  If so, use ToIPv4 to convert to the IPv4-specific type.
func (rng *SequentialRange[T]) IsIPv4() bool { // returns false when lower is nil
	if rng != nil {
		t := any(rng.GetLower())
		if _, ok := t.(*IPv4Address); ok {
			return true
		} else if addr, ok := t.(*IPAddress); ok {
			return addr.IsIPv4()
		}
	}
	return false
}

// IsIPv6 returns true if this sequential address range is an IPv6 sequential address range.  If so, use ToIPv6 to convert to the IPv6-specific type.
func (rng *SequentialRange[T]) IsIPv6() bool { // returns false when lower is nil
	if rng != nil {
		t := any(rng.GetLower())
		if _, ok := t.(*IPv6Address); ok {
			return true
		} else if addr, ok := t.(*IPAddress); ok {
			return addr.IsIPv6()
		}
	}
	return false
}

// ToIPv4 converts to a SequentialRange[*IPv4Address] if this address range is an IPv4 address range.
// If not, ToIPv4 returns nil.
//
// ToIPv4 can be called with a nil receiver, enabling you to chain this method with methods that might return a nil pointer.
func (rng *SequentialRange[T]) ToIPv4() *SequentialRange[*IPv4Address] {
	if rng != nil {
		if ipv4, ok := any(rng).(*SequentialRange[*IPv4Address]); ok {
			return ipv4
		} else {
			t := any(rng.GetLower())
			if addr, ok := t.(*IPAddress); ok && addr.IsIPv4() {
				t = any(rng.GetUpper())
				return newSequRangeUnchecked(addr.ToIPv4(), t.(*IPAddress).ToIPv4(), rng.IsMultiple())
			}
		}
	}
	return nil
}

// ToIPv6 converts to a SequentialRange[*IPv6Address] if this address range is an IPv6 address range.
// If not, ToIPv6 returns nil.
//
// ToIPv6 can be called with a nil receiver, enabling you to chain this method with methods that might return a nil pointer.
func (rng *SequentialRange[T]) ToIPv6() *SequentialRange[*IPv6Address] {
	if rng != nil {
		if ipv6, ok := any(rng).(*SequentialRange[*IPv6Address]); ok {
			return ipv6
		} else {
			t := any(rng.GetLower())
			if addr, ok := t.(*IPAddress); ok && addr.IsIPv6() {
				t = any(rng.GetUpper())
				return newSequRangeUnchecked(addr.ToIPv6(), t.(*IPAddress).ToIPv6(), rng.IsMultiple())
			}
		}
	}
	return nil
}

// ToIP converts to a SequentialRange[*IPAddress], a polymorphic type usable with all IP address sequential ranges.
//
// ToIP can be called with a nil receiver, enabling you to chain this method with methods that might return a nil pointer.
func (rng *SequentialRange[T]) ToIP() *SequentialRange[*IPAddress] {
	if rng != nil {
		if ip, ok := any(rng).(*SequentialRange[*IPAddress]); ok {
			return ip
		}
		rng = rng.init()
		return newSequRangeUnchecked(rng.lower.ToIP(), rng.upper.ToIP(), rng.IsMultiple())
	}
	return nil
}

func newSequRangeUnchecked[T SequentialRangeConstraint[T]](lower, upper T, isMult bool) *SequentialRange[T] {
	return &SequentialRange[T]{
		lower: lower,
		upper: upper,
		cache: &rangeCache{isMultiple: isMult},
	}
}

func newSequRangeCheckSize[T SequentialRangeConstraint[T]](lower, upper T) *SequentialRange[T] {
	return newSequRangeUnchecked(lower, upper, !lower.equalsSingleSameVersion(upper))
}

func newSequRange[T SequentialRangeConstraint[T]](first, other T) *SequentialRange[T] {
	var lower, upper T
	var isMult bool
	if f := first.Contains(other); f || other.Contains(first) {
		var addr T
		if f {
			addr = first.WithoutPrefixLen()
		} else {
			addr = other.WithoutPrefixLen()
		}
		lower = addr.GetLower()
		if isMult = addr.IsMultiple(); isMult {
			upper = addr.GetUpper()
		} else {
			upper = lower
		}
	} else {
		// We find the lowest and the highest from both supplied addresses
		firstLower, firstUpper := first.GetLowerAndUpper()
		otherLower, otherUpper := other.GetLowerAndUpper()
		if comp := compareLowerValues(firstLower, otherLower); comp > 0 {
			isMult = true
			lower = otherLower
		} else {
			isMult = comp < 0
			lower = firstLower
		}
		if comp := compareLowerValues(firstUpper, otherUpper); comp < 0 {
			isMult = true
			upper = otherUpper
		} else {
			isMult = isMult || comp > 0
			upper = firstUpper
		}
		if isMult = isMult || compareLowerValues(lower, upper) != 0; isMult {
			lower = lower.WithoutPrefixLen()
			upper = upper.WithoutPrefixLen()
		} else {
			if lower.IsPrefixed() {
				if upper.IsPrefixed() {
					lower = lower.WithoutPrefixLen()
					upper = lower
				} else {
					lower = upper
				}
			} else {
				upper = lower
			}
		}
	}
	// note that this method ensures that the init method has been called on both lower and upper,
	// which also means it is safe to call getAddrType on either one (getAddrType should not be used when init not yet called)
	return newSequRangeUnchecked(lower, upper, isMult)
}

func newSequRangeOrdered[T SequentialRangeConstraint[T]](lower, upper T) *SequentialRange[T] {
	lower, upper = lower.WithoutPrefixLen(), upper.WithoutPrefixLen()
	comp := compareLowerValues(lower, upper)
	return newSequRangeUnchecked(lower, upper, comp != 0)
}

// NewSequentialRange creates a sequential range from the given addresses.
// A nil value argument is equivalent to the zero value of the type of T, which then needs to be inferred by the other argument or the function call.
// If the type of T is *IPAddress and the versions of lower and upper do not match (one is IPv4, one IPv6), then nil is returned.
// Otherwise, the range is returned.
func NewSequentialRange[T SequentialRangeConstraint[T]](lower, upper T) *SequentialRange[T] {
	newVal, lowerIsNil := nilConvert(lower)
	_, upperIsNil := nilConvert(upper)
	if lowerIsNil { // nil for pointers
		if upperIsNil {
			lower = newVal
			upper = newVal
		} else {
			lower = upper
		}
	} else if upperIsNil {
		upper = lower
	} else {
		// this check only matters when T is *IPAddress
		// Using getAddrType is NOT safe here because T may be *IPv4Address or *IPv6Address, and so we need to ensure init() is called before calling getAddrType
		//if lower.getAddrType() != upper.getAddrType() {
		if !lower.GetIPVersion().Equal(upper.GetIPVersion()) {
			// when both are zero-type, we do not go in here
			// but if only one is, we return nil.  zero-type is "indeterminate", so we cannot "infer" a different version for it
			// However, nil is the absence of a version/type, so we can and do
			return nil
		}
	}
	return newSequRange(lower, upper)
}

// NewIPSeqRange creates an IP sequential range from the given addresses.
// It is here for backwards compatibility. NewSequentialRange is recommended instead.
// If the type of T is *IPAddress and the versions of lower and upper do not match (one is IPv4, one IPv6), then nil is returned.
// Otherwise, the range is returned.
func NewIPSeqRange(lower, upper *IPAddress) *SequentialRange[*IPAddress] { // for backwards compatibility
	if lower == nil {
		if upper == nil {
			lower = zeroIPAddr
			upper = zeroIPAddr
		} else {
			lower = upper
		}
	} else if upper == nil {
		upper = lower
	} else {
		// Using getAddrType is safe here because we use IPAddress so if it is zeroType that is accurate
		if lower.getAddrType() != upper.getAddrType() {
			// when both are zero-type, we do not go in here
			// but if only one is, we do go in here and return nil.  zero-type is "indeterminate", chosen to be neither IPv4 or IPv6, so we cannot "infer" a different version for it
			// However, nil is the absence of a version/type so we can infer the version, and we do
			return nil
		}
	}
	return newSequRange(lower, upper)
}

// NewIPv4SeqRange creates an IPv4 sequential range from the given addresses.
// It is here for backwards compatibility. NewSequentialRange is recommended instead.
func NewIPv4SeqRange(lower, upper *IPv4Address) *SequentialRange[*IPv4Address] { // for backwards compatibility
	if lower == nil {
		if upper == nil {
			lower = zeroIPv4
			upper = zeroIPv4
		} else {
			lower = upper
		}
	} else if upper == nil {
		upper = lower
	}
	return newSequRange(lower, upper)
}

// NewIPv6SeqRange creates an IPv6 sequential range from the given addresses.
// It is here for backwards compatibility. NewSequentialRange is recommended instead.
func NewIPv6SeqRange(lower, upper *IPv6Address) *SequentialRange[*IPv6Address] { // for backwards compatibility
	if lower == nil {
		if upper == nil {
			lower = zeroIPv6
			upper = zeroIPv6
		} else {
			lower = upper
		}
	} else if upper == nil {
		upper = lower
	}
	return newSequRange(lower, upper)
}

func joinRanges[T SequentialRangeConstraint[T]](ranges []*SequentialRange[T], canAlterInitial, isFinal bool) (ret []*SequentialRange[T]) {
	if !canAlterInitial {
		ranges = append(make([]*SequentialRange[T], 0, len(ranges)), ranges...)
	}
	// nil entries are automatic joins
	joinedCount := 0
	rangesLen := len(ranges)
	for i, j := 0, rangesLen-1; i <= j; i++ {
		if ranges[i] == nil {
			joinedCount++
			for ranges[j] == nil && j > i {
				j--
				joinedCount++
			}
			if j > i {
				ranges[i] = ranges[j]
				ranges[j] = nil
				j--
			}
		}
	}
	rangesLen = rangesLen - joinedCount
	ranges = ranges[:rangesLen]
	joinedCount = 0
	sort.Slice(ranges, func(i, j int) bool {
		return LowValueComparator.CompareRanges(ranges[i], ranges[j]) < 0
	})
	for i := 0; i < rangesLen; {
		rng := ranges[i]
		currentLower, currentUpper := rng.GetLowerAndUpper()
		var isMultiJoin, didJoin bool
		j := i + 1
		for ; j < rangesLen; j++ {
			rng2 := ranges[j]
			nextLower := rng2.GetLower()
			if nextLower.getAddrType() != currentUpper.getAddrType() {
				break
			}
			doJoin := compareLowerValues(currentUpper, nextLower) >= 0
			if !doJoin {
				doJoin = currentUpper.upperIsAdjacentTo(nextLower)
				isMultiJoin = true
			}
			if doJoin {
				//Join them
				joinedCount++
				nextUpper := rng2.GetUpper()
				if compareLowerValues(currentUpper, nextUpper) < 0 {
					currentUpper = nextUpper
				}
				ranges[j] = nil
				isMultiJoin = isMultiJoin || rng.IsMultiple() || rng2.IsMultiple()
				didJoin = true
			} else {
				break
			}
		}
		if didJoin {
			ranges[i] = newSequRangeUnchecked(currentLower, currentUpper, isMultiJoin)
		}
		i = j
	}
	if isFinal {
		finalLen := rangesLen - joinedCount
		if finalLen > 0 {
			for i, j := 0, 0; ; i++ {
				rng := ranges[i]
				if rng == nil {
					continue
				}
				ranges[j] = rng
				j++
				if j >= finalLen {
					break
				}
			}
		}
		ret = ranges[:finalLen]
	} else {
		ret = ranges
	}
	return
}

// getMinPrefixLenForBlock returns the smallest prefix length such that the upper and lower values span the block of values for that prefix length.
// The given bit count indicates the bits that matter in the two values, the remaining bits are ignored.
//
// If the entire range can be described this way, then this method returns the same value as GetPrefixLenForSingleBlock.
//
// There may be a single prefix, or multiple possible prefix values in this item for the returned prefix length.
// Use getPrefixLenForSingleBlock to avoid the case of multiple prefix values.
func getMinPrefixLenForBlock(lower, upper DivInt, bitCount BitCount) BitCount {
	if lower == upper {
		return bitCount
	} else if lower == 0 {
		maxValue := ^(^DivInt(0) << uint(bitCount))
		if upper == maxValue {
			return 0
		}
	}
	result := bitCount
	lowerZeros := bits.TrailingZeros64(lower)
	if lowerZeros != 0 {
		upperOnes := bits.TrailingZeros64(^upper)
		if upperOnes != 0 {
			var prefixedBitCount int
			if lowerZeros < upperOnes {
				prefixedBitCount = lowerZeros
			} else {
				prefixedBitCount = upperOnes
			}
			result -= BitCount(prefixedBitCount)
		}
	}
	return result
}

// getPrefixLenForSingleBlock returns a prefix length for which the given lower and upper values share the same prefix,
// and the range spanned by those values matches exactly the block of all values for that prefix.
// The given bit count indicates the bits that matter in the two values, the remaining bits are ignored.
//
// If the range can be described this way, then this method returns the same value as GetMinPrefixLenForBlock.
//
// If no such prefix length exists, returns nil.
//
// If lower and upper values are the same, this returns the bit count.
func getPrefixLenForSingleBlock(lower, upper DivInt, bitCount BitCount) PrefixLen {
	prefixLen := getMinPrefixLenForBlock(lower, upper, bitCount)
	if prefixLen == bitCount {
		if lower == upper {
			return cacheBitCount(prefixLen)
		}
	} else {
		shift := bitCount - prefixLen
		if lower>>uint(shift) == upper>>uint(shift) {
			return cacheBitCount(prefixLen)
		}
	}
	return nil
}

func createEmptyRanges[T SequentialRangeConstraint[T]]() []*SequentialRange[T] {
	return make([]*SequentialRange[T], 0, 0)
}

func createSingleRange[T SequentialRangeConstraint[T]](rng *SequentialRange[T]) []*SequentialRange[T] {
	return []*SequentialRange[T]{rng}
}

func createDoubleRange[T SequentialRangeConstraint[T]](rng1, rng2 *SequentialRange[T]) []*SequentialRange[T] {
	return []*SequentialRange[T]{rng1, rng2}
}

func createEmptyRangeList[T SequentialRangeConstraint[T]]() *SequentialRangeList[T] {
	return &SequentialRangeList[T]{}
}

func createSingleRangeList[T SequentialRangeConstraint[T]](rng *SequentialRange[T]) *SequentialRangeList[T] {
	return &SequentialRangeList[T]{
		ranges: []SequentialRange[T]{*rng},
	}
}

func createDoubleRangeList[T SequentialRangeConstraint[T]](rng1, rng2 *SequentialRange[T]) *SequentialRangeList[T] {
	return &SequentialRangeList[T]{
		ranges: []SequentialRange[T]{*rng1, *rng2},
	}
}

type (
	IPAddressSeqRange   = SequentialRange[*IPAddress]
	IPv4AddressSeqRange = SequentialRange[*IPv4Address]
	IPv6AddressSeqRange = SequentialRange[*IPv6Address]
)
