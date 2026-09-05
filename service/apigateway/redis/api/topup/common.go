package topup_cache

import "time"

const (
	topupAllCacheKey     = "apigw:topup:all:page:%d:pageSize:%d:search:%s"
	topupByCardCacheKey  = "apigw:topup:card_number:%s:page:%d:pageSize:%d:search:%s"
	topupByIdCacheKey    = "apigw:topup:id:%d"
	topupActiveCacheKey  = "apigw:topup:active:page:%d:pageSize:%d:search:%s"
	topupTrashedCacheKey = "apigw:topup:trashed:page:%d:pageSize:%d:search:%s"

	ttlDefault = 5 * time.Minute
)
