package codec

import "github.com/dlclark/regexp2"

func NewR50kBase() *Codec {
	r50kBaseVocabOnce.Do(r50kBaseVocabInit)
	return &Codec{
		name:        "r50k_base",
		vocabulary:  r50kBaseVocab,
		splitRegexp: regexp2.MustCompile(`'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+(?!\S)|\s+`, regexp2.None),
		specialTokens: map[string]uint{
			"<|endoftext|>": 50256,
		},
	}
}
