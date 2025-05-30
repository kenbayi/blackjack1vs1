package def

// Pointer return pointer on any value.
func Pointer[T any](v T) *T { return &v }
func Deref[T any](ptr *T, defaultVal T) T {
	if ptr != nil {
		return *ptr
	}
	return defaultVal
}
