.PHONY: all
all:
	protoc --go_out=web/measurement --python_out=client_python/loggers measurement.proto
