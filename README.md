# Adminsuite

Adminsuite is a comprehensive management system designed to handle various aspects of organizational operations, including HR, CRM, project management, and more.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

- Go (1.18 or later)
- Node.js (20 or later)
- Docker
- Kubernetes

### Installing

1. Clone the repository:
   ```
   git clone https://github.com/Josy-coder/adminsuite.git
   cd adminsuite
   ```

2. Set up the backend:
   ```
   cd backend
   go mod tidy
   ```

3. Set up the frontend:
   ```
   cd ../frontend
   npm install
   ```

### Running the application

1. To run the backend with hot reloading:
   ```
   cd backend
   air
   ```

2. To run the frontend:
   ```
   cd frontend
   npm run dev
   ```

## Running the tests

To run the backend tests:
```
cd backend
go test ./...
```

To run the frontend tests:
```
cd frontend
npm test
```

## Deployment

For production deployment, please refer to the `deploy/` directory and the documentation in `docs/`.

## Built With

- [Go](https://golang.org/) - Backend language
- [Gin](https://github.com/gin-gonic/gin) - Web framework for Go
- [React](https://reactjs.org/) - Frontend library
- [Next.js](https://nextjs.org/) - React framework
