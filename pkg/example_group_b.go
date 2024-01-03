package pkg

import (
	"context"
	"fmt"

	egbApi "github.com/Yu-Jack/wrangler-test/apis/example.group.b/v1beta1"
	egb "github.com/Yu-Jack/wrangler-test/generated/controllers/example.group.b"
	"k8s.io/client-go/rest"
)

type exampleGroupBFactory struct {
	egbFactory *egb.Factory
}

func NewExampleGroupBFactory(restConfig *rest.Config) Register {
	egbFactory, err := egb.NewFactoryFromConfig(restConfig)

	if err != nil {
		panic(err)
	}

	return &exampleGroupBFactory{
		egbFactory: egbFactory,
	}
}

func (egbf *exampleGroupBFactory) Setup() {
	cronJob := egbf.egbFactory.Example().V1beta1().CronJob()
	cronJob.OnChange(context.Background(), "example.group.b-cronjob-change", func(id string, obj *egbApi.CronJob) (*egbApi.CronJob, error) {
		fmt.Println("example.group.b-obj.Spec.Foo: ", obj.Spec.Foo)
		obj.Spec.Foo = "fixed--jack-self!!!!"
		return cronJob.Update(obj)
	})

	if err := egbf.egbFactory.Start(context.Background(), 50); err != nil {
		panic(err)
	}
}
