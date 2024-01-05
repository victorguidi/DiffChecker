#STAGE: FRONTEND
FROM node:alpine as frontend
WORKDIR /app
COPY ./frontend/package.json .
RUN npm install @azure/msal-browser
RUN npm install
COPY ./frontend .
RUN npm run build
EXPOSE 3000
ENV NODE_ENV=production
CMD ["npx", "serve", "-s", "build"]

#STAGE: BACKEND
FROM golang:alpine AS build
WORKDIR /app
COPY ./backend/go.mod backend/go.sum ./
RUN go mod download
COPY ./backend .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o docdiff ./src/
CMD ["./docdiff"]

FROM alpine:latest as prod
WORKDIR /app
RUN apk add --no-cache poppler-utils
RUN adduser -D -H -h /app appuser # This steps are for security reason
# 755 is the code for rwxr-xr-x (read write execute for owner, read and execute for group and others)
RUN chown -R appuser /app && \
    chmod -R 755 /app
USER appuser
ENV MONGO_URI=${MONGO_URI}
COPY --from=build /app/docdiff /app/docdiff
EXPOSE 5000
CMD ["./docdiff"]
