package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"solana/src/models"

	"github.com/valyala/fasthttp"
)

type SolanaClient struct {
	rpcURL    string
	cache     *sync.Map
	rateLimit *time.Ticker
}

func NewSolanaClient(rpcURL string) *SolanaClient {
	return &SolanaClient{
		rpcURL:    rpcURL,
		cache:     &sync.Map{},
		rateLimit: time.NewTicker(time.Millisecond * 100),
	}
}

func (c *SolanaClient) GetLatestSlot() (uint64, error) {
	<-c.rateLimit.C

	reqBody := `{
        "jsonrpc": "2.0",
        "id": 1,
        "method": "getSlot",
        "params": [{"commitment": "finalized"}]
    }`

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(c.rpcURL)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	req.SetBodyString(reqBody)

	if err := fasthttp.Do(req, resp); err != nil {
		return 0, fmt.Errorf("failed to send request: %v", err)
	}

	var result struct {
		Result uint64 `json:"result"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return 0, fmt.Errorf("failed to parse response: %v", err)
	}

	return result.Result, nil
}

func (c *SolanaClient) GetBlock(slot uint64) (*models.BlockResponse, error) {
	<-c.rateLimit.C

	// 檢查緩存中是否為空槽
	if _, ok := c.cache.Load(slot); ok {
		return nil, fmt.Errorf("empty slot")
	}

	reqBody := fmt.Sprintf(`{
        "jsonrpc": "2.0",
        "id": 1,
        "method": "getBlock",
        "params": [
            %d,
            {
                "encoding": "json",
                "transactionDetails": "full",
                "rewards": false,
                "maxSupportedTransactionVersion": 0
            }
        ]
    }`, slot)

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(c.rpcURL)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	req.SetBodyString(reqBody)

	if err := fasthttp.Do(req, resp); err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	var blockResponse models.BlockResponse
	if err := json.Unmarshal(resp.Body(), &blockResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &blockResponse, nil
}

func (c *SolanaClient) CheckSlotStatus(slot uint64) (string, error) {
	<-c.rateLimit.C

	reqBody := fmt.Sprintf(`{
        "jsonrpc": "2.0",
        "id": 1,
        "method": "getBlock",
        "params": [
            %d,
            {
                "encoding": "json",
                "transactionDetails": "none",
                "rewards": false
            }
        ]
    }`, slot)

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(c.rpcURL)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	req.SetBodyString(reqBody)

	if err := fasthttp.Do(req, resp); err != nil {
		return "NOT_AVAILABLE", err
	}

	var result struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return "NOT_AVAILABLE", err
	}

	if result.Error.Code == -32004 ||
		strings.Contains(result.Error.Message, "Block not available") {
		// 將空槽加入緩存
		c.cache.Store(slot, true)
		return "EMPTY", nil
	}

	return "CONFIRMED", nil
}
