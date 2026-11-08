import { h } from "../dom.js";
import { chips, hud, byId, avatar, gain, button, waiting } from "../common.js";

let draft = "";

export default {
  tv(S, C) {
    const g = S.g, P = byId(S);
    if (g.phase === "answer") {
      return h("section", { class: "stage-in" }, hud(S, C), h("h1", { class: "huge", key: "p" + g.round }, g.prompt), h("p", { class: "hint" }, C.t("echo.hint")), chips(S, g.done));
    }
    return h("section", { class: "stage-in" }, hud(S, C), h("h2", { class: "big" }, g.prompt),
      h("div", { class: "bubbles" }, g.groups.map((gr, i) => h("div", { key: "g" + gr.text + i, class: `bubble${gr.ids.length > 1 ? " match" : ""}`, style: { "--i": i, "--n": gr.ids.length } },
        h("span", { class: "tx" }, gr.text), h("div", { class: "who" }, gr.ids.map((id) => avatar(P[id] || { name: "?", color: 0 }, "s"))),
        gr.ids.length > 1 ? h("b", { class: "gain pos" }, "+", (gr.ids.length - 1) * 100) : null))),
      chips(S, [], true));
  },

  phone(S, C) {
    const g = S.g, me = S.me.g || {};
    if (S.me.aud) return waiting(C, C.t("audience.watching"));
    if (g.phase === "answer") {
      const send = () => draft.trim() && C.send({ k: "submit", s: draft });
      return h("section", { class: "ctl" }, h("p", { class: "kicker" }, C.t("echo.match")), h("h2", { class: "big" }, g.prompt),
        h("input", { key: "in" + g.round, class: "field", maxlength: 30, autocomplete: "off", placeholder: C.t("echo.placeholder"), oninput: (e) => { draft = e.target.value; }, onkeydown: (e) => e.key === "Enter" && send(), ref: (el) => { draft = ""; setTimeout(() => el.focus(), 100); } }),
        button(me.text ? C.t("update") : C.t("send"), send, "primary wide"),
        me.text ? h("p", { class: "hint ok" }, C.t("echo.locked", { text: me.text })) : null);
    }
    return h("section", { class: "ctl center" }, h("h2", { class: "big" }, me.text || "—"), gain(me.gain) || h("p", { class: "hint" }, C.t("echo.nomatch")));
  },
};
