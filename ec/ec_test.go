package elementcontroller

import "testing"

type testApi interface {
	TestFunction() string
}

type testApiController struct{}

func newTestApiController() testApi {
	return &testApiController{}
}

func (*testApiController) TestFunction() string {
	return testSend()
}

func testSend() string {
	return getParentFunctionName()
}

func TestGetFunctionName(t *testing.T) {
	ctrl := newTestApiController()

	name := ctrl.TestFunction()
	if "TestFunction" != name {
		t.Errorf("Test Function Name Lookup Failed: %s", name)
	}
}
