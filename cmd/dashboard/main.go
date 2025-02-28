/*
Copyright 2023 The Nuclio Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"os"

	"github.com/nuclio/nuclio/cmd/dashboard/app"
	"github.com/nuclio/nuclio/pkg/common"
	_ "github.com/nuclio/nuclio/pkg/dashboard/resource"

	"github.com/nuclio/errors"
)

func main() {
	defaultOffline := os.Getenv("NUCLIO_DASHBOARD_OFFLINE") == "true"

	externalIPAddressesDefault := os.Getenv("NUCLIO_DASHBOARD_EXTERNAL_IP_ADDRESSES")

	defaultPlatformAuthorizationMode := os.Getenv("NUCLIO_DASHBOARD_PLATFORM_AUTHORIZATION_MODE")
	if defaultPlatformAuthorizationMode == "" {
		defaultPlatformAuthorizationMode = "service-account"
	}

	// git templating env vars
	templatesGitRepository := flag.String("templates-git-repository", common.GetEnvOrDefaultString("NUCLIO_TEMPLATES_GIT_REPOSITORY", ""), "Git templates repo's name")
	templatesGitBranch := flag.String("templates-git-ref", common.GetEnvOrDefaultString("NUCLIO_TEMPLATES_GIT_REF", ""), "Git templates repo's branch name")
	templatesGitUsername := flag.String("templates-git-username", common.GetEnvOrDefaultString("NUCLIO_TEMPLATES_GIT_USERNAME", ""), "Git repo's username")
	templatesGitPassword := flag.String("templates-git-password", common.GetEnvOrDefaultString("NUCLIO_TEMPLATES_GIT_PASSWORD", ""), "Git repo's user password")
	templatesGithubAccessToken := flag.String("templates-github-access-token", common.GetEnvOrDefaultString("NUCLIO_TEMPLATES_GITHUB_ACCESS_TOKEN", ""), "Github templates repo's access token")
	templatesArchiveAddress := flag.String("templates-archive-address", common.GetEnvOrDefaultString("NUCLIO_TEMPLATES_ARCHIVE_ADDRESS", "file://tmp/templates.zip"), "Function Templates zip file address")
	templatesGitCaCertContents := flag.String("templates-git-ca-cert-contents", common.GetEnvOrDefaultString("NUCLIO_TEMPLATES_GIT_CA_CERT_CONTENTS", ""), "Base64 encoded ca certificate contents used in git requests to templates repo")

	listenAddress := flag.String("listen-addr", ":8070", "IP/port on which the dashboard listens")
	dockerKeyDir := flag.String("docker-key-dir", "", "Directory to look for docker keys for secure registries")
	platformType := flag.String("platform", common.AutoPlatformName, "One of kube/local/auto")
	defaultRegistryURL := flag.String("registry", os.Getenv("NUCLIO_DASHBOARD_REGISTRY_URL"), "Default registry URL")
	defaultRunRegistryURL := flag.String("run-registry", os.Getenv("NUCLIO_DASHBOARD_RUN_REGISTRY_URL"), "Default run registry URL")
	noPullBaseImages := flag.Bool("no-pull", common.GetEnvOrDefaultBool("NUCLIO_DASHBOARD_NO_PULL_BASE_IMAGES", false), "Whether to pull base images (Default: false)")
	credsRefreshInterval := flag.String("creds-refresh-interval", os.Getenv("NUCLIO_DASHBOARD_CREDS_REFRESH_INTERVAL"), "Default credential refresh interval, or 'none' (12h by default)")
	externalIPAddresses := flag.String("external-ip-addresses", externalIPAddressesDefault, "Comma delimited list of external IP addresses")
	namespace := flag.String("namespace", "", "Namespace in which all actions apply to, if not passed in request")
	offline := flag.Bool("offline", defaultOffline, "If true, assumes no internet connectivity")
	platformConfigurationPath := flag.String("platform-config", "/etc/nuclio/config/platform/platform.yaml", "Path of platform configuration file")
	imageNamePrefixTemplate := flag.String("image-name-prefix-template", os.Getenv("NUCLIO_DASHBOARD_IMAGE_NAME_PREFIX_TEMPLATE"), "Go template for the image names prefix")
	platformAuthorizationMode := flag.String("platform-authorization-mode", defaultPlatformAuthorizationMode, "One of service-account (default) / authorization-header-oidc")
	dependantImageRegistryURL := flag.String("dependant-image-registry", os.Getenv("NUCLIO_DASHBOARD_DEPENDANT_IMAGE_REGISTRY_URL"), "If passed, replaces base/on-build registry URLs with this value")
	monitorDockerDeamon := flag.Bool("monitor-docker-deamon", common.GetEnvOrDefaultBool("NUCLIO_MONITOR_DOCKER_DAEMON", true), "Monitor connectivity to docker deamon (in conjunction to 'docker' as container builder kind")
	monitorDockerDeamonIntervalStr := flag.String("monitor-docker-deamon-interval", common.GetEnvOrDefaultString("NUCLIO_MONITOR_DOCKER_DAEMON_INTERVAL", "5s"), "Docker deamon connectivity monitor interval (used in conjunction with 'monitor-docker-deamon')")
	monitorDockerDeamonMaxConsecutiveErrorsStr := flag.String("monitor-docker-deamon-max-consecutive-errors", common.GetEnvOrDefaultString("NUCLIO_MONITOR_DOCKER_DAEMON_MAX_CONSECUTIVE_ERRORS", "5"), "Docker deamon connectivity monitor max consecutive errors before declaring docker connection is unhealthy (used in conjunction with 'monitor-docker-deamon')")

	// auth options
	authConfigKind := flag.String("auth-config-kind", common.GetEnvOrDefaultString("NUCLIO_AUTH_KIND", "nop"), "Authentication kind, either nop or iguazio")
	authConfigIguazioVerificationURL := flag.String("auth-config-iguazio-verification-url", common.GetEnvOrDefaultString("NUCLIO_AUTH_IGUAZIO_VERIFICATION_URL", ""), "Iguazio authentication verification url")
	authConfigIguazioVerificationMethod := flag.String("auth-config-iguazio-verification-method", common.GetEnvOrDefaultString("NUCLIO_AUTH_IGUAZIO_VERIFICATION_METHOD", "POST"), "Iguazio authentication verification method")
	authConfigIguazioVerificationDataEnrichmentURL := flag.String("auth-config-iguazio-verification-data-enrichment-url", common.GetEnvOrDefaultString("NUCLIO_AUTH_IGUAZIO_VERIFICATION_DATA_ENRICHMENT_URL", ""), "Iguazio authentication verification and data enrichment url")
	authConfigIguazioTimeout := flag.String("auth-config-iguazio-timeout", common.GetEnvOrDefaultString("NUCLIO_AUTH_IGUAZIO_TIMEOUT", ""), "Iguazio authentication request timeout (golang duration string)")
	authConfigIguazioCacheSize := flag.String("auth-config-iguazio-cache-size", common.GetEnvOrDefaultString("NUCLIO_AUTH_IGUAZIO_CACHE_SIZE", ""), "Iguazio authentication cache size")
	authConfigIguazioCacheTimeout := flag.String("auth-config-iguazio-cache-expiration-timeout", common.GetEnvOrDefaultString("NUCLIO_AUTH_IGUAZIO_CACHE_EXPIRATION_TIMEOUT", "30s"), "Iguazio authentication cache expiration timeout (golang duration string)")

	// get the namespace from args -> env -> default
	*namespace = common.ResolveNamespace(*namespace, "NUCLIO_DASHBOARD_NAMESPACE")

	flag.Parse()

	if err := app.Run(
		*listenAddress,
		*dockerKeyDir,
		*defaultRegistryURL,
		*defaultRunRegistryURL,
		*platformType,
		*noPullBaseImages,
		*credsRefreshInterval,
		*externalIPAddresses,
		*namespace,
		*offline,
		*platformConfigurationPath,
		*templatesGitRepository,
		*templatesGitBranch,
		*templatesArchiveAddress,
		*templatesGitUsername,
		*templatesGitPassword,
		*templatesGithubAccessToken,
		*templatesGitCaCertContents,
		*imageNamePrefixTemplate,
		*platformAuthorizationMode,
		*dependantImageRegistryURL,
		*monitorDockerDeamon,
		*monitorDockerDeamonIntervalStr,
		*monitorDockerDeamonMaxConsecutiveErrorsStr,
		*authConfigKind,
		*authConfigIguazioTimeout,
		*authConfigIguazioVerificationURL,
		*authConfigIguazioVerificationDataEnrichmentURL,
		*authConfigIguazioCacheSize,
		*authConfigIguazioCacheTimeout,
		*authConfigIguazioVerificationMethod,
	); err != nil {

		errors.PrintErrorStack(os.Stderr, err, 5)

		os.Exit(1)
	}

	os.Exit(0)
}
