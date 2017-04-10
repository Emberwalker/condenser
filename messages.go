package main

import "time"

type (
	ShortenRequest struct {
		URL  string `json:"url"`
		Code string `json:"code,omitempty"`
		Meta string `json:"meta,omitempty"`
	}

	ShortenResponse struct {
		ShortURL string `json:"short_url"`
	}

	DeleteRequest struct {
		Code string `json:"code"`
	}

	DeleteResponse struct {
		Code   string `json:"code"`
		Status string `json:"status"`
	}

	CodeMeta struct {
		Owner    string    `json:"owner"`
		Time     time.Time `json:"time"`
		UserMeta string    `json:"user_meta,omitempty"`
	}

	CodeMetaResponse struct {
		FullURL string   `json:"full_url"`
		Meta    CodeMeta `json:"meta"`
	}
)
