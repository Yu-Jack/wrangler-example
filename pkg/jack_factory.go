package pkg

import (
	"context"
	"fmt"
	"reflect"
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

	watchedEvents    map[string]struct{}
	oldWatchedEvents map[string]struct{}
	watchedEventMux  sync.Mutex
	watchedEvent     chan *k8sv1.Event
	watcherCtx       context.Context
	watcherCancel    context.CancelFunc
	watcherWG        sync.WaitGroup
	watcherCounter   int
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
		jFactory:         jFactory,
		coreFactory:      coreFactory,
		recorder:         recorder,
		watchedEvents:    make(map[string]struct{}),
		oldWatchedEvents: make(map[string]struct{}),
		watchedEventMux:  sync.Mutex{},
		watchedEvent:     make(chan *k8sv1.Event),
		watcherWG:        sync.WaitGroup{},
	}

	jf.watcherCtx, jf.watcherCancel = context.WithCancel(context.Background())

	go jf.WatchEvent()

	return jf
}

func (j *jackFactory) appendWatchedEvents(name string) {
	j.watchedEventMux.Lock()
	j.watchedEvents[name] = struct{}{}
	j.fanInEvents()
	j.watchedEventMux.Unlock()

	j.oldWatchedEvents = j.watchedEvents
}

func (j *jackFactory) deleteWatchedEvents(name string) {
	j.watchedEventMux.Lock()
	delete(j.watchedEvents, name)
	j.fanInEvents()
	j.watchedEventMux.Unlock()

	j.oldWatchedEvents = j.watchedEvents
}

func (j *jackFactory) fanInEvents() {
	if reflect.DeepEqual(j.oldWatchedEvents, j.watchedEvents) {
		return
	}

	j.watcherCancel()
	j.watcherWG.Wait()

	j.watcherCtx, j.watcherCancel = context.WithCancel(context.Background())
	j.watcherWG = sync.WaitGroup{}

	go func() {
		for event, _ := range j.watchedEvents {
			watcher, err := j.coreFactory.Core().V1().Event().Watch("jack-operator-test-system", metav1.ListOptions{
				FieldSelector: fmt.Sprintf("involvedObject.name=%s", event),
			})

			if err != nil {
				panic(err)
			}

			select {
			case <-j.watcherCtx.Done():
				watcher.Stop()
				return
			default:
			}

			j.watcherWG.Add(1)
			j.watcherCounter++

			fmt.Println("Start watcher: ", j.watcherCounter)

			go func() {
				defer j.watcherWG.Done()
				defer watcher.Stop()

				for {
					select {
					case <-time.After(3 * time.Second):
						// timeout case
						return
					case <-j.watcherCtx.Done():
						// cancel case
						return
					case result, ok := <-watcher.ResultChan():
						if !ok {
							return
						}
						e := result.Object.(*k8sv1.Event)
						j.watchedEvent <- e
					}
				}
			}()
		}

		j.watcherWG.Wait()
	}()
}

func (j *jackFactory) WatchEvent() {
	for event := range j.watchedEvent {
		fmt.Println("---Start to print event---")
		fmt.Println(event.Source, event.Type, event.Name, event.Reason, event.Message)
		fmt.Println("---End to print event---")
	}
}

func (j *jackFactory) Setup() {
	cronJob := j.jFactory.Jack().V1alpha1().CronJob()
	cronJob.OnChange(context.Background(), "jack-cronjob-change", func(id string, obj *api.CronJob) (*api.CronJob, error) {
		if obj == nil {
			return nil, nil
		}

		fmt.Println("jack-obj.Spec.Foo: ", obj.Spec.Foo)
		j.recorder.Event(obj, "Normal", "CronJob", obj.Spec.Foo)
		j.appendWatchedEvents(obj.Name)

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
