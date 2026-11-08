<p align="center"><img src="assets/logo.svg" width="96" alt="Couchcade logo"></p>
<h1 align="center">Couchcade</h1>
<p align="center">TV 한 대와 모두의 휴대폰으로 즐기는 무료 파티 게임. Go 단일 실행 파일이고 계정도 앱도 필요 없습니다.</p>

<p align="center">
<a href="https://github.com/Arthur031221/Couchcade/actions/workflows/ci.yml"><img src="https://github.com/Arthur031221/Couchcade/actions/workflows/ci.yml/badge.svg" alt="ci"></a>
<a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT"></a>
</p>

<p align="center"><a href="README.md">English</a> | <a href="README.zh-TW.md">繁體中文</a> | <a href="README.zh-CN.md">简体中文</a> | <a href="README.ja.md">日本語</a> | <b>한국어</b> | <a href="README.es.md">Español</a> | <a href="README.fr.md">Français</a> | <a href="README.de.md">Deutsch</a> | <a href="README.pt-BR.md">Português</a></p>

<p align="center"><img src="assets/demo.gif" width="760" alt="TV의 Squiggle 라운드와 투표하는 휴대폰"></p>

TV나 노트북에서 Couchcade를 열고, 친구들이 QR 코드를 스캔해 네 글자 코드를 입력하면 각자의 휴대폰으로 바로 즐길 수 있습니다. 오리지널 게임 6개, 많은 인원을 위한 관객 모드, 사람이 모자랄 때 채워 주는 봇, 그리고 언어마다 따로 쓴 문제가 들어 있는 9개 언어를 지원합니다. 점수는 게임이 바뀌어도 파티 종합 순위에 누적되고, 플레이 중에는 휴대폰 화면이 꺼지지 않습니다.

**바로 해 보기:** [온라인 데모](https://arthur031221.github.io/Couchcade/)는 같은 엔진을 WebAssembly로 컴파일해 브라우저에서 실행하며, 빈 자리는 봇이 채웁니다.

## 측정한 숫자

가상 휴대폰 100대(착석 8, 관객 92)로 시험했을 때, 호스트가 시작을 누르고 모든 휴대폰에 첫 문제가 도착하기까지 60밀리초 이내였고, 12번 중 12번 모두 그랬습니다.

측정 방법: `couchcade loadtest -phones 100`이 같은 프로세스 안에서 서버를 띄우고, loopback으로 실제 WebSocket 연결 100개를 열고, 호스트 연결에서 Brain Teams 한 라운드를 시작한 뒤, 문제가 담긴 첫 state 메시지가 각 휴대폰에 도착한 시각을 기록합니다. Ryzen 5 7500F에서 12번 반복했고 전달 시간은 37.5에서 48.1밀리초였습니다. 서버가 변경 사항을 50밀리초마다 모아서 보내기 때문에 이 숫자의 대부분은 그 간격입니다. 측정 대상은 서버이지 Wi-Fi가 아니며, Wi-Fi는 자체 지연을 더합니다. 화면은 그리지 않으므로 휴대폰 소켓까지의 전달 시간을 잰 것입니다.

## 사용 전과 후

|  | 일반적인 유료 파티 게임 | Couchcade |
| --- | --- | --- |
| 비용 | 팩을 따로 구매 | 무료, MIT 라이선스 |
| 실행 위치 | 배급사의 앱과 서버 | 내 네트워크의 실행 파일 하나 |
| 문제 | 배급사가 고정 | 직접 고치고 공유하는 JSON 파일 |
| 언어 | 작품마다 다름 | 9개, 언어마다 전용 문제 |
| 많은 인원 | 게임마다 다름 | 플레이어 8명과 최대 500명의 관객 |

Couchcade는 어떤 상용 게임의 복제품도 아닙니다. 여섯 게임과 모든 문제가 오리지널입니다.

## 게임

| 게임 | 방법 |
| --- | --- |
| Squiggle | 제시어를 몰래 그리고 가장 잘 그린 그림에 투표 |
| Fibber | 가짜 답을 쓰고 진짜 답을 찾아내기 |
| Brain Teams | 두 팀이 객관식 문제를 빠르게 풀기 |
| Zap | 화면이 초록색이 될 때까지 기다렸다가 가장 빨리 누르기 |
| Echo | 생각을 맞추기, 답이 겹치면 득점 |
| Mole | 비밀 단어를 모르는 한 사람 찾기 |

규칙과 인원은 [docs/how-to-play.md](docs/how-to-play.md)에 있습니다.

## 빠른 시작

```sh
couchcade serve
```

출력된 첫 번째 주소를 TV에서 열고 휴대폰으로 QR 코드를 스캔합니다. 네트워크 카드가 여러 개라 QR 코드가 엉뚱한 주소를 가리키면 `-public-url http://192.168.1.20:8080`을 붙이세요.

<details>
<summary>설치</summary>

[릴리스 페이지](https://github.com/Arthur031221/Couchcade/releases)에서 실행 파일을 받거나 다음을 사용하세요.

```sh
# Go 1.24+
go install github.com/Arthur031221/Couchcade/cmd/couchcade@latest

# Docker
docker build -t couchcade https://github.com/Arthur031221/Couchcade.git
docker run --rm -p 8080:8080 couchcade serve -addr :8080 -public-url http://YOUR-LAN-IP:8080

# Source
git clone https://github.com/Arthur031221/Couchcade && cd Couchcade
go build -o couchcade ./cmd/couchcade
```

</details>

## 나만의 문제

문제는 평범한 JSON입니다. 직접 만든 팩을 폴더에 넣고 `-packs ./my-packs`로 시작하세요.

```json
{"id": "office", "lang": "en", "items": {"squiggle": [{"text": "a tired printer"}]}}
```

`couchcade packs validate FILE`은 팩을 검사하고 `couchcade packs stats`는 개수를 셉니다. 형식은 [docs/content-packs.md](docs/content-packs.md), 리버스 프록시 뒤에서 쓰는 법은 [docs/hosting.md](docs/hosting.md), 설계는 [docs/architecture.md](docs/architecture.md)에 있습니다.

## 테스트 내용

* 콘텐츠 검증, 엔진, 모든 게임에 대한 Go 단위 테스트. CI에서는 race detector를 켭니다.
* 실제 WebSocket을 쓰는 부하 테스트.
* 실제 Chrome으로 실제 실행 파일을 조작하는 16개의 엔드투엔드 테스트: 여섯 게임 각각을 사람과 봇으로, 재접속, 강퇴, 8명 초과 참가, 라이트와 다크 테마, 모든 게임에서 320px 폭에서도 가로 스크롤이 없는지, 아홉 개 언어 전부(버튼 라벨 번역 포함).
* WebAssembly 데모에서 한 판을 끝까지 플레이하는 브라우저 테스트.

## 한계

* 휴대폰과 TV 모두 서버에 닿아야 하므로 보통은 같은 Wi-Fi입니다. 인터넷으로 즐기려면 직접 리버스 프록시를 설정해야 합니다.
* 방은 메모리에 있습니다. 서버를 다시 시작하면 진행 중인 게임이 끝납니다.
* 로그인이 없습니다. 믿을 수 있는 프록시 없이 포트를 인터넷에 열지 마세요.
* 그리기는 손가락이나 마우스만 지원하며 실행 취소와 지우기는 있지만 필압은 지원하지 않습니다.

## 기여하기

문제와 번역이 가장 쉬운 시작점입니다. [CONTRIBUTING.md](CONTRIBUTING.md)를 보세요.

## 라이선스

[MIT](LICENSE)

Assisted by Claude/Codex.
