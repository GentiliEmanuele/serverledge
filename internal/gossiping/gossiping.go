package gossiping

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/serverledge-faas/serverledge/internal/registration"
)

var requests *Requests

// Gossiping build the topology and send a request to some node (casually)
func Gossiping(receivedRequest Request) error {
	// Parse the topology
	nodeList, err := getTopology()
	if err != nil {
		return err
	}

	// If the requests list is nil create it
	if requests == nil {
		requests = &Requests{}
	}

	// Check if the requests list contain the received request
	if requests.find(receivedRequest.F, receivedRequest.Timestamp) {
		log.Print("Request already processed\n")
		return nil
	}

	// If the request has not been already processed save it into the list and send it tho others causal nodes
	requests.add(receivedRequest.F, receivedRequest.Timestamp)

	// Convert in JSON the function struct to be sent to the other nodes
	jsonData, err := json.Marshal(receivedRequest)
	if err != nil {
		return err
	}

	for _, node := range *nodeList {
		// Choose casually if you use or not the current node
		if !useCurrentNode() {
			log.Printf("Skipping node %s\n", node.IPAddress)
			continue
		}

		log.Printf("Sending request to %s\n", node.IPAddress)
		err = sendGossipMessage(&node, jsonData)
		if err != nil {
			fmt.Printf("Error while sending a gossip to %s", node.IPAddress)
		}
	}

	return nil
}

// sendGossipMessage send the given json data to the specified node
func sendGossipMessage(node *registration.NodeRegistration, jsonData []byte) error {
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
