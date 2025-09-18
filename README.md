# Go Gin API Server

A robust REST API server built with Go, Gin framework, and PostgreSQL, featuring user authentication, post management, and comprehensive testing.

## Features

- **Authentication System**
  - User registration and login
  - JWT token-based authentication
  - Token refresh mechanism
  - User activation/deactivation with permission control
- **Post Management**
  - Create, read, update, delete posts
  - Cursor-based pagination
  - Author-based filtering
- **User Management**
  - User profile management
  - Username/email lookup
  - User status management
- **Comprehensive Testing**
  - Unit tests for all layers
  - Integration tests for end-to-end flows
  - Handler tests for HTTP layer

## Todo List

### High Priority

#### ✅ Completed Features

- [x] **Authentication System**

  - [x] User registration and login
  - [x] JWT token-based authentication
  - [x] Token refresh mechanism
  - [x] User activation/deactivation with permission control

- [x] **Post Management System**

  - [x] Create, read, update, delete posts
  - [x] Cursor-based pagination
  - [x] Author-based filtering
  - [x] Permission-based access control

- [x] **User Management System**

  - [x] User profile management
  - [x] Username/email lookup
  - [x] User status management
  - [x] Profile update functionality

- [x] **Comprehensive Testing**

  - [x] Unit tests for all layers (handler, service, repository)
  - [x] Integration tests for end-to-end flows
  - [x] Test database setup with Docker
  - [x] Test coverage and quality assurance

- [x] **Infrastructure Setup**
  - [x] Docker Compose for development environment
  - [x] Database migrations with golang-migrate
  - [x] Nginx load balancer setup
  - [x] Project structure and organization

#### Core Features

- [x] **Role-Based Access Control**

  - [x] Implement user roles (admin, user)
  - [x] Add role-based permissions middleware
  - [x] Update user activation/deactivation to use roles
  - [x] Post ownership permission checks
  - [x] Comprehensive RBAC testing

- [ ] **Enhanced Security**

  - [ ] Implement rate limiting
  - [ ] Add CORS configuration
  - [ ] Input validation improvements
  - [ ] SQL injection prevention audit

- [ ] **API Documentation**
  - [ ] Generate OpenAPI/Swagger documentation
  - [ ] Add API versioning strategy
  - [ ] Create interactive API docs

#### Database & Performance

- [ ] **Database Optimization**

  - [ ] Add database indexes for performance
  - [ ] Implement database connection pooling
  - [x] Add database migrations management
  - [ ] Query optimization review

- [ ] **Caching Layer**
  - [ ] Implement Redis for session storage
  - [ ] Add caching for frequently accessed data
  - [ ] Cache invalidation strategy

### Medium Priority

#### Enhanced Features

- [ ] **Post Features**

  - [ ] Add post categories/tags
  - [ ] Implement post search functionality
  - [ ] Add post likes/comments system
  - [ ] File upload for post attachments

- [ ] **User Features**

  - [ ] User profile pictures
  - [ ] User following/followers system
  - [ ] User activity feed
  - [ ] Password reset functionality

- [ ] **Notification System**
  - [ ] Email notifications
  - [ ] In-app notifications
  - [ ] Push notifications (future)

#### DevOps & Deployment

- [ ] **Containerization**

  - [ ] Optimize Docker images
  - [ ] Add multi-stage builds
  - [x] Basic Docker setup with Compose
  - [ ] Container security scanning

- [ ] **CI/CD Pipeline**

  - [x] GitHub Actions workflow
  - [x] Automated testing on PR
  - [ ] Automated deployment
  - [x] Code quality checks (linting, formatting)

- [ ] **Monitoring & Observability**
  - [ ] Add application metrics (Prometheus)
  - [ ] Implement health checks
  - [ ] Add distributed tracing
  - [ ] Log aggregation and analysis

### Low Priority / Future

#### Advanced Features

- [ ] **Real-time Features**

  - [ ] WebSocket support for real-time updates
  - [ ] Live chat functionality
  - [ ] Real-time notifications

- [ ] **Advanced Post Features**

  - [ ] Rich text editor integration
  - [ ] Post scheduling
  - [ ] Post analytics
  - [ ] Content moderation tools

- [ ] **Social Features**
  - [ ] User mentions (@username)
  - [ ] Hashtag support
  - [ ] Post sharing functionality
  - [ ] Community/group features

#### Performance & Scale

- [ ] **Load Testing**

  - [ ] Performance benchmarking
  - [ ] Load testing with realistic data
  - [ ] Stress testing for edge cases
  - [ ] Performance monitoring

- [ ] **Microservices Architecture**
  - [ ] Service decomposition analysis
  - [ ] API gateway implementation
  - [ ] Service discovery
  - [ ] Distributed logging

## Technology Stack

- **Backend**: Go 1.21+, Gin Framework
- **Database**: PostgreSQL
- **Authentication**: JWT tokens
- **Testing**: Go testing package, testify
- **Containerization**: Docker, Docker Compose
- **Migration**: golang-migrate/migrate

## Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd go-gin-api-server

# Start with Docker Compose (includes database, API server, nginx, and migrations)
docker-compose up -d
```

Your API server will be available at `http://localhost` (nginx proxy) or `http://localhost:8080` (direct API access).

## Testing Setup

```bash
# Start test database (includes test migrations)
docker-compose -f docker-compose.test.yml up -d
```

## Project Structure

```
├── cmd/                    # Application entry points
├── config/                 # Configuration management
├── internal/               # Private application code
│   ├── handler/           # HTTP handlers
│   ├── middleware/        # HTTP middleware
│   ├── model/             # Data models
│   ├── repository/        # Data access layer
│   ├── service/           # Business logic
│   └── server/            # Server setup
├── integration/           # Integration tests
├── test/                  # Unit tests
├── migrations/            # Database migrations
├── scripts/               # Utility scripts
└── pkg/                   # Public packages
```

## Testing

### Prerequisites

Make sure you have the test database running:

```bash
docker-compose -f docker-compose.test.yml up -d
```

### Running Tests

```bash
# Run all tests
go test ./... -v

# Run tests with coverage
go test ./... -cover
```

## API Endpoints

### Authentication

- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Token refresh
- `POST /api/v1/auth/activate/:userID` - Activate user (admin)
- `POST /api/v1/auth/deactivate/:userID` - Deactivate user

### Posts

- `GET /api/v1/posts` - List posts with pagination
- `POST /api/v1/posts` - Create post
- `GET /api/v1/posts/:id` - Get post by ID
- `PATCH /api/v1/posts/:id` - Update post
- `DELETE /api/v1/posts/:id` - Delete post

### Users

- `GET /api/v1/users/:id` - Get user by ID
- `GET /api/v1/users/username/:username` - Get user by username
- `GET /api/v1/users/email/:email` - Get user by email
- `GET /api/v1/users/profile/:username` - Get user profile
- `PATCH /api/v1/users/:id` - Update user profile
- ~~`DELETE /api/v1/users/:id` - Delete user~~

## License

This project is licensed under the MIT License.

---

**Note**: This is a work in progress. See the Todo List above for planned features and improvements.
