# Splatoon (Wii U) replacement server

> Fork of [PretendoNetwork/splatoon](https://github.com/PretendoNetwork/splatoon) for setups that do not run the Pretendo account server. AGPL-3.0, same as upstream.
>
> - Login tokens are opened locally (`PN_NEX_TOKEN_AES_KEY`, Pretendo's AES token format) and NEX passwords are derived from the PID with `PN_NEX_PASSWORD_SECRET`, so no account gRPC service is needed. `globals/local_token.go` is adapted from `nex-protocols-common-go` v2.4.0, which later moved this check to the account server.
> - The friends gRPC service is still used for friends matchmaking.
> - Game parameters (retail Splatoon: access key `6f599f81`, game server ID `10162B00`, NEX 3.8.15) match the retail executable.
> - Tested on Cemu only.

Includes both the authentication and secure servers

## Compiling

### Setup
Install [Go](https://go.dev/doc/install) and [git](https://git-scm.com/downloads), then clone and enter the repository

```bash
$ git clone https://github.com/PretendoNetwork/Splatoon
$ cd Splatoon
```

### Compiling using `go`
To compile using Go, `go get` the required modules and then `go build` to your desired location. You may also want to tidy the go modules, though this is optional

```bash
$ go get 
$ go mod tidy
$ go build -o build/splatoon
```

The server is now built to `build/splatoon`

When compiling with only Go, the authentication servers build string is not automatically set. This should not cause any issues with gameplay, but it means that the server build will not be visible in any packet dumps or logs a title may produce

To compile the servers with the authentication server build string, add `-ldflags "-X 'main.serverBuildString=BUILD_STRING_HERE'"` to the build command, or use `make` to compile the server

### Compiling using `make`
Compiling using `make` will read the local `.git` directory to create a dynamic authentication server build string, based on your repositories remote origin and current commit. It will also use the current folders name as the executables name

Install `make` onto your system (this varies by OS), and run `make` while inside the repository

```bash
$ make
```

The server is now built to `build/splatoon` with the authentication server build string already set

### Developing core libraries
This server depends on [nex-protocols-common-go](https://github.com/PretendoNetwork/nex-protocols-common-go),
[nex-protocols-go](https://github.com/PretendoNetwork/nex-protocols-go), and
[nex-go](https://github.com/PretendoNetwork/nex-go) to implement the bulk of its functionality. Sometimes features or
bugfixes depend on unreleased development versions of these libraries; or changes have to be made at the library level
to work on this server.

To use these unreleased libraries, add a `go.work` file with something like the following content:
```go
go 1.25

use (
	.
)

replace (
	github.com/PretendoNetwork/nex-go/v2 => ../nex-go
	github.com/PretendoNetwork/nex-protocols-go/v2 => ../nex-protocols-go
	github.com/PretendoNetwork/nex-protocols-common-go/v2 => ../nex-protocols-common-go
)
```

The build will then build these libraries out of the folders specified instead of pulling the released versions.

## Configuration
All configuration options are handled via environment variables

`.env` files are supported

| Name                                      | Description                                                                                     | Required                                      |
|-------------------------------------------|-------------------------------------------------------------------------------------------------|-----------------------------------------------|
| `PN_SPLATOON_POSTGRES_URI`               | Fully qualified URI to your Postgres server (Example `postgres://username:password@localhost/splatoon?sslmode=disable`)  | Yes                   |
| `PN_SPLATOON_AUTHENTICATION_SERVER_PORT` | Port for the authentication server                                                              | Yes                                            |
| `PN_SPLATOON_SECURE_SERVER_HOST`         | Host name for the secure server (should point to the same address as the authentication server) | Yes                                            |
| `PN_SPLATOON_SECURE_SERVER_PORT`         | Port for the secure server                                                                      | Yes                                            |
| `PN_NEX_TOKEN_AES_KEY`                   | AES-256 key (64 hex chars) used to open the NEX login token                                      | Yes |
| `PN_NEX_PASSWORD_SECRET`                 | HMAC secret (hex, 32+ bytes) used to derive per-PID NEX passwords                                | Yes |
| `PN_SPLATOON_FRIENDS_GRPC_HOST`          | Host name for Friends gRPC service (used for friends matchmaking)                               | Yes                                            |
| `PN_SPLATOON_FRIENDS_GRPC_PORT`          | Port for Friends gRPC service                                                                   | Yes                                            |
| `PN_SPLATOON_FRIENDS_GRPC_API_KEY`       | API key for Friends gRPC service                                                                | No (Assumed to be an open gRPC API)            |
| `PN_SPLATOON_LOCAL_AUTH_MODE`            | Use local file for NEX credentials rather than talking to account server (INSECURE)             | No                                             |
| `PN_SPLATOON_HEALTH_CHECK_PORT`          | Port for the UDP health check                                                                   | No                                             |