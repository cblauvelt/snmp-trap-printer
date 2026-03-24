package config

import (
	"flag"
	"os"
	"strings"

	"github.com/gosnmp/gosnmp"
)

// Config holds all runtime configuration derived from CLI flags and environment variables.
type Config struct {
	Address  string
	Port     uint16
	Output   string
	MIBPaths []string

	// SNMPv3 credentials from environment variables
	V3Username     string
	V3AuthProtocol gosnmp.SnmpV3AuthProtocol
	V3AuthPassword string
	V3PrivProtocol gosnmp.SnmpV3PrivProtocol
	V3PrivPassword string
	V3SecurityLevel gosnmp.SnmpV3MsgFlags
}

type mibPathList []string

func (m *mibPathList) String() string {
	return strings.Join(*m, ",")
}

func (m *mibPathList) Set(v string) error {
	*m = append(*m, v)
	return nil
}

// Parse reads CLI flags and environment variables into a Config.
func Parse() *Config {
	cfg := &Config{}

	var mibPaths mibPathList
	var port uint

	flag.StringVar(&cfg.Address, "address", "0.0.0.0", "Bind address")
	flag.UintVar(&port, "port", 162, "UDP port")
	flag.StringVar(&cfg.Output, "output", "human", "Output format: human or json")
	flag.Var(&mibPaths, "mib-path", "Extra MIB directory (repeatable)")
	flag.Parse()

	cfg.Port = uint16(port)
	cfg.MIBPaths = mibPaths

	cfg.V3Username = os.Getenv("SNMP_V3_USERNAME")
	cfg.V3AuthPassword = os.Getenv("SNMP_V3_AUTH_PASSWORD")
	cfg.V3PrivPassword = os.Getenv("SNMP_V3_PRIV_PASSWORD")
	cfg.V3AuthProtocol = parseAuthProtocol(os.Getenv("SNMP_V3_AUTH_PROTOCOL"))
	cfg.V3PrivProtocol = parsePrivProtocol(os.Getenv("SNMP_V3_PRIV_PROTOCOL"))
	cfg.V3SecurityLevel = inferSecurityLevel(cfg)

	return cfg
}

func parseAuthProtocol(s string) gosnmp.SnmpV3AuthProtocol {
	switch strings.ToUpper(s) {
	case "MD5":
		return gosnmp.MD5
	case "SHA":
		return gosnmp.SHA
	case "SHA224":
		return gosnmp.SHA224
	case "SHA256":
		return gosnmp.SHA256
	case "SHA384":
		return gosnmp.SHA384
	case "SHA512":
		return gosnmp.SHA512
	default:
		return gosnmp.NoAuth
	}
}

func parsePrivProtocol(s string) gosnmp.SnmpV3PrivProtocol {
	switch strings.ToUpper(s) {
	case "DES":
		return gosnmp.DES
	case "AES":
		return gosnmp.AES
	case "AES192":
		return gosnmp.AES192
	case "AES256":
		return gosnmp.AES256
	default:
		return gosnmp.NoPriv
	}
}

func inferSecurityLevel(cfg *Config) gosnmp.SnmpV3MsgFlags {
	if cfg.V3Username == "" {
		return gosnmp.NoAuthNoPriv
	}
	if cfg.V3AuthProtocol == gosnmp.NoAuth {
		return gosnmp.NoAuthNoPriv
	}
	if cfg.V3PrivProtocol == gosnmp.NoPriv {
		return gosnmp.AuthNoPriv
	}
	return gosnmp.AuthPriv
}
