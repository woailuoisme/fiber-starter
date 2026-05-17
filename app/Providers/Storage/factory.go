package storage

import (
	"fmt"

	"fiber-starter/app/Providers/Storage/Contracts"
	"fiber-starter/app/Providers/Storage/Drivers"
	"fiber-starter/configs"
)

func createDisk(name string, cfg *configs.Config) (Contracts.Disk, error) {
	driver := name
	if driver == "" {
		driver = cfg.Storage.Driver
	}

	switch driver {
	case "local":
		root := "./storage/app"
		url := "/storage"
		if cfg.Storage.Local != nil {
			if cfg.Storage.Local.Root != "" {
				root = cfg.Storage.Local.Root
			}
			if cfg.Storage.Local.URL != "" {
				url = cfg.Storage.Local.URL
			}
		}
		return Drivers.NewLocalDriver(root, url), nil

	case "public":
		root := "./storage/app/public"
		url := "/storage"
		if cfg.Storage.Public != nil {
			if cfg.Storage.Public.Root != "" {
				root = cfg.Storage.Public.Root
			}
			if cfg.Storage.Public.URL != "" {
				url = cfg.Storage.Public.URL
			}
		}
		return Drivers.NewLocalDriver(root, url), nil

	case "s3", "garage", "minio", "r2", "oss":
		return Drivers.NewS3Driver(cfg, driver), nil

	default:
		return nil, fmt.Errorf("storage driver [%s] is not supported", driver)
	}
}
