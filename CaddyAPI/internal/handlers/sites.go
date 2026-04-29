package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/caddyweb/caddyapi/internal/caddy"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Site represents a sub-route (sub-site) within a domain
type Site struct {
	Name        string `json:"name"`         // Site name, e.g., "first"
	Host        string `json:"host"`         // Full domain, e.g., "first.example.com"
	Type        string `json:"type"`         // Type: static, reverse_proxy
	Upstream    string `json:"upstream"`     // Upstream server (reverse proxy)
	Root        string `json:"root"`         // Static files directory
	IndexNames  string `json:"index_names"`  // Default documents, space-separated
	HealthCheck bool   `json:"health_check"` // Health check
	ID          string `json:"id"`           // Handle ID, e.g., "uuid"
	ServerID    string `json:"server_id"`    // Server ID
}

// ListSites returns all sites (sub-routes) from all domains
func ListSites(c *gin.Context) {
	httpConfig, err := caddyClient.GetHTTPConfig()
	if err != nil {
		log.Printf("[ERROR] Failed to get http config: %v", err)
		InternalServerError(c, "Failed to get sites")
		return
	}

	allSites := make([]Site, 0)

	for serverID, server := range httpConfig.Servers {
		for domainID, namedRoute := range server.NamedRoutes {
			for _, handle := range namedRoute.Handle {
				if handle.Handler == "subroute" && len(handle.Routes) > 0 {
					for _, subroute := range handle.Routes {
						r := Site{ServerID: serverID}
						if subroute.Id != "" {
							r.ID = subroute.Id
						}
						for _, match := range subroute.Match {
							if len(match.Host) > 0 {
								r.Host = match.Host[0]
								host := match.Host[0]
								if strings.Contains(host, ".") {
									parts := strings.SplitN(host, ".", 2)
									r.Name = parts[0]
								} else {
									r.Name = host
								}
							}
						}
						leaf := findLeafHandler(subroute.Handle)
						if leaf != nil {
							if leaf.Id != "" {
								r.ID = leaf.Id
							}
							switch leaf.Handler {
							case "reverse_proxy":
								r.Type = "reverse_proxy"
								if len(leaf.Upstreams) > 0 {
									r.Upstream = leaf.Upstreams[0].Dial
								}
								r.HealthCheck = leaf.HealthChecks != nil && leaf.HealthChecks.Active != nil
							case "file_server":
								r.Type = "static"
								r.Root = leaf.Root
								r.IndexNames = strings.Join(leaf.IndexNames, " ")
							}
						}
						if r.ID == "" {
							r.ID = domainID
						}
						allSites = append(allSites, r)
					}
				}
			}
		}
	}

	Success(c, allSites)
}

// GetSite returns a single site by domain and name
func GetSite(c *gin.Context) {
	domainName := c.Param("name")
	siteName := c.Param("site")
	fullHost := siteName + "." + domainName

	httpConfig, err := caddyClient.GetHTTPConfig()
	if err != nil {
		InternalServerError(c, "Failed to get site")
		return
	}

	server, ok := httpConfig.Servers[domainName]
	if !ok {
		NotFound(c, "Domain not found")
		return
	}

	for _, route := range server.Routes {
		for _, handle := range route.Handle {
			if handle.Handler == "subroute" && len(handle.Routes) > 0 {
				for _, subroute := range handle.Routes {
					found := false
					for _, match := range subroute.Match {
						if len(match.Host) > 0 && match.Host[0] == fullHost {
							found = true
							break
						}
					}
					if found {
						site := Site{Name: siteName, Host: fullHost}
						for _, h := range subroute.Handle {
							site.Type = h.Handler
							if h.Id != "" {
								site.ID = h.Id
							}
							if h.Handler == "reverse_proxy" && len(h.Upstreams) > 0 {
								site.Upstream = h.Upstreams[0].Dial
								site.HealthCheck = h.HealthChecks != nil && h.HealthChecks.Active != nil
							} else if h.Handler == "file_server" {
								site.Root = h.Root
								site.IndexNames = strings.Join(h.IndexNames, " ")
							}
						}
						Success(c, site)
						return
					}
				}
			}
		}
	}

	NotFound(c, "Site not found")
}

