package codec

import (
	"fmt"
	"math"
	"regexp"
)

type Codec struct {
	vocabulary        vocab
	reverseVocabulary reverse
	specialTokens     map[string]uint
	splitRegexp       *regexp.Regexp
	name              string
}

func (c *Codec) GetName() string {
	return c.name
}

func (c *Codec) Encode(input string) ([]uint, []string, error) {
	var (
		ids    []uint
		tokens []string
	)

	for _, piece := range c.splitRegexp.FindAll([]byte(input), -1) {
		if id, ok := c.vocabulary[string(piece)]; ok {
			ids = append(ids, id)
			tokens = append(tokens, string(piece))
			continue
		}
		newIds, newTokens := c.bpe(piece)
		ids = append(ids, newIds...)
		tokens = append(tokens, newTokens...)
	}
	return ids, tokens, nil
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

func (c *Codec) bpe(piece []byte) ([]uint, []string) {
	type part struct {
		offset int
		rank   uint
	}

	parts := make([]part, len(piece)+1)
	for i := 0; i < len(parts); i++ {
		parts[i] = part{i, math.MaxUint}
	}

	getRank := func(index, skip int) uint {
		if index+skip+2 < len(parts) {
			start := parts[index].offset
			end := parts[index+skip+2].offset
			if rank, ok := c.vocabulary[string(piece[start:end])]; ok {
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

	ids := make([]uint, len(parts)-1)
	tokens := make([]string, len(parts)-1)
	for i := 0; i < len(ids); i++ {
		token := string(piece[parts[i].offset:parts[i+1].offset])
		tokens[i] = token
		ids[i] = c.vocabulary[token]
	}
	return ids, tokens
}
