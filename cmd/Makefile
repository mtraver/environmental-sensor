BUILD_ARMV6 := GOOS=linux GOARCH=arm GOARM=6 go build
BUILD_ARMV7 := GOOS=linux GOARCH=arm GOARM=7 go build

OUT_DIR := out

all: iotcorelogger readtemp

.PHONY: iotcorelogger
iotcorelogger:
	go build -o $(OUT_DIR)/$@ ./$@/...
	$(BUILD_ARMV7) -o $(OUT_DIR)/armv7/$@ ./$@/...
	$(BUILD_ARMV6) -o $(OUT_DIR)/armv6/$@ ./$@/...

.PHONY: readtemp
readtemp:
	go build -o $(OUT_DIR)/$@ ./$@/...
	$(BUILD_ARMV7) -o $(OUT_DIR)/armv7/$@ ./$@/...
	$(BUILD_ARMV6) -o $(OUT_DIR)/armv6/$@ ./$@/...

clean:
	rm -rf $(OUT_DIR)
