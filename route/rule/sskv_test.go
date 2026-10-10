package rule

import (
	"bufio"
	"errors"
	"os"
	"strings"
	"testing"
	"unicode"
)

func newKvSet(domainTxtFile string) (*Set, error) {
	if _, err := os.Stat(domainTxtFile); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(domainTxtFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	scanner := bufio.NewScanner(f)
	var strs []string
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		line = strings.TrimFunc(line, func(r rune) bool {
			return !unicode.IsGraphic(r)
		})
		strs = append(strs, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(strs) == 0 {
		return nil, errors.New("This file has no domain.")
	}
	return NewSet(strs), nil
}

func BenchmarkSetHas(b *testing.B) {
	domainTxtFile := "../../dist/xx.txt"

	set, err := newKvSet(domainTxtFile)
	if err != nil {
		b.Error(err)
		return
	}

	domain := "x.alibaba.cdn.steampipe.steamcontent.com"
	b.Logf("size %.3f MB", float64(set.Size())/(1024*1024))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		set.Has(domain)
	}
}

//BenchmarkSetHas-16       1567300               726.5 ns/op             0 B/op
//0 allocs/op
//PASS
