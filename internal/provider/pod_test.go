package provider

import (
	"strings"
	"testing"

	"github.com/virtual-kubelet/systemk/internal/ospkg"
	corev1 "k8s.io/api/core/v1"
)

func TestNameSplitting(t *testing.T) {
	name := "systemk.default.openssh-server.openssh-server-container.service"
	if x := Container(name); x != "openssh-server-container" {
		t.Errorf("expected Image to be %s, got %s", "openssh-server-container", x)
	}
	if x := Pod(name); x != "openssh-server" {
		t.Errorf("expected Pod to be %s, got %s", "openssh-server", x)
	}
	if x := Name(name); x != "default.openssh-server" {
		t.Errorf("expected Name to be %s, got %s", "default.openssh-server", x)
	}
	if x := Namespace(name); x != "default" {
		t.Errorf("expected Namespace to be %s, got %s", "default", x)
	}
}
func TestArgs(t *testing.T) {
	manifest := ospkg.Manifest{
		Config: ospkg.Config{
			Entrypoint: []string{
				"/docker-entrypoint.sh",
			},
			CMD: []string{
				"nginx",
				"-g",
				"daemon off;",
			},
		},
	}
	c := corev1.Container{
		Command: []string{
			"/usr/sbin/nginx",
		},
		Args: []string{
			"-g",
			"daemon off;",
		},
	}

	result := ""
	result = strings.Join(manifest.Config.Entrypoint, " ")
	if len(manifest.Config.Entrypoint) > 0 {
		if len(c.Args) > 0 {
			result = strings.Join(manifest.Config.Entrypoint, " ") + " " + strings.Join(c.Args, " ")
		} else if len(manifest.Config.CMD) > 0 {
			result = strings.Join(manifest.Config.Entrypoint, " ") + " " + strings.Join(manifest.Config.CMD, " ")
		} else {
			result = strings.Join(manifest.Config.Entrypoint, " ")
		}
	}
	if len(c.Command) > 0 {
		if len(c.Args) > 0 {
			result = strings.Join(c.Command, " ") + " " + strings.Join(c.Args, " ")
		} else {
			result = strings.Join(c.Command, " ")
		}
	}
	t.Log(result)
}
