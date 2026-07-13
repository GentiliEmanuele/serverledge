package gossiping

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/serverledge-faas/serverledge/internal/function"
	"github.com/serverledge-faas/serverledge/internal/registration"
)

func Gossip(f *function.Function) error {
	// Parse the topology
	nodeList, err := getTopology()
	if err != nil {
		return err
	}

	// Convert in JSON the function struct to be sent to the other nodes
	jsonData, err := json.Marshal(f)
	if err != nil {
		return err
	}

	for _, node := range *nodeList {
		_ = sendGossip(&node, jsonData)
	}

	return nil
}

func sendGossip(node *registration.NodeRegistration, jsonData []byte) error {
	// Format the request
	req, err := http.NewRequest("POST", getUrlFromNode(*node, "/update-remote"), bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print("Error closing body")
		}
	}(response.Body)

	if response.StatusCode != 200 {
		return fmt.Errorf("error while executing gossiping algorithm: %v", response.Status)
	}

	return nil
}
