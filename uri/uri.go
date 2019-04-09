package uri

import (
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"strings"
)

func PruneString(raw string) (string, error) {

	allowed := [][]int{
		[]int{48, 57},  // (0-9)
		[]int{65, 90},  // (A-Z)
		[]int{97, 122}, // (a-z)
	}

	clean, err := PruneStringWithAllowed(raw, allowed)

	if err != nil {
		return "", err
	}

	return strings.ToLower(clean), nil
}

func PruneStringWithAllowed(raw string, allowed [][]int) (string, error) {

	rm_func := func(r rune) bool {

		is_allowed := false

		for _, bookends := range allowed {

			r_int := int(r)

			if r_int >= bookends[0] && r_int <= bookends[1] {
				is_allowed = true
				break
			}
		}

		if is_allowed {
			return false
		}

		return true
	}

	tr := transform.Chain(norm.NFD, transform.RemoveFunc(rm_func), norm.NFC)
	clean, _, err := transform.String(tr, raw)

	return clean, err
}
