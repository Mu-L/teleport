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
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/lib/autoupdate"
	"github.com/gravitational/teleport/lib/utils"
)

const (
	// checksumType for Teleport tgzs
	checksumType = "sha256"
	// checksumHexLen is the length of the Teleport checksum.
	checksumHexLen = sha256.Size * 2 // bytes to hex
	// maxServiceFileSize is the maximum size allowed for a systemd service file.
	maxServiceFileSize = 1_000_000 // 1 MB
	// configFileMode is the mode used for new configuration files.
	configFileMode = 0644
	// systemDirMode is the mode used for new directories.
	systemDirMode = 0755
)

const (
	// serviceDir contains the relative path to the Teleport SystemD service dir.
	serviceDir = "lib/systemd/system"
	// serviceName contains the upstream name of the Teleport SystemD service file.
	serviceName = "teleport.service"
)

// ServiceFile represents a systemd service file for a Teleport binary.
//
// ExampleName and ExampleFunc are used to parse an example configuration
// file and copy it into the service directory. This mechanism of service
// file generation is only provided to support downgrading to older versions
// of the updater that do not install the service during the setup phase using
// logic from lib/config.
type ServiceFile struct {
	// Path is the full path to the linked service file.
	Path string
	// Binary is the corresponding linked binary name.
	Binary string
	// ExampleName is the name of the example service file.
	// Deprecated.
	ExampleName string
	// ExampleFunc can be used to create the service during linking from
	// an archived example, instead of creating it during the setup phase.
	// Deprecated.
	ExampleFunc ExampleFunc
}

// A ExampleFunc generates a systemd service by replacing an example file.
// Deprecated.
type ExampleFunc func(cfg []byte, path string, flags autoupdate.InstallFlags) []byte

// LocalInstaller manages the creation and removal of installations
// of Teleport.
// SetRequiredUmask must be called before any methods are executed.
type LocalInstaller struct {
	// InstallDir contains each installation, named by version.
	InstallDir string
	// TargetServiceFile contains a copy of the linked installation's systemd service.
	TargetServices []ServiceFile
	// SystemBinDir contains binaries for the system (packaged) install of Teleport.
	SystemBinDir string
	// SystemServiceDir contains the systemd service directory for the system (packaged) install of Teleport.
	SystemServiceDir string
	// HTTP is an HTTP client for downloading Teleport.
	HTTP *http.Client
	// Log contains a logger.
	Log *slog.Logger
	// ReservedFreeTmpDisk is the amount of disk that must remain free in /tmp
	ReservedFreeTmpDisk uint64
	// ReservedFreeInstallDisk is the amount of disk that must remain free in the install directory.
	ReservedFreeInstallDisk uint64
	// ValidateBinary returns true if a file is a linkable binary, or
	// false if a file should not be linked.
	ValidateBinary func(ctx context.Context, path string) (bool, error)
	// Template is download URI Template of Teleport packages.
	Template string
}

// Remove a Teleport version directory from InstallDir.
// This function is idempotent.
// See Installer interface for additional specs.
func (li *LocalInstaller) Remove(ctx context.Context, rev Revision) error {
	// os.RemoveAll is dangerous because it can remove an entire directory tree.
	// We must validate the version to ensure that we remove only a single path
	// element under the InstallDir, and not InstallDir or its parents.
	// revisionDir performs these validations.
	versionDir, err := li.revisionDir(rev)
	if err != nil {
		return trace.Wrap(err)
	}

	// invalidate checksum first, to protect against partially-removed
	// directory with valid checksum.
	err = os.Remove(filepath.Join(versionDir, checksumType))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return trace.Wrap(err)
	}
	if err := os.RemoveAll(versionDir); err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// Install a Teleport version directory in InstallDir.
