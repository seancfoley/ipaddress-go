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
	"math/bits"
	"net"
	"sort"
	"strings"
	"sync/atomic"
	"unsafe"
)

const DefaultSeqRangeSeparator = " -> "

type rangeCache struct {
	cachedCount *big.Int
}

type ipAddressSeqRangeInternal struct {
	lower, upper *IPAddress
	isMult       bool // set on construction, even for zero values
	cache        *rangeCache
}

func (rng *ipAddressSeqRangeInternal) isMultiple() bool {
	return rng.isMult
}

func (rng *ipAddressSeqRangeInternal) getCount() *big.Int {
	return rng.getCachedCount(true)
}

func (rng *ipAddressSeqRangeInternal) getCachedCount(copy bool) (res *big.Int) {
	cache := rng.cache
	count := cache.cachedCount
	if count == nil {
		if !rng.isMultiple() {
			count = bigOne()
		} else if ipv4Range := rng.toIPv4SequentialRange(); ipv4Range != nil {
			upper := int64(ipv4Range.GetUpper().Uint32Value())
			lower := int64(ipv4Range.GetLower().Uint32Value())
			val := upper - lower + 1
			count = new(big.Int).SetInt64(val)
		} else {
			count = rng.upper.GetValue()
			res = rng.lower.GetValue()
			count.Sub(count, res).Add(count, bigOneConst())
			res.Set(count)
		}
		dataLoc := (*unsafe.Pointer)(unsafe.Pointer(&cache.cachedCount))
		atomic.StorePointer(dataLoc, unsafe.Pointer(count))
	}
	if res == nil {
		if copy {
			res = new(big.Int).Set(count)
		} else {
			res = count
		}
	}
	return
}

// GetPrefixCountLen returns the count of the number of distinct values within the prefix part of the range of addresses
func (rng *ipAddressSeqRangeInternal) GetPrefixCountLen(prefixLen BitCount) *big.Int {
	if !rng.isMultiple() { // also checks for zero-ranges
		return bigOne()
	}
	bitCount := rng.lower.GetBitCount()
	if prefixLen <= 0 {
		return bigOne()
	} else if prefixLen >= bitCount {
		return rng.getCount()
	}
	shiftAdjustment := bitCount - prefixLen
	if ipv4Range := rng.toIPv4SequentialRange(); ipv4Range != nil {
		upper := ipv4Range.GetUpper()
		lower := ipv4Range.GetLower()
		upperAdjusted := upper.Uint32Value() >> uint(shiftAdjustment)
		lowerAdjusted := lower.Uint32Value() >> uint(shiftAdjustment)
		result := int64(upperAdjusted) - int64(lowerAdjusted) + 1
		return new(big.Int).SetInt64(result)
	}
	upper := rng.upper.GetValue()
	ushiftAdjustment := uint(shiftAdjustment)
	upper.Rsh(upper, ushiftAdjustment)
	lower := rng.lower.GetValue()
	lower.Rsh(lower, ushiftAdjustment)
	upper.Sub(upper, lower).Add(upper, bigOneConst())
	return upper
}

// CompareSize returns whether this range has a larger count than the other
func (rng *ipAddressSeqRangeInternal) compareSize(other IPAddressSeqRangeType) int {
	if other == nil || other.ToIP() == nil {
		// our size is 1 or greater, other 0
		return 1
	}
	if !rng.isMultiple() {
		if other.IsMultiple() {
			return -1
		}
		return 0
	} else if !other.IsMultiple() {
		return 1
	}
	return rng.getCachedCount(false).CmpAbs(other.ToIP().getCachedCount(false))
}

func (rng *ipAddressSeqRangeInternal) contains(other IPAddressType) bool {
	if other == nil {
		return true
	}
	otherAddr := other.ToIP()
	if otherAddr == nil {
		return true
	}
	return compareLowIPAddressValues(otherAddr.GetLower(), rng.lower) >= 0 &&
		compareLowIPAddressValues(otherAddr.GetUpper(), rng.upper) <= 0
}

func (rng *ipAddressSeqRangeInternal) equals(other IPAddressSeqRangeType) bool {
	if other == nil {
		return false
	}
	otherRng := other.ToIP()
	if otherRng == nil {
		return false
	}
	return rng.lower.Equal(otherRng.GetLower()) && rng.upper.Equal(otherRng.GetUpper())
}

func (rng *ipAddressSeqRangeInternal) containsRange(other IPAddressSeqRangeType) bool {
	if other == nil {
		return true
	}
	otherRange := other.ToIP()
	if otherRange == nil {
		return true
	}
	return compareLowIPAddressValues(otherRange.GetLower(), rng.lower) >= 0 &&
		compareLowIPAddressValues(otherRange.GetUpper(), rng.upper) <= 0
}

func (rng *ipAddressSeqRangeInternal) toIPv4SequentialRange() *IPv4AddressSeqRange {
	if rng.lower != nil && rng.lower.IsIPv4() {
		return (*IPv4AddressSeqRange)(unsafe.Pointer(rng))
	}
	return nil
}

func (rng *ipAddressSeqRangeInternal) toIPSequentialRange() *IPAddressSeqRange {
	return (*IPAddressSeqRange)(unsafe.Pointer(rng))
}

