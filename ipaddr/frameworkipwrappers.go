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
	"github.com/seancfoley/ipaddress-go/ipaddr/addrerr"
	"github.com/seancfoley/ipaddress-go/ipaddr/addrstr"
)

// ExtendedIPSegmentSeries wraps either an IPAddress or IPAddressSection.
// ExtendedIPSegmentSeries can be used to write code that works with both IP Addresses and IP Address Sections,
// going further than IPAddressSegmentSeries to offer additional methods, methods with the series types in their signature.
type ExtendedIPSegmentSeries interface {
	IPAddressSegmentSeries

	// Unwrap returns the wrapped *IPAddress or *IPAddressSection as an interface, IPAddressSegmentSeries
	Unwrap() IPAddressSegmentSeries

	// Equal returns whether the given address series is equal to this address series.
	// Two address series are equal if they represent the same set of series.
	// Both must be equal addresses or both must be equal sections.
	Equal(ExtendedIPSegmentSeries) bool

	// Contains returns whether this is same type and version as the given address series and whether it contains all values in the given series.
	//
	// Series must also have the same number of segments to be comparable, otherwise false is returned.
	Contains(ExtendedIPSegmentSeries) bool

	// CompareSize compares the counts of two address series, the number of individual series represented in each.
	//
	// Rather than calculating counts with GetCount, there can be more efficient ways of comparing whether one series represents more individual address series than another.
	//
	// CompareSize returns a positive integer if this address series has a larger count than the one given, 0 if they are the same, or a negative integer if the other has a larger count.
	CompareSize(ExtendedIPSegmentSeries) int

	// GetSection returns the backing section for this series, comprising all segments.
	GetSection() *IPAddressSection

	// GetTrailingSection returns an ending subsection of the full address section
	GetTrailingSection(index int) *IPAddressSection

	// GetSubSection returns a subsection of the full address section
	GetSubSection(index, endIndex int) *IPAddressSection

	GetNetworkSection() *IPAddressSection
	GetHostSection() *IPAddressSection
	GetNetworkSectionLen(BitCount) *IPAddressSection
	GetHostSectionLen(BitCount) *IPAddressSection

	GetNetworkMask() ExtendedIPSegmentSeries
	GetHostMask() ExtendedIPSegmentSeries

	// GetSegment returns the segment at the given index.
	// The first segment is at index 0.
	// GetSegment will panic given a negative index or index larger than the segment count.
	GetSegment(index int) *IPAddressSegment

	// GetSegments returns a slice with the address segments.  The returned slice is not backed by the same array as this section.
	GetSegments() []*IPAddressSegment

	// CopySegments copies the existing segments into the given slice,
	// as much as can be fit into the slice, returning the number of segments copied
	CopySegments(segs []*IPAddressSegment) (count int)

	// CopySubSegments copies the existing segments from the given start index until but not including the segment at the given end index,
	// into the given slice, as much as can be fit into the slice, returning the number of segments copied
	CopySubSegments(start, end int, segs []*IPAddressSegment) (count int)

	// IsIPv4 returns true if this series originated as an IPv4 series.  If so, use ToIPv4 to convert back to the IPv4-specific type.
	IsIPv4() bool

	// IsIPv6 returns true if this series originated as an IPv6 series.  If so, use ToIPv6 to convert back to the IPv6-specific type.
	IsIPv6() bool

	ToIPv4() IPv4AddressSegmentSeries
	ToIPv6() IPv6AddressSegmentSeries

	// ToBlock creates a new series block by changing the segment at the given index to have the given lower and upper value,
	// and changing the following segments to be full-range.
	ToBlock(segmentIndex int, lower, upper SegInt) ExtendedIPSegmentSeries

	// ToPrefixBlock returns the series with the same prefix as this series while the remaining bits span all values.
	// The series will be the block of all series with the same prefix.
	//
	// If this series has no prefix, this series is returned.
	ToPrefixBlock() ExtendedIPSegmentSeries

	// ToPrefixBlockLen returns the series with the same prefix of the given length as this series while the remaining bits span all values.
	// The returned series will be the block of all series with the same prefix.
	ToPrefixBlockLen(BitCount) ExtendedIPSegmentSeries

	ToZeroHostLen(BitCount) (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError)
	ToZeroHost() (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError)
	ToMaxHostLen(BitCount) (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError)
	ToMaxHost() (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError)
	ToZeroNetwork() ExtendedIPSegmentSeries

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
	Increment(int64) ExtendedIPSegmentSeries

	// IncrementBoundary returns the item that is the given increment from the range boundaries of this item.
	//
	// If the given increment is positive, adds the value to the highest (GetUpper) in the range to produce a new item.
	// If the given increment is negative, adds the value to the lowest (GetLower) in the range to produce a new item.
	// If the increment is zero, returns this.
	//
	// If this represents just a single value, this item is simply incremented by the given increment value, positive or negative.
	//
	// On overflow or underflow, IncrementBoundary returns nil.
	IncrementBoundary(int64) ExtendedIPSegmentSeries

	// GetLower returns the series in the range with the lowest numeric value,
	// which will be the same series if it represents a single value.
	// For example, for "1.2-3.4.5-6", the series "1.2.4.5" is returned.
	GetLower() ExtendedIPSegmentSeries

	// GetUpper returns the series in the range with the highest numeric value,
	// which will be the same series if it represents a single value.
	// For example, for "1.2-3.4.5-6", the series "1.3.4.6" is returned.
	GetUpper() ExtendedIPSegmentSeries

	// AssignPrefixForSingleBlock returns the equivalent prefix block that matches exactly the range of values in this series.
	// The returned block will have an assigned prefix length indicating the prefix length for the block.
	//
	// There may be no such series - it is required that the range of values match the range of a prefix block.
	// If there is no such series, then nil is returned.
	AssignPrefixForSingleBlock() ExtendedIPSegmentSeries

	// AssignMinPrefixForBlock returns an equivalent series, assigned the smallest prefix length possible,
	// such that the prefix block for that prefix length is in this series.
	//
	// In other words, this method assigns a prefix length to this series matching the largest prefix block in this series.
	AssignMinPrefixForBlock() ExtendedIPSegmentSeries

	// Iterator provides an iterator to iterate through the individual series of this series.
	//
	// When iterating, the prefix length is preserved.  Remove it using WithoutPrefixLen prior to iterating if you wish to drop it from all individual series.
	//
	// Call IsMultiple to determine if this instance represents multiple series, or GetCount for the count.
	Iterator() ExtendedIPSegmentSeriesIterator

	// PrefixIterator provides an iterator to iterate through the individual prefixes of this series,
	// each iterated element spanning the range of values for its prefix.
	//
	// It is similar to the prefix block iterator, except for possibly the first and last iterated elements, which might not be prefix blocks,
	// instead constraining themselves to values from this series.
	//
	// If the series has no prefix length, then this is equivalent to Iterator.
	PrefixIterator() ExtendedIPSegmentSeriesIterator

	// PrefixBlockIterator provides an iterator to iterate through the individual prefix blocks, one for each prefix of this series.
	// Each iterated series will be a prefix block with the same prefix length as this series.
	//
	// If this series has no prefix length, then this is equivalent to Iterator.
	PrefixBlockIterator() ExtendedIPSegmentSeriesIterator

	// SequentialBlockIterator iterates through the sequential series that make up this series.
	//
	// Practically, this means finding the count of segments for which the segments that follow are not full range, and then using BlockIterator with that segment count.
	//
	// Use GetSequentialBlockCount to get the number of iterated elements.
	SequentialBlockIterator() ExtendedIPSegmentSeriesIterator

	// BlockIterator Iterates through the series that can be obtained by iterating through all the upper segments up to the given segment count.
	// The segments following remain the same in all iterated series.
	BlockIterator(segmentCount int) ExtendedIPSegmentSeriesIterator

	SpanWithPrefixBlocks() []ExtendedIPSegmentSeries
	SpanWithSequentialBlocks() []ExtendedIPSegmentSeries

	CoverWithPrefixBlock() ExtendedIPSegmentSeries

	// AdjustPrefixLen increases or decreases the prefix length by the given increment.
	//
	// A prefix length will not be adjusted lower than zero or beyond the bit length of the series.
	//
	// If this series has no prefix length, then the prefix length will be set to the adjustment if positive,
	// or it will be set to the adjustment added to the bit count if negative.
	AdjustPrefixLen(BitCount) ExtendedIPSegmentSeries

	// AdjustPrefixLenZeroed increases or decreases the prefix length by the given increment while zeroing out the bits that have moved into or outside the prefix.
	//
	// A prefix length will not be adjusted lower than zero or beyond the bit length of the series.
	//
	// If this series has no prefix length, then the prefix length will be set to the adjustment if positive,
	// or it will be set to the adjustment added to the bit count if negative.
	//
	// When prefix length is increased, the bits moved within the prefix become zero.
	// When a prefix length is decreased, the bits moved outside the prefix become zero.
	//
	// If the result cannot be zeroed because zeroing out bits results in a non-contiguous segment, an error is returned.
	AdjustPrefixLenZeroed(BitCount) (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError)

	// SetPrefixLen sets the prefix length.
	//
	// A prefix length will not be set to a value lower than zero or beyond the bit length of the series.
	// The provided prefix length will be adjusted to these boundaries if necessary.
	SetPrefixLen(BitCount) ExtendedIPSegmentSeries

	// SetPrefixLenZeroed sets the prefix length.
	//
	// A prefix length will not be set to a value lower than zero or beyond the bit length of the series.
	// The provided prefix length will be adjusted to these boundaries if necessary.
	//
	// If this series has a prefix length, and the prefix length is increased when setting the new prefix length, the bits moved within the prefix become zero.
	// If this series has a prefix length, and the prefix length is decreased when setting the new prefix length, the bits moved outside the prefix become zero.
	//
	// In other words, bits that move from one side of the prefix length to the other (ie bits moved into the prefix or outside the prefix) are zeroed.
	//
	// If the result cannot be zeroed because zeroing out bits results in a non-contiguous segment, an error is returned.
	SetPrefixLenZeroed(BitCount) (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError)

	// WithoutPrefixLen provides the same address series but with no prefix length.  The values remain unchanged.
	WithoutPrefixLen() ExtendedIPSegmentSeries

	// ReverseBytes returns a new segment series with the bytes reversed.  Any prefix length is dropped.
	//
	// If each segment is more than 1 byte long, and the bytes within a single segment cannot be reversed because the segment represents a range,
	// and reversing the segment values results in a range that is not contiguous, then this returns an error.
	//
	// In practice this means that to be reversible, a range must include all values except possibly the largest and/or smallest, which reverse to themselves.
	ReverseBytes() (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError)

	// ReverseBits returns a new segment series with the bits reversed.  Any prefix length is dropped.
	//
	// If the bits within a single segment cannot be reversed because the segment represents a range,
	// and reversing the segment values results in a range that is not contiguous, this returns an error.
	//
	// In practice this means that to be reversible, a range must include all values except possibly the largest and/or smallest, which reverse to themselves.
	//
	// If perByte is true, the bits are reversed within each byte, otherwise all the bits are reversed.
	ReverseBits(perByte bool) (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError)

	// ReverseSegments returns a new series with the segments reversed.
	ReverseSegments() ExtendedIPSegmentSeries

	// ToCustomString creates a customized string from this series according to the given string option parameters
	ToCustomString(stringOptions addrstr.IPStringOptions) string
}

