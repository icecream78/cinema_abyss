package main

import (
	"log/slog"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/icecream78/cinema_abyss/src/microservices/proxy/internal/config"
	"github.com/icecream78/cinema_abyss/src/microservices/proxy/internal/handler"
	"github.com/icecream78/cinema_abyss/src/microservices/proxy/internal/models"
	"github.com/icecream78/cinema_abyss/src/microservices/proxy/internal/service/proxy_handler"
	"github.com/icecream78/cinema_abyss/src/microservices/proxy/internal/service/reverse_proxy"
	"github.com/icecream78/cinema_abyss/src/microservices/proxy/internal/service/rules_keeper"
)

type exitCode int

const (
	succesExitCode exitCode = 0
	errorExitCode  exitCode = 1
)

func main() {
	os.Exit(int(run()))
}

func run() exitCode {
	appConfig, err := config.Get()
	if err != nil {
		slog.Error("during get config", slog.String("error", err.Error()))
		return errorExitCode
	}

	e := echo.New()

	if appConfig.Environment != "production" {
		e.Use(middleware.Logger())
	}

	fallbackServer, err := reverse_proxy.New("monolith", appConfig.FallbackServerURL)
	if err != nil {
		slog.Error("create proxy server from", slog.String("error", err.Error()))
		return errorExitCode
	}
	proxyRules := make([]*models.ProxyRule, 0, len(appConfig.ProxyRules))

	for _, proxyRuleRaw := range appConfig.ProxyRules {
		proxy, err := reverse_proxy.New(proxyRuleRaw.Description, proxyRuleRaw.ToURL)
		if err != nil {
			slog.Error("create proxy server to", slog.String("error", err.Error()))
			return errorExitCode
		}

		proxyRule, err := models.NewProxyRule(
			proxyRuleRaw.RouteRegexpString,
			fallbackServer,
			proxy,
			proxyRuleRaw.ProxyPercent,
		)
		if err != nil {
			slog.Error("create proxy rule", slog.String("error", err.Error()))
			return errorExitCode
		}

		proxyRules = append(proxyRules, proxyRule)
	}

	rulesKeeper, err := rules_keeper.New(proxyRules)
	if err != nil {
		slog.Error("create rules keeper", slog.String("error", err.Error()))
		return errorExitCode
	}

	proxyHandler, err := proxy_handler.New(rulesKeeper, fallbackServer)
	if err != nil {
		slog.Error("create proxy handler", slog.String("error", err.Error()))
		return errorExitCode
	}

	if err = handler.RegisterRouteHandlers(e, rulesKeeper, proxyHandler); err != nil {
		slog.Error("during register http handlers", slog.String("error", err.Error()))
		return errorExitCode
	}

	e.Logger.Fatal(e.Start(appConfig.HttpServer.ListenURL))

	return succesExitCode
}
