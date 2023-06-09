package cmd

import (
	"github.com/fi-ts/gardener-extension-accounting/pkg/controller"
	// "github.com/fi-ts/gardener-extension-accounting/pkg/webhook/kapiserver"
	controllercmd "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	webhookcmd "github.com/gardener/gardener/extensions/pkg/webhook/cmd"
)

// ControllerSwitchOptions are the controllercmd.SwitchOptions for the provider controllers.
func ControllerSwitchOptions() *controllercmd.SwitchOptions {
	return controllercmd.NewSwitchOptions(
		controllercmd.Switch(controller.ControllerName, controller.AddToManager),
	)
}

// WebhookSwitchOptions are the webhookcmd.SwitchOptions for the provider webhooks.
func WebhookSwitchOptions() *webhookcmd.SwitchOptions {
	return webhookcmd.NewSwitchOptions(
	// webhookcmd.Switch("fits-accounting-webhook", kapiserver.New),
	)
}
