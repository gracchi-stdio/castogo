package formactions

type FormActionsArgs struct {
	FormID       string // required — submit button binds to this via form="..."
	SubmitLabel  string // defaults to "Save"
	SubmittingLabel string // defaults to "Saving..." (shown while $fetching)
	CancelHref   string // optional — when set, renders a Cancel link
	CancelLabel  string // defaults to "Cancel"
}
