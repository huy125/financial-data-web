# Finscope

<p align="center"">
  <img src="img/finscope_logo.png" alt="Finscope Logo" width="200">
</p>

A financial advisor API server that aggregates information
and gives investment recommendations based on the financial market.
It also facilitates the investment decision-making, analyses and gives investors the vision about their profile.

## Getting Started

Follow these steps to set up and run the application locally.

### 1. Clone the Repository

```bash
git clone https://github.com/huy125/financial-data-web.git
cd finscope
```

### 2. Configure Environment Variables

Create a `.env` file in the root of the project directory and add the configuration:

```plaintext
ALPHA_VANTAGE_API_KEY = your_api_key
DATA_SOURCE_NAME = postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable
```

### 3. Build and Run the Application

Run the following command to start the application using Docker Compose:

```bash
docker compose up --build
```

This command will build the Docker images and start the containers defined in the `docker-compose.yml` file.

### 4. Access the Application

Once the containers are running, you can access the application at:

```
http://localhost:8080
```
