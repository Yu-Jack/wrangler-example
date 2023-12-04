package pkg

import (
	"context"
	"fmt"

	jackselfapi "github.com/Yu-Jack/wrangler-test/apis/jackself/v1beta1"
	"github.com/Yu-Jack/wrangler-test/generated/controllers/jackself.testing"
	"k8s.io/client-go/rest"
)

type jackSelfFactory struct {
	jsFactory *jackself.Factory
}

func NewJackSelfFactory(restConfig *rest.Config) Register {
	jsFactory, err := jackself.NewFactoryFromConfig(restConfig)

	if err != nil {
		panic(err)
	}

	return &jackSelfFactory{
		jsFactory: jsFactory,
	}
}

func (j *jackSelfFactory) Setup() {
	cronJob := j.jsFactory.Jackself().V1beta1().CronJob()
	cronJob.OnChange(context.Background(), "jackself-cronjob-change", func(id string, obj *jackselfapi.CronJob) (*jackselfapi.CronJob, error) {
		fmt.Println("jackself-obj.Spec.Foo: ", obj.Spec.Foo)
		obj.Spec.Foo = "fixed--jack-self!!!!"
		return cronJob.Update(obj)
	})

	if err := j.jsFactory.Start(context.Background(), 50); err != nil {
		panic(err)
	}
}
