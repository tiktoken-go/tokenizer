package tokenizer_test

import (
	"fmt"
	"testing"

	"github.com/tiktoken-go/tokenizer"
)

type testTokenizer struct {
	encoding tokenizer.Encoding
	data     []testTokenizerData
}

type testTokenizerData struct {
	text string
	ids  []uint
}

var (
	tokenizerTests = []testTokenizer{
		{
			encoding: tokenizer.O200kBase,
			data: []testTokenizerData{
				{text: "hello world", ids: []uint{24912, 2375}},
				{text: "hello  world", ids: []uint{24912, 220, 2375}},
				{text: "hello   world", ids: []uint{24912, 256, 2375}},
				{text: "supercalifragilistic", ids: []uint{17789, 5842, 366, 17764, 311, 6207}},
				{text: "We know what we are, but know not what we may be.", ids: []uint{2167, 1761, 1412, 581, 553, 11, 889, 1761, 625, 1412, 581, 1340, 413, 13}},
			},
		},
		{
			encoding: tokenizer.Cl100kBase,
			data: []testTokenizerData{
				{text: "hello world", ids: []uint{15339, 1917}},
				{text: "hello  world", ids: []uint{15339, 220, 1917}},
				{text: "hello   world", ids: []uint{15339, 256, 1917}},
				{text: "supercalifragilistic", ids: []uint{13066, 3035, 278, 333, 4193, 321, 4633}},
				{text: "We know what we are, but know not what we may be.", ids: []uint{1687, 1440, 1148, 584, 527, 11, 719, 1440, 539, 1148, 584, 1253, 387, 13}},
			},
		},
		{
			encoding: tokenizer.R50kBase,
			data: []testTokenizerData{
				{text: "hello world", ids: []uint{31373, 995}},
				{text: "hello  world", ids: []uint{31373, 220, 995}},
				{text: "hello   world", ids: []uint{31373, 220, 220, 995}},
				{text: "supercalifragilistic", ids: []uint{16668, 9948, 361, 22562, 346, 2569}},
				{text: "We know what we are, but know not what we may be.", ids: []uint{1135, 760, 644, 356, 389, 11, 475, 760, 407, 644, 356, 743, 307, 13}},
			},
		},
		{
			encoding: tokenizer.P50kBase,
			data: []testTokenizerData{
				{text: "hello world", ids: []uint{31373, 995}},
				{text: "hello  world", ids: []uint{31373, 220, 995}},
				{text: "hello   world", ids: []uint{31373, 50257, 995}},
				{text: "supercalifragilistic", ids: []uint{16668, 9948, 361, 22562, 346, 2569}},
				{text: "We know what we are, but know not what we may be.", ids: []uint{1135, 760, 644, 356, 389, 11, 475, 760, 407, 644, 356, 743, 307, 13}},
			},
		},
		{
			encoding: tokenizer.P50kEdit,
			data: []testTokenizerData{
				{text: "hello world", ids: []uint{31373, 995}},
				{text: "hello  world", ids: []uint{31373, 220, 995}},
				{text: "hello   world", ids: []uint{31373, 50257, 995}},
				{text: "supercalifragilistic", ids: []uint{16668, 9948, 361, 22562, 346, 2569}},
				{text: "We know what we are, but know not what we may be.", ids: []uint{1135, 760, 644, 356, 389, 11, 475, 760, 407, 644, 356, 743, 307, 13}},
			},
		},
	}
)

func TestTokenizer(t *testing.T) {
	for _, test := range tokenizerTests {
		tokenizer, err := tokenizer.Get(test.encoding)
		if err != nil {
			t.Fatalf("can't create tokenizer")
		}

		for _, data := range test.data {
			t.Run(fmt.Sprintf("%s: %s", test.encoding, data.text), func(t *testing.T) {
				ids, _, err := tokenizer.Encode(data.text)
				if err != nil {
					t.Fatalf("error encoding: %v", err)
				}

				if !sliceEqual(ids, data.ids) {
					t.Fatalf("input: %s want: %v got: %v", data.text, data.ids, ids)
				}

				text, err := tokenizer.Decode(ids)
				if err != nil {
					t.Fatalf("error decoding: %v", err)
				}

				if text != data.text {
					t.Fatalf("input: %v want: %s got: %s", data.ids, data.text, text)
				}
			})
		}
	}
}

var tokens []uint

func BenchmarkTokenizer(b *testing.B) {
	for _, test := range tokenizerTests {
		tokenizer, err := tokenizer.Get(test.encoding)
		if err != nil {
			b.Fatalf("can't create tokenizer")
		}

		for _, data := range test.data {
			b.Run(fmt.Sprintf("%s: %s", test.encoding, data.text), func(b *testing.B) {
				for i := 0; i < b.N; i++ {

					tokens, _, _ = tokenizer.Encode(data.text)
				}

				_ = tokens
			})
		}
	}
}

func sliceEqual(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}

	for i, elem := range a {
		if elem != b[i] {
			return false
		}
	}

	return true
}
