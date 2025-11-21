# Gizmatron - Architecture & Project Overview

## Project Summary

**Gizmatron** is a robotics project built in Go that runs on a Raspberry Pi. It's designed as a robot with a camera mounted on an articulated robotic arm that can see and interact with humans. The project represents a well-structured foundation for an interactive robotic system with clean separation between hardware control, computer vision, and API layers.

## Core Architecture

### Hardware Platform
- **Target Device:** Raspberry Pi 3 B+ (with plans for newer platforms)
- **Operating System:** Raspberry Pi OS Lite Bookworm 64-bit
- **Kernel:** Linux Kernel 6.6.20
- **Deployment:** Containerized using Docker for easy deployment and portability

### Main Hardware Components

#### 1. Robotic Arm (`robot/arm.go`, `robot/servo.go`)
- **Controller:** 5-servo articulated arm controlled via PCA9685 I2C driver
- **Servo Layout:**
  - Base Servo (Channel 0) - 2.0mm link length
  - Joint 1 Servo (Channel 1) - 10.3mm link length  
  - Joint 2 Servo (Channel 2) - 2.8mm link length
  - Joint 3 Servo (Channel 3) - 10.3mm link length
  - Joint 4 Servo (Channel 4) - 2.0mm link length
- **Features:**
  - Configurable direction, angle limits, and link lengths per servo
  - PWM frequency set to 50Hz for servo control
  - Designed for precise positioning and movement control
  - Default positioning with gradual movement capabilities

#### 2. Camera System (`robot/camera.go`)
- **Technology:** OpenCV-based computer vision using GoCV (Go wrapper for OpenCV)
- **Capabilities:**
  - Face detection (currently disabled due to technical issues)
  - Video streaming support with MJPEG format
  - Still image capture functionality
  - Real-time video processing
- **Integration:** HTTP streaming endpoints for web interface access

#### 3. LED Status Indicators (`robot/led.go`)
- **Running LED:** GPIO 26 (pin 37) - indicates robot is operational
- **Server LED:** GPIO 13 (pin 33) - indicates API server status  
- **Arm LED:** GPIO 5 (pin 29) - indicates arm operational status
- **Control:** GPIO control via `github.com/warthog618/go-gpiocdev` library

#### 4. HTTP API Server (`server/server.go`, `server/handlers.go`)
- **Framework:** Native Go HTTP server with custom middleware
- **Port:** 8080
- **Features:** 
  - RESTful API design
  - CORS-enabled for web interface integration
  - Request logging and client tracking
  - JSON response formatting

## Technology Stack

### Core Technologies
- **Programming Language:** Go 1.23+ (ARM64 architecture)
- **Computer Vision:** GoCV v0.40.0 (OpenCV 4.11 wrapper)
- **Hardware Libraries:**
  - `github.com/warthog618/go-gpiocdev` v0.9.1 - GPIO control
  - `periph.io` - I2C communication with servo controller
  - `gobot.io/x/gobot/v2` v2.5.0 - Robotics framework
- **Streaming:** `github.com/hybridgroup/mjpeg` - MJPEG video streaming
- **Deployment:** Docker with multi-platform support (linux/amd64, linux/arm64)

### External Dependencies
- **OpenCV:** 4.11 for computer vision processing
- **Docker:** Containerization and deployment
- **I2C Tools:** Hardware communication utilities

## API Endpoints

The robot exposes a comprehensive HTTP API for control and monitoring:

### Core Robot Control
- `GET /ping` - Health check and server availability
- `GET /api/v1/bot-status` - Get comprehensive robot operational status
- `POST /api/v1/bot-start` - Start robot operations and initialize components
- `POST /api/v1/bot-stop` - Stop robot operations and return to safe state

### Camera Operations
- `GET /api/v1/video` - Real-time video streaming (MJPEG format)
- `POST /api/v1/takepicture` - Capture and return still image
- `POST /api/v1/detectfaces` - Enable/disable face detection feature
- `POST /api/v1/start/stream` - Initialize video streaming
- `POST /api/v1/stop/stream` - Stop video streaming

### Device Management
Each endpoint returns detailed device status information including:
- Individual component operational status
- Error states and diagnostic information
- Runtime configuration and capabilities

## Project Structure

```
gizmatron/
├── main.go                 # Application entry point and initialization
├── go.mod                  # Go module dependencies
├── docker-compose.yml      # Local development deployment
├── Dockerfile             # Container build configuration
├── openapi.yml           # API specification
├── README.md             # Project documentation
├── ARCHITECTURE.md       # This architecture overview
├── build/                # Multi-platform build configurations
│   ├── Dockerfile.AlpineOpencv
│   ├── Dockerfile.buildapp
│   ├── Dockerfile.buildbase
│   ├── Dockerfile.buildGizmatron
│   ├── Dockerfile.gocv-gobot
│   └── Dockerfile.multiplatform
├── robot/                # Hardware abstraction layer
│   ├── robot.go          # Main robot controller and device management
│   ├── arm.go            # Robotic arm control and kinematics
│   ├── servo.go          # Individual servo motor control
│   ├── camera.go         # Camera operations and computer vision
│   ├── led.go            # LED status indicator control
│   └── PCA9685Driver.go  # I2C servo driver implementation
└── server/               # HTTP API layer
    ├── server.go         # HTTP server setup and routing
    └── handlers.go       # API endpoint implementations
```

