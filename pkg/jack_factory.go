package pkg

import (
	"context"
	"fmt"

	api "github.com/Yu-Jack/wrangler-test/apis/jack.jack.operator.test/v1alpha1"
	"github.com/Yu-Jack/wrangler-test/generated/controllers/jack.jack.operator.test"
	"k8s.io/client-go/rest"
)

type jackFactory struct {
	mgmt *jack.Factory
}

func NewJackFactory(restConfig *rest.Config) Register {
	factory, err := jack.NewFactoryFromConfig(restConfig)

	if err != nil {
		panic(err)
	}

	return &jackFactory{
		mgmt: factory,
	}
}

func (j *jackFactory) Setup() {
	cronJob := j.mgmt.Jack().V1alpha1().CronJob()
	cronJob.OnChange(context.Background(), "jack-cronjob-change", func(id string, obj *api.CronJob) (*api.CronJob, error) {
		fmt.Println("jack-obj.Spec.Foo: ", obj.Spec.Foo)
		obj.Spec.Foo = "fixed!!!!"
		return cronJob.Update(obj)
	})

	if err := j.mgmt.Start(context.Background(), 50); err != nil {
		panic(err)
	}
}