// This function is idempotent.
// See Installer interface for additional specs.
func (li *LocalInstaller) Install(ctx context.Context, rev Revision, baseURL string, force bool) (err error) {
	versionDir, err := li.revisionDir(rev)
	if err != nil {
		return trace.Wrap(err)
	}
	sumPath := filepath.Join(versionDir, checksumType)

	// generate download URI from Template
	uri, err := autoupdate.MakeURL(li.Template, baseURL, autoupdate.DefaultPackage, rev.Version, rev.Flags)
	if err != nil {
		return trace.Wrap(err)
	}

	// Get new and old checksums. If they match, skip download.
	// Otherwise, clear the old version directory and re-download.
	checksumURI := uri + "." + checksumType
	newSum, err := li.getChecksum(ctx, checksumURI)
	if err != nil {
		return trace.Wrap(err, "failed to download checksum from %s", checksumURI)
	}
	oldSum, err := readChecksum(sumPath)
	versionPresent := err == nil
	if versionPresent && bytes.Equal(oldSum, newSum) {
		li.Log.InfoContext(ctx, "Version already present.", "version", rev)
		return nil
	}
	if versionPresent {
		li.Log.WarnContext(ctx, "Removing version that does not match checksum.", "version", rev)
	} else if !errors.Is(err, os.ErrNotExist) {
		li.Log.WarnContext(ctx, "Removing version with unreadable checksum.", "version", rev, "error", err)
	}
	if versionPresent || !errors.Is(err, os.ErrNotExist) {
		if force {
			if err := li.Remove(ctx, rev); err != nil {
				return trace.Wrap(err)
			}
		} else {
			return trace.Errorf("refusing to remove linked installation of Teleport")
		}
	}

	// Verify that we have enough free temp space, then download tgz
	freeTmp, err := utils.FreeDiskWithReserve(os.TempDir(), li.ReservedFreeTmpDisk)
	if err != nil {
		return trace.Wrap(err, "failed to calculate free disk")
	}
	f, err := os.CreateTemp("", "teleport-update-")
	if err != nil {
		return trace.Wrap(err, "failed to create temporary file")
	}
	defer func() {
		_ = f.Close() // data never read after close
		if err := os.Remove(f.Name()); err != nil {
			li.Log.WarnContext(ctx, "Failed to cleanup temporary download.", "error", err)
		}
	}()
	pathSum, err := li.download(ctx, f, int64(freeTmp), uri)
	if err != nil {
		return trace.Wrap(err, "failed to download teleport")
	}
	// Seek to the start of the tgz file after writing
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return trace.Wrap(err, "failed seek to start of download")
	}

	// If interrupted, close the file immediately to stop extracting.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	context.AfterFunc(ctx, func() {
		_ = f.Close() // safe to close file multiple times
	})
	// Check integrity before decompression
	if !bytes.Equal(newSum, pathSum) {
		return trace.Errorf("mismatched checksum, download possibly corrupt")
	}
	// Get uncompressed size of the tgz
	n, err := uncompressedSize(f)
	if err != nil {
		return trace.Wrap(err, "failed to determine uncompressed size")
	}
	// Seek to start of tgz after reading size
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return trace.Wrap(err, "failed seek to start")
	}

	// If there's an error after we start extracting, delete the version dir.
	defer func() {
		if err != nil {
			if err := os.RemoveAll(versionDir); err != nil {
				li.Log.WarnContext(ctx, "Failed to cleanup broken version extraction.", "error", err, "dir", versionDir)
			}
		}
	}()

	// Extract tgz into version directory.
	if err := li.extract(ctx, versionDir, f, n, rev.Flags); err != nil {
		return trace.Wrap(err, "failed to extract teleport")
	}
	// Write the checksum last. This marks the version directory as valid.
	if err := os.WriteFile(sumPath, []byte(hex.EncodeToString(newSum)), configFileMode); err != nil {
		return trace.Wrap(err, "failed to write checksum")
	}
	return nil
}

