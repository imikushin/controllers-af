/*
Copyright 2021 Ivan Mikushin

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

package reconciler

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAddListToCache(t *testing.T) {
	cache := cache{}
	const cm1uid = "cm1"
	cm1 := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			UID: cm1uid,
		},
		Data: map[string]string{
			"a": "a",
			"b": "b",
		},
	}
	objectList := &corev1.ConfigMapList{Items: []corev1.ConfigMap{
		*cm1,
	}}

	addListToCache(cache, objectList)
	assert.EqualValues(t, cm1, cache[cm1uid])

	objectList.Items[0].Data["a"] = "c"
	addListToCache(cache, objectList)
	assert.EqualValues(t, cm1, cache[cm1uid])
}

func TestPanicWithArg(t *testing.T) {
	expectedErr := errors.New("expected")
	err := func() (retErr error) {
		defer func() {
			retErr = panicErr(recover(), retErr)
		}()
		panic(expectedErr)
	}()
	assert.Equal(t, expectedErr, err)
}
