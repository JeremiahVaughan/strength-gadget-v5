package model

type CloudflareIpsResponse struct {
	Result   CloudflareIpsResponseResult `json:"result"`
	Success  bool                        `json:"success"`
	Errors   []interface{}               `json:"errors"`
	Messages []interface{}               `json:"messages"`
}

type CloudflareIpsResponseResult struct {
	Ipv4Cidrs []string `json:"ipv4_cidrs"`
	Ipv6Cidrs []string `json:"ipv6_cidrs"`
	Etag      string   `json:"etag"`
}
