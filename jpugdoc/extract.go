package jpugdoc

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
)

type Pair struct {
	en string
	ja string
}

// コメント（英語原文）と続く文書（日本語翻訳）のペア、残り文字列、エラーを返す
// <!--
// english
// -->
// japanese
// の形式に一致しない場合はエラーを返す
func enjaPair(para []byte) (Pair, []byte, error) {
	re := EXCOMMENT.FindSubmatch(para)
	if len(re) < 3 {
		return Pair{}, nil, fmt.Errorf("no match")
	}
	enstr := strings.ReplaceAll(string(re[1]), "\n", " ")
	enstr = MultiSpace.ReplaceAllString(enstr, " ")
	enstr = strings.TrimSpace(enstr)

	jastr := strings.TrimSpace(string(re[2]))
	pair := Pair{
		en: enstr,
		ja: jastr,
	}

	if string(re[3]) == "<!--" {
		return pair, para[len(re[0])+3:], nil
	}
	if string(re[3]) == "<itemizedlist>" {
		left := para[len(re[0])+14:]
		left = bytes.ReplaceAll(left, []byte("<listitem>"), []byte(""))
		return pair, left, nil
	}
	return pair, nil, nil
}

func enCandidate(en string) string {
	en = RECOMMENTSTART.ReplaceAllString(en, "")
	en = RECOMMENTEND.ReplaceAllString(en, "")
	en = MultiSpace.ReplaceAllString(en, " ")
	en = strings.ReplaceAll(en, "\n", " ")
	en = strings.TrimSpace(en)
	return en
}

// src を原文と日本語訳の対の配列に変換する
func Extraction(src []byte) []Pair {
	var pairs []Pair

	// title
	for _, titles := range RECHECKTITLE.FindAll(src, -1) {
		ts := RETITLE.FindAll(titles, -1)
		if len(ts) == 2 {
			pair := Pair{
				en: string(ts[0]),
				ja: string(ts[1]),
			}
			pairs = append(pairs, pair)
		}
	}

	paras := REPARA.FindAll([]byte(src), -1)
	en := ""
	for _, para := range paras {
		pair, left, err := enjaPair(para)
		if err != nil {
			if STARTCOMMENT.Match(para) && en == "" {
				en = enCandidate(string(para))
			} else {
				pair.en = en
				pair.ja = string(para)
				en = ""
			}
		}
		pairs = append(pairs, pair)
		for len(left) > 0 {
			tmpLeft := left
			pair, left, err = enjaPair(left)
			if err != nil {
				if en == "" {
					en = enCandidate(string(tmpLeft))
				} else {
					pair.en = en
					pair.ja = string(para)
					en = ""
				}
				continue
			}
			pairs = append(pairs, pair)
		}
	}

	rows := REROWS.FindAll([]byte(src), -1)
	for _, row := range rows {
		re := EXCOMMENT.FindSubmatch(row)
		if len(re) < 3 {
			continue
		}
		enstr := string(re[1])
		enstr = ENTRYSTRIP.ReplaceAllString(enstr, "")
		enstr = MultiNL.ReplaceAllString(enstr, " ")
		enstr = MultiSpace.ReplaceAllString(enstr, " ")
		enstr = strings.TrimSpace(enstr)

		jastr := string(re[2])
		jastr = ENTRYSTRIP.ReplaceAllString(jastr, "")
		jastr = strings.TrimSpace(jastr)

		pair := Pair{
			en: enstr,
			ja: jastr,
		}
		pairs = append(pairs, pair)
	}
	return pairs
}

func Extract(fileNames []string) {
	for _, fileName := range fileNames {
		src, err := ReadAllFile(fileName)
		if err != nil {
			log.Fatal(err)
		}

		pairs := Extraction(src)

		dicname := DICDIR + fileName + ".t"
		f, err := os.Create(dicname)
		if err != nil {
			log.Fatal(err)
		}

		for _, pair := range pairs {
			fmt.Fprintf(f, "␝%s␟", pair.en)
			fmt.Fprintf(f, "%s␞\n", pair.ja)
		}
		f.Close()
	}
}
