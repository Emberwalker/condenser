package main

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var e = echo.New()

func main() {
	e.HTTPErrorHandler = echo.HTTPErrorHandler(condenserHTTPErrorHandler)
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())

	e.POST("/api/shorten", shorten, apiKeyMiddleware)
	e.POST("/api/delete", delete, apiKeyMiddleware)
	e.GET("/api/meta/:code", meta)
	e.GET("/:code", shortcode)

	e.Logger.Fatal(e.Start(":8000"))
}

func shorten(c echo.Context) error {
	req := new(shortenRequest)
	if err := c.Bind(req); err != nil {
		return err // already in HTTPError if applicable.
	}

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

	resp, errCode := deleteCode(req.Code)
	switch errCode {
	case deleteSuccess:
		return c.JSON(http.StatusOK, resp)
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "Something went wrong on our end.")
	}
}

func meta(c echo.Context) error {
	meta, errCode := getCodeMeta(c.Param("code"))
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
	target, errCode := getFullURL(c.Param("code"))
	switch errCode {
	case lookupSuccess:
		return c.Redirect(http.StatusTemporaryRedirect, target)
	case lookupNoSuchCode:
		return echo.NewHTTPError(http.StatusNotFound, "Shortcode does not exist.")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "Something went wrong on our end.")
	}
}
