# JPUG-DOC tool

[PostgreSQL](http://www.postgresql.org/)の[文書](http://www.postgresql.org/docs/manuals/)を[日本語に翻訳](http://www.postgresql.jp/document/)している
[jpug-doc](https://github.com/pgsql-jp/jpug-doc)を助けるツールです。

前バージョンの翻訳を新しいバージョンに適用したり、翻訳のチェックが可能です。

※ 現在は以下のような形式になっている文書を対象としています。titleやindexterm等の翻訳には対応していません。

```
<para>
<!--
英語原文
-->
日本語訳
</para>
```

すべての文書形式に完璧に対応はせずに、そこそこで諦める方針です。

## インストール

```sh
go install github.com/noborus/jpug-doc-tool
```

## 使い方

[jpug-doc](https://github.com/pgsql-jp/jpug-doc/)を別途チェックアウトしてある状態で、`doc/src/sgml/` に移動して、jpug-doc-toolを実行します。

```sh
$ cd github.com/pgsql-jp/jpug-doc/doc/src/sgml
$ jpug-doc-tool サブコマンド
```

※ 一部文字色を変えて出力されます。デフォルトでは端末出力の場合のみ色が付き、リダイレクトした場合は付きません。
環境変数`FORCE_COLOR`により色付きの条件を変更できます。

色を変更しない
```sh
export FORCE_COLOR=0
```

色を必ず（リダイレクトしても）変更する
```sh
export FORCE_COLOR=1
```

## 英文、日本語文の抽出

まず最初に元ドキュメントから英文と翻訳文を抽出します。`doc_ja_12`のブランチ名か`pg124tail`のようなタグ名に切り替えます。

```sh
git checkout doc_ja_12
```

抽出するには `jpug-doc-tool extract`を実行します。

```sh
cd github.com/pgsql-jp/jpug-doc/doc/src/sgml
jpug-doc-tool extract
```

実行したディレクトリに `.jpug-doc-tool/acronyms.sgml.t` のようにsgmlに対応した対訳のセットファイルが作られます。

## 英文、日本語文の出力

`list`サブコマンドにより .tファイルの内容を見やすく出力します。

引数がない場合は全部を出力します。

```sh
jpug-doc-tool list
```
![list.png](https://raw.githubusercontent.com/noborus/jpug-doc-tool/main/doc/list.png)

sgmlファイルを指定すれば、そのsgmlファイルに対応している英文、日本語文を出力します。

```sh
jpug-doc-tool list acronyms.sgml
```

オプションにより英語のみ(`--en`)、日本語のみ(`--ja`)を指定できます。

```sh
jpug-doc-tool list --en acronyms.sgml
```

## 置き換え

英文、翻訳文の抽出した翻訳文を新しいバージョンに適用して、英語のみの文書から英語、翻訳文の形式に置き換えます。新しいブランチに切り替えてから `replace`を実行します。ファイル名を指定しなかった場合は全*sgmlファイルを置き換え対象にします。

```sh
git checkout doc_ja_13
jpug-doc-tool replace [ファイル名.sgml]
```

置き換えるのは、para内にコメント（英語原文）がない部分のみです。すでに翻訳済みの部分は何もしません。

## チェック

para内にコメントがない部分があったら表示します。 単純にコメントが含まれていないかをチェックするだけなので、
修正する必要があるとは限りません。目で見て必要な場合に修正します。

```sh
jpug-doc-tool check
```

![list.png](https://raw.githubusercontent.com/noborus/jpug-doc-tool/main/doc/check.png)

以下のようなparaは未翻訳であろうと推測して出力します。

NGなので出力
```
<para>
test
</para>
```

OKなのでスルー
```
<para>
<!--
test
-->
テスト
</para>
```

## 英単語チェック

抽出した、英文、日本語文から日本語文に含まれる英単語が英文にも含まれているかチェックします。
これは最新の翻訳状態でチェックする必要があるので、実行する前に`jpug-doc-tool extract`により翻訳のリストを更新してから実行します。
```

```sh
jpug-doc-tool check -w
```

以下のようになっている箇所では`ok`が英単語なので、コメントの方に`ok`が含まれているかをチェックします。

```
<para>
<!--
test is ok
-->
テストはok
</para>
```

これによりURLが古くなっている場合に検出できる可能性が高いです。


