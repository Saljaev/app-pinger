package mockqueue

import (
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/mock"
)

type MockRabbitMQ struct {
	mock.Mock
}

func (m *MockRabbitMQ) Consume() (<-chan amqp091.Delivery, error) {
	args := m.Mock.Called()
	return args.Get(0).(chan amqp091.Delivery), args.Error(1)
}

func (m *MockRabbitMQ) Publish(data interface{}) error {
	args := m.Called(data)
	return args.Error(0)
}

func (m *MockRabbitMQ) Close() {
	m.Called()
}
