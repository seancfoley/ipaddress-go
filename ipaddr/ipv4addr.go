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
	"github.com/seancfoley/ipaddress-go/ipaddr/addrstr"
)

const (
	IPv4SegmentSeparator      = '.'
	IPv4SegmentSeparatorStr   = "."
	IPv4BitsPerSegment        = 8
	IPv4BytesPerSegment       = 1
	IPv4SegmentCount          = 4
	IPv4ByteCount             = 4
	IPv4BitCount              = 32
	IPv4DefaultTextualRadix   = 10
	IPv4MaxValuePerSegment    = 0xff
	IPv4MaxValue              = 0xffffffff
	IPv4ReverseDnsSuffix      = ".in-addr.arpa"
	IPv4SegmentMaxChars       = 3
	ipv4BitsToSegmentBitshift = 3
)

func newIPv4Address(section *IPv4AddressSection) *IPv4Address {
	return createAddress(section.ToSectionBase(), NoZone).ToIPv4()
}

func NewIPv4Address(section *IPv4AddressSection) (*IPv4Address, addrerr.AddressValueError) {
	if section == nil {
		return zeroIPv4, nil
	}
	segCount := section.GetSegmentCount()
	if segCount != IPv4SegmentCount {
		return nil, &addressValueError{
			addressError: addressError{key: "ipaddress.error.invalid.size"},
			val:          segCount,
		}
	}
	return createAddress(section.ToSectionBase(), NoZone).ToIPv4(), nil
}

func NewIPv4AddressFromSegs(segments []*IPv4AddressSegment) (*IPv4Address, addrerr.AddressValueError) {
	segCount := len(segments)
	if segCount != IPv4SegmentCount {
		return nil, &addressValueError{
			addressError: addressError{key: "ipaddress.error.invalid.size"},
			val:          segCount,
		}
	}
	section := NewIPv4Section(segments)
	return createAddress(section.ToSectionBase(), NoZone).ToIPv4(), nil
}

func NewIPv4AddressFromPrefixedSegs(segments []*IPv4AddressSegment, prefixLength PrefixLen) (*IPv4Address, addrerr.AddressValueError) {
	segCount := len(segments)
	if segCount != IPv4SegmentCount {
		return nil, &addressValueError{
			addressError: addressError{key: "ipaddress.error.invalid.size"},
			val:          segCount,
		}
	}
	section := NewIPv4PrefixedSection(segments, prefixLength)
	return createAddress(section.ToSectionBase(), NoZone).ToIPv4(), nil
}

func NewIPv4AddressFromUint32(val uint32) *IPv4Address {
	section := NewIPv4SectionFromUint32(val, IPv4SegmentCount)
	return createAddress(section.ToSectionBase(), NoZone).ToIPv4()
}

func NewIPv4AddressFromPrefixedUint32(val uint32, prefixLength PrefixLen) *IPv4Address {
	section := NewIPv4SectionFromPrefixedUint32(val, IPv4SegmentCount, prefixLength)
	return createAddress(section.ToSectionBase(), NoZone).ToIPv4()
}

func NewIPv4AddressFromBytes(bytes []byte) (addr *IPv4Address, err addrerr.AddressValueError) {
	if ipv4 := net.IP(bytes).To4(); ipv4 != nil {
		bytes = ipv4
	}
	section, err := NewIPv4SectionFromSegmentedBytes(bytes, IPv4SegmentCount)
	if err == nil {
		addr = newIPv4Address(section)
	}
	return
}

func NewIPv4AddressFromPrefixedBytes(bytes []byte, prefixLength PrefixLen) (addr *IPv4Address, err addrerr.AddressValueError) {
	if ipv4 := net.IP(bytes).To4(); ipv4 != nil {
		bytes = ipv4
	}
	section, err := NewIPv4SectionFromPrefixedBytes(bytes, IPv4SegmentCount, prefixLength)
	if err == nil {
		addr = newIPv4Address(section)
	}
	return
}

func NewIPv4AddressFromVals(vals IPv4SegmentValueProvider) *IPv4Address {
	section := NewIPv4SectionFromVals(vals, IPv4SegmentCount)
	return newIPv4Address(section)
}

func NewIPv4AddressFromPrefixedVals(vals IPv4SegmentValueProvider, prefixLength PrefixLen) *IPv4Address {
	section := NewIPv4SectionFromPrefixedVals(vals, IPv4SegmentCount, prefixLength)
	return newIPv4Address(section)
}

func NewIPv4AddressFromRange(vals, upperVals IPv4SegmentValueProvider) *IPv4Address {
	section := NewIPv4SectionFromRange(vals, upperVals, IPv4SegmentCount)
	return newIPv4Address(section)
}

func NewIPv4AddressFromPrefixedRange(vals, upperVals IPv4SegmentValueProvider, prefixLength PrefixLen) *IPv4Address {
	section := NewIPv4SectionFromPrefixedRange(vals, upperVals, IPv4SegmentCount, prefixLength)
	return newIPv4Address(section)
}

func newIPv4AddressFromPrefixedSingle(vals, upperVals IPv4SegmentValueProvider, prefixLength PrefixLen) *IPv4Address {
	section := newIPv4SectionFromPrefixedSingle(vals, upperVals, IPv4SegmentCount, prefixLength, true)
	return newIPv4Address(section)
}

var zeroIPv4 = initZeroIPv4()
var ipv4All = zeroIPv4.ToPrefixBlockLen(0)

func initZeroIPv4() *IPv4Address {
	div := zeroIPv4Seg
	segs := []*IPv4AddressSegment{div, div, div, div}
	section := NewIPv4Section(segs)
	return newIPv4Address(section)
}

//
//
// IPv4Address is an IPv4 address, or a subnet of multiple IPv4 addresses.  Each segment can represent a single value or a range of values.
// The zero value is 0.0.0.0
type IPv4Address struct {
	ipAddressInternal
}

func (addr *IPv4Address) init() *IPv4Address {
	if addr.section == nil {
		return zeroIPv4
	}
	return addr
}

