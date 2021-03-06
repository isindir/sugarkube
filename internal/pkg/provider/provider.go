package provider

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sugarkube/sugarkube/internal/pkg/kapp"
	"github.com/sugarkube/sugarkube/internal/pkg/log"
	"github.com/sugarkube/sugarkube/internal/pkg/vars"
	"os"
	"path/filepath"
)

const valuesFile = "values.yaml"

type Values = map[string]interface{}

type Provider interface {
	// Returns the variables loaded by the Provider
	getVars() Values
	// Associate provider variables with the provider
	setVars(Values)
	// Returns variables installers should pass on to kapps
	getInstallerVars() Values
	// Method that returns all paths in a config directory relevant to the
	// target profile/cluster/region, etc. that should be searched for values
	// files to merge.
	varsDirs(sc *kapp.StackConfig) ([]string, error)
}

// implemented providers
const LOCAL = "local"
const AWS = "aws"

// Factory that creates providers
func newProviderImpl(name string) (Provider, error) {
	if name == LOCAL {
		return &LocalProvider{}, nil
	}

	if name == AWS {
		return &AwsProvider{}, nil
	}

	return nil, errors.New(fmt.Sprintf("Provider '%s' doesn't exist", name))
}

// Instantiates a Provider and returns it along with the stack config vars it can
// load, or an error.
func NewProvider(stackConfig *kapp.StackConfig) (Provider, error) {
	providerImpl, err := newProviderImpl(stackConfig.Provider)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	stackConfigVars, err := stackConfigVars(providerImpl, stackConfig)
	if err != nil {
		log.Warn("Error loading stack config variables")
		return nil, errors.WithStack(err)
	}
	log.Debugf("Provider loaded vars: %#v", stackConfigVars)

	if len(stackConfigVars) == 0 {
		log.Fatal("No values loaded for stack")
		return nil, errors.New("Failed to load values for stack")
	}

	providerImpl.setVars(stackConfigVars)

	return providerImpl, nil
}

// Searches for values.yaml files in configured directories and returns the
// result of merging them.
func stackConfigVars(p Provider, sc *kapp.StackConfig) (Values, error) {
	stackConfigVars := Values{}

	varsDirs, err := p.varsDirs(sc)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for _, varFile := range varsDirs {
		valuePath := filepath.Join(varFile, valuesFile)

		_, err := os.Stat(valuePath)
		if err != nil {
			log.Debugf("Skipping merging non-existent path %s", valuePath)
			continue
		}

		err = vars.Merge(&stackConfigVars, valuePath)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return stackConfigVars, nil
}

// Return vars loaded from configs
func GetVars(p Provider) Values {
	return p.getVars()
}

// Return vars loaded from configs that should be passed on to kapps by Installers
func GetInstallerVars(p Provider) Values {
	return p.getInstallerVars()
}
