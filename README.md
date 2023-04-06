# Tokenizer

This is a pure go port of OpenAI's tokenizer.

<a href="https://www.buymeacoffee.com/mwahlmann" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/default-blue.png" alt="Buy Me A Coffee" height="41" width="174"></a>

## Usage

```go
package main

import (
    "fmt"
	"github.com/tiktoken-go/tokenizer"
)

func main {
    enc, err := tokenizer.Get(tokenizer.Cl100kBase)
    if err != nil {
        panic("oh oh")
    }

    // this should print a list of token ids
    ids, token, _ := enc.Encode("supercalifragilistic")
    fmt.Println(ids)

    // this should print the original string back
    text, _ := enc.Decode(ids)
    fmt.Println(text)
}
```

## Caveats

This library embeds OpenAI's vocabularies - which are not small (~4Mb) - as go
maps. This is different than what the way python version of tiktoken works, 
which downloads the dictionaries and puts them in a cache folder.

However, since the dictionaries are compiled during the go build process
the performance and start-up times should be better than downloading and loading
them at runtime.

## Alternatives

Here is a list of other libraries that do something similar.

- [https://github.com/sugarme/tokenizer](https://github.com/sugarme/tokenizer)
- [https://github.com/pandodao/tokenizer-go](https://github.com/pandodao/tokenizer-go)


