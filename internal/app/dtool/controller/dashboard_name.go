package controller

import (
	"strings"

	"github.com/spf13/cast"
)

func dashboardNameCounts(rows []map[string]any, key string) map[string]int {
	counts := make(map[string]int, len(rows))
	for _, row := range rows {
		name := strings.TrimSpace(cast.ToString(row[key]))
		if name == `` {
			continue
		}
		counts[strings.ToLower(name)]++
	}
	return counts
}

func dashboardHasDuplicateName(counts map[string]int, name any) bool {
	key := strings.ToLower(strings.TrimSpace(cast.ToString(name)))
	return key != `` && counts[key] > 1
}

func dashboardTokenPart(value any) string {
	return strings.Join(strings.Fields(strings.TrimSpace(cast.ToString(value))), `_`)
}

func dashboardJoinName(parts ...any) string {
	ret := make([]string, 0, len(parts))
	for _, part := range parts {
		token := dashboardTokenPart(part)
		if token == `` {
			continue
		}
		ret = append(ret, token)
	}
	return strings.Join(ret, `-`)
}
