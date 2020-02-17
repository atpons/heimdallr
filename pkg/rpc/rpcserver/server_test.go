package rpcserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/f110/lagrangian-proxy/pkg/auth"
	"github.com/f110/lagrangian-proxy/pkg/cert"
	"github.com/f110/lagrangian-proxy/pkg/config"
	"github.com/f110/lagrangian-proxy/pkg/database"
	"github.com/f110/lagrangian-proxy/pkg/database/memory"
	"github.com/f110/lagrangian-proxy/pkg/logger"
	"github.com/f110/lagrangian-proxy/pkg/netutil"
	"github.com/f110/lagrangian-proxy/pkg/rpc"
)

func TestNewServer(t *testing.T) {
	conf := &config.Config{
		General: &config.General{
			CertificateAuthority: &config.CertificateAuthority{},
		},
		RPCServer: &config.RPCServer{
			Bind: ":0",
		},
		Logger: &config.Logger{},
	}
	if err := logger.Init(conf.Logger); err != nil {
		t.Fatal(err)
	}

	v := NewServer(
		conf,
		memory.NewUserDatabase(),
		memory.NewTokenDatabase(),
		memory.NewClusterDatabase(),
		memory.NewRelayLocator(),
		memory.NewCA(conf.General.CertificateAuthority),
	)
	if v == nil {
		t.Fatal("NewServer should return a value")
	}
}

