# Deploying ToughRadius with Coolify

[Coolify](https://coolify.io/) makes it easy to deploy ToughRadius using Docker. Since we have updated the [Dockerfile](file:///home/faris/Downloads/toughradius/toughradius/Dockerfile) to handle automatic database setup, the process is very straightforward.

## 1. Create a New Application
- Go to your Coolify dashboard.
- Click **Add New Resource** -> **Application**.
- Select **Public Repository** (or private if your fork is private).
- Enter your repository URL (e.g., `https://github.com/farisnoaman/kart`).

## 2. Configure Build Settings
- **Build Pack**: Select `Docker`.
- **Dockerfile Path**: Set to [./Dockerfile](file:///home/faris/Downloads/toughradius/toughradius/Dockerfile) (default).
- **Branch**: Set to [main](file:///home/faris/Downloads/toughradius/toughradius/main.go#57-123).

## 3. Configure Domain & Ports
- **Domain**: Enter `https://radius.hayataxi.online`.
- **Internal Port**: Set to `1816`.
- **Additional Ports (RADIUS)**:
  Coolify usually creates a proxy for HTTP (1816). For RADIUS (UDP), you need to map them manually in the **Network** tab of your application:
  - `1812:1812/udp` (RADIUS Auth)
  - `1813:1813/udp` (RADIUS Acct)

## 4. Set Environment Variables
Go to the **Environment Variables** tab and add:
- `TOUGHRADIUS_WEB_PORT`: `1816`
- `TOUGHRADIUS_DB_TYPE`: `sqlite`
- `TOUGHRADIUS_DB_NAME`: `/var/toughradius/data/toughradius.db`
- `TOUGHRADIUS_SYSTEM_WORKER_DIR`: `/var/toughradius`
- `TOUGHRADIUS_LOGGER_MODE`: `production`
- `TOUGHRADIUS_WEB_SECRET`: [(generate a random 32+ character string here)](file:///home/faris/Downloads/toughradius/toughradius/internal/app/app.go#63-66)

## 5. Persistent Storage (Volumes)
Go to the **Storages** tab and add a persistent volume:
- **Mount Path**: `/var/toughradius`
- This ensures your database (`sqlite`) and logs are saved even when you redeploy.

## 6. Deploy
- Click **Deploy**.
- Coolify will build the Docker image (this takes a few minutes as it compiles the frontend and backend).
- Once finished, your admin panel will be live at `https://radius.hayataxi.online`.

---

> [!TIP]
> **First Run**: The first time you deploy, the [entrypoint.sh](file:///home/faris/Downloads/toughradius/toughradius/scripts/entrypoint.sh) script will automatically run the `-initdb` command to create your database tables and default admin user.
