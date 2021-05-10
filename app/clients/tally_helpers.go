package clients

import (
	"github.com/go-masonry/mortar/interfaces/monitor"
	"github.com/uber-go/tally"
	"time"
)

type (
	cachedReporter struct {
		metrics monitor.Metrics
	}

	cachedCounter struct {
		counter monitor.Counter
	}

	cachedGauge struct {
		gauge monitor.Gauge
	}

	cachedTimer struct {
		timer monitor.Timer
	}

	cachedHistogram struct {
		histogram monitor.Histogram
	}

	cachedHistogramBucket struct {
		histogram  monitor.Histogram
		upperBound float64
	}
)

func createCachedReporter(metrics monitor.Metrics) *cachedReporter {
	return &cachedReporter{
		metrics: metrics,
	}
}

func (impl *cachedReporter) Capabilities() tally.Capabilities {
	return impl
}

func (impl *cachedReporter) Flush() {}

// AllocateCounter pre allocates a counter data structure with name & tags.
func (impl *cachedReporter) AllocateCounter(name string, tags map[string]string) tally.CachedCount {
	return &cachedCounter{impl.metrics.WithTags(tags).Counter(name, name)}
}

// AllocateGauge pre allocates a gauge data structure with name & tags.
func (impl *cachedReporter) AllocateGauge(name string, tags map[string]string) tally.CachedGauge {
	return &cachedGauge{impl.metrics.WithTags(tags).Gauge(name, name)}
}

// AllocateTimer pre allocates a timer data structure with name & tags.
func (impl *cachedReporter) AllocateTimer(name string, tags map[string]string) tally.CachedTimer {
	return &cachedTimer{impl.metrics.WithTags(tags).Timer(name, name)}
}

// AllocateHistogram pre allocates a histogram data structure with name, tags,
// value buckets and duration buckets.
func (impl *cachedReporter) AllocateHistogram(name string, tags map[string]string, buckets tally.Buckets) tally.CachedHistogram {
	monitorBuckets := make(monitor.Buckets, len(buckets.AsValues()))
	for i, value := range buckets.AsValues() {
		monitorBuckets[i] = value
	}
	return &cachedHistogram{impl.metrics.WithTags(tags).Histogram(name, name, monitorBuckets)}
}

func (impl *cachedReporter) Tagging() bool {
	return true
}

func (impl *cachedReporter) Reporting() bool {
	return true
}

func (impl *cachedCounter) ReportCount(value int64) {
	impl.counter.Add(float64(value))
}

func (impl *cachedGauge) ReportGauge(value float64) {
	impl.gauge.Set(value)
}

func (impl *cachedTimer) ReportTimer(interval time.Duration) {
	impl.timer.Record(interval)
}

func (impl *cachedHistogramBucket) ReportSamples(value int64) {
	for i := int64(0); i < value; i++ {
		impl.histogram.Record(impl.upperBound)
	}
}

func (impl *cachedHistogram) ValueBucket(_, bucketUpperBound float64) tally.CachedHistogramBucket {
	return &cachedHistogramBucket{histogram: impl.histogram, upperBound: bucketUpperBound}
}

func (impl *cachedHistogram) DurationBucket(_, bucketUpperBound time.Duration) tally.CachedHistogramBucket {
	upperBound := float64(bucketUpperBound) / float64(time.Second)
	return &cachedHistogramBucket{histogram: impl.histogram, upperBound: upperBound}
}
