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

// syncMapInc atomically increments an *atomic.Int64 stored in a sync.Map.
// Safe for concurrent use; allocates a new counter only on first access for a key.
func syncMapInc(m *sync.Map, key any) {
	if v, ok := m.Load(key); ok {
		v.(*atomic.Int64).Add(1)
		return
	}
	newVal := &atomic.Int64{}
	newVal.Store(1)
	if actual, loaded := m.LoadOrStore(key, newVal); loaded {
		// Another goroutine stored a counter first — increment theirs.
		actual.(*atomic.Int64).Add(1)
	}
}

// secondBucket holds metrics for a single second of the test.
type secondBucket struct {
	requests     int64
	success      int64
	fail         int64
	totalLatency int64 // microseconds
	totalBytes   int64
	statusCodes  sync.Map // map[int]*atomic.Int64

	// Double-buffered histograms: Add() writes to histograms[activeHist],
	// Snapshot() swaps the active index, merges the retired histogram into
	// cumulative, then reads quantiles from cumulative without holding histMu.
	histograms [2]*hdrhistogram.Histogram
	activeHist atomic.Int32
	histMu     sync.Mutex
	cumulative *hdrhistogram.Histogram
}

// Monitor handles real-time metrics collection.
type Monitor struct {
	// Atomic counters — lock-free hot path.
	requests          int64
	success           int64
	fail              int64
	assertionFailures int64

	// sync.Map values are *atomic.Int64 for true atomic increments.
	statusCodes     sync.Map // map[int]*atomic.Int64
	errors          sync.Map // map[string]*atomic.Int64
	assertionErrors sync.Map // map[string]*atomic.Int64
	protocolCounts  sync.Map // map[string]*atomic.Int64

	totalBytes int64

	// Double-buffered global histogram.
	// Add() records into histograms[activeHist] under histMu.
	// Snapshot() swaps activeHist, merges the retired histogram into cumulative,
	// resets it, then reads quantiles from cumulative outside histMu.
	histograms [2]*hdrhistogram.Histogram
	activeHist atomic.Int32
	histMu     sync.Mutex
	cumulative *hdrhistogram.Histogram

	startTime time.Time

	secondBuckets []*secondBucket
	bucketMu      sync.RWMutex

	// Pre-allocated snapshot buffers — reused on every Snapshot() call to
	// eliminate repeated heap allocations (previously 5 fresh maps every 100ms).
	snapStatusMap    map[string]int
	snapErrorMap     map[string]int
	snapAssertionMap map[string]int
	snapProtocolMap  map[string]int
	snapTimeSeries   []models.SecondStats
}

func NewMonitor() *Monitor {
	return &Monitor{
		startTime: time.Now(),
		histograms: [2]*hdrhistogram.Histogram{
			hdrhistogram.New(1, 30000000, 3),
			hdrhistogram.New(1, 30000000, 3),
		},
		cumulative:    hdrhistogram.New(1, 30000000, 3),
		secondBuckets: make([]*secondBucket, 0),
		// Pre-allocate with reasonable initial capacities.
		snapStatusMap:    make(map[string]int, 8),
		snapErrorMap:     make(map[string]int, 16),
		snapAssertionMap: make(map[string]int, 16),
		snapProtocolMap:  make(map[string]int, 4),
		snapTimeSeries:   make([]models.SecondStats, 0, 300),
	}
}

func (m *Monitor) getOrCreateBucket(second int) *secondBucket {
	m.bucketMu.Lock()
	defer m.bucketMu.Unlock()

	for len(m.secondBuckets) <= second {
		m.secondBuckets = append(m.secondBuckets, &secondBucket{
			histograms: [2]*hdrhistogram.Histogram{
				hdrhistogram.New(1, 30000000, 3),
				hdrhistogram.New(1, 30000000, 3),
			},
			cumulative: hdrhistogram.New(1, 30000000, 3),
		})
	}
	return m.secondBuckets[second]
}

