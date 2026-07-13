package ptr

func Clone[T any](p *T) *T {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

func Equal[T comparable](a, b *T) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return *a == *b
	}
}
