package jpugdoc

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jwalton/gchalk"
)

// Conflict は、原文が同じで日本語訳が異なるものを抽出して標準出力に表示する。
func Conflict(vTag string, fileNames []string) error {
	common, err := extract(vTag, true, fileNames)
	if err != nil {
		return fmt.Errorf("Conflict: %w", err)
	}
	common = catalogsSplits(common)
	seen := toSeen(common)

	notEq := findNotSameCommon(seen)
	for _, catalog := range notEq {
		fmt.Println(gchalk.Green(catalog.en))
		fmt.Println(catalog.ja)
		fmt.Println()
	}
	return nil
}

// toSeen はカタログの配列を原文をキーとして、日本語訳の配列を値とするマップに変換する
func toSeen(catalogs Catalogs) map[string][]string {
	seen := make(map[string][]string)
	for _, catalog := range catalogs {
		if catalog.en == "" || catalog.ja == "" || catalog.ja == "no translation" {
			continue
		}
		en := stripNL(catalog.en)
		en = strings.Join(strings.Fields(en), " ")
		// 最初の連続スペースを一つのスペースに変換
		ja := STARTSPACE.ReplaceAllString(catalog.ja, " ")
		seen[en] = append(seen[en], ja)
	}
	seen = addTitle(seen)
	return seen
}

// seenから、原文が同じで日本語訳が異なるものを抽出する
func findNotSameCommon(seen map[string][]string) Catalogs {
	unique := make(map[string]Catalog)
	for en, jas := range seen {
		if len(en) <= 12 { // 原文が12文字以下のものは除外する
			continue
		}
		if len(jas) <= 1 {
			continue
		}
		uniqJa := uniqueStrings(jas)
		if len(uniqJa) > 1 {
			unique[en] = Catalog{en: en, ja: strings.Join(uniqJa, "\n")}
		}
	}

	return sortedCatalogs(unique)
}

func sortedCatalogs(unique map[string]Catalog) Catalogs {
	uniques := Catalogs{}
	for _, catalog := range unique {
		uniques = append(uniques, catalog)
	}
	sort.Slice(uniques, func(i, j int) bool {
		return uniques[i].en < uniques[j].en
	})
	return uniques
}

func uniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	unique := []string{}

	for _, v := range slice {
		j := stripNL(v)
		if !seen[j] {
			seen[j] = true
			unique = append(unique, v)
		}
	}

	return unique
}
