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
	//"github.com/seancfoley/ipaddress-go/ipaddr/addrerr"
	"math"
	"math/big"
	"sort"
	"strings"
)

// PrefixBlockConstraint is the generic type constraint used for a prefix block allocator
type PrefixBlockConstraint[T any] interface {
	SequentialRangeConstraint[T]

	MergeToPrefixBlocks(...T) []T
	SetPrefixLen(BitCount) T
	PrefixBlockIterator() Iterator[T]
}

var (
	_ = PrefixBlockAllocator[*IPAddress]{}
	_ = PrefixBlockAllocator[*IPv4Address]{}
	_ = PrefixBlockAllocator[*IPv6Address]{}
)

type PrefixBlockAllocator[T PrefixBlockConstraint[T]] struct {
	version IPVersion
	blocks  [][]T
	reservedCount,
	totalBlockCount int
}

// GetBlockCount returns the count of available blocks in this allocator
func (alloc *PrefixBlockAllocator[T]) GetBlockCount() int {
	return alloc.totalBlockCount
}

// GetVersion returns the IP version of the available blocks in the allocator,
// which is determined by the version of the first block made available to the allocator.
func (alloc *PrefixBlockAllocator[T]) GetVersion() IPVersion {
	return alloc.version
}

// GetTotalCount returns the total of the count of all individual addresses available in this allocator,
// which is the total number of individual addresses in all the blocks.
func (alloc *PrefixBlockAllocator[T]) GetTotalCount() *big.Int {
	if alloc.GetBlockCount() == 0 {
		return bigZero()
	}
	result := bigZero()
	version := alloc.version
	for i := len(alloc.blocks) - 1; i >= 0; i-- {
		if blockCount := len(alloc.blocks[i]); blockCount != 0 {
			size := BlockSize(uint(version.GetBitCount() - i))
			size.Mul(size, big.NewInt(int64(blockCount)))
			result.Add(result, size)
		}
	}
	return result
}

// SetReserved sets the additional number of addresses to be included in any size allocation.
// Any request for a block of a given size will adjust that size by the given number.
// This can be useful when the size requests do not include the count of additional addresses that must be included in every block.
// For IPv4, it is common to reserve two addresses, the network and broadcast addresses.
// If the reservedCount is negative, then every request will be shrunk by that number, useful for cases where
// insufficient space requires that all subnets be reduced in size by an equal number.
func (alloc *PrefixBlockAllocator[T]) SetReserved(reservedCount int) {
	alloc.reservedCount = reservedCount
}

// GetReserved returns the reserved count.  Use SetReserved to change the reserved count.
func (alloc *PrefixBlockAllocator[T]) GetReserved() (reservedCount int) {
	return alloc.reservedCount
}

// AddAvailable provides the given blocks to the allocator for allocating
func (alloc *PrefixBlockAllocator[T]) AddAvailable(blocks ...T) {
	if len(blocks) == 0 {
		return
	}
	for _, block := range blocks {
		version := alloc.version
		if version.IsIndeterminate() {
			alloc.version = block.GetIPVersion()
		} else if !version.Equal(block.GetIPVersion()) {
			panic("mismatched versions")
		}
	}

	if alloc.blocks == nil {
		size := alloc.version.GetBitCount() + 1
		alloc.blocks = make([][]T, size)
	} else {
		for i, existingBlocks := range alloc.blocks {
			blocks = append(blocks, existingBlocks...)
			alloc.blocks[i] = nil
		}
	}

	blocks = blocks[0].MergeToPrefixBlocks(blocks...)

	alloc.insertBlocks(blocks)
}

func (alloc *PrefixBlockAllocator[T]) insertBlocks(blocks []T) {
	for _, block := range blocks {
		prefLen := block.GetPrefixLen().bitCount()
		alloc.blocks[prefLen] = append(alloc.blocks[prefLen], block)
		alloc.totalBlockCount++
	}
}

// GetAvailable returns a list of all the blocks available for allocating in the allocator
func (alloc *PrefixBlockAllocator[T]) GetAvailable() (blocks []T) {
	for _, block := range alloc.blocks {
		blocks = append(blocks, block...)
	}
	return
}

// AllocateSize returns a block of sufficient size,
// the size indicating the number of distinct addresses required in the block,
// or nil if no such block is available in the allocator,
// or if the allocated size needs to be zero.
// The returned block will be able to accommodate sizeRequired hosts as well as the reserved count, if any.
func (alloc *PrefixBlockAllocator[T]) AllocateSize(sizeRequired uint64) T {
	var bitsRequired int
	if alloc.reservedCount < 0 {
		adjustment := uint64(-alloc.reservedCount)
		if adjustment >= sizeRequired {
			var t T
			return t
		}
		sizeRequired -= adjustment
		bitsRequired = BitsForCount(sizeRequired)
	} else if math.MaxUint64-uint64(alloc.reservedCount) < sizeRequired {
		// 64 bits holds MaxUint64 + 1 addresses
		sizeRequired += uint64(alloc.reservedCount) // overflow
		bitsRequired = BitsForCount(sizeRequired) + 64
	} else {
		sizeRequired += uint64(alloc.reservedCount)
		bitsRequired = BitsForCount(sizeRequired)
	}
	return alloc.AllocateBitLen(bitsRequired)
}

