package adapters

import (
	"context"
	"fmt"

	"github.com/0glabs/0g-storage-client/common/blockchain"
	"github.com/0glabs/0g-storage-client/core"
	"github.com/0glabs/0g-storage-client/indexer"
	"github.com/0glabs/0g-storage-client/node"
	"github.com/0glabs/0g-storage-client/transfer"
	"github.com/openweb3/web3go"
)

// ZeroGStore implements the Store interface for 0G storage
type ZeroGStore struct {
	w3client      *web3go.Client
	indexerClient *indexer.Client
	tempDir       string
	nodes         []*node.ZgsClient
	fileIndex     map[string]string // Maps file paths to their root hashes
	config        *ZeroGConfig
}

type ZeroGConfig struct {
	EVMRPC          string   `mapstructure:"evm_rpc"`
	IndexerRPC      string   `mapstructure:"indexer_rpc"`
	PrivateKey      string   `mapstructure:"private_key"`
	SegmentNumber   uint64   `mapstructure:"segment_number"`
	ExpectedReplica uint     `mapstructure:"expected_replica"`
	ExcludedNodes   []string `mapstructure:"excluded_nodes"`
	TempDir         string   `mapstructure:"temp_dir"`
}

// NewZeroGStore creates a new 0G storage adapter
func NewZeroGStore(config *ZeroGConfig) *ZeroGStore {
	return &ZeroGStore{
		tempDir:   config.TempDir,
		fileIndex: make(map[string]string),
	}
}

// Connect initializes connections to 0G network
func (s *ZeroGStore) Connect(ctx context.Context) error {
	// Initialize Web3 client
	w3client := blockchain.MustNewWeb3(s.config.EVMRPC, s.config.PrivateKey)
	s.w3client = w3client

	// Initialize indexer client
	indexerClient, err := indexer.NewClient(s.config.IndexerRPC)
	if err != nil {
		return fmt.Errorf("failed to create indexer client: %w", err)
	}
	s.indexerClient = indexerClient

	// Select initial nodes
	nodes, err := s.indexerClient.SelectNodes(ctx, s.config.SegmentNumber, s.config.ExpectedReplica, s.config.ExcludedNodes)
	if err != nil {
		return fmt.Errorf("failed to select nodes: %w", err)
	}
	s.nodes = nodes
	return nil
}

func (s *ZeroGStore) UploadFile(ctx context.Context, filePath string, data []byte) (string, string, error) {
	uploader, err := transfer.NewUploader(ctx, s.w3client, s.nodes)
	if err != nil {
		return "", "", fmt.Errorf("failed to create uploader: %w", err)
	}

	txHash, rootHash, err := uploader.UploadFile(ctx, filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to upload file: %w", err)
	}

	return txHash.String(), rootHash.String(), nil
}

func (s *ZeroGStore) DownloadFile(ctx context.Context, rootHash string, outputPath string, withProof bool) error {
	downloader, err := transfer.NewDownloader(s.nodes)
	if err != nil {
		return fmt.Errorf("failed to create downloader: %w", err)
	}

	err = downloader.Download(ctx, rootHash, outputPath, withProof)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	return nil
}

// calc file hash
func (s *ZeroGStore) calcFileHash(filePath string) (string, error) {
	rootHash, err := core.MerkleRoot(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to calculate file hash: %w", err)
	}
	return rootHash.String(), nil
}
