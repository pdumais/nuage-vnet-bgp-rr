# Description

## High Availability
Multiple instances of the app can run in order to achieve HA (2 or more). Each instance should have a different router ID defined.
The one with the lowest ID will be active and the others will be standby. HA is achieved through BGP. A path for a /32 destination
using the router ID will be advertised to the NSGs. Other instances will receive this path through ebgp and this will allow
them to discover each other. But the requirements to make it work are:
    - Each instance must have a different AS, and different from the AS of the NSG
    - Each instance must have a unique router ID using an IP from a range that will not create conflicts 
      (for example, an IP within 192.0.2.0/24)
    - Other instances will recognize the path because of its community ID (242,242). So a policy could be created
      in the VSD to limit where these routes would be advertised.

# Building
Building the sources is not required for running the application. The application is available on Dockerhub. But if
you want to build it, docker can be used:
    - docker run --rm -v "$PWD":/go/src -w /go/src golang:1.11 go get
    - docker run --rm -v "$PWD":/go/src -w /go/src golang:1.11 go build -v -o vnet-bgp-monitor



# Running
    docker run 

# Troubleshooting
## Logs
 docker inspect --format='{{.LogPath}}'
TODO: can leverage docker to send logs to graylog, journald, syslog
## Peering
