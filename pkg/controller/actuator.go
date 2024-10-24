package controller

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/fi-ts/gardener-extension-accounting/pkg/apis/accounting/v1alpha1"
	"github.com/fi-ts/gardener-extension-accounting/pkg/apis/config"
	"github.com/fi-ts/gardener-extension-accounting/pkg/imagevector"
	"github.com/gardener/gardener/extensions/pkg/controller"
	gardenercontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/extensions"
	gutil "github.com/gardener/gardener/pkg/utils/gardener"
	"github.com/gardener/gardener/pkg/utils/managedresources"
	"github.com/go-logr/logr"
	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-go/api/client/project"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/cache"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	metalhelper "github.com/metal-stack/gardener-extension-provider-metal/pkg/apis/metal/helper"
	metalv1alpha1 "github.com/metal-stack/gardener-extension-provider-metal/pkg/apis/metal/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewActuator returns an actuator responsible for Extension resources.
func NewActuator(mgr manager.Manager, config config.ControllerConfiguration) extension.Actuator {
	a := &actuator{
		client:  mgr.GetClient(),
		decoder: serializer.NewCodecFactory(mgr.GetScheme(), serializer.EnableStrict).UniversalDecoder(),
		config:  config,
	}
	a.projects = cache.NewFetchAll(30*time.Minute, a.fetchAllProjects)
	return a
}

type actuator struct {
	client  client.Client
	decoder runtime.Decoder
	config  config.ControllerConfiguration

	projects *cache.FetchAllCache[string, *models.V1ProjectResponse]
}

// ForceDelete implements extension.Actuator.
func (a *actuator) ForceDelete(context.Context, logr.Logger, *extensionsv1alpha1.Extension) error {
	return nil
}

// Reconcile the Extension resource.
func (a *actuator) Reconcile(ctx context.Context, log logr.Logger, ex *extensionsv1alpha1.Extension) error {
	namespace := ex.GetNamespace()

	cluster, err := controller.GetCluster(ctx, a.client, namespace)
	if err != nil {
		return err
	}

	accountingConfig := &v1alpha1.AccountingConfig{}
	if ex.Spec.ProviderConfig != nil {
		if _, _, err := a.decoder.Decode(ex.Spec.ProviderConfig.Raw, nil, accountingConfig); err != nil {
			return fmt.Errorf("failed to decode provider config: %w", err)
		}
	}

	if err := a.createResources(ctx, log, accountingConfig, cluster, namespace); err != nil {
		return err
	}

	return nil
}

// Delete the Extension resource.
func (a *actuator) Delete(ctx context.Context, log logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return a.deleteResources(ctx, log, ex.GetNamespace())
}

// Restore the Extension resource.
func (a *actuator) Restore(ctx context.Context, log logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return a.Reconcile(ctx, log, ex)
}

// Migrate the Extension resource.
func (a *actuator) Migrate(ctx context.Context, log logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return nil
}

func (a *actuator) createResources(ctx context.Context, log logr.Logger, _ *v1alpha1.AccountingConfig, cluster *controller.Cluster, namespace string) error {
	shootAccessSecret := gutil.NewShootAccessSecret(gutil.SecretNamePrefixShootAccess+"accounting-exporter", namespace)
	if err := shootAccessSecret.Reconcile(ctx, a.client); err != nil {
		return err
	}

	infrastructureConfig := &metalv1alpha1.InfrastructureConfig{}
	err := metalhelper.DecodeRawExtension(cluster.Shoot.Spec.Provider.InfrastructureConfig, infrastructureConfig, a.decoder)
	if err != nil {
		return fmt.Errorf("unable decoding infrastructure config: %w", err)
	}

	resp, err := a.projects.Get(ctx, infrastructureConfig.ProjectID)
	if err != nil {
		return fmt.Errorf("error fetching cluster project from metal-api: %w", err)
	}

	shootObjects := shootObjects()

	seedObjects, err := seedObjects(&a.config, infrastructureConfig, resp, cluster, namespace, shootAccessSecret.Secret.Name)
	if err != nil {
		return err
	}

	shootResources, err := managedresources.NewRegistry(kubernetes.ShootScheme, kubernetes.ShootCodec, kubernetes.ShootSerializer).AddAllAndSerialize(shootObjects...)
	if err != nil {
		return err
	}

	seedResources, err := managedresources.NewRegistry(kubernetes.SeedScheme, kubernetes.SeedCodec, kubernetes.SeedSerializer).AddAllAndSerialize(seedObjects...)
	if err != nil {
		return err
	}

	if err := managedresources.CreateForShoot(ctx, a.client, namespace, v1alpha1.ShootAccountingResourceName, "fits-accounting", false, shootResources); err != nil {
		return err
	}

	log.Info("managed resource created successfully", "name", v1alpha1.ShootAccountingResourceName)

	if err := managedresources.CreateForSeed(ctx, a.client, namespace, v1alpha1.SeedAccountingResourceName, false, seedResources); err != nil {
		return err
	}

	log.Info("managed resource created successfully", "name", v1alpha1.SeedAccountingResourceName)

	return nil
}

