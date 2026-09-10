package container

import (
	"context"
	"errors"
	"time"

	"github.com/serverledge-faas/serverledge/utils"
)

type Mechanism interface {
	PullImage(cf *DockerFactory, img string) (string, error)
}

var mechanismType Mechanism

func ChooseMechanism(conf string) {
	switch conf {
	case "node-driven-push-v1":
		mechanismType = &NodeDrivenPushV1{}
	case "node-driven-push-v2":
		mechanismType = &NodeDrivenPushV2{}
	case "recursive-push":
		mechanismType = &RecursivePush{}
	default:
		mechanismType = &BasePush{}
	}
}

func getLocalRegistryAddress() (string, error) {
	cli, err := utils.GetEtcdClient()
	if err != nil {
		return "", err
	}

	// Create a context with timer for the read
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := cli.Get(ctx, "registry")
	if err != nil {
		return "", err
	}

	// Check if the key exists
	if len(resp.Kvs) == 0 {
		return "", errors.New("no registry found")
	}

	// Return the key value
	return string(resp.Kvs[0].Value), nil
}
