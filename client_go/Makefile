PB_OUT_DIR = measurement
BINARY_NAME = iotcorelogger

all: pb
	go build -o $(BINARY_NAME) ./cmd/...
	GOOS=linux GOARCH=arm GOARM=7 go build -o $(BINARY_NAME)_armv7 ./cmd/...
	GOOS=linux GOARCH=arm GOARM=6 go build -o $(BINARY_NAME)_armv6 ./cmd/...

pb_out_dir:
	mkdir -p $(PB_OUT_DIR)

pb: pb_out_dir
	protoc --go_out=$(PB_OUT_DIR) --proto_path=.. measurement.proto

clean:
	rm -rf $(PB_OUT_DIR) $(BINARY_NAME)*
