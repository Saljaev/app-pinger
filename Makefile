ENV_FILE=.env
CONFIG_DIR=backend/config
CONFIG_FILE=backend/config/verifier_config.yaml

.PHONY: prepare build start stop delete docs-start docs-stop docs-delete

prepare:
	echo "Создаю файл .env в $(ENV_FILE)"
	touch $(ENV_FILE)
	echo "#Frontend" > $(ENV_FILE)
	echo "REACT_APP_API_LOCATION=/api/v1/" >> $(ENV_FILE)
	echo "REACT_APP_REFRESH_INTERVAL=10" >> $(ENV_FILE)
	echo "REACT_APP_API_KEY=frontend-secret-key" >> $(ENV_FILE)
	echo "" >> $(ENV_FILE)
	echo "#Backend" >> $(ENV_FILE)
	echo "BACKEND_HOST=backend" >> $(ENV_FILE)
	echo "BACKEND_ADDRESS=0.0.0.0" >> $(ENV_FILE)
	echo "BACKEND_PORT=8082" >> $(ENV_FILE)
	echo "BACKEND_LOG_LEVEL=info" >> $(ENV_FILE)
	echo "TIMEOUT=4s" >> $(ENV_FILE)
	echo "IDLE_TIMEOUT=60s" >> $(ENV_FILE)
	echo "" >> $(ENV_FILE)
	echo "#Pinger service" >> $(ENV_FILE)
	echo "PINGER_HOST=pinger" >> $(ENV_FILE)
	echo "PINGER_LOG_LEVEL=info" >> $(ENV_FILE)
	echo "PINGER_PACKETS_COUNT=4" >> $(ENV_FILE)
	echo "PINGER_PING_TIMEOUT=5s" >> $(ENV_FILE)
	echo "PINGER_SVC_PING_TIMEOUT=15s" >> $(ENV_FILE)
	echo "" >> $(ENV_FILE)
	echo "#PostgreSQL DB" >> $(ENV_FILE)
	echo "PG_CONTAINER=db" >> $(ENV_FILE)
	echo "POSTGRES_USER=user1" >> $(ENV_FILE)
	echo "POSTGRES_PASSWORD=secret" >> $(ENV_FILE)
	echo "DB=db1" >> $(ENV_FILE)
	echo "" >> $(ENV_FILE)
	echo "#RabbitMQ" >> $(ENV_FILE)
	echo "RABBITMQ_USER=guest" >> $(ENV_FILE)
	echo "RABBITMQ_PASS=guest" >> $(ENV_FILE)
	echo "RABBITMQ_HOST=rabbitmq" >> $(ENV_FILE)
	echo "RABBITMQ_QUEUE=ping_results" >> $(ENV_FILE)
	echo "Файл .env создан успешно!"
	echo "Создаю файл verifier_config.yaml в $(CONFIG_FILE)"
	mkdir $(CONFIG_DIR)
	touch $(CONFIG_FILE)
	echo "rate_limit: 5" > $(CONFIG_FILE)
	echo "rate_time: 30s" >> $(CONFIG_FILE)
	echo "keys:" >> $(CONFIG_FILE)
	echo "  - frontend-secret-key" >> $(CONFIG_FILE)
	echo "Файл verifier_config.yaml создан успешно!"

build:
	docker compose build

start:
	docker compose up -d

stop:
	docker compose down

delete:
	docker compose down -v

docs-start:
	docker compose -f ./docs/docker-compose.yaml -p docs up -d

docs-stop:
	docker compose -f ./docs/docker-compose.yaml -p docs down

docs-delete:
	docker compose -f ./docs/docker-compose.yaml -p docs down -v

