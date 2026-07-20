package layout

import "github.com/gracchi-stdio/castogo/internal/view/components/sidebar"

func sidebarSections(currentPath string) []sidebar.SidebarSection {
	return []sidebar.SidebarSection{
		{
			Title: "General",
			Items: []sidebar.SidebarItem{
				{Title: "Dashboard", Href: "/admin", Icon: "layout-grid"},
			},
		},
		{
			Title: "Podcast",
			Items: []sidebar.SidebarItem{
				{Title: "All Episodes", Href: "/admin/episodes", Icon: "list"},
				{Title: "New Episode", Href: "/admin/episodes/create", Icon: "circle-plus"},
			},
		},
		{
			Title: "Site",
			Items: []sidebar.SidebarItem{
				{Title: "Pages", Href: "/admin/pages", Icon: "file-text"},
			},
		},
		{
			Title: "Docs",
			Items: []sidebar.SidebarItem{
				{Title: "Developer", Href: "/admin/docs/developer", Icon: "code"},
				{Title: "User", Href: "/admin/docs/user", Icon: "book-open"},
			},
		},
		{
			Title: "System",
			Items: []sidebar.SidebarItem{
				{Title: "Settings", Href: "/admin/settings", Icon: "settings"},
			},
		},
	}
}
