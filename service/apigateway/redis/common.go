package mencache

import "time"

const (
	ttlDefault = 5 * time.Minute

	cacheMerchantKey = "apigw:merchant_api_key:%s"
	cacheRoleKey     = "apigw:user_roles:%s"
)
