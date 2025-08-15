package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"

	"github.com/icecream78/cinema_abyss/src/microservices/proxy/internal/models"
)

type rulesKeeper interface {
	Rules() []*models.ProxyRule
}

type proxyHandler interface {
	HandleRequest(
		incomingRequest *http.Request,
		incomingResponse http.ResponseWriter,
	) error
}

type handler struct {
	rulesKeeper  rulesKeeper
	proxyHandler proxyHandler
}

func RegisterRouteHandlers(
	e *echo.Echo,
	rulesKeeper rulesKeeper,
	proxyHandler proxyHandler,
) error {
	h := handler{
		rulesKeeper:  rulesKeeper,
		proxyHandler: proxyHandler,
	}

	// перехватываем весь входящий трафик
	e.Any("*", h.proxyRoutes)

	e.GET("/health", h.healthCheck)

	return nil
}

type healthCheckAnswer struct {
	IsHealthy                    bool `json:"status"`
	ResourcesAvailabilityPerRule map[string]resourceAvailabilityPerRule
}

type resourceAvailabilityPerRule struct {
	IsMainHealthy   bool
	IsMirrorHealthy bool
}

func (h *handler) healthCheck(c echo.Context) error {
	ctx := c.Request().Context()

	registeredRules := h.rulesKeeper.Rules()

	// список ресурсов, по которым происходит проксирование
	resourcesAvailability := make(map[string]resourceAvailabilityPerRule, len(registeredRules))
	eg, egCtx := errgroup.WithContext(ctx)
	for _, rule := range registeredRules {
		var isMainHealthy, isMirrorHealthy bool

		eg.Go(func() error {
			isMainHealthy, _ = rule.MainProxyServer.IsHealthy(egCtx)

			return nil
		})

		eg.Go(func() error {
			isMirrorHealthy, _ = rule.MirrorProxyServer.IsHealthy(egCtx)
			return nil
		})

		_ = eg.Wait()

		resourcesAvailability[rule.MatchURI()] = resourceAvailabilityPerRule{
			IsMainHealthy:   isMainHealthy,
			IsMirrorHealthy: isMirrorHealthy,
		}
	}

	c.JSON(http.StatusOK, healthCheckAnswer{
		IsHealthy:                    true,
		ResourcesAvailabilityPerRule: resourcesAvailability,
	})

	return nil
}

func (h *handler) proxyRoutes(c echo.Context) error {
	err := h.proxyHandler.HandleRequest(c.Request(), c.Response())
	if err != nil {
		// TODO: перевести на slog
		fmt.Println(err)
	}
	return nil
}
