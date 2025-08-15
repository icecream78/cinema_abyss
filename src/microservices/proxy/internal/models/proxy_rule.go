package models

import (
	"context"
	"math/rand/v2"
	"net/http"
	"regexp"

	"github.com/pkg/errors"
)

type ProxyMode int

const (
	ProxyOnlyToMain         ProxyMode = 0 // default working mode.
	ProxyToBothServers      ProxyMode = 1
	ProxyOnlyToMirrorServer ProxyMode = 2
)

type ReverseProxyServer interface {
	ServeHTTP(response http.ResponseWriter, request *http.Request)
	IsHealthy(ctx context.Context) (bool, error)
}

type ProxyRule struct {
	proxyMode         ProxyMode
	uriPattern        *regexp.Regexp
	MainProxyServer   ReverseProxyServer
	MirrorProxyServer ReverseProxyServer
	splitPercentage   int
}

func NewProxyRule(
	matchUriPattern string,
	mainProxy ReverseProxyServer,
	mirrorProxy ReverseProxyServer,
	splitPercentage int,
) (*ProxyRule, error) {
	parsedUriPattern, err := regexp.Compile(matchUriPattern)
	if err != nil {
		return nil, errors.Wrap(err, "compile regexp")
	}

	var proxyMode ProxyMode

	switch splitPercentage {
	case 0:
		proxyMode = ProxyOnlyToMain
	case 100:
		proxyMode = ProxyOnlyToMirrorServer
	default:
		proxyMode = ProxyToBothServers
	}

	return &ProxyRule{
		uriPattern:        parsedUriPattern,
		proxyMode:         proxyMode,
		MainProxyServer:   mainProxy,
		MirrorProxyServer: mirrorProxy,
		splitPercentage:   splitPercentage,
	}, nil
}

func (pr *ProxyRule) IsRuleForRequest(request *http.Request) bool {
	return pr.uriPattern.Match([]byte(request.RequestURI))
}

func (pr *ProxyRule) MatchURI() string {
	return pr.uriPattern.String()
}

func (pr *ProxyRule) DecideWhichServerToUseAsProxy(request *http.Request) ReverseProxyServer {
	switch pr.proxyMode {
	case ProxyOnlyToMain:
		return pr.MainProxyServer
	case ProxyOnlyToMirrorServer:
		return pr.MirrorProxyServer
	default:
		if rand.IntN(100) < pr.splitPercentage {
			return pr.MirrorProxyServer
		}

		return pr.MainProxyServer
	}
}
