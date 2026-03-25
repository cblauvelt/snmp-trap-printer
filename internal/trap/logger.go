package trap

import (
	"fmt"
	"log/slog"
)

// slogAdapter bridges gosnmp's LoggerInterface to slog at Warn level.
// gosnmp uses this logger to report authentication and decryption failures.
type slogAdapter struct{}

func (slogAdapter) Print(v ...any)                { slog.Warn(fmt.Sprint(v...)) }
func (slogAdapter) Printf(f string, v ...any)     { slog.Warn(fmt.Sprintf(f, v...)) }
