package controller

import "github.com/labstack/echo/v4"

type IController interface {
	Initialize(e *echo.Echo)
	Dispose() error
}
