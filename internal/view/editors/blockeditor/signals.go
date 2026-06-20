package blockEditor

import (
	"github.com/a-h/templ"
	"github.com/gracchi-stdio/castogo/internal/domain"
)

func BlockSignalAttr(block *domain.PageBlock) templ.Attributes {
	t := Lookup(block.BlockType)
	if t == nil {
		return nil
	}
	return t.SignalAttrs(block.ID, parseBlockContent(block))
}
