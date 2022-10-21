package ipaddr

// IPv6PrefixBlockAllocator allocates CIDR prefix block subnets for assignment to groups of hosts.
// It will perform variable length subnetting to supply prefix blocks of the required size from a set of initial blocks supplied to the allocator.
// The zero value is an allocator, ready to use.
type IPv6PrefixBlockAllocator struct {
	prefixBlockAllocator
}

// AddAvailable provides the given blocks to the allocator for allocating
func (alloc *IPv6PrefixBlockAllocator) AddAvailable(blocks ...*IPv6Address) {
	alloc.prefixBlockAllocator.AddAvailable(cloneIPv6AddrsToIPAddrs(blocks)...)
}

// AllocateSize returns a block of sufficient size,
// the size indicating the number of distinct addresses required in the block,
// or nil if no such block is available in the allocator,
// or if the allocated size needs to be zero.
// The returned block will be able to accommodate sizeRequired hosts as well as the reserved count, if any.
func (alloc *IPv6PrefixBlockAllocator) AllocateSize(sizeRequired uint64) *IPv6Address {
	return alloc.prefixBlockAllocator.AllocateSize(sizeRequired).ToIPv6()
}

// AllocateSizes returns multiple blocks of sufficient size for the given size required,
// or nil if there is insufficient space in the allocator
func (alloc *IPv6PrefixBlockAllocator) AllocateSizes(blockSizes ...uint64) []IPv6AllocatedBlock {
	return cloneIPAllocatedBlocksToIPv6AllocatedBlocks(alloc.prefixBlockAllocator.AllocateSizes(blockSizes...))
}

// AllocateBitLen allocates a block with the given bit-length,
// the bit-length being the number of bits extending beyond the prefix length,
// or nil if no such block is available in the allocator
func (alloc *IPv6PrefixBlockAllocator) AllocateBitLen(bitLength BitCount) *IPv6Address {
	return alloc.prefixBlockAllocator.AllocateBitLen(bitLength).ToIPv6()
}

// AllocateMultiBitLens returns multiple blocks of the given bit-lengths,
// or nil if there is insufficient space in the allocator
func (alloc *IPv6PrefixBlockAllocator) AllocateMultiBitLens(bitLengths ...BitCount) []IPv6AllocatedBlock {
	return cloneIPAllocatedBlocksToIPv6AllocatedBlocks(alloc.prefixBlockAllocator.AllocateMultiBitLens(bitLengths...))
}

// GetAvailable returns a list of all the blocks available for allocating in the allocator
func (alloc *IPv6PrefixBlockAllocator) GetAvailable() (blocks []*IPv6Address) {
	return cloneIPAddrsToIPv6Addrs(alloc.prefixBlockAllocator.GetAvailable())
}

// IPv6AllocatedBlock represents an IPv4 prefix block allocated for a group of hosts of a given size
type IPv6AllocatedBlock struct {
	allocatedBlock
}

// GetAddress returns the block
func (alloc IPv6AllocatedBlock) GetAddress() *IPv6Address {
	return alloc.block.ToIPv6()
}

func cloneIPAllocatedBlocksToIPv6AllocatedBlocks(orig []IPAllocatedBlock) []IPv6AllocatedBlock {
	result := make([]IPv6AllocatedBlock, len(orig))
	for i := range orig {
		result[i] = orig[i].toIPv6()
	}
	return result
}
