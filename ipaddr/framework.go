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

	"github.com/seancfoley/ipaddress-go/ipaddr/addrerr"
)

// AddressItem represents all addresses, division groupings, divisions, and sequential ranges.
// Any address item can be compared to any other.
type AddressItem interface {
	// GetValue returns the lowest individual address item in the address item range as an integer value
	GetValue() *big.Int

	// GetUpperValue returns the highest individual address item in the address item range as an integer value
	GetUpperValue() *big.Int

	// CopyBytes copies the value of the lowest individual address item in this address item range into a byte slice.
	//
	// If the value can fit in the given slice, the value is copied into that slice and a length-adjusted sub-slice is returned.
	// Otherwise, a new slice is created and returned with the value.
	CopyBytes(bytes []byte) []byte

	// CopyUpperBytes copies the value of the highest individual address item in this address item range into a byte slice.
	//
	// If the value can fit in the given slice, the value is copied into that slice and a length-adjusted sub-slice is returned.
	// Otherwise, a new slice is created and returned with the value.
	CopyUpperBytes(bytes []byte) []byte

	// Bytes returns the lowest individual address item in the address item range as a byte slice
	Bytes() []byte

	// UpperBytes returns the highest individual address item in the address item range as a byte slice
	UpperBytes() []byte

	// GetCount provides the number of address items represented by this AddressItem, for example the subnet size for IP addresses
	GetCount() *big.Int

	// IsMultiple returns whether this item represents multiple values (the count is larger than 1)
	IsMultiple() bool

	// GetByteCount returns the number of bytes required for each value comprising this address item,
	// rounding up if the bit count is not a multiple of 8.
	GetByteCount() int

	// GetBitCount returns the number of bits in each value comprising this address item
	GetBitCount() BitCount

	// IsFullRange returns whether this address item represents all possible values attainable by an address item of this type.
	//
	// This is true if and only if both IncludesZero and IncludesMax return true.
	IsFullRange() bool

	// IncludesZero returns whether this item includes the value of zero within its range
	IncludesZero() bool

	// IncludesMax returns whether this item includes the max value, the value whose bits are all ones, within its range
	IncludesMax() bool

	// IsZero returns whether this address item matches exactly the value of zero
	IsZero() bool

	// IsMax returns whether this address item matches exactly the maximum possible value, the value whose bits are all ones
	IsMax() bool

	// ContainsPrefixBlock returns whether the values of this item contains the prefix block for the given prefix length.
	// Unlike ContainsSinglePrefixBlock, whether there are multiple prefix values for the given prefix length makes no difference.
	ContainsPrefixBlock(BitCount) bool

	// ContainsSinglePrefixBlock returns whether the values of this series contains a single prefix block for the given prefix length.
	// This means there is only one prefix of the given length in this item, and this item contains the prefix block for that given prefix, all items with that same prefix.
	ContainsSinglePrefixBlock(BitCount) bool

	// GetPrefixLenForSingleBlock returns a prefix length for which there is only one prefix of that length in this item,
	// and the range of this item matches the block of all values for that prefix.
	//
	// If the entire range can be described this way, then this method returns the same value as GetMinPrefixLenForBlock.
	//
	// If no such prefix length exists, returns nil.
	//
	// If this item represents a single value, this returns the bit count.
	GetPrefixLenForSingleBlock() PrefixLen

	// GetMinPrefixLenForBlock returns the smallest prefix length possible such that this item includes the block of all values for that prefix length.
	//
	// If the entire range can be dictated this way, then this method returns the same value as GetPrefixLenForSingleBlock.
	//
	// There may be a single prefix, or multiple possible prefix values in this item for the returned prefix length.
	// Use GetPrefixLenForSingleBlock to avoid the case of multiple prefix values.
	//
	// If this item represents a single value, this returns the bit count.
	GetMinPrefixLenForBlock() BitCount

	// GetPrefixCountLen returns the count of the number of distinct values within the prefix part of the range of values for this item
	GetPrefixCountLen(BitCount) *big.Int

	// Compare returns a negative integer, zero, or a positive integer if this address item is less than, equal, or greater than the given item.
	// Any address item is comparable to any other.  All address items use CountComparator to compare.
	Compare(item AddressItem) int

	fmt.Stringer
}

