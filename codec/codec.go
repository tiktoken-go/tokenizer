package codec

import (
	"container/heap"
	"fmt"
	"math"

	"github.com/dlclark/regexp2/v2"
)

type Codec struct {
	vocabulary        vocab
	reverseVocabulary reverse
	specialTokens     map[string]uint
	splitRegexp       *regexp2.Regexp
	name              string
}

func (c *Codec) GetName() string {
	return c.name
}

// Count returns the number of tokens in the input string.
func (c *Codec) Count(input string) (int, error) {
	var count int

	err := c.tokenize(input, func(_ uint, _ string) {
		count++
	})

	return count, err
}

// Encode returns the token IDs and tokens for the input string.
func (c *Codec) Encode(input string) ([]uint, []string, error) {

	var ids []uint
	var tokens []string

	err := c.tokenize(input, func(id uint, token string) {
		ids = append(ids, id)
		tokens = append(tokens, token)
	})

	return ids, tokens, err
}

func (c *Codec) tokenize(input string, yield func(uint, string)) error {
	match, err := c.splitRegexp.FindStringMatch(input)
	if err != nil {
		return fmt.Errorf("error matching: %v", err)
	}
	for match != nil {
		piece := match.String()
		if id, ok := c.vocabulary[piece]; ok {
			yield(id, piece)
		} else {
			parts := c.mergePairs(piece)

			for i := range len(parts) - 1 {
				token := piece[parts[i].offset:parts[i+1].offset]
				yield(c.vocabulary[token], token)
			}
		}
		match, err = c.splitRegexp.FindNextMatch(match)
		if err != nil {
			return fmt.Errorf("error matching: %v", err)
		}
	}

	return nil
}

func (c *Codec) Decode(tokens []uint) (string, error) {
	if c.reverseVocabulary == nil {
		c.reverseVocabulary = make(map[uint]string)
		for k, v := range c.vocabulary {
			c.reverseVocabulary[v] = k
		}
	}

	var out string
	for _, t := range tokens {
		piece, ok := c.reverseVocabulary[t]
		if !ok {
			return "", fmt.Errorf("invalid token: %d", t)
		}
		out += piece
	}
	return out, nil
}

type part struct {
	offset int
	rank   uint
}

// heapMergeThreshold is the piece length, in bytes, from which mergePairs
// switches from mergePairsScan to mergePairsHeap. Both return the same parts.
// On o200k Chinese (BenchmarkMergePairs) the scan was still faster at 126
// bytes, the heap twice as fast at 255, 4.5 times at 1,023, and 25 times at
// 8,190; the threshold sits between the first two.
const heapMergeThreshold = 192

// mergePairs splits piece into the parts byte-pair encoding merges it into.
func (c *Codec) mergePairs(piece string) []part {
	if len(piece) < heapMergeThreshold {
		return c.mergePairsScan(piece)
	}

	return c.mergePairsHeap(piece)
}

// mergePairsScan repeatedly merges the leftmost pair of parts with the lowest
// rank, finding it with a linear scan. Each merge costs O(n), so a piece costs
// O(n²): a piece of 60,000 bytes, such as Chinese prose without punctuation,
// takes seconds.
func (c *Codec) mergePairsScan(piece string) []part {
	parts := make([]part, len(piece)+1)
	for i := range len(parts) {
		parts[i] = part{i, math.MaxUint}
	}

	getRank := func(index, skip int) uint {
		if index+skip+2 < len(parts) {
			start := parts[index].offset
			end := parts[index+skip+2].offset
			if rank, ok := c.vocabulary[piece[start:end]]; ok {
				return rank
			}
		}
		return math.MaxUint
	}

	for i := 0; i < len(parts)-2; i++ {
		parts[i].rank = getRank(i, 0)
	}

	for {
		if len(parts) == 1 {
			break
		}

		minRank := uint(math.MaxUint)
		minIndex := 0
		for i, p := range parts[:len(parts)-1] {
			if p.rank < minRank {
				minRank = p.rank
				minIndex = i
			}
		}

		if minRank == math.MaxUint {
			break
		}

		parts[minIndex].rank = getRank(minIndex, 1)

		if minIndex > 0 {
			parts[minIndex-1].rank = getRank(minIndex-1, 1)
		}

		parts = append(parts[:minIndex+1], parts[minIndex+2:]...)
	}

	return parts
}

