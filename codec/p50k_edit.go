package codec

import "github.com/dlclark/regexp2"

func NewP50kEdit() *Codec {
	p50kBaseVocabOnce.Do(p50kBaseVocabInit)
	return &Codec{
		name:        "p50k_edit",
		vocabulary:  p50kBaseVocab,
		splitRegexp: regexp2.MustCompile(`'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+(?!\S)|\s+`, regexp2.None),
		specialTokens: map[string]uint{
			"<|endoftext|>":  50256,
			"<|fim_prefix|>": 50281,
			"<|fim_middle|>": 50282,
			"<|fim_suffix|>": 50283,
		},
	}
}
