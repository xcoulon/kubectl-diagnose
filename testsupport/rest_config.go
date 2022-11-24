package testsupport

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

func NewConfig(host, apiPath string) *rest.Config {
	return &rest.Config{
		Host:    host,
		APIPath: apiPath,
		// These fields need to be set when using the REST client ¯\_(ツ)_/¯
		ContentConfig: rest.ContentConfig{
			GroupVersion:         &corev1.SchemeGroupVersion,
			NegotiatedSerializer: scheme.Codecs,
		},
	}
}