// GetCount returns the count of addresses that this address or subnet represents.
//
// If just a single address, not a subnet of multiple addresses, returns 1.
//
// For instance, the IP address subnet 1.2.0.0/16 has the count of 2 to the power of 16.
//
// Use IsMultiple if you simply want to know if the count is greater than 1.
func (addr *IPv4Address) GetCount() *big.Int {
	if addr == nil {
		return bigZero()
	}
	return addr.getCount()
}

// IsMultiple returns true if this represents more than a single individual address, whether it is a subnet of multiple addresses.
func (addr *IPv4Address) IsMultiple() bool {
	return addr != nil && addr.isMultiple()
}

// IsPrefixed returns whether this address has an associated prefix length
func (addr *IPv4Address) IsPrefixed() bool {
	return addr != nil && addr.isPrefixed()
}

// GetBitCount returns the number of bits comprising this address,
// or each address in the range if a subnet, which is 32.
func (addr *IPv4Address) GetBitCount() BitCount {
	return IPv4BitCount
}

// GetByteCount returns the number of bytes required for this address,
// or each address in the range if a subnet, which is 4.
func (addr *IPv4Address) GetByteCount() int {
	return IPv4ByteCount
}

// GetBitsPerSegment returns the number of bits comprising each segment in this address.  Segments in the same address are equal length.
func (addr *IPv4Address) GetBitsPerSegment() BitCount {
	return IPv4BitsPerSegment
}

// GetBytesPerSegment returns the number of bytes comprising each segment in this address or subnet.  Segments in the same address are equal length.
func (addr *IPv4Address) GetBytesPerSegment() int {
	return IPv4BytesPerSegment
}

func (addr *IPv4Address) GetSection() *IPv4AddressSection {
	return addr.init().section.ToIPv4()
}

// GetTrailingSection gets the subsection from the series starting from the given index
// The first segment is at index 0.
func (addr *IPv4Address) GetTrailingSection(index int) *IPv4AddressSection {
	return addr.GetSection().GetTrailingSection(index)
}

// GetSubSection gets the subsection from the series starting from the given index and ending just before the give endIndex
// The first segment is at index 0.
func (addr *IPv4Address) GetSubSection(index, endIndex int) *IPv4AddressSection {
	return addr.GetSection().GetSubSection(index, endIndex)
}

func (addr *IPv4Address) GetNetworkSection() *IPv4AddressSection {
	return addr.GetSection().GetNetworkSection()
}

func (addr *IPv4Address) GetNetworkSectionLen(prefLen BitCount) *IPv4AddressSection {
	return addr.GetSection().GetNetworkSectionLen(prefLen)
}

func (addr *IPv4Address) GetHostSection() *IPv4AddressSection {
	return addr.GetSection().GetHostSection()
}

func (addr *IPv4Address) GetHostSectionLen(prefLen BitCount) *IPv4AddressSection {
	return addr.GetSection().GetHostSectionLen(prefLen)
}

func (addr *IPv4Address) GetNetworkMask() *IPv4Address {
	return addr.getNetworkMask(ipv4Network).ToIPv4()
}

func (addr *IPv4Address) GetHostMask() *IPv4Address {
	return addr.getHostMask(ipv4Network).ToIPv4()
}

// CopySubSegments copies the existing segments from the given start index until but not including the segment at the given end index,
// into the given slice, as much as can be fit into the slice, returning the number of segments copied
func (addr *IPv4Address) CopySubSegments(start, end int, segs []*IPv4AddressSegment) (count int) {
	return addr.GetSection().CopySubSegments(start, end, segs)
}

// CopySubSegments copies the existing segments from the given start index until but not including the segment at the given end index,
// into the given slice, as much as can be fit into the slice, returning the number of segments copied
func (addr *IPv4Address) CopySegments(segs []*IPv4AddressSegment) (count int) {
	return addr.GetSection().CopySegments(segs)
}

// GetSegments returns a slice with the address segments.  The returned slice is not backed by the same array as this address.
func (addr *IPv4Address) GetSegments() []*IPv4AddressSegment {
	return addr.GetSection().GetSegments()
}

// GetSegment returns the segment at the given index.
// The first segment is at index 0.
// GetSegment will panic given a negative index or index larger than the segment count.
func (addr *IPv4Address) GetSegment(index int) *IPv4AddressSegment {
	return addr.init().getSegment(index).ToIPv4()
}

// GetSegmentCount returns the segment count
func (addr *IPv4Address) GetSegmentCount() int {
	return addr.GetDivisionCount()
}

// GetGenericDivision returns the segment at the given index as an DivisionType
func (addr *IPv4Address) GetGenericDivision(index int) DivisionType {
	return addr.init().getDivision(index)
}

// GetGenericSegment returns the segment at the given index as an AddressSegmentType
func (addr *IPv4Address) GetGenericSegment(index int) AddressSegmentType {
	return addr.init().getSegment(index)
}

// GetDivisionCount returns the segment count
func (addr *IPv4Address) GetDivisionCount() int {
	return addr.init().getDivisionCount()
}

func (addr *IPv4Address) GetIPVersion() IPVersion {
	return IPv4
}

func (addr *IPv4Address) checkIdentity(section *IPv4AddressSection) *IPv4Address {
	if section == nil {
		return nil
	}
	sec := section.ToSectionBase()
	if sec == addr.section {
		return addr
	}
	return newIPv4Address(section)
}

func (addr *IPv4Address) Mask(other *IPv4Address) (masked *IPv4Address, err addrerr.IncompatibleAddressError) {
	return addr.maskPrefixed(other, true)
}

func (addr *IPv4Address) maskPrefixed(other *IPv4Address, retainPrefix bool) (masked *IPv4Address, err addrerr.IncompatibleAddressError) {
	addr = addr.init()
	sect, err := addr.GetSection().maskPrefixed(other.GetSection(), retainPrefix)
	if err == nil {
		masked = addr.checkIdentity(sect)
	}
	return
}

func (addr *IPv4Address) BitwiseOr(other *IPv4Address) (masked *IPv4Address, err addrerr.IncompatibleAddressError) {
	return addr.bitwiseOrPrefixed(other, true)
}

