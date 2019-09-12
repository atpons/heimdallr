package identityprovider

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/f110/lagrangian-proxy/pkg/config"
	"github.com/f110/lagrangian-proxy/pkg/database"
	"github.com/f110/lagrangian-proxy/pkg/logger"
	"github.com/f110/lagrangian-proxy/pkg/session"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/oauth2"
	"golang.org/x/xerrors"
)

var allowCipherSuites = []uint16{
	tls.TLS_AES_128_GCM_SHA256,
	tls.TLS_AES_256_GCM_SHA384,
	tls.TLS_CHACHA20_POLY1305_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
}

type claims struct {
	Email    string `json:"email"`
	Verified bool   `json:"email_verified"`
}

type Server struct {
	Config *config.IdentityProvider

	server       *http.Server
	database     userDatabase
	sessionStore session.Store
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
}

type userDatabase interface {
	Get(ctx context.Context, id string) (*database.User, error)
}

func NewServer(conf *config.IdentityProvider, database userDatabase, store session.Store) (*Server, error) {
	issuer := ""
	switch conf.Provider {
	case "google":
		issuer = "https://accounts.google.com"
	}
	provider, err := oidc.NewProvider(context.Background(), issuer)
	if err != nil {
		return nil, xerrors.Errorf(": %v", err)
	}
	scopes := []string{oidc.ScopeOpenID}
	if len(conf.ExtraScopes) > 0 {
		scopes = append(scopes, conf.ExtraScopes...)
	}
	oauth2Config := oauth2.Config{
		ClientID:     conf.ClientId,
		ClientSecret: conf.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  conf.RedirectUrl,
		Scopes:       scopes,
	}

	s := &Server{
		Config:       conf,
		database:     database,
		sessionStore: store,
		oauth2Config: oauth2Config,
		verifier:     provider.Verifier(&oidc.Config{ClientID: conf.ClientId}),
	}

	mux := httprouter.New()
	mux.GET("/auth", s.handleAuth)
	mux.GET("/auth/callback", s.handleCallback)
	server := &http.Server{
		ErrorLog: logger.CompatibleLogger,
		Handler:  mux,
	}
	s.server = server

	return s, nil
}

func (s *Server) Start() error {
	l, err := net.Listen("tcp", s.Config.Bind)
	if err != nil {
		return err
	}
	listener := tls.NewListener(l, &tls.Config{
		MinVersion:   tls.VersionTLS12,
		CipherSuites: allowCipherSuites,
		Certificates: []tls.Certificate{s.Config.Certificate},
	})

	if err := http2.ConfigureServer(s.server, &http2.Server{}); err != nil {
		return err
	}

	logger.Log.Info("Start IdentityProvide Server", zap.String("listen", s.Config.Bind))
	return s.server.Serve(listener)
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	logger.Log.Info("Shutdown IdentityProvide Server")
	return s.server.Shutdown(ctx)
}

func (s *Server) handleAuth(w http.ResponseWriter, req *http.Request, _params httprouter.Params) {
	http.Redirect(w, req, s.oauth2Config.AuthCodeURL(""), http.StatusFound)
}

func (s *Server) handleCallback(w http.ResponseWriter, req *http.Request, _params httprouter.Params) {
	token, err := s.oauth2Config.Exchange(req.Context(), req.URL.Query().Get("code"))
	if err != nil {
		logger.Log.Info("Failed exchange token", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	rawIdToken, ok := token.Extra("id_token").(string)
	if !ok {
		logger.Log.Info("Failed covert token to raw id token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	idToken, err := s.verifier.Verify(req.Context(), rawIdToken)
	if err != nil {
		logger.Log.Info("Not verified id token", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	c := &claims{}
	if err := idToken.Claims(c); err != nil {
		logger.Log.Info("Failed extract claims", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if c.Email == "" {
		logger.Log.Info("Could not get email address. Probably, you should set more scope.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := s.database.Get(req.Context(), c.Email)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	sess := session.New(user.Id)
	if err := s.sessionStore.SetSession(w, sess); err != nil {
		logger.Log.Info("Failed write session", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.WriteString(w, "success!")
}