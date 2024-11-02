# financial-data-web

A financial information aggregator service that provides a REST API.

## Prerequisites

Before you begin, ensure you have the following installed on your machine:

- [Docker](https://www.docker.com/get-started) (including Docker Compose)
- [Postman](https://www.postman.com/downloads/) (optional, for testing the API)

## Getting Started

Follow these steps to set up and run the application locally.

### 1. Clone the Repository

```bash
git clone https://github.com/huy125/financial-data-web.git
cd financial-data-web
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
docker-compose up --build
```

This command will build the Docker images and start the containers defined in the `docker-compose.yml` file.

### 4. Access the Application

Once the containers are running, you can access the application at:

```
http://localhost:8080
```

### 5. Testing the API

You can use Postman to test the API. Here’s how to make a request to get stock data:

1. Open Postman.
2. Create a new GET request.
3. Enter the following URL:

   ```
   http://localhost:8080
   ```

4. The response should return:

    ```
    Welcome to my financial server!!!
    ```

### 6. Stopping the Application

To stop the application, use:

```bash
docker-compose down
```
