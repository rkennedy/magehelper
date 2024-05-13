package tools

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// releaseAsset is a subset of the asset information reported in a Github release.
type releaseAsset struct {
	Name               string `json:"name"`
	Size               int    `json:"size"`
	ContentType        string `json:"content_type"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// releaseInfo is a subset of the information reported in a Github release.
type releaseInfo struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

func findTarFile(reader *tar.Reader, targetFileName string) (io.Reader, fs.FileMode, error) {
	for {
		head, err := reader.Next()
		if err != nil {
			return nil, 0, err
		}
		if filepath.Base(head.Name) == targetFileName {
			return reader, fs.FileMode(head.Mode), nil
		}
	}
}

// Given a source stream pointing to the golangci-lint binary we want to extract, write it to the desired location and
// apply the given file permissions.
func writeArchivedFile(source io.Reader, dest string, mode fs.FileMode) error {
	// The file we'll be writing to.
	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("Could not open target for writing: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, source)
	if err != nil {
		return fmt.Errorf("Could not write file to disk: %w", err)
	}

	return nil
}

// Receive a zip file and return the same file, plus its file size. We do it this way to keep function lengths short
// like our linter wants.
func withZipSize(f fs.File, err error) (io.Reader, fs.FileMode, error) {
	if err != nil {
		return nil, 0, err
	}
	info, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return nil, 0, err
	}
	return f, header.Mode(), nil
}

// Find the target file within the zip file represented by the body.
func openZip(size int, body io.Reader, targetFileName string) (io.Reader, fs.FileMode, error) {
	// zip.NewReader requires a ReaderAt, which http doesn't provide. We'll store the resource in memory and
	// then open the zip file from the in-memory buffer.
	zipBuffer := make([]byte, size)

	written, err := io.Copy(bytes.NewBuffer(zipBuffer), body)
	if err != nil {
		return nil, 0, fmt.Errorf("Could not store zip file: %w", err)
	}
	if int(written) != size {
		return nil, 0, fmt.Errorf("Expected %d bytes but read %d instead", size, written)
	}

	unzipper, err := zip.NewReader(bytes.NewReader(zipBuffer), 0)
	if err != nil {
		return nil, 0, fmt.Errorf("Could not open zip file: %w", err)
	}
	return withZipSize(unzipper.Open(targetFileName))
}

// In the releaseInfo's list of assets, find the first tarball or zip file that has a GOOS and GOARCH matching the
// current runtime environment.
func findAsset(info releaseInfo) (*releaseAsset, error) {
	targetSubstring := fmt.Sprintf("-%s-%s.", runtime.GOOS, runtime.GOARCH)
	idx := slices.IndexFunc(info.Assets, func(asset releaseAsset) bool {
		return (asset.ContentType == "application/gzip" || asset.ContentType == "application/zip") &&
			strings.Contains(asset.Name, targetSubstring)
	})
	if idx == -1 {
		return nil, fmt.Errorf("No binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	logV("Found release for %s\n", info.Assets[idx].Name)
	return &info.Assets[idx], nil
}

// Return the platform binary name for golangci-lint.
func golangciLintBinary() string {
	// The name of the file in the archive.
	targetFileName := "golangci-lint"
	if runtime.GOOS == "windows" {
		targetFileName += ".exe"
	}
	return targetFileName
}

// Given an HTTP body, use the asset's content type to optn the body as a tarball or zip file and extract the
// golangci-lint binary to the desired binary location.
//
//revive:disable-next-line:function-length 14 instructions is fine here.
func writeDownloadedFile(body io.Reader, asset *releaseAsset, bin string) (err error) {
	var source io.Reader
	var mode fs.FileMode
	switch asset.ContentType {
	case "application/gzip":
		unzipper, err := gzip.NewReader(body)
		if err != nil {
			return fmt.Errorf("Could not open payload as gzip: %w", err)
		}
		defer unzipper.Close()

		source, mode, err = findTarFile(tar.NewReader(unzipper), golangciLintBinary())
		if err != nil {
			return fmt.Errorf("Could not find target file in tar stream: %w", err)
		}
	case "application/zip":
		source, mode, err = openZip(asset.Size, body, golangciLintBinary())
		if err != nil {
			return err
		}
	}
	return writeArchivedFile(source, bin, mode)
}

// Download the distribution package for the current platform (GOOS and GOARCH), unpack it, and store it at the given
// binary location.
func fetchAndWriteGolangciLint(info releaseInfo, bin string) error {
	asset, err := findAsset(info)
	if err != nil {
		return err
	}

	resp, err := http.Get(asset.BrowserDownloadURL)
	if err != nil {
		return fmt.Errorf("Could not download file: %w", err)
	}
	defer resp.Body.Close()

	return writeDownloadedFile(resp.Body, asset, bin)
}

// GolangciLintDep creates a dependency on the given version of golangci-lint to be installed at the given binary
// location. Pass this to [mg.Deps] or [mg.CtxDeps]. See [InstallGolangciLint].
func GolangciLintDep(bin string, version string) mg.Fn {
	return mg.F(InstallGolangciLint, bin, version)
}

// Get the version of the program at the current location.
func golangcilintVersion(bin string) (string, error) {
	output, err := sh.Output(bin, "--version")
	if err != nil {
		return "", err
	}
	// golangci-lint has version 1.57.2 built with go1.22.1 from 77a8601a on 2024-03-28T19:01:11Z
	re, err := regexp.Compile(`golangci-lint has version ([.0-9]+) `)
	if err != nil {
		return "", err
	}
	matches := re.FindStringSubmatch(output)
	if len(matches) <= 1 {
		return "", fmt.Errorf("Could not find version in %s", output)
	}
	return "v" + matches[1], nil
}

func decodeBody(body io.Reader) (releaseInfo, error) {
	dec := json.NewDecoder(body)
	var info releaseInfo
	err := dec.Decode(&info)
	return info, err
}

// Fetch golangci-lint release information from Github for the given version, which can be either "latest" or a release
// tag.
func getReleaseInfo(version string) (releaseInfo, error) {
	// Download release info for requested version. (If the requested version is "latest" then that will also tell
	// us the actual version, so we can compare with the current file's version.)
	url := "https://api.github.com/repos/golangci/golangci-lint/releases/latest"
	if version != "latest" {
		url = fmt.Sprintf("https://api.github.com/repos/golangci/golangci-lint/releases/tags/%s", version)
	}
	resp, err := http.Get(url)
	if err != nil {
		return releaseInfo{}, err
	}
	defer resp.Body.Close()

	return decodeBody(resp.Body)
}

// InstallGolangciLint installs the given version of golangci-lint at the given binary location. If another version is
// already installed, then it is overwritten with the requested version. Fetching the requested version requires access
// to github.com.
func InstallGolangciLint(bin string, version string) error {
	fileVersion, err := golangcilintVersion(bin)
	// Mage doesn't use %w to wrap errors. Every error is just a string, so Is(ErrNotExist) doesn't work.
	if err != nil && !errors.Is(err, fs.ErrNotExist) && !strings.Contains(err.Error(), "no such file or directory") {
		return err
	}

	info, err := getReleaseInfo(version)
	if err != nil {
		return err
	}

	if fileVersion == info.TagName {
		logV("Command is up to date; file %s, tag %s.\n", fileVersion, info.TagName)
		return nil
	}
	return fetchAndWriteGolangciLint(info, bin)
}
