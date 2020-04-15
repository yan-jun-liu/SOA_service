FROM golang:alpine

#Copy into container
ADD src /src
#Chage working directory to copied location
WORKDIR /src
#Build runable binary
RUN go build -o server server.go

# Move to /dist directory as the place for resulting binary folder
WORKDIR /go

# Copy binary from root to go folder
RUN cp /src/server .
# Expose port inside of the container
EXPOSE 8080
# run binary
CMD ["/go/server"]