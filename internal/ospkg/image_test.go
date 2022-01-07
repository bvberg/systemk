package ospkg

import (
	"strings"
	"testing"
)

func TestImage(t *testing.T) {
	d := new(ImageManager)
	ok, err := d.Install("docker.io/library/busybox:latest", "")
	if err != nil || !ok {
		// not installed
		t.Error(err)
	}
	path, err := d.Unitfile("docker.io/library/busybox:latest")
	if err != nil || !ok {
		// not installed
		t.Error(err)
	}
	if path == "" {
		t.Errorf("invalid path: %s", path)
	}
}

func TestGetRootDirectory(t *testing.T) {
	pkg := "docker.io/library/busybox:latest"
	full := GetImageRootDirectory(pkg, true)
	if !strings.Contains(full, "rootfs") {
		t.Error("path should contain the rootfs suffix")
	}
	base := GetImageRootDirectory(pkg, false)
	if strings.Contains(base, "rootfs") {
		t.Error("path should not contain the rootfs suffix")
	}
}

func TestGetImageManifest(t *testing.T) {
	pkg := "docker.io/library/busybox:latest"
	config, err := GetImageManifest(GetImageRootDirectory(pkg, false))
	if err != nil {
		t.Error(err)
	}
	if config == nil {
		t.Error("config is null")
	}
	t.Log(config)
}
