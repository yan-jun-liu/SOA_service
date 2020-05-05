FROM golang:1.13 as builder

#Copy into container
ADD src /src
#Chage working directory to copied location
WORKDIR /src
# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o server

FROM alpine:3
RUN apk add --no-cache ca-certificates
# Copy the binary to the production image from the builder stage.
COPY --from=builder /src/server /server
# Expose port of the container
EXPOSE 8080
# Run the web service on container startup.
CMD ["/server"]