func (addr *IPv4Address) bitwiseOrPrefixed(other *IPv4Address, retainPrefix bool) (masked *IPv4Address, err addrerr.IncompatibleAddressError) {
	addr = addr.init()
	sect, err := addr.GetSection().bitwiseOrPrefixed(other.GetSection(), retainPrefix)
	if err == nil {
		masked = addr.checkIdentity(sect)
	}
	return
}

func (addr *IPv4Address) Subtract(other *IPv4Address) []*IPv4Address {
	addr = addr.init()
	sects, _ := addr.GetSection().Subtract(other.GetSection())
	sectLen := len(sects)
	if sectLen == 0 {
		return nil
	} else if sectLen == 1 {
		sec := sects[0]
		if sec.ToSectionBase() == addr.section {
			return []*IPv4Address{addr}
		}
	}
	res := make([]*IPv4Address, sectLen)
	for i, sect := range sects {
		res[i] = newIPv4Address(sect)
	}
	return res
}

func (addr *IPv4Address) Intersect(other *IPv4Address) *IPv4Address {
	addr = addr.init()
	section, _ := addr.GetSection().Intersect(other.GetSection())
	if section == nil {
		return nil
	}
	return addr.checkIdentity(section)
}

func (addr *IPv4Address) SpanWithRange(other *IPv4Address) *IPv4AddressSeqRange {
	return NewIPv4SeqRange(addr.init(), other.init())
}

// GetLower returns the address in the subnet with the lowest numeric value,
// which will be the same address if it represents a single value.
// For example, for "1.2-3.4.5-6", the series "1.2.4.5" is returned.
func (addr *IPv4Address) GetLower() *IPv4Address {
	return addr.init().getLower().ToIPv4()
}

// GetUpper returns the address in the subnet with the highest numeric value,
// which will be the same address if it represents a single value.
// For example, for "1.2-3.4.5-6", the series "1.3.4.6" is returned.
func (addr *IPv4Address) GetUpper() *IPv4Address {
	return addr.init().getUpper().ToIPv4()
}

// GetLowerIPAddress implements the IPAddressRange interface
func (addr *IPv4Address) GetLowerIPAddress() *IPAddress {
	return addr.GetLower().ToIP()
}

// GetUpperIPAddress implements the IPAddressRange interface
func (addr *IPv4Address) GetUpperIPAddress() *IPAddress {
	return addr.GetUpper().ToIP()
}

func (addr *IPv4Address) IsZeroHostLen(prefLen BitCount) bool {
	return addr.init().isZeroHostLen(prefLen)
}

func (addr *IPv4Address) ToZeroHost() (*IPv4Address, addrerr.IncompatibleAddressError) {
	res, err := addr.init().toZeroHost(false)
	return res.ToIPv4(), err
}

func (addr *IPv4Address) ToZeroHostLen(prefixLength BitCount) (*IPv4Address, addrerr.IncompatibleAddressError) {
	res, err := addr.init().toZeroHostLen(prefixLength)
	return res.ToIPv4(), err
}

func (addr *IPv4Address) ToZeroNetwork() *IPv4Address {
	return addr.init().toZeroNetwork().ToIPv4()
}

func (addr *IPv4Address) IsMaxHostLen(prefLen BitCount) bool {
	return addr.init().isMaxHostLen(prefLen)
}

func (addr *IPv4Address) ToMaxHost() (*IPv4Address, addrerr.IncompatibleAddressError) {
	res, err := addr.init().toMaxHost()
	return res.ToIPv4(), err
}

func (addr *IPv4Address) ToMaxHostLen(prefixLength BitCount) (*IPv4Address, addrerr.IncompatibleAddressError) {
	res, err := addr.init().toMaxHostLen(prefixLength)
	return res.ToIPv4(), err
}

func (addr *IPv4Address) Uint32Value() uint32 {
	return addr.GetSection().Uint32Value()
}

func (addr *IPv4Address) UpperUint32Value() uint32 {
	return addr.GetSection().UpperUint32Value()
}

func (addr *IPv4Address) ToPrefixBlock() *IPv4Address {
	return addr.init().toPrefixBlock().ToIPv4()
}

func (addr *IPv4Address) ToPrefixBlockLen(prefLen BitCount) *IPv4Address {
	return addr.init().toPrefixBlockLen(prefLen).ToIPv4()
}

func (addr *IPv4Address) ToBlock(segmentIndex int, lower, upper SegInt) *IPv4Address {
	return addr.init().toBlock(segmentIndex, lower, upper).ToIPv4()
}

func (addr *IPv4Address) WithoutPrefixLen() *IPv4Address {
	if !addr.IsPrefixed() {
		return addr
	}
	return addr.init().withoutPrefixLen().ToIPv4()
}

func (addr *IPv4Address) SetPrefixLen(prefixLen BitCount) *IPv4Address {
	return addr.init().setPrefixLen(prefixLen).ToIPv4()
}

func (addr *IPv4Address) SetPrefixLenZeroed(prefixLen BitCount) (*IPv4Address, addrerr.IncompatibleAddressError) {
	res, err := addr.init().setPrefixLenZeroed(prefixLen)
	return res.ToIPv4(), err
}

func (addr *IPv4Address) AdjustPrefixLen(prefixLen BitCount) *IPv4Address {
	return addr.init().adjustPrefixLen(prefixLen).ToIPv4()
}

func (addr *IPv4Address) AdjustPrefixLenZeroed(prefixLen BitCount) (*IPv4Address, addrerr.IncompatibleAddressError) {
	res, err := addr.init().adjustPrefixLenZeroed(prefixLen)
	return res.ToIPv4(), err
}

func (addr *IPv4Address) AssignPrefixForSingleBlock() *IPv4Address {
	return addr.init().assignPrefixForSingleBlock().ToIPv4()
}

func (addr *IPv4Address) AssignMinPrefixForBlock() *IPv4Address {
	return addr.init().assignMinPrefixForBlock().ToIPv4()
}

// ToSinglePrefixBlockOrAddress converts to a single prefix block or address.
// If the given address is a single prefix block, it is returned.
// If it can be converted to a single prefix block by assigning a prefix length, the converted block is returned.
// If it is a single address, any prefix length is removed and the address is returned.
// Otherwise, nil is returned.
// This method provides the address formats used by tries.
func (addr *IPv4Address) ToSinglePrefixBlockOrAddress() *IPv4Address {
	return addr.init().toSinglePrefixBlockOrAddress().ToIPv4()
}

