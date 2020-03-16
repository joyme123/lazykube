package client

import (
	"strings"

	"github.com/prometheus/common/log"
)

// ReplaceMode 替换模式
type ReplaceMode string

const (
	// PrefixReplace 前缀替换
	PrefixReplace ReplaceMode = "prefix"

	// DefaultReplace dockerhub 替换
	DefaultReplace ReplaceMode = "default"
)

// ReplaceStrategy 替换策略
type ReplaceStrategy struct {
	Mode  ReplaceMode
	Value string
}

// ReplaceStrategies 镜像地址替换策略
type ReplaceStrategies map[string]ReplaceStrategy

var strategies ReplaceStrategies

// RegisterReplaceStrategy 注册镜像地址替换策略
func RegisterReplaceStrategy(replace string, mode ReplaceMode, to string) {
	if strategies == nil {
		strategies = make(ReplaceStrategies, 0)
	}
	strategies[replace] = ReplaceStrategy{
		Mode:  mode,
		Value: to,
	}
}

// 根据替换策略，替换镜像地址
func replace(image string) string {
	for replace, s := range strategies {

		switch s.Mode {
		case PrefixReplace:
			if strings.HasPrefix(image, replace) {
				newImg := strings.Replace(image, replace, s.Value, 1)
				log.Infof("old image: %s, new image: %s\n", image, newImg)
				return newImg
			}
			break
		case DefaultReplace:
			arr := strings.Split(image, "/")
			if len(arr) == 1 { // case similar: mysql:5.6
				newImg := s.Value + "/library/" + image
				log.Infof("old image: %s, new image: %s\n", image, newImg)
				return newImg
			} else if len(arr) == 2 && !strings.ContainsAny(arr[0], ".") { // case similar: joyme/mysql:5.6
				newImg := s.Value + "/" + image
				log.Infof("old image: %s, new image: %s\n", image, newImg)
				return newImg
			}
			break
		default:
			log.Warnf("不支持的替换模式, mode: %s, value: %s\n", s.Mode, s.Value)
		}
	}

	log.Infof("replace don't apply, image: %s", image)

	return image
}
