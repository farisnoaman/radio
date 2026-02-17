# Starting ToughRadius in Development

This guide outlines the steps to start the ToughRadius application in a development environment.

## Prerequisites

- **Go**: Version 1.24.0 or higher.
- **Node.js**: Version 18 or higher.
- **Configuration**: A `toughradius.dev.yml` file in the root directory.

## Quick Start (Script)

We have provided a script to automate the startup process.

```bash
chmod +x start_dev.sh
./start_dev.sh
```

This script will:
1.  Check for Go and Node.js.
2.  Install frontend dependencies.
3.  Initialize the database.
4.  Start the backend server (on port 1816) and frontend dev server (on port 3000) concurrently.

## Manual Startup

If you prefer to run steps manually:

### 1. Build Frontend Assets
The backend requires frontend assets to be present in `web/dist` for embedding.

```bash
cd web
npm install
npm run build
cd ..
```

### 2. Initialize Database
Initialize the SQLite database (created in `rundata/data/toughradius.db`).

```bash
go run main.go -initdb -c toughradius.dev.yml
```

### 3. Start Backend Server
Run the Go backend.

```bash
go run main.go -c toughradius.dev.yml
```

The backend API will be available at `http://localhost:1816`.

### 4. Start Frontend Dev Server
Open a new terminal and run the Vite dev server.

```bash
cd web
npm run dev
```

The frontend will be available at `http://localhost:3000`.

## Default Credentials

- **Username**: `admin`
- **Password**: `toughradius`


## Notes

- Ensure the `toughradius.dev.yml` file is properly configured for your development environment.
- The script assumes a Unix-like environment. Adjust paths or commands if using Windows.

## Fetch Remote chanes:
```bash
git status && git remote -v
git fetch --all
git merge upstream/main
