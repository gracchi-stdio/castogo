package selectcomponent

import (
	"fmt"

	"github.com/gracchi-stdio/castogo/internal/view/utils"
)

// HighlightedItem creates a data-class expression for highlighted select/list items
func highlightedItem(signalPath string, index int) string {
	return utils.NewDataClass().
		Add("bg-accent", fmt.Sprintf("$%s === %d", signalPath, index)).
		Add("text-accent-foreground", fmt.Sprintf("$%s === %d", signalPath, index)).
		Build()
}

// SelectedItem creates a data-class expression for selected items
func selectedItem(signalPath, value string) string {
	return utils.NewDataClass().
		Add("bg-primary", fmt.Sprintf("$%s === '%s'", signalPath, value)).
		Add("text-primary-foreground", fmt.Sprintf("$%s === '%s'", signalPath, value)).
		Build()
}

// SelectTriggerHandler creates handlers for select trigger functionality
type SelectTriggerHandler struct {
	selectID string
	signals  *utils.SignalManager
}

// NewSelectTriggerHandler creates a select trigger handler
func NewSelectTriggerHandler(selectID string, signals *utils.SignalManager) *SelectTriggerHandler {
	return &SelectTriggerHandler{
		selectID: selectID,
		signals:  signals,
	}
}

// detectSideExpr returns a JS expression that measures viewport space and sets the side signal.
func (s *SelectTriggerHandler) detectSideExpr() string {
	return fmt.Sprintf(
		"(function(){var r=el.getBoundingClientRect();%s=(window.innerHeight-r.bottom<200&&r.top>window.innerHeight-r.bottom)?'top':'bottom'})()",
		s.signals.Signal("side"),
	)
}

// BuildClickHandler creates the trigger click handler
func (s *SelectTriggerHandler) BuildClickHandler() string {
	return fmt.Sprintf(
		"%s ? (%s, %s) : (%s, %s, %s)",
		s.signals.Signal("open"),
		s.signals.Set("open", "false"),
		s.signals.Set("highlighted", "-1"),
		s.detectSideExpr(),
		s.signals.Set("open", "true"),
		s.signals.Set("highlighted", "-1"),
	)
}

// BuildKeyboardHandler creates the trigger keyboard navigation handler
func (s *SelectTriggerHandler) BuildKeyboardHandler() string {
	expr := utils.NewExpression()
	detect := s.detectSideExpr()

	// ArrowDown: open dropdown and highlight first item
	expr.Statement(fmt.Sprintf("evt.key === 'ArrowDown' && !%s ? (%s, %s, %s, evt.preventDefault()) : null",
		s.signals.Signal("open"),
		detect,
		s.signals.Set("open", "true"),
		s.signals.Set("highlighted", "0")))

	// ArrowUp: open dropdown and highlight last item
	expr.Statement(fmt.Sprintf("evt.key === 'ArrowUp' && !%s ? (%s, %s, %s, evt.preventDefault()) : null",
		s.signals.Signal("open"),
		detect,
		s.signals.Set("open", "true"),
		s.signals.Set("highlighted", "document.querySelector('[data-select-id=\""+s.selectID+"\"]').querySelectorAll('[data-select-item]:not([data-disabled])').length - 1")))

	// Space: open dropdown without highlighting
	expr.Statement(fmt.Sprintf("evt.key === ' ' && !%s ? (%s, %s, %s, evt.preventDefault()) : null",
		s.signals.Signal("open"),
		detect,
		s.signals.Set("open", "true"),
		s.signals.Set("highlighted", "-1")))

	// Enter: just open dropdown if closed, don't highlight anything
	expr.Statement(fmt.Sprintf("evt.key === 'Enter' && !%s ? (%s, %s, %s, evt.preventDefault()) : null",
		s.signals.Signal("open"),
		detect,
		s.signals.Set("open", "true"),
		s.signals.Set("highlighted", "-1")))

	return expr.Build()
}

// SelectContentHandler creates handlers for select content functionality
type SelectContentHandler struct {
	selectID string
	signals  *utils.SignalManager
}

