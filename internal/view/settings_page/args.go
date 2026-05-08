package settings_page

import (
	"github.com/gracchi-stdio/castogo/internal/domain"
	selectcomponent "github.com/gracchi-stdio/castogo/internal/view/components/select"
)

var categoryOptions = make([]selectcomponent.SelectOptionArgs, 0)

func init() {
	for _, category := range domain.CategoryNames() {
		categoryOptions = append(categoryOptions, selectcomponent.SelectOptionArgs{
			Value: category,
			Label: category,
		})
	}

}

func getSubcategoryOptions(category string) []selectcomponent.SelectOptionArgs {
	subcategories := domain.SubcategoriesFor(category)
	options := make([]selectcomponent.SelectOptionArgs, 0)
	for _, sub := range subcategories {
		options = append(options, selectcomponent.SelectOptionArgs{
			Value: sub,
			Label: sub,
		})
	}
	return options
}
