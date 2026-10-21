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
		c.reverseVocabulary = make(map[uint]string)
		for k, v := range c.vocabulary {
			c.reverseVocabulary[v] = k
		}
	})

	var out string
	for _, t := range tokens {
		piece, ok := c.reverseVocabulary[t]
		if !ok {
			return "", fmt.Errorf("invalid token: %d", t)
		}
		out += piece
	}

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
