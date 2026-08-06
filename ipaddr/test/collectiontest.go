//
// Copyright 2026 Sean C Foley
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

package test

import (
	"fmt"
	"math"
	"math/big"

	"github.com/seancfoley/ipaddress-go/ipaddr"
)

type collectionTester struct {
	testBase

	rangeListTestCount, rangeListFailCount,
	multiListTestCount, singleListTestCount, addressTestCount int
}

type TestResult interface {
	runTest()
}

var printPass, print, printResults bool

type rangeTest struct {
	isIPv6 bool

	range1, range2, expectedIntersection, expectedUnion, range1RemoveRange2, range2RemoveRange1 *ipaddr.IPAddressSeqRangeList

	t *collectionTester
}

func newRangeResult(t *collectionTester, isIPv6 bool, range1, range2, intersection, union, range1RemoveRange2, range2RemoveRange1 *ipaddr.IPAddressSeqRangeList) *rangeTest {
	return &rangeTest{
		isIPv6:               isIPv6,
		range1:               range1,
		range2:               range2,
		expectedIntersection: intersection,
		expectedUnion:        union,
		range1RemoveRange2:   range1RemoveRange2,
		range2RemoveRange1:   range2RemoveRange1,
		t:                    t,
	}
}

func (rangeTest *rangeTest) addRangeFailure(message string, list ipaddr.IPAddressCollection) {
	rangeTest.t.addRangeFailure(message, list)
}

func (rangeTest *rangeTest) testOverlapIndex(range1, range2 *ipaddr.IPAddressSeqRangeList) {
	previousOverlaps := false
	firstTime := true
	overlapsCannotChange := false
	rng := rangeTest.range1.Clone()
	for {
		intersection := rng.IntersectIntoNew(range2)
		overlaps := !intersection.IsEmpty()
		if !firstTime && overlapsCannotChange && overlaps != previousOverlaps {
			rangeTest.t.addRangeFailure("fail overlaps change for range 1: "+range1.String()+" and range 2: "+range2.String(), range1)
		}
		index := rng.IndexOfSeqRangeOverlappingSeqRangeList(range2)
		if overlaps {
			if index < 0 {
				rangeTest.addRangeFailure("fail overlap index negative for range 1: "+range1.String()+" and range 2: "+range2.String(), range1)
			} else {
				if printPass {
					fmt.Println("pass")
				}
			}
		} else {
			if index >= 0 {
				rangeTest.addRangeFailure("fail overlap index positive for range 1: "+range1.String()+" and range 2: "+range2.String(), range1)
			} else {
				if printPass {
					fmt.Println("pass")
				}
			}
		}
		if rng.IsEmpty() {
			break
		}
		overlapsCannotChange = index != 0
		previousOverlaps = overlaps
		rng.RemoveSeqRangeAt(0)
		firstTime = false
	}
}

func (rangeTest *rangeTest) runTest() {

	t := rangeTest.t

	t.multiListTestCount++

	range1 := rangeTest.range1
	range2 := rangeTest.range2
	expectedUnion := rangeTest.expectedUnion
	expectedIntersection := rangeTest.expectedIntersection

	trie1, trie2 := convertListToTrie(range1), convertListToTrie(range2)
	expectedUnionTrie := convertListToTrie(expectedUnion)

	testOp(rangeTest, range1, range2, (*ipaddr.IPAddressSeqRangeList).JoinIntoNew, "join", expectedUnion)
	testOp(rangeTest, trie1, trie2, (*ipaddr.IPAddressContainmentTrie).JoinIntoNew, "join", expectedUnionTrie)

	testOp(rangeTest, range2, range1, (*ipaddr.IPAddressSeqRangeList).JoinIntoNew, "join", expectedUnion)
	testOp(rangeTest, trie2, trie1, (*ipaddr.IPAddressContainmentTrie).JoinIntoNew, "join", expectedUnionTrie)

	expectedIntersectionTrie := convertListToTrie(expectedIntersection)

	testOp(rangeTest, range1, range2, (*ipaddr.IPAddressSeqRangeList).IntersectIntoNew, "intersect", expectedIntersection)
	testOp(rangeTest, trie1, trie2, (*ipaddr.IPAddressContainmentTrie).IntersectIntoNew, "intersect", expectedIntersectionTrie)

	testOp(rangeTest, range2, range1, (*ipaddr.IPAddressSeqRangeList).IntersectIntoNew, "intersect", expectedIntersection)
	testOp(rangeTest, trie2, trie1, (*ipaddr.IPAddressContainmentTrie).IntersectIntoNew, "intersect", expectedIntersectionTrie)

	expectedOverlap := !expectedIntersection.IsEmpty()

	testBooleanOp(rangeTest, range1, range2, (*ipaddr.IPAddressSeqRangeList).OverlapsOther, "overlaps", expectedOverlap)
	testBooleanOp(rangeTest, trie1, trie2, (*ipaddr.IPAddressContainmentTrie).OverlapsOther, "overlaps", expectedOverlap)

	testBooleanOp(rangeTest, range2, range1, (*ipaddr.IPAddressSeqRangeList).OverlapsOther, "overlaps", expectedOverlap)
	testBooleanOp(rangeTest, trie2, trie1, (*ipaddr.IPAddressContainmentTrie).OverlapsOther, "overlaps", expectedOverlap)

	rangeTest.testOverlapIndex(range1, range2)
	rangeTest.testOverlapIndex(range2, range1)

	range1RemoveRange2 := rangeTest.range1RemoveRange2
	range2RemoveRange1 := rangeTest.range2RemoveRange1

	trie1RemoveTrie2 := convertListToTrie(range1RemoveRange2)

	testOp(rangeTest, range1, range2, (*ipaddr.IPAddressSeqRangeList).RemoveIntoNew, "remove", range1RemoveRange2)
	testOp(rangeTest, trie1, trie2, (*ipaddr.IPAddressContainmentTrie).RemoveIntoNew, "remove", trie1RemoveTrie2)

	trie2RemoveTrie1 := convertListToTrie(range2RemoveRange1)

	testOp(rangeTest, range2, range1, (*ipaddr.IPAddressSeqRangeList).RemoveIntoNew, "remove", range2RemoveRange1)
	testOp(rangeTest, trie2, trie1, (*ipaddr.IPAddressContainmentTrie).RemoveIntoNew, "remove", trie2RemoveTrie1)

	// A contains (A intersect B)

	intersection := binaryOp(range1, range2, (*ipaddr.IPAddressSeqRangeList).IntersectIntoNew, "intersect")
	contains(t, range1, intersection, true)
	contains(t, range2, intersection, true)

	intersectionTrie := binaryOp(trie1, trie2, (*ipaddr.IPAddressContainmentTrie).IntersectIntoNew, "intersect")
	t.collectionsMatch(intersectionTrie, intersection)

	contains(t, trie1, intersectionTrie, true)
	contains(t, trie2, intersectionTrie, true)

	// (A intersect B) contains A iff B contains A

	r1ContainsAllr2 := range1.ContainsOther(range2)
	r2ContainsAllr1 := range2.ContainsOther(range1)
	contains(t, intersection, range1, r2ContainsAllr1)
	contains(t, intersection, range2, r1ContainsAllr2)

	t1ContainsAllt2 := trie1.ContainsOther(trie2)
	t2ContainsAllt1 := trie2.ContainsOther(trie1)
	contains(t, intersectionTrie, trie1, t2ContainsAllt1)
	contains(t, intersectionTrie, trie2, t1ContainsAllt2)

	union := binaryOp(range1, range2, (*ipaddr.IPAddressSeqRangeList).JoinIntoNew, "join")
	contains(t, union, range1, true)
	contains(t, union, range2, true)

	unionTrie := binaryOp(trie1, trie2, (*ipaddr.IPAddressContainmentTrie).JoinIntoNew, "join")
	t.collectionsMatch(unionTrie, union)
	contains(t, unionTrie, trie1, true)
	contains(t, unionTrie, trie2, true)
	//fmt.Println("got to here 6")

	contains(t, range1, union, r1ContainsAllr2)
	contains(t, range2, union, r2ContainsAllr1)

	contains(t, trie1, unionTrie, t1ContainsAllt2)
	contains(t, trie2, unionTrie, t2ContainsAllt1)

	empty := &ipaddr.IPAddressSeqRangeList{}

	// removal of oneself always results in nothing
	testOp(rangeTest, range1, range1, (*ipaddr.IPAddressSeqRangeList).RemoveIntoNew, "remove", empty)
	testOp(rangeTest, range2, range2, (*ipaddr.IPAddressSeqRangeList).RemoveIntoNew, "remove", empty)

	emptyTrie := &ipaddr.IPAddressContainmentTrie{}
	testOp(rangeTest, trie1, trie1, (*ipaddr.IPAddressContainmentTrie).RemoveIntoNew, "remove", emptyTrie)
	testOp(rangeTest, trie2, trie2, (*ipaddr.IPAddressContainmentTrie).RemoveIntoNew, "remove", emptyTrie)

	// intersection with oneself results in the same
	testOp(rangeTest, range1, range1, (*ipaddr.IPAddressSeqRangeList).IntersectIntoNew, "intersect", range1)
	testOp(rangeTest, range2, range2, (*ipaddr.IPAddressSeqRangeList).IntersectIntoNew, "intersect", range2)

	testOp(rangeTest, trie1, trie1, (*ipaddr.IPAddressContainmentTrie).IntersectIntoNew, "intersect", trie1)
	testOp(rangeTest, trie2, trie2, (*ipaddr.IPAddressContainmentTrie).IntersectIntoNew, "intersect", trie2)

	// union with oneself results in the same
	testOp(rangeTest, range1, range1, (*ipaddr.IPAddressSeqRangeList).JoinIntoNew, "join", range1)
	testOp(rangeTest, range2, range2, (*ipaddr.IPAddressSeqRangeList).JoinIntoNew, "join", range2)

	testOp(rangeTest, trie1, trie1, (*ipaddr.IPAddressContainmentTrie).JoinIntoNew, "intersect", trie1)
	testOp(rangeTest, trie2, trie2, (*ipaddr.IPAddressContainmentTrie).JoinIntoNew, "intersect", trie2)

	// removing one from the other, removing the other from the one, then taking the union, is the same as removing the intersection from the union
	t.collectionsMatch(
		binaryOp(binaryOp(range1, range2, (*ipaddr.IPAddressSeqRangeList).RemoveIntoNew, "remove"),
			binaryOp(range2, range1, (*ipaddr.IPAddressSeqRangeList).RemoveIntoNew, "remove"),
			(*ipaddr.IPAddressSeqRangeList).JoinIntoNew, "join"),
		binaryOp(binaryOp(range1, range2, (*ipaddr.IPAddressSeqRangeList).JoinIntoNew, "join"),
			binaryOp(range1, range2, (*ipaddr.IPAddressSeqRangeList).IntersectIntoNew, "intersect"),
			(*ipaddr.IPAddressSeqRangeList).RemoveIntoNew, "remove"))

	t.collectionsMatch(
		binaryOp(binaryOp(trie1, trie2, (*ipaddr.IPAddressContainmentTrie).RemoveIntoNew, "remove"),
			binaryOp(trie2, trie1, (*ipaddr.IPAddressContainmentTrie).RemoveIntoNew, "remove"),
			(*ipaddr.IPAddressContainmentTrie).JoinIntoNew, "join"),
		binaryOp(binaryOp(trie1, trie2, (*ipaddr.IPAddressContainmentTrie).JoinIntoNew, "join"),
			binaryOp(trie1, trie2, (*ipaddr.IPAddressContainmentTrie).IntersectIntoNew, "intersect"),
			(*ipaddr.IPAddressContainmentTrie).RemoveIntoNew, "remove"))

	var everythingStr string
	if rangeTest.isIPv6 {
		everythingStr = "::/0"
	} else {
		everythingStr = "0.0.0.0/0"
	}
	everythingAddr := ipaddr.NewIPAddressString(everythingStr).GetAddress().ToPrefixBlock()

	// De Morgan's Law 1
	// complement of the union is the same as the intersection of the complements
	t.collectionsMatch(complementWrapper(
		binaryOp(range1, range2, (*ipaddr.IPAddressSeqRangeList).JoinIntoNew, "join"), everythingAddr),
		binaryOp(complementWrapper(range1, everythingAddr),
			complementWrapper(range2, everythingAddr), (*ipaddr.IPAddressSeqRangeList).IntersectIntoNew, "intersect"))

	t.collectionsMatch(complementWrapper(
		binaryOp(trie1, trie2, (*ipaddr.IPAddressContainmentTrie).JoinIntoNew, "join"), everythingAddr),
		binaryOp(complementWrapper(trie1, everythingAddr),
			complementWrapper(trie2, everythingAddr), (*ipaddr.IPAddressContainmentTrie).IntersectIntoNew, "intersect"))

	// De Morgan's Law 2
	// complement of the intersection is the same as the union of the complements
	t.collectionsMatch(
		complementWrapper(binaryOp(range1, range2, (*ipaddr.IPAddressSeqRangeList).IntersectIntoNew, "intersect"), everythingAddr),
		binaryOp(complementWrapper(range1, everythingAddr),
			complementWrapper(range2, everythingAddr), (*ipaddr.IPAddressSeqRangeList).JoinIntoNew, "join"))

	t.collectionsMatch(
		complementWrapper(binaryOp(trie1, trie2, (*ipaddr.IPAddressContainmentTrie).IntersectIntoNew, "intersect"), everythingAddr),
		binaryOp(complementWrapper(trie1, everythingAddr),
			complementWrapper(trie2, everythingAddr), (*ipaddr.IPAddressContainmentTrie).JoinIntoNew, "join"))

	removeIntersection(rangeTest, range1, range2)
	removeIntersection(rangeTest, range2, range1)

	removeIntersection(rangeTest, trie1, trie2)
	removeIntersection(rangeTest, trie2, trie1)

	unionDoubleIntersect(rangeTest, range1, range2, everythingAddr)
	unionDoubleIntersect(rangeTest, range2, range1, everythingAddr)

	unionDoubleIntersect(rangeTest, trie1, trie2, everythingAddr)
	unionDoubleIntersect(rangeTest, trie2, trie1, everythingAddr)

	// double complement results in the original

	doubleComplement(rangeTest, range1, everythingAddr)
	doubleComplement(rangeTest, range2, everythingAddr)

	doubleComplement(rangeTest, trie1, everythingAddr)
	doubleComplement(rangeTest, trie2, everythingAddr)

	intersectUnion(rangeTest, range1, range2)
	intersectUnion(rangeTest, range2, range1)

	intersectUnion(rangeTest, trie1, trie2)
	intersectUnion(rangeTest, trie2, trie1)

	unionIntersect(rangeTest, range1, range2)
	unionIntersect(rangeTest, range2, range1)

	unionIntersect(rangeTest, trie1, trie2)
	unionIntersect(rangeTest, trie2, trie1)

	removeIntersectComplement(rangeTest, range1, range2, everythingAddr)
	removeIntersectComplement(rangeTest, range2, range1, everythingAddr)

	everythingNothing(rangeTest, range1, everythingAddr)
	everythingNothing(rangeTest, range2, everythingAddr)

	everythingNothing(rangeTest, trie1, everythingAddr)
	everythingNothing(rangeTest, trie2, everythingAddr)

	rangeTest.testRangeListSpans(expectedUnion, range1, range2)

	t.testIterateRangeList(range1)
	t.testIterateRangeList(range2)

	t.testIterateTrie(trie1)
	t.testIterateTrie(trie2)

	rangeTest.testIntegerOps(range1)
	rangeTest.testIntegerOps(range2)

	testIntegerOps(rangeTest, trie1, range1)
	testIntegerOps(rangeTest, trie2, range2)

	testIntegerOps(rangeTest, range1, range1)
	testIntegerOps(rangeTest, range2, range2)

	// if we had 3 to play with, these two require 3 sets:
	// A union (B intersect C) = (A union B) intersect (A union C)
	// A intersect (B union C) = (A intersect B) union (A intersect C)

}

func (r *rangeTest) testRangeListSpans(list, joined1, joined2 *ipaddr.IPAddressSeqRangeList) {
	resultPrefixBlocks := list.SpanWithPrefixBlocks()
	resultSequentialBlocks := list.SpanWithSequentialBlocks()
	prefixBlocks1 := joined1.SpanWithPrefixBlocks()
	prefixBlocks2 := joined2.SpanWithPrefixBlocks()
	seqBlocks1 := joined1.SpanWithSequentialBlocks()
	seqBlocks2 := joined2.SpanWithSequentialBlocks()

	t := r.t
	t.testPrefixBlockSpan(resultPrefixBlocks, prefixBlocks1, prefixBlocks2, list)
	t.testPrefixBlockSpan(resultPrefixBlocks, seqBlocks1, seqBlocks2, list)
	t.testSequentialBlockSpan(resultSequentialBlocks, seqBlocks1, seqBlocks2, list)
	t.testSequentialBlockSpan(resultSequentialBlocks, prefixBlocks1, prefixBlocks2, list)
}

// removing the intersection is the same as removing the original
func removeIntersection[S ipaddr.IPAddressCollConstraint[S, *ipaddr.IPAddress]](rangeTest *rangeTest, range1, range2 S) {
	rangeTest.t.collectionsMatch(binaryOp(range1, binaryOp(range1, range2, S.IntersectIntoNew, "intersect"), S.RemoveIntoNew, "remove"),
		binaryOp(range1, range2, S.RemoveIntoNew, "remove"))
}

// double complement results in the original
func doubleComplement[S ipaddr.IPAddressCollConstraint[S, *ipaddr.IPAddress]](rangeTest *rangeTest, list S, everythingAddr *ipaddr.IPAddress) {
	t := rangeTest.t
	complement := complementWrapper(list, everythingAddr)
	complementCount := complement.GetCount()
	matchesCount(t, everythingAddr.GetCount(), complementCount.Add(complementCount, list.GetCount()), list)
	//complement = t.convertEmptyVersionForComplement(complement, everythingAddr)
	t.collectionsMatch(
		list,
		complementWrapper(complement, everythingAddr))
}

// A intersect (A union B) = A
func intersectUnion[S ipaddr.IPAddressCollConstraint[S, *ipaddr.IPAddress]](rangeTest *rangeTest, coll1, coll2 S) {
	t := rangeTest.t
	t.collectionsMatch(
		coll1,
		binaryOp(coll1, binaryOp(coll1, coll2, S.JoinIntoNew, "join"),
			S.IntersectIntoNew, "intersect"))
}

// intersect with complement and then with original, take union, should be the original
func unionDoubleIntersect[S ipaddr.IPAddressCollConstraint[S, *ipaddr.IPAddress]](rangeTest *rangeTest, range1, range2 S, everythingAddr *ipaddr.IPAddress) {
	rangeTest.t.collectionsMatch(
		range2,
		binaryOp(
			binaryOp(range1, range2, S.IntersectIntoNew, "intersect"),
			binaryOp(
				complementWrapper(range1, everythingAddr),
				range2,
				S.IntersectIntoNew, "intersect"),
			S.JoinIntoNew, "join"))
}

// A union (A intersect B) = A
func unionIntersect[S ipaddr.IPAddressCollConstraint[S, *ipaddr.IPAddress]](rangeTest *rangeTest, range1, range2 S) {
	rangeTest.t.collectionsMatch(
		range1,
		binaryOp(range1, binaryOp(range1, range2, S.IntersectIntoNew, "intersect"),
			S.JoinIntoNew, "join"))
}

// A remove B = A intersect (B complement)
func removeIntersectComplement[S ipaddr.IPAddressCollConstraint[S, *ipaddr.IPAddress]](rangeTest *rangeTest, coll1, coll2 S, everythingAddr *ipaddr.IPAddress) {
	rangeTest.t.collectionsMatch(
		binaryOp(coll1, coll2, S.RemoveIntoNew, "remove"),
		binaryOp(coll1, complementWrapper(coll2, everythingAddr),
			S.IntersectIntoNew, "intersect"))
}

// A Union (A complement) = everything
// A intersect (A complement) = nothing
func everythingNothing[S ipaddr.IPAddressCollConstraint[S, *ipaddr.IPAddress]](rangeTest *rangeTest, range1 S, everythingAddr *ipaddr.IPAddress) {
	t := rangeTest.t
	if !range1.IsEmpty() {
		complement := complementWrapper(range1, everythingAddr)
		nothing := ipaddr.IPAddressSeqRangeList{}
		t.collectionsMatch(
			&nothing,
			binaryOp(range1, complement, S.IntersectIntoNew, "intersect"))

		nothing.Add(everythingAddr)
		everything := &nothing
		t.collectionsMatch(
			everything,
			binaryOp(range1, complement, S.JoinIntoNew, "join"))
	}
}

func testOp[S ipaddr.IPAddressCollConstraint[S, *ipaddr.IPAddress]](rangeTest *rangeTest, coll1, coll2 S, op func(one, two S) S, opName string, expected S) {
	t := rangeTest.t
	t.rangeListTestCount++
	res := op(coll1, coll2)
	if res.Equal(expected) {
		if res.GetCount().Cmp(expected.GetCount()) != 0 {
			t.addRangeFailure("counts fail "+opName+" for "+coll1.String()+" and "+coll2.String(), coll1)
		} else {
			l1, ok1 := any(res).(*ipaddr.IPAddressSeqRangeList)
			l2, ok2 := any(expected).(*ipaddr.IPAddressSeqRangeList)
			if ok1 && ok2 && l1.GetSeqRangeCount() != l2.GetSeqRangeCount() {
				t.addRangeFailure("range counts fail "+opName+" for "+coll1.String()+" and "+coll2.String(), coll1)
			} else if printPass {
				fmt.Println("pass " + opName)
			}
		}
	} else {
		t.addRangeFailure("fail "+opName+" for range 1: "+coll2.String()+" and  range 2: "+coll2.String()+" expected: "+expected.String()+" actual: "+res.String(), coll1)
	}
}

func testBooleanOp[S ipaddr.IPAddressCollConstraint[S, *ipaddr.IPAddress]](rangeTest *rangeTest, coll1, coll2 S, op func(one, two S) bool, opName string, expected bool) {
	t := rangeTest.t
	t.rangeListTestCount++
	res := op(coll1, coll2)
	if res == expected {
		if printPass {
			fmt.Println("pass " + opName)
		}
	} else {
		t.addRangeFailure("fail "+opName+" for range 1: "+coll1.String()+" and  range 2: "+coll2.String()+" expected: "+fmt.Sprint(expected)+" actual: "+fmt.Sprint(res), coll1)
	}
}

func testIntegerOps[R ipaddr.IPAddressCollConstraint[R, *ipaddr.IPAddress], S ipaddr.IPAddressCollConstraint[S, *ipaddr.IPAddress]](rangeTest *rangeTest, coll R, equivalentColl S) {

	count := coll.GetCount()

	if count.Cmp(equivalentColl.GetCount()) != 0 {
		rangeTest.addRangeFailure("inconsistent collection counts", coll)
	}
	var trie *ipaddr.IPAddressContainmentTrie
	anyColl := any(coll)
	list, ok := anyColl.(*ipaddr.IPAddressSeqRangeList)
	if !ok {
		trie, _ = anyColl.(*ipaddr.IPAddressContainmentTrie)
	}
	if trie != nil && trie.GetPrefixBlockCount() > 50 {
		return
	} else if list != nil && list.GetSeqRangeCount() > 100 {
		return
	}

	iterColl := coll.SpanningPrefixBlockIterator()
	iterEquiv := coll.SpanningPrefixBlockIterator()

	t := rangeTest.t

	if coll.IsEmpty() {
		if iterColl.HasNext() || iterEquiv.HasNext() {
			rangeTest.addRangeFailure("iterator not empty", coll)
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					// nothing in the collection
				}
			}()
			getAddr := list.GetBig(bigZeroConst())
			t.addRangeFailure("unexpected address "+getAddr.String()+" at collection address index 0", coll)
		}()
		func() {
			defer func() {
				if r := recover(); r != nil {
					// nothing in the collection
				}
			}()
			getAddr := list.Get(0)
			t.addRangeFailure("unexpected address "+getAddr.String()+" at collection address index 0", coll)
		}()
		return
	}

	firstBlock := iterColl.Next()
	equivBlock := iterEquiv.Next()

	if !firstBlock.Equal(equivBlock) {
		t.addRangeFailure("first prefix block mismatch "+firstBlock.String()+" vs expected "+equivBlock.String(), coll)
	}
	if trie != nil && !firstBlock.Equal(trie.GetLowerPrefixBlock()) {
		t.addRangeFailure("first prefix block mismatch "+firstBlock.String()+" vs trie "+trie.GetLowerPrefixBlock().String(), coll)
	}

	targetIndex1 := bigZero().Set(count)
	targetIndex1.Div(targetIndex1, big.NewInt(3))

	targetIndex2 := bigZero().Set(targetIndex1)
	targetIndex2.Mul(targetIndex2, big.NewInt(2))

	var targetBlock1, targetBlock2, lastBlock *ipaddr.IPAddress
	var targetBlock1FirstAddrIndex, targetBlock2FirstAddrIndex, lastBlockFirstAddrIndex *big.Int
	var hasTargetBlock1, hasTargetBlock2, hasLastBlock bool

	currentCount := firstBlock.GetCount()
	if targetIndex1.Cmp(currentCount) < 0 {
		targetBlock1 = firstBlock
	}
	//fmt.Println("current count is", currentCount, "following", firstBlock)
	var (
		next      *ipaddr.IPAddress
		prevCount *big.Int
	)
	for {
		if !iterColl.HasNext() {
			if iterEquiv.HasNext() {
				rangeTest.addRangeFailure("iterator unexpectedly not empty", coll)
			}
			if next != nil {
				lastBlock = next
				hasLastBlock = true
				lastBlockFirstAddrIndex = prevCount
				if hasTargetBlock1 && lastBlock.Equal(targetBlock1) {
					hasTargetBlock1 = false
				}
				if hasTargetBlock2 && lastBlock.Equal(targetBlock2) {
					hasTargetBlock2 = false
				}
				if trie != nil && !lastBlock.Equal(trie.GetUpperPrefixBlock()) {
					t.addRangeFailure("last prefix block mismatch "+firstBlock.String()+" vs trie "+trie.GetLowerPrefixBlock().String(), coll)
				}
			} else {
				// last block is first block
				if trie != nil && !firstBlock.Equal(trie.GetUpperPrefixBlock()) {
					t.addRangeFailure("first/last prefix block mismatch "+firstBlock.String()+" vs trie "+trie.GetLowerPrefixBlock().String(), coll)
				}
			}
			break
		}
		if !iterEquiv.HasNext() {
			rangeTest.addRangeFailure("iterator unexpectedly empty", coll)
		}
		next = iterColl.Next()
		equivNext := iterEquiv.Next()
		if !next.Equal(equivNext) {
			t.addRangeFailure("prefix block mismatch "+firstBlock.String()+" vs expected "+equivBlock.String(), coll)
		}
		prevCount = bigZero().Set(currentCount)
		currentCount.Add(currentCount, next.GetCount())
		if targetBlock1 == nil {
			if targetIndex1.Cmp(currentCount) < 0 {
				targetBlock1 = next
				hasTargetBlock1 = true
				targetBlock1FirstAddrIndex = prevCount
			}
		}
		if targetBlock2 == nil {
			if targetIndex2.Cmp(currentCount) < 0 {
				targetBlock2 = next
				hasTargetBlock2 = !targetBlock1.Equal(targetBlock2)
				targetBlock2FirstAddrIndex = prevCount
			}
		}
	}

	// at this point we 1, 2, 3, o4 4 blocks to check:
	// firstBlock, maybe lastBlock, maybe targetBlock1, maybe targetBlock2

	removeAndPutBacks(rangeTest, coll, trie, count, firstBlock, bigZeroConst())
	if hasTargetBlock1 {
		removeAndPutBacks(rangeTest, coll, trie, count, targetBlock1, targetBlock1FirstAddrIndex)
	}
	if hasTargetBlock2 {
		removeAndPutBacks(rangeTest, coll, trie, count, targetBlock2, targetBlock2FirstAddrIndex)
	}
	if hasLastBlock {
		removeAndPutBacks(rangeTest, coll, trie, count, lastBlock, lastBlockFirstAddrIndex)
		count2 := bigZero().Set(lastBlockFirstAddrIndex)
		count2.Add(count2, lastBlock.GetCount())
		if count.Cmp(count2) != 0 {
			t.addRangeFailure("count mismatch "+count.String()+" vs expected "+count2.String(), coll)
		}
	} else {
		if count.Cmp(firstBlock.GetCount()) != 0 {
			t.addRangeFailure("count mismatch "+count.String()+" vs expected "+firstBlock.GetCount().String(), coll)
		}
	}

	checkOutOfBounds(rangeTest, coll, trie, count, big.NewInt(-1))

	negCount := firstBlock.GetValue()
	if negCount.Sign() == 0 {
		negCount.Sub(negCount, bigOneConst())
	} else {
		negCount.Neg(negCount)
	}
	checkOutOfBounds(rangeTest, coll, trie, count, negCount)

	negCount.Sub(negCount, bigOneConst())
	checkOutOfBounds(rangeTest, coll, trie, count, negCount)

	checkOutOfBounds(rangeTest, coll, trie, count, count)

	countPlus1 := bigZero().Set(count)
	countPlus1.Add(countPlus1, bigOneConst())
	checkOutOfBounds(rangeTest, coll, trie, count, countPlus1)

	checkOutOfBounds(rangeTest, coll, trie, count, big.NewInt(math.MaxInt64))

	checkOutOfBounds(rangeTest, coll, trie, count, bigZero().SetUint64(math.MaxUint64))
}

