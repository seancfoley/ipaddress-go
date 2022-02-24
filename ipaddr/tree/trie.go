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

package tree

import (
	"fmt"
	"strings"
	"unsafe"
)

// BinTrie is a binary trie.
//
// To use BinTrie, your keys implement TrieKey.
//
// All keys are either fixed, in which the key value does not change,
// or comprising of a prefix in which an initial sequence of bits does not change, and the the remaining bits represent all bit values.
// The length of the initial fixed sequence of bits is the prefix length.
// The total bit length is the same for all keys.
//
// A key with a prefix is also known as a prefix block, and represents all bit sequences with the same prefix.
//
// The zero value for BinTrie is a binary trie ready for use.
type BinTrie struct {
	binTree
}

func (trie *BinTrie) toBinTree() *binTree {
	return (*binTree)(unsafe.Pointer(trie))
}

// GetRoot returns the root of this trie (in the case of bounded tries, this would be the bounded root)
func (trie *BinTrie) GetRoot() (root *BinTrieNode) {
	if trie != nil {
		root = trie.root.toTrieNode()
	}
	return
}

// Returns the root of this trie (in the case of bounded tries, the absolute root ignores the bounds)
func (trie *BinTrie) absoluteRoot() (root *BinTrieNode) {
	if trie != nil {
		root = trie.root.toTrieNode()
	}
	return
}

// Size returns the number of elements in the tree.
// Only nodes for which IsAdded() returns true are counted.
// When zero is returned, IsEmpty() returns true.
func (trie *BinTrie) Size() int {
	return trie.toBinTree().Size()
}

// NodeSize returns the number of nodes in the tree, which is always more than the number of elements.
func (trie *BinTrie) NodeSize() int {
	return trie.toBinTree().NodeSize()
}

func (trie *BinTrie) ensureRoot(key TrieKey) *BinTrieNode {
	root := trie.root
	if root == nil {
		root = trie.setRoot(key.ToPrefixBlockLen(0))
	}
	return root.toTrieNode()
}

func (trie *BinTrie) setRoot(key TrieKey) *binTreeNode {
	root := &binTreeNode{
		item:     key,
		cTracker: &changeTracker{},
	}
	root.setAddr()
	trie.root = root
	return root
}

// Iterator returns an iterator that iterates through the elements of the sub-tree with this node as the root.
// The iteration is in sorted element order.
func (trie *BinTrie) Iterator() TrieKeyIterator {
	return trie.GetRoot().Iterator()
}

// DescendingIterator returns an iterator that iterates through the elements of the subtrie with this node as the root.
// The iteration is in reverse sorted element order.
func (trie *BinTrie) DescendingIterator() TrieKeyIterator {
	return trie.GetRoot().DescendingIterator()
}

// ConstructAddedNodesTree provides an associative trie in which the root and each added node of this trie are mapped to a list of their respective direct added sub-nodes.
// This trie provides an alternative non-binary tree structure of the added nodes.
// It is used by ToAddedNodesTreeString to produce a string showing the alternative structure.
// If there are no non-added nodes in this trie,
// then the alternative tree structure provided by this method is the same as the original trie.
// The trie values of this trie are of type []*BinTrieNode
func (trie *BinTrie) ConstructAddedNodesTree() BinTrie {
	newRoot := trie.root.clone()
	newRoot.ClearValue()
	newRoot.cTracker = &changeTracker{}
	emptyTrie := BinTrie{binTree{newRoot}}
	//AssociativeAddressTrie<E, ? extends List<? extends AssociativeTrieNode<E, ?>>> emptyTrie;
	// // returns AssociativeAddressTrie<E, ? extends List<? extends AssociativeTrieNode<E, ?>>> of keys mapped to subnodes
	emptyTrie.Clear()
	emptyTrie.addTrie(trie.absoluteRoot(), true)

	// this trie goes into the new trie, then as we iterate,
	// we find our parent and add ourselves to that parent's list of subnodes

	cachingIterator := emptyTrie.ContainingFirstAllNodeIterator(true)
	for next := cachingIterator.Next(); next != nil; next = cachingIterator.Next() {
		cachingIterator.CacheWithLowerSubNode(next)
		cachingIterator.CacheWithUpperSubNode(next)

		// the cached object is our parent
		if next.IsAdded() {
			parent := cachingIterator.GetCached().(*BinTrieNode)
			if parent != nil {
				// find added parent, or the root if no added parent
				// this part would be tricky if we accounted for the bounds,
				// maybe we'd have to filter on the bounds, and also look for the sub-root
				for !parent.IsAdded() {
					parentParent := parent.GetParent()
					if parentParent == nil {
						break
					}
					parent = parentParent
				}
				// store ourselves with that added parent or root
				val := parent.GetValue()
				var list []*BinTrieNode
				if val == nil {
					list = make([]*BinTrieNode, 0, 3)
				} else {
					list = val.([]*BinTrieNode)
				}
				list = append(list, next)
				parent.SetValue(list)
			} // else root
		}
	}
	return emptyTrie
}

