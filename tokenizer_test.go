package tokenizer_test

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"sync"
	"testing"
	"unicode"

	"github.com/tiktoken-go/tokenizer"
)

type testCase struct {
	text string
	ids  []uint
}

func TestO200kBase(t *testing.T) {
	tok, err := tokenizer.Get(tokenizer.O200kBase)
	if err != nil {
		t.Fatalf("can't create tokenizer: %v", err)
	}

	tests := []testCase{
		{text: "'RE", ids: []uint{6, 1099}},
		{text: "hello world", ids: []uint{24912, 2375}},
		{text: "hello  world", ids: []uint{24912, 220, 2375}},
		{text: "hello   world", ids: []uint{24912, 256, 2375}},
		{text: "supercalifragilistic", ids: []uint{17789, 5842, 366, 17764, 311, 6207}},
		{text: "We know what we are, but know not what we may be.", ids: []uint{2167, 1761, 1412, 581, 553, 11, 889, 1761, 625, 1412, 581, 1340, 413, 13}},
	}

	runTests(t, tok, tests)
}

func TestCl100kBase(t *testing.T) {
	tok, err := tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
		t.Fatalf("can't create tokenizer: %v", err)
	}

	tests := []testCase{
		{text: "'RE", ids: []uint{95253}},
		{text: "hello world", ids: []uint{15339, 1917}},
		{text: "hello  world", ids: []uint{15339, 220, 1917}},
		{text: "hello   world", ids: []uint{15339, 256, 1917}},
		{text: "supercalifragilistic", ids: []uint{13066, 3035, 278, 333, 4193, 321, 4633}},
		{text: "We know what we are, but know not what we may be.", ids: []uint{1687, 1440, 1148, 584, 527, 11, 719, 1440, 539, 1148, 584, 1253, 387, 13}},
	}

	runTests(t, tok, tests)
}

func TestR50kBase(t *testing.T) {
	tok, err := tokenizer.Get(tokenizer.R50kBase)
	if err != nil {
		t.Fatalf("can't create tokenizer: %v", err)
	}

	tests := []testCase{
		{text: "hello world", ids: []uint{31373, 995}},
		{text: "hello  world", ids: []uint{31373, 220, 995}},
		{text: "hello   world", ids: []uint{31373, 220, 220, 995}},
		{text: "supercalifragilistic", ids: []uint{16668, 9948, 361, 22562, 346, 2569}},
		{text: "We know what we are, but know not what we may be.", ids: []uint{1135, 760, 644, 356, 389, 11, 475, 760, 407, 644, 356, 743, 307, 13}},
	}

	runTests(t, tok, tests)
}

func TestP50kBase(t *testing.T) {
	tok, err := tokenizer.Get(tokenizer.P50kBase)
	if err != nil {
		t.Fatalf("can't create tokenizer: %v", err)
	}

	tests := []testCase{
		{text: "hello world", ids: []uint{31373, 995}},
		{text: "hello  world", ids: []uint{31373, 220, 995}},
		{text: "hello   world", ids: []uint{31373, 50257, 995}},
		{text: "supercalifragilistic", ids: []uint{16668, 9948, 361, 22562, 346, 2569}},
		{text: "We know what we are, but know not what we may be.", ids: []uint{1135, 760, 644, 356, 389, 11, 475, 760, 407, 644, 356, 743, 307, 13}},
	}

	runTests(t, tok, tests)
}

func TestO200kBaseUnicodeMismatch(t *testing.T) {
	major, _, _ := strings.Cut(unicode.Version, ".")
	if v, err := strconv.Atoi(major); err != nil || v < 16 {
		t.Skipf("requires Unicode 16+, have %s", unicode.Version)
	}

	tok, err := tokenizer.Get(tokenizer.O200kBase)
	if err != nil {
		t.Fatalf("can't create tokenizer: %v", err)
	}

	tests := []testCase{
		{text: "\u1C89\\u", ids: []uint{157, 110, 231, 7570}},
	}

	runTests(t, tok, tests)
}

func TestDecodeConcurrent(t *testing.T) {
	tok, err := tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
		t.Fatalf("Can't create tokenizer: %v", err)
	}

	ids, _, err := tok.Encode("hello world")
	if err != nil {
		t.Fatal(err)
	}

	const goroutines = 16
	var (
		start sync.WaitGroup
		done  sync.WaitGroup
	)
	start.Add(1)
	for range goroutines {
		done.Go(func() {
			start.Wait()
			got, err := tok.Decode(ids)
			if err != nil {
				t.Error(err)
				return
			}
			if got != "hello world" {
				t.Errorf("Decode = %q, want %q", got, "hello world")
			}
		})
	}
	start.Done()
	done.Wait()
}

