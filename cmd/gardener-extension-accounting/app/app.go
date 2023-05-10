package app

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fi-ts/gardener-extension-accounting/pkg/apis/accounting/install"
	"github.com/fi-ts/gardener-extension-accounting/pkg/controller"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/util"

	componentbaseconfig "k8s.io/component-base/config"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	firewallv1 "github.com/metal-stack/firewall-controller/api/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewControllerManagerCommand creates a new command that is used to start the controller.
func NewControllerManagerCommand() *cobra.Command {
	options := NewOptions()

	cmd := &cobra.Command{
		Use:           "gardener-extension-accounting",
		Short:         "provides cluster authentication and authorization in the shoot clusters.",
		SilenceErrors: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			if err := options.optionAggregator.Complete(); err != nil {
				return fmt.Errorf("error completing options: %w", err)
			}

			cmd.SilenceUsage = true
			return options.run(cmd.Context())
		},
	}

	options.optionAggregator.AddFlags(cmd.Flags())

	return cmd
}

func (o *Options) run(ctx context.Context) error {
	// TODO: Make these flags configurable via command line parameters or component config file.
	util.ApplyClientConnectionConfigurationToRESTConfig(&componentbaseconfig.ClientConnectionConfiguration{
		QPS:   100.0,
		Burst: 130,
	}, o.restOptions.Completed().Config)

	mgrOpts := o.managerOptions.Completed().Options()

	mgrOpts.ClientDisableCacheFor = []client.Object{
		&corev1.Secret{},    // applied for ManagedResources
		&corev1.ConfigMap{}, // applied for monitoring config
	}

	mgr, err := manager.New(o.restOptions.Completed().Config, mgrOpts)
	if err != nil {
		return fmt.Errorf("could not instantiate controller-manager: %w", err)
	}

	if err := extensionscontroller.AddToScheme(mgr.GetScheme()); err != nil {
		return fmt.Errorf("could not update manager scheme: %w", err)
	}

	if err := install.AddToScheme(mgr.GetScheme()); err != nil {
		return fmt.Errorf("could not update manager scheme: %w", err)
	}

	ctrlConfig := o.accountingOptions.Completed()
	ctrlConfig.Apply(&controller.DefaultAddOptions.Config)
	o.controllerOptions.Completed().Apply(&controller.DefaultAddOptions.ControllerOptions)
	o.reconcileOptions.Completed().Apply(&controller.DefaultAddOptions.IgnoreOperationAnnotation)

	if err := o.controllerSwitches.Completed().AddToManager(mgr); err != nil {
		return fmt.Errorf("could not add controllers to manager: %w", err)
	}

	// if _, _, err := o.webhookOptions.Completed().AddToManager(ctx, mgr); err != nil {
	// 	return fmt.Errorf("could not add the mutating webhook to manager: %w", err)
	// }

	if err := deployAccountingCWNP(mgr); err != nil {
		return err
	}

	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("error running manager: %w", err)
	}

	return nil
}

func deployAccountingCWNP(mgr manager.Manager) error {
	scheme := runtime.NewScheme()
	utilruntime.Must(firewallv1.AddToScheme(scheme))

	c, err := client.New(mgr.GetConfig(), client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}

	cp := &firewallv1.ClusterwideNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "egress-allow-accounting-api",
			Namespace: "firewall",
		},
	}

	_, err = controllerutil.CreateOrUpdate(context.Background(), c, cp, func() error {
		port9000 := intstr.FromInt(9000)
		tcp := corev1.ProtocolTCP

		cp.Spec.Egress = []firewallv1.EgressRule{
			{
				Ports: []networkingv1.NetworkPolicyPort{
					{
						Port:     &port9000,
						Protocol: &tcp,
					},
				},
				To: []networkingv1.IPBlock{
					{
						CIDR: "0.0.0.0/0",
					},
				},
			},
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("unable to deploy clusterwide network policy for accounting-api into seed firewall namespace %w", err)
	}

	return nil
}
