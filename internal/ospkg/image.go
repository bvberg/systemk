package ospkg

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ForAllSecure/rootfs_builder/rootfs"
	"github.com/virtual-kubelet/systemk/internal/system"
	"github.com/virtual-kubelet/systemk/internal/unit"
)

// ImageManager manages unitfiles based on images.
type ImageManager struct{}

var _ Manager = (*ImageManager)(nil)

const (
	imageBaseRootPath = "/tmp/fleet"
)

func (p *ImageManager) Install(pkg, version string) (bool, error) {
	rootPath := GetImageRootDirectory(pkg, false)
	if _, err := os.Stat(rootPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(rootPath, os.ModePerm)
			if err != nil {
				return false, err
			}
		}
	}
	image := rootfs.PullableImage{
		Name:    pkg,
		Retries: 1,
		Spec: rootfs.Spec{
			Dest: rootPath,
			User: "root",
		},
	}

	pulledManifest, err := image.Pull()
	if err != nil {
		return false, fmt.Errorf("failed to pull image manifest: %+v", err)
	}

	// Extract rootfs
	err = pulledManifest.Extract()
	if err != nil {
		return false, fmt.Errorf("failed to extract rootfs: %+v", err)
	}

	basicPath := getBasicPath(pkg)
	_, err = os.Stat(basicPath)
	if os.IsNotExist(err) {
		_, err = os.Create(basicPath)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func (p *ImageManager) Unitfile(pkg string) (string, error) {
	return getBasicPath(pkg), nil
}

func getBasicPath(pkg string) string {
	basicPath := ""
	serviceName := prepareServiceName(pkg)
	// Determine OS
	systemID := system.ID()
	switch systemID {
	case "debian", "ubuntu":
		basicPath = debianSystemdUnitfilesPathPrefix + serviceName + unit.ServiceSuffix
	case "arch":
		basicPath = archlinuxSystemdUnitfilesPathPrefix + serviceName + unit.ServiceSuffix
	}
	return basicPath
}

// GetImageRootDirectory returns the root directory for the service to use.
// The fullPath flag is whether to include the suffix of the rootfs or not
func GetImageRootDirectory(pkg string, fullPath bool) string {
	base := fmt.Sprintf("%s/%s", imageBaseRootPath, prepareServiceName(pkg))
	if fullPath {
		base += "/rootfs"
	}
	return base
}

func prepareServiceName(pkg string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(pkg, "/", "-"), ":", "-"), ".", "-")
}
