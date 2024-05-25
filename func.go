package main

import (
	"bufio"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"strings"

	fdk "github.com/fnproject/fdk-go"
)

var (
	//go:embed public
	fs embed.FS
)

var upper = strings.NewReplacer(
	"ぁ", "あ",
	"ぃ", "い",
	"ぅ", "う",
	"ぇ", "え",
	"ぉ", "お",
	"ゃ", "や",
	"ゅ", "ゆ",
	"ょ", "よ",
)

func kana2hira(s string) string {
	return strings.Map(func(r rune) rune {
		if 0x30A1 <= r && r <= 0x30F6 {
			return r - 0x0060
		}
		return r
	}, s)
}

func hira2kana(s string) string {
	return strings.Map(func(r rune) rune {
		if 0x3041 <= r && r <= 0x3096 {
			return r + 0x0060
		}
		return r
	}, s)
}

func search(text string) (string, error) {
	rs := []rune(text)
	r := rs[len(rs)-1]

	f, err := fs.Open("public/dict.txt")
	if err != nil {
		return "", err
	}
	defer f.Close()
	buf := bufio.NewReader(f)

	words := []string{}
	for {
		b, _, err := buf.ReadLine()
		if err != nil {
			break
		}
		line := string(b)
		if ([]rune(line))[0] == r {
			words = append(words, line)
		}
	}
	if len(words) == 0 {
		return "", errors.New("empty dictionary")
	}
	return words[rand.Int()%len(words)], nil
}

func shiritori(text string) (string, error) {
	text = strings.Replace(text, "ー", "", -1)
	if rand.Int()%2 == 0 {
		text = hira2kana(text)
	} else {
		text = kana2hira(text)
	}
	return search(text)
}

func handleText(text string) (string, error) {
	rs := []rune(strings.TrimSpace(text))
	if len(rs) == 0 {
		return "", errors.New("なんやねん")
	}
	if rs[len(rs)-1] == 'ん' || rs[len(rs)-1] == 'ン' {
		return "", errors.New("出直して来い")
	}
	s, err := shiritori(text)
	if err != nil {
		return "", err
	}
	if s == "" {
		return "", errors.New("わかりません")
	}
	rs = []rune(s)
	if rs[len(rs)-1] == 'ん' || rs[len(rs)-1] == 'ン' {
		s += "\nあっ..."
	}
	return s, nil
}

func main() {
	fdk.Handle(fdk.HandlerFunc(myHandler))
}

type Siritori struct {
	Word string `json:"word,omitempty"`
	Err  string `json:"err,omitempty"`
}

func myHandler(ctx context.Context, in io.Reader, out io.Writer) {
	var s Siritori
	json.NewDecoder(in).Decode(&s)
	var err error
	s.Word, err = handleText(s.Word)
	if err != nil {
		s.Err = err.Error()
	}
	fdk.SetHeader(out, "content-type", "application/json")
	json.NewEncoder(out).Encode(&s)
}
