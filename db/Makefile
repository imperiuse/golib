.EXPORT_ALL_VARIABLES:

.PHONY: tests_all
tests_all: test_env_up tests test_env_down

.PHONY: test_env_up
test_env_up:
	@echo "Start test environment"
	@echo "1. Export test env variables and Run docker-compose test"
	docker-compose -f docker-compose.test.yml up -d
	sleep 5  # sleep driven development, pg init

.PHONY: test_env_down
test_env_down:
	@echo "Stop test environment"
	docker-compose -f docker-compose.test.yml down

.PHONY: tests
tests:
	@echo "Running integrations tests"
	go test -timeout 5m -race -short `go list ./... `

.PHONY: docker_clean_all
docker_clean_all:
	docker rm -f `docker ps -aq`
	docker volume rm -f `docker volume ls -q`
