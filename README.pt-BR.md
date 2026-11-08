<p align="center"><img src="assets/logo.svg" width="96" alt="Couchcade logo"></p>
<h1 align="center">Couchcade</h1>
<p align="center">Jogos de festa grátis para uma TV e os celulares de todo mundo. Um único binário em Go, sem conta e sem app.</p>

<p align="center">
<a href="https://github.com/Arthur031221/Couchcade/actions/workflows/ci.yml"><img src="https://github.com/Arthur031221/Couchcade/actions/workflows/ci.yml/badge.svg" alt="ci"></a>
<a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT"></a>
</p>

<p align="center"><a href="README.md">English</a> | <a href="README.zh-TW.md">繁體中文</a> | <a href="README.zh-CN.md">简体中文</a> | <a href="README.ja.md">日本語</a> | <a href="README.ko.md">한국어</a> | <a href="README.es.md">Español</a> | <a href="README.fr.md">Français</a> | <a href="README.de.md">Deutsch</a> | <b>Português</b></p>

<p align="center"><img src="assets/demo.gif" width="760" alt="Uma rodada de Squiggle na TV com celulares votando"></p>

Abra o Couchcade na TV ou no notebook. Os amigos escaneiam o QR code, digitam o código de quatro letras e jogam pelo próprio celular. Ele traz seis jogos originais, um modo plateia para grupos grandes, bots para quando falta gente e nove idiomas com textos escritos para cada um.

**Experimente agora:** a [demo online](https://arthur031221.github.io/Couchcade/) roda o mesmo motor compilado para WebAssembly no seu navegador, com bots ocupando os lugares.

## O número

Com 100 celulares simulados (8 sentados e 92 na plateia), cada celular recebeu a primeira pergunta em menos de 60 ms depois que o anfitrião apertou Iniciar, em 12 de 12 execuções.

Como foi medido: `couchcade loadtest -phones 100` sobe um servidor no mesmo processo, abre 100 conexões WebSocket reais por loopback, inicia uma rodada de Brain Teams pela conexão do anfitrião e registra quando a primeira mensagem de estado com a pergunta chega em cada celular. A medição foi repetida 12 vezes em um Ryzen 5 7500F. A entrega variou de 37,5 a 48,1 ms porque o servidor agrupa as mudanças e envia a cada 50 ms, então a maior parte desse número é o intervalo de agrupamento. Isso mede o servidor, não o seu Wi-Fi, que vai somar o próprio atraso. Nada é desenhado, então isso mede a entrega até o socket do celular.

## Antes e depois

|  | Um jogo de festa pago típico | Couchcade |
| --- | --- | --- |
| Custo | Pacotes pagos | Grátis, licença MIT |
| Onde roda | Apps e servidores do editor | Um binário na sua própria rede |
| Conteúdo | Definido pelo editor | Arquivos JSON que você edita e compartilha |
| Idiomas | Depende do título | 9, com textos escritos para cada um |
| Grupos grandes | Depende do jogo | 8 jogadores e até 500 na plateia |

O Couchcade não é clone de nenhum jogo comercial. Os seis jogos e todos os textos são originais.

## Os jogos

| Jogo | O que se faz |
| --- | --- |
| Squiggle | Desenhe o tema em segredo e vote no melhor desenho |
| Fibber | Escreva uma resposta falsa e descubra a verdadeira |
| Brain Teams | Dois times correm por perguntas de múltipla escolha |
| Zap | Espere o verde e toque mais rápido que todo mundo |
| Echo | Pense igual aos outros, respostas iguais pontuam |
| Mole | Um jogador não tem a palavra secreta, encontre quem é |

As regras e o número de jogadores estão em [docs/how-to-play.md](docs/how-to-play.md).

## Início rápido

```sh
couchcade serve
```

Abra na TV o primeiro endereço mostrado e escaneie o QR code com os celulares. Se o computador tem várias placas de rede e o QR code aponta para a errada, adicione `-public-url http://192.168.1.20:8080`.

<details>
<summary>Instalar</summary>

Baixe um binário na [página de versões](https://github.com/Arthur031221/Couchcade/releases) ou use:

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

## Faça do seu jeito

Os textos são JSON comum. Coloque seu próprio pacote em uma pasta e inicie com `-packs ./my-packs`:

```json
{"id": "office", "lang": "en", "items": {"squiggle": [{"text": "a tired printer"}]}}
```

`couchcade packs validate FILE` confere um pacote e `couchcade packs stats` conta o que você tem. O formato está em [docs/content-packs.md](docs/content-packs.md), o uso atrás de um proxy em [docs/hosting.md](docs/hosting.md) e o desenho em [docs/architecture.md](docs/architecture.md).

## O que é testado

* Testes unitários em Go da validação de conteúdo, do motor e de cada jogo, com o detector de corridas na CI.
* Um teste de carga com WebSockets reais.
* 13 testes de ponta a ponta que controlam um Chrome real contra o binário real: cada um dos seis jogos com uma pessoa e bots, reconexão, expulsão, mais de 8 jogadores, temas claro e escuro, sem rolagem lateral em 320 px em todos os jogos e os nove idiomas.
* Um teste de navegador que joga uma partida completa na demo em WebAssembly.

## Limites

* Os celulares e a TV precisam alcançar o servidor, então o normal é uma única rede Wi-Fi. Jogar pela internet exige um proxy reverso que você mesmo configura.
* As salas ficam na memória. Reiniciar o servidor encerra as partidas em andamento.
* Não há login. Não exponha a porta à internet sem um proxy confiável.
* O desenho é só com dedo ou mouse, com desfazer e limpar, mas sem sensibilidade à pressão.

## Contribuir

Textos e traduções são a entrada mais fácil. Veja [CONTRIBUTING.md](CONTRIBUTING.md).

## Licença

[MIT](LICENSE)

Assisted by Claude/Codex.