func (a *actuator) fetchAllProjects(ctx context.Context) (map[string]*models.V1ProjectResponse, error) {
	// we need to lookup the project name from the metal-api
	// unfortunately we do not have it anywhere in the cluster spec
	mclient, err := metalgo.NewDriver(a.config.Accounting.MetalURL, "", a.config.Accounting.MetalHMAC, metalgo.AuthType(a.config.Accounting.MetalAuthType))
	if err != nil {
		return nil, fmt.Errorf("error creating metal client: %w", err)
	}

	projects, err := mclient.Project().ListProjects(project.NewListProjectsParams().WithContext(ctx), nil)
	if err != nil {
		return nil, fmt.Errorf("error fetching projects from metal-api: %w", err)
	}

	result := make(map[string]*models.V1ProjectResponse)
	for _, p := range projects.Payload {
		result[p.Meta.ID] = p
	}

	return result, nil
}

func (a *actuator) deleteResources(ctx context.Context, log logr.Logger, namespace string) error {
	log.Info("deleting managed resource for registry cache")

	if err := managedresources.Delete(ctx, a.client, namespace, v1alpha1.ShootAccountingResourceName, false); err != nil {
		return err
	}

	if err := managedresources.Delete(ctx, a.client, namespace, v1alpha1.SeedAccountingResourceName, false); err != nil {
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	if err := managedresources.WaitUntilDeleted(timeoutCtx, a.client, namespace, v1alpha1.ShootAccountingResourceName); err != nil {
		return err
	}

	if err := managedresources.WaitUntilDeleted(timeoutCtx, a.client, namespace, v1alpha1.SeedAccountingResourceName); err != nil {
		return err
	}

	return nil
}

func seedObjects(cc *config.ControllerConfiguration, infrastructureConfig *metalv1alpha1.InfrastructureConfig, project *models.V1ProjectResponse, cluster *controller.Cluster, namespace, shootAccessSecretName string) ([]client.Object, error) {
	accountingExporterImage, err := imagevector.ImageVector().FindImage("accounting-exporter")
	if err != nil {
		return nil, fmt.Errorf("failed to find accounting-exporter image: %w", err)
	}

	replicas := int32(1)
	if gardenercontroller.IsHibernated(cluster) {
		replicas = 0
	}

	accountingExporterDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "accounting-exporter",
			Namespace: namespace,
			Labels: map[string]string{
				"k8s-app": "accounting-exporter",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Pointer(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"k8s-app": "accounting-exporter",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"k8s-app": "accounting-exporter",
						"app":     "accounting-exporter",
						"networking.gardener.cloud/from-prometheus":                     "allowed",
						"networking.gardener.cloud/to-dns":                              "allowed",
						"networking.gardener.cloud/to-shoot-apiserver":                  "allowed",
						"networking.gardener.cloud/to-public-networks":                  "allowed",
						"networking.resources.gardener.cloud/to-kube-apiserver-tcp-443": "allowed",
					},
					Annotations: map[string]string{
						"scheduler.alpha.kubernetes.io/critical-pod": "",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "accounting-exporter",
							Image:           accountingExporterImage.String(),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: []corev1.ContainerPort{
								{
									Name:          "health",
									ContainerPort: 3000,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/health",
										Port:   intstr.FromString("health"),
										Scheme: corev1.URISchemeHTTP,
									},
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/health",
										Port:   intstr.FromString("health"),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								FailureThreshold:    1,
								InitialDelaySeconds: 120,
							},
							Env: []corev1.EnvVar{
								{
									Name:  "KUBE_COUNTER_BIND_ADDR",
									Value: "0.0.0.0",
								},
								{
									Name:  "KUBE_COUNTER_KUBECONFIG",
									Value: gutil.PathGenericKubeconfig,
								},
								{
									Name:  "KUBE_COUNTER_PARTITION",
									Value: infrastructureConfig.PartitionID,
								},
								{
									Name:  "KUBE_COUNTER_TENANT",
									Value: project.TenantID,
								},
								{
									Name:  "KUBE_COUNTER_PROJECT_ID",
									Value: infrastructureConfig.ProjectID,
								},
								{
									Name:  "KUBE_COUNTER_PROJECT_NAME",
									Value: project.Name,
								},
								{
									Name:  "KUBE_COUNTER_CLUSTER_ID",
									Value: string(cluster.Shoot.UID),
								},
								{
									Name:  "KUBE_COUNTER_CLUSTER_NAME",
									Value: cluster.Shoot.Name,
								},
								{
									Name:  "KUBE_COUNTER_ACCOUNTING_API_HOSTNAME",
									Value: cc.Accounting.AccountingHost,
								},
								{
									Name:  "KUBE_COUNTER_ACCOUNTING_API_PORT",
									Value: cc.Accounting.AccountingPort,
								},
								{
									Name:  "KUBE_COUNTER_NETWORK_TRAFFIC_ENABLED",
									Value: strconv.FormatBool(true),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: "/certs",
									Name:      "certs",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "certs",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "accounting-exporter-tls",
								},
							},
						},
					},
				},
			},
		},
	}

	if err := gutil.InjectGenericKubeconfig(accountingExporterDeployment, extensions.GenericTokenKubeconfigSecretNameFromCluster(cluster), shootAccessSecretName); err != nil {
		return nil, err
	}

	objects := []client.Object{
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "accounting-exporter-tls",
				Namespace: namespace,
			},
			StringData: map[string]string{
				"ca.pem":         cc.Accounting.CA,
				"client.pem":     cc.Accounting.ClientCert,
				"client-key.pem": cc.Accounting.ClientKey,
			},
		},
		accountingExporterDeployment,
	}

	if cc.ImagePullSecret != nil && cc.ImagePullSecret.DockerConfigJSON != "" {
		content, err := base64.StdEncoding.DecodeString(cc.ImagePullSecret.DockerConfigJSON)
		if err != nil {
			return nil, fmt.Errorf("unable to decode image pull secret: %w", err)
		}

		objects = append(objects, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "accounting-exporter-registry-credentials",
				Namespace: namespace,
				Labels: map[string]string{
					"app": "accounting-exporter-registry-credentials",
				},
			},
			Type: corev1.SecretTypeDockerConfigJson,
			Data: map[string][]byte{
				".dockerconfigjson": content,
			},
		})

		accountingExporterDeployment.Spec.Template.Spec.ImagePullSecrets = append(accountingExporterDeployment.Spec.Template.Spec.ImagePullSecrets, corev1.LocalObjectReference{
			Name: "accounting-exporter-registry-credentials",
		})
	}

	return objects, nil
}

func shootObjects() []client.Object {
	return []client.Object{
		&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: "system:accounting-exporter",
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{
						"namespaces",
						"pods",
						"persistentvolumes",
						"persistentvolumeclaims",
					},
					Verbs: []string{
						"get",
						"list",
						"watch",
					},
				},
				{
					APIGroups: []string{"storage.k8s.io"},
					Resources: []string{
						"storageclasses",
					},
					Verbs: []string{
						"get",
					},
				},
				{
					APIGroups: []string{"metal-stack.io"},
					Resources: []string{
						"firewalls",
					},
					Verbs: []string{
						"get",
					},
				},
				{
					APIGroups: []string{"firewall.metal-stack.io"},
					Resources: []string{
						"firewallmonitors",
					},
					Verbs: []string{
						"get",
						"list",
						"watch",
					},
				},
			},
		},
		&rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "system:accounting-exporter",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "accounting-exporter",
					Namespace: "kube-system",
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "system:accounting-exporter",
			},
		},
	}
}
