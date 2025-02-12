package codec

import "github.com/dlclark/regexp2"

func NewO200kBase() *Codec {
	o200kBaseVocabOnce.Do(o200kBaseVocabInit)

	splitRegexp := regexp2.MustCompile(
		`[^\r\n\p{L}\p{N}]?[\p{Lu}\p{Lt}\p{Lm}\p{Lo}\p{M}]*[\p{Ll}\p{Lm}\p{Lo}\p{M}]+(?i:'s|'t|'re|'ve|'m|'ll|'d)?|[^\r\n\p{L}\p{N}]?[\p{Lu}\p{Lt}\p{Lm}\p{Lo}\p{M}]+[\p{Ll}\p{Lm}\p{Lo}\p{M}]*(?i:'s|'t|'re|'ve|'m|'ll|'d)?|\p{N}{1,3}| ?[^\s\p{L}\p{N}]+[\r\n/]*|\s*[\r\n]+|\s+(?!\S)|\s+`,
		regexp2.None)

	return &Codec{
		name:        "o200k_base",
		vocabulary:  o200kBaseVocab,
		splitRegexp: splitRegexp,
		specialTokens: map[string]uint{
			"<|endoftext|>":   199999,
			"<|endofprompt|>": 200018,
		},
	}
}