// ContainsPrefixBlock returns whether the range of this address or subnet contains the block of addresses for the given prefix length.
//
// Unlike ContainsSinglePrefixBlock, whether there are multiple prefix values in this item for the given prefix length makes no difference.
//
// Use GetMinPrefixLenForBlock to determine the smallest prefix length for which this method returns true.
func (addr *IPv4Address) ContainsPrefixBlock(prefixLen BitCount) bool {
	return addr.init().ipAddressInternal.ContainsPrefixBlock(prefixLen)
}

// ContainsSinglePrefixBlock returns whether this address contains a single prefix block for the given prefix length.
//
// This means there is only one prefix value for the given prefix length, and it also contains the full prefix block for that prefix, all addresses with that prefix.
//
// Use GetPrefixLenForSingleBlock to determine whether there is a prefix length for which this method returns true.
func (addr *IPv4Address) ContainsSinglePrefixBlock(prefixLen BitCount) bool {
	return addr.init().ipAddressInternal.ContainsSinglePrefixBlock(prefixLen)
}

// GetMinPrefixLenForBlock returns the smallest prefix length such that this includes the block of addresses for that prefix length.
//
// If the entire range can be described this way, then this method returns the same value as GetPrefixLenForSingleBlock.
//
// There may be a single prefix, or multiple possible prefix values in this item for the returned prefix length.
// Use GetPrefixLenForSingleBlock to avoid the case of multiple prefix values.
//
// If this represents just a single address, returns the bit length of this address.
func (addr *IPv4Address) GetMinPrefixLenForBlock() BitCount {
	return addr.init().ipAddressInternal.GetMinPrefixLenForBlock()
}

// GetPrefixLenForSingleBlock returns a prefix length for which the range of this address subnet matches exactly the block of addresses for that prefix.
//
// If the range can be described this way, then this method returns the same value as GetMinPrefixLenForBlock.
//
// If no such prefix exists, returns nil.
//
// If this segment grouping represents a single value, returns the bit length of this address division series.
//
// IP address examples:
// 1.2.3.4 returns 32
// 1.2.3.4/16 returns 32
// 1.2.*.* returns 16
// 1.2.*.0/24 returns 16
// 1.2.0.0/16 returns 16
// 1.2.*.4 returns null
// 1.2.252-255.* returns 22
func (addr *IPv4Address) GetPrefixLenForSingleBlock() PrefixLen {
	return addr.init().ipAddressInternal.GetPrefixLenForSingleBlock()
}

func (addr *IPv4Address) GetValue() *big.Int {
	return addr.init().section.GetValue()
}

func (addr *IPv4Address) GetUpperValue() *big.Int {
	return addr.init().section.GetUpperValue()
}

func (addr *IPv4Address) GetNetIP() net.IP {
	return addr.Bytes()
}

func (addr *IPv4Address) CopyNetIP(ip net.IP) net.IP {
	if ipv4 := ip.To4(); ipv4 != nil { // this shrinks the arg to 4 bytes if it was 16
		ip = ipv4
	}
	return addr.CopyBytes(ip)
}

func (addr *IPv4Address) GetUpperNetIP() net.IP {
	return addr.UpperBytes()
}

func (addr *IPv4Address) CopyUpperNetIP(ip net.IP) net.IP {
	if ipv4 := ip.To4(); ipv4 != nil { // this shrinks the arg to 4 bytes if it was 16
		ip = ipv4
	}
	return addr.CopyUpperBytes(ip)
}

func (addr *IPv4Address) Bytes() []byte {
	return addr.init().section.Bytes()
}

func (addr *IPv4Address) UpperBytes() []byte {
	return addr.init().section.UpperBytes()
}

func (addr *IPv4Address) CopyBytes(bytes []byte) []byte {
	return addr.init().section.CopyBytes(bytes)
}

func (addr *IPv4Address) CopyUpperBytes(bytes []byte) []byte {
	return addr.init().section.CopyUpperBytes(bytes)
}

func (addr *IPv4Address) IsMax() bool {
	return addr.init().section.IsMax()
}

func (addr *IPv4Address) IncludesMax() bool {
	return addr.init().section.IncludesMax()
}

// TestBit returns true if the bit in the lower value of this address at the given index is 1, where index 0 refers to the least significant bit.
// In other words, it computes (bits & (1 << n)) != 0), using the lower value of this address.
// TestBit will panic if n < 0, or if it matches or exceeds the bit count of this item.
func (addr *IPv4Address) TestBit(n BitCount) bool {
	return addr.init().testBit(n)
}

// IsOneBit returns true if the bit in the lower value of this address at the given index is 1, where index 0 refers to the most significant bit.
// IsOneBit will panic if bitIndex < 0, or if it is larger than the bit count of this item.
func (addr *IPv4Address) IsOneBit(bitIndex BitCount) bool {
	return addr.init().isOneBit(bitIndex)
}

func (addr *IPv4Address) PrefixEqual(other AddressType) bool {
	return addr.init().prefixEquals(other)
}

func (addr *IPv4Address) PrefixContains(other AddressType) bool {
	return addr.init().prefixContains(other)
}

func (addr *IPv4Address) Contains(other AddressType) bool {
	if other == nil || other.ToAddressBase() == nil {
		return true
	} else if addr == nil {
		return false
	}
	addr = addr.init()
	otherAddr := other.ToAddressBase()
	if addr.ToAddressBase() == otherAddr {
		return true
	}
	return otherAddr.getAddrType() == ipv4Type && addr.section.sameCountTypeContains(otherAddr.GetSection())
}

func (addr *IPv4Address) Compare(item AddressItem) int {
	return CountComparator.Compare(addr, item)
}

func (addr *IPv4Address) Equal(other AddressType) bool {
	if addr == nil {
		return other == nil || other.ToAddressBase() == nil
	} else if other.ToAddressBase() == nil {
		return false
	}
	return other.ToAddressBase().getAddrType() == ipv4Type && addr.init().section.sameCountTypeEquals(other.ToAddressBase().GetSection())
}

