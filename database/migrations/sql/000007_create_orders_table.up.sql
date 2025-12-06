-- 创建订单表
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    device_id BIGINT NOT NULL,
    order_number VARCHAR(100) UNIQUE NOT NULL,
    total_amount INTEGER NOT NULL CHECK (total_amount >= 0),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'completed', 'cancelled', 'timeout')),
    qr_code VARCHAR(255) UNIQUE NOT NULL,
    paid_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE RESTRICT
);

-- 创建索引
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_device_id ON orders(device_id);
CREATE INDEX idx_orders_order_number ON orders(order_number);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);
CREATE INDEX idx_orders_deleted_at ON orders(deleted_at);

-- 为 orders 表创建触发器
CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 添加注释
COMMENT ON TABLE orders IS '订单表';
COMMENT ON COLUMN orders.id IS '订单ID';
COMMENT ON COLUMN orders.user_id IS '用户ID';
COMMENT ON COLUMN orders.device_id IS '设备ID';
COMMENT ON COLUMN orders.order_number IS '订单号（唯一）';
COMMENT ON COLUMN orders.total_amount IS '订单总金额（以分为单位）';
COMMENT ON COLUMN orders.status IS '订单状态：pending-待支付, paid-已支付, completed-已完成, cancelled-已取消, timeout-已超时';
COMMENT ON COLUMN orders.qr_code IS '取货二维码';
COMMENT ON COLUMN orders.paid_at IS '支付时间';
COMMENT ON COLUMN orders.completed_at IS '完成时间';
