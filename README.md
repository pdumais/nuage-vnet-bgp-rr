# Description
## Architecture
## High Availability
# Building
Building the sources is not required for running the application. The application is available on Dockerhub. But if
you want to build it, docker can be used:
    - docker run --rm -v "$PWD":/go/src -w /go/src golang:1.11 go get
    - docker run --rm -v "$PWD":/go/src -w /go/src golang:1.11 go build -v -o vnet-bgp-rr 



# Running
    docker run 

# Troubleshooting
## Logs
 docker inspect --format='{{.LogPath}}'
TODO: can leverage docker to send logs to graylog, journald, syslog
## Peering
