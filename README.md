# URL Shortener

A simple, efficient URL shortener service built with Go and MongoDB.

## Features

- Shorten long URLs into compact, easy-to-share links
- Retrieve original URLs via shortened links
- MongoDB storage for persistence
- Rate limiting to prevent abuse
- RESTful API endpoints
- Docker support

## Prerequisites

- Go 1.22 or higher
- MongoDB Atlas account (or local MongoDB instance)
- Environment variables configured

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```
DB_USER=your_mongodb_username
DB_PASSWORD=your_mongodb_password
DB_CLUSTER=your_mongodb_cluster
PORT=3001
```

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/url-shortener.git
   cd url-shortener
   ```

2. Install dependencies:
   ```
   go mod download
   ```

3. Run the application:
   ```
   go run main.go
   ```

The server will start on the port specified in your `.env` file (default: 3001).

## API Endpoints

### Health Check
```
GET /health
```
Returns a simple message indicating the service is running.

### Shorten URL
```
POST /shorten
```
Request body (JSON):
```json
{
  "originalUrl": "https://example.com/very/long/url/that/needs/shortening"
}
```

Response (JSON):
```json
{
  "success": true,
  "originalUrl": "https://example.com/very/long/url/that/needs/shortening",
  "shortenedUrl": "http://short.link/abc123"
}
```

### Redirect to Original URL
```
GET /{id}
```
Where `{id}` is the shortened ID (e.g., `abc123`).

This endpoint redirects to the original URL associated with the shortened ID.

## Architecture

- **main.go**: Entry point, server setup, and database connection
- **handlers.go**: HTTP request handlers
- **utils/shorten.go**: URL shortening algorithms and utilities

## How It Works

1. **URL Shortening**: 
   - When a new URL is submitted, the system checks if it already exists in the database
   - If it exists, the existing shortened URL is returned
   - Otherwise, a new unique ID is generated, converted to a Base62 string, and stored in the database

2. **URL Redirect**:
   - When a user visits a shortened URL, the system looks up the ID in the database
   - If found, it redirects to the original URL
   - If not found, returns a 404 error

3. **Rate Limiting**:
   - Limits requests per IP address to prevent abuse
   - Configurable rate limit settings

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 
