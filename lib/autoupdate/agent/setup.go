/*
 * Teleport
 * Copyright (C) 2024  Gravitational, Inc.
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

package agent

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/google/renameio/v2"
	"github.com/gravitational/trace"
	"gopkg.in/yaml.v3"

	"github.com/gravitational/teleport/lib/defaults"
	libdefaults "github.com/gravitational/teleport/lib/defaults"
	libutils "github.com/gravitational/teleport/lib/utils"
)

// Base paths for constructing namespaced directories.
const (
	defaultInstallDir  = "/opt/teleport"
	defaultPathDir     = "/usr/local/bin"
	systemdAdminDir    = "/etc/systemd/system"
	systemdPIDDir      = "/run"
	needrestartConfDir = "/etc/needrestart/conf.d"
	versionsDirName    = "versions"
	lockFileName       = "update.lock"
	defaultNamespace   = "default"
	systemNamespace    = "system"
)

const (
	// deprecatedTimerName is the timer for the deprecated upgrader should be disabled on setup.
	deprecatedTimerName = "teleport-upgrade.timer"
)

const (
	updateServiceTemplate = `# teleport-update
# DO NOT EDIT THIS FILE
[Unit]
Description=Teleport auto-update service

[Service]
Type=oneshot
ExecStart={{.UpdaterBinary}} --install-suffix={{.InstallSuffix}} --install-dir="{{escape .InstallDir}}" update
`
	updateTimerTemplate = `# teleport-update
# DO NOT EDIT THIS FILE
[Unit]
Description=Teleport auto-update timer unit

[Timer]
OnActiveSec=1m
OnUnitActiveSec=5m
RandomizedDelaySec=1m

[Install]
WantedBy={{.TeleportService}}
`
	teleportDropInTemplate = `# teleport-update
# DO NOT EDIT THIS FILE
[Service]
Environment="TELEPORT_UPDATE_CONFIG_FILE={{escape .UpdaterConfigFile}}"
Environment="TELEPORT_UPDATE_INSTALL_DIR={{escape .InstallDir}}"
`
	// This configuration sets the default value for needrestart-trigger automatic restarts for teleport.service to disabled.
	// Users may still choose to enable needrestart for teleport.service when installing packaging interactively (or via dpkg config),
	// but doing so will result in a hard restart that disconnects the agent whenever any dependent libraries are updated.
	// Other network services, like openvpn, follow this pattern.
	// It is possible to configure needrestart to trigger a soft restart (via restart.d script), but given that Teleport subprocesses
	// can use a wide variety of installed binaries (when executed by the user), this could trigger many unexpected reloads.
	needrestartConfTemplate = `$nrconf{override_rc}{qr(^{{replace .TeleportService "." "\\."}})} = 0;
`
)

type confParams struct {
	TeleportService   string
	UpdaterBinary     string
	InstallSuffix     string
	InstallDir        string
	Path              string
	UpdaterConfigFile string
}

// Namespace represents a namespace within various system paths for a isolated installation of Teleport.
type Namespace struct {
	log *slog.Logger
	// name of namespace
	name string
	// installDir for Teleport namespaces (/opt/teleport)
	installDir string
	// defaultPathDir for Teleport binaries (ns: /opt/teleport/myns/bin)
	defaultPathDir string
	// dataDir parsed from teleport.yaml, if present
	dataDir string
	// defaultProxyAddr parsed from teleport.yaml, if present
	defaultProxyAddr string
	// serviceFile for the Teleport systemd service (ns: /etc/systemd/system/teleport_myns.service)
	serviceFile string
	// configFile for Teleport config (ns: /etc/teleport_myns.yaml)
	configFile string
	// pidFile for Teleport (ns: /run/teleport_myns.pid)
	pidFile string
	// updaterServiceFile is the systemd service path for the updater
	updaterServiceFile string
	// updaterTimerFile is the systemd timer path for the updater
	updaterTimerFile string
	// dropInFile is the Teleport systemd drop-in path extending Teleport
	dropInFile string
	// needrestartConfFile is the path to needrestart configuration for Teleport
	needrestartConfFile string
}

var alphanum = regexp.MustCompile("^[a-zA-Z0-9-]*$")

// NewNamespace validates and returns a Namespace.
// Namespaces must be alphanumeric + `-`.
// defaultPathDir overrides the destination directory for namespace setup (i.e., /usr/local)
func NewNamespace(ctx context.Context, log *slog.Logger, name, installDir string) (ns *Namespace, err error) {
	defer ns.overrideFromConfig(ctx)

	if name == defaultNamespace ||
		name == systemNamespace {
		return nil, trace.Errorf("namespace %s is reserved", name)
	}
	if !alphanum.MatchString(name) {
		return nil, trace.Errorf("invalid namespace name %s, must be alphanumeric", name)
	}
	if installDir == "" {
		installDir = defaultInstallDir
	}
	if name == "" {
		linkDir := defaultPathDir
		return &Namespace{
			log:                 log,
			name:                name,
			installDir:          installDir,
			defaultPathDir:      linkDir,
			dataDir:             defaults.DataDir,
			serviceFile:         filepath.Join("/", serviceDir, serviceName),
			configFile:          defaults.ConfigFilePath,
			pidFile:             filepath.Join(systemdPIDDir, "teleport.pid"),
			updaterServiceFile:  filepath.Join(systemdAdminDir, BinaryName+".service"),
			updaterTimerFile:    filepath.Join(systemdAdminDir, BinaryName+".timer"),
			dropInFile:          filepath.Join(systemdAdminDir, "teleport.service.d", BinaryName+".conf"),
			needrestartConfFile: filepath.Join(needrestartConfDir, BinaryName+".conf"),
		}, nil
	}

	prefix := "teleport_" + name
	linkDir := filepath.Join(installDir, name, "bin")
	return &Namespace{
		log:                 log,
		name:                name,
		installDir:          installDir,
		defaultPathDir:      linkDir,
		dataDir:             filepath.Join(filepath.Dir(defaults.DataDir), prefix),
		serviceFile:         filepath.Join(systemdAdminDir, prefix+".service"),
		configFile:          filepath.Join(filepath.Dir(defaults.ConfigFilePath), prefix+".yaml"),
		pidFile:             filepath.Join(systemdPIDDir, prefix+".pid"),
		updaterServiceFile:  filepath.Join(systemdAdminDir, BinaryName+"_"+name+".service"),
		updaterTimerFile:    filepath.Join(systemdAdminDir, BinaryName+"_"+name+".timer"),
		dropInFile:          filepath.Join(systemdAdminDir, prefix+".service.d", BinaryName+"_"+name+".conf"),
		needrestartConfFile: filepath.Join(needrestartConfDir, BinaryName+"_"+name+".conf"),
	}, nil
}

func (ns *Namespace) Dir() string {
	name := ns.name
	if name == "" {
		name = defaultNamespace
	}
	return filepath.Join(ns.installDir, name)
}

// Init create the initial directory structure and returns the lockfile for a Namespace.
func (ns *Namespace) Init() (lockFile string, err error) {
	if err := os.MkdirAll(filepath.Join(ns.Dir(), versionsDirName), systemDirMode); err != nil {
		return "", trace.Wrap(err)
	}
	return filepath.Join(ns.Dir(), lockFileName), nil
}

// Setup installs service and timer files for the teleport-update binary.
// Afterwords, Setup reloads systemd and enables the timer with --now.
func (ns *Namespace) Setup(ctx context.Context, path string) error {
	if ok, err := hasSystemD(); err == nil && !ok {
		ns.log.WarnContext(ctx, "Systemd is not running, skipping updater installation.")
		return nil
	}

	err := ns.writeConfigFiles(ctx, path)
	if err != nil {
		return trace.Wrap(err, "failed to write teleport-update systemd config files")
	}
	timer := &SystemdService{
		ServiceName: filepath.Base(ns.updaterTimerFile),
		Log:         ns.log,
	}
	if err := timer.Sync(ctx); err != nil {
		return trace.Wrap(err, "failed to sync systemd config")
	}
	if err := timer.Enable(ctx, true); err != nil {
		return trace.Wrap(err, "failed to enable teleport-update systemd timer")
	}
	if ns.name == "" {
		oldTimer := &SystemdService{
			ServiceName: deprecatedTimerName,
			Log:         ns.log,
		}
		// If the old teleport-upgrade script is detected, disable it to ensure they do not interfere.
		// Note that the schedule is also set to nop by the Teleport agent -- this just prevents restarts.
		enabled, err := isActiveOrEnabled(ctx, oldTimer)
		if err != nil {
			return trace.Wrap(err, "failed to determine if deprecated teleport-upgrade systemd timer is enabled")
		}
		if enabled {
			if err := oldTimer.Disable(ctx, true); err != nil {
				ns.log.ErrorContext(ctx, "The deprecated teleport-ent-updater package is installed on this server, and it cannot be disabled due to an error. You must remove the teleport-ent-updater package after verifying that teleport-update is working.", errorKey, err)
			} else {
				ns.log.WarnContext(ctx, "The deprecated teleport-ent-updater package is installed on this server. This package has been disabled to prevent conflicts. Please remove the teleport-ent-updater package after verifying that teleport-update is working.")
			}
		}
	}
	return nil
}

// Teardown removes all traces of the auto-updater, including its configuration.
func (ns *Namespace) Teardown(ctx context.Context) error {
	if ok, err := hasSystemD(); err == nil && !ok {
		ns.log.WarnContext(ctx, "Systemd is not running, skipping updater removal.")
		if err := os.RemoveAll(ns.Dir()); err != nil {
			return trace.Wrap(err, "failed to remove versions directory")
		}
		return nil
	}

	svc := &SystemdService{
		ServiceName: filepath.Base(ns.updaterTimerFile),
		Log:         ns.log,
	}
	if err := svc.Disable(ctx, true); err != nil {
		ns.log.WarnContext(ctx, "Unable to disable teleport-update systemd timer before removing.", errorKey, err)
	}
	for _, p := range []string{
		ns.updaterServiceFile,
		ns.updaterTimerFile,
		ns.dropInFile,
		ns.needrestartConfFile,
	} {
		if err := os.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return trace.Wrap(err, "failed to remove %s", filepath.Base(p))
		}
	}
	if err := svc.Sync(ctx); err != nil {
		return trace.Wrap(err, "failed to sync systemd config")
	}
	if err := os.RemoveAll(ns.Dir()); err != nil {
		return trace.Wrap(err, "failed to remove versions directory")
	}
	if ns.name == "" {
		oldTimer := &SystemdService{
			ServiceName: deprecatedTimerName,
			Log:         ns.log,
		}
		// If the old upgrader exists, attempt to re-enable it automatically
		present, err := oldTimer.IsPresent(ctx)
		if err != nil {
			return trace.Wrap(err, "failed to determine if deprecated teleport-upgrade systemd timer is present")
		}
		if present {
			if err := oldTimer.Enable(ctx, true); err != nil {
				ns.log.ErrorContext(ctx, "The deprecated teleport-ent-updater package is installed on this server, and it cannot be re-enabled due to an error. Please fix the teleport-ent-updater package if you intend to use the deprecated updater.", errorKey, err)
			} else {
				ns.log.WarnContext(ctx, "The deprecated teleport-ent-updater package is installed on this server. This package has been re-enabled to ensure continued updates. To disable automatic updates entirely, please remove the teleport-ent-updater package.")
			}
		}
	}
	return nil
}

func (ns *Namespace) writeConfigFiles(ctx context.Context, path string) error {
	teleportService := filepath.Base(ns.serviceFile)
	params := confParams{
		TeleportService:   teleportService,
		UpdaterBinary:     filepath.Join(path, BinaryName),
		InstallSuffix:     ns.name,
		InstallDir:        ns.installDir,
		Path:              path,
		UpdaterConfigFile: filepath.Join(ns.Dir(), updateConfigName),
	}
	err := writeSystemTemplate(ns.updaterServiceFile, updateServiceTemplate, params)
	if err != nil {
		return trace.Wrap(err)
	}
	err = writeSystemTemplate(ns.updaterTimerFile, updateTimerTemplate, params)
	if err != nil {
		return trace.Wrap(err)
	}
	err = writeSystemTemplate(ns.dropInFile, teleportDropInTemplate, params)
	if err != nil {
		return trace.Wrap(err)
	}
	// Needrestart config is non-critical for updater functionality.
	_, err = os.Stat(filepath.Dir(ns.needrestartConfFile))
	if os.IsNotExist(err) {
		return nil // needrestart is not present
	}
	if err != nil {
		ns.log.ErrorContext(ctx, "Unable to disable needrestart.", errorKey, err)
		return nil
	}
	ns.log.InfoContext(ctx, "Disabling needrestart.", unitKey, teleportService)
	err = writeSystemTemplate(ns.needrestartConfFile, needrestartConfTemplate, params)
	if err != nil {
		ns.log.ErrorContext(ctx, "Unable to disable needrestart.", errorKey, err)
		return nil
	}
	return nil
}

// writeSystemTemplate atomically writes a template to a system file, creating any needed directories.
// Temporarily files are stored in the target path to ensure the file has needed SELinux contexts.
func writeSystemTemplate(path, t string, values any) error {
	dir, file := filepath.Split(path)
	if err := os.MkdirAll(dir, systemDirMode); err != nil {
		return trace.Wrap(err)
	}
	opts := []renameio.Option{
		renameio.WithPermissions(configFileMode),
		renameio.WithExistingPermissions(),
		renameio.WithTempDir(dir),
	}
	f, err := renameio.NewPendingFile(path, opts...)
	if err != nil {
		return trace.Wrap(err)
	}
	defer f.Cleanup()

	tmpl, err := template.New(file).Funcs(template.FuncMap{
		"replace": func(s, old, new string) string {
			return strings.ReplaceAll(s, old, new)
		},
		// escape is a best-effort function for escaping quotes in systemd service templates.
		// Paths that are escaped with this method should not be advertised to the user as
		// configurable until a more robust escaping mechanism is shipped.
		// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.syntax.html
		"escape": func(s string) string {
			replacer := strings.NewReplacer(
				`"`, `\"`,
				`\`, `\\`,
			)
			return replacer.Replace(s)
		},
	}).Parse(t)
	if err != nil {
		return trace.Wrap(err)
	}
	err = tmpl.Execute(f, values)
	if err != nil {
		return trace.Wrap(err)
	}
	return trace.Wrap(f.CloseAtomicallyReplace())
}

// replaceTeleportService replaces the default paths in the Teleport service config with namespaced paths.
func (ns *Namespace) replaceTeleportService(cfg []byte) []byte {
	for _, rep := range []struct {
		old, new string
	}{
		{
			old: "/usr/local/bin/",
			new: ns.defaultPathDir + "/",
		},
		{
			old: "/etc/teleport.yaml",
			new: ns.configFile,
		},
		{
			old: "/run/teleport.pid",
			new: ns.pidFile,
		},
	} {
		cfg = bytes.ReplaceAll(cfg, []byte(rep.old), []byte(rep.new))
	}
	return cfg
}

func (ns *Namespace) LogWarning(ctx context.Context) {
	ns.log.WarnContext(ctx, "Custom install suffix specified. Teleport data_dir must be configured in the config file.",
		"data_dir", ns.dataDir,
		"path", ns.defaultPathDir,
		"config", ns.configFile,
		"service", filepath.Base(ns.serviceFile),
		"pid", ns.pidFile,
	)
}

// unversionedConfig is used to read all versions of teleport.yaml, including
// versions that may now be unsupported.
type unversionedConfig struct {
	Teleport unversionedTeleport `yaml:"teleport"`
}

type unversionedTeleport struct {
	AuthServers []string `yaml:"auth_servers"`
	AuthServer  string   `yaml:"auth_server"`
	ProxyServer string   `yaml:"proxy_server"`
	DataDir     string   `yaml:"data_dir"`
}

// overrideFromConfig loads fields from teleport.yaml into the namespace, overriding any defaults.
func (ns *Namespace) overrideFromConfig(ctx context.Context) {
	if ns == nil || ns.configFile == "" {
		return
	}
	path := ns.configFile
	f, err := libutils.OpenFileAllowingUnsafeLinks(path)
	if err != nil {
		ns.log.DebugContext(ctx, "Unable to open Teleport config to read proxy or data dir", "config", path, errorKey, err)
		return
	}
	defer f.Close()
	var cfg unversionedConfig
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		ns.log.DebugContext(ctx, "Unable to parse Teleport config to read proxy or data dir", "config", path, errorKey, err)
		return
	}
	if cfg.Teleport.DataDir != "" {
		ns.dataDir = cfg.Teleport.DataDir
	}

	// Any implicitly defaulted port in teleport.yaml is explicitly defaulted (to 3080).

	var addr string
	var port int
	switch t := cfg.Teleport; {
	case t.ProxyServer != "":
		addr = t.ProxyServer
		port = libdefaults.HTTPListenPort
	case t.AuthServer != "":
		addr = t.AuthServer
		port = libdefaults.AuthListenPort
	case len(t.AuthServers) > 0:
		addr = t.AuthServers[0]
		port = libdefaults.AuthListenPort
	default:
		ns.log.DebugContext(ctx, "Unable to find proxy in Teleport config", "config", path, errorKey, err)
		return
	}
	netaddr, err := libutils.ParseHostPortAddr(addr, port)
	if err != nil {
		ns.log.DebugContext(ctx, "Unable to parse proxy in Teleport config", "config", path, "proxy_addr", addr, "proxy_port", port, errorKey, err)
		return
	}
	ns.defaultProxyAddr = netaddr.String()
}
