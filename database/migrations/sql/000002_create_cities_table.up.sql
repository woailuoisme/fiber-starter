-- 创建城市表
CREATE TABLE IF NOT EXISTS cities (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_cities_status ON cities(status);
CREATE INDEX idx_cities_deleted_at ON cities(deleted_at);

-- 为 cities 表创建触发器
CREATE TRIGGER update_cities_updated_at BEFORE UPDATE ON cities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 添加注释
COMMENT ON TABLE cities IS '城市表';
COMMENT ON COLUMN cities.id IS '城市ID';
COMMENT ON COLUMN cities.name IS '城市名称';
COMMENT ON COLUMN cities.status IS '城市状态：active-活跃, inactive-未激活';
