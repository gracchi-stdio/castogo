import WaveSurfer from "wavesurfer.js";

/** @type {WeakMap<HTMLElement, import('wavesurfer.js').default>} */
const instances = new WeakMap();

let active = null; // { ws, playBtn, timeEl, container }

const PLAY_SVG = `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="6 3 20 12 6 21 6 3"/></svg>`;
const PAUSE_SVG = `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="14" y="4" width="4" height="16" rx="0"/><rect x="6" y="4" width="4" height="16" rx="0"/></svg>`;

function formatTime(seconds) {
  const m = Math.floor(seconds / 60);
  const s = Math.floor(seconds % 60);
  return `${m}:${s.toString().padStart(2, "0")}`;
}

function readThemeColors() {
  const el = document.querySelector(".theme-public") || document.documentElement;
  const style = getComputedStyle(el);
  const fg = style.getPropertyValue("--muted-foreground").trim();
  const primary = style.getPropertyValue("--primary").trim();
  return {
    waveColor: fg ? `hsl(${fg} / 0.25)` : "rgba(128,128,128,0.25)",
    progressColor: primary ? `hsl(${primary})` : "hsl(0 82% 58%)",
  };
}

function setIcon(btn, playing) {
  btn.innerHTML = playing ? PAUSE_SVG : PLAY_SVG;
  btn.setAttribute("aria-label", playing ? "Pause episode" : "Play episode");
}

function pauseActive() {
  if (!active) return;
  active.ws.pause();
  setIcon(active.playBtn, false);
  active = null;
}

function handleClick(root) {
  const playBtn = root.querySelector("[data-episode-play]");
  const waveformEl = root.querySelector("[data-episode-waveform]");
  const timeEl = root.querySelector("[data-episode-time]");
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

    ws.on("ready", () => {
      root.classList.remove("episode-player--loading");
      ws.play();
      active = { ws, playBtn, timeEl, container: waveformEl };
      setIcon(playBtn, true);
    });

    ws.on("timeupdate", (currentTime) => {
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

export function initEpisodePlayers() {
  document.querySelectorAll("[data-episode-player]").forEach(handleClick);
}

export function destroyActivePlayer() {
  if (!active) return;
  active.ws.pause();
  setIcon(active.playBtn, false);
  active = null;
}
