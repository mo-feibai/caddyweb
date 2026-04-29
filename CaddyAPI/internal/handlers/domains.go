package handlers

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/caddyweb/caddyapi/internal/caddy"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Domain represents a domain configuration
type Domain struct {
	Name       string   `json:"name"`        // Domain name, e.g., "example.com"
	Wildcard   string   `json:"wildcard"`    // Wildcard domain, e.g., "*.example.com"
	ServerID   string   `json:"server_id"`   // Server ID, e.g., "caddyweb_https"
	Listen     []string `json:"listen"`      // Listen addresses, e.g., [":443"]
	TLSEnabled bool     `json:"tls_enabled"` // Whether TLS is enabled
	ID         string   `json:"id"`          // Handle ID, e.g., "uuid"
}

// SiteRoute represents a sub-site under a domain
type SiteRoute struct {
	Name        string `json:"name"`         // Site name, e.g., "first"
	Host        string `json:"host"`         // Full domain, e.g., "first.example.com"
	Type        string `json:"type"`         // Type: static, reverse_proxy
	Upstream    string `json:"upstream"`     // Upstream server (reverse proxy)
	Root        string `json:"root"`         // Static files directory
	IndexNames  string `json:"index_names"`  // Default documents, space-separated
	HealthCheck bool   `json:"health_check"` // Health check
	ID          string `json:"id"`           // Route ID
	DomainID    string `json:"domainID"`     // Domain ID
}

// ListDomains returns all registered domains
func ListDomains(c *gin.Context) {
	httpConfig, err := caddyClient.GetHTTPConfig()
	if err != nil {
		log.Printf("[ERROR] Failed to get http config: %v", err)
		InternalServerError(c, "Failed to get domains")
		return
	}

	servers := httpConfig.Servers
	domainMap := make(map[string]Domain)

	for serverID, server := range servers {
		for _, route := range server.Routes {
			var baseDomain, wildcardDomain string
			handleID := route.Id
			for _, match := range route.Match {
				for _, h := range match.Host {
					if len(h) > 1 && h[0] == '*' {
						wildcardDomain = h
					} else if h[0] != '*' {
						baseDomain = h
					}
				}
			}
			for _, h := range route.Handle {
				if h.Handler == "subroute" && h.Id != "" {
					handleID = h.Id
					break
				}
			}

			if baseDomain != "" {
				if _, exists := domainMap[baseDomain]; !exists {
					domainMap[baseDomain] = Domain{
						Name:       baseDomain,
						Wildcard:   wildcardDomain,
						ServerID:   serverID,
						Listen:     server.Listen,
						TLSEnabled: len(server.TLSPolicies) > 0,
						ID:         handleID,
					}
				}
			}
		}
	}

	domains := make([]Domain, 0, len(domainMap))
	for _, d := range domainMap {
		domains = append(domains, d)
	}

	Success(c, domains)
}

// GetDomain returns a single domain
func GetDomain(c *gin.Context) {
	name := c.Param("name")

	httpConfig, err := caddyClient.GetHTTPConfig()
	if err != nil {
		InternalServerError(c, "Failed to get domain")
		return
	}

	for serverID, server := range httpConfig.Servers {
		for _, route := range server.Routes {
			for _, match := range route.Match {
				for _, h := range match.Host {
					if h == name || h == "*."+name {
						domain := Domain{
							Name:       name,
							ServerID:   serverID,
							TLSEnabled: len(server.TLSPolicies) > 0,
						}
						if len(h) > 1 && h[0] == '*' {
							domain.Wildcard = h
						}
						for _, hh := range route.Handle {
							if hh.Handler == "subroute" && hh.Id != "" {
								domain.ID = hh.Id
								break
							}
						}
						Success(c, domain)
						return
					}
				}
			}
		}
	}

	NotFound(c, "Domain not found")
}

