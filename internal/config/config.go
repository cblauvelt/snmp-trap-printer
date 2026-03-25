package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gosnmp/gosnmp"
)

// OutputFormat is the output format for trap printing.
type OutputFormat string

const (
	OutputHuman OutputFormat = "human"
	OutputJSON  OutputFormat = "json"
)

// V3Config holds SNMPv3 credentials loaded from environment variables.
type V3Config struct {
	Username     string
	AuthProtocol gosnmp.SnmpV3AuthProtocol
	AuthPassword string
	PrivProtocol gosnmp.SnmpV3PrivProtocol
	PrivPassword string
}

// SecurityLevel returns the appropriate SnmpV3MsgFlags based on which fields are populated.
func (v *V3Config) SecurityLevel() gosnmp.SnmpV3MsgFlags {
	if v.Username == "" || v.AuthProtocol == gosnmp.NoAuth || v.AuthProtocol == 0 {
		return gosnmp.NoAuthNoPriv
	}
	if v.PrivProtocol == gosnmp.NoPriv || v.PrivProtocol == 0 {
		return gosnmp.AuthNoPriv
	}
	return gosnmp.AuthPriv
}

// LoadV3Config reads SNMPv3 credentials from environment variables.
// Invalid protocol strings log a warning and fall back to NoAuth/NoPriv.
func LoadV3Config() (*V3Config, error) {
	cfg := &V3Config{
		Username:     os.Getenv("SNMP_V3_USERNAME"),
		AuthPassword: os.Getenv("SNMP_V3_AUTH_PASSWORD"),
		PrivPassword: os.Getenv("SNMP_V3_PRIV_PASSWORD"),
		AuthProtocol: gosnmp.NoAuth,
		PrivProtocol: gosnmp.NoPriv,
	}

	if authProto := os.Getenv("SNMP_V3_AUTH_PROTOCOL"); authProto != "" {
		proto := parseAuthProtocol(authProto)
		if proto == gosnmp.NoAuth {
			log.Printf("warning: unknown SNMP_V3_AUTH_PROTOCOL %q, falling back to noAuthNoPriv", authProto)
		}
		cfg.AuthProtocol = proto
	}

	if privProto := os.Getenv("SNMP_V3_PRIV_PROTOCOL"); privProto != "" {
		proto := parsePrivProtocol(privProto)
		if proto == gosnmp.NoPriv {
			log.Printf("warning: unknown SNMP_V3_PRIV_PROTOCOL %q, falling back to noPriv", privProto)
		}
		cfg.PrivProtocol = proto
	}

	return cfg, nil
}

// Config holds all runtime configuration derived from CLI flags and environment variables.
type Config struct {
	Address  string
	Port     uint16
	Output   OutputFormat
	MIBPaths []string

	// V3 holds SNMPv3 credentials loaded from environment variables.
	V3 *V3Config
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
	var output string

	flag.StringVar(&cfg.Address, "address", "0.0.0.0", "Bind address")
	flag.UintVar(&port, "port", 162, "UDP port")
	outputDefault := os.Getenv("OUTPUT_FORMAT")
	if outputDefault == "" {
		outputDefault = string(OutputHuman)
	}
	flag.StringVar(&output, "output", outputDefault, fmt.Sprintf("Output format: %s or %s", OutputHuman, OutputJSON))
	flag.Var(&mibPaths, "mib-path", "Extra MIB directory (repeatable)")
	flag.Parse()

	cfg.Port = uint16(port)
	cfg.MIBPaths = mibPaths
	cfg.Output = OutputFormat(output)

	switch cfg.Output {
	case OutputHuman, OutputJSON:
		// valid
	default:
		fmt.Fprintf(flag.CommandLine.Output(), "invalid --output %q: must be \"human\" or \"json\"\n\n", output)
		flag.Usage()
		os.Exit(2)
	}

	v3, _ := LoadV3Config()
	cfg.V3 = v3

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
