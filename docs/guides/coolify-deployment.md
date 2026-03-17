# 🚀 Deploying RADIO on Coolify

This guide explains how to deploy **RADIO (ISP Billing & Management)** on your own server using [Coolify](https://coolify.io/).

## 📋 Prerequisites

- A running Coolify instance.
- A private or public Git repository (e.g., GitHub, GitLab) containing the RADIO source code.

## 🛠️ Deployment Steps

### 1. Create a New Resource
- Open your Coolify dashboard.
- Click on **"New Resource"** or **"Create"**.
- Select **"Application"**.

### 2. Connect Your Repository
- Select your Git source (e.g., GitHub).
- Choose the `radio` repository.
- Select the `main` branch.
- Click **"Next"**.

### 3. Build Configuration
Coolify should automatically detect the `Dockerfile`.
- **Build Pack**: Select **"Dockerfile"**.
- **Dockerfile Path**: Ensure it's set to `./Dockerfile`.

### 4. Configure Ports
The application needs several ports exposed:
- **Web Interface (Admin/Portal)**: `1816` (HTTP, TCP)
- **RADIUS Auth**: `1812` (UDP)
- **RADIUS Acct**: `1813` (UDP)
- **RadSec**: `2083` (TCP)

In Coolify's **"Network"** or **"Ports"** section:
1. Map a public port (or use a domain) to internal port **1816**.
2. If you need external RADIUS access, ensure your server firewall allows UDP **1812** and **1813**.

### 5. Set Environment Variables
Go to the **"Variables"** tab and add the following defaults:

| Variable | Recommended Value | Description |
| :--- | :--- | :--- |
| `TOUGHRADIUS_SYSTEM_DOMAIN` | `https://your-domain.com` | Your public URL |
| `TOUGHRADIUS_LOGGER_MODE` | `production` | Enables JSON logging |
| `TOUGHRADIUS_SYSTEM_DEBUG` | `false` | Disables debug logs |
| `TOUGHRADIUS_WEB_SECRET` | `(generate-a-random-string)` | Secret for JWT tokens |

### 6. Configure Persistence (Volumes)
**CRITICAL**: You must mount a volume to prevent data loss when the container restarts.

- Go to the **"Storages / Volumes"** tab.
- Add a new volume:
  - **Source**: `radio-data` (or a specific path on host)
  - **Destination**: `/var/toughradius`

This single mount handles the database, logs, and backups.

### 7. Deploy
- Click **"Deploy"**.
- Once the build finishes, your application will be live!

---

## 🔍 Post-Deployment Check
1. Access your domain (e.g., `https://radius.hayataxi.online/admin`).
2. Log in with the default credentials provided in the deployment logs (Terminal output).
3. Change your admin password immediately.
