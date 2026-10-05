package jpugdoc

import "testing"

func TestFoundReplaceRewrittenPre(t *testing.T) {
	c := Catalog{pre: "<para>\nHello\n", ja: "追加"}
	plain := []byte("<para>\nHello\nrest\n")
	if got := foundReplace(plain, c); got != len("<para>\nHello\n") {
		t.Errorf("plain: got %d", got)
	}
	rewritten := []byte("<para>\n<!--\nHello\n-->\nこんにちは\nrest\n")
	want := len("<para>\n<!--\nHello\n-->\nこんにちは\n")
	if got := foundReplace(rewritten, c); got != want {
		t.Errorf("rewritten: got %d, want %d", got, want)
	}
}

func TestFoundReplaceAdditionalLine(t *testing.T) {
	c := Catalog{pre: "-- 3. Initialize user session data\nCREATE TEMP TABLE s (x float);\n\n-- 4. Log the connection time\n", ja: "-- 4. 接続時刻を記録する。"}
	src := []byte("-- 3. Initialize user session data\n-- 3. ユーザのセッションデータを初期化する。\nCREATE TEMP TABLE s (x float);\n\n-- 4. Log the connection time\nrest\n")
	want := len(src) - len("rest\n")
	if got := foundReplace(src, c); got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}