func removeAndPutBacks[S ipaddr.IPAddressCollConstraint[S, *ipaddr.IPAddress]](
	rangeTest *rangeTest,
	coll S,
	trie *ipaddr.IPAddressContainmentTrie,
	collCount *big.Int,
	block *ipaddr.IPAddress,
	blockFirstAddrIndex *big.Int) {

	first := block.GetLower()
	count := block.GetCount()
	cmp := count.Cmp(big.NewInt(2))
	removeAndPutBack(rangeTest, coll, trie, collCount, first, blockFirstAddrIndex)
	if cmp >= 0 {
		lastIndex := bigZero().Set(blockFirstAddrIndex)
		lastIndex.Add(lastIndex, count).Sub(lastIndex, bigOneConst())
		last := block.GetUpper()
		removeAndPutBack(rangeTest, coll, trie, collCount, last, lastIndex)
		if cmp > 0 {
			count.Div(count, big.NewInt(2))
			index := bigZero().Set(blockFirstAddrIndex)
			index.Add(index, count)
			addr := first.IncrementBig(count)
			removeAndPutBack(rangeTest, coll, trie, collCount, addr, index)
		}
	}
}

func removeAndPutBack[S ipaddr.IPAddressCollConstraint[S, *ipaddr.IPAddress]](
	rangeTest *rangeTest,
	coll S,
	trie *ipaddr.IPAddressContainmentTrie,
	collCount *big.Int,
	target *ipaddr.IPAddress,
	targetAddressIndex *big.Int) {

	// removeAndPutBack checks that the addresses are at their indices,
	// enumerate gives the same index,
	// and that things work fine as you add and remove the addresses

	var originalPrefixBlockCount int
	if trie != nil {
		originalPrefixBlockCount = trie.GetPrefixBlockCount()
	}

	t := rangeTest.t

	isLongish := targetAddressIndex.Cmp(big.NewInt(math.MaxInt64)) <= 0 && targetAddressIndex.Cmp(big.NewInt(math.MinInt64)) >= 0

	originalColl := coll.Clone()
	if originalColl.GetCount().Cmp(collCount) != 0 {
		t.addRangeFailure("inconsistent address counts", coll)
	}
	countMinusOne := bigZero().Set(originalColl.GetCount())
	countMinusOne = countMinusOne.Sub(countMinusOne, bigOneConst())
	isLastAddress := targetAddressIndex.Cmp(countMinusOne) == 0
	isFirstAddress := targetAddressIndex.Cmp(bigZeroConst()) == 0
	getAddr := coll.GetBig(targetAddressIndex)
	if getAddr.IsMultiple() || getAddr.IsPrefixed() {
		t.addRangeFailure("unexpected multiple or prefixed subnet", coll)
	}
	if !getAddr.Equal(target) {
		t.addRangeFailure("unexpected address "+getAddr.String()+" at list address index "+targetAddressIndex.String()+", expected "+target.String(), coll)
		fmt.Println("unexpected address " + getAddr.String() + " at list address index " + targetAddressIndex.String() + ", expected " + target.String() + coll.String())
		coll.GetBig(targetAddressIndex) // the expected is wrong, but in the caller you can see the expected is what we want, the index is wrong
	}
	if isLongish {
		getAddr = coll.Get(targetAddressIndex.Int64())
		if getAddr.IsMultiple() || getAddr.IsPrefixed() {
			t.addRangeFailure("unexpected multiple or prefixed subnet", coll)
		}
		if !getAddr.Equal(target) {
			t.addRangeFailure("unexpected address at list address index", coll)
		}
	}

	if coll.Enumerate(getAddr).Cmp(targetAddressIndex) != 0 {
		t.addRangeFailure("unexpected enumerated address at list increment address index", coll)
	}
	added := coll.Add(getAddr)
	if added {
		t.addRangeFailure("unexpected add to list of address expected to be in list already", coll)
	}
	if coll.GetCount().Cmp(collCount) != 0 {
		t.addRangeFailure("unexpected change in list count", coll)
	}

	// remove it
	var removedAddr *ipaddr.IPAddress
	useLong := false
	if isLongish {
		counter++
		useLong = (counter%2 == 0)
	}
	if useLong {
		removedAddr = coll.RemoveAt(targetAddressIndex.Int64())
	} else {
		removedAddr = coll.RemoveAtBig(targetAddressIndex)
	}
	if removedAddr.IsPrefixed() {
		t.addRangeFailure("addresses from collections should not be prefixed", coll)
	}
	if !removedAddr.Equal(target) {
		t.addRangeFailure("removed address "+removedAddr.String()+" not the expected address "+getAddr.String(), coll)
	}
	if !removedAddr.Equal(getAddr) {
		t.addRangeFailure("removed address "+removedAddr.String()+" not the expected address "+getAddr.String()+" looked up", coll)
	}
	// confirm it is gone
	func() {
		defer func() {
			if r := recover(); r != nil {
				if !isLastAddress {
					t.addRangeFailure("unexpected exception", coll)
				}
			}
		}()
		getAddrAgain := coll.GetBig(targetAddressIndex)
		if removedAddr.Equal(getAddrAgain) || isLastAddress {
			t.addRangeFailure("removed address not gone", coll)
		}
	}()
	if isLongish {
		func() {
			defer func() {
				if r := recover(); r != nil {
					if !isLastAddress {
						t.addRangeFailure("unexpected exception", coll)
					}
				}
			}()
			getAddrAgain := coll.Get(targetAddressIndex.Int64())
			if removedAddr.Equal(getAddrAgain) || isLastAddress {
				t.addRangeFailure("removed address not gone", coll)
			}
		}()
	}

	// confirm it is gone by removing it again
	removedAgain := coll.Remove(removedAddr)
	if removedAgain {
		t.addRangeFailure("removed address not removed", coll)
	}

	// check counts
	countPlusOne := coll.GetCount()
	if originalColl.GetCount().Cmp(countPlusOne.Add(countPlusOne, bigOneConst())) != 0 {
		t.addRangeFailure("list count unexpected after removing address", coll)
	}
	// check enumerate
	isBorderAddress := isLastAddress || isFirstAddress
	enumerated := coll.Enumerate(removedAddr)
	var fails bool
	if isBorderAddress && !coll.IsEmpty() {
		fails = enumerated == nil
	} else {
		fails = enumerated != nil
	}
	if fails {
		t.addRangeFailure("removed address found in collection after being removed", coll)
	}
	enumerated = coll.Enumerate(getAddr)
	if isBorderAddress && !coll.IsEmpty() {
		fails = enumerated == nil
	} else {
		fails = enumerated != nil
	}
	if fails {
		t.addRangeFailure("address found in collection after being removed", coll)
	}

	// add it back
	added = coll.Add(getAddr)
	if !added {
		t.addRangeFailure("address not added as expected after being removed", coll)
	}

	// confirm it is back
	getAddrAgain := coll.GetBig(targetAddressIndex)
	if !removedAddr.Equal(getAddrAgain) {
		t.addRangeFailure("address not found in list as expected after being added back", coll)
	}
	if isLongish {
		getAddrAgain = coll.Get(targetAddressIndex.Int64())
		if !removedAddr.Equal(getAddrAgain) {
			t.addRangeFailure("address not found in list as expected after being added back", coll)
		}
	}

	// check counts
	if trie != nil {
		if originalPrefixBlockCount != trie.GetPrefixBlockCount() {
			t.addRangeFailure("prefix block count not restored to original", coll)
		}
	}
	if collCount.Cmp(coll.GetCount()) != 0 {
		t.addRangeFailure("address count not restored to original", coll)
	}
	// check enumerate
	if coll.Enumerate(removedAddr).Cmp(targetAddressIndex) != 0 {
		t.addRangeFailure("address not found at expeceted location", coll)
	}
	if coll.Enumerate(getAddr).Cmp(targetAddressIndex) != 0 {
		t.addRangeFailure("address not located at expeceted location", coll)
	}
	// confirm the list is back to the same
	if !coll.Equal(originalColl) {
		t.addRangeFailure("restored list not equal to original", coll)
	}
}

// checkAboveBelow checks negative indices and indices beyond the list size, checking increments below the list and beyond the upper value of the list
func checkOutOfBounds[S ipaddr.IPAddressCollConstraint[S, *ipaddr.IPAddress]](
	rangeTest *rangeTest,
	coll S,
	trie *ipaddr.IPAddressContainmentTrie,
	collCount *big.Int,
	targetAddressIndex *big.Int) {

	t := rangeTest.t

	// checkAboveBelow checks Get, GetBig, Enumerate above and below the list.  Which means each one should panic, except Enumerate maybe.

	isLongish := targetAddressIndex.Cmp(big.NewInt(math.MaxInt64)) <= 0 && targetAddressIndex.Cmp(big.NewInt(math.MinInt64)) >= 0

	func() {
		defer func() {
			if r := recover(); r != nil {
				// no longer there
			}
		}()
		getAddr := coll.GetBig(targetAddressIndex)
		t.addRangeFailure("unexpected address "+getAddr.String()+" at collection address index "+targetAddressIndex.String(), coll)
	}()

	if isLongish {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// no longer there
				}
			}()
			getAddr := coll.Get(targetAddressIndex.Int64())
			t.addRangeFailure("unexpected address "+getAddr.String()+" at collection address index", coll)
		}()
	}

	// remove it
	func() {
		defer func() {
			if r := recover(); r != nil {
				// no longer there
			}
		}()
		removedAddr := coll.RemoveAtBig(targetAddressIndex)
		t.addRangeFailure("unexpected address "+removedAddr.String()+" at collection address index", coll)
	}()
	if isLongish {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// no longer there
				}
			}()
			removedAddr := coll.RemoveAt(targetAddressIndex.Int64())
			t.addRangeFailure("unexpected address "+removedAddr.String()+" at collection address index", coll)
		}()
	}

	if collCount.Cmp(coll.GetCount()) != 0 {
		t.addRangeFailure("collection count unexpected after attempting to remove address", coll)
	}
}

func (rangeTest *rangeTest) testIntegerOps(list *ipaddr.IPAddressSeqRangeList) {
	if list.IsEmpty() {
		rangeTest.testIntegerOpsEmptyList(list)
		return
	}
	ranges := list.GetSeqRanges()
	count := bigZero()
	for i := 0; i < len(ranges); i++ {
		count.Add(count, ranges[i].GetCount())
	}

	originalCount := bigZero().Set(count)
	originalRangeCount := len(ranges)
	if originalRangeCount != list.GetSeqRangeCount() {
		rangeTest.addRangeFailure("inconsistent range counts", list)
	}

	targetIndex1 := bigZero().Set(count)
	targetIndex1.Div(targetIndex1, big.NewInt(2))
	targetIndex2 := bigZero().Set(targetIndex1)

	targetIndex1.Add(targetIndex1, big.NewInt(-7))
	targetIndex2.Add(targetIndex2, big.NewInt(7))

	targetRangeIndex1, targetRangeIndex2 := -1, -1

	count = bigZero()

	var (
		rng1, rng2 *ipaddr.IPAddressSeqRange

		rngTargetIndex1, rngTargetIndex2, rng1LowerIndex, rng2LowerIndex, rng1UpperIndex, rng2UpperIndex *big.Int

		target1, target2 *ipaddr.IPAddress

		skipRange1, skipRange2 bool
	)

	// this code is finding targetRangeIndex1 and targetRangeIndex2, the ranges, for the given target indices targetIndex1 and targetIndex2
	// It goes through the ranges until it finds the range that contains the target, then it saves the range, the target index within the range, the target address, and the lower and upper indices of that range.
	for i := 0; i < len(ranges) && (targetRangeIndex1 < 0 || targetRangeIndex2 < 0); i++ {
		rng := &ranges[i]
		previousCount := bigZero().Set(count)
		count.Add(count, rng.GetCount())
		if targetRangeIndex1 < 0 && targetIndex1.Cmp(count) < 0 {
			targetRangeIndex1 = i
			rng1 = rng
			rngTargetIndex1 = bigZero().Set(targetIndex1)
			rngTargetIndex1.Sub(rngTargetIndex1, previousCount)
			target1 = rng.GetLower().IncrementBig(rngTargetIndex1)
			if target1 == nil {
				skipRange1 = true
			} else {
				rng1LowerIndex = previousCount
				rng1UpperIndex = bigZero().Set(count)
				rng1UpperIndex.Sub(rng1UpperIndex, bigOneConst())
			}
		}
		if targetRangeIndex2 < 0 && targetIndex2.Cmp(count) < 0 {
			targetRangeIndex2 = i
			rng2 = rng
			rngTargetIndex2 = bigZero().Set(targetIndex2)
			rngTargetIndex2.Sub(rngTargetIndex2, previousCount)
			target2 = rng.GetLower().IncrementBig(rngTargetIndex2)
			if target2 == nil {
				skipRange1 = true
			} else {
				rng2LowerIndex = previousCount
				rng2UpperIndex = bigZero().Set(count)
				rng2UpperIndex.Sub(rng2UpperIndex, bigOneConst())
			}
		}
	}

	if !skipRange1 && targetRangeIndex1 >= 0 && targetIndex1.Sign() >= 0 {
		rangeTest.removeAndPutBack(list, originalCount, originalRangeCount, targetRangeIndex1, rng1, rngTargetIndex1, targetIndex1, target1)
	}
	if !skipRange2 && targetRangeIndex2 >= 0 && targetIndex2.Sign() >= 0 {
		rangeTest.removeAndPutBack(list, originalCount, originalRangeCount, targetRangeIndex2, rng2, rngTargetIndex2, targetIndex2, target2)
	}
	if targetRangeIndex1 >= 0 && !skipRange1 {
		if rng1.GetCount().Cmp(bigOneConst()) > 0 {
			lower := rng1.GetLower()
			if !lower.Equal(target1) {
				rangeTest.removeAndPutBack(list, originalCount, originalRangeCount, targetRangeIndex1, rng1, bigZeroConst(), rng1LowerIndex, lower)
			}
			upper := rng1.GetUpper()
			if !upper.Equal(target1) {
				count := rng1.GetCount()
				rangeTest.removeAndPutBack(list, originalCount, originalRangeCount, targetRangeIndex1, rng1, count.Sub(count, bigOneConst()), rng1UpperIndex, upper)
			}
		}
	}
	if targetRangeIndex2 >= 0 && !skipRange2 {
		if !rng1.Equal(rng2) && rng2.GetCount().Cmp(bigOneConst()) > 0 {
			lower := rng2.GetLower()
			if !lower.Equal(target2) {
				rangeTest.removeAndPutBack(list, originalCount, originalRangeCount, targetRangeIndex2, rng2, bigZeroConst(), rng2LowerIndex, lower)
			}
			upper := rng2.GetUpper()
			if !upper.Equal(target2) {
				count := rng2.GetCount()
				rangeTest.removeAndPutBack(list, originalCount, originalRangeCount, targetRangeIndex2, rng2, count.Sub(count, bigOneConst()), rng2UpperIndex, upper)
			}
		}
	}

	// Do the same with the lowest range
	rng := list.GetLowerSeqRange()
	if !rng.Equal(rng1) {
		rngTargetIndex := rng.GetCount()
		rngTargetIndex.Div(rngTargetIndex, big.NewInt(2))
		target := rng.GetLower().IncrementBig(rngTargetIndex)
		rangeTest.removeAndPutBack(list, originalCount, originalRangeCount, 0, rng, rngTargetIndex, rngTargetIndex, target)
		if rng.GetCount().Cmp(bigOneConst()) > 0 {
			lower := rng.GetLower()
			if !lower.Equal(target) {
				rangeTest.removeAndPutBack(list, originalCount, originalRangeCount, 0, rng, bigZeroConst(), bigZeroConst(), lower)
			}
			upper := rng.GetUpper()
			if !upper.Equal(target) {
				rngTargetIndex = rng.GetCount()
				rngTargetIndex.Sub(rngTargetIndex, bigOneConst())
				rangeTest.removeAndPutBack(list, originalCount, originalRangeCount, 0, rng, rngTargetIndex, rngTargetIndex, upper)
			}
		}
	}

	//Do the same with the highest range
	rng = list.GetUpperSeqRange()
	if !rng.Equal(rng2) {
		rngCount := rng.GetCount()
		var rngTargetIndex *big.Int
		if rngCount.Cmp(bigOneConst()) == 0 {
			rngTargetIndex = bigZeroConst()
		} else {
			rngTargetIndex = bigZero().Set(rngCount)
			rngTargetIndex.Add(rngTargetIndex, bigOneConst())
			rngTargetIndex.Div(rngTargetIndex, big.NewInt(2))
		}
		target := rng.GetLower().IncrementBig(rngTargetIndex)
		lastRangeAddressIndex := bigZero().Set(list.GetCount())
		lastRangeAddressIndex.Sub(lastRangeAddressIndex, rng.GetCount())
		added := bigZero().Set(lastRangeAddressIndex)
		added.Add(added, rngTargetIndex)
		rangeTest.removeAndPutBack(list, originalCount, originalRangeCount, list.GetSeqRangeCount()-1, rng, rngTargetIndex, added, rng.GetLower().IncrementBig(rngTargetIndex))
		if rng.GetCount().Cmp(bigOneConst()) > 0 {
			lower := rng.GetLower()
			if !lower.Equal(target) {
				rangeTest.removeAndPutBack(list, originalCount, originalRangeCount, list.GetSeqRangeCount()-1, rng, bigZeroConst(), lastRangeAddressIndex, lower)
			}
			upper := rng.GetUpper()
			if !upper.Equal(target) {
				lastRangeAddressUpperIndex := bigZero().Set(rng.GetCount())
				lastRangeAddressUpperIndex.Sub(lastRangeAddressUpperIndex, bigOneConst())
				rangeTest.removeAndPutBack(list, originalCount, originalRangeCount, list.GetSeqRangeCount()-1, rng, lastRangeAddressUpperIndex, lastRangeAddressIndex.Add(lastRangeAddressIndex, lastRangeAddressUpperIndex), upper)
			}
		}
	}

	lower := list.GetLower()
	val := lower.GetValue()
	if val.Sign() != 0 {
		zero, _ := lower.ToZeroHostLen(0)
		if val.Cmp(bigOneConst()) > 0 {
			halfVal := bigZero().Set(val)
			halfVal.Div(halfVal, big.NewInt(2))
			halfVal.Neg(halfVal)
			rangeTest.checkAboveBelow(list, originalCount, originalRangeCount, halfVal, lower.IncrementBig(halfVal))
			negVal := bigZero().Set(val)
			negVal.Neg(negVal)
			rangeTest.checkAboveBelow(list, originalCount, originalRangeCount, negVal, zero)
			negPlusOneVal := bigZero().Set(val)
			negPlusOneVal.Add(negPlusOneVal, bigOneConst())
			negPlusOneVal.Neg(negPlusOneVal)
			rangeTest.checkOutOfBounds(list, originalCount, originalRangeCount, negPlusOneVal)
		} else { // val is one
			rangeTest.checkAboveBelow(list, originalCount, originalRangeCount, big.NewInt(-1), zero)
			rangeTest.checkOutOfBounds(list, originalCount, originalRangeCount, big.NewInt(-2))
			rangeTest.checkOutOfBounds(list, originalCount, originalRangeCount, big.NewInt(-7))
		}
	} else { // val is 0.0.0.0 or ::
		rangeTest.checkOutOfBounds(list, originalCount, originalRangeCount, big.NewInt(-1))
		rangeTest.checkOutOfBounds(list, originalCount, originalRangeCount, big.NewInt(-7))
	}
	max, _ := lower.ToMaxHostLen(0)
	beyond := list.GetUpper().Enumerate(max)
	if !list.GetUpper().IsMax() {
		rangeTest.checkAboveBelow(list, originalCount, originalRangeCount, list.GetCount(), list.GetUpper().IncrementSingle())
		beyondCount := bigZero().Set(list.GetCount())
		beyondCount.Add(beyondCount, beyond)
		beyondCount.Sub(beyondCount, bigOneConst())
		rangeTest.checkAboveBelow(list, originalCount, originalRangeCount, beyondCount, max)
	}
	beyondCount := bigZero().Set(list.GetCount())
	beyondCount.Add(beyondCount, beyond)
	rangeTest.checkOutOfBounds(list, originalCount, originalRangeCount, beyondCount)
}

var counter int

func (rangeTest *rangeTest) removeAndPutBack(
	list *ipaddr.IPAddressSeqRangeList,
	originalCount *big.Int,
	originalRangeCount,
	targetRangeIndex int,
	targetRange *ipaddr.IPAddressSeqRange,
	targetAddressIndexInRange,
	targetAddressIndex *big.Int,
	target *ipaddr.IPAddress) {

	t := rangeTest.t

	isLongish := targetAddressIndex.Cmp(big.NewInt(math.MaxInt64)) <= 0 && targetAddressIndex.Cmp(big.NewInt(math.MinInt64)) >= 0

	isFirstInRange := targetAddressIndexInRange.Sign() == 0
	nextAddr := bigZero().Set(targetAddressIndexInRange)
	isLastInRange := nextAddr.Add(nextAddr, bigOneConst()).Cmp(targetRange.GetCount()) == 0
	isLastAddress := targetRangeIndex == list.GetSeqRangeCount()-1 && isLastInRange
	isFirstAddress := targetRangeIndex == 0 && isFirstInRange

	originalList := list.Clone()
	if list.GetCount().Cmp(originalCount) != 0 {
		t.addRangeFailure("inconsistent address counts", list)
	}
	rng := list.GetSeqRange(targetRangeIndex)
	if !rng.Equal(targetRange) {
		t.addRangeFailure("unexpected range at sequential range index", list)
	}
	rng = list.GetContainingSeqRangeBig(targetAddressIndex)
	if !rng.Equal(targetRange) {
		t.addRangeFailure("unexpected range at address index", list)
	}
	if isLongish {
		rng = list.GetContainingSeqRange(targetAddressIndex.Int64())
		if !rng.Equal(targetRange) {
			t.addRangeFailure("unexpected range at address index", list)
		}
	}
	getAddr := list.GetBig(targetAddressIndex)
	if getAddr.IsMultiple() || getAddr.IsPrefixed() {
		t.addRangeFailure("unexpected multiple or prefixed subnet", list)
	}
	if !getAddr.Equal(target) {
		t.addRangeFailure("unexpected address at list address index", list)
	}
	if isLongish {
		getAddr = list.Get(targetAddressIndex.Int64())
		if getAddr.IsMultiple() || getAddr.IsPrefixed() {
			t.addRangeFailure("unexpected multiple or prefixed subnet", list)
		}
		if !getAddr.Equal(target) {
			t.addRangeFailure("unexpected address at list address index", list)
		}
	}
	incrementAddr := list.IncrementBig(targetAddressIndex)
	if !getAddr.Equal(incrementAddr) {
		t.addRangeFailure("unexpected address at list increment address index", list)
	}
	if isLongish {
		incrementAddr = list.Increment(targetAddressIndex.Int64())
		if !getAddr.Equal(incrementAddr) {
			t.addRangeFailure("unexpected address at list increment address index", list)
		}
	}
	if list.Enumerate(getAddr).Cmp(targetAddressIndex) != 0 {
		t.addRangeFailure("unexpected enumerated address at list increment address index", list)
	}
	added := list.Add(getAddr)
	if added {
		t.addRangeFailure("unexpected add to list of address expected to be in list already", list)
	}
	if list.GetCount().Cmp(originalCount) != 0 {
		t.addRangeFailure("unexpected change in list count", list)
	}

	// remove it
	var removedAddr *ipaddr.IPAddress
	useLong := false
	if isLongish {
		counter++
		useLong = (counter%2 == 0)
	}
	if useLong {
		removedAddr = list.RemoveAt(targetAddressIndex.Int64())
	} else {
		removedAddr = list.RemoveAtBig(targetAddressIndex)
	}
	if removedAddr.IsPrefixed() {
		t.addRangeFailure("addresses in sequential ranges lists should not be prefixed", list)
	}
	if !removedAddr.Equal(target) {
		t.addRangeFailure("removed address "+removedAddr.String()+" not the expected address "+getAddr.String(), list)
	}
	if !removedAddr.Equal(getAddr) {
		t.addRangeFailure("removed address "+removedAddr.String()+" not the expected address "+getAddr.String()+" looked up", list)
	}
	// confirm it is gone
	func() {
		defer func() {
			if r := recover(); r != nil {
				if !isLastAddress {
					t.addRangeFailure("unexpected exception", list)
				}
			}
		}()
		getAddrAgain := list.GetBig(targetAddressIndex)
		if removedAddr.Equal(getAddrAgain) || isLastAddress {
			t.addRangeFailure("removed address not gone", list)
		}
	}()
	if isLongish {
		func() {
			defer func() {
				if r := recover(); r != nil {
					if !isLastAddress {
						t.addRangeFailure("unexpected exception", list)
					}
				}
			}()
			getAddrAgain := list.Get(targetAddressIndex.Int64())
			if removedAddr.Equal(getAddrAgain) || isLastAddress {
				t.addRangeFailure("removed address not gone", list)
			}
		}()
	}
	// confirm it is gone by removing it again
	removedAgain := list.Remove(removedAddr)
	if removedAgain {
		t.addRangeFailure("removed address not removed", list)
	}
	if !isLastAddress {
		// confirm it is still gone
		incrementAddrAgain := list.IncrementBig(targetAddressIndex)
		if removedAddr.Equal(incrementAddrAgain) {
			t.addRangeFailure("removed address "+removedAddr.String()+" not removed", list)
		}
		if isLongish {
			incrementAddrAgain = list.Increment(targetAddressIndex.Int64())
			if removedAddr.Equal(incrementAddrAgain) {
				t.addRangeFailure("removed address "+removedAddr.String()+" not removed", list)
			}
		}
	}
	// check counts
	countAdjustment := 0
	if !isFirstInRange && !isLastInRange {
		countAdjustment++
	} else if !targetRange.IsMultiple() {
		countAdjustment--
	}
	if originalRangeCount+countAdjustment != list.GetSeqRangeCount() {
		t.addRangeFailure("range count unexpected after removing address: original "+fmt.Sprint(originalRangeCount)+" vs "+fmt.Sprint(list.GetSeqRangeCount())+" and is multiple "+fmt.Sprint(targetRange.IsMultiple()), list)
	}
	if originalCount.Cmp(bigOneConst()) > 0 {
		if targetRange.IsMultiple() || targetRangeIndex < originalRangeCount-1 {
			rng = list.GetSeqRange(targetRangeIndex)
			if rng.Equal(targetRange) {
				t.addRangeFailure("range in list still matches after removing address", list)
			}
		}
	} else if list.GetSeqRangeCount() != 0 {
		t.addRangeFailure("ranges should be gone after removng the only address", list)
	}
	countPlusOne := list.GetCount()
	if originalCount.Cmp(countPlusOne.Add(countPlusOne, bigOneConst())) != 0 {
		t.addRangeFailure("range count unexpected after removing address", list)
	}
	// check enumerate
	isBorderAddress := isLastAddress || isFirstAddress
	enumerated := list.Enumerate(removedAddr)
	var fails bool
	if isBorderAddress && !list.IsEmpty() {
		fails = enumerated == nil
	} else {
		fails = enumerated != nil
	}
	if fails {
		t.addRangeFailure("removed address found in range list after being removed", list)
	}
	enumerated = list.Enumerate(getAddr)
	if isBorderAddress && !list.IsEmpty() {
		fails = enumerated == nil
	} else {
		fails = enumerated != nil
	}
	if fails {
		t.addRangeFailure("address found in range list after being removed", list)
	}

	// add it back
	added = list.Add(getAddr)
	if !added {
		t.addRangeFailure("address not added as expected after being removed", list)
	}

	// confirm it is back
	getAddrAgain := list.GetBig(targetAddressIndex)
	if !removedAddr.Equal(getAddrAgain) {
		t.addRangeFailure("address not found in list as expected after being added back", list)
	}
	if isLongish {
		getAddrAgain = list.Get(targetAddressIndex.Int64())
		if !removedAddr.Equal(getAddrAgain) {
			t.addRangeFailure("address not found in list as expected after being added back", list)
		}
	}
	incrementAddrAgain := list.GetBig(targetAddressIndex)
	if !removedAddr.Equal(incrementAddrAgain) {
		t.addRangeFailure("address added back not at expected location", list)
	}
	if isLongish {
		incrementAddrAgain = list.Get(targetAddressIndex.Int64())
		if !removedAddr.Equal(incrementAddrAgain) {
			t.addRangeFailure("address added back not at expected location", list)
		}
	}
	// check counts
	if originalRangeCount != list.GetSeqRangeCount() {
		t.addRangeFailure("sequential range count not restored to original", list)
	}
	if originalCount.Cmp(list.GetCount()) != 0 {
		t.addRangeFailure("address count not restored to original", list)
	}
	// check enumerate
	if list.Enumerate(removedAddr).Cmp(targetAddressIndex) != 0 {
		t.addRangeFailure("address not found at expeceted location", list)
	}
	if list.Enumerate(getAddr).Cmp(targetAddressIndex) != 0 {
		t.addRangeFailure("address not located at expeceted location", list)
	}
	// confirm the list is back to the same
	if !list.Equal(originalList) {
		t.addRangeFailure("restored list not equal to original", list)
	}
}

