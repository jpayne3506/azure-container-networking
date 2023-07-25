package npmconfig

import "github.com/Azure/azure-container-networking/npm/util"

const (
	defaultResyncPeriod         = 15
	defaultApplyMaxBatches      = 100
	defaultApplyInterval        = 500
	defaultMaxBatchedACLsPerPod = 30
	defaultMaxPendingNetPols    = 100
	defaultNetPolInterval       = 500
	defaultListeningPort        = 10091
	defaultGrpcPort             = 10092
	defaultGrpcServicePort      = 9002
	// ConfigEnvPath is what's used by viper to load config path
	ConfigEnvPath = "NPM_CONFIG"

	v1 = 1
	v2 = 2
)

// DefaultConfig is the guaranteed configuration NPM can run in out of the box
var DefaultConfig = Config{
	ResyncPeriodInMinutes: defaultResyncPeriod,

	ListeningPort:    defaultListeningPort,
	ListeningAddress: "0.0.0.0",

	Transport: GrpcServerConfig{
		Address:     "0.0.0.0",
		Port:        defaultGrpcPort,
		ServicePort: defaultGrpcServicePort,
	},

	WindowsNetworkName:          util.AzureNetworkName,
	ApplyMaxBatches:             defaultApplyMaxBatches,
	ApplyIntervalInMilliseconds: defaultApplyInterval,
	MaxBatchedACLsPerPod:        defaultMaxBatchedACLsPerPod,

	MaxPendingNetPols:            defaultMaxPendingNetPols,
	NetPolInvervalInMilliseconds: defaultNetPolInterval,

	Toggles: Toggles{
		EnablePrometheusMetrics: true,
		EnablePprof:             true,
		EnableHTTPDebugAPI:      true,
		EnableV2NPM:             true,
		PlaceAzureChainFirst:    util.PlaceAzureChainAfterKubeServices,
		ApplyIPSetsOnNeed:       false,
		// ApplyInBackground is currently used in Windows to apply the following in background: IPSets and NetPols for new/updated Pods
		ApplyInBackground: true,
		// NetPolInBackground is currently used in Linux to apply NetPol controller Add events in the background
		NetPolInBackground: true,
	},
}

type GrpcServerConfig struct {
	// Address is the address on which the gRPC server will listen
	Address string `json:"Address,omitempty"`
	// Port is the port on which the gRPC server will listen
	Port int `json:"Port,omitempty"`
	// ServicePort is the service port for the client to connect to the gRPC server
	ServicePort int `json:"ServicePort,omitempty"`
}

type Config struct {
	ResyncPeriodInMinutes int              `json:"ResyncPeriodInMinutes,omitempty"`
	ListeningPort         int              `json:"ListeningPort,omitempty"`
	ListeningAddress      string           `json:"ListeningAddress,omitempty"`
	Transport             GrpcServerConfig `json:"Transport,omitempty"`
	// WindowsNetworkName can be either 'azure' or 'Calico' (case sensitive).
	// It can also be the empty string, which results in the default value of 'azure'.
	WindowsNetworkName string `json:"WindowsNetworkName,omitempty"`
	// Apply options for Windows only. Relevant when ApplyInBackground is true.
	ApplyMaxBatches             int `json:"ApplyDataPlaneMaxBatches,omitempty"`
	ApplyIntervalInMilliseconds int `json:"ApplyDataPlaneMaxWaitInMilliseconds,omitempty"`
	// MaxBatchedACLsPerPod is the maximum number of ACLs that can be added to a Pod at once in Windows.
	// The zero value is valid.
	// A NetworkPolicy's ACLs are always in the same batch, and there will be at least one NetworkPolicy per batch.
	MaxBatchedACLsPerPod         int     `json:"MaxBatchedACLsPerPod,omitempty"`
	MaxPendingNetPols            int     `json:"MaxPendingNetPols,omitempty"`
	NetPolInvervalInMilliseconds int     `json:"NetPolInvervalInMilliseconds,omitempty"`
	Toggles                      Toggles `json:"Toggles,omitempty"`
}

type Toggles struct {
	EnablePrometheusMetrics bool
	EnablePprof             bool
	EnableHTTPDebugAPI      bool
	EnableV2NPM             bool
	PlaceAzureChainFirst    bool
	ApplyIPSetsOnNeed       bool
	// ApplyInBackground applies for Windows only
	ApplyInBackground bool
	// NetPolInBackground
	NetPolInBackground bool
}

type Flags struct {
	KubeConfigPath string `json:"KubeConfigPath"`
}

// NPMVersion returns 1 if EnableV2NPM=false and 2 otherwise
func (c Config) NPMVersion() int {
	if c.Toggles.EnableV2NPM {
		return v2
	}
	return v1
}