// AddressComponent represents all addresses, address sections, and address segments
type AddressComponent interface { //AddressSegment and above, AddressSegmentSeries and above
	// TestBit returns true if the bit in the lower value of the address component at the given index is 1, where index 0 refers to the least significant bit.
	// In other words, it computes (bits & (1 << n)) != 0), using the lower value of this address component.
	// TestBit will panic if n < 0, or if it matches or exceeds the bit count of this address component.
	TestBit(index BitCount) bool

	// IsOneBit returns true if the bit in the lower value of this address component at the given index is 1, where index 0 refers to the most significant bit.
	// IsOneBit will panic if bitIndex < 0, or if it is larger than the bit count of this address component.
	IsOneBit(index BitCount) bool

	// ToHexString writes this address component as a single hexadecimal value (possibly two values if a range that is not a prefixed block),
	// the number of digits according to the bit count, with or without a preceding "0x" prefix.
	//
	// If a multiple-valued component cannot be written as a single prefix block or a range of two values, an error is returned.
	ToHexString(bool) (string, addrerr.IncompatibleAddressError)

	// ToNormalizedString produces a string that is consistent for all address components of the same type and version,
	ToNormalizedString() string
}

// StandardDivGroupingType represents any standard division grouping (division groupings or address sections where all divisions are 64 bits or less)
// including AddressSection, IPAddressSection, IPv4AddressSection, IPv6AddressSection, MACAddressSection, and AddressDivisionGrouping
type StandardDivGroupingType interface {
	AddressDivisionSeries

	// IsAdaptiveZero returns true if the division grouping was originally created as an implicitly zero-valued section or grouping (eg IPv4AddressSection{}),
	// meaning it was not constructed using a constructor function.
	// Such a grouping, which has no divisions or segments, is convertible to an implicitly zero-valued grouping of any type or version, whether IPv6, IPv4, MAC, etc
	IsAdaptiveZero() bool

	// CompareSize compares the counts of two division groupings, the number of individual division groupings within.
	//
	// Rather than calculating counts with GetCount, there can be more efficient ways of comparing whether one grouping represents more individual groupings than another.
	//
	// CompareSize returns a positive integer if this address division grouping has a larger count than the one given, 0 if they are the same, or a negative integer if the other has a larger count.
	CompareSize(StandardDivGroupingType) int

	// ToDivGrouping converts to an AddressDivisionGrouping, a polymorphic type usable with all address sections and division groupings.
	//
	// ToDivGrouping implementations can be called with a nil receiver, enabling you to chain this method with methods that might return a nil pointer.
	ToDivGrouping() *AddressDivisionGrouping
}

var _, _ StandardDivGroupingType = &AddressDivisionGrouping{},
	&IPv6v4MixedAddressGrouping{}

