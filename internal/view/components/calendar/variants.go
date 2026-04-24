package calendar

import "github.com/gracchi-stdio/castogo/internal/view/utils"

// calendarVariants generates CSS classes for the calendar container
func calendarVariants(className string) string {
	base := "p-3"

	return utils.TwMerge(base, className)
}
