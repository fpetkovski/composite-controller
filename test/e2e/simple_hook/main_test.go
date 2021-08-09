package simple_hook

import (
	"testing"

	"github.com/fpetkovski/composite-controller/test/e2e/framework"
)

var f *framework.Framework

func TestMain(m *testing.M) {
	main(m)
}

func main(m *testing.M) {
	f = framework.NewFramework()
	f.StartOperator()
	m.Run()
	f.StopOperator()
}
