package destructor

import "strings"

// combine does a join() on all the non-empty terms
func combine(sep string, terms ...string) string {
	var nonemptyTerms []string
	for _, s := range terms {
		if s != "" {
			nonemptyTerms = append(nonemptyTerms, s)
		}
	}
	return strings.Join(nonemptyTerms, sep)
}
