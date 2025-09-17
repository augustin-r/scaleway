# Overview

This project provides a agent that manages a load balancer on the fly.

It is a proof of concept:
- Should only be used locally
- Huge lack of security
- Require mocked servers for testing (provided in the docker compose)

Main techs:
- Agent is developed in Golang (required by assessment)
- Load balancer is haproxy (makes sense as it is used by Scaleway).
- Load balancer conf is updated with its Data Plane API
- Mocked servers are nginx (no specific reason, just the first solution that popped in GPT)
- The load balancer and  is wrapped in a docker compose (makes it easy to set up)

# Install

Project is expected to be used from WSL.

Following softwares are required:
- Docker / Docker compose
- Go 1.24
- Postman

This project relies on a docker image that needs to be built:
- Go to `./haproxy`
- Run `docker build --no-cache -t my-haproxy .`

This project exposes a gRPC server. Proto need to be compiled:
- Go to the project root folder
- Run `protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/lb_agent.proto`

You're good to go!

# Launch

Start the docker compose to launch the load balancer, its API and the mock servers : 
- Go to `./haproxy`
- Run `docker compose up -d`

Start the Go agent :
- Go to the project root folder
- Run `go run main.go`

Launch Postman (or equivalent). Use reflection to discover available gRPC services. Test.

# Debug

## Target the load balancer directly through the Data Plane API

See https://www.haproxy.com/documentation/haproxy-data-plane-api/tutorials/ for easy curl commands to copy/paste.
