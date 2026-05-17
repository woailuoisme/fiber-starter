#!/bin/bash

# 从 .env 文件读取数据库配置
source .env

# 创建数据库
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USERNAME -d postgres -c "CREATE DATABASE \"$DB_DATABASE\";"

echo "Database $DB_DATABASE created successfully!"
