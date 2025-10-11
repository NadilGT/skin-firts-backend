FROM golang:1.24-alpine

WORKDIR /app

COPY go.* ./

RUN go mod tidy

COPY . .

RUN go build -o main main.go
# RUN go build -o /TicketMgt (build all the files)

EXPOSE 3000
# Set environment variables (optional, can be overridden)
ENV PORT=3000
ENV FIREBASE_PROJECT_ID=skin-firts
ENV MONGODB_URI=mongodb+srv://admin:W6ptbj7HPS3RJ4cU@cluster0.tgypip5.mongodb.net/

# Command to run the application
CMD ["./main"]