package gossiping

import (
	"context"
	"fmt"
	"math/rand/v2"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/serverledge-faas/serverledge/internal/function"
	"github.com/serverledge-faas/serverledge/internal/node"
	"github.com/serverledge-faas/serverledge/internal/registration"
	"github.com/serverledge-faas/serverledge/utils"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// getTopology contact the Etcd server and read all the nodes' information
func getTopology() (*NodeList, error) {
	cli, err := utils.GetEtcdClient()
	if err != nil {
		return nil, err
	}

	// Create a context with timer for the read
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := cli.Get(ctx, "", clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("error while reading etcd: %w", err)
	}

	var nodeList NodeList

	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		value := string(kv.Value)

		// Parse the node information and add it to the list
		nodePtr, err := parseNode(key, value)
		if err == nil {
			nodeList = append(nodeList, *nodePtr)
		}
	}

	return &nodeList, nil
}

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

// useCurrentNode return true the current node must be used for this gossiping message
func useCurrentNode() bool {
	return rand.IntN(2) == 1
}

// trackRequest save the request with the timestamp of updating request.
func trackRequest(f function.Function, timestamp time.Time) {

}
