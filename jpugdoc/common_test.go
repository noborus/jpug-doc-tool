package jpugdoc

import (
	"bytes"
	"testing"
)

func TestMatchCommonLineEndOnly(t *testing.T) {
	cats := regCompile(Catalogs{{en: "not implemented", ja: "実装されていません。"}})
	src := []byte("   aggregations; but that is not implemented\n   yet.)\n")
	if got := matchCommon(src, cats[0]); !bytes.Equal(got, src) {
		t.Errorf("unexpected replace: %q", got)
	}
	src = []byte("   not implemented\n")
	if got := matchCommon(src, cats[0]); bytes.Equal(got, src) {
		t.Errorf("not replaced: %q", got)
	}
}
