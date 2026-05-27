package imageupload

import "github.com/gracchi-stdio/castogo/internal/view/utils"

func imageUploadDropZone(class string) string {
	return utils.TwMerge(
		"border-2 border-dashed border-muted-foreground/25 p-8 text-center cursor-pointer hover:border-muted-foreground/50 transition-colors",
		class,
	)
}

func imageUploadPreview(w, h string) string {
	if w == "" {
		w = "w-48"
	}
	if h == "" {
		h = "h-48"
	}
	return w + " " + h + " object-cover border"
}
