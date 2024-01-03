package pkg

import (
	"flag"
	"path/filepath"
	"time"

	api "github.com/Yu-Jack/wrangler-test/apis/example.group.a/v1alpha1"
	jackselfapi "github.com/Yu-Jack/wrangler-test/apis/example.group.b/v1beta1"
	controllergen "github.com/rancher/wrangler/pkg/controller-gen"
	"github.com/rancher/wrangler/pkg/controller-gen/args"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Register interface {
	Setup()
}

func Setup() {
	setupTypes()

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	restConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	registers := []Register{
		NewExampleGroupAFactory(restConfig),
		NewExampleGroupBFactory(restConfig),
	}

	for _, r := range registers {
		r.Setup()
	}

	time.Sleep(100 * time.Minute)
}

func setupTypes() {
	controllergen.Run(args.Options{
		OutputPackage: "github.com/Yu-Jack/wrangler-test/generated",
		Boilerplate:   "scripts/boilerplate.txt",
		Groups: map[string]args.Group{
			"example.group.a": {
				PackageName: "example.group.a",
				Types: []interface{}{
					api.CronJob{},
				},
				GenerateTypes:   true,
				GenerateClients: true,
			},
			"example.group.b": {
				PackageName: "example.group.b",
				Types: []interface{}{
					jackselfapi.CronJob{},
				},
				GenerateTypes:   true,
				GenerateClients: true,
			},
			corev1.GroupName: {
				Types: []interface{}{
					corev1.Event{},
				},
				InformersPackage: "k8s.io/client-go/informers",
				ClientSetPackage: "k8s.io/client-go/kubernetes",
				ListersPackage:   "k8s.io/client-go/listers",
			},
		},
	})
}
