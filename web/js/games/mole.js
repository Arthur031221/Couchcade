import { h } from "../dom.js";
import { chips, hud, byId, avatar, gain, button, waiting } from "../common.js";

let draft = "";

function clueCards(S, C) {
  const g = S.g, P = byId(S);
  return h("div", { class: "clues" }, g.ids.map((id, i) => {
    const p = P[id];
    if (!p) return null;
    const rev = g.phase === "reveal";
    return h("div", { key: "c" + id, class: `clue${rev && g.mole === id ? " mole" : ""}`, style: { "--i": i, "--c": p.color } }, avatar(p), h("b", {}, p.name), h("span", { class: "tx" }, (g.clues || {})[id] || "…"),
      rev ? h("small", {}, (Object.entries(g.votes || {}).filter(([, t]) => t === id).length) + " ✋", " ", gain(g.gain[id])) : null);
  }));
}

export default {
  tv(S, C) {
    const g = S.g, P = byId(S);
    if (g.phase === "clue") {
      return h("section", { class: "stage-in" }, hud(S, C), h("p", { class: "kicker" }, C.t("mole.category")), h("h1", { class: "huge", key: "c" + g.round }, g.category), h("p", { class: "hint" }, C.t("mole.hint")), chips(S, g.done));
    }
    if (g.phase === "vote") {
      return h("section", { class: "stage-in" }, hud(S, C), h("h2", { class: "big" }, C.t("mole.who")), clueCards(S, C), chips(S, g.done));
    }
    if (g.phase === "guess") {
      return h("section", { class: "stage-in" }, hud(S, C), h("h2", { class: "big" }, C.t("mole.caught", { name: (P[g.mole] || {}).name })), h("p", { class: "hint" }, C.t("mole.guessing")), clueCards(S, C));
    }
    return h("section", { class: "stage-in" }, hud(S, C),
      h("h2", { class: "big" }, g.caught ? C.t("mole.was_caught", { name: (P[g.mole] || {}).name }) : C.t("mole.escaped", { name: (P[g.mole] || {}).name })),
      h("p", { class: "kicker" }, C.t("mole.word"), ": ", h("b", {}, g.word), g.guess ? h("span", {}, " · ", C.t("mole.guess"), ": ", g.guess) : null),
      clueCards(S, C), chips(S, [], true));
  },

  phone(S, C) {
    const g = S.g, me = S.me.g || {};
    if (S.me.aud || !g.ids.includes(S.me.id)) return waiting(C, C.t("audience.watching"));
    const role = me.mole
      ? h("div", { class: "role mole" }, h("b", {}, C.t("mole.you_mole")), h("small", {}, C.t("mole.category"), ": ", g.category))
      : h("div", { class: "role agent" }, h("small", {}, C.t("mole.secret")), h("b", {}, me.word), h("small", {}, g.category));
    if (g.phase === "clue") {
      const send = () => draft.trim() && C.send({ k: "submit", s: draft });
      return h("section", { class: "ctl" }, role, h("input", { key: "in" + g.round, class: "field", maxlength: 24, autocomplete: "off", placeholder: C.t("mole.placeholder"), oninput: (e) => { draft = e.target.value; }, onkeydown: (e) => e.key === "Enter" && send(), ref: (el) => { draft = ""; } }),
        button(me.clue ? C.t("update") : C.t("send"), send, "primary wide"), me.clue ? h("p", { class: "hint ok" }, C.t("mole.locked", { text: me.clue })) : null);
    }
    if (g.phase === "vote") {
      const P = byId(S);
      return h("section", { class: "ctl" }, role, h("h2", { class: "big small" }, C.t("mole.who")),
        h("div", { class: "votes" }, g.ids.map((id, i) => id === S.me.id ? null : h("button", { key: "v" + id, type: "button", class: `vote${me.vote === i ? " picked" : ""}`, style: { "--c": (P[id] || {}).color }, onclick: () => C.send({ k: "vote", i }) },
          avatar(P[id] || { name: "?", color: 0 }), h("b", {}, (P[id] || {}).name), h("span", {}, (g.clues || {})[id])))));
    }
    if (g.phase === "guess" && me.mole) {
      const send = () => draft.trim() && C.send({ k: "guess", s: draft });
      return h("section", { class: "ctl" }, h("h2", { class: "big" }, C.t("mole.guess_prompt")), h("input", { key: "gu", class: "field", maxlength: 40, autocomplete: "off", oninput: (e) => { draft = e.target.value; }, onkeydown: (e) => e.key === "Enter" && send(), ref: (el) => { draft = ""; setTimeout(() => el.focus(), 100); } }), button(C.t("send"), send, "primary wide"));
    }
    if (g.phase === "reveal") {
      const nm = (byId(S)[g.mole] || {}).name;
      return h("section", { class: "ctl center" }, h("h2", { class: "big" }, g.caught ? C.t("mole.was_caught", { name: nm }) : C.t("mole.escaped", { name: nm })), gain(me.gain));
    }
    return waiting(C, C.t("mole.waiting"));
  },
};
