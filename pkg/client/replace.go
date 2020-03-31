package client

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v1"
	corev1 "k8s.io/api/core/v1"
)

var globalConfig = &LazykubeConfig{
	ReplaceStrategies: make([]ReplaceStrategy, 0),
}

// LazykubeConfig 配置信息
type LazykubeConfig struct {
	ReplaceStrategies []ReplaceStrategy `yaml:"replaceStrategies"`
}

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
	Case  string      `yaml:"case"`
	Mode  ReplaceMode `yaml:"mode"`
	Value string      `yaml:"value"`
}

// RegisterReplaceStrategy 注册镜像地址替换策略
func (c *LazykubeConfig) RegisterReplaceStrategy(replace string, mode ReplaceMode, to string) {
	for i := range c.ReplaceStrategies {
		if c.ReplaceStrategies[i].Case == replace {
			c.ReplaceStrategies[i].Mode = mode
			c.ReplaceStrategies[i].Value = to
			return
		}
	}
	c.ReplaceStrategies = append(c.ReplaceStrategies, ReplaceStrategy{
		Case:  replace,
		Mode:  mode,
		Value: to,
	})
}

// UpdateConfig 更新替换策略
func (c *LazykubeConfig) UpdateConfig(cm *corev1.ConfigMap) error {
	data := cm.Data["config"]

	log.Info("new config:", data)
	var tmpConfig LazykubeConfig
	err := yaml.Unmarshal([]byte(data), &tmpConfig)
	if err != nil {
		return err
	}

	*c = tmpConfig
	return nil
}

// Replace 根据替换策略，替换镜像地址
func (c *LazykubeConfig) Replace(image string) string {
	for _, s := range c.ReplaceStrategies {
		replace := s.Case
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
			if len(arr) == 1 { // case : mysql:5.6
				newImg := s.Value + "/library/" + image
				log.Infof("old image: %s, new image: %s\n", image, newImg)
				return newImg
			} else if len(arr) == 2 && !strings.ContainsAny(arr[0], ".") { // case: joyme/mysql:5.6
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
