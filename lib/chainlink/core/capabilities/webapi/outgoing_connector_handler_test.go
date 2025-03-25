package webapi

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink/v2/core/logger"
	"github.com/smartcontractkit/chainlink/v2/core/services/gateway/api"
	gcmocks "github.com/smartcontractkit/chainlink/v2/core/services/gateway/connector/mocks"
	ghcapabilities "github.com/smartcontractkit/chainlink/v2/core/services/gateway/handlers/capabilities"
	"github.com/smartcontractkit/chainlink/v2/core/services/gateway/handlers/common"
	"github.com/smartcontractkit/chainlink/v2/core/utils/matches"
)

func TestHandleSingleNodeRequest(t *testing.T) {
	t.Run("uses default timeout if no timeout is provided", func(t *testing.T) {
		ctx := tests.Context(t)
		msgID := "msgID"
		testURL := "http://localhost:8080"
		connector, connectorHandler := newFunctionWithDefaultConfig(
			t,
			func(gc *gcmocks.GatewayConnector) {
				gc.EXPECT().DonID().Return("donID")
				gc.EXPECT().AwaitConnection(matches.AnyContext, "gateway1").Return(nil)
				gc.EXPECT().GatewayIDs().Return([]string{"gateway1"})
			},
		)

		// build the expected body with the default timeout
		req := ghcapabilities.Request{
			URL:       testURL,
			TimeoutMs: defaultFetchTimeoutMs,
		}
		payload, err := json.Marshal(req)
		require.NoError(t, err)

		expectedBody := &api.MessageBody{
			MessageId: msgID,
			DonId:     connector.DonID(),
			Method:    ghcapabilities.MethodComputeAction,
			Payload:   payload,
		}

		// expect the request body to contain the default timeout
		connector.EXPECT().SignAndSendToGateway(mock.Anything, "gateway1", expectedBody).Run(func(ctx context.Context, gatewayID string, msg *api.MessageBody) {
			connectorHandler.HandleGatewayMessage(ctx, "gateway1", gatewayResponse(t, msgID))
		}).Return(nil).Times(1)

		_, err = connectorHandler.HandleSingleNodeRequest(ctx, msgID, ghcapabilities.Request{
			URL: testURL,
		})
		require.NoError(t, err)
	})

	t.Run("subsequent request uses gateway 2", func(t *testing.T) {
		ctx := tests.Context(t)
		msgID := "msgID"
		testURL := "http://localhost:8080"
		connector, connectorHandler := newFunctionWithDefaultConfig(
			t,
			func(gc *gcmocks.GatewayConnector) {
				gc.EXPECT().DonID().Return("donID")
				gc.EXPECT().AwaitConnection(matches.AnyContext, "gateway1").Return(nil).Once()
				gc.EXPECT().AwaitConnection(matches.AnyContext, "gateway2").Return(nil).Once()
				gc.EXPECT().GatewayIDs().Return([]string{"gateway1", "gateway2"})
			},
		)

		// build the expected body with the default timeout
		req := ghcapabilities.Request{
			URL:       testURL,
			TimeoutMs: defaultFetchTimeoutMs,
		}
		payload, err := json.Marshal(req)
		require.NoError(t, err)

		expectedBody := &api.MessageBody{
			MessageId: msgID,
			DonId:     connector.DonID(),
			Method:    ghcapabilities.MethodComputeAction,
			Payload:   payload,
		}

		// expect call to be made to gateway 1
		connector.EXPECT().SignAndSendToGateway(mock.Anything, "gateway1", expectedBody).Run(func(ctx context.Context, gatewayID string, msg *api.MessageBody) {
			connectorHandler.HandleGatewayMessage(ctx, "gateway1", gatewayResponse(t, msgID))
		}).Return(nil).Times(1)

		_, err = connectorHandler.HandleSingleNodeRequest(ctx, msgID, ghcapabilities.Request{
			URL: testURL,
		})
		require.NoError(t, err)

		// expect call to be made to gateway 2
		connector.EXPECT().SignAndSendToGateway(mock.Anything, "gateway2", expectedBody).Run(func(ctx context.Context, gatewayID string, msg *api.MessageBody) {
			connectorHandler.HandleGatewayMessage(ctx, "gateway2", gatewayResponse(t, msgID))
		}).Return(nil).Times(1)

		_, err = connectorHandler.HandleSingleNodeRequest(ctx, msgID, ghcapabilities.Request{
			URL: testURL,
		})
		require.NoError(t, err)
	})

	t.Run("uses timeout", func(t *testing.T) {
		ctx := tests.Context(t)
		msgID := "msgID"
		testURL := "http://localhost:8080"
		connector, connectorHandler := newFunctionWithDefaultConfig(
			t,
			func(gc *gcmocks.GatewayConnector) {
				gc.EXPECT().DonID().Return("donID")
				gc.EXPECT().AwaitConnection(matches.AnyContext, "gateway1").Return(nil)
				gc.EXPECT().GatewayIDs().Return([]string{"gateway1"})
			},
		)

		// build the expected body with the defined timeout
		req := ghcapabilities.Request{
			URL:       testURL,
			TimeoutMs: 40000,
		}
		payload, err := json.Marshal(req)
		require.NoError(t, err)

		expectedBody := &api.MessageBody{
			MessageId: msgID,
			DonId:     connector.DonID(),
			Method:    ghcapabilities.MethodComputeAction,
			Payload:   payload,
		}

		// expect the request body to contain the defined timeout
		connector.EXPECT().SignAndSendToGateway(mock.Anything, "gateway1", expectedBody).Run(func(ctx context.Context, gatewayID string, msg *api.MessageBody) {
			connectorHandler.HandleGatewayMessage(ctx, "gateway1", gatewayResponse(t, msgID))
		}).Return(nil).Times(1)

		_, err = connectorHandler.HandleSingleNodeRequest(ctx, msgID, ghcapabilities.Request{
			URL:       testURL,
			TimeoutMs: 40000,
		})
		_, found := connectorHandler.responses.get(msgID)
		assert.False(t, found)
		require.NoError(t, err)
	})

	t.Run("cleans up in event of a timeout", func(t *testing.T) {
		ctx := tests.Context(t)
		msgID := "msgID"
		testURL := "http://localhost:8080"
		connector, connectorHandler := newFunctionWithDefaultConfig(
			t,
			func(gc *gcmocks.GatewayConnector) {
				gc.EXPECT().DonID().Return("donID")
				gc.EXPECT().AwaitConnection(matches.AnyContext, "gateway1").Return(nil)
				gc.EXPECT().GatewayIDs().Return([]string{"gateway1"})
			},
		)

		// build the expected body with the defined timeout
		req := ghcapabilities.Request{
			URL:       testURL,
			TimeoutMs: 10,
		}
		payload, err := json.Marshal(req)
		require.NoError(t, err)

		expectedBody := &api.MessageBody{
			MessageId: msgID,
			DonId:     connector.DonID(),
			Method:    ghcapabilities.MethodComputeAction,
			Payload:   payload,
		}

		// expect the request body to contain the defined timeout
		connector.EXPECT().SignAndSendToGateway(mock.Anything, "gateway1", expectedBody).Run(func(ctx context.Context, gatewayID string, msg *api.MessageBody) {
			// don't call HandleGatewayMessage here; i.e. simulate a failure to receive a response
		}).Return(nil).Times(1)

		_, err = connectorHandler.HandleSingleNodeRequest(ctx, msgID, ghcapabilities.Request{
			URL:       testURL,
			TimeoutMs: 10,
		})
		_, found := connectorHandler.responses.get(msgID)
		assert.False(t, found)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("rate limits outgoing traffic", func(t *testing.T) {
		ctx := tests.Context(t)
		msgID := "msgID"
		testURL := "http://localhost:8080"
		var config = ServiceConfig{
			OutgoingRateLimiter: common.RateLimiterConfig{
				GlobalRPS:      2.0,
				GlobalBurst:    2,
				PerSenderRPS:   1.0,
				PerSenderBurst: 1,
			},
			RateLimiter: common.RateLimiterConfig{
				GlobalRPS:      100.0,
				GlobalBurst:    100,
				PerSenderRPS:   100.0,
				PerSenderBurst: 100,
			},
		}
		connector, connectorHandler := newFunction(
			t,
			func(gc *gcmocks.GatewayConnector) {
				gc.EXPECT().DonID().Return("donID")
				gc.EXPECT().AwaitConnection(matches.AnyContext, "gateway1").Return(nil)
				gc.EXPECT().GatewayIDs().Return([]string{"gateway1"})
			},
			config,
		)

		// build the expected body with the default timeout
		req := ghcapabilities.Request{
			URL:        testURL,
			WorkflowID: "1",
			TimeoutMs:  defaultFetchTimeoutMs,
		}
		payload, err := json.Marshal(req)
		require.NoError(t, err)

		expectedBody := &api.MessageBody{
			MessageId: msgID,
			DonId:     connector.DonID(),
			Method:    ghcapabilities.MethodComputeAction,
			Payload:   payload,
		}

		// expect the request body to contain the default timeout
		connector.EXPECT().SignAndSendToGateway(mock.Anything, "gateway1", expectedBody).Run(func(ctx context.Context, gatewayID string, msg *api.MessageBody) {
			connectorHandler.HandleGatewayMessage(ctx, "gateway1", gatewayResponse(t, msgID))
		}).Return(nil).Times(1)

		_, err = connectorHandler.HandleSingleNodeRequest(ctx, msgID, ghcapabilities.Request{
			URL:        testURL,
			WorkflowID: "1",
		})
		require.NoError(t, err)

		// Second request should error from workflow ratelimit
		_, err = connectorHandler.HandleSingleNodeRequest(ctx, msgID, ghcapabilities.Request{
			URL:        testURL,
			WorkflowID: "1",
		})
		require.Error(t, err)
		require.ErrorContains(t, err, errorOutgoingRatelimitWorkflow)

		// Third request should error from global ratelimit
		_, err = connectorHandler.HandleSingleNodeRequest(ctx, msgID, ghcapabilities.Request{
			URL:        testURL,
			WorkflowID: "2",
		})
		require.Error(t, err)
		require.ErrorContains(t, err, errorOutgoingRatelimitGlobal)
	})
}

func newFunctionWithDefaultConfig(t *testing.T, mockFn func(*gcmocks.GatewayConnector)) (*gcmocks.GatewayConnector, *OutgoingConnectorHandler) {
	var defaultConfig = ServiceConfig{
		OutgoingRateLimiter: common.RateLimiterConfig{
			GlobalRPS:      100.0,
			GlobalBurst:    100,
			PerSenderRPS:   100.0,
			PerSenderBurst: 100,
		},
		RateLimiter: common.RateLimiterConfig{
			GlobalRPS:      100.0,
			GlobalBurst:    100,
			PerSenderRPS:   100.0,
			PerSenderBurst: 100,
		},
	}
	return newFunction(t, mockFn, defaultConfig)
}

func newFunction(t *testing.T, mockFn func(*gcmocks.GatewayConnector), serviceConfig ServiceConfig) (*gcmocks.GatewayConnector, *OutgoingConnectorHandler) {
	log := logger.TestLogger(t)
	connector := gcmocks.NewGatewayConnector(t)

	mockFn(connector)

	connectorHandler, err := NewOutgoingConnectorHandler(connector, serviceConfig, ghcapabilities.MethodComputeAction, log)
	require.NoError(t, err)
	return connector, connectorHandler
}

func gatewayResponse(t *testing.T, msgID string) *api.Message {
	headers := map[string]string{"Content-Type": "application/json"}
	body := []byte("response body")
	responsePayload, err := json.Marshal(ghcapabilities.Response{
		StatusCode:     200,
		Headers:        headers,
		Body:           body,
		ExecutionError: false,
	})
	require.NoError(t, err)
	return &api.Message{
		Body: api.MessageBody{
			MessageId: msgID,
			Method:    ghcapabilities.MethodWebAPITarget,
			Payload:   responsePayload,
		},
	}
}

func TestServiceConfigDefaults(t *testing.T) {
	t.Run("fills default RateLimiterConfigs", func(t *testing.T) {
		var cfg ServiceConfig

		tomlErr := toml.Unmarshal([]byte{}, &cfg)
		require.NoError(t, tomlErr)

		iRLConf := incomingRateLimiterConfigDefaults(cfg.RateLimiter)
		require.Equal(t, DefaultGlobalBurst, iRLConf.GlobalBurst)
		require.InDelta(t, DefaultGlobalRPS, iRLConf.GlobalRPS, 0.001)
		require.Equal(t, DefaultPerSenderBurst, iRLConf.PerSenderBurst)
		require.InDelta(t, DefaultPerSenderRPS, iRLConf.PerSenderRPS, 0.001)

		oRLConf := outgoingRateLimiterConfigDefaults(cfg.RateLimiter)
		require.Equal(t, DefaultGlobalBurst, oRLConf.GlobalBurst)
		require.InDelta(t, DefaultGlobalRPS, oRLConf.GlobalRPS, 0.001)
		require.Equal(t, DefaultWorkflowBurst, oRLConf.PerSenderBurst)
		require.InDelta(t, DefaultWorkflowRPS, oRLConf.PerSenderRPS, 0.001)
	})
}
