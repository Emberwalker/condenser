package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"
	"strings"
)

var e = echo.New()

func main() {
	e.HTTPErrorHandler = echo.HTTPErrorHandler(CondenserHTTPErrorHandler)
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())

	e.POST("/api/shorten", shorten, APIKeyMiddleware)
	e.POST("/api/delete", delete, APIKeyMiddleware)
	e.GET("/api/meta/:code", meta)
	e.GET("/:code", shortcode)

	e.Logger.Fatal(e.Start(":8000"))
}

func shorten(c echo.Context) error {
	req := new(ShortenRequest)
	if err := c.Bind(req); err != nil {
		return err // already in HTTPError if applicable.
	}

	lowerURL := strings.ToLower(req.URL)
	if !(strings.HasPrefix(lowerURL, "http://") || strings.HasPrefix(lowerURL, "https://")) {
		return echo.NewHTTPError(http.StatusBadRequest, "URLs must start with http(s)://")
	}

	var fullUrl string
	errCode := InsertDBError
	key := c.Get("APIKey").(APIKey)
	if req.Code != "" {
		fullUrl, errCode = addURLWithCode(req.URL, req.Code, req.Meta, key)
	} else {
		fullUrl, errCode = addURL(req.URL, req.Meta, key)
	}

	switch errCode {
	case InsertSuccess:
		return c.JSON(http.StatusOK, &ShortenResponse{
			ShortURL: fullUrl,
		})
	case InsertConflict:
		return echo.NewHTTPError(http.StatusConflict, "Code already exists.")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "Something went wrong on our end.")
	}
}

func delete(c echo.Context) error {
	req := new(DeleteRequest)
	if err := c.Bind(req); err != nil {
		return err // already in HTTPError if applicable.
	}

	resp, errCode := deleteCode(req.Code)
	switch errCode {
	case DeleteSuccess:
		return c.JSON(http.StatusOK, resp)
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "Something went wrong on our end.")
	}
}

func meta(c echo.Context) error {
	meta, errCode := getCodeMeta(c.Param("code"))
	switch errCode {
	case LookupSuccess:
		return c.JSON(http.StatusOK, meta)
	case LookupNoSuchCode:
		return echo.NewHTTPError(http.StatusNotFound, "Shortcode does not exist.")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "Something went wrong on our end.")
	}
}

func shortcode(c echo.Context) error {
	target, errCode := getFullURL(c.Param("code"))
	switch errCode {
	case LookupSuccess:
		return c.Redirect(http.StatusTemporaryRedirect, target)
	case LookupNoSuchCode:
		return echo.NewHTTPError(http.StatusNotFound, "Shortcode does not exist.")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "Something went wrong on our end.")
	}
}
