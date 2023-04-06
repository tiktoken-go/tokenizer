package codec

import (
	"regexp"
)

func NewCl100kBase() *Codec {
	return &Codec{
		name:        "cl100k_base",
		vocabulary:  cl100kBaseVocab,
		splitRegexp: regexp.MustCompile(`(?i:'s|'t|'re|'ve|'m|'ll|'d)|[^\r\n\p{L}\p{N}]?\p{L}+|\p{N}{1,3}| ?[^\s\p{L}\p{N}]+[\r\n]*|\s*[\r\n]+|\s+[^(\S)]|\s+`),
		specialTokens: map[string]uint{
			"<|endoftext|>":   100257,
			"<|fim_prefix|>":  100258,
			"<|fim_middle|>":  100259,
			"<|fim_suffix|>":  100260,
			"<|endofprompt|>": 100276,
		},
	}
}
