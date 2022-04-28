//
// Copyright 2022 Sean C Foley
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

package ipaddr

import (
	"fmt"
	"github.com/seancfoley/bintree/tree"
	"math/bits"
	"unsafe"
)

// We use pointer receivers for addressTrieKey to implement TrieKey because interfaces store pointers to type data anyway.
// So we gain nothing by using value receivers instead.

type addressTrieKey struct {
	*Address
}

var _ tree.TrieKey = &addressTrieKey{}

func (a *addressTrieKey) GetPrefixLen() tree.PrefixLen {
	return tree.PrefixLen(a.Address.GetPrefixLen())
}

// ToPrefixBlockLen returns the address key associated with the prefix length provided,
// the address key whose prefix of that length matches the prefix of this address key, and the remaining bits span all values.
//
// The returned address key will represent all addresses with the same prefix as this one, the prefix "block".
func (a *addressTrieKey) ToPrefixBlockLen(bitCount BitCount) tree.TrieKey {
	return &addressTrieKey{a.Address.ToPrefixBlockLen(bitCount)}
}

// Compare compares to provide the same ordering used by the trie,
// an ordering that works with prefix block subnets and individual addresses.
// The comparator is consistent with the equality of address instances
// and can be used in other contexts.  However, it only works with prefix blocks and individual addresses,
// not with addresses like 1-2.3.4.5-6 which cannot be differentiated using this comparator from 1.3.4.5
// and is thus not consistent with equality for subnets that are not CIDR prefix blocks.
//
// The comparator first compares the prefix of addresses, with the full address value considered the prefix when
// there is no prefix length, ie when it is a single address.  It takes the minimum m of the two prefix lengths and
// compares those m prefix bits in both addresses.  The ordering is determined by which of those two values is smaller or larger.
//
// If both prefix lengths match then both addresses are equal.
// Otherwise it looks at bit m in the address with larger prefix.  If 1 it is larger and if 0 it is smaller than the other.
//
// When comparing an address with a prefix p and an address without, the first p bits in both are compared, and if equal,
// the bit at index p in the non-prefixed address determines the ordering, if 1 it is larger and if 0 it is smaller than the other.
//
// When comparing an address with prefix length matching the bit count to an address with no prefix, they are considered equal if the bits match.
// For instance, 1.2.3.4/32 is equal to 1.2.3.4, and thus the trie does not allow 1.2.3.4/32 in the trie since it is indistinguishable from 1.2.3.4,
// instead 1.2.3.4/32 is converted to 1.2.3.4 when inserted into the trie.
//
// When comparing 0.0.0.0/0, which has no prefix, to other addresses, the first bit in the other address determines the ordering.
// If 1 it is larger and if 0 it is smaller than 0.0.0.0/0.
func (a *addressTrieKey) Compare(other tree.TrieKey) int {
	res, _ := a.Address.TrieCompare(other.(*addressTrieKey).Address)
	return res
}

// MatchBits returns false if we need to keep going and try to match sub-nodes
// MatchBits returns true if the bits do not match, or the bits match to the very end
func (a addressTrieKey) MatchBits(key tree.TrieKey, bitIndex int, handleMatch tree.KeyCompareResult) bool {
	existingAddr := key.(*addressTrieKey).Address
	bitsPerSegment := existingAddr.GetBitsPerSegment()
	bytesPerSegment := existingAddr.GetBytesPerSegment()
	newAddr := a.Address
	segmentIndex := getHostSegmentIndex(bitIndex, bytesPerSegment, bitsPerSegment)
	segmentCount := existingAddr.GetSegmentCount()
	if newAddr.GetSegmentCount() != segmentCount || bitsPerSegment != newAddr.GetBitsPerSegment() {
		panic("mismatched bit length between address trie keys")
	}
	existingPref := existingAddr.GetPrefixLen()
	newPrefLen := newAddr.GetPrefixLen()

	// this block handles cases like where we matched matching ::ffff:102:304 to ::ffff:102:304/127,
	// and we found a subnode to match, but we know the final bit is a match due to the subnode being lower or upper,
	// so there is actually not more bits to match
	if segmentIndex >= segmentCount {
		// all the bits match
		handleMatch.BitsMatch()
		return true
	}

	bitsMatchedSoFar := segmentIndex * bitsPerSegment
	for {
		existingSegment := existingAddr.getSegment(segmentIndex)
		newSegment := newAddr.getSegment(segmentIndex)
		segmentPref := getSegmentPrefLen(existingAddr, existingPref, bitsPerSegment, bitsMatchedSoFar, existingSegment)
		newSegmentPref := getSegmentPrefLen(newAddr, newPrefLen, bitsPerSegment, bitsMatchedSoFar, newSegment)
		if segmentPref != nil {
			segmentPrefLen := segmentPref.Len()
			newPrefixLen := newSegmentPref.Len()
			if newSegmentPref != nil && newPrefixLen <= segmentPrefLen {
				matchingBits := getMatchingBits(existingSegment, newSegment, newPrefixLen, bitsPerSegment)
				if matchingBits >= newPrefixLen {
					handleMatch.BitsMatch()
				} else {
					// no match - the bits don't match
					// matchingBits < newPrefLen < segmentPrefLen
					handleMatch.BitsDoNotMatch(bitsMatchedSoFar + matchingBits)
				}
			} else {
				matchingBits := getMatchingBits(existingSegment, newSegment, segmentPrefLen, bitsPerSegment)
				if matchingBits >= segmentPrefLen { // match - the current subnet/address is a match so far, and we must go further to check smaller subnets
					return false
				}
				// matchingBits < segmentPrefLen - no match - the bits in current prefix do not match the prefix of the existing address
				handleMatch.BitsDoNotMatch(bitsMatchedSoFar + matchingBits)
			}
			return true
		} else if newSegmentPref != nil {
			newPrefixLen := newSegmentPref.Len()
			matchingBits := getMatchingBits(existingSegment, newSegment, newPrefixLen, bitsPerSegment)
			if matchingBits >= newPrefixLen { // the current bits match the current prefix, but the existing has no prefix
				handleMatch.BitsMatch()
			} else {
				// no match - the current subnet does not match the existing address
				handleMatch.BitsDoNotMatch(bitsMatchedSoFar + matchingBits)
			}
			return true
		} else {
			matchingBits := getMatchingBits(existingSegment, newSegment, bitsPerSegment, bitsPerSegment)
			if matchingBits < bitsPerSegment { // no match - the current subnet/address is not here
				handleMatch.BitsDoNotMatch(bitsMatchedSoFar + matchingBits)
				return true
			} else {
				segmentIndex++
				if segmentIndex == segmentCount { // match - the current subnet/address is a match
					// note that "added" is already true here, we can only be here if explicitly inserted already since it is a non-prefixed full address
					handleMatch.BitsMatch()
					return true
				}
			}
			bitsMatchedSoFar += bitsPerSegment
		}
	}
}

