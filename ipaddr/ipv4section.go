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
	"math/big"
	"sync/atomic"
	"unsafe"

	"github.com/seancfoley/ipaddress-go/ipaddr/addrerr"
	"github.com/seancfoley/ipaddress-go/ipaddr/addrstr"
)

func createIPv4Section(segments []*AddressDivision) *IPv4AddressSection {
	return &IPv4AddressSection{
		ipAddressSectionInternal{
			addressSectionInternal{
				addressDivisionGroupingInternal{
					addressDivisionGroupingBase: addressDivisionGroupingBase{
						divisions: standardDivArray{segments},
						addrType:  ipv4Type,
						cache: &valueCache{
							stringCache: stringCache{
								ipStringCache:   &ipStringCache{},
								ipv4StringCache: &ipv4StringCache{},
							},
						},
					},
				},
			},
		},
	}
}

func newIPv4SectionParsed(segments []*AddressDivision, isMultiple bool) (res *IPv4AddressSection) {
	res = createIPv4Section(segments)
	res.isMult = isMultiple
	return
}

// this one is used by that parsing code when there are prefix lengths to be applied
func newPrefixedIPv4SectionParsed(segments []*AddressDivision, isMultiple bool, prefixLength PrefixLen, singleOnly bool) (res *IPv4AddressSection) {
	res = createIPv4Section(segments)
	res.isMult = isMultiple
	if prefixLength != nil {
		assignPrefix(prefixLength, segments, res.ToIP(), singleOnly, false, BitCount(len(segments)<<ipv4BitsToSegmentBitshift))
	}
	return
}

func NewIPv4Section(segments []*IPv4AddressSegment) *IPv4AddressSection {
	return createIPv4SectionFromSegs(segments, nil)
}

func NewIPv4PrefixedSection(segments []*IPv4AddressSegment, prefixLen PrefixLen) *IPv4AddressSection {
	return createIPv4SectionFromSegs(segments, prefixLen)
}

func createIPv4SectionFromSegs(orig []*IPv4AddressSegment, prefLen PrefixLen) (result *IPv4AddressSection) {
	divs, newPref, isMultiple := createDivisionsFromSegs(
		func(index int) *IPAddressSegment {
			return orig[index].ToIP()
		},
		len(orig),
		ipv4BitsToSegmentBitshift,
		IPv4BitsPerSegment,
		IPv4BytesPerSegment,
		IPv4MaxValuePerSegment,
		zeroIPv4Seg.ToIP(),
		zeroIPv4SegZeroPrefix.ToIP(),
		zeroIPv4SegPrefixBlock.ToIP(),
		prefLen)
	result = createIPv4Section(divs)
	result.prefixLength = newPref
	result.isMult = isMultiple
	return result
}

func NewIPv4SectionFromUint32(bytes uint32, segmentCount int) (res *IPv4AddressSection) {
	return NewIPv4SectionFromPrefixedUint32(bytes, segmentCount, nil)
}

func NewIPv4SectionFromPrefixedUint32(bytes uint32, segmentCount int, prefixLength PrefixLen) (res *IPv4AddressSection) {
	if segmentCount < 0 {
		segmentCount = IPv4SegmentCount
	}
	segments := createSegmentsUint64(
		segmentCount,
		0,
		uint64(bytes),
		IPv4BytesPerSegment,
		IPv4BitsPerSegment,
		ipv4Network.getIPAddressCreator(),
		prefixLength)
	res = createIPv4Section(segments)
	if prefixLength != nil {
		assignPrefix(prefixLength, segments, res.ToIP(), false, false, BitCount(segmentCount<<ipv4BitsToSegmentBitshift))
	}
	return
}

func NewIPv4SectionFromBytes(bytes []byte) (res *IPv4AddressSection, err addrerr.AddressValueError) {
	return newIPv4SectionFromBytes(bytes, len(bytes), nil, false)
}

// Useful if the byte array has leading zeros
func NewIPv4SectionFromSegmentedBytes(bytes []byte, segmentCount int) (res *IPv4AddressSection, err addrerr.AddressValueError) {
	return newIPv4SectionFromBytes(bytes, segmentCount, nil, false)
}

func NewIPv4SectionFromPrefixedBytes(bytes []byte, segmentCount int, prefixLength PrefixLen) (res *IPv4AddressSection, err addrerr.AddressValueError) {
	return newIPv4SectionFromBytes(bytes, segmentCount, prefixLength, false)
}

func newIPv4SectionFromBytes(bytes []byte, segmentCount int, prefixLength PrefixLen, singleOnly bool) (res *IPv4AddressSection, err addrerr.AddressValueError) {
	if segmentCount < 0 {
		segmentCount = len(bytes)
	}
	expectedByteCount := segmentCount
	segments, err := toSegments(
		bytes,
		segmentCount,
		IPv4BytesPerSegment,
		IPv4BitsPerSegment,
		ipv4Network.getIPAddressCreator(),
		prefixLength)
	if err == nil {
		res = createIPv4Section(segments)
		if prefixLength != nil {
			assignPrefix(prefixLength, segments, res.ToIP(), singleOnly, false, BitCount(segmentCount<<ipv4BitsToSegmentBitshift))
		}
		if expectedByteCount == len(bytes) && len(bytes) > 0 {
			bytes = cloneBytes(bytes)
			res.cache.bytesCache = &bytesCache{lowerBytes: bytes}
			if !res.isMult { // not a prefix block
				res.cache.bytesCache.upperBytes = bytes
			}
		}
	}
	return
}

func NewIPv4SectionFromVals(vals IPv4SegmentValueProvider, segmentCount int) (res *IPv4AddressSection) {
	res = NewIPv4SectionFromPrefixedRange(vals, nil, segmentCount, nil)
	return
}

func NewIPv4SectionFromPrefixedVals(vals IPv4SegmentValueProvider, segmentCount int, prefixLength PrefixLen) (res *IPv4AddressSection) {
	return NewIPv4SectionFromPrefixedRange(vals, nil, segmentCount, prefixLength)
}

func NewIPv4SectionFromRange(vals, upperVals IPv4SegmentValueProvider, segmentCount int) (res *IPv4AddressSection) {
	res = NewIPv4SectionFromPrefixedRange(vals, upperVals, segmentCount, nil)
	return
}