// String returns a visual representation of the tree with one node per line.
func (trie *BinTrie) String() string {
	if trie == nil {
		return nilString()
	}
	return trie.binTree.String()
}

// TreeString returns a visual representation of the tree with one node per line, with or without the non-added keys.
func (trie *BinTrie) TreeString(withNonAddedKeys bool) string {
	if trie == nil {
		return "\n" + nilString()
	}
	return trie.binTree.TreeString(withNonAddedKeys)
}

// AddedNodesTreeString provides a flattened version of the trie showing only the contained added nodes and their containment structure, which is non-binary.
// The root node is included, which may or may not be added.
func (trie *BinTrie) AddedNodesTreeString() string {
	if trie == nil {
		return "\n" + nilString()
	}
	type indentsNode struct {
		inds indents
		node *BinTrieNode
	}
	//var stack list.List
	addedTree := trie.ConstructAddedNodesTree()
	var stack []indentsNode
	root := addedTree.absoluteRoot()
	builder := strings.Builder{}
	builder.WriteByte('\n')
	nodeIndent, subNodeIndent := "", ""
	nextNode := root
	for {
		builder.WriteString(nodeIndent)
		if nextNode.IsAdded() {
			builder.WriteString(addedNodeCircle)
		} else {
			builder.WriteString(nonAddedNodeCircle)
		}
		builder.WriteByte(' ')
		builder.WriteString(fmt.Sprint(nextNode.GetKey()))
		builder.WriteByte('\n')
		nextVal := nextNode.GetValue()
		var nextNodes []*BinTrieNode
		if nextVal != nil {
			nextNodes = nextVal.([]*BinTrieNode)
		}
		if nextNodes != nil && len(nextNodes) > 0 {
			i := len(nextNodes) - 1
			lastIndents := indents{
				nodeIndent: subNodeIndent + rightElbow,
				subNodeInd: subNodeIndent + belowElbows,
			}
			nNode := nextNodes[i]
			if stack == nil {
				stack = make([]indentsNode, 0, addedTree.Size())
			}
			//next := nNode
			stack = append(stack, indentsNode{lastIndents, nNode})
			if len(nextNodes) > 1 {
				firstIndents := indents{
					nodeIndent: subNodeIndent + leftElbow,
					subNodeInd: subNodeIndent + inBetweenElbows,
				}
				for i--; i >= 0; i-- {
					nNode = nextNodes[i]
					//next =  nNode;
					stack = append(stack, indentsNode{firstIndents, nNode})
				}
			}
		}
		if stack == nil {
			break
		}
		stackLen := len(stack)
		if stackLen == 0 {
			break
		}
		newLen := stackLen - 1
		nextItem := stack[newLen]
		stack = stack[:newLen]
		nextNode = nextItem.node
		nextIndents := nextItem.inds
		nodeIndent = nextIndents.nodeIndent
		subNodeIndent = nextIndents.subNodeInd
	}
	return builder.String()
}

// Add adds the given key to the trie, returning true if not there already.
func (trie *BinTrie) Add(key TrieKey) bool {
	root := trie.ensureRoot(key)
	result := &opResult{
		key: key,
		op:  insert,
	}
	root.matchBits(result)
	return !result.exists
}

