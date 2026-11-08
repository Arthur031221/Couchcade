// Shared shell for the TV and phone pages: language, theme, sound, render loop.
import { h, patch } from "./dom.js";
import * as i18n from "./i18n.js";
import { isMuted, setMuted } from "./sound.js";

// The hosted demo shows several pages of one origin side by side, so each frame
// names its own slot to keep tokens and names apart.
const NS = (() => { try { return (window.frameElement && window.frameElement.dataset.ns) || ""; } catch { return ""; } })();
export const store = {
  get: (k, d) => { try { return localStorage.getItem(`cc.${NS}${k}`) ?? d; } catch { return d; } },
  set: (k, v) => { try { localStorage.setItem(`cc.${NS}${k}`, v); } catch { /* private mode */ } },
};

export function applyTheme(mode) {
  const theme = mode || store.get("theme") || (matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark");
  document.documentElement.dataset.theme = theme;
  return theme;
}

export function toggleTheme() {
  const next = document.documentElement.dataset.theme === "light" ? "dark" : "light";
  store.set("theme", next);
  applyTheme(next);
}

export function toggleSound() {
  setMuted(!isMuted());
  store.set("muted", isMuted() ? "1" : "0");
}

export function initShell() {
  setMuted(store.get("muted") === "1");
  applyTheme();
}

// mount renders view(state) into root and batches repeated calls into one frame.
export function mount(root, view) {
  let queued = false;
  const draw = () => { queued = false; patch(root, view()); };
  return () => { if (!queued) { queued = true; requestAnimationFrame(draw); } };
}

export function langSelect(current, langs, onchange) {
  return h("select", { class: "sel", "aria-label": "Language", value: current, onchange: (e) => onchange(e.target.value) },
    i18n.LANGS.filter((l) => !langs || langs.includes(l[0]) || true).map((l) => h("option", { key: l[0], value: l[0], selected: l[0] === current }, l[1])));
}

export { i18n };
