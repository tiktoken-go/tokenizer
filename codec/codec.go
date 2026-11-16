package codec

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/dlclark/regexp2/v2"
)

// ErrorMode controls how Encode, Count and Decode handle invalid UTF-8. See
// the tokenizer package for a description of each mode.
type ErrorMode int

const (
	ErrorModeNone    ErrorMode = iota // return bytes unchanged (default)
	ErrorModeReplace                  // errors="replace"
	ErrorModeIgnore                   // errors="ignore"
	ErrorModeStrict                   // errors="strict"
)

// ErrInvalidUTF8 is returned in ErrorModeStrict when the input to Encode or
// Count, or the output of Decode, is not valid UTF-8.
var ErrInvalidUTF8 = errors.New("invalid UTF-8")

type Codec struct {
	vocabulary            vocab
	reverseVocabulary     reverse
	reverseVocabularyOnce sync.Once
	specialTokens         map[string]uint
	splitRegexp           *regexp2.Regexp
	name                  string
	errorMode             ErrorMode
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

// Decode returns the string for the input token IDs
func (c *Codec) Decode(tokens []uint) (string, error) {
	c.reverseVocabularyOnce.Do(func() {
		var maxID uint
		for _, v := range c.vocabulary {
			maxID = max(maxID, v)
		}
		c.reverseVocabulary = make(reverse, maxID+1)
		for k, v := range c.vocabulary {
			c.reverseVocabulary[v] = k
		}
	})

	var sb strings.Builder
	for _, t := range tokens {
		var piece string
		if t < uint(len(c.reverseVocabulary)) {
			piece = c.reverseVocabulary[t]
		}
		if piece == "" {
			return "", fmt.Errorf("invalid token: %d", t)
		}
		sb.WriteString(piece)
	}
	out := sb.String()

	switch c.errorMode {
	case ErrorModeReplace:
		return replaceUTF8(out), nil
	case ErrorModeIgnore:
		return strings.ToValidUTF8(out, ""), nil
	case ErrorModeStrict:
		if err := validateUTF8(out); err != nil {
			return "", err
		}
		return out, nil
	default:
		return out, nil
	}
}

// ErrorMode returns the codec's current error mode.
func (c *Codec) ErrorMode() ErrorMode {
	return c.errorMode
}

// SetErrorMode sets the codec's behavior for invalid UTF-8 sequences.
func (c *Codec) SetErrorMode(mode ErrorMode) {
	c.errorMode = mode
}

func (c *Codec) tokenize(input string, yield func(uint, string)) error {
	switch c.errorMode {
	case ErrorModeReplace:
		input = replaceUTF8(input)
	case ErrorModeIgnore:
		input = strings.ToValidUTF8(input, "")
	case ErrorModeStrict:
		if err := validateUTF8(input); err != nil {
			return err
		}
	}

	match, err := c.splitRegexp.FindStringMatch(input)
	if err != nil {
		return fmt.Errorf("error matching: %v", err)
	}
	for match != nil {
		piece := match.String()
		if id, ok := c.vocabulary[piece]; ok {
			yield(id, piece)
		} else {
			c.mergePairs(piece, yield)
		}
		match, err = c.splitRegexp.FindNextMatch(match)
		if err != nil {
			return fmt.Errorf("error matching: %v", err)
		}
	}

	return nil
}

func validateUTF8(s string) error {
	if utf8.ValidString(s) {
		return nil
	}
	for i := 0; i < len(s); {
		r, n := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && n == 1 {
			return fmt.Errorf("%w at byte %d", ErrInvalidUTF8, i)
		}
		i += n
	}
	return nil
}

func replaceUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	var sb strings.Builder
	sb.Grow(len(s))
	for len(s) > 0 {
		r, n := utf8.DecodeRuneInString(s)
		if r == utf8.RuneError && n == 1 {
			sb.WriteRune(utf8.RuneError)
			n = maximalSubpart(s)
		} else {
			sb.WriteString(s[:n])
		}
		s = s[n:]
	}
	return sb.String()
}

func maximalSubpart(s string) int {
	need, lo, hi := 0, byte(0x80), byte(0xBF)
	switch c := s[0]; {
	case c >= 0xC2 && c <= 0xDF:
		need = 1
	case c == 0xE0:
		need, lo = 2, 0xA0
	case c == 0xED:
		need, hi = 2, 0x9F
	case c >= 0xE1 && c <= 0xEF:
		need = 2
	case c == 0xF0:
		need, lo = 3, 0x90
	case c >= 0xF1 && c <= 0xF3:
		need = 3
	case c == 0xF4:
		need, hi = 3, 0x8F
	default:
		return 1
	}
	i := 1
	for ; i <= need && i < len(s); i++ {
		c := s[i]
		if i == 1 && (c < lo || c > hi) || i > 1 && (c < 0x80 || c > 0xBF) {
			break
		}
	}
	return i
}

type part struct {
	offset int
	rank   uint
}

// mergeHeapThreshold is the piece length, in bytes, from which mergePairs
// switches from mergePairsScan to mergePairsHeap.
const mergeHeapThreshold = 96

