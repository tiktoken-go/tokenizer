package codec

import (
	"regexp"
)

func NewP50kBase() *Codec {
	return &Codec{
		name:        "p50k_base",
		vocabulary:  p50kBaseVocab,
		splitRegexp: regexp.MustCompile(`'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+[^(\S)]|\s+`),
		specialTokens: map[string]uint{
			"<|endoftext|>": 50256,
		},
	}
}
