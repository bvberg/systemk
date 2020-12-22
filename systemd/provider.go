package systemd

import (
	"fmt"
	"os"

	vkmanager "github.com/virtual-kubelet/node-cli/manager"
	"github.com/virtual-kubelet/node-cli/provider"
	"github.com/virtual-kubelet/systemk/pkg/manager"
	"github.com/virtual-kubelet/systemk/pkg/packages"
	"github.com/virtual-kubelet/systemk/pkg/system"
	"github.com/virtual-kubelet/virtual-kubelet/node/nodeutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// unitDir is where systemk stores the modified unit files.
var unitDir = "/var/run/systemk"

// P is a systemd provider.
type P struct {
	m   manager.Manager
	pkg packages.PackageManager
	rm  *vkmanager.ResourceManager
	w   *Watcher

	NodeInternalIP *corev1.NodeAddress
	NodeExternalIP *corev1.NodeAddress
	DaemonPort     int32

	ClusterDomain string
	Host          string
}

// New returns a new systemd provider.
func New(cfg provider.InitConfig) (*P, error) {
	if err := os.MkdirAll(unitDir, 0750); err != nil {
		return nil, err
	}
	m, err := manager.New(unitDir, false)
	if err != nil {
		return nil, err
	}
	p := &P{m: m}
	switch system.ID() {
	default:
		return nil, fmt.Errorf("unsupported system")
	case "debian", "ubuntu":
		p.pkg = new(packages.DebianPackageManager)

		// Just installed pre-requisites instead of pointing to the docs.
		klog.Infof("Installing %s, to prevent installed daemons from starting", "policyrcd-script-zg2")
		ok, err := p.pkg.Install("policyrcd-script-zg2", "")
		if err != nil {
			klog.Warningf("Failed to install %s, %s. Continuing anyway", "policyrcd-script-zg2", err)
		}
		if ok {
			klog.Infof("%s is already installed", "policyrcd-script-zg2")
		}

	case "arch":
		p.pkg = new(packages.ArchlinuxPackageManager)
	case "noop":
		p.pkg = new(packages.NoopPackageManager)
	}

	p.rm = cfg.ResourceManager
	p.DaemonPort = cfg.DaemonPort
	p.ClusterDomain = cfg.KubeClusterDomain

	if cfg.ConfigPath == "" {
		return p, nil
	}

	// parse the config yet again, to gain access to the Host field, so we can properly set the environment variables
	restConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: cfg.ConfigPath},
		&clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return p, err
	}

	p.Host = restConfig.Host

	clientset, err := nodeutil.ClientsetFromEnv(cfg.ConfigPath)
	if err != nil {
		return p, err
	}
	// Get new clientset.
	w := newWatcher(clientset)
	go func() {
		if err := w.run(p); err != nil {
			klog.Fatal(err)
		}
	}()
	p.w = w
	return p, nil
}

func (p *P) SetNodeIPs(nodeIP, nodeEIP string) {
	// Get the addresses.
	internal, external := nodeAddresses()
	if nodeIP != "" {
		p.NodeInternalIP = &corev1.NodeAddress{Address: nodeIP, Type: corev1.NodeInternalIP}
	} else {
		p.NodeInternalIP = internal
	}
	if nodeEIP != "" {
		p.NodeExternalIP = &corev1.NodeAddress{Address: nodeEIP, Type: corev1.NodeExternalIP}
	} else {
		p.NodeExternalIP = external
	}
	if p.NodeExternalIP == nil && p.NodeInternalIP == nil {
		klog.Fatal("Can not find internal or external IP address")
	}
	if p.NodeExternalIP == nil {
		p.NodeExternalIP = p.NodeInternalIP
	}
	if p.NodeInternalIP == nil {
		p.NodeInternalIP = p.NodeExternalIP
	}
}

var _ provider.Provider = new(P)
