package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func requestLogger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(
		middleware.RequestLoggerConfig{
			LogMethod:   true,
			LogURIPath:  true,
			LogStatus:   true,
			HandleError: true,

			LogValuesFunc: func(
				c *echo.Context,
				v middleware.RequestLoggerValues,
			) error {
				level := slog.LevelInfo

				switch {
				case v.Status >= http.StatusInternalServerError:
					level = slog.LevelError
				case v.Status >= http.StatusBadRequest:
					level = slog.LevelWarn
				}

				attrs := []slog.Attr{
					slog.String("method", v.Method),
					slog.String("path", v.URIPath),
					slog.Int("status", v.Status),
				}

				if level == slog.LevelError && v.Error != nil {
					attrs = append(
						attrs,
						slog.Any("error", v.Error),
					)
				}

				c.Logger().LogAttrs(
					c.Request().Context(),
					level,
					"HTTP request",
					attrs...,
				)

				return nil
			},
		},
	)
}
