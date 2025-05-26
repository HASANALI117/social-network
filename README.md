# Social Network

A full-stack social network application built with Go backend and Next.js frontend, featuring real-time messaging, group management, post sharing, and comprehensive user interactions.

## 🚀 Features

### Core Functionality

- **User Authentication & Authorization**: Secure login/logout with session management
- **User Profiles**: Customizable profiles with avatar upload and bio
- **Posts & Content Sharing**: Create posts with text and image content
- **Privacy Controls**: Public/private posts with selective follower sharing
- **Social Interactions**: Follow/unfollow users, like and comment on posts
- **Real-time Chat**: Private messaging between users with WebSocket support
- **Group Management**: Create/join groups, group-specific posts and chat
- **Event System**: Create and manage group events with RSVP functionality
- **Notification System**: Real-time notifications for social interactions

### Privacy & Security

- **Profile Privacy**: Choose between public and private profiles
- **Follow Requests**: Private profiles require follow approval
- **Content Control**: Share posts with specific followers
- **Secure File Upload**: Image uploads with MinIO object storage

## 🏗️ Architecture

### Backend (Go)

- **Framework**: Pure Go with standard library
- **Database**: SQLite with custom repository pattern
- **Real-time**: WebSocket connections for chat and notifications
- **File Storage**: MinIO for image and media storage
- **Authentication**: Session-based with secure cookie management

### Frontend (Next.js)

- **Framework**: Next.js 14 with App Router
- **Styling**: Tailwind CSS with custom components
- **State Management**: Zustand for global state
- **Real-time**: WebSocket client for live updates
- **Forms**: React Hook Form with validation
- **UI Components**: Custom component library

### Infrastructure

- **Containerization**: Docker with multi-container setup
- **Reverse Proxy**: Nginx for load balancing
- **Development**: Hot reload with Air (Go) and Next.js dev server

## 📁 Project Structure

```
social-network/
├── backend/                 # Go backend application
│   ├── cmd/server/         # Application entry point
│   ├── pkg/                # Core business logic
│   │   ├── auth/          # Authentication services
│   │   ├── db/            # Database layer and migrations
│   │   ├── handlers/      # HTTP request handlers
│   │   ├── middleware/    # HTTP middleware
│   │   ├── models/        # Data models and structs
│   │   ├── routes/        # Route definitions
│   │   ├── services/      # Business logic services
│   │   ├── utils/         # Utility functions
│   │   └── websocket/     # WebSocket management
│   └── Dockerfile         # Backend container config
├── frontend/               # Next.js frontend application
│   ├── src/
│   │   ├── app/           # Next.js app router pages
│   │   ├── components/    # Reusable UI components
│   │   ├── hooks/         # Custom React hooks
│   │   ├── lib/           # Utility libraries
│   │   ├── store/         # Zustand state management
│   │   └── types/         # TypeScript type definitions
│   └── Dockerfile         # Frontend container config
├── minio-scripts/          # MinIO setup scripts
├── docker-compose.yml      # Multi-container orchestration
└── README.md
```

## 🛠️ Technology Stack

### Backend Technologies

- **Go 1.21+**: Core backend language
- **SQLite**: Embedded database
- **WebSockets**: Real-time communication
- **MinIO**: S3-compatible object storage
- **Docker**: Containerization

### Frontend Technologies

- **Next.js 14**: React framework with App Router
- **TypeScript**: Type-safe JavaScript
- **Tailwind CSS**: Utility-first CSS framework
- **Zustand**: Lightweight state management
- **React Hook Form**: Form handling and validation
- **WebSocket API**: Real-time client connections

## 🚀 Getting Started

### Prerequisites

- Docker and Docker Compose
- Git

### Installation & Setup

1. **Clone the repository**

   ```bash
   git clone https://github.com/your-username/social-network.git
   cd social-network
   ```

2. **Environment Configuration**
   Create a .env file in the root directory:

   ```env
   # Database
   DB_PATH=./database.db

   # MinIO Configuration
   MINIO_ROOT_USER=ak-123456
   MINIO_ROOT_PASSWORD=sk-123456
   MINIO_BUCKET_NAME=images

   # Application Settings
   NEXT_PUBLIC_API_URL=http://localhost:8080
   NEXT_PUBLIC_WEBSOCKET_URL=localhost:8080
   NEXT_PUBLIC_MINIO_ENDPOINT=http://localhost:9000
   ```

