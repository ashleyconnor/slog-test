default: docker

docker:
	docker-compose up --build

request:
	curl -X GET http://localhost:8080/hello

jaeger:
	open http://localhost:16686
