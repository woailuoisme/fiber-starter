package medialibrary

import (
	"fmt"
	"strings"
)

type CollectionPolicy struct {
	collection *MediaCollection
}

func NewCollectionPolicy(collection *MediaCollection) CollectionPolicy {
	return CollectionPolicy{collection: collection}
}

func (p CollectionPolicy) Validate(fileSize int64, mimeType, collectionName string) error {
	if p.collection == nil {
		return nil
	}
	if p.collection.MaxFileSize > 0 && fileSize > p.collection.MaxFileSize {
		return fmt.Errorf("%w: file size %d exceeds collection max size limit %d", ErrFileTooLarge, fileSize, p.collection.MaxFileSize)
	}
	if len(p.collection.AllowedMimes) == 0 {
		return nil
	}
	for _, allowed := range p.collection.AllowedMimes {
		allowed = strings.TrimSpace(allowed)
		if strings.EqualFold(allowed, mimeType) {
			return nil
		}
		if strings.HasSuffix(allowed, "/*") && strings.HasPrefix(mimeType, strings.TrimSuffix(allowed, "*")) {
			return nil
		}
	}
	return fmt.Errorf("%w: mime type %s is not allowed in collection %s", ErrMimeNotAllowed, mimeType, collectionName)
}
