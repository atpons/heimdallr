package main

import (
	"context"
	"flag"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog"

	"github.com/f110/lagrangian-proxy/operator/pkg/controllers"
	"github.com/f110/lagrangian-proxy/operator/pkg/signals"
)

func main() {
	id := ""
	metricsAddr := ""
	enableLeaderElection := false
	leaseLockName := ""
	leaseLockNamespace := ""
	clusterDomain := ""
	dev := false
	fs := flag.NewFlagSet("operator", flag.ExitOnError)
	fs.StringVar(&id, "id", uuid.New().String(), "the holder identity name")
	fs.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	fs.BoolVar(&enableLeaderElection, "enable-leader-election", enableLeaderElection,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	fs.StringVar(&leaseLockName, "lease-lock-name", "", "the lease lock resource name")
	fs.StringVar(&leaseLockNamespace, "lease-lock-namespace", "", "the lease lock resource namespace")
	fs.StringVar(&clusterDomain, "cluster-domain", clusterDomain, "Cluster domain")
	fs.BoolVar(&dev, "dev", dev, "development mode")
	klog.InitFlags(fs)
	if err := fs.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	signals.SetupSignalHandler(cancelFunc)

	cfg, err := clientcmd.BuildConfigFromFlags("", "/home/dexter/.kube/config")
	if err != nil {
		os.Exit(1)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		os.Exit(1)
	}

	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      leaseLockName,
			Namespace: leaseLockNamespace,
		},
		Client: kubeClient.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: id,
		},
	}
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   30 * time.Second,
		RenewDeadline:   15 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				c, err := controllers.New(ctx, kubeClient, cfg)
				if err != nil {
					klog.Error(err)
					os.Exit(1)
				}

				e, err := controllers.NewEtcdController(ctx, kubeClient, cfg, clusterDomain, dev)
				if err != nil {
					klog.Error(err)
					os.Exit(1)
				}

				var wg sync.WaitGroup
				wg.Add(1)
				go func() {
					defer wg.Done()

					e.Run(ctx, 1)
				}()

				wg.Add(1)
				go func() {
					defer wg.Done()

					c.Run(ctx, 1)
				}()

				wg.Wait()
			},
			OnStoppedLeading: func() {
				klog.Infof("leader lost: %s", id)
				os.Exit(0)
			},
			OnNewLeader: func(identity string) {
				if identity == id {
					return
				}
				klog.Infof("new leader elected: %s", identity)
			},
		},
	})
}