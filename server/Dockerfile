FROM golang:1.26-alpine

WORKDIR /app

RUN go install github.com/air-verse/air@latest

COPY go.mod* go.sum* ./
RUN if [ -f go.mod ]; then go mod download; fi

COPY . .

EXPOSE 4000

CMD ["air"]
