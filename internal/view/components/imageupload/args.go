package imageupload

import "github.com/a-h/templ"

type ImageUploadArgs struct {
	SignalName string // Datastar signal for the image URL (e.g. "block_42_background_image")
	UploadURL  string // SSE POST endpoint (e.g. "/admin/pages/blocks/upload-image")
	FormID     string // Parent form element ID for multipart submission
	Accept     string // File types (default "image/jpeg,image/png,image/webp")
	Label      string // Label text (e.g. "Background Image")
	PreviewW   string // Preview width class (default "w-48")
	PreviewH   string // Preview height class (default "h-48")
	ID         string // Unique ID for the file input element
	Hint       string // Help text below upload area
	Class      string // Additional wrapper classes
	Attributes templ.Attributes
}
