package codec

import "github.com/dlclark/regexp2"

func NewCl100kBase() *Codec {
	cl100kBaseVocabOnce.Do(cl100kBaseVocabInit)

	splitRegexp := regexp2.MustCompile(`(?i:'s|'t|'re|'ve|'m|'ll|'d)|[^\r\n\p{L}\p{N}]?\p{L}+|\p{N}{1,3}| ?[^\s\p{L}\p{N}]+[\r\n]*|\s*[\r\n]+|\s+(?!\S)|\s+`, regexp2.None)

	return &Codec{
		name:        "cl100k_base",
		vocabulary:  cl100kBaseVocab,
		splitRegexp: splitRegexp,
		specialTokens: map[string]uint{
			"<|endoftext|>":   100257,
			"<|fim_prefix|>":  100258,
			"<|fim_middle|>":  100259,
			"<|fim_suffix|>":  100260,
			"<|endofprompt|>": 100276,
		},
	}
}