// AddressDivisionSeries serves as a common interface to all division groupings and addresses
type AddressDivisionSeries interface {
	AddressItem

	// GetDivisionCount returns the number of divisions
	GetDivisionCount() int

	// GetPrefixCount returns the count of prefixes in this series for its prefix length, or the total count if it has no prefix length
	GetPrefixCount() *big.Int

	// GetBlockCount returns the count of distinct values in the given number of initial (more significant) segments.
	GetBlockCount(divisionCount int) *big.Int

	// GetSequentialBlockIndex gets the minimal division index for which all following divisions are full-range blocks.
	//
	// The division at this index is not a full-range block unless all divisions are full-range.
	// The division at this index and all following divisions form a sequential range.
	// For the full series to be sequential, the preceding divisions must be single-valued.
	GetSequentialBlockIndex() int

	// GetSequentialBlockCount provides the count of elements from the sequential block iterator, the minimal number of sequential address division series that comprise this address division series
	GetSequentialBlockCount() *big.Int

	// IsSequential returns  whether the series represents a range of values that are sequential.
	//
	// Generally, this means that any division covering a range of values must be followed by divisions that are full range, covering all values.
	IsSequential() bool

	// IsPrefixBlock returns whether this address division series has a prefix length and includes the block associated with its prefix length.
	//
	// This is different from ContainsPrefixBlock in that this method returns
	// false if the series has no prefix length or a prefix length that differs from prefix lengths for which ContainsPrefixBlock returns true.
	IsPrefixBlock() bool

	// IsSinglePrefixBlock returns whether the range of values matches a single subnet block for the prefix length.
	//
	// This is different from ContainsSinglePrefixBlock in that this method returns
	// false if this series has no prefix length or a prefix length that differs from the prefix lengths for which ContainsSinglePrefixBlock returns true.
	IsSinglePrefixBlock() bool

	// IsPrefixed returns whether this address has an associated prefix length
	IsPrefixed() bool

	// GetPrefixLen returns the prefix length, or nil if there is no prefix length.
	//
	// A prefix length indicates the number of bits in the initial part (most significant bits) of the series that comprise the prefix.
	//
	// A prefix is a part of the series that is not specific to that series but common amongst a group, such as a CIDR prefix block subnet.
	GetPrefixLen() PrefixLen

	// GetGenericDivision returns the division at the given index as a DivisionType.
	// The first division is at index 0.
	// GetGenericDivision will panic given a negative index or index larger than the division count.
	GetGenericDivision(index int) DivisionType // useful for comparisons
}

// AddressSegmentSeries serves as a common interface to all address sections and addresses
type AddressSegmentSeries interface { // Address and above, AddressSection and above, IPAddressSegmentSeries, ExtendedIPSegmentSeries
	AddressComponent

	AddressDivisionSeries

	// GetMaxSegmentValue returns the maximum possible segment value for this type of series.
	//
	// Note this is not the maximum of the range of segment values in this specific series,
	// this is the maximum value of any segment for this series type and version, determined by the number of bits per segment.
	GetMaxSegmentValue() SegInt

	// GetSegmentCount returns the number of segments, which is the same as the division count since the segments are also the divisions
	GetSegmentCount() int

	// GetBitsPerSegment returns the number of bits comprising each segment in this series.  Segments in the same series are equal length.
	GetBitsPerSegment() BitCount

	// GetBytesPerSegment returns the number of bytes comprising each segment in this series.  Segments in the same series are equal length.
	GetBytesPerSegment() int

	// ToCanonicalString produces a canonical string for the address series.
	//
	// For IPv4, dotted octet format, also known as dotted decimal format, is used.
	// https://datatracker.ietf.org/doc/html/draft-main-ipaddr-text-rep-00#section-2.1
	//
	// For IPv6, RFC 5952 describes the canonical string representation.
	// https://en.wikipedia.org/wiki/IPv6_address#Representation
	// http://tools.ietf.org/html/rfc5952
	//
	// For MAC, it uses the canonical standardized IEEE 802 MAC address representation of xx-xx-xx-xx-xx-xx.  An example is "01-23-45-67-89-ab".
	// For range segments, '|' is used: 11-22-33|44-55-66
	//
	// Each address has a unique canonical string, not counting the prefix length.
	// With IP addresses and sections, the prefix length is included in the string, and the prefix length can cause two equal addresses to have different strings, for example "1.2.3.4/16" and "1.2.3.4".
	// It can also cause two different addresses to have the same string, such as "1.2.0.0/16" for the individual address "1.2.0.0" and also the prefix block "1.2.*.*".
	ToCanonicalString() string

	// ToCompressedString produces a short representation of this series while remaining within the confines of standard representation(s) of the series.
	//
	// For IPv4, it is the same as the canonical string.
	//
	// For IPv6, it differs from the canonical string.  It compresses the maximum number of zeros and/or host segments with the IPv6 compression notation '::'.
	//
	// For MAC, it differs from the canonical string.  It produces a shorter string for the address that has no leading zeros.
	ToCompressedString() string

	// ToBinaryString writes this address series as a single binary value (possibly two values if a range that is not a prefixed block),
	// the number of digits according to the bit count, with or without a preceding "0b" prefix.
	//
	// If a multiple-valued series cannot be written as a single prefix block or a range of two values, an error is returned.
	ToBinaryString(with0bPrefix bool) (string, addrerr.IncompatibleAddressError)

	// ToOctalString writes this address series as a single octal value (possibly two values if a range that is not a prefixed block),
	// the number of digits according to the bit count, with or without a preceding "0" prefix.
	//
	// If a multiple-valued series cannot be written as a single prefix block or a range of two values, an error is returned.
	ToOctalString(withPrefix bool) (string, addrerr.IncompatibleAddressError)

	// GetSegmentStrings returns an array with the strings of each segment being the string that is normalized with wildcards.
	GetSegmentStrings() []string

	// GetGenericSegment returns the segment at the given index as an AddressSegmentType.
	// The first segment is at index 0.
	// GetGenericSegment will panic given a negative index or index larger than the segment count.
	GetGenericSegment(index int) AddressSegmentType
}