// checkAboveBelow checks negative indices and indices beyond the list size, checking increments below the list and beyond the upper value of the list
func (rangeTest *rangeTest) checkAboveBelow(
	list *ipaddr.IPAddressSeqRangeList,
	originalCount *big.Int,
	originalRangeCount int,
	targetAddressIndex *big.Int,
	target *ipaddr.IPAddress) {

	t := rangeTest.t

	isLongish := targetAddressIndex.Cmp(big.NewInt(math.MaxInt64)) <= 0 && targetAddressIndex.Cmp(big.NewInt(math.MinInt64)) >= 0

	if list.GetCount().Cmp(originalCount) != 0 {
		t.addRangeFailure("inconsistent address counts", list)
	}
	func() {
		defer func() {
			if r := recover(); r != nil {
				// no longer there
			}
		}()
		getAddr := list.GetBig(targetAddressIndex)
		t.addRangeFailure("unexpected address "+getAddr.String()+" at list address index", list)
	}()

	if isLongish {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// no longer there
				}
			}()
			getAddr := list.Get(targetAddressIndex.Int64())
			t.addRangeFailure("unexpected address "+getAddr.String()+" at list address index", list)
		}()
	}
	incrementAddr := list.IncrementBig(targetAddressIndex)
	if !incrementAddr.Equal(target) {
		t.addRangeFailure("unexpected address "+incrementAddr.String()+" at list increment address index, expected "+target.String(), list)
	}
	if isLongish {
		incrementAddr = list.Increment(targetAddressIndex.Int64())
		if !incrementAddr.Equal(target) {
			t.addRangeFailure("unexpected address "+incrementAddr.String()+" at list increment address index, expected "+target.String(), list)
		}
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				// no longer there
			}
		}()
		rng := list.GetContainingSeqRangeBig(targetAddressIndex)
		t.addRangeFailure("unexpected range "+rng.String()+" at address index", list)
	}()

	if isLongish {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// no longer there
				}
			}()
			rng := list.GetContainingSeqRange(targetAddressIndex.Int64())
			t.addRangeFailure("unexpected range "+rng.String()+"  at address index", list)
		}()
	}

	// remove it
	func() {
		defer func() {
			if r := recover(); r != nil {
				// no longer there
			}
		}()
		removedAddr := list.RemoveAtBig(targetAddressIndex)
		t.addRangeFailure("unexpected address "+removedAddr.String()+" at list address index", list)
	}()
	if isLongish {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// no longer there
				}
			}()
			removedAddr := list.RemoveAt(targetAddressIndex.Int64())
			t.addRangeFailure("unexpected address "+removedAddr.String()+" at list address index", list)
		}()
	}
	if originalRangeCount != list.GetSeqRangeCount() {
		t.addRangeFailure("range count unexpected after removing address not in list: original "+fmt.Sprint(originalRangeCount)+" vs "+fmt.Sprint(list.GetSeqRangeCount()), list)
	}
	if originalCount.Cmp(list.GetCount()) != 0 {
		t.addRangeFailure("range count unexpected after removing address", list)
	}
	// check enumerate
	if incrementAddr != nil {
		enumerated := list.Enumerate(incrementAddr)
		var fails bool
		if !list.IsEmpty() {
			fails = enumerated == nil
		} else {
			fails = enumerated != nil
		}
		if fails {
			t.addRangeFailure("enumerated address found in range list unexpectedly", list)
		}
		if !list.IsEmpty() {
			if targetAddressIndex.Cmp(enumerated) != 0 {
				t.addRangeFailure("enumerated address not the inverse of increment", list)
			}
		}
	}
}

func (rangeTest *rangeTest) checkOutOfBounds(
	list *ipaddr.IPAddressSeqRangeList,
	originalCount *big.Int,
	originalRangeCount int,
	targetAddressIndex *big.Int) {

	t := rangeTest.t

	isLongish := targetAddressIndex.Cmp(big.NewInt(math.MaxInt64)) <= 0 && targetAddressIndex.Cmp(big.NewInt(math.MinInt64)) >= 0

	if list.GetCount().Cmp(originalCount) != 0 {
		t.addRangeFailure("inconsistent address counts", list)
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				// pass
			}
		}()
		getAddr := list.GetBig(targetAddressIndex)
		t.addRangeFailure("unexpected address "+getAddr.String()+" at list address index", list)
	}()

	if isLongish {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// pass
				}
			}()
			getAddr := list.Get(targetAddressIndex.Int64())
			t.addRangeFailure("unexpected address "+getAddr.String()+" at list address index", list)
		}()

	}

	incrementAddr := list.IncrementBig(targetAddressIndex)
	if incrementAddr != nil {
		t.addRangeFailure("unexpected address "+incrementAddr.String()+" at list increment address index "+targetAddressIndex.String()+" which should throw", list)
	}

	if isLongish {
		incrementAddr := list.Increment(targetAddressIndex.Int64())
		if incrementAddr != nil {
			t.addRangeFailure("unexpected address "+incrementAddr.String()+" at list increment address index "+targetAddressIndex.String()+" which should throw", list)
		}
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				// out of bounds
			}
		}()
		rng := list.GetContainingSeqRangeBig(targetAddressIndex)
		t.addRangeFailure("unexpected range "+rng.String()+" at address index", list)
	}()

	if isLongish {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// out of bounds
				}
			}()
			rng := list.GetContainingSeqRange(targetAddressIndex.Int64())
			t.addRangeFailure("unexpected range "+rng.String()+"  at address index", list)
		}()
	}

	// remove it

	func() {
		defer func() {
			if r := recover(); r != nil {
				// pass
			}
		}()
		removedAddr := list.RemoveAtBig(targetAddressIndex)
		t.addRangeFailure("removed address "+removedAddr.String()+" not the expected address", list)
	}()

	if isLongish {
		func() {

			defer func() {
				if r := recover(); r != nil {
					// pass
				}
			}()
			removedAddr := list.RemoveAt(targetAddressIndex.Int64())
			t.addRangeFailure("removed address "+removedAddr.String()+" not the expected address", list)
		}()
	}

	if originalRangeCount != list.GetSeqRangeCount() {
		t.addRangeFailure("range count unexpected after removing address not in list: original "+fmt.Sprint(originalRangeCount)+" vs "+fmt.Sprint(list.GetSeqRangeCount()), list)
	}
	if originalCount.Cmp(list.GetCount()) != 0 {
		t.addRangeFailure("range count unexpected after removing address", list)
	}
}

func (rangeTest *rangeTest) testIntegerOpsEmptyList(list *ipaddr.IPAddressSeqRangeList) {
	rangeTest.testIntegerOpsEmptyListIndex(list, bigZeroConst())
	rangeTest.testIntegerOpsEmptyListIndex(list, bigOneConst())
	rangeTest.testIntegerOpsEmptyListIndex(list, big.NewInt(-1))
}

func (rangeTest *rangeTest) testIntegerOpsEmptyListIndex(list *ipaddr.IPAddressSeqRangeList, targetAddressIndex *big.Int) {
	t := rangeTest.t
	if list.GetSeqRangeCount() != 0 {
		t.addRangeFailure("unexpected range count empty list", list)
	}

	if list.GetCount().Sign() != 0 {
		t.addRangeFailure("unexpected count empty list", list)
	}

	isLongish := targetAddressIndex.Cmp(big.NewInt(math.MaxInt64)) <= 0 && targetAddressIndex.Cmp(big.NewInt(math.MinInt64)) >= 0

	func() {
		defer func() {
			if r := recover(); r != nil {
				// pass
			}
		}()
		getAddr := list.GetBig(targetAddressIndex)
		rangeTest.addRangeFailure("unexpected address "+getAddr.String()+" at list address index", list)
	}()
	if isLongish {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// pass
				}
			}()
			getAddr := list.Get(targetAddressIndex.Int64())
			t.addRangeFailure("unexpected address "+getAddr.String()+" at list address index", list)
		}()
	}

	incrementAddr := list.IncrementBig(targetAddressIndex)
	if incrementAddr != nil {
		t.addRangeFailure("unexpected address "+incrementAddr.String()+" at list increment address index "+targetAddressIndex.String()+" which should throw", list)
	}
	if isLongish {
		incrementAddr = list.Increment(targetAddressIndex.Int64())
		if incrementAddr != nil {
			t.addRangeFailure("unexpected address "+incrementAddr.String()+" at list increment address index "+targetAddressIndex.String()+" which should throw", list)
		}
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				// pass
			}
		}()
		rng := list.GetContainingSeqRangeBig(targetAddressIndex)
		t.addRangeFailure("unexpected range "+rng.String()+" at address index", list)
	}()
	if isLongish {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// pass
				}
			}()
			rng := list.GetContainingSeqRange(targetAddressIndex.Int64())
			t.addRangeFailure("unexpected range "+rng.String()+"  at address index", list)
		}()
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				// pass
			}
		}()
		removedAddr := list.RemoveAtBig(targetAddressIndex)
		t.addRangeFailure("removed address "+removedAddr.String()+" not the expected address", list)
	}()
	if isLongish {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// pass
				}
			}()
			removedAddr := list.RemoveAt(targetAddressIndex.Int64())
			t.addRangeFailure("removed address "+removedAddr.String()+" not the expected address", list)
		}()
	}
}

func (rangeTest *rangeTest) incrementTestCount() {
	rangeTest.t.rangeListTestCount++
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
//
//
//
// end of rangeTest
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
//
//
//

func (t *collectionTester) testIterateRangeList(rangeList *ipaddr.IPAddressSeqRangeList) {
	t.testIterate(rangeList,
		ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].Iterator,
		nil,
		ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].GetCount,
		func(coll ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress]) ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress] {
			return coll.(*ipaddr.IPAddressSeqRangeList).Clone()
		},
		ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].GetLower,
		ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].Add,
		ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].Remove,
		ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].ContainsAddress,
		func(coll ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress], addr *ipaddr.IPAddress) int {
			return coll.(*ipaddr.IPAddressSeqRangeList).IndexOfSeqRangeContainingAddress(addr)
		},
		ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].IsEmpty)
}

func (t *collectionTester) testIterate(
	collection ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress],
	iteratorFunc func(ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress]) ipaddr.Iterator[*ipaddr.IPAddress],
	iteratorWithRemoveFunc func(ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress]) ipaddr.IteratorWithRemove[*ipaddr.IPAddress],
	collectionSizeFunc func(ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress]) *big.Int,
	cloneFunc func(ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress]) ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress],
	firstAddressFunc func(ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress]) *ipaddr.IPAddress,
	addFunc,
	removeFunc,
	containsFunc func(ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress], *ipaddr.IPAddress) bool,
	indexOfRangeFunc func(ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress], *ipaddr.IPAddress) int,
	isEmptyFunc func(ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress]) bool) {

	// iterate the list, confirm the size by counting
	// clone the list, iterate again, but remove each time, confirm the size
	// confirm list is empty at the end
	totalSize := collectionSizeFunc(collection)
	if totalSize.Sign() > 0 {
		clonedRangeList := cloneFunc(collection)
		addr := firstAddressFunc(clonedRangeList)
		toAdd := addr
		removeFunc(clonedRangeList, toAdd)
		modIterator := iteratorFunc(clonedRangeList)
		mod := collectionSizeFunc(clonedRangeList).Uint64() / 2
		if totalSize.Cmp(big.NewInt(1000)) > 0 {
			mod = 11
		}
		var i uint64
		shouldThrow := false

		// https://go.dev/play/p/j2MXBj1_75S

		func() {
			defer func() {
				if r := recover(); r != nil {
					if !shouldThrow {
						t.addRangeFailure("unexpected throw ", clonedRangeList)
					}
				}
			}()
			for modIterator.HasNext() {
				i++
				if i == mod {
					shouldThrow = true
					addFunc(clonedRangeList, toAdd)
				}
				modIterator.Next()
				if shouldThrow {
					t.addRangeFailure("expected throw ", clonedRangeList)
					shouldThrow = false
					break
				}
			}
		}()
	}
	removeAllowed := iteratorWithRemoveFunc != nil

	if totalSize.Cmp(big.NewInt(1000)) < 0 {
		for i := 0; i < 3; i++ {
			firstTime := i == 0
			secondTime := i == 1

			expectedSize := collectionSizeFunc(collection).Uint64()
			var actualSize int
			var iteratorWithRemove ipaddr.IteratorWithRemove[*ipaddr.IPAddress]
			var iterator ipaddr.Iterator[*ipaddr.IPAddress]
			if removeAllowed {
				iteratorWithRemove = iteratorWithRemoveFunc((collection))
				iterator = iteratorWithRemove
			} else {
				iterator = iteratorFunc(collection)
			}

			j := 0
			for iterator.HasNext() {
				j++

				next := iterator.Next()
				actualSize++

				if firstTime || (secondTime && ((j % 3) != 1)) {
					if !containsFunc(collection, next) {
						t.addRangeFailure("after iteration "+next.String()+" not in list ", collection)
						containsFunc(collection, next)
					} else if indexOfRangeFunc != nil && indexOfRangeFunc(collection, next) < 0 {
						t.addRangeFailure("after iteration address "+next.String()+" not in list ", collection)
					}
				} else if removeAllowed {
					iteratorWithRemove.Remove()
					if containsFunc(collection, next) {
						t.addRangeFailure("after removal "+next.String()+" still in trie ", collection)
					}
				}

			}
			if uint64(actualSize) != expectedSize {
				t.addRangeFailure("count was "+fmt.Sprint(actualSize)+" instead of expected "+fmt.Sprint(expectedSize), collection)
				fmt.Println("count was "+fmt.Sprint(actualSize)+" instead of expected "+fmt.Sprint(expectedSize), collection)
			}
			if i == 2 {
				break
			}
			collection = cloneFunc(collection)
		}
		if removeAllowed {
			if !isEmptyFunc(collection) {
				t.addRangeFailure("list not empty, size "+collectionSizeFunc(collection).String()+" after removing everything", collection)
			} else if collectionSizeFunc(collection).Sign() > 0 {
				t.addRangeFailure("range list size not 0, "+collectionSizeFunc(collection).String()+" after removing everything", collection)
			}
		}
	}
	t.incrementTestCount()
}

func (t *collectionTester) testIterateTrie(trie *ipaddr.IPAddressContainmentTrie) {
	t.testIterate(trie,
		ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].Iterator,
		nil,
		ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].GetCount,
		func(coll ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress]) ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress] {
			return coll.(*ipaddr.IPAddressContainmentTrie).Clone()
		},
		ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].GetLower,
		ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].Add,
		ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].Remove,
		ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].ContainsAddress,
		nil,
		ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].IsEmpty)
}

func newSingleRangeResult(
	t *collectionTester,
	isIPv6 bool,
	list *ipaddr.IPAddressSeqRangeList,
	rng *ipaddr.IPAddressSeqRange,
	intersection,
	union,
	remove *ipaddr.IPAddressSeqRangeList) *singleRangeTest {
	return &singleRangeTest{
		isIPv6:               isIPv6,
		list:                 list,
		rng:                  rng,
		expectedIntersection: intersection,
		expectedUnion:        union,
		expectedRemove:       remove,
		t:                    t,
	}
}

type singleRangeTest struct {
	isIPv6 bool

	list, expectedIntersection, expectedUnion, expectedRemove *ipaddr.IPAddressSeqRangeList

	rng *ipaddr.IPAddressSeqRange

	t *collectionTester
}

func convertListToTrie(list *ipaddr.IPAddressSeqRangeList) *ipaddr.IPAddressContainmentTrie {
	trie := &ipaddr.IPAddressContainmentTrie{}
	iter := list.SeqRangeIterator()
	for iter.HasNext() {
		trie.AddSeqRange(iter.Next())
	}
	return trie
}

func convertRangeToTrie(rng *ipaddr.IPAddressSeqRange) *ipaddr.IPAddressContainmentTrie {
	trie := &ipaddr.IPAddressContainmentTrie{}
	trie.AddSeqRange(rng)
	return trie
}

func convertAddrToTrie(addr *ipaddr.IPAddress) *ipaddr.IPAddressContainmentTrie {
	trie := &ipaddr.IPAddressContainmentTrie{}
	trie.Add(addr)
	return trie
}

func (r *singleRangeTest) runTest() {

	t := r.t

	listTrie, rngTrie := convertListToTrie(r.list), convertRangeToTrie(r.rng)

	t.singleListTestCount++

	expectedUnionTrie := convertListToTrie(r.expectedUnion)

	testCollectionOpSingleRange[*ipaddr.IPAddressSeqRangeList](r, r.list, r.rng, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].AddSeqRange, "add", r.expectedUnion)        // collection test
	testCollectionOpSingleRange[*ipaddr.IPAddressContainmentTrie](r, listTrie, r.rng, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].AddSeqRange, "add", expectedUnionTrie) // collection test

	r.testOpSingleRange(r.list, r.rng, (*ipaddr.IPAddressSeqRangeList).IntersectSeqRange, "intersect", r.expectedIntersection)

	expectedOverlap := !r.expectedIntersection.IsEmpty()

	r.testCollectionBooleanOpSingleRange(r.list, r.rng, listTrie, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].OverlapsSeqRange, "overlaps", expectedOverlap) // collection test

	expectedRemoveTrie := convertListToTrie(r.expectedRemove)

	testCollectionOpSingleRange[*ipaddr.IPAddressSeqRangeList](r, r.list, r.rng, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].RemoveSeqRange, "remove", r.expectedRemove)        // collection test
	testCollectionOpSingleRange[*ipaddr.IPAddressContainmentTrie](r, listTrie, r.rng, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].RemoveSeqRange, "remove", expectedRemoveTrie) // collection test

	// A contains (A intersect B)

	intersection := r.binaryOpSingleRange(r.list, r.rng, (*ipaddr.IPAddressSeqRangeList).IntersectSeqRange, "intersect")
	t.contains(r.list, intersection, true)

	// (A intersect B) contains A iff B contains A

	r.containsRange(intersection, r.rng, r.list.ContainsRange(r.rng))

	union := r.binaryOpSingleRange(r.list, r.rng, (*ipaddr.IPAddressSeqRangeList).AddSeqRange, "add")
	t.contains(union, r.list, true)

	r.containsRange(union, r.rng, true)

	t.contains(r.list, union, r.list.ContainsRange(r.rng))

	unionTrie := r.collectionBinaryOpSingleRange(listTrie, r.rng, (*ipaddr.IPAddressContainmentTrie).AddSeqRange, "add")

	t.collectionsMatch(union, unionTrie)
	r.containsRange(unionTrie, r.rng, true) // collection test

	empty := ipaddr.IPAddressSeqRangeList{}
	emptyTrie := ipaddr.IPAddressContainmentTrie{}

	// removal of oneself always results in nothing
	testCollectionOpSingleRange[*ipaddr.IPAddressSeqRangeList](r, inList(r.rng), r.rng, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].RemoveSeqRange, "remove", &empty)  // collection test
	testCollectionOpSingleRange[*ipaddr.IPAddressContainmentTrie](r, rngTrie, r.rng, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].RemoveSeqRange, "remove", &emptyTrie) // collection test

	// intersection with oneself results in the same
	r.testOpSingleRange(inList(r.rng), r.rng, (*ipaddr.IPAddressSeqRangeList).IntersectSeqRange, "intersect", inList(r.rng))

	// union with oneself results in the same
	testCollectionOpSingleRange[*ipaddr.IPAddressSeqRangeList](r, inList(r.rng), r.rng, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].AddSeqRange, "add", inList(r.rng)) // collection test
	testCollectionOpSingleRange[*ipaddr.IPAddressContainmentTrie](r, rngTrie, r.rng, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].AddSeqRange, "add", inTrie(r.rng))    // collection test

	// removing one from the other, removing the other from the one, then taking the union, is the same as removing the intersection from the union
	t.matches(
		t.binaryOp(r.binaryOpSingleRange(r.list, r.rng, (*ipaddr.IPAddressSeqRangeList).RemoveSeqRange, "remove"),
			t.binaryOp(inList(r.rng), r.list, (*ipaddr.IPAddressSeqRangeList).RemoveIntoNew, "remove"),
			(*ipaddr.IPAddressSeqRangeList).JoinIntoNew, "join"),
		t.binaryOp(r.binaryOpSingleRange(r.list, r.rng, (*ipaddr.IPAddressSeqRangeList).AddSeqRange, "add"),
			r.binaryOpSingleRange(r.list, r.rng, (*ipaddr.IPAddressSeqRangeList).IntersectSeqRange, "intersect"),
			(*ipaddr.IPAddressSeqRangeList).RemoveIntoNew, "remove"))

	everythingStr := "0.0.0.0/0"
	if r.isIPv6 {
		everythingStr = "::/0"
	}
	everythingAddr := ipaddr.NewIPAddressString(everythingStr).GetAddress().ToPrefixBlock()

	// De Morgan's Law 1
	// complement of the union is the same as the intersection of the complements
	t.matches(t.complementWrapper(
		r.binaryOpSingleRange(r.list, r.rng, (*ipaddr.IPAddressSeqRangeList).AddSeqRange, "add"), everythingAddr),
		t.binaryOp(t.complementWrapper(r.list, everythingAddr),
			t.complementWrapper(inList(r.rng), everythingAddr), (*ipaddr.IPAddressSeqRangeList).IntersectIntoNew, "intersect"))

	// De Morgan's Law 2
	// complement of the intersection is the same as the union of the complements
	t.matches(
		t.complementWrapper(r.binaryOpSingleRange(r.list, r.rng, (*ipaddr.IPAddressSeqRangeList).IntersectSeqRange, "intersect"), everythingAddr),
		t.binaryOp(t.complementWrapper(r.list, everythingAddr),
			t.unaryOp(inList(r.rng), (*ipaddr.IPAddressSeqRangeList).ComplementIntoNew, "complement"), (*ipaddr.IPAddressSeqRangeList).JoinIntoNew, "join"))

	r.removeIntersection(r.list, r.rng)

	r.unionDoubleIntersect(r.list, r.rng, everythingAddr)

	r.doubleComplement(r.rng, everythingAddr)

	r.intersectUnion(r.list, r.rng)

	r.unionIntersect(r.list, r.rng)

	r.removeIntersectComplement(r.list, r.rng)

	r.everythingNothing(r.rng, everythingAddr)

	r.testRangeListAndRangeSpans(r.expectedUnion, r.list, r.rng, listTrie) // collection test

	t.testIterateRangeList(inList(r.rng))

	t.testIterateTrie(listTrie) // collection test

	t.testEdges(r.list, listTrie, r.isIPv6)

	t.testCoverRange(r.list, listTrie, r.rng)
}

func (r *singleRangeTest) testRangeListAndRangeSpans(list, joined1 *ipaddr.IPAddressSeqRangeList, joined2 *ipaddr.IPAddressSeqRange, joined1Trie *ipaddr.IPAddressContainmentTrie) {
	resultPrefixBlocks := list.SpanWithPrefixBlocks()
	resultSequentialBlocks := list.SpanWithSequentialBlocks()
	prefixBlocks1 := joined1.SpanWithPrefixBlocks()
	prefixBlocks2 := joined2.SpanWithPrefixBlocks()
	seqBlocks1 := joined1.SpanWithSequentialBlocks()
	seqBlocks2 := joined2.SpanWithSequentialBlocks()

	t := r.t
	t.testPrefixBlockSpan(resultPrefixBlocks, prefixBlocks1, prefixBlocks2, list)
	t.testPrefixBlockSpan(resultPrefixBlocks, seqBlocks1, seqBlocks2, list)
	t.testSequentialBlockSpan(resultSequentialBlocks, seqBlocks1, seqBlocks2, list)
	t.testSequentialBlockSpan(resultSequentialBlocks, prefixBlocks1, prefixBlocks2, list)

	t.compareSpanningBlocks(prefixBlocks1, joined1Trie.PrefixBlockIterator(), joined1Trie.GetPrefixBlockCount(), list)
}

// removing the intersection is the same as removing the original
func (r *singleRangeTest) removeIntersection(list *ipaddr.IPAddressSeqRangeList, rng *ipaddr.IPAddressSeqRange) {
	t := r.t
	t.matches(t.binaryOp(list, r.binaryOpSingleRange(list, rng, (*ipaddr.IPAddressSeqRangeList).IntersectSeqRange, "intersect"), (*ipaddr.IPAddressSeqRangeList).RemoveIntoNew, "remove"),
		r.binaryOpSingleRange(list, rng, (*ipaddr.IPAddressSeqRangeList).RemoveSeqRange, "remove"))
}

// double complement results in the original
func (r *singleRangeTest) doubleComplement(rng *ipaddr.IPAddressSeqRange, everythingAddr *ipaddr.IPAddress) {
	t := r.t
	complement := rng.Complement()
	count := bigZero()
	for _, r := range complement {
		count.Add(count, r.GetCount())
	}
	t.matchesCount(everythingAddr.GetCount(), count.Add(count, rng.GetCount()), r.list)
	listComplement := inList(complement...)
	t.matches(
		inList(rng),
		complementWrapper(listComplement, everythingAddr))

	listCompl := rng.ComplementIntoList()
	listComplCount := listCompl.GetCount()
	t.matchesCount(everythingAddr.GetCount(), listComplCount.Add(listComplCount, rng.GetCount()), r.list)
	t.matches(
		inList(rng),
		complementWrapper(listComplement, everythingAddr))
}

// A intersect (A union B) = A
func (r *singleRangeTest) intersectUnion(list *ipaddr.IPAddressSeqRangeList, rng *ipaddr.IPAddressSeqRange) {
	t := r.t
	t.matches(
		list,
		t.binaryOp(list, r.binaryOpSingleRange(list, rng, (*ipaddr.IPAddressSeqRangeList).AddSeqRange, "add"),
			(*ipaddr.IPAddressSeqRangeList).IntersectIntoNew, "intersect"))
}

// A union (A intersect B) = A
func (r *singleRangeTest) unionIntersect(list *ipaddr.IPAddressSeqRangeList, rng *ipaddr.IPAddressSeqRange) {
	t := r.t
	t.matches(
		list,
		t.binaryOp(list, r.binaryOpSingleRange(list, rng, (*ipaddr.IPAddressSeqRangeList).IntersectSeqRange, "intersect"),
			(*ipaddr.IPAddressSeqRangeList).JoinIntoNew, "join"))
}

// intersect with complement and then with original, take union, should be the original
func (r *singleRangeTest) unionDoubleIntersect(list *ipaddr.IPAddressSeqRangeList, rng *ipaddr.IPAddressSeqRange, everythingAddr *ipaddr.IPAddress) {
	t := r.t
	t.matches(
		t.binaryOp(
			r.binaryOpSingleRange(list, rng, (*ipaddr.IPAddressSeqRangeList).IntersectSeqRange, "intersect"),
			r.binaryOpSingleRange(
				t.complementWrapper(list, everythingAddr),
				rng,
				(*ipaddr.IPAddressSeqRangeList).IntersectSeqRange, "intersect"),
			(*ipaddr.IPAddressSeqRangeList).JoinIntoNew, "join"),
		inList(rng))
}

// A remove B = A intersect (B complement)
func (r *singleRangeTest) removeIntersectComplement(list *ipaddr.IPAddressSeqRangeList, rng *ipaddr.IPAddressSeqRange) {
	t := r.t
	t.matches(
		r.binaryOpSingleRange(list, rng, (*ipaddr.IPAddressSeqRangeList).RemoveSeqRange, "remove"),
		t.binaryOp(list, inList(rng.Complement()...), (*ipaddr.IPAddressSeqRangeList).IntersectIntoNew, "intersect"))
}

// A Union (A complement) = everything
// A intersect (A complement) = nothing
func (r *singleRangeTest) everythingNothing(rng *ipaddr.IPAddressSeqRange, everythingAddr *ipaddr.IPAddress) {
	t := r.t
	complement := rng.Complement()
	nothing := &ipaddr.IPAddressSeqRangeList{}
	t.matches(
		nothing,
		r.binaryOpSingleRange(inList(complement...), rng, (*ipaddr.IPAddressSeqRangeList).IntersectSeqRange, "intersect"))

	nothing.Add(everythingAddr)
	var everything *ipaddr.IPAddressSeqRangeList = nothing
	t.matches(
		everything,
		r.binaryOpSingleRange(inList(complement...), rng, (*ipaddr.IPAddressSeqRangeList).AddSeqRange, "add"))
}

func (r *singleRangeTest) containsRange(containing ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress], contained *ipaddr.IPAddressSeqRange, expected bool) {
	t := r.t
	t.rangeListTestCount++
	if containing.ContainsRange(contained) == expected {
		if expected && containing.GetCount().Cmp(contained.GetCount()) < 0 {
			t.addRangeFailure("failed count for containment for list: "+containing.String()+" and single range: "+contained.String()+" expected containment: "+fmt.Sprint(expected)+" count of containing "+containing.GetCount().String()+" count of contained "+contained.GetCount().String(), r.list)
		} else {
			if printPass {
				fmt.Println("pass")
			}
		}
	} else {
		t.addRangeFailure("fail contains for collection: "+containing.String()+" and contained address: "+contained.String()+" expected containment: "+fmt.Sprint(expected), r.list)
	}
}

