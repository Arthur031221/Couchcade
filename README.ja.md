<p align="center"><img src="assets/logo.svg" width="96" alt="Couchcade logo"></p>
<h1 align="center">Couchcade</h1>
<p align="center">テレビ1台とみんなのスマホで遊べる無料のパーティーゲーム。Go製の単一バイナリで、アカウントもアプリも不要です。</p>

<p align="center">
<a href="https://github.com/Arthur031221/Couchcade/actions/workflows/ci.yml"><img src="https://github.com/Arthur031221/Couchcade/actions/workflows/ci.yml/badge.svg" alt="ci"></a>
<a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT"></a>
</p>

<p align="center"><a href="README.md">English</a> | <a href="README.zh-TW.md">繁體中文</a> | <a href="README.zh-CN.md">简体中文</a> | <b>日本語</b> | <a href="README.ko.md">한국어</a> | <a href="README.es.md">Español</a> | <a href="README.fr.md">Français</a> | <a href="README.de.md">Deutsch</a> | <a href="README.pt-BR.md">Português</a></p>

<p align="center"><img src="assets/demo.gif" width="760" alt="テレビ上のSquiggleのラウンドと、投票するスマホ"></p>

テレビやノートPCでCouchcadeを開き、友だちがQRコードを読み取って4文字のコードを入力すれば、自分のスマホで遊べます。オリジナルのゲームが6つ、大人数向けの観戦モード、人数が足りないときのボット、そして9言語に対応していて、お題は言語ごとに書き下ろしています。

**すぐ試す：** [オンラインデモ](https://arthur031221.github.io/Couchcade/)では、同じエンジンをWebAssemblyにしてブラウザで動かし、空席はボットが埋めます。

## 計測した数字

100台の疑似スマホ（着席8、観戦92）で、ホストが開始を押してから、すべてのスマホに最初の問題が届くまで60ミリ秒以内で、12回中12回とも満たしました。

計測方法：`couchcade loadtest -phones 100` が同じプロセス内でサーバーを起動し、loopback上に本物のWebSocket接続を100本開き、ホスト接続からBrain Teamsの1ラウンドを始め、問題を含む最初のstateメッセージが各スマホに届いた時刻を記録します。Ryzen 5 7500Fで12回繰り返し、配信時間は37.5から48.1ミリ秒でした。サーバーは変更を50ミリ秒ごとにまとめて送るため、この数字の大半はその間隔です。測っているのはサーバーであり、Wi-Fiではありません。Wi-Fiではそこに遅延が加わります。画面は描画しないため、スマホのソケットへの配信までを測っています。

## 導入前と導入後

|  | 一般的な有料パーティーゲーム | Couchcade |
| --- | --- | --- |
| 費用 | パックごとに購入 | 無料、MITライセンス |
| 動く場所 | 配信元のアプリとサーバー | 自分のネットワーク上の1つのバイナリ |
| お題 | 配信元が固定 | 自分で編集して共有できるJSONファイル |
| 言語 | 作品による | 9言語、言語ごとにお題あり |
| 大人数 | ゲームによる | プレイヤー8人と最大500人の観戦 |

Couchcadeは市販ゲームのクローンではありません。6つのゲームとすべてのお題はオリジナルです。

## ゲーム

| ゲーム | 遊び方 |
| --- | --- |
| Squiggle | お題をこっそり描いて、いちばん良い絵に投票する |
| Fibber | 偽の答えを書いて、本物の答えを見破る |
| Brain Teams | 2チームで選択式の問題を早押しする |
| Zap | 画面が緑になるのを待って、誰よりも早くタップする |
| Echo | 考えを揃える。答えが一致すると得点 |
| Mole | 秘密の言葉を知らない1人を探し出す |

ルールと人数は [docs/how-to-play.md](docs/how-to-play.md) にあります。

## クイックスタート

```sh
couchcade serve
```

表示された最初のアドレスをテレビで開き、スマホでQRコードを読み取ります。ネットワークカードが複数あってQRコードが違うアドレスを指す場合は、`-public-url http://192.168.1.20:8080` を付けてください。

<details>
<summary>インストール</summary>

[リリースページ](https://github.com/Arthur031221/Couchcade/releases)からバイナリをダウンロードするか、次の方法を使います。

```sh
# Go 1.24+
go install github.com/Arthur031221/Couchcade/cmd/couchcade@latest

# Docker
docker run --rm -p 8080:8080 ghcr.io/arthur031221/couchcade serve -addr :8080 -public-url http://YOUR-LAN-IP:8080

# Homebrew
brew install Arthur031221/tap/couchcade

# Source
git clone https://github.com/Arthur031221/Couchcade && cd Couchcade
go build -o couchcade ./cmd/couchcade
```

</details>

## 自分のお題を入れる

お題は普通のJSONです。自分のパックをフォルダに入れて、`-packs ./my-packs` で起動します。

```json
{"id": "office", "lang": "en", "items": {"squiggle": [{"text": "a tired printer"}]}}
```

`couchcade packs validate FILE` でパックを検査し、`couchcade packs stats` で件数を数えます。形式は [docs/content-packs.md](docs/content-packs.md)、リバースプロキシの内側での運用は [docs/hosting.md](docs/hosting.md)、設計は [docs/architecture.md](docs/architecture.md) にあります。

## テスト内容

* コンテンツ検証、エンジン、全ゲームのGoユニットテスト。CIではrace detectorを有効にしています。
* 本物のWebSocketを使った負荷テスト。
* 本物のChromeで本物のバイナリを操作する13件のエンドツーエンドテスト：6つのゲームそれぞれを人間とボットで、再接続、キック、9人目以降の参加、ライトとダークのテーマ、全ゲームで幅320pxでも横スクロールしないこと、9言語すべて。
* WebAssemblyデモで1ゲームを最後まで遊ぶブラウザテスト。

## 制限

* スマホとテレビの両方がサーバーに届く必要があるため、通常は同じWi-Fiです。インターネット越しに遊ぶには、自分でリバースプロキシを用意します。
* ルームはメモリ上にあります。サーバーを再起動すると進行中のゲームは終わります。
* ログインはありません。信頼できるプロキシなしでポートをインターネットに公開しないでください。
* お絵描きは指かマウスのみで、元に戻すと消去はありますが、筆圧には対応していません。

## コントリビュート

お題と翻訳がいちばん参加しやすいです。[CONTRIBUTING.md](CONTRIBUTING.md) を見てください。

## ライセンス

[MIT](LICENSE)

Assisted by Claude/Codex.
