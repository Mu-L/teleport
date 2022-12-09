/*
Copyright 2016 Gravitational, Inc.

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

package client

import (
	"bufio"
	"context"
	"crypto"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	osfs "io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gravitational/trace"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/gravitational/teleport"
	"github.com/gravitational/teleport/api/constants"
	"github.com/gravitational/teleport/api/profile"
	"github.com/gravitational/teleport/api/utils/keypaths"
	"github.com/gravitational/teleport/api/utils/keys"
	apisshutils "github.com/gravitational/teleport/api/utils/sshutils"
	"github.com/gravitational/teleport/lib/auth"
	"github.com/gravitational/teleport/lib/sshutils"
	"github.com/gravitational/teleport/lib/utils"
	"github.com/gravitational/teleport/lib/utils/agentconn"
)

const (
	// profileDirPerms is the default permissions applied to the profile
	// directory (usually ~/.tsh)
	profileDirPerms os.FileMode = 0700

	// keyFilePerms is the default permissions applied to key files (.cert, .key, pub)
	// under ~/.tsh
	keyFilePerms os.FileMode = 0600

	// tshConfigFileName is the name of the directory containing the
	// tsh config file.
	tshConfigFileName = "config"
)

// KeyStore interface allows for different storage backends for tsh to
// load/save its keys.
//
// The _only_ filesystem-based implementation of KeyStore is declared
// below (FSLocalKeyStore)
type KeyStore interface {
	NonSessionKeyStore

	// AddKey adds the given key to the store.
	AddKey(key *Key) error

	// GetKey returns the user's key including the specified certs.
	GetKey(idx KeyIndex, opts ...CertOption) (*Key, error)

	// DeleteKey deletes the user's key with all its certs.
	DeleteKey(idx KeyIndex) error

	// DeleteUserCerts deletes only the specified certs of the user's key,
	// keeping the private key intact.
	DeleteUserCerts(idx KeyIndex, opts ...CertOption) error

	// DeleteKeys removes all session keys.
	DeleteKeys() error

	// GetSSHCertificates gets all certificates signed for the given user and proxy,
	// including certificates for trusted clusters.
	GetSSHCertificates(proxyHost, username string) ([]*ssh.Certificate, error)
}

type NonSessionKeyStore interface {
	ProfileStore

	// GetKnownHostsFile returns a known hosts file.
	GetKnownHostsFile() ([]byte, error)

	// AddKnownHostKeys adds the public key to the list of known hosts for
	// a hostname.
	AddKnownHostKeys(hostname, proxyHost string, keys []ssh.PublicKey) error

	// GetKnownHostKeys returns all known public host keys for the given host name.
	GetKnownHostKeys(hostname string) ([]ssh.PublicKey, error)

	// SaveTrustedCerts saves trusted TLS certificates of certificate authorities.
	SaveTrustedCerts(proxyHost string, cas []auth.TrustedCerts) error

	// GetTrustedCertsPEM gets trusted TLS certificates of certificate authorities.
	// Each returned byte slice contains an individual PEM block.
	GetTrustedCertsPEM(proxyHost string) ([][]byte, error)
}

// FSLocalKeyStore implements LocalKeyStore interface using the filesystem.
//
// The FS store uses the file layout outlined in `api/utils/keypaths.go`.
type FSLocalKeyStore struct {
	fsLocalNonSessionKeyStore
}

// NewFSLocalKeyStore creates a new filesystem-based local keystore object
// and initializes it.
//
// If dirPath is empty, sets it to ~/.tsh.
func NewFSLocalKeyStore(dirPath string) (s *FSLocalKeyStore, err error) {
	dirPath, err = initKeysDir(dirPath)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &FSLocalKeyStore{
		fsLocalNonSessionKeyStore: fsLocalNonSessionKeyStore{
			log:    logrus.WithField(trace.Component, teleport.ComponentKeyStore),
			KeyDir: dirPath,
		},
	}, nil
}

// initKeysDir initializes the keystore root directory, usually `~/.tsh`.
func initKeysDir(dirPath string) (string, error) {
	dirPath = profile.FullProfilePath(dirPath)
	if err := os.MkdirAll(dirPath, os.ModeDir|profileDirPerms); err != nil {
		return "", trace.ConvertSystemError(err)
	}
	return dirPath, nil
}

func (fs *FSLocalKeyStore) CurrentProfile() (string, error) {
	profileName, err := profile.GetCurrentProfileName(fs.KeyDir)
	if err != nil {
		return "", trace.Wrap(err)
	}
	return profileName, nil
}

func (fs *FSLocalKeyStore) ListProfiles() ([]string, error) {
	profileNames, err := profile.ListProfileNames(fs.KeyDir)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return profileNames, nil
}

func (fs *FSLocalKeyStore) GetProfile(profileName string) (*profile.Profile, error) {
	// Read in the profile for this proxy.
	profile, err := profile.FromDir(fs.KeyDir, profileName)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return profile, nil
}

// userKeyPath returns the private key path for the given KeyIndex.
func (fs *FSLocalKeyStore) userKeyPath(idx KeyIndex) string {
	return keypaths.UserKeyPath(fs.KeyDir, idx.ProxyHost, idx.Username)
}

// publicKeyPath returns the public key path for the given KeyIndex.
func (fs *FSLocalKeyStore) publicKeyPath(idx KeyIndex) string {
	return keypaths.PublicKeyPath(fs.KeyDir, idx.ProxyHost, idx.Username)
}

// tlsCertPath returns the TLS certificate path given KeyIndex.
func (fs *FSLocalKeyStore) tlsCertPath(idx KeyIndex) string {
	return keypaths.TLSCertPath(fs.KeyDir, idx.ProxyHost, idx.Username)
}

// ppkFilePath returns the PPK (PuTTY-formatted) keypair path for the given KeyIndex.
func (fs *FSLocalKeyStore) ppkFilePath(idx KeyIndex) string {
	return keypaths.PPKFilePath(fs.KeyDir, idx.ProxyHost, idx.Username)
}

// sshDir returns the SSH certificate path for the given KeyIndex.
func (fs *FSLocalKeyStore) sshDir(proxy, user string) string {
	return keypaths.SSHDir(fs.KeyDir, proxy, user)
}

// sshCertPath returns the SSH certificate path for the given KeyIndex.
func (fs *FSLocalKeyStore) sshCertPath(idx KeyIndex) string {
	return keypaths.SSHCertPath(fs.KeyDir, idx.ProxyHost, idx.Username, idx.ClusterName)
}

// appCertPath returns the TLS certificate path for the given KeyIndex and app name.
func (fs *FSLocalKeyStore) appCertPath(idx KeyIndex, appname string) string {
	return keypaths.AppCertPath(fs.KeyDir, idx.ProxyHost, idx.Username, idx.ClusterName, appname)
}

// databaseCertPath returns the TLS certificate path for the given KeyIndex and database name.
func (fs *FSLocalKeyStore) databaseCertPath(idx KeyIndex, dbname string) string {
	return keypaths.DatabaseCertPath(fs.KeyDir, idx.ProxyHost, idx.Username, idx.ClusterName, dbname)
}

// kubeCertPath returns the TLS certificate path for the given KeyIndex and kube cluster name.
func (fs *FSLocalKeyStore) kubeCertPath(idx KeyIndex, kubename string) string {
	return keypaths.KubeCertPath(fs.KeyDir, idx.ProxyHost, idx.Username, idx.ClusterName, kubename)
}

// AddKey adds the given key to the store.
func (fs *FSLocalKeyStore) AddKey(key *Key) error {
	if err := key.KeyIndex.Check(); err != nil {
		return trace.Wrap(err)
	}

	if err := fs.writeBytes(key.PrivateKeyPEM(), fs.userKeyPath(key.KeyIndex)); err != nil {
		return trace.Wrap(err)
	}

	if err := fs.writeBytes(key.MarshalSSHPublicKey(), fs.publicKeyPath(key.KeyIndex)); err != nil {
		return trace.Wrap(err)
	}

	// Store TLS cert
	if err := fs.writeBytes(key.TLSCert, fs.tlsCertPath(key.KeyIndex)); err != nil {
		return trace.Wrap(err)
	}
	if runtime.GOOS == constants.WindowsOS {
		ppkFile, err := key.PPKFile()
		if err == nil {
			if err := fs.writeBytes(ppkFile, fs.ppkFilePath(key.KeyIndex)); err != nil {
				return trace.Wrap(err)
			}
		} else if !trace.IsBadParameter(err) {
			return trace.Wrap(err)
		}
		// PPKFile can only be generated from an RSA private key.
		fs.log.WithError(err).Debugf("Failed to convert private key to PPK-formatted keypair.")
	}

	// Store per-cluster key data.
	if len(key.Cert) > 0 {
		if err := fs.writeBytes(key.Cert, fs.sshCertPath(key.KeyIndex)); err != nil {
			return trace.Wrap(err)
		}
	}

	// TODO(awly): unit test this.
	for kubeCluster, cert := range key.KubeTLSCerts {
		// Prevent directory traversal via a crafted kubernetes cluster name.
		//
		// This will confuse cluster cert loading (GetKey will return
		// kubernetes cluster names different from the ones stored here), but I
		// don't expect any well-meaning user to create bad names.
		kubeCluster = filepath.Clean(kubeCluster)

		path := fs.kubeCertPath(key.KeyIndex, kubeCluster)
		if err := fs.writeBytes(cert, path); err != nil {
			return trace.Wrap(err)
		}
	}
	for db, cert := range key.DBTLSCerts {
		path := fs.databaseCertPath(key.KeyIndex, filepath.Clean(db))
		if err := fs.writeBytes(cert, path); err != nil {
			return trace.Wrap(err)
		}
	}
	for app, cert := range key.AppTLSCerts {
		path := fs.appCertPath(key.KeyIndex, filepath.Clean(app))
		if err := fs.writeBytes(cert, path); err != nil {
			return trace.Wrap(err)
		}
	}

	return nil
}

func (fs *FSLocalKeyStore) writeBytes(bytes []byte, fp string) error {
	if err := os.MkdirAll(filepath.Dir(fp), os.ModeDir|profileDirPerms); err != nil {
		fs.log.Error(err)
		return trace.ConvertSystemError(err)
	}
	err := os.WriteFile(fp, bytes, keyFilePerms)
	if err != nil {
		fs.log.Error(err)
	}
	return trace.ConvertSystemError(err)
}

// DeleteKey deletes the user's key with all its certs.
func (fs *FSLocalKeyStore) DeleteKey(idx KeyIndex) error {
	files := []string{
		fs.userKeyPath(idx),
		fs.publicKeyPath(idx),
		fs.tlsCertPath(idx),
	}
	for _, fn := range files {
		if err := os.Remove(fn); err != nil {
			return trace.ConvertSystemError(err)
		}
	}
	// we also need to delete the extra PuTTY-formatted .ppk file when running on Windows,
	// but it may not exist when upgrading from v9 -> v10 and logging into an existing cluster.
	// as such, deletion should be best-effort and not generate an error if it fails.
	if runtime.GOOS == constants.WindowsOS {
		os.Remove(fs.ppkFilePath(idx))
	}

	// Clear ClusterName to delete the user certs stored for all clusters.
	idx.ClusterName = ""
	return fs.DeleteUserCerts(idx, WithAllCerts...)
}

// DeleteUserCerts deletes only the specified certs of the user's key,
// keeping the private key intact.
// Empty clusterName indicates to delete the certs for all clusters.
//
// Useful when needing to log out of a specific service, like a particular
// database proxy.
func (fs *FSLocalKeyStore) DeleteUserCerts(idx KeyIndex, opts ...CertOption) error {
	for _, o := range opts {
		certPath := o.certPath(fs.KeyDir, idx)
		if err := os.RemoveAll(certPath); err != nil {
			return trace.ConvertSystemError(err)
		}
	}
	return nil
}

// DeleteKeys removes all session keys.
func (fs *FSLocalKeyStore) DeleteKeys() error {
	files, err := os.ReadDir(fs.KeyDir)
	if err != nil {
		return trace.ConvertSystemError(err)
	}
	for _, file := range files {
		if file.IsDir() && file.Name() == tshConfigFileName {
			continue
		}
		if file.IsDir() {
			err := os.RemoveAll(filepath.Join(fs.KeyDir, file.Name()))
			if err != nil {
				return trace.ConvertSystemError(err)
			}
			continue
		}
		err := os.Remove(filepath.Join(fs.KeyDir, file.Name()))
		if err != nil {
			return trace.ConvertSystemError(err)
		}
	}
	return nil
}

// GetKey returns the user's key including the specified certs.
// If the key is not found, returns trace.NotFound error.
func (fs *FSLocalKeyStore) GetKey(idx KeyIndex, opts ...CertOption) (*Key, error) {
	if len(opts) > 0 {
		if err := idx.Check(); err != nil {
			return nil, trace.Wrap(err, "GetKey with CertOptions requires a fully specified KeyIndex")
		}
	}

	if _, err := os.ReadDir(fs.KeyDir); err != nil && trace.IsNotFound(err) {
		return nil, trace.Wrap(err, "no session keys for %+v", idx)
	}

	tlsCertFile := fs.tlsCertPath(idx)
	tlsCert, err := os.ReadFile(tlsCertFile)
	if err != nil {
		fs.log.Error(err)
		return nil, trace.ConvertSystemError(err)
	}
	tlsCA, err := fs.GetTrustedCertsPEM(idx.ProxyHost)
	if err != nil {
		fs.log.Error(err)
		return nil, trace.ConvertSystemError(err)
	}

	priv, err := keys.LoadKeyPair(fs.userKeyPath(idx), fs.publicKeyPath(idx))
	if err != nil {
		fs.log.Error(err)
		return nil, trace.ConvertSystemError(err)
	}

	key := NewKey(priv)
	key.KeyIndex = idx
	key.TLSCert = tlsCert
	key.TrustedCA = []auth.TrustedCerts{{
		TLSCertificates: tlsCA,
	}}

	tlsCertExpiration, err := key.TeleportTLSCertValidBefore()
	if err != nil {
		return nil, trace.Wrap(err)
	}
	fs.log.Debugf("Returning Teleport TLS certificate %q valid until %q.", tlsCertFile, tlsCertExpiration)

	for _, o := range opts {
		if err := fs.updateKeyWithCerts(o, key); err != nil && !trace.IsNotFound(err) {
			fs.log.Error(err)
			return nil, trace.Wrap(err)
		}
	}

	// Note, we may be returning expired certificates here, that is okay. If a
	// certificate is expired, it's the responsibility of the TeleportClient to
	// perform cleanup of the certificates and the profile.

	return key, nil
}

func (fs *FSLocalKeyStore) updateKeyWithCerts(o CertOption, key *Key) error {
	certPath := o.certPath(fs.KeyDir, key.KeyIndex)
	info, err := os.Stat(certPath)
	if err != nil {
		return trace.ConvertSystemError(err)
	}

	fs.log.Debugf("Reading certificates from path %q.", certPath)

	if info.IsDir() {
		certDataMap := map[string][]byte{}
		certFiles, err := os.ReadDir(certPath)
		if err != nil {
			return trace.ConvertSystemError(err)
		}
		for _, certFile := range certFiles {
			name := keypaths.TrimCertPathSuffix(certFile.Name())
			if isCert := name != certFile.Name(); isCert {
				data, err := os.ReadFile(filepath.Join(certPath, certFile.Name()))
				if err != nil {
					return trace.ConvertSystemError(err)
				}
				certDataMap[name] = data
			}
		}
		return o.updateKeyWithMap(key, certDataMap)
	}

	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		return trace.ConvertSystemError(err)
	}
	return o.updateKeyWithBytes(key, certBytes)
}

// GetSSHCertificates gets all certificates signed for the given user and proxy.
func (fs *FSLocalKeyStore) GetSSHCertificates(proxyHost, username string) ([]*ssh.Certificate, error) {
	certDir := fs.sshDir(proxyHost, username)
	certFiles, err := os.ReadDir(certDir)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	sshCerts := make([]*ssh.Certificate, len(certFiles))
	for i, certFile := range certFiles {
		data, err := os.ReadFile(filepath.Join(certDir, certFile.Name()))
		if err != nil {
			return nil, trace.ConvertSystemError(err)
		}

		sshCerts[i], err = apisshutils.ParseCertificate(data)
		if err != nil {
			return nil, trace.Wrap(err)
		}
	}

	return sshCerts, nil
}

// CertOption is an additional step to run when loading/deleting user certificates.
type CertOption interface {
	// certPath returns a path to the cert (or to a dir holding the certs)
	// within the given key dir. For use with FSLocalKeyStore.
	certPath(keyDir string, idx KeyIndex) string
	// updateKeyWithBytes adds the cert bytes to the key and performs related checks.
	updateKeyWithBytes(key *Key, certBytes []byte) error
	// updateKeyWithMap adds the cert data map to the key and performs related checks.
	updateKeyWithMap(key *Key, certMap map[string][]byte) error
	// deleteFromKey deletes the cert data from the key.
	deleteFromKey(key *Key)
}

// WithAllCerts lists all known CertOptions.
var WithAllCerts = []CertOption{WithSSHCerts{}, WithKubeCerts{}, WithDBCerts{}, WithAppCerts{}}

// WithSSHCerts is a CertOption for handling SSH certificates.
type WithSSHCerts struct{}

func (o WithSSHCerts) certPath(keyDir string, idx KeyIndex) string {
	if idx.ClusterName == "" {
		return keypaths.SSHDir(keyDir, idx.ProxyHost, idx.Username)
	}
	return keypaths.SSHCertPath(keyDir, idx.ProxyHost, idx.Username, idx.ClusterName)
}

func (o WithSSHCerts) updateKeyWithBytes(key *Key, certBytes []byte) error {
	key.Cert = certBytes

	// Validate the SSH certificate.
	if err := key.CheckCert(); err != nil {
		if !utils.IsCertExpiredError(err) {
			return trace.Wrap(err)
		}
	}

	return nil
}

func (o WithSSHCerts) updateKeyWithMap(key *Key, certMap map[string][]byte) error {
	return trace.NotImplemented("WithSSHCerts does not implement updateKeyWithMap")
}

func (o WithSSHCerts) deleteFromKey(key *Key) {
	key.Cert = nil
}

// WithKubeCerts is a CertOption for handling kubernetes certificates.
type WithKubeCerts struct{}

func (o WithKubeCerts) certPath(keyDir string, idx KeyIndex) string {
	if idx.ClusterName == "" {
		return keypaths.KubeDir(keyDir, idx.ProxyHost, idx.Username)
	}
	return keypaths.KubeCertDir(keyDir, idx.ProxyHost, idx.Username, idx.ClusterName)
}

func (o WithKubeCerts) updateKeyWithBytes(key *Key, certBytes []byte) error {
	return trace.NotImplemented("WithKubeCerts does not implement updateKeyWithBytes")
}

func (o WithKubeCerts) updateKeyWithMap(key *Key, certMap map[string][]byte) error {
	key.KubeTLSCerts = certMap
	return nil
}

func (o WithKubeCerts) deleteFromKey(key *Key) {
	key.KubeTLSCerts = nil
}

// WithDBCerts is a CertOption for handling database access certificates.
type WithDBCerts struct {
	dbName string
}

func (o WithDBCerts) certPath(keyDir string, idx KeyIndex) string {
	if idx.ClusterName == "" {
		return keypaths.DatabaseDir(keyDir, idx.ProxyHost, idx.Username)
	}
	if o.dbName == "" {
		return keypaths.DatabaseCertDir(keyDir, idx.ProxyHost, idx.Username, idx.ClusterName)
	}
	return keypaths.DatabaseCertPath(keyDir, idx.ProxyHost, idx.Username, idx.ClusterName, o.dbName)
}

func (o WithDBCerts) updateKeyWithBytes(key *Key, certBytes []byte) error {
	return trace.NotImplemented("WithDBCerts does not implement updateKeyWithBytes")
}

func (o WithDBCerts) updateKeyWithMap(key *Key, certMap map[string][]byte) error {
	key.DBTLSCerts = certMap
	return nil
}

func (o WithDBCerts) deleteFromKey(key *Key) {
	key.DBTLSCerts = nil
}

// WithAppCerts is a CertOption for handling application access certificates.
type WithAppCerts struct {
	appName string
}

func (o WithAppCerts) certPath(keyDir string, idx KeyIndex) string {
	if idx.ClusterName == "" {
		return keypaths.AppDir(keyDir, idx.ProxyHost, idx.Username)
	}
	if o.appName == "" {
		return keypaths.AppCertDir(keyDir, idx.ProxyHost, idx.Username, idx.ClusterName)
	}
	return keypaths.AppCertPath(keyDir, idx.ProxyHost, idx.Username, idx.ClusterName, o.appName)
}

func (o WithAppCerts) updateKeyWithBytes(key *Key, certBytes []byte) error {
	return trace.NotImplemented("WithAppCerts does not implement updateKeyWithBytes")
}

func (o WithAppCerts) updateKeyWithMap(key *Key, certMap map[string][]byte) error {
	key.AppTLSCerts = certMap
	return nil
}

func (o WithAppCerts) deleteFromKey(key *Key) {
	key.AppTLSCerts = nil
}

// fsLocalNonSessionKeyStore is a FS-based store implementing methods
// for CA certificates and known host fingerprints. It is embedded
// in both FSLocalKeyStore and MemLocalKeyStore.
type fsLocalNonSessionKeyStore struct {
	// log holds the structured logger.
	log logrus.FieldLogger

	// KeyDir is the directory where all keys are stored.
	KeyDir string
}

// proxyKeyDir returns the keystore's keys directory for the given proxy.
func (fs *fsLocalNonSessionKeyStore) proxyKeyDir(proxy string) string {
	return keypaths.ProxyKeyDir(fs.KeyDir, proxy)
}

// casDir returns path to trusted clusters certificates directory.
func (fs *fsLocalNonSessionKeyStore) casDir(proxy string) string {
	return keypaths.CAsDir(fs.KeyDir, proxy)
}

// clusterCAPath returns path to cluster certificate.
func (fs *fsLocalNonSessionKeyStore) clusterCAPath(proxy, clusterName string) string {
	return keypaths.TLSCAsPathCluster(fs.KeyDir, proxy, clusterName)
}

// knownHostsPath returns the keystore's known hosts file path.
func (fs *fsLocalNonSessionKeyStore) knownHostsPath() string {
	return keypaths.KnownHostsPath(fs.KeyDir)
}

// tlsCAsPath returns the TLS CA certificates path for the given KeyIndex.
func (fs *fsLocalNonSessionKeyStore) tlsCAsPath(proxy string) string {
	return keypaths.TLSCAsPath(fs.KeyDir, proxy)
}

// AddKnownHostKeys adds a new entry to `known_hosts` file.
func (fs *fsLocalNonSessionKeyStore) AddKnownHostKeys(hostname, proxyHost string, hostKeys []ssh.PublicKey) (retErr error) {
	// We're trying to serialize our writes to the 'known_hosts' file to avoid corruption, since there
	// are cases when multiple tsh instances will try to write to it.
	unlock, err := utils.FSTryWriteLockTimeout(context.Background(), fs.knownHostsPath(), 5*time.Second)
	if err != nil {
		return trace.WrapWithMessage(err, "could not acquire lock for the `known_hosts` file")
	}
	defer utils.StoreErrorOf(unlock, &retErr)

	fp, err := os.OpenFile(fs.knownHostsPath(), os.O_CREATE|os.O_RDWR, 0640)
	if err != nil {
		return trace.ConvertSystemError(err)
	}
	defer utils.StoreErrorOf(fp.Close, &retErr)
	// read all existing entries into a map (this removes any pre-existing dupes)
	entries := make(map[string]int)
	output := make([]string, 0)
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		line := scanner.Text()
		if _, exists := entries[line]; !exists {
			output = append(output, line)
			entries[line] = 1
		}
	}
	// check if the scanner ran into an error
	if err := scanner.Err(); err != nil {
		return trace.Wrap(err)
	}
	// add every host key to the list of entries
	for i := range hostKeys {
		fs.log.Debugf("Adding known host %s with proxy %s and key: %v", hostname, proxyHost, sshutils.Fingerprint(hostKeys[i]))
		bytes := ssh.MarshalAuthorizedKey(hostKeys[i])

		// Write keys in an OpenSSH-compatible format. A previous format was not
		// quite OpenSSH-compatible, so we may write a duplicate entry here. Any
		// duplicates will be pruned below.
		// We include both the proxy server and original hostname as well as the
		// root domain wildcard. OpenSSH clients match against both the proxy
		// host and nodes (via the wildcard). Teleport itself occasionally uses
		// the root cluster name.
		line := fmt.Sprintf(
			"@cert-authority %s,%s,*.%s %s type=host",
			proxyHost, hostname, hostname, strings.TrimSpace(string(bytes)),
		)
		if _, exists := entries[line]; !exists {
			output = append(output, line)
		}
	}
	// Prune any duplicate host entries for migrated hosts. Note that only
	// duplicates matching the current hostname/proxyHost will be pruned; others
	// will be cleaned up at subsequent logins.
	output = pruneOldHostKeys(output)
	// re-create the file:
	_, err = fp.Seek(0, 0)
	if err != nil {
		return trace.Wrap(err)
	}
	if err = fp.Truncate(0); err != nil {
		return trace.Wrap(err)
	}
	for _, line := range output {
		fmt.Fprintf(fp, "%s\n", line)
	}
	return fp.Sync()
}

// GetKnownHostsFile returns a known hosts file.
func (fs *fsLocalNonSessionKeyStore) GetKnownHostsFile() (knownHosts []byte, retErr error) {
	var err error
	unlock, err := utils.FSTryReadLockTimeout(context.Background(), fs.knownHostsPath(), 5*time.Second)
	if err != nil {
		return nil, trace.WrapWithMessage(err, "could not acquire lock for the `known_hosts` file")
	}
	defer utils.StoreErrorOf(unlock, &retErr)

	bytes, err := os.ReadFile(fs.knownHostsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, trace.Wrap(err)
	}
	return bytes, nil
}

// GetKnownHostKeys returns all known public host keys for the given host name.
func (fs *fsLocalNonSessionKeyStore) GetKnownHostKeys(hostname string) ([]ssh.PublicKey, error) {
	knownHosts, err := fs.GetKnownHostsFile()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return apisshutils.ParseKnownHosts([][]byte{knownHosts}, hostname)
}

// SaveTrustedCerts saves trusted TLS certificates of certificate authorities.
func (fs *fsLocalNonSessionKeyStore) SaveTrustedCerts(proxyHost string, cas []auth.TrustedCerts) (retErr error) {
	if err := os.MkdirAll(fs.proxyKeyDir(proxyHost), os.ModeDir|profileDirPerms); err != nil {
		fs.log.Error(err)
		return trace.ConvertSystemError(err)
	}

	// Save trusted clusters certs in CAS directory.
	if err := fs.saveTrustedCertsInCASDir(proxyHost, cas); err != nil {
		return trace.Wrap(err)
	}

	// For backward compatibility save trusted in legacy certs.pem file.
	if err := fs.saveTrustedCertsInLegacyCAFile(proxyHost, cas); err != nil {
		return trace.Wrap(err)
	}

	return nil
}

func (fs *fsLocalNonSessionKeyStore) saveTrustedCertsInCASDir(proxyHost string, cas []auth.TrustedCerts) error {
	casDirPath := filepath.Join(fs.casDir(proxyHost))
	if err := os.MkdirAll(casDirPath, os.ModeDir|profileDirPerms); err != nil {
		fs.log.Error(err)
		return trace.ConvertSystemError(err)
	}

	for _, ca := range cas {
		if !isSafeClusterName(ca.ClusterName) {
			fs.log.Warnf("Skipped unsafe cluster name: %q", ca.ClusterName)
			continue
		}
		// Create CA files in cas dir for each cluster.
		caFile, err := os.OpenFile(fs.clusterCAPath(proxyHost, ca.ClusterName), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0640)
		if err != nil {
			return trace.ConvertSystemError(err)
		}

		if err := writeClusterCertificates(caFile, ca.TLSCertificates); err != nil {
			return trace.Wrap(err)
		}
	}
	return nil
}

func (fs *fsLocalNonSessionKeyStore) saveTrustedCertsInLegacyCAFile(proxyHost string, cas []auth.TrustedCerts) (retErr error) {
	certsFile := fs.tlsCAsPath(proxyHost)
	fp, err := os.OpenFile(certsFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0640)
	if err != nil {
		return trace.ConvertSystemError(err)
	}
	defer utils.StoreErrorOf(fp.Close, &retErr)
	for _, ca := range cas {
		for _, cert := range ca.TLSCertificates {
			if _, err := fp.Write(cert); err != nil {
				return trace.ConvertSystemError(err)
			}
			if _, err := fmt.Fprintln(fp); err != nil {
				return trace.ConvertSystemError(err)
			}
		}
	}
	return fp.Sync()
}

// isSafeClusterName check if cluster name is safe and doesn't contain miscellaneous characters.
func isSafeClusterName(name string) bool {
	return !strings.Contains(name, "..")
}

func writeClusterCertificates(f *os.File, tlsCertificates [][]byte) error {
	defer f.Close()
	for _, cert := range tlsCertificates {
		if _, err := f.Write(cert); err != nil {
			return trace.ConvertSystemError(err)
		}
	}
	if err := f.Sync(); err != nil {
		return trace.ConvertSystemError(err)
	}
	return nil
}

// GetTrustedCertsPEM returns trusted TLS certificates of certificate authorities PEM
// blocks.
func (fs *fsLocalNonSessionKeyStore) GetTrustedCertsPEM(proxyHost string) ([][]byte, error) {
	dir := fs.casDir(proxyHost)

	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return nil, trace.NotFound("please relogin, tsh user profile doesn't contain CAS directory: %s", dir)
		}
		return nil, trace.ConvertSystemError(err)
	}

	var blocks [][]byte
	err := filepath.Walk(dir, func(path string, info osfs.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		for len(data) > 0 {
			if err != nil {
				return trace.Wrap(err)
			}
			block, rest := pem.Decode(data)
			if block == nil {
				break
			}
			if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
				fs.log.Debugf("Skipping PEM block type=%v headers=%v.", block.Type, block.Headers)
				data = rest
				continue
			}
			// rest contains the remainder of data after reading a block.
			// Therefore, the block length is len(data) - len(rest).
			// Use that length to slice the block from the start of data.
			blocks = append(blocks, data[:len(data)-len(rest)])
			data = rest
		}
		return nil
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return blocks, nil
}

// noLocalKeyStore is a LocalKeyStore representing the absence of a keystore.
// All methods return errors. This exists to avoid nil checking everywhere in
// LocalKeyAgent and prevent nil pointer panics.
type noLocalKeyStore struct{}

var errNoLocalKeyStore = trace.NotFound("there is no local keystore")

func (noLocalKeyStore) AddKey(key *Key) error {
	return errNoLocalKeyStore
}
func (noLocalKeyStore) GetKey(idx KeyIndex, opts ...CertOption) (*Key, error) {
	return nil, errNoLocalKeyStore
}
func (noLocalKeyStore) DeleteKey(idx KeyIndex) error {
	return errNoLocalKeyStore
}
func (noLocalKeyStore) DeleteUserCerts(idx KeyIndex, opts ...CertOption) error {
	return errNoLocalKeyStore
}
func (noLocalKeyStore) DeleteKeys() error { return errNoLocalKeyStore }
func (noLocalKeyStore) AddKnownHostKeys(hostname, proxyHost string, keys []ssh.PublicKey) error {
	return errNoLocalKeyStore
}
func (noLocalKeyStore) GetKnownHostsFile() ([]byte, error) {
	return nil, errNoLocalKeyStore
}
func (noLocalKeyStore) GetKnownHostKeys(hostname string) ([]ssh.PublicKey, error) {
	return nil, errNoLocalKeyStore
}
func (noLocalKeyStore) SaveTrustedCerts(proxyHost string, cas []auth.TrustedCerts) error {
	return errNoLocalKeyStore
}
func (noLocalKeyStore) GetTrustedCertsPEM(proxyHost string) ([][]byte, error) {
	return nil, errNoLocalKeyStore
}
func (noLocalKeyStore) GetSSHCertificates(proxyHost, username string) ([]*ssh.Certificate, error) {
	return nil, errNoLocalKeyStore
}
func (noLocalKeyStore) CurrentProfile() (string, error) {
	return "", errNoLocalKeyStore
}
func (noLocalKeyStore) ListProfiles() ([]string, error) {
	return nil, errNoLocalKeyStore
}
func (noLocalKeyStore) GetProfile(profileName string) (*profile.Profile, error) {
	return nil, errNoLocalKeyStore
}

// MemLocalKeyStore is an in-memory session keystore implementation.
type MemLocalKeyStore struct {
	log *logrus.Entry
	NonSessionKeyStore
	keyCache memLocalKeyStoreMap
}

// memLocalKeyStoreMap is a three-dimensional map indexed by [proxyHost][username][clusterName]
type memLocalKeyStoreMap = map[string]map[string]map[string]*Key

// NewMemLocalKeyStore initializes a MemLocalKeyStore.
// The key directory here is only used for storing CA certificates and known
// host fingerprints.
func NewMemLocalKeyStore(s NonSessionKeyStore) KeyStore {
	return &MemLocalKeyStore{
		log:                logrus.WithField(trace.Component, teleport.ComponentKeyStore),
		NonSessionKeyStore: s,
		keyCache:           memLocalKeyStoreMap{},
	}
}

// AddKey writes a key to the underlying key store.
func (s *MemLocalKeyStore) AddKey(key *Key) error {
	if err := key.KeyIndex.Check(); err != nil {
		return trace.Wrap(err)
	}
	_, ok := s.keyCache[key.ProxyHost]
	if !ok {
		s.keyCache[key.ProxyHost] = map[string]map[string]*Key{}
	}
	_, ok = s.keyCache[key.ProxyHost][key.Username]
	if !ok {
		s.keyCache[key.ProxyHost][key.Username] = map[string]*Key{}
	}
	s.keyCache[key.ProxyHost][key.Username][key.ClusterName] = key
	return nil
}

// GetKey returns the user's key including the specified certs.
func (s *MemLocalKeyStore) GetKey(idx KeyIndex, opts ...CertOption) (*Key, error) {
	var key *Key
	if idx.ClusterName == "" {
		// If clusterName is not specified then the cluster-dependent fields
		// are not considered relevant and we may simply return any key
		// associated with any cluster name whatsoever.
		for _, found := range s.keyCache[idx.ProxyHost][idx.Username] {
			key = found
			break
		}
	} else {
		key = s.keyCache[idx.ProxyHost][idx.Username][idx.ClusterName]
	}
	if key == nil {
		return nil, trace.NotFound("key for %+v not found", idx)
	}

	// It is not necessary to handle opts because all the optional certs are
	// already part of the Key struct as stored in memory.

	tlsCertExpiration, err := key.TeleportTLSCertValidBefore()
	if err != nil {
		return nil, trace.Wrap(err)
	}
	s.log.Debugf("Returning Teleport TLS certificate from memory, valid until %q.", tlsCertExpiration)

	// Validate the SSH certificate.
	if err := key.CheckCert(); err != nil {
		if !utils.IsCertExpiredError(err) {
			return nil, trace.Wrap(err)
		}
	}

	return key, nil
}

// DeleteKey deletes the user's key with all its certs.
func (s *MemLocalKeyStore) DeleteKey(idx KeyIndex) error {
	delete(s.keyCache[idx.ProxyHost], idx.Username)
	return nil
}

// DeleteKeys removes all session keys.
func (s *MemLocalKeyStore) DeleteKeys() error {
	s.keyCache = memLocalKeyStoreMap{}
	return nil
}

// DeleteUserCerts deletes only the specified certs of the user's key,
// keeping the private key intact.
// Empty clusterName indicates to delete the certs for all clusters.
//
// Useful when needing to log out of a specific service, like a particular
// database proxy.
func (s *MemLocalKeyStore) DeleteUserCerts(idx KeyIndex, opts ...CertOption) error {
	var keys []*Key
	if idx.ClusterName != "" {
		key, ok := s.keyCache[idx.ProxyHost][idx.Username][idx.ClusterName]
		if !ok {
			return nil
		}
		keys = []*Key{key}
	} else {
		keys = make([]*Key, 0, len(s.keyCache[idx.ProxyHost][idx.Username]))
		for _, key := range s.keyCache[idx.ProxyHost][idx.Username] {
			keys = append(keys, key)
		}
	}

	for _, key := range keys {
		for _, o := range opts {
			o.deleteFromKey(key)
		}
	}
	return nil
}

// GetSSHCertificates gets all certificates signed for the given user and proxy.
func (s *MemLocalKeyStore) GetSSHCertificates(proxyHost, username string) ([]*ssh.Certificate, error) {
	var sshCerts []*ssh.Certificate
	for _, key := range s.keyCache[proxyHost][username] {
		sshCert, err := key.SSHCert()
		if err != nil {
			return nil, trace.Wrap(err)
		}
		sshCerts = append(sshCerts, sshCert)
	}

	return sshCerts, nil
}

type memLocalNonSessionKeyStore struct {
	profile    *profile.Profile
	KnownHosts []byte
	TrustedCAs []auth.TrustedCerts
}

// CurrentProfile returns the current active profile.
func (ms *memLocalNonSessionKeyStore) CurrentProfile() (string, error) {
	return ms.profile.Name(), nil
}

// ListProfiles returns a list of all active profiles.
func (ms *memLocalNonSessionKeyStore) ListProfiles() ([]string, error) {
	return []string{ms.profile.Name()}, nil
}

// GetProfile returns the requested profile.
func (ms *memLocalNonSessionKeyStore) GetProfile(profileName string) (*profile.Profile, error) {
	return ms.profile, nil
}

// GetKnownHostsFile returns a known hosts file.
func (ms *memLocalNonSessionKeyStore) GetKnownHostsFile() ([]byte, error) {
	return ms.KnownHosts, nil
}

// AddKnownHostKeys adds a new entry to `known_hosts` file.
func (ms *memLocalNonSessionKeyStore) AddKnownHostKeys(hostname, proxyHost string, hostKeys []ssh.PublicKey) error {
	return nil
}

// GetKnownHostKeys returns all known public host keys for the given host name.
func (ms *memLocalNonSessionKeyStore) GetKnownHostKeys(hostname string) ([]ssh.PublicKey, error) {
	// var sshHostKeys []ssh.PublicKey
	// for _, ca := range ms.TrustedCAs {
	// 	hostKeys, err := ca.SSHCertPublicKeys()
	// 	if err != nil {
	// 		return nil, trace.Wrap(err)
	// 	}
	// 	sshHostKeys = append(sshHostKeys, hostKeys...)
	// }

	knownHosts, err := ms.GetKnownHostsFile()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return apisshutils.ParseKnownHosts([][]byte{knownHosts}, hostname)
}

// SaveTrustedCerts saves trusted TLS certificates of certificate authorities.
func (ms *memLocalNonSessionKeyStore) SaveTrustedCerts(_ string, cas []auth.TrustedCerts) error {
	ms.TrustedCAs = append(ms.TrustedCAs, cas...)
	return nil
}

// GetTrustedCertsPEM gets trusted TLS certificates of certificate authorities.
// Each returned byte slice contains an individual PEM block.
func (ms *memLocalNonSessionKeyStore) GetTrustedCertsPEM(proxyHost string) ([][]byte, error) {
	var tlsHostCerts [][]byte
	for _, ca := range ms.TrustedCAs {
		tlsHostCerts = append(tlsHostCerts, ca.TLSCertificates...)
	}
	return tlsHostCerts, nil
}

func NewMemLocalKeyStoreFromAgent(sshAuthSocket string) (KeyStore, error) {
	conn, err := agentconn.Dial(sshAuthSocket)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	defer conn.Close()

	agentClient := agent.NewClient(conn)
	keyExtensionResp, err := KeyExtension(agentClient)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	var profile profile.Profile
	if err := json.Unmarshal(keyExtensionResp.ProfileBlob, &profile); err != nil {
		return nil, trace.Wrap(err)
	}

	var forwardedKey forwardedKey
	if err := json.Unmarshal(keyExtensionResp.ForwardedKeyBlob, &forwardedKey); err != nil {
		return nil, trace.Wrap(err)
	}

	sshCert, err := apisshutils.ParseCertificate(forwardedKey.SSHCertificate)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	priv, err := keys.NewPrivateKey(newAgentSigner(sshAuthSocket, sshCert), nil)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	keyStore := NewMemLocalKeyStore(&memLocalNonSessionKeyStore{
		profile:    &profile,
		KnownHosts: keyExtensionResp.KnownHosts,
		TrustedCAs: forwardedKey.TrustedCerts,
	})

	key := NewKey(priv)
	key.KeyIndex = forwardedKey.KeyIndex
	key.Cert = forwardedKey.SSHCertificate
	key.TLSCert = forwardedKey.TLSCertificate
	key.TrustedCA = forwardedKey.TrustedCerts

	// Preload the client key from the agent.
	if err := keyStore.AddKey(key); err != nil {
		return nil, trace.Wrap(err)
	}

	return keyStore, nil
}

type agentSigner struct {
	sshAuthSock string
	sshCert     *ssh.Certificate
}

func newAgentSigner(sshAuthSock string, sshCert *ssh.Certificate) crypto.Signer {
	return &agentSigner{
		sshAuthSock: sshAuthSock,
		sshCert:     sshCert,
	}
}

func (as *agentSigner) Public() crypto.PublicKey {
	if pubGetter, ok := as.sshCert.Key.(ssh.CryptoPublicKey); ok {
		return pubGetter.CryptoPublicKey()
	}
	return nil
}

func (as *agentSigner) Sign(_ io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	conn, err := agentconn.Dial(as.sshAuthSock)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	defer conn.Close()

	agentClient := agent.NewClient(conn)
	return SignExtension(agentClient, as.sshCert, digest, opts)
}

func NewMemLocalKeyStoreFromIdentityFile(identityFile string, proxyAddr string, clusterName string) (KeyStore, error) {
	key, err := KeyFromIdentityFile(identityFile)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	proxyHost, err := utils.Host(proxyAddr)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	if clusterName == "" {
		clusterName, err = key.RootClusterName()
		if err != nil {
			return nil, trace.Wrap(err)
		}
	}

	certUsername, err := key.CertUsername()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	// Configure missing KeyIndex fields.
	key.KeyIndex = KeyIndex{
		ProxyHost:   proxyHost,
		ClusterName: clusterName,
		Username:    certUsername,
	}

	keyStore := NewMemLocalKeyStore(&memLocalNonSessionKeyStore{
		TrustedCAs: key.TrustedCA,
		profile: &profile.Profile{
			WebProxyAddr: proxyAddr,
			SiteName:     clusterName,
			Username:     certUsername,
		},
	})

	// Preload the client key from the agent.
	if err := keyStore.AddKey(key); err != nil {
		return nil, trace.Wrap(err)
	}

	return keyStore, nil
}
