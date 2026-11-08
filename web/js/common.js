import { h } from "./dom.js";

export const PCOL = ["#ff5d73", "#ffbe3b", "#8fd14f", "#4cc9f0", "#8b5cf6", "#ff8fab", "#2ec4b6", "#ff9f1c", "#5e60ce", "#06d6a0", "#f15bb5", "#00b4d8"];
export const GAMES = ["squiggle", "fibber", "trivia", "zap", "echo", "mole"];
export const ICON = { squiggle: "✎", fibber: "🎭", trivia: "❓", zap: "⚡", echo: "🔔", mole: "🕵" };
export const LETTERS = ["A", "B", "C", "D", "E", "F"];

export const color = (p) => PCOL[(p ? p.color : 0) % PCOL.length];
export const initial = (name) => ([...(name || "?")][0] || "?").toUpperCase();

export function avatar(p, size = "m") {
  return h("span", { class: `avatar ${size}${p.on === false ? " off" : ""}`, style: { "--c": color(p) }, title: p.name }, initial(p.name));
}

// A player chip: avatar, name and a state mark (done, muted, offline).
export function chip(p, done, extra) {
  return h("div", { key: p.id, class: `chip${done ? " done" : ""}${p.on === false ? " off" : ""}${p.muted ? " muted" : ""}`, style: { "--c": color(p) } },
    avatar(p), h("span", { class: "nm" }, p.name), extra, done ? h("span", { class: "tick" }, "✓") : null);
}

export function chips(S, doneIds, withScore) {
  const done = new Set(doneIds || []);
  return h("div", { class: "chips" }, S.room.players.map((p) => chip(p, done.has(p.id), withScore ? h("b", { class: "sc" }, num(p.score)) : null)));
}

let timers = new Set();
setInterval(() => {
  timers.forEach((el) => {
    if (!el.isConnected) return timers.delete(el);
    const s = Math.max(0, Math.ceil((el._end - performance.now()) / 1000));
    const b = el.querySelector("b");
    if (b && b.textContent !== String(s)) b.textContent = s;
  });
}, 250);

// The bar is created once per phase, so the CSS transition runs unbroken.
export function timer(k, ms) {
  if (!ms) return null;
  return h("div", {
    key: "timer-" + k, class: "timer", style: { "--ms": ms + "ms" },
    ref: (el) => {
      el._end = performance.now() + ms;
      timers.add(el);
      requestAnimationFrame(() => requestAnimationFrame(() => el.classList.add("run")));
    },
  }, h("i"), h("b", {}, Math.ceil(ms / 1000)));
}

export function num(v) {
  return h("span", {
    class: "num", upd: (el) => {
      const from = el._shown ?? 0;
      if (from === v) { el.textContent = v; return; }
      el._shown = v;
      const t0 = performance.now();
      const step = (t) => {
        const k = Math.min(1, (t - t0) / 700);
        el.textContent = Math.round(from + (v - from) * (1 - (1 - k) ** 3));
        if (k < 1) requestAnimationFrame(step);
      };
      requestAnimationFrame(step);
    },
  }, v);
}

export function hud(S, C) {
  const g = S.g;
  return h("div", { class: "hud" },
    h("span", { class: "pill game" }, ICON[S.room.game], " ", C.t(`${S.room.game}.name`)),
    h("span", { class: "pill" }, C.t("round", { n: g.round, total: g.rounds })),
    timer(`${g.phase}-${g.round}`, g.ms));
}

export const byId = (S) => Object.fromEntries(S.room.players.map((p) => [p.id, p]));

export function gain(n) {
  if (!n) return null;
  return h("b", { class: `gain ${n < 0 ? "neg" : "pos"}` }, (n > 0 ? "+" : "") + n);
}

export function button(label, onclick, cls = "", extra = {}) {
  return h("button", { class: "btn " + cls, onclick, type: "button", ...extra }, label);
}

export function waiting(C, text) {
  return h("div", { class: "wait" }, h("span", { class: "dots" }, h("i"), h("i"), h("i")), h("p", {}, text));
}
