FROM gocv/opencv:latest

ENV PKG_CONFIG_PATH /usr/local/lib/pkgconfig
ENV LD_LIBRARY_PATH /usr/local/lib
ENV CGO_CPPFLAGS -I/usr/local/include
ENV CGO_CXXFLAGS "--std=c++1z"
ENV CGO_LDFLAGS "-L/usr/local/lib -lopencv_core -lopencv_face -lopencv_videoio -lopencv_imgproc -lopencv_highgui -lopencv_imgcodecs -lopencv_objdetect -lopencv_features2d -lopencv_video -lopencv_dnn -lopencv_xfeatures2d -lopencv_plot -lopencv_tracking"

ENV CGO_ENABLED=1

# configure Go
ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

# Set GOPATH mode
#RUN go env -w GO111MODULE=auto
RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin
#RUN sudo modprobe bcm2835-v4l2

# Install dependencies including GStreamer and libcamera for camera support
RUN apt-get update && apt-get install -y \
    git \
    gcc \
    libc-dev \
    i2c-tools \
    gstreamer1.0-tools \
    gstreamer1.0-plugins-base \
    gstreamer1.0-plugins-good \
    gstreamer1.0-plugins-bad \
    libgstreamer1.0-dev \
    libgstreamer-plugins-base1.0-dev \
    libcamera-dev \
    libcamera-tools \
    v4l-utils \
    && rm -rf /var/lib/apt/lists/*

RUN git clone https://github.com/arabenjamin/gizmatron.git

WORKDIR gizmatron/

#RUN go build -tags customenv
RUN go mod tidy
RUN go build -a . 


CMD ["./gizmatron"]

EXPOSE 8080
