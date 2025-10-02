package cli

import (
	"fmt"
	"io"
	"mrunc/internal/config"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v2"
)

func initCommand(ctx *cli.Context) error {
	// allow override for testing
	baseDir := os.Getenv("MRUNC_BASE")
	if baseDir == "" {
		baseDir = config.BaseImageDir
	}
	ubuntuDir := filepath.Join(baseDir, "ubuntu")
	verifyFile := filepath.Join(ubuntuDir, "etc", "os-release")

	// If rootfs already installed: skip download/extract but ensure ubuntuDir exists
	if _, err := os.Stat(verifyFile); err == nil {
		fmt.Println("Ubuntu rootfs already initialized, skipping rootfs download/extract.")
		if err := os.MkdirAll(ubuntuDir, 0755); err != nil {
			return fmt.Errorf("failed to ensure ubuntu dir: %v", err)
		}
	} else {
		// need to install rootfs
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			return fmt.Errorf("failed to create base images dir: %v", err)
		}

		tarball := "/tmp/ubuntu-24.04-rootfs.tar.xz"
		ubuntuURL := "https://cloud-images.ubuntu.com/minimal/releases/noble/release/ubuntu-24.04-minimal-cloudimg-amd64-root.tar.xz"

		// download tarball with retries
		if _, err := os.Stat(tarball); os.IsNotExist(err) {
			fmt.Println("Downloading Ubuntu rootfs...")
			if err := downloadWithRetries(tarball, ubuntuURL, 3, 30*time.Second); err != nil {
				return fmt.Errorf("failed to download rootfs: %v", err)
			}
		}

		// ensure target dir
		if err := os.MkdirAll(ubuntuDir, 0755); err != nil {
			return fmt.Errorf("failed to create ubuntu dir: %v", err)
		}

		// extract using system tar (handles symlinks, devices, permissions)
		fmt.Println("Extracting Ubuntu rootfs (using tar)...")
		if err := extractTarXzWithTarCmd(tarball, ubuntuDir); err != nil {
			return fmt.Errorf("failed to extract rootfs: %v", err)
		}

		// verify
		if _, err := os.Stat(verifyFile); err != nil {
			return fmt.Errorf("rootfs verification failed (missing %s): %v", verifyFile, err)
		}
		// cleanup
		_ = os.Remove(tarball)
		fmt.Println("Ubuntu rootfs extracted successfully.")
	}

	// ALWAYS download the config from GitHub (overwrites previous)
	configURL := config.ConfigURLTemplate
	destPath := filepath.Join(ubuntuDir, "ubuntu.json")
	fmt.Printf("Downloading default ubuntu.json config from %s ...\n", configURL)
	if err := downloadWithRetries(destPath, configURL, 3, 10*time.Second); err != nil {
		return fmt.Errorf("failed to download ubuntu.json: %v", err)
	}
	// ensure sane permissions
	_ = os.Chmod(destPath, 0644)

	fmt.Println("Config written to", destPath)
	return nil
}

func downloadWithRetries(dest, url string, retries int, delay time.Duration) error {
	var lastErr error
	for i := range retries {
		if err := downloadFile(dest, url); err != nil {
			lastErr = err
			fmt.Printf("download attempt %d/%d failed: %v\n", i+1, retries, err)
			time.Sleep(delay)
			continue
		}
		return nil
	}
	return lastErr
}

func downloadFile(dest, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status when downloading %s: %s", url, resp.Status)
	}

	tmp := dest + ".tmp"
	// ensure parent exists
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		_ = os.Remove(tmp)
		return err
	}
	// atomic replace
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func extractTarXzWithTarCmd(src, dest string) error {
	// requires `tar` on host. -xJf -> extract xz. --numeric-owner avoids name lookup
	cmd := exec.Command("tar", "--numeric-owner", "-xJf", src, "-C", dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
