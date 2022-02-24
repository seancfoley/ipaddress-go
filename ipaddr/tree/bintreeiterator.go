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
	"container/heap"
)

type HasNext interface {
	HasNext() bool
}

type keyIterator interface {
	HasNext

	Next() e

	// Remove removes the last iterated element from the underlying trie, and returns that element.
	// If there is no such element, it returns nil.
	Remove() e
}

type nodeIteratorRem interface {
	nodeIterator

	// Remove removes the last iterated element from the underlying trie, and returns that element.
	// If there is no such element, it returns nil.
	Remove() *binTreeNode
}

type nodeIterator interface {
	HasNext

	Next() *binTreeNode
}

type binTreeKeyIterator struct {
	nodeIteratorRem
}

func (iter binTreeKeyIterator) Next() e {
	node := iter.nodeIteratorRem.Next()
	if node != nil {
		return node.getKey()
	}
	return nil
}

func (iter binTreeKeyIterator) Remove() e {
	node := iter.nodeIteratorRem.Remove()
	if node != nil {
		return node.getKey()
	}
	return nil
}

func newNodeIterator(forward, addedOnly bool, start, end *binTreeNode, ctracker *changeTracker) nodeIteratorRem {
	var nextOperator func(current *binTreeNode, end *binTreeNode) *binTreeNode
	if forward {
		nextOperator = (*binTreeNode).nextNodeBounded
	} else {
		nextOperator = (*binTreeNode).previousNodeBounded
	}
	if addedOnly {
		wrappedOp := nextOperator
		nextOperator = func(currentNode *binTreeNode, endNode *binTreeNode) *binTreeNode {
			return currentNode.nextAdded(endNode, wrappedOp)
		}
	}
	res := binTreeNodeIterator{end: end}
	res.setChangeTracker(ctracker)
	res.operator = nextOperator
	res.next = res.getStart(start, end, nil, addedOnly)
	return &res
}

type binTreeNodeIterator struct {
	// takes current node and end as args
	operator      func(currentNode *binTreeNode, endNode *binTreeNode) (nextNode *binTreeNode)
	end           *binTreeNode // a non-nil node that denotes the end, possibly parent of the starting node
	cTracker      *changeTracker
	currentChange change

	current, next *binTreeNode
}

func (iter *binTreeNodeIterator) getStart(
	start,
	end *binTreeNode,
	bounds *bounds,
	addedOnly bool) *binTreeNode {
	if start == end || start == nil {
		return nil
	}
	if !addedOnly || start.IsAdded() {
		if bounds == nil || bounds.isInBounds(start.getKey()) {
			return start
		}
	}
	return iter.toNext(start)
}

func (iter *binTreeNodeIterator) setChangeTracker(ctracker *changeTracker) {
	if ctracker != nil {
		iter.cTracker, iter.currentChange = ctracker, ctracker.getCurrent()
	}
}

func (iter *binTreeNodeIterator) HasNext() bool {
	return iter.next != nil
}

func (iter *binTreeNodeIterator) Next() *binTreeNode {
	if !iter.HasNext() {
		return nil
	}
	cTracker := iter.cTracker
	if cTracker != nil && cTracker.changedSince(iter.currentChange) {
		panic("the tree has been modified since the iterator was created")
	}
	iter.current = iter.next
	iter.next = iter.toNext(iter.next)
	return iter.current
}

func (iter *binTreeNodeIterator) toNext(current *binTreeNode) *binTreeNode {
	return iter.operator(current, iter.end)
}

func (iter *binTreeNodeIterator) Remove() *binTreeNode {
	if iter.current == nil {
		return nil
	}
	cTracker := iter.cTracker
	if cTracker != nil && cTracker.changedSince(iter.currentChange) {
		panic("the tree has been modified since the iterator was created")
	}
	result := iter.current
	result.Remove()
	iter.current = nil
	if cTracker != nil {
		iter.currentChange = cTracker.getCurrent()
	}
	return result
}

var _ nodeIteratorRem = &binTreeNodeIterator{}

