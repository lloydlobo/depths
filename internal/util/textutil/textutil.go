package textutil

// TextSubtext
func Subtext(s string, start, end int) string {
	if end >= len(s) {
		return s
	}
	return s[start:end]
}
