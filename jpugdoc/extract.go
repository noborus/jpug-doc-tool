package jpugdoc

import (
	"bytes"
	"fmt"
	"strings"
)

type Catalog struct {
	pre      string
	en       string
	ja       string
	cdatapre string
}

// 一つのファイルから原文と日本語訳の対の配列を抽出する
func PARAExtraction(src []byte) []Catalog {
	var pairs []Catalog

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

		pair := Catalog{
			en: enstr,
			ja: jastr,
		}
		pairs = append(pairs, pair)
	}
	return pairs
}

// コメント（英語原文）と続く文書（日本語翻訳）のペア、残り文字列、エラーを返す
// <!--
// english
// -->
// japanese
// の形式に一致しない場合はエラーを返す
func enjaPair(para []byte) (Catalog, []byte, error) {
	re := EXCOMMENT.FindSubmatch(para)
	if len(re) < 3 {
		return Catalog{}, nil, fmt.Errorf("no match")
	}
	enstr := strings.ReplaceAll(string(re[1]), "\n", " ")
	enstr = MultiSpace.ReplaceAllString(enstr, " ")
	enstr = strings.TrimSpace(enstr)

	jastr := strings.TrimSpace(string(re[2]))
	pair := Catalog{
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
