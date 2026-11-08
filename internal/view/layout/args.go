package layout

import "github.com/gracchi-stdio/castogo/internal/domain"

type SettingsArgs struct {
	Title string
}

type AdminSharedData struct {
	User        *domain.User
	CurrentPath string
}