// NewSelectContentHandler creates a select content handler
func NewSelectContentHandler(selectID string, signals *utils.SignalManager) *SelectContentHandler {
	return &SelectContentHandler{
		selectID: selectID,
		signals:  signals,
	}
}

// BuildKeyboardHandler creates the content keyboard navigation handler
func (s *SelectContentHandler) BuildKeyboardHandler() string {
	maxItemsExpr := fmt.Sprintf(
		"document.querySelector('[data-select-id=\"%s\"]').querySelectorAll('[data-select-item]:not([data-disabled])').length - 1",
		s.selectID,
	)

	selectOpenCheck := fmt.Sprintf(
		"document.querySelector('[data-select-id=\"%s\"]') && %s",
		s.selectID,
		s.signals.Signal("open"),
	)

	// Build individual handlers using comma operator for multiple statements
	arrowDown := fmt.Sprintf(
		"evt.key === 'ArrowDown' && %s ? (evt.preventDefault(), evt.stopPropagation(), %s) : null",
		selectOpenCheck,
		s.signals.Set("highlighted", fmt.Sprintf("Math.min(%s, %s + 1)", maxItemsExpr, s.signals.Signal("highlighted"))),
	)

	arrowUp := fmt.Sprintf(
		"evt.key === 'ArrowUp' && %s ? (evt.preventDefault(), evt.stopPropagation(), %s) : null",
		selectOpenCheck,
		s.signals.Set("highlighted", fmt.Sprintf("Math.max(0, %s - 1)", s.signals.Signal("highlighted"))),
	)

	enterSpace := fmt.Sprintf(
		"(evt.key === 'Enter' || evt.key === ' ') && %s && %s >= 0 ? (evt.preventDefault(), evt.stopPropagation(), document.querySelector('[data-select-id=\"%s\"]').querySelector('[data-select-item][data-index=\"' + %s + '\"]')?.click()) : null",
		selectOpenCheck,
		s.signals.Signal("highlighted"),
		s.selectID,
		s.signals.Signal("highlighted"),
	)

	escape := fmt.Sprintf(
		"evt.key === 'Escape' && %s ? (evt.preventDefault(), evt.stopPropagation(), %s) : null",
		selectOpenCheck,
		s.signals.Set("open", "false"),
	)

	tab := fmt.Sprintf(
		"evt.key === 'Tab' && %s ? %s : null",
		selectOpenCheck,
		s.signals.Set("open", "false"),
	)

	return utils.NewExpression().
		Statement(arrowDown).
		Statement(arrowUp).
		Statement(enterSpace).
		Statement(escape).
		Statement(tab).
		Build()
}

// SelectItemHandler creates handlers for select item functionality
type SelectItemHandler struct {
	selectID string
	value    string
	onChange string
}

// NewSelectItemHandler creates a select item handler
func NewSelectItemHandler(selectID, value, onChange string) *SelectItemHandler {
	return &SelectItemHandler{
		selectID: selectID,
		value:    value,
		onChange: onChange,
	}
}

// BuildClickHandler creates the item click handler
func (s *SelectItemHandler) BuildClickHandler() string {
	// Extract label from clicked item
	labelExpr := "evt.currentTarget.querySelector('.select-item-text')?.textContent.trim() || ''"
	valueExpr := fmt.Sprintf("%q", s.value)

	base := fmt.Sprintf(`(function(){ $%[1]s.value = %[2]s; $%[1]s.label = %[3]s; $%[1]s.open = false; var input = document.querySelector('[data-select-id="%[1]s"] input[name]'); if (input) { input.value = %[2]s; } })()`,
		s.selectID,
		valueExpr,
		labelExpr,
	)

	if s.onChange != "" {
		return base + "; " + s.onChange
	}
	return base
}

// BuildKeyboardHandler creates the item keyboard handler
func (s *SelectItemHandler) BuildKeyboardHandler() string {
	return `(evt.key === 'Enter' || evt.key === ' ') ? (evt.preventDefault(), evt.target.click()) : null`
}
