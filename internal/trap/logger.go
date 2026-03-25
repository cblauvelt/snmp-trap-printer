package trap

import (
	"fmt"
	"log/slog"
)

// slogAdapter bridges gosnmp's LoggerInterface to slog at Debug level.
// gosnmp uses this logger for internal trace messages as well as errors;
// Debug keeps the noise out of normal output while remaining visible with -v.
type slogAdapter struct{}

func (slogAdapter) Print(v ...any)                { slog.Debug(fmt.Sprint(v...)) }
func (slogAdapter) Printf(f string, v ...any)     { slog.Debug(fmt.Sprintf(f, v...)) }
