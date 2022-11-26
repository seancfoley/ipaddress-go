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
	"net"
	"sync"
	"unsafe"
)

type addressNetwork interface {
	getAddressCreator() parsedAddressCreator
}

// IPAddressNetwork represents a network of addresses of a single IP version providing a collection of standard address components for that version, such as masks and loopbacks.
type IPAddressNetwork interface {
	GetLoopback() *IPAddress

	GetNetworkMask(prefixLength BitCount) *IPAddress

	GetPrefixedNetworkMask(prefixLength BitCount) *IPAddress

	GetHostMask(prefixLength BitCount) *IPAddress

	GetPrefixedHostMask(prefixLength BitCount) *IPAddress

	getIPAddressCreator() ipAddressCreator

	addressNetwork
}

type ipAddressNetwork struct {
	subnetsMasksWithPrefix, subnetMasks, hostMasksWithPrefix, hostMasks []*IPAddress
}

//
//
//
//
//
type ipv6AddressNetwork struct {
	ipAddressNetwork
	creator ipv6AddressCreator
}

func (network *ipv6AddressNetwork) getIPAddressCreator() ipAddressCreator {
	return &network.creator
}

func (network *ipv6AddressNetwork) getAddressCreator() parsedAddressCreator {
	return &network.creator
}

func (network *ipv6AddressNetwork) GetLoopback() *IPAddress {
	return ipv6loopback.ToIP()
}

func (network *ipv6AddressNetwork) GetNetworkMask(prefLen BitCount) *IPAddress {
	return getMask(IPv6, zeroIPv6Seg.ToDiv(), prefLen, network.subnetMasks, true, false)
}

func (network *ipv6AddressNetwork) GetPrefixedNetworkMask(prefLen BitCount) *IPAddress {
	return getMask(IPv6, zeroIPv6Seg.ToDiv(), prefLen, network.subnetsMasksWithPrefix, true, true)
}

func (network *ipv6AddressNetwork) GetHostMask(prefLen BitCount) *IPAddress {
	return getMask(IPv6, zeroIPv6Seg.ToDiv(), prefLen, network.hostMasks, false, false)
}

func (network *ipv6AddressNetwork) GetPrefixedHostMask(prefLen BitCount) *IPAddress {
	return getMask(IPv6, zeroIPv6Seg.ToDiv(), prefLen, network.hostMasksWithPrefix, false, true)
}

var _ IPAddressNetwork = &ipv6AddressNetwork{}

// IPv6AddressNetwork is the implementation of IPAddressNetwork for IPv6
type IPv6AddressNetwork struct {
	*ipv6AddressNetwork
}

func (network IPv6AddressNetwork) GetLoopback() *IPv6Address {
	return ipv6loopback
}

func (network IPv6AddressNetwork) GetNetworkMask(prefLen BitCount) *IPv6Address {
	return network.ipv6AddressNetwork.GetNetworkMask(prefLen).ToIPv6()
}

func (network IPv6AddressNetwork) GetPrefixedNetworkMask(prefLen BitCount) *IPv6Address {
	return network.ipv6AddressNetwork.GetPrefixedNetworkMask(prefLen).ToIPv6()
}

func (network IPv6AddressNetwork) GetHostMask(prefLen BitCount) *IPv6Address {
	return network.ipv6AddressNetwork.GetHostMask(prefLen).ToIPv6()
}

func (network IPv6AddressNetwork) GetPrefixedHostMask(prefLen BitCount) *IPv6Address {
	return network.ipv6AddressNetwork.GetPrefixedHostMask(prefLen).ToIPv6()
}

var ipv6Network = &ipv6AddressNetwork{
	ipAddressNetwork: ipAddressNetwork{
		make([]*IPAddress, IPv6BitCount+1),
		make([]*IPAddress, IPv6BitCount+1),
		make([]*IPAddress, IPv6BitCount+1),
		make([]*IPAddress, IPv6BitCount+1),
	},
}

var IPv6Network = &IPv6AddressNetwork{ipv6Network}

//
//
//
//
//

type ipv4AddressNetwork struct {
	ipAddressNetwork
	creator ipv4AddressCreator
}

func (network *ipv4AddressNetwork) getIPAddressCreator() ipAddressCreator {
	return &network.creator
}

func (network *ipv4AddressNetwork) getAddressCreator() parsedAddressCreator {
	return &network.creator
}

func (network *ipv4AddressNetwork) GetLoopback() *IPAddress {
	return ipv4loopback.ToIP()
}

func (network *ipv4AddressNetwork) GetNetworkMask(prefLen BitCount) *IPAddress {
	return getMask(IPv4, zeroIPv4Seg.ToDiv(), prefLen, network.subnetMasks, true, false)
}

func (network *ipv4AddressNetwork) GetPrefixedNetworkMask(prefLen BitCount) *IPAddress {
	return getMask(IPv4, zeroIPv4Seg.ToDiv(), prefLen, network.subnetsMasksWithPrefix, true, true)
}