// CompareSize compares the counts of two subnets or addresses, the number of individual addresses within.
//
// Rather than calculating counts with GetCount, there can be more efficient ways of comparing whether one subnet represents more individual addresses than another.
//
// CompareSize returns a positive integer if this address or subnet has a larger count than the one given, 0 if they are the same, or a negative integer if the other has a larger count.
func (addr *IPv4Address) CompareSize(other AddressType) int {
	if addr == nil {
		if other != nil && other.ToAddressBase() != nil {
			// we have size 0, other has size >= 1
			return -1
		}
		return 0
	}
	return addr.init().compareSize(other)
}

// TrieCompare compares two addresses according to address trie ordering.
// It returns a number less than zero, zero, or a number greater than zero if the first address argument is less than, equal to, or greater than the second.
//
// The comparison is intended for individual addresses and CIDR prefix blocks.
// If an address is neither an individual address nor a prefix block, it is treated like one:
//
//	- ranges that occur inside the prefix length are ignored, only the lower value is used.
//	- ranges beyond the prefix length are assumed to be the full range across all hosts for that prefix length.
func (addr *IPv4Address) TrieCompare(other *IPv4Address) int {
	return addr.init().trieCompare(other.ToAddressBase())
}

// TrieIncrement returns the next address or block according to address trie ordering
//
// If an address is neither an individual address nor a prefix block, it is treated like one:
//
//	- ranges that occur inside the prefix length are ignored, only the lower value is used.
//	- ranges beyond the prefix length are assumed to be the full range across all hosts for that prefix length.
func (addr *IPv4Address) TrieIncrement() *IPv4Address {
	return addr.trieIncrement().ToIPv4()
}

// TrieDecrement returns the previous address or block according to address trie ordering
//
// If an address is neither an individual address nor a prefix block, it is treated like one:
//
//	- ranges that occur inside the prefix length are ignored, only the lower value is used.
//	- ranges beyond the prefix length are assumed to be the full range across all hosts for that prefix length.
func (addr *IPv4Address) TrieDecrement() *IPv4Address {
	return addr.trieDecrement().ToIPv4()
}

func (addr *IPv4Address) MatchesWithMask(other *IPv4Address, mask *IPv4Address) bool {
	return addr.init().GetSection().MatchesWithMask(other.GetSection(), mask.GetSection())
}

// GetMaxSegmentValue returns the maximum possible segment value for this type of address.
//
// Note this is not the maximum of the range of segment values in this specific address,
// this is the maximum value of any segment for this address type and version, determined by the number of bits per segment.
func (addr *IPv4Address) GetMaxSegmentValue() SegInt {
	return addr.init().getMaxSegmentValue()
}

func (addr *IPv4Address) ToSequentialRange() *IPv4AddressSeqRange {
	if addr == nil {
		return nil
	}
	addr = addr.init().WithoutPrefixLen()
	return newSeqRangeUnchecked(addr.GetLower().ToIP(), addr.GetUpper().ToIP(), addr.isMultiple()).ToIPv4()
}

// ToBroadcastAddress returns the IPv4 broadcast address.
// The broadcast address has the same prefix but a host that is all 1 bits.
// If this address or subnet is not prefixed, this returns the address of all 1 bits, the "max" address.
// This returns an error if a prefixed and ranged-valued segment cannot be converted to a host of all ones and remain a range of consecutive values.
func (addr *IPv4Address) ToBroadcastAddress() (*IPv4Address, addrerr.IncompatibleAddressError) {
	return addr.ToMaxHost()
}

// ToNetworkAddress returns the IPv4 network address.
// The network address has the same prefix but a zero host.
// If this address or subnet is not prefixed, this returns the zero "any" address.
// This returns an error if a prefixed and ranged-valued segment cannot be converted to a host of all zeros and remain a range of consecutive values.
func (addr *IPv4Address) ToNetworkAddress() (*IPv4Address, addrerr.IncompatibleAddressError) {
	return addr.ToZeroHost()
}

func (addr *IPv4Address) ToAddressString() *IPAddressString {
	return addr.init().ToIP().ToAddressString()
}

func (addr *IPv4Address) IncludesZeroHostLen(networkPrefixLength BitCount) bool {
	return addr.init().includesZeroHostLen(networkPrefixLength)
}

func (addr *IPv4Address) IncludesMaxHostLen(networkPrefixLength BitCount) bool {
	return addr.init().includesMaxHostLen(networkPrefixLength)
}

// IsLinkLocal returns whether the address is link local, whether unicast or multicast.
func (addr *IPv4Address) IsLinkLocal() bool {
	if addr.IsMulticast() {
		//224.0.0.252	Link-local Multicast Name Resolution	[RFC4795]
		return addr.GetSegment(0).Matches(224) && addr.GetSegment(1).IsZero() && addr.GetSegment(2).IsZero() && addr.GetSegment(3).Matches(252)
	}
	return addr.GetSegment(0).Matches(169) && addr.GetSegment(1).Matches(254)
}

func (addr *IPv4Address) IsPrivate() bool {
	// refer to RFC 1918
	// 10/8 prefix
	// 172.16/12 prefix (172.16.0.0 – 172.31.255.255)
	// 192.168/16 prefix
	seg0, seg1 := addr.GetSegment(0), addr.GetSegment(1)
	return seg0.Matches(10) ||
		(seg0.Matches(172) && seg1.MatchesWithPrefixMask(16, 4)) ||
		(seg0.Matches(192) && seg1.Matches(168))
}

func (addr *IPv4Address) IsMulticast() bool {
	// 1110...
	//224.0.0.0/4
	return addr.GetSegment(0).MatchesWithPrefixMask(0xe0, 4)
}

