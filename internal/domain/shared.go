package domain

type AdminSharedData struct {
	User        *User
	CurrentPath string
}

type DashboardStats struct {
	Total     int
	Published int
	Drafts    int
	Scheduled int
}
