package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// WARN We should implement our own Http client over the default from http instead of using
// the default one just raw (timeout conf among other things)

const (
	versionRoute = "/v3/services/haproxy/configuration/version"
	bindRoute    = "/v3/services/haproxy/configuration/frontends/my-frontend/binds"
	// WARN Should not be hardcoded but injected from the same source as the one defined in haproxy.cfg
	backendRoute = "/v3/services/haproxy/configuration/backends/web-backend/servers"
)

const (
	httpTimeOut = 5 * time.Second
)

type Bind struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Port    int    `json:"port"`
}

type BackendServer struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Port    int    `json:"port"`
}

type LoadBalancerWrapper interface {
	AddBind(bind Bind) error
	DeleteBind(name string) error
	AddBackendServer(bind BackendServer) error
	DeleteBackendServer(name string) error
	// WARN Would be logical to add interface to list existing bindings & backend servers. But same principle than upper ones.
}

type LoadBalancerWrapperImpl struct {
	logger   *zerolog.Logger
	address  string // Expected format : "http://myAddress"
	user     string
	password string
	client   http.Client
}

// WARN We should check address format
func NewLoadBalancerWrapperImpl(address, user, password string) *LoadBalancerWrapperImpl {
	logger := log.With().Str("service", "wrapper").Logger()
	return &LoadBalancerWrapperImpl{
		logger:   &logger,
		address:  address,
		user:     user,
		password: password,
		client:   http.Client{Timeout: httpTimeOut},
	}
}

func (impl *LoadBalancerWrapperImpl) AddBind(bind Bind) error {
	body, errMarshal := json.Marshal(bind)
	if errMarshal != nil {
		return errMarshal
	}

	_, errRequest := impl.requestBuilder(http.MethodPost, bindRoute, bytes.NewBuffer(body), true)

	return errRequest
}

func (impl *LoadBalancerWrapperImpl) DeleteBind(name string) error {
	_, errRequest := impl.requestBuilder(http.MethodDelete, fmt.Sprintf("%v/%v", bindRoute, name), http.NoBody, true)
	return errRequest
}

func (impl *LoadBalancerWrapperImpl) AddBackendServer(backend BackendServer) error {
	body, errMarshal := json.Marshal(backend)
	if errMarshal != nil {
		return errMarshal
	}

	_, errRequest := impl.requestBuilder(http.MethodPost, backendRoute, bytes.NewBuffer(body), true)

	return errRequest
}

func (impl *LoadBalancerWrapperImpl) DeleteBackendServer(name string) error {
	_, errRequest := impl.requestBuilder(http.MethodDelete, fmt.Sprintf("%v/%v", backendRoute, name), http.NoBody, true)
	return errRequest
}

// WARN Version should be cached some way, instead of retrieving it each time.
func (impl *LoadBalancerWrapperImpl) getVersion() (string, error) {
	res, errGet := impl.requestBuilder(http.MethodGet, versionRoute, http.NoBody, false)
	if errGet != nil {
		impl.logger.Error().Err(errGet).Msgf("Failed to get HAProxy version")
		return "", errGet
	}

	// WARN Should be packaged with checks
	return string(res[:len(res)-1]), nil
}

func (impl *LoadBalancerWrapperImpl) requestBuilder(method, route string, body io.Reader, versionRequired bool) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeOut)
	defer cancel()
	req, errNewRequest := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%v%v", impl.address, route), body)
	if errNewRequest != nil {
		return nil, errNewRequest
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Add("Version", "HTTP/1.1")

	// Authentication
	req.SetBasicAuth(impl.user, impl.password)

	// Add version query param when necessary
	if versionRequired {
		if v, errGetVersion := impl.getVersion(); errGetVersion != nil {
			return nil, errGetVersion
		} else {
			q := req.URL.Query()
			q.Add("version", v)
			req.URL.RawQuery = q.Encode()
		}
	}

	// Call
	resp, errDo := impl.client.Do(req)
	if errDo != nil {
		return nil, errDo
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			impl.logger.Error().Err(err).Msg(err.Error())
			return
		}
	}()

	// Handle result
	contents, errReadAll := io.ReadAll(resp.Body)
	if errReadAll != nil {
		return nil, errReadAll
	}
	if resp.StatusCode-http.StatusOK >= http.StatusContinue {
		return nil, errors.New(http.StatusText(resp.StatusCode))
	}
	return contents, nil
}
