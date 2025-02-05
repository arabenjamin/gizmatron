#!/bin/bash
# This bash script compiles OpenCV on Debian-based systems

# Update and install dependencies
sudo apt-get update
sudo apt-get install -y build-essential cmake git pkg-config libjpeg-dev libtiff-dev \
      libpng-dev libavcodec-dev libavformat-dev libswscale-dev libv4l-dev \
      libxvidcore-dev libx264-dev libgtk-3-dev libatlas-base-dev gfortran python3-dev

# Download OpenCV and OpenCV Contrib
cd ~
git clone https://github.com/opencv/opencv.git
git clone https://github.com/opencv/opencv_contrib.git

# Checkout the desired version
cd opencv
git checkout 4.11.0
cd ../opencv_contrib
git checkout 4.11.0

# Build and install OpenCV
cd ~/opencv
mkdir build
cd build


cmake -D CMAKE_BUILD_TYPE=RELEASE \
    -D CMAKE_INSTALL_PREFIX=/usr/local \
    -D OPENCV_EXTRA_MODULES_PATH=../../opencv_contrib/modules \
    -D WITH_FFMPEG=YES \
    -D WITH_V4L=ON \
    -D WITH_LIBV4L=ON \
    -D WITH_PROTOBUF=ON \
    -D INSTALL_C_EXAMPLES=NO \
    -D INSTALL_PYTHON_EXAMPLES=NO \
    -D BUILD_ANDROID_EXAMPLES=NO \
    -D BUILD_DOCS=NO \
    -D BUILD_TESTS=NO \
    -D BUILD_PERF_TESTS=NO \
    -D BUILD_EXAMPLES=NO \
    -D BUILD_opencv_java=NO \
    -D BUILD_opencv_python=NO \
    -D BUILD_opencv_python2=NO \
    -D BUILD_opencv_python3=YES \
    -D OPENCV_GENERATE_PKGCONFIG=ON \
    -D CMAKE_INSTALL_RPATH=/usr/local/lib ..


make -j$(nproc)
sudo make install
sudo ldconfig

# Verification
echo "OpenCV version installed:"
python3 -c "import cv2; print(cv2.__version__)"