type WrappedIPAddress struct {
	*IPAddress
}

// Unwrap returns the wrapped *IPAddress as an interface, IPAddressSegmentSeries
func (addr WrappedIPAddress) Unwrap() IPAddressSegmentSeries {
	res := addr.IPAddress
	if res == nil {
		return nil
	}
	return res
}

func (addr WrappedIPAddress) ToIPv4() IPv4AddressSegmentSeries {
	return addr.IPAddress.ToIPv4()
}

func (addr WrappedIPAddress) ToIPv6() IPv6AddressSegmentSeries {
	return addr.IPAddress.ToIPv6()
}

func (addr WrappedIPAddress) GetNetworkMask() ExtendedIPSegmentSeries {
	return wrapIPAddress(addr.IPAddress.GetNetworkMask())
}

func (addr WrappedIPAddress) GetHostMask() ExtendedIPSegmentSeries {
	return wrapIPAddress(addr.IPAddress.GetHostMask())
}

// SequentialBlockIterator iterates through the sequential series that make up this series.
//
// Practically, this means finding the count of segments for which the segments that follow are not full range, and then using BlockIterator with that segment count.
//
// Use GetSequentialBlockCount to get the number of iterated elements.
func (addr WrappedIPAddress) SequentialBlockIterator() ExtendedIPSegmentSeriesIterator {
	return ipaddressSeriesIterator{addr.IPAddress.SequentialBlockIterator()}
}

