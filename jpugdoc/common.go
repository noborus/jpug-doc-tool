package jpugdoc

import (
	"log"
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
<title><link linkend="charset">Localization</link></title>,<title><link linkend="charset">多言語対応</link></title>
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
`

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
