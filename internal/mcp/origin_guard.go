package mcp

import (
	"net"
	"net/http"
	"net/url"
	"strings"

	"kapua-mcp-server/internal/config"
	"kapua-mcp-server/pkg/utils"
)

type originMiddleware struct {
	allowAll bool
	patterns []originPattern
	logger   *utils.Logger
}

type originPattern struct {
	scheme       string
	host         string
	wildcardPort bool
	port         string
}

func newOriginMiddleware(cfg config.MCPConfig, logger *utils.Logger, next http.Handler) http.Handler {
	policy := compileOriginPolicy(cfg.AllowedOrigins, logger)
	if policy.allowAll {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if originAllowed(r, policy) {
			next.ServeHTTP(w, r)
			return
		}

		logger.Warn("Blocked request with disallowed origin %q", r.Header.Get("Origin"))
		http.Error(w, "origin not allowed", http.StatusForbidden)
	})
}

func compileOriginPolicy(origins []string, logger *utils.Logger) originMiddleware {
	policy := originMiddleware{logger: logger}
	for _, value := range origins {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if value == "*" {
			policy.allowAll = true
			policy.patterns = nil
			return policy
		}

		pattern, err := parseOriginPattern(value)
		if err != nil {
			logger.Warn("Ignoring invalid origin pattern %q: %v", value, err)
			continue
		}
		policy.patterns = append(policy.patterns, pattern)
	}
	return policy
}

func originAllowed(r *http.Request, policy originMiddleware) bool {
	originHeaders := r.Header.Values("Origin")
	if len(originHeaders) == 0 {
		return true
	}

	reqHost, reqPort := splitHostPort(r.Host)
	for _, value := range originHeaders {
		originURL, err := url.Parse(value)
		if err != nil {
			continue
		}
		if originURL.Scheme == "" {
			originURL.Scheme = "http"
		}

		if defaultOriginMatch(originURL, reqHost, reqPort) {
			return true
		}

		if matchConfiguredPatterns(originURL, policy.patterns) {
			return true
		}
	}

	return false
}

func defaultOriginMatch(origin *url.URL, reqHost, reqPort string) bool {
	originHost := strings.ToLower(origin.Hostname())
	originPort := origin.Port()
	if originPort == "" {
		originPort = defaultPortForScheme(origin.Scheme)
	}

	if hostsEquivalent(originHost, reqHost) {
		if portsEquivalent(originPort, reqPort, origin.Scheme) {
			return true
		}
	}

	if isLoopback(originHost) && isLoopback(reqHost) {
		return true
	}

	return false
}

func matchConfiguredPatterns(origin *url.URL, patterns []originPattern) bool {
	if len(patterns) == 0 {
		return false
	}

	host := strings.ToLower(origin.Hostname())
	port := origin.Port()
	if port == "" {
		port = defaultPortForScheme(origin.Scheme)
	}

	for _, pattern := range patterns {
		if pattern.scheme != "" && !strings.EqualFold(pattern.scheme, origin.Scheme) {
			continue
		}
		if pattern.host != "" && !strings.EqualFold(pattern.host, host) {
			continue
		}
		if pattern.wildcardPort {
			return true
		}
		if pattern.port == port {
			return true
		}
	}

	return false
}

func parseOriginPattern(value string) (originPattern, error) {
	var pattern originPattern
	raw := value

	if strings.Contains(value, "://") {
		u, err := url.Parse(value)
		if err != nil {
			return pattern, err
		}
		pattern.scheme = strings.ToLower(u.Scheme)
		pattern.host = strings.ToLower(u.Hostname())
		if u.Port() == "" {
			pattern.wildcardPort = true
		} else {
			pattern.port = u.Port()
		}
		return pattern, nil
	}

	if strings.Contains(value, ":") {
		host, port, err := net.SplitHostPort(value)
		if err != nil {
			// Treat as host only if split fails.
			pattern.host = strings.ToLower(value)
			pattern.wildcardPort = true
			return pattern, nil
		}
		pattern.host = strings.ToLower(host)
		pattern.port = port
		return pattern, nil
	}

	pattern.host = strings.ToLower(raw)
	pattern.wildcardPort = true
	return pattern, nil
}

func splitHostPort(value string) (string, string) {
	if value == "" {
		return "", ""
	}
	if strings.Contains(value, ":") {
		host, port, err := net.SplitHostPort(value)
		if err == nil {
			return strings.ToLower(host), port
		}
	}
	return strings.ToLower(value), ""
}

func portsEquivalent(originPort, reqPort, scheme string) bool {
	if originPort == reqPort {
		return true
	}
	if originPort == "" {
		originPort = defaultPortForScheme(scheme)
	}
	if reqPort == "" {
		reqPort = defaultPortForScheme(scheme)
	}
	return originPort == reqPort
}

func defaultPortForScheme(scheme string) string {
	switch strings.ToLower(scheme) {
	case "https":
		return "443"
	default:
		return "80"
	}
}

func hostsEquivalent(a, b string) bool {
	if strings.EqualFold(a, b) {
		return true
	}
	if isLoopback(a) && isLoopback(b) {
		return true
	}
	return false
}

func isLoopback(host string) bool {
	if host == "" {
		return false
	}
	host = strings.Trim(host, "[]")
	if host == "localhost" || host == "host.docker.internal" {
		return true
	}
	if strings.HasPrefix(host, "127.") {
		return true
	}
	if host == "0.0.0.0" {
		return true
	}
	if host == "::1" {
		return true
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return true
	}
	return false
}
