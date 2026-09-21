// Audio-synced slide deck ([data-slideshow], rendered by
// pageview/blocks/slideshow.templ) — one slide visible at a time.
//
// Each slideshow names its audio element via data-slideshow-audio. Playback
// drives the deck: crossing a slide's time cue (data-slide-time, seconds)
// makes it the visible slide. Navigation drives the audio: the arrows, a slide
// click, or scrolling to a slide seek playback to that slide's cue (its own,
// or the latest cue at or before it) and hold it there until playback reaches
// the next cue. Scrubbing the audio directly re-enables follow mode.

const MOUNTED = "data-ss-mounted";

/** "45" | "03:12" | "1:02:03" → seconds; right-aligned clock parts. */
function parseSlideTime(v: string | undefined): number | null {
  if (!v) return null;
  const parts = v.split(":");
  if (parts.length === 0 || parts.length > 3) return null;
  let total = 0;
  for (const p of parts) {
    const n = Number.parseInt(p.trim(), 10);
    if (Number.isNaN(n) || n < 0) return null;
    total = total * 60 + n;
  }
  return total;
}

function wire(root: HTMLElement): void {
  const trackEl = root.querySelector<HTMLElement>("[data-slideshow-track]");
  const slides = Array.from(root.querySelectorAll<HTMLElement>("[data-slide]"));
  if (!trackEl || slides.length === 0) return;
  // TS won't carry the null-narrowing into the closures below — bind it.
  const track: HTMLElement = trackEl;

  const audioID = root.getAttribute("data-slideshow-audio");
  const audio = audioID ? (document.getElementById(audioID) as HTMLAudioElement | null) : null;

  if (slides.length < 2) {
    root.querySelector("[data-slideshow-prev]")?.setAttribute("hidden", "");
    root.querySelector("[data-slideshow-next]")?.setAttribute("hidden", "");
  }

  // Per-slide cue: the slide's own time, else the latest cue at or before it.
  const cues: (number | null)[] = [];
  let lastCue: number | null = null;
  for (const s of slides) {
    const t = parseSlideTime(s.dataset.slideTime);
    if (t !== null) lastCue = t;
    cues.push(t ?? lastCue);
  }

  let current = 0;
  let selfSeek = false; // the in-flight audio seek was ours
  let manualHold = false; // the user picked a slide; hold it until the next cue
  let holdUntil = Infinity;
  let programmaticScroll = false;

  function paint(i: number): void {
    slides.forEach((s, idx) => {
      if (idx === i) {
        s.setAttribute("data-active", "");
        s.setAttribute("aria-current", "true");
      } else {
        s.removeAttribute("data-active");
        s.removeAttribute("aria-current");
      }
    });
  }

  function show(i: number): void {
    programmaticScroll = true;
    const el = slides[i];
    track.scrollTo({
      left: el.offsetLeft - (track.clientWidth - el.clientWidth) / 2,
      behavior: "smooth",
    });
  }

  function nextCueAfter(i: number): number {
    for (let j = i + 1; j < slides.length; j++) {
      const t = parseSlideTime(slides[j].dataset.slideTime);
      if (t !== null) return t;
    }
    return Infinity;
  }

  function goTo(i: number, fromAudio: boolean): void {
    const clamped = Math.max(0, Math.min(slides.length - 1, i));
    const changed = clamped !== current;
    current = clamped;
    paint(clamped);
    if (changed) show(clamped);
    if (fromAudio || !audio) return;

    // User navigation: seek to this slide's cue and hold here until playback
    // reaches the cue of the next timed slide after it.
    manualHold = true;
    holdUntil = nextCueAfter(clamped);
    const cue = cues[clamped];
    if (cue !== null) {
      selfSeek = true;
      audio.currentTime = cue;
    }
  }

  root.querySelector("[data-slideshow-prev]")?.addEventListener("click", () => goTo(current - 1, false));
  root.querySelector("[data-slideshow-next]")?.addEventListener("click", () => goTo(current + 1, false));
  slides.forEach((s, i) => s.addEventListener("click", () => goTo(i, false)));

  // A settled scroll we did not cause is the user swiping to a slide.
  const onScrollSettled = (): void => {
    if (programmaticScroll) {
      programmaticScroll = false;
      return;
    }
    const center = track.scrollLeft + track.clientWidth / 2;
    let best = 0;
    let bestDist = Infinity;
    slides.forEach((s, i) => {
      const dist = Math.abs(s.offsetLeft + s.clientWidth / 2 - center);
      if (dist < bestDist) {
        bestDist = dist;
        best = i;
      }
    });
    goTo(best, false);
  };
  // Feature-detect on window — `"onscrollend" in track` narrows track's DOM
  // type to never in the else branch (lib.dom declares it on Element).
  if ("onscrollend" in window) {
    track.addEventListener("scrollend", onScrollSettled);
  } else {
    let timer: number | undefined;
    track.addEventListener("scroll", () => {
      window.clearTimeout(timer);
      timer = window.setTimeout(onScrollSettled, 150);
    });
  }

  if (!audio) return;

  // A seek we did not cause (the user scrubbing) ends any manual hold; an
  // ended track also releases it so a replay follows the cues again.
  audio.addEventListener("seeking", () => {
    if (!selfSeek) manualHold = false;
  });
  audio.addEventListener("seeked", () => {
    selfSeek = false;
  });
  audio.addEventListener("ended", () => {
    manualHold = false;
  });

  audio.addEventListener("timeupdate", () => {
    if (manualHold && audio.currentTime < holdUntil) return;
    manualHold = false;
    // The latest slide whose own cue is at or before playback.
    let j = -1;
    for (let i = 0; i < slides.length; i++) {
      const t = parseSlideTime(slides[i].dataset.slideTime);
      if (t !== null && t <= audio.currentTime) j = i;
    }
    if (j >= 0) goTo(j, true);
  });

  paint(current);
}

/** Wire every unmounted slideshow. Idempotent — safe to call repeatedly. */
export function initEpisodeSlideshows(root: ParentNode = document): void {
  for (const el of root.querySelectorAll<HTMLElement>("[data-slideshow]")) {
    if (el.hasAttribute(MOUNTED)) continue;
    el.setAttribute(MOUNTED, "");
    wire(el);
  }
}