// mergePairsHeap returns the same parts as mergePairsScan in O(n log n).
//
// The part boundaries 0..len(piece) form a doubly linked list. A boundary i
// has a rank when the boundary after its successor exists: the rank of the
// token piece[i:next(next(i))] that merging i with its successor would form.
// mergePairsScan merges the leftmost boundary with the lowest rank; here every
// ranked boundary sits in a min-heap ordered by rank, then offset, which is the
// same choice. Merging i removes next(i), and that changes the rank of i and of
// prev(i) only, so those two are pushed again. A popped entry whose boundary
// was removed, or whose rank is no longer the boundary's current one, is stale
// and skipped. It cannot match by accident: a recomputed rank always belongs to
// a longer token than before, and every token has its own rank.
func (c *Codec) mergePairsHeap(piece string) []part {
	n := len(piece)
	none := n + 1

	next := make([]int, n+1)
	prev := make([]int, n+1)
	ranks := make([]uint, n+1)
	removed := make([]bool, n+1)

	for i := range n + 1 {
		next[i] = i + 1
		prev[i] = i - 1
		ranks[i] = math.MaxUint
	}

	rankOf := func(i int) uint {
		j := next[i]
		if j == none {
			return math.MaxUint
		}

		k := next[j]
		if k == none {
			return math.MaxUint
		}

		if rank, ok := c.vocabulary[piece[i:k]]; ok {
			return rank
		}

		return math.MaxUint
	}

	candidates := make(mergeHeap, 0, n)
	for i := 0; i+2 <= n; i++ {
		ranks[i] = rankOf(i)
		if ranks[i] != math.MaxUint {
			candidates = append(candidates, mergeCandidate{rank: ranks[i], offset: i})
		}
	}

	heap.Init(&candidates)

	for candidates.Len() > 0 {
		top := heap.Pop(&candidates).(mergeCandidate)

		i := top.offset
		if removed[i] || ranks[i] != top.rank {
			continue
		}

		j := next[i]
		removed[j] = true
		next[i] = next[j]

		if next[j] != none {
			prev[next[j]] = i
		}

		ranks[i] = rankOf(i)
		if ranks[i] != math.MaxUint {
			heap.Push(&candidates, mergeCandidate{rank: ranks[i], offset: i})
		}

		if p := prev[i]; p >= 0 {
			ranks[p] = rankOf(p)
			if ranks[p] != math.MaxUint {
				heap.Push(&candidates, mergeCandidate{rank: ranks[p], offset: p})
			}
		}
	}

	parts := make([]part, 0, n+1)
	for i := 0; i != none; i = next[i] {
		parts = append(parts, part{offset: i, rank: ranks[i]})
	}

	return parts
}

type mergeCandidate struct {
	rank   uint
	offset int
}

// mergeHeap orders candidates by rank, then by offset, so that among equal
// ranks the leftmost pair merges first, as in mergePairsScan.
type mergeHeap []mergeCandidate

func (h mergeHeap) Len() int { return len(h) }

func (h mergeHeap) Less(a, b int) bool {
	if h[a].rank != h[b].rank {
		return h[a].rank < h[b].rank
	}

	return h[a].offset < h[b].offset
}

func (h mergeHeap) Swap(a, b int) { h[a], h[b] = h[b], h[a] }

func (h *mergeHeap) Push(x any) { *h = append(*h, x.(mergeCandidate)) }

func (h *mergeHeap) Pop() any {
	old := *h
	last := old[len(old)-1]
	*h = old[:len(old)-1]

	return last
}
