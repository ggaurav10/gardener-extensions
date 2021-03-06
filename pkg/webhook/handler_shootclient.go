// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gardener/gardener-extensions/pkg/util"

	v1alpha1constants "github.com/gardener/gardener/pkg/apis/core/v1alpha1/constants"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/api/admission/v1beta1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	admissionScheme = runtime.NewScheme()
	admissionCodecs = serializer.NewCodecFactory(admissionScheme)
)

// NewHandlerWithShootClient creates a new handler for the given types, using the given mutator, and logger.
func NewHandlerWithShootClient(mgr manager.Manager, types []runtime.Object, mutator MutatorWithShootClient, logger logr.Logger) (*handlerShootClient, error) {
	// Build a map of the given types keyed by their GVKs
	typesMap, err := buildTypesMap(mgr, types)
	if err != nil {
		return nil, err
	}

	// Create and return a handler
	return &handlerShootClient{
		typesMap: typesMap,
		mutator:  mutator,
		logger:   logger.WithName("handlerShootClient"),
	}, nil
}

type handlerShootClient struct {
	typesMap map[metav1.GroupVersionKind]runtime.Object
	mutator  MutatorWithShootClient
	client   client.Client
	decoder  *admission.Decoder
	logger   logr.Logger
}

// InjectDecoder injects the given decoder into the handler.
func (h *handlerShootClient) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

// InjectClient injects the given client into the mutator.
// TODO Replace this with the more generic InjectFunc when controller runtime supports it
func (h *handlerShootClient) InjectClient(client client.Client) error {
	h.client = client
	if _, err := inject.ClientInto(client, h.mutator); err != nil {
		return errors.Wrap(err, "could not inject the client into the mutator")
	}
	return nil
}

func (h *handlerShootClient) HandleWithRequest(ctx context.Context, req admission.Request, r *http.Request) admission.Response {
	f := func(ctx context.Context, newObj runtime.Object, r *http.Request) error {
		ipPort := strings.Split(r.RemoteAddr, ":")
		if len(ipPort) < 1 {
			return fmt.Errorf("remote address not parseable: %s", r.RemoteAddr)
		}
		ip := ipPort[0]

		podList := &corev1.PodList{}
		if err := h.client.List(ctx, podList, client.MatchingLabels(map[string]string{
			v1alpha1constants.LabelApp:  v1alpha1constants.LabelKubernetes,
			v1alpha1constants.LabelRole: v1alpha1constants.LabelAPIServer,
		})); err != nil {
			return errors.Wrapf(err, "error while listing all pods")
		}

		var shootNamespace string
		for _, pod := range podList.Items {
			if pod.Status.PodIP == ip {
				shootNamespace = pod.Namespace
				break
			}
		}

		if len(shootNamespace) == 0 {
			return fmt.Errorf("could not find shoot namespace for webhook request")
		}

		_, shootClient, err := util.NewClientForShoot(ctx, h.client, shootNamespace, client.Options{})
		if err != nil {
			return errors.Wrapf(err, "could not create shoot client")
		}

		return h.mutator.Mutate(ctx, newObj, shootClient)
	}

	return handle(ctx, req, r, f, h.typesMap, h.decoder, h.logger)
}

// ServeHTTP is a handler for serving an HTTP endpoint that is used for shoot webhooks.
func (h *handlerShootClient) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body []byte
	var err error

	var reviewResponse admission.Response
	if r.Body != nil {
		if body, err = ioutil.ReadAll(r.Body); err != nil {
			h.logger.Error(err, "unable to read the body from the incoming request")
			reviewResponse = admission.Errored(http.StatusBadRequest, err)
			h.writeResponse(w, reviewResponse)
			return
		}
	} else {
		err = errors.New("request body is empty")
		h.logger.Error(err, "bad request")
		reviewResponse = admission.Errored(http.StatusBadRequest, err)
		h.writeResponse(w, reviewResponse)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		err = fmt.Errorf("contentType=%s, expected application/json", contentType)
		h.logger.Error(err, "unable to process a request with an unknown content type", "content type", contentType)
		reviewResponse = admission.Errored(http.StatusBadRequest, err)
		h.writeResponse(w, reviewResponse)
		return
	}

	req := admission.Request{}
	ar := admissionv1beta1.AdmissionReview{
		// avoid an extra copy
		Request: &req.AdmissionRequest,
	}
	if _, _, err := admissionCodecs.UniversalDeserializer().Decode(body, nil, &ar); err != nil {
		h.logger.Error(err, "unable to decode the request")
		reviewResponse = admission.Errored(http.StatusBadRequest, err)
		h.writeResponse(w, reviewResponse)
		return
	}

	reviewResponse = h.HandleWithRequest(context.Background(), req, r)
	h.writeResponse(w, reviewResponse)
}

func (h *handlerShootClient) writeResponse(w io.Writer, response admission.Response) {
	if err := json.NewEncoder(w).Encode(v1beta1.AdmissionReview{
		Response: &response.AdmissionResponse,
	}); err != nil {
		h.logger.Error(err, "unable to encode the response")
		h.writeResponse(w, admission.Errored(http.StatusInternalServerError, err))
	}
}
