package tokenizer_test

import (
	"testing"

	"github.com/tiktoken-go/tokenizer"
)

func TestCl100kEncoding(t *testing.T) {
	tokenizer, err := tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
		t.Fatalf("can't create tokenizer")
	}

	tests := []struct {
		input string
		want  []uint
	}{
		{input: "hello world", want: []uint{15339, 1917}},
		{input: "supercalifragilistic", want: []uint{9712, 5531, 97816, 37135, 1638, 292}},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			tokens, _, err := tokenizer.Encode(test.input)
			if err != nil {
				t.Fatalf("error encoding: %v", err)
			}

			if !sliceEqual(tokens, test.want) {
				t.Fatalf("input: %s want: %v got: %v", test.input, test.want, tokens)
			}
		})
	}
}

func TestCl100kDecoding(t *testing.T) {
	tokenizer, err := tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
		t.Fatalf("can't create tokenizer")
	}

	tests := []struct {
		input []uint
		want  string
	}{
		{want: "hello world", input: []uint{15339, 1917}},
		{want: "supercalifragilistic", input: []uint{9712, 5531, 97816, 37135, 1638, 292}},
	}

	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			text, err := tokenizer.Decode(test.input)
			if err != nil {
				t.Fatalf("error encoding: %v", err)
			}

			if text != test.want {
				t.Fatalf("input: %v want: %v got: %v", test.input, test.want, text)
			}
		})
	}
}

func TestP50kBaseEncoding(t *testing.T) {
	tokenizer, err := tokenizer.Get(tokenizer.P50kBase)
	if err != nil {
		t.Fatalf("can't create tokenizer")
	}

	tests := []struct {
		input string
		want  []uint
	}{
		{input: "hello world", want: []uint{31373, 995}},
		{input: "supercalifragilistic", want: []uint{16668, 9948, 361, 22562, 2403, 11268}},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			tokens, _, err := tokenizer.Encode(test.input)
			if err != nil {
				t.Fatalf("error encoding: %v", err)
			}

			if !sliceEqual(tokens, test.want) {
				t.Fatalf("input: %s want: %v got: %v", test.input, test.want, tokens)
			}
		})
	}
}

func TestP50kDecoding(t *testing.T) {
	tokenizer, err := tokenizer.Get(tokenizer.P50kBase)
	if err != nil {
		t.Fatalf("can't create tokenizer")
	}

	tests := []struct {
		input []uint
		want  string
	}{
		{want: "hello world", input: []uint{31373, 995}},
		{want: "supercalifragilistic", input: []uint{16668, 9948, 361, 22562, 2403, 11268}},
	}

	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			text, err := tokenizer.Decode(test.input)
			if err != nil {
				t.Fatalf("error encoding: %v", err)
			}

			if text != test.want {
				t.Fatalf("input: %v want: %v got: %v", test.input, test.want, text)
			}
		})
	}
}

func TestR50kBaseEncoding(t *testing.T) {
	tokenizer, err := tokenizer.Get(tokenizer.R50kBase)
	if err != nil {
		t.Fatalf("can't create tokenizer")
	}

	tests := []struct {
		input string
		want  []uint
	}{
		{input: "hello world", want: []uint{31373, 995}},
		{input: "supercalifragilistic", want: []uint{16668, 9948, 361, 22562, 2403, 11268}},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			tokens, _, err := tokenizer.Encode(test.input)
			if err != nil {
				t.Fatalf("error encoding: %v", err)
			}

			if !sliceEqual(tokens, test.want) {
				t.Fatalf("input: %s want: %v got: %v", test.input, test.want, tokens)
			}
		})
	}
}

func TestR50kDecoding(t *testing.T) {
	tokenizer, err := tokenizer.Get(tokenizer.R50kBase)
	if err != nil {
		t.Fatalf("can't create tokenizer")
	}

	tests := []struct {
		input []uint
		want  string
	}{
		{want: "hello world", input: []uint{31373, 995}},
		{want: "supercalifragilistic", input: []uint{16668, 9948, 361, 22562, 2403, 11268}},
	}

	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			text, err := tokenizer.Decode(test.input)
			if err != nil {
				t.Fatalf("error encoding: %v", err)
			}

			if text != test.want {
				t.Fatalf("input: %v want: %v got: %v", test.input, test.want, text)
			}
		})
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
