package handlers

import (
	"encoding/json"
	"fmt"
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
	IndexNames  string `json:"index_names"`      // Default documents, space-separated
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
		for id, route := range server.NamedRoutes {
			var baseDomain, wildcardDomain string
			handleID := id
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

	serverPath := req.ServerID + "/named_routes/" + req.ID

	newRoute := caddy.Route{
		Match: []caddy.Match{
			{
				Host: []string{domain, wildcard},
			},
		},
		Handle: []caddy.Handle{
			{
				Handler: "subroute",
				Id:      req.ID,
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

// UpdateDomain updates a domain configuration
func UpdateDomain(c *gin.Context) {
	name := c.Param("name")

	var req struct {
		Listen     string `json:"listen"`
		TLSEnabled bool   `json:"tls"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid request body")
		return
	}

	// Get existing config
	path := "apps/http/servers/" + name
	data, err := caddyClient.GetConfigPath(path, true)
	if err != nil {
		NotFound(c, "Domain not found")
		return
	}

	var serverConfig caddy.Server
	if err := json.Unmarshal(data, &serverConfig); err != nil {
		InternalServerError(c, "Failed to parse domain config")
		return
	}

	// Update fields
	if req.Listen != "" {
		serverConfig.Listen = []string{req.Listen}
	}

	if req.TLSEnabled && len(serverConfig.TLSPolicies) == 0 {
		serverConfig.TLSPolicies = []caddy.TLSPolicy{{}}
	} else if !req.TLSEnabled && len(serverConfig.TLSPolicies) > 0 {
		serverConfig.TLSPolicies = nil
	}

	// Save
	if err := caddyClient.PutConfigPath(path, false, serverConfig); err != nil {
		log.Printf("[ERROR] Failed to update domain: %v", err)
		InternalServerError(c, "Failed to update domain")
		return
	}

	SuccessWithMessage(c, "Domain updated", nil)
}

// DeleteDomain deletes a domain
func DeleteDomain(c *gin.Context) {
	serverID := c.Param("server_id")
	domainID := c.Param("domain_id")

	if serverID == "" || domainID == "" {
		BadRequest(c, "Server ID and Domain ID are required")
		return
	}

	path := fmt.Sprintf("%s/named_routes/%s", serverID, domainID)
	data, err := caddyClient.GetConfigPath(path, true)
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

	var serverConfig struct {
		Routes []struct {
			Id       string         `json:"@id"`
			Match    []caddy.Match  `json:"match"`
			Handle   []caddy.Handle `json:"handle"`
			Terminal bool           `json:"terminal"`
		} `json:"routes"`
	}
	if err := json.Unmarshal(data, &serverConfig); err != nil {
		InternalServerError(c, "Failed to parse domain config")
		return
	}

	sites := make([]SiteRoute, 0)

	for _, route := range serverConfig.Routes {
		r := SiteRoute{DomainID: domainID}
		if route.Id != "" {
			r.ID = route.Id
		}
		for _, m := range route.Match {
			if len(m.Host) > 0 {
				host := m.Host[0]
				r.Host = host
				if base, ok := strings.CutPrefix(host, "*."); ok {
					// wildcard domain like *.example.com -> DomainID = example.com, Name unknown
					r.DomainID = base
					r.Name = ""
				} else if strings.Contains(host, ".") {
					parts := strings.SplitN(host, ".", 2)
					r.Name = parts[0]
				} else {
					r.Name = host
				}
			}
		}
		for _, sub := range route.Handle {
			if sub.Handler == "subroute" {
				for _, leaf := range sub.Routes {
					// leaf's own handles determine type/upstream/root
					for _, hl := range leaf.Handle {
						r.Type = hl.Handler
						if hl.Id != "" {
							r.ID = hl.Id
						}
						if hl.Handler == "reverse_proxy" && len(hl.Upstreams) > 0 {
							r.Upstream = hl.Upstreams[0].Dial
							r.HealthCheck = hl.HealthChecks != nil && hl.HealthChecks.Active != nil
						} else if hl.Handler == "file_server" {
							r.Root = hl.Root
							r.IndexNames = strings.Join(hl.IndexNames, " ")
						}
					}
					sites = append(sites, r)
				}
			}
		}
	}

	Success(c, sites)
}
