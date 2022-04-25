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

// ExtendedSegmentSeries wraps either an Address or AddressSection.
// ExtendedSegmentSeries can be used to write code that works with both Addresses and Address Sections,
// going further than AddressSegmentSeries to offer additional methods with the series types in their signature.
type ExtendedSegmentSeries interface {
	AddressSegmentSeries

	// Unwrap returns the wrapped *Address or *AddressSection as an interface, AddressSegmentSeries
	Unwrap() AddressSegmentSeries

	Equal(ExtendedSegmentSeries) bool

	// Contains returns whether this is same type and version as the given address series and whether it contains all values in the given series.
	//
	// Series must also have the same number of segments to be comparable, otherwise false is returned.
	Contains(ExtendedSegmentSeries) bool

	// CompareSize compares the counts of two address series, the number of individual series represented in each.
	//
	// Rather than calculating counts with GetCount, there can be more efficient ways of comparing whether one series represents more individual address series than another.
	//
	// CompareSize returns a positive integer if this address series has a larger count than the one given, 0 if they are the same, or a negative integer if the other has a larger count.
	CompareSize(ExtendedSegmentSeries) int

	// GetSection returns the full address section
	GetSection() *AddressSection

	// GetTrailingSection returns an ending subsection of the full address section
	GetTrailingSection(index int) *AddressSection

	// GetSubSection returns a subsection of the full address section
	GetSubSection(index, endIndex int) *AddressSection

	// GetSegment returns the segment at the given index.
	// The first segment is at index 0.
	// GetSegment will panic given a negative index or index larger than the segment count.
	GetSegment(index int) *AddressSegment

	// GetSegments returns a slice with the address segments.  The returned slice is not backed by the same array as this section.
	GetSegments() []*AddressSegment

	// CopySegments copies the existing segments into the given slice,
	// as much as can be fit into the slice, returning the number of segments copied
	CopySegments(segs []*AddressSegment) (count int)

	// CopySubSegments copies the existing segments from the given start index until but not including the segment at the given end index,
	// into the given slice, as much as can be fit into the slice, returning the number of segments copied
	CopySubSegments(start, end int, segs []*AddressSegment) (count int)

	// IsIP returns true if this series originated as an IPv4 or IPv6 series, or a zero-length IP series.  If so, use ToIP to convert back to the IP-specific type.
	IsIP() bool

	// IsIPv4 returns true if this series originated as an IPv4 series.  If so, use ToIPv4 to convert back to the IPv4-specific type.
	IsIPv4() bool

	// IsIPv6 returns true if this series originated as an IPv6 series.  If so, use ToIPv6 to convert back to the IPv6-specific type.
	IsIPv6() bool

	// IsMAC returns true if this series originated as a MAC series.  If so, use ToMAC to convert back to the MAC-specific type.
	IsMAC() bool

	ToIP() IPAddressSegmentSeries
	ToIPv4() IPv4AddressSegmentSeries
	ToIPv6() IPv6AddressSegmentSeries
	ToMAC() MACAddressSegmentSeries

	// ToBlock creates a new series block by changing the segment at the given index to have the given lower and upper value,
	// and changing the following segments to be full-range.
	ToBlock(segmentIndex int, lower, upper SegInt) ExtendedSegmentSeries

	// ToPrefixBlock returns the series with the same prefix as this series while the remaining bits span all values.
	// The series will be the block of all series with the same prefix.
	//
	// If this series has no prefix, this series is returned.
	ToPrefixBlock() ExtendedSegmentSeries

	// ToPrefixBlockLen returns the series with the same prefix of the given length as this series while the remaining bits span all values.
	// The returned series will be the block of all series with the same prefix.
	ToPrefixBlockLen(prefLen BitCount) ExtendedSegmentSeries

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
	Increment(int64) ExtendedSegmentSeries

	// IncrementBoundary returns the item that is the given increment from the range boundaries of this item.
	//
	// If the given increment is positive, adds the value to the highest (GetUpper) in the range to produce a new item.
	// If the given increment is negative, adds the value to the lowest (GetLower) in the range to produce a new item.
	// If the increment is zero, returns this.
	//
	// If this represents just a single value, this item is simply incremented by the given increment value, positive or negative.
	//
	// On overflow or underflow, IncrementBoundary returns nil.
	IncrementBoundary(int64) ExtendedSegmentSeries

	// GetLower returns the series in the range with the lowest numeric value,
	// which will be the same series if it represents a single value.
	GetLower() ExtendedSegmentSeries

	// GetUpper returns the series in the range with the highest numeric value,
	// which will be the same series if it represents a single value.
	GetUpper() ExtendedSegmentSeries

	AssignPrefixForSingleBlock() ExtendedSegmentSeries
	AssignMinPrefixForBlock() ExtendedSegmentSeries

	Iterator() ExtendedSegmentSeriesIterator
	PrefixIterator() ExtendedSegmentSeriesIterator
	PrefixBlockIterator() ExtendedSegmentSeriesIterator

	AdjustPrefixLen(BitCount) ExtendedSegmentSeries
	AdjustPrefixLenZeroed(BitCount) (ExtendedSegmentSeries, addrerr.IncompatibleAddressError)
	SetPrefixLen(BitCount) ExtendedSegmentSeries
	SetPrefixLenZeroed(BitCount) (ExtendedSegmentSeries, addrerr.IncompatibleAddressError)
	WithoutPrefixLen() ExtendedSegmentSeries

	// ReverseBytes returns a new segment series with the bytes reversed.  Any prefix length is dropped.
	//
	// If each segment is more than 1 byte long, and the bytes within a single segment cannot be reversed because the segment represents a range,
	// and reversing the segment values results in a range that is not contiguous, then this returns an error.
	//
	// In practice this means that to be reversible, a range must include all values except possibly the largest and/or smallest, which reverse to themselves.
	ReverseBytes() (ExtendedSegmentSeries, addrerr.IncompatibleAddressError)

	// ReverseBits returns a new segment series with the bits reversed.  Any prefix length is dropped.
	//
	// If the bits within a single segment cannot be reversed because the segment represents a range,
	// and reversing the segment values results in a range that is not contiguous, this returns an error.
	//
	// In practice this means that to be reversible, a range must include all values except possibly the largest and/or smallest, which reverse to themselves.
	ReverseBits(perByte bool) (ExtendedSegmentSeries, addrerr.IncompatibleAddressError)

	// ReverseSegments returns a new series with the segments reversed.
	ReverseSegments() ExtendedSegmentSeries

	ToCustomString(stringOptions addrstr.StringOptions) string
}