// readChecksum from the version directory.
func readChecksum(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	defer f.Close()
	var buf bytes.Buffer
	_, err = io.CopyN(&buf, f, checksumHexLen)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	raw := buf.String()
	sum, err := hex.DecodeString(raw)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return sum, nil
}

func (li *LocalInstaller) getChecksum(ctx context.Context, url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	resp, err := li.HTTP.Do(req)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, trace.Errorf("checksum not found: %s", url)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, trace.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}

	var buf bytes.Buffer
	_, err = io.CopyN(&buf, resp.Body, checksumHexLen)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	sum, err := hex.DecodeString(buf.String())
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return sum, nil
}

func (li *LocalInstaller) download(ctx context.Context, w io.Writer, max int64, url string) (sum []byte, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	startTime := time.Now()
	resp, err := li.HTTP.Do(req)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, trace.Errorf("Teleport download not found: %s", url)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, trace.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}
	li.Log.InfoContext(ctx, "Downloading Teleport tarball.", "url", url, "size", resp.ContentLength)

	// Ensure there's enough space in /tmp for the download.
	size := resp.ContentLength
	if size < 0 {
		li.Log.WarnContext(ctx, "Content length missing from response, unable to verify Teleport download size.")
		size = max
	} else if size > max {
		return nil, trace.Errorf("size of download (%d bytes) exceeds available disk space (%d bytes)", resp.ContentLength, max)
	}
	// Calculate checksum concurrently with download.
	shaReader := sha256.New()
	tee := io.TeeReader(resp.Body, shaReader)
	tee = io.TeeReader(tee, &progressLogger{
		ctx:   ctx,
		log:   li.Log,
		level: slog.LevelInfo,
		name:  path.Base(resp.Request.URL.Path),
		max:   int(resp.ContentLength),
		lines: 5,
	})
	n, err := io.CopyN(w, tee, size)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	if resp.ContentLength >= 0 && n != resp.ContentLength {
		return nil, trace.Errorf("mismatch in Teleport download size")
	}
	li.Log.InfoContext(ctx, "Download complete.", "duration", time.Since(startTime), "size", n)
	return shaReader.Sum(nil), nil
}

func (li *LocalInstaller) extract(ctx context.Context, dstDir string, src io.Reader, max int64, flags autoupdate.InstallFlags) error {
	if err := os.MkdirAll(dstDir, systemDirMode); err != nil {
		return trace.Wrap(err)
	}
	free, err := utils.FreeDiskWithReserve(dstDir, li.ReservedFreeInstallDisk)
	if err != nil {
		return trace.Wrap(err, "failed to calculate free disk in %s", dstDir)
	}
	// Bail if there's not enough free disk space at the target
	if d := int64(free) - max; d < 0 {
		return trace.Errorf("%s needs %d additional bytes of disk space for decompression", dstDir, -d)
	}
	zr, err := gzip.NewReader(src)
	if err != nil {
		return trace.Wrap(err, "requires gzip-compressed body")
	}
	li.Log.InfoContext(ctx, "Extracting Teleport tarball.", "path", dstDir, "size", max)

	err = utils.Extract(zr, dstDir, tgzExtractPaths(flags&(autoupdate.FlagEnterprise|autoupdate.FlagFIPS) != 0)...)
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// tgzExtractPaths describes how to extract the Teleport tgz.
// See utils.Extract for more details on how this list is parsed.
// Paths must use tarball-style / separators (not filepath).
func tgzExtractPaths(ent bool) []utils.ExtractPath {
	prefix := "teleport"
	if ent {
		prefix += "-ent"
	}
	return []utils.ExtractPath{
		{Src: path.Join(prefix, "examples/systemd/teleport.service"), Dst: filepath.Join(serviceDir, serviceName), DirMode: systemDirMode},
		{Src: path.Join(prefix, "examples"), Skip: true, DirMode: systemDirMode},
		{Src: path.Join(prefix, "install"), Skip: true, DirMode: systemDirMode},
		{Src: path.Join(prefix, "README.md"), Dst: "share/README.md", DirMode: systemDirMode},
		{Src: path.Join(prefix, "CHANGELOG.md"), Dst: "share/CHANGELOG.md", DirMode: systemDirMode},
		{Src: path.Join(prefix, "VERSION"), Dst: "share/VERSION", DirMode: systemDirMode},
		{Src: path.Join(prefix, "LICENSE-community"), Dst: "share/LICENSE-community", DirMode: systemDirMode},
		{Src: prefix, Dst: "bin", DirMode: systemDirMode},
	}
}

func uncompressedSize(f io.Reader) (int64, error) {
	// NOTE: The gzip length trailer is very unreliable,
	//   but we could optimize this in the future if
	//   we are willing to verify that all published
	//   Teleport tarballs have valid trailers.
	r, err := gzip.NewReader(f)
	if err != nil {
		return 0, trace.Wrap(err)
	}
	n, err := io.Copy(io.Discard, r)
	if err != nil {
		return 0, trace.Wrap(err)
	}
	return n, nil
}

// List installed versions of Teleport.
func (li *LocalInstaller) List(ctx context.Context) (revs []Revision, err error) {
	entries, err := os.ReadDir(li.InstallDir)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		rev, err := NewRevisionFromDir(entry.Name())
		if err != nil {
			return nil, trace.Wrap(err)
		}
		revs = append(revs, rev)
	}
	return revs, nil
}

