package processing

import (
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
)

func slicesEqual(a1, a2 []string) bool {
	if len(a1) != len(a2) {
		return false
	}

	sort.Strings(a1)
	sort.Strings(a2)

	for i := 0; i < len(a1); i++ {
		if !strings.EqualFold(a1[i], a2[i]) {
			log.Debug().Msg("Both the Slices are NotEqual")
			return false
		}
	}
	log.Debug().Msg("Both the Slices are Equal")
	return true
}
