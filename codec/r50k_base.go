package codec

import (
	"regexp"
)

func NewR50kBase() *Codec {
	return &Codec{
		name:        "r50k_base",
		vocabulary:  r50kBaseVocab,
		splitRegexp: regexp.MustCompile(`'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+[^(\S)]|\s+`),
		specialTokens: map[string]uint{
			"<|endoftext|>": 50256,
		},
	}
}
