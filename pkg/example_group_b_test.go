package pkg

import (
	"context"
	"testing"

	egbApi "github.com/Yu-Jack/wrangler-test/apis/example.group.b/v1beta1"
	"github.com/Yu-Jack/wrangler-test/generated/clientset/versioned/fake"
	"github.com/Yu-Jack/wrangler-test/generated/clientset/versioned/typed/example.group.b/v1beta1"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

type ExampleGroupBSuite struct {
	suite.Suite

	namespace string
	clientSet *fake.Clientset
	egbf      *exampleGroupBFactory
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(ExampleGroupBSuite))
}

func (s *ExampleGroupBSuite) SetupSuite() {
	s.namespace = "test"
	s.clientSet = fake.NewSimpleClientset()
	s.egbf = &exampleGroupBFactory{
		cronJobClient: FakeExampleGroupBClient(s.clientSet.ExampleV1beta1().CronJobs),
	}
}

func (s *ExampleGroupBSuite) TestOnChange() {
	description := "normal case"
	cronJob := &egbApi.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cronjob-01",
			Namespace: s.namespace,
		},
		Spec: egbApi.CronJobSpec{
			Foo: "this is cron job",
		},
	}

	_, err := s.clientSet.ExampleV1beta1().CronJobs(s.namespace).Create(context.TODO(), cronJob, metav1.CreateOptions{})
	s.Require().NoError(err, description)

	newCronJob, err := s.egbf.OnChange("", cronJob)
	s.Require().NoError(err, description)
	s.Require().Equal(fixedFoo, newCronJob.Spec.Foo, description)
}

type FakeExampleGroupBClient func(namespace string) v1beta1.CronJobInterface

func (f FakeExampleGroupBClient) Create(job *egbApi.CronJob) (*egbApi.CronJob, error) {
	//TODO implement me
	panic("implement me")
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
