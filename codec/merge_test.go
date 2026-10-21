package codec

import (
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"
	"time"
)

func allCodecs() []*Codec {
	return []*Codec{
		NewO200kBase(),
		NewCl100kBase(),
		NewP50kBase(),
		NewP50kEdit(),
		NewR50kBase(),
	}
}

func requireSameParts(t *testing.T, c *Codec, piece string) {
	t.Helper()

	scan := c.mergePairsScan(piece)
	byHeap := c.mergePairsHeap(piece)

	if len(scan) != len(byHeap) {
		t.Fatalf("%s: %d bytes: scan gives %d parts, heap %d",
			c.GetName(), len(piece), len(scan), len(byHeap))
	}

	for i := range scan {
		if scan[i] != byHeap[i] {
			t.Fatalf("%s: %d bytes: part %d is %+v by scan, %+v by heap",
				c.GetName(), len(piece), i, scan[i], byHeap[i])
		}
	}
}

// The heap must reproduce the scan exactly, offsets and remaining ranks
// alike, on every encoding and on inputs that stress its tie-breaking.
func TestMergePairsHeapMatchesScan(t *testing.T) {
	const chinese = "网关在转发请求之前需要估算输入的令牌数量以便检查余额和选择超时档位" +
		"如果一段中文没有任何标点那么分词器会把整段当成一个片段再做合并"

	alphabets := []string{
		chinese,
		"abcdefghijklmnopqrstuvwxyz",
		"aaaaaaaaab",         // long runs of one letter: many equal ranks
		"0123456789",         // digits
		" \t\n",              // whitespace
		"=-*#/\\|_.:;",       // symbols
		"*́̀",                // symbols and combining marks
		"ab\xff\xfe\xe4\xb8", // invalid UTF-8
		"áéíóúñçãõ日本語テキスト한국어",
	}

	rng := rand.New(rand.NewPCG(20261021, 2))

	for _, c := range allCodecs() {
		for _, alphabet := range alphabets {
			runes := []rune(alphabet)
			bytes := []byte(alphabet)

			// Mostly around and above heapMergeThreshold, where the heap runs;
			// kept short enough that the quadratic scan stays quick under -race.
			for range 60 {
				length := 1 + rng.IntN(1200)

				var b strings.Builder
				for b.Len() < length {
					if rng.IntN(8) == 0 {
						b.WriteByte(bytes[rng.IntN(len(bytes))])
					} else {
						b.WriteRune(runes[rng.IntN(len(runes))])
					}
				}

				requireSameParts(t, c, b.String())
			}
		}

		// Either side of the dispatch, on input whose merges are not trivial.
		for _, length := range []int{
			heapMergeThreshold - 1, heapMergeThreshold, heapMergeThreshold + 1,
		} {
			for _, unit := range []string{chinese, "ab", " ", "0", "*́"} {
				piece := strings.Repeat(unit, length/len(unit)+1)[:length]
				requireSameParts(t, c, piece)
			}
		}

		for _, piece := range []string{
			"",
			"a",
			"ab",
			strings.Repeat("a", 4096),
			strings.Repeat("ab", 2048),
			strings.Repeat(" ", 4096),
			strings.Repeat("0", 4096),
			strings.Repeat(chinese, 40),
			strings.Repeat("*́", 2000),
		} {
			requireSameParts(t, c, piece)
		}
	}
}

// Above the threshold the two paths must also agree through Count, on whole
// texts that the pre-tokenizer splits into pieces of every length.
func TestCountUnchangedAcrossThreshold(t *testing.T) {
	const chinese = "网关在转发请求之前需要估算输入的令牌数量以便检查余额和选择超时档位"

	var b strings.Builder
	for i := range 200 {
		b.WriteString(strings.Repeat(chinese, i%7))
		b.WriteString(" func main() { fmt.Println(\"hello\") }\n")
		b.WriteString(strings.Repeat("=", i%90))
		b.WriteString(strings.Repeat(" ", i%70))
		b.WriteString("\n")
	}

	text := b.String()

	for _, c := range allCodecs() {
		byPieces := 0

		match, err := c.splitRegexp.FindStringMatch(text)
		if err != nil {
			t.Fatal(err)
		}

		for match != nil {
			piece := match.String()
			if _, ok := c.vocabulary[piece]; ok {
				byPieces++
			} else {
				byPieces += len(c.mergePairsScan(piece)) - 1
			}

			match, err = c.splitRegexp.FindNextMatch(match)
			if err != nil {
				t.Fatal(err)
			}
		}

		got, err := c.Count(text)
		if err != nil {
			t.Fatal(err)
		}

		if got != byPieces {
			t.Errorf("%s: Count %d, scan-only count %d", c.GetName(), got, byPieces)
		}
	}
}

// A single 180,000-byte piece takes minutes with the scan.
func TestMergePairsIsNotQuadratic(t *testing.T) {
	c := NewO200kBase()
	piece := strings.Repeat("网关在转发请求之前需要估算输入的令牌数量", 3000)

	start := time.Now()
	parts := c.mergePairs(piece)

	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Fatalf("merging %d bytes took %s", len(piece), elapsed)
	}

	if len(parts) < 2 {
		t.Fatalf("got %d parts", len(parts))
	}
}

func BenchmarkMergePairs(b *testing.B) {
	c := NewO200kBase()
	source := strings.Repeat("网关在转发请求之前需要估算输入的令牌数量", 1000)

	for _, length := range []int{8, 16, 32, 48, 64, 96, 128, 256, 1024, 8192} {
		piece := source[:length/3*3]

		b.Run("scan/"+strconv.Itoa(len(piece)), func(b *testing.B) {
			for b.Loop() {
				c.mergePairsScan(piece)
			}
		})
		b.Run("heap/"+strconv.Itoa(len(piece)), func(b *testing.B) {
			for b.Loop() {
				c.mergePairsHeap(piece)
			}
		})
	}
}