3. **Start the application**

   ```bash
   docker-compose up --build
   ```

4. **Access the application**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - MinIO Console: http://localhost:9001

### Development Setup

For development with hot reload:

1. **Backend development**

   ```bash
   cd backend
   go mod download
   make dev  # Uses Air for hot reload
   ```

2. **Frontend development**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

## 📖 API Documentation

### Authentication Endpoints

- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `POST /api/auth/logout` - User logout
- `GET /api/auth/check` - Check authentication status

### User Management

- `GET /api/users/profile` - Get current user profile
- `PUT /api/users/profile` - Update user profile
- `GET /api/users/{id}` - Get user by ID
- `POST /api/users/{id}/follow` - Follow/unfollow user
- `GET /api/users/{id}/followers` - Get user followers
- `GET /api/users/{id}/following` - Get users being followed

### Posts & Content

- `GET /api/posts` - Get all posts (with privacy filtering)
- `POST /api/posts` - Create new post
- `GET /api/posts/{id}` - Get specific post
- `POST /api/posts/{id}/like` - Like/unlike post
- `POST /api/posts/{id}/comment` - Add comment to post

### Groups & Events

- `GET /api/groups` - Get all groups
- `POST /api/groups` - Create new group
- `POST /api/groups/{id}/join` - Join group
- `GET /api/groups/{id}/events` - Get group events
- `POST /api/groups/{id}/events` - Create group event

### Real-time Features

- `WebSocket /ws` - Real-time messaging and notifications

## 🔧 Configuration

### Environment Variables

#### Backend Configuration

```env
PORT=8080
DB_PATH=./database.db
SESSION_SECRET=your-session-secret
MINIO_ENDPOINT=http://minio_local_storage:9000
MINIO_ACCESS_KEY=ak-123456
MINIO_SECRET_KEY=sk-123456
MINIO_BUCKET_NAME=images
```

#### Frontend Configuration

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WEBSOCKET_URL=localhost:8080
NEXT_PUBLIC_MINIO_ENDPOINT=http://localhost:9000
NEXT_PUBLIC_MINIO_ACCESS_KEY=ak-123456
NEXT_PUBLIC_MINIO_SECRET_KEY=sk-123456
NEXT_PUBLIC_MINIO_BUCKET_NAME=images
```

### Docker Services

The application runs with the following services:

- **backend_app**: Go backend server (port 8080)
- **frontend_app**: Next.js frontend (port 3000)
- **minio_local_storage**: MinIO object storage (port 9000)
- **nginx**: Reverse proxy and load balancer (port 80)

## 🧪 Testing

### Backend Testing

```bash
cd backend
go test ./...
```

### Frontend Testing

```bash
cd frontend
npm test
```

### API Testing

Use the included Postman collection (postman_collection.json) for API endpoint testing.

## 📝 Database Schema

The application uses SQLite with the following main tables:

- `users` - User accounts and profiles
- `posts` - User posts and content
- `comments` - Post comments
- `likes` - Post likes
- `follows` - User follow relationships
- `groups` - Group information
- `group_members` - Group membership
- `events` - Group events
- `messages` - Private messages
- `notifications` - User notifications

## 🔐 Security Features

- **Session Management**: Secure cookie-based sessions
- **Input Validation**: Comprehensive input sanitization
- **CORS Protection**: Configured for secure cross-origin requests
- **File Upload Security**: Type validation and size limits
- **Privacy Controls**: Granular privacy settings for posts and profiles

## 🚀 Deployment

### Production Deployment

1. Set production environment variables
2. Build and deploy with Docker Compose:
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

### Environment-specific Configuration

- Use .env.production for production frontend settings
- Configure reverse proxy settings in production
- Set up proper SSL certificates for HTTPS

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is part of the 01-edu curriculum. See the [original project requirements](https://github.com/01-edu/public/blob/master/subjects/social-network/README.md) for educational context.

## 🆘 Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure ports 3000, 8080, 9000, and 9001 are available
2. **Docker issues**: Try `docker-compose down && docker-compose up --build`
3. **Database issues**: Check database migrations in migrations
4. **MinIO connectivity**: Verify MinIO container is healthy and accessible

### Logs and Debugging

```bash
# View all service logs
docker-compose logs

# View specific service logs
docker-compose logs backend_app
docker-compose logs frontend_app
```

For additional support, please refer to the project documentation or create an issue in the repository.
