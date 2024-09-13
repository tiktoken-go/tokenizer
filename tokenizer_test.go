package tokenizer_test

import (
	"testing"

	"github.com/tiktoken-go/tokenizer"
)

func TestO200kBaseEncoding(t *testing.T) {
	tokenizer, err := tokenizer.Get(tokenizer.O200kBase)
	if err != nil {
		t.Fatalf("can't create tokenizer")
	}

	tests := []struct {
		text string
		ids  []uint
	}{
		{text: "hello world", ids: []uint{24912, 2375}},
		{text: "hello  world", ids: []uint{24912, 220, 2375}},
		{text: "hello   world", ids: []uint{24912, 256, 2375}},
		{text: "supercalifragilistic", ids: []uint{17789, 5842, 366, 17764, 311, 6207}},
		{text: "We know what we are, but know not what we may be.", ids: []uint{2167, 1761, 1412, 581, 553, 11, 889, 1761, 625, 1412, 581, 1340, 413, 13}},
	}

	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			ids, _, err := tokenizer.Encode(test.text)
			if err != nil {
				t.Fatalf("error encoding: %v", err)
			}

			if !sliceEqual(ids, test.ids) {
				t.Fatalf("input: %s want: %v got: %v", test.text, test.ids, ids)
			}

			text, err := tokenizer.Decode(ids)
			if err != nil {
				t.Fatalf("error decoding: %v", err)
			}

			if text != test.text {
				t.Fatalf("input: %v want: %s got: %s", test.ids, test.text, text)
			}
		})
	}
}

func TestCl100kEncoding(t *testing.T) {
	tokenizer, err := tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
		t.Fatalf("can't create tokenizer")
	}

	tests := []struct {
		text string
		ids  []uint
	}{
		{text: "hello world", ids: []uint{15339, 1917}},
		{text: "hello  world", ids: []uint{15339, 220, 1917}},
		{text: "hello   world", ids: []uint{15339, 256, 1917}},
		{text: "supercalifragilistic", ids: []uint{13066, 3035, 278, 333, 4193, 321, 4633}},
		{text: "We know what we are, but know not what we may be.", ids: []uint{1687, 1440, 1148, 584, 527, 11, 719, 1440, 539, 1148, 584, 1253, 387, 13}},
	}

	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			ids, _, err := tokenizer.Encode(test.text)
			if err != nil {
				t.Fatalf("error encoding: %v", err)
			}

			if !sliceEqual(ids, test.ids) {
				t.Fatalf("input: %s want: %v got: %v", test.text, test.ids, ids)
			}

			text, err := tokenizer.Decode(ids)
			if err != nil {
				t.Fatalf("error decoding: %v", err)
			}

			if text != test.text {
				t.Fatalf("input: %v want: %s got: %s", test.ids, test.text, text)
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
		text string
		ids  []uint
	}{
		{text: "hello world", ids: []uint{31373, 995}},
		{text: "hello  world", ids: []uint{31373, 220, 995}},
		{text: "hello   world", ids: []uint{31373, 220, 220, 995}},
		{text: "supercalifragilistic", ids: []uint{16668, 9948, 361, 22562, 346, 2569}},
		{text: "We know what we are, but know not what we may be.", ids: []uint{1135, 760, 644, 356, 389, 11, 475, 760, 407, 644, 356, 743, 307, 13}},
	}

	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			ids, _, err := tokenizer.Encode(test.text)
			if err != nil {
				t.Fatalf("error encoding: %v", err)
			}

			if !sliceEqual(ids, test.ids) {
				t.Fatalf("input: %s want: %v got: %v", test.text, test.ids, ids)
			}

			text, err := tokenizer.Decode(ids)
			if err != nil {
				t.Fatalf("error decoding: %v", err)
			}

			if text != test.text {
				t.Fatalf("input: %v want: %s got: %s", test.ids, test.text, text)
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
		text string
		ids  []uint
	}{
		{text: "hello world", ids: []uint{31373, 995}},
		{text: "hello  world", ids: []uint{31373, 220, 995}},
		{text: "hello   world", ids: []uint{31373, 50257, 995}},
		{text: "supercalifragilistic", ids: []uint{16668, 9948, 361, 22562, 346, 2569}},
		{text: "We know what we are, but know not what we may be.", ids: []uint{1135, 760, 644, 356, 389, 11, 475, 760, 407, 644, 356, 743, 307, 13}},
	}

	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			ids, _, err := tokenizer.Encode(test.text)
			if err != nil {
				t.Fatalf("error encoding: %v", err)
			}

			if !sliceEqual(ids, test.ids) {
				t.Fatalf("input: %s want: %v got: %v", test.text, test.ids, ids)
			}

			text, err := tokenizer.Decode(ids)
			if err != nil {
				t.Fatalf("error decoding: %v", err)
			}

			if text != test.text {
				t.Fatalf("input: %v want: %s got: %s", test.ids, test.text, text)
			}
		})
	}
}

func TestP50kEditEncoding(t *testing.T) {
	tokenizer, err := tokenizer.Get(tokenizer.P50kEdit)
	if err != nil {
		t.Fatalf("can't create tokenizer")
	}

	tests := []struct {
		text string
		ids  []uint
	}{
		{text: "hello world", ids: []uint{31373, 995}},
		{text: "hello  world", ids: []uint{31373, 220, 995}},
		{text: "hello   world", ids: []uint{31373, 50257, 995}},
		{text: "supercalifragilistic", ids: []uint{16668, 9948, 361, 22562, 346, 2569}},
		{text: "We know what we are, but know not what we may be.", ids: []uint{1135, 760, 644, 356, 389, 11, 475, 760, 407, 644, 356, 743, 307, 13}},
	}

	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			ids, _, err := tokenizer.Encode(test.text)
			if err != nil {
				t.Fatalf("error encoding: %v", err)
			}

			if !sliceEqual(ids, test.ids) {
				t.Fatalf("input: %s want: %v got: %v", test.text, test.ids, ids)
			}

			text, err := tokenizer.Decode(ids)
			if err != nil {
				t.Fatalf("error decoding: %v", err)
			}

			if text != test.text {
				t.Fatalf("input: %v want: %s got: %s", test.ids, test.text, text)
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
