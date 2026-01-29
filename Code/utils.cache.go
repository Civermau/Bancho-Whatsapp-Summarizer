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

// isAliasCached checks the in-memory alias cache for a (chatJID, senderJID) pair.
// If not present, it falls back to the database and, on hit, populates the cache.
func isAliasCached(aliasCache *AliasCache, chatJID, senderJID string, db *AppDB) (string, bool, error) {
	if aliasCache == nil {
		return "", false, nil
	}

	key := chatJID + "|" + senderJID

	aliasCache.mu.RLock()
	alias, ok := aliasCache.aliases[key]
	aliasCache.mu.RUnlock()

	if ok {
		return alias, true, nil
	}

	dbAlias, found, err := db.GetAlias(context.Background(), chatJID, senderJID)
	if err != nil {
		return "", false, err
	}
	if !found {
		return "", false, nil
	}

	aliasCache.mu.Lock()
	aliasCache.aliases[key] = dbAlias
	aliasCache.mu.Unlock()

	return dbAlias, true, nil
}

// setNewAliasCache updates the in-memory alias cache and persists it to the database.
func setNewAliasCache(aliasCache *AliasCache, chatJID, senderJID, alias string, db *AppDB) error {
	if aliasCache == nil {
		return nil
	}

	key := chatJID + "|" + senderJID

	aliasCache.mu.Lock()
	aliasCache.aliases[key] = alias
	aliasCache.mu.Unlock()

	if err := db.SetAlias(context.Background(), chatJID, senderJID, alias); err != nil {
		return err
	}

	return nil
}
