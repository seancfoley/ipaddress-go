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

// Iterator iterates collections, such as subnets and sequential address ranges
type Iterator[T any] interface {
	hasNext

	// Next returns the next item, or the zero value for T if there is none left.
	Next() T
}

// IteratorRem is an iterator that provides a removal operation
type IteratorRem[T any] interface { //TODO NOW rename IteratorRemover or IteratorRemove?  But we are not removing the iterator.  IteratorWithRemove?  IterateRemover?  That last one is not bad.   Actually, I sorta like IteratorWithRemove
	Iterator[T]

	// Remove removes the last iterated item from the underlying data structure or collection, and returns that element.
	// If there is no such element, it returns the zero value for T.
	Remove() T
}

//type singleAddrIterator struct {
//	original *Address
//}
//
//func (it *singleAddrIterator) HasNext() bool {
//	return it.original != nil
//}
//
//func (it *singleAddrIterator) Next() (res *Address) {
//	if it.HasNext() {
//		res = it.original
//		it.original = nil
//	}
//	return
//}

type singleIterator[T any] struct {
	empty    bool
	original T
}

func (it *singleIterator[T]) HasNext() bool {
	return !it.empty
}

func (it *singleIterator[T]) Next() (res T) {
	if it.HasNext() {
		res = it.original
		it.empty = true
	}
	return
}

type multiAddrIterator struct {
	Iterator[*AddressSection]
	zone Zone
}

func (it multiAddrIterator) Next() (res *Address) {
	if it.HasNext() {
		sect := it.Iterator.Next()
		res = createAddress(sect, it.zone)
	}
	return
}

func nilAddrIterator() Iterator[*Address] {
	return &singleIterator[*Address]{}
}

func nilIterator[T any]() Iterator[T] {
	return &singleIterator[T]{}
}

func addrIterator(
	single bool,
	original *Address,
	prefixLen PrefixLen,
	valsAreMultiple bool,
	iterator Iterator[[]*AddressDivision]) Iterator[*Address] {
	if single {
		return &singleIterator[*Address]{original: original}
	}
	return multiAddrIterator{
		Iterator: &multiSectionIterator{
			original:        original.section,
			iterator:        iterator,
			valsAreMultiple: valsAreMultiple,
			prefixLen:       prefixLen,
		},
		zone: original.zone,
	}
}

func prefixAddrIterator(
	single bool,
	original *Address,
	prefixLen PrefixLen,
	iterator Iterator[[]*AddressDivision]) Iterator[*Address] {
	if single {
		return &singleIterator[*Address]{original: original}
	}
	var zone Zone
	if original != nil {
		zone = original.zone
	}
	return multiAddrIterator{
		Iterator: &prefixSectionIterator{
			original:  original.section,
			iterator:  iterator,
			prefixLen: prefixLen,
		},
		zone: zone,
	}
}

// this one is used by the sequential ranges
func rangeAddrIterator(
	single bool,
	original *Address,
	prefixLen PrefixLen,
	valsAreMultiple bool,
	iterator Iterator[[]*AddressDivision]) Iterator[*Address] {
	return addrIterator(single, original, prefixLen, valsAreMultiple, iterator)
}

type ipAddrIterator struct {
	Iterator[*Address]
}

func (iter ipAddrIterator) Next() *IPAddress {
	return iter.Iterator.Next().ToIP()
}

//type ipAddrSliceIterator struct {
//	addrs []*IPAddress
//}
//
//func (iter *ipAddrSliceIterator) HasNext() bool {
//	return len(iter.addrs) > 0
//}
//
//func (iter *ipAddrSliceIterator) Next() (res *IPAddress) {
//	if iter.HasNext() {
//		res = iter.addrs[0]
//		iter.addrs = iter.addrs[1:]
//	}
//	return
//}

//func (iter *ipAddrSliceIterator) append(addrs []*IPAddress) {
//	iter.addrs = append(iter.addrs, addrs...)
//}

type sliceIterator[T any] struct {
	elements []T
}

func (iter *sliceIterator[T]) HasNext() bool {
	return len(iter.elements) > 0
}

func (iter *sliceIterator[T]) Next() (res T) {
	if iter.HasNext() {
		res = iter.elements[0]
		iter.elements = iter.elements[1:]
	}
	return
}

//func (iter *sliceIterator[T]) append(addrs []T) {
//	iter.elements = append(iter.elements, addrs...)
//}

type ipv4AddressIterator struct {
	Iterator[*Address]
}

func (iter ipv4AddressIterator) Next() *IPv4Address {
	return iter.Iterator.Next().ToIPv4()
}

//type ipv4IPAddressIterator struct {
//	Iterator[*IPAddress]
//}
//
//func (iter ipv4IPAddressIterator) Next() *IPv4Address {
//	return iter.Iterator.Next().ToIPv4()
//}

type ipv6AddressIterator struct {
	Iterator[*Address]
}

func (iter ipv6AddressIterator) Next() *IPv6Address {
	return iter.Iterator.Next().ToIPv6()
}

//type ipv6IPAddressIterator struct {
//	Iterator[*IPAddress]
//}
//
//func (iter ipv6IPAddressIterator) Next() *IPv6Address {
//	return iter.Iterator.Next().ToIPv6()
//}

type macAddressIterator struct {
	Iterator[*Address]
}

func (iter macAddressIterator) Next() *MACAddress {
	return iter.Iterator.Next().ToMAC()
}

type addressSeriesIterator struct {
	Iterator[*Address]
}

func (iter addressSeriesIterator) Next() ExtendedSegmentSeries {
	return wrapAddress(iter.Iterator.Next())
}

//type wrapperIterator[T WrappedConstraint[T, R, S], R AddressSectionType, S AddressSegmentType] struct {
//	Iterator[T]
//}
//
//func (iter wrapperIterator[T, R, S]) Next() ExtendedSegmentSeries {
//	return wrap(iter.Iterator.Next())
//}

type ipaddressSeriesIterator struct {
	Iterator[*IPAddress]
}

func (iter ipaddressSeriesIterator) Next() ExtendedIPSegmentSeries {
	return iter.Iterator.Next().Wrap()
}

type sectionSeriesIterator struct {
	Iterator[*AddressSection]
}

func (iter sectionSeriesIterator) Next() ExtendedSegmentSeries {
	return wrapSection(iter.Iterator.Next())
}

type ipsectionSeriesIterator struct {
	Iterator[*IPAddressSection]
}

func (iter ipsectionSeriesIterator) Next() ExtendedIPSegmentSeries {
	return wrapIPSection(iter.Iterator.Next())
}
