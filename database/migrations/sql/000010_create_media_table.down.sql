-- 删除触发器
DROP TRIGGER IF EXISTS update_media_updated_at ON media;

-- 删除媒体表
DROP TABLE IF EXISTS media;
