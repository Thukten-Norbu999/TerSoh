FROM golang:latest


# 1) Set the working directory
WORKDIR /app

# 2) Copy only go.mod & go.sum (from the repo root) to leverage Docker’s cache
COPY go.mod go.sum ./
RUN go mod download

# 3) Copy all your code (cmd/, controllers/, internal/, migrations/, models/, pkg/, routes/, sql/, web/, etc.)
COPY . .

# 4) (Optional) If you ever need schema.sql in a "database/" folder
RUN mkdir -p database && cp sql/schema.sql database/schema.sql

# 5) Compile your server binary from the true entrypoint
RUN go build -o tersoh-server cmd/server/main.go

# 6) Expose and run
EXPOSE 8080
CMD ["./tersoh-server"]