// IsLocal returns true if the address is link local, site local, organization local, administered locally, or unspecified.
// This includes both unicast and multicast.
func (addr *IPv4Address) IsLocal() bool {
	if addr.IsMulticast() {
		//1110...
		seg0 := addr.GetSegment(0)
		//http://www.tcpipguide.com/free/t_IPMulticastAddressing.htm
		//rfc4607 and https://www.iana.org/assignments/multicast-addresses/multicast-addresses.xhtml

		//239.0.0.0-239.255.255.255 organization local
		if seg0.matches(239) {
			return true
		}
		seg1, seg2 := addr.GetSegment(1), addr.GetSegment(2)

		// 224.0.0.0 to 224.0.0.255 local
		// includes link local multicast name resolution https://tools.ietf.org/html/rfc4795 224.0.0.252
		return (seg0.matches(224) && seg1.IsZero() && seg2.IsZero()) ||
			//232.0.0.1 - 232.0.0.255	Reserved for IANA allocation	[RFC4607]
			//232.0.1.0 - 232.255.255.255	Reserved for local host allocation	[RFC4607]
			(seg0.matches(232) && !(seg1.IsZero() && seg2.IsZero()))
	}
	return addr.IsLinkLocal() || addr.IsPrivate() || addr.IsAnyLocal()
}

// IsUnspecified returns whether this is the unspecified address.  The unspecified address is the address that is all zeros.
func (addr *IPv4Address) IsUnspecified() bool {
	return addr.section == nil || addr.IsZero()
}

// IsAnyLocal returns whether this address is the address which binds to any address on the local host.
// This is the address that has the value of 0, aka the unspecified address.
func (addr *IPv4Address) IsAnyLocal() bool {
	return addr.section == nil || addr.IsZero()
}

// IsLoopback returns whether this address is a loopback address, such as
// [::1] (aka [0:0:0:0:0:0:0:1]) or 127.0.0.1
func (addr *IPv4Address) IsLoopback() bool {
	return addr.section != nil && addr.GetSegment(0).Matches(127)
}

func (addr *IPv4Address) Iterator() IPv4AddressIterator {
	if addr == nil {
		return ipv4AddressIterator{nilAddrIterator()}
	}
	return ipv4AddressIterator{addr.init().addrIterator(nil)}
}

func (addr *IPv4Address) PrefixIterator() IPv4AddressIterator {
	return ipv4AddressIterator{addr.init().prefixIterator(false)}
}

func (addr *IPv4Address) PrefixBlockIterator() IPv4AddressIterator {
	return ipv4AddressIterator{addr.init().prefixIterator(true)}
}

func (addr *IPv4Address) BlockIterator(segmentCount int) IPv4AddressIterator {
	return ipv4AddressIterator{addr.init().blockIterator(segmentCount)}
}

func (addr *IPv4Address) SequentialBlockIterator() IPv4AddressIterator {
	return ipv4AddressIterator{addr.init().sequentialBlockIterator()}
}

func (addr *IPv4Address) GetSequentialBlockIndex() int {
	return addr.init().getSequentialBlockIndex()
}

func (addr *IPv4Address) GetSequentialBlockCount() *big.Int {
	return addr.getSequentialBlockCount()
}

// IncrementBoundary returns the address that is the given increment from the range boundaries of this subnet.
//
// If the given increment is positive, adds the value to the upper address ({@link #getUpper()}) in the subnet range to produce a new address.
// If the given increment is negative, adds the value to the lower address ({@link #getLower()}) in the subnet range to produce a new address.
// If the increment is zero, returns this address.
//
// If this is a single address value, that address is simply incremented by the given increment value, positive or negative.
//
// On address overflow or underflow, IncrementBoundary returns nil.
func (addr *IPv4Address) IncrementBoundary(increment int64) *IPv4Address {
	return addr.init().incrementBoundary(increment).ToIPv4()
}

// Increment returns the address from the subnet that is the given increment upwards into the subnet range,
// with the increment of 0 returning the first address in the range.
//
// If the increment i matches or exceeds the subnet size count c, then i - c + 1
// is added to the upper address of the range.
// An increment matching the subnet count gives you the address just above the highest address in the subnet.
//
// If the increment is negative, it is added to the lower address of the range.
// To get the address just below the lowest address of the subnet, use the increment -1.
//
// If this is just a single address value, the address is simply incremented by the given increment, positive or negative.
//
// If this is a subnet with multiple values, a positive increment i is equivalent i + 1 values from the subnet iterator and beyond.
// For instance, a increment of 0 is the first value from the iterator, an increment of 1 is the second value from the iterator, and so on.
// An increment of a negative value added to the subnet count is equivalent to the same number of iterator values preceding the upper bound of the iterator.
// For instance, an increment of count - 1 is the last value from the iterator, an increment of count - 2 is the second last value, and so on.
//
// On address overflow or underflow, Increment returns nil.
func (addr *IPv4Address) Increment(increment int64) *IPv4Address {
	return addr.init().increment(increment).ToIPv4()
}

func (addr *IPv4Address) SpanWithPrefixBlocks() []*IPv4Address {
	if addr.IsSequential() {
		if addr.IsSinglePrefixBlock() {
			return []*IPv4Address{addr}
		}
		wrapped := wrapIPAddress(addr.ToIP())
		spanning := getSpanningPrefixBlocks(wrapped, wrapped)
		return cloneToIPv4Addrs(spanning)
	}
	wrapped := wrapIPAddress(addr.ToIP())
	return cloneToIPv4Addrs(spanWithPrefixBlocks(wrapped))
}

func (addr *IPv4Address) SpanWithPrefixBlocksTo(other *IPv4Address) []*IPv4Address {
	return cloneToIPv4Addrs(
		getSpanningPrefixBlocks(
			wrapIPAddress(addr.ToIP()),
			wrapIPAddress(other.ToIP()),
		),
	)
}

func (addr *IPv4Address) SpanWithSequentialBlocks() []*IPv4Address {
	if addr.IsSequential() {
		return []*IPv4Address{addr}
	}
	wrapped := wrapIPAddress(addr.ToIP())
	return cloneToIPv4Addrs(spanWithSequentialBlocks(wrapped))
}

func (addr *IPv4Address) SpanWithSequentialBlocksTo(other *IPv4Address) []*IPv4Address {
	return cloneToIPv4Addrs(
		getSpanningSequentialBlocks(
			wrapIPAddress(addr.ToIP()),
			wrapIPAddress(other.ToIP()),
		),
	)
}

func (addr *IPv4Address) CoverWithPrefixBlockTo(other *IPv4Address) *IPv4Address {
	return addr.init().coverWithPrefixBlockTo(other.ToIP()).ToIPv4()
}

