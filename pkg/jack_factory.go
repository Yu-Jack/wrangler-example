package pkg

import (
	"context"
	"fmt"
	"log"
	"sync"
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

	watchedEvents   map[string]string
	watchedEventMux sync.RWMutex
	watchedEvent    chan *k8sv1.Event
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
		jFactory:        jFactory,
		coreFactory:     coreFactory,
		recorder:        recorder,
		watchedEvents:   make(map[string]string),
		watchedEventMux: sync.RWMutex{},
		watchedEvent:    make(chan *k8sv1.Event),
	}

	go jf.WatchEvent(context.Background())

	return jf
}

func (j *jackFactory) appendWatchedEvents(name, namespace string) {
	exists := func() bool {
		j.watchedEventMux.RLock()
		defer j.watchedEventMux.RUnlock()
		_, ok := j.watchedEvents[name]
		return ok
	}()

	if exists {
		return
	}

	j.watchedEventMux.Lock()
	j.watchedEvents[name] = namespace
	j.watchedEventMux.Unlock()
}

func (j *jackFactory) deleteWatchedEvents(name string) {
	notExists := func() bool {
		j.watchedEventMux.RLock()
		defer j.watchedEventMux.RUnlock()
		_, ok := j.watchedEvents[name]
		return !ok
	}()

	if notExists {
		return
	}

	j.watchedEventMux.Lock()
	delete(j.watchedEvents, name)
	j.watchedEventMux.Unlock()
}

func (j *jackFactory) syncEvents() {
	j.watchedEventMux.RLock()
	defer j.watchedEventMux.RUnlock()

	time.Sleep(3 * time.Second) // simulate more time consuming

	if len(j.watchedEvents) == 0 {
		log.Println("no event to sync")
		return
	}

	for resourceName, namespace := range j.watchedEvents {
		eventList, err := j.coreFactory.Core().V1().Event().List(namespace, metav1.ListOptions{
			FieldSelector: fmt.Sprintf("involvedObject.name=%s", resourceName),
		})

		if err != nil {
			panic(err)
		}

		log.Println("---Start to print event---")
		var events []string
		for _, event := range eventList.Items {
			events = append(events, fmt.Sprintf("%s %s %s %s %s", event.Source, event.Type, event.Name, event.Reason, event.Message))
		}

		cronJob, err := j.jFactory.Jack().V1alpha1().CronJob().Get(namespace, resourceName, metav1.GetOptions{})
		if err != nil {
			panic(err)
		}
		cronJobDp := cronJob.DeepCopy()
		if cronJobDp.Annotations["events"] == fmt.Sprintf("%s", events) {
			log.Println("events is equal, skip it")
			continue
		}

		cronJob.Annotations["events"] = fmt.Sprintf("%s", events)

		if _, err = j.jFactory.Jack().V1alpha1().CronJob().Update(cronJob); err != nil {
			panic(err)
		}
		log.Println("---End to print event---")
	}
}

func (j *jackFactory) WatchEvent(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ticker.C:
			log.Println("start to sync events")
			j.syncEvents()
		case <-ctx.Done():
			return
		}
	}
}

func (j *jackFactory) Setup() {
	cronJob := j.jFactory.Jack().V1alpha1().CronJob()
	cronJob.OnChange(context.Background(), "jack-cronjob-change", func(id string, obj *api.CronJob) (*api.CronJob, error) {
		if obj == nil {
			return nil, nil
		}

		log.Println("jack-obj.Spec.Foo: ", obj.Spec.Foo)
		j.recorder.Event(obj, "Normal", "CronJob", obj.Spec.Foo)
		j.appendWatchedEvents(obj.Name, obj.Namespace)

		return obj, nil
	})

	cronJob.OnRemove(context.Background(), "jack-cronjob-change", func(id string, obj *api.CronJob) (*api.CronJob, error) {
		j.deleteWatchedEvents(obj.Name)
		return obj, nil
	})

	if err := j.jFactory.Start(context.Background(), 50); err != nil {
		panic(err)
	}
}
