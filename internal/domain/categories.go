package domain

import "sort"

// iTunes categories mapped to their subcategories.
// Categories with no subcategories have an empty slice (e.g. True Crime).
// This is the single source of truth — the settings page, SSE endpoint,
// and future RSS feed renderer all read from here.
var ITunesCategories = map[string][]string{
	"Arts":                    {"Books", "Design", "Fashion & Beauty", "Food", "Performing Arts", "Visual Arts"},
	"Business":                {"Careers", "Entrepreneurship", "Investing", "Management", "Marketing", "Non-Profit"},
	"Comedy":                  {"Comedy Interviews", "Improv", "Stand-Up"},
	"Education":               {"Courses", "How To", "Language Learning", "Self-Improvement"},
	"Fiction":                 {"Comedy Fiction", "Drama", "Science Fiction"},
	"Government":              {},
	"History":                 {},
	"Health & Fitness":        {"Alternative Health", "Fitness", "Medicine", "Mental Health", "Nutrition", "Sexuality"},
	"Kids & Family":           {"Education for Kids", "Family", "Parenting", "Pets & Animals", "Stories for Kids"},
	"Leisure":                 {"Animation & Manga", "Automotive", "Aviation", "Crafts", "Games", "Hobbies", "Home & Garden", "Video Games"},
	"Music":                   {"Music Commentary", "Music History", "Music Interviews"},
	"News":                    {"Business News", "Daily News", "Entertainment News", "News Commentary", "Politics", "Sports News", "Tech News"},
	"Religion & Spirituality": {"Buddhism", "Christianity", "Hinduism", "Islam", "Judaism", "Religion", "Spirituality"},
	"Science":                 {"Astronomy", "Chemistry", "Earth Sciences", "Life Sciences", "Mathematics", "Natural Sciences", "Nature", "Physics", "Social Sciences"},
	"Society & Culture":       {"Documentary", "Education", "History", "Personal Journals", "Philosophy", "Places & Travel", "Relationships"},
	"Sports":                  {"Baseball", "Basketball", "Cricket", "Fantasy Sports", "Football", "Golf", "Hockey", "Rugby", "Running", "Soccer", "Swimming", "Tennis", "Volleyball", "Wrestling"},
	"Technology":              {},
	"True Crime":              {},
	"TV & Film":               {"After Shows", "Film History", "Film Interviews", "Film Reviews", "TV Reviews"},
}

// CategoryNames returns all top-level iTunes category names in alphabetical order.
// Used to populate the category select dropdown.
func CategoryNames() []string {
	names := make([]string, 0, len(ITunesCategories))
	for name := range ITunesCategories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// SubcategoriesFor returns the subcategories for a given category.
// Returns nil if the category doesn't exist.
// Returns an empty (non-nil) slice if the category exists but has no subcategories
// (e.g. True Crime, Technology).
func SubcategoriesFor(category string) []string {
	subs, ok := ITunesCategories[category]
	if !ok {
		return nil
	}
	// Return a copy to prevent callers from mutating the original.
	result := make([]string, len(subs))
	copy(result, subs)
	return result
}

// HasSubcategories reports whether a category has any subcategories.
// Returns false for unknown categories or categories with empty subcategory lists.
func HasSubcategories(category string) bool {
	return len(ITunesCategories[category]) > 0
}

// IsValidCategory reports whether the category exists and the subcategory is
// valid for that category.
func IsValidCategory(category, subcategory string) bool {
	subcategories, ok := ITunesCategories[category]
	if !ok {
		return false
	}
	if subcategory == "" {
		return len(subcategories) == 0
	}
	for _, candidate := range subcategories {
		if candidate == subcategory {
			return true
		}
	}
	return false
}
