package pkg

import (
	"context"
	"fmt"

	egbApi "github.com/Yu-Jack/wrangler-test/apis/example.group.b/v1beta1"
	egb "github.com/Yu-Jack/wrangler-test/generated/controllers/example.group.b"
	"github.com/Yu-Jack/wrangler-test/generated/controllers/example.group.b/v1beta1"
	"k8s.io/client-go/rest"
)

type exampleGroupBFactory struct {
	egbFactory        *egb.Factory
	cronJobClient     v1beta1.CronJobClient
	cronJobController v1beta1.CronJobController
}

func NewExampleGroupBFactory(restConfig *rest.Config) Register {
	egbFactory, err := egb.NewFactoryFromConfig(restConfig)
	cronJobClient := egbFactory.Example().V1beta1().CronJob()

	if err != nil {
		panic(err)
	}

	return &exampleGroupBFactory{
		egbFactory:        egbFactory,
		cronJobClient:     cronJobClient,
		cronJobController: cronJobClient,
	}
}

func (egbf *exampleGroupBFactory) Setup() {
	egbf.cronJobController.OnChange(context.Background(), "example.group.b-cronjob-change", egbf.OnChange)

	if err := egbf.egbFactory.Start(context.Background(), 50); err != nil {
		panic(err)
	}
}

func (egbf *exampleGroupBFactory) OnChange(id string, obj *egbApi.CronJob) (*egbApi.CronJob, error) {
	fmt.Println("example.group.b-obj.Spec.Foo: ", obj.Spec.Foo)
	obj.Spec.Foo = "fixed--jack-self!!!!"
	return egbf.cronJobClient.Update(obj)
}
