package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// total clicks per alias, broken down by who clicked
	URLClicksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "url_clicks_total",
			Help: "Total clicks per shortened URL",
		},
		[]string{"alias"},
	)

	// total clicks per alias, broken down by referrer domain
	URLClicksByReferrerTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "url_clicks_by_referrer_total",
			Help: "Total clicks per shortened URL broken down by referrer",
		},
		[]string{"alias", "referrer"},
	)

	// total clicks per alias, broken down by browser family
	URLClicksByUserAgentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "url_clicks_by_user_agent_total",
			Help: "Total clicks per shortened URL broken down by user agent",
		},
		[]string{"alias", "user_agent"},
	)

	// timestamp of the most recent click per alias
	URLLastClickTimestampSeconds = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "url_last_click_timestamp_seconds",
			Help: "Unix timestamp of the most recent click per alias",
		},
		[]string{"alias"},
	)
)