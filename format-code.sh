#!/bin/bash
find container-runtime -type f -name '*.go' -not -path './vendor/*' -print0 | xargs -0 gofmt -w
