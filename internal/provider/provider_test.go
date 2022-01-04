package provider

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/virtual-kubelet/systemk/internal/kubernetes"
	"github.com/virtual-kubelet/systemk/internal/ospkg"
	"github.com/virtual-kubelet/systemk/internal/unit"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
)

const dir = "../testdata/provider"

func TestProviderPodSpecUnits(t *testing.T) {
	p := initProvider()
	testFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatalf("could not read %s: %q", dir, err)
	}
	for _, f := range testFiles {
		if f.IsDir() {
			continue
		}

		if filepath.Ext(f.Name()) != ".yaml" {
			continue
		}
		base := f.Name()[:len(f.Name())-5]
		t.Run("Testing: "+base, func(t *testing.T) {
			testPodSpecUnit(t, p, base)
		})
	}
}

func TestCreatePod(t *testing.T) {
	p := initProvider()
	p.config.ExtractImage = true
	p.pkgManager = &ospkg.ImageManager{}
	pod := corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name: "nginx",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nginx",
					Image: "nginx",
					Command: []string{
						"nginx",
					},
					Args: []string{
						"-g",
						"daemon off;",
					},
				},
			},
		},
	}
	err := p.CreatePod(context.TODO(), &pod)
	if err != nil {
		t.Error(err)
		return
	}
}

func initProvider() *p {
	log = &noopLogger{}
	p := new(p)
	p.pkgManager = &ospkg.NoopManager{}
	p.unitManager, _ = unit.NewMockManager()
	p.config = &Opts{
		NodeName:       "localhost",
		NodeInternalIP: []byte{192, 168, 1, 1},
		NodeExternalIP: []byte{172, 16, 0, 1},
	}

	p.podResourceManager = kubernetes.NewPodResourceWatcher(informers.NewSharedInformerFactory(nil, 0))
	return p
}

func testPodSpecUnit(t *testing.T, p *p, base string) {
	yamlFile := filepath.Join(dir, base+".yaml")
	pod, err := kubernetes.PodFromFile(yamlFile)
	if err != nil {
		t.Error(err)
		return
	}
	unitFile := filepath.Join(dir, base+".units")
	unit, _ := ioutil.ReadFile(unitFile)

	if err := p.CreatePod(context.TODO(), pod); err != nil {
		t.Errorf("failed to call CreatePod: %v", err)
		return
	}
	got := ""
	for _, c := range pod.Spec.Containers {
		name := podToUnitName(pod, c.Name)
		got += p.unitManager.Unit(name)
	}

	if got != string(unit) {
		t.Errorf("got unexpected result: %s", diff.LineDiff(got, string(unit)))
	}
}
