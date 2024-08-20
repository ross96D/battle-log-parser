package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/ross96D/cwbattle_parser/parser"
	"github.com/rs/zerolog/log"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrNoUrlParam      Error = "required url param not set"
	ErrInvalidUrlParam Error = "invalid url param"
)

func Server() *echo.Echo {
	s := echo.New()

	s.POST("/parse", parse)

	return s
}

func parse(c echo.Context) error {
	urlStr := c.QueryParam("url")
	if urlStr == "" {
		return ErrNoUrlParam
	}

	_, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("url.Parse() %s %w", urlStr, ErrInvalidUrlParam)
	}

	resp, err := http.Get(urlStr)
	if err != nil {
		return fmt.Errorf("http.Get %s %w", urlStr, err)
	}
	defer resp.Body.Close()

	b, err := parser.Parse(resp.Body)
	if err != nil {
		// TODO log io.Reader data
		log.Error().Err(err).Msg("parsing function error")
		return fmt.Errorf("parser.Parse %w", err)
	}

	jenc := json.NewEncoder(c.Response())
	err = jenc.Encode(b)
	if err != nil {
		return fmt.Errorf("jenc.Encode %w", err)
	}

	return nil
}

func ValidateUrlParam(uri string) error {
	_, err := url.Parse(uri)

	return err
}
