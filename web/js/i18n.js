// Locale files live in /locales/<lang>.json. Missing keys fall back to English.
export const LANGS = [
  ["en", "English"], ["zh-TW", "繁體中文"], ["zh-CN", "简体中文"], ["ja", "日本語"], ["ko", "한국어"],
  ["es", "Español"], ["fr", "Français"], ["de", "Deutsch"], ["pt-BR", "Português (Brasil)"],
];
const CODES = LANGS.map((l) => l[0]);
let dict = {};
let fallback = {};
let current = "en";

export function detect(saved) {
  if (saved && CODES.includes(saved)) return saved;
  for (const raw of navigator.languages || [navigator.language || "en"]) {
    const l = raw.toLowerCase();
    if (l === "zh-tw" || l === "zh-hk" || l === "zh-hant" || l.startsWith("zh-hant")) return "zh-TW";
    if (l.startsWith("zh")) return "zh-CN";
    if (l.startsWith("pt")) return "pt-BR";
    const hit = CODES.find((c) => c.toLowerCase() === l.split("-")[0]);
    if (hit) return hit;
  }
  return "en";
}

async function load(lang) {
  const r = await fetch(`locales/${lang}.json`);
  return r.ok ? r.json() : {};
}

export async function setLang(lang) {
  if (!Object.keys(fallback).length) fallback = await load("en");
  dict = lang === "en" ? fallback : await load(lang);
  current = lang;
  document.documentElement.lang = lang;
}

export const lang = () => current;

export function t(key, vars) {
  let s = dict[key] ?? fallback[key] ?? key;
  if (vars) for (const k in vars) s = s.replaceAll(`{${k}}`, vars[k]);
  return s;
}
