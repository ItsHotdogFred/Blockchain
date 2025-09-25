package network

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"time"
)

const (
	seedNodesFile = "./tmp/seed_nodes.json"
	discoveryPort = 3001 // Port for peer discovery
)

type SeedNode struct {
	Address   string    `json:"address"`
	LastSeen  time.Time `json:"last_seen"`
	IsHealthy bool      `json:"is_healthy"`
}

type PeerDiscovery struct {
	SeedNodes []SeedNode `json:"seed_nodes"`
}

func NewPeerDiscovery() *PeerDiscovery {
	return &PeerDiscovery{
		SeedNodes: []SeedNode{
			{Address: "158.178.141.60:3000", LastSeen: time.Now(), IsHealthy: true},
		},
	}
}

func (pd *PeerDiscovery) LoadSeedNodes() error {
	if _, err := os.Stat(seedNodesFile); os.IsNotExist(err) {
		return pd.SaveSeedNodes()
	}

	data, err := ioutil.ReadFile(seedNodesFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &pd.SeedNodes)
}

func (pd *PeerDiscovery) SaveSeedNodes() error {
	data, err := json.MarshalIndent(pd, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(seedNodesFile, data, 0644)
}

func (pd *PeerDiscovery) AddSeedNode(address string) {
	for _, node := range pd.SeedNodes {
		if node.Address == address {
			return
		}
	}
	pd.SeedNodes = append(pd.SeedNodes, SeedNode{
		Address:   address,
		LastSeen:  time.Now(),
		IsHealthy: true,
	})
	pd.SaveSeedNodes()
}

func (pd *PeerDiscovery) GetHealthyNodes() []string {
	var healthyNodes []string
	for _, node := range pd.SeedNodes {
		if node.IsHealthy {
			healthyNodes = append(healthyNodes, node.Address)
		}
	}
	return healthyNodes
}

func (pd *PeerDiscovery) CheckNodeHealth(address string) bool {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (pd *PeerDiscovery) UpdateKnownNodes() {
	var updatedNodes []SeedNode

	for _, node := range pd.SeedNodes {
		if pd.CheckNodeHealth(node.Address) {
			node.LastSeen = time.Now()
			node.IsHealthy = true
			updatedNodes = append(updatedNodes, node)
		} else {
			node.IsHealthy = false
			// Keep unhealthy nodes for a while before removing them
			if time.Since(node.LastSeen) < time.Hour*24 {
				updatedNodes = append(updatedNodes, node)
			}
		}
	}

	pd.SeedNodes = updatedNodes
	pd.SaveSeedNodes()
}

func GetRandomSeedNode() string {
	pd := NewPeerDiscovery()
	pd.LoadSeedNodes()
	pd.UpdateKnownNodes()

	healthyNodes := pd.GetHealthyNodes()
	if len(healthyNodes) == 0 {
		return "localhost:3000" // Fallback
	}

	rand.Seed(time.Now().UnixNano())
	return healthyNodes[rand.Intn(len(healthyNodes))]
}

func DiscoverPeers() []string {
	pd := NewPeerDiscovery()
	pd.LoadSeedNodes()
	pd.UpdateKnownNodes()

	// Start with healthy seed nodes
	peers := pd.GetHealthyNodes()

	// Ask each peer for their known peers
	for _, peer := range peers {
		additionalPeers := requestPeerList(peer)
		for _, additionalPeer := range additionalPeers {
			pd.AddSeedNode(additionalPeer)
		}
	}

	return pd.GetHealthyNodes()
}

func requestPeerList(address string) []string {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil
	}
	defer conn.Close()

	// Send peer discovery request
	request := CmdToBytes("getpeers")
	conn.Write(request)

	// Read response (simplified - in real implementation would handle response)
	return nil
}

func InitializeDecentralizedNetwork() []string {
	peers := DiscoverPeers()

	// Update global KnownNodes
	KnownNodes = peers

	return peers
}
