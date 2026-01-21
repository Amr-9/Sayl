package stats

import (
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Amr-9/sayl/pkg/models"
	"github.com/HdrHistogram/hdrhistogram-go"
)

// isTimeout checks if the error is a timeout
func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}
	if os.IsTimeout(err) {
		return true
	}
	return false
}

// secondBucket holds metrics for a single second
type secondBucket struct {
	requests     int64
	success      int64
	fail         int64
	totalLatency int64 // in microseconds
	totalBytes   int64
	histogram    *hdrhistogram.Histogram
	statusCodes  sync.Map
	mu           sync.Mutex
}

// Monitor handles real-time metrics collection using atomic counters and HDR Histogram
type Monitor struct {
	requests          int64
	success           int64
	fail              int64
	assertionFailures int64    // Separate counter for assertion failures
	statusCodes       sync.Map // map[int]int
	errors            sync.Map // map[string]int (network/server errors)
	assertionErrors   sync.Map // map[string]int (assertion failures)

	totalBytes int64

	mu        sync.Mutex
	histogram *hdrhistogram.Histogram

	startTime time.Time

	// Per-second tracking
	secondBuckets []*secondBucket
	bucketMu      sync.RWMutex
}

func NewMonitor() *Monitor {
	return &Monitor{
		startTime: time.Now(),
		// min 1µs, max 30s (in µs), 3 significant figures
		histogram:     hdrhistogram.New(1, 30000000, 3),
		secondBuckets: make([]*secondBucket, 0),
	}
}

func (m *Monitor) getOrCreateBucket(second int) *secondBucket {
	m.bucketMu.Lock()
	defer m.bucketMu.Unlock()

	// Extend slice if needed
	for len(m.secondBuckets) <= second {
		m.secondBuckets = append(m.secondBuckets, &secondBucket{
			histogram: hdrhistogram.New(1, 30000000, 3),
		})
	}
	return m.secondBuckets[second]
}

// Add records a single result into the monitor
func (m *Monitor) Add(res models.Result, isSuccess bool) {
	atomic.AddInt64(&m.requests, 1)
	atomic.AddInt64(&m.totalBytes, res.Bytes)

	// Track assertion failures separately from network failures
	hasAssertionError := res.AssertionError != nil
	if hasAssertionError {
		atomic.AddInt64(&m.assertionFailures, 1)
		// Track assertion error message
		errStr := res.AssertionError.Error()
		count, _ := m.assertionErrors.LoadOrStore(errStr, 0)
		m.assertionErrors.Store(errStr, count.(int)+1)
	}

	if isSuccess && !hasAssertionError {
		atomic.AddInt64(&m.success, 1)
	} else {
		atomic.AddInt64(&m.fail, 1)
	}

	// If request failed with 0 status but has an error, check if it's a timeout
	if res.Status == 0 && res.Error != nil {
		if isTimeout(res.Error) {
			res.Status = 1 // 1 for Timeout
		}
	}

	// Update status codes
	count, _ := m.statusCodes.LoadOrStore(res.Status, 0)
	m.statusCodes.Store(res.Status, count.(int)+1)

	// Update network/server errors (separate from assertion errors)
	if res.Error != nil {
		errStr := sanitizeError(res.Error.Error())
		count, _ := m.errors.LoadOrStore(errStr, 0)
		m.errors.Store(errStr, count.(int)+1)
	}

	// Update latencies in microseconds ONLY if it's not a transport error
	// We want to track latency for successful requests or server errors (e.g. 500),
	// but NOT for immediate transport failures (e.g. dial tcp: refused) which skew min latency.
	latencyUs := res.Latency.Microseconds()

	// If status is 0 and error is present, it's likely a transport error (timeout, connection refused etc)
	// We might have set Status=1 for timeout above, so check original error presence mostly.
	// Actually, simplified check: if we have a network error that prevented a response (Status < 100), skip latency.

	// Better logic: Record latency only if we got a response (Status > 0) OR if it's a specific interesting error?
	// The user specifically complained about 1ms min latency.
	// Converting Timeout to Status 1 happens above.
	// Let's rely on: if res.Error != nil, we generally don't trust the latency as "server response time".
	// BUT, we might want to track how long it took to fail.
	// The user's issue is "1ms" which implies immediate failure.
	// Let's skip recording if error != nil.

	if res.Error == nil {
		m.mu.Lock()
		_ = m.histogram.RecordValue(latencyUs)
		m.mu.Unlock()
	}

	// Per-second tracking
	second := int(time.Since(m.startTime).Seconds())
	bucket := m.getOrCreateBucket(second)

	atomic.AddInt64(&bucket.requests, 1)
	atomic.AddInt64(&bucket.totalLatency, latencyUs)
	atomic.AddInt64(&bucket.totalBytes, res.Bytes)
	if isSuccess && !hasAssertionError {
		atomic.AddInt64(&bucket.success, 1)
	} else {
		atomic.AddInt64(&bucket.fail, 1)
	}

	// Update per-second status codes
	cnt, _ := bucket.statusCodes.LoadOrStore(res.Status, 0)
	bucket.statusCodes.Store(res.Status, cnt.(int)+1)

	// Update per-second histogram if no error
	if res.Error == nil {
		bucket.mu.Lock()
		_ = bucket.histogram.RecordValue(latencyUs)
		bucket.mu.Unlock()
	}
}