type CachingIterator interface {
	// GetCached returns an object previously cached with the current iterated node.
	// After Next has returned a node,
	// if an object was cached by a call to CacheWithLowerSubNode or CacheWithUpperSubNode
	// was called when that node's parent was previously returned by Next,
	// then this returns that cached object.
	GetCached() C

	// CacheWithLowerSubNode caches an object with the lower sub-node of the current iterated node.
	// After Next has returned a node,
	// calling this method caches the provided object with the lower sub-node so that it can
	// be retrieved with GetCached when the lower sub-node is visited later.
	//
	// Returns false if it could not be cached, either because the node has since been removed with a call to Remove,
	// because Next has not been called yet, or because there is no lower sub node for the node previously returned by  Next.
	//
	// The caching and retrieval is done in constant time.
	CacheWithLowerSubNode(C) bool

	// CacheWithUpperSubNode caches an object with the upper sub-node of the current iterated node.
	// After Next has returned a node,
	// calling this method caches the provided object with the upper sub-node so that it can
	// be retrieved with GetCached when the upper sub-node is visited later.
	//
	// Returns false if it could not be cached, either because the node has since been removed with a call to Remove,
	// because Next has not been called yet, or because there is no upper sub node for the node previously returned by Next.
	//
	// The caching and retrieval is done in constant time.
	CacheWithUpperSubNode(C) bool
}

type cachingNodeIterator interface {
	nodeIteratorRem

	CachingIterator
}

// see https://pkg.go.dev/container/heap
type nodePriorityQueue struct {
	queue      []interface{}
	comparator func(one, two interface{}) int // -1, 0 or 1 if one is <, == or > two
	//queue      []*binTreeNode
	//comparator func(one, two *binTreeNode) int // -1, 0 or 1 if one is <, == or > two
}

func (prioQueue nodePriorityQueue) Len() int {
	return len(prioQueue.queue)
}

func (prioQueue nodePriorityQueue) Less(i, j int) bool {
	queue := prioQueue.queue
	return prioQueue.comparator(queue[i], queue[j]) < 0
}

func (prioQueue nodePriorityQueue) Swap(i, j int) {
	queue := prioQueue.queue
	queue[i], queue[j] = queue[j], queue[i]
}

func (prioQueue *nodePriorityQueue) Push(x interface{}) {
	prioQueue.queue = append(prioQueue.queue, x)
}

func (prioQueue *nodePriorityQueue) Pop() interface{} {
	current := prioQueue.queue
	queueLen := len(current)
	topNode := current[queueLen-1]
	current[queueLen-1] = nil
	prioQueue.queue = current[:queueLen-1]
	return topNode
}

func newPriorityNodeIterator(
	treeSize int,
	addedOnly bool,
	start *binTreeNode,
	comparator func(e, e) int,
) binTreeNodeIterator {
	return newPriorityNodeIteratorBounded(
		nil,
		treeSize,
		addedOnly,
		start,
		comparator)
}

func newPriorityNodeIteratorBounded(
	bnds *bounds,
	treeSize int,
	addedOnly bool,
	start *binTreeNode,
	comparator func(e, e) int) binTreeNodeIterator {

	comp := func(one, two interface{}) int {
		node1, node2 := one.(*binTreeNode), two.(*binTreeNode)
		addr1, addr2 := node1.getKey(), node2.getKey()
		return comparator(addr1, addr2)
	}
	queue := &nodePriorityQueue{comparator: comp}
	if treeSize > 0 {
		//queue.queue = make([]*binTreeNode, 0, (treeSize+2)>>1)
		queue.queue = make([]interface{}, 0, (treeSize+2)>>1)
	}
	op := func(currentNode *binTreeNode, endNode *binTreeNode) *binTreeNode {
		lower := currentNode.getLowerSubNode()
		if lower != nil {
			heap.Push(queue, lower)
		}
		upper := currentNode.getUpperSubNode()
		if upper != nil {
			heap.Push(queue, upper)
		}
		var node *binTreeNode
		if queue.Len() > 0 {
			node = heap.Pop(queue).(*binTreeNode)
		}
		if node == endNode {
			return nil
		}
		return node

	}
	if addedOnly {
		wrappedOp := op
		op = func(currentNode *binTreeNode, endNode *binTreeNode) *binTreeNode {
			return currentNode.nextAdded(endNode, wrappedOp)
		}
	}
	if bnds != nil {
		wrappedOp := op
		op = func(currentNode *binTreeNode, endNode *binTreeNode) *binTreeNode {
			return currentNode.nextInBounds(endNode, wrappedOp, bnds)
		}
	}
	res := binTreeNodeIterator{operator: op}
	start = res.getStart(start, nil, bnds, addedOnly)
	if start != nil {
		res.next = start
		res.setChangeTracker(start.cTracker)
	}
	return res
}

