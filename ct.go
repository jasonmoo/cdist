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
	"sync"
	"time"
	"unicode"

	"github.com/jasonmoo/oc"
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

	filter := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}

	switch {
	case *words:

		dict := oc.NewOc()

		if *stem {

			const chunksize = 64 << 10

			stems := make(chan string, chunksize)
			stemswg := new(sync.WaitGroup)
			stemswg.Add(1)
			go func() {
				for word := range stems {
					dict.Increment(word, 1)
				}
				stemswg.Done()
			}()

			wg := new(sync.WaitGroup)

			for _, input := range inputs {
				br := bufio.NewReaderSize(input, 1<<20)
				for {
					buf := make([]byte, chunksize)

					n, err := br.Read(buf)
					if err != nil && err != io.EOF {
						log.Fatal(err)
					}
					if n == 0 {
						break
					}
					buf = buf[:n]

					// read to a newline to prevent
					// spanning a word across two buffers
					if buf[len(buf)-1] != '\n' {
						extra, err := br.ReadBytes('\n')
						if err != nil && err != io.EOF {
							log.Fatal(err)
						}
						buf = append(buf, extra...)
					}

					wg.Add(1)
					go func(b []byte) {
						for _, word := range strings.FieldsFunc(string(b), filter) {
							stems <- english.Stem(word, false)
						}
						wg.Done()
					}(buf)
				}
			}

			wg.Wait()
			close(stems)
			stemswg.Wait()

		} else {
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

					for _, word := range strings.FieldsFunc(line, filter) {
						dict.Increment(word, 1)
					}
				}
			}
		}

		dict.SortByCt(oc.DESC)

		for dict.Next() {
			key, ct := dict.KeyValue()
			fmt.Printf("%s: %d\n", key, ct)
		}

		fmt.Fprintf(os.Stderr, "%d words counted in %s\n", dict.Len(), time.Since(start))

	default:

		dict := oc.NewOc()

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
				dict.Increment(strconv.QuoteRune(r), 1)
			}
		}

		dict.SortByCt(oc.DESC)

		for dict.Next() {
			key, ct := dict.KeyValue()
			fmt.Printf("%s: %d\n", key, ct)
		}

		fmt.Fprintf(os.Stderr, "%d runes counted in %s\n", dict.Len(), time.Since(start))

	}

}
