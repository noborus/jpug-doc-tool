package jpugdoc

import "testing"

func TestExtractionCDATAPre(t *testing.T) {
	diff := "@@ -1,6 +1,9 @@\n <para>\n+<!--\n text\n+-->\n+テキスト\n </para>\n <programlisting><![CDATA[\n int x;\n \n+]]><!--\n     if (!A(f))  /* internal error */\n+--><![CDATA[\n+    if (!A(f))  /* 内部エラー */\n         elog();\n ]]></programlisting>\n"
	cs := Extraction([]byte(diff))
	if len(cs) == 0 {
		t.Fatal("no catalog")
	}
	if got := cs[len(cs)-1].pre; got == "<para>" {
		t.Errorf("stale pre: %q", got)
	}
}
