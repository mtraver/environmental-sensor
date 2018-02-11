.PHONY: all
all:
	protoc --go_out=receiver/measurement --python_out=loggers measurement.proto
