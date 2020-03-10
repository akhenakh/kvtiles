package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	versionGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "insided",
		Name:      "version",
		Help:      "App version.",
	}, []string{"version"})

	dataVersionGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "insided",
		Name:      "dataset_version",
		Help:      "Dataset version.",
	}, []string{"version"})
)
