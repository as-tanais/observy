#!/bin/bash

# Устанавливаем директории
PROTO_DIR="internal/proto/metrics"
OUT_DIR="internal/proto/metrics"

# Создаем выходную директорию если её нет
mkdir -p ${OUT_DIR}

# Генерируем код
protoc \
  --go_out=${OUT_DIR} \
  --go_opt=paths=source_relative \
  --go-grpc_out=${OUT_DIR} \
  --go-grpc_opt=paths=source_relative \
  ${PROTO_DIR}/metrics.proto

echo "Proto files generated successfully in ${OUT_DIR}"