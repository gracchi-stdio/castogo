package blockEditor

// ContentFromSignal looks up the block type and delegates to its
// ContentFromSignals method. Minimal pass-through for Slice 0; the
// underlying method body is filled in by Slice 1.1.
func ContentFromSignal(blockID int64, blockType string, signals map[string]any) map[string]any {
	t := Lookup(blockType)
	if t == nil {
		return nil
	}
	return t.ContentFromSignals(blockID, signals)
}
