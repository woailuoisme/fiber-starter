package storage

import (
	"fmt"

	"lfiber/configs"
	"lfiber/internal/providers/storage/contracts"
	"lfiber/internal/providers/storage/drivers"
)

func createDisk(name string, cfg *configs.Config) (contracts.Disk, error) {
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
		return drivers.NewLocalDriver(root, url), nil

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
		return drivers.NewLocalDriver(root, url), nil

	case "s3", "garage", "minio", "r2", "oss":
		return drivers.NewS3Driver(cfg, driver), nil

	case "noop", "null":
		return drivers.NewNoopDisk(), nil

	default:
		return nil, fmt.Errorf("storage driver [%s] is not supported", driver)
	}
}
