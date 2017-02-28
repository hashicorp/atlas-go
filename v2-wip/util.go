package atlas

// maskString masks all but the first few characters of a string for display
// output. This is useful for tokens so we can display them to the user without
// showing the full output.
func maskString(s string) string {
	if len(s) <= 3 {
		return "*** (masked)"
	}

	return s[0:3] + "*** (masked)"
}