func newCachingPriorityNodeIterator(
	start *binTreeNode,
	comparator func(e, e) int,
) cachingPriorityNodeIterator {
	return newCachingPriorityNodeIteratorSized(
		0,
		start,
		comparator)
}

func newCachingPriorityNodeIteratorSized(
	treeSize int,
	start *binTreeNode,
	comparator func(e, e) int) cachingPriorityNodeIterator {

	comp := func(one, two interface{}) int {
		cached1, cached2 := one.(*cached), two.(*cached)
		node1, node2 := cached1.node, cached2.node
		addr1, addr2 := node1.getKey(), node2.getKey()
		return comparator(addr1, addr2)
	}
	queue := &nodePriorityQueue{comparator: comp}
	if treeSize > 0 {
		queue.queue = make([]interface{}, 0, (treeSize+2)>>1)
	}
	res := cachingPriorityNodeIterator{cached: &cachedObjs{}}
	res.operator = res.getNextOperation(queue)
	start = res.getStart(start, nil, nil, false)
	if start != nil {
		res.next = start
		res.setChangeTracker(start.cTracker)
	}
	return res
}

type cachedObjs struct {
	cacheItem                    C
	nextCachedItem               *cached
	lowerCacheObj, upperCacheObj *cached
}

type cachingPriorityNodeIterator struct {
	binTreeNodeIterator
	cached *cachedObjs
}

func (iter *cachingPriorityNodeIterator) getNextOperation(queue *nodePriorityQueue) func(currentNode *binTreeNode, endNode *binTreeNode) *binTreeNode {
	return func(currentNode *binTreeNode, endNode *binTreeNode) *binTreeNode {
		lower := currentNode.getLowerSubNode()
		cacheObjs := iter.cached
		if lower != nil {
			cachd := &cached{
				node: lower,
			}
			cacheObjs.lowerCacheObj = cachd
			heap.Push(queue, cachd)
		} else {
			cacheObjs.lowerCacheObj = nil
		}
		upper := currentNode.getUpperSubNode()
		if upper != nil {
			cachd := &cached{
				node: upper,
			}
			cacheObjs.upperCacheObj = cachd
			heap.Push(queue, cachd)
		} else {
			cacheObjs.upperCacheObj = nil
		}
		if cacheObjs.nextCachedItem != nil {
			cacheObjs.cacheItem = cacheObjs.nextCachedItem.cached
		}
		var item interface{}
		if queue.Len() > 0 {
			item = heap.Pop(queue)
		}
		if item != nil {
			cachd := item.(*cached)
			node := cachd.node
			if node != endNode {
				cacheObjs.nextCachedItem = cachd
				return node
			}
		}
		cacheObjs.nextCachedItem = nil
		return nil
	}
}

func (iter *cachingPriorityNodeIterator) GetCached() C {
	return iter.cached.cacheItem
}

func (iter *cachingPriorityNodeIterator) CacheWithLowerSubNode(object C) bool {
	cached := iter.cached
	if cached.lowerCacheObj != nil {
		cached.lowerCacheObj.cached = object
		return true
	}
	return false
}

func (iter *cachingPriorityNodeIterator) CacheWithUpperSubNode(object C) bool {
	cached := iter.cached
	if cached.upperCacheObj != nil {
		cached.upperCacheObj.cached = object
		return true
	}
	return false
}

type cached struct {
	node   *binTreeNode
	cached C
}

var _ cachingNodeIterator = &cachingPriorityNodeIterator{}

// The caching only useful when in reverse order, since you have to visit parent nodes first for it to be useful.
func newPostOrderNodeIterator(
	forward, addedOnly bool,
	start, end *binTreeNode,
	ctracker *changeTracker,
) subNodeCachingIterator {
	return newPostOrderNodeIteratorBounded(
		nil,
		forward, addedOnly,
		start, end,
		ctracker)
}

