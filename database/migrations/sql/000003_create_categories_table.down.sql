-- 删除触发器
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;

-- 删除产品分类表
DROP TABLE IF EXISTS categories;
