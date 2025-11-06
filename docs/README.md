# Healthy Pay Backend - Clean Architecture

A Golang backend API following Clean Architecture principles with Domain-Driven Design.

## Architecture

```
healthy_pay_backend/
├── main.go
├── internal/
│   ├── domain/                    # Business Logic Layer
│   │   ├── entities/             # Domain entities
│   │   └── repositories/         # Repository interfaces
│   ├── application/              # Application Layer
│   │   ├── services/            # Business logic services
│   │   └── dtos/                # Data Transfer Objects
│   ├── infrastructure/          # Infrastructure Layer
│   │   ├── database/           # Database connection
│   │   └── repositories/       # Repository implementations
│   ├── presentation/           # Presentation Layer
│   │   └── controllers/        # HTTP controllers
│   ├── middleware/             # HTTP middleware
│   ├── config/                 # Configuration
│   └── utils/                  # Utility functions
```

## Clean Architecture Layers

### 1. Domain Layer (`internal/domain/`)
- **Entities**: Core business objects
- **Repository Interfaces**: Data access contracts

### 2. Application Layer (`internal/application/`)
- **Services**: Business logic implementation
- **DTOs**: Request/Response objects

### 3. Infrastructure Layer (`internal/infrastructure/`)
- **Repository Implementations**: Data access logic
- **Database**: Connection and configuration

### 4. Presentation Layer (`internal/presentation/`)
- **Controllers**: HTTP request handlers
- **Middleware**: Cross-cutting concerns

## Features

- Clean Architecture with dependency inversion
- Repository pattern for data access
- Service layer for business logic
- DTO pattern for data transfer
- JWT authentication
- MongoDB integration
- CORS support

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user

## Setup

1. Install dependencies:
```bash
go mod tidy
```

2. Set up environment:
```bash
cp .env.example .env
# Edit .env with your MongoDB URI
```

3. Run the server:
```bash
go run main.go
```

## Benefits of Clean Architecture

- **Testability**: Easy to unit test business logic
- **Maintainability**: Clear separation of concerns
- **Flexibility**: Easy to swap implementations
- **Scalability**: Well-organized code structure
- **Independence**: Framework and database agnostic core