// GetStats returns current counters for circuit breaker checks
func (m *Monitor) GetStats() (totalRequests, failures, assertionFailures int64) {
	return atomic.LoadInt64(&m.requests),
		atomic.LoadInt64(&m.fail),
		atomic.LoadInt64(&m.assertionFailures)
}

// Snapshot returns a current report of the metrics
func (m *Monitor) Snapshot() models.Report {
	reqs := atomic.LoadInt64(&m.requests)
	succ := atomic.LoadInt64(&m.success)
	fail := atomic.LoadInt64(&m.fail)

	totalBytes := atomic.LoadInt64(&m.totalBytes)

	duration := time.Since(m.startTime).Seconds()
	rps := 0.0
	throughput := 0.0 // MB/s
	if duration > 0 {
		rps = float64(reqs) / duration
		throughput = float64(totalBytes) / duration / 1024 / 1024
	}

	successRate := 0.0
	if reqs > 0 {
		successRate = float64(succ) / float64(reqs) * 100
	}

	m.mu.Lock()
	h := m.histogram
	p50 := time.Duration(h.ValueAtQuantile(50)) * time.Microsecond
	p75 := time.Duration(h.ValueAtQuantile(75)) * time.Microsecond
	p90 := time.Duration(h.ValueAtQuantile(90)) * time.Microsecond
	p95 := time.Duration(h.ValueAtQuantile(95)) * time.Microsecond
	p99 := time.Duration(h.ValueAtQuantile(99)) * time.Microsecond
	max := time.Duration(h.Max()) * time.Microsecond
	min := time.Duration(h.Min()) * time.Microsecond
	m.mu.Unlock()

	statusMap := make(map[string]int)
	m.statusCodes.Range(func(key, value interface{}) bool {
		code := key.(int)
		var sKey string
		if code == 1 {
			sKey = "Timeout"
		} else {
			sKey = fmt.Sprintf("%d", code)
		}
		statusMap[sKey] = value.(int)
		return true
	})

	errorMap := make(map[string]int)
	m.errors.Range(func(key, value interface{}) bool {
		errorMap[key.(string)] = value.(int)
		return true
	})

	// Build time series data
	m.bucketMu.RLock()
	timeSeriesData := make([]models.SecondStats, len(m.secondBuckets))
	for i, bucket := range m.secondBuckets {
		bucketReqs := atomic.LoadInt64(&bucket.requests)
		bucketSucc := atomic.LoadInt64(&bucket.success)
		bucketFail := atomic.LoadInt64(&bucket.fail)
		bucketLatency := atomic.LoadInt64(&bucket.totalLatency)

		avgLatency := 0.0
		if bucketReqs > 0 {
			avgLatency = float64(bucketLatency) / float64(bucketReqs) / 1000.0 // Convert to ms
		}

		bucket.mu.Lock()
		bh := bucket.histogram
		bp50 := time.Duration(bh.ValueAtQuantile(50)) * time.Microsecond
		bp75 := time.Duration(bh.ValueAtQuantile(75)) * time.Microsecond
		bp90 := time.Duration(bh.ValueAtQuantile(90)) * time.Microsecond
		bp95 := time.Duration(bh.ValueAtQuantile(95)) * time.Microsecond
		bp99 := time.Duration(bh.ValueAtQuantile(99)) * time.Microsecond
		bucket.mu.Unlock()

		bucketStatusCodes := make(map[string]int)
		bucket.statusCodes.Range(func(key, value interface{}) bool {
			code := key.(int)
			var sKey string
			if code == 1 {
				sKey = "Timeout"
			} else {
				sKey = fmt.Sprintf("%d", code)
			}
			bucketStatusCodes[sKey] = value.(int)
			return true
		})

		timeSeriesData[i] = models.SecondStats{
			Second:      i + 1,
			Requests:    bucketReqs,
			Success:     bucketSucc,
			Failures:    bucketFail,
			AvgLatency:  avgLatency,
			P50:         bp50,
			P75:         bp75,
			P90:         bp90,
			P95:         bp95,
			P99:         bp99,
			StatusCodes: bucketStatusCodes,
		}
	}
	m.bucketMu.RUnlock()

	// Collect assertion errors separately
	assertionErrorMap := make(map[string]int)
	m.assertionErrors.Range(func(key, value interface{}) bool {
		assertionErrorMap[key.(string)] = value.(int)
		return true
	})

	return models.Report{
		TotalRequests:     reqs,
		SuccessCount:      succ,
		FailureCount:      fail,
		AssertionFailures: atomic.LoadInt64(&m.assertionFailures),
		SuccessRate:       successRate,
		TotalBytes:        totalBytes,
		Throughput:        throughput,
		RPS:               rps,
		P50:               p50,
		P75:               p75,
		P90:               p90,
		P95:               p95,
		P99:               p99,
		Max:               max,
		Min:               min,
		StatusCodes:       statusMap,
		Errors:            errorMap,
		AssertionErrors:   assertionErrorMap,
		TimeSeriesData:    timeSeriesData,
	}
}