// Link the specified version into pathDir and TargetServiceFile.
// The revert function restores the previous linking.
// If force is true, Link will overwrite files that are not symlinks.
// See Installer interface for additional specs.
func (li *LocalInstaller) Link(ctx context.Context, rev Revision, pathDir string, force bool) (revert func(context.Context) bool, err error) {
	revert = func(context.Context) bool { return true }
	versionDir, err := li.revisionDir(rev)
	if err != nil {
		return revert, trace.Wrap(err)
	}
	revert, err = li.forceLinks(ctx,
		filepath.Join(versionDir, "bin"),
		filepath.Join(versionDir, serviceDir),
		pathDir, force, rev.Flags,
	)
	if err != nil {
		return revert, trace.Wrap(err)
	}
	return revert, nil
}

// LinkSystem links the system (package) version into defaultPathDir and TargetServiceFile.
// This prevents namespaced installations in /opt/teleport from linking to the system package.
// The revert function restores the previous linking.
// See Installer interface for additional specs.
func (li *LocalInstaller) LinkSystem(ctx context.Context) (revert func(context.Context) bool, err error) {
	// The system package service file is always removed without flags, so pass
	// no flags here to match the behavior.
	revert, err = li.forceLinks(ctx, li.SystemBinDir, li.SystemServiceDir, defaultPathDir, false, 0)
	return revert, trace.Wrap(err)
}

// TryLink links the specified version into pathDir, but only in the case that
// no installation of Teleport is already linked or partially linked.
// See Installer interface for additional specs.
func (li *LocalInstaller) TryLink(ctx context.Context, revision Revision, pathDir string) error {
	versionDir, err := li.revisionDir(revision)
	if err != nil {
		return trace.Wrap(err)
	}
	return trace.Wrap(li.tryLinks(ctx,
		filepath.Join(versionDir, "bin"),
		filepath.Join(versionDir, serviceDir),
		pathDir, revision.Flags,
	))
}

// TryLinkSystem links the system installation to defaultPathDir, but only in the case that
// no installation of Teleport is already linked or partially linked.
// See Installer interface for additional specs.
func (li *LocalInstaller) TryLinkSystem(ctx context.Context) error {
	// The system package service file is always removed without flags, so pass
	// no flags here to match the behavior.
	return trace.Wrap(li.tryLinks(ctx, li.SystemBinDir, li.SystemServiceDir, defaultPathDir, 0))
}

