# JPUG-DOC tool

[PostgreSQL](http://www.postgresql.org/)の[文書](http://www.postgresql.org/docs/manuals/)を[日本語に翻訳](http://www.postgresql.jp/document/)している
[jpug-doc](https://github.com/pgsql-jp/jpug-doc)を助けるツールです。

前バージョンの翻訳を新しいバージョンに適用したり、翻訳のチェックが可能です。

## インストール

```console
go install github.com/noborus/jpug-doc-tool@latest
```

[https://github.com/noborus/jpug-doc-tool/releases/latest](https://github.com/noborus/jpug-doc-tool/releases/latest)に各OSバイナリもあります。

## 使い方

[jpug-doc](https://github.com/pgsql-jp/jpug-doc/)を別途チェックアウトしてある状態で、`doc/src/sgml/` に移動して、jpug-doc-toolを実行します。

```console
cd github.com/pgsql-jp/jpug-doc/doc/src/sgml
jpug-doc-tool サブコマンド
```

サブコマンドには以下があります。

- `extract` 英文、日本語文の抽出
- `replace` 英文、日本語文の置き換え
- `mtreplace` 機械翻訳の置き換え
- `list` 英文、日本語文の出力
- `check` 英文、日本語文のチェック
- `word` 英単語と日本語の単語のチェック
- `mt` 指定した英文を日本語に機械翻訳

PostgreSQLバージョンアップ時には、`extract`（抽出）→`replace`（置き換え）→`mtreplace`（機械翻訳）の順で実行します。

`list`,`check`,`mt`は最新バージョン翻訳に役立てるためのコマンドです。

### 抽出コマンド

`extract`サブコマンドにより、英文、日本語文の抽出を行います。基本的に完了しているバージョンのブランチで行います。
例えば完了しているバージョンがdoc_ja_15(15.4)で、新しいバージョンがdoc_ja_16(16.0)の例を示します。

```console
git checkout doc_ja_15
cd doc/src/sgml
jpug-doc-tool extract
```

内部的には `git diff REL_15_4 doc_ja_15`を実行して、変更箇所から英語と日本語を抽出します。
抽出は`.jpug-doc-tool`ディレクトリに`ファイル名.sgml.t`ファイルを作成します。

### 置き換えコマンド

`replace`サブコマンドにより、英文、翻訳文の抽出した翻訳文（doc/src/sgml/.jpug-doc-toolディレクトリにあるファイル）を新しいバージョンに適用して、英語のみの文書から英語、翻訳文の形式に置き換えます。
オプションなしの場合は、完全一致のみ置き換えます
（引数が無ければ、全sgmlファイル。引数としてファイル名を指定すると、そのファイルのみを置き換えます）。

```console
git checkout doc_ja_16
cd doc/src/sgml
jpug-doc-tool replace
```

オプションを追加することで、類似度や機械翻訳の置き換えを行います（ここでは機械翻訳はせずに機械翻訳のマークを付けるだけです）。

```console
jpug-doc-tool replace  --similar 50 --mt
```

`--similar 50`は類似度が50ポイント以上の場合に置き換えます。さらに90ポイント以下（デフォルト）の場合は`--mt`は機械翻訳のマークを付けます。
90ポイントを変更する場合は、`--mts`オプションで指定します。

実行結果は以下のようになります。`«`と`»`で囲まれた部分が機械翻訳のマークで、中身の英文が翻訳対象です。

```xml
     <para>
<!--
      <literal>PLAIN</literal> prevents either compression or
      out-of-line storage.  This is the only possible strategy for
      columns of non-<acronym>TOAST</acronym>-able data types.
-->
《マッチ度[56.994819]》<literal>PLAIN</literal>は圧縮や行外の格納を防止します。
さらにvarlena型での単一バイトヘッダの使用を無効にします。
これは<acronym>TOAST</acronym>化不可能のデータ型の列に対してのみ取り得る戦略です。
《機械翻訳》« <literal>PLAIN</literal> prevents either compression or out-of-line storage.  This is the only possible strategy for columns of non-<acronym>TOAST</acronym>-able data types. »
     </para>
```

### 機械翻訳置き換えコマンド

機械翻訳はAPIを使用して翻訳するため、後述する[APIの設定](#machine-translation-setting)が必要です。

mtreplaceサブコマンドにより、機械翻訳のみを置き換えます。`*sgml`、`ref/*sgml`ファイルをチェックして、`«`と`»`で囲まれた部分を機械翻訳で置き換えます。

```console
jpug-doc-tool mtreplace
```

結果は以下のようになります。

```xml
     <para>
<!--
      <literal>PLAIN</literal> prevents either compression or
      out-of-line storage.  This is the only possible strategy for
      columns of non-<acronym>TOAST</acronym>-able data types.
-->
《マッチ度[56.994819]》<literal>PLAIN</literal>は圧縮や行外の格納を防止します。
さらにvarlena型での単一バイトヘッダの使用を無効にします。
これは<acronym>TOAST</acronym>化不可能のデータ型の列に対してのみ取り得る戦略です。
《機械翻訳》<literal>PLAIN</literal>は圧縮も行外の格納も防止します。
これは<acronym>TOAST</acronym>不可能なデータ型の列に対する唯一の可能な戦略です。
     </para>
```

### リストコマンド

`list`サブコマンドにより、.tファイルの内容を見やすく出力します。

```console
jpug-doc-tool list
```

オプション無しは全てのファイルを対象に英語、日本語訳を出力します。これは色を付けて出力されます。

![list.png](https://raw.githubusercontent.com/noborus/jpug-doc-tool/main/doc/list.png)

`--en`オプションは英語のみ、`--ja`オプションは日本語のみを出力します。

`--pre`オプションは、（もしあれば）抽出時の前後の行を出力します。

`--tsv`オプションはタブ区切りで出力します。

sgmlファイルを指定すれば、そのsgmlファイルに対応している英文、日本語文を出力します。

```console
jpug-doc-tool list acronyms.sgml
```

### チェックコマンド

`check`サブコマンドにより、git diffを解析やファイル自体の原文と英語をチェックして問題がありそうな箇所を表示します。
git diffを解析して原文と英語をチェックして問題がありそうな箇所を表示します。
表示された内容が修正する必要があるとは限りません。目で見て必要な場合に修正します。

```console
jpug-doc-tool check [-w] [-n] [-t] [-p]
```

#### コメント形式のチェック(-pオプション)

`-p`オプションを指定すると、コメント形式のチェックを行います。

```console
jpug-doc-tool check -p
```

![para.png](https://raw.githubusercontent.com/noborus/jpug-doc-tool/main/doc/check.png)

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

#### 英単語チェック

**翻訳の日本語文**に含まれる英単語が**英文**にも含まれているかチェックします。

```console
jpug-doc-tool check -w
```

```console
ref/set_session_auth.sgml
<========================================
[SQL]が含まれていません
 The privileges necessary to execute this command are left implementation-defined by the standard.
-----------------------------------------
標準SQLでは、このコマンドを実行するために必要な権限は、実装に依存するとされています。
========================================>
```

このチェックは、以下のようになっている箇所では`ok`が英単語なので、コメントの方に`ok`が含まれているかをチェックします。

```xml
<para>
<!--
test is ok
-->
テストはok
</para>
```

これによりURLが古くなっている場合に検出できる可能性が高いです。

#### 数値チェック

**英文**にある数値が**翻訳の日本語文**にもあるかチェックします。

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

#### タグチェック

**英文**にある内部タグが**翻訳の日本語文**にもあるかチェックします。

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

#### 貢献者チェック

relase-*sgmlには、項目毎に貢献者が記載されています。このチェックは、貢献者が記載されているかをチェックします。

```console
jpug-doc-tool check -x
```

```text
<========================================
原文にある[(Masahiko Sawada, Noriyoshi Shinoda)]が含まれていません
 Add speculative lock information to the <link linkend="view-pg-locks"><structname>pg_locks</structname></link> system view (Masahiko Sawada, Noriyoshi Shinoda)
-----------------------------------------
投機的ロックの情報を<link linkend="view-pg-locks"><structname>pg_locks</structname></link>システムビューに追加しました。 (Sawada Masahiko, Shinoda Noriyoshi)
========================================>
```

#### 英語と日本語の単語のチェック

英単語と日本語の単語を指定して、その英単語が英語文にあり、日本語訳に日本語の単語がない文を出力します。

```console
jpug-doc-tool word -w "database" -j "データベース"
```

auth-delay.sgml

```xml
  <filename>auth_delay</filename> causes the server to pause briefly before
  reporting authentication failure, to make brute-force attacks on database
  passwords more difficult.  Note that it does nothing to prevent
  denial-of-service attacks, and may even exacerbate them, since processes
  that are waiting before reporting authentication failure will still consume
  connection slots.
<filename>auth_delay</filename>はパスワードの総当たり攻撃をより難しくするために認証エラーの報告を行う前にわずかにサーバを停止させます。
これはDoS攻撃を防ぐためのものでは無いことに注意してください。認証エラーを待たせ、コネクションスロットを消費させるため、DoS攻撃の影響を増長させるかもしれません。
```

### 機械翻訳コマンド

`mt`サブコマンドにより、引数で指定した英文を日本語に機械翻訳します。[APIの設定](#machine-translation-setting)が必要です。

```console
jpug-doc-tool mt "This is a pen."
これはペンです。
```

## 機械翻訳設定{#machine-translation-setting}

[みんなの自動翻訳＠TexTra®](https://mt-auto-minhon-mlt.ucri.jgn-x.jp/)のAPIを利用して、翻訳します。

みんなの自動翻訳＠TexTraのアカウントが必要です。まずアカウントを作成してください。

アカウントを作成したら、`$(HOME)/.jpug-doc-tool.yaml` にAPIの設定を書きます。

```yaml
APIKEY: 1234567890abcdef1234567dummy # (API key)
APISecret: e123123456456abcdefabcdefabdummy # (API secret)
APIName: "noborus" # (ログインID)
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
c-1640_en_ja: これはペンです。
generalNT_en_ja: これはペンです。
```

この翻訳設定は置き換えでも使用します。

## 注意事項

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