func NewIPv4SectionFromPrefixedRange(vals, upperVals IPv4SegmentValueProvider, segmentCount int, prefixLength PrefixLen) (res *IPv4AddressSection) {
	return newIPv4SectionFromPrefixedSingle(vals, upperVals, segmentCount, prefixLength, false)
}

func newIPv4SectionFromPrefixedSingle(vals, upperVals IPv4SegmentValueProvider, segmentCount int, prefixLength PrefixLen, singleOnly bool) (res *IPv4AddressSection) {
	if segmentCount < 0 {
		segmentCount = 0
	}
	segments, isMultiple := createSegments(
		WrappedIPv4SegmentValueProvider(vals),
		WrappedIPv4SegmentValueProvider(upperVals),
		segmentCount,
		IPv4BitsPerSegment,
		ipv4Network.getIPAddressCreator(),
		prefixLength)
	res = createIPv4Section(segments)
	res.isMult = isMultiple
	if prefixLength != nil {
		assignPrefix(prefixLength, segments, res.ToIP(), singleOnly, false, BitCount(segmentCount<<ipv4BitsToSegmentBitshift))
	}
	return
}

// IPv4AddressSection represents a section of an IPv4 address comprising 0 to 4 IPv4 address segments.
// The zero values is a section with zero segments.
type IPv4AddressSection struct {
	ipAddressSectionInternal
}

// Contains returns whether this is same type and version as the given address section and whether it contains all values in the given section.
//
// Sections must also have the same number of segments to be comparable, otherwise false is returned.
func (section *IPv4AddressSection) Contains(other AddressSectionType) bool {
	if section == nil {
		return other == nil || other.ToSectionBase() == nil
	}
	return section.contains(other)
}

// Equal returns whether the given address section is equal to this address section.
// Two address sections are equal if they represent the same set of sections.
// They must match:
//  - type/version: IPv4
//  - segment counts
//  - segment value ranges
// Prefix lengths are ignored.
func (section *IPv4AddressSection) Equal(other AddressSectionType) bool {
	if section == nil {
		return other == nil || other.ToSectionBase() == nil
	}
	return section.equal(other)
}

func (section *IPv4AddressSection) Compare(item AddressItem) int {
	return CountComparator.Compare(section, item)
}

// CompareSize compares the counts of two address sections, the number of individual sections represented.
//
// Rather than calculating counts with GetCount, there can be more efficient ways of comparing whether one section represents more individual address sections than another.
//
// CompareSize returns a positive integer if this address section has a larger count than the one given, 0 if they are the same, or a negative integer if the other has a larger count.
func (section *IPv4AddressSection) CompareSize(other StandardDivGroupingType) int {
	if section == nil {
		if other != nil && other.ToDivGrouping() != nil {
			// we have size 0, other has size >= 1
			return -1
		}
		return 0
	}
	return section.compareSize(other)
}

// GetBitsPerSegment returns the number of bits comprising each segment in this section.  Segments in the same address section are equal length.
func (section *IPv4AddressSection) GetBitsPerSegment() BitCount {
	return IPv4BitsPerSegment
}

// GetBytesPerSegment returns the number of bytes comprising each segment in this section.  Segments in the same address section are equal length.
func (section *IPv4AddressSection) GetBytesPerSegment() int {
	return IPv4BytesPerSegment
}

// GetIPVersion returns IPv4, the IP version of this address section
func (section *IPv4AddressSection) GetIPVersion() IPVersion {
	return IPv4
}

// GetCount returns the count of possible distinct values for this item.
// If not representing multiple values, the count is 1,
// unless this is a division grouping with no divisions, or an address section with no segments, in which case it is 0.
//
// Use IsMultiple if you simply want to know if the count is greater than 1.
func (section *IPv4AddressSection) GetCount() *big.Int {
	if section == nil {
		return bigZero()
	}
	return section.cacheCount(func() *big.Int {
		return bigZero().SetUint64(section.GetIPv4Count())
	})
}

// IsMultiple returns  whether this section represents multiple values
func (section *IPv4AddressSection) IsMultiple() bool {
	return section != nil && section.isMultiple()
}

// IsPrefixed returns whether this section has an associated prefix length
func (section *IPv4AddressSection) IsPrefixed() bool {
	return section != nil && section.isPrefixed()
}

// GetPrefixCount returns the number of distinct prefix values in this item.
//
// The prefix length is given by GetPrefixLen.
//
// If this has a non-nil prefix length, returns the number of distinct prefix values.
//
// If this has a nil prefix length, returns the same value as GetCount
func (section *IPv4AddressSection) GetPrefixCount() *big.Int {
	return section.cachePrefixCount(func() *big.Int {
		return bigZero().SetUint64(section.GetIPv4PrefixCount())
	})
}

// GetPrefixCountLen returns the number of distinct prefix values in this item for the given prefix length
func (section *IPv4AddressSection) GetPrefixCountLen(prefixLen BitCount) *big.Int {
	if prefixLen <= 0 {
		return bigOne()
	} else if bc := section.GetBitCount(); prefixLen > bc {
		prefixLen = bc
	}
	return section.calcCount(func() *big.Int { return new(big.Int).SetUint64(section.GetIPv4PrefixCountLen(prefixLen)) })
}

// GetBlockCount returns the count of distinct values in the given number of initial (more significant) segments.
func (section *IPv4AddressSection) GetBlockCount(segmentCount int) *big.Int {
	if segmentCount <= 0 {
		return bigOne()
	}
	return section.calcCount(func() *big.Int { return new(big.Int).SetUint64(section.GetIPv4BlockCount(segmentCount)) })
}

// GetIPv4PrefixCountLen gives count available as a uint64 instead of big.Int
func (section *IPv4AddressSection) GetIPv4PrefixCountLen(prefixLength BitCount) uint64 {
	if !section.isMultiple() {
		return 1
	} else if prefixLength >= section.GetBitCount() {
		return section.GetIPv4Count()
	} else if prefixLength < 0 {
		prefixLength = 0
	}
	return longPrefixCount(section.ToSectionBase(), prefixLength)
}

func (section *IPv4AddressSection) GetIPv4PrefixCount() uint64 {
	prefixLength := section.getPrefixLen()
	if prefixLength == nil {
		return section.GetIPv4Count()
	}
	return section.GetIPv4PrefixCountLen(prefixLength.bitCount())
}

func (section *IPv4AddressSection) GetIPv4Count() uint64 {
	if !section.isMultiple() {
		return 1
	}
	return longCount(section.ToSectionBase(), section.GetSegmentCount())
}

