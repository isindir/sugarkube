package provisioner

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sugarkube/sugarkube/internal/pkg/log"
	"github.com/sugarkube/sugarkube/internal/pkg/provider"
	"github.com/sugarkube/sugarkube/internal/pkg/provisioner/clustersot"
	"github.com/sugarkube/sugarkube/internal/pkg/vars"
	"time"
)

type Provisioner interface {
	// Returns the ClusterSot for this provisioner
	ClusterSot() (clustersot.ClusterSot, error)
	// Creates a cluster
	Create(sc *vars.StackConfig, values provider.Values, dryRun bool) error
	// Returns whether the cluster is already running
	IsAlreadyOnline(sc *vars.StackConfig, values provider.Values) (bool, error)
	// Update the cluster config if supported by the provisioner
	Update(sc *vars.StackConfig, values provider.Values) error
	// Wait for a cluster to become ready to install Kapps into
	WaitForClusterReadiness(sc *vars.StackConfig, values provider.Values) error
}

// key in Values that relates to this provisioner
const PROVISIONER_KEY = "provisioner"

// Implemented provisioner names
const MINIKUBE = "minikube"
const KOPS = "kops"

// Factory that creates providers
func NewProvisioner(name string) (Provisioner, error) {
	if name == MINIKUBE {
		return MinikubeProvisioner{}, nil
	}

	if name == KOPS {
		return KopsProvisioner{}, nil
	}

	return nil, errors.New(fmt.Sprintf("Provisioner '%s' doesn't exist", name))
}

// Creates a cluster using an implementation of a Provisioner
func Create(p Provisioner, sc *vars.StackConfig, values provider.Values, dryRun bool) error {
	return p.Create(sc, values, dryRun)
}

// Return whether the cluster is already online
func IsAlreadyOnline(p Provisioner, sc *vars.StackConfig, values provider.Values) (bool, error) {

	log.Infof("Checking whether cluster '%s' is already online...", sc.Cluster)

	online, err := p.IsAlreadyOnline(sc, values)
	if err != nil {
		return false, errors.WithStack(err)
	}

	if online {
		log.Infof("Cluster '%s' is online", sc.Cluster)
	} else {
		log.Infof("Cluster '%s' is not online", sc.Cluster)
	}

	sc.Status.IsOnline = online
	return online, nil
}

// Wait for a cluster to come online, then to become ready.
func WaitForClusterReadiness(p Provisioner, sc *vars.StackConfig, values provider.Values) error {
	clusterSot, err := p.ClusterSot()
	if err != nil {
		return errors.WithStack(err)
	}

	// only check whether the cluster is online if we started it, otherwise assume it is so
	// we don't wait pointlessly when resuming previous runs/working with existing clusters.
	if sc.Status.StartedThisRun {
		timeoutTime := time.Now().Add(time.Second * time.Duration(sc.OnlineTimeout))

		for time.Now().Before(timeoutTime) {
			online, err := clusterSot.IsOnline(sc, values)
			if err != nil {
				return errors.WithStack(err)
			}

			if online {
				log.Info("Cluster is online")
				break
			} else {
				time.Sleep(time.Duration(5) * time.Second)
			}
		}

		if !sc.Status.IsOnline {
			return errors.New("Timed out waiting for the cluster to come online")
		}

		// sleep for sc.Status.SleepBeforeReadyCheck seconds before proceeding
		sleepTime := sc.Status.SleepBeforeReadyCheck
		log.Infof("Sleeping for %d seconds before checking cluster readiness", sleepTime)
		time.Sleep(time.Second * time.Duration(sleepTime))
	}

	readinessTimeoutTime := time.Now().Add(time.Second * time.Duration(sc.OnlineTimeout))
	for time.Now().Before(readinessTimeoutTime) {
		ready, err := clusterSot.IsReady(sc, values)
		if err != nil {
			return errors.WithStack(err)
		}

		if ready {
			log.Info("Cluster is ready")
			break
		} else {
			time.Sleep(time.Duration(5) * time.Second)
		}
	}

	if !sc.Status.IsReady {
		return errors.New("Timed out waiting for the cluster to become ready")
	}

	return nil
}
