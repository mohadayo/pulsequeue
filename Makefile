.PHONY: up down build test test-python test-go test-ts lint clean logs health

up:
	docker compose up -d --build

down:
	docker compose down

build:
	docker compose build

test: test-python test-go test-ts

test-python:
	cd api-gateway && pip install -r requirements.txt -q && pytest -v

test-go:
	cd worker-engine && go test -v ./...

test-ts:
	cd dashboard && npm install && npm test

lint: lint-python lint-ts

lint-python:
	cd api-gateway && pip install flake8 -q && flake8 app.py --max-line-length=120

lint-ts:
	cd dashboard && npm install && npx eslint src/

logs:
	docker compose logs -f

health:
	@echo "API Gateway:" && curl -s http://localhost:8080/health | python3 -m json.tool
	@echo "Worker Engine:" && curl -s http://localhost:8081/health | python3 -m json.tool
	@echo "Dashboard:" && curl -s http://localhost:8082/health | python3 -m json.tool

clean:
	docker compose down -v --rmi local
	rm -rf dashboard/node_modules dashboard/dist
	rm -rf worker-engine/worker-engine
