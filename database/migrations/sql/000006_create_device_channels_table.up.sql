-- 创建设备货道表
CREATE TABLE IF NOT EXISTS device_channels (
    id BIGSERIAL PRIMARY KEY,
    device_id BIGINT NOT NULL,
    channel_number INTEGER NOT NULL CHECK (channel_number >= 1 AND channel_number <= 53),
    product_id BIGINT,
    virtual_stock INTEGER DEFAULT 0 CHECK (virtual_stock >= 0 AND virtual_stock <= 4),
    actual_stock INTEGER DEFAULT 0 CHECK (actual_stock >= 0 AND actual_stock <= 4),
    max_capacity INTEGER DEFAULT 4,
    status VARCHAR(20) DEFAULT 'normal' CHECK (status IN ('normal', 'fault', 'maintenance', 'disabled')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL,
    UNIQUE (device_id, channel_number)
);

-- 创建索引
CREATE INDEX idx_device_channels_device_id ON device_channels(device_id);
CREATE INDEX idx_device_channels_product_id ON device_channels(product_id);
CREATE INDEX idx_device_channels_status ON device_channels(status);
CREATE INDEX idx_device_channels_deleted_at ON device_channels(deleted_at);

-- 为 device_channels 表创建触发器
CREATE TRIGGER update_device_channels_updated_at BEFORE UPDATE ON device_channels
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 添加注释
COMMENT ON TABLE device_channels IS '设备货道表';
COMMENT ON COLUMN device_channels.id IS '货道ID';
COMMENT ON COLUMN device_channels.device_id IS '所属设备ID';
COMMENT ON COLUMN device_channels.channel_number IS '货道编号（1-53）';
COMMENT ON COLUMN device_channels.product_id IS '产品ID';
COMMENT ON COLUMN device_channels.virtual_stock IS '虚拟库存（用于下单扣减）';
COMMENT ON COLUMN device_channels.actual_stock IS '实际库存（用于取货扣减）';
COMMENT ON COLUMN device_channels.max_capacity IS '最大容量';
COMMENT ON COLUMN device_channels.status IS '货道状态：normal-正常, fault-故障, maintenance-维护中, disabled-已禁用';
