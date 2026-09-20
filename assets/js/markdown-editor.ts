// Markdown formatting toolbar over a bound textarea.
//
// Renders nothing in the DOM — the toolbar markup (buttons + icons) lives in the
// Templ `markdownToolbar` snippet so it stays inside the shared component library.
// This module only wires behaviour: each toolbar button carries a `data-md-action`,
// and the toolbar container carries `data-md-target` (the textarea's id). Clicking a
// button wraps or prefixes the textarea's current selection with markdown, then
// dispatches a synthetic `input` event so Datastar's `data-bind` picks up the
// programmatic change — it only syncs the signal on real input events, so without
// this the save would silently lose the edit.

const MOUNTED = "data-md-mounted";

type Action = (ta: HTMLTextAreaElement) => void;

/** Wrap the selection (or drop markers at the cursor) with before/after. */
function wrap(ta: HTMLTextAreaElement, before: string, after: string): void {
  const { selectionStart: s, selectionEnd: e, value } = ta;
  const sel = value.slice(s, e);
  ta.value = value.slice(0, s) + before + sel + after + value.slice(e);
  if (sel.length > 0) {
    ta.selectionStart = s + before.length;
    ta.selectionEnd = s + before.length + sel.length;
  } else {
    ta.selectionStart = ta.selectionEnd = s + before.length;
  }
  ta.focus();
  sync(ta);
}

/** Toggle a line prefix (e.g. "# ", "> ", "- ") on every line in the selection. */
function prefixLines(ta: HTMLTextAreaElement, prefix: string): void {
  const { selectionStart: s, selectionEnd: e, value } = ta;
  const lineStart = value.lastIndexOf("\n", s - 1) + 1;
  const nl = value.indexOf("\n", e);
  const lineEnd = nl === -1 ? value.length : nl;
  const lines = value.slice(lineStart, lineEnd).split("\n");
  // Toggle: if every selected line already has the prefix, strip it; else add it.
  const allHave = lines.every((l) => l.startsWith(prefix));
  const replaced = lines
    .map((l) => (allHave ? l.slice(prefix.length) : prefix + l))
    .join("\n");
  ta.value = value.slice(0, lineStart) + replaced + value.slice(lineEnd);
  ta.selectionStart = lineStart;
  ta.selectionEnd = lineStart + replaced.length;
  ta.focus();
  sync(ta);
}

/** Insert a markdown link, prompting for the URL. Any selection becomes the text. */
function link(ta: HTMLTextAreaElement): void {
  const { selectionStart: s, selectionEnd: e, value } = ta;
  const sel = value.slice(s, e);
  const url = window.prompt("Link URL", "https://");
  if (url === null) return; // cancelled
  const text = sel || url;
  const insert = `[${text}](${url})`;
  ta.value = value.slice(0, s) + insert + value.slice(e);
  ta.selectionStart = s;
  ta.selectionEnd = s + insert.length;
  ta.focus();
  sync(ta);
}

const ACTIONS: Record<string, Action> = {
  bold: (ta) => wrap(ta, "**", "**"),
  italic: (ta) => wrap(ta, "*", "*"),
  code: (ta) => wrap(ta, "`", "`"),
  h1: (ta) => prefixLines(ta, "# "),
  h2: (ta) => prefixLines(ta, "## "),
  quote: (ta) => prefixLines(ta, "> "),
  "bullet-list": (ta) => prefixLines(ta, "- "),
  "numbered-list": (ta) => prefixLines(ta, "1. "),
  link,
};

/** Push the textarea's value into its bound Datastar signal. */
function sync(ta: HTMLTextAreaElement): void {
  ta.dispatchEvent(new Event("input", { bubbles: true }));
}

/** Wire every unmounted toolbar. Idempotent — safe to call repeatedly. */
export function initMarkdownEditors(root: ParentNode = document): void {
  for (const bar of root.querySelectorAll<HTMLElement>("[data-md-toolbar]")) {
    if (bar.hasAttribute(MOUNTED)) continue;
    bar.setAttribute(MOUNTED, "");
    bar.addEventListener("click", (ev: MouseEvent) => {
      const btn = (ev.target as HTMLElement).closest<HTMLElement>(
        "[data-md-action]",
      );
      if (!btn) return;
      ev.preventDefault();
      const action = btn.getAttribute("data-md-action");
      const targetId = bar.getAttribute("data-md-target");
      if (!action || !targetId) return;
      const ta = document.getElementById(targetId) as HTMLTextAreaElement | null;
      if (!ta) return;
      ACTIONS[action]?.(ta);
    });
  }
}