func (network *ipv4AddressNetwork) GetHostMask(prefLen BitCount) *IPAddress {
	return getMask(IPv4, zeroIPv4Seg.ToDiv(), prefLen, network.hostMasks, false, false)
}

func (network *ipv4AddressNetwork) GetPrefixedHostMask(prefLen BitCount) *IPAddress {
	return getMask(IPv4, zeroIPv4Seg.ToDiv(), prefLen, network.hostMasksWithPrefix, false, true)
}

var _ IPAddressNetwork = &ipv4AddressNetwork{}

// IPv4AddressNetwork is the implementation of IPAddressNetwork for IPv4
type IPv4AddressNetwork struct {
	*ipv4AddressNetwork
}

func (network IPv4AddressNetwork) GetLoopback() *IPv4Address {
	return ipv4loopback
}

func (network IPv4AddressNetwork) GetNetworkMask(prefLen BitCount) *IPv4Address {
	return network.ipv4AddressNetwork.GetNetworkMask(prefLen).ToIPv4()
}

func (network IPv4AddressNetwork) GetPrefixedNetworkMask(prefLen BitCount) *IPv4Address {
	return network.ipv4AddressNetwork.GetPrefixedNetworkMask(prefLen).ToIPv4()
}

func (network IPv4AddressNetwork) GetHostMask(prefLen BitCount) *IPv4Address {
	return network.ipv4AddressNetwork.GetHostMask(prefLen).ToIPv4()
}

func (network IPv4AddressNetwork) GetPrefixedHostMask(prefLen BitCount) *IPv4Address {
	return network.ipv4AddressNetwork.GetPrefixedHostMask(prefLen).ToIPv4()
}

var ipv4Network = &ipv4AddressNetwork{
	ipAddressNetwork: ipAddressNetwork{
		make([]*IPAddress, IPv4BitCount+1),
		make([]*IPAddress, IPv4BitCount+1),
		make([]*IPAddress, IPv4BitCount+1),
		make([]*IPAddress, IPv4BitCount+1),
	},
}

var IPv4Network = &IPv4AddressNetwork{ipv4Network}

var maskMutex sync.Mutex

