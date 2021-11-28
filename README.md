# JPUG-DOC tool

[PostgreSQL](http://www.postgresql.org/)の[文書](http://www.postgresql.org/docs/manuals/)を[日本語に翻訳](http://www.postgresql.jp/document/)している
[jpug-doc](https://github.com/pgsql-jp/jpug-doc)を助けるツールです。

前バージョンの翻訳を新しいバージョンに適用したり、翻訳のチェックが可能です。

※ 現在は以下のような形式になっている文書を対象としています。titleやindexterm等の翻訳には対応していません。

```xml
<para>
<!--
英語原文
-->
日本語訳
</para>
```

すべての文書形式に完璧に対応はせずに、そこそこで諦める方針です。

## インストール

```console
go install github.com/noborus/jpug-doc-tool@latest
```

## 使い方

[jpug-doc](https://github.com/pgsql-jp/jpug-doc/)を別途チェックアウトしてある状態で、`doc/src/sgml/` に移動して、jpug-doc-toolを実行します。

```console
cd github.com/pgsql-jp/jpug-doc/doc/src/sgml
jpug-doc-tool サブコマンド
```

※ 一部文字色を変えて出力されます。デフォルトでは端末出力の場合のみ色が付き、リダイレクトした場合は付きません。
環境変数`FORCE_COLOR`により色付きの条件を変更できます。

色を変更しない

```console
export FORCE_COLOR=0
```

色を必ず（リダイレクトしても）変更する

```console
export FORCE_COLOR=1
```

## 機械翻訳

[みんなの自動翻訳＠TexTra®](https://mt-auto-minhon-mlt.ucri.jgn-x.jp/)のAPをを利用して、翻訳します。

みんなの自動翻訳＠TexTraのアカウントが必要です。まずアカウントを作成してください。

アカウントを作成したら、`$(HOME)/.jpug-doc-tool.yaml` にAPIの設定を書きます。

```yaml
ClientID: 1234567890abcdef123456789abc # (API key)
ClientSecret: e123123456456abcdefabcdefabcdef1 # (API secret)
Name: "noborus" # (ログインID)
APIAutoTranslate: "mt" #
APIAutoTranslateType: "c-1640_en_ja" # （翻訳エンジン） 汎用NT は "generalNT_en_ja" になります。
```

ログイン後 [☁ Web API] → [自動翻訳リクエスト　一覧] → [ℹ API] 等から `API key`と`API secret`、`name`の`ログインID:` を確認してコピー、ペーストしてください。

![API](https://raw.githubusercontent.com/noborus/jpug-doc-tool/main/doc/APIkey.png)

翻訳エンジンは "c-1640_en_ja" がPostgreSQLマニュアル翻訳用にカスタマイズしたエンジンです。
もしエラーになる場合は "generalNT_en_ja" にして試してみてください。

 "c-1640_en_ja" が使用できない場合は、ユーザーのグループ化が必要かもしれません。みんなの自動翻訳＠TexTraのアカウントを伝えてください。

設定できたら機械翻訳を試します。 `mt`サブコマンドで英語→日本語の翻訳ができます。

```console
$ jpug-doc-tool mt "This is a pen."
Using config file: /home/noborus/.jpug-doc-tool.yaml
これはペンです。
```

この翻訳設定は置き換えでも使用します。

## 英文、日本語文の抽出

まず最初に元ドキュメントから英文と翻訳文を抽出します。`doc_ja_13`のブランチ名か`pg131tail`のようなタグ名に切り替えます。

```console
git checkout doc_ja_13
```

抽出するには `jpug-doc-tool extract`を実行します。

```console
cd github.com/pgsql-jp/jpug-doc/doc/src/sgml
jpug-doc-tool extract
```

実行したディレクトリに `.jpug-doc-tool/acronyms.sgml.t` のようにsgmlに対応した対訳のセットファイルが作られます。

## 英文、日本語文の出力

`list`サブコマンドにより .tファイルの内容を見やすく出力します。

引数がない場合は全部を出力します。

```console
jpug-doc-tool list
```

![list.png](https://raw.githubusercontent.com/noborus/jpug-doc-tool/main/doc/list.png)

sgmlファイルを指定すれば、そのsgmlファイルに対応している英文、日本語文を出力します。

```console
jpug-doc-tool list acronyms.sgml
```

オプションにより英語のみ(`--en`)、日本語のみ(`--ja`)を指定できます。

```console
jpug-doc-tool list --en acronyms.sgml
```

## 置き換え

英文、翻訳文の抽出した翻訳文を新しいバージョンに適用して、英語のみの文書から英語、翻訳文の形式に置き換えます。新しいブランチに切り替えてから `replace`を実行します。ファイル名を指定しなかった場合は全*sgmlファイルを置き換え対象にします。

```console
git checkout doc_ja_13
jpug-doc-tool replace [ファイル名.sgml]
```

置き換えるのは、para内にコメント（英語原文）がない部分のみです。すでに翻訳済みの部分は何もしません。

オプションを付けずに`replace`を実行した場合は、スペース、改行等を除いて完全に一致した場合のみ置き換えます。

### 機械翻訳で翻訳する

`replace`に `--mt`オプションをつけると、機械翻訳によって置き換えます。時間もかかるのでファイルを指定しての実行をオススメします。

```console
jpug-doc-tool --mt [ファイル名.sgml]
```

実行すると時間がかかるためAPI問い合わせした場合は以下のように`API...`、`Done`と表示されます。

```console
API...Done
API...Done
...
```

置き換えした箇所は以下のように`《機械翻訳》｀のコメントの後に翻訳文が追加されています。

```diff
  <para>
--- a/doc/src/sgml/hash.sgml
+++ b/doc/src/sgml/hash.sgml
+<!--
   Hash indexes support only single-column indexes and do not allow
   uniqueness checking.
+-->
+<!-- 《機械翻訳》 -->
+ハッシュのインデックスはサポートの単一カラムのインデックスのみで、一意性のチェックはできません。
  </para>
```

### 類似文を対象にする

`-s` 又は `--similar`にスコア（100点満点）のオプションをつけると「レーベンシュタイン距離」により文字列の類似度を測って指定したスコア以上であれば置き換えます。時間もかかるのでファイルを指定しての実行をオススメします。

```console
jpug-doc-tool replace -s 90 [ファイル名.sgml]
```

90点以上であれば、数文字が違うだけの少し変更した文章についても置き換えます。完全一致ではないときには以下のようにマッチ度と元文章を付けて置き換えます。目で見て、不要な部分を消して修正する必要があります。

```diff
-- a/doc/src/sgml/func.sgml
+++ b/doc/src/sgml/func.sgml
@@ -12337,8 +12337,14 @@ SELECT EXTRACT(CENTURY FROM TIMESTAMP '2001-02-16 20:38:40');
       <term><literal>day</literal></term>
       <listitem>
        <para>
+<!--
         For <type>timestamp</type> values, the day (of the month) field
         (1&ndash;31) ; for <type>interval</type> values, the number of days
+-->
+<!-- マッチ度[94.656489]
+For <type>timestamp</type> values, the day (of the month) field (1 - 31) ; for <type>interval</type> values, the number of days
+-->
+<type>timestamp</type>値については、(月内の)日付フィールド(1〜31)。<type>interval</type>値については日数。
        </para>
```

## チェック

para内の原文と英語をチェックして問題がありそうな箇所を表示します。
表示された内容が修正する必要があるとは限りません。目で見て必要な場合に修正します。

### コメント形式のチェック

オプションがない場合はコメント形式を単純にコメントが含まれていないかをチェックするだけなので、

```console
jpug-doc-tool check
```

![list.png](https://raw.githubusercontent.com/noborus/jpug-doc-tool/main/doc/check.png)

以下のようなparaは未翻訳であろうと推測して出力します。

NGなので出力

```xml
<para>
test
</para>
```

OKなのでスルー

```xml
<para>
<!--
test
-->
テスト
</para>
```

### 英単語チェック

抽出した、英文、日本語文から日本語文に含まれる英単語が英文にも含まれているかチェックします。

```console
jpug-doc-tool check -w
```

以下のようになっている箇所では`ok`が英単語なので、コメントの方に`ok`が含まれているかをチェックします。

```xml
<para>
<!--
test is ok
-->
テストはok
</para>
```

これによりURLが古くなっている場合に検出できる可能性が高いです。

### 数値チェック

英文にある数値が日本語にもあるかチェックします。

```console
jpug-doc-tool check -n
```

数値の表記方法が変わっている場合に正しいかどうか判断できないので出力されます。

```text
config.sgml
<========================================
原文にある[200]が含まれていません
Vacuum also allows removal of old files from the <filename>pg_xact</filename> subdirectory, which is why the default is a relatively low 200 million transactions. This parameter can only be set at server start, but the setting can be reduced for individual tables by changing table storage parameters. For more information see <xref linkend="vacuum-for-wraparound"/>.
-----------------------------------------
vacuumは同時に<filename>pg_xact</filename>サブディレクトリから古いファイルの削除を許可します。
       これが、比較的低い2億トランザクションがデフォルトである理由です。
       このパラメータはサーバ起動時にのみ設定可能です。
しかし、この設定はテーブルストレージパラメータの変更により、それぞれのテーブルで減らすことができます。
詳細は<xref linkend="vacuum-for-wraparound"/>を参照してください。
========================================>
```

### タグチェック

英文にある内部タグが日本語にもあるかチェックします。

```console
jpug-doc-tool check -t
```

```text
<========================================
原文にある[<emphasis>]が含まれていません
It is recommended that you use the <application>pg_dump</application> and <application>pg_dumpall</application> programs from the <emphasis>newer</emphasis> version of <productname>PostgreSQL</productname>, to take advantage of enhancements that might have been made in these programs. Current releases of the dump programs can read data from any server version back to 7.0.
-----------------------------------------
新しいバージョンの<productname>PostgreSQL</productname>の<application>pg_dump</application>と<application>pg_dumpall</application>を使用することを勧めます。
これらのプログラムで拡張された機能を利用する可能性があるためです。
現在のリリースのダンププログラムは7.0以降のバージョンのサーバからデータを読み取ることができます。
========================================>
```