var _, _ AddressSegmentSeries = &Address{}, &AddressSection{}

// IPAddressSegmentSeries serves as a common interface to all IP address sections and IP addresses
type IPAddressSegmentSeries interface { // IPAddress and above, IPAddressSection and above, ExtendedIPSegmentSeries
	AddressSegmentSeries

	IncludesZeroHost() bool
	IncludesZeroHostLen(prefLen BitCount) bool
	IncludesMaxHost() bool
	IncludesMaxHostLen(prefLen BitCount) bool
	IsZeroHost() bool
	IsZeroHostLen(BitCount) bool
	IsMaxHost() bool
	IsMaxHostLen(BitCount) bool
	IsSingleNetwork() bool

	GetIPVersion() IPVersion

	GetBlockMaskPrefixLen(network bool) PrefixLen

	// GetLeadingBitCount returns the number of consecutive leading one or zero bits.
	// If ones is true, returns the number of consecutive leading one bits.
	// Otherwise, returns the number of consecutive leading zero bits.
	//
	// This method applies to the lower value of the range if this series represents multiple values.
	GetLeadingBitCount(ones bool) BitCount

	// GetTrailingBitCount returns the number of consecutive trailing one or zero bits.
	// If ones is true, returns the number of consecutive trailing zero bits.
	// Otherwise, returns the number of consecutive trailing one bits.
	//
	// This method applies to the lower value of the range if this series represents multiple values.
	GetTrailingBitCount(ones bool) BitCount

	ToFullString() string
	ToPrefixLenString() string
	ToSubnetString() string
	ToNormalizedWildcardString() string
	ToCanonicalWildcardString() string
	ToCompressedWildcardString() string
	ToSegmentedBinaryString() string
	ToSQLWildcardString() string
	ToReverseDNSString() (string, addrerr.IncompatibleAddressError)
}

var _, _ IPAddressSegmentSeries = &IPAddress{}, &IPAddressSection{}

type IPv6AddressSegmentSeries interface {
	IPAddressSegmentSeries

	// GetTrailingSection returns an ending subsection of the full address or address section
	GetTrailingSection(index int) *IPv6AddressSection

	// GetSubSection returns a subsection of the full address or address section
	GetSubSection(index, endIndex int) *IPv6AddressSection

	GetNetworkSection() *IPv6AddressSection
	GetHostSection() *IPv6AddressSection
	GetNetworkSectionLen(BitCount) *IPv6AddressSection
	GetHostSectionLen(BitCount) *IPv6AddressSection

	// GetSegments returns a slice with the address segments.  The returned slice is not backed by the same array as the receiver.
	GetSegments() []*IPv6AddressSegment

	// CopySegments copies the existing segments into the given slice,
	// as much as can be fit into the slice, returning the number of segments copied
	CopySegments(segs []*IPv6AddressSegment) (count int)

	// CopySubSegments copies the existing segments from the given start index until but not including the segment at the given end index,
	// into the given slice, as much as can be fit into the slice, returning the number of segments copied
	CopySubSegments(start, end int, segs []*IPv6AddressSegment) (count int)

	// GetSegment returns the segment at the given index.
	// The first segment is at index 0.
	// GetSegment will panic given a negative index or index larger than the segment count.
	GetSegment(index int) *IPv6AddressSegment
}

