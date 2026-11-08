<p align="center"><img src="assets/logo.svg" width="96" alt="Couchcade logo"></p>
<h1 align="center">Couchcade</h1>
<p align="center">Des jeux de soirée gratuits pour une télé et les téléphones de tout le monde. Un seul binaire Go, sans compte ni appli.</p>

<p align="center">
<a href="https://github.com/Arthur031221/Couchcade/actions/workflows/ci.yml"><img src="https://github.com/Arthur031221/Couchcade/actions/workflows/ci.yml/badge.svg" alt="ci"></a>
<a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT"></a>
</p>

<p align="center"><a href="README.md">English</a> | <a href="README.zh-TW.md">繁體中文</a> | <a href="README.zh-CN.md">简体中文</a> | <a href="README.ja.md">日本語</a> | <a href="README.ko.md">한국어</a> | <a href="README.es.md">Español</a> | <b>Français</b> | <a href="README.de.md">Deutsch</a> | <a href="README.pt-BR.md">Português</a></p>

<p align="center"><img src="assets/demo.gif" width="760" alt="Une manche de Squiggle sur la télé avec des téléphones qui votent"></p>

Ouvrez Couchcade sur une télé ou un ordinateur portable. Vos amis scannent le QR code, saisissent le code de quatre lettres et jouent depuis leur téléphone. Il y a six jeux originaux, un mode public pour les grandes tablées, des bots quand il manque du monde et neuf langues avec des textes écrits pour chacune. Les points s'additionnent d'un jeu à l'autre dans un classement de la soirée, et les téléphones restent allumés pendant la partie.

**À essayer tout de suite :** la [démo en ligne](https://arthur031221.github.io/Couchcade/) fait tourner le même moteur compilé en WebAssembly dans votre navigateur, avec des bots pour remplir les places.

## Le chiffre

Avec 100 téléphones simulés (8 assis et 92 dans le public), chaque téléphone a reçu la première question moins de 60 ms après l'appui du maître de jeu sur Démarrer, sur 12 essais sur 12.

Méthode : `couchcade loadtest -phones 100` lance un serveur dans le même processus, ouvre 100 vraies connexions WebSocket en loopback, démarre une manche de Brain Teams depuis la connexion de l'hôte et relève l'heure à laquelle le premier message d'état contenant la question arrive sur chaque téléphone. La mesure a été répétée 12 fois sur un Ryzen 5 7500F. La diffusion a varié de 37,5 à 48,1 ms parce que le serveur regroupe les changements et les envoie toutes les 50 ms, donc l'essentiel de ce chiffre est cet intervalle. Cela mesure le serveur, pas votre Wi-Fi, qui ajoutera son propre délai. Rien n'est affiché, donc cela mesure la livraison jusqu'au socket du téléphone.

## Avant et après

|  | Un jeu de soirée payant typique | Couchcade |
| --- | --- | --- |
| Coût | Des packs payants | Gratuit, licence MIT |
| Où ça tourne | Les applis et serveurs de l'éditeur | Un binaire sur votre propre réseau |
| Contenu | Fixé par l'éditeur | Des fichiers JSON que vous modifiez et partagez |
| Langues | Selon le titre | 9, avec des textes écrits pour chacune |
| Grands groupes | Selon le jeu | 8 joueurs et jusqu'à 500 spectateurs |

Couchcade n'est le clone d'aucun jeu commercial. Les six jeux et tous les textes sont originaux.

## Les jeux

| Jeu | Ce qu'on fait |
| --- | --- |
| Squiggle | Dessiner le thème en secret, puis voter pour le meilleur dessin |
| Fibber | Écrire une fausse réponse, puis repérer la vraie |
| Brain Teams | Deux équipes enchaînent des questions à choix multiple |
| Zap | Attendre le vert, puis toucher plus vite que tout le monde |
| Echo | Penser pareil, les réponses identiques marquent des points |
| Mole | Un joueur n'a pas le mot secret, le trouver |

Les règles et le nombre de joueurs sont dans [docs/how-to-play.md](docs/how-to-play.md).

## Démarrage rapide

```sh
couchcade serve
```

Ouvrez sur la télé la première adresse affichée et scannez le QR code avec les téléphones. Si votre machine a plusieurs cartes réseau et que le QR code pointe vers la mauvaise, ajoutez `-public-url http://192.168.1.20:8080`.

<details>
<summary>Installer</summary>

Téléchargez un binaire depuis la [page des versions](https://github.com/Arthur031221/Couchcade/releases) ou utilisez :

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

## À vous de jouer

Les textes sont du JSON ordinaire. Mettez votre propre pack dans un dossier et lancez avec `-packs ./my-packs` :

```json
{"id": "office", "lang": "en", "items": {"squiggle": [{"text": "a tired printer"}]}}
```

`couchcade packs validate FILE` vérifie un pack et `couchcade packs stats` compte ce que vous avez. Le format est dans [docs/content-packs.md](docs/content-packs.md), l'hébergement derrière un proxy dans [docs/hosting.md](docs/hosting.md) et la conception dans [docs/architecture.md](docs/architecture.md).

## Ce qui est testé

* Des tests unitaires Go pour la validation du contenu, le moteur et chaque jeu, avec le détecteur de courses en CI.
* Un test de charge avec de vrais WebSockets.
* 16 tests de bout en bout qui pilotent un vrai Chrome contre le vrai binaire : chacun des six jeux avec un humain et des bots, la reconnexion, l'exclusion, plus de 8 joueurs, les thèmes clair et sombre, aucun défilement horizontal à 320 px dans chaque jeu et les neuf langues avec les libellés des boutons traduits.
* Un test de navigateur qui joue une partie complète sur la démo WebAssembly.

## Limites

* Les téléphones et la télé doivent joindre le serveur, donc le cas courant est un seul réseau Wi-Fi. Jouer par internet demande un proxy inverse que vous configurez.
* Les salons vivent en mémoire. Redémarrer le serveur met fin aux parties en cours.
* Il n'y a pas de connexion. N'exposez pas le port à internet sans un proxy de confiance.
* Le dessin se fait au doigt ou à la souris, avec annuler et effacer, mais sans sensibilité à la pression.

## Contribuer

Les textes et les traductions sont la porte d'entrée la plus simple. Voir [CONTRIBUTING.md](CONTRIBUTING.md).

## Licence

[MIT](LICENSE)

Assisted by Claude/Codex.
