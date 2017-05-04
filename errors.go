package main

import "github.com/labstack/echo"

type genericError struct {
	Err string `json:"error"`
	Msg string `json:"message,omitempty"`
}

func condenserHTTPErrorHandler(errIn error, ctx echo.Context) {
	if err, ok := errIn.(*echo.HTTPError); ok {
		code := err.Code
		switch code {
		case 400:
			if jsonErr := ctx.JSON(code, genericError{
				Err: "nokey",
				Msg: "No API key in X-API-Key header.",
			}); jsonErr != nil {
				e.Logger.Error(jsonErr)
			}
			return
		case 401:
			if jsonErr := ctx.JSON(code, genericError{
				Err: "invalidkey",
				Msg: "Invalid API key in X-API-Key header.",
			}); jsonErr != nil {
				e.Logger.Error(jsonErr)
			}
			return
		}
	}
	e.DefaultHTTPErrorHandler(errIn, ctx)
}
