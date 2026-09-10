package container

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/docker/docker/api/types/image"
)

type BasePush struct{}

func (bp *BasePush) PullImage(cf *DockerFactory, img string) (string, error) {
	// Get registry info from etcd
	localRegistryAddress, err := getLocalRegistryAddress()
	if err != nil {
		fmt.Printf("No registry found")
	}
	// Format the name for local image using local registry address
	localImage := strings.Join([]string{localRegistryAddress, img}, "/")

	// Try to pull first from local registry
	pullResp, err := cf.cli.ImagePull(cf.ctx, localImage, image.PullOptions{})
	if err != nil {
		// If an error occur try to pull from remote registry
		pullResp, err = cf.cli.ImagePull(cf.ctx, img, image.PullOptions{})
		if err != nil {
			return img, fmt.Errorf("Could not pull image '%s': %v", img, err)
		}

		// If the image was pulled from Docker Hub the image is the latter specified in runtime
		// In this case is not needed to tag the images
	} else {
		img = localImage
	}

	defer func(pullResp io.ReadCloser) {
		err := pullResp.Close()
		if err != nil {
			log.Printf("Could not close the docker image pull response\n")
		}
	}(pullResp)

	// This seems to be necessary to wait for the img to be pulled:
	_, _ = io.Copy(io.Discard, pullResp)
	log.Printf("Pulled image: %s\n", img)
	refreshedImages[img] = true

	return img, nil
}
