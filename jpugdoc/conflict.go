package jpugdoc

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/jwalton/gchalk"
)

const (
	ConflictFormatText = "text"
	ConflictFormatGHMD = "gh-md"
)

type ConflictOption struct {
	Format string
}

type conflictCatalog struct {
	en  string
	jas map[string]int
}

// Conflict は、原文が同じで日本語訳が異なるものを抽出して標準出力に表示する。
func Conflict(vTag string, fileNames []string, opt ConflictOption) error {
	if opt.Format == "" {
		opt.Format = ConflictFormatText
	}

	common, err := extract(vTag, true, fileNames)
	if err != nil {
		return fmt.Errorf("Conflict: %w", err)
	}
	common = catalogsSplits(common)
	seen := toSeen(common)

	return writeConflict(os.Stdout, findNotSameCommon(seen), opt)
}

func writeConflict(w io.Writer, catalogs []conflictCatalog, opt ConflictOption) error {
	switch opt.Format {
	case ConflictFormatText:
		for _, catalog := range catalogs {
			fmt.Fprintln(w, gchalk.Green(catalog.en))
			for ja := range catalog.jas {
				fmt.Fprintln(w, ja)
			}
			fmt.Fprintln(w)
		}
	case ConflictFormatGHMD:
		for i, catalog := range catalogs {
			fmt.Fprintf(w, "### Conflict %d\n\n", i+1)
			fmt.Fprintln(w, "**英語**")
			fmt.Fprintln(w, "```xml")
			fmt.Fprintln(w, catalog.en)
			fmt.Fprintln(w, "```")
			fmt.Fprintln(w)
			fmt.Fprintln(w, "**日本語候補**")
			i := 1
			for ja := range catalog.jas {
				fmt.Fprintf(w, "%d. (%d)\n", i, catalog.jas[ja])
				fmt.Fprintln(w, "```xml")
				fmt.Fprintln(w, ja)
				fmt.Fprintln(w, "```")
				i++
			}
			fmt.Fprintln(w)
		}
	default:
		return fmt.Errorf("unsupported conflict format: %s", opt.Format)
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
func findNotSameCommon(seen map[string][]string) []conflictCatalog {
	uniques := []conflictCatalog{}
	for en, jas := range seen {
		if len(en) <= 12 { // 原文が12文字以下のものは除外する
			continue
		}
		if len(jas) <= 1 {
			continue
		}
		countJa := countStrings(jas)
		if len(countJa) > 1 {
			uniques = append(uniques, conflictCatalog{en: en, jas: countJa})
		}
	}
	sort.Slice(uniques, func(i, j int) bool {
		return uniques[i].en < uniques[j].en
	})
	return uniques
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

// countStrings は、文字列のスライスを受け取り、各文字列の出現回数をマップで返す
func countStrings(slice []string) map[string]int {
	counts := make(map[string]int)
	for _, v := range slice {
		j := stripNL(v)
		counts[j]++
	}
	return counts
}
