package jpugdoc

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteConflictGHMD(t *testing.T) {
	catalogs := []conflictCatalog{
		{
			en: "<para>Accessing a Database</para>",
			jas: map[string]int{
				"<para>データベースへのアクセス</para>": 1,
				"<para>DBアクセス</para>":       1,
			},
		},
	}

	var b bytes.Buffer
	err := writeConflict(&b, catalogs, ConflictOption{Format: ConflictFormatGHMD})
	if err != nil {
		t.Fatalf("writeConflict() error = %v", err)
	}

	got := b.String()
	want := strings.Join([]string{
		"### Conflict 1",
		"",
		"**英語**",
		"```xml",
		"<para>Accessing a Database</para>",
		"```",
		"",
		"**日本語候補**",
		"1. (1)",
		"```xml",
		"<para>DBアクセス</para>",
		"```",
		"2. (1)",
		"```xml",
		"<para>データベースへのアクセス</para>",
		"```",
		"",
	}, "\n")

	if strings.TrimSpace(got) != strings.TrimSpace(want) {
		t.Errorf("writeConflict() =\n%s\nwant\n%s", got, want)
	}
}

func TestWriteConflictInvalidFormat(t *testing.T) {
	var b bytes.Buffer
	err := writeConflict(&b, nil, ConflictOption{Format: "unknown"})
	if err == nil {
		t.Fatal("writeConflict() expected error")
	}
}

func TestFindNotSameCommonSortAndFilter(t *testing.T) {
	seen := map[string][]string{
		"short":                 {"訳1", "訳2"},
		"This is enough length": {"同じ", "同じ"},
		"Z conflict sample":     {"訳Z1", "訳Z2"},
		"A conflict sample":     {"訳A1", "訳A2"},
	}

	got := findNotSameCommon(seen)
	if len(got) != 2 {
		t.Fatalf("findNotSameCommon() len = %d, want 2", len(got))
	}
	if got[0].en != "A conflict sample" {
		t.Errorf("findNotSameCommon()[0].en = %s, want A conflict sample", got[0].en)
	}
	if got[1].en != "Z conflict sample" {
		t.Errorf("findNotSameCommon()[1].en = %s, want Z conflict sample", got[1].en)
	}
}
