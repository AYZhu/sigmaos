package wc

//
// a word-count application for MapReduce.
//

import (
	"bufio"
	"io"
	"strconv"
	"unicode"
	"unicode/utf8"

	"ulambda/mr"
)

//
// The map function is called once for each file of input. The first
// argument is the name of the input file, and the second is the
// file's complete contents. You should ignore the input file name,
// and look only at the contents argument. The return value is a slice
// of key/value pairs.
//

func scanWords(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip leading non letters
	start := 0
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if unicode.IsLetter(r) {
			break
		}
	}
	// Scan until non letter
	for width, i := 0, start; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])
		if !unicode.IsLetter(r) {
			return i + width, data[start:i], nil
		}
	}
	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}
	// Request more data.
	return start, nil, nil

}

func Map(filename string, rdr io.Reader) []mr.KeyValue {
	scanner := bufio.NewScanner(rdr)
	scanner.Split(scanWords)

	kva := make([]mr.KeyValue, 0)
	for scanner.Scan() {
		kv := mr.KeyValue{scanner.Text(), "1"}
		kva = append(kva, kv)
	}
	return kva
}

//
// The reduce function is called once for each key generated by the
// map tasks, with a list of all the values created for that key by
// any map task.
//
func Reduce(key string, values []string) string {
	// return the number of occurrences of this word.
	return strconv.Itoa(len(values))
}
