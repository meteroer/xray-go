package geo

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/net/proxy"
	"xray-go/config"
)

const (
	geoipURL    = "https://github.com/v2fly/geoip/releases/latest/download/geoip.dat"
	geositeURL  = "https://github.com/v2fly/domain-list-community/releases/latest/download/dlc.dat"
	maxAge      = 7 * 24 * time.Hour
	downloadTimeout = 120 * time.Second
)

var (
	geoipPath   string
	geositePath string
)

func init() {
	dir, _ := config.ConfigDir()
	if dir != "" {
		geoipPath = filepath.Join(dir, "geoip.dat")
		geositePath = filepath.Join(dir, "geosite.dat")
	}
}

// Paths returns the paths to geoip.dat and geosite.dat in the config directory.
func Paths() (geoip, geosite string) {
	return geoipPath, geositePath
}

// NeedUpdate checks if geo data files need to be updated (missing or older than maxAge).
func NeedUpdate() bool {
	for _, path := range []string{geoipPath, geositePath} {
		if path == "" {
			return true
		}
		info, err := os.Stat(path)
		if err != nil {
			return true
		}
		if time.Since(info.ModTime()) > maxAge {
			return true
		}
	}
	return false
}

// Ensure copies geo data files from config dir to the working directory if they exist.
// This allows Xray-core to find them automatically.
func Ensure(workDir string) error {
	for _, src := range []string{geoipPath, geositePath} {
		if src == "" {
			continue
		}
		if _, err := os.Stat(src); err != nil {
			continue
		}
		dst := filepath.Join(workDir, filepath.Base(src))
		if sameFile(src, dst) {
			continue
		}
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("copy %s: %w", filepath.Base(src), err)
		}
	}
	return nil
}

// DownloadAll downloads geoip.dat and geosite.dat via the given SOCKS5 proxy.
// If proxyAddr is empty, downloads directly.
func DownloadAll(proxyAddr string) error {
	if geoipPath == "" || geositePath == "" {
		return fmt.Errorf("config directory not available")
	}

	if err := downloadFile(geoipURL, geoipPath, proxyAddr); err != nil {
		return fmt.Errorf("download geoip.dat: %w", err)
	}
	if err := downloadFile(geositeURL, geositePath, proxyAddr); err != nil {
		return fmt.Errorf("download geosite.dat: %w", err)
	}
	return nil
}

func downloadFile(url, dstPath, proxyAddr string) error {
	var client *http.Client
	if proxyAddr != "" {
		dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
		if err != nil {
			return fmt.Errorf("socks5 dialer: %w", err)
		}
		ctxDialer, ok := dialer.(proxy.ContextDialer)
		if !ok {
			return fmt.Errorf("socks5 dialer does not support DialContext")
		}
		client = &http.Client{
			Timeout: downloadTimeout,
			Transport: &http.Transport{
				DialContext: ctxDialer.DialContext,
			},
		}
	} else {
		client = &http.Client{Timeout: downloadTimeout}
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	tmpPath := dstPath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer os.Remove(tmpPath)

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, dstPath)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func sameFile(a, b string) bool {
	fiA, err := os.Stat(a)
	if err != nil {
		return false
	}
	fiB, err := os.Stat(b)
	if err != nil {
		return false
	}
	return os.SameFile(fiA, fiB)
}
