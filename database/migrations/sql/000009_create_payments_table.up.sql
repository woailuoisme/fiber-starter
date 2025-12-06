-- 创建支付表
CREATE TABLE IF NOT EXISTS payments (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    payment_method VARCHAR(20) NOT NULL CHECK (payment_method IN ('wechat', 'alipay')),
    payment_number VARCHAR(100) UNIQUE NOT NULL,
    amount INTEGER NOT NULL CHECK (amount >= 0),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'success', 'failed', 'refunded')),
    transaction_id VARCHAR(255),
    paid_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE RESTRICT
);

-- 创建索引
CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payments_payment_number ON payments(payment_number);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_transaction_id ON payments(transaction_id);

-- 为 payments 表创建触发器
CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 添加注释
COMMENT ON TABLE payments IS '支付表';
COMMENT ON COLUMN payments.id IS '支付ID';
COMMENT ON COLUMN payments.order_id IS '订单ID';
COMMENT ON COLUMN payments.payment_method IS '支付方式：wechat-微信支付, alipay-支付宝';
COMMENT ON COLUMN payments.payment_number IS '支付单号（唯一）';
COMMENT ON COLUMN payments.amount IS '支付金额（以分为单位）';
COMMENT ON COLUMN payments.status IS '支付状态：pending-待支付, success-成功, failed-失败, refunded-已退款';
COMMENT ON COLUMN payments.transaction_id IS '第三方交易ID';
COMMENT ON COLUMN payments.paid_at IS '支付时间';