func newPostOrderNodeIteratorBounded(
	bnds *bounds,
	forward, addedOnly bool,
	start, end *binTreeNode,
	ctracker *changeTracker) subNodeCachingIterator {
	var op func(current *binTreeNode, end *binTreeNode) *binTreeNode
	if forward {
		op = (*binTreeNode).nextPostOrderNode
	} else {
		op = (*binTreeNode).previousPostOrderNode
	}
	// do the added-only filter first, because it is simpler
	if addedOnly {
		wrappedOp := op
		op = func(currentNode *binTreeNode, endNode *binTreeNode) *binTreeNode {
			return currentNode.nextAdded(endNode, wrappedOp)
		}
	}
	if bnds != nil {
		wrappedOp := op
		op = func(currentNode *binTreeNode, endNode *binTreeNode) *binTreeNode {
			return currentNode.nextInBounds(endNode, wrappedOp, bnds)
		}
	}
	return newSubNodeCachingIterator(
		bnds,
		forward, addedOnly,
		start, end,
		ctracker,
		op,
		!forward,
		!forward || addedOnly)
}

// The caching only useful when in forward order, since you have to visit parent nodes first for it to be useful.
func newPreOrderNodeIterator(
	forward, addedOnly bool,
	start, end *binTreeNode,
	ctracker *changeTracker) subNodeCachingIterator {
	return newPreOrderNodeIteratorBounded(
		nil,
		forward, addedOnly,
		start, end,
		ctracker)
}

func newPreOrderNodeIteratorBounded(
	bnds *bounds,
	forward, addedOnly bool,
	start, end *binTreeNode,
	ctracker *changeTracker) subNodeCachingIterator {
	var op func(current *binTreeNode, end *binTreeNode) *binTreeNode
	if forward {
		op = (*binTreeNode).nextPreOrderNode
	} else {
		op = (*binTreeNode).previousPreOrderNode
	}
	// do the added-only filter first, because it is simpler
	if addedOnly {
		wrappedOp := op
		op = func(currentNode *binTreeNode, endNode *binTreeNode) *binTreeNode {
			return currentNode.nextAdded(endNode, wrappedOp)
		}
	}
	if bnds != nil {
		wrappedOp := op
		op = func(currentNode *binTreeNode, endNode *binTreeNode) *binTreeNode {
			return currentNode.nextInBounds(endNode, wrappedOp, bnds)
		}
	}
	return newSubNodeCachingIterator(
		bnds,
		forward, addedOnly,
		start, end,
		ctracker,
		op,
		forward,
		forward || addedOnly)
}

func newSubNodeCachingIterator(
	bnds *bounds,
	forward, addedOnly bool,
	start, end *binTreeNode,
	ctracker *changeTracker,
	nextOperator func(current *binTreeNode, end *binTreeNode) *binTreeNode,
	allowCaching,
	allowRemove bool,
) subNodeCachingIterator {
	res := subNodeCachingIterator{
		allowCaching:        allowCaching,
		allowRemove:         allowRemove,
		stackIndex:          -1,
		bnds:                bnds,
		isForward:           forward,
		addedOnly:           addedOnly,
		binTreeNodeIterator: binTreeNodeIterator{end: end},
	}
	res.setChangeTracker(ctracker)
	res.operator = nextOperator
	res.next = res.getStart(start, end, bnds, addedOnly)
	return res
}

const ipv6BitCount = 128
const stackSize = ipv6BitCount + 2 // 129 for prefixes /0 to /128 and also 1 more for non-prefixed

type subNodeCachingIterator struct {
	binTreeNodeIterator

	cacheItem  C
	nextKey    e
	nextCached C
	stack      []interface{}
	stackIndex int

	bnds                 *bounds
	addedOnly, isForward bool

	// Both these fields are not really necessary because
	// the caching and removal functionality should not be exposed when it is not usable.
	// The interfaces will not include the caching and Remove() methods in the cases where they are not usable.
	// So these fields are both runtime checks for coding errors.
	allowCaching, allowRemove bool
}

func (iter *subNodeCachingIterator) Next() *binTreeNode {
	result := iter.binTreeNodeIterator.Next()
	if result != nil && iter.allowCaching {
		iter.populateCacheItem(result)
	}
	return result
}

func (iter *subNodeCachingIterator) GetCached() C {
	if !iter.allowCaching {
		panic("no caching allowed, this code path should not be accessible")
	}
	return iter.cacheItem
}

