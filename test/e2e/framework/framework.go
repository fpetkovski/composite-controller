package framework

import (
	"context"
	"log"
	"sync"

	"github.com/fpetkovski/composite-controller/pkg/hook"
	"github.com/fpetkovski/composite-controller/pkg/operator"
	"k8s.io/klog/v2/klogr"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Framework struct {
	Operator manager.Manager

	wg     sync.WaitGroup
	cancel context.CancelFunc
}

func NewFramework() *Framework {
	return &Framework{
		Operator: makeOperator(),
		wg:       sync.WaitGroup{},
	}
}

func (f *Framework) StartOperator() {
	ctx, cancel := context.WithCancel(context.Background())
	f.cancel = cancel
	f.wg.Add(1)
	go func() {
		if err := f.Operator.Start(ctx); err != nil {
			log.Fatal(err)
		}
		f.wg.Done()
	}()
}

func (f *Framework) StopOperator() {
	f.cancel()
	f.wg.Wait()
}

func makeOperator() manager.Manager {
	mapper := hook.Mapper{}
	cfg := controllerruntime.GetConfigOrDie()
	logger := klogr.New()
	op, err := operator.New(cfg, logger, mapper)
	if err != nil {
		log.Fatal()
	}

	return op
}
