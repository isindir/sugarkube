package cacher

import (
	"github.com/pkg/errors"
	"github.com/sugarkube/sugarkube/internal/pkg/acquirer"
	"github.com/sugarkube/sugarkube/internal/pkg/kapp"
	"github.com/sugarkube/sugarkube/internal/pkg/log"
	"os"
	"path/filepath"
	"strings"
)

const CACHE_DIR = ".sugarkube"

// Returns the cache dir for a manifest
func GetManifestCachePath(cacheDir string, manifest kapp.Manifest) string {
	return filepath.Join(cacheDir, manifest.Id)
}

// Returns the root path in the cache for a kapp
func GetKappRootPath(manifestCacheDir string, kappObj kapp.Kapp) string {
	return filepath.Join(manifestCacheDir, kappObj.Id)
}

// Returns the path of a kapp's cache dir where the different sources are
// checked out to
func getKappCachePath(kappRootPath string) string {
	return filepath.Join(kappRootPath, CACHE_DIR)
}

// Build a cache for a manifest into a directory
func CacheManifest(manifest kapp.Manifest, cacheDir string, dryRun bool) error {

	// create a directory to cache all kapps in this manifest in
	manifestCacheDir := GetManifestCachePath(cacheDir, manifest)

	log.Debugf("Creating manifest cache dir: %s", manifestCacheDir)
	err := os.MkdirAll(manifestCacheDir, 0755)
	if err != nil {
		return errors.WithStack(err)
	}

	// acquire each kapp and cache it
	for _, kappObj := range manifest.Kapps {
		// build a directory path for the kapp in the manifest cache directory
		kappRootPath := GetKappRootPath(manifestCacheDir, kappObj)
		// build a directory path for the kapp's .sugarkube cache directory
		kappCacheDir := getKappCachePath(kappRootPath)

		log.Debugf("Creating kapp cache dir: %s", kappCacheDir)
		err := os.MkdirAll(kappCacheDir, 0755)
		if err != nil {
			return errors.WithStack(err)
		}

		err = acquireSource(manifest, kappObj.Sources, kappRootPath, kappCacheDir, dryRun)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// Acquires each source and symlinks it to the target path in the cache directory.
// Runs all acquirers in parallel.
func acquireSource(manifest kapp.Manifest, acquirers []acquirer.Acquirer, rootDir string,
	cacheDir string, dryRun bool) error {
	doneCh := make(chan bool)
	errCh := make(chan error)

	log.Debugf("Acquiring sources for manifest: %s", manifest.Id)

	for _, acquirerImpl := range acquirers {
		go func(a acquirer.Acquirer) {
			acquirerId, err := a.Id()
			if err != nil {
				errCh <- errors.Wrap(err, "Invalid acquirer ID")
			}

			sourceDest := filepath.Join(cacheDir, acquirerId)

			if dryRun {
				log.Debugf("Dry run: Would acquire source into: %s", sourceDest)
			} else {
				err := acquirer.Acquire(a, sourceDest)
				if err != nil {
					errCh <- errors.WithStack(err)
				}
			}

			// todo - this doesn't actually create relative symlinks. Probably need
			// need to use exec.Command and set `command.Dir`, using `ln` directly.
			sourcePath := filepath.Join(sourceDest, a.Path())
			sourcePath = strings.TrimPrefix(sourcePath, rootDir)
			sourcePath = strings.TrimPrefix(sourcePath, "/")

			symLinkTarget := filepath.Join(rootDir, a.Name())

			if dryRun {
				log.Debugf("Dry run. Would symlink cached source %s to %s", sourcePath, symLinkTarget)
			} else {
				if _, err := os.Stat(filepath.Join(rootDir, sourcePath)); err != nil {
					errCh <- errors.Wrapf(err, "Symlink source '%s' doesn't exist", sourcePath)
				}

				log.Debugf("Symlinking cached source %s to %s", sourcePath, symLinkTarget)
				err := os.Symlink(sourcePath, symLinkTarget)
				if err != nil {
					errCh <- errors.Wrapf(err, "Error symlinking source")
				}
			}

			doneCh <- true
		}(acquirerImpl)
	}

	for success := 0; success < len(acquirers); success++ {
		select {
		case err := <-errCh:
			close(doneCh)
			log.Warnf("Error in acquirer goroutines: %s", err)
			return errors.Wrapf(err, "Error running acquirer in goroutine "+
				"for manifest '%s'", manifest.Id)
		case <-doneCh:
			log.Debugf("%d acquirer(s) successfully completed for manifest '%s'",
				success+1, manifest.Id)
		}
	}

	log.Debugf("Finished acquiring sources for manifest: %s", manifest.Id)

	return nil
}

// Diffs a set of manifests against a cache directory and reports any differences
//func DiffCache(manifests []kapp.Manifest, cacheDir string) (???, error) {
// todo - implement
//}
