package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// APIKey represents a single API key + Name pair in the configuration.
type APIKey struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

var apiKeyMiddleware = middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
	KeyLookup:  "header:X-API-Key",
	AuthScheme: "",
	Validator:  middleware.KeyAuthValidator(checkKey),
})

func checkKey(key string, ctx echo.Context) bool {
	for i := 0; i < len(getConfig().APIKeys); i++ {
		if currKey := getConfig().APIKeys[i]; currKey.Key == key {
			ctx.Set("APIKey", currKey)
			return true
		}
	}
	return false
}
