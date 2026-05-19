package loading

import "github.com/a-h/templ"

type LoadingOverlayArgs struct {
	Signal     string           // signal name controlling visibility (e.g. "uploading")
	Status     string           // signal name for status text (e.g. "uploading_status")
	Class      string           // additional CSS classes
	Attributes templ.Attributes // additional HTML attributes
}
