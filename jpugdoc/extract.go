package jpugdoc

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

type Catalog struct {
	pre      string
	en       string
	ja       string
	cdatapre string
}

func versionTag() (string, error) {
	versionFile := "version.sgml"
	src, err := ReadAllFile(versionFile)
	if err != nil {
		return "", err
	}
	ver := regexp.MustCompile(`<!ENTITY version "([0-9\.]+)">`)
	re := ver.FindSubmatch(src)
	if len(re) < 1 {
		return "", fmt.Errorf("no version")
	}
	v := strings.ReplaceAll(string(re[1]), ".", "_")
	tag := fmt.Sprintf("REL_%s", v)
	return tag, nil
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

// diff を原文と日本語訳の対の配列に変換する
func Extraction(diffSrc []byte) []Catalog {
	reader := bytes.NewReader(diffSrc)
	scanner := bufio.NewScanner(reader)
	var en, ja, index, indexj strings.Builder
	pre := make([]string, 4)
	prefix := ""
	cdatapre := ""
	var similer bool
	var pairs []Catalog
	var comment, jadd, indexF bool
	for i := 0; i < 3; i++ {
		if !scanner.Scan() {
			//fmt.Println("scaner error")
			return pairs
		}
	}

	for scanner.Scan() {
		l := scanner.Text()
		line := strings.TrimSpace(l)
		pre = append(pre[1:], l[1:])
		// マッチ度コメントをスキップ
		if similer || strings.Contains(line, "マッチ度[") {
			if strings.Contains(line, "-->") {
				similer = false
				continue
			}
			similer = true
			continue
		}

		if m := STARTADDCOMMENTWITHC.FindAllStringSubmatch(line, 1); len(m) > 0 {
			pair := Catalog{
				pre:      prefix,
				en:       strings.Trim(en.String(), "\n"),
				ja:       strings.Trim(ja.String(), "\n"),
				cdatapre: cdatapre,
			}
			if en.Len() != 0 {
				pairs = append(pairs, pair)
			}
			en.Reset()
			ja.Reset()
			prefix = pre[0]
			if len(m[0]) == 1 {
				en.WriteString("\n")
				comment = true
				continue
			}
			cdatapre = strings.Join(m[0][1:], "")
			en.WriteString(cdatapre)
			en.WriteString("\n")
			comment = true
			continue
		} else if STARTADDCOMMENT.MatchString(line) {
			pair := Catalog{
				pre: prefix,
				en:  strings.Trim(en.String(), "\n"),
				ja:  strings.Trim(ja.String(), "\n"),
			}
			if en.Len() != 0 {
				pairs = append(pairs, pair)
			}
			en.Reset()
			ja.Reset()
			prefix = pre[0]
			en.WriteString("\n")
			comment = true
			continue
		} else if ENDADDCOMMENT.MatchString(line) || ENDADDCOMMENTWITHC.MatchString(line) {
			comment = false
			jadd = true
			continue
		}

		if comment {
			if l[0] == '-' {
				continue
			}
			l = REPHIGHHUN.ReplaceAllString(l, "-")
			en.WriteString(l[1:])
			en.WriteString("\n")
		} else {
			if jadd && strings.HasPrefix(l, "+") {
				ja.WriteString(l[1:])
				ja.WriteString("\n")
			} else {
				jadd = false
			}
		}

		if comment {
			continue
		}

		// indexterm
		if !strings.HasPrefix(l, "+") {
			// original
			if STARTINDEXTERM.MatchString(line) {
				index.Reset()
				indexF = true
				if ENDINDEXTERM.MatchString(line) {
					index.WriteString(l[1:])
					index.WriteString("\n")
					indexF = false
				}
			} else if ENDINDEXTERM.MatchString(line) {
				index.WriteString(l[1:])
				index.WriteString("\n")
				indexF = false
			}
			if indexF {
				index.WriteString(l[1:])
				index.WriteString("\n")
			}
		} else {
			// Add ja indexterm
			if STARTINDEXTERM.MatchString(line) {
				indexF = true
				if ENDINDEXTERM.MatchString(line) {
					indexj.WriteString(l[1:])
					indexj.WriteString("\n")
					indexF = false
					pair := Catalog{
						pre: index.String(),
						ja:  strings.Trim(indexj.String(), "\n"),
					}
					pairs = append(pairs, pair)
					index.Reset()
					indexj.Reset()
				}
			} else if ENDINDEXTERM.MatchString(line) {
				indexj.WriteString(l[1:])
				indexj.WriteString("\n")
				indexF = false
				pair := Catalog{
					pre: index.String(),
					ja:  strings.Trim(indexj.String(), "\n"),
				}
				pairs = append(pairs, pair)
				index.Reset()
				indexj.Reset()
			}
			if indexF {
				indexj.WriteString(l[1:])
				indexj.WriteString("\n")
			}
			if !indexF && !jadd {
				if !strings.Contains(l, "</indexterm>") {
					if pre[0] != "" && pre[1] != "" && pre[2] != "" && pre[3] != "" {
						if l != "+" {
							pair := Catalog{
								pre: strings.Join(pre[:len(pre)-1], "\n") + "\n",
								ja:  l[1:],
							}
							pairs = append(pairs, pair)
						}
					}
				}

			}
		}
	}
	// last
	if en.Len() != 0 {
		pair := Catalog{
			pre: prefix,
			en:  strings.Trim(en.String(), "\n"),
			ja:  strings.Trim(ja.String(), "\n"),
		}
		pairs = append(pairs, pair)
	}

	return pairs
}

/*
	func extractFromDiff(fileName string, diff []byte) {
		pairs := Extraction(diff)
		writeDIC(fileName, pairs)
	}
*/
func Extract(fileNames []string) {
	vTag, err := versionTag()
	if err != nil {
		log.Fatal(err)
	}

	for _, fileName := range fileNames {
		src := getDiff(vTag, fileName)
		pairs := Extraction(src)
		writeDIC(fileName, pairs)
	}
}

func writeDIC(fileName string, pairs []Catalog) {
	dicname := DICDIR + fileName + ".t"
	f, err := os.Create(dicname)
	if err != nil {
		log.Fatal(err)
	}

	for _, pair := range pairs {
		fmt.Fprintf(f, "␝%s␟", pair.pre)
		fmt.Fprintf(f, "%s␟", pair.en)
		fmt.Fprintf(f, "%s␞", pair.ja)
		fmt.Fprintf(f, "%s␞\n", pair.cdatapre)
	}
	f.Close()
}
