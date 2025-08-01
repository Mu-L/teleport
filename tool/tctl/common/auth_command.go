/*
 * Teleport
 * Copyright (C) 2023  Gravitational, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package common

import (
	"context"
	"crypto/x509"
	"encoding/base32"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/gravitational/trace"
	"github.com/jonboulle/clockwork"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/gravitational/teleport"
	"github.com/gravitational/teleport/api/client/proto"
	"github.com/gravitational/teleport/api/client/webclient"
	apidefaults "github.com/gravitational/teleport/api/defaults"
	trustpb "github.com/gravitational/teleport/api/gen/proto/go/teleport/trust/v1"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/api/utils/keys"
	"github.com/gravitational/teleport/lib/auth/authclient"
	"github.com/gravitational/teleport/lib/client"
	"github.com/gravitational/teleport/lib/client/db"
	"github.com/gravitational/teleport/lib/client/identityfile"
	"github.com/gravitational/teleport/lib/cryptosuites"
	"github.com/gravitational/teleport/lib/defaults"
	"github.com/gravitational/teleport/lib/service/servicecfg"
	"github.com/gravitational/teleport/lib/services"
	"github.com/gravitational/teleport/lib/utils"
	"github.com/gravitational/teleport/lib/winpki"
	commonclient "github.com/gravitational/teleport/tool/tctl/common/client"
	tctlcfg "github.com/gravitational/teleport/tool/tctl/common/config"
)

// authCommandClient is aggregated client interface for auth command.
type authCommandClient interface {
	authclient.ClientI
}

// AuthCommand implements `tctl auth` group of commands
type AuthCommand struct {
	config                     *servicecfg.Config
	authType                   string
	genPubPath                 string
	genPrivPath                string
	genUser                    string
	genHost                    string
	format                     string
	genTTL                     time.Duration
	exportAuthorityFingerprint string
	exportPrivateKeys          bool
	output                     string
	outputFormat               identityfile.Format
	compatVersion              string
	compatibility              string
	proxyAddr                  string
	leafCluster                string
	kubeCluster                string
	appName                    string
	dbService                  string
	dbName                     string
	dbUser                     string
	windowsUser                string
	windowsDomain              string
	windowsPKIDomain           string
	windowsSID                 string
	omitCDP                    bool
	signOverwrite              bool
	password                   string
	caType                     string
	streamTarfile              bool
	identityWriter             identityfile.ConfigWriter
	integration                string

	authRotate authRotateCommand

	authGenerate *kingpin.CmdClause
	authExport   *kingpin.CmdClause
	authSign     *kingpin.CmdClause
	authLS       *kingpin.CmdClause
	authCRL      *kingpin.CmdClause
	// testInsecureSkipVerify is used to skip TLS verification during tests
	// when connecting to the proxy ping address.
	testInsecureSkipVerify bool
}

// Initialize allows TokenCommand to plug itself into the CLI parser
func (a *AuthCommand) Initialize(app *kingpin.Application, _ *tctlcfg.GlobalCLIFlags, config *servicecfg.Config) {
	a.config = config
	// operations with authorities
	auth := app.Command("auth", "Operations with user and host certificate authorities (CAs).").Hidden()
	a.authExport = auth.Command("export", "Export public cluster CA certificates to stdout.")
	a.authExport.Flag("keys", "if set, will print private keys").BoolVar(&a.exportPrivateKeys)
	a.authExport.Flag("fingerprint", "filter authority by fingerprint").StringVar(&a.exportAuthorityFingerprint)
	a.authExport.Flag("compat", "export certificates compatible with specific version of Teleport").StringVar(&a.compatVersion)
	a.authExport.Flag("type",
		fmt.Sprintf("export certificate type (%v)", strings.Join(allowedCertificateTypes, ", "))).
		EnumVar(&a.authType, allowedCertificateTypes...)
	a.authExport.Flag("integration", "Name of the integration. Only applies to \"github\" CAs.").StringVar(&a.integration)
	a.authExport.
		Flag("out", "If set writes exported authorities to files with the given path prefix").
		StringVar(&a.output)

	a.authGenerate = auth.Command("gen", "Generate a new SSH keypair.").Hidden()
	a.authGenerate.Flag("pub-key", "path to the public key").Required().StringVar(&a.genPubPath)
	a.authGenerate.Flag("priv-key", "path to the private key").Required().StringVar(&a.genPrivPath)

	a.authSign = auth.Command("sign", "Create an identity file(s) for a given user.")
	a.authSign.Flag("user", "Teleport user name").StringVar(&a.genUser)
	a.authSign.Flag("host", "Teleport host name").StringVar(&a.genHost)
	a.authSign.Flag("out", "Identity output").Short('o').Required().StringVar(&a.output)
	a.authSign.Flag("format",
		fmt.Sprintf("Identity format: %s. %s is the default.",
			identityfile.KnownFileFormats.String(), identityfile.DefaultFormat)).
		Default(string(identityfile.DefaultFormat)).
		StringVar((*string)(&a.outputFormat))
	a.authSign.Flag("ttl", "TTL (time to live) for the generated certificate.").
		Default(fmt.Sprintf("%v", apidefaults.CertDuration)).
		DurationVar(&a.genTTL)
	a.authSign.Flag("compat", "OpenSSH compatibility flag").StringVar(&a.compatibility)
	a.authSign.Flag("proxy", `Address of the Teleport proxy. When --format is set to "kubernetes", this address will be set as cluster address in the generated kubeconfig file`).StringVar(&a.proxyAddr)
	a.authSign.Flag("overwrite", "Whether to overwrite existing destination files. When not set, user will be prompted before overwriting any existing file.").BoolVar(&a.signOverwrite)
	a.authSign.Flag("tar", "Create a tarball of the resulting certificates and stream to stdout.").BoolVar(&a.streamTarfile)
	// --kube-cluster was an unfortunately chosen flag name, before teleport
	// supported kubernetes_service and registered kubernetes clusters that are
	// not trusted teleport clusters.
	// It's kept as an alias for --leaf-cluster for backwards-compatibility,
	// but hidden.
	a.authSign.Flag("kube-cluster", `Leaf cluster to generate identity file for when --format is set to "kubernetes"`).Hidden().StringVar(&a.leafCluster)
	a.authSign.Flag("leaf-cluster", `Leaf cluster to generate identity file for when --format is set to "kubernetes"`).StringVar(&a.leafCluster)
	a.authSign.Flag("kube-cluster-name", `Kubernetes cluster to generate identity file for when --format is set to "kubernetes"`).StringVar(&a.kubeCluster)
	a.authSign.Flag("app-name", `Application to generate identity file for. Mutually exclusive with "--db-service".`).StringVar(&a.appName)
	a.authSign.Flag("db-service", `Database to generate identity file for. Mutually exclusive with "--app-name".`).StringVar(&a.dbService)
	a.authSign.Flag("db-user", `Database user placed on the identity file. Only used when "--db-service" is set.`).StringVar(&a.dbUser)
	a.authSign.Flag("db-name", `Database name placed on the identity file. Only used when "--db-service" is set.`).StringVar(&a.dbName)
	a.authSign.Flag("windows-user", `Window user placed on the identity file. Only used when --format is set to "windows"`).StringVar(&a.windowsUser)
	a.authSign.Flag("windows-domain", `Active Directory domain for which this cert is valid. Only used when --format is set to "windows"`).StringVar(&a.windowsDomain)
	a.authSign.Flag("windows-pki-domain", `Active Directory domain where CRLs will be located. Only used when --format is set to "windows"`).StringVar(&a.windowsPKIDomain)
	a.authSign.Flag("windows-sid", `Optional Security Identifier to embed in the certificate. Only used when --format is set to "windows"`).StringVar(&a.windowsSID)
	a.authSign.Flag("omit-cdp", `Omit CRL Distribution Points from the cert. Only used when --format is set to "windows"`).BoolVar(&a.omitCDP)

	a.authRotate.Initialize(auth)

	a.authLS = auth.Command("ls", "List connected auth servers.")
	a.authLS.Flag("format", "Output format: 'yaml', 'json' or 'text'").Default(teleport.YAML).StringVar(&a.format)

	a.authCRL = auth.Command("crl", "Export empty certificate revocation list (CRL) for certificate authorities.")
	a.authCRL.Flag("type", fmt.Sprintf("Certificate authority type, one of: %s", strings.Join(allowedCRLCertificateTypes, ", "))).Required().EnumVar(&a.caType, allowedCRLCertificateTypes...)
	a.authCRL.Flag("out", "If set, writes exported revocation lists to files with the given path prefix").StringVar(&a.output)
}

// TryRun takes the CLI command as an argument (like "auth gen") and executes it
// or returns match=false if 'cmd' does not belong to it
func (a *AuthCommand) TryRun(ctx context.Context, cmd string, clientFunc commonclient.InitFunc) (match bool, err error) {
	if match, err := a.authRotate.TryRun(ctx, cmd, clientFunc); match || err != nil {
		return match, trace.Wrap(err)
	}

	var commandFunc func(ctx context.Context, client authCommandClient) error
	switch cmd {
	case a.authGenerate.FullCommand():
		commandFunc = a.GenerateKeys
	case a.authExport.FullCommand():
		commandFunc = a.ExportAuthorities
	case a.authSign.FullCommand():
		commandFunc = a.GenerateAndSignKeys
	case a.authLS.FullCommand():
		commandFunc = a.ListAuthServers
	case a.authCRL.FullCommand():
		commandFunc = a.GenerateCRLForCA
	default:
		return false, nil
	}
	client, closeFn, err := clientFunc(ctx)
	if err != nil {
		return false, trace.Wrap(err)
	}
	err = commandFunc(ctx, client)
	closeFn(ctx)

	return true, trace.Wrap(err)
}

var allowedCertificateTypes = []string{
	"user",
	"host",
	"tls-host",
	"tls-user",
	"tls-user-der",
	"tls-spiffe",
	"windows",
	"db",
	"db-der",
	"db-client",
	"db-client-der",
	"openssh",
	"saml-idp",
	"github",
	"awsra",
}

// allowedCRLCertificateTypes list of certificate authorities types that can
// have a CRL exported.
var allowedCRLCertificateTypes = []string{
	string(types.HostCA),
	string(types.DatabaseCA),
	string(types.DatabaseClientCA),
	string(types.UserCA),
}

func (a *AuthCommand) exportAuthorities(ctx context.Context, clt authCommandClient) ([]*client.ExportedAuthority, error) {
	switch {
	case client.IsIntegrationAuthorityType(a.authType):
		if a.exportPrivateKeys {
			return nil, trace.BadParameter("exporting private keys is not supported for integration authorities")
		}
		return client.ExportIntegrationAuthorities(ctx, clt, client.ExportIntegrationAuthoritiesRequest{
			AuthType:         a.authType,
			MatchFingerprint: a.exportAuthorityFingerprint,
			Integration:      a.integration,
		})

	case a.exportPrivateKeys:
		return client.ExportAllAuthoritiesSecrets(ctx, clt, client.ExportAuthoritiesRequest{
			AuthType:                   a.authType,
			ExportAuthorityFingerprint: a.exportAuthorityFingerprint,
			UseCompatVersion:           a.compatVersion == "1.0",
		})
	default:
		return client.ExportAllAuthorities(ctx, clt, client.ExportAuthoritiesRequest{
			AuthType:                   a.authType,
			ExportAuthorityFingerprint: a.exportAuthorityFingerprint,
			UseCompatVersion:           a.compatVersion == "1.0",
		})
	}
}

// ExportAuthorities outputs the list of authorities in OpenSSH compatible formats
// If --type flag is given, only prints keys for CAs of this type, otherwise
// prints all keys
func (a *AuthCommand) ExportAuthorities(ctx context.Context, clt authCommandClient) error {
	authorities, err := a.exportAuthorities(ctx, clt)
	if err != nil {
		return trace.Wrap(err)
	}

	if l := len(authorities); l > 1 && a.output == "" {
		return trace.BadParameter("found %d authorities to export, use --out to export all", l)
	}

	if a.output != "" {
		perms := os.FileMode(0644)
		if a.exportPrivateKeys {
			perms = 0600
		}

		fmt.Fprintf(os.Stderr, "Writing %d files with prefix %q\n", len(authorities), a.output)
		for i, authority := range authorities {
			name := fmt.Sprintf("%s%d.cer", a.output, i)
			if err := os.WriteFile(name, authority.Data, perms); err != nil {
				return trace.Wrap(err)
			}
			fmt.Fprintln(os.Stderr, name)
		}
		return nil
	}

	// Only a single CA is exported if we got this far.
	fmt.Printf("%s\n", authorities[0].Data)
	return nil
}

// GenerateKeys generates a new keypair
func (a *AuthCommand) GenerateKeys(ctx context.Context, clusterAPI authCommandClient) error {
	signer, err := cryptosuites.GenerateKey(ctx,
		cryptosuites.GetCurrentSuiteFromPing(clusterAPI),
		cryptosuites.UserSSH)
	if err != nil {
		return trace.Wrap(err)
	}
	key, err := keys.NewPrivateKey(signer)
	if err != nil {
		return trace.Wrap(err)
	}

	pubBytes := key.MarshalSSHPublicKey()
	privBytes, err := key.MarshalSSHPrivateKey()
	if err != nil {
		return trace.Wrap(err)
	}

	err = os.WriteFile(a.genPubPath, pubBytes, 0o600)
	if err != nil {
		return trace.Wrap(err)
	}

	err = os.WriteFile(a.genPrivPath, privBytes, 0o600)
	if err != nil {
		return trace.Wrap(err)
	}

	fmt.Printf("wrote public key to: %v and private key to: %v\n", a.genPubPath, a.genPrivPath)
	return nil
}

// certificateSigner is an interface for the methods used by GenerateAndSignKeys
// to sign certificates using the Auth Server.
type certificateSigner interface {
	GenerateDatabaseCert(context.Context, *proto.DatabaseCertRequest) (*proto.DatabaseCertResponse, error)
	GenerateUserCerts(ctx context.Context, req proto.UserCertsRequest) (*proto.Certs, error)
	GenerateWindowsDesktopCert(context.Context, *proto.WindowsDesktopCertRequest) (*proto.WindowsDesktopCertResponse, error)
	GetApplicationServers(ctx context.Context, namespace string) ([]types.AppServer, error)
	GetCertAuthorities(ctx context.Context, caType types.CertAuthType, loadKeys bool) ([]types.CertAuthority, error)
	GetCertAuthority(ctx context.Context, id types.CertAuthID, loadKeys bool) (types.CertAuthority, error)
	GetClusterName(ctx context.Context) (types.ClusterName, error)
	GetClusterNetworkingConfig(ctx context.Context) (types.ClusterNetworkingConfig, error)
	GetDatabaseServers(ctx context.Context, namespace string, opts ...services.MarshalOption) ([]types.DatabaseServer, error)
	GetProxies() ([]types.Server, error)
	GetRemoteClusters(ctx context.Context) ([]types.RemoteCluster, error)
	TrustClient() trustpb.TrustServiceClient
	Ping(context.Context) (proto.PingResponse, error)
}

// GenerateAndSignKeys generates a new keypair and signs it for role
func (a *AuthCommand) GenerateAndSignKeys(ctx context.Context, clusterAPI authCommandClient) error {
	if a.streamTarfile {
		tarWriter := newTarWriter(clockwork.NewRealClock())
		defer tarWriter.Archive(os.Stdout)
		a.identityWriter = tarWriter
	}

	switch a.outputFormat {
	case identityfile.FormatDatabase, identityfile.FormatMongo, identityfile.FormatCockroach,
		identityfile.FormatRedis, identityfile.FormatElasticsearch:
		return a.generateDatabaseKeys(ctx, clusterAPI)
	case identityfile.FormatCassandra, identityfile.FormatScylla:
		jskPass, err := utils.CryptoRandomHex(32)
		if err != nil {
			return trace.Wrap(err)
		}
		a.password = jskPass
		return a.generateDatabaseKeys(ctx, clusterAPI)
	case identityfile.FormatSnowflake:
		return a.generateSnowflakeKey(ctx, clusterAPI)
	case identityfile.FormatWindows:
		return a.generateWindowsCert(ctx, clusterAPI)
	case identityfile.FormatOracle:
		oracleWalletPass, err := utils.CryptoRandomHex(32)
		if err != nil {
			return trace.Wrap(err)
		}
		a.password = oracleWalletPass
		return a.generateDatabaseKeys(ctx, clusterAPI)

	}
	switch {
	case a.genUser != "" && a.genHost == "":
		return a.generateUserKeys(ctx, clusterAPI)
	case a.genUser == "" && a.genHost != "":
		return a.generateHostKeys(ctx, clusterAPI)
	default:
		return trace.BadParameter("--user or --host must be specified")
	}
}

func (a *AuthCommand) generateWindowsCert(ctx context.Context, clusterAPI certificateSigner) error {
	var missingFlags []string
	if len(a.windowsUser) == 0 {
		missingFlags = append(missingFlags, "--windows-user")
	}
	if len(a.windowsDomain) == 0 {
		missingFlags = append(missingFlags, "--windows-domain")
	}
	if len(missingFlags) > 0 {
		return trace.BadParameter("the following flags are missing: %v",
			strings.Join(missingFlags, ", "))
	}

	cn, err := clusterAPI.GetClusterName(ctx)
	if err != nil {
		return trace.Wrap(err)
	}

	certDER, _, err := winpki.GenerateWindowsDesktopCredentials(ctx, clusterAPI, &winpki.GenerateCredentialsRequest{
		CAType:             types.UserCA,
		Username:           a.windowsUser,
		Domain:             a.windowsDomain,
		PKIDomain:          a.windowsPKIDomain,
		ActiveDirectorySID: a.windowsSID,
		TTL:                a.genTTL,
		ClusterName:        cn.GetClusterName(),
		OmitCDP:            a.omitCDP,
	})
	if err != nil {
		return trace.Wrap(err)
	}

	_, err = identityfile.Write(ctx, identityfile.WriteConfig{
		OutputPath:           a.output,
		WindowsDesktopCerts:  map[string][]byte{a.windowsUser: certDER},
		Format:               a.outputFormat,
		OverwriteDestination: a.signOverwrite,
		Writer:               a.identityWriter,
	})
	if err != nil {
		return trace.Wrap(err)
	}

	return nil
}

// generateSnowflakeKey exports DatabaseCA public key in the format required by Snowflake
// Ref: https://docs.snowflake.com/en/user-guide/key-pair-auth.html#step-2-generate-a-public-key
func (a *AuthCommand) generateSnowflakeKey(ctx context.Context, clusterAPI certificateSigner) error {
	keyRing, err := generateKeyRing(ctx, clusterAPI, cryptosuites.DatabaseClient)
	if err != nil {
		return trace.Wrap(err)
	}

	cn, err := clusterAPI.GetClusterName(ctx)
	if err != nil {
		return trace.Wrap(err)
	}
	certAuthID := types.CertAuthID{
		Type:       types.DatabaseClientCA,
		DomainName: cn.GetClusterName(),
	}
	dbClientCA, err := clusterAPI.GetCertAuthority(ctx, certAuthID, false)
	if err != nil {
		return trace.Wrap(err)
	}
	keyRing.TrustedCerts = []authclient.TrustedCerts{{TLSCertificates: services.GetTLSCerts(dbClientCA)}}

	filesWritten, err := identityfile.Write(ctx, identityfile.WriteConfig{
		OutputPath:           a.output,
		KeyRing:              keyRing,
		Format:               a.outputFormat,
		OverwriteDestination: a.signOverwrite,
		Writer:               a.identityWriter,
	})
	if err != nil {
		return trace.Wrap(err)
	}

	return trace.Wrap(
		writeHelperMessageDBmTLS(a.helperMsgDst(), filesWritten, "", a.outputFormat, "", a.streamTarfile))
}

// ListAuthServers prints a list of connected auth servers
func (a *AuthCommand) ListAuthServers(ctx context.Context, clusterAPI authCommandClient) error {
	servers, err := clusterAPI.GetAuthServers()
	if err != nil {
		return trace.Wrap(err)
	}

	sc := &serverCollection{servers}

	switch a.format {
	case teleport.Text:
		// auth servers don't have labels.
		verbose := false
		return sc.writeText(os.Stdout, verbose)
	case teleport.YAML:
		return writeYAML(sc, os.Stdout)
	case teleport.JSON:
		return writeJSON(sc, os.Stdout)
	}

	return nil
}

// GenerateCRLForCA generates a certificate revocation list for a certificate
// authority.
func (a *AuthCommand) GenerateCRLForCA(ctx context.Context, clusterAPI authCommandClient) error {
	certType := types.CertAuthType(a.caType)
	if err := certType.Check(); err != nil {
		return trace.Wrap(err)
	}
	clusterName, err := clusterAPI.GetClusterName(ctx)
	if err != nil {
		return trace.Wrap(err)
	}
	authority, err := clusterAPI.GetCertAuthority(ctx, types.CertAuthID{
		Type:       certType,
		DomainName: clusterName.GetClusterName(),
	}, false)
	if err != nil {
		return trace.Wrap(err)
	}

	tlsKeys := authority.GetActiveKeys().TLS
	if len(tlsKeys) == 0 {
		return trace.BadParameter("CA has no active keys")
	}

	if a.output == "" {
		if len(tlsKeys) > 1 {
			return trace.BadParameter("CA has multiple active keys, use --out to export all CRLs")
		}
		crl := tlsKeys[0].CRL
		if len(crl) == 0 {
			fmt.Fprintf(os.Stderr, "keypair is missing CRL for %v authority %v, generating legacy fallback", authority.GetType(), authority.GetName())
			crl, err = clusterAPI.GenerateCertAuthorityCRL(ctx, certType)
			if err != nil {
				return trace.Wrap(err)
			}
		}
		fmt.Print(string(crl))
		return nil
	}

	// collect the CRLs ahead of time so we can print a message
	// like we do with tctl auth export
	type output struct{ cert, crl []byte }
	var results []output
	for i, keypair := range tlsKeys {
		crl := keypair.CRL
		if len(crl) == 0 {
			fmt.Fprintf(os.Stderr, "keypair %v is missing CRL for %v authority %v, generating legacy fallback", i, authority.GetType(), authority.GetName())
			crl, err = clusterAPI.GenerateCertAuthorityCRL(ctx, certType)
			if err != nil {
				return trace.Wrap(err)
			}
		}
		results = append(results, output{keypair.Cert, crl})
	}

	fmt.Fprintf(os.Stderr, "Writing %d files with prefix %q\n", len(results), a.output)
	commands := make([]string, len(results))
	for i, out := range results {
		block, _ := pem.Decode(out.cert)
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return trace.Wrap(err)
		}

		cn := base32.HexEncoding.EncodeToString(cert.SubjectKeyId) + "_" + cert.Subject.CommonName
		filename := fmt.Sprintf("%s-%v-%v.crl", a.output, certType, cn)
		if err := os.WriteFile(filename, out.crl, os.FileMode(0644)); err != nil {
			return trace.Wrap(err)
		}
		fmt.Fprintln(os.Stderr, filename)
		commands[i] = fmt.Sprintf("certutil -dspublish %s TeleportDB %s", filename, cn)
	}

	if certType == types.DatabaseClientCA && len(results) > 1 {
		fmt.Fprintln(os.Stderr, "\nTo publish CRLs, run the following in Windows:")
		for _, command := range commands {
			fmt.Fprintln(os.Stderr, "  "+command)
		}
	}

	return nil
}

func (a *AuthCommand) generateHostKeys(ctx context.Context, clusterAPI certificateSigner) error {
	// only format=openssh is supported
	if a.outputFormat != identityfile.FormatOpenSSH {
		return trace.BadParameter("invalid --format flag %q, only %q is supported", a.outputFormat, identityfile.FormatOpenSSH)
	}

	// split up comma separated list
	principals := strings.Split(a.genHost, ",")

	// Generate an SSH key.
	signer, err := cryptosuites.GenerateKey(ctx,
		cryptosuites.GetCurrentSuiteFromPing(clusterAPI),
		cryptosuites.HostSSH)
	if err != nil {
		return trace.Wrap(err)
	}
	key, err := keys.NewPrivateKey(signer)
	if err != nil {
		return trace.Wrap(err)
	}

	cn, err := clusterAPI.GetClusterName(ctx)
	if err != nil {
		return trace.Wrap(err)
	}
	clusterName := cn.GetClusterName()

	res, err := clusterAPI.TrustClient().GenerateHostCert(ctx, &trustpb.GenerateHostCertRequest{
		Key:         key.MarshalSSHPublicKey(),
		HostId:      "",
		NodeName:    "",
		Principals:  principals,
		ClusterName: clusterName,
		Role:        string(types.RoleNode),
		Ttl:         durationpb.New(0),
	})
	if err != nil {
		return trace.Wrap(err)
	}

	hostCAs, err := clusterAPI.GetCertAuthorities(ctx, types.HostCA, false)
	if err != nil {
		return trace.Wrap(err)
	}
	keyRing := &client.KeyRing{
		SSHPrivateKey: key,
		Cert:          res.SshCertificate,
		TrustedCerts:  authclient.AuthoritiesToTrustedCerts(hostCAs),
	}

	// if no name was given, take the first name on the list of principals
	filePath := a.output
	if filePath == "" {
		filePath = principals[0]
	}

	filesWritten, err := identityfile.Write(ctx, identityfile.WriteConfig{
		OutputPath:           filePath,
		KeyRing:              keyRing,
		Format:               a.outputFormat,
		OverwriteDestination: a.signOverwrite,
		Writer:               a.identityWriter,
	})
	if err != nil {
		return trace.Wrap(err)
	}

	fmt.Fprintf(a.helperMsgDst(), "\nThe credentials have been written to %s\n", strings.Join(filesWritten, ", "))

	return nil
}

// generateDatabaseKeys generates a new unsigned key and signs it with Teleport
// CA for database access.
func (a *AuthCommand) generateDatabaseKeys(ctx context.Context, clusterAPI certificateSigner) error {
	keyRing, err := generateKeyRing(ctx, clusterAPI, cryptosuites.DatabaseClient)
	if err != nil {
		return trace.Wrap(err)
	}
	return a.generateDatabaseKeysForKeyRing(ctx, clusterAPI, keyRing)
}

// generateDatabaseKeysForKeyRing signs the provided unsigned key with Teleport CA
// for database access.
func (a *AuthCommand) generateDatabaseKeysForKeyRing(ctx context.Context, clusterAPI certificateSigner, keyRing *client.KeyRing) error {
	principals := strings.Split(a.genHost, ",")

	dbCertReq := db.GenerateDatabaseCertificatesRequest{
		ClusterAPI:         clusterAPI,
		Principals:         principals,
		OutputFormat:       a.outputFormat,
		OutputCanOverwrite: a.signOverwrite,
		OutputLocation:     a.output,
		TTL:                a.genTTL,
		KeyRing:            keyRing,
		Password:           a.password,
		IdentityFileWriter: a.identityWriter,
	}
	filesWritten, err := db.GenerateDatabaseServerCertificates(ctx, dbCertReq)
	if err != nil {
		return trace.Wrap(err)
	}

	return trace.Wrap(writeHelperMessageDBmTLS(a.helperMsgDst(), filesWritten, a.output, a.outputFormat, a.password, a.streamTarfile))
}

var mapIdentityFileFormatHelperTemplate = map[identityfile.Format]*template.Template{
	identityfile.FormatDatabase:      dbAuthSignTpl,
	identityfile.FormatMongo:         mongoAuthSignTpl,
	identityfile.FormatCockroach:     cockroachAuthSignTpl,
	identityfile.FormatRedis:         redisAuthSignTpl,
	identityfile.FormatSnowflake:     snowflakeAuthSignTpl,
	identityfile.FormatElasticsearch: elasticsearchAuthSignTpl,
	identityfile.FormatCassandra:     cassandraAuthSignTpl,
	identityfile.FormatScylla:        scyllaAuthSignTpl,
	identityfile.FormatOracle:        oracleAuthSignTpl,
}

func writeHelperMessageDBmTLS(writer io.Writer, filesWritten []string, output string, outputFormat identityfile.Format, password string, tarOutput bool) error {
	if writer == nil {
		return nil
	}

	tpl, found := mapIdentityFileFormatHelperTemplate[outputFormat]
	if !found {
		// This format doesn't have a recommended configuration.
		// Consider adding one to ease the installation for the end-user
		return nil
	}
	tplVars := map[string]any{
		"files":     strings.Join(filesWritten, ", "),
		"password":  password,
		"output":    output,
		"tarOutput": tarOutput,
	}
	switch outputFormat {
	case defaults.ProtocolCockroachDB:
		tplVars["clientCAPath"] = "/path/to/client-ca.key"
	case defaults.ProtocolOracle:
		tplVars["manualOrapkiFlow"] = len(filesWritten) != 1
		// use a generic example path since they will have to copy the files
		// to the oracle server.
		tplVars["walletDir"] = "/path/to/oracleWalletDir"
		var caCertPaths []string
		for _, f := range filesWritten {
			if strings.HasSuffix(f, ".crt") {
				caCertPaths = append(caCertPaths, f)
			}
		}
		tplVars["caCertPaths"] = caCertPaths
	}

	return trace.Wrap(tpl.Execute(writer, tplVars))
}

var (
	// dbAuthSignTpl is printed when user generates credentials for a self-hosted database.
	dbAuthSignTpl = template.Must(template.New("").Parse(
		`{{if .tarOutput }}
To unpack the tar archive, pipe the output of tctl to 'tar x'. For example:

$ tctl auth sign ${FLAGS} | tar -xv
{{else}}
Database credentials have been written to {{.files}}.
{{end}}

To enable mutual TLS on your PostgreSQL server, add the following to its postgresql.conf configuration file:

ssl = on
ssl_cert_file = '/path/to/{{.output}}.crt'
ssl_key_file = '/path/to/{{.output}}.key'
ssl_ca_file = '/path/to/{{.output}}.cas'

To enable mutual TLS on your MySQL server, add the following to its mysql.cnf configuration file:

[mysqld]
require_secure_transport=ON
ssl-cert=/path/to/{{.output}}.crt
ssl-key=/path/to/{{.output}}.key
ssl-ca=/path/to/{{.output}}.cas
`))
	// mongoAuthSignTpl is printed when user generates credentials for a MongoDB database.
	mongoAuthSignTpl = template.Must(template.New("").Parse(
		`{{- if .tarOutput -}}
To unpack the tar archive, pipe the output of tctl to 'tar -x'. For example:

$ tctl auth sign ${FLAGS} | tar -x
{{- else -}}
Database credentials have been written to {{.files}}.
{{- end }}

To enable mutual TLS on your MongoDB server, add the following to its
mongod.yaml configuration file:

net:
  tls:
    mode: requireTLS
    certificateKeyFile: /path/to/{{.output}}.crt
    CAFile: /path/to/{{.output}}.cas
`))
	cockroachAuthSignTpl = template.Must(template.New("").Parse(`Database credentials have been written to {{.files}}.

To enable mutual TLS on your CockroachDB server, generate a client CA and client
certs for your node:

# --overwrite flag prepends the client CA cert to {{.output}}/ca-client.crt
cockroach cert create-client-ca \
    --certs-dir={{.output}} \
    --ca-key={{.clientCAPath}} \
    --overwrite

cockroach cert create-client node \
    --certs-dir={{.output}}
    --ca-key={{.clientCAPath}}

Then point cockroach to the certs directory using the --certs-dir flag:

cockroach start \
  --certs-dir={{.output}} \
  # other flags...

For more information about creating a client CA and issuing certs, see:
https://www.cockroachlabs.com/docs/stable/cockroach-cert

Teleport uses a split CA architecture for database access.
For more information about using a split CA with CockroachDB, see:
https://www.cockroachlabs.com/docs/stable/authentication#using-split-ca-certificates
`))

	redisAuthSignTpl = template.Must(template.New("").Parse(
		`{{- if .tarOutput }}
Unpack the tar archive by piping the output of tctl to 'tar x'. For example:

$ tctl auth sign ${CERT_FLAGS} | tar -xv
{{else}}
Database credentials have been written to {{.files}}.
{{end}}

To enable mutual TLS on your Redis server, add the following to your redis.conf:

tls-ca-cert-file /path/to/{{.output}}.cas
tls-cert-file /path/to/{{.output}}.crt
tls-key-file /path/to/{{.output}}.key
tls-protocols "TLSv1.2 TLSv1.3"

For information on enabling Redis Cluster bus communication TLS, see:
https://goteleport.com/docs/database-access/guides/redis-cluster
`))

	snowflakeAuthSignTpl = template.Must(template.New("").Parse(`Database credentials have been written to {{.files}}.

Please add the generated key to the Snowflake users as described here:
https://docs.snowflake.com/en/user-guide/key-pair-auth.html#step-4-assign-the-public-key-to-a-snowflake-user
`))

	elasticsearchAuthSignTpl = template.Must(template.New("").Parse(
		`{{- if .tarOutput -}}
To unpack the tar archive, pipe the output of tctl to 'tar -x'. For example:

$ tctl auth sign ${FLAGS} | tar -x
{{- else -}}
Database credentials have been written to {{.files}}.
{{- end }}

To enable mutual TLS on your Elasticsearch server, add the following to your elasticsearch.yml:

xpack.security.http.ssl:
  certificate_authorities: /path/to/{{.output}}.cas
  certificate: /path/to/{{.output}}.crt
  key: /path/to/{{.output}}.key
  enabled: true
  client_authentication: required
  verification_mode: certificate

xpack.security.authc.realms.pki.pki1:
  order: 1
  enabled: true
  certificate_authorities: /path/to/{{.output}}.cas

For more information on configuring security settings in Elasticsearch, see:
https://www.elastic.co/guide/en/elasticsearch/reference/current/security-settings.html
`))

	cassandraAuthSignTpl = template.Must(template.New("").Parse(
		`{{- if .tarOutput -}}
To unpack the tar archive, pipe the output of tctl to 'tar -x'. For example:

$ tctl auth sign ${FLAGS} | tar -x
{{- else -}}
Database credentials have been written to {{.files}}.
{{- end }}

To enable mutual TLS on your Cassandra server, add the following to your
cassandra.yaml configuration file:
client_encryption_options:
   enabled: true
   optional: false
   keystore: /path/to/{{.output}}.keystore
   keystore_password: "{{.password}}"
   require_client_auth: true
   truststore: /path/to/{{.output}}.truststore
   truststore_password: "{{.password}}"
   protocol: TLS
   algorithm: SunX509
   store_type: JKS
   cipher_suites: [TLS_RSA_WITH_AES_256_CBC_SHA]
`))

	oracleAuthSignTpl = template.Must(template.New("").Parse(
		`{{- if .tarOutput -}}
To unpack the tar archive, pipe the output of tctl to 'tar -x'. For example:

$ tctl auth sign ${FLAGS} | tar -x
{{- end }}
{{- if .manualOrapkiFlow}}
Orapki binary was not found. Please create oracle wallet file manually by running the following commands on the Oracle server:

WALLET_DIR="{{.walletDir}}"
orapki wallet create -wallet "$WALLET_DIR" -auto_login_only
orapki wallet import_pkcs12 -wallet "$WALLET_DIR" -auto_login_only -pkcs12file {{.output}}.p12 -pkcs12pwd {{.password}}
{{- range $certPath := .caCertPaths }}
orapki wallet add -wallet "$WALLET_DIR" -trusted_cert -auto_login_only -cert {{ $certPath }}
{{- end}}

If copying these files to your Oracle server, ensure the cert file permissions are readable by the "oracle" user.
{{else}}
Oracle wallet stored in {{.output}} directory created with Oracle Orapki.

{{end}}
To enable mutual TLS on your Oracle server, add the following settings to Oracle sqlnet.ora configuration file:

WALLET_LOCATION = (SOURCE = (METHOD = FILE)(METHOD_DATA = (DIRECTORY = {{.walletDir}})))
SSL_CLIENT_AUTHENTICATION = TRUE
SQLNET.AUTHENTICATION_SERVICES = (TCPS)


To enable mutual TLS on your Oracle server, add the following TCPS entries to listener.ora configuration file:

LISTENER =
  (DESCRIPTION_LIST =
    (DESCRIPTION =
      (ADDRESS = (PROTOCOL = TCPS)(HOST = 0.0.0.0)(PORT = 2484))
    )
  )

WALLET_LOCATION = (SOURCE = (METHOD = FILE)(METHOD_DATA = (DIRECTORY = {{.walletDir}})))
SSL_CLIENT_AUTHENTICATION = TRUE
`))

	scyllaAuthSignTpl = template.Must(template.New("").Parse(`Database credentials have been written to {{.files}}.

To enable mutual TLS on your Scylla server, add the following to your
scylla.yaml configuration file:

client_encryption_options:
   enabled: true
   certificate: /path/to/{{.output}}.crt
   keyfile: /path/to/{{.output}}.key
   truststore:  /path/to/{{.output}}.cas
   require_client_auth: True
`))
)

// generateKeyRing generates and returns a keyring using a key algorithm
// determined by the current cluster signature algorithm suite and [purpose].
// The returned KeyRing always uses a single private key for both SSH and TLS,
// this is a deliberate choice for `tctl auth sign` which either only uses a
// single protocol anyway, or writes to an identity file which only supports a
// single private key.
func generateKeyRing(ctx context.Context, clusterAPI certificateSigner, purpose cryptosuites.KeyPurpose) (*client.KeyRing, error) {
	signer, err := cryptosuites.GenerateKey(ctx,
		cryptosuites.GetCurrentSuiteFromPing(clusterAPI),
		purpose)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	key, err := keys.NewPrivateKey(signer)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return client.NewKeyRing(key, key), nil
}

func (a *AuthCommand) generateUserKeys(ctx context.Context, clusterAPI certificateSigner) error {
	// Validate --proxy flag.
	if err := a.checkProxyAddr(ctx, clusterAPI); err != nil {
		return trace.Wrap(err)
	}
	// parse compatibility parameter
	certificateFormat, err := utils.CheckCertificateFormatFlag(a.compatibility)
	if err != nil {
		return trace.Wrap(err)
	}

	// The output here is likely to be written to an identity file which only
	// supports a single key for SSH and TLS. SSH supports all the TLS key
	// algorithms but TLS does not support all the SSH key algorithms (Ed25519),
	// so use the UserTLS key purpose unless we know this is only for SSH.
	var keyPurpose cryptosuites.KeyPurpose
	switch a.outputFormat {
	case identityfile.FormatOpenSSH:
		keyPurpose = cryptosuites.UserSSH
	default:
		keyPurpose = cryptosuites.UserTLS
	}
	keyRing, err := generateKeyRing(ctx, clusterAPI, keyPurpose)
	if err != nil {
		return trace.Wrap(err)
	}

	if a.leafCluster != "" {
		if err := a.checkLeafCluster(clusterAPI); err != nil {
			return trace.Wrap(err)
		}
	} else {
		cn, err := clusterAPI.GetClusterName(ctx)
		if err != nil {
			return trace.Wrap(err)
		}
		a.leafCluster = cn.GetClusterName()
	}
	keyRing.ClusterName = a.leafCluster

	if err := a.checkKubeCluster(); err != nil {
		return trace.Wrap(err)
	}

	var (
		routeToApp      proto.RouteToApp
		routeToDatabase proto.RouteToDatabase
		certUsage       proto.UserCertsRequest_CertUsage
	)

	// `appName` and `db` are mutually exclusive.
	if a.appName != "" && a.dbService != "" {
		return trace.BadParameter("only --app-name or --db-service can be set, not both")
	}

	switch {
	case a.appName != "":
		server, err := getApplicationServer(ctx, clusterAPI, a.appName)
		if err != nil {
			return trace.Wrap(err)
		}

		routeToApp = proto.RouteToApp{
			Name:        a.appName,
			PublicAddr:  server.GetApp().GetPublicAddr(),
			ClusterName: a.leafCluster,
			URI:         server.GetApp().GetURI(),
		}

		certUsage = proto.UserCertsRequest_App
	case a.dbService != "":
		server, err := getDatabaseServer(ctx, clusterAPI, a.dbService)
		if err != nil {
			return trace.Wrap(err)
		}

		routeToDatabase = proto.RouteToDatabase{
			ServiceName: a.dbService,
			Protocol:    server.GetDatabase().GetProtocol(),
			Database:    a.dbName,
			Username:    a.dbUser,
		}
		certUsage = proto.UserCertsRequest_Database
	}

	sshPublicKey := keyRing.SSHPrivateKey.MarshalSSHPublicKey()
	tlsPublicKey, err := keys.MarshalPublicKey(keyRing.TLSPrivateKey.Public())
	if err != nil {
		return trace.Wrap(err)
	}

	reqExpiry := time.Now().UTC().Add(a.genTTL)
	// Request signed certs from `auth` server.
	certs, err := clusterAPI.GenerateUserCerts(ctx, proto.UserCertsRequest{
		SSHPublicKey:      sshPublicKey,
		TLSPublicKey:      tlsPublicKey,
		Username:          a.genUser,
		Expires:           reqExpiry,
		Format:            certificateFormat,
		RouteToCluster:    a.leafCluster,
		KubernetesCluster: a.kubeCluster,
		RouteToApp:        routeToApp,
		Usage:             certUsage,
		RouteToDatabase:   routeToDatabase,
	})
	if err != nil {
		return trace.Wrap(err)
	}
	keyRing.Cert = certs.SSH
	keyRing.TLSCert = certs.TLS

	hostCAs, err := clusterAPI.GetCertAuthorities(ctx, types.HostCA, false)
	if err != nil {
		return trace.Wrap(err)
	}
	keyRing.TrustedCerts = authclient.AuthoritiesToTrustedCerts(hostCAs)

	// Is TLS routing enabled?
	proxyListenerMode := types.ProxyListenerMode_Separate
	if a.config != nil && a.config.Auth.NetworkingConfig != nil {
		proxyListenerMode = a.config.Auth.NetworkingConfig.GetProxyListenerMode()
	}
	if networkConfig, err := clusterAPI.GetClusterNetworkingConfig(ctx); err == nil {
		proxyListenerMode = networkConfig.GetProxyListenerMode()
	}

	// If we're in multiplexed mode get SNI name for kube from single multiplexed proxy addr
	kubeTLSServerName := ""
	if proxyListenerMode == types.ProxyListenerMode_Multiplex {
		slog.DebugContext(ctx, "Using Proxy SNI for kube TLS server name")
		u, err := parseURL(a.proxyAddr)
		if err != nil {
			return trace.Wrap(err)
		}
		// extract host part if present
		split := strings.Split(u.Host, ":")
		kubeTLSServerName = client.GetKubeTLSServerName(split[0])
	}

	expires, err := keyRing.TeleportTLSCertValidBefore()
	if err != nil {
		slog.WarnContext(ctx, "Failed to check TTL validity", "error", err)
		// err swallowed on purpose
	} else if reqExpiry.Sub(expires) > time.Minute {
		maxAllowedTTL := time.Until(expires).Round(time.Second)
		return trace.BadParameter(`The credential was not issued because the requested TTL of %s exceeded the maximum allowed value of %s. To successfully request a credential, please reduce the requested TTL.`,
			a.genTTL,
			maxAllowedTTL)
	}

	// write the cert+private key to the output:
	filesWritten, err := identityfile.Write(ctx, identityfile.WriteConfig{
		OutputPath:           a.output,
		KeyRing:              keyRing,
		Format:               a.outputFormat,
		KubeProxyAddr:        a.proxyAddr,
		KubeClusterName:      a.kubeCluster,
		KubeTLSServerName:    kubeTLSServerName,
		OverwriteDestination: a.signOverwrite,
		Writer:               a.identityWriter,
	})
	if err != nil {
		return trace.Wrap(err)
	}
	// Print a tip guiding people towards Machine ID. We use stderr here in case
	// someone is programatically parsing stdout.
	_, _ = fmt.Fprintln(
		os.Stderr,
		"\nGenerating credentials to allow a machine access to Teleport? We recommend Teleport's Machine ID! Find out more at https://goteleport.com/r/machineid-tip",
	)

	fmt.Fprintf(a.helperMsgDst(), "The credentials have been written to %s\n", strings.Join(filesWritten, ", "))

	return nil
}

func (a *AuthCommand) checkLeafCluster(clusterAPI certificateSigner) error {
	if a.outputFormat != identityfile.FormatKubernetes && a.leafCluster != "" {
		// User set --cluster but it's not actually used for the chosen --format.
		// Print a warning but continue.
		fmt.Printf("Note: --cluster is only used with --format=%q, ignoring for --format=%q\n", identityfile.FormatKubernetes, a.outputFormat)
	}

	if a.outputFormat != identityfile.FormatKubernetes {
		return nil
	}

	clusters, err := clusterAPI.GetRemoteClusters(context.TODO())
	if err != nil {
		return trace.WrapWithMessage(err, "couldn't load leaf clusters")
	}

	for _, cluster := range clusters {
		if cluster.GetMetadata().Name == a.leafCluster {
			return nil
		}
	}

	return trace.BadParameter("couldn't find leaf cluster named %q", a.leafCluster)
}

func (a *AuthCommand) checkKubeCluster() error {
	if a.kubeCluster == "" {
		return nil
	}
	if a.outputFormat != identityfile.FormatKubernetes && a.kubeCluster != "" {
		// User set --kube-cluster-name but it's not actually used for the chosen --format.
		// Print a warning but continue.
		fmt.Fprintf(a.helperMsgDst(), "Note: --kube-cluster-name is only used with --format=%q, ignoring for --format=%q\n", identityfile.FormatKubernetes, a.outputFormat)
	}
	if a.outputFormat != identityfile.FormatKubernetes {
		return nil
	}

	return nil
}

func (a *AuthCommand) checkProxyAddr(ctx context.Context, clusterAPI certificateSigner) error {
	if a.outputFormat != identityfile.FormatKubernetes && a.proxyAddr != "" {
		// User set --proxy but it's not actually used for the chosen --format.
		// Print a warning but continue.
		fmt.Fprintf(a.helperMsgDst(), "Note: --proxy is only used with --format=%q, ignoring for --format=%q\n", identityfile.FormatKubernetes, a.outputFormat)
		return nil
	}
	if a.outputFormat != identityfile.FormatKubernetes {
		return nil
	}
	if a.proxyAddr != "" {
		// User set --proxy. Validate it and set its scheme to https in case it was omitted.
		u, err := parseURL(a.proxyAddr)
		if err != nil {
			return trace.WrapWithMessage(err, "specified --proxy URL is invalid")
		}
		switch u.Scheme {
		case "":
			u.Scheme = "https"
			a.proxyAddr = u.String()
			return nil
		case "http", "https":
			return nil
		default:
			return trace.BadParameter("expected --proxy URL with http or https scheme")
		}
	}

	// User didn't specify --proxy for kubeconfig. Let's try to guess it.
	//
	// Is the auth server also a proxy?
	if a.config.Proxy.Kube.Enabled {
		var err error
		if a.config.Auth.NetworkingConfig != nil &&
			a.config.Auth.NetworkingConfig.GetProxyListenerMode() == types.ProxyListenerMode_Multiplex {
			a.proxyAddr, err = a.config.Proxy.WebPublicAddr()
			return trace.Wrap(err)
		}
		a.proxyAddr, err = a.config.Proxy.KubeAddr()
		return trace.Wrap(err)
	}
	netConfig, err := clusterAPI.GetClusterNetworkingConfig(ctx)
	if err != nil {
		return trace.WrapWithMessage(err, "couldn't load cluster network configuration, try setting --proxy manually")
	}
	// Fetch proxies known to auth server and try to find a public address.
	proxies, err := clusterAPI.GetProxies()
	if err != nil {
		return trace.WrapWithMessage(err, "couldn't load registered proxies, try setting --proxy manually")
	}
	for _, p := range proxies {
		addr := p.GetPublicAddr()
		if addr == "" {
			continue
		}
		// if the proxy is multiplexing, the public address is the web proxy address.
		if netConfig.GetProxyListenerMode() == types.ProxyListenerMode_Multiplex {
			u := url.URL{
				Scheme: "https",
				Host:   addr,
			}
			a.proxyAddr = u.String()
			return nil
		}

		_, err := utils.ParseAddr(addr)
		if err != nil {
			slog.WarnContext(ctx, "Invalid public address on the proxy",
				"proxy", p.GetName(),
				"public_address", addr,
				"error", err,
			)
			continue
		}

		ping, err := webclient.Ping(
			&webclient.Config{
				Context:   ctx,
				ProxyAddr: addr,
				Timeout:   5 * time.Second,
				Insecure:  a.testInsecureSkipVerify,
			},
		)
		if err != nil {
			slog.WarnContext(ctx, "Unable to ping proxy public address on the proxy",
				"proxy", p.GetName(),
				"public_address", addr,
				"error", err,
			)
			continue
		}

		if !ping.Proxy.Kube.Enabled || ping.Proxy.Kube.PublicAddr == "" {
			continue
		}

		u := url.URL{
			Scheme: "https",
			Host:   ping.Proxy.Kube.PublicAddr,
		}
		a.proxyAddr = u.String()
		return nil
	}

	return trace.BadParameter("couldn't find registered public proxies, specify --proxy when using --format=%q", identityfile.FormatKubernetes)
}

func parseURL(rawurl string) (*url.URL, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	// If no scheme is provided url.Parse fails the parsing and considers the host
	// as scheme, leaving the host empty.
	if u.Host == "" {
		return &url.URL{
			Host: rawurl,
		}, nil
	}

	return u, nil
}

func getApplicationServer(ctx context.Context, clusterAPI certificateSigner, appName string) (types.AppServer, error) {
	servers, err := clusterAPI.GetApplicationServers(ctx, apidefaults.Namespace)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	for _, s := range servers {
		if s.GetName() == appName {
			return s, nil
		}
	}
	return nil, trace.NotFound("app %q not found", appName)
}

// getDatabaseServer fetches a single `DatabaseServer` by name using the
// provided `*auth.Client`.
func getDatabaseServer(ctx context.Context, clientAPI certificateSigner, dbName string) (types.DatabaseServer, error) {
	servers, err := clientAPI.GetDatabaseServers(ctx, apidefaults.Namespace)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	for _, server := range servers {
		if server.GetName() == dbName {
			return server, nil
		}
	}

	return nil, trace.NotFound("database %q not found", dbName)
}

func getCertAuthTypes() []string {
	t := make([]string, 0, len(types.CertAuthTypes)+1)
	for _, at := range types.CertAuthTypes {
		t = append(t, string(at))
	}
	return t
}

func (a *AuthCommand) helperMsgDst() io.Writer {
	if a.streamTarfile {
		return os.Stderr
	}
	return os.Stdout
}