// WrappedAddress is the implementation of ExtendedSegmentSeries for Address
type WrappedAddress struct {
	*Address
}

func (addr WrappedAddress) Unwrap() AddressSegmentSeries {
	res := addr.Address
	if res == nil {
		return nil
	}
	return res
}

func (addr WrappedAddress) ToIPv4() IPv4AddressSegmentSeries {
	return addr.Address.ToIPv4()
}

func (addr WrappedAddress) ToIPv6() IPv6AddressSegmentSeries {
	return addr.Address.ToIPv6()
}

func (addr WrappedAddress) ToIP() IPAddressSegmentSeries {
	return addr.Address.ToIP()
}

func (addr WrappedAddress) ToMAC() MACAddressSegmentSeries {
	return addr.Address.ToMAC()
}

func (addr WrappedAddress) Iterator() ExtendedSegmentSeriesIterator {
	return addressSeriesIterator{addr.Address.Iterator()}
}

func (addr WrappedAddress) PrefixIterator() ExtendedSegmentSeriesIterator {
	return addressSeriesIterator{addr.Address.PrefixIterator()}
}

func (addr WrappedAddress) PrefixBlockIterator() ExtendedSegmentSeriesIterator {
	return addressSeriesIterator{addr.Address.PrefixBlockIterator()}
}

// ToBlock creates a new series block by changing the segment at the given index to have the given lower and upper value,
// and changing the following segments to be full-range.
func (addr WrappedAddress) ToBlock(segmentIndex int, lower, upper SegInt) ExtendedSegmentSeries {
	return WrapAddress(addr.Address.ToBlock(segmentIndex, lower, upper))
}

// ToPrefixBlock returns the series with the same prefix as this series while the remaining bits span all values.
// The series will be the block of all series with the same prefix.
//
// If this series has no prefix, this series is returned.
func (addr WrappedAddress) ToPrefixBlock() ExtendedSegmentSeries {
	return WrapAddress(addr.Address.ToPrefixBlock())
}

