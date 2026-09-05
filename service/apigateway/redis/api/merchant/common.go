package merchant_cache

import "time"

const (
	merchantAllCacheKey = "apigw:merchant:all:page:%d:pageSize:%d:search:%s"

	merchantByIdCacheKey = "apigw:merchant:id:%d"

	merchantActiveCacheKey = "apigw:merchant:active:page:%d:pageSize:%d:search:%s"

	merchantTrashedCacheKey = "apigw:merchant:trashed:page:%d:pageSize:%d:search:%s"

	merchantByApiKeyCacheKey = "apigw:merchant:api_key:%s"

	merchantByUserIdCacheKey = "apigw:merchant:user_id:%d"

	merchantTransactionsCacheKey = "apigw:merchant:transaction:search:%s:page:%d:pageSize:%d"

	merchantTransactionApikeyCacheKey = "apigw:merchant:transaction:apikey:%s:search:%s:page:%d:pageSize:%d"

	merchantTransactionCacheKey = "apigw:merchant:transaction:merchant:%d:search:%s:page:%d:pageSize:%d"

	ttlDefault = 5 * time.Minute
)
