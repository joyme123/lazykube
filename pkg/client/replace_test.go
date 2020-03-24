package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_replace(t *testing.T) {
	globalConfig.RegisterReplaceStrategy("quay.io", PrefixReplace, "quay.azk8s.cn")
	globalConfig.RegisterReplaceStrategy("gcr.io", PrefixReplace, "gcr.azk8s.cn")
	globalConfig.RegisterReplaceStrategy("k8s.gcr.io", PrefixReplace, "registry.aliyuncs.com/google_containers")
	globalConfig.RegisterReplaceStrategy("docker.io", PrefixReplace, "dockerhub.azk8s.cn")
	globalConfig.RegisterReplaceStrategy("default", DefaultReplace, "dockerhub.azk8s.cn")

	testcases := []struct {
		image  string
		expect string
	}{
		{"quay.io/dexidp/dex:v2.10.0", "quay.azk8s.cn/dexidp/dex:v2.10.0"},
		{"gcr.io/dexidp/dex:v2.10.0", "gcr.azk8s.cn/dexidp/dex:v2.10.0"},
		{"k8s.gcr.io/dex:v2.10.0", "registry.aliyuncs.com/google_containers/dex:v2.10.0"},
		{"docker.io/dexidp/dex:v2.10.0", "dockerhub.azk8s.cn/dexidp/dex:v2.10.0"},
		{"dexidp/dex:v2.10.0", "dockerhub.azk8s.cn/dexidp/dex:v2.10.0"},
		{"dex:v2.10.0", "dockerhub.azk8s.cn/library/dex:v2.10.0"},
	}

	for row, testcase := range testcases {
		acturl := globalConfig.Replace(testcase.image)
		assert.Equal(t, testcase.expect, acturl, "testcase %d failed", row)
	}
}