// ToPrefixBlockLen returns the series with the same prefix of the given length as this series while the remaining bits span all values.
// The returned series will be the block of all series with the same prefix.
func (addr WrappedAddress) ToPrefixBlockLen(prefLen BitCount) ExtendedSegmentSeries {
	return WrapAddress(addr.Address.ToPrefixBlockLen(prefLen))
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
func (addr WrappedAddress) Increment(i int64) ExtendedSegmentSeries {
	return convAddrToIntf(addr.Address.Increment(i))
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
func (addr WrappedAddress) IncrementBoundary(i int64) ExtendedSegmentSeries {
	return convAddrToIntf(addr.Address.IncrementBoundary(i))
}

// GetLower returns the series in the range with the lowest numeric value,
// which will be the same series if it represents a single value.
func (addr WrappedAddress) GetLower() ExtendedSegmentSeries {
	return WrapAddress(addr.Address.GetLower())
}

// GetUpper returns the series in the range with the highest numeric value,
// which will be the same series if it represents a single value.
func (addr WrappedAddress) GetUpper() ExtendedSegmentSeries {
	return WrapAddress(addr.Address.GetUpper())
}

func (addr WrappedAddress) GetSection() *AddressSection {
	return addr.Address.GetSection()
}

func (addr WrappedAddress) AssignPrefixForSingleBlock() ExtendedSegmentSeries {
	return convAddrToIntf(addr.Address.AssignPrefixForSingleBlock())
}

func (addr WrappedAddress) AssignMinPrefixForBlock() ExtendedSegmentSeries {
	return WrapAddress(addr.Address.AssignMinPrefixForBlock())
}

func (addr WrappedAddress) WithoutPrefixLen() ExtendedSegmentSeries {
	return WrapAddress(addr.Address.WithoutPrefixLen())
}

// Contains returns whether this is same type and version as the given address series and whether it contains all values in the given series.
//
// Series must also have the same number of segments to be comparable, otherwise false is returned.
func (addr WrappedAddress) Contains(other ExtendedSegmentSeries) bool {
	a, ok := other.Unwrap().(AddressType)
	return ok && addr.Address.Contains(a)
}

func (addr WrappedAddress) Equal(other ExtendedSegmentSeries) bool {
	a, ok := other.Unwrap().(AddressType)
	return ok && addr.Address.Equal(a)
}

// CompareSize compares the counts of two address series, the number of individual series represented in each.
//
// Rather than calculating counts with GetCount, there can be more efficient ways of comparing whether one series represents more individual address series than another.
//
// CompareSize returns a positive integer if this address series has a larger count than the one given, 0 if they are the same, or a negative integer if the other has a larger count.
func (addr WrappedAddress) CompareSize(other ExtendedSegmentSeries) int {
	if a, ok := other.Unwrap().(AddressType); ok {
		return addr.Address.CompareSize(a)
	}
	return addr.GetCount().Cmp(other.GetCount())
}

func (addr WrappedAddress) SetPrefixLen(prefixLen BitCount) ExtendedSegmentSeries {
	return WrapAddress(addr.Address.SetPrefixLen(prefixLen))
}

func (addr WrappedAddress) SetPrefixLenZeroed(prefixLen BitCount) (ExtendedSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapAddrWithErr(addr.Address.SetPrefixLenZeroed(prefixLen))
}

func (addr WrappedAddress) AdjustPrefixLen(prefixLen BitCount) ExtendedSegmentSeries {
	return WrapAddress(addr.Address.AdjustPrefixLen(prefixLen))
}

func (addr WrappedAddress) AdjustPrefixLenZeroed(prefixLen BitCount) (ExtendedSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapAddrWithErr(addr.Address.AdjustPrefixLenZeroed(prefixLen))
}

// ReverseBytes returns a new segment series with the bytes reversed.  Any prefix length is dropped.
//
// If each segment is more than 1 byte long, and the bytes within a single segment cannot be reversed because the segment represents a range,
// and reversing the segment values results in a range that is not contiguous, then this returns an error.
//
// In practice this means that to be reversible, a range must include all values except possibly the largest and/or smallest, which reverse to themselves.
func (addr WrappedAddress) ReverseBytes() (ExtendedSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapAddrWithErr(addr.Address.ReverseBytes())
}

// ReverseBits returns a new segment series with the bits reversed.  Any prefix length is dropped.
//
// If the bits within a single segment cannot be reversed because the segment represents a range,
// and reversing the segment values results in a range that is not contiguous, this returns an error.
//
// In practice this means that to be reversible, a range must include all values except possibly the largest and/or smallest, which reverse to themselves.
//
// If perByte is true, the bits are reversed within each byte, otherwise all the bits are reversed.
func (addr WrappedAddress) ReverseBits(perByte bool) (ExtendedSegmentSeries, addrerr.IncompatibleAddressError) {
	a, err := addr.Address.ReverseBits(perByte)
	if err != nil {
		return nil, err
	}
	return WrapAddress(a), nil
}

// ReverseSegments returns a new series with the segments reversed.
func (addr WrappedAddress) ReverseSegments() ExtendedSegmentSeries {
	return WrapAddress(addr.Address.ReverseSegments())
}

type WrappedAddressSection struct {
	*AddressSection
}

func (section WrappedAddressSection) Unwrap() AddressSegmentSeries {
	res := section.AddressSection
	if res == nil {
		return nil
	}
	return res
}

func (section WrappedAddressSection) ToIPv4() IPv4AddressSegmentSeries {
	return section.AddressSection.ToIPv4()
}

func (section WrappedAddressSection) ToIPv6() IPv6AddressSegmentSeries {
	return section.AddressSection.ToIPv6()
}

func (section WrappedAddressSection) ToIP() IPAddressSegmentSeries {
	return section.AddressSection.ToIP()
}

func (section WrappedAddressSection) ToMAC() MACAddressSegmentSeries {
	return section.AddressSection.ToMAC()
}

func (section WrappedAddressSection) Iterator() ExtendedSegmentSeriesIterator {
	return sectionSeriesIterator{section.AddressSection.Iterator()}
}

func (section WrappedAddressSection) PrefixIterator() ExtendedSegmentSeriesIterator {
	return sectionSeriesIterator{section.AddressSection.PrefixIterator()}
}

func (section WrappedAddressSection) PrefixBlockIterator() ExtendedSegmentSeriesIterator {
	return sectionSeriesIterator{section.AddressSection.PrefixBlockIterator()}
}

// ToBlock creates a new series block by changing the segment at the given index to have the given lower and upper value,
// and changing the following segments to be full-range.
func (section WrappedAddressSection) ToBlock(segmentIndex int, lower, upper SegInt) ExtendedSegmentSeries {
	return WrapSection(section.AddressSection.ToBlock(segmentIndex, lower, upper))
}

// ToPrefixBlock returns the series with the same prefix as this series while the remaining bits span all values.
// The series will be the block of all series with the same prefix.
//
// If this series has no prefix, this series is returned.
func (section WrappedAddressSection) ToPrefixBlock() ExtendedSegmentSeries {
	return WrapSection(section.AddressSection.ToPrefixBlock())
}

// ToPrefixBlockLen returns the series with the same prefix of the given length as this series while the remaining bits span all values.
// The returned series will be the block of all series with the same prefix.
func (section WrappedAddressSection) ToPrefixBlockLen(prefLen BitCount) ExtendedSegmentSeries {
	return WrapSection(section.AddressSection.ToPrefixBlockLen(prefLen))
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
func (section WrappedAddressSection) Increment(i int64) ExtendedSegmentSeries {
	return convSectToIntf(section.AddressSection.Increment(i))
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
func (section WrappedAddressSection) IncrementBoundary(i int64) ExtendedSegmentSeries {
	return convSectToIntf(section.AddressSection.IncrementBoundary(i))
}

// GetLower returns the series in the range with the lowest numeric value,
// which will be the same series if it represents a single value.
func (section WrappedAddressSection) GetLower() ExtendedSegmentSeries {
	return WrapSection(section.AddressSection.GetLower())
}

// GetUpper returns the series in the range with the highest numeric value,
// which will be the same series if it represents a single value.
func (section WrappedAddressSection) GetUpper() ExtendedSegmentSeries {
	return WrapSection(section.AddressSection.GetUpper())
}

func (section WrappedAddressSection) GetSection() *AddressSection {
	return section.AddressSection
}

func (section WrappedAddressSection) AssignPrefixForSingleBlock() ExtendedSegmentSeries {
	return convSectToIntf(section.AddressSection.AssignPrefixForSingleBlock())
}

func (section WrappedAddressSection) AssignMinPrefixForBlock() ExtendedSegmentSeries {
	return WrapSection(section.AddressSection.AssignMinPrefixForBlock())
}

func (section WrappedAddressSection) WithoutPrefixLen() ExtendedSegmentSeries {
	return WrapSection(section.AddressSection.WithoutPrefixLen())
}

// Contains returns whether this is same type and version as the given address series and whether it contains all values in the given series.
//
// Series must also have the same number of segments to be comparable, otherwise false is returned.
func (section WrappedAddressSection) Contains(other ExtendedSegmentSeries) bool {
	s, ok := other.Unwrap().(AddressSectionType)
	return ok && section.AddressSection.Contains(s)
}

// CompareSize compares the counts of two address series, the number of individual series represented in each.
//
// Rather than calculating counts with GetCount, there can be more efficient ways of comparing whether one series represents more individual address series than another.
//
// CompareSize returns a positive integer if this address series has a larger count than the one given, 0 if they are the same, or a negative integer if the other has a larger count.
func (section WrappedAddressSection) CompareSize(other ExtendedSegmentSeries) int {
	if s, ok := other.Unwrap().(AddressSectionType); ok {
		return section.AddressSection.CompareSize(s)
	}
	return section.GetCount().Cmp(other.GetCount())
}

func (section WrappedAddressSection) Equal(other ExtendedSegmentSeries) bool {
	s, ok := other.Unwrap().(AddressSectionType)
	return ok && section.AddressSection.Equal(s)
}

func (section WrappedAddressSection) SetPrefixLen(prefixLen BitCount) ExtendedSegmentSeries {
	return WrapSection(section.AddressSection.SetPrefixLen(prefixLen))
}

func (section WrappedAddressSection) SetPrefixLenZeroed(prefixLen BitCount) (ExtendedSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapSectWithErr(section.AddressSection.SetPrefixLenZeroed(prefixLen))
}

func (section WrappedAddressSection) AdjustPrefixLen(adjustment BitCount) ExtendedSegmentSeries {
	return WrapSection(section.AddressSection.AdjustPrefixLen(adjustment))
}

func (section WrappedAddressSection) AdjustPrefixLenZeroed(adjustment BitCount) (ExtendedSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapSectWithErr(section.AddressSection.AdjustPrefixLenZeroed(adjustment))
}

// ReverseBytes returns a new segment series with the bytes reversed.  Any prefix length is dropped.
//
// If each segment is more than 1 byte long, and the bytes within a single segment cannot be reversed because the segment represents a range,
// and reversing the segment values results in a range that is not contiguous, then this returns an error.
//
// In practice this means that to be reversible, a range must include all values except possibly the largest and/or smallest, which reverse to themselves.
func (section WrappedAddressSection) ReverseBytes() (ExtendedSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapSectWithErr(section.AddressSection.ReverseBytes())
}

// ReverseBits returns a new segment series with the bits reversed.  Any prefix length is dropped.
//
// If the bits within a single segment cannot be reversed because the segment represents a range,
// and reversing the segment values results in a range that is not contiguous, this returns an error.
//
// In practice this means that to be reversible, a range must include all values except possibly the largest and/or smallest, which reverse to themselves.
//
// If perByte is true, the bits are reversed within each byte, otherwise all the bits are reversed.
func (section WrappedAddressSection) ReverseBits(perByte bool) (ExtendedSegmentSeries, addrerr.IncompatibleAddressError) {
	return wrapSectWithErr(section.AddressSection.ReverseBits(perByte))
}

// ReverseSegments returns a new series with the segments reversed.
func (section WrappedAddressSection) ReverseSegments() ExtendedSegmentSeries {
	return WrapSection(section.AddressSection.ReverseSegments())
}

var _ ExtendedSegmentSeries = WrappedAddress{}
var _ ExtendedSegmentSeries = WrappedAddressSection{}

// In go, a nil value is not coverted to a nil interface, it is converted to a non-nil interface instance with underlying value nil
func convAddrToIntf(addr *Address) ExtendedSegmentSeries {
	if addr == nil {
		return nil
	}
	return WrapAddress(addr)
}

func convSectToIntf(sect *AddressSection) ExtendedSegmentSeries {
	if sect == nil {
		return nil
	}
	return WrapSection(sect)
}

func wrapSectWithErr(section *AddressSection, err addrerr.IncompatibleAddressError) (ExtendedSegmentSeries, addrerr.IncompatibleAddressError) {
	if err == nil {
		return WrapSection(section), nil
	}
	return nil, err
}

func wrapAddrWithErr(addr *Address, err addrerr.IncompatibleAddressError) (ExtendedSegmentSeries, addrerr.IncompatibleAddressError) {
	if err == nil {
		return WrapAddress(addr), nil
	}
	return nil, err
}

func WrapAddress(addr *Address) WrappedAddress {
	return WrappedAddress{addr}
}

func WrapSection(section *AddressSection) WrappedAddressSection {
	return WrappedAddressSection{section}
}
