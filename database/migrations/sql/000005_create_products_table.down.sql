-- 删除触发器
DROP TRIGGER IF EXISTS update_products_updated_at ON products;

-- 删除产品表
DROP TABLE IF EXISTS products;