// mergePairs yields the tokens byte-pair encoding merges piece into.
func (c *Codec) mergePairs(piece string, yield func(uint, string)) {
	if len(piece) >= mergeHeapThreshold {
		c.mergePairsHeap(piece, yield)
		return
	}

	parts := c.mergePairsScan(piece)
	for i := range len(parts) - 1 {
		token := piece[parts[i].offset:parts[i+1].offset]
		yield(c.vocabulary[token], token)
	}
}

// mergePairsScan repeatedly merges the leftmost pair of parts with the lowest
// rank, finding it with a linear scan. Each merge costs O(n), so a piece costs
// O(n^2).
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

// noRank marks a boundary that can't merge with its successor.
const noRank = math.MaxUint32

// noSlot marks a boundary that has no candidate in the heap.
const noSlot = -1

// mergeNode is a part boundary in mergePairsHeap's doubly linked list. slot is
// where the heap keeps the candidate that merges this part with the next one.
//
// candidate belongs to the heap, not to the boundary: nodes[k].candidate is the
// heap's k-th entry. There is never more than one candidate per boundary, so
// the heap fits in the nodes and the two share a single allocation.
type mergeNode struct {
	prev, next int32
	slot       int32
	candidate  mergeCandidate
}

type mergeCandidate struct {
	rank   uint32
	offset int32
}

// mergePairsHeap yields the same tokens as mergePairsScan in O(n log n).
func (c *Codec) mergePairsHeap(piece string, yield func(uint, string)) {
	n := int32(len(piece))
	nodes := make([]mergeNode, n+1)
	for i := range nodes {
		nodes[i] = mergeNode{prev: int32(i) - 1, next: int32(i) + 1, slot: noSlot}
	}
	nodes[n].next = -1

	rankOf := func(i int32) uint32 {
		j := nodes[i].next
		if j < 0 {
			return noRank
		}
		k := nodes[j].next
		if k < 0 {
			return noRank
		}
		if rank, ok := c.vocabulary[piece[i:k]]; ok {
			return uint32(rank)
		}
		return noRank
	}

	candidates := mergeHeap{nodes: nodes}
	for i := int32(0); i+2 <= n; i++ {
		if rank := rankOf(i); rank != noRank {
			candidates.place(candidates.size, mergeCandidate{rank, i})
			candidates.size++
		}
	}
	candidates.init()

	for candidates.size > 0 {
		i := nodes[0].candidate.offset
		j := nodes[i].next
		k := nodes[j].next
		candidates.set(j, noRank)
		nodes[i].next = k
		if k >= 0 {
			nodes[k].prev = i
		}

		for _, b := range [2]int32{i, nodes[i].prev} {
			if b >= 0 {
				candidates.set(b, rankOf(b))
			}
		}
	}

	for i := int32(0); nodes[i].next >= 0; i = nodes[i].next {
		token := piece[i:nodes[i].next]
		yield(c.vocabulary[token], token)
	}
}

// mergeHeap is a binary min-heap of merge candidates ordered by rank, then by
// offset, so that among equal ranks the leftmost pair merges first. It stores
// its entries in nodes[:size] and keeps every boundary's slot up to date, so a
// candidate that a merge invalidates is removed instead of left behind: the
// heap never grows beyond the candidates that can still merge.
type mergeHeap struct {
	nodes []mergeNode
	size  int32
}

func (a mergeCandidate) less(b mergeCandidate) bool {
	if a.rank != b.rank {
		return a.rank < b.rank
	}
	return a.offset < b.offset
}

// place stores c as the heap's k-th entry.
func (h *mergeHeap) place(k int32, c mergeCandidate) {
	h.nodes[k].candidate = c
	h.nodes[c.offset].slot = k
}

func (h *mergeHeap) init() {
	for k := h.size/2 - 1; k >= 0; k-- {
		h.down(k, h.nodes[k].candidate)
	}
}

// set makes rank the rank of the candidate of the boundary at offset: it adds
// the candidate, moves it to where its new rank belongs or, for noRank, takes
// it out of the heap.
func (h *mergeHeap) set(offset int32, rank uint32) {
	k := h.nodes[offset].slot
	c := mergeCandidate{rank, offset}

	switch {
	case rank != noRank && k == noSlot:
		h.size++
		h.up(h.size-1, c)
		return
	case rank == noRank && k == noSlot:
		return
	case rank == noRank:
		h.nodes[offset].slot = noSlot
		h.size--
		if k == h.size {
			return
		}
		c = h.nodes[h.size].candidate // the last entry takes the place of the one removed
	}

	if k > 0 && c.less(h.nodes[(k-1)/2].candidate) {
		h.up(k, c)
	} else {
		h.down(k, c)
	}
}

// up places c at k or above, moving the entries it passes one level down.
func (h *mergeHeap) up(k int32, c mergeCandidate) {
	for k > 0 {
		parent := (k - 1) / 2
		if !c.less(h.nodes[parent].candidate) {
			break
		}
		h.place(k, h.nodes[parent].candidate)
		k = parent
	}
	h.place(k, c)
}

// down places c at k or below, moving the entries it passes one level up.
func (h *mergeHeap) down(k int32, c mergeCandidate) {
	for k < h.size/2 {
		child := 2*k + 1
		if right := child + 1; right < h.size && h.nodes[right].candidate.less(h.nodes[child].candidate) {
			child = right
		}
		if !h.nodes[child].candidate.less(c) {
			break
		}
		h.place(k, h.nodes[child].candidate)
		k = child
	}
	h.place(k, c)
}
