ROOTPROJECT ?= ../root
APIPROJECT = .
PROTOBUF_FILES=streaming/sample.pb.go
include ${ROOTPROJECT}/include.mk

# Dummy targets for cluster/up and cluster/teardown
.PHONY: up down
up:
down: