package codec

import (
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
)

// TestMergePairsHeap checks that mergePairsHeap yields exactly the tokens of
// mergePairsScan, for every encoding and for pieces on both sides of
// mergeHeapThreshold.
func TestMergePairsHeap(t *testing.T) {
	codecs := []*Codec{NewO200kBase(), NewCl100kBase(), NewP50kBase(), NewR50kBase()}
	alphabets := []struct {
		name string
		text string
	}{
		{"chinese", "网关在转发请求之前需要估算输入的令牌数量以便检查余额和选择超时档位"},
		{"latin", "abcdefghijklmnopqrstuvwxyz"},
		{"spaces", " "},
		{"digits", "0123456789"},
		{"symbols", "=-*#/\\|_.:;"},
		{"mixed", "the of and 网关 12 \t\n=-éñ\u0301😀"},
		{"bytes", "\x00\x7f\x80\xbf\xc3\xe4\xf0\xff"}, // most of these are not valid UTF-8 on their own
	}
	lengths := []int{0, 1, 2, 3, 7, mergeHeapThreshold - 1, mergeHeapThreshold, mergeHeapThreshold + 1, 257, 1024, 4099}

	rng := rand.New(rand.NewPCG(1, 2))
	for _, c := range codecs {
		for _, alphabet := range alphabets {
			units := strings.Split(alphabet.text, "") // UTF-8 sequences, and single bytes where there is none

			for _, n := range lengths {
				var b strings.Builder
				for b.Len() < n {
					b.WriteString(units[rng.IntN(len(units))])
				}
				piece := b.String()[:n] // may end in the middle of a UTF-8 sequence

				var want []string
				parts := c.mergePairsScan(piece)
				for i := range len(parts) - 1 {
					want = append(want, piece[parts[i].offset:parts[i+1].offset])
				}

				var got []string
				c.mergePairsHeap(piece, func(id uint, token string) {
					if id != c.vocabulary[token] {
						t.Errorf("%s/%s/%d: token %q yielded with id %d, want %d", c.name, alphabet.name, n, token, id, c.vocabulary[token])
					}
					got = append(got, token)
				})

				if !slices.Equal(got, want) {
					t.Errorf("%s/%s/%d: mergePairsHeap(%q)\n got %q\nwant %q", c.name, alphabet.name, n, piece, got, want)
				}
			}
		}
	}
}

// TestMergePairsHeapAllocs checks that the working memory of mergePairsHeap is
// a single allocation, as it is for mergePairsScan.
func TestMergePairsHeapAllocs(t *testing.T) {
	c := NewO200kBase()
	piece := strings.Repeat("网关在转发请求之前需要估算输入的令牌数量", 64)

	allocs := testing.AllocsPerRun(10, func() {
		c.mergePairsHeap(piece, func(uint, string) {})
	})
	if allocs != 1 {
		t.Errorf("mergePairsHeap allocated %v times, want 1", allocs)
	}
}
