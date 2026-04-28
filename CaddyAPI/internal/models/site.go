package models

// Site 代表一个站点配置，对应 Caddy 中的一个 server
type Site struct {
	Name       string   `json:"name"`        // 站点标识名称，如 "exp"
	Domain     string   `json:"domain"`      // 基础域名，如 "example.com"
	Wildcard   string   `json:"wildcard"`    // 泛域名，如 "*.example.com"
	Listen     []string `json:"listen"`      // 监听端口，如 [":443"]
	TLSEnabled bool     `json:"tls_enabled"` // 是否启用 TLS
	Routes     []Route  `json:"routes"`      // 子路由列表
}

// Route 代表一个子路由，对应 subroute 中的一个子站点
type Route struct {
	Name        string `json:"name"`         // 子路由名称，如 "first"
	Host        string `json:"host"`         // 完整域名，如 "first.example.com"
	Type        string `json:"type"`         // 类型：static, reverse_proxy
	Upstream    string `json:"upstream"`     // 上游服务器（反向代理）
	Root        string `json:"root"`         // 静态文件目录
	IndexNames  string `json:"index_names"`      // 默认文档，多个用空格分隔
	HealthCheck bool   `json:"health_check"` // 健康检查
}
