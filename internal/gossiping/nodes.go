package gossiping

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	sledgeConfig "github.com/serverledge-faas/serverledge/internal/config"
	"github.com/serverledge-faas/serverledge/utils"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// populate add to the nodes list all the node registered in etcd
func (nodes *NodeList) populate() error {
	cli, err := utils.GetEtcdClient()
	if err != nil {
		return err
	}

	// Create a context with timer for the read
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := cli.Get(ctx, "", clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("error while reading etcd: %w", err)
	}

	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		value := string(kv.Value)

		// Parse the node information and add it to the list
		nodePtr, err := parseNode(key, value)
		if err != nil {
			return err
		}

		*nodes = append(*nodes, *nodePtr)
	}

	return nil
}

// getNodeSubList return a sub list that contains fanOut elements chosen randomly
func (nodes *NodeList) getNodeSubList() (*NodeList, error) {
	// Get the fan out parameter from configuration
	fanOut := sledgeConfig.GetInt(sledgeConfig.GOSSIPING_FAN_OUT, 1)

	// Check if the configured fan out is more than the size of the nodes list
	if fanOut > len(*nodes) {
		fanOut = len(*nodes)
	}

	// Create a copy of the nodes list
	nodesCopy := make(NodeList, len(*nodes))
	copy(nodesCopy, *nodes)

	// Shuffle the copy
	rand.Shuffle(len(nodesCopy), func(i, j int) {
		nodesCopy[i], nodesCopy[j] = nodesCopy[j], nodesCopy[i]
	})

	subList := nodesCopy[:fanOut]

	// Extract the first fanOut element
	return &subList, nil
}