// AddNode is similar to Add but returns the new or existing node.
func (trie *BinTrie) AddNode(key TrieKey) *BinTrieNode {
	root := trie.ensureRoot(key)
	result := &opResult{
		key: key,
		op:  insert,
	}
	root.matchBits(result)
	node := result.existingNode
	if node == nil {
		node = result.inserted
	}
	return node
}

func (trie *BinTrie) addNode(result *opResult, fromNode, nodeToAdd *BinTrieNode, withValues bool) *BinTrieNode {
	if withValues {
		result.newValue = nodeToAdd.GetValue()
	}
	fromNode.matchBitsFromIndex(fromNode.GetKey().GetPrefixLen().Len(), result)
	node := result.existingNode
	if node == nil {
		return result.inserted
	}
	return node
}

// Note: this method not called from sets or maps, so bounds does not apply
func (trie *BinTrie) addTrie(addedTree *BinTrieNode, withValues bool) *BinTrieNode {
	iterator := addedTree.ContainingFirstAllNodeIterator(true)
	toAdd := iterator.Next()
	result := &opResult{
		key: toAdd.GetKey(),
		op:  insert,
	}
	var firstNode *BinTrieNode
	root := trie.absoluteRoot()
	firstAdded := toAdd.IsAdded()
	if firstAdded {
		firstNode = trie.addNode(result, root, toAdd, withValues)
	} else {
		firstNode = root
	}
	lastAddedNode := firstNode
	for iterator.HasNext() {
		iterator.CacheWithLowerSubNode(lastAddedNode)
		iterator.CacheWithUpperSubNode(lastAddedNode)
		toAdd = iterator.Next()
		cachedNode := iterator.GetCached().(*BinTrieNode)
		if toAdd.IsAdded() {
			addrNext := toAdd.GetKey()
			result.key = addrNext
			result.existingNode = nil
			result.inserted = nil
			lastAddedNode = trie.addNode(result, cachedNode, toAdd, withValues)
		} else {
			lastAddedNode = cachedNode
		}
	}
	if !firstAdded {
		firstNode = trie.GetNode(addedTree.GetKey())
	}
	return firstNode
}

func (trie *BinTrie) AddTrie(trieNode *BinTrieNode) *BinTrieNode {
	if trieNode == nil {
		return nil
	}
	trie.ensureRoot(trieNode.GetKey())
	return trie.addTrie(trieNode, false)
}

func (trie *BinTrie) Contains(key TrieKey) bool {
	return trie.absoluteRoot().Contains(key)
}

func (trie *BinTrie) Remove(key TrieKey) bool {
	return trie.absoluteRoot().RemoveNode(key)
}

func (trie *BinTrie) RemoveElementsContainedBy(key TrieKey) *BinTrieNode {
	return trie.absoluteRoot().RemoveElementsContainedBy(key)
}

func (trie *BinTrie) ElementsContainedBy(key TrieKey) *BinTrieNode {
	return trie.absoluteRoot().ElementsContainedBy(key)
}

func (trie *BinTrie) ElementsContaining(key TrieKey) *BinTrieNode {
	return trie.absoluteRoot().ElementsContaining(key)
}

// LongestPrefixMatch finds the key with the longest matching prefix.
func (trie *BinTrie) LongestPrefixMatch(key TrieKey) TrieKey {
	return trie.absoluteRoot().LongestPrefixMatch(key)
}

// LongestPrefixMatchNode finds the node with the longest matching prefix.
func (trie *BinTrie) LongestPrefixMatchNode(key TrieKey) *BinTrieNode {
	return trie.absoluteRoot().LongestPrefixMatchNode(key)
}

func (trie *BinTrie) ElementContains(key TrieKey) bool {
	return trie.absoluteRoot().ElementContains(key)
}

// GetNode gets the node in the sub-trie corresponding to the given address,
// or returns nil if not such element exists.
//
// It returns any node, whether added or not,
// including any prefix block node that was not added.
func (trie *BinTrie) GetNode(key TrieKey) *BinTrieNode {
	return trie.absoluteRoot().GetNode(key)
}

// GetAddedNode gets trie nodes representing added elements.
//
// Use Contains to check for the existence of a given address in the trie,
// as well as GetNode to search for all nodes including those not-added but also auto-generated nodes for subnet blocks.
func (trie *BinTrie) GetAddedNode(key TrieKey) *BinTrieNode {
	return trie.absoluteRoot().GetAddedNode(key)
}