// ToMaxLower changes this key to a new key with a 0 at the given index followed by all ones, and with no prefix length
func (a addressTrieKey) ToMaxLower() tree.TrieKey {
	return &addressTrieKey{a.Address.toMaxLower()}
}

// ToMinUpper changes this key to a new key with a 1 at the given index followed by all zeros, and with no prefix length
func (a addressTrieKey) ToMinUpper() tree.TrieKey {
	return &addressTrieKey{a.Address.toMinUpper()}
}

func getSegmentPrefLen(
	addr AddressSegmentSeries,
	prefLen PrefixLen,
	bitsPerSegment,
	bitsMatchedSoFar BitCount,
	segment *AddressSegment) PrefixLen {
	if ipSeg := segment.ToIP(); ipSeg != nil {
		return ipSeg.GetSegmentPrefixLen()
	} else if prefLen != nil {
		result := prefLen.Len() - bitsMatchedSoFar
		if result <= bitsPerSegment {
			if result < 0 {
				result = 0
			}
			return cacheBitCount(result)
		}
	}
	return nil
}

func getMatchingBits(segment1, segment2 *AddressSegment, maxBits, bitsPerSegment BitCount) BitCount {
	if maxBits == 0 {
		return 0
	}
	val1 := segment1.getSegmentValue()
	val2 := segment2.getSegmentValue()
	xor := val1 ^ val2
	switch bitsPerSegment {
	case IPv4BitsPerSegment:
		return BitCount(bits.LeadingZeros8(uint8(xor)))
	case IPv6BitsPerSegment:
		return BitCount(bits.LeadingZeros16(uint16(xor)))
	default:
		return BitCount(bits.LeadingZeros32(xor)) - 32 + bitsPerSegment
	}
}

type addressTrieNode struct {
	tree.BinTrieNode
}

// getKey gets the key used for placing the node in the tree.
func (node *addressTrieNode) getKey() *Address {
	val := node.toTrieNode().GetKey()
	if val == nil {
		return nil
	}
	key := val.(*addressTrieKey)
	return key.Address
}

func (node *addressTrieNode) get(addr *Address) NodeValue {
	addr = mustBeBlockOrAddress(addr)
	return node.toTrieNode().Get(&addressTrieKey{addr})
}

func (node *addressTrieNode) lowerAddedNode(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(node.toTrieNode().LowerAddedNode(&addressTrieKey{addr}))
}

func (node *addressTrieNode) floorAddedNode(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(node.toTrieNode().FloorAddedNode(&addressTrieKey{addr}))
}

func (node *addressTrieNode) higherAddedNode(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(node.toTrieNode().HigherAddedNode(&addressTrieKey{addr}))
}

func (node *addressTrieNode) ceilingAddedNode(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(node.toTrieNode().CeilingAddedNode(&addressTrieKey{addr}))
}

// iterator returns an iterator that iterates through the elements of the sub-tree with this node as the root.
// The iteration is in sorted element order.
func (node *addressTrieNode) iterator() AddressIterator {
	return addressKeyIterator{node.toTrieNode().Iterator()}
}

// descendingIterator returns an iterator that iterates through the elements of the subtrie with this node as the root.
// The iteration is in reverse sorted element order.
func (node *addressTrieNode) descendingIterator() AddressIterator {
	return addressKeyIterator{node.toTrieNode().DescendingIterator()}
}

// nodeIterator iterates through the added nodes of the sub-tree with this node as the root, in forward or reverse tree order.
func (node *addressTrieNode) nodeIterator(forward bool) AddressTrieNodeIteratorRem {
	return addrTrieNodeIteratorRem{node.toTrieNode().NodeIterator(forward)}
}

// allNodeIterator iterates through all the nodes of the sub-tree with this node as the root, in forward or reverse tree order.
func (node *addressTrieNode) allNodeIterator(forward bool) AddressTrieNodeIteratorRem {
	return addrTrieNodeIteratorRem{node.toTrieNode().AllNodeIterator(forward)}
}

// blockSizeNodeIterator iterates the added nodes, ordered by keys from largest prefix blocks to smallest and then to individual addresses,
// in the sub-trie with this node as the root.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order is taken.
func (node *addressTrieNode) blockSizeNodeIterator(lowerSubNodeFirst bool) AddressTrieNodeIteratorRem {
	return addrTrieNodeIteratorRem{node.toTrieNode().BlockSizeNodeIterator(lowerSubNodeFirst)}
}

// blockSizeAllNodeIterator iterates all the nodes, ordered by keys from largest prefix blocks to smallest and then to individual addresses,
// in the sub-trie with this node as the root.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (node *addressTrieNode) blockSizeAllNodeIterator(lowerSubNodeFirst bool) AddressTrieNodeIteratorRem {
	return addrTrieNodeIteratorRem{node.toTrieNode().BlockSizeAllNodeIterator(lowerSubNodeFirst)}
}

// blockSizeCachingAllNodeIterator iterates all nodes, ordered by keys from largest prefix blocks to smallest and then to individual addresses,
// in the sub-trie with this node as the root.
func (node *addressTrieNode) blockSizeCachingAllNodeIterator() CachingAddressTrieNodeIterator {
	return cachingAddressTrieNodeIterator{node.toTrieNode().BlockSizeCachingAllNodeIterator()}
}

func (node *addressTrieNode) containingFirstIterator(forwardSubNodeOrder bool) CachingAddressTrieNodeIterator {
	return cachingAddressTrieNodeIterator{node.toTrieNode().ContainingFirstIterator(forwardSubNodeOrder)}
}

func (node *addressTrieNode) containingFirstAllNodeIterator(forwardSubNodeOrder bool) CachingAddressTrieNodeIterator {
	return cachingAddressTrieNodeIterator{node.toTrieNode().ContainingFirstAllNodeIterator(forwardSubNodeOrder)}
}

func (node *addressTrieNode) containedFirstIterator(forwardSubNodeOrder bool) AddressTrieNodeIteratorRem {
	return addrTrieNodeIteratorRem{node.toTrieNode().ContainedFirstIterator(forwardSubNodeOrder)}
}

func (node *addressTrieNode) containedFirstAllNodeIterator(forwardSubNodeOrder bool) AddressTrieNodeIterator {
	return addrTrieNodeIterator{node.toTrieNode().ContainedFirstAllNodeIterator(forwardSubNodeOrder)}
}

func (node *addressTrieNode) contains(addr *Address) bool {
	addr = mustBeBlockOrAddress(addr)
	return node.toTrieNode().Contains(&addressTrieKey{addr})
}

func (node *addressTrieNode) removeNode(addr *Address) bool {
	addr = mustBeBlockOrAddress(addr)
	return node.toTrieNode().RemoveNode(&addressTrieKey{addr})
}

func (node *addressTrieNode) removeElementsContainedBy(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(node.toTrieNode().RemoveElementsContainedBy(&addressTrieKey{addr}))
}

func (node *addressTrieNode) elementsContainedBy(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(node.toTrieNode().ElementsContainedBy(&addressTrieKey{addr}))
}

