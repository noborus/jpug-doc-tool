package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

var REPARA = regexp.MustCompile(`(?s)(<para>\n*)(.*?)(\s+</para>)`)
var REROWS = regexp.MustCompile(`(?s)(<row>\n*)(.*?)(\s+</row>)`)
var RECOMMENT = regexp.MustCompile(`(?s)<!--(.*?)-->`)
var EXCOMMENT = regexp.MustCompile(`(?s)<!--(.*?)-->(.*?)(</row>|<synopsis>|<variablelist>|<programlisting>|<itemizedlist>|</para>)`)
var ENTRYSTRIP = regexp.MustCompile(`</?entry>`)

type Pair struct {
	en string
	ja string
}

var Catalog []Pair

func paraReplace(src []byte) []byte {
	if RECOMMENT.Match(src) {
		return src
	}
	re := REPARA.FindSubmatch(src)
	en := strings.TrimRight(string(re[2]), "\n")

	// fmt.Printf("[%s]\n", strings.TrimSpace(en))
	for _, cata := range Catalog {
		if cata.en == strings.TrimSpace(en) {
			para := fmt.Sprintf("$1<!--\n%s\n-->\n%s$3", en, strings.TrimRight(cata.ja, "\n"))
			ret := REPARA.ReplaceAll(src, []byte(para))
			return ret
		}
	}
	return src
}

var SPLITCATALOG = regexp.MustCompile(`(?s)｛([^｝]*)｝([^｛]*)`)

func loadCatalog(fileName string) {
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	src, err := io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	catas := SPLITCATALOG.FindAll(src, -1)
	for _, cata := range catas {
		re := SPLITCATALOG.FindSubmatch(cata)
		en := string(re[1])
		ja := string(re[2])
		pair := Pair{
			en: en,
			ja: ja,
		}
		Catalog = append(Catalog, pair)
	}
}

func Extraction(src []byte) []Pair {
	var pairs []Pair
	paras := REPARA.FindAll([]byte(src), -1)
	for _, para := range paras {
		re := EXCOMMENT.FindSubmatch(para)
		if len(re) < 3 {
			continue
		}
		enstr := strings.TrimSpace(string(re[1]))
		jastr := strings.TrimSpace(string(re[2]))
		pair := Pair{
			en: enstr,
			ja: jastr,
		}
		pairs = append(pairs, pair)
	}
	rows := REROWS.FindAll([]byte(src), -1)
	for _, row := range rows {
		re := EXCOMMENT.FindSubmatch(row)
		if len(re) < 3 {
			continue
		}
		enstr := string(re[1])
		enstr = ENTRYSTRIP.ReplaceAllString(enstr, "")
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

func main() {
	replace := true

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	src, err := io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	if replace {
		loadCatalog(os.Args[2])
		fu := paraReplace
		ret := REPARA.ReplaceAllFunc(src, fu)

		// _ = string(ret)
		fmt.Println(string(ret))
		return
	}

	pairs := Extraction(src)
	for _, pair := range pairs {
		fmt.Printf("｛%s｝", pair.en)
		fmt.Printf("%s\n", pair.ja)
	}
}
