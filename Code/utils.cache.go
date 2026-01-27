package main

import "context"

func isImageCached(imgCache *ImageDescriptionCache, hash string, db *AppDB) (string, error) {
	imgCache.mu.RLock()
	cachedImage, ok := imgCache.descriptions[hash]
	imgCache.mu.RUnlock()

	if ok {
		return cachedImage, nil
	}

	cachedImage, err := db.GetImageDescription(context.Background(), hash)
	if err != nil {
		return "", err
	}

	return cachedImage, nil
}

func setNewImageCache(imgCache *ImageDescriptionCache, hash string, description string, db *AppDB) error {
	imgCache.mu.Lock()
	imgCache.descriptions[hash] = description
	imgCache.mu.Unlock()

	err := db.AddImageDescription(context.Background(), hash, description)
	if err != nil {
		return err
	}

	return nil
}
