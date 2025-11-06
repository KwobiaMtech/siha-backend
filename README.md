# HealthyPay Backend

## Project Structure

```
backend/
├── main.go                 # Application entry point
├── go.mod                  # Go module dependencies
├── go.sum                  # Go module checksums
├── .env                    # Environment variables
├── Dockerfile              # Docker configuration
├── docker-compose.yml      # Docker compose setup
├── internal/               # Internal application code
│   ├── handlers/           # HTTP request handlers
│   ├── models/             # Data models
│   ├── services/           # Business logic services
│   ├── middleware/         # HTTP middleware
│   └── ...
├── docs/                   # Documentation files
│   ├── README.md           # Project documentation
│   ├── *.md                # Various documentation
│   └── *.postman_collection.json  # API collections
├── test/                   # Test files
│   ├── *.go                # Go test files
│   ├── *.sh                # Shell test scripts
│   ├── *.http              # HTTP test files
│   └── *.js                # JavaScript test utilities
├── tests/                  # Additional test suites
├── scripts/                # Utility scripts
└── logs/                   # Application logs
```

## Getting Started

1. **Install dependencies:**
   ```bash
   go mod tidy
   ```

2. **Run the application:**
   ```bash
   go run main.go
   ```

3. **Build the application:**
   ```bash
   go build -o healthypay .
   ```

4. **Run tests:**
   ```bash
   cd test && ./run_tests.sh
   ```

## Features

- Multi-currency support (GHS, USD, KES, ZMW)
- Investment tracking with dynamic currency
- Mobile money integration
- Stellar blockchain wallet support
- RESTful API with comprehensive documentation
