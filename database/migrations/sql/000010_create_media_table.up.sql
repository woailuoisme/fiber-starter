-- 创建媒体表
CREATE TABLE IF NOT EXISTS media (
    id BIGSERIAL PRIMARY KEY,
    model_type VARCHAR(255) NOT NULL,
    model_id BIGINT NOT NULL,
    collection_name VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    mime_type VARCHAR(255),
    disk VARCHAR(50) DEFAULT 'local',
    size BIGINT,
    manipulations JSONB,
    custom_properties JSONB,
    order_column INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_media_model ON media(model_type, model_id);
CREATE INDEX idx_media_collection ON media(collection_name);
CREATE INDEX idx_media_order ON media(order_column);
CREATE INDEX idx_media_deleted_at ON media(deleted_at);

-- 为 media 表创建触发器
CREATE TRIGGER update_media_updated_at BEFORE UPDATE ON media
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 添加注释
COMMENT ON TABLE media IS '媒体文件表';
COMMENT ON COLUMN media.id IS '媒体ID';
COMMENT ON COLUMN media.model_type IS '关联模型类型';
COMMENT ON COLUMN media.model_id IS '关联模型ID';
COMMENT ON COLUMN media.collection_name IS '集合名称（如：images, avatars, documents）';
COMMENT ON COLUMN media.name IS '媒体名称';
COMMENT ON COLUMN media.file_name IS '文件名';
COMMENT ON COLUMN media.mime_type IS 'MIME类型';
COMMENT ON COLUMN media.disk IS '存储驱动：local-本地, minio-MinIO, s3-AWS S3';
COMMENT ON COLUMN media.size IS '文件大小（字节）';
COMMENT ON COLUMN media.manipulations IS '图片处理配置（JSON）';
COMMENT ON COLUMN media.custom_properties IS '自定义属性（JSON）';
COMMENT ON COLUMN media.order_column IS '排序字段';
