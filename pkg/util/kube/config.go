/*
Copyright 2018 The Jetstack cert-manager contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kube

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/jetstack/cert-manager/pkg/util"
	"github.com/jetstack/cert-manager/pkg/util/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// KubeConfig will return a rest.Config for communicating with the Kubernetes API server.
// If apiServerHost is specified, a config without authentication that is configured
// to talk to the apiServerHost URL will be returned. Else, the in-cluster config will be loaded,
// and failing this, the config will be loaded from the users local kubeconfig directory
func KubeConfig(apiServerHost string) (*rest.Config, error) {
	var err error
	var cfg *rest.Config

	if len(apiServerHost) > 0 {
		cfg = new(rest.Config)
		cfg.Host = apiServerHost
	} else if cfg, err = rest.InClusterConfig(); err != nil {
		apiCfg, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()

		if err != nil {
			return nil, fmt.Errorf("error loading cluster config: %s", err.Error())
		}

		cfg, err = clientcmd.NewDefaultClientConfig(*apiCfg, &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("error loading cluster client config: %s", err.Error())
		}
	}
	cfg.UserAgent = util.CertManagerUserAgent

	return cfg, nil
}

// ConfigFromSecret will return a rest.Config for communicating with the Kubernetes API server.
// The instance from rest.Config is created out of a secret within the cluster which contains a Base64 encoded Kubeconfig.
func ConfigFromSecret(namespace, secretname, kubeConfigKey string) (*rest.Config, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		glog.Errorln("An error occurred while getting the in-cluster config")
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Errorln("An error occurred while creating a clientset out of the in-cluster config")
		return nil, err
	}

	secret, err := clientset.CoreV1().Secrets(namespace).Get(secretname, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("Secret with name %s is not available in namespace %s\n", secretname, namespace)
		return nil, err
	}

	return newRESTConfigFromSecret(secret, kubeConfigKey)
}

func newRESTConfigFromSecret(secret *v1.Secret, kubeconfigKey string) (*rest.Config, error) {
	kubeConfigData, ok := secret.Data[kubeconfigKey]
	if !ok {
		err := errors.NewInvalidData("Invalid Kubeconfig key %s for secret %s", kubeconfigKey, secret.GetName())
		return nil, err
	}
	cfg, err := clientcmd.Load(kubeConfigData)
	if err != nil {
		return nil, err
	}

	clientConfig := clientcmd.NewDefaultClientConfig(*cfg, &clientcmd.ConfigOverrides{})
	return clientConfig.ClientConfig()
}