func (node *addressTrieNode) elementsContaining(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(node.toTrieNode().ElementsContaining(&addressTrieKey{addr}))
}

func (node *addressTrieNode) longestPrefixMatch(addr *Address) *Address {
	addr = mustBeBlockOrAddress(addr)
	return node.toTrieNode().LongestPrefixMatch(&addressTrieKey{addr}).(*addressTrieKey).Address
}

func (node *addressTrieNode) longestPrefixMatchNode(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(node.toTrieNode().LongestPrefixMatchNode(&addressTrieKey{addr}))
}

func (node *addressTrieNode) elementContains(addr *Address) bool {
	addr = mustBeBlockOrAddress(addr)
	return node.toTrieNode().ElementContains(&addressTrieKey{addr})
}

func (node *addressTrieNode) getNode(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(node.toTrieNode().GetNode(&addressTrieKey{addr}))
}

func (node *addressTrieNode) getAddedNode(addr *Address) *AddressTrieNode {
	addr = mustBeBlockOrAddress(addr)
	return toAddressTrieNode(node.toTrieNode().GetAddedNode(&addressTrieKey{addr}))
}

func (node *addressTrieNode) toTrieNode() *tree.BinTrieNode {
	return (*tree.BinTrieNode)(unsafe.Pointer(node))
}

func toAddressTrieNode(node *tree.BinTrieNode) *AddressTrieNode {
	return (*AddressTrieNode)(unsafe.Pointer(node))
}

func toAssociativeAddressTrieNode(node *tree.BinTrieNode) *AssociativeAddressTrieNode {
	return (*AssociativeAddressTrieNode)(unsafe.Pointer(node))
}

//
//
//
//
//
//
//
//
//
//
//

type AddressTrieNode struct {
	addressTrieNode
}

func (node *AddressTrieNode) toTrieNode() *tree.BinTrieNode {
	return (*tree.BinTrieNode)(unsafe.Pointer(node))
}

func (node *AddressTrieNode) ToIPv4() *IPv4AddressTrieNode {
	if node == nil || node.GetKey().IsIPv4() {
		return (*IPv4AddressTrieNode)(node)
	}
	return nil
}

func (node *AddressTrieNode) ToIPv6() *IPv6AddressTrieNode {
	if node == nil || node.GetKey().IsIPv6() {
		return (*IPv6AddressTrieNode)(node)
	}
	return nil
}

func (node *AddressTrieNode) ToMAC() *MACAddressTrieNode {
	if node == nil || node.GetKey().IsMAC() {
		return (*MACAddressTrieNode)(node)
	}
	return nil
}

func (node *AddressTrieNode) ToAssociative() *AssociativeAddressTrieNode {
	return (*AssociativeAddressTrieNode)(node)
}

func (node *AddressTrieNode) ToIPv4Associative() *IPv4AddressAssociativeTrieNode {
	if node == nil || node.GetKey().IsIPv4() {
		return (*IPv4AddressAssociativeTrieNode)(node)
	}
	return nil
}

func (node *AddressTrieNode) ToIPv6Associative() *IPv6AddressAssociativeTrieNode {
	if node == nil || node.GetKey().IsIPv6() {
		return (*IPv6AddressAssociativeTrieNode)(node)
	}
	return nil
}

func (node *AddressTrieNode) ToMACAssociative() *MACAddressAssociativeTrieNode {
	if node == nil || node.GetKey().IsMAC() {
		return (*MACAddressAssociativeTrieNode)(node)
	}
	return nil
}

// tobase is used to convert the pointer rather than doing a field dereference, so that nil pointer handling can be done in *addressTrieNode
func (node *AddressTrieNode) tobase() *addressTrieNode {
	return (*addressTrieNode)(unsafe.Pointer(node))
}

// GetKey gets the key used for placing the node in the tree.
func (node *AddressTrieNode) GetKey() *Address {
	return node.tobase().getKey()
}

// IsRoot returns whether this is the root of the backing tree.
func (node *AddressTrieNode) IsRoot() bool {
	return node.toTrieNode().IsRoot()
}

// IsAdded returns whether the node was "added".
// Some binary tree nodes are considered "added" and others are not.
// Those nodes created for key elements added to the tree are "added" nodes.
// Those that are not added are those nodes created to serve as junctions for the added nodes.
// Only added elements contribute to the size of a tree.
// When removing nodes, non-added nodes are removed automatically whenever they are no longer needed,
// which is when an added node has less than two added sub-nodes.
func (node *AddressTrieNode) IsAdded() bool {
	return node.toTrieNode().IsAdded()
}

// SetAdded makes this node an added node, which is equivalent to adding the corresponding key to the tree.
// If the node is already an added node, this method has no effect.
// You cannot set an added node to non-added, for that you should Remove the node from the tree by calling Remove.
// A non-added node will only remain in the tree if it needs to in the tree.
func (node *AddressTrieNode) SetAdded() {
	node.toTrieNode().SetAdded()
}

// Clear removes this node and all sub-nodes from the tree, after which isEmpty() will return true.
func (node *AddressTrieNode) Clear() {
	node.toTrieNode().Clear()
}

// IsLeaf returns whether this node is in the tree (a node for which IsAdded() is true)
// and there are no elements in the sub-tree with this node as the root.
func (node *AddressTrieNode) IsLeaf() bool {
	return node.toTrieNode().IsLeaf()
}

// GetUpperSubNode gets the direct child node whose key is largest in value
func (node *AddressTrieNode) GetUpperSubNode() *AddressTrieNode {
	return toAddressTrieNode(node.toTrieNode().GetUpperSubNode())
}

// GetLowerSubNode gets the direct child node whose key is smallest in value
func (node *AddressTrieNode) GetLowerSubNode() *AddressTrieNode {
	return toAddressTrieNode(node.toTrieNode().GetLowerSubNode())
}

// GetParent gets the node from which this node is a direct child node, or nil if this is the root.
func (node *AddressTrieNode) GetParent() *AddressTrieNode {
	return toAddressTrieNode(node.toTrieNode().GetParent())
}

// PreviousAddedNode returns the previous node in the tree that is an added node, following the tree order in reverse,
// or nil if there is no such node.
func (node *AddressTrieNode) PreviousAddedNode() *AddressTrieNode {
	return toAddressTrieNode(node.toTrieNode().PreviousAddedNode())
}

// NextAddedNode returns the next node in the tree that is an added node, following the tree order,
// or nil if there is no such node.
func (node *AddressTrieNode) NextAddedNode() *AddressTrieNode {
	return toAddressTrieNode(node.toTrieNode().NextAddedNode())
}

// NextNode returns the node that follows this node following the tree order
func (node *AddressTrieNode) NextNode() *AddressTrieNode {
	return toAddressTrieNode(node.toTrieNode().NextNode())
}

// PreviousNode eturns the node that precedes this node following the tree order.
func (node *AddressTrieNode) PreviousNode() *AddressTrieNode {
	return toAddressTrieNode(node.toTrieNode().PreviousNode())
}

