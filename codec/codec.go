package codec

import (
	"fmt"
	"math"

	"github.com/dlclark/regexp2"
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

	err := c.tokenize(input, func(_ uint) {
		count++
	})

	return count, err
}

// Encode returns the token IDs and tokens for the input string.
func (c *Codec) Encode(input string) ([]uint, []string, error) {

	var ids []uint
	var tokens []string

	err := c.tokenize(input, func(id uint) {
		ids = append(ids, id)
		tokens = append(tokens, c.reverseVocabulary[id])
	})

	return ids, tokens, err
}

func (c *Codec) tokenize(input string, yield func(uint)) error {
	match, err := c.splitRegexp.FindStringMatch(input)
	if err != nil {
		return fmt.Errorf("error matching: %v", err)
	}
	for match != nil {
		piece := match.String()
		if id, ok := c.vocabulary[piece]; ok {
			yield(id)
		} else {
			parts := c.mergePairs(piece)

			for i := range len(parts) - 1 {
				token := piece[parts[i].offset:parts[i+1].offset]
				yield(c.vocabulary[token])
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

func (c *Codec) mergePairs(piece string) []part {
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
