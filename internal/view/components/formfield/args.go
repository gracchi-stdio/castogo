package formfield

import "github.com/a-h/templ"

type FormFieldArgs struct {
	Name        string           // field name — also used to derive error signal name (`<Name>_error`)
	Label       string           // label text
	Required    bool             // show required indicator
	Description string           // optional help text under the control
	ErrorSignal string           // signal name for reactive error (e.g. "title_error"); empty disables error display
	Class       string           // additional classes on the FormItem wrapper
	Attributes  templ.Attributes // additional HTML attributes on the FormItem wrapper
}
