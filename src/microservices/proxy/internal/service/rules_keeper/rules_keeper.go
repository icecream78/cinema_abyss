package rules_keeper

import (
	"net/http"

	"github.com/icecream78/cinema_abyss/src/microservices/proxy/internal/models"
)

type RulesKeeper struct {
	rules []*models.ProxyRule
}

func New(rules []*models.ProxyRule) (*RulesKeeper, error) {
	return &RulesKeeper{
		rules: rules,
	}, nil
}

func (s *RulesKeeper) SearchRule(request *http.Request) (*models.ProxyRule, error) {
	for _, rule := range s.rules {
		if rule.IsRuleForRequest(request) {
			return rule, nil
		}
	}

	return nil, nil
}

func (s *RulesKeeper) Rules() []*models.ProxyRule {
	return s.rules
}
