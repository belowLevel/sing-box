package rule

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"os"
	"strings"
	"testing"
	"unicode"
)

func TestTxt2json(t *testing.T) {
	txtFile := "../../dist/xx.txt"
	jsonFile := "../../dist/xx.json"
	type Rule struct {
		DomainSuffix []string `json:"domain_suffix"`
	}

	type JsonStrct struct {
		Version int     `json:"version"`
		Rules   []*Rule `json:"rules"`
	}

	data := JsonStrct{
		Version: 2,
		Rules: []*Rule{
			{
				DomainSuffix: []string{},
			},
		},
	}
	f, err := os.OpenFile(txtFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		line = strings.TrimFunc(line, func(r rune) bool {
			return !unicode.IsGraphic(r)
		})
		data.Rules[0].DomainSuffix = append(data.Rules[0].DomainSuffix, "."+line)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}

	bytes, err := json.Marshal(&data)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(jsonFile, bytes, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}

func newSet(jsonFile string) (*LocalRuleSet, error) {
	return NewLocalRuleSet(context.Background(), log.NewNOPFactory().Logger(), "test", option.RuleSet{
		Type:         "local",
		Format:       "source",
		LocalOptions: option.LocalRuleSet{Path: jsonFile},
	})
}

func Benchmark(b *testing.B) {
	jsonFile := "../../dist/xx.json"
	set, err := newSet(jsonFile)
	if err != nil {
		b.Fatal(err)
	}

	metadata := &adapter.InboundContext{
		Domain: "x.alibaba.cdn.steampipe.steamcontent.com",
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		set.Match(metadata)
	}
}

//Benchmark-16             1000000              1033 ns/op              96 B/op
//2 allocs/op
//PASS