// Unlink unlinks a version from pathDir and TargetServiceFile.
// See Installer interface for additional specs.
func (li *LocalInstaller) Unlink(ctx context.Context, rev Revision, pathDir string) error {
	versionDir, err := li.revisionDir(rev)
	if err != nil {
		return trace.Wrap(err)
	}
	return trace.Wrap(li.removeLinks(ctx, filepath.Join(versionDir, "bin"), pathDir))
}

// UnlinkSystem unlinks the system (package) version from defaultPathDir and TargetServiceFile.
// See Installer interface for additional specs.
func (li *LocalInstaller) UnlinkSystem(ctx context.Context) error {
	return trace.Wrap(li.removeLinks(ctx, li.SystemBinDir, defaultPathDir))
}

// symlink from oldname to newname
type symlink struct {
	oldname, newname string
}

// smallFile is a file small enough to be stored in memory.
type smallFile struct {
	name string
	data []byte
	mode os.FileMode
}

// forceLinks replaces binary links and service files using files in binDir and svcDir.
// Existing links and files are replaced, but mismatched links and files will result in error.
// forceLinks will revert any overridden links or files if it hits an error.
// If successful, forceLinks may also be reverted after it returns by calling revert.
// The revert function returns true if reverting succeeds.
// If force is true, non-link files will be overwritten.
func (li *LocalInstaller) forceLinks(ctx context.Context, srcBinDir, srcSvcDir, dstBinDir string, force bool, flags autoupdate.InstallFlags) (revert func(context.Context) bool, err error) {
	// setup revert function
	var (
		revertLinks []symlink
		revertFiles []smallFile
	)
	revert = func(ctx context.Context) bool {
		// This function is safe to call repeatedly.
		// Returns true only when all changes are successfully reverted.
		var (
			keepLinks []symlink
			keepFiles []smallFile
		)
		for _, l := range revertLinks {
			err := atomicSymlink(l.oldname, l.newname)
			if err != nil {
				keepLinks = append(keepLinks, l)
				li.Log.ErrorContext(ctx, "Failed to revert symlink", "oldname", l.oldname, "newname", l.newname, errorKey, err)
			}
		}
		for _, f := range revertFiles {
			err := writeFileAtomicWithinDir(f.name, f.data, f.mode)
			if err != nil {
				keepFiles = append(keepFiles, f)
				li.Log.ErrorContext(ctx, "Failed to revert files", "name", f.name, errorKey, err)
			}
		}
		revertLinks = keepLinks
		revertFiles = keepFiles
		return len(revertLinks) == 0 && len(revertFiles) == 0
	}
	// revert immediately on error, so caller can ignore revert arg
	defer func() {
		if err != nil {
			revert(ctx)
		}
	}()

	// ensure source directory exists
	entries, err := os.ReadDir(srcBinDir)
	if errors.Is(err, os.ErrNotExist) {
		return revert, trace.Wrap(ErrNoBinaries)
	}
	if err != nil {
		return revert, trace.Wrap(err, "failed to read Teleport binary directory")
	}

	// ensure target directories exist before trying to create links
	err = os.MkdirAll(dstBinDir, systemDirMode)
	if err != nil {
		return revert, trace.Wrap(err)
	}
	for _, s := range li.TargetServices {
		err = os.MkdirAll(filepath.Dir(s.Path), systemDirMode)
		if err != nil {
			return revert, trace.Wrap(err)
		}
	}

	// create binary links
	var linked int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		oldname := filepath.Join(srcBinDir, entry.Name())
		newname := filepath.Join(dstBinDir, entry.Name())
		exec, err := li.ValidateBinary(ctx, oldname)
		if err != nil {
			return revert, trace.Wrap(err)
		}
		if !exec {
			continue
		}
		orig, err := forceLink(oldname, newname, force)
		if err != nil && !errors.Is(err, os.ErrExist) {
			return revert, trace.Wrap(err, "failed to create symlink for %s", entry.Name())
		}
		if orig != "" {
			revertLinks = append(revertLinks, symlink{
				oldname: orig,
				newname: newname,
			})
		}
		linked++
	}
	if linked == 0 {
		return revert, trace.Wrap(ErrNoBinaries)
	}

	// create systemd service files

	for _, s := range li.TargetServices {
		orig, err := copyService(s, srcSvcDir, dstBinDir, flags)
		if err != nil && !errors.Is(err, os.ErrExist) {
			return revert, trace.Wrap(err, "failed to copy service %s", filepath.Base(s.Path))
		}
		if orig != nil {
			revertFiles = append(revertFiles, *orig)
		}
	}

	return revert, nil
}

