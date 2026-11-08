import { h } from "../dom.js";
import { chips, hud, byId, avatar, gain, button, LETTERS, waiting } from "../common.js";

let draft = "";

function options(S, C, onPick, me) {
  const P = byId(S);
  const g = S.g;
  return h("div", { class: "options" }, g.options.map((o, i) => {
    const reveal = g.phase === "reveal";
    const mine = me && me.own.includes(i);
    return h("button", {
      key: "o" + i, type: "button", disabled: !onPick || mine, onclick: onPick ? () => onPick(i) : null, style: { "--i": i },
      class: `opt${reveal && o.real ? " real" : ""}${reveal && !o.real ? " fake" : ""}${me && me.pick === i ? " picked" : ""}${mine ? " mine" : ""}`,
    }, h("span", { class: "ltr" }, LETTERS[i]), h("span", { class: "tx" }, o.text),
    reveal ? h("div", { class: "by" }, o.real ? h("em", {}, C.t("fibber.real")) : h("em", {}, C.t("fibber.by"), " ", (o.owners || []).map((id) => (P[id] || {}).name).join(", ") || "—"),
      (o.pickers || []).map((id) => avatar(P[id] || { name: "?", color: 0 }, "s"))) : null);
  }));
}

export default {
  tv(S, C) {
    const g = S.g;
    return h("section", { class: "stage-in" }, hud(S, C),
      h("h1", { class: "quote", key: "q" + g.round }, g.q),
      g.phase === "write" ? h("p", { class: "hint" }, C.t("fibber.write_hint")) : options(S, C),
      g.phase === "pick" ? h("p", { class: "hint" }, C.t("fibber.pick_hint")) : null,
      chips(S, g.phase === "reveal" ? [] : g.done, g.phase === "reveal"));
  },

  phone(S, C) {
    const g = S.g, me = S.me.g || {};
    if (S.me.aud) return waiting(C, C.t("audience.watching"));
    if (g.phase === "write") {
      const send = () => draft.trim() && C.send({ k: "submit", s: draft });
      return h("section", { class: "ctl" }, h("p", { class: "kicker" }, C.t("fibber.write")), h("h2", { class: "big" }, g.q),
        h("input", { key: "in" + g.round, class: "field", maxlength: 40, autocomplete: "off", placeholder: C.t("fibber.placeholder"), oninput: (e) => { draft = e.target.value; }, onkeydown: (e) => e.key === "Enter" && send(), ref: (el) => { draft = ""; setTimeout(() => el.focus(), 100); } }),
        button(me.fake ? C.t("update") : C.t("send"), send, "primary wide"),
        me.fake ? h("p", { class: "hint ok" }, C.t("fibber.locked", { text: me.fake })) : null);
    }
    if (g.phase === "pick") {
      return h("section", { class: "ctl" }, h("p", { class: "kicker" }, g.q), h("h2", { class: "big" }, C.t("fibber.pick")), options(S, C, (i) => C.send({ k: "pick", i }), me));
    }
    return h("section", { class: "ctl center" }, options(S, C, null, me), gain((g.gain || {})[S.me.id]));
  },
};
