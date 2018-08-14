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
	"strings"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestInvalidKubeConfigKey(t *testing.T) {
	const secretKey = "AnyKey"

	metaObj := metav1.ObjectMeta{Name: "TestSecret"}
	testSecret := v1.Secret{ObjectMeta: metaObj}

	cfg, err := newRESTConfigFromSecret(&testSecret, secretKey)

	if cfg != nil {
		t.Errorf("Kubeconfig is expected to be nil")
	}
	if err == nil {
		t.Errorf("expected an error object")
	} else {
		errorText := err.Error()
		if !strings.Contains(errorText, secretKey) {
			t.Errorf("error message is expected to contain key \"%s\": \"%s\"", secretKey, errorText)
		}
	}
}
