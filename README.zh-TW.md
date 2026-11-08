<p align="center"><img src="assets/logo.svg" width="96" alt="Couchcade logo"></p>
<h1 align="center">Couchcade</h1>
<p align="center">免費的派對遊戲，一台電視加上大家的手機。單一 Go 執行檔，不用帳號，不用裝 App。</p>

<p align="center">
<a href="https://github.com/Arthur031221/Couchcade/actions/workflows/ci.yml"><img src="https://github.com/Arthur031221/Couchcade/actions/workflows/ci.yml/badge.svg" alt="ci"></a>
<a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT"></a>
</p>

<p align="center"><a href="README.md">English</a> | <b>繁體中文</b> | <a href="README.zh-CN.md">简体中文</a> | <a href="README.ja.md">日本語</a> | <a href="README.ko.md">한국어</a> | <a href="README.es.md">Español</a> | <a href="README.fr.md">Français</a> | <a href="README.de.md">Deutsch</a> | <a href="README.pt-BR.md">Português</a></p>

<p align="center"><img src="assets/demo.gif" width="760" alt="電視上的 Squiggle 回合，手機正在投票"></p>

在電視或筆電上開啟 Couchcade，朋友掃描 QR code，輸入四個字母的房間代碼，就能用自己的手機玩。內建六款原創遊戲、可容納大批觀眾的觀眾模式、人數不夠時可以補位的機器人，以及九種語言，每種語言的題目都是專門寫的。

**立即試玩：** [線上展示](https://arthur031221.github.io/Couchcade/)把同一個引擎編譯成 WebAssembly 在瀏覽器裡執行，座位由機器人補滿。

## 實測數字

用 100 支模擬手機（8 個座位，92 位觀眾）測試，主持人按下開始後，每支手機都在 60 毫秒內收到第一題，12 次執行全數達成。

測量方法：`couchcade loadtest -phones 100` 在同一個程序內啟動伺服器，透過 loopback 開 100 條真實的 WebSocket 連線，由主持人連線開始一輪 Brain Teams，並在每支手機上記錄第一則帶有題目的狀態訊息的時間。測試在 Ryzen 5 7500F 上重複 12 次，擴散時間落在 37.5 到 48.1 毫秒，因為伺服器每 50 毫秒合併一次變更再送出，這個數字大部分是合併的間隔。它測的是伺服器，不是你的 Wi-Fi，Wi-Fi 會再加上自己的延遲。畫面並未渲染，所以它量到的是送達手機 socket 的時間。

## 使用前後

|  | 一般的付費派對遊戲盒 | Couchcade |
| --- | --- | --- |
| 費用 | 要買擴充包 | 免費，MIT 授權 |
| 執行位置 | 發行商的 App 與伺服器 | 你自己網路上的一個執行檔 |
| 題目 | 由發行商決定 | 可自行編輯與分享的 JSON 檔 |
| 語言 | 依作品而定 | 9 種，每種都有專屬題目 |
| 大型聚會 | 依遊戲而定 | 8 位玩家加上最多 500 位觀眾 |

Couchcade 不是任何商業遊戲的複製品，六款遊戲和所有題目都是原創。

## 遊戲

| 遊戲 | 玩法 |
| --- | --- |
| Squiggle | 偷偷畫出題目，再投票選出最好的畫作 |
| Fibber | 寫一個假答案，再找出真答案 |
| Brain Teams | 兩隊搶答選擇題 |
| Zap | 等畫面變綠，比誰點得快 |
| Echo | 心有靈犀，答案相同就得分 |
| Mole | 有一位玩家沒有秘密詞，把他找出來 |

規則與人數請看 [docs/how-to-play.md](docs/how-to-play.md)。

## 快速開始

```sh
couchcade serve
```

在電視上開啟它印出的第一個網址，手機掃描 QR code。如果你的電腦有多張網卡，QR code 指到錯的位址，請加上 `-public-url http://192.168.1.20:8080`。

<details>
<summary>安裝</summary>

從[發行頁面](https://github.com/Arthur031221/Couchcade/releases)下載執行檔，或使用：

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

## 自己的題目

題目就是普通的 JSON。把自己的題庫放進一個資料夾，用 `-packs ./my-packs` 啟動：

```json
{"id": "office", "lang": "en", "items": {"squiggle": [{"text": "a tired printer"}]}}
```

`couchcade packs validate FILE` 會檢查題庫，`couchcade packs stats` 會統計數量。格式見 [docs/content-packs.md](docs/content-packs.md)，放在反向代理後面見 [docs/hosting.md](docs/hosting.md)，設計見 [docs/architecture.md](docs/architecture.md)。

## 測試內容

* 內容驗證、引擎與每款遊戲的 Go 單元測試，CI 中開啟 race detector。
* 使用真實 WebSocket 的負載測試。
* 13 項端對端測試，用真實的 Chrome 操作真實的執行檔：六款遊戲各自搭配真人與機器人、斷線重連、踢人、超過 8 位玩家、淺色與深色主題、每款遊戲在 320 px 寬度下不會橫向捲動，以及全部九種語言。
* 在 WebAssembly 展示頁上完整玩一場遊戲的瀏覽器測試。

## 限制

* 手機和電視都要連得到伺服器，所以一般情況是同一個 Wi-Fi。要透過網際網路玩，需要自己設定反向代理。
* 房間存在記憶體裡，重新啟動伺服器會中斷進行中的遊戲。
* 沒有登入機制。沒有可信任的代理時，不要把連接埠暴露到網際網路。
* 繪圖只支援手指或滑鼠，有復原與清除，但沒有壓力感應。

## 參與貢獻

題目和翻譯是最容易的入口，請看 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 授權

[MIT](LICENSE)

Assisted by Claude/Codex.