func (section *IPv4AddressSection) GetIPv4BlockCount(segmentCount int) uint64 {
	if !section.isMultiple() {
		return 1
	}
	return longCount(section.ToSectionBase(), segmentCount)
}

// GetSegment returns the segment at the given index.
// The first segment is at index 0.
// GetSegment will panic given a negative index or index larger than the segment count.
func (section *IPv4AddressSection) GetSegment(index int) *IPv4AddressSegment {
	return section.getDivision(index).ToIPv4()
}

// GetTrailingSection gets the subsection from the series starting from the given index.
// The first segment is at index 0.
func (section *IPv4AddressSection) GetTrailingSection(index int) *IPv4AddressSection {
	return section.GetSubSection(index, section.GetSegmentCount())
}

// GetSubSection gets the subsection from the series starting from the given index and ending just before the give endIndex.
// The first segment is at index 0.
func (section *IPv4AddressSection) GetSubSection(index, endIndex int) *IPv4AddressSection {
	return section.getSubSection(index, endIndex).ToIPv4()
}

func (section *IPv4AddressSection) GetNetworkSection() *IPv4AddressSection {
	return section.getNetworkSection().ToIPv4()
}

func (section *IPv4AddressSection) GetNetworkSectionLen(prefLen BitCount) *IPv4AddressSection {
	return section.getNetworkSectionLen(prefLen).ToIPv4()
}

func (section *IPv4AddressSection) GetHostSection() *IPv4AddressSection {
	return section.getHostSection().ToIPv4()
}

func (section *IPv4AddressSection) GetHostSectionLen(prefLen BitCount) *IPv4AddressSection {
	return section.getHostSectionLen(prefLen).ToIPv4()
}

func (section *IPv4AddressSection) GetNetworkMask() *IPv4AddressSection {
	return section.getNetworkMask(ipv4Network).ToIPv4()
}

func (section *IPv4AddressSection) GetHostMask() *IPv4AddressSection {
	return section.getHostMask(ipv4Network).ToIPv4()
}

// CopySubSegments copies the existing segments from the given start index until but not including the segment at the given end index,
// into the given slice, as much as can be fit into the slice, returning the number of segments copied
func (section *IPv4AddressSection) CopySubSegments(start, end int, segs []*IPv4AddressSegment) (count int) {
	return section.visitSubDivisions(start, end, func(index int, div *AddressDivision) bool { segs[index] = div.ToIPv4(); return false }, len(segs))
}

// CopySegments copies the existing segments into the given slice,
// as much as can be fit into the slice, returning the number of segments copied
func (section *IPv4AddressSection) CopySegments(segs []*IPv4AddressSegment) (count int) {
	return section.visitDivisions(func(index int, div *AddressDivision) bool { segs[index] = div.ToIPv4(); return false }, len(segs))
}

// GetSegments returns a slice with the address segments.  The returned slice is not backed by the same array as this section.
func (section *IPv4AddressSection) GetSegments() (res []*IPv4AddressSegment) {
	res = make([]*IPv4AddressSegment, section.GetSegmentCount())
	section.CopySegments(res)
	return
}

func (section *IPv4AddressSection) Mask(other *IPv4AddressSection) (res *IPv4AddressSection, err addrerr.IncompatibleAddressError) {
	return section.maskPrefixed(other, true)
}

func (section *IPv4AddressSection) maskPrefixed(other *IPv4AddressSection, retainPrefix bool) (res *IPv4AddressSection, err addrerr.IncompatibleAddressError) {
	sec, err := section.mask(other.ToIP(), retainPrefix)
	if err == nil {
		res = sec.ToIPv4()
	}
	return
}

func (section *IPv4AddressSection) BitwiseOr(other *IPv4AddressSection) (res *IPv4AddressSection, err addrerr.IncompatibleAddressError) {
	return section.bitwiseOrPrefixed(other, true)
}

func (section *IPv4AddressSection) bitwiseOrPrefixed(other *IPv4AddressSection, retainPrefix bool) (res *IPv4AddressSection, err addrerr.IncompatibleAddressError) {
	sec, err := section.bitwiseOr(other.ToIP(), retainPrefix)
	if err == nil {
		res = sec.ToIPv4()
	}
	return
}

func (section *IPv4AddressSection) MatchesWithMask(other *IPv4AddressSection, mask *IPv4AddressSection) bool {
	return section.matchesWithMask(other.ToIP(), mask.ToIP())
}

func (section *IPv4AddressSection) Subtract(other *IPv4AddressSection) (res []*IPv4AddressSection, err addrerr.SizeMismatchError) {
	sections, err := section.subtract(other.ToIP())
	if err == nil {
		res = cloneIPSectsToIPv4Sects(sections)
	}
	return
}

func (section *IPv4AddressSection) Intersect(other *IPv4AddressSection) (res *IPv4AddressSection, err addrerr.SizeMismatchError) {
	sec, err := section.intersect(other.ToIP())
	if err == nil {
		res = sec.ToIPv4()
	}
	return
}

// GetLower returns the section in the range with the lowest numeric value,
// which will be the same section if it represents a single value.
// For example, for "1.2-3.4.5-6", the section "1.2.4.5" is returned.
func (section *IPv4AddressSection) GetLower() *IPv4AddressSection {
	return section.getLower().ToIPv4()
}

// GetUpper returns the section in the range with the highest numeric value,
// which will be the same section if it represents a single value.
// For example, for "1.2-3.4.5-6", the section "1.3.4.6" is returned.
func (section *IPv4AddressSection) GetUpper() *IPv4AddressSection {
	return section.getUpper().ToIPv4()
}

func (section *IPv4AddressSection) ToZeroHost() (*IPv4AddressSection, addrerr.IncompatibleAddressError) {
	res, err := section.toZeroHost(false)
	return res.ToIPv4(), err
}

func (section *IPv4AddressSection) ToZeroHostLen(prefixLength BitCount) (*IPv4AddressSection, addrerr.IncompatibleAddressError) {
	res, err := section.toZeroHostLen(prefixLength)
	return res.ToIPv4(), err
}

func (section *IPv4AddressSection) ToZeroNetwork() *IPv4AddressSection {
	return section.toZeroNetwork().ToIPv4()
}

func (section *IPv4AddressSection) ToMaxHost() (*IPv4AddressSection, addrerr.IncompatibleAddressError) {
	res, err := section.toMaxHost()
	return res.ToIPv4(), err
}

