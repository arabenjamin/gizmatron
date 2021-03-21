FROM arabenjamin/opencv-go:latest

ENV PKG_CONFIG_PATH /usr/local/lib64/pkgconfig
ENV LD_LIBRARY_PATH /usr/local/lib64
ENV CGO_CPPFLAGS -I/usr/local/include
ENV CGO_CXXFLAGS "--std=c++1z"
ENV CGO_LDFLAGS "-L/usr/local/lib -lopencv_core -lopencv_face -lopencv_videoio -lopencv_imgproc -lopencv_highgui -lopencv_imgcodecs -lopencv_objdetect -lopencv_features2d -lopencv_video -lopencv_dnn -lopencv_xfeatures2d -lopencv_plot -lopencv_tracking"


RUN apk update
RUN git config --global url."https://${GIT-TOKEN}:@github.com/".insteadOf "https://github.com/"


# configure Go
ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
# Set GOPATH mode
RUN go env -w GO111MODULE=auto
RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin
#RUN sudo modprobe bcm2835-v4l2Z

# Install Dependancyies 
#RUN go get -u -d gocv.io/x/gocv
#RUN go get -u github.com/hybridgroup/mjpeg
#RUN go get -u github.com/stianeikeland/go-rpio


# Copy project files
COPY . $GOPATH/src/robot/
WORKDIR $GOPATH/src/robot

RUN go build

CMD ["./gizmatron"]


EXPOSE 8080
