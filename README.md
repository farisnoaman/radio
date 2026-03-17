# RADIO - ISP Billing & Management

> [!NOTE]
> This project is a customized and enhanced version of [ToughRADIUS](https://github.com/talkincode/toughradius). It builds upon the powerful core of ToughRADIUS with specialized features for ISP management, improved UI/UX, and extensive localization.

A powerful RADIUS server and ISP management platform designed for ISPs, enterprise networks, and carriers. Supports standard RADIUS protocols, RadSec (RADIUS over TLS), and a modern Web management interface with full RTL support.

## ✨ Core Features

### 📡 Network & RADIUS
- 🔐 **Standard RADIUS** - Full support for RFC 2865/2866 authentication and accounting protocols.
- 🔒 **RadSec** - TLS encrypted RADIUS over TCP (RFC 6614).
- 🌐 **Mikrotik Auto-Discovery** - Automatically discover and configure Mikrotik routers on your network.
- ⚡ **High Performance** - Built with Go, supporting high concurrency processing for large-scale deployments.
- 🔌 **Dynamic Configuration** - Configure RADIUS server IP and other critical settings directly from the system management page.

### 🎟️ Advanced Voucher Management
- 📊 **Quota & Validity Support** - Create vouchers with specific data limits (Quota) and time limits (Validity).
- 🔄 **Voucher Transfers** - Easily transfer vouchers between different user accounts or batches.
- 🗑️ **Batch Operations** - Support for batch delete with confirmation and status tracking.
- 📜 **Detailed Analytics** - Monitor voucher activation, usage, and expiration in real-time.

### 🌍 Localization & UI/UX
- 🇸🇦 **Full Arabic Support** - Completely translated into Arabic for seamless regional operation.
- ↔️ **Professional RTL Layout** - Full Right-to-Left alignment for the entire management interface.
- 📈 **Enhanced Dashboards** - Visualize network performance and user trends with integrated **ECharts**.
- 📱 **Mobile Responsive** - Optimized interface for management on the go, including specialized mobile navigation bars.


## 🚀 Quick Start

### Prerequisites

- Go 1.24+ (for building from source)
- PostgreSQL or SQLite
- Node.js 18+ (for frontend development)

### Installation

#### 1. Build from Source

```bash
# Clone repository
git clone https://github.com/farisnoaman/radio.git
cd radio

# Build frontend
cd web
npm install
npm run build
cd ..

# Build backend
go build -o toughradius main.go
```

#### 2. Use Pre-compiled Version

Download the latest version from the [Releases](https://github.com/talkincode/toughradius/releases) page.

### Configuration

1. Copy the configuration template:

```bash
cp toughradius.yml toughradius.prod.yml
```

2. Edit `toughradius.prod.yml` configuration file:

```yaml
system:
  appid: ToughRADIUS
  location: Asia/Shanghai
  workdir: ./rundata

database:
  type: sqlite # or postgres
  name: toughradius.db
  # PostgreSQL configuration
  # host: localhost
  # port: 5432
  # user: toughradius
  # passwd: your_password

radiusd:
  enabled: true
  host: 0.0.0.0
  auth_port: 1812 # RADIUS authentication port
  acct_port: 1813 # RADIUS accounting port
  radsec_port: 2083 # RadSec port

web:
  host: 0.0.0.0
  port: 1816 # Web management interface port
```

### EAP Configuration

You can fine-tune authentication behavior via system configuration (`sys_config`):

- `radius.EapMethod`: Preferred EAP method (default `eap-md5`).
- `radius.EapEnabledHandlers`: List of allowed EAP handlers, separated by commas, e.g., `eap-md5,eap-mschapv2`. Use `*` to enable all registered handlers.

This allows you to quickly disable unauthorized EAP methods without interrupting the service.

### Running

```bash
# Initialize database
./toughradius -initdb -c toughradius.prod.yml

# Start service
./toughradius -c toughradius.prod.yml
```

Access Web Management Interface: <http://localhost:1816>

Default Admin Account:

- Username: admin
- Password: Please check the initialization log output

## 📖 Documentation

- [Architecture](docs/v9-architecture.md) - v9 version architecture design
- [React Admin Refactor](docs/react-admin-refactor.md) - Frontend management interface explanation
- [SQLite Support](docs/sqlite-support.md) - SQLite database configuration
- [Environment Variables](docs/environment-variables.md) - Environment variable configuration guide

## 🏗️ Project Structure

```text
toughradius/
├── cmd/             # Application entry points
├── internal/        # Private application code
│   ├── adminapi/   # Admin API (New version)
│   ├── radiusd/    # RADIUS service core
│   ├── domain/     # Data models
│   └── webserver/  # Web server
├── pkg/            # Public libraries
├── web/            # React Admin frontend
└── docs/           # Documentation

## 🚢 Deployment

### 🐳 Coolify Deployment (Recommended)
RADIO is optimized for one-click deployment on Coolify.
Follow our [Detailed Coolify Guide](docs/guides/coolify-deployment.md) for step-by-step instructions on volumes, ports, and environment variables.

### 🐧 Manual Linux Deployment
For manual deployment, see our [Production Setup Guide](docs/guides/production-setup.md).
```

## 🔧 Development

### Backend Development

```bash
# Run tests
go test ./...

# Run benchmark tests
go test -bench=. ./internal/radiusd/

# Start development mode
go run main.go -c toughradius.yml
```

### Frontend Development

```bash
cd web
npm install
npm run dev       # Development server
npm run build     # Production build
npm run lint      # Code linting
```

## 🤝 Contribution

We welcome contributions in various forms, including but not limited to:

- 🐛 Submitting Bug reports and feature requests
- 📝 Improving documentation
- 💻 Submitting code patches and new features
- 🌍 Helping with translation

## 📜 License

This project is licensed under the [MIT License](LICENSE).

### Third-Party Resources

The RADIUS dictionary files in the `share/` directory are derived from the [FreeRADIUS](https://freeradius.org/) project and are licensed under the [Creative Commons Attribution 4.0 International License (CC BY 4.0)](share/LICENSE).

## 💎 Attribution

This project is built upon the excellent work of the [ToughRADIUS](https://github.com/talkincode/toughradius) project. We thank the original authors for their contributions to the open-source community.