var _, _, _ IPv6AddressSegmentSeries = &IPv6Address{},
	&IPv6AddressSection{},
	&EmbeddedIPv6AddressSection{}

type IPv4AddressSegmentSeries interface {
	IPAddressSegmentSeries

	// GetTrailingSection returns an ending subsection of the full address section
	GetTrailingSection(index int) *IPv4AddressSection

	// GetSubSection returns a subsection of the full address section
	GetSubSection(index, endIndex int) *IPv4AddressSection

	GetNetworkSection() *IPv4AddressSection
	GetHostSection() *IPv4AddressSection
	GetNetworkSectionLen(BitCount) *IPv4AddressSection
	GetHostSectionLen(BitCount) *IPv4AddressSection

	// GetSegments returns a slice with the address segments.  The returned slice is not backed by the same array as the receiver.
	GetSegments() []*IPv4AddressSegment

	// CopySegments copies the existing segments into the given slice,
	// as much as can be fit into the slice, returning the number of segments copied
	CopySegments(segs []*IPv4AddressSegment) (count int)

	// CopySubSegments copies the existing segments from the given start index until but not including the segment at the given end index,
	// into the given slice, as much as can be fit into the slice, returning the number of segments copied
	CopySubSegments(start, end int, segs []*IPv4AddressSegment) (count int)

	// GetSegment returns the segment at the given index.
	// The first segment is at index 0.
	// GetSegment will panic given a negative index or index larger than the segment count.
	GetSegment(index int) *IPv4AddressSegment
}

var _, _ IPv4AddressSegmentSeries = &IPv4Address{}, &IPv4AddressSection{}

type MACAddressSegmentSeries interface {
	AddressSegmentSeries

	// GetTrailingSection returns an ending subsection of the full address section
	GetTrailingSection(index int) *MACAddressSection

	// GetSubSection returns a subsection of the full address section
	GetSubSection(index, endIndex int) *MACAddressSection

	// GetSegments returns a slice with the address segments.  The returned slice is not backed by the same array as the receiver.
	GetSegments() []*MACAddressSegment

	// CopySegments copies the existing segments into the given slice,
	// as much as can be fit into the slice, returning the number of segments copied
	CopySegments(segs []*MACAddressSegment) (count int)

	// CopySubSegments copies the existing segments from the given start index until but not including the segment at the given end index,
	// into the given slice, as much as can be fit into the slice, returning the number of segments copied
	CopySubSegments(start, end int, segs []*MACAddressSegment) (count int)

	// GetSegment returns the segment at the given index.
	// The first segment is at index 0.
	// GetSegment will panic given a negative index or index larger than the segment count.
	GetSegment(index int) *MACAddressSegment
}

var _, _ MACAddressSegmentSeries = &MACAddress{}, &MACAddressSection{}

// AddressSectionType represents any address section
// that can be converted to/from the base type AddressSection,
// including AddressSection, IPAddressSection, IPv4AddressSection, IPv6AddressSection, and MACAddressSection
type AddressSectionType interface {
	StandardDivGroupingType

	// Equal returns whether the given address section is equal to this address section.
	// Two address sections are equal if they represent the same set of sections.
	// They must match:
	//  - type/version (IPv4, IPv6, MAC, etc)
	//  - segment counts
	//  - bits per segment
	//  - segment value ranges
	// Prefix lengths are ignored.
	Equal(AddressSectionType) bool

	// Contains returns whether this is same type and version as the given address section and whether it contains all values in the given section.
	//
	// Sections must also have the same number of segments to be comparable, otherwise false is returned.
	Contains(AddressSectionType) bool

	// PrefixEqual determines if the given section matches this section up to the prefix length of this section.
	// It returns whether the argument section has the same address section prefix values as this.
	//
	// The entire prefix of this section must be present in the other section to be comparable.
	PrefixEqual(AddressSectionType) bool

	// PrefixContains returns whether the prefix values in the given address section
	// are prefix values in this address section, using the prefix length of this section.
	// If this address section has no prefix length, the entire address is compared.
	//
	// It returns whether the prefix of this address contains all values of the same prefix length in the given address.
	//
	// All prefix bits of this section must be present in the other section to be comparable.
	PrefixContains(AddressSectionType) bool

	// ToSectionBase converts to an AddressSection, a polymorphic type usable with all address sections.
	//
	// ToSectionBase implementations can be called with a nil receiver, enabling you to chain this method with methods that might return a nil pointer.
	ToSectionBase() *AddressSection
}