func (r *singleRangeTest) binaryOpSingleRange(list *ipaddr.IPAddressSeqRangeList, list2 *ipaddr.IPAddressSeqRange, op func(*ipaddr.IPAddressSeqRangeList, *ipaddr.IPAddressSeqRange) bool, opName string) *ipaddr.IPAddressSeqRangeList {
	t := r.t
	res := list.Clone()
	val := op(res, list2)
	if list.Equal(res) == val { // val true changed, list equal res then false, then false != true is true
		t.addRangeFailure("failed return value for "+opName+" for list: "+list.String()+" and single range: "+list2.String()+" expected same: "+fmt.Sprint(val)+" original: "+list.String()+" result: "+res.String(), list)
	} else {
		if printPass {
			fmt.Println("pass " + opName)
		}
	}
	if print {
		fmt.Println(list)
		fmt.Println(list2)
		fmt.Println(opName)
		fmt.Println(res)
		fmt.Println(val)
		fmt.Println()
	}
	return res
}

func (r *singleRangeTest) collectionBinaryOpSingleRange(coll *ipaddr.IPAddressContainmentTrie, list2 *ipaddr.IPAddressSeqRange, op func(*ipaddr.IPAddressContainmentTrie, *ipaddr.IPAddressSeqRange) bool, opName string) ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress] {
	t := r.t
	res := coll.Clone()
	val := op(res, list2)
	if coll.Equal(res) == val { // val true changed, list equal res then false, then false != true is true
		t.addRangeFailure("failed return value for "+opName+" for collection: "+coll.String()+" and single range: "+list2.String()+" expected same: "+fmt.Sprint(val)+" original: "+coll.String()+" result: "+res.String(), r.list)
	} else {
		if printPass {
			fmt.Println("pass " + opName)
		}
	}
	if print {
		fmt.Println(coll)
		fmt.Println(list2)
		fmt.Println(opName)
		fmt.Println(res)
		fmt.Println(val)
		fmt.Println()
	}
	return res
}

func testCollectionOpSingleRange[T ipaddr.IPAddressCollConstraint[T, *ipaddr.IPAddress]](
	singleRangeTest *singleRangeTest,
	coll ipaddr.IPAddressCollConstraint[T, *ipaddr.IPAddress],
	rng *ipaddr.IPAddressSeqRange,
	op func(ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress], *ipaddr.IPAddressSeqRange) bool,
	opName string,
	expected T) {

	t := singleRangeTest.t
	t.rangeListTestCount++
	res := coll.Clone()
	val := op(res, rng)
	if res.Equal(expected) {
		if res.GetCount().Cmp(expected.GetCount()) != 0 /*|| res.GetSeqRangeCount() != expected.GetSeqRangeCount()*/ {
			t.addRangeFailure("failed count for "+opName+" for collection: "+coll.String()+" and single range: "+rng.String()+" expected: "+expected.String()+" result: "+res.String(), coll)
		} else if coll.Equal(res) == val { // val true changed, list equal res then false, then false != true is true
			t.addRangeFailure("failed return value for "+opName+" for collection: "+coll.String()+" and single range: "+rng.String()+" expected same: "+fmt.Sprint(val)+" result: "+res.String(), coll)
		} else {
			if printPass {
				fmt.Println("pass " + opName)
			}
		}
	} else {
		t.addRangeFailure("fail "+opName+" for list: "+coll.String()+" and single range: "+rng.String()+" expected: "+expected.String()+" actual: "+res.String(), coll)
	}
}

func (r *singleRangeTest) testOpSingleRange(list *ipaddr.IPAddressSeqRangeList, rng *ipaddr.IPAddressSeqRange, op func(*ipaddr.IPAddressSeqRangeList, *ipaddr.IPAddressSeqRange) bool, opName string, expected *ipaddr.IPAddressSeqRangeList) {
	t := r.t
	t.rangeListTestCount++
	res := list.Clone()
	val := op(res, rng)
	if res.Equal(expected) {
		if res.GetCount().Cmp(expected.GetCount()) != 0 || res.GetSeqRangeCount() != expected.GetSeqRangeCount() {
			t.addRangeFailure("failed count for "+opName+" for list: "+list.String()+" and single range: "+rng.String()+" expected: "+expected.String()+" result: "+res.String(), list)
		} else if list.Equal(res) == val { // val true changed, list equal res then false, then false != true is true
			t.addRangeFailure("failed return value for "+opName+" for list: "+list.String()+" and single range: "+rng.String()+" expected same: "+fmt.Sprint(val)+" original: "+list.String()+" result: "+res.String(), list)
		} else {
			if printPass {
				fmt.Println("pass " + opName)
			}
		}
	} else {
		t.addRangeFailure("fail "+opName+" for list: "+list.String()+" and single range: "+rng.String()+" expected: "+expected.String()+" actual: "+res.String(), list)
	}
}

func (r *singleRangeTest) testCollectionBooleanOpSingleRange(list1 *ipaddr.IPAddressSeqRangeList, list2 *ipaddr.IPAddressSeqRange, containmentTrie *ipaddr.ContainmentTrieBase[*ipaddr.IPAddress], op func(ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress], *ipaddr.IPAddressSeqRange) bool, opName string, expected bool) {
	t := r.t
	t.rangeListTestCount++
	res := op(list1, list2)
	if res == expected {
		if printPass {
			fmt.Println("pass " + opName)
		}
	} else {
		t.addRangeFailure("fail "+opName+" for list: "+list1.String()+" and single range: "+list2.String()+" expected: "+fmt.Sprint(expected)+" actual: "+fmt.Sprint(res), list1)
	}

	t.rangeListTestCount++
	res = op(containmentTrie, list2)
	if res == expected {
		if printPass {
			fmt.Println("pass " + opName)
		}
	} else {
		t.addRangeFailure("fail "+opName+" for trie: "+containmentTrie.String()+" and single range: "+list2.String()+" expected: "+fmt.Sprint(expected)+" actual: "+fmt.Sprint(res), containmentTrie)
	}
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
//
//
//
// end of singleRangeTest
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
//
//
//

func (t *collectionTester) testCoverAddr(list *ipaddr.IPAddressSeqRangeList, listTrie *ipaddr.IPAddressContainmentTrie, addr *ipaddr.IPAddress) {
	t.testCover(list, listTrie)
	t.testCover(addr.IntoSequentialRangeList(), convertAddrToTrie(addr))
}

func (t *collectionTester) testCoverRange(list *ipaddr.IPAddressSeqRangeList, listTrie *ipaddr.IPAddressContainmentTrie, rng *ipaddr.IPAddressSeqRange) {
	t.testCover(list, listTrie)
	t.testCover(rng.IntoSequentialRangeList(), convertRangeToTrie(rng))
}

func (t *collectionTester) testCover(list *ipaddr.IPAddressSeqRangeList, listTrie *ipaddr.IPAddressContainmentTrie) {
	if !list.CoverWithPrefixBlock().Equal(listTrie.CoverWithPrefixBlock()) {
		t.addRangeFailure("cover with prefix block mismatch, "+list.CoverWithPrefixBlock().String()+" and "+listTrie.CoverWithPrefixBlock().String(), list)
	} else if !list.CoverWithSequentialRange().Equal(listTrie.CoverWithSequentialRange()) {
		t.addRangeFailure("cover with sequential range mismatch, "+list.CoverWithSequentialRange().String()+" and "+listTrie.CoverWithSequentialRange().String(), list)
	} else if list.IsSequential() != listTrie.IsSequential() {
		t.addRangeFailure("isSequential mismatch, "+fmt.Sprint(list.IsSequential())+" and "+fmt.Sprint(listTrie.IsSequential()), list)
	}
	t.rangeListTestCount++
}

func (t *collectionTester) testEdges(list *ipaddr.IPAddressSeqRangeList, listTrie *ipaddr.IPAddressContainmentTrie, isIPv6 bool) {
	rangeCount := list.GetSeqRangeCount()
	if rangeCount == 0 {
		everythingStr := "0.0.0.0/0"
		if isIPv6 {
			everythingStr = "::/0"
		}
		everythingAddr := ipaddr.NewIPAddressString(everythingStr).GetAddress()
		t.matchesWithColl(list.Floor(everythingAddr), nil, list)
		t.matchesWithColl(list.Ceiling(everythingAddr), nil, list)
		t.matchesWithColl(list.Lower(everythingAddr), nil, list)
		t.matchesWithColl(list.Higher(everythingAddr), nil, list)

		t.matchesWithColl(listTrie.Floor(everythingAddr), nil, listTrie)
		t.matchesWithColl(listTrie.Ceiling(everythingAddr), nil, listTrie)
		t.matchesWithColl(listTrie.Lower(everythingAddr), nil, listTrie)
		t.matchesWithColl(listTrie.Higher(everythingAddr), nil, listTrie)

		zeroStr := "0.0.0.0"
		if isIPv6 {
			zeroStr = "::"
		}

		zeroAddr := ipaddr.NewIPAddressString(zeroStr).GetAddress()
		t.matchesWithColl(list.Floor(zeroAddr), nil, list)
		t.matchesWithColl(list.Ceiling(zeroAddr), nil, list)
		t.matchesWithColl(list.Lower(zeroAddr), nil, list)
		t.matchesWithColl(list.Higher(zeroAddr), nil, list)

		t.matchesWithColl(listTrie.Floor(zeroAddr), nil, listTrie)
		t.matchesWithColl(listTrie.Ceiling(zeroAddr), nil, listTrie)
		t.matchesWithColl(listTrie.Lower(zeroAddr), nil, listTrie)
		t.matchesWithColl(listTrie.Higher(zeroAddr), nil, listTrie)

		maxStr := "255.255.255.255"
		if isIPv6 {
			maxStr = "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"
		}

		maxAddr := ipaddr.NewIPAddressString(maxStr).GetAddress()
		t.matchesWithColl(list.Floor(maxAddr), nil, list)
		t.matchesWithColl(list.Ceiling(maxAddr), nil, list)
		t.matchesWithColl(list.Lower(maxAddr), nil, list)
		t.matchesWithColl(list.Higher(maxAddr), nil, list)

		t.matchesWithColl(listTrie.Floor(maxAddr), nil, listTrie)
		t.matchesWithColl(listTrie.Ceiling(maxAddr), nil, listTrie)
		t.matchesWithColl(listTrie.Lower(maxAddr), nil, listTrie)
		t.matchesWithColl(listTrie.Higher(maxAddr), nil, listTrie)

		str := "1.2.3.4"
		if isIPv6 {
			str = "1:2:3:4::"
		}

		addr := ipaddr.NewIPAddressString(str).GetAddress()
		t.matchesWithColl(list.Floor(addr), nil, list)
		t.matchesWithColl(list.Ceiling(addr), nil, list)
		t.matchesWithColl(list.Lower(addr), nil, list)
		t.matchesWithColl(list.Higher(addr), nil, list)

		t.matchesWithColl(listTrie.Floor(addr), nil, listTrie)
		t.matchesWithColl(listTrie.Ceiling(addr), nil, listTrie)
		t.matchesWithColl(listTrie.Lower(addr), nil, listTrie)
		t.matchesWithColl(listTrie.Higher(addr), nil, listTrie)

		t.rangeListTestCount++

	} else if rangeCount == 1 {
		rng := list.GetSeqRange(0)
		t.testRangeEdges(list, listTrie, nil, rng, nil)
	} else if rangeCount == 2 {
		rng0 := list.GetSeqRange(0)
		rng1 := list.GetSeqRange(1)
		t.testRangeEdges(list, listTrie, nil, rng0, rng1)
		t.testRangeEdges(list, listTrie, rng0, rng1, nil)
	} else if rangeCount == 3 {
		rng0 := list.GetSeqRange(0)
		rng1 := list.GetSeqRange(1)
		rng2 := list.GetSeqRange(2)
		t.testRangeEdges(list, listTrie, nil, rng0, rng1)
		t.testRangeEdges(list, listTrie, rng0, rng1, rng2)
		t.testRangeEdges(list, listTrie, rng1, rng2, nil)
	} else {
		//take the lower range, one in middle, and the upper range

		rng0 := list.GetLowerSeqRange()
		rng1 := list.GetSeqRange(1)
		t.testRangeEdges(list, listTrie, nil, rng0, rng1)

		middleIndex := rangeCount >> 1
		rngMiddlePrevious := list.GetSeqRange(middleIndex - 1)
		rngMiddle := list.GetSeqRange(middleIndex)
		rngMiddleNext := list.GetSeqRange(middleIndex + 1)
		t.testRangeEdges(list, listTrie, rngMiddlePrevious, rngMiddle, rngMiddleNext)

		rngUpperPrevious := list.GetSeqRange(rangeCount - 2)
		rngUpper := list.GetUpperSeqRange()
		t.testRangeEdges(list, listTrie, rngUpperPrevious, rngUpper, nil)
	}
}

func (t *collectionTester) testRangeEdges(list *ipaddr.IPAddressSeqRangeList, listTrie *ipaddr.IPAddressContainmentTrie, left, middle, right *ipaddr.IPAddressSeqRange) {
	hasLeft := left != nil
	hasRight := right != nil
	isMultiple := middle.IsMultiple()

	lowerAddr := middle.GetLower()

	// lower: highest < addr
	// floor: highest <= addr
	// ceiling: lowest >= addr
	// higher: lower > addr

	// address to the left
	if !lowerAddr.IsZero() {
		var lower *ipaddr.IPAddress
		if hasLeft {
			lower = left.GetUpper()
		}
		t.testFourEdges(list, listTrie, lowerAddr.DecrementSingle(), lower, lower, lowerAddr, lowerAddr)
	}

	// lower boundary
	var expectedLower, expectedHigher *ipaddr.IPAddress
	if hasLeft {
		expectedLower = left.GetUpper()
	}
	if isMultiple {
		expectedHigher = lowerAddr.IncrementSingle()
	} else if hasRight {
		expectedHigher = right.GetLower()
	}
	t.testFourEdges(list, listTrie, lowerAddr, expectedLower, lowerAddr, lowerAddr, expectedHigher)

	if isMultiple || !lowerAddr.IsMax() {
		lowerNext := lowerAddr.IncrementSingle()
		upperAddr := middle.GetUpper()

		if isMultiple {
			//addresses in the middle
			if !lowerNext.Equal(upperAddr) {
				lowerNextNext := lowerNext.IncrementSingle()
				t.testFourEdges(list, listTrie, lowerNext, lowerAddr, lowerNext, lowerNext, lowerNextNext)

				if !lowerNextNext.Equal(upperAddr) {
					upperPrevious := upperAddr.DecrementSingle()
					t.testFourEdges(list, listTrie, upperPrevious, upperPrevious.DecrementSingle(), upperPrevious, upperPrevious, upperAddr)
				}
			}

			// upper boundary
			if isMultiple {
				expectedLower = upperAddr.DecrementSingle()
			} else if hasLeft {
				expectedLower = left.GetUpper()
			} else {
				expectedLower = nil
			}
			if hasRight {
				expectedHigher = right.GetLower()
			} else {
				expectedHigher = nil
			}
			t.testFourEdges(list, listTrie, upperAddr, expectedLower, upperAddr, upperAddr, expectedHigher)
		}

		// address to the right
		if !upperAddr.IsMax() {
			var upper *ipaddr.IPAddress
			if hasRight {
				upper = right.GetLower()
			}
			t.testFourEdges(list, listTrie, upperAddr.IncrementSingle(), upperAddr, upperAddr, upper, upper)
		}
	}
}

func (t *collectionTester) testFourEdges(list *ipaddr.IPAddressSeqRangeList, listTrie *ipaddr.IPAddressContainmentTrie, addr,
	expectedLower, expectedFloor, expectedCeiling, expectedHigher *ipaddr.IPAddress) {
	t.matchesWithColl(list.Lower(addr), expectedLower, list)
	t.matchesWithColl(list.Floor(addr), expectedFloor, list)
	t.matchesWithColl(list.Ceiling(addr), expectedCeiling, list)
	t.matchesWithColl(list.Higher(addr), expectedHigher, list)

	t.matchesWithColl(listTrie.Lower(addr), expectedLower, listTrie)
	t.matchesWithColl(listTrie.Floor(addr), expectedFloor, listTrie)
	t.matchesWithColl(listTrie.Ceiling(addr), expectedCeiling, listTrie)
	t.matchesWithColl(listTrie.Higher(addr), expectedHigher, listTrie)

	t.rangeListTestCount++
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
//
//
//
//
//
//
//

func newAddressResult(
	t *collectionTester,
	isIPv6 bool,
	list *ipaddr.IPAddressSeqRangeList,
	addr *ipaddr.IPAddress,
	intersection,
	union,
	remove *ipaddr.IPAddressSeqRangeList) *addrTest {
	return &addrTest{
		isIPv6:               isIPv6,
		list:                 list,
		addr:                 addr,
		expectedIntersection: intersection,
		expectedUnion:        union,
		expectedRemove:       remove,
		t:                    t,
	}
}

type addrTest struct {
	isIPv6 bool

	list, expectedIntersection, expectedUnion, expectedRemove *ipaddr.IPAddressSeqRangeList

	addr *ipaddr.IPAddress

	t *collectionTester
}

func (addrTest *addrTest) runTest() {
	t := addrTest.t

	t.addressTestCount++

	listTrie, addrTrie := convertListToTrie(addrTest.list), convertAddrToTrie(addrTest.addr)

	expectedUnionTrie := convertListToTrie(addrTest.expectedUnion)

	testCollectionOpSingleAddress(addrTest, addrTest.list, addrTest.addr, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].Add, "add", addrTest.expectedUnion, addrTest.expectedUnion)
	testCollectionOpSingleAddress(addrTest, listTrie, addrTest.addr, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].Add, "add", addrTest.expectedUnion, expectedUnionTrie)

	addrTest.testOpSingleAddress(addrTest.list, addrTest.addr, (*ipaddr.IPAddressSeqRangeList).Intersect, "intersect", addrTest.expectedIntersection)

	expectedOverlap := !addrTest.expectedIntersection.IsEmpty()
	testCollectionBooleanOpSingleAddress(addrTest, addrTest.list, addrTest.addr, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].OverlapsAddress, "overlaps", expectedOverlap)
	testCollectionBooleanOpSingleAddress(addrTest, listTrie, addrTest.addr, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].OverlapsAddress, "overlaps", expectedOverlap)

	expectedRemoveTrie := convertListToTrie(addrTest.expectedRemove)
	testCollectionOpSingleAddress(addrTest, addrTest.list, addrTest.addr, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].Remove, "remove", addrTest.expectedRemove, addrTest.expectedRemove)
	testCollectionOpSingleAddress(addrTest, listTrie, addrTest.addr, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].Remove, "remove", addrTest.expectedRemove, expectedRemoveTrie)

	// A contains (A intersect B)

	intersection := addrTest.binaryOpSingleAddress(addrTest.list, addrTest.addr, (*ipaddr.IPAddressSeqRangeList).Intersect, "intersect")
	t.contains(addrTest.list, intersection, true)

	// (A intersect B) contains A iff B contains A

	addrTest.containsAddress(intersection, addrTest.addr, addrTest.list.Contains(addrTest.addr))

	union := addrTest.binaryOpSingleAddress(addrTest.list, addrTest.addr, (*ipaddr.IPAddressSeqRangeList).Add, "add")
	t.contains(union, addrTest.list, true)

	addrTest.containsAddress(union, addrTest.addr, true)

	t.contains(addrTest.list, union, addrTest.list.Contains(addrTest.addr))

	unionTrie := addrTest.collectionBinaryOpSingleAddress(listTrie, addrTest.addr, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].Add, "add")

	t.collectionsMatch(union, unionTrie) // collection test

	addrTest.containsAddress(unionTrie, addrTest.addr, true) // collection test

	empty := &ipaddr.IPAddressSeqRangeList{}
	emptyTrie := &ipaddr.IPAddressContainmentTrie{}

	// removal of oneself always results in nothing
	testCollectionOpSingleAddress(addrTest, addressInList(addrTest.addr), addrTest.addr, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].Remove, "remove", empty, empty) // collection test
	testCollectionOpSingleAddress(addrTest, addrTrie, addrTest.addr, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].Remove, "remove", empty, emptyTrie)                 // collection test

	// intersection with oneself results in the same
	addrTest.testOpSingleAddress(addressInList(addrTest.addr), addrTest.addr, (*ipaddr.IPAddressSeqRangeList).Intersect, "intersect", addressInList(addrTest.addr))

	// union with oneself results in the same
	testCollectionOpSingleAddress(addrTest, addressInList(addrTest.addr), addrTest.addr, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].Add, "add", addressInList(addrTest.addr), addressInList(addrTest.addr)) // collection test
	testCollectionOpSingleAddress(addrTest, addrTrie, addrTest.addr, ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress].Add, "add", addressInList(addrTest.addr), addrTrie)                                         // collection test

	// removing one from the other, removing the other from the one, then taking the union, is the same as removing the intersection from the union
	t.matches(
		t.binaryOp(addrTest.binaryOpSingleAddress(addrTest.list, addrTest.addr, (*ipaddr.IPAddressSeqRangeList).Remove, "remove"),
			t.binaryOp(addressInList(addrTest.addr), addrTest.list, (*ipaddr.IPAddressSeqRangeList).RemoveIntoNew, "remove"),
			(*ipaddr.IPAddressSeqRangeList).JoinIntoNew, "join"),
		t.binaryOp(addrTest.binaryOpSingleAddress(addrTest.list, addrTest.addr, (*ipaddr.IPAddressSeqRangeList).Add, "add"),
			addrTest.binaryOpSingleAddress(addrTest.list, addrTest.addr, (*ipaddr.IPAddressSeqRangeList).Intersect, "intersect"),
			(*ipaddr.IPAddressSeqRangeList).RemoveIntoNew, "remove"))

	everythingStr := "0.0.0.0/0"
	if addrTest.isIPv6 {
		everythingStr = "::/0"
	}
	everythingAddr := ipaddr.NewIPAddressString(everythingStr).GetAddress().ToPrefixBlock()

	// De Morgan's Law 1
	// complement of the union is the same as the intersection of the complements
	t.matches(t.complementWrapper(
		addrTest.binaryOpSingleAddress(addrTest.list, addrTest.addr, (*ipaddr.IPAddressSeqRangeList).Add, "add"), everythingAddr),
		t.binaryOp(t.complementWrapper(addrTest.list, everythingAddr),
			t.complementWrapper(addressInList(addrTest.addr), everythingAddr), (*ipaddr.IPAddressSeqRangeList).IntersectIntoNew, "intersect"))

	// De Morgan's Law 2
	// complement of the intersection is the same as the union of the complements
	t.matches(
		t.complementWrapper(addrTest.binaryOpSingleAddress(addrTest.list, addrTest.addr, (*ipaddr.IPAddressSeqRangeList).Intersect, "intersect"), everythingAddr),
		t.binaryOp(t.complementWrapper(addrTest.list, everythingAddr),
			t.complementWrapper(addressInList(addrTest.addr), everythingAddr), (*ipaddr.IPAddressSeqRangeList).JoinIntoNew, "join"))

	addrTest.removeIntersection(addrTest.list, addrTest.addr)

	addrTest.unionDoubleIntersect(addrTest.list, addrTest.addr, everythingAddr)

	// double complement results in the original

	addrTest.doubleComplement(addrTest.addr, everythingAddr)

	addrTest.intersectUnion(addrTest.list, addrTest.addr)

	addrTest.unionIntersect(addrTest.list, addrTest.addr)

	addrTest.removeIntersectComplement(addrTest.list, addrTest.addr)

	addrTest.everythingNothing(addrTest.addr, everythingAddr)

	addrTest.testRangeListAndAddressSpans(addrTest.expectedUnion, addrTest.list, addrTest.addr, listTrie) // collection test

	t.testIterateRangeList(addressInList(addrTest.addr))

	t.testIterateTrie(listTrie) // collection test

	t.testEdges(addrTest.list, listTrie, addrTest.isIPv6)

	t.testCoverAddr(addrTest.list, listTrie, addrTest.addr)

	addrTest.testTreeGrowThenShrink(addrTrie, addrTest.addr)
}

func (r *addrTest) testTreeGrowThenShrink(addrTrie *ipaddr.IPAddressContainmentTrie, addr *ipaddr.IPAddress) {
	if !addr.IsSequential() {
		return
	}
	t := r.t
	trie := addrTrie.Clone()
	if !trie.EqualAggregation(addr) {
		t.addRangeFailure("unexpected mismatch, "+trie.String()+" and "+addr.String(), addrTrie)
	}
	if !trie.Equal(addrTrie) {
		t.addRangeFailure("unexpected mismatch, "+trie.String()+" and "+addrTrie.String(), addrTrie)
	}
	newBlock := addr.ToPrefixBlockLen(16)
	trie.Add(newBlock)
	lpb := trie.GetLowerPrefixBlock()
	if !lpb.Equal(newBlock) {
		t.addRangeFailure("unexpected mismatch, "+lpb.String()+" and "+newBlock.String(), addrTrie)
	}
	upb := trie.GetUpperPrefixBlock()
	if !upb.Equal(newBlock) {
		t.addRangeFailure("unexpected mismatch, "+upb.String()+" and "+newBlock.String(), addrTrie)
	}
	if !lpb.GetLower().Equal(addr.GetLower()) {
		trie.RemoveSeqRange(lpb.GetLower().SpanWithRange(addr.DecrementSingle()))
	}
	if !upb.GetUpper().Equal(addr.GetUpper()) {
		trie.RemoveSeqRange(upb.GetUpper().SpanWithRange(addr.IncrementBoundarySingle()))
	}
	if !trie.EqualAggregation(addr) {
		t.addRangeFailure("unexpected mismatch, "+trie.String()+" and "+addr.String(), addrTrie)
	}
	if !trie.Equal(addrTrie) {
		t.addRangeFailure("unexpected mismatch, "+trie.String()+" and "+addrTrie.String(), addrTrie)
	}
}

func (r *addrTest) testRangeListAndAddressSpans(list, joined1 *ipaddr.IPAddressSeqRangeList, joined2 *ipaddr.IPAddress, joined1Trie *ipaddr.IPAddressContainmentTrie) {
	t := r.t
	resultPrefixBlocks := list.SpanWithPrefixBlocks()
	resultSequentialBlocks := list.SpanWithSequentialBlocks()
	prefixBlocks1 := joined1.SpanWithPrefixBlocks()
	prefixBlocks2 := joined2.SpanWithPrefixBlocks()
	seqBlocks1 := joined1.SpanWithSequentialBlocks()
	seqBlocks2 := joined2.SpanWithSequentialBlocks()

	t.testPrefixBlockSpan(resultPrefixBlocks, prefixBlocks1, prefixBlocks2, list)
	t.testPrefixBlockSpan(resultPrefixBlocks, seqBlocks1, seqBlocks2, list)
	t.testPrefixBlockSpan(resultPrefixBlocks, prefixBlocks1, []*ipaddr.IPAddress{joined2}, list)
	t.testSequentialBlockSpan(resultSequentialBlocks, seqBlocks1, seqBlocks2, list)
	t.testSequentialBlockSpan(resultSequentialBlocks, prefixBlocks1, prefixBlocks2, list)
	t.testSequentialBlockSpan(resultSequentialBlocks, seqBlocks1, []*ipaddr.IPAddress{joined2}, list)

	t.compareSpanningBlocks(prefixBlocks1, joined1Trie.PrefixBlockIterator(), joined1Trie.GetPrefixBlockCount(), list)
}

// removing the intersection is the same as removing the original
func (r *addrTest) removeIntersection(list *ipaddr.IPAddressSeqRangeList, addr *ipaddr.IPAddress) {
	t := r.t
	t.matches(t.binaryOp(list, r.binaryOpSingleAddress(list, addr, (*ipaddr.IPAddressSeqRangeList).Intersect, "intersect"), (*ipaddr.IPAddressSeqRangeList).RemoveIntoNew, "remove"),
		r.binaryOpSingleAddress(list, addr, (*ipaddr.IPAddressSeqRangeList).Remove, "remove"))
}

// double complement results in the original
func (r *addrTest) doubleComplement(addr, everythingAddr *ipaddr.IPAddress) {
	t := r.t
	list := &ipaddr.IPAddressSeqRangeList{}
	list.Add(addr)
	complement := list.ComplementIntoNew()
	count := complement.GetCount()
	t.matchesCount(everythingAddr.GetCount(), count.Add(count, addr.GetCount()), list)
	t.matches(
		addressInList(addr),
		complementWrapper(complement, everythingAddr))
}

// A intersect (A union B) = A
func (r *addrTest) intersectUnion(list *ipaddr.IPAddressSeqRangeList, addr *ipaddr.IPAddress) {
	t := r.t
	t.matches(
		list,
		t.binaryOp(list, r.binaryOpSingleAddress(list, addr, (*ipaddr.IPAddressSeqRangeList).Add, "add"),
			(*ipaddr.IPAddressSeqRangeList).IntersectIntoNew, "intersect"))
}

