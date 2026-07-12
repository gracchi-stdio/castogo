import WaveSurfer from "wavesurfer.js";

interface ActivePlayer {
  ws: WaveSurfer;
  playBtn: HTMLButtonElement;
  timeEl: HTMLElement | null;
  container: HTMLElement;
}

interface ThemeColors {
  waveColor: string;
  progressColor: string;
}

const instances = new WeakMap<HTMLElement, WaveSurfer>();

let active: ActivePlayer | null = null;

const PLAY_SVG =
  '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="6 3 20 12 6 21 6 3"/></svg>';
const PAUSE_SVG =
  '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="14" y="4" width="4" height="16" rx="0"/><rect x="6" y="4" width="4" height="16" rx="0"/></svg>';

function formatTime(seconds: number): string {
  const m = Math.floor(seconds / 60);
  const s = Math.floor(seconds % 60);
  return `${m}:${s.toString().padStart(2, "0")}`;
}

function formatClock(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = Math.floor(seconds % 60);
  return `${h}:${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
}

function readThemeColors(): ThemeColors {
  const el =
    document.querySelector(".theme-public") ?? document.documentElement;
  const style = getComputedStyle(el);
  const fg = style.getPropertyValue("--muted").trim();
  const foreground = style.getPropertyValue("--muted-foreground").trim();
  return {
    waveColor: fg ? `hsl(${fg})` : "rgba(128,128,128,0.25)",
    progressColor: foreground ? `hsl(${foreground})` : "hsl(0 82% 58%)",
  };
}

function setIcon(btn: HTMLButtonElement, playing: boolean): void {
  btn.innerHTML = playing ? PAUSE_SVG : PLAY_SVG;
  btn.setAttribute("aria-label", playing ? "Pause episode" : "Play episode");
}

function pauseActive(): void {
  if (!active) return;
  active.ws.pause();
  setIcon(active.playBtn, false);
  active = null;
}

function handleClick(root: HTMLElement): void {
  const playBtn = root.querySelector<HTMLButtonElement>("[data-episode-play]");
  const waveformEl = root.querySelector<HTMLElement>("[data-episode-waveform]");
  const timeEl = root.querySelector<HTMLElement>("[data-episode-time]");
  const audioUrl = root.dataset.audioSrc;
  if (!playBtn || !waveformEl || !audioUrl) return;

  playBtn.addEventListener("click", (e) => {
    e.stopPropagation();

    // Toggle if same episode
    if (active && active.ws === instances.get(waveformEl)) {
      active.ws.playPause();
      setIcon(active.playBtn, active.ws.isPlaying());
      return;
    }

    // Pause whatever is playing
    pauseActive();

    // Already initialized — just play
    const existing = instances.get(waveformEl);
    if (existing) {
      existing.play();
      active = { ws: existing, playBtn, timeEl, container: waveformEl };
      setIcon(playBtn, true);
      return;
    }

    // First-time lazy init
    root.classList.add("episode-player--loading");
    setIcon(playBtn, false);
    const colors = readThemeColors();

    const ws = WaveSurfer.create({
      container: waveformEl,
      url: audioUrl,
      waveColor: colors.waveColor,
      progressColor: colors.progressColor,
      cursorColor: "transparent",
      barWidth: 2,
      barGap: 1,
      barRadius: 0,
      height: 20,
      normalize: true,
      hideScrollbar: true,
      interact: true,
    });

    instances.set(waveformEl, ws);

    // Hover: primary fill from start → cursor, plus an HH:MM:SS tooltip.
    // CSS toggles visibility via :hover; JS only updates geometry + label.
    const hoverFill = document.createElement("div");
    hoverFill.className = "episode-hover-fill";
    const hoverTip = document.createElement("div");
    hoverTip.className = "episode-hover-tip";
    waveformEl.append(hoverFill, hoverTip);

    waveformEl.addEventListener("mousemove", (e) => {
      const rect = waveformEl.getBoundingClientRect();
      const ratio = Math.min(
        1,
        Math.max(0, (e.clientX - rect.left) / rect.width),
      );
      hoverFill.style.width = `${ratio * 100}%`;
      hoverTip.style.left = `${ratio * 100}%`;
      const dur = ws.getDuration();
      hoverTip.textContent = formatClock((dur || 0) * ratio);
    });

    ws.on("ready", () => {
      root.classList.remove("episode-player--loading");
      ws.play();
      active = { ws, playBtn, timeEl, container: waveformEl };
      setIcon(playBtn, true);
    });

    ws.on("timeupdate", (currentTime: number) => {
      if (timeEl) timeEl.textContent = formatTime(currentTime);
    });

    ws.on("finish", () => {
      setIcon(playBtn, false);
      if (timeEl) timeEl.textContent = formatTime(ws.getDuration());
      active = null;
    });

    ws.on("error", () => {
      root.classList.remove("episode-player--loading");
      waveformEl.classList.add("episode-player--error");
      setIcon(playBtn, false);
    });
  });
}

export function initEpisodePlayers(): void {
  document
    .querySelectorAll<HTMLElement>("[data-episode-player]")
    .forEach(handleClick);
}

export function destroyActivePlayer(): void {
  if (!active) return;
  active.ws.pause();
  setIcon(active.playBtn, false);
  active = null;
}
