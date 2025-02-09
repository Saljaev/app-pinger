package containershandler

import (
	"app-pinger/backend/internal/api/utilapi"
	"app-pinger/backend/internal/entity"
	storagemock "app-pinger/backend/internal/usecase/repo/mock"
	"app-pinger/pkg/contracts"
	mockqueue "app-pinger/pkg/queue/mock"
	"bytes"
	"encoding/json"
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestContainersHandler_GetAll(t *testing.T) {
	tests := []struct {
		name string
		want interface{}
	}{
		{
			name: "Valid",
			want: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := storagemock.NewMockRepo(entity.Container{
				IP:          "192.168.0.1",
				IsReachable: true,
				LastPing:    time.Now(),
			})
			mockRabbit := new(mockqueue.MockRabbitMQ)
			h := NewContainersHandler(mockRepo, mockRabbit)

			r := utilapi.NewRouter(slog.Default())
			r.Handle("/", h.GetAll)

			req := httptest.NewRequest(http.MethodGet, "/", nil)

			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.want {
				t.Errorf("expected: %v get: %v", tt.want, w.Code)
			}
		})
	}
}

func TestContainersHandler_ProcessQueue(t *testing.T) {
	tests := []struct {
		name      string
		container contracts.PingData
		want      interface{}
	}{
		{
			name: "Valid container",
			container: contracts.PingData{
				IPAddress:   "192.168.1.1",
				IsReachable: true,
				LastPing:    time.Now().Format(time.DateTime),
			},
			want: "",
		},
		{
			name: "Invalid container (IP)",
			container: contracts.PingData{
				IPAddress:   "",
				IsReachable: true,
				LastPing:    time.Now().Format(time.DateTime),
			},
			want: "failed decode json",
		},
		{
			name: "Invalid container (Last ping - not data)",
			container: contracts.PingData{
				IPAddress:   "",
				IsReachable: true,
				LastPing:    "1000-10-10",
			},
			want: "failed decode json",
		},
		{
			name: "Invalid container (Last ping - zero data)",
			container: contracts.PingData{
				IPAddress:   "",
				IsReachable: true,
				LastPing:    "",
			},
			want: "failed decode json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := storagemock.NewMockRepo(entity.Container{
				IP:          "192.168.0.1",
				IsReachable: true,
				LastPing:    time.Now(),
			})
			mockRabbit := new(mockqueue.MockRabbitMQ)
			h := NewContainersHandler(mockRepo, mockRabbit)

			testReq := contracts.ContainerAddReq{
				Containers: []contracts.PingData{
					tt.container,
				},
			}

			body, _ := json.Marshal(testReq)
			msgChan := make(chan amqp091.Delivery, 1)
			msgChan <- amqp091.Delivery{Body: body}
			close(msgChan)

			mockRabbit.On("Consume").Return(msgChan, nil)

			var logBuffer bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

			h.ProcessQueue(logger)

			mockRabbit.AssertExpectations(t)

			if tt.want != "" {
				require.Contains(t, logBuffer.String(), tt.want, "Expected log error not found")
			}
		})
	}
}
