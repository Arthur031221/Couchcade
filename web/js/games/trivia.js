import { h } from "../dom.js";
import { hud, chips, LETTERS, num, waiting, gain } from "../common.js";

const TEAM = ["#ff5d73", "#4cc9f0"];

function tiles(S, C, onPick, me) {
  const g = S.g;
  const rev = g.phase === "reveal" || g.phase === "end";
  const max = rev ? Math.max(1, ...g.counts, ...g.crowd) : 1;
  return h("div", { class: "tiles" }, g.options.map((o, i) => h("button", {
    key: "t" + i, type: "button", disabled: !onPick || me.pick >= 0, onclick: onPick ? () => onPick(i) : null,
    class: `tile t${i}${rev && g.answer === i ? " right" : ""}${rev && g.answer !== i ? " dim" : ""}${me && me.pick === i ? " picked" : ""}`,
  }, h("span", { class: "ltr" }, LETTERS[i]), h("span", { class: "tx" }, o),
  rev ? h("span", { class: "bar", style: { "--w": ((g.counts[i] + g.crowd[i]) / max) * 100 + "%" } }, g.counts[i] + g.crowd[i] || "") : null)));
}

function teams(S, C) {
  const g = S.g;
  const names = (k) => S.room.players.filter((p) => g.teams[p.id] === k).map((p) => p.name).join(", ");
  return h("div", { class: "teams" }, [0, 1].map((k) => h("div", { key: "tm" + k, class: "team", style: { "--c": TEAM[k] } },
    h("b", {}, C.t("trivia.team", { n: "AB"[k] })), num(g.totals[k]), h("small", {}, names(k)))));
}

export default {
  tv(S, C) {
    const g = S.g;
    return h("section", { class: "stage-in" }, hud(S, C), teams(S, C),
      h("h1", { class: "quote", key: "q" + g.round }, g.q), tiles(S, C),
      g.phase === "ask" && g.crowdN ? h("p", { class: "hint" }, C.t("crowd_votes", { n: g.crowdN })) : null,
      chips(S, g.done));
  },

  phone(S, C) {
    const g = S.g, me = S.me.g || { pick: -1 };
    const head = h("p", { class: "kicker", style: { color: TEAM[me.team] } }, S.me.aud ? C.t("audience.label") : C.t("trivia.team", { n: "AB"[me.team] }));
    if (g.phase === "ask") {
      return h("section", { class: "ctl" }, head, h("h2", { class: "big small" }, g.q), tiles(S, C, (i) => C.send({ k: "answer", i }), me),
        me.pick >= 0 ? waiting(C, C.t("trivia.locked")) : null);
    }
    const ok = me.pick === g.answer;
    return h("section", { class: "ctl center" }, head, h("h2", { class: `verdict ${ok ? "yes" : "no"}` }, me.pick < 0 ? C.t("trivia.noanswer") : ok ? C.t("trivia.right") : C.t("trivia.wrong")),
      tiles(S, C, null, me), gain(me.gain));
  },
};
