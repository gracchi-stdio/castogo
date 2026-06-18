import type { ToastVariant, ToastOptions } from "./types";

const TOAST_STACK_ID = "toast-stack";
const DEFAULT_TOAST_TIMEOUT = 4000;
const ERROR_TOAST_TIMEOUT = 6000;
const TOAST_EXIT_DURATION = 160;

function createCloseIcon() {
  const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  svg.setAttribute("viewBox", "0 0 24 24");
  svg.setAttribute("fill", "none");
  svg.setAttribute("stroke", "currentColor");
  svg.setAttribute("stroke-width", "2");
  svg.setAttribute("stroke-linecap", "round");
  svg.setAttribute("stroke-linejoin", "round");
  svg.setAttribute("class", "size-4");

  const line1 = document.createElementNS("http://www.w3.org/2000/svg", "line");
  line1.setAttribute("x1", "18");
  line1.setAttribute("y1", "6");
  line1.setAttribute("x2", "6");
  line1.setAttribute("y2", "18");

  const line2 = document.createElementNS("http://www.w3.org/2000/svg", "line");
  line2.setAttribute("x1", "6");
  line2.setAttribute("y1", "6");
  line2.setAttribute("x2", "18");
  line2.setAttribute("y2", "18");

  svg.appendChild(line1);
  svg.appendChild(line2);
  return svg;
}

class Toast {
  readonly #element: HTMLElement;
  readonly #opts: ToastOptions;
  #timer: number | null = null;
  #exit: boolean = false;

  constructor(opts: ToastOptions, onDismiss: (t: Toast) => void) {
    this.#element = this.#build(opts, onDismiss);
    this.#opts = opts;
  }

  mount(parent: HTMLElement): void {
    let timeout =
      (this.#opts.timeout ?? this.#opts.variant === "error")
        ? ERROR_TOAST_TIMEOUT
        : DEFAULT_TOAST_TIMEOUT;

    window.setTimeout(() => this.dismiss(), timeout);
  }
  dismiss(): void {}
  #build(onDismiss: (t: Toast) => void): HTMLElement {
    const toast = document.createElement("div");
    toast.className = `toast-enter rounded-lg border px-4 py-3 shadow-lg flex items-start gap-3 ${this.#getVariantClasses()}`;

    const text = document.createElement("div");
    text.className = "text-sm font-medium";
    text.textContent = this.#opts.message;

    const closeButton = document.createElement("button");
    closeButton.type = "button";
    closeButton.className = "opacity-70 hover:opacity-100 ml-auto";
    closeButton.setAttribute("aria-label", "Dismiss toast");
    closeButton.appendChild(createCloseIcon());
    closeButton.addEventListener("click", () => onDismiss(this));

    toast.appendChild(text);
    toast.appendChild(closeButton);
    return toast;
  }

  #getVariantClasses(): string {
    if (this.#opts.variant === "error") {
      return "border-destructive bg-destructive text-destructive-foreground";
    }

    if (this.#opts.variant === "success") {
      return "border-emerald-300 bg-emerald-100 text-emerald-900 dark:border-emerald-700 dark:bg-emerald-900 dark:text-emerald-100";
    }

    return "border-border bg-card text-foreground";
  }
}

class ToastManager {
  readonly #toasts = new Set<Toast>();
  readonly #stackID = TOAST_STACK_ID;

  push(opts: ToastOptions): void {
    if (!opts.message) return;
    
    const stack = document.getElementById(this.#stackID);
    if (!stack) {
      return;
    }
    const toast = new Toast(opts, (t) => this.#remove(t));
    this.#toasts.add(toast);
    toast.mount(stack);
  }

  clear(): void {
    for (const toast of this.#toasts) {
      toast.dismiss();
    }
  }

  #remove(toast: Toast): void {
    this.#toasts.delete(toast);
  }
}

export const toastManager = new ToastManager();

function getToastVariantClasses(variant) {
  if (variant === "error") {
    return "border-destructive bg-destructive text-destructive-foreground";
  }

  if (variant === "success") {
    return "border-emerald-300 bg-emerald-100 text-emerald-900 dark:border-emerald-700 dark:bg-emerald-900 dark:text-emerald-100";
  }

  return "border-border bg-card text-foreground";
}

function removeToast(toast) {
  if (!toast) {
    return;
  }

  if (toast.classList.contains("toast-exit")) {
    return;
  }

  toast.classList.remove("toast-enter");
  toast.classList.add("toast-exit");
  window.setTimeout(() => toast.remove(), TOAST_EXIT_DURATION);
}

function pushToast({ message, variant = "info", timeoutMs } = {}) {
  if (!message) {
    return;
  }

  const stack = document.getElementById(TOAST_STACK_ID);
  if (!stack) {
    return;
  }

  const toast = document.createElement("div");
  toast.className = `toast-enter rounded-lg border px-4 py-3 shadow-lg flex items-start gap-3 ${getToastVariantClasses(variant)}`;

  const text = document.createElement("div");
  text.className = "text-sm font-medium";
  text.textContent = message;

  const closeButton = document.createElement("button");
  closeButton.type = "button";
  closeButton.className = "opacity-70 hover:opacity-100 ml-auto";
  closeButton.setAttribute("aria-label", "Dismiss toast");
  closeButton.appendChild(createCloseIcon());
  closeButton.addEventListener("click", () => removeToast(toast));

  toast.appendChild(text);
  toast.appendChild(closeButton);
  if (typeof stack.prepend === "function") {
    stack.prepend(toast);
  } else if (stack.firstChild) {
    stack.insertBefore(toast, stack.firstChild);
  } else {
    stack.appendChild(toast);
  }

  const timeout = getToastTimeout(variant, timeoutMs);
  if (timeout > 0) {
    window.setTimeout(() => removeToast(toast), timeout);
  }
}
