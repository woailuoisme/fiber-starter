#!/bin/bash

# 显示统一构建配置

set -e

if [ -f ".buildconfig" ]; then
    set -a
    . ./.buildconfig
    set +a
fi

BUILD_DIR=${BUILD_DIR:-build}
COVERAGE_DIR=${COVERAGE_DIR:-coverage}
SERVER_BINARY_NAME=${SERVER_BINARY_NAME:-fiber-starter}
CLI_BINARY_NAME=${CLI_BINARY_NAME:-fiber-starter-cli}
APP_LOG_DIR=${APP_LOG_DIR:-storage/logs}
DEPLOY_DIR=${DEPLOY_DIR:-deploy}

cat <<EOF
BUILD_DIR=$BUILD_DIR
COVERAGE_DIR=$COVERAGE_DIR
SERVER_BINARY_NAME=$SERVER_BINARY_NAME
CLI_BINARY_NAME=$CLI_BINARY_NAME
APP_LOG_DIR=$APP_LOG_DIR
DEPLOY_DIR=$DEPLOY_DIR
EOF