func (iter *subNodeCachingIterator) populateCacheItem(current *binTreeNode) {
	nextKey := iter.nextKey
	if nextKey != nil && current.getKey() == nextKey {
		iter.cacheItem = iter.nextCached
		iter.nextCached = nil
		nextKey = nil
	} else {
		stack := iter.stack
		if stack != nil {
			stackIndex := iter.stackIndex
			if stackIndex >= 0 && stack[stackIndex] == current.getKey() {
				iter.cacheItem = stack[stackIndex+stackSize].(C)
				stack[stackIndex+stackSize] = nil
				stack[stackIndex] = nil
				iter.stackIndex--
			} else {
				iter.cacheItem = nil
			}
		} else {
			iter.cacheItem = nil
		}
	}
}

func (iter *subNodeCachingIterator) Remove() *binTreeNode {
	if !iter.allowRemove {
		// Example:
		// Suppose we are at right sub-node, just visited left.  Next node is parent, but not added.
		// When right is removed, so is the parent, so that the left takes its place.
		// But parent is our next node.  Now our next node is invalid.  So we are lost.
		// This is avoided for iterators that are "added" only.
		panic("no removal allowed, this code path should not be accessible")
		return nil
	}
	return iter.binTreeNodeIterator.Remove()
}

func (iter *subNodeCachingIterator) checkCaching() {
	if !iter.allowCaching {
		panic("no caching allowed, this code path should not be accessible")
	}
}

func (iter *subNodeCachingIterator) CacheWithLowerSubNode(object C) bool {
	iter.checkCaching()
	if iter.isForward {
		return iter.cacheWithFirstSubNode(object)
	}
	return iter.cacheWithSecondSubNode(object)

}

func (iter *subNodeCachingIterator) CacheWithUpperSubNode(object C) bool {
	iter.checkCaching()
	if iter.isForward {
		return iter.cacheWithSecondSubNode(object)
	}
	return iter.cacheWithFirstSubNode(object)
}

// the sub-node will be the next visited node
func (iter *subNodeCachingIterator) cacheWithFirstSubNode(object C) bool {
	iter.checkCaching()
	if iter.current != nil {
		var firstNode *binTreeNode
		if iter.isForward {
			firstNode = iter.current.getLowerSubNode()
		} else {
			firstNode = iter.current.getUpperSubNode()
		}
		if firstNode != nil {
			if (iter.addedOnly && !firstNode.IsAdded()) || (iter.bnds != nil && !iter.bnds.isInBounds(firstNode.getKey())) {
				//firstNode = getToNextOperation().apply(firstNode, current);
				firstNode = iter.operator(firstNode, iter.current)
			}
			if firstNode != nil {
				// the lower sub-node is always next if it exists
				iter.nextKey = firstNode.getKey()
				//System.out.println(current + " cached with " + firstNode + ": " + object);
				iter.nextCached = object
				return true
			}
		}
	}
	return false
}

// the sub-node will only be the next visited node if there is no other sub-node,
// otherwise it might not be visited for a while
func (iter *subNodeCachingIterator) cacheWithSecondSubNode(object C) bool {
	iter.checkCaching()
	if iter.current != nil {
		var secondNode *binTreeNode
		if iter.isForward {
			secondNode = iter.current.getUpperSubNode()
		} else {
			secondNode = iter.current.getLowerSubNode()
		}
		if secondNode != nil {
			if (iter.addedOnly && !secondNode.IsAdded()) || (iter.bnds != nil && !iter.bnds.isInBounds(secondNode.getKey())) {
				//secondNode = getToNextOperation().apply(secondNode, current);
				secondNode = iter.operator(secondNode, iter.current)
			}
			if secondNode != nil {
				// if there is no lower node, we can use the nextCached field since upper is next when no lower sub-node
				var firstNode *binTreeNode
				if iter.isForward {
					firstNode = iter.current.getLowerSubNode()
				} else {
					firstNode = iter.current.getUpperSubNode()
				}
				if firstNode == nil {
					iter.nextKey = secondNode.getKey()
					iter.nextCached = object
				} else {
					if iter.stack == nil {
						iter.stack = make([]interface{}, stackSize<<1)
					}
					iter.stackIndex++
					iter.stack[iter.stackIndex] = secondNode.getKey()
					iter.stack[iter.stackIndex+stackSize] = object
				}
				return true
			}
		}
	}
	return false
}

var _ cachingNodeIterator = &subNodeCachingIterator{}
