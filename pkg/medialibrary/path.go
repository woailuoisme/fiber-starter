package medialibrary

import (
	"fmt"
	"path/filepath"
	"strings"
)

type PathGenerator struct{}

func (PathGenerator) OriginalPath(mediaUUID, fileName string) string {
	return fmt.Sprintf("media/%s/%s", mediaUUID, fileName)
}

func (PathGenerator) ConversionPath(mediaUUID, fileName, conversionName, format string) string {
	ext := filepath.Ext(fileName)
	base := strings.TrimSuffix(fileName, ext)
	targetExt := "." + strings.TrimPrefix(format, ".")
	return fmt.Sprintf("media/%s/conversions/%s-%s%s", mediaUUID, base, conversionName, targetExt)
}

func (PathGenerator) MediaDirectory(mediaUUID string) string {
	return fmt.Sprintf("media/%s", mediaUUID)
}
