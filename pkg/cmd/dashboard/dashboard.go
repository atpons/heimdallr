package dashboard

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	"go.f110.dev/heimdallr/pkg/cmd"
	"go.f110.dev/heimdallr/pkg/config"
	"go.f110.dev/heimdallr/pkg/config/configreader"
	"go.f110.dev/heimdallr/pkg/dashboard"
	"go.f110.dev/heimdallr/pkg/logger"
)

const (
	stateInit cmd.State = iota
	stateSetup
	stateOpenConnection
	stateStart
	stateShutdown
)

type mainProcess struct {
	*cmd.FSM
	ConfFile string

	config        *config.Config
	rpcServerConn *grpc.ClientConn
	dashboard     *dashboard.Server
}

func New() *mainProcess {
	m := &mainProcess{}
	m.FSM = cmd.NewFSM(
		map[cmd.State]cmd.StateFunc{
			stateInit:           m.init,
			stateSetup:          m.setup,
			stateOpenConnection: m.openConnection,
			stateStart:          m.start,
			stateShutdown:       m.shutdown,
		},
		stateInit,
		stateShutdown,
	)

	return m
}

func (m *mainProcess) init() (cmd.State, error) {
	conf, err := configreader.ReadConfig(m.ConfFile)
	if err != nil {
		return cmd.UnknownState, xerrors.Errorf(": %w", err)
	}
	m.config = conf

	return stateSetup, nil
}

func (m *mainProcess) setup() (cmd.State, error) {
	if err := logger.Init(m.config.Logger); err != nil {
		return cmd.UnknownState, xerrors.Errorf(": %w", err)
	}

	return stateOpenConnection, nil
}

func (m *mainProcess) openConnection() (cmd.State, error) {
	cred := credentials.NewTLS(&tls.Config{
		ServerName: m.config.General.ServerNameHost,
		RootCAs:    m.config.General.CertificateAuthority.CertPool,
	})
	conn, err := grpc.Dial(
		m.config.General.RpcTarget,
		grpc.WithTransportCredentials(cred),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: 20 * time.Second, Timeout: time.Second, PermitWithoutStream: true}),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor()),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor()),
	)
	if err != nil {
		return cmd.UnknownState, xerrors.Errorf(": %v", err)
	}
	m.rpcServerConn = conn

	return stateStart, nil
}

func (m *mainProcess) start() (cmd.State, error) {
	dashboardServer := dashboard.NewServer(m.config, m.rpcServerConn)
	m.dashboard = dashboardServer
	if err := m.dashboard.Start(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
	}

	return cmd.WaitState, nil
}

func (m *mainProcess) shutdown() (cmd.State, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFunc()

	if m.dashboard != nil {
		if err := m.dashboard.Shutdown(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "%+v\n", err)
		}
	}

	return cmd.CloseState, nil
}
