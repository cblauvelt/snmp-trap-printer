package output

import (
	"io"

	"github.com/cblauvelt/snmp-trap-printer/internal/trap"
)

// Formatter writes a single trap to the given writer.
type Formatter interface {
	Format(w io.Writer, t *trap.Trap) error
}
