import { h } from "./dom.js";
import { Link } from "./net.js";
import { initShell, store, mount, toggleTheme, toggleSound, i18n } from "./app.js";
import { GAMES, ICON, avatar, button, waiting, chip } from "./common.js";
import { keepAwake } from "./awake.js";
import { sfx, isMuted } from "./sound.js";
import squiggle from "./games/squiggle.js";
import fibber from "./games/fibber.js";
import trivia from "./games/trivia.js";
import zap from "./games/zap.js";
import echo from "./games/echo.js";
import mole from "./games/mole.js";

const VIEWS = { squiggle, fibber, trivia, zap, echo, mole };
const params = new URLSearchParams(location.search);
const S = { room: null, g: null, heavy: null, hk: "", me: null, conn: "idle", err: "", code: (params.get("c") || "").toUpperCase(), name: params.get("name") || store.get("name", ""), joined: false, manage: false };
let link;

const send = (a) => link.send({ t: "act", a });
const C = { t: i18n.t, render: () => render(), send };
const host = (op, extra) => link.send({ t: "host", op, ...extra });
const root = document.getElementById("app");
const render = mount(root, view);

function toast(code) {
  S.err = code;
  clearTimeout(S.errT);
  S.errT = setTimeout(() => { S.err = ""; render(); }, 3500);
}

let lastKey = "";
function onMessage(m) {
  if (m.t === "joined") {
    S.joined = true;
    store.set("tok." + m.code, m.token);
    store.set("name", S.name);
    S.code = m.code;
    history.replaceState(null, "", `?c=${m.code}`);
    sfx.join();
    keepAwake();
  } else if (m.t === "state") {
    S.room = m.room;
    S.g = m.g;
    S.me = m.me;
    if (m.hk !== undefined) { if (m.heavy !== undefined) S.heavy = m.heavy; S.hk = m.hk; } else { S.heavy = null; S.hk = ""; }
    if (S.langApplied !== m.room.lang) {
      S.langApplied = m.room.lang;
      i18n.setLang(m.room.lang).then(render);
    }
    const key = m.room.phase + (m.g ? m.g.phase + m.g.round : "");
    if (key !== lastKey) {
      lastKey = key;
      if (m.g && m.g.phase === "go") { sfx.go(); navigator.vibrate && navigator.vibrate(60); }
      else if (m.g) sfx.tick();
    }
  } else if (m.t === "err") {
    if (m.code === "room_not_found" || m.code === "name_required" || m.code === "kicked" || m.code === "room_full") {
      S.joined = false;
      S.room = null;
      link.close();
      S.conn = "idle";
    }
    toast(m.code);
  } else if (m.t === "kicked") {
    S.joined = false;
    S.room = null;
    link.close();
    S.conn = "idle";
    toast("kicked");
  }
  render();
}

function connect() {
  if (S.code.length !== 4 || !S.name.trim()) return;
  keepAwake(); // from the Join tap, so iOS allows the fallback video to start
  if (link) link.close();
  link = new Link({
    onmessage: onMessage,
    onstatus: (s) => { S.conn = s; render(); },
    hello: () => ({ t: "join", code: S.code, name: S.name.trim(), token: store.get("tok." + S.code, "") }),
  });
  link.connect();
}

function joinForm() {
  return h("main", { class: "joinform stage-in" },
    h("img", { src: "logo.svg", alt: "", width: 72, height: 72 }),
    h("h1", {}, "Couchcade"),
    h("label", {}, i18n.t("join.code"), h("input", { class: "field code", maxlength: 4, autocapitalize: "characters", autocomplete: "off", value: S.code, oninput: (e) => { S.code = e.target.value.toUpperCase().replace(/[^A-Z]/g, ""); e.target.value = S.code; render(); } })),
    h("label", {}, i18n.t("join.name"), h("input", { class: "field", maxlength: 12, autocomplete: "nickname", value: S.name, oninput: (e) => { S.name = e.target.value; render(); }, onkeydown: (e) => e.key === "Enter" && connect() })),
    button(i18n.t("join.go"), connect, "primary wide", { disabled: S.code.length !== 4 || !S.name.trim() }),
    S.err ? h("p", { class: "hint bad", role: "alert" }, i18n.t("err." + S.err)) : null);
}

