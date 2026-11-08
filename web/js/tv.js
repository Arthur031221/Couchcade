import { h } from "./dom.js";
import { Link } from "./net.js";
import { initShell, store, mount, toggleTheme, toggleSound, i18n } from "./app.js";
import { GAMES, ICON, chip, avatar, num, button, byId, color } from "./common.js";
import { sfx, isMuted } from "./sound.js";
import squiggle from "./games/squiggle.js";
import fibber from "./games/fibber.js";
import trivia from "./games/trivia.js";
import zap from "./games/zap.js";
import echo from "./games/echo.js";
import mole from "./games/mole.js";

const VIEWS = { squiggle, fibber, trivia, zap, echo, mole };
const S = { room: null, g: null, heavy: null, hk: "", conn: "connecting", rounds: 0, err: "" };
let link;

const C = { t: i18n.t, render: () => render(), send: () => {} };
const host = (op, extra) => link.send({ t: "host", op, ...extra });
const root = document.getElementById("app");
const render = mount(root, view);

let lastPhase = "", lastKey = "";
function onMessage(m) {
  if (m.t === "tvok") {
    store.set("tv.code", m.code);
    store.set("tv.token", m.token);
    S.code = m.code;
  } else if (m.t === "state") {
    S.room = m.room;
    S.g = m.g;
    if (m.hk !== undefined) { if (m.heavy !== undefined) S.heavy = m.heavy; S.hk = m.hk; } else { S.heavy = null; S.hk = ""; }
    if (S.langApplied !== m.room.lang) {
      S.langApplied = m.room.lang;
      i18n.setLang(m.room.lang).then(() => { store.set("lang", m.room.lang); render(); });
    }
    cues(m);
  } else if (m.t === "err") {
    S.err = m.code;
    clearTimeout(S.errT);
    S.errT = setTimeout(() => { S.err = ""; render(); }, 3500);
  }
  render();
}

function cues(m) {
  const key = m.room.phase + (m.g ? m.g.phase + m.g.round : "");
  if (key === lastKey) return;
  lastKey = key;
  if (m.room.phase === "final") sfx.win();
  else if (m.g && m.g.phase === "go") sfx.go();
  else if (m.g && m.g.phase === "reveal") sfx.reveal();
  else if (m.g) sfx.ok();
  lastPhase = m.room.phase;
}

function start(game, info) {
  host("start", { game, rounds: S.rounds || info.rounds });
}

