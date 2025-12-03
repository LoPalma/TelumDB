package client

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// ParseURL parses a TelumDB connection URL and extracts connection parameters
// Supported formats:
// - telumdb://[user[:password]@]host[:port][/database]
// - host:port (legacy format)
func ParseURL(serverURL string) (*ConnectionParams, error) {
	// Handle legacy host:port format
	if !strings.Contains(serverURL, "://") {
		host, port, err := parseHostPort(serverURL)
		if err != nil {
			return nil, fmt.Errorf("invalid host:port format: %w", err)
		}
		return &ConnectionParams{
			Host: host,
			Port: port,
		}, nil
	}

	// Parse URL format
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL format: %w", err)
	}

	if u.Scheme != "telumdb" {
		return nil, fmt.Errorf("unsupported URL scheme: %s (expected 'telumdb')", u.Scheme)
	}

	params := &ConnectionParams{
		Host:     u.Hostname(),
		Database: strings.TrimPrefix(u.Path, "/"),
	}

	// Extract port
	if u.Port() != "" {
		port, err := strconv.Atoi(u.Port())
		if err != nil {
			return nil, fmt.Errorf("invalid port number: %w", err)
		}
		params.Port = port
	} else {
		params.Port = 5432 // default port
	}

	// Extract user credentials
	if u.User != nil {
		params.Username = u.User.Username()
		if password, ok := u.User.Password(); ok {
			params.Password = password
		}
	}

	// Validate host
	if params.Host == "" {
		return nil, fmt.Errorf("host is required in connection URL")
	}

	return params, nil
}

// ConnectionParams holds parsed connection parameters
type ConnectionParams struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

// Address returns the network address (host:port) for dialing
func (p *ConnectionParams) Address() string {
	return net.JoinHostPort(p.Host, strconv.Itoa(p.Port))
}

// String returns a string representation of the connection parameters
func (p *ConnectionParams) String() string {
	var parts []string

	if p.Username != "" {
		if p.Password != "" {
			parts = append(parts, fmt.Sprintf("%s:%s@", p.Username, p.Password))
		} else {
			parts = append(parts, fmt.Sprintf("%s@", p.Username))
		}
	}

	parts = append(parts, p.Host)

	if p.Port != 5432 {
		parts = append(parts, fmt.Sprintf(":%d", p.Port))
	}

	if p.Database != "" {
		parts = append(parts, fmt.Sprintf("/%s", p.Database))
	}

	return "telumdb://" + strings.Join(parts, "")
}

// parseHostPort parses a legacy host:port string
func parseHostPort(hostPort string) (string, int, error) {
	host, portStr, err := net.SplitHostPort(hostPort)
	if err != nil {
		return "", 0, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("invalid port number: %w", err)
	}

	if port <= 0 || port > 65535 {
		return "", 0, fmt.Errorf("port number out of range: %d", port)
	}

	return host, port, nil
}