// BlockIterator Iterates through the series that can be obtained by iterating through all the upper segments up to the given segment count.
// The segments following remain the same in all iterated series.
func (addr WrappedIPAddress) BlockIterator(segmentCount int) ExtendedIPSegmentSeriesIterator {
	return ipaddressSeriesIterator{addr.IPAddress.BlockIterator(segmentCount)}
}

// Iterator provides an iterator to iterate through the individual series of this series.
//
// When iterating, the prefix length is preserved.  Remove it using WithoutPrefixLen prior to iterating if you wish to drop it from all individual series.
//
// Call IsMultiple to determine if this instance represents multiple series, or GetCount for the count.
func (addr WrappedIPAddress) Iterator() ExtendedIPSegmentSeriesIterator {
	return ipaddressSeriesIterator{addr.IPAddress.Iterator()}
}

// PrefixIterator provides an iterator to iterate through the individual prefixes of this series,
// each iterated element spanning the range of values for its prefix.
//
// It is similar to the prefix block iterator, except for possibly the first and last iterated elements, which might not be prefix blocks,
// instead constraining themselves to values from this series.
//
// If the series has no prefix length, then this is equivalent to Iterator.
func (addr WrappedIPAddress) PrefixIterator() ExtendedIPSegmentSeriesIterator {
	return ipaddressSeriesIterator{addr.IPAddress.PrefixIterator()}
}

// PrefixBlockIterator provides an iterator to iterate through the individual prefix blocks, one for each prefix of this series.
// Each iterated series will be a prefix block with the same prefix length as this series.
//
// If this series has no prefix length, then this is equivalent to Iterator.
func (addr WrappedIPAddress) PrefixBlockIterator() ExtendedIPSegmentSeriesIterator {
	return ipaddressSeriesIterator{addr.IPAddress.PrefixBlockIterator()}
}

// ToBlock creates a new series block by changing the segment at the given index to have the given lower and upper value,
// and changing the following segments to be full-range.
func (addr WrappedIPAddress) ToBlock(segmentIndex int, lower, upper SegInt) ExtendedIPSegmentSeries {
	return wrapIPAddress(addr.IPAddress.ToBlock(segmentIndex, lower, upper))
}

