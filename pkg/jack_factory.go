package pkg

import (
	"context"
	"fmt"
	"strings"

	api "github.com/Yu-Jack/wrangler-test/apis/jack.jack.operator.test/v1alpha1"
	corev1 "github.com/Yu-Jack/wrangler-test/generated/controllers/core"
	"github.com/Yu-Jack/wrangler-test/generated/controllers/jack.jack.operator.test"
	"github.com/rancher/wrangler/pkg/leader"
	"github.com/sirupsen/logrus"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
)

type jackFactory struct {
	jFactory    *jack.Factory
	coreFactory *corev1.Factory
	recorder    record.EventRecorder
	clientSet   *kubernetes.Clientset
}

func NewJackFactory(restConfig *rest.Config) Register {
	jFactory, err := jack.NewFactoryFromConfig(restConfig)
	if err != nil {
		panic(err)
	}

	coreFactory, err := corev1.NewFactoryFromConfig(restConfig)
	if err != nil {
		panic(err)
	}

	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		panic(err)
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(logrus.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: clientSet.CoreV1().Events("jack-operator-test-system")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, k8sv1.EventSource{})

	jf := &jackFactory{
		jFactory:    jFactory,
		coreFactory: coreFactory,
		recorder:    recorder,
		clientSet:   clientSet,
	}

	return jf
}
func (j *jackFactory) Setup() {
	cronJob := j.jFactory.Jack().V1alpha1().CronJob()
	cronJob.OnChange(context.Background(), "jack-cronjob-change", func(id string, obj *api.CronJob) (*api.CronJob, error) {
		if obj == nil {
			return obj, nil
		}

		apiVersion, _ := obj.GroupVersionKind().ToAPIVersionAndKind()
		apiPath := fmt.Sprintf("/apis/%s/namespaces/%s/%s/%s", apiVersion, obj.Namespace, api.CronJobResourceName, obj.Name)
		apiPath = strings.ToLower(apiPath)
		fmt.Println(apiPath)

		res, err := j.clientSet.RESTClient().Delete().AbsPath(apiPath).DoRaw(context.Background())
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(res))

		return obj, nil
	})

	leader.RunOrDie(context.Background(), "", "jack-controller", j.clientSet, func(cb context.Context) {
		if err := j.jFactory.Start(context.Background(), 50); err != nil {
			panic(err)
		}
	})
}
