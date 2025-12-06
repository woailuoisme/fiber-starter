-- 删除触发器
DROP TRIGGER IF EXISTS update_cities_updated_at ON cities;

-- 删除城市表
DROP TABLE IF EXISTS cities;
