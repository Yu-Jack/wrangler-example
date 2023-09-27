package pkg

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	api "github.com/Yu-Jack/wrangler-test/apis/jack.jack.operator.test/v1alpha1"
	jack "github.com/Yu-Jack/wrangler-test/generated/controllers/jack.jack.operator.test"
	controllergen "github.com/rancher/wrangler/pkg/controller-gen"
	"github.com/rancher/wrangler/pkg/controller-gen/args"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func Setup() {
	setupTypes()

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	restConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	_ = restConfig

	mgmt, err := jack.NewFactoryFromConfig(restConfig)
	if err != nil {
		panic(err)
	}

	cronJob := mgmt.Jack().V1alpha1().CronJob()

	cronJob.OnChange(context.Background(), "jack-cronjob-change", func(id string, obj *api.CronJob) (*api.CronJob, error) {
		fmt.Println("obj.Spec.Foo: ", obj.Spec.Foo)
		obj.Spec.Foo = "fixed!!!!"
		return cronJob.Update(obj)
	})

	mgmt.Start(context.Background(), 50)
	time.Sleep(100 * time.Second)
}

func setupTypes() {
	controllergen.Run(args.Options{
		OutputPackage: "github.com/Yu-Jack/wrangler-test/generated",
		Boilerplate:   "scripts/boilerplate.txt",
		Groups: map[string]args.Group{
			"jack.jack.operator.test": {
				PackageName: "jack.jack.operator.test",
				Types: []interface{}{
					api.CronJob{},
				},
				GenerateTypes: true,
			},
		},
	})
}
