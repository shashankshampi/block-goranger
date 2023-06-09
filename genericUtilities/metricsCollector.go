package genericUtilities

import (
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"goranger/config"
	"goranger/constants"
	"net/http"
	"time"
)

var uniqueMetrics = make(map[string]prometheus.Counter)
var uniqueMetrics_Latency = make(map[string]prometheus.Summary)
var uniqueMetrics_Gauge = make(map[string]prometheus.Gauge)

func Prometheuscollect(port string) {
	if config.IsPrometheusEnabled() {
		p := port
		glog.Infoln("PROMETHEUS_COLLECTION STARTED")
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":"+p, nil)
		glog.Infoln("PROMETHEUS_COLLECTED")
	}
}

func RegisterTasksToPrometheus(taskName string) {
	if config.IsPrometheusEnabled() {
		registerThroughputTaskToProm(taskName, "Total no of requests processed")
		registerThroughputTaskToProm(taskName+constants.REQUEST_ERROR_COUNT, "Total no of 500 errors")
		registerThroughputTaskToProm(taskName+constants.RESPONSE_NOT_EXPECTED_COUNT, "Total no of unexpected response")
		registerThroughputTaskToProm(taskName+constants.REQUEST_CACHE_MISS, "Total no of cache misses")
		registerLatencyTaskToProm(taskName+constants.REQUEST_LATENCY, "Latency of every api request")
		registerConcurrencyTaskToProm(taskName+constants.PRODUCER_COUNT, "Total no of producers")
		registerConcurrencyTaskToProm(taskName+constants.EXECUTOR_COUNT, "Total no of executors")
		registerLatencyTaskToProm(taskName+constants.PRODUCER_LATENCY, "Latency of producers")
		registerLatencyTaskToProm(taskName+constants.EXECUTOR_LATENCY, "Executor of producers")
		registerThroughputTaskToProm(taskName+constants.REDIS_ERROR_COUNT, "Total no errors in redis")
		registerThroughputTaskToProm(taskName+constants.REDIS_PUSH_COUNT, "Total no push in redis")
		registerThroughputTaskToProm(taskName+constants.REDIS_GET_COUNT, "Total no get in redis")
		registerLatencyTaskToProm(taskName+constants.REDIS_PUSH_LATENCY, "Latency of redis push call")
		registerLatencyTaskToProm(taskName+constants.REDIS_GET_LATENCY, "Latency of redis get call")
	}

}

func registerThroughputTaskToProm(metricName string, helpText string) {

	opsProcessed := promauto.NewCounter(prometheus.CounterOpts{
		Name: metricName,
		Help: helpText,
	})
	uniqueMetrics[metricName] = opsProcessed
	glog.Infoln("TASKNAME:", metricName)
}

func registerLatencyTaskToProm(metricName string, helpText string) {
	latency := promauto.NewSummary(prometheus.SummaryOpts{
		Name: metricName,
		Help: helpText,
	})
	uniqueMetrics_Latency[metricName] = latency
	glog.Infoln("TASKNAME:", metricName)
}

func registerConcurrencyTaskToProm(metricName string, helpText string) {
	gauge := promauto.NewGauge(prometheus.GaugeOpts{
		Name: metricName,
		Help: helpText,
	})
	uniqueMetrics_Gauge[metricName] = gauge
	glog.Infoln("TASKNAME:", metricName)
}

func Increment(metricName string) {
	if config.IsPrometheusEnabled() {
		go func(metricName string) {
			uniqueMetrics[metricName].Inc()
		}(metricName)
	}
}

func PushLatency(metricsName string, startTime time.Time) {
	if config.IsPrometheusEnabled() {
		go func(metricsName string, startTime time.Time) {
			endtime := time.Now()
			diff := endtime.Sub(startTime).Seconds()
			uniqueMetrics_Latency[metricsName].Observe(diff)
		}(metricsName, startTime)
	}
}

func PushCurrentCount(metricsName string, count int) {
	if config.IsPrometheusEnabled() {
		go func(metricName string) {
			uniqueMetrics_Gauge[metricName].Set(float64(count))
		}(metricsName)
	}
}