func (section *IPv4AddressSection) ToMaxHostLen(prefixLength BitCount) (*IPv4AddressSection, addrerr.IncompatibleAddressError) {
	res, err := section.toMaxHostLen(prefixLength)
	return res.ToIPv4(), err
}

func (section *IPv4AddressSection) Uint32Value() uint32 {
	lower, _ := section.getIntValues()
	return lower
}

func (section *IPv4AddressSection) UpperUint32Value() uint32 {
	_, upper := section.getIntValues()
	return upper
}

func (section *IPv4AddressSection) getIntValues() (lower, upper uint32) {
	segCount := section.GetSegmentCount()
	if segCount == 0 {
		return 0, 0
	}
	cache := section.cache
	if cache == nil {
		return section.calcIntValues()
	}
	cached := cache.intsCache
	if cached == nil {
		cached = &intsCache{}
		cached.cachedLowerVal, cached.cachedUpperVal = section.calcIntValues()
		dataLoc := (*unsafe.Pointer)(unsafe.Pointer(&cache.intsCache))
		atomic.StorePointer(dataLoc, unsafe.Pointer(cached))
	}
	lower = cached.cachedLowerVal
	upper = cached.cachedUpperVal
	return
}

func (section *IPv4AddressSection) calcIntValues() (lower, upper uint32) {
	segCount := section.GetSegmentCount()
	isMult := section.isMultiple()
	if segCount == 4 {
		lower = (uint32(section.GetSegment(0).GetSegmentValue()) << 24) |
			(uint32(section.GetSegment(1).GetSegmentValue()) << 16) |
			(uint32(section.GetSegment(2).GetSegmentValue()) << 8) |
			uint32(section.GetSegment(3).GetSegmentValue())
		if isMult {
			upper = (uint32(section.GetSegment(0).GetUpperSegmentValue()) << 24) |
				(uint32(section.GetSegment(1).GetUpperSegmentValue()) << 16) |
				(uint32(section.GetSegment(2).GetUpperSegmentValue()) << 8) |
				uint32(section.GetSegment(3).GetUpperSegmentValue())
		} else {
			upper = lower
		}
		return
	}
	seg := section.GetSegment(0)
	lower = uint32(seg.GetSegmentValue())
	if isMult {
		upper = uint32(seg.GetUpperSegmentValue())
	}
	bitsPerSegment := section.GetBitsPerSegment()
	for i := 1; i < segCount; i++ {
		seg = section.GetSegment(i)
		lower = (lower << uint(bitsPerSegment)) | uint32(seg.GetSegmentValue())
		if isMult {
			upper = (upper << uint(bitsPerSegment)) | uint32(seg.GetUpperSegmentValue())
		}
	}
	if !isMult {
		upper = lower
	}
	return
}

// ToPrefixBlock returns the section with the same prefix as this section while the remaining bits span all values.
// The returned section will be the block of all sections with the same prefix.
//
// If this section has no prefix, this section is returned.
func (section *IPv4AddressSection) ToPrefixBlock() *IPv4AddressSection {
	return section.toPrefixBlock().ToIPv4()
}

// ToPrefixBlockLen returns the section with the same prefix of the given length as this section while the remaining bits span all values.
// The returned section will be the block of all sections with the same prefix.
func (section *IPv4AddressSection) ToPrefixBlockLen(prefLen BitCount) *IPv4AddressSection {
	return section.toPrefixBlockLen(prefLen).ToIPv4()
}

// ToBlock creates a new block of address sections by changing the segment at the given index to have the given lower and upper value,
// and changing the following segments to be full-range.
func (section *IPv4AddressSection) ToBlock(segmentIndex int, lower, upper SegInt) *IPv4AddressSection {
	return section.toBlock(segmentIndex, lower, upper).ToIPv4()
}

// WithoutPrefixLen provides the same address section but with no prefix length.  The values remain unchanged.
func (section *IPv4AddressSection) WithoutPrefixLen() *IPv4AddressSection {
	if !section.IsPrefixed() {
		return section
	}
	return section.withoutPrefixLen().ToIPv4()
}

// SetPrefixLen sets the prefix length.
//
// A prefix length will not be set to a value lower than zero or beyond the bit length of the address section.
// The provided prefix length will be adjusted to these boundaries if necessary.
func (section *IPv4AddressSection) SetPrefixLen(prefixLen BitCount) *IPv4AddressSection {
	return section.setPrefixLen(prefixLen).ToIPv4()
}

// SetPrefixLenZeroed sets the prefix length.
//
// A prefix length will not be set to a value lower than zero or beyond the bit length of the address section.
// The provided prefix length will be adjusted to these boundaries if necessary.
//
// If this address section has a prefix length, and the prefix length is increased when setting the new prefix length, the bits moved within the prefix become zero.
// If this address section has a prefix length, and the prefix length is decreased when setting the new prefix length, the bits moved outside the prefix become zero.
//
// In other words, bits that move from one side of the prefix length to the other (ie bits moved into the prefix or outside the prefix) are zeroed.
//
// If the result cannot be zeroed because zeroing out bits results in a non-contiguous segment, an error is returned.
func (section *IPv4AddressSection) SetPrefixLenZeroed(prefixLen BitCount) (*IPv4AddressSection, addrerr.IncompatibleAddressError) {
	res, err := section.setPrefixLenZeroed(prefixLen)
	return res.ToIPv4(), err
}

// AdjustPrefixLen increases or decreases the prefix length by the given increment.
//
// A prefix length will not be adjusted lower than zero or beyond the bit length of the address section.
//
// If this address section has no prefix length, then the prefix length will be set to the adjustment if positive,
// or it will be set to the adjustment added to the bit count if negative.
func (section *IPv4AddressSection) AdjustPrefixLen(prefixLen BitCount) *IPv4AddressSection {
	return section.adjustPrefixLen(prefixLen).ToIPv4()
}

// AdjustPrefixLenZeroed increases or decreases the prefix length by the given increment while zeroing out the bits that have moved into or outside the prefix.
//
// A prefix length will not be adjusted lower than zero or beyond the bit length of the address section.
//
// If this address section has no prefix length, then the prefix length will be set to the adjustment if positive,
// or it will be set to the adjustment added to the bit count if negative.
//
// When prefix length is increased, the bits moved within the prefix become zero.
// When a prefix length is decreased, the bits moved outside the prefix become zero.
//
// If the result cannot be zeroed because zeroing out bits results in a non-contiguous segment, an error is returned.
func (section *IPv4AddressSection) AdjustPrefixLenZeroed(prefixLen BitCount) (*IPv4AddressSection, addrerr.IncompatibleAddressError) {
	res, err := section.adjustPrefixLenZeroed(prefixLen)
	return res.ToIPv4(), err
}

