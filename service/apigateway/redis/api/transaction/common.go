package transaction_cache

import "time"

const (
	transactionAllCacheKey          = "apigw:transaction:all:page:%d:pageSize:%d:search:%s"
	transactionByIdCacheKey         = "apigw:transaction:id:%d"
	transactionActiveCacheKey       = "apigw:transaction:active:page:%d:pageSize:%d:search:%s"
	transactionTrashedCacheKey      = "apigw:transaction:trashed:page:%d:pageSize:%d:search:%s"
	transactionByCardCacheKey       = "apigw:transaction:card_number:%s:page:%d:pageSize:%d:search:%s"
	transactionByMerchantIdCacheKey = "apigw:transaction:merchant_id:%d"

	ttlDefault = 5 * time.Minute
)
