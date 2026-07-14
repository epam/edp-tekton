package formatter

import "unicode/utf8"

const truncationNotice = "\n\n… (truncated)"

// Truncate caps the comment body at maxBytes, cutting on a valid UTF-8 rune
// boundary and appending a truncation notice within the budget.
func Truncate(body string, maxBytes int) string {
	if len(body) <= maxBytes {
		return body
	}

	limit := maxBytes - len(truncationNotice)
	if limit < 0 {
		limit = 0
	}

	body = body[:limit]

	for len(body) > 0 {
		r, size := utf8.DecodeLastRuneInString(body)
		if r != utf8.RuneError || size > 1 {
			break
		}

		body = body[:len(body)-size]
	}

	return body + truncationNotice
}
