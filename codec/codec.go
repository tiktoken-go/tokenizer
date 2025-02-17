package codec

import (
	"fmt"
	"iter"
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
func (c *Codec) Count(input string) (count int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error encoding: %v", r)
		}
	}()

	for _, _ = range c.tokenize(input) {
		count++
	}

	return count, err
}

// Encode returns the token IDs and tokens for the input string.
func (c *Codec) Encode(input string) (ids []uint, tokens []string, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error encoding: %v", r)
		}
	}()

	for id, token := range c.tokenize(input) {
		ids = append(ids, id)
		tokens = append(tokens, token)
	}

	return ids, tokens, err
}

func (c *Codec) tokenize(input string) iter.Seq2[uint, string] {
	return func(yield func(uint, string) bool) {
		match, err := c.splitRegexp.FindStringMatch(input)
		if err != nil {
			panic(fmt.Errorf("error matching: %v", err))
		}
		for match != nil {
			piece := match.String()
			if id, ok := c.vocabulary[piece]; ok {
				if !yield(id, piece) {
					break
				}
			} else {
				parts := c.mergePairs([]byte(piece))

				for i := 0; i < len(parts)-1; i++ {
					token := string(piece[parts[i].offset:parts[i+1].offset])
					if !yield(c.vocabulary[token], token) {
						break
					}
				}
			}
			m, err := c.splitRegexp.FindNextMatch(match)
			if err != nil {
				break
			}
			match = m
		}
	}
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

func (c *Codec) mergePairs(piece []byte) []part {
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

	return parts
}