// copyService copies a systemd service file from src to dst.
// The contents of both src and dst must be smaller than n.
//
// Copied data is processed by s.ExampleFunc.
// If s.ExampleFunc nil, no data is copied, but the original file contents are still returned.
//
// See prepCopy and forceCopy for more details.
func copyService(s ServiceFile, exampleDir string, dstBinDir string, flags autoupdate.InstallFlags) (orig *smallFile, err error) {
	const n = maxServiceFileSize
	if s.ExampleFunc != nil {
		srcData, err := readFileAtMost(filepath.Join(exampleDir, s.ExampleName), n)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		orig, err = forceCopy(s.Path, s.ExampleFunc(srcData, dstBinDir, flags), n)
		return orig, trace.Wrap(err)
	}
	orig, err = prepCopy(s.Path, n)
	return orig, trace.Wrap(err)
}

// forceLink attempts to create a symlink, atomically replacing an existing link if already present.
// If a non-symlink file or directory exists in newname already and force is false, forceLink errors with ErrFilePresent.
// If the link is already present with the desired oldname, forceLink returns os.ErrExist.
func forceLink(oldname, newname string, force bool) (orig string, err error) {
	orig, err = os.Readlink(newname)
	if errors.Is(err, os.ErrInvalid) ||
		errors.Is(err, syscall.EINVAL) { // workaround missing ErrInvalid wrapper
		if force {
			return "", trace.Wrap(atomicSymlink(oldname, newname))
		}
		// important: do not attempt to replace a non-linked install of Teleport without force
		return "", trace.Wrap(ErrFilePresent, "refusing to replace file at %s", newname)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", trace.Wrap(err)
	}
	if orig == oldname {
		return "", trace.Wrap(os.ErrExist)
	}
	err = atomicSymlink(oldname, newname)
	if err != nil {
		return "", trace.Wrap(err)
	}
	return orig, nil
}

// forceCopy atomically copies a file from srcData to dst, replacing an existing file at dst if needed.
// The contents of dst must be smaller than n.
// forceCopy returns the original file path, mode, and contents as orig.
// If an irregular file, too large file, or directory exists in dst already, forceCopy errors.
// If the file is already present with the desired contents, forceCopy returns os.ErrExist.
func forceCopy(dst string, srcData []byte, n int64) (orig *smallFile, err error) {
	orig, err = prepCopy(dst, n)
	if err != nil {
		return orig, trace.Wrap(err)
	}
	if orig != nil && bytes.Equal(srcData, orig.data) {
		return nil, trace.Wrap(os.ErrExist)
	}
	err = writeFileAtomicWithinDir(dst, srcData, configFileMode)
	if err != nil {
		return orig, trace.Wrap(err)
	}
	return orig, nil
}

// prepCopy validates and returns a preserved original copy of a file with
// length <= n at the path specified by dst.
func prepCopy(dst string, n int64) (orig *smallFile, err error) {
	fi, err := os.Lstat(dst)
	if errors.Is(err, os.ErrNotExist) {
		return orig, nil
	}
	if err != nil {
		return nil, trace.Wrap(err)
	}
	orig = &smallFile{
		name: dst,
		mode: fi.Mode(),
	}
	if !orig.mode.IsRegular() {
		return nil, trace.Errorf("refusing to replace irregular file at %s", dst)
	}
	orig.data, err = readFileAtMost(dst, n)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return orig, nil
}