// CreateDomain creates a domain configuration
func CreateDomain(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"` // Domain name
		ServerID string `json:"server_id"`               // Server ID, can be caddyweb_https or caddyweb_http
		ID       string `json:"id"`                      // Optional ID, will be auto-generated if not provided
		TLS      bool   `json:"tls"`                     // Enable TLS
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid request body")
		return
	}

	if req.ServerID == "" {
		req.ServerID = "caddyweb_https"
	}

	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	domain := req.Name
	wildcard := "*." + domain

	serverPath := req.ServerID + "/routes"

	newRoute := caddy.Route{
		Id: req.ID,
		Match: []caddy.Match{
			{
				Host: []string{domain, wildcard},
			},
		},
		Handle: []caddy.Handle{
			{
				Handler: "subroute",
				Routes:  make([]caddy.Route, 0),
			},
		},
		Terminal: true,
	}

	if err := caddyClient.PostConfigPath(serverPath, true, newRoute); err != nil {
		log.Printf("[ERROR] Failed to create domain: %v", err)
		InternalServerError(c, "Failed to create domain")
		return
	}

	SuccessWithMessage(c, "Domain created", gin.H{"name": req.Name, "server_id": req.ServerID, "id": req.ID})
}

// DeleteDomain deletes a domain
func DeleteDomain(c *gin.Context) {
	serverID := c.Param("server_id")
	domainID := c.Param("domain_id")

	if serverID == "" || domainID == "" {
		BadRequest(c, "Server ID and Domain ID are required")
		return
	}

	path := domainID
	data, err := caddyClient.GetConfigPath(domainID, true)
	if err != nil {
		NotFound(c, "Domain not found")
		return
	}

	var routeConfig struct {
		Handle []struct {
			Handler string `json:"handler"`
			Routes  []struct {
				Handle []struct {
					Handler string `json:"handler"`
				} `json:"handle"`
			} `json:"routes"`
		} `json:"handle"`
	}
	if err := json.Unmarshal(data, &routeConfig); err == nil {
		hasSubRoutes := false

		for _, h := range routeConfig.Handle {
			if h.Handler == "subroute" && len(h.Routes) > 0 {
				hasSubRoutes = true
				break
			}
		}
		if hasSubRoutes {
			BadRequest(c, "Cannot delete domain: there are sub-sites under this domain. Please delete all sub-sites first.")
			return
		}
	}

	if err := caddyClient.DeleteConfigPath(path, true); err != nil {
		log.Printf("[ERROR] Failed to delete domain: %v", err)
		InternalServerError(c, "Failed to delete domain")
		return
	}

	SuccessWithMessage(c, "Domain deleted", nil)
}

// GetDomainSites returns all sites (sub-routes) for a domain
func GetDomainSites(c *gin.Context) {
	domainID := c.Param("domain_id")

	// Fetch the whole server config for this domain id
	path := domainID
	data, err := caddyClient.GetConfigPath(path, true)
	if err != nil {
		NotFound(c, "Domain not found")
		return
	}

	var domainRoute caddy.Route
	if err := json.Unmarshal(data, &domainRoute); err != nil {
		InternalServerError(c, "Failed to parse domain config")
		return
	}

	sites := make([]SiteRoute, 0)

	r := SiteRoute{DomainID: domainID}
	for _, m := range domainRoute.Match {
		if len(m.Host) > 0 {
			host := m.Host[0]
			r.Host = host
			if _, ok := strings.CutPrefix(host, "*."); ok {
				// wildcard domain like *.example.com -> DomainID = example.com, Name unknown
				r.Name = ""
			} else if strings.Contains(host, ".") {
				parts := strings.SplitN(host, ".", 2)
				r.Name = parts[0]
			} else {
				r.Name = host
			}
		}
	}
	for _, sub := range domainRoute.Handle {
		if sub.Handler == "subroute" {
			for _, leaf := range sub.Routes {
				if leaf.Id != "" {
					r.ID = leaf.Id
				}
				handle := findLeafHandler(leaf.Handle)

				r.Type = handle.Handler
				if handle.Handler == "reverse_proxy" && len(handle.Upstreams) > 0 {
					r.Upstream = handle.Upstreams[0].Dial
					r.HealthCheck = handle.HealthChecks != nil && handle.HealthChecks.Active != nil
				} else if handle.Handler == "file_server" {
					r.Root = handle.Root
					r.IndexNames = strings.Join(handle.IndexNames, " ")
				}

				sites = append(sites, r)
			}
		}
	}

	Success(c, sites)
}
