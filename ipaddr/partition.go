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
)

// A Partition is a collection of addresses partitioned from an original address.
// Much like an iterator, the elements of the partition can be iterated just once, after which it becomes empty.
type Partition struct {
	original *IPAddress

	single *IPAddress

	iterator IPAddressIterator

	count *big.Int
}

// IPv6Partition is a Partition of an IPv6 address
type IPv6Partition struct {
	p *Partition
}

// ForEach calls the action with each partition element
func (p IPv6Partition) ForEach(action func(*IPv6Address)) {
	p.p.ForEach(func(address *IPAddress) {
		action(address.ToIPv6())
	})
}

// Iterator provides an iterator to iterate through each element of the partition.
func (p IPv6Partition) Iterator() IPv6AddressIterator {
	return ipv6IPAddressIterator{p.p.Iterator()}
}

// PredicateForEach applies the supplied predicate operation argument to each element of the partition,
// returning true if they all return true, false otherwise
func (p IPv6Partition) PredicateForEach(predicate func(*IPv6Address) bool) bool {
	return p.p.PredicateForEach(func(address *IPAddress) bool {
		return predicate(address.ToIPv6())
	})
}

// PredicateForEachEarly applies the supplied predicate operation argument to each element of the partition,
// returning false if the given predicate returns false for any of the elements.
//
// The method returns when one application of the predicate returns false (determining the overall result)
func (p IPv6Partition) PredicateForEachEarly(predicate func(*IPv6Address) bool) bool {
	return p.p.PredicateForEachEarly(func(address *IPAddress) bool {
		return predicate(address.ToIPv6())
	})
}

// PredicateForAnyEarly applies the operation to each element of the partition,
// returning true if the given predicate returns true for any of the elements.
//
// The method returns when one application of the predicate returns true (determining the overall result)
func (p IPv6Partition) PredicateForAnyEarly(predicate func(*IPv6Address) bool) bool {
	return p.p.PredicateForAnyEarly(func(address *IPAddress) bool {
		return predicate(address.ToIPv6())
	})
}

// PredicateForAny applies the supplied predicate operation argument to each element of the partition,
// returning true if the given predicate returns true for any of the elements.
func (p IPv6Partition) PredicateForAny(predicate func(*IPv6Address) bool) bool {
	return p.p.PredicateForAny(func(address *IPAddress) bool {
		return predicate(address.ToIPv6())
	})
}

// IPv4Partition is a Partition of an IPv4 address
type IPv4Partition struct {
	p *Partition
}

// ForEach calls the action with each partition element
func (p IPv4Partition) ForEach(action func(*IPv4Address)) {
	p.p.ForEach(func(address *IPAddress) {
		action(address.ToIPv4())
	})
}

// Iterator provides an iterator to iterate through each element of the partition.
func (p IPv4Partition) Iterator() IPv4AddressIterator {
	return ipv4IPAddressIterator{p.p.Iterator()}
}

// PredicateForEach applies the supplied predicate operation argument to each element of the partition,
// returning true if they all return true, false otherwise
func (p IPv4Partition) PredicateForEach(predicate func(*IPv4Address) bool) bool {
	return p.p.PredicateForEach(func(address *IPAddress) bool {
		return predicate(address.ToIPv4())
	})
}

// PredicateForEachEarly applies the supplied predicate operation argument to each element of the partition,
// returning false if the given predicate returns false for any of the elements.
//
// The method returns when one application of the predicate returns false (determining the overall result)
func (p IPv4Partition) PredicateForEachEarly(predicate func(*IPv4Address) bool) bool {
	return p.p.PredicateForEachEarly(func(address *IPAddress) bool {
		return predicate(address.ToIPv4())
	})
}

// PredicateForAnyEarly applies the supplied predicate operation argument to each element of the partition,
// returning true if the given predicate returns true for any of the elements.
//
// The method returns when one application of the predicate returns true (determining the overall result)
func (p IPv4Partition) PredicateForAnyEarly(predicate func(*IPv4Address) bool) bool {
	return p.p.PredicateForAnyEarly(func(address *IPAddress) bool {
		return predicate(address.ToIPv4())
	})
}

