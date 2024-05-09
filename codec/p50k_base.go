package codec

import "github.com/dlclark/regexp2"

func NewP50kBase() *Codec {
	p50kBaseVocabOnce.Do(p50kBaseVocabInit)
	return &Codec{
		name:        "p50k_base",
		vocabulary:  p50kBaseVocab,
		splitRegexp: regexp2.MustCompile(`'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+(?!\S)|\s+`, regexp2.None),
		specialTokens: map[string]uint{
			"<|endoftext|>": 50256,
		},
	}
}