func TestErrorModeDefault(t *testing.T) {
	tok, err := tokenizer.Get(tokenizer.O200kBase)
	if err != nil {
		t.Fatalf("can't create tokenizer: %v", err)
	}

	if got := tok.ErrorMode(); got != tokenizer.ErrorModeNone {
		t.Errorf("default error mode - want: %v got: %v", tokenizer.ErrorModeNone, got)
	}

	tok.SetErrorMode(tokenizer.ErrorModeStrict)
	if got := tok.ErrorMode(); got != tokenizer.ErrorModeStrict {
		t.Errorf("error mode after set - want: %v got: %v", tokenizer.ErrorModeStrict, got)
	}
}

func TestDecodeErrorMode(t *testing.T) {
	tok, err := tokenizer.Get(tokenizer.O200kBase)
	if err != nil {
		t.Fatalf("can't create tokenizer: %v", err)
	}

	tests := []struct {
		name    string
		ids     []uint
		none    string
		replace string
		ignore  string
		invalid bool
		offset  int // byte offset of the first invalid sequence, for strict
	}{
		{name: "valid", ids: []uint{24912, 2375}, none: "hello world", replace: "hello world", ignore: "hello world"},
		{name: "truncated", ids: []uint{35895}, none: "\xe8\x96", replace: "�", ignore: "", invalid: true, offset: 0},
		{name: "truncated twice", ids: []uint{35895, 35895}, none: "\xe8\x96\xe8\x96", replace: "��", ignore: "", invalid: true, offset: 0},
		{name: "truncated in text", ids: []uint{64, 35895, 65}, none: "a\xe8\x96b", replace: "a�b", ignore: "ab", invalid: true, offset: 1},
		{name: "invalid bytes", ids: []uint{187, 186}, none: "\xff\xfe", replace: "��", ignore: "", invalid: true, offset: 0},
		{name: "surrogate", ids: []uint{64, 169, 254, 222, 65}, none: "a\xed\xa0\x80b", replace: "a���b", ignore: "ab", invalid: true, offset: 1},
		{name: "truncated before valid", ids: []uint{35895, 377}, none: "\xe8\x96é", replace: "�é", ignore: "é", invalid: true, offset: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, mode := range []struct {
				mode tokenizer.ErrorMode
				name string
				want string
			}{
				{tokenizer.ErrorModeNone, "none", test.none},
				{tokenizer.ErrorModeReplace, "replace", test.replace},
				{tokenizer.ErrorModeIgnore, "ignore", test.ignore},
			} {
				tok.SetErrorMode(mode.mode)
				got, err := tok.Decode(test.ids)
				if err != nil {
					t.Errorf("%s: error decoding: %v", mode.name, err)
					continue
				}
				if got != mode.want {
					t.Errorf("%s: decoding mismatch - want: %q got: %q", mode.name, mode.want, got)
				}
			}

			tok.SetErrorMode(tokenizer.ErrorModeStrict)
			got, err := tok.Decode(test.ids)
			switch {
			case test.invalid:
				checkInvalidUTF8(t, "strict decode", err, test.offset)
			case err != nil:
				t.Errorf("strict: error decoding: %v", err)
			case got != test.none:
				t.Errorf("strict: decoding mismatch - want: %q got: %q", test.none, got)
			}
		})
	}
}

func TestEncodeErrorMode(t *testing.T) {
	tok, err := tokenizer.Get(tokenizer.O200kBase)
	if err != nil {
		t.Fatalf("can't create tokenizer: %v", err)
	}
	ref, err := tokenizer.Get(tokenizer.O200kBase)
	if err != nil {
		t.Fatalf("can't create tokenizer: %v", err)
	}

	tests := []struct {
		name    string
		ids     []uint
		none    string
		replace string
		ignore  string
		invalid bool
		offset  int // byte offset of the first invalid sequence, for strict
	}{
		{name: "valid", ids: []uint{24912, 2375}, none: "hello world", replace: "hello world", ignore: "hello world"},
		{name: "truncated", ids: []uint{35895}, none: "\xe8\x96", replace: "�", ignore: "", invalid: true, offset: 0},
		{name: "truncated twice", ids: []uint{35895, 35895}, none: "\xe8\x96\xe8\x96", replace: "��", ignore: "", invalid: true, offset: 0},
		{name: "truncated in text", ids: []uint{64, 35895, 65}, none: "a\xe8\x96b", replace: "a�b", ignore: "ab", invalid: true, offset: 1},
		{name: "invalid bytes", ids: []uint{187, 186}, none: "\xff\xfe", replace: "��", ignore: "", invalid: true, offset: 0},
		{name: "surrogate", ids: []uint{64, 169, 254, 222, 65}, none: "a\xed\xa0\x80b", replace: "a���b", ignore: "ab", invalid: true, offset: 1},
		{name: "truncated before valid", ids: []uint{35895, 377}, none: "\xe8\x96é", replace: "�é", ignore: "é", invalid: true, offset: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, mode := range []struct {
				mode tokenizer.ErrorMode
				name string
				want string
			}{
				{tokenizer.ErrorModeReplace, "replace", test.replace},
				{tokenizer.ErrorModeIgnore, "ignore", test.ignore},
			} {
				want, _, err := ref.Encode(mode.want)
				if err != nil {
					t.Fatalf("%s: error encoding reference: %v", mode.name, err)
				}

				tok.SetErrorMode(mode.mode)
				ids, _, err := tok.Encode(test.none)
				if err != nil {
					t.Errorf("%s: error encoding: %v", mode.name, err)
					continue
				}
				if !sliceEqual(ids, want) {
					t.Errorf("%s: encoding mismatch - want: %v got: %v", mode.name, want, ids)
				}

				count, err := tok.Count(test.none)
				if err != nil {
					t.Errorf("%s: error counting: %v", mode.name, err)
					continue
				}
				if count != len(want) {
					t.Errorf("%s: count mismatch - want: %d got: %d", mode.name, len(want), count)
				}
			}

			tok.SetErrorMode(tokenizer.ErrorModeStrict)
			_, _, err := tok.Encode(test.none)
			if test.invalid {
				checkInvalidUTF8(t, "strict encode", err, test.offset)
			} else if err != nil {
				t.Errorf("strict: error encoding: %v", err)
			}
			_, err = tok.Count(test.none)
			if test.invalid {
				checkInvalidUTF8(t, "strict count", err, test.offset)
			} else if err != nil {
				t.Errorf("strict: error counting: %v", err)
			}
		})
	}
}