func (node *AddressTrieNode) FirstNode() *AddressTrieNode {
	return toAddressTrieNode(node.toTrieNode().FirstNode())
}

func (node *AddressTrieNode) FirstAddedNode() *AddressTrieNode {
	return toAddressTrieNode(node.toTrieNode().FirstAddedNode())
}

func (node *AddressTrieNode) LastNode() *AddressTrieNode {
	return toAddressTrieNode(node.toTrieNode().LastNode())
}

func (node *AddressTrieNode) LastAddedNode() *AddressTrieNode {
	return toAddressTrieNode(node.toTrieNode().LastAddedNode())
}

// LowerAddedNode returns the added node, in this sub-trie with this node as root, whose address is the highest address strictly less than the given address.
func (node *AddressTrieNode) LowerAddedNode(addr *Address) *AddressTrieNode {
	return node.tobase().lowerAddedNode(addr)
}

// FloorAddedNode returns the added node, in this sub-trie with this node as root, whose address is the highest address less than or equal to the given address.
func (node *AddressTrieNode) FloorAddedNode(addr *Address) *AddressTrieNode {
	return node.tobase().floorAddedNode(addr)
}

// HigherAddedNode returns the added node, in this sub-trie with this node as root, whose address is the lowest address strictly greater than the given address.
func (node *AddressTrieNode) HigherAddedNode(addr *Address) *AddressTrieNode {
	return node.tobase().higherAddedNode(addr)
}

// CeilingAddedNode returns the added node, in this sub-trie with this node as root, whose address is the lowest address greater than or equal to the given address.
func (node *AddressTrieNode) CeilingAddedNode(addr *Address) *AddressTrieNode {
	return node.tobase().ceilingAddedNode(addr)
}

// Iterator returns an iterator that iterates through the elements of the sub-tree with this node as the root.
// The iteration is in sorted element order.
func (node *AddressTrieNode) Iterator() AddressIterator {
	return node.tobase().iterator()
}

// DescendingIterator returns an iterator that iterates through the elements of the subtrie with this node as the root.
// The iteration is in reverse sorted element order.
func (node *AddressTrieNode) DescendingIterator() AddressIterator {
	return node.tobase().descendingIterator()
}

// NodeIterator returns an iterator that iterates through the added nodes of the sub-tree with this node as the root, in forward or reverse tree order.
func (node *AddressTrieNode) NodeIterator(forward bool) AddressTrieNodeIteratorRem {
	return node.tobase().nodeIterator(forward)
}

// AllNodeIterator returns an iterator that iterates through all the nodes of the sub-tree with this node as the root, in forward or reverse tree order.
func (node *AddressTrieNode) AllNodeIterator(forward bool) AddressTrieNodeIteratorRem {
	return node.tobase().allNodeIterator(forward)
}

// BlockSizeNodeIterator returns an iterator that iterates the added nodes, ordered by keys from largest prefix blocks to smallest and then to individual addresses,
// in the sub-trie with this node as the root.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order is taken.
func (node *AddressTrieNode) BlockSizeNodeIterator(lowerSubNodeFirst bool) AddressTrieNodeIteratorRem {
	return node.tobase().blockSizeNodeIterator(lowerSubNodeFirst)
}

// BlockSizeAllNodeIterator returns an iterator that iterates all the nodes, ordered by keys from largest prefix blocks to smallest and then to individual addresses,
// in the sub-trie with this node as the root.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (node *AddressTrieNode) BlockSizeAllNodeIterator(lowerSubNodeFirst bool) AddressTrieNodeIteratorRem {
	return node.tobase().blockSizeAllNodeIterator(lowerSubNodeFirst)
}

// BlockSizeCachingAllNodeIterator returns an iterator that iterates all nodes, ordered by keys from largest prefix blocks to smallest and then to individual addresses,
// in the sub-trie with this node as the root.
func (node *AddressTrieNode) BlockSizeCachingAllNodeIterator() CachingAddressTrieNodeIterator {
	return node.tobase().blockSizeCachingAllNodeIterator()
}

func (node *AddressTrieNode) ContainingFirstIterator(forwardSubNodeOrder bool) CachingAddressTrieNodeIterator {
	return node.tobase().containingFirstIterator(forwardSubNodeOrder)
}

func (node *AddressTrieNode) ContainingFirstAllNodeIterator(forwardSubNodeOrder bool) CachingAddressTrieNodeIterator {
	return node.tobase().containingFirstAllNodeIterator(forwardSubNodeOrder)
}

func (node *AddressTrieNode) ContainedFirstIterator(forwardSubNodeOrder bool) AddressTrieNodeIteratorRem {
	return node.tobase().containedFirstIterator(forwardSubNodeOrder)
}

func (node *AddressTrieNode) ContainedFirstAllNodeIterator(forwardSubNodeOrder bool) AddressTrieNodeIterator {
	return node.tobase().containedFirstAllNodeIterator(forwardSubNodeOrder)
}

// Clone clones the node.
// Keys remain the same, but the parent node and the lower and upper sub-nodes are all set to nil.
func (node *AddressTrieNode) Clone() *AddressTrieNode {
	return toAddressTrieNode(node.toTrieNode().Clone())
}

// CloneTree clones the sub-tree starting with this node as root.
// The nodes are cloned, but their keys and values are not cloned.
func (node *AddressTrieNode) CloneTree() *AddressTrieNode {
	return toAddressTrieNode(node.toTrieNode().CloneTree())
}

// AsNewTrie creates a new sub-trie, copying the nodes starting with this node as root.
// The nodes are copies of the nodes in this sub-trie, but their keys and values are not copies.
func (node *AddressTrieNode) AsNewTrie() *AddressTrie {
	return toAddressTrie(node.toTrieNode().AsNewTrie())
}

// Compare returns -1, 0 or 1 if this node is less than, equal, or greater than the other, according to the key and the trie order.
func (node *AddressTrieNode) Compare(other *AddressTrieNode) int {
	return node.toTrieNode().Compare(other.toTrieNode())
}

// Equal returns whether the address and and mapped value match those of the given node
func (node *AddressTrieNode) Equal(other *AddressTrieNode) bool {
	return node.toTrieNode().Equal(other.toTrieNode())
}

// TreeEqual returns whether the sub-tree represented by this node as the root node matches the given sub-tree
func (node *AddressTrieNode) TreeEqual(other *AddressTrieNode) bool {
	return node.toTrieNode().TreeEqual(other.toTrieNode())
}

// Remove removes this node from the collection of added nodes, and also from the trie if possible.
// If it has two sub-nodes, it cannot be removed from the trie, in which case it is marked as not "added",
// nor is it counted in the trie size.
// Only added nodes can be removed from the trie.  If this node is not added, this method does nothing.
func (node *AddressTrieNode) Remove() {
	node.toTrieNode().Remove()
}

// Contains returns whether the given address or prefix block subnet is in the sub-trie, as an added element, with this node as the root.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the prefix block or address address exists already in the trie, false otherwise.
//
// Use GetAddedNode  to get the node for the address rather than just checking for its existence.
func (node *AddressTrieNode) Contains(addr *Address) bool {
	return node.tobase().contains(addr)
}

