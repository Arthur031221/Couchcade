<p align="center"><img src="assets/logo.svg" width="96" alt="Couchcade logo"></p>
<h1 align="center">Couchcade</h1>
<p align="center">免费的派对游戏，一台电视加上大家的手机。单个 Go 可执行文件，无需账号，无需安装 App。</p>

<p align="center">
<a href="https://github.com/Arthur031221/Couchcade/actions/workflows/ci.yml"><img src="https://github.com/Arthur031221/Couchcade/actions/workflows/ci.yml/badge.svg" alt="ci"></a>
<a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT"></a>
</p>

<p align="center"><a href="README.md">English</a> | <a href="README.zh-TW.md">繁體中文</a> | <b>简体中文</b> | <a href="README.ja.md">日本語</a> | <a href="README.ko.md">한국어</a> | <a href="README.es.md">Español</a> | <a href="README.fr.md">Français</a> | <a href="README.de.md">Deutsch</a> | <a href="README.pt-BR.md">Português</a></p>

<p align="center"><img src="assets/demo.gif" width="760" alt="电视上的 Squiggle 回合，手机正在投票"></p>

在电视或笔记本上打开 Couchcade，朋友扫描二维码，输入四个字母的房间代码，就能用自己的手机玩。内置六款原创游戏、可容纳大批观众的观众模式、人数不够时可以补位的机器人，以及九种语言，每种语言的题目都是专门写的。

**马上试玩：** [在线演示](https://arthur031221.github.io/Couchcade/)把同一个引擎编译成 WebAssembly 在浏览器里运行，座位由机器人补满。

## 实测数字

用 100 台模拟手机（8 个座位，92 位观众）测试，主持人按下开始后，每台手机都在 60 毫秒内收到第一题，12 次运行全部达成。

测量方法：`couchcade loadtest -phones 100` 在同一个进程内启动服务器，通过 loopback 打开 100 条真实的 WebSocket 连接，由主持人连接开始一轮 Brain Teams，并在每台手机上记录第一条带有题目的状态消息的时间。测试在 Ryzen 5 7500F 上重复 12 次，扩散时间在 37.5 到 48.1 毫秒之间，因为服务器每 50 毫秒合并一次变更再发出，这个数字大部分是合并的间隔。它测的是服务器，不是你的 Wi-Fi，Wi-Fi 会再加上自己的延迟。画面并未渲染，所以它量到的是送达手机 socket 的时间。

## 使用前后

|  | 一般的付费派对游戏盒 | Couchcade |
| --- | --- | --- |
| 费用 | 要买扩展包 | 免费，MIT 许可 |
| 运行位置 | 发行商的 App 与服务器 | 你自己网络上的一个可执行文件 |
| 题目 | 由发行商决定 | 可自行编辑和分享的 JSON 文件 |
| 语言 | 视作品而定 | 9 种，每种都有专属题目 |
| 大型聚会 | 视游戏而定 | 8 位玩家加上最多 500 位观众 |

Couchcade 不是任何商业游戏的复制品，六款游戏和所有题目都是原创。

## 游戏

| 游戏 | 玩法 |
| --- | --- |
| Squiggle | 偷偷画出题目，再投票选出最好的画 |
| Fibber | 写一个假答案，再找出真答案 |
| Brain Teams | 两队抢答选择题 |
| Zap | 等画面变绿，比谁点得快 |
| Echo | 心有灵犀，答案相同就得分 |
| Mole | 有一位玩家没有秘密词，把他找出来 |

规则与人数见 [docs/how-to-play.md](docs/how-to-play.md)。

## 快速开始

```sh
couchcade serve
```

在电视上打开它打印出的第一个地址，手机扫描二维码。如果你的电脑有多张网卡，二维码指向了错误的地址，请加上 `-public-url http://192.168.1.20:8080`。

<details>
<summary>安装</summary>

从[发布页面](https://github.com/Arthur031221/Couchcade/releases)下载可执行文件，或使用：

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

## 自己的题目

题目就是普通的 JSON。把自己的题库放进一个文件夹，用 `-packs ./my-packs` 启动：

```json
{"id": "office", "lang": "en", "items": {"squiggle": [{"text": "a tired printer"}]}}
```

`couchcade packs validate FILE` 会检查题库，`couchcade packs stats` 会统计数量。格式见 [docs/content-packs.md](docs/content-packs.md)，放在反向代理后面见 [docs/hosting.md](docs/hosting.md)，设计见 [docs/architecture.md](docs/architecture.md)。

## 测试内容

* 内容校验、引擎和每款游戏的 Go 单元测试，CI 中开启 race detector。
* 使用真实 WebSocket 的负载测试。
* 13 项端到端测试，用真实的 Chrome 操作真实的可执行文件：六款游戏各自搭配真人和机器人、断线重连、踢人、超过 8 位玩家、浅色和深色主题、每款游戏在 320 px 宽度下不会横向滚动，以及全部九种语言。
* 在 WebAssembly 演示页上完整玩一局游戏的浏览器测试。

## 限制

* 手机和电视都要连得到服务器，所以一般情况是同一个 Wi-Fi。要通过互联网玩，需要自己配置反向代理。
* 房间保存在内存里，重启服务器会中断进行中的游戏。
* 没有登录机制。没有可信任的代理时，不要把端口暴露到互联网。
* 绘图只支持手指或鼠标，有撤销和清除，但没有压感。

## 参与贡献

题目和翻译是最容易的入口，请看 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 许可证

[MIT](LICENSE)

Assisted by Claude/Codex.
