package card_cache

import "time"

const (
	ttlDefault = 5 * time.Minute

	cardAllCacheKey       = "apigw:card:all:page:%d:pageSize:%d:search:%s"
	cardByIdCacheKey      = "apigw:card:id:%d"
	cardActiveCacheKey    = "apigw:card:active:page:%d:pageSize:%d:search:%s"
	cardTrashedCacheKey   = "apigw:card:trashed:page:%d:pageSize:%d:search:%s"
	cardByUserIdCacheKey  = "apigw:card:user_id:%d"
	cardByCardNumCacheKey = "apigw:card:card_number:%s"
)
