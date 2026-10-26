// Package socket
//
// Copyright (c) 2025-2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  types.go
package socket

import (
	"github.com/GabeCordo/DistributedFunctions/internal/shared/logging"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/socket"
)

const privateKeyPathEnvVar = "DistributedFunctions_SERVER_TLS_KEY"
const certificatePathEnvVar = "DistributedFunctions_SERVER_TLS_CERT"

type UseCases struct {
	Socket socket.Server
	Logger logging.Logger
}
