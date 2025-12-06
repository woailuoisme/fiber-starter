-- 创建设备表
CREATE TABLE IF NOT EXISTS devices (
    id BIGSERIAL PRIMARY KEY,
    city_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100) UNIQUE NOT NULL,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    status VARCHAR(20) DEFAULT 'offline' CHECK (status IN ('online', 'offline', 'fault')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (city_id) REFERENCES cities(id) ON DELETE RESTRICT
);

-- 创建索引
CREATE INDEX idx_devices_city_id ON devices(city_id);
CREATE INDEX idx_devices_code ON devices(code);
CREATE INDEX idx_devices_status ON devices(status);
CREATE INDEX idx_devices_deleted_at ON devices(deleted_at);

-- 为 devices 表创建触发器
CREATE TRIGGER update_devices_updated_at BEFORE UPDATE ON devices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 添加注释
COMMENT ON TABLE devices IS '设备表';
COMMENT ON COLUMN devices.id IS '设备ID';
COMMENT ON COLUMN devices.city_id IS '所属城市ID';
COMMENT ON COLUMN devices.name IS '设备名称';
COMMENT ON COLUMN devices.code IS '设备编码（唯一）';
COMMENT ON COLUMN devices.latitude IS '纬度';
COMMENT ON COLUMN devices.longitude IS '经度';
COMMENT ON COLUMN devices.status IS '设备状态：online-在线, offline-离线, fault-故障';