func checkInvalidUTF8(t *testing.T, name string, err error, offset int) {
	t.Helper()
	if !errors.Is(err, tokenizer.ErrInvalidUTF8) {
		t.Errorf("%s: want ErrInvalidUTF8, got: %v", name, err)
		return
	}
	if want := fmt.Sprintf("at byte %d", offset); !strings.HasSuffix(err.Error(), want) {
		t.Errorf("%s: want error ending in %q, got: %v", name, want, err)
	}
}

func TestDecodeInvalidToken(t *testing.T) {
	tests := []struct {
		encoding tokenizer.Encoding
		id       uint
	}{
		{encoding: tokenizer.O200kBase, id: 199998},   // first ID past the vocabulary
		{encoding: tokenizer.O200kBase, id: ^uint(0)}, // max uint
		{encoding: tokenizer.P50kBase, id: 50256},     // gap: <|endoftext|> is a special token
	}

	for _, test := range tests {
		tok, err := tokenizer.Get(test.encoding)
		if err != nil {
			t.Fatalf("can't create tokenizer: %v", err)
		}
		if got, err := tok.Decode([]uint{64, test.id}); err == nil {
			t.Errorf("%s: expected error decoding %d, got: %q", test.encoding, test.id, got)
		}
	}
}

func longPiece(alphabet string, n int) string {
	rng := rand.New(rand.NewPCG(1, 2))
	runes := []rune(alphabet)

	var b strings.Builder
	for {
		r := runes[rng.IntN(len(runes))]
		if b.Len()+len(string(r)) > n {
			return b.String()
		}
		b.WriteRune(r)
	}
}

// BenchmarkLongPiece measures Count on a single piece of increasing length for
// each alphabet, which shows how merging scales with piece length.
func BenchmarkLongPiece(b *testing.B) {
	var longPieceAlphabets = []struct {
		name     string
		encoding tokenizer.Encoding
		runes    string
	}{
		{"chinese", tokenizer.O200kBase, "网关在转发请求之前需要估算输入的令牌数量以便检查余额和选择超时档位"},
		{"latin", tokenizer.O200kBase, "abcdefghijklmnopqrstuvwxyz"},
		{"spaces", tokenizer.O200kBase, " "},
		{"digits", tokenizer.P50kBase, "0123456789"},
		{"symbols", tokenizer.O200kBase, "=-*#/\\|_.:;"},
	}

	for _, alphabet := range longPieceAlphabets {
		tok, err := tokenizer.Get(alphabet.encoding)
		if err != nil {
			b.Fatalf("can't create tokenizer: %v", err)
		}

		for _, n := range []int{16, 64, 256, 1024, 4096} {
			piece := longPiece(alphabet.runes, n)

			b.Run(fmt.Sprintf("%s/%d", alphabet.name, len(piece)), func(b *testing.B) {
				b.SetBytes(int64(len(piece)))
				for b.Loop() {
					if _, err := tok.Count(piece); err != nil {
						b.Fatal(err)
					}
				}
				b.ReportAllocs()
			})
		}
	}
}

func runTests(t *testing.T, tok tokenizer.Codec, tests []testCase) {
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			ids, _, err := tok.Encode(test.text)
			if err != nil {
				t.Fatalf("error encoding: %v", err)
			}
			if !sliceEqual(ids, test.ids) {
				t.Errorf("encoding mismatch - want: %v got: %v", test.ids, ids)
			}

			text, err := tok.Decode(ids)
			if err != nil {
				t.Fatalf("error decoding: %v", err)
			}
			if text != test.text {
				t.Errorf("decoding mismatch - want: %s got: %s", test.text, text)
			}

			count, err := tok.Count(test.text)
			if err != nil {
				t.Fatalf("error counting: %v", err)
			}
			if count != len(test.ids) {
				t.Errorf("count mismatch - want: %d got: %d", len(test.ids), count)
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
