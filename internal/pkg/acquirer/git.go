package acquirer

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitAcquirer struct {
	Acquirer
	url    string
	branch string
	path   string
}

// todo - make configurable
const GIT_PATH = "git"

const URL = "url"
const BRANCH = "branch"
const PATH = "path"

// Acquires kapps via git and saves them to `dest`.
func (a GitAcquirer) Acquire(dest string) error {

	// create the dest dir if it doesn't exist
	err := os.MkdirAll(dest, 0755)
	if err != nil {
		return errors.Wrapf(err, "Error creating directory %s", dest)
	}

	var stderrBuf bytes.Buffer

	// git init
	initCmd := exec.Command(GIT_PATH, "init")
	initCmd.Dir = dest
	initCmd.Stderr = &stderrBuf
	err = initCmd.Run()
	if err != nil {
		return errors.Wrapf(err, "Error running: %s. Stderr: %s",
			strings.Join(initCmd.Args, " "), stderrBuf.String())
	}

	stderrBuf.Reset()

	// add origin
	remoteAddCmd := exec.Command(GIT_PATH, "remote", "add", "origin", a.url)
	remoteAddCmd.Dir = dest
	remoteAddCmd.Stderr = &stderrBuf
	err = remoteAddCmd.Run()
	if err != nil {
		return errors.Wrapf(err, "Error running: %s. Stderr=%s",
			strings.Join(remoteAddCmd.Args, " "), stderrBuf.String())
	}

	stderrBuf.Reset()

	fetchCmd := exec.Command(GIT_PATH, "fetch")
	fetchCmd.Dir = dest
	fetchCmd.Stderr = &stderrBuf
	err = fetchCmd.Run()
	if err != nil {
		return errors.Wrapf(err, "Error running: %s. Stderr=%s",
			strings.Join(fetchCmd.Args, " "), stderrBuf.String())
	}

	stderrBuf.Reset()

	configCmd := exec.Command(GIT_PATH, "config", "core.sparsecheckout", "true")
	configCmd.Dir = dest
	configCmd.Stderr = &stderrBuf
	err = configCmd.Run()
	if err != nil {
		return errors.Wrapf(err, "Error running: %s. Stderr=%s",
			strings.Join(configCmd.Args, " "), stderrBuf.String())
	}

	err = appendToFile(filepath.Join(dest, ".git/info/sparse-checkout"),
		fmt.Sprintf("%s/*\n", strings.TrimSuffix(a.path, "/")))
	if err != nil {
		return errors.WithStack(err)
	}

	stderrBuf.Reset()

	checkoutCmd := exec.Command(GIT_PATH, "checkout", a.branch)
	checkoutCmd.Dir = dest
	checkoutCmd.Stderr = &stderrBuf
	err = checkoutCmd.Run()
	if err != nil {
		return errors.Wrapf(err, "Error running: %s. Stderr=%s",
			strings.Join(checkoutCmd.Args, " "), stderrBuf.String())
	}

	return nil
}

// Appends text to a file
func appendToFile(filename string, text string) error {
	// create the file if it doesn't exist
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0744)
	if err != nil {
		return errors.Wrapf(err, "Error opening file %s", filename)
	}

	defer f.Close()

	if _, err = f.WriteString(text); err != nil {
		return errors.Wrapf(err, "Error writing to file %s", filename)
	}

	return nil
}
