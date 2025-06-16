# Build output directory
build_dir := "build"

# Default build (ARM Linux)
# build: build-arm-linux
# Build all Linux architectures
build-all: build-arm-linux build-arm64-linux build-amd64-linux

# Build for ARM Linux
build-arm-linux:
    GOARCH=arm GOOS=linux go build -o {{build_dir}}/modbus-serial2tcp-arm-linux .

# Build for ARM64 Linux
build-arm64-linux:
    GOARCH=arm64 GOOS=linux go build -o {{build_dir}}/modbus-serial2tcp-arm64-linux .

# Build for x86_64 Linux
build-amd64-linux:
    GOARCH=amd64 GOOS=linux go build -o {{build_dir}}/modbus-serial2tcp-amd64-linux .


# Clean build directory
clean:
    rm -rf {{build_dir}}