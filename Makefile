.PHONY: all
all:
	protoc --go_out=receiver/measurement --python_out=client_python/loggers measurement.proto