## Current Operational Status

### ✅ Working Features
- **Core Infrastructure:** Robot initialization and device management
- **API Server:** HTTP server with comprehensive endpoint coverage
- **Hardware Control:** LED status indicators and GPIO management
- **Servo Control:** Full 5-servo arm control system with I2C communication
- **Containerization:** Docker deployment with proper device access
- **Multi-platform Support:** ARM64 and AMD64 container builds

### ⚠️ Known Issues
- **Camera System:** GoCV camera functionality currently non-operational
- **Face Detection:** Disabled due to camera system issues
- **Error Handling:** Some incomplete error handling in device initialization
- **Stream Management:** Video streaming endpoints need stabilization

### 🚧 Development Areas
- **Device Recovery:** Improved error handling and device reconnection
- **Configuration Management:** Runtime configuration system
- **Logging:** Enhanced structured logging throughout the system

## Deployment Architecture

### Local Development Environment
```yaml
# docker-compose.yml configuration
services:
  gizmatron:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - "/dev/video0:/dev/video0"    # Camera access
      - "/dev/gpiomem:/dev/gpiomem"  # GPIO memory access
    devices:
      - "/dev/i2c-1:/dev/i2c-1"     # I2C bus access
    privileged: true                # Hardware access permissions
```

### Remote Access Infrastructure
- **Twingate VPN:** Zero-trust network access for secure remote robot control
- **GitHub Actions:** CI/CD pipeline running directly on Raspberry Pi
- **Container Registry:** Multi-platform Docker images for deployment flexibility

### Security Considerations
- **Network Isolation:** Twingate ensures secure communication channels
- **Container Security:** Privileged access limited to necessary hardware interfaces
- **API Security:** Request logging and client identification for monitoring

## Future Development Roadmap

### Immediate Priorities
1. **Camera System Recovery:** Debug and fix GoCV integration issues
2. **Face Detection Restoration:** Re-enable computer vision capabilities
3. **Error Handling Enhancement:** Robust device failure recovery mechanisms
4. **API Stabilization:** Comprehensive testing and error response standardization

### Medium-term Goals
- **AI Integration:** Large Language Model (LLM) integration for conversational AI
- **Autonomous Behavior:** AI agent integration for independent operation
- **Enhanced Vision:** Advanced computer vision capabilities beyond face detection
- **Motion Planning:** Sophisticated arm movement and trajectory planning

### Long-term Vision
- **Cloud Integration:** Backend infrastructure for enhanced processing capabilities
- **Multi-robot Coordination:** Support for multiple Gizmatron units
- **Learning Capabilities:** Machine learning integration for adaptive behavior
- **Advanced Interaction:** Natural language processing and response generation

## Development Guidelines

### Code Organization
- **Separation of Concerns:** Clear boundaries between hardware, API, and business logic
- **Error Handling:** Comprehensive error propagation and recovery mechanisms
- **Device Abstraction:** Hardware-agnostic interfaces for component swapping
- **Logging Strategy:** Structured logging with appropriate verbosity levels

### Testing Strategy
- **Unit Testing:** Component-level testing for individual modules
- **Integration Testing:** Hardware-in-the-loop testing for device interfaces
- **API Testing:** Comprehensive endpoint testing and validation
- **Performance Testing:** Real-time operation validation and optimization

### Contribution Workflow
- **Branching Strategy:** Feature branches with pull request workflow
- **Documentation:** Inline code documentation and architectural updates
- **Hardware Testing:** Physical device validation for hardware-related changes
- **Container Testing:** Multi-platform container build verification

## Hardware Requirements

### Minimum Specifications
- **Raspberry Pi 3 B+** or newer
- **8GB microSD card** (minimum, 32GB recommended)
- **2.5A power supply** for stable operation
- **USB camera** (compatible with Video4Linux)
- **PCA9685 servo driver board** for arm control
- **5 servo motors** for articulated arm
- **3 LEDs** for status indication
- **I2C and GPIO connectivity** for hardware interfaces

### Optional Enhancements
- **Raspberry Pi 4** or **Pi 5** for improved performance
- **High-resolution camera** for enhanced computer vision
- **Additional sensors** (IMU, distance sensors, etc.)
- **Audio capabilities** for voice interaction
- **Network connectivity** improvements (WiFi 6, Ethernet)

---

*This architecture document provides a comprehensive overview of the Gizmatron robotics project. For specific implementation details, refer to the individual source files and inline documentation.*