// AssignPrefixForSingleBlock returns the equivalent prefix block that matches exactly the range of values in this address section.
// The returned block will have an assigned prefix length indicating the prefix length for the block.
//
// There may be no such address section - it is required that the range of values match the range of a prefix block.
// If there is no such address section, then nil is returned.
func (section *IPv4AddressSection) AssignPrefixForSingleBlock() *IPv4AddressSection {
	return section.assignPrefixForSingleBlock().ToIPv4()
}

// AssignMinPrefixForBlock returns an equivalent address section, assigned the smallest prefix length possible,
// such that the prefix block for that prefix length is in this address section.
//
// In other words, this method assigns a prefix length to this address section matching the largest prefix block in this address section.
func (section *IPv4AddressSection) AssignMinPrefixForBlock() *IPv4AddressSection {
	return section.assignMinPrefixForBlock().ToIPv4()
}

// Iterator provides an iterator to iterate through the individual address sections of this address section.
//
// When iterating, the prefix length is preserved.  Remove it using WithoutPrefixLen prior to iterating if you wish to drop it from all individual address sections.
//
// Call IsMultiple to determine if this instance represents multiple address sections, or GetCount for the count.
func (section *IPv4AddressSection) Iterator() IPv4SectionIterator {
	if section == nil {
		return ipv4SectionIterator{nilSectIterator()}
	}
	return ipv4SectionIterator{section.sectionIterator(nil)}
}

// PrefixIterator provides an iterator to iterate through the individual prefixes of this address section,
// each iterated element spanning the range of values for its prefix.
//
// It is similar to the prefix block iterator, except for possibly the first and last iterated elements, which might not be prefix blocks,
// instead constraining themselves to values from this address section.
//
// If the series has no prefix length, then this is equivalent to Iterator.
func (section *IPv4AddressSection) PrefixIterator() IPv4SectionIterator {
	return ipv4SectionIterator{section.prefixIterator(false)}
}

// PrefixBlockIterator provides an iterator to iterate through the individual prefix blocks, one for each prefix of this address section.
// Each iterated address section will be a prefix block with the same prefix length as this address section.
//
// If this address section has no prefix length, then this is equivalent to Iterator.
func (section *IPv4AddressSection) PrefixBlockIterator() IPv4SectionIterator {
	return ipv4SectionIterator{section.prefixIterator(true)}
}

// BlockIterator Iterates through the address sections that can be obtained by iterating through all the upper segments up to the given segment count.
// The segments following remain the same in all iterated sections.
func (section *IPv4AddressSection) BlockIterator(segmentCount int) IPv4SectionIterator {
	return ipv4SectionIterator{section.blockIterator(segmentCount)}
}

// SequentialBlockIterator iterates through the sequential address sections that make up this address section.
//
// Practically, this means finding the count of segments for which the segments that follow are not full range, and then using BlockIterator with that segment count.
//
// Use GetSequentialBlockCount to get the number of iterated elements.
func (section *IPv4AddressSection) SequentialBlockIterator() IPv4SectionIterator {
	return ipv4SectionIterator{section.sequentialBlockIterator()}
}

func (section *IPv4AddressSection) ToDivGrouping() *AddressDivisionGrouping {
	return section.ToSectionBase().ToDivGrouping()
}

func (section *IPv4AddressSection) ToSectionBase() *AddressSection {
	return section.ToIP().ToSectionBase()
}

// ToIP converts to an IPAddressSection, a polymorphic type usable with all IP address sections.
//
// ToIP can be called with a nil receiver, enabling you to chain this method with methods that might return a nil pointer.
func (section *IPv4AddressSection) ToIP() *IPAddressSection {
	return (*IPAddressSection)(section)
}

// IncrementBoundary returns the item that is the given increment from the range boundaries of this item.
//
// If the given increment is positive, adds the value to the highest (GetUpper) in the range to produce a new item.
// If the given increment is negative, adds the value to the lowest (GetLower) in the range to produce a new item.
// If the increment is zero, returns this.
//
// If this represents just a single value, this item is simply incremented by the given increment value, positive or negative.
//
// On overflow or underflow, IncrementBoundary returns nil.
func (section *IPv4AddressSection) IncrementBoundary(increment int64) *IPv4AddressSection {
	return section.incrementBoundary(increment).ToIPv4()
}

func getIPv4MaxValueLong(segmentCount int) uint64 {
	return macMaxValues[segmentCount]
}

// Increment returns the item that is the given increment upwards into the range,
// with the increment of 0 returning the first in the range.
//
// If the increment i matches or exceeds the range count c, then i - c + 1
// is added to the upper item of the range.
// An increment matching the count gives you the item just above the highest in the range.
//
// If the increment is negative, it is added to the lowest of the range.
// To get the item just below the lowest of the range, use the increment -1.
//
// If this represents just a single value, the item is simply incremented by the given increment, positive or negative.
//
// If this item represents multiple values, a positive increment i is equivalent i + 1 values from the iterator and beyond.
// For instance, a increment of 0 is the first value from the iterator, an increment of 1 is the second value from the iterator, and so on.
// An increment of a negative value added to the count is equivalent to the same number of iterator values preceding the last value of the iterator.
// For instance, an increment of count - 1 is the last value from the iterator, an increment of count - 2 is the second last value, and so on.
//
// On overflow or underflow, Increment returns nil.
func (section *IPv4AddressSection) Increment(inc int64) *IPv4AddressSection {
	if inc == 0 && !section.isMultiple() {
		return section
	}
	lowerValue := uint64(section.Uint32Value())
	upperValue := uint64(section.UpperUint32Value())
	count := section.GetIPv4Count()
	isOverflow := checkOverflow(inc, lowerValue, upperValue, count-1, getIPv4MaxValueLong(section.GetSegmentCount()))
	if isOverflow {
		return nil
	}
	return increment(
		section.ToSectionBase(),
		inc,
		ipv4Network.getIPAddressCreator(),
		count-1,
		lowerValue,
		upperValue,
		section.getLower,
		section.getUpper,
		section.getPrefixLen()).ToIPv4()
}