// ToPrefixBlockLen returns the series with the same prefix of the given length as this series while the remaining bits span all values.
// The returned series will be the block of all series with the same prefix.
func (addr WrappedIPAddress) ToPrefixBlockLen(bitCount BitCount) ExtendedIPSegmentSeries {
	return wrapIPAddress(addr.IPAddress.ToPrefixBlockLen(bitCount))
}

// ToPrefixBlock returns the series with the same prefix as this series while the remaining bits span all values.
// The series will be the block of all series with the same prefix.
//
// If this series has no prefix, this series is returned.
func (addr WrappedIPAddress) ToPrefixBlock() ExtendedIPSegmentSeries {
	return wrapIPAddress(addr.IPAddress.ToPrefixBlock())
}

func (addr WrappedIPAddress) ToZeroHostLen(bitCount BitCount) (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapIPAddrWithErr(addr.IPAddress.ToZeroHostLen(bitCount)) //in IPAddress/Section
}

func (addr WrappedIPAddress) ToZeroHost() (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapIPAddrWithErr(addr.IPAddress.ToZeroHost()) // in IPAddress/Section/Segment
}

func (addr WrappedIPAddress) ToMaxHostLen(bitCount BitCount) (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapIPAddrWithErr(addr.IPAddress.ToMaxHostLen(bitCount))
}

func (addr WrappedIPAddress) ToMaxHost() (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapIPAddrWithErr(addr.IPAddress.ToMaxHost())
}

