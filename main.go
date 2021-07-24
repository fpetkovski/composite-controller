package main

import (
	"log"

	"github.com/fpetkovski/composite-controller/pkg/hook"

	"github.com/fpetkovski/composite-controller/pkg/operator"
	"k8s.io/klog/v2/klogr"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func main() {
	op, err := operator.New(config.GetConfigOrDie(), klogr.New(), hook.Mapper{})
	if err != nil {
		log.Fatal(err)
	}

	ctx := controllerruntime.SetupSignalHandler()
	if err := op.Start(ctx); err != nil {
		log.Fatal(err)
	}
}
