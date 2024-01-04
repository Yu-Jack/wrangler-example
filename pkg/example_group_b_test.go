package pkg

import (
	"context"
	"testing"

	egbApi "github.com/Yu-Jack/wrangler-test/apis/example.group.b/v1beta1"
	"github.com/Yu-Jack/wrangler-test/generated/clientset/versioned/fake"
	"github.com/Yu-Jack/wrangler-test/generated/clientset/versioned/typed/example.group.b/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

func Test_OnChange(t *testing.T) {
	namespace := "test"
	clientSet := fake.NewSimpleClientset()

	egbf := exampleGroupBFactory{
		cronJobClient: FakeExampleGroupBClient(clientSet.ExampleV1beta1().CronJobs),
	}

	cronJob := &egbApi.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cronjob-01",
			Namespace: namespace,
		},
		Spec: egbApi.CronJobSpec{
			JobName: "jack-self-job-name~~~",
			Foo:     "this is cron job",
		},
	}
	_, err := clientSet.ExampleV1beta1().CronJobs(namespace).Create(context.TODO(), cronJob, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	newCronJob, err := egbf.OnChange("", cronJob)
	if err != nil {
		t.Fatal(err)
	}
	if newCronJob.Spec.Foo != fixedFoo {
		t.Fatal("newCronJob.Spec.Foo: ", newCronJob.Spec.Foo)
	}
}

type FakeExampleGroupBClient func(namespace string) v1beta1.CronJobInterface

func (f FakeExampleGroupBClient) Create(job *egbApi.CronJob) (*egbApi.CronJob, error) {
	return f(job.Namespace).Create(context.TODO(), job, metav1.CreateOptions{})
}

func (f FakeExampleGroupBClient) Update(job *egbApi.CronJob) (*egbApi.CronJob, error) {
	return f(job.Namespace).Update(context.TODO(), job, metav1.UpdateOptions{})
}

func (f FakeExampleGroupBClient) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	//TODO implement me
	panic("implement me")
}

func (f FakeExampleGroupBClient) Get(namespace, name string, options metav1.GetOptions) (*egbApi.CronJob, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeExampleGroupBClient) List(namespace string, opts metav1.ListOptions) (*egbApi.CronJobList, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeExampleGroupBClient) Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeExampleGroupBClient) Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *egbApi.CronJob, err error) {
	//TODO implement me
	panic("implement me")
}