// PredicateForAny applies the supplied predicate operation argument to each element of the partition,
// returning true if the given predicate returns true for any of the elements.
func (p IPv4Partition) PredicateForAny(predicate func(*IPv4Address) bool) bool {
	return p.p.PredicateForAny(func(address *IPAddress) bool {
		return predicate(address.ToIPv4())
	})
}

// TODO LATER this needs generics (of course, the whole Partition type could be generic as well)
// Otherwise our map function would have to map *IPAddress to interface{}
//
// Supplies to the given function each element of this partition,
// inserting non-null return values into the returned map.
//	public <R> Map<E, R> ApplyForEach(Function<? super E, ? extends R> func) {
//		TreeMap<E, R> results = new TreeMap<>();
//		forEach(address -> {
//			R result = func.apply(address);
//			if(result != null) {
//				results.put(address, result);
//			}
//		});
//		return results;
//	}

// ForEach calls the given action on each partition element.
func (p *Partition) ForEach(action func(*IPAddress)) {
	if p.iterator == nil {
		item := p.single
		if item != nil {
			p.single = nil
			action(item)
		}
	} else {
		iterator := p.iterator
		for iterator.HasNext() {
			action(iterator.Next())
		}
		p.iterator = nil
	}
}

// Iterator provides an iterator to iterate through each element of the partition.
func (p *Partition) Iterator() IPAddressIterator {
	if p.iterator == nil {
		item := p.single
		if item != nil {
			res := &ipAddrIterator{&singleAddrIterator{item.ToAddressBase()}}
			item = nil
			return res
		}
		return nil
	}
	res := p.iterator
	p.iterator = nil
	return res
}

// PredicateForEach applies the supplied predicate operation to each element of the partition,
// returning true if they all return true, false otherwise
//
// Use IPAddressPredicateAdapter to pass in a function that takes *Address as argument instead.
func (p *Partition) PredicateForEach(predicate func(*IPAddress) bool) bool {
	return p.predicateForEach(predicate, false)
}

// PredicateForEachEarly applies the supplied predicate operation to each element of the partition,
// returning false if the given predicate returns false for any of the elements.
//
// The method returns when one application of the predicate returns false (determining the overall result)
//
// Use IPAddressPredicateAdapter to pass in a function that takes *Address as argument instead.
func (p *Partition) PredicateForEachEarly(predicate func(*IPAddress) bool) bool {
	return p.predicateForEach(predicate, false)
}

func (p *Partition) predicateForEach(predicate func(*IPAddress) bool, returnEarly bool) bool {
	if p.iterator == nil {
		return predicate(p.single)
	}
	result := true
	iterator := p.iterator
	for iterator.HasNext() {
		if !predicate(iterator.Next()) {
			result = false
			if returnEarly {
				break
			}
		}
	}
	return result
}

// PredicateForAnyEarly applies the supplied predicate operation to each element of the partition,
// returning true if the given predicate returns true for any of the elements.
//
// The method returns when one application of the predicate returns true (determining the overall result)
func (p *Partition) PredicateForAnyEarly(predicate func(*IPAddress) bool) bool {
	return p.predicateForAny(predicate, true)
}

// PredicateForAny applies the supplied predicate operation to each element of the partition,
// returning true if the given predicate returns true for any of the elements.
func (p *Partition) PredicateForAny(predicate func(*IPAddress) bool) bool {
	return p.predicateForAny(predicate, false)
}

func (p *Partition) predicateForAny(predicate func(address *IPAddress) bool, returnEarly bool) bool {
	return !p.predicateForEach(func(addr *IPAddress) bool {
		return !predicate(addr)
	}, returnEarly)
}

// PartitionIpv6WithSpanningBlocks partitions the address series into prefix blocks and single addresses.
//
// This method iterates through a list of prefix blocks of different sizes that span the entire subnet.
func PartitionIpv6WithSpanningBlocks(newAddr *IPv6Address) IPv6Partition {
	return IPv6Partition{PartitionIPWithSpanningBlocks(newAddr.ToIP())}
}

