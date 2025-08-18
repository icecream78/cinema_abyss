package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/icecream78/cinema_abyss/src/microservices/events/internal/dto"
	service_movie "github.com/icecream78/cinema_abyss/src/microservices/events/internal/service/movie"
	service_payment "github.com/icecream78/cinema_abyss/src/microservices/events/internal/service/payment"
	service_user "github.com/icecream78/cinema_abyss/src/microservices/events/internal/service/user"
)

type movieService interface {
	PublishMovieUpdate(event service_movie.MovieEvent) (*dto.PublishedEvent, error)
}

type userService interface {
	PublishUserUpdate(event service_user.UserEvent) (*dto.PublishedEvent, error)
}

type paymentService interface {
	PublisNewPayment(event service_payment.NewPaymentEvent) (*dto.PublishedEvent, error)
}

type handler struct {
	movieService   movieService
	userService    userService
	paymentService paymentService
}

func RegisterRouteHandlers(
	e *echo.Echo,
	movieService movieService,
	userService userService,
	paymentService paymentService,
) error {
	h := handler{
		movieService:   movieService,
		userService:    userService,
		paymentService: paymentService,
	}

	e.POST("/api/events/movie", h.publishMovieEvent)
	e.POST("/api/events/user", h.publishUserEvent)
	e.POST("/api/events/payment", h.publishPaymentEvent)

	// healt check
	e.GET("/api/events/health", h.healthCheck)

	return nil
}

type healthCheckAnswer struct {
	IsHealthy bool `json:"status"`
	Resources resourcesHealthCheck
}

type resourcesHealthCheck struct {
	Eventbus bool
}

func (h *handler) healthCheck(c echo.Context) error {
	c.JSON(http.StatusOK, healthCheckAnswer{
		IsHealthy: true,
		Resources: resourcesHealthCheck{
			Eventbus: true,
		},
	})

	return nil
}

type movieEventIn struct {
	ID          int64    `json:"movie_id"`
	Title       string   `json:"title"`
	Action      string   `json:"action"`
	UserID      int64    `json:"user_id"`
	Rating      float64  `json:"rating"`
	Genres      []string `json:"genres"`
	Description string   `json:"description"`
}

type movieEventOut struct {
	eventPublishResponse
}

type eventPublishResponse struct {
	Status    string            `json:"status"`
	Partition int64             `json:"partition"`
	Offset    int64             `json:"offset"`
	Event     eventInfoResponse `json:"event"`
}

type eventInfoResponse struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Timestamp string            `json:"timestamp"`
	Payload   map[string]string `json:"payload"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *handler) publishMovieEvent(c echo.Context) error {
	var in movieEventIn
	if err := c.Bind(&in); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{
			Error: "Invalid input",
		})
	}
	if in.ID == 0 || in.Title == "" || in.Action == "" {
		return c.JSON(http.StatusBadRequest, errorResponse{
			Error: "Must be passed required fields: movie_id, title, action",
		})
	}

	event := service_movie.MovieEvent{
		ID:          in.ID,
		Title:       in.Title,
		Action:      in.Action,
		UserID:      in.UserID,
		Rating:      in.Rating,
		Genres:      in.Genres,
		Description: in.Description,
	}

	publishedEventInfo, err := h.movieService.PublishMovieUpdate(event)
	if err != nil {
		slog.Error("failed publish new movie", slog.String("error", err.Error()))

		return c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "Internal server error",
		})
	}

	c.JSON(http.StatusCreated, movieEventOut{
		eventPublishResponse: eventPublishResponse{
			Status:    "success",
			Partition: publishedEventInfo.Details.Partition,
			Offset:    publishedEventInfo.Details.Offset,
			Event: eventInfoResponse{
				ID:        publishedEventInfo.Payload.ID,
				Type:      publishedEventInfo.Payload.Type,
				Timestamp: publishedEventInfo.Payload.Timestamp.Format(time.RFC3339),
				Payload:   publishedEventInfo.Payload.Payload,
			},
		},
	})

	return nil
}

type userEventIn struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
}

type userEventOut struct {
	eventPublishResponse
}

func (h *handler) publishUserEvent(c echo.Context) error {
	var in userEventIn
	if err := c.Bind(&in); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{
			Error: "Invalid input",
		})
	}
	if in.UserID == 0 || in.Timestamp == "" || in.Action == "" {
		return c.JSON(http.StatusBadRequest, errorResponse{
			Error: "Must be passed required fields: user_id, timestamp, action",
		})
	}

	event := service_user.UserEvent{
		Action:    in.Action,
		UserID:    in.UserID,
		Username:  in.Username,
		Email:     in.Email,
		Timestamp: in.Timestamp,
	}

	publishedEventInfo, err := h.userService.PublishUserUpdate(event)
	if err != nil {
		slog.Error("failed publish user event", slog.String("error", err.Error()))

		return c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "Internal server error",
		})
	}

	c.JSON(http.StatusCreated, userEventOut{
		eventPublishResponse: eventPublishResponse{
			Status:    "success",
			Partition: publishedEventInfo.Details.Partition,
			Offset:    publishedEventInfo.Details.Offset,
			Event: eventInfoResponse{
				ID:        publishedEventInfo.Payload.ID,
				Type:      publishedEventInfo.Payload.Type,
				Timestamp: publishedEventInfo.Payload.Timestamp.Format(time.RFC3339),
				Payload:   publishedEventInfo.Payload.Payload,
			},
		},
	})

	return nil
}

type paymentEventIn struct {
	UserID      int64   `json:"user_id"`
	PaymentID   int64   `json:"payment_id"`
	Amount      float64 `json:"amount"`
	Status      string  `json:"status"`
	Timestamp   string  `json:"timestamp"`
	PaymentType string  `json:"method_type"`
}

type paymentEventOut struct {
	eventPublishResponse
}

func (h *handler) publishPaymentEvent(c echo.Context) error {
	var in paymentEventIn
	if err := c.Bind(&in); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{
			Error: "Invalid input",
		})
	}
	if in.UserID == 0 || in.PaymentID == 0 || in.Amount == 0 || in.Status == "" || in.Timestamp == "" {
		return c.JSON(http.StatusBadRequest, errorResponse{
			Error: "Must be passed required fields: payment_id, user_id, amount, status, timestamp",
		})
	}

	event := service_payment.NewPaymentEvent{
		UserID:      in.UserID,
		PaymentID:   in.PaymentID,
		Amount:      in.Amount,
		Status:      in.Status,
		PaymentType: in.PaymentType,
		Timestamp:   in.Timestamp,
	}

	publishedEventInfo, err := h.paymentService.PublisNewPayment(event)
	if err != nil {
		slog.Error("failed publish user event", slog.String("error", err.Error()))

		return c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "Internal server error",
		})
	}

	c.JSON(http.StatusCreated, userEventOut{
		eventPublishResponse: eventPublishResponse{
			Status:    "success",
			Partition: publishedEventInfo.Details.Partition,
			Offset:    publishedEventInfo.Details.Offset,
			Event: eventInfoResponse{
				ID:        publishedEventInfo.Payload.ID,
				Type:      publishedEventInfo.Payload.Type,
				Timestamp: publishedEventInfo.Payload.Timestamp.Format(time.RFC3339),
				Payload:   publishedEventInfo.Payload.Payload,
			},
		},
	})

	return nil
}
