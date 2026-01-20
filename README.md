# 🌊 Sayl - High-Performance HTTP Load Testing Tool

<p align="center">
  <img src="https://img.shields.io/badge/🌊_Sayl-Ride_the_Wave_of_Load_Testing-0077B6?style=for-the-badge" alt="Sayl Logo">
</p>

<p align="center">
  <strong>🚀 Modern • ⚡ Fast • 🎨 Beautiful</strong><br>
  <em>A professional-grade load testing tool with an interactive TUI and powerful YAML configuration</em>
</p>

<p align="center">
  <img alt="Go" src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go">
  <img alt="License" src="https://img.shields.io/badge/license-GPL--3.0-blue?style=flat-square">
  <img alt="Platform" src="https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey?style=flat-square">
  <img alt="Release" src="https://img.shields.io/github/v/release/Amr-9/sayl?style=flat-square&color=green">
  <img alt="Stars" src="https://img.shields.io/github/stars/Amr-9/sayl?style=flat-square&color=yellow">
</p>

<p align="center">
  <a href="#-quick-start">🏃 Quick Start</a> •
  <a href="#-features">✨ Features</a> •
  <a href="#-yaml-configuration-guide">📘 YAML Guide</a> •
  <a href="#-examples">📂 Examples</a> •
  <a href="#-contributing">🤝 Contributing</a>
</p>

---

## 📖 Table of Contents

