package acquirer

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sugarkube/sugarkube/internal/pkg/log"
	"strings"
)

type Acquirer interface {
	acquire(dest string) error
	Id() (string, error)
	Name() string
	Path() string
}

const ACQUIRER_KEY = "acquirer"
const GIT = "git"

// Factory that creates acquirers
func acquirerFactory(name string, settings map[string]string) (Acquirer, error) {
	log.Debugf("Returning new %s acquirer", name)

	if name == GIT {
		if settings[URI] == "" || settings[BRANCH] == "" || settings[PATH] == "" {
			return nil, errors.New("Invalid git parameters. The uri, " +
				"branch and path are all mandatory.")
		}

		return NewGitAcquirer(settings[NAME], settings[URI], settings[BRANCH],
			settings[PATH]), nil
	}

	return nil, errors.New(fmt.Sprintf("Acquirer '%s' doesn't exist", name))
}

// Identifies the requirer based on its settings, and returns a new instance of it
func NewAcquirer(settings map[string]string) (Acquirer, error) {
	// perhaps the acquirer is explicitly declared in settings
	acquirer := settings[ACQUIRER_KEY]

	uri := settings[URI]

	if strings.Contains(uri, ".git") || acquirer == GIT {
		return acquirerFactory(GIT, settings)
	}

	return nil, errors.New(fmt.Sprintf("Couldn't identify acquirer for URI '%s'", uri))
}

// Delegate to an acquirer implementation
func Acquire(a Acquirer, dest string) error {
	return a.acquire(dest)
}
