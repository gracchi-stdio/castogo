package blockEditor

import (
	"fmt"

	"github.com/gracchi-stdio/castogo/internal/domain"
)

func TypeLabel(name string) string {
	if t := Lookup(name); t != nil {
		return t.Label
	}
	return name
}

func SummaryExpr(block *domain.PageBlock) string {
	t := Lookup(block.BlockType)
	if t == nil {
		return "'Empty block'"
	}
	return fmt.Sprintf(t.Summary, blockEditPrefix(block, ""))
}
