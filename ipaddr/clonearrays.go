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

func cloneIPv4Sections(sect *IPv4AddressSection, orig []*IPv4AddressSection) []ExtendedIPSegmentSeries {
	converter := func(a *IPv4AddressSection) ExtendedIPSegmentSeries { return wrapIPSection(a.ToIP()) }
	if sect == nil {
		return cloneTo(orig, converter)
	}
	return cloneToExtra(sect, orig, converter) // return types matter with interfaces - https://play.golang.org/p/HZR8FSp42a9 )
}

func cloneIPv6Sections(sect *IPv6AddressSection, orig []*IPv6AddressSection) []ExtendedIPSegmentSeries {
	converter := func(a *IPv6AddressSection) ExtendedIPSegmentSeries { return wrapIPSection(a.ToIP()) }
	if sect == nil {
		return cloneTo(orig, converter)
	}
	return cloneToExtra(sect, orig, converter)
}

// returns a slice of addresses that match the same IP version as the given
func filterCloneIPAddrs(addr *IPAddress, orig []*IPAddress) []ExtendedIPSegmentSeries {
	addrType := addr.getAddrType()
	result := make([]ExtendedIPSegmentSeries, 0, len(orig)+1)
	result = append(result, wrapIPAddress(addr))
	for _, a := range orig {
		if addrType == a.getAddrType() {
			result = append(result, a.Wrap())
		}
	}
	return result
}

func cloneIPv4Addrs(sect *IPv4Address, orig []*IPv4Address) []ExtendedIPSegmentSeries {
	converter := func(a *IPv4Address) ExtendedIPSegmentSeries { return wrapIPAddress(a.ToIP()) }
	if sect == nil {
		return cloneTo(orig, converter)
	}
	return cloneToExtra(sect, orig, converter)
}

func cloneIPv6Addrs(sect *IPv6Address, orig []*IPv6Address) []ExtendedIPSegmentSeries {
	converter := func(a *IPv6Address) ExtendedIPSegmentSeries { return wrapIPAddress(a.ToIP()) }
	if sect == nil {
		return cloneTo(orig, converter)
	}
	return cloneToExtra(sect, orig, converter)
}

func cloneToIPSections(orig []ExtendedIPSegmentSeries) []*IPAddressSection {
	return cloneTo(orig, func(a ExtendedIPSegmentSeries) *IPAddressSection { return a.(WrappedIPAddressSection).IPAddressSection })
}

func cloneToIPv4Sections(orig []ExtendedIPSegmentSeries) []*IPv4AddressSection {
	return cloneTo(orig, func(a ExtendedIPSegmentSeries) *IPv4AddressSection {
		return a.(WrappedIPAddressSection).IPAddressSection.ToIPv4()
	})
}

func cloneToIPv6Sections(orig []ExtendedIPSegmentSeries) []*IPv6AddressSection {
	return cloneTo(orig, func(a ExtendedIPSegmentSeries) *IPv6AddressSection {
		return a.(WrappedIPAddressSection).IPAddressSection.ToIPv6()
	})
}

func cloneToIPAddrs(orig []ExtendedIPSegmentSeries) []*IPAddress {
	return cloneTo(orig, func(a ExtendedIPSegmentSeries) *IPAddress { return a.(WrappedIPAddress).IPAddress })
}

func cloneToIPv4Addrs(orig []ExtendedIPSegmentSeries) []*IPv4Address {
	return cloneTo(orig, func(a ExtendedIPSegmentSeries) *IPv4Address { return a.(WrappedIPAddress).IPAddress.ToIPv4() })
}

func cloneToIPv6Addrs(orig []ExtendedIPSegmentSeries) []*IPv6Address {
	return cloneTo(orig, func(a ExtendedIPSegmentSeries) *IPv6Address { return a.(WrappedIPAddress).IPAddress.ToIPv6() })
}

func cloneTo[T any, U any](orig []T, conv func(T) U) []U {
	result := make([]U, len(orig))
	for i := range orig {
		result[i] = conv(orig[i])
	}
	return result
}

func cloneToExtra[T any, U any](sect T, orig []T, conv func(T) U) []U {
	origCount := len(orig)
	result := make([]U, origCount+1)
	result[origCount] = conv(sect)
	for i := range orig {
		result[i] = conv(orig[i])
	}
	return result
}

func copyTo[T any, U any](dest []U, orig []T, conv func(T) U) {
	for i := range orig {
		if i == len(dest) {
			break
		}
		dest[i] = conv(orig[i])
	}
	return
}
