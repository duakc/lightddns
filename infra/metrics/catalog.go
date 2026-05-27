package metrics

import "github.com/prometheus/client_golang/prometheus"

type counterDef struct {
	name   string
	help   string
	labels []string
}

type gaugeDef struct {
	name   string
	help   string
	labels []string
}

type histogramDef struct {
	name    string
	help    string
	labels  []string
	buckets []float64
}

var (
	acquiredCounters   []counterDef
	acquiredGauges     []gaugeDef
	acquiredHistograms []histogramDef
)

func RegisterCounter(name, help string, labels []string) {
	acquiredCounters = append(acquiredCounters, counterDef{name: name, help: help, labels: labels})
}

func RegisterGauge(name, help string, labels []string) {
	acquiredGauges = append(acquiredGauges, gaugeDef{name: name, help: help, labels: labels})
}

func RegisterHistogram(name, help string, labels []string, buckets []float64) {
	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}
	acquiredHistograms = append(acquiredHistograms, histogramDef{name: name, help: help, labels: labels, buckets: buckets})
}
