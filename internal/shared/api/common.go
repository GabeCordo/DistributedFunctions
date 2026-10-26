// Package api
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  common.go
package api

import (
	"fmt"
	"io"
	"net/http"
)

var client = http.Client{}

const coreHost = "http://localhost:8136"

func closeBody(body io.ReadCloser) {
	if err := body.Close(); err != nil {
		fmt.Print(err)
	}
}

type Response struct {
	Success     bool   `json:"success"`
	Description string `json:"description"`
	Data        any    `json:"data"`
}

type dataResponse[T any] struct {
	Data T `json:"data"`
}

type runResponse struct {
	Id uint64 `json:"id"`
}
