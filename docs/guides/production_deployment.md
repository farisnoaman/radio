# Production Deployment Guide

This guide explains how to move your ToughRadius project from a local development environment to a production server and make it accessible over the internet.

## Phase 1: Preparation (Remote Server)

1.  **Get a VPS**: Use a provider like AWS (LightSail/EC2), DigitalOcean, Hetzner, or Linode.
    - **Recommended OS**: Ubuntu 22.04 or 24.04 LTS.
    - **Resources**: Minimum 2GB RAM.
2.  **Domain & IP**: Ensure your server has a **Public IP Address** (which you have) and point your domain `radius.hayataxi.online` to that IP.
3.  **Security (Firewall)**: Open the following ports in your cloud provider's firewall:
    - `1816/TCP`: Web Management Interface.
    - `1812/UDP`: RADIUS Authentication.
    - `1813/UDP`: RADIUS Accounting.
    - `80/TCP` & `443/TCP`: For Nginx and SSL.

---

## Phase 2: Build the Production Binary

On your local machine or build server, create a single, self-contained binary that includes both the backend and the frontend.

```bash
# In the project root
make build
```
The output will be in [release/toughradius](file:///home/faris/Downloads/toughradius/toughradius/release/toughradius). This file is all you need to run the app.

---

## Phase 3: Setup on the Production Server

1.  **Transfer the binary**:
    ```bash
    scp release/toughradius root@your-vps-ip:/opt/toughradius/
    ```
2.  **Initialize the Environment**:
    ```bash
    cd /opt/toughradius
    # Create the config
    cp toughradius.yml toughradius.prod.yml
    # Initialize the database
    ./toughradius -initdb -c toughradius.prod.yml
    ```

---

## Phase 4: Service Management (systemd)

To ensure ToughRadius starts automatically and stays running, create a systemd service file: `/etc/systemd/system/toughradius.service`

```ini
[Unit]
Description=ToughRadius Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/toughradius
ExecStart=/opt/toughradius/toughradius -c /opt/toughradius/toughradius.prod.yml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

**Activate the service**:
```bash
sudo systemctl daemon-reload
sudo systemctl enable toughradius
sudo systemctl start toughradius
```

---

## Phase 5: Reverse Proxy (Nginx + SSL)

1.  **Install Nginx**: `sudo apt install nginx certbot python3-certbot-nginx`
2.  **Configure Nginx**: Create `/etc/nginx/sites-available/toughradius`
    ```nginx
    server {
        server_name radius.hayataxi.online;

        location / {
            proxy_pass http://localhost:1816;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
    ```
3.  **Enable Site & SSL**:
    ```bash
    sudo ln -s /etc/nginx/sites-available/toughradius /etc/nginx/sites-enabled/
    sudo certbot --nginx -d radius.hayataxi.online
    sudo systemctl restart nginx
    ```

---

## Summary Checklist
- [x] Domain `radius.hayataxi.online` pointed to Static IP.
- [ ] Ports 1812(UDP), 1813(UDP), 1816(TCP) open.
- [ ] `make build` created a single binary.
- [ ] Systemd service running.
- [ ] Nginx + Let's Encrypt SSL configured.
