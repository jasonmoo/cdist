package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"unicode/utf8"
)

func main() {

	dict := make(map[rune]int)

	for {
		buf := make([]byte, 1<<20)

		n, err := os.Stdin.Read(buf)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		if n == 0 {
			break
		}

		buf = buf[:n]

		for len(buf) > 0 {
			r, size := utf8.DecodeRune(buf)
			dict[r]++
			buf = buf[size:]
		}
	}

	for c, ct := range dict {
		fmt.Printf("%s: %d\n", strconv.QuoteRune(c), ct)
	}

}
