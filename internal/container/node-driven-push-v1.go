package container

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api/types/image"
)

type NodeDrivenPushV1 struct{}

func (ndv1 *NodeDrivenPushV1) PullImage(cf *DockerFactory, img string) (string, error) {
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
		// If the image was pulled from Docker Hub, local registry hasn't the required image, or the local registry is not available.
		// Try to push the image to the local registry using a goroutine
		go func() {
			// First push the image retag it
			err := cf.cli.ImageTag(cf.ctx, img, localImage)
			if err != nil {
				log.Printf("Could not tag image '%s': %v", img, err)
				return
			}

			out, err := cf.cli.ImagePush(cf.ctx, localImage, image.PushOptions{})
			if err != nil {
				log.Printf("Could not push image '%s': %v", img, err)
				return
			}

			defer func(out io.ReadCloser) {
				if out != nil {
					err := out.Close()
					if err != nil {
						log.Printf("Could not close the docker image push response")
					}
				}
			}(out)

			_, err = io.Copy(os.Stdout, out)
			if err != nil {
				log.Printf("Could not copy the docker image push response")
			}
		}()
	}

	return img, nil
}
