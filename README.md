# Foo Service
golang backend service scaffold

### Features
- [x] API server
- [x] worker for background processing or consumers
- [x] telemetry
- [x] automatic database migration
- [x] env config override
- [x] unit tests

### Requirements
- go 1.21
- docker 24

### Setup
- copy `.env.sample` to `.env` and change values accordingly
- run postgres database `docker run --rm -d -e POSTGRES_USER=root -e POSTGRES_PASSWORD=password -p 5432:5432 postgres:15.4`
- *(optional)* run telemetry exporter `docker run --rm -p 4317:4317 otel/opentelemetry-collector-contrib:0.82.0`

### Running locally
- run server `make run-server`
- run worker `make run-worker`