func (section *IPv4AddressSection) SpanWithPrefixBlocks() []*IPv4AddressSection {
	if section.IsSequential() {
		if section.IsSinglePrefixBlock() {
			return []*IPv4AddressSection{section}
		}
		wrapped := wrapIPSection(section.ToIP())
		spanning := getSpanningPrefixBlocks(wrapped, wrapped)
		return cloneToIPv4Sections(spanning)
	}
	wrapped := wrapIPSection(section.ToIP())
	return cloneToIPv4Sections(spanWithPrefixBlocks(wrapped))
}

func (section *IPv4AddressSection) SpanWithPrefixBlocksTo(other *IPv4AddressSection) ([]*IPv4AddressSection, addrerr.SizeMismatchError) {
	if err := section.checkSectionCount(other.ToIP()); err != nil {
		return nil, err
	}
	return cloneToIPv4Sections(
		getSpanningPrefixBlocks(
			wrapIPSection(section.ToIP()),
			wrapIPSection(other.ToIP()),
		),
	), nil
}

func (section *IPv4AddressSection) SpanWithSequentialBlocks() []*IPv4AddressSection {
	if section.IsSequential() {
		return []*IPv4AddressSection{section}
	}
	wrapped := wrapIPSection(section.ToIP())
	return cloneToIPv4Sections(spanWithSequentialBlocks(wrapped))
}

func (section *IPv4AddressSection) SpanWithSequentialBlocksTo(other *IPv4AddressSection) ([]*IPv4AddressSection, addrerr.SizeMismatchError) {
	if err := section.checkSectionCount(other.ToIP()); err != nil {
		return nil, err
	}
	return cloneToIPv4Sections(
		getSpanningSequentialBlocks(
			wrapIPSection(section.ToIP()),
			wrapIPSection(other.ToIP()),
		),
	), nil
}

func (section *IPv4AddressSection) CoverWithPrefixBlockTo(other *IPv4AddressSection) (*IPv4AddressSection, addrerr.SizeMismatchError) {
	res, err := section.coverWithPrefixBlockTo(other.ToIP())
	return res.ToIPv4(), err
}

func (section *IPv4AddressSection) CoverWithPrefixBlock() *IPv4AddressSection {
	return section.coverWithPrefixBlock().ToIPv4()
}

func (section *IPv4AddressSection) checkSectionCounts(sections []*IPv4AddressSection) addrerr.SizeMismatchError {
	segCount := section.GetSegmentCount()
	length := len(sections)
	for i := 0; i < length; i++ {
		section2 := sections[i]
		if section2 == nil {
			continue
		}
		if section2.GetSegmentCount() != segCount {
			return &sizeMismatchError{incompatibleAddressError{addressError{key: "ipaddress.error.sizeMismatch"}}}
		}
	}
	return nil
}

//
// MergeToSequentialBlocks merges this with the list of sections to produce the smallest array of blocks that are sequential
//
// The resulting array is sorted from lowest address value to highest, regardless of the size of each prefix block.
func (section *IPv4AddressSection) MergeToSequentialBlocks(sections ...*IPv4AddressSection) ([]*IPv4AddressSection, addrerr.SizeMismatchError) {
	if err := section.checkSectionCounts(sections); err != nil {
		return nil, err
	}
	series := cloneIPv4Sections(section, sections)
	blocks := getMergedSequentialBlocks(series)
	return cloneToIPv4Sections(blocks), nil
}

//
// MergeToPrefixBlocks merges this with the list of sections to produce the smallest array of prefix blocks.
//
// The resulting array is sorted from lowest address value to highest, regardless of the size of each prefix block.
func (section *IPv4AddressSection) MergeToPrefixBlocks(sections ...*IPv4AddressSection) ([]*IPv4AddressSection, addrerr.SizeMismatchError) {
	if err := section.checkSectionCounts(sections); err != nil {
		return nil, err
	}
	series := cloneIPv4Sections(section, sections)
	blocks := getMergedPrefixBlocks(series)
	return cloneToIPv4Sections(blocks), nil
}

// ReverseBits returns a new section with the bits reversed.  Any prefix length is dropped.
//
// If the bits within a single segment cannot be reversed because the segment represents a range,
// and reversing the segment values results in a range that is not contiguous, this returns an error.
//
// In practice this means that to be reversible, a range must include all values except possibly the largest and/or smallest, which reverse to themselves.
//
// If perByte is true, the bits are reversed within each byte, otherwise all the bits are reversed.
func (section *IPv4AddressSection) ReverseBits(perByte bool) (*IPv4AddressSection, addrerr.IncompatibleAddressError) {
	res, err := section.reverseBits(perByte)
	return res.ToIPv4(), err
}

// ReverseBytes returns a new section with the bytes reversed.  Any prefix length is dropped.
func (section *IPv4AddressSection) ReverseBytes() *IPv4AddressSection {
	return section.ReverseSegments()
}

// ReverseSegments returns a new section with the segments reversed.
func (section *IPv4AddressSection) ReverseSegments() *IPv4AddressSection {
	if section.GetSegmentCount() <= 1 {
		if section.IsPrefixed() {
			return section.WithoutPrefixLen()
		}
		return section
	}
	res, _ := section.reverseSegments(
		func(i int) (*AddressSegment, addrerr.IncompatibleAddressError) {
			return section.GetSegment(i).WithoutPrefixLen().ToSegmentBase(), nil
		},
	)
	return res.ToIPv4()
}

func (section *IPv4AddressSection) Append(other *IPv4AddressSection) *IPv4AddressSection {
	count := section.GetSegmentCount()
	return section.ReplaceLen(count, count, other, 0, other.GetSegmentCount())
}

func (section *IPv4AddressSection) Insert(index int, other *IPv4AddressSection) *IPv4AddressSection {
	return section.insert(index, other.ToIP(), ipv4BitsToSegmentBitshift).ToIPv4()
}

// Replace replaces the segments of this section starting at the given index with the given replacement segments
func (section *IPv4AddressSection) Replace(index int, replacement *IPv4AddressSection) *IPv4AddressSection {
	return section.ReplaceLen(index, index+replacement.GetSegmentCount(), replacement, 0, replacement.GetSegmentCount())
}

