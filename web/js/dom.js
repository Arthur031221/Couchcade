// A small keyed virtual DOM: views are plain functions that return h() trees,
// patch() updates the page in place so animations and inputs survive state pushes.
const SVG_TAGS = new Set(["svg", "path", "g", "circle", "rect", "line", "text", "polyline"]);
const SVG_NS = "http://www.w3.org/2000/svg";

export function h(tag, props, ...kids) {
  props = props || {};
  const out = [];
  const add = (k) => {
    if (Array.isArray(k)) k.forEach(add);
    else if (k !== null && k !== undefined && k !== false && k !== true) out.push(k);
  };
  kids.forEach(add);
  return { tag, props, kids: out, key: props.key };
}

function setProp(el, name, value, old) {
  if (name === "key" || name === "ref" || name === "upd") return;
  if (name.startsWith("on") && typeof (value ?? old) === "function") {
    const ev = name.slice(2).toLowerCase();
    el._on = el._on || {};
    if (el._on[ev]) el.removeEventListener(ev, el._on[ev]);
    if (value) {
      el._on[ev] = value;
      el.addEventListener(ev, value, { passive: ev.startsWith("touch") ? true : undefined });
    }
    return;
  }
  if (name === "class") {
    el.setAttribute("class", value || "");
  } else if (name === "style" && typeof value === "object") {
    el.removeAttribute("style");
    for (const k in value) el.style.setProperty(k, value[k]);
  } else if (name === "value" || name === "checked" || name === "disabled" || name === "selected") {
    if (el[name] !== value) el[name] = value;
  } else if (value === false || value === null || value === undefined) {
    el.removeAttribute(name);
  } else {
    el.setAttribute(name, value === true ? "" : value);
  }
}

function create(v, svg) {
  if (typeof v !== "object") return document.createTextNode(String(v));
  svg = svg || v.tag === "svg";
  const el = svg || SVG_TAGS.has(v.tag) ? document.createElementNS(SVG_NS, v.tag) : document.createElement(v.tag);
  for (const k in v.props) setProp(el, k, v.props[k]);
  v.kids.forEach((k) => el.appendChild(create(k, svg)));
  el._v = v;
  if (v.props.ref) v.props.ref(el);
  if (v.props.upd) v.props.upd(el);
  return el;
}

function same(el, v) {
  if (typeof v !== "object") return el.nodeType === 3;
  return el.nodeType === 1 && el._v && el._v.tag === v.tag;
}

function patchEl(el, v, svg) {
  if (typeof v !== "object") {
    if (el.data !== String(v)) el.data = String(v);
    return;
  }
  svg = svg || v.tag === "svg";
  const old = el._v.props;
  for (const k in v.props) if (v.props[k] !== old[k] || k === "value") setProp(el, k, v.props[k], old[k]);
  for (const k in old) if (!(k in v.props)) setProp(el, k, null, old[k]);
  patchKids(el, v.kids, svg);
  el._v = v;
  if (v.props.upd) v.props.upd(el);
}

function patchKids(parent, kids, svg) {
  const old = Array.from(parent.childNodes);
  const byKey = new Map();
  const unkeyed = [];
  old.forEach((n) => {
    const k = n._v && n._v.key;
    if (k !== undefined && k !== null) byKey.set(k, n);
    else unkeyed.push(n);
  });
  let u = 0;
  const used = new Set();
  const next = kids.map((k) => {
    let el;
    const key = typeof k === "object" ? k.key : undefined;
    if (key !== undefined && key !== null) el = byKey.get(key);
    else el = unkeyed[u++];
    if (el && !used.has(el) && same(el, k)) {
      patchEl(el, k, svg);
    } else {
      el = create(k, svg);
    }
    used.add(el);
    return el;
  });
  old.forEach((n) => {
    if (!used.has(n)) parent.removeChild(n);
  });
  next.forEach((n, i) => {
    if (parent.childNodes[i] !== n) parent.insertBefore(n, parent.childNodes[i] || null);
  });
}

export function patch(root, vnode) {
  patchKids(root, vnode === null ? [] : [vnode], false);
}
