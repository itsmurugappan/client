// Copyright © 2019 The Knative Authors
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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands"
)

// NewRevisionDeleteCommand represent 'revision delete' command
func NewRevisionDeleteCommand(p *commands.KnParams) *cobra.Command {
	var waitFlags commands.WaitFlags

	RevisionDeleteCommand := &cobra.Command{
		Use:   "delete NAME [NAME ...]",
		Short: "Delete revisions",
		Example: `
  # Delete a revision 'svc1-abcde' in default namespace
  kn revision delete svc1-abcde`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("'kn revision delete' requires one or more revision name")
			}
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			client, err := p.NewServingClient(namespace, cmd)
			if err != nil {
				return err
			}

			errs := []string{}
			for _, name := range args {
				timeout := time.Duration(0)
				if waitFlags.Wait {
					timeout = time.Duration(waitFlags.TimeoutInSeconds) * time.Second
				}
				err = client.DeleteRevision(name, timeout)
				if err != nil {
					errs = append(errs, err.Error())
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "Revision '%s' deleted in namespace '%s'.\n", name, namespace)
				}
			}
			if len(errs) > 0 {
				return errors.New("Error: " + strings.Join(errs, "\nError: "))
			}
			return nil
		},
	}
	commands.AddNamespaceFlags(RevisionDeleteCommand.Flags(), false)
	waitFlags.AddConditionWaitFlags(RevisionDeleteCommand, commands.WaitDefaultTimeout, "delete", "revision", "deleted")
	return RevisionDeleteCommand
}
