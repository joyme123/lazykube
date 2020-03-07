package client

import (
	"strings"

	"github.com/prometheus/common/log"
)

// ReplaceStrategy 镜像地址替换策略
type ReplaceStrategy map[string]string

var strategies ReplaceStrategy

// RegisterReplaceStrategy 注册镜像地址替换策略
func RegisterReplaceStrategy(replace, to string) {
	if strategies == nil {
		strategies = make(ReplaceStrategy, 0)
	}
	strategies[replace] = to
}

// 根据替换策略，替换镜像地址
func replace(image string) string {
	for replace, to := range strategies {
		if strings.HasPrefix(image, replace) {
			newImg := strings.Replace(image, replace, to, 1)
			log.Infof("old image: %s, new image: %s", image, newImg)
			return newImg
		}
	}

	log.Infof("replace don't apply, image: %s", image)

	return image
}