// Add records a single result. Called from a single goroutine (processResults).
func (m *Monitor) Add(res models.Result, isSuccess bool) {
	atomic.AddInt64(&m.requests, 1)
	atomic.AddInt64(&m.totalBytes, res.Bytes)

	hasAssertionError := res.AssertionError != nil
	if hasAssertionError {
		atomic.AddInt64(&m.assertionFailures, 1)
		syncMapInc(&m.assertionErrors, res.AssertionError.Error())
	}

	if isSuccess && !hasAssertionError {
		atomic.AddInt64(&m.success, 1)
	} else {
		atomic.AddInt64(&m.fail, 1)
	}

	// Classify transport timeouts as status 1 for grouping.
	if res.Status == 0 && res.Error != nil {
		if isTimeout(res.Error) {
			res.Status = 1
		}
	}

	syncMapInc(&m.statusCodes, res.Status)

	if res.Error != nil {
		syncMapInc(&m.errors, sanitizeError(res.Error.Error()))
	}

	if res.Protocol != "" {
		syncMapInc(&m.protocolCounts, res.Protocol)
	}

	latencyUs := res.Latency.Microseconds()

	// Record latency only for requests that received a response.
	if res.Error == nil {
		m.histMu.Lock()
		_ = m.histograms[m.activeHist.Load()].RecordValue(latencyUs)
		m.histMu.Unlock()
	}

	// Per-second tracking.
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

	syncMapInc(&bucket.statusCodes, res.Status)

	if res.Error == nil {
		bucket.histMu.Lock()
		_ = bucket.histograms[bucket.activeHist.Load()].RecordValue(latencyUs)
		bucket.histMu.Unlock()
	}
}

// GetStats returns current counters for circuit breaker checks.
func (m *Monitor) GetStats() (totalRequests, failures, assertionFailures int64) {
	return atomic.LoadInt64(&m.requests),
		atomic.LoadInt64(&m.fail),
		atomic.LoadInt64(&m.assertionFailures)
}

