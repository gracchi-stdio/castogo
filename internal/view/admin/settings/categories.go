package settings

import (
	"github.com/gracchi-stdio/castogo/internal/domain"
	selectcomponent "github.com/gracchi-stdio/castogo/internal/view/components/select"
)

// CategoryOptions returns the static list of all podcast categories.
func CategoryOptions() []selectcomponent.SelectOptionArgs {
	options := make([]selectcomponent.SelectOptionArgs, 0, len(domain.CategoryNames()))
	for _, category := range domain.CategoryNames() {
		options = append(options, selectcomponent.SelectOptionArgs{
			Value: category,
			Label: category,
		})
	}
	return options
}

// SubcategoryOptions returns the subcategory list for a given category.
// Empty when the category has no subcategories.
func SubcategoryOptions(category string) []selectcomponent.SelectOptionArgs {
	subs := domain.SubcategoriesFor(category)
	options := make([]selectcomponent.SelectOptionArgs, 0, len(subs))
	for _, sub := range subs {
		options = append(options, selectcomponent.SelectOptionArgs{
			Value: sub,
			Label: sub,
		})
	}
	return options
}