// RemoveNode removes the given single address or prefix block subnet from the trie with this node as the root.
//
// Removing an element will not remove contained elements (nodes for contained blocks and addresses).
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the prefix block or address was removed, false if not already in the trie.
//
// You can also remove by calling GetAddedNode to get the node and then calling Remove on the node.
//
// When an address is removed, the corresponding node may remain in the trie if it remains a subnet block for two sub-nodes.
// If the corresponding node can be removed from the trie, it will be.
func (node *AddressTrieNode) RemoveNode(addr *Address) bool {
	return node.tobase().removeNode(addr)
}

// RemoveElementsContainedBy removes any single address or prefix block subnet from the trie, with this node as the root, that is contained in the given individual address or prefix block subnet.
//
// Goes further than Remove, not requiring a match to an inserted node, and also removing all the sub-nodes of any removed node or sub-node.
//
// For example, after inserting 1.2.3.0 and 1.2.3.1, passing 1.2.3.0/31 to RemoveElementsContainedBy will remove them both,
// while the Remove method will remove nothing.
// After inserting 1.2.3.0/31, then #remove(Address) will remove 1.2.3.0/31, but will leave 1.2.3.0 and 1.2.3.1 in the trie.
//
// It cannot partially delete a node, such as deleting a single address from a prefix block represented by a node.
// It can only delete the whole node if the whole address or block represented by that node is contained in the given address or block.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
//Returns the root node of the subtrie that was removed from the trie, or nil if nothing was removed.
func (node *AddressTrieNode) RemoveElementsContainedBy(addr *Address) *AddressTrieNode {
	return node.tobase().removeElementsContainedBy(addr)
}

// ElementsContainedBy checks if a part of this trie, with this node as the root, is contained by the given prefix block subnet or individual address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the root node of the contained subtrie, or nil if no subtrie is contained.
// The node returned need not be an "added" node, see IsAdded for more details on added nodes.
// The returned subtrie is backed by this trie, so changes in this trie are reflected in those nodes and vice-versa.
func (node *AddressTrieNode) ElementsContainedBy(addr *Address) *AddressTrieNode {
	return node.tobase().elementsContainedBy(addr)
}

// ElementsContaining finds the trie nodes in the trie, with this sub-node as the root,
// containing the given key and returns them as a linked list.
// Only added nodes are added to the linked list
//
// If the argument is not a single address nor prefix block, this method will panic.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (node *AddressTrieNode) ElementsContaining(addr *Address) *AddressTrieNode {
	return node.tobase().elementsContaining(addr)
}

// LongestPrefixMatch returns the address pr subnet with the longest prefix of all the added subnets or address whose prefix matches the given address.
// This is equivalent to finding the containing subnet or address with the smallest subnet size.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns nil if no added subnet or address contains the given argument.
//
// Use ElementContains to check for the existence of a containing address.
// To get all the containing addresses (subnets with matching prefix), use ElementsContaining.
// To get the node corresponding to the result of this method, use LongestPrefixMatchNode.
func (node *AddressTrieNode) LongestPrefixMatch(addr *Address) *Address {
	return node.tobase().longestPrefixMatch(addr)
}

// LongestPrefixMatchNode finds the containing subnet or address in the trie with the smallest subnet size,
// which is equivalent to finding the subnet or address with the longest matching prefix.
// Returns the node corresponding to that subnet.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns nil if no added subnet or address contains the given argument.
//
// Use ElementContains to check for the existence of a containing address.
// To get all the containing addresses, use ElementsContaining.
// Use LongestPrefixMatch to get only the address corresponding to the result of this method.
func (node *AddressTrieNode) LongestPrefixMatchNode(addr *Address) *AddressTrieNode {
	return node.tobase().longestPrefixMatchNode(addr)
}

// ElementContains checks if a prefix block subnet or address in the trie, with this node as the root, contains the given subnet or address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the subnet or address is contained by a trie element, false otherwise.
//
// To get all the containing addresses, use ElementsContaining.
func (node *AddressTrieNode) ElementContains(addr *Address) bool {
	return node.tobase().elementContains(addr)
}

// GetNode gets the node in the trie, with this subnode as the root, corresponding to the given address,
// or returns nil if not such element exists.
//
// It returns any node, whether added or not,
// including any prefix block node that was not added.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (node *AddressTrieNode) GetNode(addr *Address) *AddressTrieNode {
	return node.tobase().getNode(addr)
}

// GetAddedNode gets trie nodes representing added elements.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Use Contains to check for the existence of a given address in the trie,
// as well as GetNode to search for all nodes including those not-added but also auto-generated nodes for subnet blocks.
func (node *AddressTrieNode) GetAddedNode(addr *Address) *AddressTrieNode {
	return node.tobase().getAddedNode(addr)
}

// NodeSize returns the number of nodes in the trie with this node as the root, which is more than the number of added addresses or blocks.
func (node *AddressTrieNode) NodeSize() int {
	//if node == nil {
	//	return 0
	//}
	//return node.nodeSize()
	return node.toTrieNode().NodeSize()
}

// Size returns the number of elements in the tree.
// Only nodes for which IsAdded returns true are counted.
// When zero is returned, IsEmpty returns true.
func (node *AddressTrieNode) Size() int {
	//if node == nil {
	//	return 0
	//}
	//return node.size()
	return node.toTrieNode().Size()
}

// IsEmpty returns whether the size is 0
func (node *AddressTrieNode) IsEmpty() bool {
	return node.Size() == 0
}

// TreeString returns a visual representation of the sub-tree with this node as root, with one node per line.
//
// withNonAddedKeys: whether to show nodes that are not added nodes
// withSizes: whether to include the counts of added nodes in each sub-tree
func (node *AddressTrieNode) TreeString(withNonAddedKeys, withSizes bool) string {
	return node.toTrieNode().TreeString(withNonAddedKeys, withSizes)
}

// String returns a visual representation of this node including the key, with an open circle indicating this node is not an added node,
// a closed circle indicating this node is an added node.
func (node *AddressTrieNode) String() string {
	return node.toTrieNode().String()
}

// For some reason Format must be here and not in addressTrieNode for nil node.
// It panics in fmt code either way, but if in here then it is handled by a recover() call in fmt properly in the debugger.

// Format implements the fmt.Formatter interface
func (node AddressTrieNode) Format(state fmt.State, verb rune) {
	node.toTrieNode().Format(state, verb)
}

//
//
//
//
//
//
//
//
//
//
//
//

// NodeValue represents the value stored with each node in an associative trie.
type NodeValue = tree.V

//
//
//
//
//

type AssociativeAddressTrieNode struct {
	addressTrieNode
}

func (node *AssociativeAddressTrieNode) toTrieNode() *tree.BinTrieNode {
	return (*tree.BinTrieNode)(unsafe.Pointer(node))
}

func (node *AssociativeAddressTrieNode) toBase() *addressTrieNode {
	return (*addressTrieNode)(unsafe.Pointer(node))
}

