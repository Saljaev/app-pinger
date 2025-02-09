package service

import (
	"app-pinger/pkg/contracts"
	mockqueue "app-pinger/pkg/queue/mock"
	"fmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGoPinger_SendRequest(t *testing.T) {
	tests := []struct {
		name          string
		data          []contracts.PingData
		mockError     error
		expectedError string
	}{
		{
			name: "Valid send",
			data: []contracts.PingData{
				{
					IPAddress:   "192.168.1.1",
					IsReachable: true,
					LastPing:    time.Now().Format(time.DateTime),
				},
			},
			mockError:     nil,
			expectedError: "",
		},
		{
			name: "Invalid send (rabbitmq error)",
			data: []contracts.PingData{
				{
					IPAddress:   "192.168.1.1",
					IsReachable: true,
					LastPing:    time.Now().Format(time.DateTime),
				},
			},
			mockError:     fmt.Errorf("rabbitmq connection error"),
			expectedError: "failed to publish data: rabbitmq connection error",
		},
		{
			name: "Invalid send (not valid data - zero IP)",
			data: []contracts.PingData{
				{
					IPAddress:   "",
					IsReachable: true,
					LastPing:    time.Now().Format(time.DateTime),
				},
			},
			mockError:     nil,
			expectedError: "invalid request",
		},
		{
			name: "Invalid send (not valid data - zero time)",
			data: []contracts.PingData{
				{
					IPAddress:   "",
					IsReachable: true,
					LastPing:    "",
				},
			},
			mockError:     nil,
			expectedError: "invalid request",
		},
		{
			name: "Invalid send (not valid data - not correct time)",
			data: []contracts.PingData{
				{
					IPAddress:   "",
					IsReachable: true,
					LastPing:    "1000-10-10",
				},
			},
			mockError:     nil,
			expectedError: "invalid request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRabbit := &mockqueue.MockRabbitMQ{}
			pinger := &GoPinger{rabbitMQ: mockRabbit}

			req := contracts.ContainerAddReq{Containers: tt.data}

			if req.IsValid() {
				mockRabbit.On("Publish", mock.Anything).Return(tt.mockError)
			}

			err := pinger.SendRequest(tt.data)

			if tt.expectedError == "" {
				require.NoError(t, err)
				mockRabbit.AssertCalled(t, "Publish", mock.Anything)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)

				if req.IsValid() {
					mockRabbit.AssertCalled(t, "Publish", mock.Anything)
				} else {
					mockRabbit.AssertNotCalled(t, "Publish", mock.Anything)
				}
			}

			mockRabbit.AssertExpectations(t)
		})
	}
}
