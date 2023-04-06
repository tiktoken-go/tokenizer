package codec

import (
	"errors"
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

// Greedy algo
// func (c *Codec) bpe(piece []byte) ([]uint, []string) {
// 	var ids []uint
// 	var tokens []string
// 	i := 0
// 	for i < len(piece) {
// 		for j := len(piece); j > i; j-- {
// 			token := string(piece[i:j])
// 			if id, ok := c.vocabulary[token]; ok {
// 				tokens = append(tokens, token)
// 				ids = append(ids, id)
// 				i = j
// 				break
// 			}
// 		}
// 	}
// 	return ids, tokens
// }

func (c *Codec) bpe(piece []byte) ([]uint, []string) {
	type byteRange struct {
		start int
		end   int
	}

	parts := make([]byteRange, len(piece)+1)
	for i := 0; i < len(parts); i++ {
		parts[i] = byteRange{i, math.MaxInt64}
	}

	getRank := func(parts []byteRange, startIdx, skip int) (uint, error) {
		if startIdx+skip+2 < len(parts) {
			chunk := string(piece[parts[startIdx].start:parts[startIdx+skip+2].start])
			if rank, ok := c.vocabulary[chunk]; ok {
				return rank, nil
			}
		}
		return math.MaxInt64, errors.New("not found")
	}

	for i := 0; i < len(parts)-2; i++ {
		if r, err := getRank(parts, i, 0); err == nil {
			parts[i].end = int(r)
		}
	}

	for {
		if len(parts) == 1 {
			break
		}

		minRank := byteRange{math.MaxInt64, 0}
		for i, p := range parts[:len(parts)-1] {
			if p.end < minRank.start {
				minRank = byteRange{p.end, i}
			}
		}

		if minRank.start != math.MaxInt64 {
			i := minRank.end

			parts[i].end = math.MaxInt64
			if r, err := getRank(parts, i, 1); err == nil {
				parts[i].end = int(r)
			}
			if i > 0 {
				parts[i-1].end = math.MaxInt64
				if r, err := getRank(parts, i-1, 1); err == nil {
					parts[i-1].end = int(r)
				}
			}

			parts = append(parts[:i+1], parts[i+2:]...)
		} else {
			break
		}
	}

	ids := make([]uint, len(parts)-1)
	tokens := make([]string, len(parts)-1)
	for i := 0; i < len(ids); i++ {
		token := string(piece[parts[i].start:parts[i+1].start])
		tokens[i] = token
		ids[i] = c.vocabulary[token]
	}
	return ids, tokens
}
