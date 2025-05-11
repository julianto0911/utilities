package utilities

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
)

const (
	TYPE_QUERY = 1
	TYPE_JSON  = 2
)

type MockNetRequest struct{}

func (m MockNetRequest) RequestWithJSON(reqType string, address string, data any, header map[string]string) ([]byte, error) {
	return nil, nil
}

func (m MockNetRequest) RequestWithQuery(reqType string, address string, data, header map[string]string) ([]byte, error) {
	return nil, nil
}

func NewNetRequest(rt http.RoundTripper, timeOut time.Duration, log *zap.Logger, debug bool) NetRequest {
	return netRequest{
		debug: debug,
		log:   log,
		client: &http.Client{
			Transport: rt,
			Timeout:   timeOut,
		},
	}
}

type NetRequest interface {
	RequestWithJSON(reqType string, address string, data any, header map[string]string) (int, []byte, error)
	RequestWithQuery(reqType string, address string, data, header map[string]string) (int, []byte, error)
}

type netRequest struct {
	debug  bool
	log    *zap.Logger
	client *http.Client
}

func makeHeader(data map[string]string) http.Header {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	for i, val := range data {
		h.Set(i, val)
	}
	return h
}

func (n netRequest) RequestWithJSON(reqType string, address string, data any, header map[string]string) (int, []byte, error) {
	errHandle := func(err error) (int, []byte, error) {
		n.log.Error("request with json",
			zap.String("address", address),
			zap.Any("data", data),
			zap.Any("header", header),
			zap.Error(err),
		)
		return 0, nil, err
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return errHandle(fmt.Errorf("marshal %w", err))
	}

	//parse url to safety string
	host, err := url.Parse(address)
	if err != nil {
		return errHandle(fmt.Errorf("parse url %w", err))
	}

	//init http request
	req, err := http.NewRequest(reqType, host.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return errHandle(fmt.Errorf("init request %w", err))
	}

	//work with header
	req.Header = makeHeader(header)

	//do request
	resp, err := n.client.Do(req)
	if err != nil {
		return errHandle(fmt.Errorf("do request %w", err))
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errHandle(fmt.Errorf("read body %w", err))
	}

	if n.debug {
		n.log.Info("request with json",
			zap.String("address", address),
			zap.Any("data", data),
			zap.Any("header", header),
			zap.String("response", string(body)),
		)
	}

	return resp.StatusCode, body, nil
}

func (n netRequest) RequestWithQuery(reqType string, address string, data, header map[string]string) (int, []byte, error) {
	errHandle := func(err error) (int, []byte, error) {
		n.log.Error("request with query",
			zap.String("address", address),
			zap.Any("data", data),
			zap.Any("header", header),
			zap.Error(err),
		)
		return 0, nil, err
	}

	//parse url to safety string
	host, err := url.Parse(address)
	if err != nil {
		return errHandle(fmt.Errorf("parse url %w", err))
	}

	//init http GET request
	req, err := http.NewRequest(reqType, host.String(), nil)
	if err != nil {
		return errHandle(fmt.Errorf("init request %w", err))
	}

	//work with query
	q := req.URL.Query()
	for k, v := range data {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	//work with header
	req.Header = makeHeader(header)

	//do request
	resp, err := n.client.Do(req)
	if err != nil {
		return errHandle(fmt.Errorf("do request %w", err))
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errHandle(fmt.Errorf("read body %w", err))
	}

	if n.debug {
		n.log.Info("request with json",
			zap.String("address", address),
			zap.Any("data", data),
			zap.Any("header", header),
			zap.String("response", string(body)),
		)
	}

	return resp.StatusCode, body, nil
}
