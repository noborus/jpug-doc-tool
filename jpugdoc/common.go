package jpugdoc

import (
	"log"
	"regexp"
	"strings"
)

var titleData = `
<title>Arguments</title>,<title>引数</title>
<title>Arrays</title>,<title>配列</title>
<title>Author</title>,<title>作者</title>
<title>Authors</title>,<title>作者</title>
<title>Built-in Operator Classes</title>,<title>組み込み演算子クラス</title>
<title>Caveats</title>,<title>警告</title>
<title>Client Interfaces</title>,<title>クライアントインタフェース</title>
<title>Compatibility</title>,<title>互換性</title>
<title>Composite Types</title>,<title>複合型</title>
<title>Concepts</title>,<title>概念</title>
<title>Configuration Parameters</title>,<title>設定パラメータ</title>
<title>Configuration</title>,<title>設定</title>
<title>Data Types</title>,<title>データ型</title>
<title>Description</title>,<title>説明</title>
<title>Developer Options</title>,<title>開発者向けオプション</title>
<title>Diagnostics</title>,<title>診断</title>
<title>Environment Variables</title>,<title>環境変数</title>
<title>Environment</title>,<title>環境</title>
<title>Error Handling</title>,<title>エラー処理</title>
<title>Example</title>,<title>例</title>
<title>Examples</title>,<title>例</title>
<title>Exit Status</title>,<title>終了ステータス</title>
<title>Extensibility</title>,<title>拡張性</title>
<title>Functional Dependencies</title>,<title>関数従属性</title>
<title>Functions and Operators</title>,<title>関数と演算子</title>
<title>Functions</title>,<title>関数</title>
<title>Implementation</title>,<title>実装</title>
<title>Indexes</title>,<title>インデックス</title>
<title>Inheritance</title>,<title>継承</title>
<title>Introduction</title>,<title>はじめに</title>
<title>Limitations</title>,<title>制限事項</title>
<title>Miscellaneous</title>,<title>その他</title>
<title>Monitoring</title>,<title>監視</title>
<title>Notes</title>,<title>注釈</title>
<title>Options</title>,<title>オプション</title>
<title>Outputs</title>,<title>出力</title>
<title>Overview</title>,<title>概要</title>
<title>Parameters</title>,<title>パラメータ</title>
<title>Pseudo-Types</title>,<title>疑似データ型</title>
<title>Rationale</title>,<title>原理</title>
<title>Regression Tests</title>,<title>リグレッションテスト</title>
<title>Requirements</title>,<title>必要条件</title>
<title>Return Value</title>,<title>戻り値</title>
<title>Sample Output</title>,<title>サンプル出力</title>
<title>See Also</title>,<title>関連項目</title>
<title>Transaction Management</title>,<title>トランザクション制御</title>
<title>Transforms</title>,<title>変換</title>
<title>Trigger Functions</title>,<title>トリガ関数</title>
<title>Usage</title>,<title>使用方法</title>
<title>Release Notes</title>,<title>リリースノート</title>
<title>Release date:</title>,<title>リリース日:</title>
<title>Changes</title>,<title>変更点</title>
<title>Server</title>,<title>サーバ</title>
<title>Optimizer</title>,<title>オプティマイザ</title>
<title>General Performance</title>,<title>性能一般</title>
<title>Server Configuration</title>,<title>サーバ設定</title>
<title><link linkend="charset">Localization</link></title>,<title><link linkend="charset">ローカライゼーション</link></title>
<title><link linkend="logical-replication">Logical Replication</link></title>,<title><link linkend="logical-replication">論理レプリケーション</link></title>
<title>Utility Commands</title>,<title>ユーティリティコマンド</title>
<title>General Queries</title>,<title>問い合わせ一般</title>
<title>Client Applications</title>,<title>クライアントアプリケーション</title>
<title>Server Applications</title>,<title>サーバアプリケーション</title>
<title>Source Code</title>,<title>ソースコード</title>
<title>Additional Modules</title>,<title>追加モジュール</title>
<title>Acknowledgments</title>,<title>謝辞</title>
<title><acronym>Authentication</acronym></title>,<title><acronym>認証</acronym></title>
<title>Documentation</title>,<title>ドキュメンテーション</title>
<title>Privileges</title>,<title>権限</title>
`

// commonRefMark は commonData の訳文中で正規表現のN番目のグループを参照する記号(§1 など)。
const commonRefMark = "§"

// commonData はリリースノートの定型句。各項目は "----" 行、原文(正規表現)と訳文は "====" 行で区切る。
var commonData = `<title>Release (\d+(?:\.\d+)?)</title>
====
 <title>リリース§1</title>
----
<title>Migration to Version (\d+(?:\.\d+)?)</title>
====
 <title>バージョン§1への移行</title>
----
<productname>PostgreSQL</productname> (\d+(?:\.\d+)?) contains many new features
\s+and enhancements, including:
====
<productname>PostgreSQL</productname> §1には、以下をはじめとする多数の新機能と拡張が含まれています。
----
The above items and other new features of
\s+<productname>PostgreSQL</productname> (\d+(?:\.\d+)?) are explained in more detail
\s+in the sections below\.
====
<productname>PostgreSQL</productname> §1の上記の項目とその他の新機能は次節でより詳しく説明されます。
----
Version (\d+(?:\.\d+)?) contains a number of changes that may affect compatibility
\s+with previous releases\.\s+Observe the following incompatibilities:
====
バージョン§1には、以前のバージョンとの互換性に影響するかもしれない多数の変更点が含まれています。
以下の非互換性に注意してください。
----
Below you will find a detailed account of the changes between
\s+<productname>PostgreSQL</productname> (\d+(?:\.\d+)?) and the previous major
\s+release\.
====
<productname>PostgreSQL</productname> §1と前メジャーリリースとの詳細な変更点を記載しました。`

// regexpCatalog は原文を正規表現として扱う共通カタログを作る。
func regexpCatalog(en, ja string) Catalog {
	return Catalog{
		en:        en,
		ja:        ja,
		isRegexp:  true,
		commonReg: regexp.MustCompile(`(.*\n)?[^\n]*` + en + `\n`),
	}
}

// commonCatalogs は commonData を Catalog の配列に変換する。
func commonCatalogs() []Catalog {
	var catalogs []Catalog
	for entry := range strings.SplitSeq(commonData, "\n----\n") {
		en, ja, ok := strings.Cut(entry, "\n====\n")
		if !ok {
			log.Printf("Unexpected format in commonData: %s", entry)
			continue
		}
		catalogs = append(catalogs, regexpCatalog(en, ja))
	}
	return catalogs
}

func titleMap() map[string]string {
	lines := strings.Split(titleData, "\n")
	m := make(map[string]string)
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			log.Printf("Unexpected format in titleData: %s", line)
			continue
		}
		m[parts[0]] = parts[1]
	}
	return m
}
