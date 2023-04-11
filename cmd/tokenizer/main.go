package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/tiktoken-go/tokenizer"
)

func main() {
	model := flag.String("model", "gpt-3.5-turbo", "the target OpenAI model to generate tokens for")
	encoding := flag.String("encoding", "", "the encoding format. (not that you can't specify both model and encoding)")
	encode := flag.String("encode", "", "text to encode")
	decode := flag.String("decode", "", "space separated list of token ids to decode")
	emitTokens := flag.Bool("tokens", false, "if true will output the tokens instead of the token ids")
	listModels := flag.Bool("list-models", false, "list all supported models")
	listEncodings := flag.Bool("list-encodings", false, "list all supported encoding formats")
	flag.Parse()

	if *listModels {
		printModels()
		os.Exit(0)
	}

	if *listEncodings {
		printEncodings()
		os.Exit(0)
	}

	// either model or encoding should be specified
	if (*model != "" && *encoding != "") || (*model == "" && *encoding == "") {
		flag.PrintDefaults()
	}

	// either encode or decode operations should be requested
	if (*encode != "" && *decode != "") || (*encode == "" && *decode == "") {
		flag.PrintDefaults()
	}

	codec := getCodec(*model, *encoding)

	if *encode != "" {
		encodeInput(codec, *encode, *emitTokens)
	} else {
		decodeInput(codec, *decode+" "+strings.Join(flag.Args(), " "))
	}
}

func getCodec(model, encoding string) tokenizer.Codec {
	if model != "" {
		c, err := tokenizer.ForModel(tokenizer.Model(model))
		if err != nil {
			log.Fatalf("error creating tokenizer: %v", err)
		}
		return c
	} else {
		c, err := tokenizer.Get(tokenizer.Encoding(encoding))
		if err != nil {
			log.Fatalf("error creating tokenizer: %v", err)
		}
		return c
	}
}

func encodeInput(codec tokenizer.Codec, text string, wantTokens bool) {
	ids, tokens, err := codec.Encode(text)
	if err != nil {
		log.Fatalf("error encoding: %v", err)
	}

	if wantTokens {
		fmt.Println(strings.Join(tokens, " "))
	} else {
		var textIds []string
		for _, id := range ids {
			textIds = append(textIds, strconv.Itoa(int(id)))
		}
		fmt.Println(strings.Join(textIds, " "))
	}
}

func decodeInput(codec tokenizer.Codec, tokens string) {
	var ids []uint
	for _, t := range strings.Split(tokens, " ") {
		id, err := strconv.Atoi(t)
		if err != nil {
			log.Fatalf("invalid token id: %s", t)
		}
		ids = append(ids, uint(id))
	}

	text, err := codec.Decode(ids)
	if err != nil {
		log.Fatalf("error decoding: %v", err)
	}
	fmt.Println(text)
}

func printEncodings() {
	encodings := []tokenizer.Encoding{
		tokenizer.R50kBase,
		tokenizer.P50kBase,
		tokenizer.P50kEdit,
		tokenizer.Cl100kBase,
	}

	for _, e := range encodings {
		fmt.Println(e)
	}
}

func printModels() {
	models := []tokenizer.Model{tokenizer.GPT4,
		tokenizer.GPT35Turbo,
		tokenizer.TextEmbeddingAda002,
		tokenizer.TextDavinci003,
		tokenizer.TextDavinci002,
		tokenizer.CodeDavinci002,
		tokenizer.CodeDavinci001,
		tokenizer.CodeCushman002,
		tokenizer.CodeCushman001,
		tokenizer.DavinciCodex,
		tokenizer.CushmanCodex,
		tokenizer.TextDavinci001,
		tokenizer.TextCurie001,
		tokenizer.TextBabbage001,
		tokenizer.TextAda001,
		tokenizer.Davinci,
		tokenizer.Curie,
		tokenizer.Babbage,
		tokenizer.Ada,
		tokenizer.TextSimilarityDavinci001,
		tokenizer.TextSimilarityCurie001,
		tokenizer.TextSimilarityBabbage001,
		tokenizer.TextSimilarityAda001,
		tokenizer.TextSearchDavinciDoc001,
		tokenizer.TextSearchCurieDoc001,
		tokenizer.TextSearchAdaDoc001,
		tokenizer.TextSearchBabbageDoc001,
		tokenizer.CodeSearchBabbageCode001,
		tokenizer.CodeSearchAdaCode001,
		tokenizer.TextDavinciEdit001,
		tokenizer.CodeDavinciEdit001}

	for _, m := range models {
		fmt.Println(m)
	}
}