func TestServer_Start(t *testing.T) {
	certBytes, caPrivateKey, err := cert.CreateCertificateAuthority("test", "test", "test", "jp")
	if err != nil {
		t.Fatal(err)
	}
	crt, err := x509.ParseCertificate(certBytes)
	if err != nil {
		t.Fatal(err)
	}
	port, err := netutil.FindUnusedPort()
	if err != nil {
		t.Fatal(err)
	}
	metricPort, err := netutil.FindUnusedPort()
	if err != nil {
		t.Fatal(err)
	}

	conf := &config.Config{
		General: &config.General{
			CertificateAuthority: &config.CertificateAuthority{
				Certificate: crt,
				PrivateKey:  caPrivateKey,
			},
		},
		RPCServer: &config.RPCServer{
			Bind:        fmt.Sprintf(":%d", port),
			MetricsBind: fmt.Sprintf(":%d", metricPort),
		},
		Logger: &config.Logger{},
	}
	if err := logger.Init(conf.Logger); err != nil {
		t.Fatal(err)
	}

	v := NewServer(
		conf,
		memory.NewUserDatabase(),
		memory.NewTokenDatabase(),
		memory.NewClusterDatabase(),
		memory.NewRelayLocator(),
		memory.NewCA(conf.General.CertificateAuthority),
	)
	go func() {
		if err := v.Start(); err != nil {
			t.Fatal(err)
		}
	}()

	if err := netutil.WaitListen(fmt.Sprintf(":%d", port), 5*time.Second); err != nil {
		t.Fatal(err)
	}
	if err := netutil.WaitListen(fmt.Sprintf(":%d", metricPort), 5*time.Second); err != nil {
		t.Fatal(err)
	}

	if err := v.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestServicesViaServer(t *testing.T) {
	certBytes, caPrivateKey, err := cert.CreateCertificateAuthority("test", "test", "test", "jp")
	if err != nil {
		t.Fatal(err)
	}
	crt, err := x509.ParseCertificate(certBytes)
	if err != nil {
		t.Fatal(err)
	}
	port, err := netutil.FindUnusedPort()
	if err != nil {
		t.Fatal(err)
	}

	conf := &config.Config{
		General: &config.General{
			ServerNameHost: "test.example.com",
			CertificateAuthority: &config.CertificateAuthority{
				Certificate: crt,
				PrivateKey:  caPrivateKey,
			},
			InternalToken: "internal-token",
			RootUsers:     []string{database.SystemUser.Id},
			Backends: []*config.Backend{
				{Name: "test"},
			},
			Roles: []*config.Role{
				{
					Name: "test-admin",
					Bindings: []*config.Binding{
						{Rpc: "test-admin"},
					},
				},
				{Name: "test-admin2"},
			},
			RpcPermissions: []*config.RpcPermission{
				{Name: "test-admin", Allow: []string{"/proxy.rpc.Admin/*"}},
			},
		},
		RPCServer: &config.RPCServer{
			Bind: fmt.Sprintf(":%d", port),
		},
		Logger: &config.Logger{},
	}
	if err := conf.General.Load(conf.General.Backends, conf.General.Roles, conf.General.RpcPermissions); err != nil {
		t.Fatal(err)
	}
	if err := logger.Init(conf.Logger); err != nil {
		t.Fatal(err)
	}
	u := memory.NewUserDatabase(database.SystemUser)
	token := memory.NewTokenDatabase()
	auth.InitInterceptor(conf, u, token)

	testUser := &database.User{
		Id:    "test@example.com",
		Type:  database.UserTypeNormal,
		Roles: []string{"test-admin"},
	}
	testUser.Setup()
	if err := u.Set(nil, testUser); err != nil {
		t.Fatal(err)
	}
	code, err := token.NewCode(nil, testUser.Id, "", "")
	if err != nil {
		t.Fatal(err)
	}
	userToken, err := token.IssueToken(nil, code.Code, "")
	if err != nil {
		t.Fatal(err)
	}

	s := NewServer(
		conf,
		u,
		token,
		memory.NewClusterDatabase(),
		memory.NewRelayLocator(),
		memory.NewCA(conf.General.CertificateAuthority),
	)
	go func() {
		if err := s.Start(); err != nil {
			t.Fatal(err)
		}
	}()
	if err := netutil.WaitListen(fmt.Sprintf(":%d", port), 5*time.Second); err != nil {
		t.Fatal(err)
	}

	caPool := x509.NewCertPool()
	caPool.AddCert(crt)
	transCreds := credentials.NewTLS(&tls.Config{
		ServerName: conf.General.ServerNameHost,
		RootCAs:    caPool,
	})
	conn, err := grpc.Dial(fmt.Sprintf(":%d", port), grpc.WithTransportCredentials(transCreds))
	if err != nil {
		t.Fatal(err)
	}

	md := metadata.New(map[string]string{rpc.InternalTokenMetadataKey: "internal-token"})
	systemUserCtx := metadata.NewOutgoingContext(context.Background(), md)

	md = metadata.New(map[string]string{rpc.TokenMetadataKey: userToken.Token})
	testUserCtx := metadata.NewOutgoingContext(context.Background(), md)

	t.Run("Admin", func(t *testing.T) {
		t.Parallel()

		adminClient := rpc.NewAdminClient(conn)
		res, err := adminClient.Ping(context.Background(), &rpc.RequestPing{})
		if err != nil {
			t.Fatal(err)
		}
		if res == nil {
			t.Error("Expect return a value")
		}

		t.Run("Get Config", func(t *testing.T) {
			backendListRes, err := adminClient.BackendList(systemUserCtx, &rpc.RequestBackendList{})
			if err != nil {
				t.Fatal(err)
			}
			if len(backendListRes.Items) != 1 {
				t.Errorf("BackendList should return an array that have 1 element: %d", len(backendListRes.Items))
			}

			roleListRes, err := adminClient.RoleList(systemUserCtx, &rpc.RequestRoleList{})
			if err != nil {
				t.Fatal(err)
			}
			if len(roleListRes.Items) != 3 {
				t.Errorf("RoleList should return an array that have 3 elements: %d", len(roleListRes.Items))
			}
		})

		t.Run("Management User", func(t *testing.T) {
			addRes, err := adminClient.UserAdd(systemUserCtx, &rpc.RequestUserAdd{
				Id:   testUser.Id,
				Type: rpc.UserType_NORMAL,
				Role: "test-admin2",
			})
			if err != nil {
				t.Fatal(err)
			}
			if !addRes.Ok {
				t.Error("Expect return ok")
			}

			getRes, err := adminClient.UserGet(systemUserCtx, &rpc.RequestUserGet{Id: testUser.Id, WithTokens: true})
			if err != nil {
				t.Fatal(err)
			}
			if getRes.User.Id != "test@example.com" {
				t.Errorf("Unexpected Id: %s", getRes.User.Id)
			}
			if getRes.User.Roles[0] != "test-admin" || getRes.User.Roles[1] != "test-admin2" {
				t.Errorf("Unexpected role: %v", getRes.User.Roles)
			}
			if getRes.User.Type != rpc.UserType_NORMAL {
				t.Errorf("Unexpected user type: %v", getRes.User.Type)
			}

			becomeRes, err := adminClient.BecomeMaintainer(systemUserCtx, &rpc.RequestBecomeMaintainer{Id: testUser.Id, Role: "test-admin"})
			if err != nil {
				t.Fatal(err)
			}
			if !becomeRes.Ok {
				t.Error("Expect return ok")
			}

			getRes, err = adminClient.UserGet(systemUserCtx, &rpc.RequestUserGet{Id: "test@example.com"})
			if err != nil {
				t.Fatal(err)
			}
			if len(getRes.User.MaintainRoles) != 1 {
				t.Error("test user should have a privilege to maintain 1 role")
			}

			userListRes, err := adminClient.UserList(systemUserCtx, &rpc.RequestUserList{})
			if err != nil {
				t.Fatal(err)
			}
			if len(userListRes.Items) != 2 {
				t.Errorf("Expect 2 users: %d users", len(userListRes.Items))
			}
			userListRes, err = adminClient.UserList(systemUserCtx, &rpc.RequestUserList{Role: "test-admin"})
			if err != nil {
				t.Fatal(err)
			}
			if len(userListRes.Items) != 1 {
				t.Errorf("Expect 1 user: %d users", len(userListRes.Items))
			}
			userListRes, err = adminClient.UserList(systemUserCtx, &rpc.RequestUserList{ServiceAccount: true})
			if err != nil {
				t.Fatal(err)
			}
			if len(userListRes.Items) != 1 {
				t.Errorf("Expect 1 user: %d users", len(userListRes.Items))
			}

			toggleRes, err := adminClient.ToggleAdmin(systemUserCtx, &rpc.RequestToggleAdmin{Id: testUser.Id})
			if err != nil {
				t.Fatal(err)
			}
			if !toggleRes.Ok {
				t.Error("Expect return ok")
			}
			userListRes, err = adminClient.UserList(testUserCtx, &rpc.RequestUserList{})
			if err != nil {
				t.Fatal(err)
			}
			if len(userListRes.Items) != 2 {
				t.Errorf("Expect 2 users: %d users", len(userListRes.Items))
			}

			tokenRes, err := adminClient.TokenNew(systemUserCtx, &rpc.RequestTokenNew{Name: "test", UserId: testUser.Id})
			if err != nil {
				t.Fatal(err)
			}
			if tokenRes.Item.Value == "" {
				t.Error("Expect return a value")
			}
			getRes, err = adminClient.UserGet(systemUserCtx, &rpc.RequestUserGet{Id: testUser.Id, WithTokens: true})
			if err != nil {
				t.Fatal(err)
			}
			if len(getRes.User.Tokens) != 1 {
				t.Errorf("Expect 1 token: %d tokens", len(getRes.User.Tokens))
			}
		})
	})
}