func (node *AssociativeAddressTrieNode) ToBase() *AddressTrieNode {
	return (*AddressTrieNode)(node)
}

func (node *AssociativeAddressTrieNode) ToIPv4() *IPv4AddressAssociativeTrieNode {
	if node == nil || node.GetKey().IsIPv4() {
		return (*IPv4AddressAssociativeTrieNode)(node)
	}
	return nil
}

func (node *AssociativeAddressTrieNode) ToIPv6() *IPv6AddressAssociativeTrieNode {
	if node == nil || node.GetKey().IsIPv6() {
		return (*IPv6AddressAssociativeTrieNode)(node)
	}
	return nil
}

func (node *AssociativeAddressTrieNode) ToMAC() *MACAddressAssociativeTrieNode {
	if node == nil || node.GetKey().IsMAC() {
		return (*MACAddressAssociativeTrieNode)(node)
	}
	return nil
}

// GetKey gets the key used for placing the node in the tree.
func (node *AssociativeAddressTrieNode) GetKey() *Address {
	return node.toBase().getKey()
}

// IsRoot returns whether this is the root of the backing tree.
func (node *AssociativeAddressTrieNode) IsRoot() bool {
	return node.toTrieNode().IsRoot()
}

// IsAdded returns whether the node was "added".
// Some binary tree nodes are considered "added" and others are not.
// Those nodes created for key elements added to the tree are "added" nodes.
// Those that are not added are those nodes created to serve as junctions for the added nodes.
// Only added elements contribute to the size of a tree.
// When removing nodes, non-added nodes are removed automatically whenever they are no longer needed,
// which is when an added node has less than two added sub-nodes.
func (node *AssociativeAddressTrieNode) IsAdded() bool {
	return node.toTrieNode().IsAdded()
}

// SetAdded makes this node an added node, which is equivalent to adding the corresponding key to the tree.
// If the node is already an added node, this method has no effect.
// You cannot set an added node to non-added, for that you should Remove the node from the tree by calling Remove.
// A non-added node will only remain in the tree if it needs to in the tree.
func (node *AssociativeAddressTrieNode) SetAdded() {
	node.toTrieNode().SetAdded()
}

// Clear removes this node and all sub-nodes from the tree, after which isEmpty() will return true.
func (node *AssociativeAddressTrieNode) Clear() {
	node.toTrieNode().Clear()
}

// IsLeaf returns whether this node is in the tree (a node for which IsAdded() is true)
// and there are no elements in the sub-tree with this node as the root.
func (node *AssociativeAddressTrieNode) IsLeaf() bool {
	return node.toTrieNode().IsLeaf()
}

// ClearValue makes the value associated with this node the nil value
func (node *AssociativeAddressTrieNode) ClearValue() {
	node.toTrieNode().ClearValue()
}

// SetValue sets the value associated with this node
func (node *AssociativeAddressTrieNode) SetValue(val NodeValue) {
	node.toTrieNode().SetValue(val)
}

// GetValue sets the value associated with this node
func (node *AssociativeAddressTrieNode) GetValue() NodeValue {
	return node.toTrieNode().GetValue()
}

// GetUpperSubNode gets the direct child node whose key is largest in value
func (node *AssociativeAddressTrieNode) GetUpperSubNode() *AssociativeAddressTrieNode {
	return toAssociativeAddressTrieNode(node.toTrieNode().GetUpperSubNode())
}

// GetLowerSubNode gets the direct child node whose key is smallest in value
func (node *AssociativeAddressTrieNode) GetLowerSubNode() *AssociativeAddressTrieNode {
	return toAssociativeAddressTrieNode(node.toTrieNode().GetLowerSubNode())
}

// GetParent gets the node from which this node is a direct child node, or nil if this is the root.
func (node *AssociativeAddressTrieNode) GetParent() *AssociativeAddressTrieNode {
	return toAssociativeAddressTrieNode(node.toTrieNode().GetParent())
}

// PreviousAddedNode returns the first added node that precedes this node following the tree order
func (node *AssociativeAddressTrieNode) PreviousAddedNode() *AssociativeAddressTrieNode {
	return toAssociativeAddressTrieNode(node.toTrieNode().PreviousAddedNode())
}

// NextAddedNode returns the first added node that follows this node following the tree order
func (node *AssociativeAddressTrieNode) NextAddedNode() *AssociativeAddressTrieNode {
	return toAssociativeAddressTrieNode(node.toTrieNode().NextAddedNode())
}

// NextNode returns the node that follows this node following the tree order
func (node *AssociativeAddressTrieNode) NextNode() *AssociativeAddressTrieNode {
	return toAssociativeAddressTrieNode(node.toTrieNode().NextNode())
}

// PreviousNode returns the node that precedes this node following the tree order.
func (node *AssociativeAddressTrieNode) PreviousNode() *AssociativeAddressTrieNode {
	return toAssociativeAddressTrieNode(node.toTrieNode().PreviousNode())
}

func (node *AssociativeAddressTrieNode) FirstNode() *AssociativeAddressTrieNode {
	return toAssociativeAddressTrieNode(node.toTrieNode().FirstNode())
}

func (node *AssociativeAddressTrieNode) FirstAddedNode() *AssociativeAddressTrieNode {
	return toAssociativeAddressTrieNode(node.toTrieNode().FirstAddedNode())
}

func (node *AssociativeAddressTrieNode) LastNode() *AssociativeAddressTrieNode {
	return toAssociativeAddressTrieNode(node.toTrieNode().LastNode())
}

func (node *AssociativeAddressTrieNode) LastAddedNode() *AssociativeAddressTrieNode {
	return toAssociativeAddressTrieNode(node.toTrieNode().LastAddedNode())
}

// LowerAddedNode returns the added node, in this sub-trie with this node as root, whose address is the highest address strictly less than the given address.
func (node *AssociativeAddressTrieNode) LowerAddedNode(addr *Address) *AssociativeAddressTrieNode {
	return node.toBase().lowerAddedNode(addr).ToAssociative()
}

// FloorAddedNode returns the added node, in this sub-trie with this node as root, whose address is the highest address less than or equal to the given address.
func (node *AssociativeAddressTrieNode) FloorAddedNode(addr *Address) *AssociativeAddressTrieNode {
	return node.toBase().floorAddedNode(addr).ToAssociative()
}

// HigherAddedNode returns the added node, in this sub-trie with this node as root, whose address is the lowest address strictly greater than the given address.
func (node *AssociativeAddressTrieNode) HigherAddedNode(addr *Address) *AssociativeAddressTrieNode {
	return node.toBase().higherAddedNode(addr).ToAssociative()
}

// CeilingAddedNode returns the added node, in this sub-trie with this node as root, whose address is the lowest address greater than or equal to the given address.
func (node *AssociativeAddressTrieNode) CeilingAddedNode(addr *Address) *AssociativeAddressTrieNode {
	return node.toBase().ceilingAddedNode(addr).ToAssociative()
}

