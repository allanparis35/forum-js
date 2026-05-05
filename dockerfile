FROM golang:alpine

# Installation des dépendances système pour Go et Postgres
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copie des fichiers de modules
COPY go.mod ./

RUN go mod download

# Copie du reste du code
COPY . .

# Compilation
RUN go build -o main ./api/main.go

EXPOSE 8080

CMD ["./main"]