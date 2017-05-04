package main

import "time"

type (
	shortenRequest struct {
		URL  string `json:"url"`
		Code string `json:"code,omitempty"`
		Meta string `json:"meta,omitempty"`
	}

	shortenResponse struct {
		ShortURL string `json:"short_url"`
	}

	deleteRequest struct {
		Code string `json:"code"`
	}

	deleteResponse struct {
		Code   string `json:"code"`
		Status string `json:"status"`
	}

	codeMeta struct {
		Owner    string    `json:"owner"`
		Time     time.Time `json:"time"`
		UserMeta string    `json:"user_meta,omitempty"`
	}

	codeMetaResponse struct {
		FullURL string   `json:"full_url"`
		Meta    codeMeta `json:"meta"`
	}
)
