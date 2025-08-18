package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Config struct {
	Environment       string
	HttpServer        HttpServer
	ProxyRules        []ProxyRule
	FallbackServerURL string
}

type HttpServer struct {
	ListenURL string
}

type ProxyRule struct {
	Description       string
	ToURL             string
	RouteRegexpString string
	ProxyPercent      int
}

func Get() (*Config, error) {
	environment := "production"
	if env := os.Getenv("ENV"); env != "" {
		environment = env
	}

	httpListenURL := "0.0.0.0:8081"
	if port := os.Getenv("PORT"); port != "" {
		httpListenURL = "0.0.0.0:" + port
	}

	fallbackServerURL := os.Getenv("MONOLITH_URL")
	if fallbackServerURL == "" {
		return nil, errors.New("empty monolith_url")
	}

	var proxyRules []ProxyRule
	if os.Getenv("GRADUAL_MIGRATION") == "true" {
		for i := 1; ; i++ {
			// перебираем правила проксирования, пока они не закончатся
			envKeyForMapping := fmt.Sprintf("PROXY_RULE_%d_MAPPING", i)
			envKeyForPercent := fmt.Sprintf("PROXY_RULE_%d_PERCENT", i)

			proxyPercent, _ := strconv.Atoi(os.Getenv(envKeyForPercent))

			// считаем, что правила закончились
			if proxyPercent == 0 {
				break
			}

			if proxyPercent < 0 || proxyPercent > 100 {
				return nil, errors.New("incorrect proxy percent value")
			}

			ruleMappingSlice := strings.Split(os.Getenv(envKeyForMapping), " -> ")

			// считаем, что правила закончились
			if len(ruleMappingSlice) == 0 {
				break
			}

			proxyRules = append(proxyRules, ProxyRule{
				Description:       os.Getenv(envKeyForMapping),
				RouteRegexpString: ruleMappingSlice[0],
				ToURL:             ruleMappingSlice[1],
				ProxyPercent:      proxyPercent,
			})
		}
	}

	return &Config{
		Environment: environment,
		HttpServer: HttpServer{
			ListenURL: httpListenURL,
		},
		ProxyRules:        proxyRules,
		FallbackServerURL: fallbackServerURL,
	}, nil
}
