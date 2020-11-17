// Copyright © 2018 The Knative Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package revision

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/assert"
	"k8s.io/client-go/tools/clientcmd"
	knflags "knative.dev/client/pkg/kn/flags"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/kn/commands"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
	"knative.dev/client/pkg/util"
)

// Helper methods
var blankConfig clientcmd.ClientConfig

func TestExtractTrafficAndTag(t *testing.T) {

	service := &servingv1.Service{
		Status: servingv1.ServiceStatus{
			RouteStatusFields: servingv1.RouteStatusFields{
				Traffic: []servingv1.TrafficTarget{
					createTarget("myv1", 10, "v1"),
					createTarget("myv2", 100, "v1"),
					createTarget("myv1", 20, "stable"),
				},
			},
		},
	}

	percent, tags := trafficAndTagsForRevision("myv1", service)

	assert.Equal(t, percent, int64(30), "expected percentage to be added up")
	assert.Check(t, util.ContainsAll(strings.Join(tags, ","), "v1", "stable"), "all tags included")

}

func createTarget(rev string, percent int64, tag string) servingv1.TrafficTarget {
	return servingv1.TrafficTarget{
		Tag:          tag,
		RevisionName: rev,
		Percent:      &percent,
	}
}

func executeRevisionCommand(client clientservingv1.KnServingClient, args ...string) (string, error) {
	knParams := &commands.KnParams{}
	knParams.ClientConfig = blankConfig

	output := new(bytes.Buffer)
	knParams.Output = output
	knParams.NewServingClient = func(namespace string, cmd *cobra.Command) (clientservingv1.KnServingClient, error) {
		return client, nil
	}
	cmd := NewRevisionCommand(knParams)
	cmd.SetArgs(args)
	cmd.SetOutput(output)

	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return knflags.ReconcileBoolFlags(cmd.Flags())
	}
	err := cmd.Execute()
	return output.String(), err
}
