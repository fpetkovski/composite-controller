package main

import (
	"context"
	"github.com/fpetkovski/composite-controller/pkg/operator"
	"k8s.io/klog/v2/klogr"
	"log"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func main() {
	logger := klogr.New()
	controllerruntime.SetLogger(logger)

	cfg := config.GetConfigOrDie()
	op, err := operator.New(cfg, logger)
	if err != nil {
		log.Fatal(err)
	}

	if err := op.Start(context.Background()); err != nil {
		log.Fatal(err)
	}
}

