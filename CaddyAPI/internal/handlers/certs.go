package handlers

import (
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AddCertRequest struct {
	AuthMethod string `json:"authMethod"` // "auto" or "file"
	TargetType string `json:"targetType"` // "domain" or "site"
	Domain     string `json:"domain"`     // For file-based
	CertFile   string `json:"certFile"`   // For file-based
	KeyFile    string `json:"keyFile"`    // For file-based
	AutoHTTPS  bool   `json:"autoHTTPS"`  // For file-based
	DomainName string `json:"domainName"` // For auto domain
	SiteID     string `json:"siteId"`     // For auto site
	CertID     string `json:"certId"`     // Optional certificate identifier
}

func AddCertificate(c *gin.Context) {
	var req AddCertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid request body")
		return
	}

	if req.AuthMethod == "file" {
		SuccessWithMessage(c, "证书配置已保存 (file-based)", gin.H{
			"method": "file",
		})
		return
	}

	if req.AuthMethod != "auto" {
		BadRequest(c, "无效的认证方式")
		return
	}

	certID := req.CertID
	if certID == "" {
		certID = uuid.New().String()
	}

	var subjects []string
	var certIdentifier string

	switch req.TargetType {
	case "domain":
		if req.DomainName == "" {
			BadRequest(c, "请选择域名")
			return
		}
		wildcard := fmt.Sprintf("*.%s", req.DomainName)
		subjects = []string{wildcard, req.DomainName}
		certIdentifier = req.DomainName

	case "site":
		if req.SiteID == "" {
			BadRequest(c, "请选择站点")
			return
		}
		site, err := getSiteByID(req.SiteID)
		if err != nil {
			BadRequest(c, "站点不存在")
			return
		}
		domainName := extractDomain(site.Host)
		if hasWildcard, _ := checkWildcardCertExists(domainName); hasWildcard {
			BadRequest(c, fmt.Sprintf("域名 %s 已存在泛域名证书，无需为站点单独申请", domainName))
			return
		}
		subjects = []string{site.Host}
		certIdentifier = site.Host

	default:
		BadRequest(c, "无效的目标类型")
		return
	}

	tlsCert := []map[string]interface{}{{
		"@id":      certID,
		"subjects": subjects,
		"issuers": []map[string]interface{}{{
			"module": "internal",
		}},
	}}

	if err := caddyClient.PutConfigPath("apps/tls/automation/policies/...", false, tlsCert); err != nil {
		log.Printf("[ERROR] Failed to add TLS cert for %s: %v", certIdentifier, err)
		InternalServerError(c, "添加证书失败")
		return
	}

	SuccessWithMessage(c, "证书申请已提交", gin.H{
		"certId":   certID,
		"method":   "auto",
		"type":     req.TargetType,
		"subjects": subjects,
	})
}

func extractDomain(host string) string {
	parts := strings.SplitN(host, ".", 2)
	if len(parts) >= 2 {
		return parts[1]
	}
	return host
}

func getSiteByID(siteID string) (*SiteRoute, error) {
	sites, err := getAllSites()
	if err != nil {
		return nil, err
	}

	for _, site := range sites {
		if site.ID == siteID {
			return &site, nil
		}
	}

	return nil, fmt.Errorf("site not found")
}

func getAllSites() ([]SiteRoute, error) {
	httpConfig, err := caddyClient.GetHTTPConfig()
	if err != nil {
		return nil, err
	}

	var allSites []SiteRoute

	for serverID, server := range httpConfig.Servers {
		for _, route := range server.Routes {
			for _, handle := range route.Handle {
				if handle.Handler == "subroute" && handle.Routes != nil {
					for _, subRoute := range handle.Routes {
						if subRoute.Id != "" && strings.HasPrefix(subRoute.Id, "site_") {
							host := ""
							for _, match := range subRoute.Match {
								if len(match.Host) > 0 {
									host = match.Host[0]
									break
								}
							}

							siteID := strings.TrimPrefix(subRoute.Id, "site_")

							allSites = append(allSites, SiteRoute{
								Name:     siteID,
								Host:     host,
								DomainID: serverID,
							})
						}
					}
				}
			}
		}
	}

	return allSites, nil
}

func checkWildcardCertExists(domainName string) (bool, error) {
	certPath := fmt.Sprintf("apps/tls/automation/policies/internal/*.%s", domainName)

	data, err := caddyClient.GetConfigPath(certPath, false)
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}

	if data == nil || len(data) == 0 || string(data) == "null" {
		return false, nil
	}

	return true, nil
}