// ReplaceLen replaces segments starting from startIndex and ending before endIndex with the segments starting at replacementStartIndex and
//ending before replacementEndIndex from the replacement section
func (section *IPv4AddressSection) ReplaceLen(startIndex, endIndex int, replacement *IPv4AddressSection, replacementStartIndex, replacementEndIndex int) *IPv4AddressSection {
	return section.replaceLen(startIndex, endIndex, replacement.ToIP(), replacementStartIndex, replacementEndIndex, ipv4BitsToSegmentBitshift).ToIPv4()
}

// IsAdaptiveZero returns true if the section was originally created as a zero-valued section (eg IPv4AddressSection{}),
// meaning it was not constructed using a constructor function.
// Such a grouping, which has no divisions or segments, is convertible to a zero-valued grouping of any type or version, whether IPv6, IPv4, MAC, etc
// It is not considered equal to constructions of specific zero length sections or groupings like NewIPv4Section(nil) which can only represent a zero-length section of a single address type.
func (section *IPv4AddressSection) IsAdaptiveZero() bool {
	return section != nil && section.matchesZeroGrouping()
}

var (
	ipv4CanonicalParams          = new(addrstr.IPv4StringOptionsBuilder).ToOptions()
	ipv4FullParams               = new(addrstr.IPv4StringOptionsBuilder).SetExpandedSegments(true).SetWildcardOptions(wildcardsRangeOnlyNetworkOnly).ToOptions()
	ipv4NormalizedWildcardParams = new(addrstr.IPv4StringOptionsBuilder).SetWildcardOptions(allWildcards).ToOptions()
	ipv4SqlWildcardParams        = new(addrstr.IPv4StringOptionsBuilder).SetWildcardOptions(allSQLWildcards).ToOptions()

	inetAtonOctalParams       = new(addrstr.IPv4StringOptionsBuilder).SetRadix(Inet_aton_radix_octal.GetRadix()).SetSegmentStrPrefix(Inet_aton_radix_octal.GetSegmentStrPrefix()).ToOptions()
	inetAtonHexParams         = new(addrstr.IPv4StringOptionsBuilder).SetRadix(Inet_aton_radix_hex.GetRadix()).SetSegmentStrPrefix(Inet_aton_radix_hex.GetSegmentStrPrefix()).ToOptions()
	ipv4ReverseDNSParams      = new(addrstr.IPv4StringOptionsBuilder).SetWildcardOptions(allWildcards).SetReverse(true).SetAddressSuffix(IPv4ReverseDnsSuffix).ToOptions()
	ipv4SegmentedBinaryParams = new(addrstr.IPStringOptionsBuilder).SetRadix(2).SetSeparator(IPv4SegmentSeparator).SetSegmentStrPrefix(BinaryPrefix).ToOptions()
)

func (section *IPv4AddressSection) ToHexString(with0xPrefix bool) (string, addrerr.IncompatibleAddressError) {
	if section == nil {
		return nilString(), nil
	}
	return section.toHexString(with0xPrefix)
}

func (section *IPv4AddressSection) ToOctalString(with0Prefix bool) (string, addrerr.IncompatibleAddressError) {
	if section == nil {
		return nilString(), nil
	}
	return section.toOctalString(with0Prefix)
}

func (section *IPv4AddressSection) ToBinaryString(with0bPrefix bool) (string, addrerr.IncompatibleAddressError) {
	if section == nil {
		return nilString(), nil
	}
	return section.toBinaryString(with0bPrefix)
}

// ToCanonicalString produces a canonical string.
//
//If this section has a prefix length, it will be included in the string.
func (section *IPv4AddressSection) ToCanonicalString() string {
	if section == nil {
		return nilString()
	}
	cache := section.getStringCache()
	if cache == nil {
		return section.toNormalizedString(ipv4CanonicalParams)
	}
	return cacheStr(&cache.canonicalString,
		func() string {
			return section.toNormalizedString(ipv4CanonicalParams)
		})
}

// ToNormalizedString produces a normalized string.
//
//If this section has a prefix length, it will be included in the string.
func (section *IPv4AddressSection) ToNormalizedString() string {
	if section == nil {
		return nilString()
	}
	return section.ToCanonicalString()
}

func (section *IPv4AddressSection) ToCompressedString() string {
	if section == nil {
		return nilString()
	}
	return section.ToCanonicalString()
}

func (section *IPv4AddressSection) ToNormalizedWildcardString() string {
	if section == nil {
		return nilString()
	}
	cache := section.getStringCache()
	if cache == nil {
		return section.toNormalizedString(ipv4NormalizedWildcardParams)
	}
	return cacheStr(&cache.normalizedWildcardString,
		func() string {
			return section.toNormalizedString(ipv4NormalizedWildcardParams)
		})
}

func (section *IPv4AddressSection) ToCanonicalWildcardString() string {
	if section == nil {
		return nilString()
	}
	return section.ToNormalizedWildcardString()
}

func (section *IPv4AddressSection) ToSegmentedBinaryString() string {
	if section == nil {
		return nilString()
	}
	cache := section.getStringCache()
	if cache == nil {
		return section.toNormalizedString(ipv4SegmentedBinaryParams)
	}
	return cacheStr(&cache.segmentedBinaryString,
		func() string {
			return section.toNormalizedString(ipv4SegmentedBinaryParams)
		})
}

func (section *IPv4AddressSection) ToSQLWildcardString() string {
	if section == nil {
		return nilString()
	}
	cache := section.getStringCache()
	if cache == nil {
		return section.toNormalizedString(ipv4SqlWildcardParams)
	}
	return cacheStr(&cache.sqlWildcardString,
		func() string {
			return section.toNormalizedString(ipv4SqlWildcardParams)
		})
}

func (section *IPv4AddressSection) ToFullString() string {
	if section == nil {
		return nilString()
	}
	cache := section.getStringCache()
	if cache == nil {
		return section.toNormalizedString(ipv4FullParams)
	}
	return cacheStr(&cache.fullString,
		func() string {
			return section.toNormalizedString(ipv4FullParams)
		})
}

// ToReverseDNSString returns the reverse DNS string.
// The method helps implement the IPAddressSegmentSeries interface.  For IPV4, the error is always nil.
func (section *IPv4AddressSection) ToReverseDNSString() (string, addrerr.IncompatibleAddressError) {
	if section == nil {
		return nilString(), nil
	}
	cache := section.getStringCache()
	if cache == nil {
		return section.toNormalizedString(ipv4ReverseDNSParams), nil
	}
	return cacheStr(&cache.reverseDNSString,
		func() string {
			return section.toNormalizedString(ipv4ReverseDNSParams)
		}), nil
}

