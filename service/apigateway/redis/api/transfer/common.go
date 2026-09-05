package transfer_cache

import "time"

const (
	transferAllCacheKey     = "apigw:transfer:all:page:%d:pageSize:%d:search:%s"
	transferByIdCacheKey    = "apigw:transfer:id:%d"
	transferActiveCacheKey  = "apigw:transfer:active:page:%d:pageSize:%d:search:%s"
	transferTrashedCacheKey = "apigw:transfer:trashed:page:%d:pageSize:%d:search:%s"

	transferByFromCacheKey = "apigw:transfer:from_card_number:%s:"
	transferByToCacheKey   = "apigw:transfer:to_card_number:%s"

	ttlDefault = 5 * time.Minute
)
