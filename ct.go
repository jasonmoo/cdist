package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/kljensen/snowball/english"
)

var (
	words = flag.Bool("w", false, "show word dist")
	stem  = flag.Bool("stem", false, "snowball stem that shit")
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
}

func main() {

	var inputs []io.Reader

	if flag.NArg() > 0 {
		for _, file_path := range flag.Args() {
			file, err := os.Open(file_path)
			if err != nil {
				log.Fatal(err)
			}
			inputs = append(inputs, file)
		}
	} else {
		inputs = []io.Reader{os.Stdin}
	}

	start := time.Now()

	switch {
	case *words:

		dict := make(map[string]int)

		for _, input := range inputs {
			buf := bufio.NewReader(input)
			for {
				line, err := buf.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						break
					}
					log.Fatal(err)
				}

				words := strings.Fields(line)
				if *stem {
					for i, word := range words {
						words[i] = english.Stem(word, false)
					}
				}

				for _, word := range words {
					dict[word]++
				}
			}
		}

		for word, ct := range dict {
			fmt.Printf("%s: %d\n", word, ct)
		}

		fmt.Fprintf(os.Stderr, "%d words counted in %s\n", len(dict), time.Since(start))

	default:

		dict := make(map[rune]int)

		for _, input := range inputs {
			buf := bufio.NewReader(input)
			for {
				r, _, err := buf.ReadRune()
				if err != nil {
					if err == io.EOF {
						break
					}
					log.Fatal(err)
				}
				dict[r]++
			}
		}

		for c, ct := range dict {
			fmt.Printf("%s: %d\n", strconv.QuoteRune(c), ct)
		}

		fmt.Fprintf(os.Stderr, "%d runes counted in %s\n", len(dict), time.Since(start))

	}

}
