package ipaddr

// IPv4PrefixBlockAllocator allocates CIDR prefix block subnets for assignment to groups of hosts.
// It will perform variable length subnetting to supply prefix blocks of the required size from a set of initial blocks supplied to the allocator.
// The zero value is an allocator, ready to use.
type IPv4PrefixBlockAllocator struct {
	prefixBlockAllocator
}

// AddAvailable provides the given blocks to the allocator for allocating
func (alloc *IPv4PrefixBlockAllocator) AddAvailable(blocks ...*IPv4Address) {
	alloc.prefixBlockAllocator.AddAvailable(cloneIPv4AddrsToIPAddrs(blocks)...)
}

// AllocateSize returns a block of sufficient size,
// the size indicating the number of distinct addresses required in the block,
// or nil if no such block is available in the allocator,
// or if the allocated size needs to be zero.
// The returned block will be able to accommodate sizeRequired hosts as well as the reserved count, if any.
func (alloc *IPv4PrefixBlockAllocator) AllocateSize(sizeRequired uint64) *IPv4Address {
	return alloc.prefixBlockAllocator.AllocateSize(sizeRequired).ToIPv4()
}

// AllocateSizes returns multiple blocks of sufficient size for the given size required,
// or nil if there is insufficient space in the allocator
func (alloc *IPv4PrefixBlockAllocator) AllocateSizes(blockSizes ...uint64) []IPv4AllocatedBlock {
	return cloneIPAllocatedBlocksToIPv4AllocatedBlocks(alloc.prefixBlockAllocator.AllocateSizes(blockSizes...))
}

// AllocateBitLen allocates a block with the given bit-length,
// the bit-length being the number of bits extending beyond the prefix length,
// or nil if no such block is available in the allocator
func (alloc *IPv4PrefixBlockAllocator) AllocateBitLen(bitLength BitCount) *IPv4Address {
	return alloc.prefixBlockAllocator.AllocateBitLen(bitLength).ToIPv4()
}

// AllocateMultiBitLens returns multiple blocks of the given bit-lengths,
// or nil if there is insufficient space in the allocator
func (alloc *IPv4PrefixBlockAllocator) AllocateMultiBitLens(bitLengths ...BitCount) []IPv4AllocatedBlock {
	return cloneIPAllocatedBlocksToIPv4AllocatedBlocks(alloc.prefixBlockAllocator.AllocateMultiBitLens(bitLengths...))
}

// GetAvailable returns a list of all the blocks available for allocating in the allocator
func (alloc *IPv4PrefixBlockAllocator) GetAvailable() (blocks []*IPv4Address) {
	return cloneIPAddrsToIPv4Addrs(alloc.prefixBlockAllocator.GetAvailable())
}

// IPv4AllocatedBlock represents an IPv4 prefix block allocated for a group of hosts of a given size
type IPv4AllocatedBlock struct {
	allocatedBlock
}

// GetAddress returns the block
func (alloc IPv4AllocatedBlock) GetAddress() *IPv4Address {
	return alloc.block.ToIPv4()
}

func cloneIPAllocatedBlocksToIPv4AllocatedBlocks(orig []IPAllocatedBlock) []IPv4AllocatedBlock {
	result := make([]IPv4AllocatedBlock, len(orig))
	for i := range orig {
		result[i] = orig[i].toIPv4()
	}
	return result
}
