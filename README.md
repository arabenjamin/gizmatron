# gizmatron

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

`docker build -t opencv-go-base -f Dockerfile.buildbase .`

`docker build --no-cache -t gizmatron -f Dockerfile.buildapp .`


`docker run --device /dev/video0:/dev/video0 -p 8080:8080 gizmatron`

