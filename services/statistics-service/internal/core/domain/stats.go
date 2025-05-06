package domain

type UserOrderStats struct {
	TotalOrders        int64
	HourlyDistribution map[string]int64
}

type UserStats struct {
	TotalItemsPurchased  int64
	AverageOrderValue    float64
	MostPurchasedItem    string
	TotalCompletedOrders int64
}
