package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	MessagesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "nerugaysya_messages_total",
			Help: "Total Telegram messages processed.",
		},
	)
	ProfaneTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "nerugaysya_profane_messages_total",
			Help: "Total messages detected as profane.",
		},
	)
	RepliesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "nerugaysya_replies_total",
			Help: "Total 'не ругайся' replies sent.",
		},
	)
	StatsRequestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "nerugaysya_stats_requests_total",
			Help: "Total /stats command requests.",
		},
	)
	DBErrorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "nerugaysya_db_errors_total",
			Help: "Total database errors.",
		},
	)
)

func init() {
	prometheus.MustRegister(MessagesTotal, ProfaneTotal, RepliesTotal, StatsRequestsTotal, DBErrorsTotal)
}

func startMetricsServer(addr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	_ = http.ListenAndServe(addr, mux)
}