// Put associates the specified value with the specified key in this map.
//
// If the argument is not a single address nor prefix block, this method will panic.
// The Partition type can be used to convert the argument to single addresses and prefix blocks before calling this method.
//
// If this map previously contained a mapping for a key,
// the old value is replaced by the specified value, and false is returned along with the old value.
// If this map did not previously contain a mapping for the key, true is returned along with nil.
// The boolean return value allows you to distinguish whether the key was previously mapped to nil or not mapped at all.
func (trie *BinTrie) Put(key TrieKey, value V) (bool, V) {
	root := trie.ensureRoot(key)
	result := &opResult{
		key:      key,
		op:       insert,
		newValue: value,
	}
	root.matchBits(result)
	return !result.exists, result.existingValue
}

func (trie *BinTrie) PutTrie(trieNode *BinTrieNode) *BinTrieNode {
	if trieNode == nil {
		return nil
	}
	trie.ensureRoot(trieNode.GetKey())
	return trie.addTrie(trieNode, true)
}

func (trie *BinTrie) PutNode(key TrieKey, value V) *BinTrieNode {
	root := trie.ensureRoot(key)
	result := &opResult{
		key:      key,
		op:       insert,
		newValue: value,
	}
	root.matchBits(result)
	resultNode := result.existingNode
	if resultNode == nil {
		resultNode = result.inserted
	}
	return resultNode
}

func (trie *BinTrie) Remap(key TrieKey, remapper func(V) V) *BinTrieNode {
	return trie.remapImpl(key,
		func(existingVal V) (V, remapAction) {
			result := remapper(existingVal)
			if result == nil {
				return nil, removeNode
			}
			return result, remapValue
		})
}

func (trie *BinTrie) RemapIfAbsent(key TrieKey, supplier func() V, insertNil bool) *BinTrieNode {
	return trie.remapImpl(key,
		func(existingVal V) (V, remapAction) {
			if existingVal == nil {
				result := supplier()
				if result != nil || insertNil {
					return result, remapValue
				}
			}
			return nil, doNothing
		})
}

func (trie *BinTrie) remapImpl(key TrieKey, remapper func(V) (V, remapAction)) *BinTrieNode {
	result := &opResult{
		key:      key,
		op:       remap,
		remapper: remapper,
	}
	root := trie.absoluteRoot()
	if root != nil {
		root.matchBits(result)
	}
	resultNode := result.existingNode
	if resultNode == nil {
		resultNode = result.inserted
	}
	return resultNode
}

func (trie *BinTrie) Get(key TrieKey) V {
	return trie.absoluteRoot().Get(key)
}

// NodeIterator returns an iterator that iterates through the added nodes of the trie in forward or reverse tree order.
func (trie *BinTrie) NodeIterator(forward bool) TrieNodeIteratorRem {
	return trie.absoluteRoot().NodeIterator(forward)
}

// AllNodeIterator returns an iterator that iterates through all the nodes of the trie in forward or reverse tree order.
func (trie *BinTrie) AllNodeIterator(forward bool) TrieNodeIteratorRem {
	return trie.absoluteRoot().AllNodeIterator(forward)
}

// BlockSizeNodeIterator returns an iterator that iterates the added nodes in the trie, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *BinTrie) BlockSizeNodeIterator(lowerSubNodeFirst bool) TrieNodeIteratorRem {
	return trie.absoluteRoot().BlockSizeNodeIterator(lowerSubNodeFirst)
}

// BlockSizeAllNodeIterator returns an iterator that iterates all nodes in the trie, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
//
// If lowerSubNodeFirst is true, for blocks of equal size the lower is first, otherwise the reverse order
func (trie *BinTrie) BlockSizeAllNodeIterator(lowerSubNodeFirst bool) TrieNodeIteratorRem {
	return trie.absoluteRoot().BlockSizeAllNodeIterator(lowerSubNodeFirst)
}

// BlockSizeCachingAllNodeIterator returns an iterator that iterates all nodes, ordered by keys from largest prefix blocks to smallest, and then to individual addresses.
func (trie *BinTrie) BlockSizeCachingAllNodeIterator() CachingTrieNodeIterator {
	return trie.absoluteRoot().BlockSizeCachingAllNodeIterator()
}

