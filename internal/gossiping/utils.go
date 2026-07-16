package gossiping

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	sledgeConfig "github.com/serverledge-faas/serverledge/internal/config"
	"github.com/serverledge-faas/serverledge/internal/node"
	"github.com/serverledge-faas/serverledge/internal/registration"
)

// parseNode return the node infos from the etcd data
func parseNode(key, value string) (*registration.NodeRegistration, error) {
	// This is the pattern that the serverledge node key must be respect
	pattern := `registry/[A-Za-z0-9_-]+/[A-Za-z0-9_-]+/[A-Za-z0-9_-]+`
	// Compile the regex
	re := regexp.MustCompile(pattern)

	if re.MatchString(key) {
		splitKey := strings.Split(key, "/")
		splitValue := strings.Split(value, ";")

		// Prevent a node from sending a message to itself
		if node.LocalNode.Key != splitKey[3] {
			newNodeId := node.NodeID{
				Area: splitKey[1],
				Key:  splitKey[3],
				Arch: splitValue[2],
			}

			apiPort, err := strconv.Atoi(splitValue[1])
			if err != nil {
				return nil, err
			}

			udpPort, err := strconv.Atoi(splitValue[2])
			if err != nil {
				return nil, err
			}

			newNode := registration.NodeRegistration{
				NodeID:         newNodeId,
				IPAddress:      splitValue[0],
				APIPort:        apiPort,
				UDPPort:        udpPort,
				IsLoadBalancer: false,
			}

			return &newNode, nil
		}
	}

	return nil, fmt.Errorf("invalid node key: %s", key)
}

// getUrlFromNode return the URL of the specified node
func getUrlFromNode(node registration.NodeRegistration, route string) string {
	return fmt.Sprintf("http://%s:%d%s", node.IPAddress, node.APIPort, route)
}

// getNodesRandomly return a sub list that contains fanOut elements chosen randomly
func getNodesRandomly(neighborInfo map[string]*registration.StatusInformation) map[string]*registration.StatusInformation {
	// Get the fan out parameter from configuration
	fanOut := sledgeConfig.GetInt(sledgeConfig.GOSSIPING_FAN_OUT, 1)

	// Check if the configured fan out is more than the size of the nodes list
	if fanOut > len(neighborInfo) {
		fanOut = len(neighborInfo)
	}

	// Extract keys
	keys := make([]string, 0, len(neighborInfo))
	for key := range neighborInfo {
		keys = append(keys, key)
	}

	// Shuffle keys
	rand.Shuffle(len(keys), func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})

	randNodes := make(map[string]*registration.StatusInformation, fanOut)
	for _, key := range keys[:fanOut] {
		randNodes[key] = neighborInfo[key]
	}

	return randNodes
}