// readFileAtMost reads a file up to n, or errors if it is too large.
func readFileAtMost(name string, n int64) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := utils.ReadAtMost(f, n)
	return data, trace.Wrap(err)
}

func (li *LocalInstaller) removeLinks(ctx context.Context, srcBinDir, dstBinDir string) error {
	entries, err := os.ReadDir(srcBinDir)
	if err != nil {
		return trace.Wrap(err, "failed to find Teleport binary directory")
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		oldname := filepath.Join(srcBinDir, entry.Name())
		newname := filepath.Join(dstBinDir, entry.Name())
		v, err := os.Readlink(newname)
		if errors.Is(err, os.ErrNotExist) ||
			errors.Is(err, os.ErrInvalid) ||
			errors.Is(err, syscall.EINVAL) {
			li.Log.DebugContext(ctx, "Link not present.", "oldname", oldname, "newname", newname)
			continue
		}
		if err != nil {
			return trace.Wrap(err, "error reading link for %s", filepath.Base(newname))
		}
		if v != oldname {
			li.Log.DebugContext(ctx, "Skipping link to different binary.", "oldname", oldname, "newname", newname)
			continue
		}
		if err := os.Remove(newname); err != nil {
			li.Log.ErrorContext(ctx, "Unable to remove link.", "oldname", oldname, "newname", newname, errorKey, err)
			continue
		}

		for _, s := range li.TargetServices {
			if filepath.Base(newname) != s.Binary {
				continue
			}
			// binRev is either the version or "system"
			binRev := filepath.Base(filepath.Dir(filepath.Dir(oldname)))
			rev, err := NewRevisionFromDir(binRev)
			if err != nil {
				li.Log.DebugContext(ctx, "Service not present.", "path", s.Path)
				continue
			}
			revMarker := genMarker(rev)
			diskMarker, err := readFileLimit(s.Path, int64(len(revMarker)))
			if errors.Is(err, os.ErrNotExist) {
				li.Log.DebugContext(ctx, "Service not present.", "path", s.Path)
				continue
			}
			if err != nil {
				return trace.Wrap(err)
			}
			// Note that old versions of teleport-update will install services without the marker.
			// Certain version combinations (before and after this commit) may leave services behind
			// if they are not replaced by the new version of teleport-update. This should only impact
			// explicit system package unlinking, which is rarely used.
			if string(diskMarker) != revMarker {
				li.Log.WarnContext(ctx, "Removed binary link, but skipping removal of custom service that does not match the binary.",
					"service", filepath.Base(s.Path), "binary", filepath.Base(newname))
				continue
			}
			if err := os.Remove(s.Path); err != nil {
				return trace.Wrap(err, "error removing copy of %s", filepath.Base(s.Path))
			}
		}
	}
	return nil
}

