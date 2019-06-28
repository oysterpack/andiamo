/*
 * Copyright (c) 2019 OysterPack, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fxapp

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"net/http"
	"sort"
	"sync"
	"time"
)

// HTTPHandler is used to group HTTPEndpoint(s) together.
// The HTTPEndpoint(s) are automatically registered with the app's HTTP server.
type HTTPHandler struct {
	fx.Out

	HTTPEndpoint `group:"HTTPHandler"`
}

// NewHTTPHandler constructs a new HTTPHandler
func NewHTTPHandler(path string, handler func(http.ResponseWriter, *http.Request)) HTTPHandler {
	return HTTPHandler{
		HTTPEndpoint: HTTPEndpoint{
			Path:    path,
			Handler: handler,
		},
	}
}

// HTTPEndpoint maps an HTTP handler to an HTTP path
type HTTPEndpoint struct {
	Path    string
	Handler func(http.ResponseWriter, *http.Request)
}

// httpServerOpts is used by the app to configure and run an HTTP server only if HTTPEndpoint(s) are discovered, i.e.,
// registered with the app via dependency injection.
//
// An http.Server can be provided when building the app. If an http.Server is not found, then the app creates one with the
// following options:
// 	- Addr:              ":8008",
//	- ReadHeaderTimeout: time.Second,
//	- MaxHeaderBytes:    1024,
type httpServerOpts struct {
	fx.In

	Server *http.Server `name:"http.Server" optional:"true"`

	Endpoints []HTTPEndpoint `group:"HTTPHandler"`
}

// validate runs the following checks:
//	- endpoint paths are unique
//	- handler funcs are not nil
func (opts httpServerOpts) validate() error {
	paths := make(map[string]bool, len(opts.Endpoints))
	for _, endpoint := range opts.Endpoints {
		if paths[endpoint.Path] {
			return fmt.Errorf("duplicate HTTP endpoint path: %v", endpoint.Path)
		}
		if endpoint.Handler == nil {
			return fmt.Errorf("http handler func is nil for: %v", endpoint.Path)
		}
		paths[endpoint.Path] = true
	}

	return nil
}

func (opts httpServerOpts) httpServerInfo() httpServerInfo {
	endpoints := make([]string, 0, len(opts.Endpoints))
	for _, endpoint := range opts.Endpoints {
		endpoints = append(endpoints, endpoint.Path)
	}
	sort.Strings(endpoints)

	return httpServerInfo{
		addr:      opts.Server.Addr,
		endpoints: endpoints,
	}
}

func runHTTPServer(opts httpServerOpts, logger *zerolog.Logger, lc fx.Lifecycle) error {
	if len(opts.Endpoints) == 0 {
		return nil
	}

	if err := opts.validate(); err != nil {
		return err
	}

	serveMux := http.NewServeMux()
	for _, endpoint := range opts.Endpoints {
		serveMux.HandleFunc(endpoint.Path, endpoint.Handler)
	}

	if opts.Server == nil {
		opts.Server = newHTTPServerWithDefaultOpts()
	}
	opts.Server.Handler = serveMux

	errorLog := httpServerErrorLog(HTTPServerError.NewLogEventer(logger, zerolog.ErrorLevel))
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			HTTPServerStarting.NewLogEventer(logger, zerolog.InfoLevel)(opts.httpServerInfo(), "starting HTTP server")
			// wait for the HTTP server go routine to start running before returning
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				wg.Done()
				err := opts.Server.ListenAndServe()
				if err != http.ErrServerClosed {
					errorLog(httpListenAndServerError{err}, "HTTP server has exited with an error")
				}
			}()
			wg.Wait()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return opts.Server.Shutdown(ctx)
		},
	})

	return nil
}

func newHTTPServerWithDefaultOpts() *http.Server {
	return &http.Server{
		Addr:              ":8008",
		ReadHeaderTimeout: time.Second,
		MaxHeaderBytes:    1024,
	}
}

// HTTP server related events
const (
	// HTTPServerError indicates an error occurred while handling a metrics scrape HTTP request.
	HTTPServerError EventTypeID = "01DEDRH8A9X3SCSJRCJ4PM7749"

	HTTPServerStarting EventTypeID = "01DEFM9FFSH58ZGNPSR7Z4C3G2"
)

type httpServerErrorLog LogEventer

func (log httpServerErrorLog) Println(v ...interface{}) {
	log(httpServerError(fmt.Sprint(v...)), "HTTP Server error")
}

type httpServerError string

func (err httpServerError) MarshalZerologObject(e *zerolog.Event) {
	e.Err(errors.New(string(err)))
}

type httpListenAndServerError struct {
	error
}

func (err httpListenAndServerError) MarshalZerologObject(e *zerolog.Event) {
	e.Err(err)
}

type httpServerInfo struct {
	addr      string
	endpoints []string
}

func (info httpServerInfo) MarshalZerologObject(e *zerolog.Event) {
	e.
		Str("addr", info.addr).
		Strs("endpoints", info.endpoints)

}
