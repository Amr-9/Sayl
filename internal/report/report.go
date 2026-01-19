package report

import (
	"fmt"
	"html/template"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Amr-9/sayl/pkg/models"
)

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sayl Load Test Report</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            min-height: 100vh;
            color: #e0e0e0;
            padding: 20px;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
        }
        .header {
            text-align: center;
            margin-bottom: 40px;
            padding: 30px;
            background: rgba(255,255,255,0.05);
            border-radius: 20px;
            backdrop-filter: blur(10px);
        }
        .header h1 {
            font-size: 3rem;
            background: linear-gradient(90deg, #00d9ff, #ff00ff);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            margin-bottom: 10px;
        }
        .header p {
            color: #888;
            font-size: 1.1rem;
        }
        .summary-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 40px;
        }
        .summary-card {
            background: rgba(255,255,255,0.08);
            border-radius: 15px;
            padding: 25px;
            text-align: center;
            border: 1px solid rgba(255,255,255,0.1);
            transition: transform 0.3s, box-shadow 0.3s;
        }
        .summary-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 10px 30px rgba(0,217,255,0.2);
        }
        .summary-card .value {
            font-size: 2.5rem;
            font-weight: bold;
            background: linear-gradient(90deg, #00d9ff, #00ff88);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        .summary-card .label {
            color: #888;
            margin-top: 10px;
            font-size: 0.9rem;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        .charts-grid {
            display: grid;
            grid-template-columns: repeat(2, 1fr);
            gap: 30px;
            margin-bottom: 40px;
        }
        @media (max-width: 1200px) {
            .charts-grid {
                grid-template-columns: 1fr;
            }
        }
        .chart-container {
            background: rgba(255,255,255,0.05);
            border-radius: 20px;
            padding: 25px;
            border: 1px solid rgba(255,255,255,0.1);
        }
        .chart-container h3 {
            margin-bottom: 20px;
            color: #00d9ff;
            font-size: 1.3rem;
        }
        .chart-wrapper {
            position: relative;
            height: 300px;
        }
        .status-table {
            background: rgba(255,255,255,0.05);
            border-radius: 20px;
            padding: 25px;
            border: 1px solid rgba(255,255,255,0.1);
        }
        .status-table h3 {
            margin-bottom: 20px;
            color: #00d9ff;
        }
        table {
            width: 100%;
            border-collapse: collapse;
        }
        th, td {
            padding: 15px;
            text-align: left;
            border-bottom: 1px solid rgba(255,255,255,0.1);
        }
        th {
            color: #00d9ff;
            font-weight: 600;
            text-transform: uppercase;
            font-size: 0.85rem;
            letter-spacing: 1px;
        }
        tr:hover {
            background: rgba(255,255,255,0.05);
        }
        .success-badge {
            background: linear-gradient(90deg, #00ff88, #00d9ff);
            color: #1a1a2e;
            padding: 5px 15px;
            border-radius: 20px;
            font-weight: bold;
            font-size: 0.85rem;
        }
        .error-badge {
            background: linear-gradient(90deg, #ff4757, #ff6b81);
            color: white;
            padding: 5px 15px;
            border-radius: 20px;
            font-weight: bold;
            font-size: 0.85rem;
        }
        .footer {
            text-align: center;
            padding: 30px;
            color: #666;
            font-size: 0.9rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⚡ Sayl Load Test Report</h1>
            <p>Generated at {{.GeneratedAt}}</p>
            <div style="margin-top: 20px; padding: 15px; background: rgba(0,0,0,0.2); border-radius: 10px; display: inline-block;">
                <div style="font-size: 1.2rem; margin-bottom: 5px;">
                    <span style="color: #00d9ff; font-weight: bold;">{{.Method}}</span> 
                    <a href="{{.TargetURL}}" style="color: #fff; text-decoration: none; border-bottom: 1px dotted #00ff88;" target="_blank">{{.TargetURL}}</a>
                </div>
                <div style="color: #888; font-size: 0.9rem;">
                    Duration: <span style="color: #00ff88">{{.TestDuration}}</span> • 
                    Concurrency: <span style="color: #00ff88">{{.Concurrency}}</span> workers
                </div>
            </div>
        </div>

        <div class="summary-grid">
            <div class="summary-card">
                <div class="value">{{.TotalRequests}}</div>
                <div class="label">Total Requests</div>
            </div>
            <div class="summary-card">
                <div class="value">{{printf "%.1f" .SuccessRate}}%</div>
                <div class="label">Success Rate</div>
            </div>
            <div class="summary-card">
                <div class="value">{{printf "%.0f" .RPS}}</div>
                <div class="label">Requests/sec</div>
            </div>
            <div class="summary-card">
                <div class="value">{{.Min}}</div>
                <div class="label">Min Latency</div>
            </div>
            <div class="summary-card">
                <div class="value">{{.P50}}</div>
                <div class="label">P50 Latency</div>
            </div>
            <div class="summary-card">
                <div class="value">{{.P99}}</div>
                <div class="label">P99 Latency</div>
            </div>
            <div class="summary-card">
                <div class="value">{{.Max}}</div>
                <div class="label">Max Latency</div>
            </div>
            <div class="summary-card">
                <div class="value">{{.SuccessCount}}</div>
                <div class="label">Successful</div>
            </div>
        </div>

        <div class="charts-grid">
            <div class="chart-container">
                <h3>📈 Requests Per Second (RPS)</h3>
                <div class="chart-wrapper">
                    <canvas id="rpsChart"></canvas>
                </div>
            </div>
            <div class="chart-container">
                <h3>⏱️ Latency Percentiles (ms)</h3>
                <div class="chart-wrapper">
                    <canvas id="latencyChart"></canvas>
                </div>
            </div>
            <div class="chart-container">
                <h3>✅ Success vs Failure</h3>
                <div class="chart-wrapper">
                    <canvas id="successChart"></canvas>
                </div>
            </div>
            <div class="chart-container">
                <h3>🔢 Status Code Distribution</h3>
                <div class="chart-wrapper">
                    <canvas id="statusChart"></canvas>
                </div>
            </div>
        </div>

        <div class="status-table">
            <h3>📊 Status Codes Breakdown</h3>
            <table>
                <thead>
                    <tr>
                        <th>Status Code</th>
                        <th>Count</th>
                        <th>Percentage</th>
                        <th>Status</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .StatusCodesTable}}
                    <tr>
                        <td>{{.Code}}</td>
                        <td>{{.Count}}</td>
                        <td>{{printf "%.2f" .Percentage}}%</td>
                        <td>
                            {{if .IsSuccess}}
                            <span class="success-badge">Success</span>
                            {{else}}
                            <span class="error-badge">Error</span>
                            {{end}}
                        </td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>

        {{if .Errors}}
        <div class="status-table" style="margin-top: 30px; border-color: rgba(255, 71, 87, 0.3);">
            <h3 style="color: #ff4757;">⚠️ Error Distribution</h3>
            <table>
                <thead>
                    <tr>
                        <th style="color: #ff4757;">Error Message</th>
                        <th style="color: #ff4757;">Count</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Errors}}
                    <tr>
                        <td style="color: #ff6b81; font-family: monospace;">{{.Message}}</td>
                        <td>{{.Count}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
        {{end}}

        <div class="footer">
            <p>Generated by Sayl - High-Performance Load Testing Tool</p>
        </div>
    </div>

    <script>
        // Chart.js global configuration
        Chart.defaults.color = '#888';
        Chart.defaults.borderColor = 'rgba(255,255,255,0.1)';

        // Time series data
        const timeLabels = [{{.TimeLabels}}];
        const rpsData = [{{.RPSData}}];
        const p50Data = [{{.P50Data}}];
        const p90Data = [{{.P90Data}}];
        const p95Data = [{{.P95Data}}];
        const p99Data = [{{.P99Data}}];
        const successData = [{{.SuccessData}}];
        const failureData = [{{.FailureData}}];

        // RPS Chart
        new Chart(document.getElementById('rpsChart'), {
            type: 'line',
            data: {
                labels: timeLabels,
                datasets: [{
                    label: 'RPS',
                    data: rpsData,
                    borderColor: '#00d9ff',
                    backgroundColor: 'rgba(0,217,255,0.1)',
                    fill: true,
                    tension: 0.4,
                    pointRadius: 3,
                    pointHoverRadius: 6
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { display: false }
                },
                scales: {
                    y: { beginAtZero: true, grid: { color: 'rgba(255,255,255,0.05)' } },
                    x: { grid: { color: 'rgba(255,255,255,0.05)' } }
                }
            }
        });

        // Latency Chart
        new Chart(document.getElementById('latencyChart'), {
            type: 'line',
            data: {
                labels: timeLabels,
                datasets: [
                    { label: 'P50', data: p50Data, borderColor: '#00ff88', tension: 0.4, pointRadius: 2 },
                    { label: 'P90', data: p90Data, borderColor: '#ffbb00', tension: 0.4, pointRadius: 2 },
                    { label: 'P95', data: p95Data, borderColor: '#ff6b6b', tension: 0.4, pointRadius: 2 },
                    { label: 'P99', data: p99Data, borderColor: '#ff00ff', tension: 0.4, pointRadius: 2 }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { position: 'top', labels: { usePointStyle: true } }
                },
                scales: {
                    y: { beginAtZero: true, grid: { color: 'rgba(255,255,255,0.05)' } },
                    x: { grid: { color: 'rgba(255,255,255,0.05)' } }
                }
            }
        });

        // Success/Failure Chart
        new Chart(document.getElementById('successChart'), {
            type: 'bar',
            data: {
                labels: timeLabels,
                datasets: [
                    { label: 'Success', data: successData, backgroundColor: '#00ff88' },
                    { label: 'Failure', data: failureData, backgroundColor: '#ff4757' }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { position: 'top', labels: { usePointStyle: true } }
                },
                scales: {
                    x: { stacked: true, grid: { color: 'rgba(255,255,255,0.05)' } },
                    y: { stacked: true, beginAtZero: true, grid: { color: 'rgba(255,255,255,0.05)' } }
                }
            }
        });

        // Status Code Chart
        new Chart(document.getElementById('statusChart'), {
            type: 'doughnut',
            data: {
                labels: [{{.StatusLabels}}],
                datasets: [{
                    data: [{{.StatusData}}],
                    backgroundColor: ['#00ff88', '#00d9ff', '#ffbb00', '#ff6b6b', '#ff00ff', '#6c5ce7']
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { position: 'right', labels: { usePointStyle: true } }
                }
            }
        });
    </script>
</body>
</html>`

// StatusCodeRow represents a row in the status codes table
type StatusCodeRow struct {
	Code       string
	Count      int
	Percentage float64
	IsSuccess  bool
}

// ErrorRow represents a row in the errors table
type ErrorRow struct {
	Message string
	Count   int
}

// TemplateData holds all data for the HTML template
type TemplateData struct {
	GeneratedAt      string
	TargetURL        string
	Method           string
	TestDuration     string
	Concurrency      int
	TotalRequests    int64
	SuccessCount     int64
	FailureCount     int64
	SuccessRate      float64
	RPS              float64
	P50              string
	P90              string
	P95              string
	P99              string
	Max              string
	Min              string
	StatusCodesTable []StatusCodeRow
	Errors           []ErrorRow
	TimeLabels       template.JS
	RPSData          template.JS
	P50Data          template.JS
	P90Data          template.JS
	P95Data          template.JS
	P99Data          template.JS
	SuccessData      template.JS
	FailureData      template.JS
	StatusLabels     template.JS
	StatusData       template.JS
}

// GenerateHTML creates an HTML report file with charts
func GenerateHTML(report models.Report, filename string) error {
	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Build time series arrays
	var timeLabels, rpsData, p50Data, p90Data, p95Data, p99Data, successData, failureData []string

	for _, s := range report.TimeSeriesData {
		timeLabels = append(timeLabels, fmt.Sprintf("'%ds'", s.Second))
		rpsData = append(rpsData, fmt.Sprintf("%d", s.Requests))
		p50Data = append(p50Data, fmt.Sprintf("%.2f", float64(s.P50.Milliseconds())))
		p90Data = append(p90Data, fmt.Sprintf("%.2f", float64(s.P90.Milliseconds())))
		p95Data = append(p95Data, fmt.Sprintf("%.2f", float64(s.P95.Milliseconds())))
		p99Data = append(p99Data, fmt.Sprintf("%.2f", float64(s.P99.Milliseconds())))
		successData = append(successData, fmt.Sprintf("%d", s.Success))
		failureData = append(failureData, fmt.Sprintf("%d", s.Failures))
	}

	// Build status code arrays
	var statusLabels, statusData []string
	var statusRows []StatusCodeRow

	// Sort status codes for consistent display
	var codes []string
	for code := range report.StatusCodes {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	for _, code := range codes {
		count := report.StatusCodes[code]
		percentage := float64(count) / float64(report.TotalRequests) * 100

		var isSuccess bool
		var codeInt int
		n, _ := fmt.Sscanf(code, "%d", &codeInt)
		if n > 0 {
			isSuccess = codeInt >= 200 && codeInt < 300
		} else {
			isSuccess = false
		}

		statusLabels = append(statusLabels, fmt.Sprintf("'%s'", code))
		statusData = append(statusData, fmt.Sprintf("%d", count))
		statusRows = append(statusRows, StatusCodeRow{
			Code:       code,
			Count:      count,
			Percentage: percentage,
			IsSuccess:  isSuccess,
		})
	}

	// Build errors table
	var errorRows []ErrorRow
	for msg, count := range report.Errors {
		errorRows = append(errorRows, ErrorRow{
			Message: msg,
			Count:   count,
		})
	}
	// Sort errors by count desc
	sort.Slice(errorRows, func(i, j int) bool {
		return errorRows[i].Count > errorRows[j].Count
	})

	data := TemplateData{
		GeneratedAt:      time.Now().Format("2006-01-02 15:04:05"),
		TargetURL:        report.TargetURL,
		Method:           report.Method,
		TestDuration:     report.Duration.String(),
		Concurrency:      report.Concurrency,
		TotalRequests:    report.TotalRequests,
		SuccessCount:     report.SuccessCount,
		FailureCount:     report.FailureCount,
		SuccessRate:      report.SuccessRate,
		RPS:              report.RPS,
		P50:              formatDuration(report.P50),
		P90:              formatDuration(report.P90),
		P95:              formatDuration(report.P95),
		P99:              formatDuration(report.P99),
		Max:              formatDuration(report.Max),
		Min:              formatDuration(report.Min),
		StatusCodesTable: statusRows,
		Errors:           errorRows,
		TimeLabels:       template.JS(strings.Join(timeLabels, ",")),
		RPSData:          template.JS(strings.Join(rpsData, ",")),
		P50Data:          template.JS(strings.Join(p50Data, ",")),
		P90Data:          template.JS(strings.Join(p90Data, ",")),
		P95Data:          template.JS(strings.Join(p95Data, ",")),
		P99Data:          template.JS(strings.Join(p99Data, ",")),
		SuccessData:      template.JS(strings.Join(successData, ",")),
		FailureData:      template.JS(strings.Join(failureData, ",")),
		StatusLabels:     template.JS(strings.Join(statusLabels, ",")),
		StatusData:       template.JS(strings.Join(statusData, ",")),
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.0fµs", float64(d.Microseconds()))
	}
	if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
