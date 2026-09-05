package withdraw_cache

import "time"

const (
	withdrawAllCacheKey     = "apigw:withdraws:all:page:%d:pageSize:%d:search:%s"
	withdrawByCardCacheKey  = "apigw:withdraws:card_number:%s:page:%d:pageSize:%d:search:%s"
	withdrawByIdCacheKey    = "apigw:withdraws:id:%d"
	withdrawActiveCacheKey  = "apigw:withdraws:active:page:%d:pageSize:%d:search:%s"
	withdrawTrashedCacheKey = "apigw:withdraws:trashed:page:%d:pageSize:%d:search:%s"

	ttlDefault = 5 * time.Minute
)