func getMask(version IPVersion, zeroSeg *AddressDivision, networkPrefixLength BitCount, cache []*IPAddress, network, withPrefixLength bool) *IPAddress {
	bits := networkPrefixLength
	addressBitLength := version.GetBitCount()
	if bits < 0 {
		bits = 0
	} else if bits > addressBitLength {
		bits = addressBitLength
	}
	cacheIndex := bits

	subnet := (*IPAddress)(atomicLoadPointer((*unsafe.Pointer)(unsafe.Pointer(&cache[cacheIndex]))))
	if subnet != nil {
		return subnet
	}
	maskMutex.Lock()
	subnet = cache[cacheIndex]
	if subnet != nil {
		maskMutex.Unlock()
		return subnet
	}
	//
	//
	//

	var onesSubnetIndex, zerosSubnetIndex int
	if network {
		onesSubnetIndex = int(addressBitLength)
		zerosSubnetIndex = 0
	} else {
		onesSubnetIndex = 0
		zerosSubnetIndex = int(addressBitLength)
	}
	segmentCount := version.GetSegmentCount()
	bitsPerSegment := version.GetBitsPerSegment()
	maxSegmentValue := version.GetMaxSegmentValue()

	onesSubnet := (*IPAddress)(atomicLoadPointer((*unsafe.Pointer)(unsafe.Pointer(&cache[onesSubnetIndex]))))
	if onesSubnet == nil {
		newSegments := createSegmentArray(segmentCount)
		if withPrefixLength {
			if network {
				segment := createAddressDivision(zeroSeg.deriveNewSeg(maxSegmentValue, nil))
				lastSegment := createAddressDivision(zeroSeg.deriveNewSeg(maxSegmentValue, cacheBitCount(bitsPerSegment) /* bitsPerSegment */))
				lastIndex := len(newSegments) - 1
				fillDivs(newSegments[:lastIndex], segment)
				newSegments[lastIndex] = lastSegment
				onesSubnet = createIPAddress(createSection(newSegments, cacheBitCount(addressBitLength), version.toType()), NoZone)
			} else {
				segment := createAddressDivision(zeroSeg.deriveNewSeg(maxSegmentValue, cacheBitCount(0)))
				fillDivs(newSegments, segment)
				onesSubnet = createIPAddress(createSection(newSegments, cacheBitCount(0), version.toType()), NoZone)
			}
		} else {
			segment := createAddressDivision(zeroSeg.deriveNewSeg(maxSegmentValue, nil))
			fillDivs(newSegments, segment)
			onesSubnet = createIPAddress(createSection(newSegments, nil, version.toType()), NoZone) /* address creation */
		}
		dataLoc := (*unsafe.Pointer)(unsafe.Pointer(&cache[onesSubnetIndex]))
		atomicStorePointer(dataLoc, unsafe.Pointer(onesSubnet))
	}
	zerosSubnet := (*IPAddress)(atomicLoadPointer((*unsafe.Pointer)(unsafe.Pointer(&cache[zerosSubnetIndex]))))
	if zerosSubnet == nil {
		newSegments := createSegmentArray(segmentCount)
		if withPrefixLength {
			prefLen := cacheBitCount(0)
			if network {
				segment := createAddressDivision(zeroSeg.deriveNewSeg(0, prefLen))
				fillDivs(newSegments, segment)
				zerosSubnet = createIPAddress(createSection(newSegments, prefLen, version.toType()), NoZone)
			} else {
				lastSegment := createAddressDivision(zeroSeg.deriveNewSeg(0, cacheBitCount(bitsPerSegment) /* bitsPerSegment */))
				lastIndex := len(newSegments) - 1
				fillDivs(newSegments[:lastIndex], zeroSeg)
				newSegments[lastIndex] = lastSegment
				zerosSubnet = createIPAddress(createSection(newSegments, cacheBitCount(addressBitLength), version.toType()), NoZone)
			}
		} else {
			segment := createAddressDivision(zeroSeg.deriveNewSeg(0, nil))
			fillDivs(newSegments, segment)
			zerosSubnet = createIPAddress(createSection(newSegments, nil, version.toType()), NoZone)
		}
		dataLoc := (*unsafe.Pointer)(unsafe.Pointer(&cache[zerosSubnetIndex]))
		atomicStorePointer(dataLoc, unsafe.Pointer(zerosSubnet))
	}
	prefix := bits
	onesSegment := onesSubnet.getDivision(0)
	zerosSegment := zerosSubnet.getDivision(0)
	newSegments := createSegmentArray(segmentCount)[:0]
	i := 0
	for ; bits > 0; i, bits = i+1, bits-bitsPerSegment {
		if bits <= bitsPerSegment {
			var segment *AddressDivision

			//first do a check whether we have already created a segment like the one we need
			offset := ((bits - 1) % bitsPerSegment) + 1
			for j, entry := 0, offset; j < segmentCount; j, entry = j+1, entry+bitsPerSegment {
				//for j := 0, entry = offset; j < segmentCount; j++, entry += bitsPerSegment {
				if entry != cacheIndex { //we already know that the entry at cacheIndex is nil
					prev := cache[entry]
					if prev != nil {
						segment = prev.getDivision(j)
						break
					}
				}
			}

			//if none of the other addresses with a similar segment are created yet, we need a new segment.
			if segment == nil {
				if network {
					mask := maxSegmentValue & (maxSegmentValue << uint(bitsPerSegment-bits))
					if withPrefixLength {
						segment = createAddressDivision(zeroSeg.deriveNewSeg(mask, getDivisionPrefixLength(bitsPerSegment, bits)))
					} else {
						segment = createAddressDivision(zeroSeg.deriveNewSeg(mask, nil))
					}
				} else {
					mask := maxSegmentValue & ^(maxSegmentValue << uint(bitsPerSegment-bits))
					if withPrefixLength {
						segment = createAddressDivision(zeroSeg.deriveNewSeg(mask, getDivisionPrefixLength(bitsPerSegment, bits)))
					} else {
						segment = createAddressDivision(zeroSeg.deriveNewSeg(mask, nil))
					}
				}
			}
			newSegments = append(newSegments, segment)
		} else {
			if network {
				newSegments = append(newSegments, onesSegment)
			} else {
				newSegments = append(newSegments, zerosSegment)
			}
		}
	}
	for ; i < segmentCount; i++ {
		if network {
			newSegments = append(newSegments, zerosSegment)
		} else {
			newSegments = append(newSegments, onesSegment)
		}
	}
	var prefLen PrefixLen
	if withPrefixLength {
		prefLen = cacheBitCount(prefix)
	}
	subnet = createIPAddress(createSection(newSegments, prefLen, version.toType()), NoZone)
	dataLoc := (*unsafe.Pointer)(unsafe.Pointer(&cache[cacheIndex]))
	atomicStorePointer(dataLoc, unsafe.Pointer(subnet))
	maskMutex.Unlock()
	return subnet
}

type macAddressNetwork struct {
	creator macAddressCreator
}

func (network *macAddressNetwork) getAddressCreator() parsedAddressCreator {
	return &network.creator
}

var macNetwork = &macAddressNetwork{}

var _ addressNetwork = &macAddressNetwork{}

var ipv4loopback = createIPv4Loopback()
var ipv6loopback = createIPv6Loopback()

func createIPv6Loopback() *IPv6Address {
	ipv6loopback, _ := NewIPv6AddressFromBytes(net.IPv6loopback)
	return ipv6loopback
}

func createIPv4Loopback() *IPv4Address {
	ipv4loopback, _ := NewIPv4AddressFromBytes([]byte{127, 0, 0, 1})
	return ipv4loopback
}
