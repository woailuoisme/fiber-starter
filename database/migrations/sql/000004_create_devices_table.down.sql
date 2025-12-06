-- 删除触发器
DROP TRIGGER IF EXISTS update_devices_updated_at ON devices;

-- 删除设备表
DROP TABLE IF EXISTS devices;