// Iterator returns an iterator that iterates through the elements of the sub-tree with this node as the root.
// The iteration is in sorted element order.
func (node *AssociativeAddressTrieNode) Iterator() AddressIterator {
	return node.toBase().iterator()
}

// DescendingIterator returns an iterator that iterates through the elements of the subtrie with this node as the root.
// The iteration is in reverse sorted element order.
func (node *AssociativeAddressTrieNode) DescendingIterator() AddressIterator {
	return node.toBase().descendingIterator()
}

// NodeIterator returns an iterator that iterates through the added nodes of the sub-tree with this node as the root, in forward or reverse tree order.
func (node *AssociativeAddressTrieNode) NodeIterator(forward bool) AssociativeAddressTrieNodeIteratorRem {
	return associativeAddressTrieNodeIteratorRem{node.toBase().nodeIterator(forward)}
}

// AllNodeIterator returns an iterator that iterates through all the nodes of the sub-tree with this node as the root, in forward or reverse tree order.
func (node *AssociativeAddressTrieNode) AllNodeIterator(forward bool) AssociativeAddressTrieNodeIteratorRem {
	return associativeAddressTrieNodeIteratorRem{node.toBase().allNodeIterator(forward)}
}

// BlockSizeNodeIterator returns an iterator that iterates the added nodes, ordered by keys from largest prefix blocks to smallest and then to individual addresses,
// in the sub-trie with this node as the root.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order is taken.
func (node *AssociativeAddressTrieNode) BlockSizeNodeIterator(lowerSubNodeFirst bool) AssociativeAddressTrieNodeIteratorRem {
	return associativeAddressTrieNodeIteratorRem{node.toBase().blockSizeNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeAllNodeIterator returns an iterator that iterates all the nodes, ordered by keys from largest prefix blocks to smallest and then to individual addresses,
// in the sub-trie with this node as the root.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (node *AssociativeAddressTrieNode) BlockSizeAllNodeIterator(lowerSubNodeFirst bool) AssociativeAddressTrieNodeIteratorRem {
	return associativeAddressTrieNodeIteratorRem{node.toBase().blockSizeAllNodeIterator(lowerSubNodeFirst)}
}

// BlockSizeCachingAllNodeIterator returns an iterator that iterates all nodes, ordered by keys from largest prefix blocks to smallest and then to individual addresses,
// in the sub-trie with this node as the root.
func (node *AssociativeAddressTrieNode) BlockSizeCachingAllNodeIterator() CachingAssociativeAddressTrieNodeIterator {
	return cachingAssociativeAddressTrieNodeIterator{node.toBase().blockSizeCachingAllNodeIterator()}
}

func (node *AssociativeAddressTrieNode) ContainingFirstIterator(forwardSubNodeOrder bool) CachingAssociativeAddressTrieNodeIterator {
	return cachingAssociativeAddressTrieNodeIterator{node.toBase().containingFirstIterator(forwardSubNodeOrder)}
}

func (node *AssociativeAddressTrieNode) ContainingFirstAllNodeIterator(forwardSubNodeOrder bool) CachingAssociativeAddressTrieNodeIterator {
	return cachingAssociativeAddressTrieNodeIterator{node.toBase().containingFirstAllNodeIterator(forwardSubNodeOrder)}
}

func (node *AssociativeAddressTrieNode) ContainedFirstIterator(forwardSubNodeOrder bool) AssociativeAddressTrieNodeIteratorRem {
	return associativeAddressTrieNodeIteratorRem{node.toBase().containedFirstIterator(forwardSubNodeOrder)}
}

func (node *AssociativeAddressTrieNode) ContainedFirstAllNodeIterator(forwardSubNodeOrder bool) AssociativeAddressTrieNodeIterator {
	return associativeAddressTrieNodeIterator{node.toBase().containedFirstAllNodeIterator(forwardSubNodeOrder)}
}

// Clone clones the node.
// Keys remain the same, but the parent node and the lower and upper sub-nodes are all set to nil.
func (node *AssociativeAddressTrieNode) Clone() *AssociativeAddressTrieNode {
	//return node.clone().ToAssociative()
	return toAssociativeAddressTrieNode(node.toTrieNode().Clone())
}

// CloneTree clones the sub-tree starting with this node as root.
// The nodes are cloned, but their keys and values are not cloned.
func (node *AssociativeAddressTrieNode) CloneTree() *AssociativeAddressTrieNode {
	//return node.cloneTree().ToAssociative()
	return toAssociativeAddressTrieNode(node.toTrieNode().CloneTree())
}

// AsNewTrie creates a new sub-trie, copying the nodes starting with this node as root.
// The nodes are copies of the nodes in this sub-trie, but their keys and values are not copies.
func (node *AssociativeAddressTrieNode) AsNewTrie() *AssociativeAddressTrie {
	return toAddressTrie(node.toTrieNode().AsNewTrie()).ToAssociative()
}

// Compare returns -1, 0 or 1 if this node is less than, equal, or greater than the other, according to the key and the trie order.
func (node *AssociativeAddressTrieNode) Compare(other *AssociativeAddressTrieNode) int {
	return node.toTrieNode().Compare(other.toTrieNode())
}

// Equal returns whether the key and and mapped value match those of the given node
func (node *AssociativeAddressTrieNode) Equal(other *AssociativeAddressTrieNode) bool {
	return node.toTrieNode().Equal(other.toTrieNode())
}

// TreeEqual returns whether the sub-tree represented by this node as the root node matches the given sub-tree
func (node *AssociativeAddressTrieNode) TreeEqual(other *AssociativeAddressTrieNode) bool {
	return node.toTrieNode().TreeEqual(other.toTrieNode())
}

// DeepEqual returns whether the key is equal to that of the given node and the value is deep equal to that of the given node
func (node *AssociativeAddressTrieNode) DeepEqual(other *AssociativeAddressTrieNode) bool {
	return node.toTrieNode().DeepEqual(other.toTrieNode())
}

// TreeDeepEqual returns whether the sub-tree represented by this node as the root node matches the given sub-tree, matching with Compare on the keys and reflect.DeepEqual on the values
func (node *AssociativeAddressTrieNode) TreeDeepEqual(other *AssociativeAddressTrieNode) bool {
	return node.toTrieNode().TreeDeepEqual(other.toTrieNode())
}

///////////////////////////////////////////////////////////////////////////

// Remove removes this node from the collection of added nodes, and also from the trie if possible.
// If it has two sub-nodes, it cannot be removed from the trie, in which case it is marked as not "added",
// nor is it counted in the trie size.
// Only added nodes can be removed from the trie.  If this node is not added, this method does nothing.
func (node *AssociativeAddressTrieNode) Remove() {
	node.toTrieNode().Remove()
}

// Contains returns whether the given address or prefix block subnet is in the sub-trie, as an added element, with this node as the root.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the prefix block or address address exists already in the trie, false otherwise.
//
// Use GetAddedNode  to get the node for the address rather than just checking for its existence.
func (node *AssociativeAddressTrieNode) Contains(addr *Address) bool {
	return node.toBase().contains(addr)
}

// RemoveNode removes the given single address or prefix block subnet from the trie with this node as the root.
//
// Removing an element will not remove contained elements (nodes for contained blocks and addresses).
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the prefix block or address was removed, false if not already in the trie.
//
// You can also remove by calling GetAddedNode to get the node and then calling Remove on the node.
//
// When an address is removed, the corresponding node may remain in the trie if it remains a subnet block for two sub-nodes.
// If the corresponding node can be removed from the trie, it will be.
func (node *AssociativeAddressTrieNode) RemoveNode(addr *Address) bool {
	return node.toBase().removeNode(addr)
}

// RemoveElementsContainedBy removes any single address or prefix block subnet from the trie, with this node as the root, that is contained in the given individual address or prefix block subnet.
//
// Goes further than Remove, not requiring a match to an inserted node, and also removing all the sub-nodes of any removed node or sub-node.
//
// For example, after inserting 1.2.3.0 and 1.2.3.1, passing 1.2.3.0/31 to RemoveElementsContainedBy will remove them both,
// while the Remove method will remove nothing.
// After inserting 1.2.3.0/31, then #remove(Address) will remove 1.2.3.0/31, but will leave 1.2.3.0 and 1.2.3.1 in the trie.
//
// It cannot partially delete a node, such as deleting a single address from a prefix block represented by a node.
// It can only delete the whole node if the whole address or block represented by that node is contained in the given address or block.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
//Returns the root node of the subtrie that was removed from the trie, or nil if nothing was removed.
func (node *AssociativeAddressTrieNode) RemoveElementsContainedBy(addr *Address) *AssociativeAddressTrieNode {
	return node.toBase().removeElementsContainedBy(addr).ToAssociative()
}

// ElementsContainedBy checks if a part of this trie, with this node as the root, is contained by the given prefix block subnet or individual address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the root node of the contained subtrie, or nil if no subtrie is contained.
// The node returned need not be an "added" node, see IsAdded for more details on added nodes.
// The returned subtrie is backed by this trie, so changes in this trie are reflected in those nodes and vice-versa.
func (node *AssociativeAddressTrieNode) ElementsContainedBy(addr *Address) *AssociativeAddressTrieNode {
	return node.toBase().elementsContainedBy(addr).ToAssociative()
}

// ElementsContaining finds the trie nodes in the trie, with this sub-node as the root,
// containing the given key and returns them as a linked list.
// Only added nodes are added to the linked list
//
// If the argument is not a single address nor prefix block, this method will panic.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (node *AssociativeAddressTrieNode) ElementsContaining(addr *Address) *AssociativeAddressTrieNode {
	return node.toBase().elementsContaining(addr).ToAssociative()
}

