package matcher

import "strings"

// MatchesTokens checks if the includes list is matched in the supplied message string. If the includes
// list is empty everything is matched.
func MatchesTokens(includes []string, msg string, fallback bool) bool {
	// if there are no include patterns just let everything through
	if len(includes) == 0 {
		return fallback
	}

	for _, inc := range includes {
		if strings.Contains(msg, inc) {
			return true
		}
	}

	return false
}
