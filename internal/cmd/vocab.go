package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	packageName = "codec"
)

type config struct {
	url      string
	mapName  string
	filename string
}

func main() {
	encoding := flag.String("encoding", "", "encoding format. (e.g. cl100k_base)")
	flag.Parse()

	if encoding == nil {
		flag.PrintDefaults()
		os.Exit(1)
	}

	cfg := getConfig(*encoding)

	file, err := os.Create(cfg.filename)
	if err != nil {
		log.Fatal("error creating file: %v", err)
	}
	defer file.Close()

	generatePreable(file, *encoding)
	genVocabulary(file, cfg.mapName, cfg.url)
}

func generatePreable(w io.Writer, encoding string) {
	fmt.Fprintf(w, "package %s\n", packageName)
	fmt.Fprintf(w, "//go:generate go run ../internal/cmd/vocab.go -encoding %s\n", encoding)
	fmt.Fprintf(w, "// THIS FILE WAS AUTOMATICALLY GENERATED. DO NOT MODIFY\n")
}

func genVocabulary(w io.Writer, mapName string, uri string) {
	resp, err := http.Get(uri)
	if err != nil {
		log.Fatalf("error fetching file: %v", err)
	}
	defer resp.Body.Close()

	fmt.Fprintf(w, "var %s vocab = vocab{\n", mapName)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")

		if len(parts) != 2 {
			log.Fatalf("invalid line: %s", line)
		}

		word, err := base64.StdEncoding.DecodeString(parts[0])
		if err != nil {
			log.Fatalf("invalid word: %s", parts[0])
		}

		fmt.Fprintf(w, "	%s:%s,\n", strconv.Quote(string(word)), parts[1])
	}

	fmt.Fprintf(w, "}\n\n")
}

func getConfig(encoding string) config {
	switch encoding {
	case "cl100k_base":
		return config{
			mapName:  "cl100kBaseVocab",
			url:      "https://openaipublic.blob.core.windows.net/encodings/cl100k_base.tiktoken",
			filename: "cl100k_base_vocab.go",
		}
	case "r50k_base":
		return config{
			mapName:  "r50kBaseVocab",
			url:      "https://openaipublic.blob.core.windows.net/encodings/r50k_base.tiktoken",
			filename: "r50k_base_vocab.go",
		}
	case "p50k_base":
		return config{
			mapName:  "p50kBaseVocab",
			url:      "https://openaipublic.blob.core.windows.net/encodings/p50k_base.tiktoken",
			filename: "p50k_base_vocab.go",
		}
	default:
		log.Fatal("config not found")
		return config{}
	}
}
