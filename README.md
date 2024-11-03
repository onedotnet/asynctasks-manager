This project is an asynchronous task management system built with Go. It leverages the Gin web framework for handling HTTP requests and RabbitMQ for task queue management. The project structure is organized into several directories, each serving a specific purpose:

api/v1/handler: Contains HTTP handlers for various endpoints, including image uploads and task management.

image.go: Handles image uploads for tasks and artifacts.
node.go: Manages task nodes.
routes.go: Initializes API routes.
task.go: Manages task-related operations.
cmd: Contains the entry points for different commands.

migrate.go: Handles database migrations.
root.go: Defines the root command for the CLI.
start.go: Starts the HTTP server.
test.go: Contains test commands.
config: Manages configuration settings.

config.go: Loads and parses configuration from environment variables and files.
data: Placeholder for data-related files.

database: Manages database connections and operations.

db.go: Initializes the database connection.
pkg/taskmanager: Contains the core logic for task management.

amqp.go: Manages RabbitMQ connections and message handling.
node.go: Defines task node operations.
task.go: Defines task operations and states.
uploads: Directory for storing uploaded files.

utils: Placeholder for utility functions.

Dockerfile: Defines the Docker build process, including a multi-stage build to compile the Go application and run it in an Alpine Linux container.

go.mod and go.sum: Manage Go module dependencies.

.env: Contains environment variables for configuration.

The project is designed to handle asynchronous tasks, such as image processing, by distributing them across available task nodes. It supports uploading images, creating tasks, and managing task states through a RESTful API. The system uses RabbitMQ for task queuing and PostgreSQL for persistent storage.