// CreateSite creates a new site under a domain
func CreateSite(c *gin.Context) {
	domainID := c.Param("id")

	var req struct {
		Name        string `json:"name" binding:"required"` // Site name
		Type        string `json:"type" binding:"required"` // static or reverse_proxy
		ID          string `json:"id"`                      // Optional ID
		Upstream    string `json:"upstream"`                // Upstream for reverse proxy
		Root        string `json:"root"`                    // Root directory for static
		IndexNames  string `json:"index_names"`             // Default documents
		HealthCheck bool   `json:"health_check"`            // Health check
		Domain      string `json:"domain"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid request body")
		return
	}

	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	fullHost := req.Name + "." + req.Domain

	leaf := createLeaf(req.Type, req.Upstream, req.Root, req.IndexNames, req.HealthCheck)

	newSubRoute := caddy.Route{
		Id:    req.ID,
		Match: []caddy.Match{{Host: []string{fullHost}}},
		Handle: []caddy.Handle{
			{
				Handler: "subroute",
				Routes:  []caddy.Route{leaf},
			},
		},
		Terminal: true,
	}

	path := domainID + "/routes/..."
	if err := caddyClient.PostConfigPath(path, true, []caddy.Route{newSubRoute}); err != nil {
		log.Printf("[ERROR] Failed to create site: %v", err)
		InternalServerError(c, "Failed to create site")
		return
	}

	SuccessWithMessage(c, "Site created", gin.H{"domain": req.Domain, "name": req.Name, "id": req.ID})
}

// UpdateSite updates a site
func UpdateSite(c *gin.Context) {
	siteId := c.Param("id")

	var req struct {
		Name        string `json:"name"` // New name for the site (if renaming)
		Type        string `json:"type" binding:"required"`
		Upstream    string `json:"upstream"`
		Root        string `json:"root"`
		IndexNames  string `json:"index_names"`
		HealthCheck bool   `json:"health_check"`
		Domain      string `json:"domain"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid request body")
		return
	}

	data, err := caddyClient.GetConfigPath(siteId, true)
	if err != nil {
		NotFound(c, "Site not found")
		return
	}

	var existingRoute caddy.Route
	if err := json.Unmarshal(data, &existingRoute); err != nil {
		InternalServerError(c, "Failed to parse site config")
		return
	}

	fullHost := req.Name + "." + req.Domain

	leaf := createLeaf(req.Type, req.Upstream, req.Root, req.IndexNames, req.HealthCheck)

	newHandle := []caddy.Handle{
		{
			Handler: "subroute",
			Routes:  []caddy.Route{leaf},
		},
	}
	newMatch := []caddy.Match{{Host: []string{fullHost}}}

	handlePath := siteId + "/handle"
	if err := caddyClient.PatchConfigPath(handlePath, true, newHandle); err != nil {
		log.Printf("[ERROR] Failed to update site: %v", err)
		InternalServerError(c, "Failed to update site")
		return
	}

	matchPath := siteId + "/match"
	if err := caddyClient.PatchConfigPath(matchPath, true, newMatch); err != nil {
		log.Printf("[ERROR] Failed to update site: %v", err)
		InternalServerError(c, "Failed to update site")
		return
	}

	SuccessWithMessage(c, "Site updated", nil)
}

func createLeaf(siteType, upstream, root, indexes string, healthCheck bool) caddy.Route {
	var leaf caddy.Route
	if siteType == "static" {
		leaf = caddy.Route{
			Handle: []caddy.Handle{{Handler: "file_server", Root: root, IndexNames: parseIndexes(indexes)}},
		}
	} else {
		leaf = caddy.Route{
			Handle: []caddy.Handle{{Handler: "reverse_proxy", Upstreams: []caddy.Upstream{{Dial: upstream}}}},
		}
		if healthCheck {
			leaf.Handle[0].HealthChecks = &caddy.HealthChecks{Active: &caddy.ActiveHealthCheck{Path: "/", Interval: "10s", Timeout: "5s"}}
		}
	}
	return leaf
}

func DeleteSite(c *gin.Context) {
	domainID := c.Param("domain_id")
	siteID := c.Param("site_id")

	if err := deleteSiteByID(c, domainID, siteID); err != nil {
		log.Printf("[ERROR] Failed to delete site: %v", err)
		InternalServerError(c, "Failed to delete site")
	}
}

func deleteSiteByID(c *gin.Context, serverID, siteID string) error {
	if siteID == "" || serverID == "" {
		BadRequest(c, "Domain ID and Site ID are required")
		return fmt.Errorf("domain id and site id required")
	}

	if err := caddyClient.DeleteConfigPath(siteID, true); err != nil {
		log.Printf("[ERROR] Failed to delete site: %v", err)
		InternalServerError(c, "Failed to delete site")
		return err
	}

	SuccessWithMessage(c, "Site deleted", nil)
	return nil
}

func deleteSiteByHost(c *gin.Context, domainName, fullHost string) {
	path := "apps/http/servers/" + domainName
	data, err := caddyClient.GetConfigPath(path, false)
	if err != nil {
		InternalServerError(c, "Domain not found")
		return
	}

	var serverConfig caddy.Server
	if err := json.Unmarshal(data, &serverConfig); err != nil {
		InternalServerError(c, "Failed to parse domain config")
		return
	}

	for i := range serverConfig.Routes {
		for j := range serverConfig.Routes[i].Handle {
			if serverConfig.Routes[i].Handle[j].Handler == "subroute" {
				subroutes := serverConfig.Routes[i].Handle[j].Routes
				newSubroutes := make([]caddy.Route, 0, len(subroutes))
				for _, sr := range subroutes {
					keep := true
					for _, m := range sr.Match {
						if slices.Contains(m.Host, fullHost) {
							keep = false
						}
						if !keep {
							break
						}
					}
					if keep {
						newSubroutes = append(newSubroutes, sr)
					}
				}
				serverConfig.Routes[i].Handle[j].Routes = newSubroutes
			}
		}
	}

	if err := caddyClient.PostConfigPath(path, false, serverConfig); err != nil {
		log.Printf("[ERROR] Failed to delete site: %v", err)
		InternalServerError(c, "Failed to delete site")
		return
	}

	SuccessWithMessage(c, "Site deleted", nil)
}

func findLeafHandler(handles []caddy.Handle) *caddy.Handle {
	for i := range handles {
		h := &handles[i]
		if h.Handler == "subroute" && len(h.Routes) > 0 {
			for j := range h.Routes {
				if leaf := findLeafHandler(h.Routes[j].Handle); leaf != nil {
					return leaf
				}
			}
		} else {
			return h
		}
	}
	return nil
}

func parseIndexes(indexes string) []string {
	if indexes == "" {
		return []string{"index.html"}
	}
	parts := make([]string, 0)
	for _, p := range strings.Split(indexes, " ") {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	if len(parts) == 0 {
		return []string{"index.html"}
	}
	return parts
}
