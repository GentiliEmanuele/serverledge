package container

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/docker/docker/api/types/image"
)

type NodeDrivenPushV2 struct{}

func (ndv2 *NodeDrivenPushV2) PullImage(cf *DockerFactory, img string) (string, error) {
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
			return img, fmt.Errorf("could not pull image '%s': %v", img, err)
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

	// img != localImage means that the image was pulled from Docker Hub
	if img != localImage {
		// If the image was pulled from Docker hub require to the registry to pull the image
		go func() {
			err = pullRequest(localRegistryAddress, img)
			if err != nil {
				log.Printf("Error while communicating with local registry: %v\n", err)
			}
		}()
	}

	return img, nil
}

func pullRequest(localRegistryAddress, img string) error {
	// Format the URL using to invoke the API
	url := fmt.Sprintf("http://%s/pull-image", localRegistryAddress)

	// Define the payload as a map
	payload := map[string]string{"image": img}

	// Serialize data in json format
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Could not close http response\n")
		}
	}(resp.Body)

	return nil
}
