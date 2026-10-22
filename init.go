package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	pbfriends "github.com/PretendoNetwork/grpc/go/friends"
	"github.com/PretendoNetwork/nex-go/v2"

	"github.com/PretendoNetwork/plogger-go"
	"github.com/PretendoNetwork/splatoon/globals"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func init() {
	globals.Logger = plogger.NewLogger()

	var err error

	err = godotenv.Load()
	if err != nil {
		globals.Logger.Warning("Error loading .env file")
	}

	authenticationServerPort := os.Getenv("PN_SPLATOON_AUTHENTICATION_SERVER_PORT")
	secureServerHost := os.Getenv("PN_SPLATOON_SECURE_SERVER_HOST")
	secureServerPort := os.Getenv("PN_SPLATOON_SECURE_SERVER_PORT")
	friendsGRPCHost := os.Getenv("PN_SPLATOON_FRIENDS_GRPC_HOST")
	friendsGRPCPort := os.Getenv("PN_SPLATOON_FRIENDS_GRPC_PORT")
	friendsGRPCAPIKey := os.Getenv("PN_SPLATOON_FRIENDS_GRPC_API_KEY")
	localAuthMode := os.Getenv("PN_SPLATOON_LOCAL_AUTH")
	healthCheckPort := os.Getenv("PN_SPLATOON_HEALTH_CHECK_PORT")

	kerberosPassword := make([]byte, 0x10)
	_, err = rand.Read(kerberosPassword)
	if err != nil {
		globals.Logger.Error("Error generating Kerberos password")
		os.Exit(0)
	}

	globals.KerberosPassword = string(kerberosPassword)

	globals.InitAccounts()

	if strings.TrimSpace(authenticationServerPort) == "" {
		globals.Logger.Error("PN_SPLATOON_AUTHENTICATION_SERVER_PORT environment variable not set")
		os.Exit(0)
	}

	if port, err := strconv.Atoi(authenticationServerPort); err != nil {
		globals.Logger.Errorf("PN_SPLATOON_AUTHENTICATION_SERVER_PORT is not a valid port. Expected 0-65535, got %s", authenticationServerPort)
		os.Exit(0)
	} else if port < 0 || port > 65535 {
		globals.Logger.Errorf("PN_SPLATOON_AUTHENTICATION_SERVER_PORT is not a valid port. Expected 0-65535, got %s", authenticationServerPort)
		os.Exit(0)
	}

	if strings.TrimSpace(secureServerHost) == "" {
		globals.Logger.Error("PN_SPLATOON_SECURE_SERVER_HOST environment variable not set")
		os.Exit(0)
	}

	if strings.TrimSpace(secureServerPort) == "" {
		globals.Logger.Error("PN_SPLATOON_SECURE_SERVER_PORT environment variable not set")
		os.Exit(0)
	}

	if port, err := strconv.Atoi(secureServerPort); err != nil {
		globals.Logger.Errorf("PN_SPLATOON_SECURE_SERVER_PORT is not a valid port. Expected 0-65535, got %s", secureServerPort)
		os.Exit(0)
	} else if port < 0 || port > 65535 {
		globals.Logger.Errorf("PN_SPLATOON_SECURE_SERVER_PORT is not a valid port. Expected 0-65535, got %s", secureServerPort)
		os.Exit(0)
	}

	// Login tokens are opened locally (AES key) and NEX passwords derived (HMAC secret); both are
	// shared with the service that issues the tokens, so no account gRPC server is needed.
	globals.TokenAESKey, err = hex.DecodeString(os.Getenv("PN_NEX_TOKEN_AES_KEY"))
	if err != nil || len(globals.TokenAESKey) != 32 {
		globals.Logger.Error("PN_NEX_TOKEN_AES_KEY must be a 64-character hex AES-256 key")
		os.Exit(0)
	}

	globals.NEXPasswordSecret, err = hex.DecodeString(os.Getenv("PN_NEX_PASSWORD_SECRET"))
	if err != nil || len(globals.NEXPasswordSecret) < 32 {
		globals.Logger.Error("PN_NEX_PASSWORD_SECRET must be at least 32 bytes encoded as hexadecimal")
		os.Exit(0)
	}

	if strings.TrimSpace(friendsGRPCHost) == "" {
		globals.Logger.Error("PN_SPLATOON_FRIENDS_GRPC_HOST environment variable not set")
		os.Exit(0)
	}

	if strings.TrimSpace(friendsGRPCPort) == "" {
		globals.Logger.Error("PN_SPLATOON_FRIENDS_GRPC_PORT environment variable not set")
		os.Exit(0)
	}

	if port, err := strconv.Atoi(friendsGRPCPort); err != nil {
		globals.Logger.Errorf("PN_SPLATOON_FRIENDS_GRPC_PORT is not a valid port. Expected 0-65535, got %s", friendsGRPCPort)
		os.Exit(0)
	} else if port < 0 || port > 65535 {
		globals.Logger.Errorf("PN_SPLATOON_FRIENDS_GRPC_PORT is not a valid port. Expected 0-65535, got %s", friendsGRPCPort)
		os.Exit(0)
	}

	if strings.TrimSpace(friendsGRPCAPIKey) == "" {
		globals.Logger.Warning("Insecure gRPC server detected. PN_SPLATOON_FRIENDS_GRPC_API_KEY environment variable not set")
	}

	globals.GRPCFriendsClientConnection, err = grpc.NewClient(fmt.Sprintf("dns:%s:%s", friendsGRPCHost, friendsGRPCPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		globals.Logger.Criticalf("Failed to connect to account gRPC server: %v", err)
		os.Exit(0)
	}

	globals.GRPCFriendsClient = pbfriends.NewFriendsClient(globals.GRPCFriendsClientConnection)
	globals.GRPCFriendsCommonMetadata = metadata.Pairs(
		"X-API-Key", friendsGRPCAPIKey,
	)

	globals.LocalAuthMode = localAuthMode == "1"
	if globals.LocalAuthMode {
		globals.Logger.Warning("Local authentication mode is enabled. Token validation will be skipped!")
		globals.Logger.Warning("This is insecure and could allow ban bypasses!")
	}

	globals.Postgres, err = sql.Open("postgres", os.Getenv("PN_SPLATOON_POSTGRES_URI"))
	if err != nil {
		globals.Logger.Critical(err.Error())
	}
	globals.Logger.Success("Connected to Postgres!")

	if strings.TrimSpace(healthCheckPort) == "" {
		globals.Logger.Warning("Basic UDP health check will not be enabled. PN_SPLATOON_HEALTH_CHECK_PORT environment variable not set")
	} else if port, err := strconv.Atoi(healthCheckPort); err != nil {
		globals.Logger.Errorf("PN_SPLATOON_HEALTH_CHECK_PORT is not a valid port. Expected 0-65535, got %s", healthCheckPort)
		os.Exit(0)
	} else if port < 0 || port > 65535 {
		globals.Logger.Errorf("PN_SPLATOON_HEALTH_CHECK_PORT is not a valid port. Expected 0-65535, got %s", healthCheckPort)
		os.Exit(0)
	} else {
		nex.EnableBasicUDPHealthCheck(port)
	}
}