// LongestPrefixMatch returns the address pr subnet with the longest prefix of all the added subnets or address whose prefix matches the given address.
// This is equivalent to finding the containing subnet or address with the smallest subnet size.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns nil if no added subnet or address contains the given argument.
//
// Use ElementContains to check for the existence of a containing address.
// To get all the containing addresses (subnets with matching prefix), use ElementsContaining.
// To get the node corresponding to the result of this method, use LongestPrefixMatchNode.
func (node *AssociativeAddressTrieNode) LongestPrefixMatch(addr *Address) *Address {
	return node.toBase().longestPrefixMatch(addr)
}

// LongestPrefixMatchNode finds the containing subnet or address in the trie with the smallest subnet size,
// which is equivalent to finding the subnet or address with the longest matching prefix.
// Returns the node corresponding to that subnet.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns nil if no added subnet or address contains the given argument.
//
// Use ElementContains to check for the existence of a containing address.
// To get all the containing addresses, use ElementsContaining.
// Use LongestPrefixMatch to get only the address corresponding to the result of this method.
func (node *AssociativeAddressTrieNode) LongestPrefixMatchNode(addr *Address) *AssociativeAddressTrieNode {
	return node.toBase().longestPrefixMatchNode(addr).ToAssociative()
}

// ElementContains checks if a prefix block subnet or address in the trie, with this node as the root, contains the given subnet or address.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns true if the subnet or address is contained by a trie element, false otherwise.
//
// To get all the containing addresses, use ElementsContaining.
func (node *AssociativeAddressTrieNode) ElementContains(addr *Address) bool {
	return node.toBase().elementContains(addr)
}

// GetNode gets the node in the trie, with this subnode as the root, corresponding to the given address,
// or returns nil if not such element exists.
//
// It returns any node, whether added or not,
// including any prefix block node that was not added.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
func (node *AssociativeAddressTrieNode) GetNode(addr *Address) *AssociativeAddressTrieNode {
	return node.toBase().getNode(addr).ToAssociative()
}

// GetAddedNode gets trie nodes representing added elements.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Use Contains to check for the existence of a given address in the trie,
// as well as GetNode to search for all nodes including those not-added but also auto-generated nodes for subnet blocks.
func (node *AssociativeAddressTrieNode) GetAddedNode(addr *Address) *AssociativeAddressTrieNode {
	return node.toBase().getAddedNode(addr).ToAssociative()
}

// Get gets the specified value for the specified key in this mapped trie or subtrie.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// Returns the value for the given key.
// Returns nil if the contains no mapping for that key or if the mapped value is nil.
func (node *AssociativeAddressTrieNode) Get(addr *Address) NodeValue {
	return node.toBase().get(addr)
}

// NodeSize returns the number of nodes in the trie with this node as the root, which is more than the number of added addresses or blocks.
func (node *AssociativeAddressTrieNode) NodeSize() int {
	return node.toTrieNode().NodeSize()

}

// Size returns the number of elements in the tree.
// Only nodes for which IsAdded returns true are counted.
// When zero is returned, IsEmpty returns true.
func (node *AssociativeAddressTrieNode) Size() int {
	return node.toTrieNode().Size()
}

// IsEmpty returns whether the size is 0
func (node *AssociativeAddressTrieNode) IsEmpty() bool {
	return node.Size() == 0
}

// TreeString returns a visual representation of the sub-tree with this node as root, with one node per line.
//
// withNonAddedKeys: whether to show nodes that are not added nodes
// withSizes: whether to include the counts of added nodes in each sub-tree
func (node *AssociativeAddressTrieNode) TreeString(withNonAddedKeys, withSizes bool) string {
	return node.toTrieNode().TreeString(withNonAddedKeys, withSizes)
}

// String returns a visual representation of this node including the key, with an open circle indicating this node is not an added node,
// a closed circle indicating this node is an added node.
func (node *AssociativeAddressTrieNode) String() string {
	return node.toTrieNode().String()
}

// For some reason Format must be here and not in addressTrieNode for nil node.
// It panics in fmt code either way, but if in here then it is handled by a recover() call in fmt properly in the debugger.

// Format implements the fmt.Formatter interface
func (node AssociativeAddressTrieNode) Format(state fmt.State, verb rune) {
	node.ToBase().Format(state, verb)
}
