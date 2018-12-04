.PHONY: all
all:
	protoc --go_out=measurement --python_out=client_python/loggers measurement.proto
