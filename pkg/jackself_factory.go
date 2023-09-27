package pkg

import (
	"context"
	"fmt"

	jackselfapi "github.com/Yu-Jack/wrangler-test/apis/jackself/v1beta1"
	"github.com/Yu-Jack/wrangler-test/generated/controllers/jackself.testing"
	"k8s.io/client-go/rest"
)

type jackSelfFactory struct {
	mgmt *jackself.Factory
}

func NewJackSelfFactory(restConfig *rest.Config) Register {
	factory, err := jackself.NewFactoryFromConfig(restConfig)

	if err != nil {
		panic(err)
	}

	return &jackSelfFactory{
		mgmt: factory,
	}
}

func (j *jackSelfFactory) Setup() {
	cronJob := j.mgmt.Jackself().V1beta1().CronJob()
	cronJob.OnChange(context.Background(), "jackself-cronjob-change", func(id string, obj *jackselfapi.CronJob) (*jackselfapi.CronJob, error) {
		fmt.Println("jackself-obj.Spec.Foo: ", obj.Spec.Foo)
		obj.Spec.Foo = "fixed--jack-self!!!!"
		return cronJob.Update(obj)
	})
	if err := j.mgmt.Start(context.Background(), 50); err != nil {
		panic(err)
	}
}