func (addr WrappedIPAddress) ToZeroNetwork() ExtendedIPSegmentSeries {
	return wrapIPAddress(addr.IPAddress.ToZeroNetwork()) //IPAddress/Section.  ToZeroHost() is in IPAddress/Section/Segment
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
func (addr WrappedIPAddress) Increment(i int64) ExtendedIPSegmentSeries {
	return convIPAddrToIntf(addr.IPAddress.Increment(i))
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
func (addr WrappedIPAddress) IncrementBoundary(i int64) ExtendedIPSegmentSeries {
	return convIPAddrToIntf(addr.IPAddress.IncrementBoundary(i))
}

// GetLower returns the series in the range with the lowest numeric value,
// which will be the same series if it represents a single value.
// For example, for "1.2-3.4.5-6", the series "1.2.4.5" is returned.
func (addr WrappedIPAddress) GetLower() ExtendedIPSegmentSeries {
	return wrapIPAddress(addr.IPAddress.GetLower())
}

// GetUpper returns the series in the range with the highest numeric value,
// which will be the same series if it represents a single value.
// For example, for "1.2-3.4.5-6", the series "1.3.4.6" is returned.
func (addr WrappedIPAddress) GetUpper() ExtendedIPSegmentSeries {
	return wrapIPAddress(addr.IPAddress.GetUpper())
}

// GetSection returns the backing section for this series, comprising all segments.
func (addr WrappedIPAddress) GetSection() *IPAddressSection {
	return addr.IPAddress.GetSection()
}

// AssignPrefixForSingleBlock returns the equivalent prefix block that matches exactly the range of values in this series.
// The returned block will have an assigned prefix length indicating the prefix length for the block.
//
// There may be no such series - it is required that the range of values match the range of a prefix block.
// If there is no such series, then nil is returned.
func (addr WrappedIPAddress) AssignPrefixForSingleBlock() ExtendedIPSegmentSeries {
	return convIPAddrToIntf(addr.IPAddress.AssignPrefixForSingleBlock())
}

// AssignMinPrefixForBlock returns an equivalent series, assigned the smallest prefix length possible,
// such that the prefix block for that prefix length is in this series.
//
// In other words, this method assigns a prefix length to this series matching the largest prefix block in this series.
func (addr WrappedIPAddress) AssignMinPrefixForBlock() ExtendedIPSegmentSeries {
	return wrapIPAddress(addr.IPAddress.AssignMinPrefixForBlock())
}

// WithoutPrefixLen provides the same address series but with no prefix length.  The values remain unchanged.
func (addr WrappedIPAddress) WithoutPrefixLen() ExtendedIPSegmentSeries {
	return wrapIPAddress(addr.IPAddress.WithoutPrefixLen())
}

func (addr WrappedIPAddress) SpanWithPrefixBlocks() []ExtendedIPSegmentSeries {
	return addr.IPAddress.spanWithPrefixBlocks()
}

func (addr WrappedIPAddress) SpanWithSequentialBlocks() []ExtendedIPSegmentSeries {
	return addr.IPAddress.spanWithSequentialBlocks()
}

func (addr WrappedIPAddress) CoverWithPrefixBlock() ExtendedIPSegmentSeries {
	return addr.IPAddress.coverSeriesWithPrefixBlock()
}

// Contains returns whether this is same type and version as the given address series and whether it contains all values in the given series.
//
// Series must also have the same number of segments to be comparable, otherwise false is returned.
func (addr WrappedIPAddress) Contains(other ExtendedIPSegmentSeries) bool {
	a, ok := other.Unwrap().(AddressType)
	return ok && addr.IPAddress.Contains(a)
}

// CompareSize compares the counts of two address series, the number of individual series represented in each.
//
// Rather than calculating counts with GetCount, there can be more efficient ways of comparing whether one series represents more individual address series than another.
//
// CompareSize returns a positive integer if this address series has a larger count than the one given, 0 if they are the same, or a negative integer if the other has a larger count.
func (addr WrappedIPAddress) CompareSize(other ExtendedIPSegmentSeries) int {
	if a, ok := other.Unwrap().(AddressType); ok {
		return addr.IPAddress.CompareSize(a)
	}
	return addr.GetCount().Cmp(other.GetCount())
}

// Equal returns whether the given address series is equal to this address series.
// Two address series are equal if they represent the same set of series.
// Both must be equal addresses.
func (addr WrappedIPAddress) Equal(other ExtendedIPSegmentSeries) bool {
	a, ok := other.Unwrap().(AddressType)
	return ok && addr.IPAddress.Equal(a)
}

// SetPrefixLen sets the prefix length.
//
// A prefix length will not be set to a value lower than zero or beyond the bit length of the series.
// The provided prefix length will be adjusted to these boundaries if necessary.
func (addr WrappedIPAddress) SetPrefixLen(prefixLen BitCount) ExtendedIPSegmentSeries {
	return wrapIPAddress(addr.IPAddress.SetPrefixLen(prefixLen))
}

// SetPrefixLenZeroed sets the prefix length.
//
// A prefix length will not be set to a value lower than zero or beyond the bit length of the series.
// The provided prefix length will be adjusted to these boundaries if necessary.
//
// If this series has a prefix length, and the prefix length is increased when setting the new prefix length, the bits moved within the prefix become zero.
// If this series has a prefix length, and the prefix length is decreased when setting the new prefix length, the bits moved outside the prefix become zero.
//
// In other words, bits that move from one side of the prefix length to the other (ie bits moved into the prefix or outside the prefix) are zeroed.
//
// If the result cannot be zeroed because zeroing out bits results in a non-contiguous segment, an error is returned.
func (addr WrappedIPAddress) SetPrefixLenZeroed(prefixLen BitCount) (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapIPAddrWithErr(addr.IPAddress.SetPrefixLenZeroed(prefixLen))
}

// AdjustPrefixLen increases or decreases the prefix length by the given increment.
//
// A prefix length will not be adjusted lower than zero or beyond the bit length of the series.
//
// If this series has no prefix length, then the prefix length will be set to the adjustment if positive,
// or it will be set to the adjustment added to the bit count if negative.
func (addr WrappedIPAddress) AdjustPrefixLen(prefixLen BitCount) ExtendedIPSegmentSeries {
	return wrapIPAddress(addr.IPAddress.AdjustPrefixLen(prefixLen))
}

// AdjustPrefixLenZeroed increases or decreases the prefix length by the given increment while zeroing out the bits that have moved into or outside the prefix.
//
// A prefix length will not be adjusted lower than zero or beyond the bit length of the series.
//
// If this series has no prefix length, then the prefix length will be set to the adjustment if positive,
// or it will be set to the adjustment added to the bit count if negative.
//
// When prefix length is increased, the bits moved within the prefix become zero.
// When a prefix length is decreased, the bits moved outside the prefix become zero.
//
// If the result cannot be zeroed because zeroing out bits results in a non-contiguous segment, an error is returned.
func (addr WrappedIPAddress) AdjustPrefixLenZeroed(prefixLen BitCount) (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapIPAddrWithErr(addr.IPAddress.AdjustPrefixLenZeroed(prefixLen))
}

// ReverseBytes returns a new segment series with the bytes reversed.  Any prefix length is dropped.
//
// If each segment is more than 1 byte long, and the bytes within a single segment cannot be reversed because the segment represents a range,
// and reversing the segment values results in a range that is not contiguous, then this returns an error.
//
// In practice this means that to be reversible, a range must include all values except possibly the largest and/or smallest, which reverse to themselves.
func (addr WrappedIPAddress) ReverseBytes() (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapIPAddrWithErr(addr.IPAddress.ReverseBytes())
}

// ReverseBits returns a new segment series with the bits reversed.  Any prefix length is dropped.
//
// If the bits within a single segment cannot be reversed because the segment represents a range,
// and reversing the segment values results in a range that is not contiguous, this returns an error.
//
// In practice this means that to be reversible, a range must include all values except possibly the largest and/or smallest, which reverse to themselves.
//
// If perByte is true, the bits are reversed within each byte, otherwise all the bits are reversed.
func (addr WrappedIPAddress) ReverseBits(perByte bool) (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapIPAddrWithErr(addr.IPAddress.ReverseBits(perByte))
}

// ReverseSegments returns a new series with the segments reversed.
func (addr WrappedIPAddress) ReverseSegments() ExtendedIPSegmentSeries {
	return wrapIPAddress(addr.IPAddress.ReverseSegments())
}

type WrappedIPAddressSection struct {
	*IPAddressSection
}

// Unwrap returns the wrapped *IPAddressSection as an interface, IPAddressSegmentSeries
func (section WrappedIPAddressSection) Unwrap() IPAddressSegmentSeries {
	res := section.IPAddressSection
	if res == nil {
		return nil
	}
	return res
}

func (section WrappedIPAddressSection) ToIPv4() IPv4AddressSegmentSeries {
	return section.IPAddressSection.ToIPv4()
}

func (section WrappedIPAddressSection) ToIPv6() IPv6AddressSegmentSeries {
	return section.IPAddressSection.ToIPv6()
}

func (section WrappedIPAddressSection) GetNetworkMask() ExtendedIPSegmentSeries {
	return wrapIPSection(section.IPAddressSection.GetNetworkMask())
}

func (section WrappedIPAddressSection) GetHostMask() ExtendedIPSegmentSeries {
	return wrapIPSection(section.IPAddressSection.GetHostMask())
}

// SequentialBlockIterator iterates through the sequential series that make up this series.
//
// Practically, this means finding the count of segments for which the segments that follow are not full range, and then using BlockIterator with that segment count.
//
// Use GetSequentialBlockCount to get the number of iterated elements.
func (section WrappedIPAddressSection) SequentialBlockIterator() ExtendedIPSegmentSeriesIterator {
	return ipsectionSeriesIterator{section.IPAddressSection.SequentialBlockIterator()}
}

// BlockIterator Iterates through the series that can be obtained by iterating through all the upper segments up to the given segment count.
// The segments following remain the same in all iterated series.
func (section WrappedIPAddressSection) BlockIterator(segmentCount int) ExtendedIPSegmentSeriesIterator {
	return ipsectionSeriesIterator{section.IPAddressSection.BlockIterator(segmentCount)}
}

// Iterator provides an iterator to iterate through the individual series of this series.
//
// When iterating, the prefix length is preserved.  Remove it using WithoutPrefixLen prior to iterating if you wish to drop it from all individual series.
//
// Call IsMultiple to determine if this instance represents multiple series, or GetCount for the count.
func (section WrappedIPAddressSection) Iterator() ExtendedIPSegmentSeriesIterator {
	return ipsectionSeriesIterator{section.IPAddressSection.Iterator()}
}

// PrefixIterator provides an iterator to iterate through the individual prefixes of this series,
// each iterated element spanning the range of values for its prefix.
//
// It is similar to the prefix block iterator, except for possibly the first and last iterated elements, which might not be prefix blocks,
// instead constraining themselves to values from this series.
//
// If the series has no prefix length, then this is equivalent to Iterator.
func (section WrappedIPAddressSection) PrefixIterator() ExtendedIPSegmentSeriesIterator {
	return ipsectionSeriesIterator{section.IPAddressSection.PrefixIterator()}
}

// PrefixBlockIterator provides an iterator to iterate through the individual prefix blocks, one for each prefix of this series.
// Each iterated series will be a prefix block with the same prefix length as this series.
//
// If this series has no prefix length, then this is equivalent to Iterator.
func (section WrappedIPAddressSection) PrefixBlockIterator() ExtendedIPSegmentSeriesIterator {
	return ipsectionSeriesIterator{section.IPAddressSection.PrefixBlockIterator()}
}

// ToBlock creates a new series block by changing the segment at the given index to have the given lower and upper value,
// and changing the following segments to be full-range.
func (section WrappedIPAddressSection) ToBlock(segmentIndex int, lower, upper SegInt) ExtendedIPSegmentSeries {
	return wrapIPSection(section.IPAddressSection.ToBlock(segmentIndex, lower, upper))
}

// ToPrefixBlockLen returns the series with the same prefix of the given length as this series while the remaining bits span all values.
// The returned series will be the block of all series with the same prefix.
func (section WrappedIPAddressSection) ToPrefixBlockLen(bitCount BitCount) ExtendedIPSegmentSeries {
	return wrapIPSection(section.IPAddressSection.ToPrefixBlockLen(bitCount))
}

// ToPrefixBlock returns the series with the same prefix as this series while the remaining bits span all values.
// The series will be the block of all series with the same prefix.
//
// If this series has no prefix, this series is returned.
func (section WrappedIPAddressSection) ToPrefixBlock() ExtendedIPSegmentSeries {
	return wrapIPSection(section.IPAddressSection.ToPrefixBlock())
}

func (section WrappedIPAddressSection) ToZeroHostLen(bitCount BitCount) (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapIPSectWithErr(section.IPAddressSection.ToZeroHostLen(bitCount))
}

func (section WrappedIPAddressSection) ToZeroHost() (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapIPSectWithErr(section.IPAddressSection.ToZeroHost())
}

func (section WrappedIPAddressSection) ToMaxHostLen(bitCount BitCount) (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapIPSectWithErr(section.IPAddressSection.ToMaxHostLen(bitCount))
}

func (section WrappedIPAddressSection) ToMaxHost() (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapIPSectWithErr(section.IPAddressSection.ToMaxHost())
}

func (section WrappedIPAddressSection) ToZeroNetwork() ExtendedIPSegmentSeries {
	return wrapIPSection(section.IPAddressSection.ToZeroNetwork())
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
func (section WrappedIPAddressSection) Increment(i int64) ExtendedIPSegmentSeries {
	return convIPSectToIntf(section.IPAddressSection.Increment(i))
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
func (section WrappedIPAddressSection) IncrementBoundary(i int64) ExtendedIPSegmentSeries {
	return convIPSectToIntf(section.IPAddressSection.IncrementBoundary(i))
}

// GetLower returns the series in the range with the lowest numeric value,
// which will be the same series if it represents a single value.
// For example, for "1.2-3.4.5-6", the series "1.2.4.5" is returned.
func (section WrappedIPAddressSection) GetLower() ExtendedIPSegmentSeries {
	return wrapIPSection(section.IPAddressSection.GetLower())
}

// GetUpper returns the series in the range with the highest numeric value,
// which will be the same series if it represents a single value.
// For example, for "1.2-3.4.5-6", the series "1.3.4.6" is returned.
func (section WrappedIPAddressSection) GetUpper() ExtendedIPSegmentSeries {
	return wrapIPSection(section.IPAddressSection.GetUpper())
}

// GetSection returns the backing section for this series, comprising all segments.
func (section WrappedIPAddressSection) GetSection() *IPAddressSection {
	return section.IPAddressSection
}

// AssignPrefixForSingleBlock returns the equivalent prefix block that matches exactly the range of values in this series.
// The returned block will have an assigned prefix length indicating the prefix length for the block.
//
// There may be no such series - it is required that the range of values match the range of a prefix block.
// If there is no such series, then nil is returned.
func (section WrappedIPAddressSection) AssignPrefixForSingleBlock() ExtendedIPSegmentSeries {
	return convIPSectToIntf(section.IPAddressSection.AssignPrefixForSingleBlock())
}

// AssignMinPrefixForBlock returns an equivalent series, assigned the smallest prefix length possible,
// such that the prefix block for that prefix length is in this series.
//
// In other words, this method assigns a prefix length to this series matching the largest prefix block in this series.
func (section WrappedIPAddressSection) AssignMinPrefixForBlock() ExtendedIPSegmentSeries {
	return wrapIPSection(section.IPAddressSection.AssignMinPrefixForBlock())
}

// WithoutPrefixLen provides the same address series but with no prefix length.  The values remain unchanged.
func (section WrappedIPAddressSection) WithoutPrefixLen() ExtendedIPSegmentSeries {
	return wrapIPSection(section.IPAddressSection.WithoutPrefixLen())
}

func (section WrappedIPAddressSection) SpanWithPrefixBlocks() []ExtendedIPSegmentSeries {
	return section.IPAddressSection.spanWithPrefixBlocks()
}

func (section WrappedIPAddressSection) SpanWithSequentialBlocks() []ExtendedIPSegmentSeries {
	return section.IPAddressSection.spanWithSequentialBlocks()
}

func (section WrappedIPAddressSection) CoverWithPrefixBlock() ExtendedIPSegmentSeries {
	return section.IPAddressSection.coverSeriesWithPrefixBlock()
}

// Contains returns whether this is same type and version as the given address series and whether it contains all values in the given series.
//
// Series must also have the same number of segments to be comparable, otherwise false is returned.
func (section WrappedIPAddressSection) Contains(other ExtendedIPSegmentSeries) bool {
	s, ok := other.Unwrap().(AddressSectionType)
	return ok && section.IPAddressSection.Contains(s)
}

// Equal returns whether the given address series is equal to this address series.
// Two address series are equal if they represent the same set of series.
// Both must be equal sections.
func (section WrappedIPAddressSection) Equal(other ExtendedIPSegmentSeries) bool {
	s, ok := other.Unwrap().(AddressSectionType)
	return ok && section.IPAddressSection.Equal(s)
}

// CompareSize compares the counts of two address series, the number of individual series represented in each.
//
// Rather than calculating counts with GetCount, there can be more efficient ways of comparing whether one series represents more individual address series than another.
//
// CompareSize returns a positive integer if this address series has a larger count than the one given, 0 if they are the same, or a negative integer if the other has a larger count.
func (section WrappedIPAddressSection) CompareSize(other ExtendedIPSegmentSeries) int {
	if s, ok := other.Unwrap().(AddressSectionType); ok {
		return section.IPAddressSection.CompareSize(s)
	}
	return section.GetCount().Cmp(other.GetCount())
}

// SetPrefixLen sets the prefix length.
//
// A prefix length will not be set to a value lower than zero or beyond the bit length of the series.
// The provided prefix length will be adjusted to these boundaries if necessary.
func (section WrappedIPAddressSection) SetPrefixLen(prefixLen BitCount) ExtendedIPSegmentSeries {
	return wrapIPSection(section.IPAddressSection.SetPrefixLen(prefixLen))
}

// SetPrefixLenZeroed sets the prefix length.
//
// A prefix length will not be set to a value lower than zero or beyond the bit length of the series.
// The provided prefix length will be adjusted to these boundaries if necessary.
//
// If this series has a prefix length, and the prefix length is increased when setting the new prefix length, the bits moved within the prefix become zero.
// If this series has a prefix length, and the prefix length is decreased when setting the new prefix length, the bits moved outside the prefix become zero.
//
// In other words, bits that move from one side of the prefix length to the other (ie bits moved into the prefix or outside the prefix) are zeroed.
//
// If the result cannot be zeroed because zeroing out bits results in a non-contiguous segment, an error is returned.
func (section WrappedIPAddressSection) SetPrefixLenZeroed(prefixLen BitCount) (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapIPSectWithErr(section.IPAddressSection.SetPrefixLenZeroed(prefixLen))
}

// AdjustPrefixLen increases or decreases the prefix length by the given increment.
//
// A prefix length will not be adjusted lower than zero or beyond the bit length of the series.
//
// If this series has no prefix length, then the prefix length will be set to the adjustment if positive,
// or it will be set to the adjustment added to the bit count if negative.
func (section WrappedIPAddressSection) AdjustPrefixLen(prefixLen BitCount) ExtendedIPSegmentSeries {
	return wrapIPSection(section.IPAddressSection.AdjustPrefixLen(prefixLen))
}

// AdjustPrefixLenZeroed increases or decreases the prefix length by the given increment while zeroing out the bits that have moved into or outside the prefix.
//
// A prefix length will not be adjusted lower than zero or beyond the bit length of the series.
//
// If this series has no prefix length, then the prefix length will be set to the adjustment if positive,
// or it will be set to the adjustment added to the bit count if negative.
//
// When prefix length is increased, the bits moved within the prefix become zero.
// When a prefix length is decreased, the bits moved outside the prefix become zero.
//
// If the result cannot be zeroed because zeroing out bits results in a non-contiguous segment, an error is returned.
func (section WrappedIPAddressSection) AdjustPrefixLenZeroed(prefixLen BitCount) (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapIPSectWithErr(section.IPAddressSection.AdjustPrefixLenZeroed(prefixLen))
}

// ReverseBytes returns a new segment series with the bytes reversed.  Any prefix length is dropped.
//
// If each segment is more than 1 byte long, and the bytes within a single segment cannot be reversed because the segment represents a range,
// and reversing the segment values results in a range that is not contiguous, then this returns an error.
//
// In practice this means that to be reversible, a range must include all values except possibly the largest and/or smallest, which reverse to themselves.
func (section WrappedIPAddressSection) ReverseBytes() (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapIPSectWithErr(section.IPAddressSection.ReverseBytes())
}

// ReverseBits returns a new segment series with the bits reversed.  Any prefix length is dropped.
//
// If the bits within a single segment cannot be reversed because the segment represents a range,
// and reversing the segment values results in a range that is not contiguous, this returns an error.
//
// In practice this means that to be reversible, a range must include all values except possibly the largest and/or smallest, which reverse to themselves.
//
// If perByte is true, the bits are reversed within each byte, otherwise all the bits are reversed.
func (section WrappedIPAddressSection) ReverseBits(perByte bool) (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapIPSectWithErr(section.IPAddressSection.ReverseBits(perByte))
}

// ReverseSegments returns a new series with the segments reversed.
func (section WrappedIPAddressSection) ReverseSegments() ExtendedIPSegmentSeries {
	return wrapIPSection(section.IPAddressSection.ReverseSegments())
}

var _ ExtendedIPSegmentSeries = WrappedIPAddress{}
var _ ExtendedIPSegmentSeries = WrappedIPAddressSection{}

// In go, a nil value is not coverted to a nil interface, it is converted to a non-nil interface instance with underlying value nil
func convIPAddrToIntf(addr *IPAddress) ExtendedIPSegmentSeries {
	if addr == nil {
		return nil
	}
	return wrapIPAddress(addr)
}

func convIPSectToIntf(sect *IPAddressSection) ExtendedIPSegmentSeries {
	if sect == nil {
		return nil
	}
	return wrapIPSection(sect)
}

func wrapIPSectWithErr(section *IPAddressSection, err addrerr.IncompatibleAddressError) (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError) {
	if err == nil {
		return wrapIPSection(section), nil
	}
	return nil, err
}

func wrapIPAddrWithErr(addr *IPAddress, err addrerr.IncompatibleAddressError) (ExtendedIPSegmentSeries, addrerr.IncompatibleAddressError) {
	if err == nil {
		return wrapIPAddress(addr), nil
	}
	return nil, err
}

func wrapIPAddress(addr *IPAddress) WrappedIPAddress {
	return WrappedIPAddress{addr}
}

func wrapIPSection(section *IPAddressSection) WrappedIPAddressSection {
	return WrappedIPAddressSection{section}
}