// tryLinks create binary and service links for files in binDir and svcDir if links are not already present.
// Existing links that point to files outside binDir or svcDir, as well as existing non-link files, will error.
// tryLinks will not attempt to create any links if linking could result in an error.
// However, concurrent changes to links may result in an error with partially-complete linking.
func (li *LocalInstaller) tryLinks(ctx context.Context, srcBinDir, srcSvcDir, dstBinDir string, flags autoupdate.InstallFlags) error {
	// ensure source directory exists
	entries, err := os.ReadDir(srcBinDir)
	if errors.Is(err, os.ErrNotExist) {
		return trace.Wrap(ErrNoBinaries)
	}
	if err != nil {
		return trace.Wrap(err, "failed to read Teleport binary directory")
	}

	// ensure target directories exist before trying to create links
	err = os.MkdirAll(dstBinDir, systemDirMode)
	if err != nil {
		return trace.Wrap(err)
	}
	for _, s := range li.TargetServices {
		err = os.MkdirAll(filepath.Dir(s.Path), systemDirMode)
		if err != nil {
			return trace.Wrap(err)
		}
	}

	// validate that we can link all system binaries before attempting linking
	var links []symlink
	var linked int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		oldname := filepath.Join(srcBinDir, entry.Name())
		newname := filepath.Join(dstBinDir, entry.Name())
		exec, err := li.ValidateBinary(ctx, oldname)
		if err != nil {
			return trace.Wrap(err)
		}
		if !exec {
			continue
		}
		ok, err := needsLink(oldname, newname)
		if err != nil {
			return trace.Wrap(err, "error evaluating link for %s", filepath.Base(oldname))
		}
		if ok {
			links = append(links, symlink{oldname, newname})
		}
		linked++
	}
	// bail if no binaries can be linked
	if linked == 0 {
		return trace.Wrap(ErrNoBinaries)
	}

	// link binaries that are missing links
	for _, link := range links {
		if err := os.Symlink(link.oldname, link.newname); err != nil {
			return trace.Wrap(err, "failed to create symlink for %s", filepath.Base(link.oldname))
		}
	}

	for _, s := range li.TargetServices {
		_, err := copyService(s, srcSvcDir, dstBinDir, flags)
		if err != nil && !errors.Is(err, os.ErrExist) {
			return trace.Wrap(err, "failed to copy service %s", filepath.Base(s.Path))
		}
	}

	return nil
}

// needsLink returns true when a symlink from oldname to newname needs to be created, or false if it exists.
// If a non-symlink file or directory exists at newname, needsLink errors with ErrFilePresent.
// If a symlink to a different location exists, needsLink errors with ErrLinked.
func needsLink(oldname, newname string) (ok bool, err error) {
	orig, err := os.Readlink(newname)
	if errors.Is(err, os.ErrInvalid) ||
		errors.Is(err, syscall.EINVAL) { // workaround missing ErrInvalid wrapper
		// important: do not attempt to replace a non-linked install of Teleport
		return false, trace.Wrap(ErrFilePresent, "refusing to replace file at %s", newname)
	}
	if errors.Is(err, os.ErrNotExist) {
		return true, nil
	}
	if err != nil {
		return false, trace.Wrap(err)
	}
	if orig != oldname {
		return false, trace.Wrap(ErrLinked, "refusing to replace link at %s", newname)
	}
	return false, nil
}

// revisionDir returns the storage directory for a Teleport revision.
// revisionDir will fail if the revision cannot be used to construct the directory name.
// For example, it ensures that ".." cannot be provided to return a system directory.
func (li *LocalInstaller) revisionDir(rev Revision) (string, error) {
	installDir, err := filepath.Abs(li.InstallDir)
	if err != nil {
		return "", trace.Wrap(err)
	}
	versionDir := filepath.Join(installDir, rev.Dir())
	if filepath.Dir(versionDir) != filepath.Clean(installDir) {
		return "", trace.Errorf("refusing to link directory outside of version directory")
	}
	return versionDir, nil
}

// IsLinked returns true if any binaries for Revision rev are linked to pathDir.
// Returns os.ErrNotExist error if the revision does not exist.
func (li *LocalInstaller) IsLinked(ctx context.Context, rev Revision, pathDir string) (bool, error) {
	versionDir, err := li.revisionDir(rev)
	if err != nil {
		return false, trace.Wrap(err)
	}
	binDir := filepath.Join(versionDir, "bin")
	entries, err := os.ReadDir(binDir)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, trace.Wrap(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		v, err := os.Readlink(filepath.Join(pathDir, entry.Name()))
		if err != nil {
			continue
		}
		if filepath.Clean(v) == filepath.Join(binDir, entry.Name()) {
			return true, nil
		}
	}
	return false, nil
}