//Note: if we had an IPAddressSectionType we could add Wrap() WrappedIPAddressSection to it, but I guess not much else

var _, _, _, _, _ AddressSectionType = &AddressSection{},
	&IPAddressSection{},
	&IPv4AddressSection{},
	&IPv6AddressSection{},
	&MACAddressSection{}

// AddressType represents any address, all of which can be represented by the base type Address.
// This includes IPAddress, IPv4Address, IPv6Address, and MACAddress.
// It can be useful as a parameter for functions to take any address type, while inside the function you can convert to *Address using ToAddressBase
type AddressType interface {
	AddressSegmentSeries

	// Equal returns whether the given address or subnet is equal to this address or subnet.
	// Two address instances are equal if they represent the same set of addresses.
	Equal(AddressType) bool

	// Contains returns whether this is same type and version as the given address or subnet and whether it contains all addresses in the given address or subnet.
	Contains(AddressType) bool

	// CompareSize compares the counts of two subnets, the number of individual addresses within.
	//
	// Rather than calculating counts with GetCount, there can be more efficient ways of comparing whether one subnet represents more individual addresses than another.
	//
	// CompareSize returns a positive integer if this address has a larger count than the one given, 0 if they are the same, or a negative integer if the other has a larger count.
	CompareSize(AddressType) int

	// PrefixEqual determines if the given address matches this address up to the prefix length of this address.
	// If this address has no prefix length, the entire address is compared.
	//
	// It returns whether the two addresses share the same range of prefix values.
	PrefixEqual(AddressType) bool

	// PrefixContains returns whether the prefix values in the given address or subnet
	// are prefix values in this address or subnet, using the prefix length of this address or subnet.
	// If this address has no prefix length, the entire address is compared.
	//
	// It returns whether the prefix of this address contains all values of the same prefix length in the given address.
	PrefixContains(AddressType) bool

	// ToAddressBase converts to an Address, a polymorphic type usable with all addresses and subnets.
	//
	// ToAddressBase implementations can be called with a nil receiver, enabling you to chain this method with methods that might return a nil pointer.
	ToAddressBase() *Address
}

var _, _ AddressType = &Address{}, &MACAddress{}

type ipAddressRange interface {
	// GetLowerIPAddress returns the address in the subnet or address collection with the lowest numeric value,
	// which will be the same address if it represents a single value.
	// For example, for "1.2-3.4.5-6", the series "1.2.4.5" is returned.
	GetLowerIPAddress() *IPAddress

	// GetUpperIPAddress returns the address in the subnet or address collection with the highest numeric value,
	// which will be the same address if it represents a single value.
	// For example, for "1.2-3.4.5-6", the series "1.3.4.6" is returned.
	GetUpperIPAddress() *IPAddress

	CopyNetIP(bytes net.IP) net.IP
	CopyUpperNetIP(bytes net.IP) net.IP

	GetNetIP() net.IP
	GetUpperNetIP() net.IP
}

