# rooms

Rooms service, written with gRPC, proxied with grpc-gateway websockets.

## Prerequisites
* go
* git

## Usage
1. Install dependencies using `go mod tidy`*
2. Create a .env file (.env.example is present as a template)
3. Build (optionally) & run `cmd/main/main.go`

*go modules uses https to fetch code; consider configuring netrc to let go modules use your https access tokens.