// PartitionIpv4WithSpanningBlocks partitions the address series into prefix blocks and single addresses.
//
// This method iterates through a list of prefix blocks of different sizes that span the entire subnet.
func PartitionIpv4WithSpanningBlocks(newAddr *IPv4Address) IPv4Partition {
	return IPv4Partition{PartitionIPWithSpanningBlocks(newAddr.ToIP())}
}

// PartitionIPWithSpanningBlocks partitions the address series into prefix blocks and single addresses.
//
// This method iterates through a list of prefix blocks of different sizes that span the entire subnet.
func PartitionIPWithSpanningBlocks(newAddr *IPAddress) *Partition {
	if !newAddr.isMultiple() {
		if !newAddr.IsPrefixed() {
			return &Partition{
				original: newAddr,
				single:   newAddr,
				count:    bigOneConst(),
			}
		}
		return &Partition{
			original: newAddr,
			single:   newAddr.WithoutPrefixLen(),
			count:    bigOneConst(),
		}
	} else if newAddr.IsSinglePrefixBlock() {
		return &Partition{
			original: newAddr,
			single:   newAddr,
			count:    bigOneConst(),
		}
	}
	blocks := newAddr.SpanWithPrefixBlocks()
	return &Partition{
		original: newAddr,
		iterator: &ipAddrSliceIterator{blocks},
		count:    big.NewInt(int64(len(blocks))),
	}
}

// PartitionIPv6WithSingleBlockSize partitions the address series into prefix blocks and single addresses.
//
// This method chooses the maximum block size for a list of prefix blocks contained by the address or subnet,
// and then iterates to produce blocks of that size.
func PartitionIPv6WithSingleBlockSize(newAddr *IPv6Address) IPv6Partition {
	return IPv6Partition{PartitionIPWithSingleBlockSize(newAddr.ToIP())}
}

// PartitionIPv4WithSingleBlockSize partitions the address series into prefix blocks and single addresses.
//
// This method chooses the maximum block size for a list of prefix blocks contained by the address or subnet,
// and then iterates to produce blocks of that size.
func PartitionIPv4WithSingleBlockSize(newAddr *IPv4Address) IPv4Partition {
	return IPv4Partition{PartitionIPWithSingleBlockSize(newAddr.ToIP())}
}

// PartitionIPWithSingleBlockSize partitions the address series into prefix blocks and single addresses.
//
// This method chooses the maximum block size for a list of prefix blocks contained by the address or subnet,
// and then iterates to produce blocks of that size.
func PartitionIPWithSingleBlockSize(newAddr *IPAddress) *Partition {
	if !newAddr.isMultiple() {
		if !newAddr.IsPrefixed() {
			return &Partition{
				original: newAddr,
				single:   newAddr,
				count:    bigOneConst(),
			}
		}
		return &Partition{
			original: newAddr,
			single:   newAddr.WithoutPrefixLen(),
			count:    bigOneConst(),
		}
	} else if newAddr.IsSinglePrefixBlock() {
		return &Partition{
			original: newAddr,
			single:   newAddr,
			count:    bigOneConst(),
		}
	}
	// prefix blocks are handled as prefix blocks,
	// such as 1.2.*.*, which is handled as prefix block iterator for 1.2.0.0/16,
	// but 1.2.3-4.5 is handled as iterator with no prefix lengths involved
	series := newAddr.AssignMinPrefixForBlock()
	if series.getPrefixLen().bitCount() != newAddr.GetBitCount() {
		return &Partition{
			original: newAddr,
			iterator: series.PrefixBlockIterator(),
			count:    series.GetPrefixCountLen(series.getPrefixLen().bitCount()),
		}
	}
	return &Partition{
		original: newAddr,
		iterator: newAddr.WithoutPrefixLen().Iterator(),
		count:    newAddr.GetCount(),
	}
}

//TODO LATER partition ranges (not just addresses) with spanning blocks