// IPAddressRange represents all IPAddress instances and all IPAddress sequential range instances
type IPAddressRange interface { //IPAddress and above, IPAddressSeqRange and above
	AddressItem

	ipAddressRange

	// IsSequential returns whether the address item represents a range of addresses that are sequential.
	//
	// IP Address sequential ranges are sequential by definition.
	//
	// Generally, for a subnet this means that any segment covering a range of values must be followed by segments that are full range, covering all values.
	//
	// Individual addresses are sequential and CIDR prefix blocks are sequential.
	// The subnet 1.2.3-4.5 is not sequential, since the two addresses it represents, 1.2.3.5 and 1.2.4.5, are not (1.2.3.6 is in-between the two but not in the subnet).
	IsSequential() bool
}

var _, _, _, _, _, _ IPAddressRange = &IPAddress{}, &IPv4Address{}, &IPv6Address{}, &IPAddressSeqRange{},
	&IPv4AddressSeqRange{},
	&IPv6AddressSeqRange{}

// IPAddressType represents any IP address, all of which can be represented by the base type IPAddress.
// This includes IPv4Address and IPv6Address.
type IPAddressType interface {
	AddressType

	ipAddressRange

	// Wrap wraps this IP address, returning a WrappedIPAddress, an implementation of ExtendedIPSegmentSeries,
	// which can be used to write code that works with both IP addresses and IP address sections.
	Wrap() WrappedIPAddress

	// ToIP converts to an IPAddress, a polymorphic type usable with all IP addresses and subnets.
	//
	// ToIP can be called with a nil receiver, enabling you to chain this method with methods that might return a nil pointer.
	ToIP() *IPAddress

	ToAddressString() *IPAddressString
}

var _, _, _ IPAddressType = &IPAddress{},
	&IPv4Address{},
	&IPv6Address{}

// IPAddressSeqRangeType represents any IP address sequential range, all of which can be represented by the base type IPAddressSeqRange.
// This includes IPv4AddressSeqRange and IPv6AddressSeqRange.
type IPAddressSeqRangeType interface {
	AddressItem

	ipAddressRange

	// CompareSize compares the counts of two subnets, the number of individual addresses within.
	//
	// Rather than calculating counts with GetCount, there can be more efficient ways of comparing whether one subnet represents more individual addresses than another.
	//
	// CompareSize returns a positive integer if this address has a larger count than the one given, 0 if they are the same, or a negative integer if the other has a larger count.
	CompareSize(IPAddressSeqRangeType) int

	// ContainsRange returns whether all the addresses in the given sequential range are also contained in this sequential range.
	ContainsRange(IPAddressSeqRangeType) bool

	// Contains returns whether this range contains all IP addresses in the given address or subnet.
	Contains(IPAddressType) bool

	// Equal returns whether the given sequential address range is equal to this sequential address range.
	// Two sequential address ranges are equal if their lower and upper range boundaries are equal.
	Equal(IPAddressSeqRangeType) bool

	// ToCanonicalString produces a canonical string for the address range.
	// It has the format "lower -> upper" where lower and upper are the canonical strings for the lowest and highest addresses in the range, given by GetLower and GetUpper.
	ToCanonicalString() string

	// ToNormalizedString produces a normalized string for the address range.
	// It has the format "lower -> upper" where lower and upper are the normalized strings for the lowest and highest addresses in the range, given by GetLower and GetUpper.
	ToNormalizedString() string

	// ToIP converts to an IPAddressSeqRange, a polymorphic type usable with all IP address sequential ranges.
	//
	// ToIP can be called with a nil receiver, enabling you to chain this method with methods that might return a nil pointer.
	ToIP() *IPAddressSeqRange
}

var _, _, _ IPAddressSeqRangeType = &IPAddressSeqRange{},
	&IPv4AddressSeqRange{},
	&IPv6AddressSeqRange{}

// HostIdentifierString represents a string that is used to identify a host.
type HostIdentifierString interface {

	// ToNormalizedString provides a normalized String representation for the host identified by this HostIdentifierString instance
	ToNormalizedString() string

	// IsValid returns whether the wrapped string is a valid identifier for a host
	IsValid() bool

	fmt.Stringer
}

var _, _, _ HostIdentifierString = &IPAddressString{}, &MACAddressString{}, &HostName{}
