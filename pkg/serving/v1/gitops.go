// Copyright 2020 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"

	"knative.dev/client/pkg/wait"

	"k8s.io/apimachinery/pkg/watch"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

const (
	operation_not_suported_error = "this operation is not supported in gitops mode"
	ksvc_kind                    = "ksvc"
)

// knServingGitOpsClient - kn service client
// to work on a local repo instead of a remote cluster
type knServingGitOpsClient struct {
	dir       string
	namespace string
	printer   printers.ResourcePrinter
}

// NewKnServingGitOpsClient returns an instance of the
// kn service gitops client
func NewKnServingGitOpsClient(namespace, dir string) KnServingClient {
	yamlPrinter, _ := genericclioptions.NewJSONYamlPrintFlags().ToPrinter("yaml")
	return &knServingGitOpsClient{
		dir:       dir,
		namespace: namespace,
		printer:   yamlPrinter,
	}
}

func (cl *knServingGitOpsClient) getKsvcFilePath(name string) string {
	return filepath.Join(cl.dir, cl.namespace, ksvc_kind, name+".yaml")
}

// Namespace returns the namespace
func (cl *knServingGitOpsClient) Namespace() string {
	return cl.namespace
}

// GetService returns the knative service for the name
func (cl *knServingGitOpsClient) GetService(name string) (*servingv1.Service, error) {
	return getServiceFromFile(cl.getKsvcFilePath(name), name)
}

func getServiceFromFile(filePath, name string) (*servingv1.Service, error) {
	var svc servingv1.Service
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, apierrors.NewNotFound(servingv1.Resource("services"), name)
		}
		return nil, err
	}
	decoder := yaml.NewYAMLOrJSONDecoder(file, 512)
	if err := decoder.Decode(&svc); err != nil {
		return nil, err
	}
	updateServingGvk(&svc)
	return &svc, nil
}

// WatchService is not supported by this client
func (cl *knServingGitOpsClient) WatchService(name string, timeout time.Duration) (watch.Interface, error) {
	return nil, fmt.Errorf(operation_not_suported_error)
}

// WatchRevision is not supported by this client
func (cl *knServingGitOpsClient) WatchRevision(name string, timeout time.Duration) (watch.Interface, error) {
	return nil, fmt.Errorf(operation_not_suported_error)
}

// ListServices lists the services in the path provided
func (cl *knServingGitOpsClient) ListServices(config ...ListConfig) (*servingv1.ServiceList, error) {
	var root string
	var services []servingv1.Service

	switch cl.namespace {
	case "":
		root = cl.dir
	default:
		root = filepath.Join(cl.dir, cl.namespace)
	}

	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		switch {
		// skip if dir is not ksvc
		case info.IsDir():
			return nil

		// skip non yaml files
		case !strings.Contains(info.Name(), ".yaml"):
			return nil

		// skip non ksvc dir
		case !strings.Contains(path, ksvc_kind):
			return filepath.SkipDir

		default:
			svc, err := getServiceFromFile(path, "")
			if err != nil {
				return err
			}
			updateServingGvk(svc)
			services = append(services, *svc)
			return nil
		}
	}); err != nil {
		return nil, err
	}

	typeMeta := metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "List",
	}
	serviceList := &servingv1.ServiceList{
		TypeMeta: typeMeta,
		Items:    services,
	}

	return serviceList, nil
}

// CreateService saves the knative service spec in
// yaml format in the local path provided
func (cl *knServingGitOpsClient) CreateService(service *servingv1.Service) error {
	// create dir , if not present
	namespaceDir := filepath.Join(cl.dir, cl.namespace, ksvc_kind)
	if _, err := os.Stat(namespaceDir); os.IsNotExist(err) {
		os.MkdirAll(namespaceDir, 0777)
	}
	serviceFile, err := os.Create(cl.getKsvcFilePath(service.ObjectMeta.Name))
	if err != nil {
		return err
	}
	updateServingGvk(service)
	return cl.printer.PrintObj(service, serviceFile)
}

// UpdateService updates the service in
// the local directory
func (cl *knServingGitOpsClient) UpdateService(service *servingv1.Service) error {
	// check if file exist
	if _, err := cl.GetService(service.ObjectMeta.Name); err != nil {
		return err
	}
	// replace file
	return cl.CreateService(service)
}

// UpdateServiceWithRetry updates the service in the local directory
func (cl *knServingGitOpsClient) UpdateServiceWithRetry(name string, updateFunc ServiceUpdateFunc, nrRetries int) error {
	return updateServiceWithRetry(cl, name, updateFunc, nrRetries)
}

// ApplyService is not supported by this client
func (cl *knServingGitOpsClient) ApplyService(modifiedService *servingv1.Service) (bool, error) {
	return false, fmt.Errorf(operation_not_suported_error)
}

// DeleteService removes the file from the local file system
func (cl *knServingGitOpsClient) DeleteService(serviceName string, timeout time.Duration) error {
	return os.Remove(cl.getKsvcFilePath(serviceName))
}

// WaitForService always returns success for this client
func (cl *knServingGitOpsClient) WaitForService(name string, timeout time.Duration, msgCallback wait.MessageCallback) (error, time.Duration) {
	return nil, 1 * time.Second
}

// GetConfiguration not supported by this client
func (cl *knServingGitOpsClient) GetConfiguration(name string) (*servingv1.Configuration, error) {
	return nil, fmt.Errorf(operation_not_suported_error)
}

// GetRevision not supported by this client
func (cl *knServingGitOpsClient) GetRevision(name string) (*servingv1.Revision, error) {
	return nil, fmt.Errorf(operation_not_suported_error)
}

// GetBaseRevision not supported by this client
func (cl *knServingGitOpsClient) GetBaseRevision(service *servingv1.Service) (*servingv1.Revision, error) {
	return nil, fmt.Errorf(operation_not_suported_error)
}

// CreateRevision not supported by this client
func (cl *knServingGitOpsClient) CreateRevision(revision *servingv1.Revision) error {
	return fmt.Errorf(operation_not_suported_error)
}

// UpdateRevision not supported by this client
func (cl *knServingGitOpsClient) UpdateRevision(revision *servingv1.Revision) error {
	return fmt.Errorf(operation_not_suported_error)
}

// DeleteRevision not supported by this client
func (cl *knServingGitOpsClient) DeleteRevision(name string, timeout time.Duration) error {
	return fmt.Errorf(operation_not_suported_error)
}

// ListRevisions not supported by this client
func (cl *knServingGitOpsClient) ListRevisions(config ...ListConfig) (*servingv1.RevisionList, error) {
	return nil, fmt.Errorf(operation_not_suported_error)
}

// GetRoute not supported by this client
func (cl *knServingGitOpsClient) GetRoute(name string) (*servingv1.Route, error) {
	return nil, fmt.Errorf(operation_not_suported_error)
}

// ListRoutes not supported by this client
func (cl *knServingGitOpsClient) ListRoutes(config ...ListConfig) (*servingv1.RouteList, error) {
	return nil, fmt.Errorf(operation_not_suported_error)
}
