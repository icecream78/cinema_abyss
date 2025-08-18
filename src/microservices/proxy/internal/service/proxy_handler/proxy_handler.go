package proxy_handler

import (
	"net/http"

	"github.com/icecream78/cinema_abyss/src/microservices/proxy/internal/models"
)

type ruleStorage interface {
	SearchRule(request *http.Request) (*models.ProxyRule, error)
}

type fallbackServer interface {
	ServeHTTP(response http.ResponseWriter, request *http.Request)
}

type ProxyHandler struct {
	ruleStorage    ruleStorage
	fallbackServer fallbackServer
}

func New(
	ruleStorage ruleStorage,
	fallbackServer fallbackServer,
) (*ProxyHandler, error) {
	return &ProxyHandler{
		ruleStorage:    ruleStorage,
		fallbackServer: fallbackServer,
	}, nil
}

func (s *ProxyHandler) HandleRequest(
	incomingRequest *http.Request,
	incomingResponse http.ResponseWriter,
) error {
	rule, err := s.ruleStorage.SearchRule(incomingRequest)
	if err != nil || rule == nil {
		s.fallbackServer.ServeHTTP(incomingResponse, incomingRequest)
		return nil
	}

	proxyServer := rule.DecideWhichServerToUseAsProxy(incomingRequest)
	proxyServer.ServeHTTP(incomingResponse, incomingRequest)

	return nil
}
