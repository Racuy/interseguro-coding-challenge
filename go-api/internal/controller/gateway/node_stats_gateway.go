// HTTP client for calling node-api
package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go-api/internal/application"
	"go-api/internal/domain"
)

// endpoint node-api expects the QR result on
const statsPath = "/api/v1/stats"

type NodeStatsGateway struct {
	baseURL string
	client  *http.Client
}

// no client timeout on purpose, ctx controls the deadline
func NewNodeStatsGateway(baseURL string) *NodeStatsGateway {
	return &NodeStatsGateway{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// posts Q/R to node-api, forwards the token if present
func (g *NodeStatsGateway) SendForStats(ctx context.Context, result domain.QRFactorization) (domain.MatrixStats, error) {
	body, err := json.Marshal(result)
	if err != nil {
		return domain.MatrixStats{}, fmt.Errorf("gateway: encode QR result: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+statsPath, bytes.NewReader(body))
	if err != nil {
		return domain.MatrixStats{}, fmt.Errorf("gateway: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token, ok := application.TokenFromContext(ctx); ok {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return domain.MatrixStats{}, fmt.Errorf("gateway: call node-api: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.MatrixStats{}, fmt.Errorf("gateway: read response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return domain.MatrixStats{}, fmt.Errorf("gateway: node-api returned %d: %s", resp.StatusCode, respBody)
	}

	var stats domain.MatrixStats
	if err := json.Unmarshal(respBody, &stats); err != nil {
		return domain.MatrixStats{}, fmt.Errorf("gateway: decode response: %w", err)
	}

	return stats, nil
}
