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
var neighborInfo map[string]*registration.StatusInformation

// Gossiping build the topology and send a request to some node (casually)
func Gossiping(receivedRequest Request) error {
	// Get the neighbor info
	neighborInfo = registration.GetFullNeighborInfo()

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

	randNodes := getNodesRandomly(neighborInfo)

	for key, _ := range randNodes {
		node := registration.GetPeerFromKey(key)
		if node == nil {
			continue
		}
		log.Printf("Sending request to %s\n", node.IPAddress)
		err = sendGossipMessage(node, jsonData)
		if err != nil {
			fmt.Printf("Error while sending a gossip to %s: %v\n", node.IPAddress, err)
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
			fmt.Print("Error closing body\n")
		}
	}(response.Body)

	if response.StatusCode != 200 {
		return fmt.Errorf("error %d %v\n", response.StatusCode, response.Status)
	}

	return nil
}
