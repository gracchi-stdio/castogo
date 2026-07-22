package audio

import "github.com/a-h/templ"

// AudioArgs configures a native HTML5 audio player. The native controls are used
// deliberately: unlike the WaveSurfer-based public player, this is a static
// element that re-mounts cleanly when its container is SSE outer-swapped (e.g.
// the episode audio card after a replace), with no JS instance lifecycle to
// manage.
type AudioArgs struct {
	Src       string            // Audio source URL
	MimeType  string            // Optional MIME type for <source type=...> (e.g. "audio/mpeg")
	ID        string            // Optional element ID
	Preload   string            // "auto" (default) | "metadata" | "none"
	Class     string            // Additional classes on the <audio> element
	Attributes templ.Attributes // Extra HTML attributes
}
