// Package naming implementa las estrategias de renombrado de archivos de salida.
package naming

import (
	"crypto/rand"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type Strategy string

const (
	Original  Strategy = "original"
	Snowflake Strategy = "snowflake"
	UUID      Strategy = "uuid"
)

// Supported es la lista canónica de estrategias de renombrado aceptadas por CLI.
var Supported = []Strategy{Original, Snowflake, UUID}

// Contains indica si v está en la lista de estrategias dada.
func Contains(list []Strategy, v Strategy) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// Names arma "original, snowflake, uuid" para mensajes de ayuda/error.
func Names(list []Strategy) string {
	out := make([]string, len(list))
	for i, s := range list {
		out[i] = string(s)
	}
	return strings.Join(out, ", ")
}

var snowflakeSeq uint64

// Snowflake genera un ID único basado en timestamp (ms) + worker ID + secuencia,
// similar en espíritu al esquema de Twitter Snowflake.
func GenerateSnowflake() string {
	timestamp := uint64(time.Now().UnixMilli()) << 22
	const workerID = uint64(1) << 17
	sequence := atomic.AddUint64(&snowflakeSeq, 1) - 1
	sequence &= 0xFFF
	return strconv.FormatUint(timestamp|workerID|sequence, 10)
}

// GenerateUUID genera un UUID v4 aleatorio. Si /dev/urandom fallara
// (caso extremadamente improbable), cae de vuelta a GenerateSnowflake.
func GenerateUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return GenerateSnowflake()
	}
	b[6] = (b[6] & 0x0f) | 0x40 // versión 4
	b[8] = (b[8] & 0x3f) | 0x80 // variante RFC 4122
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// Resolve calcula el nombre base (sin extensión) del archivo de salida,
// según la estrategia elegida.
func Resolve(src string, strategy Strategy) string {
	switch strategy {
	case Snowflake:
		return GenerateSnowflake()
	case UUID:
		return GenerateUUID()
	default:
		base := filepath.Base(src)
		return strings.TrimSuffix(base, filepath.Ext(base))
	}
}