func (rng *ipAddressSeqRangeInternal) overlaps(other *IPAddressSeqRange) bool {
	return compareLowIPAddressValues(other.GetLower(), rng.upper) <= 0 &&
		compareLowIPAddressValues(other.GetUpper(), rng.lower) >= 0
}

// IsSequential returns whether the address or subnet represents a range of values that are sequential.
//
// IP address sequential ranges are sequential by definition, so this returns true.
func (rng *ipAddressSeqRangeInternal) IsSequential() bool {
	return true
}

// Returns the intersection of this range with the given range, a range which includes those addresses in both this and the given range.
func (rng *ipAddressSeqRangeInternal) intersect(other *IPAddressSeqRange) *IPAddressSeqRange {
	otherLower, otherUpper := other.GetLower(), other.GetUpper()
	lower, upper := rng.lower, rng.upper
	if compareLowIPAddressValues(lower, otherLower) <= 0 {
		if compareLowIPAddressValues(upper, otherUpper) >= 0 { // l, ol, ou, u
			return other
		}
		comp := compareLowIPAddressValues(upper, otherLower)
		if comp < 0 { // l, u, ol, ou
			return nil
		}
		return newSeqRangeUnchecked(otherLower, upper, comp != 0) // l, ol, u,  ou
	} else if compareLowIPAddressValues(otherUpper, upper) >= 0 {
		return rng.toIPSequentialRange()
	}
	comp := compareLowIPAddressValues(otherUpper, lower)
	if comp < 0 {
		return nil
	}
	return newSeqRangeUnchecked(lower, otherUpper, comp != 0)
}

