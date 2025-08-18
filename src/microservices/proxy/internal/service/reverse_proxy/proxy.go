package reverse_proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/pkg/errors"
)

type ReverseProxy struct {
	id           string
	url          string
	reverseProxy *httputil.ReverseProxy
}

func New(id string, proxyURL string) (*ReverseProxy, error) {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, errors.Wrap(err, "url parse")
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(parsedURL)

	return &ReverseProxy{
		id:           id,
		url:          proxyURL,
		reverseProxy: reverseProxy,
	}, nil
}

func (s *ReverseProxy) ID() string {
	return s.id
}

func (s *ReverseProxy) ProxyURL() string {
	return s.url
}

func (s *ReverseProxy) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	originalDirector := s.reverseProxy.Director

	s.reverseProxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Preserve original request path and query parameters
		req.URL.RawQuery = req.URL.Query().Encode()
	}

	// TODO: нужно реализовать какие-то базовые методы для хендлинга ошибок проксирования. без этого хендлера я не мог понять почему фейлится запрос, а не хватало http вначале
	s.reverseProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		fmt.Println(err)
	}

	s.reverseProxy.ServeHTTP(response, request)
}

// TODO: нужно реализовать поход за статусом сервиса. По хорошему, он должен быть общего формата, т.к. ы контроллируем инфру
func (s *ReverseProxy) IsHealthy(ctx context.Context) (bool, error) {
	return true, nil
}