func (trie *BinTrie) ContainingFirstIterator(forwardSubNodeOrder bool) CachingTrieNodeIterator {
	return trie.absoluteRoot().ContainingFirstIterator(forwardSubNodeOrder)
}

func (trie *BinTrie) ContainingFirstAllNodeIterator(forwardSubNodeOrder bool) CachingTrieNodeIterator {
	return trie.absoluteRoot().ContainingFirstAllNodeIterator(forwardSubNodeOrder)
}

func (trie *BinTrie) ContainedFirstIterator(forwardSubNodeOrder bool) TrieNodeIteratorRem {
	return trie.absoluteRoot().ContainedFirstIterator(forwardSubNodeOrder)
}

func (trie *BinTrie) ContainedFirstAllNodeIterator(forwardSubNodeOrder bool) TrieNodeIterator {
	return trie.absoluteRoot().ContainedFirstAllNodeIterator(forwardSubNodeOrder)
}

func (trie *BinTrie) FirstNode() *BinTrieNode {
	return trie.absoluteRoot().FirstNode()
}

func (trie *BinTrie) FirstAddedNode() *BinTrieNode {
	return trie.absoluteRoot().FirstAddedNode()
}

func (trie *BinTrie) LastNode() *BinTrieNode {
	return trie.absoluteRoot().LastNode()
}

func (trie *BinTrie) LastAddedNode() *BinTrieNode {
	return trie.absoluteRoot().LastAddedNode()
}

func (trie *BinTrie) LowerAddedNode(key TrieKey) *BinTrieNode {
	return trie.absoluteRoot().LowerAddedNode(key)
}

func (trie *BinTrie) FloorAddedNode(key TrieKey) *BinTrieNode {
	return trie.absoluteRoot().FloorAddedNode(key)
}

func (trie *BinTrie) HigherAddedNode(key TrieKey) *BinTrieNode {
	return trie.absoluteRoot().HigherAddedNode(key)
}

func (trie *BinTrie) CeilingAddedNode(key TrieKey) *BinTrieNode {
	return trie.absoluteRoot().CeilingAddedNode(key)
}

func (trie *BinTrie) Clone() *BinTrie {
	if trie == nil {
		return nil
	}
	return &BinTrie{binTree{root: trie.absoluteRoot().CloneTree().toBinTreeNode()}}
}

// DeepEqual returns whether the given argument is a trie with a set of nodes with the same keys as in this trie according to the Compare method,
// and the same values according to the reflect.DeepEqual method
func (trie *BinTrie) DeepEqual(other *BinTrie) bool {
	return trie.absoluteRoot().TreeDeepEqual(other.absoluteRoot())
}

// Equal returns whether the given argument is a trie with a set of nodes with the same keys as in this trie according to the Compare method
func (trie *BinTrie) Equal(other *BinTrie) bool {
	return trie.absoluteRoot().TreeEqual(other.absoluteRoot())
}

// For some reason Format must be here and not in addressTrieNode for nil node.
// It panics in fmt code either way, but if in here then it is handled by a recover() call in fmt properly.
// Seems to be a problem only in the debugger.

// Format implements the fmt.Formatter interface
func (trie BinTrie) Format(state fmt.State, verb rune) {
	trie.format(state, verb)
}

// NewBinTrie creates a new trie with root key.ToPrefixBlockLen(0).
// If the key argument is not Equal to its zero-length prefix block, then the key will be added as well.
func NewBinTrie(key TrieKey) BinTrie {
	trie := BinTrie{binTree{}}
	root := key.ToPrefixBlockLen(0)
	trie.setRoot(root)
	if key.Compare(root) != 0 {
		trie.Add(key)
	}
	return trie
}

func TreesString(withNonAddedKeys bool, tries ...*BinTrie) string {
	binTrees := make([]*binTree, 0, len(tries))
	for _, trie := range tries {
		binTrees = append(binTrees, tobinTree(trie))
	}
	return treesString(withNonAddedKeys, binTrees...)
}

func tobinTree(trie *BinTrie) *binTree {
	return (*binTree)(unsafe.Pointer(trie))
}
