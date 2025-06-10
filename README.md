# gizmatron

Gizmatron is a robotics project I started, to challenge mytself, learn new things, and allow my imagination to take some form. Gizmatron has a camera attached to an articulated arm. The motivation for the robot is to be able to see and interact with humans.


## Design

Gizmatron runs in Docker on a RasperbyPi. Gizmatron is written in Go, and is using GoCV, with is a Go wrapper for OpenCv, for the computer vision component. When Gizmatron starts, it first runs a server so you can connect to it and interact with it via an api. Currently the api is not extensive. 

The api will tell you Gizmatrons status, if any or all of its devices are running, and being about to interact with those devices.

At some point. Gizmatron will be connected to an LLM, and possibly an Ai Agent, so it will need a connection to the server. Currently, untill I want to move more of the Backend infrastructure to the cloud, we'll need a way to make sure that if we have internet, we can talk to the server. To do so I am using [Twingate](https://www.twingate.com/). Twingate is a zero trust network access solution that allows you to connect to your private network without exposing it to the internet. This way, Gizmatron can securely communicate with the server and any other services it needs to access. Twingatte will run in a secure environment, ensuring that all communications are protected and efficient. 

Also running on the oi is a Gihub actions runner, so that I can run my CI/CD pipelines on the pi. This allows me to test and deploy changes to Gizmatron quickly and easily. The runner is configured to run on the same network as Gizmatron, so it can access all the necessary resources and services.



## Current Problems:
 The camera is borked in GOCV

## Prequisites

Currently Gizmatron runs on the raspberry pi 3 B+. There are plans for this to run on more recent platforms as well

* Raspberry pi os lite Bookworm 64 bit
* Linux Kernel 6.6.20
* Docker
* OpenCV 4.11
* Golang 1.23.5 arm64
* GoCv latest
* github.com/warthog618/go-gpiocdev  


## Building MultiPlatform docker image

`docker buildx build --no-cache --platform linux/amd64,linux/arm64 -t arabenjamin/gizmatron:latest --push -f Dockerfile.multiplatform .`

`docker buildx build --target builder-amd64 --target builder-arm64 --platform linux/amd64,linux/arm64 -t arabenjamin/gizmatron:latest --push -f Dockerfile.multiplatform .`

## Running in Docker

`docker run --device /dev/video0:/dev/video0 -p 8080:8080 gizmatron`

## Docker Compose 
`docker compose up`