// AllocateSizes returns multiple blocks of sufficient size for the given size required,
// or nil if there is insufficient space in the allocator.
// The reserved count, if any, will be added to the required sizes.
func (alloc *PrefixBlockAllocator[T]) AllocateSizes(blockSizes ...uint64) []AllocatedBlock[T] {
	sizes := append(make([]uint64, 0, len(blockSizes)), blockSizes...)
	// sort required subnets by size, largest first
	sort.Slice(sizes, func(i, j int) bool {
		return sizes[i] > sizes[j]
	})
	result := make([]AllocatedBlock[T], 0, len(sizes))
	for _, blockSize := range sizes {
		if alloc.reservedCount < 0 && uint64(-alloc.reservedCount) >= blockSize {
			// size zero
			continue
		}
		allocated := alloc.AllocateSize(blockSize)
		if allocated.IsMultiple() || bigIsOne(allocated.GetCount()) { // count is non-zero
			result = append(result, AllocatedBlock[T]{
				blockSize:     new(big.Int).SetUint64(blockSize),
				reservedCount: alloc.reservedCount,
				block:         allocated,
			})
		} else {
			return nil
		}
	}
	return result
}

// AllocateBitLen allocates a block with the given bit-length,
// the bit-length being the number of bits extending beyond the prefix length,
// or nil if no such block is available in the allocator
// The reserved count is ignored when allocating by bit-length.
func (alloc *PrefixBlockAllocator[T]) AllocateBitLen(bitLength BitCount) T {
	if alloc.totalBlockCount == 0 {
		var t T
		return t
	}
	newPrefixBitCount := alloc.version.GetBitCount() - bitLength
	var block T
	i := newPrefixBitCount
	for ; i >= 0; i-- {
		blockRow := alloc.blocks[i]
		if len(blockRow) > 0 {
			block = blockRow[0]
			var t T
			blockRow[0] = t // just for GC
			alloc.blocks[i] = blockRow[1:]
			alloc.totalBlockCount--
			break
		}
	}
	if (!block.IsMultiple() && !bigIsOne(block.GetCount())) || i == newPrefixBitCount {
		return block
	}
	adjustedBlock := block.SetPrefixLen(newPrefixBitCount)
	blockIterator := adjustedBlock.PrefixBlockIterator()
	result := blockIterator.Next()

	// now we add the remaining from the block iterator back into the list
	alloc.insertBlocks(newSequRangeUnchecked(blockIterator.Next().GetLower(), block.GetUpper(), true).SpanWithPrefixBlocks())

	return result
}

// AllocateMultiBitLens returns multiple blocks of the given bit-lengths,
// or nil if there is insufficient space in the allocator.
// The reserved count is ignored when allocating by bit-length.
func (alloc *PrefixBlockAllocator[T]) AllocateMultiBitLens(bitLengths ...BitCount) []AllocatedBlock[T] {
	lengths := append(make([]BitCount, 0, len(bitLengths)), bitLengths...)

	// sort required subnets by size, largest first
	sort.Slice(lengths, func(i, j int) bool {
		return lengths[i] > lengths[j]
	})
	version := alloc.version
	result := make([]AllocatedBlock[T], 0, len(lengths))
	for _, bitLength := range lengths {
		allocated := alloc.AllocateBitLen(bitLength)
		if allocated.IsMultiple() || bigIsOne(allocated.GetCount()) {
			result = append(result, AllocatedBlock[T]{
				blockSize: new(big.Int).Lsh(bigOneConst(), uint(version.GetBitCount()-bitLength)),
				block:     allocated,
			})
		} else {
			return nil
		}
	}
	return result
}

// String returns a string showing the counts of available blocks for each prefix size in the allocator
func (alloc PrefixBlockAllocator[T]) String() string {
	var builder strings.Builder
	version := alloc.version
	hasBlocks := false
	builder.WriteString("available blocks:\n")
	for i := len(alloc.blocks) - 1; i >= 0; i-- {
		if blockCount := len(alloc.blocks[i]); blockCount != 0 {
			size := BlockSize(uint(version.GetBitCount() - i))
			builder.WriteString(fmt.Sprint(blockCount))
			if blockCount == 1 {
				builder.WriteString(" block with prefix length ")
			} else {
				builder.WriteString(" blocks with prefix length ")
			}
			builder.WriteString(fmt.Sprint(i))
			builder.WriteString(" size ")
			builder.WriteString(fmt.Sprint(size))
			builder.WriteString("\n")
			hasBlocks = true
		}
	}
	if !hasBlocks {
		builder.WriteString("none\n")
	}
	return builder.String()
}

type (
	IPPrefixBlockAllocator   = PrefixBlockAllocator[*IPAddress]
	IPv4PrefixBlockAllocator = PrefixBlockAllocator[*IPv4Address]
	IPv6PrefixBlockAllocator = PrefixBlockAllocator[*IPv6Address]
)

type AllocatedBlock[T AddressType] struct {
	blockSize     *big.Int
	block         T
	reservedCount int
}

// GetAddress returns the block
func (alloc AllocatedBlock[T]) GetAddress() T {
	return alloc.block
}

// GetSize returns the number of hosts for which this block was allocated
func (alloc AllocatedBlock[T]) GetSize() *big.Int {
	return alloc.blockSize
}

// GetCount returns the total number of addresses within the block
func (alloc AllocatedBlock[T]) GetCount() *big.Int {
	return alloc.block.GetCount()
}

// GetReservedCount returns the number of reserved addresses with the block
func (alloc AllocatedBlock[T]) GetReservedCount() int {
	return alloc.reservedCount
}

// String returns a string representation of the allocated block
func (alloc AllocatedBlock[T]) String() string {
	if alloc.reservedCount > 0 {
		return fmt.Sprint(alloc.block, " for ", alloc.blockSize, " hosts and ",
			alloc.reservedCount, " reserved addresses")
	}
	return fmt.Sprint(alloc.block, " for ", alloc.blockSize, " hosts")
}