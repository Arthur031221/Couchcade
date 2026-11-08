import { h } from "../dom.js";
import { chips, hud, byId, avatar, gain, button, waiting, color } from "../common.js";
import { drawingSvg, DrawPad, INK, BRUSH } from "../draw.js";

let pad = null;
const ui = { color: 0, width: 1, count: 0 };

function grid(S, C, onPick, me) {
  const P = byId(S);
  const res = S.g.results;
  const best = res ? Math.max(...res.map((r) => r.gain)) : 0;
  return h("div", { class: "gallery" }, (S.heavy || []).map((d, i) => {
    const r = res && res[i];
    const mine = me && me.own === i;
    const picked = me && me.vote === i;
    return h("button", {
      key: "d" + i, type: "button", disabled: !onPick || mine, onclick: onPick ? () => onPick(i) : null,
      class: `card art${r && r.gain === best && best > 0 ? " win" : ""}${picked ? " picked" : ""}${mine ? " mine" : ""}`,
      style: { "--i": i },
    }, drawingSvg(d),
    h("span", { class: "tag" }, mine ? C.t("squiggle.yours") : r ? "" : i + 1),
    r ? h("div", { class: "by" }, avatar(P[r.author] || { name: "?", color: 0 }, "s"), h("span", {}, (P[r.author] || {}).name),
      h("b", {}, "♥ ", r.votes + (r.crowd ? ` +${r.crowd}` : "")), gain(r.gain)) : null);
  }));
}

export default {
  tv(S, C) {
    const g = S.g;
    if (g.phase === "draw") {
      return h("section", { class: "stage-in" }, hud(S, C),
        h("p", { class: "kicker" }, C.t("squiggle.draw_this")),
        h("h1", { class: "huge", key: "p" + g.round }, g.prompt),
        chips(S, g.done), h("p", { class: "hint" }, C.t("squiggle.hint")));
    }
    return h("section", { class: "stage-in" }, hud(S, C),
      h("h2", { class: "big" }, g.prompt),
      grid(S, C),
      g.phase === "vote" ? h("p", { class: "hint" }, C.t("squiggle.vote_hint"), g.crowd ? ` · ${C.t("crowd_votes", { n: g.crowd })}` : "") : null,
      chips(S, g.voted));
  },

  phone(S, C) {
    const g = S.g, me = S.me.g || {};
    if (S.me.aud && g.phase === "draw") return waiting(C, C.t("audience.watching"));
    if (g.phase === "draw") {
      const tools = h("div", { class: "tools" },
        h("div", { class: "swatches" }, INK.map((c, i) => h("button", { key: i, type: "button", class: `sw${ui.color === i ? " on" : ""}`, style: { background: c }, "aria-label": c, onclick: () => { ui.color = i; if (pad) pad.color = i; C.render(); } }))),
        h("div", { class: "brushes" }, BRUSH.map((w, i) => h("button", { key: i, type: "button", class: `br${ui.width === i ? " on" : ""}`, onclick: () => { ui.width = i; if (pad) pad.width = i; C.render(); } }, h("i", { style: { width: 6 + i * 6 + "px", height: 6 + i * 6 + "px" } })))),
        button("↶", () => pad && pad.undo(), "ghost icon", { "aria-label": C.t("undo") }),
        button("🗑", () => pad && pad.clear(), "ghost icon", { "aria-label": C.t("clear") }));
      return h("section", { class: "ctl draw" },
        h("p", { class: "kicker" }, C.t("squiggle.draw_this")), h("h2", { class: "big" }, g.prompt),
        h("canvas", { key: "pad" + g.round, class: "pad", ref: (el) => { pad = new DrawPad(el, (p) => { ui.count = p.strokes.length; C.render(); }); pad.color = ui.color; pad.width = ui.width; } }),
        tools,
        button(me.submitted ? C.t("squiggle.resend") : C.t("squiggle.send"), () => pad && pad.strokes.length && C.send({ k: "submit", p: pad.strokes }), "primary wide", { disabled: !ui.count }),
        me.submitted ? h("p", { class: "hint ok" }, C.t("squiggle.sent")) : null);
    }
    if (g.phase === "vote") {
      return h("section", { class: "ctl" }, h("p", { class: "kicker" }, g.prompt), h("h2", { class: "big" }, C.t("squiggle.pick")),
        grid(S, C, (i) => C.send({ k: "vote", i }), me));
    }
    return h("section", { class: "ctl center" }, grid(S, C, null, me));
  },
};
