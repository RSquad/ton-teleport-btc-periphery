#!/bin/bash

mkdir -p ./pkg/ton/generated

swagger generate client \
  --spec=./ton-center-v3-swagger.json \
  --target=./pkg/ton/generated \
  --client-package=toncenterv3client \
  --model-package=toncenterv3models \
  --api-package=toncenterv3operations \
  -A TonCenterV3 \

go mod tidy