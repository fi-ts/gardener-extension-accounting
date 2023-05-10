package app

import (
	"os"

	controllercmd "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	"github.com/gardener/gardener/extensions/pkg/controller/controlplane/genericactuator"
	webhookcmd "github.com/gardener/gardener/extensions/pkg/webhook/cmd"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	accountingcmd "github.com/fi-ts/gardener-extension-accounting/pkg/cmd"
)

// ExtensionName is the name of the extension.
const ExtensionName = "extension-fits-accounting"

// Options holds configuration passed to the registry service controller.
type Options struct {
	generalOptions     *controllercmd.GeneralOptions
	accountingOptions  *accountingcmd.AccountingOptions
	restOptions        *controllercmd.RESTOptions
	managerOptions     *controllercmd.ManagerOptions
	controllerOptions  *controllercmd.ControllerOptions
	healthOptions      *controllercmd.ControllerOptions
	controllerSwitches *controllercmd.SwitchOptions
	webhookOptions     *webhookcmd.AddToManagerOptions
	reconcileOptions   *controllercmd.ReconcilerOptions
	optionAggregator   controllercmd.OptionAggregator
}

// NewOptions creates a new Options instance.
func NewOptions() *Options {
	// options for the webhook server
	webhookServerOptions := &webhookcmd.ServerOptions{
		Namespace: os.Getenv("WEBHOOK_CONFIG_NAMESPACE"),
	}

	webhookSwitches := accountingcmd.WebhookSwitchOptions()
	webhookOptions := webhookcmd.NewAddToManagerOptions(
		"fits-accounting",
		genericactuator.ShootWebhooksResourceName,
		genericactuator.ShootWebhookNamespaceSelector("metal"),
		webhookServerOptions,
		webhookSwitches,
	)

	options := &Options{
		generalOptions:    &controllercmd.GeneralOptions{},
		accountingOptions: &accountingcmd.AccountingOptions{},
		restOptions:       &controllercmd.RESTOptions{},
		managerOptions: &controllercmd.ManagerOptions{
			// These are default values.
			LeaderElection:             true,
			LeaderElectionID:           controllercmd.LeaderElectionNameID(ExtensionName),
			LeaderElectionResourceLock: resourcelock.LeasesResourceLock,
			LeaderElectionNamespace:    os.Getenv("LEADER_ELECTION_NAMESPACE"),
		},
		controllerOptions: &controllercmd.ControllerOptions{
			// This is a default value.
			MaxConcurrentReconciles: 5,
		},
		healthOptions: &controllercmd.ControllerOptions{
			// This is a default value.
			MaxConcurrentReconciles: 5,
		},
		controllerSwitches: accountingcmd.ControllerSwitchOptions(),
		reconcileOptions:   &controllercmd.ReconcilerOptions{},
		webhookOptions:     webhookOptions,
	}

	options.optionAggregator = controllercmd.NewOptionAggregator(
		options.generalOptions,
		options.restOptions,
		options.managerOptions,
		options.controllerOptions,
		options.accountingOptions,
		controllercmd.PrefixOption("healthcheck-", options.healthOptions),
		options.controllerSwitches,
		options.reconcileOptions,
		options.webhookOptions,
	)

	return options
}
