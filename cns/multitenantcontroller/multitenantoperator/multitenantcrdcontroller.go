package multitenantoperator

import (
	"context"
	"errors"
	"os"
	"sync"

	"github.com/Azure/azure-container-networking/cns/logger"
	"github.com/Azure/azure-container-networking/cns/multitenantcontroller"
	"github.com/Azure/azure-container-networking/cns/restserver"
	"github.com/Azure/azure-container-networking/crd"
	ncapi "github.com/Azure/azure-container-networking/crd/multitenantnetworkcontainer/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

const (
	nodeNameEnvVar    = "NODENAME"
	prometheusAddress = "0" // 0 means disabled
)

var _ (multitenantcontroller.RequestController) = (*requestController)(nil)

// requestController operates multi-tenant CRD.
type requestController struct {
	mgr        manager.Manager // Manager starts the reconcile loop which watches for crd status changes
	KubeClient client.Client   // KubeClient is a cached client which interacts with API server
	nodeName   string          // name of node running this program
	Reconciler *multiTenantCrdReconciler
	Started    bool
	lock       sync.Mutex
}

// New creates a new multi-tenant CRD operator.
func New(restService *restserver.HTTPRestService, kubeconfig *rest.Config) (*requestController, error) {
	// Check that logger package has been initialized.
	if logger.Log == nil {
		return nil, errors.New("Must initialize logger before calling")
	}

	// Check that NODENAME environment variable is set. NODENAME is name of node running this program.
	nodeName := os.Getenv(nodeNameEnvVar)
	if nodeName == "" {
		return nil, errors.New("Must declare " + nodeNameEnvVar + " environment variable.")
	}

	// Add client-go scheme to runtime scheme so manager can recognize it.
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, errors.New("Error adding client-go scheme to runtime scheme")
	}

	// Add CRD scheme to runtime sheme so manager can recognize it.
	if err := ncapi.AddToScheme(scheme); err != nil {
		return nil, errors.New("Error adding NetworkContainer scheme to runtime scheme")
	}

	// Create manager for multiTenantController.
	mgr, err := ctrl.NewManager(kubeconfig, ctrl.Options{
		Scheme:  scheme,
		Metrics: ctrlmetrics.Options{BindAddress: prometheusAddress},
	})
	if err != nil {
		logger.Errorf("Error creating new multiTenantController: %v", err)
		return nil, err
	}

	// Create multiTenantCrdReconciler
	reconciler := &multiTenantCrdReconciler{
		KubeClient:     mgr.GetClient(),
		NodeName:       nodeName,
		CNSRestService: restService,
	}
	if err := reconciler.SetupWithManager(mgr); err != nil {
		logger.Errorf("Error setting up new multiTenantCrdReconciler: %v", err)
		return nil, err
	}

	// Create the multiTenantController
	return &requestController{
		mgr:        mgr,
		KubeClient: mgr.GetClient(),
		nodeName:   nodeName,
		Reconciler: reconciler,
	}, nil
}

// StartMultiTenantController starts the Reconciler loop which watches for CRD status updates.
func (rc *requestController) Start(ctx context.Context) error {
	logger.Printf("Starting MultiTenantController")

	// Setting the started state
	rc.lock.Lock()
	rc.Started = true
	rc.lock.Unlock()

	logger.Printf("Starting reconcile loop")
	if err := rc.mgr.Start(ctx); err != nil {
		if crd.IsNotDefined(err) {
			logger.Errorf("multi-tenant CRD is not defined on cluster, starting reconcile loop failed: %v", err)
			os.Exit(1)
		}

		return err
	}

	return nil
}

// IsStarted return if RequestController is started
func (rc *requestController) IsStarted() bool {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	return rc.Started
}