- [✨ Features](#-features)
- [🆚 Why Sayl?](#-why-sayl)
- [📦 Installation](#-installation)
- [🏃 Quick Start](#-quick-start)
- [🛠️ Usage Workflows](#️-usage-workflows)
- [📘 YAML Configuration Guide](#-yaml-configuration-guide)
  - [🎯 Target Section](#-target-section)
  - [⚙️ Load Section](#️-load-section)
  - [🔗 Steps Section](#-steps-section)
  - [📁 Data Section](#-data-section)
- [🎲 Dynamic Variables](#-dynamic-variables)
- [🔗 Chained Scenarios](#-chained-scenarios)
- [📁 Data Feeding (CSV)](#-data-feeding-csv)
- [📈 Load Stages (Ramping)](#-load-stages-ramping)
- [🔧 CLI Flags Reference](#-cli-flags-reference)
- [📊 Output & Reports](#-output--reports)
- [📂 Examples Gallery](#-examples-gallery)
- [🏗️ Architecture](#️-architecture)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)

---

## ✨ Features

### 🖥️ Interactive TUI (Terminal User Interface)
Stop wrestling with complex CLI flags! Sayl's TUI guides you through test setup with a visual wizard:
- 📊 **Live Dashboard** with real-time metrics, sparkline charts, and status code distribution
- 🎛️ **Visual Configuration** for URL, method, headers, rate, and duration
- 📈 **Progress Tracking** with latency histograms and success rates
- 🎨 **Beautiful Styling** with colors and modern UI elements

### 📄 Powerful YAML Configuration
Define your test scenarios in simple, readable YAML files:
- 📝 **Human Readable** - No programming knowledge required
- 🔄 **Version Control Friendly** - Commit and review with your team
- 🤖 **CI/CD Ready** - Run automated benchmarks in pipelines
- 🧩 **Template Variables** - Inject dynamic data anywhere

### 🔗 Chained Scenarios (Multi-Step Flows)
Go beyond simple endpoint hitting. Create complex user flows:
1. 🔐 **Login** to get a token
2. 📤 **Extract** the token from the response (JSON or Header)
3. 🔑 **Use** the token in subsequent authenticated requests

### 🎲 Built-in Dynamic Data Generators
Test with realistic data using built-in variables - no external tools needed:
```yaml
body: '{"email": "{{random_email}}", "id": "{{uuid}}"}'
```

### 📈 Smart Load Ramping (Stages)
Simulate real-world traffic patterns with gradual ramp-up:
```yaml
stages:
  - duration: 30s   # 🌅 Warm up
    target: 10
  - duration: 2m    # 🔥 Peak load
    target: 500
  - duration: 30s   # 🌙 Cool down
    target: 0
```

### 🛡️ Reliability Features
- 🔄 **Automatic Retries** with exponential backoff for transient errors
- 🛑 **Graceful Shutdown** - Ctrl+C saves all data before exit
- 🦺 **Panic Recovery** - Never crash unexpectedly
- ✅ **Preflight Checks** - Verify target connectivity before testing

### 📊 Rich Reporting
- 🖨️ **Console Summary** with colored metrics
- 📋 **JSON Reports** for programmatic processing
- 📈 **Interactive HTML Reports** with charts and visualizations

---

## 🆚 Why Sayl?

| Feature | 🌊 Sayl | 🥕 Vegeta | 🏎️ K6 | 🦗 Locust |
| :--- | :---: | :---: | :---: | :---: |
| **🖥️ Primary Interface** | **TUI + YAML** | CLI + Pipes | JS Scripting | Python |
| **📚 Ease of Use** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ | ⭐⭐ |
| **⏱️ Learning Curve** | Minutes | Hours | Days | Days |
| **🔗 Complex Scenarios** | ✅ YAML Config | ❌ Single Endpoint | ✅ JS Scripts | ✅ Python |
| **🎲 Dynamic Variables** | ✅ Built-in | ❌ External Tools | ✅ Programmatic | ✅ Programmatic |
| **📊 Real-time Dashboard** | ✅ Rich TUI | ❌ Basic Text | ❌ Console Only | ✅ Web UI |
| **🔄 Auto Retry** | ✅ Built-in | ❌ Manual | ❌ Manual | ❌ Manual |
| **🤖 CI/CD Ready** | ✅ YAML Files | ✅ Pipes | ✅ Scripts | ✅ Scripts |
| **💻 No Coding Required** | ✅ | ✅ | ❌ | ❌ |

### 🎯 Choose Sayl when you want:
- ✅ Visual feedback without sacrificing performance
- ✅ Complex scenarios without writing code
- ✅ Quick setup for ad-hoc testing
- ✅ Professional reports for stakeholders

---

## 📦 Installation

### ⚡ Quick Download (Recommended)

> **🎉 No installation required!** Just download the binary and run it immediately.

The fastest way to get started is to download a pre-built binary directly from **GitHub Releases**:

<p align="center">
  <a href="https://github.com/Amr-9/sayl/releases/latest">
    <img src="https://img.shields.io/badge/📥_Download_Latest_Release-0077B6?style=for-the-badge" alt="Download">
  </a>
</p>

---

### 🪟 Windows Installation

**Option 1: Direct Download (Easiest)**
1. 🌐 Go to [**GitHub Releases Page**](https://github.com/Amr-9/sayl/releases/latest)
2. 📥 Click on `Sayl-windows-amd64.exe` to download
3. 📂 Move the file to your desired folder
4. ✅ Double-click or run from terminal - **That's it!**

**Option 2: Using PowerShell**
```powershell
# 📥 Download the latest release
Invoke-WebRequest -Uri "https://github.com/Amr-9/sayl/releases/latest/download/Sayl-windows-amd64.exe" -OutFile "sayl.exe"

# ✅ Run it
./sayl
```

**Option 3: Using curl**
```bash
curl -LO https://github.com/Amr-9/sayl/releases/latest/download/Sayl-windows-amd64.exe
```

---

### 🐧 Linux Installation

**Option 1: Direct Download**
1. 🌐 Go to [**GitHub Releases Page**](https://github.com/Amr-9/sayl/releases/latest)
2. 📥 Click on `Sayl-linux-amd64` to download
3. 🔓 Make it executable: `chmod +x Sayl-linux-amd64`
4. ✅ Run it: `./Sayl-linux-amd64`

**Option 2: Using Terminal (One-liner)**
```bash
# 📥 Download, make executable, and run
curl -LO https://github.com/Amr-9/sayl/releases/latest/download/Sayl-linux-amd64 && \
chmod +x Sayl-linux-amd64 && \
./Sayl-linux-amd64
```

**Option 3: Install System-wide**
```bash
# 📥 Download
curl -LO https://github.com/Amr-9/sayl/releases/latest/download/Sayl-linux-amd64

# 🔓 Make executable
chmod +x Sayl-linux-amd64

# 📂 Move to system path (requires sudo)
sudo mv Sayl-linux-amd64 /usr/local/bin/sayl

# ✅ Now you can run from anywhere
sayl --help
```

---

### 🍎 macOS Installation

**Option 1: Direct Download**
1. 🌐 Go to [**GitHub Releases Page**](https://github.com/Amr-9/sayl/releases/latest)
2. 📥 Click on `Sayl-macos-amd64` (Intel) or `Sayl-macos-arm64` (Apple Silicon M1/M2/M3)
3. 🔓 Make it executable: `chmod +x Sayl-macos-*`
4. ✅ Run it: `./Sayl-macos-*`

**Option 2: Using Terminal**
```bash
# 📥 For Intel Macs
curl -LO https://github.com/Amr-9/sayl/releases/latest/download/Sayl-macos-amd64
chmod +x Sayl-macos-amd64

# 📥 For Apple Silicon (M1/M2/M3)
curl -LO https://github.com/Amr-9/sayl/releases/latest/download/Sayl-macos-arm64
chmod +x Sayl-macos-arm64
```

> **⚠️ macOS Security Note:** If you see "cannot be opened because the developer cannot be verified", run:
> ```bash
> xattr -d com.apple.quarantine Sayl-macos-*
> ```

---

### 📋 Available Downloads

Visit the [**Releases Page**](https://github.com/Amr-9/sayl/releases/latest) to see all available downloads:

| 📦 File | 🖥️ Platform | 🏗️ Architecture |
| :--- | :--- | :--- |
| `Sayl-windows-amd64.exe` | 🪟 Windows | 64-bit Intel/AMD |
| `Sayl-linux-amd64` | 🐧 Linux | 64-bit Intel/AMD |
| `Sayl-macos-amd64` | 🍎 macOS | Intel |
| `Sayl-macos-arm64` | 🍎 macOS | Apple Silicon (M1/M2/M3) |

---

### 🔨 Build from Source

If you prefer to build from source or need a custom build:

```bash
# 📋 Prerequisites: Go 1.23 or later
go version  # Verify Go is installed

# 📥 Clone the repository
git clone https://github.com/Amr-9/sayl.git
cd sayl

# 🔧 Build the binary
go build -o sayl ./cmd/sayl

# 🚀 Or with optimizations (smaller binary)
go build -ldflags="-s -w" -o sayl ./cmd/sayl

# ✅ Run it
./sayl
```

### 📦 Go Install

If you have Go installed, you can install directly:

```bash
go install github.com/Amr-9/sayl/cmd/sayl@latest

# ✅ Run it (make sure $GOPATH/bin is in your PATH)
sayl --help
```

---

## 🏃 Quick Start

### 🖥️ Interactive Mode (TUI)
```bash
./sayl
```
Follow the visual wizard to configure and run your test.

### 📄 Configuration File Mode
```bash
./sayl -config scenario.yaml
```

### ⌨️ Command Line Mode
```bash
./sayl -url "https://api.example.com/health" -method GET -rate 100 -duration 30s -concurrency 10
```

---

## 🛠️ Usage Workflows

### 1️⃣ The Explorer Workflow (TUI)
*🎯 Best for: Ad-hoc testing, debugging, and visual feedback*

```bash
./sayl
```

The interactive wizard walks you through:
1. 🎯 **Target Selection**: Input URL and HTTP Method
2. ⚙️ **Load Configuration**: Set rate, duration, and concurrency
3. 📋 **Header Setup**: Add custom headers (optional)
4. 📊 **Live Dashboard**: Watch real-time metrics during the test

### 2️⃣ The Automation Workflow (YAML)
*🎯 Best for: CI/CD pipelines, repeatable benchmarks, and complex scenarios*

Create a `scenario.yaml`:
```yaml
target:
  url: "https://api.example.com/v1/orders"
  method: "POST"
  headers:
    Content-Type: "application/json"
    Authorization: "Bearer {{env_token}}"
  body: '{"item_id": "{{uuid}}", "qty": {{random_int}}}'
  timeout: "10s"

load:
  duration: "2m"
  rate: 100
  concurrency: 20
  success_codes: [200, 201]
```

Run it:
```bash
./sayl -config scenario.yaml
```

---

## 📘 YAML Configuration Guide

The YAML configuration file is divided into **four main sections**. Each section controls a different aspect of your load test.

```
┌─────────────────────────────────────────────────────────────┐
│                    📄 YAML Structure                        │
├─────────────────────────────────────────────────────────────┤
│  🎯 target:    WHERE to send requests (URL, method, body)   │
│  ⚙️ load:      HOW to send requests (rate, duration)        │
│  🔗 steps:     MULTI-STEP scenarios (optional)              │
│  📁 data:      EXTERNAL data sources (optional)             │
└─────────────────────────────────────────────────────────────┘
```

---

### 🎯 Target Section

The `target` section defines **WHERE** your requests go and **WHAT** they contain.

```yaml
target:
  # ═══════════════════════════════════════════════════════════
  # 🌐 URL (Required)
  # ═══════════════════════════════════════════════════════════
  # The endpoint to test. Can include variables.
  url: "https://api.example.com/v1/users/{{uuid}}"
  
  # ═══════════════════════════════════════════════════════════
  # 📨 HTTP Method (Required)
  # ═══════════════════════════════════════════════════════════
  # Supported: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS
  method: "POST"
  
  # ═══════════════════════════════════════════════════════════
  # 📋 Headers (Optional)
  # ═══════════════════════════════════════════════════════════
  # Key-value pairs for HTTP headers. Variables supported!
  headers:
    Content-Type: "application/json"           # 📝 Required for JSON bodies
    Authorization: "Bearer {{auth_token}}"     # 🔑 Auth token (can be variable)
    Accept: "application/json"                 # 📥 Expected response type
    X-Request-ID: "req-{{timestamp_ms}}"       # 🏷️ Custom tracking header
    User-Agent: "Sayl-LoadTest/1.0"            # 🤖 Custom user agent
  
  # ═══════════════════════════════════════════════════════════
  # 📦 Request Body (Optional - choose ONE method)
  # ═══════════════════════════════════════════════════════════
  
  # 📝 Method 1: Inline String Body
  # Best for: Simple JSON, form data, or text
  body: '{"username": "{{random_email}}", "password": "test123"}'
  
  # 📂 Method 2: Load from File
  # Best for: Large payloads, complex JSON, binary data
  body_file: "./payloads/create_order.json"
  
  # 🧱 Method 3: Native YAML Object (auto-converts to JSON)
  # Best for: Complex nested structures, better readability
  body_json:
    user:
      name: "{{random_name}}"
      email: "{{random_email}}"
    order:
      items:
        - product_id: "{{uuid}}"
          quantity: 2
        - product_id: "{{uuid}}"
          quantity: 1
      total: 99.99
  
  # ═══════════════════════════════════════════════════════════
  # ⏱️ Timeout (Optional)
  # ═══════════════════════════════════════════════════════════
  # Maximum time to wait for a response
  # Format: "30s", "1m", "500ms"
  # Default: 30s
  timeout: "15s"
  
  # ═══════════════════════════════════════════════════════════
  # 🔓 TLS Settings (Optional)
  # ═══════════════════════════════════════════════════════════
  # Skip TLS certificate verification (for self-signed certs)
  # ⚠️ WARNING: Only use in development/testing!
  insecure: false  # Default: false
  
  # ═══════════════════════════════════════════════════════════
  # 🔌 Connection Settings (Optional)
  # ═══════════════════════════════════════════════════════════
  # Enable HTTP keep-alive for connection reuse
  # Improves performance for high-rate tests
  keep_alive: true  # Default: true
```

#### 📝 Body Format Examples

<details>
<summary>📋 Click to expand: JSON Body Examples</summary>

```yaml
# Simple JSON
body: '{"key": "value"}'

# JSON with variables
body: '{"email": "{{random_email}}", "id": "{{uuid}}"}'

# Multi-line JSON (using YAML literal block)
body: |
  {
    "user": {
      "name": "{{random_name}}",
      "email": "{{random_email}}"
    },
    "timestamp": {{timestamp}}
  }
```
</details>

<details>
<summary>📋 Click to expand: Form Data Examples</summary>

```yaml
# URL-encoded form data
headers:
  Content-Type: "application/x-www-form-urlencoded"
body: "username={{random_email}}&password=secret123&remember=true"
```
</details>

<details>
<summary>📋 Click to expand: GraphQL Examples</summary>

```yaml
headers:
  Content-Type: "application/json"
body: |
  {
    "query": "mutation CreateUser($input: UserInput!) { createUser(input: $input) { id name } }",
    "variables": {
      "input": {
        "name": "{{random_name}}",
        "email": "{{random_email}}"
      }
    }
  }
```
</details>

---

### ⚙️ Load Section

The `load` section defines **HOW** requests are sent - rate, duration, and concurrency.

```yaml
load:
  # ═══════════════════════════════════════════════════════════
  # ⏱️ Duration (Required if no stages)
  # ═══════════════════════════════════════════════════════════
  # How long to run the test
  # Format: "30s", "5m", "1h", "1h30m"
  duration: "2m"
  
  # ═══════════════════════════════════════════════════════════
  # 🚀 Rate (Required if no stages)
  # ═══════════════════════════════════════════════════════════
  # Requests per second (RPS) to maintain
  # This is the TARGET rate - actual may vary based on server response
  rate: 100
  
  # ═══════════════════════════════════════════════════════════
  # 👥 Concurrency (Required)
  # ═══════════════════════════════════════════════════════════
  # Number of concurrent workers (goroutines)
  # 
  # 💡 TIPS:
  # - Set higher than rate for bursty traffic
  # - Set equal to rate for steady traffic
  # - For slow endpoints, use concurrency > rate
  #
  # 📊 Example: rate=100, concurrency=50
  #    Each worker handles ~2 requests/second
  concurrency: 50
  
  # ═══════════════════════════════════════════════════════════
  # ✅ Success Codes (Optional)
  # ═══════════════════════════════════════════════════════════
  # HTTP status codes to count as successful
  # Default: [200]
  success_codes: [200, 201, 202, 204]
  
  # ═══════════════════════════════════════════════════════════
  # 📈 Stages (Optional - replaces duration/rate)
  # ═══════════════════════════════════════════════════════════
  # Define variable load patterns over time
  # Rate transitions SMOOTHLY between stages (linear ramping)
  stages:
    # 🌅 Stage 1: Warm-up
    - duration: "30s"
      target: 10      # Start at 10 RPS
    
    # 📈 Stage 2: Ramp up
    - duration: "1m"
      target: 100     # Gradually increase to 100 RPS
    
    # 🔥 Stage 3: Peak load
    - duration: "5m"
      target: 100     # Hold at 100 RPS
    
    # ⚡ Stage 4: Stress test
    - duration: "30s"
      target: 500     # Spike to 500 RPS
    
    # 📉 Stage 5: Recovery
    - duration: "1m"
      target: 50      # Drop to 50 RPS
    
    # 🌙 Stage 6: Cool down
    - duration: "30s"
      target: 0       # Gradually stop
```

#### 📊 Load Pattern Visualization

```
Rate (RPS)
    │
500 │                    ╭───╮
    │                   ╱     ╲
100 │       ╭──────────╯       ╲
    │      ╱                    ╲
 50 │     ╱                      ╲──────╮
    │    ╱                              ╲
 10 │───╯                                ╲───
    │
  0 └─────────────────────────────────────────▶ Time
      30s    1m      5m    30s   1m    30s
      warm   ramp   peak  spike  cool  stop
```

---

### 🔗 Steps Section

The `steps` section defines **MULTI-STEP** scenarios for complex API flows.

```yaml
steps:
  # ═══════════════════════════════════════════════════════════
  # 🔐 Step 1: Authentication
  # ═══════════════════════════════════════════════════════════
  - name: "Login"                              # 🏷️ Step identifier (for logs)
    url: "https://api.example.com/auth/login"
    method: "POST"
    headers:
      Content-Type: "application/json"
    body: |
      {
        "email": "{{random_email}}",
        "password": "test123"
      }
    
    # 📤 Extract values from response for later use
    extract:
      # JSON path extraction (dot notation)
      auth_token: "data.access_token"    # From: {"data": {"access_token": "abc"}}
      user_id: "data.user.id"            # From: {"data": {"user": {"id": 123}}}
      expires_in: "data.expires_in"      # From: {"data": {"expires_in": 3600}}
      
      # Header extraction (prefix with "header:")
      session_id: "header:X-Session-ID"   # From response header
      rate_limit: "header:X-RateLimit-Remaining"

  # ═══════════════════════════════════════════════════════════
  # 👤 Step 2: Get User Profile (uses extracted token)
  # ═══════════════════════════════════════════════════════════
  - name: "Get Profile"
    url: "https://api.example.com/users/{{user_id}}"  # 🔄 Using extracted variable
    method: "GET"
    headers:
      Authorization: "Bearer {{auth_token}}"          # 🔑 Using extracted token
      X-Session-ID: "{{session_id}}"
    
    # 📤 Extract more data for next step
    extract:
      account_id: "data.account_id"
      subscription_tier: "data.subscription.tier"

  # ═══════════════════════════════════════════════════════════
  # 🛒 Step 3: Create Order (uses multiple extracted values)
  # ═══════════════════════════════════════════════════════════
  - name: "Create Order"
    url: "https://api.example.com/accounts/{{account_id}}/orders"
    method: "POST"
    headers:
      Authorization: "Bearer {{auth_token}}"
      Content-Type: "application/json"
    body: |
      {
        "product_id": "{{uuid}}",
        "quantity": {{random_int}},
        "user_id": "{{user_id}}",
        "tier": "{{subscription_tier}}"
      }
    
    # 💾 Save computed values for this step
    variables:
      order_timestamp: "{{timestamp_ms}}"
      order_id_prefix: "ORD-{{random_digits_8}}"
```

#### 🔄 Step Execution Flow

```
┌─────────────────────────────────────────────────────────────┐
│                     🔄 Step Execution                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   Step 1: Login                                             │
│   ┌─────────────┐      ┌─────────────┐                      │
│   │ POST /login │ ───▶ │  Response   │                      │
│   └─────────────┘      └──────┬──────┘                      │
│                               │                             │
│                        ┌──────▼──────┐                      │
│                        │   Extract    │                      │
│                        │ auth_token   │                      │
│                        │ user_id      │                      │
│                        └──────┬──────┘                      │
│                               │                             │
│   Step 2: Get Profile         ▼                             │
│   ┌─────────────────────────────────────┐                   │
│   │ GET /users/{{user_id}}              │                   │
│   │ Authorization: Bearer {{auth_token}} │                   │
│   └──────────────────┬──────────────────┘                   │
│                      │                                      │
│               ┌──────▼──────┐                               │
│               │   Extract    │                               │
│               │ account_id   │                               │
│               └──────┬──────┘                               │
│                      │                                      │
│   Step 3: Create Order                                      │
│   ┌─────────────────────────────────────────┐               │
│   │ POST /accounts/{{account_id}}/orders    │               │
│   │ Body: {"user_id": "{{user_id}}", ...}   │               │
│   └─────────────────────────────────────────┘               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

### 📁 Data Section

The `data` section defines **EXTERNAL** data sources like CSV files.

```yaml
data:
  # ═══════════════════════════════════════════════════════════
  # 👥 Users Data Source
  # ═══════════════════════════════════════════════════════════
  - name: "users"           # 🏷️ Reference name (use as {{users.column}})
    path: "./data/users.csv"  # 📂 Path to CSV file
  
  # 🛍️ Products Data Source
  - name: "products"
    path: "./data/products.csv"
  
  # 🏢 Companies Data Source
  - name: "companies"
    path: "./data/companies.csv"
```

#### 📄 CSV File Format

```csv
# 📂 data/users.csv
email,password,name,role
admin@test.com,secret123,Alice Admin,admin
user1@test.com,pass456,Bob User,user
user2@test.com,pass789,Charlie User,user
```

#### 🔄 Usage in YAML

```yaml
target:
  url: "https://api.example.com/login"
  method: "POST"
  body: |
    {
      "email": "{{users.email}}",
      "password": "{{users.password}}",
      "name": "{{users.name}}",
      "role": "{{users.role}}"
    }
```

#### 💡 Data Feeding Behavior

```
┌─────────────────────────────────────────────────────────────┐
│                   📊 CSV Data Cycling                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   CSV File:              Request Uses:                      │
│   ┌────────────────┐                                        │
│   │ Row 1 (Alice)  │ ───▶ Request 1                         │
│   │ Row 2 (Bob)    │ ───▶ Request 2                         │
│   │ Row 3 (Charlie)│ ───▶ Request 3                         │
│   │ Row 1 (Alice)  │ ───▶ Request 4  (cycles back)         │
│   │ Row 2 (Bob)    │ ───▶ Request 5                         │
│   │ ...            │ ───▶ ...                               │
│   └────────────────┘                                        │
│                                                             │
│   ♻️ Data cycles infinitely through all rows                │
│   🧵 Each worker gets next row (thread-safe)                │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 🎲 Dynamic Variables

Inject random/dynamic data anywhere in your configuration using `{{variable}}` syntax.

### 📦 Built-in Generators

| 🏷️ Variable | 📝 Description | 💡 Example Output |
| :--- | :--- | :--- |
| `{{uuid}}` | 🔑 Random UUID v4 | `d53a1f32-8c4a-4b1e-9f2c-...` |
| `{{random_int}}` | 🔢 Random integer (0-100,000) | `48293` |
| `{{random_email}}` | 📧 Random email address | `user482938@example.com` |
| `{{random_name}}` | 👤 Random name with number | `Alice 847` |
| `{{random_phone}}` | 📱 US phone number | `+1-555-0142` |
| `{{random_domain}}` | 🌐 Random subdomain | `x7k2.example.com` |
| `{{random_alphanum}}` | 🔤 10-char alphanumeric | `aZ9xK2mNpQ` |
| `{{timestamp}}` | ⏰ Unix timestamp (seconds) | `1705632847` |
| `{{timestamp_ms}}` | ⏱️ Unix timestamp (milliseconds) | `1705632847123` |

### 🔢 Dynamic Length Generators

| 🏷️ Variable | 📝 Description | 💡 Example Output |
| :--- | :--- | :--- |
| `{{random_digits_5}}` | 5️⃣ 5 random digits | `48293` |
| `{{random_digits_10}}` | 🔟 10 random digits | `4829316745` |
| `{{random_digits_N}}` | #️⃣ N random digits (max 20) | `...` |

### 💡 Usage Examples

```yaml
# 🌐 In URL
url: "https://api.example.com/users/{{uuid}}"

# 📋 In Headers
headers:
  X-Request-ID: "req-{{timestamp_ms}}"
  X-Correlation-ID: "{{uuid}}"

# 📦 In Body
body: |
  {
    "email": "{{random_email}}",
    "phone": "{{random_phone}}",
    "order_id": "ORD-{{random_digits_8}}",
    "session": "{{uuid}}",
    "created_at": {{timestamp}}
  }
```

---

## 🔧 CLI Flags Reference

| 🏷️ Flag | 🔤 Short | 📝 Description | 💡 Example |
| :--- | :---: | :--- | :--- |
| `--config` | `-f` | 📄 YAML configuration file | `-config test.yaml` |
| `--url` | | 🌐 Target URL | `--url https://api.example.com` |
| `--method` | | 📨 HTTP method | `--method POST` |
| `--rate` | | 🚀 Requests per second | `--rate 100` |
| `--duration` | | ⏱️ Test duration | `--duration 2m` |
| `--concurrency` | | 👥 Concurrent workers | `--concurrency 20` |
| `--success` | | ✅ Success status codes | `--success 200,201,204` |

### 💡 CLI Examples

```bash
# 🧪 Simple GET test
./sayl --url "https://api.example.com/health" --rate 50 --duration 1m --concurrency 10

# 📄 Override config file settings
./sayl -config base.yaml --rate 500 --duration 30s

# ✅ Test with custom success codes
./sayl --url "https://api.example.com/create" --method POST --success 200,201,202
```

---

## 📊 Output & Reports

### 🖨️ Console Summary
After each test, Sayl displays a detailed console summary:
```
╔══════════════════════════════════════════════════════════════╗
║                   🌊 SAYL LOAD TEST REPORT                   ║
╠══════════════════════════════════════════════════════════════╣
║  🎯 Target: https://api.example.com/v1/orders                ║
║  📨 Method: POST                                             ║
║  ⏱️  Duration: 2m0s                                          ║
╠══════════════════════════════════════════════════════════════╣
║  📊 METRICS                                                  ║
╠══════════════════════════════════════════════════════════════╣
║  📈 Total Requests:     12,847                               ║
║  ✅ Success Rate:       98.7%                                ║
║  ⏱️  Avg Latency:        45.2ms                              ║
║  📊 P50 Latency:        38.1ms                               ║
║  📊 P95 Latency:        89.4ms                               ║
║  📊 P99 Latency:        156.2ms                              ║
║  💾 Throughput:         2.4 MB/s                             ║
╠══════════════════════════════════════════════════════════════╣
║  📋 STATUS CODES                                             ║
╠══════════════════════════════════════════════════════════════╣
║  ✅ 200 OK:             12,542 (97.6%)                       ║
║  ✅ 201 Created:        142 (1.1%)                           ║
║  ❌ 500 Server Error:   89 (0.7%)                            ║
║  ⏱️  Timeout:            74 (0.6%)                           ║
╚══════════════════════════════════════════════════════════════╝
```

### 📄 Generated Files

| 📂 File | 📝 Description |
| :--- | :--- |
| `report.json` | 🤖 Machine-readable JSON with all metrics |
| `report.html` | 📈 Interactive HTML dashboard with charts |

### 🗂️ JSON Report Structure
```json
{
  "target_url": "https://api.example.com",
  "method": "POST",
  "duration": "2m0s",
  "total_requests": 12847,
  "success_count": 12684,
  "failure_count": 163,
  "success_rate": 98.73,
  "latency": {
    "avg_ms": 45.2,
    "min_ms": 12.1,
    "max_ms": 892.4,
    "p50_ms": 38.1,
    "p75_ms": 56.8,
    "p95_ms": 89.4,
    "p99_ms": 156.2
  },
  "throughput": {
    "total_bytes": 287654321,
    "mbps": 2.4
  },
  "status_codes": {
    "200": 12542,
    "201": 142,
    "500": 89,
    "Timeout": 74
  }
}
```

---

## 📂 Examples Gallery

The `Examples of yaml files` folder contains ready-to-use configurations:

| 📄 File | 📝 Description | 🏷️ Tags |
| :--- | :--- | :--- |
| [01_basic_get.yaml](./Examples%20of%20yaml%20files/01_basic_get.yaml) | 🌐 Simple GET request | `beginner` |
| [02_post_json.yaml](./Examples%20of%20yaml%20files/02_post_json.yaml) | 📦 POST with JSON body | `beginner` |
| [03_post_raw_body.yaml](./Examples%20of%20yaml%20files/03_post_raw_body.yaml) | 📝 POST with raw text body | `beginner` |
| [04_load_stages.yaml](./Examples%20of%20yaml%20files/04_load_stages.yaml) | 📈 Ramped load with stages | `intermediate` |
| [05_data_loader.yaml](./Examples%20of%20yaml%20files/05_data_loader.yaml) | 📁 CSV data feeding | `intermediate` |
| [06_scenario_chain.yaml](./Examples%20of%20yaml%20files/06_scenario_chain.yaml) | 🔗 Multi-step auth flow | `advanced` |
| [07_auth_headers.yaml](./Examples%20of%20yaml%20files/07_auth_headers.yaml) | 🔑 Bearer token auth | `beginner` |
| [08_advanced_config.yaml](./Examples%20of%20yaml%20files/08_advanced_config.yaml) | ⚙️ All options combined | `advanced` |
| [10_graphql_query.yaml](./Examples%20of%20yaml%20files/10_graphql_query.yaml) | 🔮 GraphQL queries | `intermediate` |
| [17_complex_json_body.yaml](./Examples%20of%20yaml%20files/17_complex_json_body.yaml) | 🧱 Nested JSON with body_json | `intermediate` |
| [19_variables_demo.yaml](./Examples%20of%20yaml%20files/19_variables_demo.yaml) | 🎲 All variable types | `intermediate` |
| [21_persistence_demo.yaml](./Examples%20of%20yaml%20files/21_persistence_demo.yaml) | 💾 Session persistence | `advanced` |

---

## 🏗️ Architecture

```
🌊 sayl/
├── 🚀 cmd/sayl/                 # Application entry point
│   └── main.go                  # CLI parsing, signal handling
├── 🔧 internal/
│   ├── ⚡ attacker/             # Load generation engine
│   │   ├── attacker.go          # HTTP client, workers, retry logic
│   │   ├── variables.go         # Template variable processor
│   │   └── csv_feeder.go        # CSV data source handler
│   ├── 📊 report/               # Report generation
│   │   └── report.go            # Console, JSON, HTML reports
│   ├── 📈 stats/                # Statistics collection
│   │   └── collector.go         # Latency histograms, percentiles
│   └── 🖥️ tui/                  # Terminal UI
│       ├── setup.go             # Configuration wizard
│       ├── dash.go              # Live dashboard
│       └── styles.go            # UI styling
└── 📦 pkg/
    ├── ⚙️ config/               # Configuration parsing
    │   └── config.go            # YAML loader, validator
    └── 🗂️ models/               # Data structures
        └── models.go            # Config, Result, Report types
```

### 🧩 Key Components

| 🔧 Component | 📝 Responsibility |
| :--- | :--- |
| ⚡ **Attacker Engine** | Manages HTTP client pool, rate limiting, and worker goroutines |
| 🎲 **Variable Processor** | Parses and replaces `{{variable}}` placeholders |
| 📊 **Stats Collector** | Aggregates latency, status codes, and throughput |
| 🖥️ **TUI Module** | Renders interactive terminal interface |
| ⚙️ **Config Loader** | Parses YAML and validates configuration |

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. 🍴 Fork the repository
2. 🌿 Create your feature branch (`git checkout -b feature/amazing-feature`)
3. 💾 Commit your changes (`git commit -m 'Add amazing feature'`)
4. 📤 Push to the branch (`git push origin feature/amazing-feature`)
5. 🔃 Open a Pull Request

---

## 📄 License

This project is licensed under the GPL-3.0 License - see the [LICENSE](LICENSE) file for details.

---

<p align="center">
  <strong>🌊 Ride the Wave of Load Testing!</strong>
</p>

<p align="center">
  Made with ❤️ by <a href="https://github.com/Amr-9">Amr</a>
</p>

<p align="center">
  ⭐ Star this repo if you find it useful! ⭐
</p>
