package gossiping

import (
	"fmt"
	"math/rand"

	sledgeConfig "github.com/serverledge-faas/serverledge/internal/config"
	"github.com/serverledge-faas/serverledge/internal/registration"
)

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
