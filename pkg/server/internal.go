package server

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"

	"go.f110.dev/heimdallr/pkg/config"
	"go.f110.dev/heimdallr/pkg/logger"
)

type Internal struct {
	Config *config.Config

	server *http.Server
}

func NewInternal(conf *config.Config, child ...ChildServer) *Internal {
	mux := httprouter.New()
	for _, v := range child {
		v.Route(mux)
	}

	return &Internal{
		Config: conf,
		server: &http.Server{
			Addr:     conf.General.BindInternalApi,
			ErrorLog: logger.CompatibleLogger,
			Handler:  mux,
		},
	}
}

func (s *Internal) Start() error {
	logger.Log.Info("Start Internal server", zap.String("listen", s.Config.General.BindInternalApi))
	return s.server.ListenAndServe()
}

func (s *Internal) Shutdown(ctx context.Context) error {
	logger.Log.Info("Shutdown Internal server")
	return s.server.Shutdown(ctx)
}
