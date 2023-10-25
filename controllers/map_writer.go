/*
MIT License

Copyright (c) 2022 Stephen Cuppett

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package controllers

import (
	"context"
	"regexp"
	"strings"

	"github.com/go-logr/logr"
	"github.com/linki/cloudformation-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type MapWriter struct {
	client.Client
	Log logr.Logger
	ChannelHub
	*runtime.Scheme
}

func createTempMap() *v1.ConfigMap {
	tempMap := v1.ConfigMap{}
	return &tempMap
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

func clearString(str string) string {
	return strings.ToLower(nonAlphanumericRegex.ReplaceAllString(str, ""))
}

// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Worker
func (w *MapWriter) Worker() {

	for {
		toBeMapped := <-w.ChannelHub.MappingChannel
		w.Log.Info("Synchronizing map", "Namespace", toBeMapped.Namespace, "Stack ID", toBeMapped.Status.StackID)

		for _, v := range toBeMapped.Status.Resources {

			// Getting the map
			m := createTempMap()
			created := false
			// w.Log.Info(clearString(toBeMapped.Name + v.LogicalId))
			namespacedName := types.NamespacedName{Namespace: toBeMapped.Namespace, Name: clearString(v.LogicalId)}
			err := w.Client.Get(context.TODO(), namespacedName, m)
			if errors.IsNotFound(err) {
				w.Log.Info("Map resource not found. To be created.", "Namespace", toBeMapped.Namespace, "Stack ID", toBeMapped.Status.StackID)
				*m = w.createMap(toBeMapped, clearString(v.LogicalId))
				created = true
			}

			temp := make(map[string]string)
			temp["LogicalId"] = v.LogicalId
			temp["PhysicalId"] = v.PhysicalId
			temp["Type"] = v.Type
			temp["Status"] = v.Status
			temp["StatusReason"] = v.StatusReason

			m.Data = temp

			// Setting the owner reference
			err = controllerutil.SetControllerReference(toBeMapped, m, w.Scheme)
			if err != nil {
				w.Log.Info(err.Error(), "Unable to set controller owner.", "Namespace", toBeMapped.Namespace, "Name", m.Name, "Stack ID", toBeMapped.Status.StackID)
			} else {
				if created {
					err = w.Client.Create(context.TODO(), m)
				} else {
					err = w.Client.Update(context.TODO(), m)
				}
			}

			if err != nil {
				w.Log.Info(err.Error(), "Failed to create or update map.", "Namespace", toBeMapped.Namespace, "Name", m.Name, "Stack ID", toBeMapped.Status.StackID)
			} else {
				w.Log.Info("Map written", "Namespace", toBeMapped.Namespace, "Name", m.Name, "Stack ID", toBeMapped.Status.StackID)
			}
		}

	}
}

func (w *MapWriter) createMap(stack *v1alpha1.Stack, name string) v1.ConfigMap {
	return v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: stack.Namespace,
		},
	}
}
