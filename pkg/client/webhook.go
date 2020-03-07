package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"k8s.io/api/admission/v1beta1"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	defaulter = runtime.ObjectDefaulter(runtimeScheme)
)

var ignoredNamespaces = []string{
	metav1.NamespaceSystem,
	metav1.NamespacePublic,
}

const (
	admissionWebhookAnnotationInjectKey = "lazykube.myway5.com/inject"
	admissionWebhookAnnotationStatusKey = "lazykube.myway5.com/status"
)

// WebhookServer 提供接口调用
type WebhookServer struct {
	server *http.Server
}

// NewWebhookServer 新建 webhook server
func NewWebhookServer(params *WhSvrParameters) (*WebhookServer, error) {

	pair, err := tls.LoadX509KeyPair(params.CertFile, params.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to load key pair: %v", err)
	}

	ws := &WebhookServer{
		server: &http.Server{
			Addr:      fmt.Sprintf(":%v", params.Port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
	}

	return ws, nil
}

// WhSvrParameters webhook server parameters
type WhSvrParameters struct {
	Port     int    // webhook server port
	CertFile string // path to the x509 certificate for https
	KeyFile  string // path to the x509 private key matching `CertFile`
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1beta1.AddToScheme(runtimeScheme)
	// defaulting with webhooks
	_ = v1.AddToScheme(runtimeScheme)

	RegisterReplaceStrategy("quay.io", "quay.azk8s.cn")
	RegisterReplaceStrategy("gcr.io", "gcr.azk8s.cn")
	RegisterReplaceStrategy("k8s.gcr.io", "registry.aliyuncs.com/google_containers")
}

// change init containers

// change containers
func createPatch(pod *corev1.Pod) ([]byte, error) {
	var patch = make([]patchOperation, 2)

	// replace initContainers
	for i := range pod.Spec.InitContainers {
		pod.Spec.InitContainers[i].Image = replace(pod.Spec.InitContainers[i].Image)
	}
	patch[0].Op = "replace"
	patch[0].Path = "/spec/initContainers"
	// new, err := json.Marshal(pod.Spec.InitContainers)
	// if err != nil {
	// 	return []byte{}, fmt.Errorf("marshal initcontainers error: %v", err)
	// }
	patch[0].Value = pod.Spec.InitContainers

	// replace containers
	for i := range pod.Spec.Containers {
		pod.Spec.Containers[i].Image = replace(pod.Spec.Containers[i].Image)
	}
	patch[1].Op = "replace"
	patch[1].Path = "/spec/containers"
	// new, err = json.Marshal(pod.Spec.Containers)
	// if err != nil {
	// 	return []byte{}, fmt.Errorf("marshal containers error: %v", err)
	// }
	patch[1].Value = pod.Spec.Containers

	return json.Marshal(patch)
}

func (whsrv *WebhookServer) mutate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	var pod corev1.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		log.Errorf("Could not unmarshal raw object: %v\n", err)

		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	log.Infof("AdmissionReview for Kind=%v, Namespace=%v (%v) UID=%v patchOperation=%v UserInfo=%v\n",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo)

	// TODO: determine whether to perform mutation

	patchBytes, err := createPatch(&pod)
	if err != nil {
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	log.Infof("AdmissionResponse: patch=%v\n", string(patchBytes))
	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func (whsrv *WebhookServer) serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if len(body) == 0 {
		log.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		log.Errorf("Can't decode body: %v", err)
		admissionResponse = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		admissionResponse = whsrv.mutate(&ar)
	}

	admissionReview := v1beta1.AdmissionReview{}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		log.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	log.Infof("Ready to write response ...")
	if _, err := w.Write(resp); err != nil {
		log.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}

// Start 启动服务
func (whsrv *WebhookServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", whsrv.serve)
	whsrv.server.Handler = mux

	if err := whsrv.server.ListenAndServeTLS("", ""); err != nil {
		return fmt.Errorf("Failed to listen and serve webhook server: %v", err)
	}

	return nil
}

// Shutdown 终止
func (whsrv *WebhookServer) Shutdown() {
	log.Info("Got OS shutdown signal, shutting down webhook server gracefully...")
	whsrv.server.Shutdown(context.Background())
}