// A union (A intersect B) = A
func (r *addrTest) unionIntersect(list *ipaddr.IPAddressSeqRangeList, addr *ipaddr.IPAddress) {
	t := r.t
	t.matches(
		list,
		t.binaryOp(list, r.binaryOpSingleAddress(list, addr, (*ipaddr.IPAddressSeqRangeList).Intersect, "intersect"),
			(*ipaddr.IPAddressSeqRangeList).JoinIntoNew, "join"))
}

// intersect with complement and then with original, take union, should be the original
func (r *addrTest) unionDoubleIntersect(list *ipaddr.IPAddressSeqRangeList, addr, everythingAddr *ipaddr.IPAddress) {
	t := r.t
	t.matches(
		t.binaryOp(
			r.binaryOpSingleAddress(list, addr, (*ipaddr.IPAddressSeqRangeList).Intersect, "intersect"),
			r.binaryOpSingleAddress(
				t.complementWrapper(list, everythingAddr),
				addr,
				(*ipaddr.IPAddressSeqRangeList).Intersect, "intersect"),
			(*ipaddr.IPAddressSeqRangeList).JoinIntoNew, "join"),
		addressInList(addr))
}

// A remove B = A intersect (B complement)
func (r *addrTest) removeIntersectComplement(list *ipaddr.IPAddressSeqRangeList, addr *ipaddr.IPAddress) {
	t := r.t
	addrList := &ipaddr.IPAddressSeqRangeList{}
	addrList.Add(addr)
	complement := addrList.ComplementIntoNew()
	t.matches(
		r.binaryOpSingleAddress(list, addr, (*ipaddr.IPAddressSeqRangeList).Remove, "remove"),
		t.binaryOp(list /* inList(addr.complement())*/, complement, (*ipaddr.IPAddressSeqRangeList).IntersectIntoNew, "intersect"))
}

// A Union (A complement) = everything
// A intersect (A complement) = nothing
func (r *addrTest) everythingNothing(addr, everythingAddr *ipaddr.IPAddress) {
	t := r.t
	list := &ipaddr.IPAddressSeqRangeList{}
	list.Add(addr)
	complement := list.ComplementIntoNew()
	nothing := &ipaddr.IPAddressSeqRangeList{}
	t.matches(
		nothing,
		r.binaryOpSingleAddress(complement, addr, (*ipaddr.IPAddressSeqRangeList).Intersect, "intersect"))

	nothing.Add(everythingAddr)
	everything := nothing
	t.matches(
		everything,
		r.binaryOpSingleAddress(complement, addr, (*ipaddr.IPAddressSeqRangeList).Add, "add"))
}

func (r *addrTest) containsAddress(containing ipaddr.IPAddressCollection, contained *ipaddr.IPAddress, expected bool) {
	t := r.t
	t.rangeListTestCount++
	if containing.Contains(contained) == expected {
		if expected && containing.GetCount().Cmp(contained.GetCount()) < 0 {
			t.addRangeFailure("failed count for containment for list: "+containing.String()+" and address: "+contained.String()+" expected containment: "+fmt.Sprint(expected), r.list)
		} else {
			if printPass {
				fmt.Println("pass")
			}
		}
	} else {
		t.addRangeFailure("fail contains for collection: "+containing.String()+" and contained address: "+contained.String()+" expected containment: "+fmt.Sprint(expected), r.list)
	}
}

func (r *addrTest) binaryOpSingleAddress(list *ipaddr.IPAddressSeqRangeList, address *ipaddr.IPAddress, op func(*ipaddr.IPAddressSeqRangeList, *ipaddr.IPAddress) bool, opName string) *ipaddr.IPAddressSeqRangeList {
	t := r.t
	res := list.Clone()
	val := op(res, address)
	if list.Equal(res) == val { // val true changed, list equal res then false, then false != true is true
		t.addRangeFailure("failed return value for "+opName+" for list: "+list.String()+" and address: "+address.String()+" expected same: "+fmt.Sprint(val)+" original: "+list.String()+" result: "+res.String(), list)
	}
	if print {
		fmt.Println(list)
		fmt.Println(address)
		fmt.Println(opName)
		fmt.Println(res)
		fmt.Println(val)
		fmt.Println()
	}
	return res
}

func (r *addrTest) collectionBinaryOpSingleAddress(
	collection *ipaddr.IPAddressContainmentTrie,
	address *ipaddr.IPAddress,
	op func(ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress],
		*ipaddr.IPAddress) bool, opName string) ipaddr.IPAddressCollection {
	t := r.t
	res := collection.Clone()
	val := op(res, address)
	if collection.EqualAggregation(res) == val { // val true changed, list equal res then false, then false != true is true
		t.addRangeFailure("failed return value for "+opName+" for collection: "+collection.String()+" and address: "+address.String()+" expected same: "+fmt.Sprint(val)+" original: "+r.list.String()+" result: "+res.String(), r.list)
	}
	if print {
		fmt.Println(collection)
		fmt.Println(address)
		fmt.Println(opName)
		fmt.Println(res)
		fmt.Println(val)
		fmt.Println()
	}
	return res
}

func (r *addrTest) testOpSingleAddress(
	list *ipaddr.IPAddressSeqRangeList,
	addr *ipaddr.IPAddress,
	op func(*ipaddr.IPAddressSeqRangeList, *ipaddr.IPAddress) bool,
	opName string,
	expected *ipaddr.IPAddressSeqRangeList) {
	t := r.t
	t.rangeListTestCount++
	res := list.Clone()
	val := op(res, addr)
	if res.Equal(expected) {
		if res.GetCount().Cmp(expected.GetCount()) != 0 || res.GetSeqRangeCount() != expected.GetSeqRangeCount() {
			t.addRangeFailure("failed count for "+opName+" for list: "+list.String()+" and address: "+addr.String()+" expected: "+expected.String()+" result: "+res.String(), list)
		} else if list.Equal(res) == val {
			t.addRangeFailure("failed return value for "+opName+" for list: "+list.String()+" and address: "+addr.String()+" expected same: "+fmt.Sprint(val)+" original: "+list.String()+" result: "+res.String(), list)
		} else {
			if printPass {
				fmt.Println("pass " + opName)
			}
		}
	} else {
		t.addRangeFailure("fail "+opName+" for list: "+list.String()+" and address: "+addr.String()+" expected: "+expected.String()+" actual: "+res.String(), list)
	}
}

func testCollectionOpSingleAddress[T ipaddr.IPAddressCollConstraint[T, *ipaddr.IPAddress]](
	r *addrTest,
	coll T,
	addr *ipaddr.IPAddress,
	op func(ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress], *ipaddr.IPAddress) bool,
	opName string,
	expectedList *ipaddr.IPAddressSeqRangeList,
	expected T) {
	t := r.t
	t.rangeListTestCount++
	res := coll.Clone()
	val := op(res, addr)
	if res.EqualAggregation(expectedList) {
		if res.EqualAggregation(expected) {
			if res.Equal(expected) {
				if res.GetCount().Cmp(expected.GetCount()) != 0 /*|| res.getSeqRangeCount() != expected.getSeqRangeCount()*/ {
					t.addRangeFailure("failed count for "+opName+" for list: "+expectedList.String()+" and address: "+addr.String()+" expected: "+expected.String()+" result: "+res.String(), coll)
				} else if coll.EqualAggregation(res) == val {
					t.addRangeFailure("failed return value for "+opName+" for list: "+expectedList.String()+" and address: "+addr.String()+" expected same: "+fmt.Sprint(val)+" original: "+coll.String()+" result: "+res.String(), coll)
				} else {
					if printPass {
						fmt.Println("pass " + opName)
					}
				}
			} else {
				t.addRangeFailure("fail "+opName+" for equal list: "+coll.String()+" and address: "+addr.String()+" expected: "+expected.String()+" actual: "+res.String(), coll)
			}
		} else {
			t.addRangeFailure("fail "+opName+" for equal list: "+coll.String()+" and address: "+addr.String()+" expected: "+expected.String()+" actual: "+res.String(), coll)
			fmt.Println("fail " + opName + " for equal list: " + coll.String() + " and address: " + addr.String() + " expected: " + expected.String() + " actual: " + res.String())
			fmt.Println("count new", res.GetCount())
			fmt.Println("count old", expected.GetCount())
			fmt.Println("expected list", expectedList)
		}
	} else {
		t.addRangeFailure("fail "+opName+" for aggregation equallist: "+coll.String()+" and address: "+addr.String()+" expected: "+expectedList.String()+" actual: "+res.String(), coll)
	}
}

