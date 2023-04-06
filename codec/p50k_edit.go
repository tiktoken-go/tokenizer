package codec

import (
	"regexp"
)

func NewP50kEdit() *Codec {
	return &Codec{
		name:        "p50k_edit",
		vocabulary:  p50kBaseVocab,
		splitRegexp: regexp.MustCompile(`'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+[^(\S)]|\s+`),
		specialTokens: map[string]uint{
			"<|endoftext|>":  50256,
			"<|fim_prefix|>": 50281,
			"<|fim_middle|>": 50282,
			"<|fim_suffix|>": 50283,
		},
	}
}
