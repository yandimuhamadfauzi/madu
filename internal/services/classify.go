package services

import (
	"net/url"
	"strings"

	"github.com/yandimuhamadfauzi/madu/internal/models"
)

func Classify(raw string, rules []models.Rule) *models.Rule {
	s := raw
	for i := 0; i < 2; i++ {
		if d, err := url.QueryUnescape(s); err == nil {
			s = d
		}
	}
	s = strings.ToLower(s)
	for i := range rules {
		if rules[i].Active && strings.Contains(s, strings.ToLower(rules[i].Pattern)) {
			return &rules[i]
		}
	}
	return nil
}
