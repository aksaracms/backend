# Use the official Golang image as the base image
FROM golang:1.17

# Set the working directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the application
RUN go build -o main .

# Set the command to run your application by default
CMD ["./main"]