// If this range overlaps with the given range,
// or if the highest value of the lower range is one below the lowest value of the higher range,
// then the two are joined into a new larger range that is returned.
// Otherwise nil is returned.
func (rng *ipAddressSeqRangeInternal) joinTo(other *IPAddressSeqRange) *IPAddressSeqRange {
	otherLower, otherUpper := other.GetLower(), other.GetUpper()
	lower, upper := rng.lower, rng.upper
	lowerComp := compareLowIPAddressValues(lower, otherLower)
	if !rng.overlaps(other) {
		if lowerComp >= 0 {
			if otherUpper.increment(1).equals(lower) {
				return newSeqRangeUnchecked(otherLower, upper, true)
			}
		} else {
			if upper.increment(1).equals(otherLower) {
				return newSeqRangeUnchecked(lower, otherUpper, true)
			}
		}
		return nil
	}
	upperComp := compareLowIPAddressValues(upper, otherUpper)
	var lowestLower, highestUpper *IPAddress
	if lowerComp >= 0 {
		if lowerComp == 0 && upperComp == 0 {
			return rng.toIPSequentialRange()
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
	return newSeqRangeUnchecked(lowestLower, highestUpper, true)
}

// extend extends this sequential range to include all address in the given range.
// If the argument has a different IP version than this, nil is returned.
// Otherwise, this method returns the range that includes this range, the given range, and all addresses in-between.
func (rng *ipAddressSeqRangeInternal) extend(other *IPAddressSeqRange) *IPAddressSeqRange {
	otherLower, otherUpper := other.GetLower(), other.GetUpper()
	lower, upper := rng.lower, rng.upper
	lowerComp := compareLowIPAddressValues(lower, otherLower)
	upperComp := compareLowIPAddressValues(upper, otherUpper)
	if lowerComp > 0 { //
		if upperComp <= 0 { // ol l u ou
			return other
		}
		// ol l ou u or ol ou l u
		return newSeqRangeUnchecked(otherLower, upper, true)
	}
	// lowerComp <= 0
	if upperComp >= 0 { // l ol ou u
		return rng.toIPSequentialRange()
	}
	return newSeqRangeUnchecked(lower, otherUpper, true) // l ol u ou or l u ol ou
}

// Subtracts the given range from this range, to produce either zero, one, or two address ranges that contain the addresses in this range and not in the given range.
// If the result has length 2, the two ranges are ordered by ascending lowest range value.
func (rng *ipAddressSeqRangeInternal) subtract(other *IPAddressSeqRange) []*IPAddressSeqRange {
	otherLower, otherUpper := other.GetLower(), other.GetUpper()
	lower, upper := rng.lower, rng.upper
	if compareLowIPAddressValues(lower, otherLower) < 0 {
		if compareLowIPAddressValues(upper, otherUpper) > 0 { // l ol ou u
			return []*IPAddressSeqRange{
				newSeqRangeCheckSize(lower, otherLower.Increment(-1)),
				newSeqRangeCheckSize(otherUpper.Increment(1), upper),
			}
		} else {
			comp := compareLowIPAddressValues(upper, otherLower)
			if comp < 0 { // l u ol ou
				return []*IPAddressSeqRange{rng.toIPSequentialRange()}
			} else if comp == 0 { // l u == ol ou
				return []*IPAddressSeqRange{newSeqRangeCheckSize(lower, upper.Increment(-1))}
			}
			return []*IPAddressSeqRange{newSeqRangeCheckSize(lower, otherLower.Increment(-1))} // l ol u ou
		}
	} else if compareLowIPAddressValues(otherUpper, upper) >= 0 { // ol l u ou
		return make([]*IPAddressSeqRange, 0, 0)
	} else {
		comp := compareLowIPAddressValues(otherUpper, lower)
		if comp < 0 {
			return []*IPAddressSeqRange{rng.toIPSequentialRange()} // ol ou l u
		} else if comp == 0 {
			return []*IPAddressSeqRange{newSeqRangeCheckSize(lower.Increment(1), upper)} // ol ou == l u
		}
		return []*IPAddressSeqRange{newSeqRangeCheckSize(otherUpper.Increment(1), upper)} // ol l ou u
	}
}

// ContainsPrefixBlock returns whether the range contains the block of addresses for the given prefix length.
//
// Unlike ContainsSinglePrefixBlock, whether there are multiple prefix values for the given prefix length makes no difference.
//
// Use GetMinPrefixLenForBlock to determine whether there is a prefix length for which this method returns true.
func (rng *ipAddressSeqRangeInternal) ContainsPrefixBlock(prefixLen BitCount) bool {
	lower := rng.lower
	if lower == nil {
		return true // returns true for 0 bits
	}
	prefixLen = checkSubnet(lower, prefixLen)
	upper := rng.upper
	divCount := lower.GetDivisionCount()
	bitsPerSegment := lower.GetBitsPerSegment()
	i := getHostSegmentIndex(prefixLen, lower.GetBytesPerSegment(), bitsPerSegment)
	if i < divCount {
		div := lower.GetSegment(i)
		upperDiv := upper.GetSegment(i)
		segmentPrefixLength := getPrefixedSegmentPrefixLength(bitsPerSegment, prefixLen, i)
		if !div.isPrefixBlockVals(div.getDivisionValue(), upperDiv.getDivisionValue(), segmentPrefixLength.bitCount()) {
			return false
		}
		for i++; i < divCount; i++ {
			div = lower.GetSegment(i)
			upperDiv = upper.GetSegment(i)
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
func (rng *ipAddressSeqRangeInternal) ContainsSinglePrefixBlock(prefixLen BitCount) bool {
	lower := rng.lower
	if lower == nil {
		return true // returns true for 0 bits
	}
	prefixLen = checkSubnet(lower, prefixLen)
	var prevBitCount BitCount
	upper := rng.upper
	divCount := lower.GetDivisionCount()
	for i := 0; i < divCount; i++ {
		div := lower.GetSegment(i)
		upperDiv := upper.GetSegment(i)
		bitCount := div.GetBitCount()
		totalBitCount := bitCount + prevBitCount
		if prefixLen >= totalBitCount {
			if !divValSame(div.getDivisionValue(), upperDiv.getDivisionValue()) {
				return false
			}
		} else {
			divPrefixLen := prefixLen - prevBitCount
			if !div.isPrefixBlockVals(div.getDivisionValue(), upperDiv.getDivisionValue(), divPrefixLen) {
				return false
			}
			for i++; i < divCount; i++ {
				div = lower.GetSegment(i)
				upperDiv = upper.GetSegment(i)
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
func (rng *ipAddressSeqRangeInternal) GetPrefixLenForSingleBlock() PrefixLen {
	lower := rng.lower
	if lower == nil {
		return cacheBitCount(0) // returns true for 0 bits
	}
	upper := rng.upper
	count := lower.GetSegmentCount()
	segBitCount := lower.GetBitsPerSegment()
	maxSegValue := ^(^SegInt(0) << uint(segBitCount))
	totalPrefix := BitCount(0)
	for i := 0; i < count; i++ {
		lowerSeg := lower.GetSegment(i)
		upperSeg := upper.GetSegment(i)
		segPrefix := GetPrefixLenForSingleBlock(lowerSeg.getDivisionValue(), upperSeg.getDivisionValue(), segBitCount)
		if segPrefix == nil {
			return nil
		}
		dabits := segPrefix.bitCount()
		totalPrefix += dabits
		if dabits < segBitCount {
			//remaining segments must be full range or we return nil
			for i++; i < count; i++ {
				lowerSeg = lower.GetSegment(i)
				upperSeg = upper.GetSegment(i)
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
func (rng *ipAddressSeqRangeInternal) GetMinPrefixLenForBlock() BitCount {
	lower := rng.lower
	if lower == nil {
		return 0 // returns true for 0 bits
	}
	upper := rng.upper
	count := lower.GetSegmentCount()
	totalPrefix := lower.GetBitCount()
	segBitCount := lower.GetBitsPerSegment()
	for i := count - 1; i >= 0; i-- {
		lowerSeg := lower.GetSegment(i)
		upperSeg := upper.GetSegment(i)
		segPrefix := GetMinPrefixLenForBlock(lowerSeg.getDivisionValue(), upperSeg.getDivisionValue(), segBitCount)
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
func (rng *ipAddressSeqRangeInternal) IsZero() bool {
	return rng.IncludesZero() && !rng.isMultiple()
}

// IncludesZero returns whether this sequential range's lower value is the zero address.
func (rng *ipAddressSeqRangeInternal) IncludesZero() bool {
	lower := rng.lower
	return lower == nil || lower.IsZero()
}

// IsMax returns whether this sequential range spans from the max address, the address whose bits are all ones, to itself.
func (rng *ipAddressSeqRangeInternal) IsMax() bool {
	return rng.IncludesMax() && !rng.isMultiple()
}

// IncludesMax returns whether this sequential range's upper value is the max value, the value whose bits are all ones.
func (rng *ipAddressSeqRangeInternal) IncludesMax() bool {
	upper := rng.upper
	return upper == nil || upper.IsMax()
}

// IsFullRange returns whether this address range covers the entire address space of this IP address version.
//
// This is true if and only if both IncludesZero and IncludesMax return true.
func (rng *ipAddressSeqRangeInternal) IsFullRange() bool {
	return rng.IncludesZero() && rng.IncludesMax()
}

func (rng *ipAddressSeqRangeInternal) toString(lowerStringer func(*IPAddress) string, separator string, upperStringer func(*IPAddress) string) string {
	builder := strings.Builder{}
	str1, str2, str3 := lowerStringer(rng.lower), separator, upperStringer(rng.upper)
	builder.Grow(len(str1) + len(str2) + len(str3))
	builder.WriteString(str1)
	builder.WriteString(str2)
	builder.WriteString(str3)
	return builder.String()
}

func (rng *ipAddressSeqRangeInternal) format(state fmt.State, verb rune) {
	rng.lower.Format(state, verb)
	_, _ = state.Write([]byte(DefaultSeqRangeSeparator))
	rng.upper.Format(state, verb)
}

// Iterates through the range of prefixes in this range instance using the given prefix length.
//
// Since a range between two arbitrary addresses cannot always be represented with a single IPAddress instance,
// the returned iterator iterates through IPAddressSeqRange instances.
//
// For instance, if iterating from 1.2.3.4 to 1.2.4.5 with prefix 8, the range shares the same prefix 1,
// but the range cannot be represented by the address 1.2.3-4.4-5 which does not include 1.2.3.255 or 1.2.4.0 both of which are in the original range.
// Nor can the range be represented by 1.2.3-4.0-255 which includes 1.2.4.6 and 1.2.3.3, both of which were not in the original range.
// An IPAddressSeqRange is thus required to represent that prefixed range.
func (rng *ipAddressSeqRangeInternal) prefixIterator(prefLength BitCount) IPAddressSeqRangeIterator {
	lower := rng.lower
	if !rng.isMultiple() {
		return &singleRangeIterator{original: rng.toIPSequentialRange()}
	}
	prefLength = checkSubnet(lower, prefLength)
	return &rangeIterator{
		rng:                 rng.toIPSequentialRange(),
		creator:             newSeqRange,
		prefixBlockIterator: ipAddrIterator{rng.prefixBlockIterator(prefLength)},
		prefixLength:        prefLength,
	}
}

func (rng *ipAddressSeqRangeInternal) prefixBlockIterator(prefLength BitCount) AddressIterator {
	lower := rng.lower
	if !rng.isMultiple() {
		return &singleAddrIterator{original: lower.ToPrefixBlockLen(prefLength).ToAddressBase()}
	} //else if prefLength > lower.GetBitCount() {
	//return rng.iterator()
	//}
	prefLength = checkSubnet(lower, prefLength)
	bitsPerSegment := lower.GetBitsPerSegment()
	bytesPerSegment := lower.GetBytesPerSegment()
	segCount := lower.GetSegmentCount()
	type segPrefData struct {
		prefLen PrefixLen
		shift   BitCount
	}
	segPrefs := make([]segPrefData, segCount)
	networkSegIndex := getNetworkSegmentIndex(prefLength, bytesPerSegment, bitsPerSegment)
	for i := networkSegIndex; i < segCount; i++ {
		segPrefLength := getPrefixedSegmentPrefixLength(bitsPerSegment, prefLength, i)
		segPrefs[i] = segPrefData{segPrefLength, bitsPerSegment - segPrefLength.bitCount()}
	}
	hostSegIndex := getHostSegmentIndex(prefLength, bytesPerSegment, bitsPerSegment)
	return rng.rangeIterator(
		true,
		cacheBitCount(prefLength),
		(*IPAddress).GetSegment,
		func(seg *IPAddressSegment, index int) IPSegmentIterator {
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
		func(seg *IPAddressSegment, index int) IPSegmentIterator {
			segPref := segPrefs[index]
			segPrefLen := segPref.prefLen
			if segPrefLen == nil {
				return seg.Iterator()
			}
			return seg.PrefixedBlockIterator(segPrefLen.bitCount())
		},
	)
}

func (rng *ipAddressSeqRangeInternal) iterator() AddressIterator {
	lower := rng.lower
	if !rng.isMultiple() {
		return &singleAddrIterator{original: lower.ToAddressBase()}
	}
	divCount := lower.GetSegmentCount()
	return rng.rangeIterator(
		false,
		nil,
		(*IPAddress).GetSegment,
		func(seg *IPAddressSegment, index int) IPSegmentIterator {
			return seg.Iterator()
		},
		func(addr1, addr2 *IPAddress, index int) bool {
			return addr1.getSegment(index).getSegmentValue() == addr2.getSegment(index).getSegmentValue()
		},
		divCount-1,
		divCount,
		nil)
}

func (rng *ipAddressSeqRangeInternal) rangeIterator(
	//creator parsedAddressCreator, /* nil for zero sections */
	valsAreMultiple bool,
	prefixLen PrefixLen,
	segProducer func(addr *IPAddress, index int) *IPAddressSegment,
	segmentIteratorProducer func(seg *IPAddressSegment, index int) IPSegmentIterator,
	segValueComparator func(seg1, seg2 *IPAddress, index int) bool,
	networkSegmentIndex,
	hostSegmentIndex int,
	prefixedSegIteratorProducer func(seg *IPAddressSegment, index int) IPSegmentIterator,
) AddressIterator {
	lower := rng.lower
	upper := rng.upper
	divCount := lower.getDivisionCount()

	// at any given point in time, this list provides an iterator for the segment at each index
	segIteratorProducerList := make([]func() IPSegmentIterator, divCount)

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
		var segIteratorProducer func(seg *IPAddressSegment, index int) IPSegmentIterator
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
				segIteratorProducerList[i] = func() IPSegmentIterator { return iterator }
			} else {
				// in the first differing segment the only iterator will go from segment value of lower address to segment value of upper address
				iterator := segIteratorProducer(
					createAddressDivision(lowerSeg.deriveNewMultiSeg(lowerSeg.getSegmentValue(), upper.getSegment(i).getSegmentValue(), nil)).ToIP(),
					i)
				wrappedFinalIterator := &wrappedIterator{
					iterator:   iterator,
					finalValue: finalValue,
					indexi:     indexi,
				}
				segIteratorProducerList[i] = func() IPSegmentIterator { return wrappedFinalIterator }
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
				createAddressDivision(lowerSeg.deriveNewMultiSeg(0, upper.getSegment(i).getSegmentValue(), nil)).ToIP(),
				i)

			// the wrapper iterator detects when the final iterator has reached its final value
			wrappedFinalIterator := &wrappedIterator{
				iterator:   finalIterator,
				finalValue: finalValue,
				indexi:     indexi,
			}
			if allSegShared == nil {
				allSegShared = createAddressDivision(lowerSeg.deriveNewMultiSeg(0, lower.getMaxSegmentValue(), nil)).ToIP()
			}
			// all iterators after the first iterator and before the final iterator go from 0 the max segment value,
			// and there will be many such iterators
			finalIteratorProducer := func() IPSegmentIterator {
				if finalValue[indexi] {
					return wrappedFinalIterator
				}
				return segIteratorProducer(allSegShared, indexi)
			}
			segIteratorProducerList[i] = func() IPSegmentIterator {
				//the first time through, we replace the iterator producer so the first iterator used only once (ie we remove this function from the list)
				segIteratorProducerList[indexi] = finalIteratorProducer
				return firstIterator
			}
		}
	}
	iteratorProducer := func(iteratorIndex int) SegmentIterator {
		iter := segIteratorProducerList[iteratorIndex]()
		return WrappedIPSegmentIterator{iter}
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

var zeroRange = newSeqRange(zeroIPAddr, zeroIPAddr)

type IPAddressSeqRange struct {
	ipAddressSeqRangeInternal
}

func (rng *IPAddressSeqRange) init() *IPAddressSeqRange {
	if rng.lower == nil {
		return zeroRange
	}
	return rng
}

// GetCount returns the count of addresses that this sequential range spans.
//
// Use IsMultiple if you simply want to know if the count is greater than 1.
func (rng *IPAddressSeqRange) GetCount() *big.Int {
	if rng == nil {
		return bigZero()
	}
	return rng.init().getCount()
}

// IsMultiple returns whether this range represents a range of multiple addresses
func (rng *IPAddressSeqRange) IsMultiple() bool {
	return rng != nil && rng.isMultiple()
}

// String implements the fmt.Stringer interface,
// returning the lower address canonical string, followed by the default separator " -> ",
// followed by the upper address canonical string.
// It returns "<nil>" if the receiver is a nil pointer.
func (rng *IPAddressSeqRange) String() string {
	if rng == nil {
		return nilString()
	}
	return rng.ToString((*IPAddress).String, DefaultSeqRangeSeparator, (*IPAddress).String)
}

// Format implements fmt.Formatter interface.
//
// It prints the string as "lower -> upper" where lower and upper are the formatted strings for the lowest and highest addresses in the range, given by GetLower and GetUpper.
// The formats, flags, and other specifications supported are those supported by Format in IPAddress.
func (rng IPAddressSeqRange) Format(state fmt.State, verb rune) {
	rng.init().format(state, verb)
}

func (rng *IPAddressSeqRange) ToString(lowerStringer func(*IPAddress) string, separator string, upperStringer func(*IPAddress) string) string {
	if rng == nil {
		return nilString()
	}
	return rng.init().toString(lowerStringer, separator, upperStringer)
}

// ToNormalizedString produces a normalized string for the address range.
// It has the format "lower -> upper" where lower and upper are the normalized strings for the lowest and highest addresses in the range, given by GetLower and GetUpper.
func (rng *IPAddressSeqRange) ToNormalizedString() string {
	return rng.ToString((*IPAddress).ToNormalizedString, DefaultSeqRangeSeparator, (*IPAddress).ToNormalizedString)
}

// ToCanonicalString produces a canonical string for the address range.
// It has the format "lower -> upper" where lower and upper are the canonical strings for the lowest and highest addresses in the range, given by GetLower and GetUpper.
func (rng *IPAddressSeqRange) ToCanonicalString() string {
	return rng.ToString((*IPAddress).ToCanonicalString, DefaultSeqRangeSeparator, (*IPAddress).ToCanonicalString)
}

// GetLowerIPAddress satisfies the IPAddressRange interface, returning the lower address in the range, same as GetLower()
func (rng *IPAddressSeqRange) GetLowerIPAddress() *IPAddress {
	return rng.GetLower()
}

// GetUpperIPAddress satisfies the IPAddressRange interface, returning the upper address in the range, same as GetUpper()
func (rng *IPAddressSeqRange) GetUpperIPAddress() *IPAddress {
	return rng.GetUpper()
}

// GetLower returns the lowest address in the range, the one with the lowest numeric value
func (rng *IPAddressSeqRange) GetLower() *IPAddress {
	return rng.init().lower
}

// GetUpper returns the highest address in the range, the one with the highest numeric value
func (rng *IPAddressSeqRange) GetUpper() *IPAddress {
	return rng.init().upper
}

// GetBitCount returns the number of bits in each address in the range
func (rng *IPAddressSeqRange) GetBitCount() BitCount {
	return rng.GetLower().GetBitCount()
}

// GetByteCount returns the number of bytes in each address in the range
func (rng *IPAddressSeqRange) GetByteCount() int {
	return rng.GetLower().GetByteCount()
}

// GetNetIP returns the lower IP address in the range as a net.IP
func (rng *IPAddressSeqRange) GetNetIP() net.IP {
	return rng.GetLower().GetNetIP()
}

func (rng *IPAddressSeqRange) CopyNetIP(bytes net.IP) net.IP {
	return rng.GetLower().CopyNetIP(bytes) // this changes the arg to 4 bytes if 16 bytes and ipv4
}

// GetUpperNetIP returns the upper IP address in the range as a net.IP
func (rng *IPAddressSeqRange) GetUpperNetIP() net.IP {
	return rng.GetUpper().GetUpperNetIP()
}

func (rng *IPAddressSeqRange) CopyUpperNetIP(bytes net.IP) net.IP {
	return rng.GetUpper().CopyUpperNetIP(bytes) // this changes the arg to 4 bytes if 16 bytes and ipv4
}

// Bytes returns the lowest address in the range, the one with the lowest numeric value, as a byte slice
func (rng *IPAddressSeqRange) Bytes() []byte {
	return rng.GetLower().Bytes()
}

// CopyBytes copies the value of the lowest address in the range into a byte slice.
//
// If the value can fit in the given slice, the value is copied into that slice and a length-adjusted sub-slice is returned.
// Otherwise, a new slice is created and returned with the value.
func (rng *IPAddressSeqRange) CopyBytes(bytes []byte) []byte {
	return rng.GetLower().CopyBytes(bytes)
}

// UpperBytes returns the highest address in the range, the one with the highest numeric value, as a byte slice
func (rng *IPAddressSeqRange) UpperBytes() []byte {
	return rng.GetUpper().UpperBytes()
}

// CopyUpperBytes copies the value of the highest address in the range into a byte slice.
//
// If the value can fit in the given slice, the value is copied into that slice and a length-adjusted sub-slice is returned.
// Otherwise, a new slice is created and returned with the value.
func (rng *IPAddressSeqRange) CopyUpperBytes(bytes []byte) []byte {
	return rng.GetUpper().CopyUpperBytes(bytes)
}

// Contains returns whether this range contains all addresses in the given address or subnet.
func (rng *IPAddressSeqRange) Contains(other IPAddressType) bool {
	if rng == nil {
		return other == nil || other.ToAddressBase() == nil
	}
	return rng.init().contains(other)
}

// ContainsRange returns whether all the addresses in the given sequential range are also contained in this sequential range.
func (rng *IPAddressSeqRange) ContainsRange(other IPAddressSeqRangeType) bool {
	if rng == nil {
		return other == nil || other.ToIP() == nil
	}
	return rng.init().containsRange(other)
}

// Equal returns whether the given sequential address range is equal to this sequential address range.
// Two sequential address ranges are equal if their lower and upper range boundaries are equal.
func (rng *IPAddressSeqRange) Equal(other IPAddressSeqRangeType) bool {
	if rng == nil {
		return other == nil || other.ToIP() == nil
	}
	return rng.init().equals(other)
}

// Compare returns a negative integer, zero, or a positive integer if this sequential address range is less than, equal, or greater than the given item.
// Any address item is comparable to any other.  All address items use CountComparator to compare.
func (rng *IPAddressSeqRange) Compare(item AddressItem) int {
	if rng != nil {
		rng = rng.init()
	}
	return CountComparator.Compare(rng, item)
}

// CompareSize compares the counts of two address ranges, the number of individual addresses within.
//
// Rather than calculating counts with GetCount, there can be more efficient ways of comparing whether one range spans more individual addresses than another.
//
// CompareSize returns a positive integer if this range has a larger count than the one given, 0 if they are the same, or a negative integer if the other has a larger count.
func (rng *IPAddressSeqRange) CompareSize(other IPAddressSeqRangeType) int {
	if rng == nil {
		if other != nil && other.ToIP() != nil {
			// we have size 0, other has size >= 1
			return -1
		}
		return 0
	}
	return rng.compareSize(other)
}

// GetValue returns the lowest address in the range, the one with the lowest numeric value, as an integer
func (rng *IPAddressSeqRange) GetValue() *big.Int {
	return rng.GetLower().GetValue()
}

// GetUpperValue returns the highest address in the range, the one with the highest numeric value, as an integer
func (rng *IPAddressSeqRange) GetUpperValue() *big.Int {
	return rng.GetUpper().GetValue()
}

// Iterator provides an iterator to iterate through the individual addresses of this address range.
//
// Call GetCount for the count.
func (rng *IPAddressSeqRange) Iterator() IPAddressIterator {
	if rng == nil {
		return ipAddrIterator{nilAddrIterator()}
	}
	return &ipAddrIterator{rng.init().iterator()}
}

// PrefixBlockIterator provides an iterator to iterate through the individual prefix blocks of the given prefix length,
// one for each prefix of that length in the address range.
func (rng *IPAddressSeqRange) PrefixBlockIterator(prefLength BitCount) IPAddressIterator {
	return &ipAddrIterator{rng.init().prefixBlockIterator(prefLength)}
}

// PrefixIterator provides an iterator to iterate through the individual prefixes of the given prefix length in this address range,
// each iterated element spanning the range of values for its prefix.
//
// It is similar to the prefix block iterator, except for possibly the first and last iterated elements, which might not be prefix blocks,
// instead constraining themselves to values from this range.
func (rng *IPAddressSeqRange) PrefixIterator(prefLength BitCount) IPAddressSeqRangeIterator {
	return rng.init().prefixIterator(prefLength)
}

// ToIP is an identity method.
//
// ToIP can be called with a nil receiver, enabling you to chain this method with methods that might return a nil pointer.
func (rng *IPAddressSeqRange) ToIP() *IPAddressSeqRange {
	return rng
}

// IsIPv4 returns true if this sequential address range originated as an IPv4 sequential address range.  If so, use ToIPv4 to convert back to the IPv4-specific type.
func (rng *IPAddressSeqRange) IsIPv4() bool { // returns false when lower is nil
	return rng != nil && rng.GetLower().IsIPv4()
}

// IsIPv6 returns true if this sequential address range originated as an IPv6 sequential address range.  If so, use ToIPv6 to convert back to the IPv6-specific type.
func (rng *IPAddressSeqRange) IsIPv6() bool { // returns false when lower is nil
	return rng != nil && rng.GetLower().IsIPv6()
}

// ToIPv4 converts to an IPv4AddressSeqRange if this address range originated as an IPv4 address range.
// If not, ToIPv4 returns nil.
//
// ToIPv4 can be called with a nil receiver, enabling you to chain this method with methods that might return a nil pointer.
func (rng *IPAddressSeqRange) ToIPv4() *IPv4AddressSeqRange {
	if rng.IsIPv4() {
		return (*IPv4AddressSeqRange)(rng)
	}
	return nil
}

// ToIPv6 converts to an IPv6AddressSeqRange if this address range originated as an IPv6 address range.
// If not, ToIPv6 returns nil.
//
// ToIPv6 can be called with a nil receiver, enabling you to chain this method with methods that might return a nil pointer.
func (rng *IPAddressSeqRange) ToIPv6() *IPv6AddressSeqRange {
	if rng.IsIPv6() {
		return (*IPv6AddressSeqRange)(rng)
	}
	return nil
}

func (rng *IPAddressSeqRange) Overlaps(other *IPAddressSeqRange) bool {
	return rng.init().overlaps(other)
}

func (rng *IPAddressSeqRange) Intersect(other *IPAddressSeqRange) *IPAddressSeqRange {
	return rng.init().intersect(other)
}

func (rng *IPAddressSeqRange) CoverWithPrefixBlock() *IPAddress {
	return rng.GetLower().CoverWithPrefixBlockTo(rng.GetUpper())
}

func (rng *IPAddressSeqRange) SpanWithPrefixBlocks() []*IPAddress {
	return rng.GetLower().SpanWithPrefixBlocksTo(rng.GetUpper())
}

func (rng *IPAddressSeqRange) SpanWithSequentialBlocks() []*IPAddress {
	res := rng.GetLower().SpanWithSequentialBlocksTo(rng.GetUpper())
	return res
}

// Join joins the given ranges into the fewest number of ranges.
// The returned array will be sorted by ascending lowest range value.
func (rng *IPAddressSeqRange) Join(ranges ...*IPAddressSeqRange) []*IPAddressSeqRange {
	ranges = append(append(make([]*IPAddressSeqRange, 0, len(ranges)+1), ranges...), rng)
	return join(ranges)
}

// JoinTo joins this range to the other.  If this range overlaps with the given range,
// or if the highest value of the lower range is one below the lowest value of the higher range,
// then the two are joined into a new larger range that is returned.
// Otherwise nil is returned.
func (rng *IPAddressSeqRange) JoinTo(other *IPAddressSeqRange) *IPAddressSeqRange {
	return rng.init().joinTo(other.init())
}

// Extend extends this sequential range to include all address in the given range.
// If the argument has a different IP version than this, nil is returned.
// Otherwise, this method returns the range that includes this range, the given range, and all addresses in-between.
func (rng *IPAddressSeqRange) Extend(other *IPAddressSeqRange) *IPAddressSeqRange {
	rng = rng.init()
	other = other.init()
	if !versionsMatch(rng.lower, other.lower) {
		return nil
	}
	return rng.extend(other)
}

// Subtract subtracts the given range from this range, to produce either zero, one, or two address ranges that contain the addresses in this range and not in the given range.
// If the result has length 2, the two ranges are ordered by ascending lowest range value.
func (rng *IPAddressSeqRange) Subtract(other *IPAddressSeqRange) []*IPAddressSeqRange {
	return rng.init().subtract(other.init())
}

func (rng *IPAddressSeqRange) ToKey() *IPAddressSeqRangeKey {
	return &IPAddressSeqRangeKey{
		lower: *rng.GetLower().ToKey(),
		upper: *rng.GetUpper().ToKey(),
	}
}

func newSeqRangeUnchecked(lower, upper *IPAddress, isMult bool) *IPAddressSeqRange {
	return &IPAddressSeqRange{
		ipAddressSeqRangeInternal{
			lower:  lower,
			upper:  upper,
			isMult: isMult,
			cache:  &rangeCache{},
		},
	}
}

func newSeqRangeCheckSize(lower, upper *IPAddress) *IPAddressSeqRange {
	return newSeqRangeUnchecked(lower, upper, !lower.equalsSameVersion(upper))
}

func newSeqRange(first, other *IPAddress) *IPAddressSeqRange {
	var lower, upper *IPAddress
	var isMult bool
	if f := first.Contains(other); f || other.Contains(first) {
		var addr *IPAddress
		if f {
			addr = first.WithoutPrefixLen()
		} else {
			addr = other.WithoutPrefixLen()
		}
		lower = addr.GetLower()
		if isMult = addr.isMultiple(); isMult {
			upper = addr.GetUpper()
		} else {
			upper = lower
		}
	} else {
		firstLower := first.GetLower()
		otherLower := other.GetLower()
		firstUpper := first.GetUpper()
		otherUpper := other.GetUpper()
		if comp := compareLowIPAddressValues(firstLower, otherLower); comp > 0 {
			isMult = true
			lower = otherLower
		} else {
			isMult = comp < 0
			lower = firstLower
		}
		if comp := compareLowIPAddressValues(firstUpper, otherUpper); comp < 0 {
			isMult = true
			upper = otherUpper
		} else {
			isMult = comp > 0
			upper = firstUpper
		}
		lower = lower.WithoutPrefixLen()
		if isMult = isMult || compareLowIPAddressValues(lower, upper) != 0; isMult {
			upper = upper.WithoutPrefixLen()
		} else {
			upper = lower
		}
	}
	return newSeqRangeUnchecked(lower, upper, isMult)
}

func join(ranges []*IPAddressSeqRange) []*IPAddressSeqRange {
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
	for i := 0; i < rangesLen; i++ {
		rng := ranges[i]
		if rng == nil {
			continue
		}
		currentLower, currentUpper := rng.GetLower(), rng.GetUpper()
		var isMultiJoin, didJoin bool
		for j := i + 1; j < rangesLen; j++ {
			rng2 := ranges[j]
			nextLower := rng2.GetLower()
			doJoin := compareLowIPAddressValues(currentUpper, nextLower) >= 0
			if !doJoin {
				doJoin = currentUpper.increment(1).equals(nextLower)
				isMultiJoin = true
			}
			if doJoin {
				//Join them
				nextUpper := rng2.GetUpper()
				if compareLowIPAddressValues(currentUpper, nextUpper) < 0 {
					currentUpper = nextUpper
				}
				ranges[j] = nil
				isMultiJoin = isMultiJoin || rng.isMult || rng2.isMult
				joinedCount++
				didJoin = true
			} else {
				break
			}
		}
		if didJoin {
			ranges[i] = newSeqRangeUnchecked(currentLower, currentUpper, isMultiJoin)
		}
	}
	finalLen := rangesLen - joinedCount
	for i, j := 0, 0; i < rangesLen; i++ {
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
	ret := ranges[:finalLen]
	return ret
}

func compareLowValues(one, two *Address) int {
	return LowValueComparator.CompareAddresses(one, two)
}

func compareLowIPAddressValues(one, two *IPAddress) int {
	return LowValueComparator.CompareAddresses(one, two)
}

func GetMinPrefixLenForBlock(lower, upper DivInt, bitCount BitCount) BitCount {
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

func GetPrefixLenForSingleBlock(lower, upper DivInt, bitCount BitCount) PrefixLen {
	prefixLen := GetMinPrefixLenForBlock(lower, upper, bitCount)
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
