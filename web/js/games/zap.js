import { h } from "../dom.js";
import { chips, hud, byId, avatar, gain, waiting } from "../common.js";

export default {
  tv(S, C) {
    const g = S.g, P = byId(S);
    if (g.phase === "wait") {
      return h("section", { class: "zap wait stage-in" }, hud(S, C), h("h1", { class: "huge" }, C.t("zap.wait")), h("p", { class: "hint" }, C.t("zap.dont_tap")),
        h("div", { class: "chips" }, (g.early || []).map((id) => h("span", { key: id, class: "pill bad" }, (P[id] || {}).name, " ✕"))));
    }
    if (g.phase === "go") {
      return h("section", { class: "zap go" }, h("h1", { class: "mega", key: "go" + g.round }, C.t("zap.go")), h("p", { class: "hint" }, g.tapped, " / ", g.total));
    }
    const top = Math.max(1, ...(g.ranks || []).map((r) => r.ms));
    return h("section", { class: "stage-in" }, hud(S, C),
      h("div", { class: "ranks" }, (g.ranks || []).map((r, i) => h("div", { key: r.id, class: "rank", style: { "--i": i, "--c": (P[r.id] || {}).color } },
        h("span", { class: "pos" }, i + 1), avatar(P[r.id] || { name: "?", color: 0 }), h("span", { class: "nm" }, (P[r.id] || {}).name),
        h("span", { class: "bar", style: { "--w": (r.ms / top) * 100 + "%" } }), h("b", {}, r.ms, " ms"), gain(g.gain[r.id])))),
      (g.early || []).length ? h("p", { class: "hint" }, C.t("zap.early"), ": ", g.early.map((id) => (P[id] || {}).name).join(", ")) : null,
      chips(S, []));
  },

  phone(S, C) {
    const g = S.g, me = S.me.g || {};
    if (S.me.aud) return waiting(C, C.t("audience.watching"));
    const tap = (e) => { e.preventDefault(); C.send({ k: "tap" }); };
    if (g.phase === "reveal" || g.phase === "end") {
      return h("section", { class: "ctl center" }, h("h2", { class: "big" }, me.early ? C.t("zap.too_early") : me.tapped ? C.t("zap.nice") : C.t("zap.missed")), gain(me.gain));
    }
    const state = me.early ? "early" : me.tapped ? "tapped" : g.phase;
    return h("section", { class: "ctl zapbox" }, h("button", {
      key: "zb" + g.round, type: "button", class: `zapbtn ${state}`, onpointerdown: tap,
      ref: () => {}, upd: (el) => { if (state === "go" && !el._buzz) { el._buzz = true; navigator.vibrate && navigator.vibrate(40); } },
    }, state === "wait" ? C.t("zap.wait") : state === "go" ? C.t("zap.go") : state === "early" ? C.t("zap.too_early") : C.t("zap.nice")));
  },
};
