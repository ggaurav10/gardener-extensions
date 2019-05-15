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

package app

import (
	"context"
	"fmt"
	"os"

	gcpinstall "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/apis/gcp/install"
	gcpcmd "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/cmd"
	gcpcontrolplane "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/controller/controlplane"
	gcpinfrastructure "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/controller/infrastructure"
	gcpworker "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/controller/worker"
	"github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/gcp"
	"github.com/gardener/gardener-extensions/pkg/controller"
	controllercmd "github.com/gardener/gardener-extensions/pkg/controller/cmd"
	"github.com/gardener/gardener-extensions/pkg/controller/infrastructure"
	webhookcmd "github.com/gardener/gardener-extensions/pkg/webhook/cmd"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// NewControllerManagerCommand creates a new command for running a GCP provider controller.
func NewControllerManagerCommand(ctx context.Context) *cobra.Command {
	var (
		restOpts = &controllercmd.RESTOptions{}
		mgrOpts  = &controllercmd.ManagerOptions{
			LeaderElection:          true,
			LeaderElectionID:        controllercmd.LeaderElectionNameID(gcp.Name),
			LeaderElectionNamespace: os.Getenv("LEADER_ELECTION_NAMESPACE"),
		}
		configFileOpts = &gcpcmd.ConfigOptions{}

		infraCtrlOpts = &controllercmd.ControllerOptions{
			MaxConcurrentReconciles: 5,
		}
		infraReconcileOpts = &infrastructure.ReconcilerOptions{
			IgnoreOperationAnnotation: true,
		}
		unprefixedInfraOpts = controllercmd.NewOptionAggregator(infraCtrlOpts, infraReconcileOpts)

		controlPlaneCtrlOpts = &controllercmd.ControllerOptions{
			MaxConcurrentReconciles: 5,
		}

		workerCtrlOpts = &controllercmd.ControllerOptions{
			MaxConcurrentReconciles: 5,
		}

		controllerSwitches = gcpcmd.ControllerSwitchOptions()
		webhookSwitches    = gcpcmd.WebhookAddToManagerOptions()

		aggOption = controllercmd.NewOptionAggregator(
			restOpts,
			mgrOpts,
			controllercmd.PrefixOption("infrastructure-", &unprefixedInfraOpts),
			controllercmd.PrefixOption("controlplane-", controlPlaneCtrlOpts),
			controllercmd.PrefixOption("worker-", workerCtrlOpts),
			controllerSwitches,
			webhookSwitches,
		)
	)
	webhookSwitches.Server = webhookcmd.ServerOptions{
		Port:             7890,
		CertDir:          "/tmp/cert",
		Mode:             webhookcmd.ServiceMode,
		Name:             "webhooks",
		Namespace:        os.Getenv("WEBHOOK_CONFIG_NAMESPACE"),
		ServiceSelectors: "{}",
		Host:             "localhost",
	}

	cmd := &cobra.Command{
		Use: fmt.Sprintf("%s-controller-manager", gcp.Name),

		Run: func(cmd *cobra.Command, args []string) {
			if err := aggOption.Complete(); err != nil {
				controllercmd.LogErrAndExit(err, "Error completing options")
			}

			if err := configFileOpts.Complete(); err != nil {
				controllercmd.LogErrAndExit(err, "Error completing config options")
			}

			mgr, err := manager.New(restOpts.Completed().Config, mgrOpts.Completed().Options())
			if err != nil {
				controllercmd.LogErrAndExit(err, "Could not instantiate manager")
			}

			if err := controller.AddToScheme(mgr.GetScheme()); err != nil {
				controllercmd.LogErrAndExit(err, "Could not update manager scheme")
			}

			if err := gcpinstall.AddToScheme(mgr.GetScheme()); err != nil {
				controllercmd.LogErrAndExit(err, "Could not update manager scheme")
			}

			configFileOpts.Completed().ApplyMachineImages(&gcpworker.DefaultAddOptions.MachineImages)
			controlPlaneCtrlOpts.Completed().Apply(&gcpcontrolplane.Options)
			infraCtrlOpts.Completed().Apply(&gcpinfrastructure.DefaultAddOptions.Controller)
			infraReconcileOpts.Completed().Apply(&gcpinfrastructure.DefaultAddOptions.IgnoreOperationAnnotation)
			workerCtrlOpts.Completed().Apply(&gcpworker.DefaultAddOptions.Controller)

			if err := controllerSwitches.Completed().AddToManager(mgr); err != nil {
				controllercmd.LogErrAndExit(err, "Could not add controllers to manager")
			}

			if err := webhookSwitches.Completed().AddToManager(mgr); err != nil {
				controllercmd.LogErrAndExit(err, "Could not add webhooks to manager")
			}

			if err := mgr.Start(ctx.Done()); err != nil {
				controllercmd.LogErrAndExit(err, "Error running manager")
			}
		},
	}

	aggOption.AddFlags(cmd.Flags())
	configFileOpts.AddFlags(cmd.Flags())

	return cmd
}
