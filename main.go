package main

import (
	"net/http"
	"strings"

	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
)

var (
	e      = echo.New()
	logger = log.New("condenser")
)

func main() {
	initLogger()

	// Force the check for a config file at launch
	_ = getConfig()

	e.HTTPErrorHandler = echo.HTTPErrorHandler(condenserHTTPErrorHandler)
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())

	// v1
	e.POST("/api/v1/shorten", shorten, apiKeyMiddleware)
	e.POST("/api/v1/delete", delete, apiKeyMiddleware)
	e.GET("/api/v1/meta/:code", meta)

	// Legacy
	e.POST("/api/shorten", shorten, apiKeyMiddleware)
	e.POST("/api/delete", delete, apiKeyMiddleware)
	e.GET("/api/meta/:code", meta)

	// Shortcodes
	e.GET("/:code", shortcode)

	if listenAddr := os.Getenv("CONDENSER_LISTEN"); listenAddr != "" {
		logger.Fatal(e.Start(listenAddr))
	} else {
		logger.Fatal(e.Start(":8000"))
	}
}

func initLogger() {
	logger.SetLevel(log.OFF)

	dbgEnv := os.Getenv("CONDENSER_DEBUG")
	if dbgEnv != "" {
		e.Logger.SetLevel(log.INFO)
		logger.SetLevel(log.DEBUG)
		logger.Warn("Running in DEBUG configuration.")
	}
}

func shorten(c echo.Context) error {
	req := new(shortenRequest)
	if err := c.Bind(req); err != nil {
		return err // already in HTTPError if applicable.
	}
	logger.Debugf("shorten: %+v", req)

	lowerURL := strings.ToLower(req.URL)
	if !(strings.HasPrefix(lowerURL, "http://") || strings.HasPrefix(lowerURL, "https://")) {
		return echo.NewHTTPError(http.StatusBadRequest, "URLs must start with http(s)://")
	}

	var fullURL string
	errCode := insertDBError
	key := c.Get("APIKey").(APIKey)
	if req.Code != "" {
		fullURL, errCode = addURLWithCode(req.URL, req.Code, req.Meta, key)
	} else {
		fullURL, errCode = addURL(req.URL, req.Meta, key)
	}

	switch errCode {
	case insertSuccess:
		return c.JSON(http.StatusOK, &shortenResponse{
			ShortURL: fullURL,
		})
	case insertConflict:
		return echo.NewHTTPError(http.StatusConflict, "Code already exists.")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "Something went wrong on our end.")
	}
}

func delete(c echo.Context) error {
	req := new(deleteRequest)
	if err := c.Bind(req); err != nil {
		return err // already in HTTPError if applicable.
	}
	logger.Debugf("delete: %+v", req)

	resp, errCode := deleteCode(req.Code)
	switch errCode {
	case deleteSuccess:
		return c.JSON(http.StatusOK, resp)
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "Something went wrong on our end.")
	}
}

func meta(c echo.Context) error {
	code := c.Param("code")
	logger.Debugf("meta: %s", code)
	meta, errCode := getCodeMeta(code)
	switch errCode {
	case lookupSuccess:
		return c.JSON(http.StatusOK, meta)
	case lookupNoSuchCode:
		return echo.NewHTTPError(http.StatusNotFound, "Shortcode does not exist.")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "Something went wrong on our end.")
	}
}

func shortcode(c echo.Context) error {
	code := c.Param("code")
	logger.Debugf("shortcode: %s", code)
	target, errCode := getFullURL(code)
	switch errCode {
	case lookupSuccess:
		return c.Redirect(http.StatusTemporaryRedirect, target)
	case lookupNoSuchCode:
		return echo.NewHTTPError(http.StatusNotFound, "Shortcode does not exist.")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "Something went wrong on our end.")
	}
}
