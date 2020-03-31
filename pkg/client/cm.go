package client

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// ConfigMapWatcher configmap 资源监控
type ConfigMapWatcher struct {
	k8sCli    kubernetes.Interface
	namespace string
	name      string
	config    *LazykubeConfig
}

// NewConfigMapWatcher 实例化 configmap 监控
func NewConfigMapWatcher(k8sCli kubernetes.Interface, namespace string, name string, config *LazykubeConfig) *ConfigMapWatcher {
	return &ConfigMapWatcher{
		k8sCli:    k8sCli,
		name:      name,
		namespace: namespace,
		config:    config,
	}
}

// SyncConfig 从 configmap 中同步配置
func (w *ConfigMapWatcher) SyncConfig() error {
	cmClient := w.k8sCli.CoreV1().ConfigMaps(w.namespace)
	cm, err := cmClient.Get(w.name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("SyncConfig error: %v", err)
	}
	return w.config.UpdateConfig(cm)
}

// Run 运行
func (w *ConfigMapWatcher) Run(ctx context.Context) {
	fieldSelector := fields.ParseSelectorOrDie(fmt.Sprintf("metadata.name=%s", w.name))

	listWatch := cache.NewListWatchFromClient(
		w.k8sCli.CoreV1().RESTClient(),
		"configmaps",
		w.namespace,
		fieldSelector,
	)

	_, controller := cache.NewInformer(listWatch, &corev1.ConfigMap{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if cm, ok := obj.(*corev1.ConfigMap); ok {
				log.Infof("Detected ConfigMap add. Updating the controller config.")
				err := w.config.UpdateConfig(cm)
				if err != nil {
					log.Errorf("Update config failed due to: %v", err)
				}
			}
		},

		UpdateFunc: func(old, new interface{}) {
			oldCM := old.(*corev1.ConfigMap)
			newCM := new.(*corev1.ConfigMap)
			if oldCM.ResourceVersion == newCM.ResourceVersion {
				return
			}
			log.Infof("Detected ConfigMap update. Updating the controller config.")
			err := w.config.UpdateConfig(newCM)
			if err != nil {
				log.Errorf("Update of config failed due to: %v", err)
			}
		},
	})

	controller.Run(ctx.Done())
}
