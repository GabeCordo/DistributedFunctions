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

const coreHostEnvVar = "DistributedFunctions_CORE_HOST"
const tlsCertEnvVar = "DistributedFunctions_CLIENT_TLS_CERT"

type UseCases struct {
	Sock   socket.Client
	Logger logging.Logger
}
