package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_replace(t *testing.T) {
	RegisterReplaceStrategy("quay.io", "quay.azk8s.cn")
	RegisterReplaceStrategy("gcr.io", "gcr.azk8s.cn")
	RegisterReplaceStrategy("k8s.gcr.io", "registry.aliyuncs.com/google_containers")

	testcases := []struct {
		image  string
		expect string
	}{
		{"quay.io/dexidp/dex:v2.10.0", "quay.azk8s.cn/dexidp/dex:v2.10.0"},
		{"gcr.io/dexidp/dex:v2.10.0", "gcr.azk8s.cn/dexidp/dex:v2.10.0"},
		{"k8s.gcr.io/dexidp/dex:v2.10.0", "registry.aliyuncs.com/google_containers/dexidp/dex:v2.10.0"},
	}

	for row, testcase := range testcases {
		acturl := replace(testcase.image)
		assert.Equal(t, testcase.expect, acturl, "testcase %d failed", row)
	}
}
