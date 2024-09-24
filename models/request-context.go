package models

import (
	"time"

	echo "github.com/labstack/echo/v4"
)

// RequestContext structure
type RequestContext struct {
	TimestampMs *int64  `json:"timestampMs"`
	ClientIP    *string `json:"clientIP"`
	EchoContext echo.Context
}

func CreateRequestContext(c echo.Context) *RequestContext {
	rc := new(RequestContext)

	clientIP := GetClientIP(c)
	nowMs := InstantEpochMilli()

	rc.EchoContext = c
	rc.ClientIP = &clientIP
	rc.TimestampMs = &nowMs

	return rc
}

func GetClientIP(c echo.Context) string {
	// WARNING : There's no guarantee that retrieved header is present in HTTP request (it depends on intermediate Proxy configuration)
	// WARNING : There's no guarantee that Client IP corresponds to Workstation IP, it may correspond to some intermediate proxy.
	return c.Request().Header.Get("X-Forwarded-For")
}

// InstantEpochMilli returns elapsed time since the Unix epoch in milliseconds
func InstantEpochMilli() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