function hostPanel() {
  const r = S.room;
  if (!S.me.host) return null;
  const lobby = r.phase === "lobby";
  return h("section", { class: "hostbox" },
    h("p", { class: "kicker" }, i18n.t("host.you")),
    lobby ? h("div", { class: "games" }, r.catalog.filter((i) => GAMES.includes(i.id)).map((info) => h("button", {
      key: info.id, type: "button", class: "gamecard", disabled: r.players.length < info.min, onclick: () => host("start", { game: info.id, rounds: info.rounds }),
    }, h("span", { class: "gi" }, ICON[info.id]), h("b", {}, i18n.t(`${info.id}.name`)), h("small", {}, r.players.length < info.min ? i18n.t("lobby.need", { n: info.min }) : i18n.t(`${info.id}.blurb`))))) : null,
    h("div", { class: "row" },
      !lobby ? button(i18n.t("end_game"), () => host("lobby"), "ghost small") : null,
      lobby ? button(i18n.t("lobby.bots"), () => host("bots", { n: 3 }), "ghost small") : null,
      button(S.manage ? i18n.t("host.hide") : i18n.t("host.players"), () => { S.manage = !S.manage; render(); }, "ghost small")),
    S.manage ? h("div", { class: "manage" }, r.players.filter((p) => p.id !== S.me.id).map((p) => h("div", { key: "m" + p.id, class: "line" },
      avatar(p, "s"), h("b", {}, p.name),
      button(p.muted ? i18n.t("host.unmute") : i18n.t("host.mute"), () => host("mute", { id: p.id }), "ghost small"),
      button(i18n.t("host.kick"), () => host("kick", { id: p.id }), "ghost small danger")))) : null);
}

function body() {
  const r = S.room;
  if (r.phase === "lobby") {
    return h("main", { class: "ctl center stage-in" },
      h("div", { class: "me-card" }, avatar(r.players.find((p) => p.id === S.me.id) || { name: S.name, color: 0 }, "l"), h("b", {}, S.name), S.me.aud ? h("small", {}, i18n.t("audience.label")) : null,
        r.played && !S.me.aud ? h("small", {}, i18n.t("final.total"), ": ", S.me.total, " ", i18n.t("pts")) : null),
      waiting(C, S.me.aud ? i18n.t("audience.lobby") : i18n.t("phone.lobby")),
      h("div", { class: "chips" }, r.players.map((p) => chip(p, false))),
      hostPanel());
  }
  if (r.phase === "final") {
    const sorted = [...r.players].sort((a, b) => b.score - a.score);
    const rank = sorted.findIndex((p) => p.id === S.me.id) + 1;
    return h("main", { class: "ctl center stage-in" }, h("p", { class: "kicker" }, i18n.t("final.title")),
      h("h1", { class: "mega" }, rank ? "#" + rank : "—"), h("p", { class: "big" }, S.me.score, " ", i18n.t("pts")),
      r.played > 1 && !S.me.aud ? h("p", { class: "hint" }, i18n.t("final.total"), ": ", S.me.total, " ", i18n.t("pts")) : null, hostPanel(),
      S.me.host ? button(i18n.t("final.again"), () => host("start", { game: r.game }), "primary wide") : null);
  }
  return h("main", { class: `game g-${r.game}`, "data-phase": S.g.phase },
    h("div", { class: "phonebar" }, h("span", {}, ICON[r.game], " ", i18n.t(`${r.game}.name`)), h("b", {}, S.me.score, " ", i18n.t("pts"))),
    VIEWS[r.game].phone({ ...S, me: { ...S.me } }, C), hostPanel());
}

function view() {
  return h("div", { class: "shell phone" },
    h("header", { class: "top" }, h("span", { class: "brand" }, h("img", { src: "logo.svg", alt: "", width: 24, height: 24 }), "Couchcade"),
      S.joined ? h("span", { class: "pill code-pill" }, S.code) : null,
      S.conn === "reconnecting" ? h("span", { class: "pill bad" }, i18n.t("net.reconnecting")) : null,
      h("span", { class: "grow" }),
      h("button", { class: "btn ghost icon", type: "button", onclick: () => { toggleSound(); render(); }, "aria-label": "Sound" }, isMuted() ? "🔇" : "🔊"),
      h("button", { class: "btn ghost icon", type: "button", onclick: () => { toggleTheme(); render(); }, "aria-label": "Theme" }, "◐")),
    S.room && S.me ? body() : S.conn === "idle" ? joinForm() : h("main", { class: "boot" }, h("p", {}, i18n.t("net." + S.conn))),
    S.err && S.room ? h("div", { class: "toast", role: "alert" }, i18n.t("err." + S.err)) : null);
}

(async function main() {
  initShell();
  await i18n.setLang(i18n.detect(store.get("lang")));
  window.__couchPhone = { S };
  if (S.code.length === 4 && S.name.trim() && (params.get("name") || store.get("tok." + S.code))) connect();
  render();
})();
