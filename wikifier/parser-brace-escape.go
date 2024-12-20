package wikifier

type braceEscape struct {
	parent catch
	*genericCatch
}

func newBraceEscape() *braceEscape {
	return &braceEscape{genericCatch: &genericCatch{}}
}

func (be *braceEscape) catchType() catchType {
	return catchTypeBraceEscape
}

func (be *braceEscape) parentCatch() catch {
	return be.parent
}

func (be *braceEscape) byteOK(b byte) bool {
	return true
}

func (be *braceEscape) shouldSkipByte(b byte) bool {
	return false
}
