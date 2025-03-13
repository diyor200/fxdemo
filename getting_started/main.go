package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Provide(
			NewHTTPServer,

			AsRoute(NewEchoHandler),
			AsRoute(NewHelloHandler),

			fx.Annotate(
				NewServeMux,
				fx.ParamTags(`group:"routes"`),
			),
			zap.NewExample,
		),
		fx.Invoke(func(*http.Server) {}),
	).Run()
}

func NewHTTPServer(lc fx.Lifecycle, mux *http.ServeMux, log *zap.Logger) *http.Server {
	srv := &http.Server{Addr: ":8080", Handler: mux}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				log.Error("Error starting HTTP server:", zap.Error(err))
				return err
			}

			log.Info("HTTP server listening on", zap.String("address", ln.Addr().String()))
			go srv.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Shutting down HTTP server")
			return srv.Shutdown(ctx)
		},
	})

	return srv
}

// handler
type EchoHandler struct {
	log *zap.Logger
}

func NewEchoHandler(log *zap.Logger) *EchoHandler {
	return &EchoHandler{log: log}
}

func (h *EchoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, err := io.Copy(w, r.Body); err != nil {
		h.log.Warn("Error serving HTTP request:", zap.Error(err))
	}
}

func (h *EchoHandler) Pattern() string {
	return "/echo"
}

// NewServeMux builds a ServeMux that will route requests
// to the given EchoHandler.
func NewServeMux(routes []Route) *http.ServeMux {
	mux := http.NewServeMux()
	for i := range routes {
		mux.Handle(routes[i].Pattern(), routes[i])
	}
	return mux
}

// Route is a http.Handler that knows the mux pattern
// under which it will be registered
type Route interface {
	http.Handler

	// Pattern reports the path at which this is registered
	Pattern() string
}

func AsRoute(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(Route)),
		fx.ResultTags(`group:"routes"`),
	)
}

// HelloHandler is an HTTP handler that
// prints a greeting to the user
type HelloHandler struct {
	log *zap.Logger
}

// NewHelloHandler builds a new HelloHandler
func NewHelloHandler(log *zap.Logger) *HelloHandler {
	return &HelloHandler{log: log}
}

func (h *HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Error("Error serving HTTP request", zap.Error(err))
		http.Error(w, "failed while processing the request body", http.StatusInternalServerError)
		return
	}

	if _, err := fmt.Fprintf(w, "Hello %s", body); err != nil {
		h.log.Error("Error serving HTTP request", zap.Error(err))
		http.Error(w, "failed writing the response", http.StatusInternalServerError)
		return
	}
}

func (h *HelloHandler) Pattern() string {
	return "/hello"
}
