package simple_hook

import (
	"github.com/fpetkovski/composite-controller/test/e2e/framework"
	"testing"
)

var f *framework.Framework

func TestMain(m *testing.M) {
	f = framework.NewFramework()
	f.StartOperator()
	m.Run()
	f.StopOperator()
}
