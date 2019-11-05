package client

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

type storeFailer struct{ error }

func (s storeFailer) SaveSubscription(arg *store.Subscription) error {
	return s.error
}

func generateCreateSubscriptionReq(id, chain, endpoint string, addresses, topics []string) CreateSubscriptionReq {
	config := struct {
		Endpoint   string `json:"endpoint"`
		ChainId    string `json:"chainId"`
		RefreshInt int    `json:"refreshInterval"`
	}{
		Endpoint: endpoint,
	}
	params := struct {
		Type   string `json:"type"`
		Config struct {
			Endpoint   string `json:"endpoint"`
			ChainId    string `json:"chainId"`
			RefreshInt int    `json:"refreshInterval"`
		} `json:"config"`
		Addresses []string `json:"addresses"`
		Topics    []string `json:"topics"`
	}{
		Type:      chain,
		Config:    config,
		Addresses: addresses,
		Topics:    topics,
	}

	return CreateSubscriptionReq{
		JobID:  id,
		Type:   "external",
		Params: params,
	}
}

func TestConfigController(t *testing.T) {
	tests := []struct {
		Name       string
		Payload    interface{}
		App        subscriptionStorer
		StatusCode int
	}{
		{
			"Create success",
			generateCreateSubscriptionReq("id", "ethereum", "http://localhost:6688", []string{"0x123"}, []string{"0x123"}),
			storeFailer{nil},
			http.StatusCreated,
		},
		{
			"Decode failed",
			"bad json format",
			storeFailer{errors.New("failed save")},
			http.StatusBadRequest,
		},
		{
			"Save failed",
			generateCreateSubscriptionReq("id", "ethereum", "http://localhost:6688", []string{"0x123"}, []string{"0x123"}),
			storeFailer{errors.New("failed save")},
			http.StatusInternalServerError,
		},
	}
	for _, test := range tests {
		t.Log(test.Name)
		body, err := json.Marshal(test.Payload)
		require.NoError(t, err)

		srv := &httpService{
			store: test.App,
		}
		srv.createRouter()

		req := httptest.NewRequest("POST", "/job", bytes.NewBuffer(body))

		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		assert.Equal(t, test.StatusCode, w.Code)

		var respJSON map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &respJSON)
		assert.NoError(t, err)
	}
}

func TestHealthController(t *testing.T) {
	tests := []struct {
		Name       string
		StatusCode int
	}{
		{
			"Is healthy",
			http.StatusOK,
		},
	}
	for _, test := range tests {
		srv := &httpService{}
		srv.createRouter()

		req := httptest.NewRequest("GET", "/health", nil)

		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		assert.Equal(t, test.StatusCode, w.Code)

		var respJSON map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &respJSON)
		assert.NoError(t, err)
	}
}