// Snapshot returns a consistent report of current metrics.
// Called from the dashboard tick goroutine (separate from Add's goroutine).
func (m *Monitor) Snapshot() models.Report {
	reqs := atomic.LoadInt64(&m.requests)
	succ := atomic.LoadInt64(&m.success)
	fail := atomic.LoadInt64(&m.fail)
	totalBytes := atomic.LoadInt64(&m.totalBytes)

	duration := time.Since(m.startTime).Seconds()
	rps := 0.0
	throughput := 0.0
	if duration > 0 {
		rps = float64(reqs) / duration
		throughput = float64(totalBytes) / duration / 1024 / 1024
	}

	successRate := 0.0
	if reqs > 0 {
		successRate = float64(succ) / float64(reqs) * 100
	}

	// Double-buffer swap: retire the current active histogram, merge it into
	// cumulative, then reset it. histMu is held only for swap+merge+reset
	// (fast), NOT for the subsequent quantile reads — that is the key improvement
	// over the previous single-histogram approach which held the lock during all
	// ValueAtQuantile calls.
	m.histMu.Lock()
	currentIdx := m.activeHist.Load()
	m.activeHist.Store(1 - currentIdx)
	m.cumulative.Merge(m.histograms[currentIdx])
	m.histograms[currentIdx].Reset()
	m.histMu.Unlock()

	// Read quantiles outside the lock — cumulative is only written here (under
	// histMu) and Snapshot is called from a single goroutine, so no race exists.
	h := m.cumulative
	p50 := time.Duration(h.ValueAtQuantile(50)) * time.Microsecond
	p75 := time.Duration(h.ValueAtQuantile(75)) * time.Microsecond
	p90 := time.Duration(h.ValueAtQuantile(90)) * time.Microsecond
	p95 := time.Duration(h.ValueAtQuantile(95)) * time.Microsecond
	p99 := time.Duration(h.ValueAtQuantile(99)) * time.Microsecond
	maxLat := time.Duration(h.Max()) * time.Microsecond
	minLat := time.Duration(h.Min()) * time.Microsecond

	// Reuse pre-allocated maps — clear entries without deallocating storage.
	clear(m.snapStatusMap)
	m.statusCodes.Range(func(key, value interface{}) bool {
		code := key.(int)
		var sKey string
		if code == 1 {
			sKey = "Timeout"
		} else {
			sKey = fmt.Sprintf("%d", code)
		}
		m.snapStatusMap[sKey] = int(value.(*atomic.Int64).Load())
		return true
	})

	clear(m.snapErrorMap)
	m.errors.Range(func(key, value interface{}) bool {
		m.snapErrorMap[key.(string)] = int(value.(*atomic.Int64).Load())
		return true
	})

	// Build time series — reuse slice backing array, extend only if needed.
	m.bucketMu.RLock()
	needed := len(m.secondBuckets)
	if cap(m.snapTimeSeries) < needed {
		m.snapTimeSeries = make([]models.SecondStats, needed, needed+64)
	} else {
		m.snapTimeSeries = m.snapTimeSeries[:needed]
	}

	for i, bucket := range m.secondBuckets {
		bucketReqs := atomic.LoadInt64(&bucket.requests)
		bucketSucc := atomic.LoadInt64(&bucket.success)
		bucketFail := atomic.LoadInt64(&bucket.fail)
		bucketLatency := atomic.LoadInt64(&bucket.totalLatency)

		avgLatency := 0.0
		if bucketReqs > 0 {
			avgLatency = float64(bucketLatency) / float64(bucketReqs) / 1000.0
		}

		// Per-second histogram double-buffer swap: same pattern as global histogram.
		bucket.histMu.Lock()
		bCurrentIdx := bucket.activeHist.Load()
		bucket.activeHist.Store(1 - bCurrentIdx)
		bucket.cumulative.Merge(bucket.histograms[bCurrentIdx])
		bucket.histograms[bCurrentIdx].Reset()
		bucket.histMu.Unlock()

		bh := bucket.cumulative
		bp50 := time.Duration(bh.ValueAtQuantile(50)) * time.Microsecond
		bp75 := time.Duration(bh.ValueAtQuantile(75)) * time.Microsecond
		bp90 := time.Duration(bh.ValueAtQuantile(90)) * time.Microsecond
		bp95 := time.Duration(bh.ValueAtQuantile(95)) * time.Microsecond
		bp99 := time.Duration(bh.ValueAtQuantile(99)) * time.Microsecond

		bucketStatusCodes := make(map[string]int)
		bucket.statusCodes.Range(func(key, value interface{}) bool {
			code := key.(int)
			var sKey string
			if code == 1 {
				sKey = "Timeout"
			} else {
				sKey = fmt.Sprintf("%d", code)
			}
			bucketStatusCodes[sKey] = int(value.(*atomic.Int64).Load())
			return true
		})

		m.snapTimeSeries[i] = models.SecondStats{
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

	clear(m.snapAssertionMap)
	m.assertionErrors.Range(func(key, value interface{}) bool {
		m.snapAssertionMap[key.(string)] = int(value.(*atomic.Int64).Load())
		return true
	})

	clear(m.snapProtocolMap)
	m.protocolCounts.Range(func(key, value interface{}) bool {
		m.snapProtocolMap[key.(string)] = int(value.(*atomic.Int64).Load())
		return true
	})

	// Note: snapshot maps are returned by reference and reused on the next call.
	// This is safe because the Bubble Tea model replaces m.report on every tick
	// before the next Snapshot() runs (single-threaded model update loop).
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
		Max:               maxLat,
		Min:               minLat,
		StatusCodes:       m.snapStatusMap,
		Errors:            m.snapErrorMap,
		AssertionErrors:   m.snapAssertionMap,
		ProtocolCounts:    m.snapProtocolMap,
		TimeSeriesData:    m.snapTimeSeries,
	}
}
