package cmd

import (
	"errors"
	"os"

	configapi "github.com/fi-ts/gardener-extension-accounting/pkg/apis/config"
	"github.com/fi-ts/gardener-extension-accounting/pkg/apis/config/v1alpha1"
	healthcheckconfig "github.com/gardener/gardener/extensions/pkg/apis/config"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var (
	scheme  *runtime.Scheme
	decoder runtime.Decoder
)

func init() {
	scheme = runtime.NewScheme()
	utilruntime.Must(configapi.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))

	decoder = serializer.NewCodecFactory(scheme).UniversalDecoder()
}

// RegistryOptions holds options related to the registry service.
type AccountingOptions struct {
	ConfigLocation string
	config         *AccountingServiceConfig
}

// AddFlags implements Flagger.AddFlags.
func (o *AccountingOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ConfigLocation, "config", "", "Path to registry service configuration")
}

// Complete implements Completer.Complete.
func (o *AccountingOptions) Complete() error {
	if o.ConfigLocation == "" {
		return errors.New("config location is not set")
	}
	data, err := os.ReadFile(o.ConfigLocation)
	if err != nil {
		return err
	}

	config := configapi.ControllerConfiguration{}
	_, _, err = decoder.Decode(data, nil, &config)
	if err != nil {
		return err
	}

	// if errs := validation.ValidateConfiguration(&config); len(errs) > 0 {
	// 	return errs.ToAggregate()
	// }

	o.config = &AccountingServiceConfig{
		config: config,
	}

	return nil
}

// Completed returns the decoded RegistryServiceConfiguration instance. Only call this if `Complete` was successful.
func (o *AccountingOptions) Completed() *AccountingServiceConfig {
	return o.config
}

// RegistryServiceConfig contains configuration information about the registry service.
type AccountingServiceConfig struct {
	config configapi.ControllerConfiguration
}

// Apply applies the RegistryOptions to the passed ControllerOptions instance.
func (c *AccountingServiceConfig) Apply(config *configapi.ControllerConfiguration) {
	*config = c.config
}

// ApplyHealthCheckConfig applies the HealthCheckConfig.
func (c *AccountingServiceConfig) ApplyHealthCheckConfig(config *healthcheckconfig.HealthCheckConfig) {
	if c.config.HealthCheckConfig != nil {
		*config = *c.config.HealthCheckConfig
	}
}
