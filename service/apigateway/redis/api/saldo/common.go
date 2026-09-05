package saldo_cache

import "time"

const (
	saldoAllCacheKey     = "apigw:saldo:all:page:%d:pageSize:%d:search:%s"
	saldoActiveCacheKey  = "apigw:saldo:active:page:%d:pageSize:%d:search:%s"
	saldoTrashedCacheKey = "apigw:saldo:trashed:page:%d:pageSize:%d:search:%s"
	saldoByIdCacheKey    = "apigw:saldo:id:%d"
	saldoByCardNumberKey = "apigw:saldo:card_number:%s"

	ttlDefault = 5 * time.Minute
)
