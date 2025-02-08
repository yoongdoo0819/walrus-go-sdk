package main

import (
	"encoding/json"
	"log"
	"net/http"

	walrus "github.com/namihq/walrus-go"

	"fmt"
)

type StoreReq struct {
	DigestArr             []string `json:"digestArr"`
	PartialDenseDigestArr []string `json:"partialDensesDigestArr"`
	VersionArr            []string `json:"versionArr"`
}

type StoreRes struct {
	Status string `json:"status"`
	BlobID string `json:"blobId"`
}

// "/store" Handler Func
func handleStore(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		var storeReq StoreReq
		if err := json.NewDecoder(r.Body).Decode(&storeReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		fmt.Println("Digest: ", storeReq.DigestArr)
		fmt.Println("PartialDenseDigest: ", storeReq.PartialDenseDigestArr)
		fmt.Println("Version: ", storeReq.VersionArr)
		blobId, err := storeWalrus(storeReq)
		res := StoreRes{
			BlobID: blobId,
		}
		if err != nil {
			res.Status = "failure"
			json.NewEncoder(w).Encode(res) // JSON Resp
		}
		res.Status = "success"
		json.NewEncoder(w).Encode(res) // JSON Resp
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func storeWalrus(storeReq StoreReq) (string, error) {
	// walrus
	wClient := walrus.NewClient(
		walrus.WithAggregatorURLs([]string{"https://aggregator.walrus-testnet.walrus.space"}),
		walrus.WithPublisherURLs([]string{"https://publisher.walrus-testnet.walrus.space"}),
	)

	// Store data
	jsonBytes, err := json.Marshal(storeReq)
	if err != nil {
		fmt.Println("JSON err:", err)
		return "", err
	}

	wData := []byte(jsonBytes)
	resp, err := wClient.Store(wData, &walrus.StoreOptions{Epochs: 10})

	if err != nil {
		log.Fatalf("Error storing data: %v", err)
		return "", err
	}

	var blobID string
	// Check response type and handle accordingly
	if resp.NewlyCreated != nil {
		blobID = resp.NewlyCreated.BlobObject.BlobID
		fmt.Printf("Stored new blob ID: %s with cost: %d\n",
			blobID, resp.NewlyCreated.Cost)
	} else if resp.AlreadyCertified != nil {
		blobID = resp.AlreadyCertified.BlobID
		fmt.Printf("Blob already exists with ID: %s, end epoch: %d\n",
			blobID, resp.AlreadyCertified.EndEpoch)
	}

	// Read data
	retrievedData, err := wClient.Read(blobID, nil)

	if err != nil {
		log.Fatalf("Error reading data: %v", err)
	}
	fmt.Printf("Retrieved data: %s\n", string(retrievedData))

	return blobID, nil
}

func main() {
	// Handler
	http.HandleFunc("/store", handleStore)

	fmt.Println("Server is running...")
	log.Fatal(http.ListenAndServe(":8083", nil))

}