func (addr *IPv4Address) CoverWithPrefixBlock() *IPv4Address {
	return addr.init().coverWithPrefixBlock().ToIPv4()
}

//
// MergeToSequentialBlocks merges this with the list of addresses to produce the smallest array of blocks that are sequential
//
// The resulting array is sorted from lowest address value to highest, regardless of the size of each prefix block.
func (addr *IPv4Address) MergeToSequentialBlocks(addrs ...*IPv4Address) []*IPv4Address {
	series := cloneIPv4Addrs(addr, addrs)
	blocks := getMergedSequentialBlocks(series)
	return cloneToIPv4Addrs(blocks)
}

//
// MergeToPrefixBlocks merges this with the list of sections to produce the smallest array of CIDR prefix blocks.
//
// The resulting array is sorted from lowest address value to highest, regardless of the size of each prefix block.
func (addr *IPv4Address) MergeToPrefixBlocks(addrs ...*IPv4Address) []*IPv4Address {
	series := cloneIPv4Addrs(addr, addrs)
	blocks := getMergedPrefixBlocks(series)
	return cloneToIPv4Addrs(blocks)
}

func (addr *IPv4Address) ReverseBytes() *IPv4Address {
	addr = addr.init()
	return addr.checkIdentity(addr.GetSection().ReverseBytes())
}

func (addr *IPv4Address) ReverseBits(perByte bool) (*IPv4Address, addrerr.IncompatibleAddressError) {
	addr = addr.init()
	res, err := addr.GetSection().ReverseBits(perByte)
	if err != nil {
		return nil, err
	}
	return addr.checkIdentity(res), nil
}

func (addr *IPv4Address) ReverseSegments() *IPv4Address {
	addr = addr.init()
	return addr.checkIdentity(addr.GetSection().ReverseSegments())
}

// ReplaceLen replaces segments starting from startIndex and ending before endIndex with the same number of segments starting at replacementStartIndex from the replacement section
// Mappings to or from indices outside the range of this or the replacement address are skipped.
func (addr *IPv4Address) ReplaceLen(startIndex, endIndex int, replacement *IPv4Address, replacementIndex int) *IPv4Address {
	startIndex, endIndex, replacementIndex =
		adjust1To1Indices(startIndex, endIndex, IPv4SegmentCount, replacementIndex, IPv4SegmentCount)
	if startIndex == endIndex {
		return addr
	}
	count := endIndex - startIndex
	addr = addr.init()
	return addr.checkIdentity(addr.GetSection().ReplaceLen(startIndex, endIndex, replacement.GetSection(), replacementIndex, replacementIndex+count))
}

// Replace replaces segments starting from startIndex with segments from the replacement section.
// Mappings to or from indices outside the range of this address or the replacement section are skipped.
func (addr *IPv4Address) Replace(startIndex int, replacement *IPv4AddressSection) *IPv4Address {
	startIndex, endIndex, replacementIndex :=
		adjust1To1Indices(startIndex, startIndex+replacement.GetSegmentCount(), IPv4SegmentCount, 0, replacement.GetSegmentCount())
	count := endIndex - startIndex
	addr = addr.init()
	return addr.checkIdentity(addr.GetSection().ReplaceLen(startIndex, endIndex, replacement, replacementIndex, replacementIndex+count))
}

func (addr *IPv4Address) GetLeadingBitCount(ones bool) BitCount {
	return addr.GetSection().GetLeadingBitCount(ones)
}

func (addr *IPv4Address) GetTrailingBitCount(ones bool) BitCount {
	return addr.GetSection().GetTrailingBitCount(ones)
}

func (addr *IPv4Address) GetNetwork() IPAddressNetwork {
	return ipv4Network
}

// GetIPv6Address creates an IPv6 mixed address using the given ipv6 segments and using this address for the embedded IPv4 segments
func (addr *IPv4Address) GetIPv6Address(section *IPv6AddressSection) (*IPv6Address, addrerr.AddressError) {
	if section.GetSegmentCount() < IPv6MixedOriginalSegmentCount {
		return nil, &addressValueError{addressError: addressError{key: "ipaddress.mac.error.not.eui.convertible"}}
	}
	newSegs := createSegmentArray(IPv6SegmentCount)
	section = section.WithoutPrefixLen()
	section.copyDivisions(newSegs)
	sect, err := createMixedSection(newSegs, addr)
	if err != nil {
		return nil, err
	}
	return newIPv6Address(sect), nil
}

func (addr *IPv4Address) GetIPv4MappedAddress() (*IPv6Address, addrerr.IncompatibleAddressError) {
	zero := zeroIPv6Seg.ToDiv()
	segs := createSegmentArray(IPv6SegmentCount)
	segs[0], segs[1], segs[2], segs[3], segs[4] = zero, zero, zero, zero, zero
	segs[5] = NewIPv6Segment(IPv6MaxValuePerSegment).ToDiv()
	var sect *IPv6AddressSection
	sect, err := createMixedSection(segs, addr.WithoutPrefixLen())
	if err != nil {
		return nil, err
	}
	return newIPv6Address(sect), nil
}

// returns an error if the first or 3rd segments have a range of values that cannot be combined with their neighbouting segments into IPv6 segments
func (addr *IPv4Address) getIPv6Address(ipv6Segs []*AddressDivision) (*IPv6Address, addrerr.IncompatibleAddressError) {
	newSegs := createSegmentArray(IPv6SegmentCount)
	copy(newSegs, ipv6Segs)
	sect, err := createMixedSection(newSegs, addr)
	if err != nil {
		return nil, err
	}
	return newIPv6Address(sect), nil
}

