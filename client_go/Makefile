BUILD_ARMV6 := GOOS=linux GOARCH=arm GOARM=6 go build
BUILD_ARMV7 := GOOS=linux GOARCH=arm GOARM=7 go build

OUT_DIR := out

all: iotcorelogger readtemp

.PHONY: iotcorelogger
iotcorelogger: pb
	go build -o $(OUT_DIR)/$@ ./cmd/$@/...
	$(BUILD_ARMV7) -o $(OUT_DIR)/$@_armv7 ./cmd/$@/...
	$(BUILD_ARMV6) -o $(OUT_DIR)/$@_armv6 ./cmd/$@/...

.PHONY: readtemp
readtemp:
	go build -o $(OUT_DIR)/$@ ./cmd/$@/...
	$(BUILD_ARMV7) -o $(OUT_DIR)/$@_armv7 ./cmd/$@/...
	$(BUILD_ARMV6) -o $(OUT_DIR)/$@_armv6 ./cmd/$@/...

measurement:
	mkdir -p $@

pb: measurement
	protoc --go_out=$< --proto_path=.. measurement.proto

clean:
	rm -rf measurement $(OUT_DIR)
