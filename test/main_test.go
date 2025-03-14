package main_test

import (
	"log/slog"
	"os"
	"testing"

	kubeassert "github.com/DWSR/kubeassert-go"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/third_party/kind"
)

var testEnv env.Environment

func TestMain(m *testing.M) {
	kindClusterName := envconf.RandomName("kind", 16)

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	testEnv = env.New().
		Setup(
			envfuncs.CreateCluster(kind.NewProvider(), kindClusterName),
			kubeassert.ApplyKustomization("../overlays/test"),
		).
		Finish(
			envfuncs.DestroyCluster(kindClusterName),
		)

	os.Exit(testEnv.Run(m))
}

func Test_Crossplane(t *testing.T) {
	nsName := "crossplane-system"
	controllerDeployName := "crossplane"
	rbacManagerDeployName := "crossplane-rbac-manager"

	namespaceOpt := kubeassert.WithNamespace(nsName)

	assertions := []kubeassert.Assertion{
		kubeassert.NewNamespaceAssertion(
			kubeassert.WithResourceName(nsName),
		).Exists().IsRestricted(),
		kubeassert.NewDeploymentAssertion(
			kubeassert.WithResourceName(controllerDeployName),
			namespaceOpt,
			namespaceOpt,
		).Exists().IsAvailable(),
		kubeassert.NewDeploymentAssertion(
			kubeassert.WithResourceName(rbacManagerDeployName),
			namespaceOpt,
		).Exists().IsAvailable(),
		kubeassert.NewPDBAssertion(
			kubeassert.WithResourceName(controllerDeployName),
			namespaceOpt,
		).Exists(),
		kubeassert.NewPDBAssertion(
			kubeassert.WithResourceName(rbacManagerDeployName),
			namespaceOpt,
		).Exists(),
		kubeassert.NewCRDAssertion(
			kubeassert.WithResourceName("deploymentruntimeconfigs.pkg.crossplane.io"),
		).Exists().HasVersion("v1beta1"),
		kubeassert.NewCRDAssertion(
			kubeassert.WithResourceName("providers.pkg.crossplane.io"),
		).Exists().HasVersion("v1"),
		kubeassert.NewPodAssertion(
			namespaceOpt,
			kubeassert.WithLabels(map[string]string{"pkg.crossplane.io/provider": "provider-kubernetes"}),
			kubeassert.WithSetup(
				kubeassert.CreateResourceFromPath("./testdata/restricted-deployment.yaml"),
				kubeassert.CreateResourceFromPath("./testdata/kube-provider.yaml"),
			),
		).Exists().IsReady(),
	}

	kubeassert.TestAssertions(t, testEnv, assertions...)
}
