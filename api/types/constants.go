/*
Copyright 2020-2021 Gravitational, Inc.

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

package types

import (
	"github.com/gravitational/teleport/api/types/common"
)

const (
	// The `Kind*` constants in this const block identify resource kinds used for
	// storage an/or and access control. Please keep these in sync with the
	// `ResourceKind` enum in
	// `web/packages/teleport/src/services/resources/types.ts`.

	// DefaultAPIGroup is a default group of permissions API,
	// lets us to add different permission types
	DefaultAPIGroup = "gravitational.io/teleport"

	// DefaultReleaseServerAddr is the default release service URL
	DefaultReleaseServerAddr = "rlz.teleport.sh"

	// ReleaseServerEnvVar is the environment variable used to overwrite
	// the default release server address
	ReleaseServerEnvVar = "RELEASE_SERVER_HOSTPORT"

	// EnterpriseReleaseEndpoint is the endpoint of Teleport Enterprise
	// releases on the release server
	EnterpriseReleaseEndpoint = "teleport-ent"

	// PackageNameOSS is the teleport package name for the OSS version.
	PackageNameOSS = "teleport"
	// PackageNameEnt is the teleport package name for the Enterprise version.
	PackageNameEnt = "teleport-ent"
	// PackageNameEntFIPS is the teleport package name for the Enterprise with FIPS enabled version.
	PackageNameEntFIPS = "teleport-ent-fips"

	// ActionRead grants read access (get, list)
	ActionRead = "read"

	// ActionWrite allows to write (create, update, delete)
	ActionWrite = "write"

	// Wildcard is a special wildcard character matching everything
	Wildcard = "*"

	// True holds "true" string value
	True = "true"

	// HomeEnvVar specifies the home location for tsh configuration
	// and data
	HomeEnvVar = "TELEPORT_HOME"

	// KindNamespace is a namespace
	KindNamespace = "namespace"

	// KindUser is a user resource
	KindUser = "user"

	// KindBot is a Machine ID bot resource
	KindBot = "bot"
	// KindBotInstance is an instance of a Machine ID bot
	KindBotInstance = "bot_instance"
	// KindSPIFFEFederation is a SPIFFE federation resource
	KindSPIFFEFederation = "spiffe_federation"

	// KindHostCert is a host certificate
	KindHostCert = "host_cert"

	// KindJWT is a JWT token signer.
	KindJWT = "jwt"

	// KindLicense is a license resource
	KindLicense = "license"

	// KindRole is a role resource
	KindRole = "role"

	// KindAccessRequest is an AccessRequest resource
	KindAccessRequest = "access_request"

	// KindAccessMonitoringRule is an access monitoring rule resource
	KindAccessMonitoringRule = "access_monitoring_rule"

	// KindPluginData is a PluginData resource
	KindPluginData = "plugin_data"

	// KindAccessPluginData is a resource directive that applies
	// only to plugin data associated with access requests.
	KindAccessPluginData = "access_plugin_data"

	// KindOIDC is OIDC connector resource
	KindOIDC = "oidc"

	// KindSAML is SAML connector resource
	KindSAML = "saml"

	// KindGithub is Github connector resource
	KindGithub = "github"

	// KindOIDCRequest is OIDC auth request resource
	KindOIDCRequest = "oidc_request"

	// KindSAMLRequest is SAML auth request resource
	KindSAMLRequest = "saml_request"

	// KindGithubRequest is Github auth request resource
	KindGithubRequest = "github_request"

	// KindSession is a recorded SSH session.
	KindSession = "session"

	// KindSSHSession represents an active SSH session in early versions of Teleport
	// prior to the introduction of moderated sessions. Note that ssh_session is not
	// a "real" resource, and it is never used as the "session kind" value in the
	// session_tracker resource.
	KindSSHSession = "ssh_session"

	// KindWebSession is a web session resource
	KindWebSession = "web_session"

	// KindWebToken is a web token resource
	KindWebToken = "web_token"

	// KindAppSession represents an application specific web session.
	KindAppSession = "app_session"

	// KindSnowflakeSession represents a Snowflake specific web session.
	KindSnowflakeSession = "snowflake_session"

	// KindEvent is structured audit logging event
	KindEvent = "event"

	// KindAuthServer is auth server resource
	KindAuthServer = "auth_server"

	// KindProxy is proxy resource
	KindProxy = "proxy"

	// KindNode is node resource. It can be either a Teleport node or
	// a registered OpenSSH (agentless) node.
	KindNode = "node"

	// SubKindTeleportNode is a Teleport node.
	SubKindTeleportNode = "teleport"

	// SubKindOpenSSHNode is a registered OpenSSH (agentless) node.
	SubKindOpenSSHNode = "openssh"

	// SubKindOpenSSHEICENode is a registered OpenSSH (agentless) node that doesn't require trust in Teleport CA.
	// For each session an SSH Key is created and uploaded to the target host using a side-channel.
	//
	// For Amazon EC2 Instances, it uploads the key using:
	// https://docs.aws.amazon.com/ec2-instance-connect/latest/APIReference/API_SendSSHPublicKey.html
	// This Key is valid for 60 seconds.
	//
	// It uses the private key created above to SSH into the host.
	SubKindOpenSSHEICENode = "openssh-ec2-ice"

	// KindUnifiedResource is a meta Kind that is used for the unified resource search present on
	// the webUI and Connect. It allows us to query and return multiple kinds at the same time
	KindUnifiedResource = "unified_resource"

	// KindAppServer is an application server resource.
	KindAppServer = "app_server"

	// KindApp is a web app resource.
	KindApp = "app"

	// SubKindMCP represents an MCP server as a subkind of app.
	SubKindMCP = KindMCP

	// KindMCP is an MCP server resource.
	// Currently, MCP servers are accessed through apps.
	// In the future, they may become a standalone resource kind.
	KindMCP = "mcp"

	// KindDatabaseServer is a database proxy server resource.
	KindDatabaseServer = "db_server"

	// KindDatabaseService is a database service resource.
	KindDatabaseService = "db_service"

	// KindDatabase is a database resource.
	KindDatabase = "db"

	// KindDatabaseObjectImportRule is a database object import rule resource.
	KindDatabaseObjectImportRule = "db_object_import_rule"

	// KindDatabaseObject is a database object resource.
	KindDatabaseObject = "db_object"

	// KindKubeServer is an kubernetes server resource.
	KindKubeServer = "kube_server"
	// KindCrownJewel is a crown jewel resource
	KindCrownJewel = "crown_jewel"
	// KindKubernetesCluster is a Kubernetes cluster.
	KindKubernetesCluster = "kube_cluster"
	// KindKubernetesResource is a Kubernetes resource within a cluster.
	KindKubernetesResource = "kube_resource"

	// KindKubePod is a Kubernetes Pod resource type.
	KindKubePod = "pod"

	// KindKubeSecret is a Kubernetes Secret resource type.
	KindKubeSecret = "secret"

	// KindKubeConfigMap is a Kubernetes Configmap resource type.
	KindKubeConfigmap = "configmap"

	// KindKubeNamespace is a Kubernetes namespace resource type.
	KindKubeNamespace = "namespace"

	// KindKubeService is a Kubernetes Service resource type.
	KindKubeService = "service"

	// KindKubeServiceAccount is an Kubernetes Service Account resource type.
	KindKubeServiceAccount = "serviceaccount"

	// KindKubeNode is a Kubernetes Node resource type.
	KindKubeNode = "kube_node"

	// KindKubePersistentVolume is a Kubernetes Persistent Volume resource type.
	KindKubePersistentVolume = "persistentvolume"

	// KindKubePersistentVolumeClaim is a Kubernetes Persistent Volume Claim resource type.
	KindKubePersistentVolumeClaim = "persistentvolumeclaim"

	// KindKubeDeployment is a Kubernetes Deployment resource type.
	KindKubeDeployment = "deployment"

	// KindKubeReplicaSet is a Kubernetes Replicaset resource type.
	KindKubeReplicaSet = "replicaset"

	// KindKubeReplicationController is a Kubernetes ReplicationController resource type.
	KindKubeReplicationController = "replicationcontroller"

	// KindKubeStatefulset is a Kubernetes Statefulset resource type.
	KindKubeStatefulset = "statefulset"

	// KindKubeDaemonSet is a Kubernetes Daemonset resource type.
	KindKubeDaemonSet = "daemonset"

	// KindKubeClusterRole is a Kubernetes ClusterRole resource type.
	KindKubeClusterRole = "clusterrole"

	// KindKubeRole is a Kubernetes Role resource type.
	KindKubeRole = "kube_role"

	// KindKubeClusterRoleBinding is a Kubernetes Cluster Role Binding resource type.
	KindKubeClusterRoleBinding = "clusterrolebinding"

	// KindKubeRoleBinding is a Kubernetes Role Binding resource type.
	KindKubeRoleBinding = "rolebinding"

	// KindKubeCronjob is a Kubernetes Cronjob resource type.
	KindKubeCronjob = "cronjob"

	// KindKubeJob is a Kubernetes job resource type.
	KindKubeJob = "job"

	// KindKubeCertificateSigningRequest is a Certificate Signing Request resource type.
	KindKubeCertificateSigningRequest = "certificatesigningrequest"

	// KindKubeIngress is a Kubernetes Ingress resource type.
	KindKubeIngress = "ingress"

	// KindKubeWaitingContainer is a Kubernetes ephemeral
	// container that are waiting to be created until moderated
	// session conditions are met.
	KindKubeWaitingContainer = "kube_ephemeral_container"

	// KindToken is a provisioning token resource
	KindToken = "token"

	// KindCertAuthority is a certificate authority resource
	KindCertAuthority = "cert_authority"

	// KindReverseTunnel is a reverse tunnel connection
	KindReverseTunnel = "tunnel"

	// KindOIDCConnector is a OIDC connector resource
	KindOIDCConnector = "oidc"

	// KindSAMLConnector is a SAML connector resource
	KindSAMLConnector = "saml"

	// KindGithubConnector is Github OAuth2 connector resource
	KindGithubConnector = "github"

	// KindConnectors is a shortcut for all authentication connector
	KindConnectors = "connectors"

	// KindClusterAuthPreference is the type of authentication for this cluster.
	KindClusterAuthPreference = "cluster_auth_preference"

	// MetaNameClusterAuthPreference is the type of authentication for this cluster.
	MetaNameClusterAuthPreference = "cluster-auth-preference"

	// KindSessionRecordingConfig is the resource for session recording configuration.
	KindSessionRecordingConfig = "session_recording_config"

	// KindRecordingEncryption is the collection of active session recording encryption keys.
	KindRecordingEncryption = "recording_encryption"

	// MetaNameSessionRecordingConfig is the exact name of the singleton resource for
	// session recording configuration.
	MetaNameSessionRecordingConfig = "session-recording-config"

	// MetaNameRecordingEncryption is the exact name of the singleton resource for
	// session recording configuration.
	MetaNameRecordingEncryption = "recording-encryption"

	// KindExternalAuditStorage the resource kind for External Audit Storage
	// configuration.
	KindExternalAuditStorage = "external_audit_storage"
	// MetaNameExternalAuditStorageDraft is the exact name of the singleton resource
	// holding External Audit Storage draft configuration.
	MetaNameExternalAuditStorageDraft = "draft"
	// MetaNameExternalAuditStorageCluster is the exact name of the singleton resource
	// holding External Audit Storage cluster configuration.
	MetaNameExternalAuditStorageCluster = "cluster"

	// KindClusterConfig is the resource that holds cluster level configuration.
	// Deprecated: This does not correspond to an actual resource anymore but is
	// still used when checking access to the new configuration resources, as an
	// alternative to their individual resource kinds.
	KindClusterConfig = "cluster_config"

	// KindAutoUpdateConfig is the resource with autoupdate configuration.
	KindAutoUpdateConfig = "autoupdate_config"

	// KindAutoUpdateVersion is the resource with autoupdate versions.
	KindAutoUpdateVersion = "autoupdate_version"

	// KindAutoUpdateAgentRollout is the resource that controls and tracks agent rollouts.
	KindAutoUpdateAgentRollout = "autoupdate_agent_rollout"

	// KindAutoUpdateAgentReport is the resource that tracks connected agents.
	KindAutoUpdateAgentReport = "autoupdate_agent_report"

	// MetaNameAutoUpdateConfig is the name of a configuration resource for autoupdate config.
	MetaNameAutoUpdateConfig = "autoupdate-config"

	// MetaNameAutoUpdateVersion is the name of a resource for autoupdate version.
	MetaNameAutoUpdateVersion = "autoupdate-version"

	// MetaNameAutoUpdateAgentRollout is the name of the autoupdate agent rollout resource.
	MetaNameAutoUpdateAgentRollout = "autoupdate-agent-rollout"

	// KindClusterAuditConfig is the resource that holds cluster audit configuration.
	KindClusterAuditConfig = "cluster_audit_config"

	// MetaNameClusterAuditConfig is the exact name of the singleton resource holding
	// cluster audit configuration.
	MetaNameClusterAuditConfig = "cluster-audit-config"

	// MetaNameUIConfig is the exact name of the singleton resource holding
	// proxy service UI configuration.
	MetaNameUIConfig = "ui-config"

	// KindClusterNetworkingConfig is the resource that holds cluster networking configuration.
	KindClusterNetworkingConfig = "cluster_networking_config"

	// MetaNameClusterNetworkingConfig is the exact name of the singleton resource holding
	// cluster networking configuration.
	MetaNameClusterNetworkingConfig = "cluster-networking-config"

	// KindSemaphore is the resource that provides distributed semaphore functionality
	KindSemaphore = "semaphore"

	// KindClusterName is a type of configuration resource that contains the cluster name.
	KindClusterName = "cluster_name"

	// MetaNameClusterName is the name of a configuration resource for cluster name.
	MetaNameClusterName = "cluster-name"

	// MetaNameWatchStatus is the name of a watch status resource.
	MetaNameWatchStatus = "watch-status"

	// KindStaticTokens is a type of configuration resource that contains static tokens.
	KindStaticTokens = "static_tokens"

	// MetaNameStaticTokens is the name of a configuration resource for static tokens.
	MetaNameStaticTokens = "static-tokens"

	// MetaNameSessionTracker is the prefix of resources used to track live sessions.
	MetaNameSessionTracker = "session-tracker"

	// KindTrustedCluster is a resource that contains trusted cluster configuration.
	KindTrustedCluster = "trusted_cluster"

	// KindAuthConnector allows access to OIDC and SAML connectors.
	KindAuthConnector = "auth_connector"

	// KindTunnelConnection specifies connection of a reverse tunnel to proxy
	KindTunnelConnection = "tunnel_connection"

	// KindRemoteCluster represents remote cluster connected via reverse tunnel
	// to proxy
	KindRemoteCluster = "remote_cluster"

	// KindUserToken is a user token used for various user related actions.
	KindUserToken = "user_token"

	// KindUserTokenSecrets is user token secrets.
	KindUserTokenSecrets = "user_token_secrets"

	// KindIdentity is local on disk identity resource
	KindIdentity = "identity"

	// KindState is local on disk process state
	KindState = "state"

	// KindMFADevice is an MFA device for a user.
	KindMFADevice = "mfa_device"

	// KindBilling represents access to cloud billing features
	KindBilling = "billing"

	// KindLock is a lock resource.
	KindLock = "lock"

	// KindNetworkRestrictions are restrictions for SSH sessions
	KindNetworkRestrictions = "network_restrictions"

	// MetaNameNetworkRestrictions is the exact name of the singleton resource for
	// network restrictions
	MetaNameNetworkRestrictions = "network-restrictions"

	// KindWindowsDesktopService is a Windows desktop service resource.
	KindWindowsDesktopService = "windows_desktop_service"

	// KindWindowsDesktop is a Windows desktop host.
	KindWindowsDesktop = "windows_desktop"

	// KindDynamicWindowsDesktop is a dynamic Windows desktop host.
	KindDynamicWindowsDesktop = "dynamic_windows_desktop"

	// KindRecoveryCodes is a resource that holds users recovery codes.
	KindRecoveryCodes = "recovery_codes"

	// KindSessionTracker is a resource that tracks a live session.
	KindSessionTracker = "session_tracker"

	// KindConnectionDiagnostic is a resource that tracks the result of testing a connection
	KindConnectionDiagnostic = "connection_diagnostic"

	// KindDatabaseCertificate is a resource to control db CA cert
	// generation.
	KindDatabaseCertificate = "database_certificate"

	// KindInstaller is a resource that holds a node installer script
	// used to install teleport on discovered nodes
	KindInstaller = "installer"

	// KindUIConfig is a resource that holds configuration for the UI
	// served by the proxy service
	KindUIConfig = "ui_config"

	// KindClusterAlert is a resource that conveys a cluster-level alert message.
	KindClusterAlert = "cluster_alert"

	// KindDevice represents a registered or trusted device.
	KindDevice = "device"

	// KindDownload represents Teleport binaries downloads.
	KindDownload = "download"

	// KindUsageEvent is an external cluster usage event. Similar to
	// KindHostCert, this kind is not backed by a real resource.
	KindUsageEvent = "usage_event"

	// KindInstance represents a teleport instance independent of any specific service.
	KindInstance = "instance"

	// KindLoginRule is a login rule resource.
	KindLoginRule = "login_rule"

	// KindPlugin represents a plugin instance
	KindPlugin = "plugin"

	// KindPluginStaticCredentials represents plugin static credentials.
	KindPluginStaticCredentials = "plugin_static_credentials"

	// KindSAMLIdPServiceProvider is a SAML service provider for the built in Teleport IdP.
	KindSAMLIdPServiceProvider = "saml_idp_service_provider"

	// KindUserGroup is an externally sourced user group.
	KindUserGroup = "user_group"

	// KindOktaImportRule is a rule for importing Okta objects.
	KindOktaImportRule = "okta_import_rule"

	// KindOktaAssignment is a set of actions to apply to Okta.
	KindOktaAssignment = "okta_assignment"

	// KindHeadlessAuthentication is a headless authentication resource.
	KindHeadlessAuthentication = "headless_authentication"

	// KindHealthCheckConfig is the resource for health check configuration.
	KindHealthCheckConfig = "health_check_config"

	// KindAccessGraph is the RBAC kind for access graph.
	KindAccessGraph = "access_graph"

	// KindIntegration is a connection to a 3rd party system API.
	KindIntegration = "integration"

	// KindUserTask is a task representing an issue with some other resource.
	KindUserTask = "user_task"

	// KindClusterMaintenanceConfig determines maintenance times for the cluster.
	KindClusterMaintenanceConfig = "cluster_maintenance_config"

	// KindServerInfo contains info that should be applied to joining Nodes.
	KindServerInfo = "server_info"

	// KindBackendInfo contains backend info.
	KindBackendInfo = "backend_info"

	// MetaNameBackendInfo name backend info entity.
	MetaNameBackendInfo = "backend-info"

	// SubKindCloudInfo is a ServerInfo that was created by the Discovery
	// service to match with a single discovered instance.
	SubKindCloudInfo = "cloud_info"

	// MetaNameClusterMaintenanceConfig is the only allowed metadata.name value for the maintenance
	// window singleton resource.
	MetaNameClusterMaintenanceConfig = "cluster-maintenance-config"

	// KindWatchStatus is a kind for WatchStatus resource which contains information about a successful Watch request.
	KindWatchStatus = "watch_status"

	// KindAccessList is an AccessList resource
	KindAccessList = "access_list"

	// KindUserLoginState is a UserLoginState resource
	KindUserLoginState = "user_login_state"

	// KindAccessListMember is an AccessListMember resource
	KindAccessListMember = "access_list_member"

	// KindAccessListReview is an AccessListReview resource
	KindAccessListReview = "access_list_review"

	// KindDiscoveryConfig is a DiscoveryConfig resource.
	// Used for adding additional matchers in Discovery Service.
	KindDiscoveryConfig = "discovery_config"
	// KindAuditQuery is an AuditQuery resource.
	KindAuditQuery = "audit_query"
	// KindSecurityReport is a SecurityReport resource.
	KindSecurityReport = "security_report"
	// KindSecurityReportState is a SecurityReportState resource.
	KindSecurityReportState = "security_report_state"
	// KindSecurityReportCostLimiter const limiter
	KindSecurityReportCostLimiter = "security_report_cost_limiter"

	// KindNotification is a notification resource.
	KindNotification = "notification"
	// KindGlobalNotification is a global notification resource.
	KindGlobalNotification = "global_notification"
	// KindUserLastSeenNotification is a resource which stores the timestamp of a user's last seen notification.
	KindUserLastSeenNotification = "user_last_seen_notification"
	// KindUserNotificationState is a resource which tracks whether a user has clicked on or dismissed a notification.
	KindUserNotificationState = "user_notification_state"
	// KindUniqueNotificationIdentifier is a resource which tracks a unique identifier for a notification and is used to prevent duplicate notifications in certain cases.
	KindUniqueNotificationIdentifier = "unique_notification_identifier"

	// KindAccessGraphSecretAuthorizedKey is a authorized key entry found in
	// a Teleport SSH node type.
	KindAccessGraphSecretAuthorizedKey = "access_graph_authorized_key"

	// KindAccessGraphSecretPrivateKey is a private key entry found in
	// a managed device.
	KindAccessGraphSecretPrivateKey = "access_graph_private_key"

	// KindProvisioningPrincipalState is a resource that tracks provisioning of a user or access
	// list in a downstream SCIM server
	KindProvisioningPrincipalState = "provisioning_principal_state"

	// KindVnetConfig is a resource which holds cluster-wide configuration for VNet.
	KindVnetConfig = "vnet_config"

	// KindAccessGraphSettings is a resource which holds cluster-wide configuration for dynamic access graph settings.
	KindAccessGraphSettings = "access_graph_settings"

	// KindStaticHostUser is a host user to be created on matching SSH nodes.
	KindStaticHostUser = "static_host_user"

	// KindIdentityCenter is an umbrella kind, representing all KindIdentityCenter*
	// resource kinds in RBAC checks. This is to simplify Role condition statements
	// so that they don't have to individually specify all of the Identity Center
	// resource kinds.
	KindIdentityCenter = "aws_identity_center"

	// KindIdentityCenterAccount describes an Identity-Center managed AWS Account
	KindIdentityCenterAccount = "aws_ic_account"

	// KindIdentityCenterPermissionSet describes an AWS Identity Center Permission Set
	KindIdentityCenterPermissionSet = "aws_ic_permission_set"

	// KindIdentityCenterPermissionSet describes an AWS Principal Assignment, representing
	// a collection Account Assignments assigned to a Teleport User or AccessList
	KindIdentityCenterPrincipalAssignment = "aws_ic_principal_assignment"

	// KindIdentityCenterAccountAssignment describes an AWS Account and Permission Set
	// pair that can be requested by a Teleport User.
	KindIdentityCenterAccountAssignment = "aws_ic_account_assignment"

	// KindContact is a resource that holds contact information
	// for Teleport Enterprise customers.
	KindContact = "contact"

	// KindWorkloadIdentity is the WorkloadIdentity resource.
	KindWorkloadIdentity = "workload_identity"

	// KindWorkloadIdentityX509Revocation is the WorkloadIdentityX509Revocation
	// resource.
	KindWorkloadIdentityX509Revocation = "workload_identity_x509_revocation"

	// KindWorkloadIdentityX509IssuerOverride is the kind of
	// teleport.workloadidentity.v1.X509IssuerOverride.
	KindWorkloadIdentityX509IssuerOverride = "workload_identity_x509_issuer_override"

	// KindWorkloadIdentityX509IssuerOverrideCSR is the pseudo-kind representing
	// the act of signing CSRs on behalf of the SPIFFE CA (with [VerbCreate]).
	KindWorkloadIdentityX509IssuerOverrideCSR = "workload_identity_x509_issuer_override_csr"

	// KindSigstorePolicy is the kind of teleport.workloadidentity.v1.SigstorePolicy.
	KindSigstorePolicy = "sigstore_policy"

	// KindGitServer represents a Git server that can proxy git commands.
	KindGitServer = "git_server"
	// SubKindGitHub specifies the GitHub subkind of a Git server.
	SubKindGitHub = "github"

	// KindStableUNIXUser is the RBAC-only kind to refer to interactions with
	// stable UNIX users.
	KindStableUNIXUser = "stable_unix_user"

	// KindInferenceModel is the kind of teleport.summarizer.v1.InferenceModel.
	KindInferenceModel = "inference_model"

	// KindInferenceSecret is the kind of teleport.summarizer.v1.InferenceSecret.
	KindInferenceSecret = "inference_secret"

	// KindInferencePolicy is the kind of teleport.summarizer.v1.InferencePolicy.
	KindInferencePolicy = "inference_policy"

	// MetaNameAccessGraphSettings is the exact name of the singleton resource holding
	// access graph settings.
	MetaNameAccessGraphSettings = "access-graph-settings"

	// MetaNameVnetConfig is the exact name of the singleton resource holding VNet config.
	MetaNameVnetConfig = "vnet-config"

	// KindRelayServer is the resource kind for a Relay service heartbeat.
	KindRelayServer = "relay_server"

	// V8 is the eighth version of resources.
	V8 = "v8"

	// V7 is the seventh version of resources.
	V7 = "v7"

	// V6 is the sixth version of resources.
	V6 = "v6"

	// V5 is the fifth version of resources.
	V5 = "v5"

	// V4 is the fourth version of resources.
	V4 = "v4"

	// V3 is the third version of resources.
	V3 = "v3"

	// V2 is the second version of resources.
	V2 = "v2"

	// V1 is the first version of resources. Note: The first version was
	// not explicitly versioned.
	V1 = "v1"
)

// PackageNameKinds is the list of valid teleport package names.
var PackageNameKinds = []string{PackageNameOSS, PackageNameEnt, PackageNameEntFIPS}

// WebSessionSubKinds lists subkinds of web session resources
var WebSessionSubKinds = []string{KindAppSession, KindWebSession, KindSnowflakeSession}

const (
	// VerbList is used to list all objects. Does not imply the ability to read a single object.
	VerbList = "list"

	// VerbCreate is used to create an object.
	VerbCreate = "create"

	// VerbRead is used to read a single object.
	VerbRead = "read"

	// VerbReadNoSecrets is used to read a single object without secrets.
	VerbReadNoSecrets = "readnosecrets"

	// VerbUpdate is used to update an object.
	VerbUpdate = "update"

	// VerbDelete is used to remove an object.
	VerbDelete = "delete"

	// VerbRotate is used to rotate certificate authorities
	// used only internally
	VerbRotate = "rotate"

	// VerbCreateEnrollToken allows the creation of device enrollment tokens.
	// Device Trust is a Teleport Enterprise feature.
	VerbCreateEnrollToken = "create_enroll_token"

	// VerbEnroll allows enrollment of trusted devices.
	// Device Trust is a Teleport Enterprise feature.
	VerbEnroll = "enroll"

	// VerbUse allows the usage of an Integration.
	// Roles with this verb can issue API calls using the integration.
	VerbUse = "use"
)

const (
	// TeleportNamespace is used as the namespace prefix for labels defined by Teleport which can
	// carry metadata such as cloud AWS account or instance. Those labels can be used for RBAC.
	//
	// If a label with this prefix is used in a config file, the associated feature must take into
	// account that the label might be removed, modified or could have been set by the user.
	//
	// See also TeleportInternalLabelPrefix and TeleportHiddenLabelPrefix.
	TeleportNamespace = common.TeleportNamespace

	// OriginLabel is a resource metadata label name used to identify a source
	// that the resource originates from.
	OriginLabel = common.OriginLabel

	// ClusterLabel is a label that identifies the current cluster when creating resources on another systems.
	// Eg, when creating a resource in AWS, this label must be set as a Tag in the resource.
	ClusterLabel = TeleportNamespace + "/cluster"

	// ADLabel is a resource metadata label name used to identify if resource is part of Active Directory
	ADLabel = TeleportNamespace + "/ad"

	// OriginDefaults is an origin value indicating that the resource was
	// constructed as a default value.
	OriginDefaults = common.OriginDefaults

	// OriginConfigFile is an origin value indicating that the resource is
	// derived from static configuration.
	OriginConfigFile = common.OriginConfigFile

	// OriginDynamic is an origin value indicating that the resource was
	// committed as dynamic configuration.
	OriginDynamic = common.OriginDynamic

	// OriginCloud is an origin value indicating that the resource was
	// imported from a cloud provider.
	OriginCloud = common.OriginCloud

	// OriginKubernetes is an origin value indicating that the resource was
	// created from the Kubernetes Operator.
	OriginKubernetes = common.OriginKubernetes

	// OriginOkta is an origin value indicating that the resource was
	// created from the Okta service.
	OriginOkta = common.OriginOkta

	// OriginIntegrationAWSOIDC is an origin value indicating that the resource was
	// created from the AWS OIDC Integration.
	OriginIntegrationAWSOIDC = common.OriginIntegrationAWSOIDC

	// OriginIntegrationAWSRolesAnywhere is an origin value indicating that the resource was
	// created from the AWS IAM Roles Anywhere Integration.
	OriginIntegrationAWSRolesAnywhere = common.OriginIntegrationAWSRolesAnywhere

	// OriginDiscoveryKubernetes indicates that the resource was imported
	// from kubernetes cluster by discovery service.
	OriginDiscoveryKubernetes = common.OriginDiscoveryKubernetes

	// OriginEntraID indicates that the resource was imported
	// from the Entra ID directory.
	OriginEntraID = common.OriginEntraID

	// IntegrationLabel is a resource metadata label name used to identify the integration name that created the resource.
	IntegrationLabel = TeleportNamespace + "/integration"

	// AWSAccountIDLabel is used to identify nodes by AWS account ID
	// found via automatic discovery, to avoid re-running installation
	// commands on the node.
	AWSAccountIDLabel = TeleportNamespace + "/account-id"
	// AWSInstanceIDLabel is used to identify nodes by EC2 instance ID
	// found via automatic discovery, to avoid re-running installation
	// commands on the node.
	AWSInstanceIDLabel = TeleportNamespace + "/instance-id"
	// AWSInstanceRegion is used to identify the region an EC2
	// instance is running in
	AWSInstanceRegion = TeleportNamespace + "/aws-region"
	// SubscriptionIDLabel is used to identify virtual machines by Azure
	// subscription ID found via automatic discovery, to avoid re-running
	// installation commands on the node.
	SubscriptionIDLabel = TeleportInternalLabelPrefix + "subscription-id"
	// VMIDLabel is used to identify virtual machines by ID found
	// via automatic discovery, to avoid re-running installation commands
	// on the node.
	VMIDLabel = TeleportInternalLabelPrefix + "vm-id"
	// projectIDLabelSuffix is the identifier for adding the GCE ProjectID to an instance.
	projectIDLabelSuffix = "project-id"
	// ProjectIDLabelDiscovery is used to identify virtual machines by GCP project
	// id found via automatic discovery, to avoid re-running
	// installation commands on the node.
	ProjectIDLabelDiscovery = TeleportInternalLabelPrefix + projectIDLabelSuffix
	// ProjectIDLabel is used to identify the project ID for a virtual machine in GCP.
	// The difference between this and ProjectIDLabelDiscovery, is that this one will be visible to the user
	// and can be used in RBAC checks.
	ProjectIDLabel = TeleportNamespace + "/" + projectIDLabelSuffix
	// RegionLabel is used to identify virtual machines by region found
	// via automatic discovery, to avoid re-running installation commands
	// on the node.
	RegionLabel = TeleportInternalLabelPrefix + "region"
	// ResourceGroupLabel is used to identify virtual machines by resource-group found
	// via automatic discovery, to avoid re-running installation commands
	// on the node.
	ResourceGroupLabel = TeleportInternalLabelPrefix + "resource-group"
	// ZoneLabelDiscovery is used to identify virtual machines by GCP zone
	// found via automatic discovery, to avoid re-running installation
	// commands on the node.
	ZoneLabelDiscovery = TeleportInternalLabelPrefix + "zone"
	// NameLabelDiscovery is used to identify virtual machines by GCP VM name
	// found via automatic discovery, to avoid re-running installation
	// commands on the node.
	NameLabelDiscovery = TeleportInternalLabelPrefix + "name"

	// CloudLabel is used to identify the cloud where the resource was discovered.
	CloudLabel = TeleportNamespace + "/cloud"

	// DatabaseAdminLabel is used to identify database admin user for auto-
	// discovered databases.
	DatabaseAdminLabel = TeleportNamespace + "/db-admin"

	// DatabaseAdminDefaultDatabaseLabel is used to identify the database that
	// the admin user logs into by default.
	DatabaseAdminDefaultDatabaseLabel = TeleportNamespace + "/db-admin-default-database"

	// cloudKubeClusterNameOverrideLabel is a cloud agnostic label key for
	// overriding kubernetes cluster name in discovered cloud kube clusters.
	// It's used for AWS, GCP, and Azure, but not exported to decouple the
	// cloud-specific labels from eachother.
	cloudKubeClusterNameOverrideLabel = "TeleportKubernetesName"

	// cloudDatabaseNameOverrideLabel is a cloud agnostic label key for
	// overriding the database name in discovered cloud databases.
	// It's used for AWS, GCP, and Azure, but not exported to decouple the
	// cloud-specific labels from eachother.
	cloudDatabaseNameOverrideLabel = "TeleportDatabaseName"

	// AzureDatabaseNameOverrideLabel is the label key containing the database
	// name override for discovered Azure databases.
	// Azure tags cannot contain these characters: "<>%&\?/", so it doesn't
	// start with the namespace prefix.
	AzureDatabaseNameOverrideLabel = cloudDatabaseNameOverrideLabel

	// AzureKubeClusterNameOverrideLabel is the label key containing the
	// kubernetes cluster name override for discovered Azure kube clusters.
	AzureKubeClusterNameOverrideLabel = cloudKubeClusterNameOverrideLabel

	// GCPKubeClusterNameOverrideLabel is the label key containing the
	// kubernetes cluster name override for discovered GCP kube clusters.
	GCPKubeClusterNameOverrideLabel = cloudKubeClusterNameOverrideLabel

	// KubernetesClusterLabel indicates name of the kubernetes cluster for auto-discovered services inside kubernetes.
	KubernetesClusterLabel = TeleportNamespace + "/kubernetes-cluster"

	// AWSRolesAnywhereProfileNameOverrideLabel indicates the name of the AWS IAM Roles Anywhere Profile's tag key
	// that Teleport will use to override the name of the discovered profile.
	// Ensure this name is unique and valid DNS label.
	AWSRolesAnywhereProfileNameOverrideLabel = "TeleportApplicationName"

	// AWSRolesAnywhereProfileARNLabel is the label key to store the Profile ARN when creating an Application
	// resource from an AWS IAM Roles Anywhere Profile.
	AWSRolesAnywhereProfileARNLabel = TeleportNamespace + "/aws-roles-anywhere-profile-arn"

	// DiscoveryTypeLabel specifies type of discovered service that should be created from Kubernetes service.
	// Also added by discovery service to indicate the type of discovered
	// resource, e.g. "rds" for RDS databases, "eks" for EKS kube clusters, etc.
	DiscoveryTypeLabel = TeleportNamespace + "/discovery-type"
	// DiscoveryPortLabel specifies preferred port for a discovered app created from Kubernetes service.
	DiscoveryPortLabel = TeleportNamespace + "/port"
	// DiscoveryProtocolLabel specifies protocol for a discovered app created from Kubernetes service.
	DiscoveryProtocolLabel = TeleportNamespace + "/protocol"
	// DiscoveryAppRewriteLabel specifies rewrite rules for a discovered app created from Kubernetes service.
	DiscoveryAppRewriteLabel = TeleportNamespace + "/app-rewrite"
	// DiscoveryAppNameLabel specifies explicitly name of an app created from Kubernetes service.
	DiscoveryAppNameLabel = TeleportNamespace + "/name"
	// DiscoveryPathLabel optionally specifies a context path for apps created from Kubernetes services.
	DiscoveryPathLabel = TeleportNamespace + "/path"
	// DiscoveryAppInsecureSkipVerify specifies the TLS verification enforcement for a discovered app created from Kubernetes service.
	DiscoveryAppInsecureSkipVerify = TeleportNamespace + "/insecure-skip-verify"
	// DiscoveryAppIgnore specifies if a Kubernetes service should be ignored by discovery service.
	DiscoveryAppIgnore = TeleportNamespace + "/ignore"
	// DiscoveryPublicAddr specifies the public address for a discovered app created from a Kubernetes service.
	DiscoveryPublicAddr = TeleportNamespace + "/public-addr"

	// ReqAnnotationApproveSchedulesLabel is the request annotation key at which schedules are stored for access plugins.
	ReqAnnotationApproveSchedulesLabel = "/schedules"
	// ReqAnnotationNotifySchedulesLabel is the request annotation key at which notify schedules are stored for access plugins.
	ReqAnnotationNotifySchedulesLabel = "/notify-services"
	// ReqAnnotationTeamsLabel is the request annotation key at which teams are stored for access plugins.
	ReqAnnotationTeamsLabel = "/teams"

	// CloudAWS identifies that a resource was discovered in AWS.
	CloudAWS = "AWS"
	// CloudAzure identifies that a resource was discovered in Azure.
	CloudAzure = "Azure"
	// CloudGCP identifies that a resource was discovered in GCP.
	CloudGCP = "GCP"

	// SchemaMCPStdio is a URI schema for MCP servers using stdio transport.
	SchemaMCPStdio = "mcp+stdio://"
	// MCPTransportStdio indicates the MCP server uses stdio transport.
	MCPTransportStdio = "stdio"

	// DiscoveredResourceNode identifies a discovered SSH node.
	DiscoveredResourceNode = "node"
	// DiscoveredResourceDatabase identifies a discovered database.
	DiscoveredResourceDatabase = "db"
	// DiscoveredResourceKubernetes identifies a discovered kubernetes cluster.
	DiscoveredResourceKubernetes = "k8s"
	// DiscoveredResourceAgentlessNode identifies a discovered agentless SSH node.
	DiscoveredResourceAgentlessNode = "node.openssh"
	// DiscoveredResourceEICENode identifies a discovered AWS EC2 Instance using the EICE access method.
	DiscoveredResourceEICENode = "node.openssh-eice"
	// DiscoveredResourceApp identifies a discovered Kubernetes App.
	DiscoveredResourceApp = "app"

	// TeleportAzureMSIEndpoint is a special URL intercepted by TSH local proxy, serving Azure credentials.
	TeleportAzureMSIEndpoint = "azure-msi." + TeleportNamespace
	// TeleportAzureIdentityEndpoint is a special URL intercepted by TSH local proxy, serving Azure credentials.
	TeleportAzureIdentityEndpoint = "azure-identity." + TeleportNamespace

	// ConnectMyComputerNodeOwnerLabel is a label used to control access to the node managed by
	// Teleport Connect as part of Connect My Computer. See [teleterm.connectmycomputer.RoleSetup].
	ConnectMyComputerNodeOwnerLabel = TeleportNamespace + "/connect-my-computer/owner"
)

var (
	// AWSKubeClusterNameOverrideLabels are the label keys that Teleport
	// supports to override the kubernetes cluster name of discovered AWS kube
	// clusters.
	// Originally Teleport supported just the namespaced label
	// "teleport.dev/kubernetes-name", but this was an invalid label key in
	// other clouds.
	// For consistency and backwards compatibility, Teleport now supports both
	// the generic cloud kube cluster name override label and the original
	// namespaced label.
	AWSKubeClusterNameOverrideLabels = []string{
		cloudKubeClusterNameOverrideLabel,
		// This is a legacy label that should continue to be supported, but
		// don't reference it in documentation or error messages anymore.
		// The generic label takes precedence.
		TeleportNamespace + "/kubernetes-name",
	}
	// AWSDatabaseNameOverrideLabels are the label keys that Teleport
	// supports to override the database name of discovered AWS databases.
	// Originally Teleport supported just the namespaced label
	// "teleport.dev/database_name", but this was an invalid label key in
	// other clouds.
	// For consistency and backwards compatibility, Teleport now supports both
	// the generic cloud database name override label and the original
	// namespaced label.
	AWSDatabaseNameOverrideLabels = []string{
		cloudDatabaseNameOverrideLabel,
		// This is a legacy label that should continue to be supported, but
		// don't reference it in documentation or error messages anymore.
		// The generic label takes precedence.
		TeleportNamespace + "/database_name",
	}
)

// Labels added by the discovery service to discovered databases,
// Kubernetes clusters, and Windows desktops.
const (
	// DiscoveryLabelRegion identifies a discovered cloud resource's region.
	DiscoveryLabelRegion = "region"
	// DiscoveryLabelAccountID is the label key containing AWS account ID.
	DiscoveryLabelAccountID = "account-id"
	// DiscoveryLabelEngine is the label key containing database engine name.
	DiscoveryLabelEngine = "engine"
	// DiscoveryLabelEngineVersion is the label key containing database engine version.
	DiscoveryLabelEngineVersion = "engine-version"
	// DiscoveryLabelEndpointType is the label key containing the endpoint type.
	DiscoveryLabelEndpointType = "endpoint-type"
	// DiscoveryLabelVPCID is the label key containing the VPC ID.
	DiscoveryLabelVPCID = "vpc-id"
	// DiscoveryLabelNamespace is the label key for namespace name.
	DiscoveryLabelNamespace = "namespace"
	// DiscoveryLabelWorkgroup is the label key for workgroup name.
	DiscoveryLabelWorkgroup = "workgroup"
	// DiscoveryLabelStatus is the label key containing the database status, e.g. "available"
	DiscoveryLabelStatus = "status"
	// DiscoveryLabelAWSArn is an internal label that contains AWS Arn of the resource.
	DiscoveryLabelAWSArn = TeleportInternalLabelPrefix + "aws-arn"

	// DiscoveryLabelAzureSubscriptionID is the label key for Azure subscription ID.
	DiscoveryLabelAzureSubscriptionID = "subscription-id"
	// DiscoveryLabelAzureResourceGroup is the label key for the Azure resource group name.
	DiscoveryLabelAzureResourceGroup = "resource-group"
	// DiscoveryLabelAzureReplicationRole is the replication role of an Azure DB Flexible server, e.g. "Source" or "Replica".
	DiscoveryLabelAzureReplicationRole = "replication-role"
	// DiscoveryLabelAzureSourceServer is the source server for replica Azure DB Flexible servers.
	// This is the source (primary) database resource name.
	DiscoveryLabelAzureSourceServer = "source-server"

	// DiscoveryLabelGCPProjectID is the label key for GCP project ID.
	DiscoveryLabelGCPProjectID = "project-id"
	// DiscoveryLabelGCPLocation is the label key for GCP location.
	DiscoveryLabelGCPLocation = "location"

	// DiscoveryLabelWindowsDNSHostName is the DNS hostname of an LDAP object.
	DiscoveryLabelWindowsDNSHostName = TeleportNamespace + "/dns_host_name"
	// DiscoveryLabelWindowsComputerName is the name of an LDAP object.
	DiscoveryLabelWindowsComputerName = TeleportNamespace + "/computer_name"
	// DiscoveryLabelWindowsOS is the operating system of an LDAP object.
	DiscoveryLabelWindowsOS = TeleportNamespace + "/os"
	// DiscoveryLabelWindowsOSVersion operating system version of an LDAP object.
	DiscoveryLabelWindowsOSVersion = TeleportNamespace + "/os_version"
	// DiscoveryLabelWindowsOU is an LDAP objects's OU.
	DiscoveryLabelWindowsOU = TeleportNamespace + "/ou"
	// DiscoveryLabelWindowsIsDomainController is whether an LDAP object is a
	// domain controller.
	DiscoveryLabelWindowsIsDomainController = TeleportNamespace + "/is_domain_controller"
	// DiscoveryLabelWindowsDomain is an Active Directory domain name.
	DiscoveryLabelWindowsDomain = TeleportNamespace + "/windows_domain"
	// DiscoveryLabelLDAPPrefix is the prefix used when applying any custom
	// labels per the discovery LDAP attribute labels configuration.
	DiscoveryLabelLDAPPrefix = "ldap/"
)

// BackSortedLabelPrefixes are label names that we want to always be at the end of
// the sorted labels list to reduce visual clutter. This will generally be automatically
// discovered cloud provider labels such as azure/aks-managed-createOperationID=123123123123
// or internal labels
var BackSortedLabelPrefixes = []string{CloudAWS, CloudAzure, CloudGCP, DiscoveryLabelLDAPPrefix, TeleportNamespace}

const (
	// TeleportInternalLabelPrefix is the prefix used by all Teleport internal labels. Those labels
	// are automatically populated by Teleport and are expected to be used by Teleport internal
	// components and not for RBAC.
	//
	// See also TeleportNamespace and TeleportHiddenLabelPrefix.
	TeleportInternalLabelPrefix = "teleport.internal/"

	// TeleportHiddenLabelPrefix is the prefix used by all user specified hidden labels.
	//
	// See also TeleportNamespace and TeleportInternalLabelPrefix.
	TeleportHiddenLabelPrefix = "teleport.hidden/"

	// TeleportDynamicLabelPrefix is the prefix used by labels that can change
	// over time and should not be used as part of a role's deny rules.
	TeleportDynamicLabelPrefix = "dynamic/"

	// DiscoveredNameLabel is a resource metadata label name used to identify
	// the discovered name of a resource, i.e. the name of a resource before a
	// uniquely distinguishing suffix is added by the discovery service.
	// See: RFD 129 - Avoid Discovery Resource Name Collisions.
	DiscoveredNameLabel = TeleportInternalLabelPrefix + "discovered-name"

	// BotLabel is a label used to identify a resource used by a certificate renewal bot.
	BotLabel = TeleportInternalLabelPrefix + "bot"

	// BotGenerationLabel is a label used to record the certificate generation counter.
	BotGenerationLabel = TeleportInternalLabelPrefix + "bot-generation"

	// InternalResourceIDLabel is a label used to store an ID to correlate between two resources
	// A pratical example of this is to create a correlation between a Node Provision Token and
	// the Node that used that token to join the cluster
	InternalResourceIDLabel = TeleportInternalLabelPrefix + "resource-id"

	// AlertOnLogin is an internal label that indicates an alert should be displayed to users on login
	AlertOnLogin = TeleportInternalLabelPrefix + "alert-on-login"

	// AlertPermitAll is an internal label that indicates that an alert is suitable for display
	// to all users.
	AlertPermitAll = TeleportInternalLabelPrefix + "alert-permit-all"

	// AlertLink is an internal label that indicates that an alert is a link.
	AlertLink = TeleportInternalLabelPrefix + "link"

	// AlertLinkText is a text that will be rendered by Web UI on the action
	// button accompanying the alert.
	AlertLinkText = TeleportInternalLabelPrefix + "link-text"

	// AlertVerbPermit is an internal label that permits a user to view the alert if they
	// hold a specific resource permission verb (e.g. 'node:list'). Note that this label is
	// a coarser control than it might initially appear and has the potential for accidental
	// misuse. Because this permitting strategy doesn't take into account constraints such as
	// label selectors or where clauses, it can't reliably protect information related to a
	// specific resource. This label should be used only for permitting of alerts that are
	// of concern to holders of a given <resource>:<verb> capability in the most general case.
	AlertVerbPermit = TeleportInternalLabelPrefix + "alert-verb-permit"

	// AlertSupersedes is an internal label used to indicate when one alert supersedes
	// another. Teleport may choose to hide the superseded alert if the superseding alert
	// is also visible to the user and of higher or equivalent severity. This intended as
	// a mechanism for reducing noise/redundancy, and is not a form of access control. Use
	// one of the "permit" labels if you need to restrict viewership of an alert.
	AlertSupersedes = TeleportInternalLabelPrefix + "alert-supersedes"

	// AlertLicenseExpired is an internal label that indicates that the license has expired.
	AlertLicenseExpired = TeleportInternalLabelPrefix + "license-expired-warning"

	// TeleportInternalDiscoveryGroupName is the label used to store the name of the discovery group
	// that the discovered resource is owned by. It is used to differentiate resources
	// that belong to different discovery services that operate on different sets of resources.
	TeleportInternalDiscoveryGroupName = TeleportInternalLabelPrefix + "discovery-group-name"

	// TeleportInternalDiscoveryIntegrationName is the label used to store the name of the integration
	// whose credentials were used to discover the resource.
	// It is used to report stats for a given Integration / DiscoveryConfig.
	TeleportInternalDiscoveryIntegrationName = TeleportInternalLabelPrefix + "discovery-integration-name"

	// TeleportInternalDiscoveryConfigName is the label used to store the name of the discovery config
	// whose matchers originated the resource.
	// It is used to report stats for a given Integration / DiscoveryConfig.
	TeleportInternalDiscoveryConfigName = TeleportInternalLabelPrefix + "discovery-config-name"

	// TeleportDowngradedLabel identifies resources that have been automatically
	// downgraded before being returned to clients on older versions that do not
	// support one or more features enabled in that resource.
	TeleportDowngradedLabel = TeleportInternalLabelPrefix + "downgraded"

	// TeleportInternalResourceType indicates the type of internal Teleport resource a resource is.
	// Valid values are:
	// - system: These resources will be automatically created and overwritten on startup. Users should
	//           not change these resources.
	// - preset: These resources will be created if they don't exist. Updates may be applied to them,
	//           but user changes to these resources will be preserved.
	TeleportInternalResourceType = TeleportInternalLabelPrefix + "resource-type"

	// TeleportResourceRevision marks a teleport-managed resource with a reversion
	// number to aid future migrations. Label value is expected to be a number.
	TeleportResourceRevision = TeleportInternalLabelPrefix + "revision"

	// SystemResource are resources that will be automatically created and overwritten on startup. Users
	// should not change these resources.
	SystemResource = "system"

	// PresetResource are resources that will be created if they don't exist. Updates may be applied
	// to them, but user changes to these resources will be preserved.
	PresetResource = "preset"

	// DemoResource are resources that demonstrates specific Teleport features.
	// These resources are typically managed internally by Teleport and enabled
	// via flags. Users should not change these resources.
	DemoResource = "demo"

	// ProxyGroupIDLabel is the internal-use label for proxy heartbeats that's
	// used by reverse tunnel agents to keep track of multiple independent sets
	// of proxies in proxy peering mode.
	ProxyGroupIDLabel = TeleportInternalLabelPrefix + "proxygroup-id"

	// ProxyGroupGenerationLabel is the internal-use label for proxy heartbeats
	// that's used by reverse tunnel agents to know which proxies in each proxy
	// group they should attempt to be connected to.
	ProxyGroupGenerationLabel = TeleportInternalLabelPrefix + "proxygroup-gen"

	// UnstableProxyPeerQUICLabel is the internal-use label for proxy heartbeats
	// that's used to signal that the proxy supports receiving proxy peering
	// connections over QUIC. The value should be "yes".
	UnstableProxyPeerQUICLabel = TeleportInternalLabelPrefix + "proxy-peer-quic"

	// OktaAppNameLabel is the individual app name label.
	OktaAppNameLabel = TeleportInternalLabelPrefix + "okta-app-name"

	// OktaAppDescriptionLabel is the individual app description label.
	OktaAppDescriptionLabel = TeleportInternalLabelPrefix + "okta-app-description"

	// OktaGroupNameLabel is the individual group name label.
	OktaGroupNameLabel = TeleportInternalLabelPrefix + "okta-group-name"

	// OktaGroupDescriptionLabel is the individual group description label.
	OktaGroupDescriptionLabel = TeleportInternalLabelPrefix + "okta-group-description"

	// OktaRoleNameLabel is the human readable name for a role sourced from Okta.
	OktaRoleNameLabel = TeleportInternalLabelPrefix + "okta-role-name"

	// PluginGenerationLabel is the label for the current generation of the plugin.
	PluginGenerationLabel = TeleportInternalLabelPrefix + "plugin-generation"

	// EntraTenantIDLabel is the label for the Entra tenant ID.
	EntraTenantIDLabel = TeleportInternalLabelPrefix + "entra-tenant"

	// EntraUniqueIDLabel is the label for the unique identifier of the object in the Entra ID directory.
	EntraUniqueIDLabel = TeleportInternalLabelPrefix + "entra-unique-id"

	// EntraUPNLabel is the label for the user principal name in Entra ID.
	EntraUPNLabel = TeleportInternalLabelPrefix + "entra-upn"

	// EntraDisplayNameLabel is the label for the display name of the object in the Entra ID directory.
	// The display name may not be unique.
	EntraDisplayNameLabel = TeleportInternalLabelPrefix + "entra-display-name"

	// EntraSAMAccountNameLabel is the label for user's on-premises sAMAccountName.
	EntraSAMAccountNameLabel = TeleportInternalLabelPrefix + "entra-sam-account-name"

	// GitHubOrgLabel is the label for GitHub organization.
	GitHubOrgLabel = TeleportInternalLabelPrefix + "github-org"
)

const (
	// NotificationTitleLabel is the label which contains the title of the notification.
	NotificationTitleLabel = TeleportInternalLabelPrefix + "title"
	// NotificationClickedLabel is the label which contains whether the notification has been clicked on by the user.
	NotificationClickedLabel = TeleportInternalLabelPrefix + "clicked"
	// NotificationScope is the label which contains the scope of the notification, either "user" or "global"
	NotificationScope = TeleportInternalLabelPrefix + "scope"
	// NotificationTextContentLabel is the label which contains the text content of a user-created notification.
	NotificationTextContentLabel = TeleportInternalLabelPrefix + "content"

	// NotificationDefaultInformationalSubKind is the default subkind for an informational notification.
	NotificationDefaultInformationalSubKind = "default-informational"
	// NotificationDefaultWarningSubKind is the default subkind for a warning notification.
	NotificationDefaultWarningSubKind = "default-warning"

	// NotificationUserCreatedInformationalSubKind is the subkind for a user-created informational notification.
	NotificationUserCreatedInformationalSubKind = "user-created-informational"
	// NotificationUserCreatedWarningSubKind is the subkind for a user-created warning notification.
	NotificationUserCreatedWarningSubKind = "user-created-warning"

	// NotificationAccessRequestPendingSubKind is the subkind for a notification for an access request pending review.
	NotificationAccessRequestPendingSubKind = "access-request-pending"
	// NotificationAccessRequestApprovedSubKind is the subkind for a notification for a user's access request being approved.
	NotificationAccessRequestApprovedSubKind = "access-request-approved"
	// NotificationAccessRequestDeniedSubKind is the subkind for a notification for a user's access request being denied.
	NotificationAccessRequestDeniedSubKind = "access-request-denied"
	// NotificationAccessRequestPromotedSubKind is the subkind for a notification for a user's access request being promoted to an access list.
	NotificationAccessRequestPromotedSubKind = "access-request-promoted"

	// NotificationAccessListReviewDue14dSubKind is the subkind for a notification for an access list review due in less than 14 days.
	NotificationAccessListReviewDue14dSubKind = "access-list-review-due-14d"

	// NotificationAccessListReviewDue7dSubKind is the subkind for a notification for an access list review due in less than 7 days.
	NotificationAccessListReviewDue7dSubKind = "access-list-review-due-7d"

	// NotificationAccessListReviewDue3dSubKind is the subkind for a notification for an access list review due in less than 3 days.
	NotificationAccessListReviewDue3dSubKind = "access-list-review-due-3d"

	// NotificationAccessListReviewDue0dSubKind is the subkind for a notification for an access list review due today.
	NotificationAccessListReviewDue0dSubKind = "access-list-review-due-0d"

	// NotificationAccessListReviewOverdue3dSubKind is the subkind for a notification for an access list review overdue by 3 days.
	NotificationAccessListReviewOverdue3dSubKind = "access-list-review-overdue-3d"

	// NotificationAccessListReviewOverdue7dSubKind is the subkind for a notification for an access list review overdue by 7 days.
	NotificationAccessListReviewOverdue7dSubKind = "access-list-review-overdue-7d"
)

const (
	// NotificationIdentifierPrefixAccessListDueReminder14d is the prefix for unique notification identifiers for 14d access list review reminders.
	NotificationIdentifierPrefixAccessListDueReminder14d = "access_list_14d_due_reminder"
	// NotificationIdentifierPrefixAccessListDueReminder7d is the prefix for unique notification identifiers for 7d access list review reminders.
	NotificationIdentifierPrefixAccessListDueReminder7d = "access_list_7d_due_reminder"
	// NotificationIdentifierPrefixAccessListDueReminder3d is the prefix for unique notification identifiers for 3d access list review reminders.
	NotificationIdentifierPrefixAccessListDueReminder3d = "access_list_3d_due_reminder"
	// NotificationIdentifierPrefixAccessListDueReminder0d is the prefix for unique notification identifiers for 0d (today) access list review reminders.
	NotificationIdentifierPrefixAccessListDueReminder0d = "access_list_0d_due_reminder"
	// NotificationIdentifierPrefixAccessListDueReminder30d is the prefix for unique notification identifiers for 3d overdue access list review reminders.
	NotificationIdentifierPrefixAccessListOverdue3d = "access_list_3d_overdue_reminder"
	// NotificationIdentifierPrefixAccessListDueReminder30d is the prefix for unique notification identifiers for 7d overdue access list review reminders.
	NotificationIdentifierPrefixAccessListOverdue7d = "access_list_7d_overdue_reminder"
)

const (
	// InstallMethodAWSOIDCDeployServiceEnvVar is the env var used to detect if the agent was installed
	// using the DeployService action of the AWS OIDC integration.
	InstallMethodAWSOIDCDeployServiceEnvVar = "TELEPORT_INSTALL_METHOD_AWSOIDC_DEPLOYSERVICE"

	// AWSOIDCAgentLabel is a label that indicates that the service was deployed into ECS/Fargate using the AWS OIDC Integration.
	AWSOIDCAgentLabel = TeleportNamespace + "/awsoidc-agent"
)

// CloudHostnameTag is the name of the tag in a cloud instance used to override a node's hostname.
const CloudHostnameTag = "TeleportHostname"

// InstanceMetadataType is the type of cloud instance metadata client.
type InstanceMetadataType string

const (
	InstanceMetadataTypeDisabled InstanceMetadataType = "disabled"
	InstanceMetadataTypeEC2      InstanceMetadataType = "EC2"
	InstanceMetadataTypeAzure    InstanceMetadataType = "Azure"
	InstanceMetadataTypeGCP      InstanceMetadataType = "GCP"
	InstanceMetadataTypeOracle   InstanceMetadataType = "Oracle"
)

// OriginValues lists all possible origin values.
var OriginValues = common.OriginValues

const (
	// RecordAtNode is the default. Sessions are recorded at Teleport nodes.
	RecordAtNode = "node"

	// RecordAtProxy enables the recording proxy which intercepts and records
	// all sessions.
	RecordAtProxy = "proxy"

	// RecordOff is used to disable session recording completely.
	RecordOff = "off"

	// RecordAtNodeSync enables the nodes to stream sessions in sync mode
	// to the auth server
	RecordAtNodeSync = "node-sync"

	// RecordAtProxySync enables the recording proxy which intercepts and records
	// all sessions, streams the records synchronously
	RecordAtProxySync = "proxy-sync"
)

// SessionRecordingModes lists all possible session recording modes.
var SessionRecordingModes = []string{RecordAtNode, RecordAtProxy, RecordOff, RecordAtNodeSync, RecordAtProxySync}

// TunnelType is the type of tunnel.
type TunnelType string

const (
	// NodeTunnel is a tunnel where the node connects to the proxy (dial back).
	NodeTunnel TunnelType = "node"

	// ProxyTunnel is a tunnel where a proxy connects to the proxy (trusted cluster).
	ProxyTunnel TunnelType = "proxy"

	// AppTunnel is a tunnel where the application proxy dials back to the proxy.
	AppTunnel TunnelType = "app"

	// KubeTunnel is a tunnel where the kubernetes service dials back to the proxy.
	KubeTunnel TunnelType = "kube"

	// DatabaseTunnel is a tunnel where a database proxy dials back to the proxy.
	DatabaseTunnel TunnelType = "db"

	// WindowsDesktopTunnel is a tunnel where the Windows desktop service dials back to the proxy.
	WindowsDesktopTunnel TunnelType = "windows_desktop"

	// OktaTunnel is a tunnel where the Okta service dials back to the proxy.
	OktaTunnel TunnelType = "okta"
)

type TunnelStrategyType string

const (
	// AgentMesh requires agents to create a reverse tunnel to
	// every proxy server.
	AgentMesh TunnelStrategyType = "agent_mesh"
	// ProxyPeering requires agents to create a reverse tunnel to a configured
	// number of proxy servers and enables proxy to proxy communication.
	ProxyPeering TunnelStrategyType = "proxy_peering"
)

const (
	// ResourceMetadataName refers to a resource metadata field named "name".
	ResourceMetadataName = "name"

	// ResourceSpecDescription refers to a resource spec field named "description".
	ResourceSpecDescription = "description"

	// ResourceSpecHostname refers to a resource spec field named "hostname".
	ResourceSpecHostname = "hostname"

	// ResourceSpecAddr refers to a resource spec field named "address".
	ResourceSpecAddr = "address"

	// ResourceSpecPublicAddr refers to a resource field named "address".
	ResourceSpecPublicAddr = "publicAddress"

	// ResourceSpecType refers to a resource field named "type".
	ResourceSpecType = "type"

	// ResourceKind refers to a resource field named "kind".
	ResourceKind = "kind"
)

// RequestableResourceKinds lists all Teleport resource kinds users can request access to.
var RequestableResourceKinds = []string{
	KindNode,
	KindKubernetesCluster,
	KindKubernetesResource,
	KindDatabase,
	KindApp,
	KindWindowsDesktop,
	KindUserGroup,
	KindSAMLIdPServiceProvider,
	KindIdentityCenterAccount,
	KindIdentityCenterAccountAssignment,
	KindGitServer,
}

// LegacyRequestableKubeResourceKinds lists all legacy Teleport resource kinds users can request access to.
// Those are the requestable Kubernetes resource kinds that were supported before the introduction of
// custom resource support. We need to keep them to maintain support with older Teleport versions.
// TODO(@creack): DELETE IN v20.0.0.
var LegacyRequestableKubeResourceKinds = []string{
	KindKubePod,
	KindKubeSecret,
	KindKubeConfigmap,
	KindKubeNamespace,
	KindKubeService,
	KindKubeServiceAccount,
	KindKubeNode,
	KindKubePersistentVolume,
	KindKubePersistentVolumeClaim,
	KindKubeDeployment,
	KindKubeReplicaSet,
	KindKubeStatefulset,
	KindKubeDaemonSet,
	KindKubeClusterRole,
	KindKubeRole,
	KindKubeClusterRoleBinding,
	KindKubeRoleBinding,
	KindKubeCronjob,
	KindKubeJob,
	KindKubeCertificateSigningRequest,
	KindKubeIngress,
}

// Prefix constants to identify kubernetes resources in access requests.
const (
	// AccessRequestPrefixKindKube denotes that the resource is a kubernetes one. Used for access requests.
	AccessRequestPrefixKindKube = "kube:"
	// AccessRequestPrefixKindKubeClusterWide denotes that the kube resource is cluster-wide.
	AccessRequestPrefixKindKubeClusterWide = AccessRequestPrefixKindKube + "cw:"
	// AccessRequestPrefixKindKubeNamespaced denotes that the kube resource is namespaced.
	AccessRequestPrefixKindKubeNamespaced = AccessRequestPrefixKindKube + "ns:"
)

// The list below needs to be kept in sync with `kubernetesResourceKindOptions`
// in `web/packages/teleport/src/Roles/RoleEditor/standardmodel.ts`. (Keeping
// this comment separate to prevent it from being included in the official
// package docs.)

// KubernetesResourcesKinds lists the supported Kubernetes resource kinds. This
// is for the latest version of Role resources; roles whose version is set to
// v6 or prior only support [KindKubePod].
var KubernetesResourcesKinds = []string{
	KindKubePod,
	KindKubeSecret,
	KindKubeConfigmap,
	KindKubeNamespace,
	KindKubeService,
	KindKubeServiceAccount,
	KindKubeNode,
	KindKubePersistentVolume,
	KindKubePersistentVolumeClaim,
	KindKubeDeployment,
	KindKubeReplicaSet,
	KindKubeStatefulset,
	KindKubeDaemonSet,
	KindKubeClusterRole,
	KindKubeRole,
	KindKubeClusterRoleBinding,
	KindKubeRoleBinding,
	KindKubeCronjob,
	KindKubeJob,
	KindKubeCertificateSigningRequest,
	KindKubeIngress,
}

// KubernetesResourceSelfSubjectAccessReview is a Kubernetes resource that
// represents a self-subject access review. This gets injected in the allow section in the roles.
var KubernetesResourceSelfSubjectAccessReview = KubernetesResource{
	Kind:     "selfsubjectaccessreviews",
	Name:     Wildcard,
	Verbs:    []string{"create"},
	APIGroup: "authorization.k8s.io",
}

// KubernetesResourcesV7KindGroups maps the legacy Teleport kube kinds
// to their kubernetes group.
// Used for validation in role >=v8 to check whether an older value has
// been accidentally used.
var KubernetesResourcesV7KindGroups = map[string]string{
	KindKubePod:                       "",
	KindKubeSecret:                    "",
	KindKubeConfigmap:                 "",
	KindKubeNamespace:                 "",
	KindKubeService:                   "",
	KindKubeServiceAccount:            "",
	KindKubeNode:                      "",
	KindKubePersistentVolume:          "",
	KindKubePersistentVolumeClaim:     "",
	KindKubeDeployment:                "apps",
	KindKubeReplicaSet:                "apps",
	KindKubeStatefulset:               "apps",
	KindKubeDaemonSet:                 "apps",
	KindKubeClusterRole:               "rbac.authorization.k8s.io",
	KindKubeRole:                      "rbac.authorization.k8s.io",
	KindKubeClusterRoleBinding:        "rbac.authorization.k8s.io",
	KindKubeRoleBinding:               "rbac.authorization.k8s.io",
	KindKubeCronjob:                   "batch",
	KindKubeJob:                       "batch",
	KindKubeCertificateSigningRequest: "certificates.k8s.io",
	KindKubeIngress:                   "networking.k8s.io",
}

// KubernetesResourcesKindsPlurals maps the legacy Teleport kube kinds
// to their kubernetes name.
// Used to upgrade roles <=v7 as well as to support existing access request
// format.
// NOTE: Namespace having a different behavior between versions, it is omitted from this map.
var KubernetesResourcesKindsPlurals = map[string]string{
	KindKubePod:                       "pods",
	KindKubeSecret:                    "secrets",
	KindKubeConfigmap:                 "configmaps",
	KindKubeService:                   "services",
	KindKubeServiceAccount:            "serviceaccounts",
	KindKubeNode:                      "nodes",
	KindKubePersistentVolume:          "persistentvolumes",
	KindKubePersistentVolumeClaim:     "persistentvolumeclaims",
	KindKubeDeployment:                "deployments",
	KindKubeReplicaSet:                "replicasets",
	KindKubeReplicationController:     "replicationcontrollers",
	KindKubeStatefulset:               "statefulsets",
	KindKubeDaemonSet:                 "daemonsets",
	KindKubeClusterRole:               "clusterroles",
	KindKubeRole:                      "roles",
	KindKubeClusterRoleBinding:        "clusterrolebindings",
	KindKubeRoleBinding:               "rolebindings",
	KindKubeCronjob:                   "cronjobs",
	KindKubeJob:                       "jobs",
	KindKubeCertificateSigningRequest: "certificatesigningrequests",
	KindKubeIngress:                   "ingresses",
}

const (
	// KubeVerbGet is the Kubernetes verb for "get".
	KubeVerbGet = "get"
	// KubeVerbCreate is the Kubernetes verb for "create".
	KubeVerbCreate = "create"
	// KubeVerbUpdate is the Kubernetes verb for "update".
	KubeVerbUpdate = "update"
	// KubeVerbPatch is the Kubernetes verb for "patch".
	KubeVerbPatch = "patch"
	// KubeVerbDelete is the Kubernetes verb for "delete".
	KubeVerbDelete = "delete"
	// KubeVerbList is the Kubernetes verb for "list".
	KubeVerbList = "list"
	// KubeVerbWatch is the Kubernetes verb for "watch".
	KubeVerbWatch = "watch"
	// KubeVerbDeleteCollection is the Kubernetes verb for "deletecollection".
	KubeVerbDeleteCollection = "deletecollection"
	// KubeVerbExec is the Kubernetes verb for "pod/exec".
	KubeVerbExec = "exec"
	// KubeVerbPortForward is the Kubernetes verb for "pod/portforward".
	KubeVerbPortForward = "portforward"
)

// The list below needs to be kept in sync with `kubernetesResourceVerbOptions`
// in `web/packages/teleport/src/Roles/RoleEditor/standardmodel.ts`. (Keeping
// this comment separate to prevent it from being included in the official
// package docs.)

// KubernetesVerbs lists the supported Kubernetes verbs.
var KubernetesVerbs = []string{
	Wildcard,
	KubeVerbGet,
	KubeVerbCreate,
	KubeVerbUpdate,
	KubeVerbPatch,
	KubeVerbDelete,
	KubeVerbList,
	KubeVerbWatch,
	KubeVerbDeleteCollection,
	KubeVerbExec,
	KubeVerbPortForward,
}

// KubernetesClusterWideResourceKinds is the list of supported Kubernetes cluster resource kinds
// that are not namespaced.
// Needed to maintain backward compatibility.
var KubernetesClusterWideResourceKinds = []string{
	KindKubeNamespace,
	KindKubeNode,
	KindKubePersistentVolume,
	KindKubeClusterRole,
	KindKubeClusterRoleBinding,
	KindKubeCertificateSigningRequest,
}

// KubernetesNamespacedResourceKinds is the list of known Kubernetes resource kinds
// that are namespaced.
//
// Generated from `kubectl api-resources --namespaced=true -o name --sort-by=name` (kind k8s v1.32.2).
// The format is "<plural>.<apigroup>".
//
// Only used in role >=v8 to attempt to validate the api_group field.
// If we have a match, we know we need a namespaced value, if we don't
// have a match, we don't know we don't. Best effort basis.
//
// Key: resource kind, value: api group.
var kubernetesNamespacedResourceKinds = map[string]string{
	"bindings":                  "",
	"configmaps":                "",
	"controllerrevisions":       "apps",
	"cronjobs":                  "batch",
	"csistoragecapacities":      "storage.k8s.io",
	"daemonsets":                "apps",
	"deployments":               "apps",
	"endpoints":                 "",
	"endpointslices":            "discovery.k8s.io",
	"events":                    "events.k8s.io",
	"horizontalpodautoscalers":  "autoscaling",
	"ingresses":                 "networking.k8s.io",
	"jobs":                      "batch",
	"leases":                    "coordination.k8s.io",
	"limitranges":               "",
	"localsubjectaccessreviews": "authorization.k8s.io",
	"networkpolicies":           "networking.k8s.io",
	"persistentvolumeclaims":    "",
	"poddisruptionbudgets":      "policy",
	"pods":                      "",
	"podtemplates":              "",
	"replicasets":               "apps",
	"replicationcontrollers":    "",
	"resourcequotas":            "",
	"rolebindings":              "rbac.authorization.k8s.io",
	"roles":                     "rbac.authorization.k8s.io",
	"secrets":                   "",
	"serviceaccounts":           "",
	"services":                  "",
	"statefulsets":              "apps",
}

// List of "" (core / legacy) resources.
//
// Used to validate the api_group field.
//
// Generated with:
//
//	(kubectl api-resources --api-group "" --output=name --namespaced=true && kubectl api-resources --api-group "" --output=name --namespaced=false) | sort
var KubernetesCoreResourceKinds = map[string]struct{}{
	"bindings":               {},
	"componentstatuses":      {},
	"configmaps":             {},
	"endpoints":              {},
	"events":                 {},
	"limitranges":            {},
	"namespaces":             {},
	"nodes":                  {},
	"persistentvolumeclaims": {},
	"persistentvolumes":      {},
	"pods":                   {},
	"podtemplates":           {},
	"replicationcontrollers": {},
	"resourcequotas":         {},
	"secrets":                {},
	"serviceaccounts":        {},
	"services":               {},
}

const (
	// TeleportDropGroup is a default group that users of the teleport automated user
	// provisioning system get added to when provisioned in INSECURE_DROP mode. This
	// prevents already existing users from being tampered with or deleted.
	TeleportDropGroup = "teleport-system"
	// TeleportKeepGroup is a default group that users of the teleport automated user
	// provisioning system get added to when provisioned in KEEP mode. This prevents
	// already existing users from being tampered with or deleted.
	TeleportKeepGroup = "teleport-keep"
	// TeleportStaticGroup is a default group that static host users get added to. This
	// prevents already existing users from being tampered with or deleted.
	TeleportStaticGroup = "teleport-static"
)

const (
	// JWTClaimsRewriteRolesAndTraits includes both roles and traits in the JWT token.
	JWTClaimsRewriteRolesAndTraits = "roles-and-traits"
	// JWTClaimsRewriteRoles includes only the roles in the JWT token.
	JWTClaimsRewriteRoles = "roles"
	// JWTClaimsRewriteTraits includes only the traits in the JWT token.
	JWTClaimsRewriteTraits = "traits"
	// JWTClaimsRewriteNone include neither traits nor roles in the JWT token.
	JWTClaimsRewriteNone = "none"
)

const (
	// DefaultInstallerScriptName is the name of the by default populated, EC2
	// installer script
	DefaultInstallerScriptName = "default-installer"

	// DefaultInstallerScriptNameAgentless is the name of the by default populated, EC2
	// installer script when agentless mode is enabled for a matcher
	DefaultInstallerScriptNameAgentless = "default-agentless-installer"
)

const (
	// ApplicationProtocolHTTP is the HTTP (Web) apps protocol
	ApplicationProtocolHTTP = "HTTP"
	// ApplicationProtocolTCP is the TCP apps protocol.
	ApplicationProtocolTCP = "TCP"
)

const (
	// HostedPluginLabel defines the name for the hosted plugin label.
	// When this label is set to "true" on a Plugin resource,
	// it indicates that the Plugin should be run by the Cloud service,
	// rather than self-hosted plugin services.
	HostedPluginLabel = TeleportNamespace + "/hosted-plugin"
)

const (
	// OktaOrgURLLabel is the label used by Okta-managed resources to indicate
	// the upstream Okta organization that they come from.
	OktaOrgURLLabel = "okta/org"

	// OktaAppIDLabel is the label for the Okta application ID on appserver objects.
	OktaAppIDLabel = TeleportInternalLabelPrefix + "okta-app-id"

	// OktaCredPurposeLabel is used by Okta-managed PluginStaticCredentials to
	// indicate their purpose
	OktaCredPurposeLabel = "okta/purpose"

	// OktaCredPurposeAuth indicates that the credential is intended for
	// authenticating with the Okta REST API
	OktaCredPurposeAuth = "okta-auth"

	// OktaCredPurposeSCIMToken indicates that theis to be used for authenticating
	// SCIM requests from the upstream organization. The content of the credential
	// is a bcrypt hash of actual token.
	OktaCredPurposeSCIMToken = "scim-bearer-token"

	// CredPurposeOKTAAPITokenWithSCIMOnlyIntegration is used when okta integration was enabled without
	// app groups sync. Due to backward compatibility when teleport was downgraded to version where the
	// AppGroupSyncDisabled flag is not supported we need to prevent plugin from starting.
	// This is done by distinguishing between OktaCredPurposeAuth and CredPurposeOKTAAPITokenWithSCIMOnlyIntegration
	// that are only set when AppGroupSyncDisabled is set to true.
	CredPurposeOKTAAPITokenWithSCIMOnlyIntegration = "okta-auth-scim-only"
)

const (
	// SCIMBaseURLLabel defines a label indicating the base URL for
	// interacting with a plugin via SCIM. Useful for diagnostic display.
	SCIMBaseURLLabel = TeleportNamespace + "/scim-base-url"
)

const (
	// DatadogCredentialLabel is used by Datadog-managed PluginStaticCredentials
	// to indiciate credential type.
	DatadogCredentialLabel = "datadog/credential"

	// DatadogCredentialAPIKey indicates that the credential is used as a
	// Datadog API key.
	DatadogCredentialAPIKey = "datadog-api-key"

	// DatadogCredentialApplicationKey indicates that the credential is used as
	// a Datadog Application key.
	DatadogCredentialApplicationKey = "datadog-application-key"
)

const (
	// GitHubOrgServerDomain is the sub domain used in the hostname of a
	// types.Server to indicate the GitHub organization of a Git server.
	GitHubOrgServerDomain = "teleport-github-org"
)

// AccessMonitoringRuleState specifies the desired state of an access monitoring
// rule subject.
type AccessMonitoringRuleState string

const (
	// AccessMonitoringRuleStateReviewed indicates that the access monitoring
	// rule subject should be reviewed.
	AccessMonitoringRuleStateReviewed = "reviewed"
)

// BuiltInAutomaticReview is used within access monitoring rules and indicates
// that the automatic_review rule should be monitored by Teleport.
const BuiltInAutomaticReview = "builtin"
