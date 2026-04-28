package models

// Domain represents a domain configuration stored separately
type Domain struct {
	Name       string   `json:"name"`        // Domain name, e.g., "example.com"
	Wildcard   string   `json:"wildcard"`    // Wildcard domain, e.g., "*.example.com"
	Listen     []string `json:"listen"`      // Listen ports, e.g., [":443"]
	TLSEnabled bool     `json:"tls_enabled"` // Whether TLS is enabled
}
