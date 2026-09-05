package merchantdocument_cache

import "time"

const (
	merchantAllCacheKey = "apigw:merchant_document:all:page:%d:pageSize:%d:search:%s"

	merchantByIdCacheKey = "apigw:merchant_document:id:%d"

	merchantActiveCacheKey = "apigw:merchant_document:active:page:%d:pageSize:%d:search:%s"

	merchantTrashedCacheKey = "apigw:merchant_document:trashed:page:%d:pageSize:%d:search:%s"

	ttlDefault = 5 * time.Minute
)
