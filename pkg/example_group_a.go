package pkg

import (
	"context"
	"fmt"
	"strings"

	api "github.com/Yu-Jack/wrangler-test/apis/example.group.a/v1alpha1"
	ega "github.com/Yu-Jack/wrangler-test/generated/controllers/example.group.a"
	v1alpha1 "github.com/Yu-Jack/wrangler-test/generated/controllers/example.group.a/v1alpha1"
	"github.com/rancher/wrangler/pkg/leader"
	"github.com/sirupsen/logrus"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
)

type exampleGroupAFactory struct {
	egaFactory        *ega.Factory
	recorder          record.EventRecorder
	clientSet         *kubernetes.Clientset
	cronJobClient     v1alpha1.CronJobClient
	cronJobController v1alpha1.CronJobController
}

func NewExampleGroupAFactory(restConfig *rest.Config) Register {
	egaFactory, err := ega.NewFactoryFromConfig(restConfig)
	if err != nil {
		panic(err)
	}

	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		panic(err)
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(logrus.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: clientSet.CoreV1().Events("example-group-a-operator-test-system")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, k8sv1.EventSource{})

	cronJobs := egaFactory.Example().V1alpha1().CronJob()

	egaf := &exampleGroupAFactory{
		egaFactory:        egaFactory,
		recorder:          recorder,
		clientSet:         clientSet,
		cronJobClient:     cronJobs,
		cronJobController: cronJobs,
	}

	cronJobs.OnChange(context.Background(), "example-group-a-cronjob-change", egaf.OnChange)

	return egaf
}
func (egaf *exampleGroupAFactory) Setup() {
	leader.RunOrDie(context.Background(), "", "example-a-controller", egaf.clientSet, func(cb context.Context) {
		if err := egaf.egaFactory.Start(context.Background(), 50); err != nil {
			panic(err)
		}
	})
}

func (egaf *exampleGroupAFactory) OnChange(id string, obj *api.CronJob) (*api.CronJob, error) {
	if obj == nil {
		return obj, nil
	}

	apiVersion, _ := obj.GroupVersionKind().ToAPIVersionAndKind()
	apiPath := fmt.Sprintf("/apis/%s/namespaces/%s/%s/%s", apiVersion, obj.Namespace, api.CronJobResourceName, obj.Name)
	apiPath = strings.ToLower(apiPath)
	fmt.Println(apiPath)

	res, err := egaf.clientSet.RESTClient().Delete().AbsPath(apiPath).DoRaw(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(res))

	return obj, nil
}
