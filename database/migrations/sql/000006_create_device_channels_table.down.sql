-- 删除触发器
DROP TRIGGER IF EXISTS update_device_channels_updated_at ON device_channels;

-- 删除设备货道表
DROP TABLE IF EXISTS device_channels;