func (section *IPv4AddressSection) ToPrefixLenString() string {
	if section == nil {
		return nilString()
	}
	return section.ToCanonicalString()
}

func (section *IPv4AddressSection) ToSubnetString() string {
	if section == nil {
		return nilString()
	}
	return section.ToNormalizedWildcardString()
}

func (section *IPv4AddressSection) ToCompressedWildcardString() string {
	if section == nil {
		return nilString()
	}
	return section.ToNormalizedWildcardString()
}

func (section *IPv4AddressSection) ToInetAtonString(radix Inet_aton_radix) string {
	if section == nil {
		return nilString()
	}
	cache := section.getStringCache()
	if radix == Inet_aton_radix_octal {
		if cache == nil {
			return section.toNormalizedString(inetAtonOctalParams)
		}
		return cacheStr(&cache.inetAtonOctalString,
			func() string {
				return section.toNormalizedString(inetAtonOctalParams)
			})
	} else if radix == Inet_aton_radix_hex {
		if cache == nil {
			return section.toNormalizedString(inetAtonHexParams)
		}
		return cacheStr(&cache.inetAtonHexString,
			func() string {
				return section.toNormalizedString(inetAtonHexParams)
			})
	} else {
		return section.ToCanonicalString()
	}
}

func (section *IPv4AddressSection) ToInetAtonJoinedString(radix Inet_aton_radix, joinedCount int) (string, addrerr.IncompatibleAddressError) {
	if section == nil {
		return nilString(), nil
	}
	if joinedCount <= 0 {
		return section.ToInetAtonString(radix), nil
	}
	var stringParams addrstr.IPStringOptions
	if radix == Inet_aton_radix_octal {
		stringParams = inetAtonOctalParams
	} else if radix == Inet_aton_radix_hex {
		stringParams = inetAtonHexParams
	} else {
		stringParams = ipv4CanonicalParams
	}
	return section.ToNormalizedJoinedString(stringParams, joinedCount)
}

func (section *IPv4AddressSection) ToNormalizedJoinedString(stringParams addrstr.IPStringOptions, joinedCount int) (string, addrerr.IncompatibleAddressError) {
	if section == nil {
		return nilString(), nil
	}
	if joinedCount <= 0 || section.GetSegmentCount() <= 1 {
		return section.toNormalizedString(stringParams), nil
	}
	equivalentPart, err := section.ToJoinedSegments(joinedCount) // AddressDivisionSeries
	if err != nil {
		return "", err
	}
	return toNormalizedIPString(stringParams, equivalentPart), nil
}

func (section *IPv4AddressSection) ToJoinedSegments(joinCount int) (AddressDivisionSeries, addrerr.IncompatibleAddressError) {
	thisCount := section.GetSegmentCount()
	if joinCount <= 0 || thisCount <= 1 {
		return section, nil
	}
	var totalCount int
	if joinCount >= thisCount {
		joinCount = thisCount - 1
		totalCount = 1
	} else {
		totalCount = thisCount - joinCount
	}
	joinedSegment, err := section.joinSegments(joinCount) //IPv4JoinedSegments
	if err != nil {
		return nil, err
	}
	notJoinedCount := totalCount - 1
	segs := make([]*AddressDivision, totalCount)
	section.copySubSegmentsToSlice(0, notJoinedCount, segs)
	segs[notJoinedCount] = joinedSegment
	equivalentPart := createInitializedGrouping(segs, section.getPrefixLen())
	return equivalentPart, nil
}

func (section *IPv4AddressSection) joinSegments(joinCount int) (*AddressDivision, addrerr.IncompatibleAddressError) {
	var lower, upper DivInt
	var prefix PrefixLen
	var networkPrefixLength BitCount

	var firstRange *IPv4AddressSegment
	firstJoinedIndex := section.GetSegmentCount() - 1 - joinCount
	bitsPerSeg := section.GetBitsPerSegment()
	for j := 0; j <= joinCount; j++ {
		thisSeg := section.GetSegment(firstJoinedIndex + j)
		if firstRange != nil {
			if !thisSeg.IsFullRange() {
				return nil, &incompatibleAddressError{addressError{key: "ipaddress.error.invalidMixedRange"}}
			}
		} else if thisSeg.isMultiple() {
			firstRange = thisSeg
		}
		lower = (lower << uint(bitsPerSeg)) | DivInt(thisSeg.getSegmentValue())
		upper = (upper << uint(bitsPerSeg)) | DivInt(thisSeg.getUpperSegmentValue())
		if prefix == nil {
			thisSegPrefix := thisSeg.getDivisionPrefixLength()
			if thisSegPrefix != nil {
				prefix = cacheBitCount(networkPrefixLength + thisSegPrefix.bitCount())
			} else {
				networkPrefixLength += thisSeg.getBitCount()
			}
		}
	}
	return NewRangePrefixDivision(lower, upper, prefix, (BitCount(joinCount)+1)<<3), nil
}

func (section *IPv4AddressSection) toNormalizedString(stringOptions addrstr.IPStringOptions) string {
	return toNormalizedIPString(stringOptions, section)
}

// String implements the fmt.Stringer interface, returning the normalized string provided by ToNormalizedString, or "<nil>" if the receiver is a nil pointer
func (section *IPv4AddressSection) String() string {
	if section == nil {
		return nilString()
	}
	return section.toString()
}

func (section *IPv4AddressSection) GetSegmentStrings() []string {
	if section == nil {
		return nil
	}
	return section.getSegmentStrings()
}

type Inet_aton_radix int

func (rad Inet_aton_radix) GetRadix() int {
	return int(rad)
}

func (rad Inet_aton_radix) GetSegmentStrPrefix() string {
	if rad == Inet_aton_radix_octal {
		return OctalPrefix
	} else if rad == Inet_aton_radix_hex {
		return HexPrefix
	}
	return ""
}

func (rad Inet_aton_radix) String() string {
	if rad == Inet_aton_radix_octal {
		return "octal"
	} else if rad == Inet_aton_radix_hex {
		return "hexadecimal"
	}
	return "decimal"
}

const (
	Inet_aton_radix_octal   Inet_aton_radix = 8
	Inet_aton_radix_hex     Inet_aton_radix = 16
	Inet_aton_radix_decimal Inet_aton_radix = 10
)
