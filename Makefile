BUILD := go build
BUILD_ARMV6 := GOOS=linux GOARCH=arm GOARM=6 $(BUILD)
BUILD_ARMV7 := GOOS=linux GOARCH=arm GOARM=7 $(BUILD)

OUT_DIR := out

all: iotcorelogger readtemp

.PHONY: iotcorelogger
iotcorelogger: proto
	$(BUILD) -o $(OUT_DIR)/$@ ./cmd/$@
	$(BUILD_ARMV7) -o $(OUT_DIR)/armv7/$@ ./cmd/$@
	$(BUILD_ARMV6) -o $(OUT_DIR)/armv6/$@ ./cmd/$@

.PHONY: readtemp
readtemp:
	$(BUILD) -o $(OUT_DIR)/$@ ./cmd/$@
	$(BUILD_ARMV7) -o $(OUT_DIR)/armv7/$@ ./cmd/$@
	$(BUILD_ARMV6) -o $(OUT_DIR)/armv6/$@ ./cmd/$@

# This also used to generate Python code, but the Python client is no longer
# maintained so it no longer does. If you wish to generate Python, add this
# flag to the protoc command:
#   --python_out=client_python/loggers
.PHONY: proto
proto:
	protoc --go_out=measurement measurement.proto

clean:
	rm -rf $(OUT_DIR)
