package card_dashboard_cache

import "time"

const (
	cacheKeyDashboardDefault    = "apigw:dashboard:card"
	cacheKeyDashboardCardNumber = "apigw:dashboard:card:number:%s"
	ttlDashboardDefault         = 5 * time.Minute
)
