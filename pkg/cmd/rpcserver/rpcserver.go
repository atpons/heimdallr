package rpcserver

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"go.etcd.io/etcd/v3/clientv3"
	"go.f110.dev/protoc-ddl/probe"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	"go.f110.dev/heimdallr/pkg/auth"
	"go.f110.dev/heimdallr/pkg/cert"
	"go.f110.dev/heimdallr/pkg/cmd"
	"go.f110.dev/heimdallr/pkg/config"
	"go.f110.dev/heimdallr/pkg/config/configreader"
	"go.f110.dev/heimdallr/pkg/database"
	"go.f110.dev/heimdallr/pkg/database/etcd"
	"go.f110.dev/heimdallr/pkg/database/mysql"
	"go.f110.dev/heimdallr/pkg/database/mysql/dao"
	"go.f110.dev/heimdallr/pkg/database/mysql/entity"
	"go.f110.dev/heimdallr/pkg/logger"
	"go.f110.dev/heimdallr/pkg/rpc/rpcserver"
)

const (
	datastoreTypeEtcd  = "etcd"
	datastoreTypeMySQL = "mysql"
)

const (
	stateInit cmd.State = iota
	stateSetup
	stateStart
	stateShutdown
)

type mainProcess struct {
	*cmd.FSM

	ConfFile string
	Config   *config.Config

	server *rpcserver.Server

	datastoreType   string
	etcdClient      *clientv3.Client
	conn            *sql.DB
	ca              *cert.CertificateAuthority
	userDatabase    database.UserDatabase
	clusterDatabase database.ClusterDatabase
	tokenDatabase   database.TokenDatabase
	relayLocator    database.RelayLocator

	mu    sync.Mutex
	ready bool
}

func New() *mainProcess {
	m := &mainProcess{}
	m.FSM = cmd.NewFSM(
		map[cmd.State]cmd.StateFunc{
			stateInit:     m.init,
			stateSetup:    m.setup,
			stateStart:    m.start,
			stateShutdown: m.shutdown,
		},
		stateInit,
		stateShutdown,
	)

	return m
}

func (m *mainProcess) init() (cmd.State, error) {
	conf, err := configreader.ReadConfig(m.ConfFile)
	if err != nil {
		return cmd.UnknownState, xerrors.Errorf(": %v", err)
	}
	m.Config = conf

	if m.Config.Datastore.Url != nil {
		m.datastoreType = datastoreTypeEtcd
	}
	if m.Config.Datastore.DSN != nil {
		m.datastoreType = datastoreTypeMySQL
	}

	return stateSetup, nil
}

func (m *mainProcess) setup() (cmd.State, error) {
	if err := logger.Init(m.Config.Logger); err != nil {
		return cmd.UnknownState, xerrors.Errorf(": %v", err)
	}
	switch m.datastoreType {
	case datastoreTypeEtcd:
		client, err := m.Config.Datastore.GetEtcdClient(m.Config.Logger)
		if err != nil {
			return cmd.UnknownState, xerrors.Errorf(": %v", err)
		}
		go m.watchGRPCConnState(client.ActiveConnection())
		m.etcdClient = client

		m.userDatabase, err = etcd.NewUserDatabase(context.Background(), client, database.SystemUser)
		if err != nil {
			return cmd.UnknownState, xerrors.Errorf(": %v", err)
		}
		caDatabase, err := etcd.NewCA(context.Background(), client)
		if err != nil {
			return cmd.UnknownState, xerrors.Errorf(": %v", err)
		}
		m.ca = cert.NewCertificateAuthority(caDatabase, m.Config.General.CertificateAuthority)
		m.clusterDatabase, err = etcd.NewClusterDatabase(context.Background(), client)
		if err != nil {
			return cmd.UnknownState, xerrors.Errorf(": %v", err)
		}

		if m.Config.General.Enable {
			m.tokenDatabase = etcd.NewTemporaryToken(client)
			m.relayLocator, err = etcd.NewRelayLocator(context.Background(), client)
			if err != nil {
				return cmd.UnknownState, xerrors.Errorf(": %v", err)
			}
		}
	case datastoreTypeMySQL:
		m.datastoreType = datastoreTypeMySQL
		conn, err := sql.Open("mysql", m.Config.Datastore.DSN.FormatDSN())
		if err != nil {
			return cmd.UnknownState, xerrors.Errorf(": %w", err)
		}
		m.conn = conn

		repository := dao.NewRepository(conn)
		m.userDatabase = mysql.NewUserDatabase(repository, database.SystemUser)
		caDatabase := mysql.NewCA(repository)
		m.ca = cert.NewCertificateAuthority(caDatabase, m.Config.General.CertificateAuthority)
		m.clusterDatabase, err = mysql.NewCluster(repository)
		if err != nil {
			return cmd.UnknownState, xerrors.Errorf(": %w", err)
		}

		if m.Config.General.Enable {
			m.tokenDatabase = mysql.NewTokenDatabase(repository)
			m.relayLocator = mysql.NewRelayLocator(repository)
		}
	default:
		return cmd.UnknownState, xerrors.New("cmd/rpcserver: required external datastore")
	}

	m.server = rpcserver.NewServer(m.Config, m.userDatabase, m.tokenDatabase, m.clusterDatabase, m.relayLocator, m.ca, m.IsReady)

	auth.InitInterceptor(m.Config, m.userDatabase, m.tokenDatabase)
	return stateStart, nil
}

func (m *mainProcess) start() (cmd.State, error) {
	go func() {
		if err := m.server.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "%+v\n", err)
		}
	}()

	if m.datastoreType == datastoreTypeEtcd {
		c, err := etcd.NewCompactor(m.etcdClient)
		if err != nil {
			return cmd.UnknownState, xerrors.Errorf(": %v", err)
		}

		go func() {
			c.Start(context.Background())
		}()
	}

	return cmd.WaitState, nil
}

func (m *mainProcess) shutdown() (cmd.State, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFunc()

	if err := m.server.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
	}

	return cmd.CloseState, nil
}

func (m *mainProcess) IsReady() bool {
	switch m.datastoreType {
	case datastoreTypeEtcd:
		m.mu.Lock()
		defer m.mu.Unlock()
		return m.ready
	case datastoreTypeMySQL:
		p := probe.NewProbe(m.conn)
		ctx, cancelFunc := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancelFunc()

		return p.Ready(ctx, entity.SchemaHash)
	}

	return false
}

func (m *mainProcess) watchGRPCConnState(conn *grpc.ClientConn) {
	state := conn.GetState()
	for conn.WaitForStateChange(context.Background(), state) {
		state = conn.GetState()
		m.mu.Lock()
		switch state {
		case connectivity.Ready:
			m.ready = true
		default:
			m.ready = false
		}
		m.mu.Unlock()
	}
}