function lobby() {
  const r = S.room;
  const demo = !!(window.parent !== window && window.parent.__couchEngine);
  const need = (info) => r.players.length < info.min;
  return h("main", { class: "lobby stage-in" },
    h("div", { class: "join" },
      demo ? null : h("p", { class: "kicker" }, i18n.t("lobby.join_at"), " ", h("b", {}, r.join.replace(/^https?:\/\//, "").replace(/\/play\?c=.*/, ""))),
      h("div", { class: "code", "aria-label": i18n.t("lobby.code") }, [...r.code].map((c, i) => h("span", { key: "cd" + i, style: { "--i": i } }, c))),
      demo ? h("p", { class: "hint" }, i18n.t("lobby.demo_note")) : h("img", { class: "qr", alt: i18n.t("a11y.qr"), src: `api/qr?c=${r.code}` })),
    h("div", { class: "side" },
      h("h2", {}, i18n.t("lobby.players"), " ", h("small", {}, r.players.length, "/", r.max), r.audience ? h("small", {}, " · ", i18n.t("lobby.audience", { n: r.audience })) : null),
      r.played ? h("p", { class: "kicker standings" }, i18n.t("lobby.standings")) : null,
      h("div", { class: "chips big" }, r.players.length ? (r.played ? [...r.players].sort((a, b) => b.total - a.total) : r.players).map((p) => chip(p, false, [
        p.bot ? h("small", { class: "pill" }, i18n.t("lobby.bot")) : null, r.played ? h("b", { class: "sc" }, num(p.total)) : null])) : h("p", { class: "hint" }, i18n.t("lobby.waiting"))),
      h("div", { class: "games" }, r.catalog.filter((i) => GAMES.includes(i.id)).map((info) => h("button", {
        key: info.id, type: "button", class: "gamecard", disabled: need(info), onclick: () => start(info.id, info),
      }, h("span", { class: "gi" }, ICON[info.id]), h("b", {}, i18n.t(`${info.id}.name`)), h("small", {}, i18n.t(`${info.id}.blurb`)),
      h("em", {}, need(info) ? i18n.t("lobby.need", { n: info.min }) : i18n.t("lobby.rounds", { n: S.rounds || info.rounds }))))),
      h("div", { class: "row" },
        button(i18n.t("lobby.bots"), () => host("bots", { n: 3 }), "ghost"),
        r.players.some((p) => p.bot) ? button(i18n.t("lobby.nobots"), () => host("nobots"), "ghost") : null,
        r.played ? button(i18n.t("lobby.reset"), () => host("resettotals"), "ghost") : null,
        h("label", { class: "toggle" }, h("input", { type: "checkbox", checked: r.family, onchange: (e) => host("family", { on: e.target.checked }) }), i18n.t("lobby.family")),
        h("label", { class: "toggle" }, i18n.t("lobby.length"), h("select", { class: "sel", onchange: (e) => { S.rounds = +e.target.value; render(); } },
          [[0, "lobby.len_default"], [2, "lobby.len_short"], [5, "lobby.len_long"]].map(([v, k]) => h("option", { key: v, value: v, selected: S.rounds === v }, i18n.t(k))))))));
}

// The party total, shown on the final screen once more than one game has been played.
function total(p) {
  return S.room.played > 1 ? h("small", { class: "total" }, i18n.t("final.total"), " ", p.total) : null;
}

function final() {
  const sorted = [...S.room.players].sort((a, b) => b.score - a.score);
  const top = sorted.slice(0, 3);
  const order = [1, 0, 2].filter((i) => top[i]);
  return h("main", { class: "final stage-in" }, h("h1", { class: "huge" }, i18n.t("final.title")),
    h("div", { class: "podium" }, order.map((i) => h("div", { key: "pd" + top[i].id, class: `step s${i + 1}`, style: { "--c": color(top[i]) } },
      avatar(top[i], "l"), h("b", {}, top[i].name), num(top[i].score), total(top[i]), h("div", { class: "block" }, i + 1)))),
    h("div", { class: "rest" }, sorted.slice(3).map((p, i) => h("div", { key: "r" + p.id, class: "line" }, h("span", {}, i + 4), avatar(p, "s"), h("b", {}, p.name), total(p), num(p.score)))),
    h("div", { class: "row" }, button(i18n.t("final.again"), () => host("start", { game: S.room.game, rounds: S.rounds }), "primary"), button(i18n.t("final.lobby"), () => host("lobby"), "ghost")));
}

function topbar() {
  const r = S.room;
  return h("header", { class: "top" },
    h("span", { class: "brand" }, h("img", { src: "logo.svg", alt: "", width: 28, height: 28 }), "Couchcade"),
    r ? h("span", { class: "pill code-pill" }, r.code) : null,
    S.conn !== "open" ? h("span", { class: "pill bad" }, i18n.t("net." + S.conn)) : null,
    h("span", { class: "grow" }),
    r && r.phase === "lobby" ? i18n.LANGS && h("select", { class: "sel", "aria-label": i18n.t("a11y.language"), onchange: (e) => host("lang", { lang: e.target.value }) },
      i18n.LANGS.map((l) => h("option", { key: l[0], value: l[0], selected: l[0] === r.lang }, l[1]))) : null,
    r && r.phase !== "lobby" ? button(i18n.t("end_game"), () => host("lobby"), "ghost small") : null,
    h("button", { class: "btn ghost icon", type: "button", onclick: () => { toggleSound(); render(); }, "aria-label": i18n.t("a11y.sound"), "aria-pressed": String(!isMuted()) }, isMuted() ? "🔇" : "🔊"),
    h("button", { class: "btn ghost icon", type: "button", onclick: () => { toggleTheme(); render(); }, "aria-label": i18n.t("a11y.theme") }, "◐"),
    h("button", { class: "btn ghost icon", type: "button", onclick: () => (document.fullscreenElement ? document.exitFullscreen() : document.documentElement.requestFullscreen()), "aria-label": i18n.t("a11y.fullscreen") }, "⛶"));
}

function view() {
  const r = S.room;
  let body;
  if (!r) body = h("main", { class: "boot" }, h("p", {}, i18n.t("net." + S.conn)));
  else if (r.phase === "lobby") body = lobby();
  else if (r.phase === "final") body = final();
  else body = h("main", { class: `game g-${r.game}`, "data-phase": S.g.phase }, VIEWS[r.game].tv(S, C));
  return h("div", { class: "shell tv" }, topbar(), body, S.err ? h("div", { class: "toast", role: "alert" }, i18n.t("err." + S.err)) : null);
}

(async function main() {
  initShell();
  await i18n.setLang(i18n.detect(store.get("lang")));
  link = new Link({
    onmessage: onMessage,
    onstatus: (s) => { S.conn = s; render(); },
    hello: () => ({ t: "tv", code: store.get("tv.code", ""), token: store.get("tv.token", "") }),
  });
  window.__couchTV = { S, host };
  link.connect();
  render();
})();