func testCollectionBooleanOpSingleAddress[T ipaddr.IPAddressCollConstraint[T, *ipaddr.IPAddress]](
	r *addrTest,
	coll T,
	addr *ipaddr.IPAddress,
	op func(ipaddr.IPAddressCollAddrConstraint[*ipaddr.IPAddress], *ipaddr.IPAddress) bool,
	opName string,
	expected bool) {
	t := r.t
	t.rangeListTestCount++
	res := op(coll, addr)
	if res == expected {
		if printPass {
			fmt.Println("pass " + opName)
		}
	} else {
		t.addRangeFailure("fail "+opName+" for list: "+coll.String()+" and address: "+addr.String()+" expected: "+fmt.Sprint(expected)+" actual: "+fmt.Sprint(res), coll)
	}
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
//
//
//
// end of addressResult
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
//
//
//

// all failures in this test module go through addRangeFailure
func (t *collectionTester) addRangeFailure(message string, list ipaddr.AddressAggregation) { //IPAddressCollection<IPAddress, IPAddressSeqRange>
	t.rangeListFailCount++
	t.addFailure(newCollectionFailure(message, list))
}

//unaryOp and binaryOp are unnecessary, we could just call the op.  But they can be useful for debugging.

func complementWrapper[S ipaddr.IPAddressCollConstraint[S, T], T ipaddr.IPAddressTypeConstraint[T]](coll S, everythingAddr T) S {

	if coll.IsEmpty() {
		result := coll.NewEmpty()
		result.Add(everythingAddr)
		return result
	}
	return unaryOp[S, T](coll, S.ComplementIntoNew, "complement")
}

func (t *collectionTester) complementWrapper(list *ipaddr.IPAddressSeqRangeList, everythingAddr *ipaddr.IPAddress) *ipaddr.IPAddressSeqRangeList {
	if list.IsEmpty() {
		return everythingAddr.IntoSequentialRangeList()
	}
	return t.unaryOp(list, (*ipaddr.IPAddressSeqRangeList).ComplementIntoNew, "complement")
}

func binaryOp[S ipaddr.IPAddressCollConstraint[S, *ipaddr.IPAddress]](list1, list2 S, op func(one, two S) S, opName string) S {
	res := op(list1, list2)
	if print {
		fmt.Println(list1)
		fmt.Println(list2)
		fmt.Println(opName)
		fmt.Println(res)
		fmt.Println()
	}
	return res
}

func (t *collectionTester) binaryOp(list1, list2 *ipaddr.IPAddressSeqRangeList, op func(one, two *ipaddr.IPAddressSeqRangeList) *ipaddr.IPAddressSeqRangeList, opName string) *ipaddr.IPAddressSeqRangeList {
	res := op(list1, list2)
	if print {
		fmt.Println(list1)
		fmt.Println(list2)
		fmt.Println(opName)
		fmt.Println(res)
		fmt.Println()
	}
	return res
}

func unaryOp[S ipaddr.IPAddressCollConstraint[S, T], T ipaddr.IPAddressTypeConstraint[T]](coll S, op func(S) S, opName string) S {
	res := op(coll)
	if print {
		fmt.Println(coll)
		fmt.Println(opName)
		fmt.Println(res)
		fmt.Println()
	}
	return res
}

func (t *collectionTester) unaryOp(list *ipaddr.IPAddressSeqRangeList, op func(*ipaddr.IPAddressSeqRangeList) *ipaddr.IPAddressSeqRangeList, opName string) *ipaddr.IPAddressSeqRangeList {
	res := op(list)
	if print {
		fmt.Println(list)
		fmt.Println(opName)
		fmt.Println(res)
		fmt.Println()
	}
	return res
}

func inList(rngs ...*ipaddr.IPAddressSeqRange) *ipaddr.IPAddressSeqRangeList {
	list := ipaddr.IPAddressSeqRangeList{}
	for _, rng := range rngs {
		list.AddSeqRange(rng)
	}
	return &list
}

func inTrie(rngs ...*ipaddr.IPAddressSeqRange) *ipaddr.IPAddressContainmentTrie {
	trie := ipaddr.IPAddressContainmentTrie{}
	for _, rng := range rngs {
		trie.AddSeqRange(rng)
	}
	return &trie
}

func addressInList(addr *ipaddr.IPAddress) *ipaddr.IPAddressSeqRangeList {
	return addr.IntoSequentialRangeList()
}

func contains[S ipaddr.IPAddressCollConstraint[S, *ipaddr.IPAddress]](t *collectionTester, containing, contained S, expected bool) {
	t.rangeListTestCount++
	if containing.ContainsOther(contained) == expected {
		if expected && containing.GetCount().Cmp(contained.GetCount()) < 0 {
			t.addRangeFailure("failed count for containment for containing list: "+containing.String()+" of count "+containing.GetCount().String()+" and contained list "+contained.String()+" of count "+contained.GetCount().String()+", expected containment: "+fmt.Sprint(expected), containing)
		} else {
			if printPass {
				fmt.Println("pass")
			}
		}
	} else {
		t.addRangeFailure("fail contains for collection: "+containing.String()+" and contained address: "+contained.String()+" expected containment: "+fmt.Sprint(expected), containing)
		fmt.Println("fail contains for collection", containing, "contained", contained, "expected containment", expected)
		containing.ContainsOther(contained)
	}
}

func (t *collectionTester) contains(containing, contained *ipaddr.IPAddressSeqRangeList, expected bool) {
	t.rangeListTestCount++
	if containing.ContainsOther(contained) == expected {
		if expected && containing.GetCount().Cmp(contained.GetCount()) < 0 {
			t.addRangeFailure("failed count for containment for containing list: "+containing.String()+" of count "+containing.GetCount().String()+" and contained list "+contained.String()+" of count "+contained.GetCount().String()+", expected containment: "+fmt.Sprint(expected), containing)
		} else {
			if printPass {
				fmt.Println("pass")
			}
		}
	} else {
		t.addRangeFailure("fail contains for list: "+containing.String()+" and contained address: "+contained.String()+" expected containment: "+fmt.Sprint(expected), containing)
	}
}

func (t *collectionTester) matchesCount(res, expected *big.Int, list *ipaddr.IPAddressSeqRangeList) {
	t.rangeListTestCount++
	if res.Cmp(expected) == 0 {
		if printPass {
			fmt.Println("pass")
		}
	} else {
		t.addRangeFailure("fail match, got: "+res.String()+", expected match: "+expected.String(), list)
	}
}

func matchesCount[S ipaddr.IPAddressCollConstraint[S, *ipaddr.IPAddress]](t *collectionTester, res, expected *big.Int, list S) {
	t.rangeListTestCount++
	if res.Cmp(expected) == 0 {
		if printPass {
			fmt.Println("pass")
		}
	} else {
		t.addRangeFailure("fail match, got: "+res.String()+", expected match: "+expected.String(), list)
	}
}

func (t *collectionTester) collectionsMatch(res, expected ipaddr.IPAddressCollection) {
	t.rangeListTestCount++
	if res.EqualAggregation(expected) && expected.EqualAggregation(res) {
		if res.GetCount().Cmp(expected.GetCount()) != 0 {
			t.addRangeFailure("failed count for list, got "+res.GetCount().String()+" expected "+expected.GetCount().String(), expected)
		} else {
			if resList, ok := any(res).(*ipaddr.IPAddressSeqRangeList); ok {
				if expectedList, ok := any(res).(*ipaddr.IPAddressSeqRangeList); ok {
					if !resList.Equal(expectedList) {
						t.addRangeFailure("fail match for list, got: "+res.String()+", expected match: "+expected.String(), expected)
					} else if resList.GetSeqRangeCount() != expectedList.GetSeqRangeCount() {
						t.addRangeFailure("failed sequential range count for list, got "+fmt.Sprint(resList.GetSeqRangeCount())+" expected "+fmt.Sprint(expectedList.GetSeqRangeCount()), expectedList)
					} else {
						if printPass {
							fmt.Println("pass")
						}
					}
				}
			} else if resTrie, ok := any(res).(*ipaddr.IPAddressContainmentTrie); ok {
				if expectedTrie, ok := any(res).(*ipaddr.IPAddressContainmentTrie); ok {
					if !resTrie.Equal(expectedTrie) {
						t.addRangeFailure("fail match for list, got: "+resTrie.String()+", expected match: "+expectedTrie.String(), expectedTrie)
					} else {
						if printPass {
							fmt.Println("pass")
						}
					}
				}
			} else {
				if printPass {
					fmt.Println("pass")
				}
			}
		}
	} else {
		t.addRangeFailure("fail match for collection, got: "+res.String()+", expected match: "+expected.String(), expected)
	}
}

func (t *collectionTester) matchesWithColl(one, two *ipaddr.IPAddress, coll ipaddr.IPAddressCollection) {
	if !one.Equal(two) {
		t.addRangeFailure("address mismatch, "+one.String()+" and "+two.String(), coll)
	} else if one != nil && one.GetPrefixLen() != nil {
		t.addRangeFailure("prefix unexpected: "+one.GetPrefixLen().String(), coll)
	} else if two != nil && two.GetPrefixLen() != nil {
		t.addRangeFailure("prefix unexpected: "+two.GetPrefixLen().String(), coll)
	}
}

func (t *collectionTester) matches(res, expected *ipaddr.IPAddressSeqRangeList) {
	t.rangeListTestCount++
	if res.Equal(expected) {
		if res.GetCount().Cmp(expected.GetCount()) != 0 {
			t.addRangeFailure("failed count for list, got "+res.GetCount().String()+" expected "+expected.GetCount().String(), expected)
		} else if res.GetSeqRangeCount() != expected.GetSeqRangeCount() {
			t.addRangeFailure("failed sequential range count for list, got "+fmt.Sprint(res.GetSeqRangeCount())+" expected "+fmt.Sprint(expected.GetSeqRangeCount()), expected)
		} else {
			if printPass {
				fmt.Println("pass")
			}
		}
	} else {
		t.addRangeFailure("fail match for list, got: "+res.String()+", expected match: "+expected.String(), expected)
	}
}

func (t *collectionTester) testPrefixBlockSpan(prefixBlocks, joined1, joined2 []*ipaddr.IPAddress, list *ipaddr.IPAddressSeqRangeList) {
	var merged []*ipaddr.IPAddress
	if len(joined1) > 0 {
		merged = joined1[0].MergeToPrefixBlocks(join(joined1, joined2)...)
	} else if len(joined2) > 0 {
		merged = joined2[0].MergeToPrefixBlocks(join(joined1, joined2)...)
	} else {
		merged = []*ipaddr.IPAddress{}
	}
	t.compareBlocks(prefixBlocks, merged, list)
}

func (t *collectionTester) testSequentialBlockSpan(sequentialBlocks, joined1, joined2 []*ipaddr.IPAddress, list *ipaddr.IPAddressSeqRangeList) {
	var merged []*ipaddr.IPAddress
	if len(joined1) > 0 {
		merged = joined1[0].MergeToSequentialBlocks(join(joined1, joined2)...)
	} else if len(joined2) > 0 {
		merged = joined2[0].MergeToSequentialBlocks(join(joined1, joined2)...)
	} else {
		merged = []*ipaddr.IPAddress{}
	}
	t.compareBlocks(sequentialBlocks, merged, list)
}

func join[T any](one, two []T) []T {
	if len(one) == 0 {
		return two
	} else if len(two) == 0 {
		return one
	}
	return append(append(make([]T, 0, len(one)+len(two)), one...), two...)
}

func (t *collectionTester) compareBlocks(blocks1, blocks2 []*ipaddr.IPAddress, list *ipaddr.IPAddressSeqRangeList) {
	if len(blocks1) != len(blocks2) {
		t.addRangeFailure("blocks mismatch, matching "+fmt.Sprint((blocks1))+" and "+fmt.Sprint(blocks2), list)
	} else {
		for i := 0; i < len(blocks1); i++ {
			if !blocks1[i].Equal(blocks2[i]) {
				t.addRangeFailure("blocks mismatch with block "+blocks1[i].String()+" and "+blocks2[i].String()+", matching "+fmt.Sprint(blocks1)+" and "+fmt.Sprint(blocks2), list)
			} else if !blocks1[i].GetPrefixLen().Equal(blocks2[i].GetPrefixLen()) {
				t.addRangeFailure("blocks prefix mismatch with block "+blocks1[i].String()+" and "+blocks2[i].String()+", matching "+fmt.Sprint(blocks1)+" and "+fmt.Sprint(blocks2), list)
			}
		}
	}
}

func (t *collectionTester) compareSpanningBlocks(spanningBlocks []*ipaddr.IPAddress, blocks2 ipaddr.Iterator[*ipaddr.IPAddress], blocks2Count int, list *ipaddr.IPAddressSeqRangeList) {
	if len(spanningBlocks) != blocks2Count {
		t.addRangeFailure("blocks mismatch, matching "+fmt.Sprint(spanningBlocks)+" with count "+fmt.Sprint(blocks2Count), list)
	} else {
		i := 0
		for blocks2.HasNext() {
			block2 := blocks2.Next()
			spanningBlock := spanningBlocks[i]
			if !spanningBlock.Equal(block2) {
				t.addRangeFailure("blocks mismatch with block "+spanningBlock.String()+" and "+block2.String()+", matching "+fmt.Sprint(spanningBlocks)+" and "+fmt.Sprint(blocks2), list)
			} else {
				spanningBlock = spanningBlock.RemoveBitCountPrefixLen()
				if !spanningBlock.GetPrefixLen().Equal(block2.GetPrefixLen()) {
					t.addRangeFailure("blocks prefix mismatch with block "+spanningBlock.String()+" and "+block2.String()+", matching "+fmt.Sprint(spanningBlocks)+" and "+fmt.Sprint(blocks2), list)
				}
			}
			i++
		}
	}
}

func (t *collectionTester) testRangeListIncrement() {
	list := &ipaddr.IPAddressSeqRangeList{}
	last := ipaddr.NewIPAddressString("2.255.3.4").GetAddress()
	increment := -1
	len := 211
	for i := 0; i < len; i++ {
		first := last.Increment(99)
		increment++
		if i%2 == 0 {
			increment += 7
			last = first.Increment(7)
		} else {
			increment += 9
			last = first.Increment(9)
		}
		rng := first.SpanWithRange(last)
		list.AddSeqRange(rng)
	}

	increment64 := int64(increment)
	t.rangeListTestCount++
	inc := list.IncrementBig(big.NewInt(increment64))
	longInc := list.Increment(increment64)
	if !inc.Equal(last) || !longInc.Equal(last) {
		t.addRangeFailure("failed increment for list "+list.String()+", increment was "+fmt.Sprint(increment)+" result was "+inc.String()+", long increment result was "+longInc.String()+", expected "+last.String(), list)
	} else {
		if list.Enumerate(inc).Int64() != increment64 {
			t.addRangeFailure("failed enumerate for list, got: "+list.Enumerate(inc).String()+", enumerate: "+inc.String()+", expected: "+fmt.Sprint(increment), list)
		} else if !list.GetUpperSeqRange().Contains(inc) {
			t.addRangeFailure("fail containment for list: upper range: "+list.GetUpperSeqRange().String()+", enumerate: "+inc.String()+", increment: "+fmt.Sprint(increment), list)
		} else {
			if printPass {
				fmt.Println("pass increment")
			}

			t.rangeListTestCount++
			// do it again, this time it does binary search on the existing range sizes
			inc = list.IncrementBig(big.NewInt(increment64))
			longInc = list.Increment(increment64)
			if !inc.Equal(last) || !longInc.Equal(last) {
				t.addRangeFailure("failed increment for list "+list.String()+", increment was "+fmt.Sprint(increment)+" result was "+inc.String()+", long increment result was "+longInc.String()+", expected "+last.String(), list)
			} else {
				if list.Enumerate(inc).Int64() != increment64 {
					t.addRangeFailure("fail enumerate for list, got: "+list.Enumerate(inc).String()+", enumerate: "+inc.String()+", expected: "+fmt.Sprint(increment), list)
				} else {
					if printPass {
						fmt.Println("pass increment")
					}
				}
			}
		}
	}

	list.Clear()

	last = ipaddr.NewIPAddressString("2.255.3.4").GetAddress()
	increment = -1
	var midRange, firstRange *ipaddr.IPAddressSeqRange
	midRangeIncrement := 0
	len = 131
	firstLen := 7
	for i := 0; i < len; i++ {
		first := last.Increment(99)
		increment++
		if i%2 == 0 {
			increment += firstLen
			last = first.Increment(7)
		} else {
			increment += 9
			last = first.Increment(9)
		}
		rng := first.SpanWithRange(last)
		if i == len/2 {
			midRangeIncrement = increment
			midRange = rng
		}
		if i == 0 {
			firstRange = rng
		}
		list.AddSeqRange(rng)
	}
	midRangeIncrement64 := int64(midRangeIncrement)
	increment64 = int64(increment)

	// now we do various increments that incrementally populate the rangeSizes

	t.rangeListTestCount++
	firstRangeIncrement := int64(firstLen/2 + firstLen/4)
	expectedFirstIncrement := firstRange.GetLower().Increment(firstRangeIncrement)
	firstIncrement := list.IncrementBig(big.NewInt(firstRangeIncrement))
	longFirstIncrement := list.Increment(firstRangeIncrement)
	if !firstIncrement.Equal(expectedFirstIncrement) || !longFirstIncrement.Equal(expectedFirstIncrement) {
		t.addRangeFailure("failed increment for list "+list.String()+", increment was "+fmt.Sprint(firstRangeIncrement)+" result was "+firstIncrement.String()+", long increment result was "+longFirstIncrement.String()+", expected "+expectedFirstIncrement.String(), list)
	} else {
		if list.Enumerate(firstIncrement).Int64() != firstRangeIncrement {
			t.addRangeFailure("fail enumerate for list "+list.String()+", got: "+list.Enumerate(firstIncrement).String()+", enumerate: "+firstIncrement.String()+", expected: "+fmt.Sprint(firstRangeIncrement), list)
		} else if !list.GetLowerSeqRange().Contains(firstIncrement) {
			t.addRangeFailure("fail containment for list: lower range: "+list.GetLowerSeqRange().String()+", enumerate: "+firstIncrement.String()+", increment: "+fmt.Sprint(firstRangeIncrement), list)
		} else {
			if printPass {
				fmt.Println("pass increment")
			}

			t.rangeListTestCount++
			// do it gain, this time it does binary search on the existing range sizes
			firstIncrement = list.IncrementBig(big.NewInt(firstRangeIncrement))
			longFirstIncrement = list.Increment(firstRangeIncrement)
			if !firstIncrement.Equal(expectedFirstIncrement) || !longFirstIncrement.Equal(expectedFirstIncrement) {
				t.addRangeFailure("failed increment for list "+list.String()+", increment was "+fmt.Sprint(firstRangeIncrement)+" result was "+firstIncrement.String()+", long increment result was "+longFirstIncrement.String()+", expected "+expectedFirstIncrement.String(), list)
			} else {
				if list.Enumerate(firstIncrement).Int64() != firstRangeIncrement {
					t.addRangeFailure("fail enumerate for list "+list.String()+", got: "+list.Enumerate(firstIncrement).String()+", enumerate: "+firstIncrement.String()+", expected: "+fmt.Sprint(firstRangeIncrement), list)
				} else if !list.GetLowerSeqRange().Contains(firstIncrement) {
					t.addRangeFailure("fail containment for list: lower range: "+list.GetLowerSeqRange().String()+", enumerate: "+firstIncrement.String()+", increment: "+fmt.Sprint(firstRangeIncrement), list)
				} else {
					if printPass {
						fmt.Println("pass increment")
					}

					// we do a mid range increment next, it creates some of the range sizes, so some of them are there for the next search
					t.rangeListTestCount++
					midIncrement := list.IncrementBig(big.NewInt(midRangeIncrement64))
					longMidIncrement := list.Increment(midRangeIncrement64)
					if !midIncrement.Equal(midRange.GetUpper()) || !longMidIncrement.Equal(midRange.GetUpper()) {
						t.addRangeFailure("failed increment for list "+list.String()+", increment was "+fmt.Sprint(midRangeIncrement)+" result was "+midIncrement.String()+", long increment result was "+longMidIncrement.String()+", expected "+midRange.GetUpper().String(), list)
					} else {
						if list.Enumerate(midIncrement).Int64() != midRangeIncrement64 {
							t.addRangeFailure("fail enumerate for list "+list.String()+", got: "+list.Enumerate(midIncrement).String()+", enumerate: "+midIncrement.String()+", expected: "+fmt.Sprint(midRangeIncrement), list)
						} else if !midRange.Contains(midIncrement) {
							t.addRangeFailure("fail containment for list: mid range: "+midRange.String()+", enumerate: "+midIncrement.String()+", increment: "+fmt.Sprint(midRangeIncrement), list)
						} else {
							if printPass {
								fmt.Println("pass increment")
							}

							t.rangeListTestCount++
							// do it gain, this time it does binary search on the existing range sizes
							midIncrement = list.IncrementBig(big.NewInt(midRangeIncrement64))
							longMidIncrement = list.Increment(midRangeIncrement64)
							if !midIncrement.Equal(midRange.GetUpper()) || !longMidIncrement.Equal(midRange.GetUpper()) {
								t.addRangeFailure("failed increment for list "+list.String()+", increment was "+fmt.Sprint(midRangeIncrement)+" result was "+midIncrement.String()+", long increment result was "+longMidIncrement.String()+", expected "+midRange.GetUpper().String(), list)
							} else {
								if list.Enumerate(midIncrement).Int64() != midRangeIncrement64 {
									t.addRangeFailure("fail enumerate for list "+list.String()+", got: "+list.Enumerate(midIncrement).String()+", enumerate: "+midIncrement.String()+", expected: "+fmt.Sprint(midRangeIncrement), list)
								} else if !midRange.Contains(midIncrement) {
									t.addRangeFailure("fail containment for list: mid range: "+midRange.String()+", enumerate: "+midIncrement.String()+", increment: "+fmt.Sprint(midRangeIncrement), list)
								} else {
									if printPass {
										fmt.Println("pass increment")
									}
									// now do a range search to the end, some of the range sizes will be there, some not, this tests the switch-over
									t.rangeListTestCount++
									inc = list.IncrementBig(big.NewInt(increment64))
									longInc = list.Increment(increment64)
									if !inc.Equal(last) || !longInc.Equal(last) {
										t.addRangeFailure("failed increment for list "+list.String()+", increment was "+fmt.Sprint(increment)+" result was "+inc.String()+", long increment result was "+longInc.String()+", expected "+last.String(), list)
									} else {
										if list.Enumerate(inc).Int64() != increment64 {
											t.addRangeFailure("fail enumerate for list "+list.String()+", got: "+list.Enumerate(inc).String()+", enumerate: "+inc.String()+", expected: "+fmt.Sprint(increment), list)
										} else if !list.GetUpperSeqRange().Contains(inc) {
											t.addRangeFailure("fail containment for list: upper range: "+list.GetUpperSeqRange().String()+", enumerate: "+inc.String()+", increment: "+fmt.Sprint(increment), list)
										} else {
											if printPass {
												fmt.Println("pass increment")
											}

											t.rangeListTestCount++
											// do it again, this time it does binary search on the existing range sizes
											inc = list.IncrementBig(big.NewInt(increment64))
											longInc = list.Increment(increment64)
											if !inc.Equal(last) || !longInc.Equal(last) {
												t.addRangeFailure("failed increment for list "+list.String()+", increment was "+fmt.Sprint(increment)+" result was "+inc.String()+", long increment result was "+longInc.String()+", expected "+last.String(), list)
											} else {
												if list.Enumerate(inc).Int64() != increment64 {
													t.addRangeFailure("fail enumerate for list "+list.String()+", got: "+list.Enumerate(inc).String()+", enumerate: "+inc.String()+", expected: "+fmt.Sprint(increment), list)
												} else if !list.GetUpperSeqRange().Contains(inc) {
													t.addRangeFailure("fail containment for list: upper range: "+list.GetUpperSeqRange().String()+", enumerate: "+inc.String()+", increment: "+fmt.Sprint(increment), list)
												} else {
													if printPass {
														fmt.Println("pass increment")
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func testEmptyAggrsAndRanges(t *collectionTester, coll ipaddr.IPAddressAggregation, rng ipaddr.IPAddressSeqRangeType) {
	if coll.ContainsRange(rng) {
		t.addRangeFailure("fail ContainsRange "+coll.String()+" for range "+rng.String(), coll)
	} else if coll.OverlapsRange(rng) {
		t.addRangeFailure("fail OverlapsRange "+coll.String()+" for range "+rng.String(), coll)
	}
}

func testEmptyAggrsAndAddrs(t *collectionTester, coll ipaddr.AddressAggregation, addr ipaddr.AddressType) {
	if coll.Contains(addr) {
		t.addRangeFailure("fail Contains "+coll.String()+" for address "+addr.String(), coll)
	} else if coll.OverlapsAddr(addr) {
		t.addRangeFailure("fail OverlapsAddr "+coll.String()+" for address "+addr.String(), coll)
	} else if coll.Enumerate(addr) != nil {
		t.addRangeFailure("fail Enumerate "+coll.String()+" for address "+addr.String(), coll)
	}
}

func (t *collectionTester) testEmptiesAndNils() {
	testEmptyAndNils(t, ipaddr.NewIPAddressString("1.2.3.4").GetAddress().ToIPv4())
	testEmptyAndNils(t, ipaddr.NewIPAddressString("1.2.3.4").GetAddress())
	ipv6Addr, _ := ipaddr.NewIPAddressString("1.2.3.4").GetAddress().ToIPv4().GetIPv4MappedAddress()
	testEmptyAndNils(t, ipv6Addr)
}

func testEmptyAndNils[T ipaddr.IPAddressTypeConstraint[T]](t *collectionTester, regularAddr T) {
	var nilAddr T
	var nilRng *ipaddr.SequentialRange[T]
	var nilRngList *ipaddr.SequentialRangeList[T]
	var nilContTrie *ipaddr.ContainmentTrieBase[T]

	emptyRngList := &ipaddr.SequentialRangeList[T]{}
	emptyContTrie := &ipaddr.ContainmentTrieBase[T]{}

	regRngList := &ipaddr.SequentialRangeList[T]{}
	regRngList.Add(regularAddr)
	regContTrie := &ipaddr.ContainmentTrieBase[T]{}
	regContTrie.Add(regularAddr)

	regularRng := regularAddr.SpanWithRange(regularAddr)

	testEqualAggregations(t, nilAddr, nilAddr)
	testEqualAggregations(t, nilAddr, nilRng)
	testEqualAggregations(t, nilAddr, nilRngList)
	testEqualAggregations(t, nilAddr, nilContTrie)
	testEqualAggregations(t, nilAddr, emptyRngList)
	testEqualAggregations(t, nilAddr, emptyContTrie)

	testEqualAggregations(t, nilRng, nilRng)
	testEqualAggregations(t, nilRng, nilRngList)
	testEqualAggregations(t, nilRng, nilContTrie)
	testEqualAggregations(t, nilRng, emptyRngList)
	testEqualAggregations(t, nilRng, emptyContTrie)

	testEqualAggregations(t, nilRngList, nilRngList)
	testEqualAggregations(t, nilRngList, nilContTrie)
	testEqualAggregations(t, nilRngList, emptyRngList)
	testEqualAggregations(t, nilRngList, emptyContTrie)

	testEqualAggregations(t, nilContTrie, nilContTrie)
	testEqualAggregations(t, nilContTrie, emptyRngList)
	testEqualAggregations(t, nilContTrie, emptyContTrie)

	testEqualAggregations(t, emptyRngList, emptyRngList)
	testEqualAggregations(t, emptyRngList, emptyContTrie)

	testEqualAggregations(t, emptyContTrie, emptyContTrie)

	testEmptyAggregations(t, nilAddr)
	testEmptyAggregations(t, nilRng)
	testEmptyAggregations(t, nilRngList)
	testEmptyAggregations(t, nilContTrie)
	testEmptyAggregations(t, emptyRngList)
	testEmptyAggregations(t, emptyContTrie)

	testEmptyCollections(t, nilRngList, nilRngList)
	testEmptyCollections(t, nilRngList, emptyRngList)

	testEmptyCollections(t, nilContTrie, nilContTrie)
	testEmptyCollections(t, nilContTrie, emptyContTrie)

	testEmptyCollections(t, emptyRngList, emptyRngList)
	testEmptyCollections(t, emptyContTrie, emptyContTrie)

	testEmptyAddrs(t, nilRngList, regularAddr)
	testEmptyAddrs(t, nilContTrie, regularAddr)
	testEmptyAddrs(t, emptyRngList, regularAddr)
	testEmptyAddrs(t, emptyContTrie, regularAddr)
	testEmptyAddrs(t, nilRngList, nilAddr)
	testEmptyAddrs(t, nilContTrie, nilAddr)
	testEmptyAddrs(t, emptyRngList, nilAddr)
	testEmptyAddrs(t, emptyContTrie, nilAddr)
	testEmptyAddrs(t, regRngList, nilAddr)
	testEmptyAddrs(t, regContTrie, nilAddr)

	testEmptyRanges(t, nilRngList, regularRng)
	testEmptyRanges(t, nilContTrie, regularRng)
	testEmptyRanges(t, emptyRngList, regularRng)
	testEmptyRanges(t, emptyContTrie, regularRng)
	testEmptyRanges(t, nilRngList, nilRng)
	testEmptyRanges(t, nilContTrie, nilRng)
	testEmptyRanges(t, emptyRngList, nilRng)
	testEmptyRanges(t, emptyContTrie, nilRng)
	testEmptyRanges(t, regRngList, nilRng)
	testEmptyRanges(t, regContTrie, nilRng)

	testEmptyAggrsAndRanges(t, nilAddr, nilRng)
	testEmptyAggrsAndRanges(t, nilAddr, regularRng)
	testEmptyAggrsAndRanges(t, nilRng, regularRng)
	testEmptyAggrsAndRanges(t, nilRng, nilRng)
	testEmptyAggrsAndRanges(t, nilRngList, nilRng)
	testEmptyAggrsAndRanges(t, nilRngList, regularRng)
	testEmptyAggrsAndRanges(t, nilContTrie, nilRng)
	testEmptyAggrsAndRanges(t, nilContTrie, regularRng)
	testEmptyAggrsAndRanges(t, emptyRngList, nilRng)
	testEmptyAggrsAndRanges(t, emptyRngList, regularRng)
	testEmptyAggrsAndRanges(t, emptyContTrie, nilRng)
	testEmptyAggrsAndRanges(t, emptyContTrie, regularRng)
	testEmptyAggrsAndRanges(t, regRngList, nilRng)
	testEmptyAggrsAndRanges(t, regContTrie, nilRng)

	testEmptyAggrsAndAddrs(t, nilAddr, nilAddr)
	testEmptyAggrsAndAddrs(t, nilAddr, regularAddr)
	testEmptyAggrsAndAddrs(t, nilRng, regularAddr)
	testEmptyAggrsAndAddrs(t, nilRng, nilAddr)
	testEmptyAggrsAndAddrs(t, nilRngList, nilAddr)
	testEmptyAggrsAndAddrs(t, nilRngList, regularAddr)
	testEmptyAggrsAndAddrs(t, nilContTrie, nilAddr)
	testEmptyAggrsAndAddrs(t, nilContTrie, regularAddr)
	testEmptyAggrsAndAddrs(t, emptyRngList, nilAddr)
	testEmptyAggrsAndAddrs(t, emptyRngList, regularAddr)
	testEmptyAggrsAndAddrs(t, emptyContTrie, nilAddr)
	testEmptyAggrsAndAddrs(t, emptyContTrie, regularAddr)
	testEmptyAggrsAndAddrs(t, regRngList, nilAddr)
	testEmptyAggrsAndAddrs(t, regContTrie, nilAddr)
}

func testEqualAggregations(t *collectionTester, coll ipaddr.IPAddressAggregation, coll2 ipaddr.IPAddressAggregation) {
	if !coll.EqualAggregation(coll2) {
		t.addRangeFailure("fail equal aggr "+coll.String()+" for collection "+coll2.String(), coll)
	} else if !coll2.EqualAggregation(coll) {
		t.addRangeFailure("fail equal aggr "+coll2.String()+" for collection "+coll.String(), coll)
	}
}

func testEmptyAggregations(t *collectionTester, coll ipaddr.AddressAggregation) {
	if coll.GetCount().Sign() != 0 {
		t.addRangeFailure("fail count empty aggr "+coll.String(), coll)
	} else if coll.IsMultiple() {
		t.addRangeFailure("fail IsMultiple aggr "+coll.String(), coll)
	}
	iterator := coll.AddressIterator()
	if iterator.HasNext() {
		t.addRangeFailure("fail AddressIterator aggr "+coll.String(), coll)
	}
	next := iterator.Next()
	if next != nil {
		t.addRangeFailure("fail AddressIterator aggr next "+coll.String(), coll)
	}
}

func testEmptyCollections[S ipaddr.IPAddressCollConstraint[S, T], T ipaddr.IPAddressTypeConstraint[T]](t *collectionTester, coll S, coll2 S) {
	if coll.ContainsOther(coll2) != coll2.IsEmpty() {
		t.addRangeFailure("fail contains "+coll2.String()+" for collection "+coll.String(), coll)
	} else if coll.OverlapsOther(coll2) {
		t.addRangeFailure("fail overlaps "+coll2.String()+" for collection "+coll.String(), coll)
	} else if !coll.Equal(coll2) {
		t.addRangeFailure("fail equal "+coll2.String()+" for collection "+coll.String(), coll)
	} else if coll2.ContainsOther(coll) != coll.IsEmpty() {
		t.addRangeFailure("fail contains "+coll.String()+" for collection "+coll2.String(), coll2)
	} else if coll2.OverlapsOther(coll) {
		t.addRangeFailure("fail overlaps "+coll.String()+" for collection "+coll2.String(), coll2)
	} else if !coll2.Equal(coll) {
		t.addRangeFailure("fail equal "+coll.String()+" for collection "+coll2.String(), coll2)
	}
}

func testEmptyAddrs[S ipaddr.IPAddressCollAddrConstraint[T], T ipaddr.IPAddressTypeConstraint[T]](t *collectionTester, coll S, emptyAddr T) {
	if coll.OverlapsAddress(emptyAddr) {
		t.addRangeFailure("fail overlaps "+emptyAddr.String()+" for collection "+coll.String(), coll)
	} else if coll.EnumerateAddress(emptyAddr) != nil {
		t.addRangeFailure("fail enumerate "+emptyAddr.String()+" for collection "+coll.String(), coll)
	} else if coll.ContainsAddress(emptyAddr) {
		t.addRangeFailure("fail contains "+emptyAddr.String()+" for collection "+coll.String(), coll)
	}
}

func testEmptyRanges[S ipaddr.IPAddressCollAddrConstraint[T], T ipaddr.IPAddressTypeConstraint[T]](t *collectionTester, coll S, emptyRng *ipaddr.SequentialRange[T]) {
	if coll.OverlapsSeqRange(emptyRng) {
		t.addRangeFailure("fail overlaps range "+emptyRng.String()+" for collection "+coll.String(), coll)
	} else if coll.ContainsSeqRange(emptyRng) {
		t.addRangeFailure("fail contains range "+emptyRng.String()+" for collection "+coll.String(), coll)
	}
}
func (t *collectionTester) testConversions() {
	t.testAddrConversions()
	t.testRangeConversions()
}

func (t *collectionTester) testAddrConversions() {
	var (
		nilMAC  *ipaddr.MACAddress
		nilIPv4 *ipaddr.IPv4Address
		nilIPv6 *ipaddr.IPv6Address
		mac     = ipaddr.NewMACAddressString("1:2:3:4:5:6").GetAddress()
		ipv4    = ipaddr.NewIPAddressString("1.2.3.4").GetAddress()
		ipv6    = ipaddr.NewIPAddressString("1:2:3:4:5:6:7:8").GetAddress()

		IPv4AsIP = ipv4.ToIP()
		IPv6AsIP = ipv6.ToIP()

		IPv4AsBase = ipv4.ToAddressBase()
		IPv6AsBase = ipv6.ToAddressBase()
		macAsBase  = mac.ToAddressBase()
	)
	isNilConversionOk := true // all nils can always be converted
	isNil := true
	convertAddr[*ipaddr.MACAddress](t, nil, isNil, isNilConversionOk)
	convertAddr[*ipaddr.MACAddress](t, nilMAC, isNil, isNilConversionOk)
	convertAddr[*ipaddr.MACAddress](t, nilIPv4, isNil, isNilConversionOk)
	convertAddr[*ipaddr.MACAddress](t, nilIPv6, isNil, isNilConversionOk)
	isNil = false
	convertAddr[*ipaddr.MACAddress](t, mac, isNil, true)
	convertAddr[*ipaddr.MACAddress](t, ipv4, isNil, false)
	convertAddr[*ipaddr.MACAddress](t, ipv6, isNil, false)
	convertAddr[*ipaddr.MACAddress](t, macAsBase, isNil, true)
	convertAddr[*ipaddr.MACAddress](t, IPv4AsBase, isNil, false)
	convertAddr[*ipaddr.MACAddress](t, IPv6AsBase, isNil, false)
	convertAddr[*ipaddr.MACAddress](t, IPv4AsIP, isNil, false)
	convertAddr[*ipaddr.MACAddress](t, IPv6AsIP, isNil, false)

	isNil = true
	convertAddr[*ipaddr.IPv4Address](t, nil, isNil, isNilConversionOk)
	convertAddr[*ipaddr.IPv4Address](t, nilMAC, isNil, isNilConversionOk)
	convertAddr[*ipaddr.IPv4Address](t, nilIPv4, isNil, isNilConversionOk)
	convertAddr[*ipaddr.IPv4Address](t, nilIPv6, isNil, isNilConversionOk)
	isNil = false
	convertAddr[*ipaddr.IPv4Address](t, mac, isNil, false)
	convertAddr[*ipaddr.IPv4Address](t, ipv4, isNil, true)
	convertAddr[*ipaddr.IPv4Address](t, ipv6, isNil, false)
	convertAddr[*ipaddr.IPv4Address](t, macAsBase, isNil, false)
	convertAddr[*ipaddr.IPv4Address](t, IPv4AsBase, isNil, true)
	convertAddr[*ipaddr.IPv4Address](t, IPv6AsBase, isNil, false)
	convertAddr[*ipaddr.IPv4Address](t, IPv4AsIP, isNil, true)
	convertAddr[*ipaddr.IPv4Address](t, IPv6AsIP, isNil, false)

	isNil = true
	convertAddr[*ipaddr.IPv6Address](t, nil, isNil, isNilConversionOk)
	convertAddr[*ipaddr.IPv6Address](t, nilMAC, isNil, isNilConversionOk)
	convertAddr[*ipaddr.IPv6Address](t, nilIPv4, isNil, isNilConversionOk)
	convertAddr[*ipaddr.IPv6Address](t, nilIPv6, isNil, isNilConversionOk)
	isNil = false
	convertAddr[*ipaddr.IPv6Address](t, mac, isNil, false)
	convertAddr[*ipaddr.IPv6Address](t, ipv4, isNil, false)
	convertAddr[*ipaddr.IPv6Address](t, ipv6, isNil, true)
	convertAddr[*ipaddr.IPv6Address](t, macAsBase, isNil, false)
	convertAddr[*ipaddr.IPv6Address](t, IPv4AsBase, isNil, false)
	convertAddr[*ipaddr.IPv6Address](t, IPv6AsBase, isNil, true)
	convertAddr[*ipaddr.IPv6Address](t, IPv4AsIP, isNil, false)
	convertAddr[*ipaddr.IPv6Address](t, IPv6AsIP, isNil, true)

	isNil = true
	convertAddr[*ipaddr.IPAddress](t, nil, isNil, isNilConversionOk)
	convertAddr[*ipaddr.IPAddress](t, nilMAC, isNil, isNilConversionOk)
	convertAddr[*ipaddr.IPAddress](t, nilIPv4, isNil, isNilConversionOk)
	convertAddr[*ipaddr.IPAddress](t, nilIPv6, isNil, isNilConversionOk)
	isNil = false
	convertAddr[*ipaddr.IPAddress](t, mac, isNil, false)
	convertAddr[*ipaddr.IPAddress](t, ipv4, isNil, true)
	convertAddr[*ipaddr.IPAddress](t, ipv6, isNil, true)
	convertAddr[*ipaddr.IPAddress](t, macAsBase, isNil, false)
	convertAddr[*ipaddr.IPAddress](t, IPv4AsBase, isNil, true)
	convertAddr[*ipaddr.IPAddress](t, IPv6AsBase, isNil, true)
	convertAddr[*ipaddr.IPAddress](t, IPv4AsIP, isNil, true)
	convertAddr[*ipaddr.IPAddress](t, IPv6AsIP, isNil, true)

	isNil = true
	convertAddr[*ipaddr.Address](t, nil, isNil, isNilConversionOk)
	convertAddr[*ipaddr.Address](t, nilMAC, isNil, isNilConversionOk)
	convertAddr[*ipaddr.Address](t, nilIPv4, isNil, isNilConversionOk)
	convertAddr[*ipaddr.Address](t, nilIPv6, isNil, isNilConversionOk)
	isNil = false
	convertAddr[*ipaddr.Address](t, mac, isNil, true)
	convertAddr[*ipaddr.Address](t, ipv4, isNil, true)
	convertAddr[*ipaddr.Address](t, ipv6, isNil, true)
	convertAddr[*ipaddr.Address](t, macAsBase, isNil, true)
	convertAddr[*ipaddr.Address](t, IPv4AsBase, isNil, true)
	convertAddr[*ipaddr.Address](t, IPv6AsBase, isNil, true)
	convertAddr[*ipaddr.Address](t, IPv4AsIP, isNil, true)
	convertAddr[*ipaddr.Address](t, IPv6AsIP, isNil, true)

	isNil = true
	convertAddr[ipaddr.AddressType](t, nil, isNil, true)
	convertAddr[ipaddr.AddressType](t, nilMAC, isNil, isNilConversionOk)
	convertAddr[ipaddr.AddressType](t, nilIPv4, isNil, isNilConversionOk)
	convertAddr[ipaddr.AddressType](t, nilIPv6, isNil, isNilConversionOk)
	isNil = false
	convertAddr[ipaddr.AddressType](t, mac, isNil, true)
	convertAddr[ipaddr.AddressType](t, ipv4, isNil, true)
	convertAddr[ipaddr.AddressType](t, ipv6, isNil, true)
	convertAddr[ipaddr.AddressType](t, macAsBase, isNil, true)
	convertAddr[ipaddr.AddressType](t, IPv4AsBase, isNil, true)
	convertAddr[ipaddr.AddressType](t, IPv6AsBase, isNil, true)
	convertAddr[ipaddr.AddressType](t, IPv4AsIP, isNil, true)
	convertAddr[ipaddr.AddressType](t, IPv6AsIP, isNil, true)
}

func (t *collectionTester) testRangeConversions() {
	var (
		nilIPv4 *ipaddr.IPv4AddressSeqRange
		nilIPv6 *ipaddr.IPv6AddressSeqRange
		ipv4    = ipaddr.NewIPAddressString("1.2.3.4").GetAddress().CoverWithSequentialRange()
		ipv6    = ipaddr.NewIPAddressString("1:2:3:4:5:6:7:8").GetAddress().CoverWithSequentialRange()

		IPv4AsIP = ipv4.ToIP()
		IPv6AsIP = ipv6.ToIP()
	)
	isNilConversionOk := true // all nils can always be converted

	isNil := true
	convertRng[*ipaddr.IPv4Address](t, nil, isNil, isNilConversionOk)
	convertRng[*ipaddr.IPv4Address](t, nilIPv4, isNil, isNilConversionOk)
	convertRng[*ipaddr.IPv4Address](t, nilIPv6, isNil, isNilConversionOk)
	isNil = false
	convertRng[*ipaddr.IPv4Address](t, ipv4, isNil, true)
	convertRng[*ipaddr.IPv4Address](t, ipv6, isNil, false)
	convertRng[*ipaddr.IPv4Address](t, IPv4AsIP, isNil, true)
	convertRng[*ipaddr.IPv4Address](t, IPv6AsIP, isNil, false)

	isNil = true
	convertRng[*ipaddr.IPv6Address](t, nil, isNil, isNilConversionOk)
	convertRng[*ipaddr.IPv6Address](t, nilIPv4, isNil, isNilConversionOk)
	convertRng[*ipaddr.IPv6Address](t, nilIPv6, isNil, isNilConversionOk)
	isNil = false
	convertRng[*ipaddr.IPv6Address](t, ipv4, isNil, false)
	convertRng[*ipaddr.IPv6Address](t, ipv6, isNil, true)
	convertRng[*ipaddr.IPv6Address](t, IPv4AsIP, isNil, false)
	convertRng[*ipaddr.IPv6Address](t, IPv6AsIP, isNil, true)

	isNil = true
	convertRng[*ipaddr.IPAddress](t, nil, isNil, isNilConversionOk)
	convertRng[*ipaddr.IPAddress](t, nilIPv4, isNil, isNilConversionOk)
	convertRng[*ipaddr.IPAddress](t, nilIPv6, isNil, isNilConversionOk)
	isNil = false
	convertRng[*ipaddr.IPAddress](t, ipv4, isNil, true)
	convertRng[*ipaddr.IPAddress](t, ipv6, isNil, true)
	convertRng[*ipaddr.IPAddress](t, IPv4AsIP, isNil, true)
	convertRng[*ipaddr.IPAddress](t, IPv6AsIP, isNil, true)
}

func convertAddr[T ipaddr.AddressType](t *collectionTester, addr ipaddr.AddressType, expectedIsNil, expectedIsOk bool) (T, bool) {
	res, resultIsNil, resultIsOk := ipaddr.ConvertAddressTypeCheckNil[T](addr)
	if resultIsNil != expectedIsNil {
		typeStr := fmt.Sprintf("Type: %T\n", *new(T))
		t.addRangeFailure("fail nil "+fmt.Sprint(resultIsNil)+" expected "+fmt.Sprint(expectedIsNil)+" addr "+addr.String()+" to type "+typeStr, addr)
	}
	if resultIsOk != expectedIsOk {
		typeStr := fmt.Sprintf("Type: %T\n", *new(T))
		fmt.Println(resultIsOk, expectedIsOk)
		t.addRangeFailure("fail ok "+fmt.Sprint(resultIsOk)+" expected "+fmt.Sprint(expectedIsOk)+"addr "+addr.String()+" to type "+typeStr, addr)
	}
	ok := false
	if resultIsOk {
		resAddr, ok := any(res).(ipaddr.AddressType)
		if ok && !resAddr.Equal(addr) {
			typeStr := fmt.Sprintf("Type: %T\n", *new(T))
			t.addRangeFailure("fail conversion to "+resAddr.String()+" of type "+typeStr+" from addr "+addr.String(), addr)
		}
	}

	res, resultIsOk = ipaddr.ConvertAddressType[T](addr)
	if resultIsOk != expectedIsOk {
		typeStr := fmt.Sprintf("Type: %T\n", *new(T))
		t.addRangeFailure("fail conversion ok "+fmt.Sprint(resultIsOk)+" expected "+fmt.Sprint(expectedIsOk)+"addr "+addr.String()+" to type "+typeStr, addr)
	}
	if resultIsOk {
		resAddr, ok := any(res).(ipaddr.AddressType)
		if ok && !resAddr.Equal(addr) {
			typeStr := fmt.Sprintf("Type: %T\n", *new(T))
			t.addRangeFailure("fail conversion to "+resAddr.String()+" of type "+typeStr+" from addr "+addr.String(), addr)
		}
	}
	return res, ok
}

func convertRng[T ipaddr.SequentialRangeConstraint[T]](t *collectionTester, addr ipaddr.IPAddressSeqRangeType, expectedIsNil, expectedIsOk bool) {
	res, resultIsNil, resultIsOk := ipaddr.ConvertRangeTypeCheckNil[T](addr)
	if resultIsNil != expectedIsNil {
		typeStr := fmt.Sprintf("Type: %T\n", *new(ipaddr.SequentialRangeConstraint[T]))
		t.addRangeFailure("fail nil "+fmt.Sprint(resultIsNil)+" expected "+fmt.Sprint(expectedIsNil)+" addr "+addr.String()+" to type "+typeStr, addr)
	}
	if resultIsOk != expectedIsOk {
		typeStr := fmt.Sprintf("Type: %T\n", *new(ipaddr.SequentialRangeConstraint[T]))
		t.addRangeFailure("fail ok "+fmt.Sprint(resultIsOk)+" expected "+fmt.Sprint(expectedIsOk)+"addr "+addr.String()+" to type "+typeStr, addr)
	}
	if resultIsOk {
		resAddr := any(res).(ipaddr.IPAddressSeqRangeType)
		if !resAddr.Equal(addr) {
			typeStr := fmt.Sprintf("Type: %T\n", *new(ipaddr.SequentialRangeConstraint[T]))
			t.addRangeFailure("fail conversion to "+resAddr.String()+" of type "+typeStr+" from addr "+addr.String(), addr)
		}
	}

	res, resultIsOk = ipaddr.ConvertRangeType[T](addr)
	if resultIsOk != expectedIsOk {
		typeStr := fmt.Sprintf("Type: %T\n", *new(ipaddr.SequentialRangeConstraint[T]))
		t.addRangeFailure("fail conversion ok "+fmt.Sprint(resultIsOk)+" expected "+fmt.Sprint(expectedIsOk)+"addr "+addr.String()+" to type "+typeStr, addr)
	}
	if resultIsOk {
		resAddr := any(res).(ipaddr.IPAddressSeqRangeType)
		if !resAddr.Equal(addr) {
			typeStr := fmt.Sprintf("Type: %T\n", *new(ipaddr.SequentialRangeConstraint[T]))
			t.addRangeFailure("fail conversion to "+resAddr.String()+" of type "+typeStr+" from addr "+addr.String(), addr)
		}
	}
}

func executeTests(tests []TestResult) {
	for _, r := range tests {
		r.runTest()
	}
}

func (t *collectionTester) run() {

	t.testConversions()

	t.testEmptiesAndNils()

	range1 := [][]string{
		{"0.0.0.1", "0.0.0.2"},
		{"0.0.0.10", "0.0.0.12"},
		{"0.0.0.20", "0.0.0.22"},
	}

	range2 := [][]string{
		{"0.0.0.1", "0.0.0.3"},
		{"0.0.0.7", "0.0.0.11"},
		{"0.0.0.21", "0.0.0.25"},
	}

	range1range2Union := [][]string{
		{"0.0.0.1", "0.0.0.3"},
		{"0.0.0.7", "0.0.0.12"},
		{"0.0.0.20", "0.0.0.25"},
	}

	range1range2Intersection := [][]string{
		{"0.0.0.1", "0.0.0.2"},
		{"0.0.0.10", "0.0.0.11"},
		{"0.0.0.21", "0.0.0.22"},
	}

	range1range2Remove := [][]string{
		{"0.0.0.12", "0.0.0.12"},
		{"0.0.0.20", "0.0.0.20"},
	}

	range2range1Remove := [][]string{
		{"0.0.0.3", "0.0.0.3"},
		{"0.0.0.7", "0.0.0.9"},
		{"0.0.0.23", "0.0.0.25"},
	}

	empty := [][]string{}

	result := t.initIPv4Lists(
		range1, range2, range1range2Intersection, range1range2Union, range1range2Remove, range2range1Remove)

	result2 := t.initIPv4Lists(
		range1, range1, range1, range1, empty, empty)

	result3 := t.initIPv4Lists(
		range2, range2, range2, range2, empty, empty)

	executeTests(result)

	executeTests(result2)

	executeTests(result3)

	range3 := [][]string{
		{"0.0.0.1", "0.0.0.3"},
		{"0.0.0.7", "0.0.0.7"},
		{"0.0.0.9", "0.0.0.25"},
	}

	range2range3Union := [][]string{
		{"0.0.0.1", "0.0.0.3"},
		{"0.0.0.7", "0.0.0.25"},
	}

	range2range3Intersection := [][]string{
		{"0.0.0.1", "0.0.0.3"},
		{"0.0.0.7", "0.0.0.7"},
		{"0.0.0.9", "0.0.0.11"},
		{"0.0.0.21", "0.0.0.25"},
	}

	range2range3Remove := [][]string{
		{"0.0.0.8", "0.0.0.8"},
	}

	range3range2Remove := [][]string{
		{"0.0.0.12", "0.0.0.20"},
	}

	result4 := t.initIPv4Lists(
		range2, range3, range2range3Intersection, range2range3Union, range2range3Remove, range3range2Remove)

	executeTests(result4)

	range3a := [][]string{
		{"0.0.0.1", "0.0.0.3"},
		{"0.0.0.7", "0.0.0.7"},
		{"0.0.0.9", "0.0.0.25"},
		{"0.0.0.50", "0.0.0.75"},
		{"0.0.0.80", "0.0.0.255"},
	}

	range2range3aUnion := [][]string{
		{"0.0.0.1", "0.0.0.3"},
		{"0.0.0.7", "0.0.0.25"},
		{"0.0.0.50", "0.0.0.75"},
		{"0.0.0.80", "0.0.0.255"},
	}

	range2range3aIntersection := [][]string{
		{"0.0.0.1", "0.0.0.3"},
		{"0.0.0.7", "0.0.0.7"},
		{"0.0.0.9", "0.0.0.11"},
		{"0.0.0.21", "0.0.0.25"},
	}

	range2range3aRemove := [][]string{
		{"0.0.0.8", "0.0.0.8"},
	}

	range3arange2Remove := [][]string{
		{"0.0.0.12", "0.0.0.20"},
		{"0.0.0.50", "0.0.0.75"},
		{"0.0.0.80", "0.0.0.255"},
	}

	executeTests(t.initIPv4Lists(
		range2, range3a, range2range3aIntersection, range2range3aUnion, range2range3aRemove, range3arange2Remove))

	range4 := [][]string{
		{"0.0.0.1", "0.0.0.3"},
		{"0.0.0.7", "0.0.0.9"},
		{"0.0.0.12", "0.0.0.25"},
	}

	range2range4Union := range2range3Union

	range2range4Intersection := [][]string{
		{"0.0.0.1", "0.0.0.3"},
		{"0.0.0.7", "0.0.0.9"},
		{"0.0.0.21", "0.0.0.25"},
	}

	range2range4Remove := [][]string{
		{"0.0.0.10", "0.0.0.11"},
	}

	range4range2Remove := [][]string{
		{"0.0.0.12", "0.0.0.20"},
	}

	result5 := t.initIPv4Lists(
		range2, range4, range2range4Intersection, range2range4Union, range2range4Remove, range4range2Remove)

	executeTests(result5)

	range3range4Union := range2range3Union

	range3range4Intersection := [][]string{
		{"0.0.0.1", "0.0.0.3"},
		{"0.0.0.7", "0.0.0.7"},
		{"0.0.0.9", "0.0.0.9"},
		{"0.0.0.12", "0.0.0.25"},
	}

	range3range4Remove := [][]string{
		{"0.0.0.10", "0.0.0.11"},
	}

	range4range3Remove := [][]string{
		{"0.0.0.8", "0.0.0.8"},
	}

	result6 := t.initIPv4Lists(
		range3, range4, range3range4Intersection, range3range4Union, range3range4Remove, range4range3Remove)

	executeTests(result6)

	range5 := [][]string{
		{"0.0.0.1", "0.0.0.3"},
		{"0.0.0.8", "0.0.0.9"},
		{"0.0.0.12", "0.0.0.25"},
	}

	range3range5Union := range2range3Union

	range3range5Intersection := [][]string{
		{"0.0.0.1", "0.0.0.3"},
		{"0.0.0.9", "0.0.0.9"},
		{"0.0.0.12", "0.0.0.25"},
	}

	range3range5Remove := [][]string{
		{"0.0.0.7", "0.0.0.7"},
		{"0.0.0.10", "0.0.0.11"},
	}

	range5range3Remove := [][]string{
		{"0.0.0.8", "0.0.0.8"},
	}

	result7 := t.initIPv4Lists(
		range3, range5, range3range5Intersection, range3range5Union, range3range5Remove, range5range3Remove)

	executeTests(result7)

	range6 := [][]string{
		{"0.0.0.0", "0.0.3.0"},
		{"0.0.5.0", "0.0.13.0"},
		{"0.0.21.0", "0.0.26.0"},
		{"0.0.36.0", "0.0.40.0"},
		{"0.0.42.0", "0.0.50.0"},
	}

	range7 := [][]string{
		{"0.0.1.0", "0.0.8.0"},
		{"0.0.13.0", "0.0.20.0"},
		{"0.0.27.0", "0.0.35.0"},
		{"0.0.37.0", "0.0.42.0"},
		{"0.0.45.0", "0.0.55.0"},
	}

	range6range7Union := [][]string{
		{"0.0.0.0", "0.0.20.0"},
		{"0.0.21.0", "0.0.26.0"},
		{"0.0.27.0", "0.0.35.0"},
		{"0.0.36.0", "0.0.55.0"},
	}

	range6range7Intersection := [][]string{
		{"0.0.1.0", "0.0.3.0"},
		{"0.0.5.0", "0.0.8.0"},
		{"0.0.13.0", "0.0.13.0"},
		{"0.0.37.0", "0.0.40.0"},
		{"0.0.42.0", "0.0.42.0"},
		{"0.0.45.0", "0.0.50.0"},
	}

	range6range7Remove := [][]string{
		{"0.0.0.0", "0.0.0.255"},
		{"0.0.8.1", "0.0.12.255"},
		{"0.0.21.0", "0.0.26.0"},
		{"0.0.36.0", "0.0.36.255"},
		{"0.0.42.1", "0.0.44.255"},
	}

	range7range6Remove := [][]string{
		{"0.0.3.1", "0.0.4.255"},
		{"0.0.13.1", "0.0.20.0"},
		{"0.0.27.0", "0.0.35.0"},
		{"0.0.40.1", "0.0.41.255"},
		{"0.0.50.1", "0.0.55.0"},
	}

	result8 := t.initIPv4Lists(
		range6, range7, range6range7Intersection, range6range7Union, range6range7Remove, range7range6Remove)

	executeTests(result8)

	range8 := [][]string{
		{"0.0.0.0", "0.0.255.0"},
	}

	range7range8Remove := [][]string{}

	range8range7Remove := [][]string{
		{"0.0.0.0", "0.0.0.255"},
		{"0.0.8.1", "0.0.12.255"},
		{"0.0.20.1", "0.0.26.255"},
		{"0.0.35.1", "0.0.36.255"},
		{"0.0.42.1", "0.0.44.255"},
		{"0.0.55.1", "0.0.255.0"},
	}

	result9 := t.initIPv4Lists(
		range7, range8, range7, range8, range7range8Remove, range8range7Remove)

	executeTests(result9)

	range9 := [][]string{
		{"0.0.0.1", "0.0.254.0"},
	}

	range7range9Remove := empty

	range9range7Remove := [][]string{
		{"0.0.0.1", "0.0.0.255"},
		{"0.0.8.1", "0.0.12.255"},
		{"0.0.20.1", "0.0.26.255"},
		{"0.0.35.1", "0.0.36.255"},
		{"0.0.42.1", "0.0.44.255"},
		{"0.0.55.1", "0.0.254.0"},
	}

	result10 := t.initIPv4Lists(
		range7, range9, range7, range9, range7range9Remove, range9range7Remove)

	executeTests(result10)

	range10 := [][]string{
		{"0.0.0.0", "0.0.0.3"},
		{"0.0.0.5", "0.0.0.13"},
		{"0.0.0.21", "0.0.0.26"},
		{"0.0.0.36", "0.0.0.40"},
		{"0.0.0.42", "0.0.0.50"},
	}

	range11 := [][]string{
		{"0.0.0.1", "0.0.0.8"},
		{"0.0.0.13", "0.0.0.20"},
		{"0.0.0.27", "0.0.0.35"},
		{"0.0.0.37", "0.0.0.42"},
		{"0.0.0.45", "0.0.0.55"},
	}

	range10range11Union := [][]string{
		{"0.0.0.0", "0.0.0.55"},
	}

	range10range11Intersection := [][]string{
		{"0.0.0.1", "0.0.0.3"},
		{"0.0.0.5", "0.0.0.8"},
		{"0.0.0.13", "0.0.0.13"},
		{"0.0.0.37", "0.0.0.40"},
		{"0.0.0.42", "0.0.0.42"},
		{"0.0.0.45", "0.0.0.50"},
	}

	range10range11Remove := [][]string{
		{"0.0.0.0", "0.0.0.0"},
		{"0.0.0.9", "0.0.0.12"},
		{"0.0.0.21", "0.0.0.26"},
		{"0.0.0.36", "0.0.0.36"},
		{"0.0.0.43", "0.0.0.44"},
	}

	range11range10Remove := [][]string{
		{"0.0.0.4", "0.0.0.4"},
		{"0.0.0.14", "0.0.0.20"},
		{"0.0.0.27", "0.0.0.35"},
		{"0.0.0.41", "0.0.0.41"},
		{"0.0.0.51", "0.0.0.55"},
	}

	executeTests(t.initIPv4Lists(
		range10, range11, range10range11Intersection, range10range11Union, range10range11Remove, range11range10Remove))

	range12 := [][]string{
		{"0.0.0.1", "0.0.0.3"},
		{"0.0.0.7", "0.0.0.25"},
	}

	range12range4Union := range12

	range12range4Intersection := range4

	range12range4Remove := [][]string{
		{"0.0.0.10", "0.0.0.11"},
	}

	range4range12Remove := empty

	executeTests(t.initIPv4Lists(
		range4, range12, range12range4Intersection, range12range4Union, range4range12Remove, range12range4Remove))

	executeTests(t.initIPv4Lists(
		[][]string{}, // multi range 1
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range 2
		[][]string{}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{}, // multi 1 remove multi 2
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi 2 remove multi 1
	))

	executeTests(t.initIPv4Lists(
		[][]string{}, // multi range 1
		[][]string{}, //  multi range 2
		[][]string{}, // intersection
		[][]string{}, // union
		[][]string{}, // multi 1 remove multi 2
		[][]string{}, // multi 2 remove multi 1
	))

	singleRange0 := []string{"0.0.0.24", "0.0.0.36"}

	range11single0Intersection := [][]string{
		{"0.0.0.27", "0.0.0.35"},
	}
	range11single0union := [][]string{
		{"0.0.0.1", "0.0.0.8"},
		{"0.0.0.13", "0.0.0.20"},
		{"0.0.0.24", "0.0.0.42"},
		{"0.0.0.45", "0.0.0.55"},
	}
	range11Single0Remove := [][]string{
		{"0.0.0.1", "0.0.0.8"},
		{"0.0.0.13", "0.0.0.20"},
		{"0.0.0.37", "0.0.0.42"},
		{"0.0.0.45", "0.0.0.55"},
	}
	range11Single0ReverseRemove := [][]string{
		{"0.0.0.24", "0.0.0.26"},
		{"0.0.0.36", "0.0.0.36"},
	}
	executeTests(t.initIPv4SingleList(
		range11,
		singleRange0,
		range11single0Intersection,
		range11single0union,
		range11Single0Remove,
		range11Single0ReverseRemove))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.24", "0.0.0.36"}, // single range
		[][]string{
			{"0.0.0.27", "0.0.0.35"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.24", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			{"0.0.0.24", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.20", "0.0.0.36"}, // single range
		[][]string{
			{"0.0.0.20", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.19"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			{"0.0.0.21", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.19", "0.0.0.36"}, // single range
		[][]string{
			{"0.0.0.19", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.18"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			{"0.0.0.21", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.14", "0.0.0.36"}, // single range
		[][]string{
			{"0.0.0.14", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.13"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			{"0.0.0.21", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.13", "0.0.0.36"}, // single range
		[][]string{
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			{"0.0.0.21", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.12", "0.0.0.36"}, // single range
		[][]string{
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.12", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			{"0.0.0.12", "0.0.0.12"},
			{"0.0.0.21", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.24", "0.0.0.37"}, // single range
		[][]string{
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.37"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.24", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.38", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			{"0.0.0.24", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.24", "0.0.0.38"}, // single range
		[][]string{
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.38"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.24", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.39", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			{"0.0.0.24", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.24", "0.0.0.41"}, // single range
		[][]string{
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.41"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.24", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.42", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			{"0.0.0.24", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.24", "0.0.0.42"}, // single range
		[][]string{
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.24", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			{"0.0.0.24", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.24", "0.0.0.43"}, // single range
		[][]string{
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.24", "0.0.0.43"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			{"0.0.0.24", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
			{"0.0.0.43", "0.0.0.43"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.24", "0.0.0.44"}, // single range
		[][]string{
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.24", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			{"0.0.0.24", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
			{"0.0.0.43", "0.0.0.44"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.24", "0.0.0.45"}, // single range
		[][]string{
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.45"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.24", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.46", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			{"0.0.0.24", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
			{"0.0.0.43", "0.0.0.44"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.0", "255.0.0.0"}, // single range
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // intersection
		[][]string{
			{"0.0.0.0", "255.0.0.0"},
		}, // union
		[][]string{}, // multi remove single
		[][]string{
			{"0.0.0.0", "0.0.0.0"},
			{"0.0.0.9", "0.0.0.12"},
			{"0.0.0.21", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
			{"0.0.0.43", "0.0.0.44"},
			{"0.0.0.56", "255.0.0.0"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.1", "255.0.0.0"}, // single range
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "255.0.0.0"},
		}, // union
		[][]string{}, // multi remove single
		[][]string{
			{"0.0.0.9", "0.0.0.12"},
			{"0.0.0.21", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
			{"0.0.0.43", "0.0.0.44"},
			{"0.0.0.56", "255.0.0.0"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.2", "255.0.0.0"}, // single range
		[][]string{
			{"0.0.0.2", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "255.0.0.0"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.1"},
		}, // multi remove single
		[][]string{
			{"0.0.0.9", "0.0.0.12"},
			{"0.0.0.21", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
			{"0.0.0.43", "0.0.0.44"},
			{"0.0.0.56", "255.0.0.0"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.0", "255.255.255.255"}, // single range
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // intersection
		[][]string{
			{"0.0.0.0", "255.255.255.255"},
		}, // union
		[][]string{}, // multi remove single
		[][]string{
			{"0.0.0.0", "0.0.0.0"},
			{"0.0.0.9", "0.0.0.12"},
			{"0.0.0.21", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
			{"0.0.0.43", "0.0.0.44"},
			{"0.0.0.56", "255.255.255.255"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.0", "255.255.255.254"}, // single range
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // intersection
		[][]string{
			{"0.0.0.0", "255.255.255.254"},
		}, // union
		[][]string{}, // multi remove single
		[][]string{
			{"0.0.0.0", "0.0.0.0"},
			{"0.0.0.9", "0.0.0.12"},
			{"0.0.0.21", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
			{"0.0.0.43", "0.0.0.44"},
			{"0.0.0.56", "255.255.255.254"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.1", "255.255.255.254"}, // single range
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "255.255.255.254"},
		}, // union
		[][]string{}, // multi remove single
		[][]string{
			{"0.0.0.9", "0.0.0.12"},
			{"0.0.0.21", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
			{"0.0.0.43", "0.0.0.44"},
			{"0.0.0.56", "255.255.255.254"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.1", "255.255.255.255"}, // single range
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "255.255.255.255"},
		}, // union
		[][]string{}, // multi remove single
		[][]string{
			{"0.0.0.9", "0.0.0.12"},
			{"0.0.0.21", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
			{"0.0.0.43", "0.0.0.44"},
			{"0.0.0.56", "255.255.255.255"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.0", "0.0.0.56"}, // single range
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // intersection
		[][]string{
			{"0.0.0.0", "0.0.0.56"},
		}, // union
		[][]string{}, // multi remove single
		[][]string{
			{"0.0.0.0", "0.0.0.0"},
			{"0.0.0.9", "0.0.0.12"},
			{"0.0.0.21", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
			{"0.0.0.43", "0.0.0.44"},
			{"0.0.0.56", "0.0.0.56"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.1", "0.0.0.56"}, // single range
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.56"},
		}, // union
		[][]string{}, // multi remove single
		[][]string{
			{"0.0.0.9", "0.0.0.12"},
			{"0.0.0.21", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
			{"0.0.0.43", "0.0.0.44"},
			{"0.0.0.56", "0.0.0.56"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.0", "0.0.0.55"}, // single range
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // intersection
		[][]string{
			{"0.0.0.0", "0.0.0.55"},
		}, // union
		[][]string{}, // multi remove single
		[][]string{
			{"0.0.0.0", "0.0.0.0"},
			{"0.0.0.9", "0.0.0.12"},
			{"0.0.0.21", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
			{"0.0.0.43", "0.0.0.44"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.1", "0.0.0.55"}, // single range
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.55"},
		}, // union
		[][]string{}, // multi remove single
		[][]string{
			{"0.0.0.9", "0.0.0.12"},
			{"0.0.0.21", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
			{"0.0.0.43", "0.0.0.44"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.0", "0.0.0.54"}, // single range
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.54"},
		}, // intersection
		[][]string{
			{"0.0.0.0", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.55", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			{"0.0.0.0", "0.0.0.0"},
			{"0.0.0.9", "0.0.0.12"},
			{"0.0.0.21", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
			{"0.0.0.43", "0.0.0.44"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.1", "0.0.0.54"}, // single range
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.54"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.55", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			//{"0.0.0.0", "0.0.0.0"},
			{"0.0.0.9", "0.0.0.12"},
			{"0.0.0.21", "0.0.0.26"},
			{"0.0.0.36", "0.0.0.36"},
			{"0.0.0.43", "0.0.0.44"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.0", "0.0.0.1"}, // single range
		[][]string{
			{"0.0.0.1", "0.0.0.1"},
		}, // intersection
		[][]string{
			{"0.0.0.0", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.2", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			{"0.0.0.0", "0.0.0.0"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.0", "0.0.0.0"}, // single range
		[][]string{
			//{"0.0.0.1", "0.0.0.1"},
		}, // intersection
		[][]string{
			{"0.0.0.0", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			{"0.0.0.0", "0.0.0.0"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.1", "0.0.0.1"}, // single range
		[][]string{
			{"0.0.0.1", "0.0.0.1"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.2", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			//{"0.0.0.0", "0.0.0.0"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.54", "0.0.0.55"}, // single range
		[][]string{
			{"0.0.0.54", "0.0.0.55"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.53"},
		}, // multi remove single
		[][]string{}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.54", "0.0.0.54"}, // single range
		[][]string{
			{"0.0.0.54", "0.0.0.54"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.53"},
			{"0.0.0.55", "0.0.0.55"},
		}, // multi remove single
		[][]string{}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.55", "0.0.0.55"}, // single range
		[][]string{
			{"0.0.0.55", "0.0.0.55"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.54"},
		}, // multi remove single
		[][]string{}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.56", "0.0.0.56"}, // single range
		[][]string{},                     // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.56"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			{"0.0.0.56", "0.0.0.56"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.14", "0.0.0.14"}, // single range
		[][]string{
			{"0.0.0.14", "0.0.0.14"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.13"},
			{"0.0.0.15", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.0", "0.0.0.30"}, // single range
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.30"},
		}, // intersection
		[][]string{
			{"0.0.0.0", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.31", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			{"0.0.0.0", "0.0.0.0"},
			{"0.0.0.9", "0.0.0.12"},
			{"0.0.0.21", "0.0.0.26"},
		}, // single remove multi
	))

	///////////////////////////////////////////////////////

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.1", "0.0.0.30"}, // single range
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.30"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.31", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			//{"0.0.0.0", "0.0.0.0"},
			{"0.0.0.9", "0.0.0.12"},
			{"0.0.0.21", "0.0.0.26"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{
			{"0.0.0.1", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, //  multi range
		[]string{"0.0.0.2", "0.0.0.30"}, // single range
		[][]string{
			{"0.0.0.2", "0.0.0.8"},
			{"0.0.0.13", "0.0.0.20"},
			{"0.0.0.27", "0.0.0.30"},
		}, // intersection
		[][]string{
			{"0.0.0.01", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.1"},
			{"0.0.0.31", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi remove single
		[][]string{
			{"0.0.0.9", "0.0.0.12"},
			{"0.0.0.21", "0.0.0.26"},
		}, // single remove multi
	))

	executeTests(t.initIPv4SingleList(
		[][]string{}, //  multi range
		[]string{
			"0.0.0.2", "0.0.0.30"}, // single range
		[][]string{}, // intersection
		[][]string{
			{"0.0.0.2", "0.0.0.30"},
		}, // union
		[][]string{}, // multi remove single
		[][]string{
			{"0.0.0.2", "0.0.0.30"},
		}, // single remove multi
	))

	executeTests(t.initIPv4Lists(
		[][]string{
			{"0.0.0.1", "0.0.0.4"},
			{"0.0.0.8", "0.0.0.14"},
			{"0.0.0.17", "0.0.0.17"},
			{"0.0.0.19", "0.0.0.19"},
			{"0.0.0.21", "0.0.0.25"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi range 1
		[][]string{
			{"0.0.0.50", "0.0.0.75"},
		}, //  multi range 2
		[][]string{
			{"0.0.0.50", "0.0.0.55"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.4"},
			{"0.0.0.8", "0.0.0.14"},
			{"0.0.0.17", "0.0.0.17"},
			{"0.0.0.19", "0.0.0.19"},
			{"0.0.0.21", "0.0.0.25"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.75"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.4"},
			{"0.0.0.8", "0.0.0.14"},
			{"0.0.0.17", "0.0.0.17"},
			{"0.0.0.19", "0.0.0.19"},
			{"0.0.0.21", "0.0.0.25"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.49"},
		}, // multi 1 remove multi 2
		[][]string{
			{"0.0.0.56", "0.0.0.75"},
		}, // multi 2 remove multi 1
	))

	executeTests(t.initIPv4Lists(
		[][]string{
			{"0.0.0.1", "0.0.0.4"},
			{"0.0.0.8", "0.0.0.14"},
			{"0.0.0.17", "0.0.0.17"},
			{"0.0.0.19", "0.0.0.19"},
			{"0.0.0.21", "0.0.0.25"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi range 1
		[][]string{
			{"0.0.0.80", "0.0.0.255"},
		}, //  multi range 2
		[][]string{}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.4"},
			{"0.0.0.8", "0.0.0.14"},
			{"0.0.0.17", "0.0.0.17"},
			{"0.0.0.19", "0.0.0.19"},
			{"0.0.0.21", "0.0.0.25"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
			{"0.0.0.80", "0.0.0.255"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.4"},
			{"0.0.0.8", "0.0.0.14"},
			{"0.0.0.17", "0.0.0.17"},
			{"0.0.0.19", "0.0.0.19"},
			{"0.0.0.21", "0.0.0.25"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi 1 remove multi 2
		[][]string{
			{"0.0.0.80", "0.0.0.255"},
		}, // multi 2 remove multi 1
	))

	executeTests(t.initIPv4Lists(
		[][]string{
			{"0.0.0.1", "0.0.0.4"},
			{"0.0.0.8", "0.0.0.14"},
			{"0.0.0.17", "0.0.0.17"},
			{"0.0.0.19", "0.0.0.19"},
			{"0.0.0.21", "0.0.0.25"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.55"},
		}, // multi range 1
		[][]string{
			{"0.0.0.50", "0.0.0.75"},
			{"0.0.0.80", "0.0.0.255"},
		}, //  multi range 2
		[][]string{
			{"0.0.0.50", "0.0.0.55"},
		}, // intersection
		[][]string{
			{"0.0.0.1", "0.0.0.4"},
			{"0.0.0.8", "0.0.0.14"},
			{"0.0.0.17", "0.0.0.17"},
			{"0.0.0.19", "0.0.0.19"},
			{"0.0.0.21", "0.0.0.25"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.75"},
			{"0.0.0.80", "0.0.0.255"},
		}, // union
		[][]string{
			{"0.0.0.1", "0.0.0.4"},
			{"0.0.0.8", "0.0.0.14"},
			{"0.0.0.17", "0.0.0.17"},
			{"0.0.0.19", "0.0.0.19"},
			{"0.0.0.21", "0.0.0.25"},
			{"0.0.0.27", "0.0.0.35"},
			{"0.0.0.37", "0.0.0.42"},
			{"0.0.0.45", "0.0.0.49"},
		}, // multi 1 remove multi 2
		[][]string{
			{"0.0.0.56", "0.0.0.75"},
			{"0.0.0.80", "0.0.0.255"},
		}, // multi 2 remove multi 1
	))

	// 		boolean isNoAutoSubnets = prefixConfiguration.prefixedSubnetsAreExplicit();
	// 		boolean allPrefixesAreSubnets = prefixConfiguration.allPrefixedAddressesAreSubnets();

	// 		String address0 = isNoAutoSubnets ? "0.0.0.0-7/29" : "0.0.0.0/29"; //0.0.0.0 to 0.0.0.7

	address0 := "0.0.0.0/29"

	range11addr0Intersection := [][]string{
		{"0.0.0.1", "0.0.0.7"},
	}
	range11addr0union := [][]string{
		{"0.0.0.0", "0.0.0.8"},
		{"0.0.0.13", "0.0.0.20"},
		{"0.0.0.27", "0.0.0.35"},
		{"0.0.0.37", "0.0.0.42"},
		{"0.0.0.45", "0.0.0.55"},
	}
	range11addr0Remove := [][]string{
		{"0.0.0.8", "0.0.0.8"},
		{"0.0.0.13", "0.0.0.20"},
		{"0.0.0.27", "0.0.0.35"},
		{"0.0.0.37", "0.0.0.42"},
		{"0.0.0.45", "0.0.0.55"},
	}
	range11addr0ReverseRemove := [][]string{
		{"0.0.0.0", "0.0.0.0"},
	}

	executeTests(t.initIPv4SingleAddress(
		range11, address0,
		range11addr0Intersection,
		range11addr0union,
		range11addr0Remove,
		range11addr0ReverseRemove))

	// 		String address1 = isNoAutoSubnets ? "0.0.0.8-15/29" : "0.0.0.8/29"; //0.0.0.8 to 0.0.0.15

	address1 := "0.0.0.8-15/29"

	range11addr1Intersection := [][]string{
		{"0.0.0.8", "0.0.0.8"},
		{"0.0.0.13", "0.0.0.15"},
	}
	range11addr1union := [][]string{
		{"0.0.0.1", "0.0.0.20"},
		{"0.0.0.27", "0.0.0.35"},
		{"0.0.0.37", "0.0.0.42"},
		{"0.0.0.45", "0.0.0.55"},
	}
	range11addr1Remove := [][]string{
		{"0.0.0.1", "0.0.0.7"},
		{"0.0.0.16", "0.0.0.20"},
		{"0.0.0.27", "0.0.0.35"},
		{"0.0.0.37", "0.0.0.42"},
		{"0.0.0.45", "0.0.0.55"},
	}
	range11addr1ReverseRemove := [][]string{
		{"0.0.0.9", "0.0.0.12"},
	}

	executeTests(t.initIPv4SingleAddress(
		range11, address1,
		range11addr1Intersection,
		range11addr1union,
		range11addr1Remove,
		range11addr1ReverseRemove))

	address2 := "30-34.2.3-5.4"

	executeTests(t.initIPv4SingleAddress(
		[][]string{
			{"31.2.2.3", "31.2.3.3"},
			{"33.2.5.4", "34.0.0.0"},
		}, //range
		address2,
		[][]string{
			{"33.2.5.4", "33.2.5.4"},
		}, // intersection
		[][]string{
			{"30.2.3.4", "30.2.3.4"},
			{"30.2.4.4", "30.2.4.4"},
			{"30.2.5.4", "30.2.5.4"},

			{"31.2.2.3", "31.2.3.4"},
			{"31.2.4.4", "31.2.4.4"},
			{"31.2.5.4", "31.2.5.4"},

			{"32.2.3.4", "32.2.3.4"},
			{"32.2.4.4", "32.2.4.4"},
			{"32.2.5.4", "32.2.5.4"},

			{"33.2.3.4", "33.2.3.4"},
			{"33.2.4.4", "33.2.4.4"},
			{"33.2.5.4", "34.0.0.0"},

			{"34.2.3.4", "34.2.3.4"},
			{"34.2.4.4", "34.2.4.4"},
			{"34.2.5.4", "34.2.5.4"},
		}, //union
		[][]string{
			{"31.2.2.3", "31.2.3.3"},
			{"33.2.5.5", "34.0.0.0"},
		}, // remove
		[][]string{
			{"30.2.3.4", "30.2.3.4"},
			{"30.2.4.4", "30.2.4.4"},
			{"30.2.5.4", "30.2.5.4"},

			{"31.2.3.4", "31.2.3.4"},
			{"31.2.4.4", "31.2.4.4"},
			{"31.2.5.4", "31.2.5.4"},

			{"32.2.3.4", "32.2.3.4"},
			{"32.2.4.4", "32.2.4.4"},
			{"32.2.5.4", "32.2.5.4"},

			{"33.2.3.4", "33.2.3.4"},
			{"33.2.4.4", "33.2.4.4"},

			{"34.2.3.4", "34.2.3.4"},
			{"34.2.4.4", "34.2.4.4"},
			{"34.2.5.4", "34.2.5.4"},
		}, // reverse remove
	))

	executeTests(t.initIPv4SingleAddress(
		[][]string{
			{"31.2.2.3", "31.2.3.3"},
			{"33.2.5.5", "34.0.0.0"},
		}, //range
		address2,
		[][]string{}, // intersection
		[][]string{
			{"30.2.3.4", "30.2.3.4"},
			{"30.2.4.4", "30.2.4.4"},
			{"30.2.5.4", "30.2.5.4"},

			{"31.2.2.3", "31.2.3.4"},
			{"31.2.4.4", "31.2.4.4"},
			{"31.2.5.4", "31.2.5.4"},

			{"32.2.3.4", "32.2.3.4"},
			{"32.2.4.4", "32.2.4.4"},
			{"32.2.5.4", "32.2.5.4"},

			{"33.2.3.4", "33.2.3.4"},
			{"33.2.4.4", "33.2.4.4"},
			{"33.2.5.4", "34.0.0.0"},

			{"34.2.3.4", "34.2.3.4"},
			{"34.2.4.4", "34.2.4.4"},
			{"34.2.5.4", "34.2.5.4"},
		}, //union
		[][]string{
			{"31.2.2.3", "31.2.3.3"},
			{"33.2.5.5", "34.0.0.0"},
		}, // remove
		[][]string{
			{"30.2.3.4", "30.2.3.4"},
			{"30.2.4.4", "30.2.4.4"},
			{"30.2.5.4", "30.2.5.4"},

			{"31.2.3.4", "31.2.3.4"},
			{"31.2.4.4", "31.2.4.4"},
			{"31.2.5.4", "31.2.5.4"},

			{"32.2.3.4", "32.2.3.4"},
			{"32.2.4.4", "32.2.4.4"},
			{"32.2.5.4", "32.2.5.4"},

			{"33.2.3.4", "33.2.3.4"},
			{"33.2.4.4", "33.2.4.4"},
			{"33.2.5.4", "33.2.5.4"},

			{"34.2.3.4", "34.2.3.4"},
			{"34.2.4.4", "34.2.4.4"},
			{"34.2.5.4", "34.2.5.4"},
		}, // reverse remove
	))

	// 		String address3 = allPrefixesAreSubnets ? "255.3-5.2.0/24" : "255.3-5.2.*/16";

	address3 := "255.3-5.2.*/16"

	executeTests(t.initIPv4SingleAddress(
		[][]string{
			{"255.3.1.0", "255.3.1.255"},
			{"255.5.3.0", "255.6.3.0"},
		}, //range
		address3,
		[][]string{}, // intersection
		[][]string{
			{"255.3.1.0", "255.3.2.255"},
			{"255.4.2.0", "255.4.2.255"},
			{"255.5.2.0", "255.6.3.0"},
		}, //union
		[][]string{
			{"255.3.1.0", "255.3.1.255"},
			{"255.5.3.0", "255.6.3.0"},
		}, // remove
		[][]string{
			{"255.3.2.0", "255.3.2.255"},
			{"255.4.2.0", "255.4.2.255"},
			{"255.5.2.0", "255.5.2.255"},
		}, // reverse remove
	))

	address4 := "200.248-255.200.25-45"
	executeTests(t.initIPv4SingleAddress(
		[][]string{
			{"200.250.200.0", "200.250.200.10"},
			{"200.250.200.20", "200.250.200.30"},
			{"200.250.200.40", "200.252.200.10"},
			{"200.252.200.20", "200.252.200.30"},
			{"200.252.200.40", "200.252.200.50"},
		}, //range
		address4,
		[][]string{
			{"200.250.200.25", "200.250.200.30"},
			{"200.250.200.40", "200.250.200.45"},
			{"200.251.200.25", "200.251.200.45"},
			{"200.252.200.25", "200.252.200.30"},
			{"200.252.200.40", "200.252.200.45"},
		}, // intersection
		[][]string{
			{"200.248.200.25", "200.248.200.45"},
			{"200.249.200.25", "200.249.200.45"},
			{"200.250.200.0", "200.250.200.10"},
			{"200.250.200.20", "200.252.200.10"},
			{"200.252.200.20", "200.252.200.50"},
			{"200.252.200.25", "200.252.200.45"},
			{"200.253.200.25", "200.253.200.45"},
			{"200.254.200.25", "200.254.200.45"},
			{"200.255.200.25", "200.255.200.45"},
		}, //union
		[][]string{
			{"200.250.200.0", "200.250.200.10"},
			{"200.250.200.20", "200.250.200.24"},
			{"200.250.200.46", "200.251.200.24"},
			{"200.250.200.46", "200.251.200.10"},
			{"200.251.200.20", "200.251.200.24"},
			{"200.251.200.46", "200.252.200.10"},
			{"200.252.200.20", "200.252.200.24"},
			{"200.252.200.46", "200.252.200.50"},
		}, // remove
		[][]string{
			{"200.248.200.25", "200.248.200.45"},
			{"200.249.200.25", "200.249.200.45"},
			{"200.250.200.31", "200.250.200.39"},
			{"200.252.200.31", "200.252.200.39"},
			{"200.253.200.25", "200.253.200.45"},
			{"200.254.200.25", "200.254.200.45"},
			{"200.255.200.25", "200.255.200.45"},
		}, // reverse remove
	))

	list1 := &ipaddr.IPAddressSeqRangeList{}
	list2 := &ipaddr.IPAddressSeqRangeList{}
	union := &ipaddr.IPAddressSeqRangeList{}
	last := ipaddr.NewIPAddressString("1.2.3.4").GetAddress()
	for i := 0; i < 200; i++ {
		first := last.Increment(100)
		if i%2 == 0 {
			last = first.Increment(20)
		} else {
			last = first.Increment(21)
		}
		rng := first.SpanWithRange(last)
		list1.AddSeqRange(rng)
		union.AddSeqRange(rng)
		if i%2 == 0 {
			first = last.Increment(11)
		} else {
			first = last.Increment(10)
		}
		last = first.Increment(11)
		rng = first.SpanWithRange(last)
		list2.AddSeqRange(rng)
		union.AddSeqRange(rng)
	}

	executeTests(t.initIPv4ListTests(
		list1,                           // range1
		list2,                           // range2
		&ipaddr.IPAddressSeqRangeList{}, // intersection
		union,                           // union
		list1,                           // range1 remove range2
		list2,                           // range2 remove range1
	))

	list1.Clear()
	list2.Clear()
	union.Clear()
	intersection := &ipaddr.IPAddressSeqRangeList{}
	range1Remove2 := &ipaddr.IPAddressSeqRangeList{}
	range2Remove1 := &ipaddr.IPAddressSeqRangeList{}
	last = ipaddr.NewIPAddressString("1.2.3.4").GetAddress()
	for i := 0; i < 200; i++ {
		first := last.Increment(100)
		if i%2 == 0 {
			last = first.Increment(20)
		} else {
			last = first.Increment(21)
		}
		rng := first.SpanWithRange(last)
		list1.AddSeqRange(rng)

		var otherFirst, otherLast *ipaddr.IPAddress
		if i%2 == 1 {
			otherFirst = first.Increment(10)
		} else {
			otherFirst = first.Increment(11)
		}
		if i%2 == 0 {
			otherLast = otherFirst.Increment(20)
		} else {
			otherLast = otherFirst.Increment(21)
		}
		list2.AddSeqRange(otherFirst.SpanWithRange(otherLast))
		union.AddSeqRange(first.SpanWithRange(otherLast))
		intersection.AddSeqRange(otherFirst.SpanWithRange(last))
		range1Remove2.AddSeqRange(first.SpanWithRange(otherFirst.DecrementSingle()))
		range2Remove1.AddSeqRange(last.IncrementSingle().SpanWithRange(otherLast))
	}

	executeTests(t.initIPv4ListTests(
		list1,         // range1
		list2,         // range2
		intersection,  // intersection
		union,         // union
		range1Remove2, // range1 remove range2
		range2Remove1, // range2 remove range1
	))

	list1.Clear()
	list2.Clear()
	union.Clear()
	intersection.Clear()
	range1Remove2.Clear()
	range2Remove1.Clear()

	last = ipaddr.NewIPAddressString("2.2.3.4").GetAddress()
	for i := 0; i < 200; i++ {
		first := last.Increment(100)
		if i%2 == 0 {
			last = first.Increment(20)
		} else {
			last = first.Increment(21)
		}
		rng := first.SpanWithRange(last)
		list1.AddSeqRange(rng)

		var otherFirst, otherLast *ipaddr.IPAddress
		if i%2 == 1 {
			otherFirst = last.IncrementSingle()
		} else {
			otherFirst = last.Increment(2)
		}
		if i%2 == 0 {
			otherLast = otherFirst.Increment(20)
		} else {
			otherLast = otherFirst.Increment(21)
		}
		list2.AddSeqRange(otherFirst.SpanWithRange(otherLast))

		if i%2 == 1 {
			union.AddSeqRange(first.SpanWithRange(otherLast))
		} else {
			union.AddSeqRange(first.SpanWithRange(last))
			union.AddSeqRange(otherFirst.SpanWithRange(otherLast))
		}
	}

	executeTests(t.initIPv4ListTests(
		list1,                           // range1
		list2,                           // range2
		&ipaddr.IPAddressSeqRangeList{}, // intersection
		union,                           // union
		list1,                           // range1 remove range2
		list2,                           // range2 remove range1
	))

	list1.Clear()
	list2.Clear()
	union.Clear()

	last = ipaddr.NewIPAddressString("100.2.3.4").GetAddress()
	for i := 0; i < 200; i++ {
		first := last.Increment(100)
		last = first.Increment(19)
		rng := first.SpanWithRange(last)
		list1.AddSeqRange(rng)

		var otherFirst *ipaddr.IPAddress
		if i%2 == 0 {
			otherFirst = first
		} else {
			range1Remove2.AddSeqRange(first.CoverWithSequentialRange())
			otherFirst = first.IncrementSingle()
		}
		for j := 0; j < 10; j++ {
			list2.AddSeqRange(otherFirst.CoverWithSequentialRange())
			if i%2 == 0 || j < 9 {
				otherFirst = otherFirst.IncrementSingle()
				range1Remove2.AddSeqRange(otherFirst.CoverWithSequentialRange())
				otherFirst = otherFirst.IncrementSingle()
			}
		}
	}

	executeTests(t.initIPv4ListTests(
		list1,                           // range1
		list2,                           // range2
		list2,                           // intersection
		list1,                           // union
		range1Remove2,                   // range1 remove range2
		&ipaddr.IPAddressSeqRangeList{}, // range2 remove range1
	))

	list1.Clear()
	list2.Clear()
	range1Remove2.Clear()
	last = ipaddr.NewIPAddressString("2.255.3.4").GetAddress()
	for i := 0; i < 201; i++ {
		first := last.Increment(99)
		if i%2 == 0 {
			last = first.Increment(7)
		} else {
			last = first.Increment(9)
		}
		rng := first.SpanWithRange(last)
		list1.AddSeqRange(rng)

		if i == 198 {
			otherFirst := first.IncrementSingle()
			otherLast := last.DecrementSingle()
			rng2 := otherFirst.SpanWithRange(otherLast)
			list2.AddSeqRange(rng2)
			intersection.AddSeqRange(rng2)
			range1Remove2.AddSeqRange(first.CoverWithSequentialRange())
			range1Remove2.AddSeqRange(last.CoverWithSequentialRange())
		} else {
			range1Remove2.AddSeqRange(rng)
		}
	}
	executeTests(t.initIPv4ListTests(
		list1,                           // range1
		list2,                           // range2
		intersection,                    // intersection
		list1,                           // union
		range1Remove2,                   // range1 remove range2
		&ipaddr.IPAddressSeqRangeList{}, // range2 remove range1
	))

	list1.Clear()
	list2.Clear()
	range1Remove2.Clear()
	intersection.Clear()
	last = ipaddr.NewIPAddressString("2.255.3.4").GetAddress()
	for i := 0; i < 201; i++ {
		first := last.Increment(99)
		if i%2 == 0 {
			last = first.Increment(7)
		} else {
			last = first.Increment(9)
		}
		rng := first.SpanWithRange(last)
		list1.AddSeqRange(rng)

		if i == 197 {
			otherFirst := last
			otherLast := last
			rng2 := otherFirst.SpanWithRange(otherLast)
			list2.AddSeqRange(rng2)
			intersection.AddSeqRange(rng2)
			range1Remove2.AddSeqRange(first.SpanWithRange(otherLast.DecrementSingle()))
		} else {
			range1Remove2.AddSeqRange(rng)
		}
	}
	executeTests(t.initIPv4ListTests(
		list1,                           // range1
		list2,                           // range2
		intersection,                    // intersection
		list1,                           // union
		range1Remove2,                   // range1 remove range2
		&ipaddr.IPAddressSeqRangeList{}, // range2 remove range1
	))

	t.testRangeListIncrement()

	if printResults {
		fmt.Println()
		fmt.Println("IPAddressSeqRangeList multi list tests:", t.multiListTestCount)
		fmt.Println("IPAddressSeqRangeList single list tests:", t.singleListTestCount)
		fmt.Println("IPAddressSeqRangeList address tests:", t.addressTestCount)
		fmt.Println("IPAddressSeqRangeList tests:", t.rangeListTestCount)
		fmt.Println("IPAddressSeqRangeList failures:", t.rangeListFailCount)
		fmt.Println()
	}

	t.incrementTestCountNum(uint64(t.rangeListTestCount))
}

func (t *collectionTester) initIPv4Lists(
	range1Strs,
	range2Strs,
	intersectionStrs,
	unionStrs,
	range1RemoveRange2Strs,
	range2RemoveRange1Strs [][]string) []TestResult {
	range1 := t.create(range1Strs)
	range2 := t.create(range2Strs)
	expectedIntersection := t.create(intersectionStrs)
	expectedUnion := t.create(unionStrs)
	range1RemoveRange2 := t.create(range1RemoveRange2Strs)
	range2RemoveRange1 := t.create(range2RemoveRange1Strs)
	return t.initIPv4ListTests(range1, range2, expectedIntersection, expectedUnion, range1RemoveRange2, range2RemoveRange1)
}

func (t *collectionTester) create(rngList [][]string) *ipaddr.IPAddressSeqRangeList {
	list := ipaddr.NewSequentialRangeList[*ipaddr.IPAddress](len(rngList) << 1)
	for _, rngStrs := range rngList {
		rng := t.createRange(rngStrs)
		list.AddSeqRange(rng)
	}
	return list
}

func (t *collectionTester) createRange(rngStrs []string) *ipaddr.IPAddressSeqRange {
	lower, upper := t.createAddr(rngStrs[0]), t.createAddr(rngStrs[1])
	rng := lower.SpanWithRange(upper)
	return rng
}

func (t *collectionTester) createAddr(addrStr string) *ipaddr.IPAddress {
	return t.createAddress(addrStr).GetAddress()
}

func (t *collectionTester) initIPv4ListTests(
	range1,
	range2,
	expectedIntersection,
	expectedUnion,
	range1RemoveRange2,
	range2RemoveRange1 *ipaddr.IPAddressSeqRangeList) []TestResult {

	isIPv6 := false

	var tests []TestResult

	tests = append(tests, newRangeResult(t, isIPv6, range1, range2, expectedIntersection, expectedUnion, range1RemoveRange2, range2RemoveRange1))
	if range1.GetSeqRangeCount() == 1 {
		rng := range1.GetLowerSeqRange()
		tests = append(tests, newSingleRangeResult(t, isIPv6, range2, rng, expectedIntersection, expectedUnion, range2RemoveRange1))
		blocks := rng.SpanWithSequentialBlocks()
		if len(blocks) == 1 {
			tests = append(tests, newAddressResult(t, isIPv6, range2, blocks[0], expectedIntersection, expectedUnion, range2RemoveRange1))
		}
	}
	if range2.GetSeqRangeCount() == 1 {
		rng := range2.GetLowerSeqRange()
		//rngTrie := createContainmentTrieRange(rng)
		tests = append(tests, newSingleRangeResult(t, isIPv6, range1, rng, expectedIntersection, expectedUnion, range1RemoveRange2))
		blocks := rng.SpanWithSequentialBlocks()
		if len(blocks) == 1 { // [0.0.0-254.*, 0.0.255.0]
			tests = append(tests, newAddressResult(t, isIPv6, range1, blocks[0], expectedIntersection, expectedUnion, range1RemoveRange2))
		}
	}

	if range1.GetSeqRangeCount() == 1 {
		rng := range1.GetLowerSeqRange()
		tests = append(tests, newSingleRangeResult(t, isIPv6, range2, rng, expectedIntersection, expectedUnion, range2RemoveRange1))
		blocks := rng.SpanWithSequentialBlocks()
		if len(blocks) == 1 {
			tests = append(tests, newAddressResult(t, isIPv6, range2, blocks[0], expectedIntersection, expectedUnion, range2RemoveRange1))
		}
	}
	if range2.GetSeqRangeCount() == 1 {
		rng := range2.GetLowerSeqRange()
		tests = append(tests, newSingleRangeResult(t, isIPv6, range1, rng, expectedIntersection, expectedUnion, range1RemoveRange2))
		blocks := rng.SpanWithSequentialBlocks()
		if len(blocks) == 1 { // [0.0.0-254.*, 0.0.255.0]
			tests = append(tests, newAddressResult(t, isIPv6, range1, blocks[0], expectedIntersection, expectedUnion, range1RemoveRange2))
		}
	}

	isIPv6 = true

	range1IPv6 := convert(range1)
	range2IPv6 := convert(range2)
	expectedIntersectionIPv6 := convert(expectedIntersection)
	expectedUnionIPv6 := convert(expectedUnion)
	range1RemoveRange2IPv6 := convert(range1RemoveRange2)
	range2RemoveRange1IPv6 := convert(range2RemoveRange1)

	tests = append(tests, newRangeResult(t, isIPv6, range1IPv6, range2IPv6, expectedIntersectionIPv6, expectedUnionIPv6, range1RemoveRange2IPv6, range2RemoveRange1IPv6))

	if range1IPv6.GetSeqRangeCount() == 1 {
		rng := range1IPv6.GetLowerSeqRange()
		tests = append(tests, newSingleRangeResult(t, isIPv6, range2IPv6, rng /*range2TrieIPv6, rngTrieIPv6,*/, expectedIntersectionIPv6, expectedUnionIPv6, range2RemoveRange1IPv6))
		blocks := rng.SpanWithSequentialBlocks()
		if len(blocks) == 1 {
			tests = append(tests, newAddressResult(t, isIPv6, range2IPv6, blocks[0] /*range2TrieIPv6, rngTrieIPv6,*/, expectedIntersectionIPv6, expectedUnionIPv6, range2RemoveRange1IPv6))
		}
	}
	if range2IPv6.GetSeqRangeCount() == 1 {
		rng := range2IPv6.GetLowerSeqRange()
		tests = append(tests, newSingleRangeResult(t, isIPv6, range1IPv6, rng /*range1TrieIPv6, rngTrieIPv6,*/, expectedIntersectionIPv6, expectedUnionIPv6, range1RemoveRange2IPv6))
		blocks := rng.SpanWithSequentialBlocks()
		if len(blocks) == 1 {
			tests = append(tests, newAddressResult(t, isIPv6, range1IPv6, blocks[0] /*range1TrieIPv6, rngTrieIPv6,*/, expectedIntersectionIPv6, expectedUnionIPv6, range1RemoveRange2IPv6))
		}
	}

	if range1IPv6.GetSeqRangeCount() == 1 {
		rng := range1IPv6.GetLowerSeqRange()
		tests = append(tests, newSingleRangeResult(t, isIPv6, range2IPv6, rng /*range2TrieIPv6, rngTrieIPv6,*/, expectedIntersectionIPv6, expectedUnionIPv6, range2RemoveRange1IPv6))
		blocks := rng.SpanWithSequentialBlocks()
		if len(blocks) == 1 {
			tests = append(tests, newAddressResult(t, isIPv6, range2IPv6, blocks[0] /*range2TrieIPv6, rngTrieIPv6,*/, expectedIntersectionIPv6, expectedUnionIPv6, range2RemoveRange1IPv6))
		}
	}
	if range2IPv6.GetSeqRangeCount() == 1 {
		rng := range2IPv6.GetLowerSeqRange()
		tests = append(tests, newSingleRangeResult(t, isIPv6, range1IPv6, rng /*range1TrieIPv6, rngTrieIPv6,*/, expectedIntersectionIPv6, expectedUnionIPv6, range1RemoveRange2IPv6))
		blocks := rng.SpanWithSequentialBlocks()
		if len(blocks) == 1 {
			tests = append(tests, newAddressResult(t, isIPv6, range1IPv6, blocks[0] /*range1TrieIPv6, rngTrieIPv6,*/, expectedIntersectionIPv6, expectedUnionIPv6, range1RemoveRange2IPv6))
		}
	}

	return tests
}

func (t *collectionTester) initIPv4SingleList(
	rangeStrs [][]string,
	rangeStr []string,
	intersectionStrs,
	unionStrs,
	removeStrs,
	reverseRemoveStrs [][]string) []TestResult {

	var tests []TestResult

	list := t.create(rangeStrs)
	rng := t.createRange(rangeStr).ToIP()
	expectedIntersection := t.create(intersectionStrs)
	expectedUnion := t.create(unionStrs)
	expectedRemove := t.create(removeStrs)
	expectedReverseRemove := t.create(reverseRemoveStrs)

	isIPv6 := false
	tests = append(tests, newSingleRangeResult(t, isIPv6, list, rng /*listTrie, rngTrie,*/, expectedIntersection, expectedUnion, expectedRemove))
	tests = append(tests, newRangeResult(t, isIPv6, list, inList(rng), expectedIntersection, expectedUnion, expectedRemove, expectedReverseRemove))
	blocks := rng.SpanWithSequentialBlocks()
	if len(blocks) == 1 {
		tests = append(tests, newAddressResult(t, isIPv6, list, blocks[0] /*listTrie, rngTrie,*/, expectedIntersection, expectedUnion, expectedRemove))
	}

	// convert to the more specific range list type IPv4SequentialRangeList

	// IPv6 (IPv4-mapped)

	isIPv6 = true

	listIPv6 := convert(list)
	rngIPv6 := convertRange(rng.ToIPv4()).ToIP()
	expectedIntersectionIPv6 := convert(expectedIntersection)
	expectedUnionIPv6 := convert(expectedUnion)
	expectedRemoveIPv6 := convert(expectedRemove)
	expectedReverseRemoveIPv6 := convert(expectedReverseRemove)

	tests = append(tests, newSingleRangeResult(t, isIPv6, listIPv6, rngIPv6 /*listTrieIPv6, rngTrieIPv6,*/, expectedIntersectionIPv6, expectedUnionIPv6, expectedRemoveIPv6))
	tests = append(tests, newRangeResult(t, isIPv6, listIPv6, inList(rngIPv6), expectedIntersectionIPv6, expectedUnionIPv6, expectedRemoveIPv6, expectedReverseRemoveIPv6))
	blocks = rngIPv6.SpanWithSequentialBlocks()
	if len(blocks) == 1 {
		tests = append(tests, newAddressResult(t, isIPv6, listIPv6, blocks[0] /*listTrieIPv6, rngTrieIPv6,*/, expectedIntersectionIPv6, expectedUnionIPv6, expectedRemoveIPv6))
	}

	// convert to the more specific range list type IPv6SequentialRangeList

	return tests
}

func (t *collectionTester) initIPv4SingleAddress(
	rangeStrs [][]string,
	addressStr string,
	intersectionStrs,
	unionStrs,
	removeStrs,
	reverseRemoveStrs [][]string) []TestResult {

	isIPv6 := false

	var tests []TestResult

	list := t.create(rangeStrs)
	addr := t.createAddr(addressStr)
	expectedIntersection := t.create(intersectionStrs)
	expectedUnion := t.create(unionStrs)
	expectedRemove := t.create(removeStrs)
	expectedReverseRemove := t.create(reverseRemoveStrs)

	tests = append(tests, newAddressResult(t, isIPv6, list, addr /* listTrie, addrTrie,*/, expectedIntersection, expectedUnion, expectedRemove))
	if addr.IsSequential() {
		tests = append(tests, newSingleRangeResult(t, isIPv6, list, addr.CoverWithSequentialRange() /*listTrie, addrTrie,*/, expectedIntersection, expectedUnion, expectedRemove))
		tests = append(tests, newRangeResult(t, isIPv6, list, inList(addr.CoverWithSequentialRange()), expectedIntersection, expectedUnion, expectedRemove, expectedReverseRemove))
	} else {
		tests = append(tests, newRangeResult(t, isIPv6, list, addressInList(addr), expectedIntersection, expectedUnion, expectedRemove, expectedReverseRemove))
	}

	isIPv6 = true

	addrIPv6 := convertAddress(addr.ToIPv4()).ToIP()

	if addrIPv6 != nil {
		listIPv6 := convert(list)
		expectedIntersectionIPv6 := convert(expectedIntersection)
		expectedUnionIPv6 := convert(expectedUnion)
		expectedRemoveIPv6 := convert(expectedRemove)
		expectedReverseRemoveIPv6 := convert(expectedReverseRemove)

		tests = append(tests, newAddressResult(t, isIPv6, listIPv6, addrIPv6, expectedIntersectionIPv6, expectedUnionIPv6, expectedRemoveIPv6))
		if addrIPv6.IsSequential() {
			tests = append(tests, newSingleRangeResult(t, isIPv6, listIPv6, addrIPv6.CoverWithSequentialRange(), expectedIntersectionIPv6, expectedUnionIPv6, expectedRemoveIPv6))
			tests = append(tests, newRangeResult(t, isIPv6, listIPv6, inList(addrIPv6.CoverWithSequentialRange()), expectedIntersectionIPv6, expectedUnionIPv6, expectedRemoveIPv6, expectedReverseRemoveIPv6))
		} else {
			tests = append(tests, newRangeResult(t, isIPv6, listIPv6, addressInList(addrIPv6), expectedIntersectionIPv6, expectedUnionIPv6, expectedRemoveIPv6, expectedReverseRemoveIPv6))
		}
	}
	return tests
}

func convert(ipv4List *ipaddr.IPAddressSeqRangeList) *ipaddr.IPAddressSeqRangeList {
	ipv6List := ipaddr.NewSequentialRangeList[*ipaddr.IPAddress](ipv4List.GetSeqRangeCount())
	iter := ipv4List.SeqRangeIterator()
	for iter.HasNext() {
		converted := convertRange(iter.Next().ToIPv4())
		ipv6List.AddSeqRange(converted.ToIP())
	}
	return ipv6List
}

func convertRange(ipv4Range *ipaddr.IPv4AddressSeqRange) *ipaddr.IPv6AddressSeqRange {
	return convertAddress(ipv4Range.GetLower()).SpanWithRange(convertAddress(ipv4Range.GetUpper()))
}

func convertAddress(addr *ipaddr.IPv4Address) *ipaddr.IPv6Address {
	ipv6Addr, _ := addr.GetIPv4MappedAddress()
	return ipv6Addr
}
