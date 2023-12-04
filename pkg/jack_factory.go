package pkg

import (
	"context"
	"fmt"
	"time"

	api "github.com/Yu-Jack/wrangler-test/apis/jack.jack.operator.test/v1alpha1"
	corev1 "github.com/Yu-Jack/wrangler-test/generated/controllers/core"
	"github.com/Yu-Jack/wrangler-test/generated/controllers/jack.jack.operator.test"
	"github.com/sirupsen/logrus"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	watchedEvents map[string]struct{}
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
		jFactory:      jFactory,
		coreFactory:   coreFactory,
		recorder:      recorder,
		watchedEvents: make(map[string]struct{}),
	}

	go jf.WatchEvent(context.Background(), "jack-operator-test-system")

	return jf
}

func (j *jackFactory) appendWatchedEvents(name string) {
	j.watchedEvents[name] = struct{}{}
}

func (j *jackFactory) WatchEvent(ctx context.Context, namespace string) {

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Finish Watching")
			return
		case <-time.After(3 * time.Second):
			for name, _ := range j.watchedEvents {
				eventList, err := j.coreFactory.Core().V1().Event().List(namespace, metav1.ListOptions{
					FieldSelector: fmt.Sprintf("involvedObject.name=%s", name),
				})
				if err != nil {
					panic(err)
				}

				fmt.Println("---Start to print event---")
				for _, event := range eventList.Items {
					fmt.Println(event.Source, event.Type, event.Name, event.Reason, event.Message)
				}
				fmt.Println("---End to print event---")
			}
		}
	}
}

func (j *jackFactory) Setup() {
	cronJob := j.jFactory.Jack().V1alpha1().CronJob()
	cronJob.OnChange(context.Background(), "jack-cronjob-change", func(id string, obj *api.CronJob) (*api.CronJob, error) {
		if obj == nil {
			return nil, nil
		}

		fmt.Println("jack-obj.Spec.Foo: ", obj.Spec.Foo)
		obj.Spec.Foo = "fixed!!!!"
		j.recorder.Event(obj, "Normal", "CronJob", "fixed!!!!")
		j.appendWatchedEvents(obj.Name)

		return cronJob.Update(obj)
	})

	if err := j.jFactory.Start(context.Background(), 50); err != nil {
		panic(err)
	}
}