func createMixedSection(newIPv6Divisions []*AddressDivision, mixedSection *IPv4Address) (res *IPv6AddressSection, err addrerr.IncompatibleAddressError) {
	ipv4Section := mixedSection.GetSection().WithoutPrefixLen()
	var seg *IPv6AddressSegment
	if seg, err = ipv4Section.GetSegment(0).Join(ipv4Section.GetSegment(1)); err == nil {
		newIPv6Divisions[6] = seg.ToDiv()
		if seg, err = ipv4Section.GetSegment(2).Join(ipv4Section.GetSegment(3)); err == nil {
			newIPv6Divisions[7] = seg.ToDiv()
			res = newIPv6SectionFromMixed(newIPv6Divisions)
			if res.cache != nil {
				nonMixedSection := res.createNonMixedSection()
				mixedGrouping := newIPv6v4MixedGrouping(
					nonMixedSection,
					ipv4Section,
				)
				mixed := &mixedCache{
					defaultMixedAddressSection: mixedGrouping,
					embeddedIPv6Section:        nonMixedSection,
					embeddedIPv4Section:        ipv4Section,
				}
				res.cache.mixed = mixed
			}
		}
	}
	return
}

func (addr IPv4Address) Format(state fmt.State, verb rune) {
	addr.init().format(state, verb)
}

// String implements the fmt.Stringer interface, returning the canonical string provided by ToCanonicalString, or "<nil>" if the receiver is a nil pointer
func (addr *IPv4Address) String() string {
	if addr == nil {
		return nilString()
	}
	return addr.init().toString()
}

func (addr *IPv4Address) GetSegmentStrings() []string {
	if addr == nil {
		return nil
	}
	return addr.init().getSegmentStrings()
}

func (addr *IPv4Address) ToCanonicalString() string {
	if addr == nil {
		return nilString()
	}
	return addr.init().toCanonicalString()
}

func (addr *IPv4Address) ToNormalizedString() string {
	if addr == nil {
		return nilString()
	}
	return addr.init().toNormalizedString()
}

func (addr *IPv4Address) ToCompressedString() string {
	if addr == nil {
		return nilString()
	}
	return addr.init().toCompressedString()
}

func (addr *IPv4Address) ToCanonicalWildcardString() string {
	if addr == nil {
		return nilString()
	}
	return addr.init().toCanonicalWildcardString()
}

func (addr *IPv4Address) ToNormalizedWildcardString() string {
	if addr == nil {
		return nilString()
	}
	return addr.init().toNormalizedWildcardString()
}

func (addr *IPv4Address) ToSegmentedBinaryString() string {
	if addr == nil {
		return nilString()
	}
	return addr.init().toSegmentedBinaryString()
}

func (addr *IPv4Address) ToSQLWildcardString() string {
	if addr == nil {
		return nilString()
	}
	return addr.init().toSQLWildcardString()
}

func (addr *IPv4Address) ToFullString() string {
	if addr == nil {
		return nilString()
	}
	return addr.init().toFullString()
}

// ToReverseDNSString returns the reverse DNS string.
// The method helps implement the IPAddressSegmentSeries interface.  For IPV4, the error is always nil.
func (addr *IPv4Address) ToReverseDNSString() (string, addrerr.IncompatibleAddressError) {
	if addr == nil {
		return nilString(), nil
	}
	str, _ := addr.init().toReverseDNSString()
	return str, nil
}

func (addr *IPv4Address) ToPrefixLenString() string {
	if addr == nil {
		return nilString()
	}
	return addr.init().toPrefixLenString()
}

func (addr *IPv4Address) ToSubnetString() string {
	if addr == nil {
		return nilString()
	}
	return addr.init().toSubnetString()
}

func (addr *IPv4Address) ToCompressedWildcardString() string {
	if addr == nil {
		return nilString()
	}
	return addr.init().toCompressedWildcardString()
}

func (addr *IPv4Address) ToHexString(with0xPrefix bool) (string, addrerr.IncompatibleAddressError) {
	if addr == nil {
		return nilString(), nil
	}
	return addr.init().toHexString(with0xPrefix)
}

func (addr *IPv4Address) ToOctalString(with0Prefix bool) (string, addrerr.IncompatibleAddressError) {
	if addr == nil {
		return nilString(), nil
	}
	return addr.init().toOctalString(with0Prefix)
}

func (addr *IPv4Address) ToBinaryString(with0bPrefix bool) (string, addrerr.IncompatibleAddressError) {
	if addr == nil {
		return nilString(), nil
	}
	return addr.init().toBinaryString(with0bPrefix)
}

func (addr *IPv4Address) ToInetAtonString(radix Inet_aton_radix) string {
	if addr == nil {
		return nilString()
	}
	return addr.GetSection().ToInetAtonString(radix)
}

func (addr *IPv4Address) ToInetAtonJoinedString(radix Inet_aton_radix, joinedCount int) (string, addrerr.IncompatibleAddressError) {
	if addr == nil {
		return nilString(), nil
	}
	return addr.GetSection().ToInetAtonJoinedString(radix, joinedCount)
}

func (addr *IPv4Address) ToCustomString(stringOptions addrstr.IPStringOptions) string {
	if addr == nil {
		return nilString()
	}
	return addr.GetSection().toCustomZonedString(stringOptions, addr.zone)
}

func (addr *IPv4Address) ToAddressBase() *Address {
	return addr.ToIP().ToAddressBase()
}

func (addr *IPv4Address) ToIP() *IPAddress {
	if addr != nil {
		addr = addr.init()
	}
	return (*IPAddress)(addr)
}

func (addr *IPv4Address) Wrap() WrappedIPAddress {
	return wrapIPAddress(addr.ToIP())
}

// ToKey creates the associated address key.
// While addresses can be compare with the Compare, TrieCompare or Equal methods as well as various provided instances of AddressComparator,
// they are not comparable with go operators.
// However, IPv4AddressKey instances are comparable with go operators, and thus can be used as map keys.
func (addr *IPv4Address) ToKey() *IPv4AddressKey {
	addr = addr.init()
	key := &IPv4AddressKey{
		Prefix: PrefixKey{
			IsPrefixed: addr.IsPrefixed(),
			PrefixLen:  PrefixBitCount(addr.GetPrefixLen().Len()),
		},
	}
	section := addr.GetSection()
	divs := section.divisions.(standardDivArray)
	for i, div := range divs.divisions {
		seg := div.ToIPv4()
		vals := &key.Values[i]
		vals.Value, vals.UpperValue = seg.GetIPv4SegmentValue(), seg.GetIPv4UpperSegmentValue()
	}
	return key
}
