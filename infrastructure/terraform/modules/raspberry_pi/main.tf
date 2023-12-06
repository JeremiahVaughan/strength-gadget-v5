locals {
  api_domain_name    = "api.${var.domain_name}"
}

data "cloudflare_zone" "this" {
  name = var.domain_name
}

resource "cloudflare_record" "this" {
  zone_id = data.cloudflare_zone.this.zone_id
  name    = local.api_domain_name
  value   = var.static_ip
  type    = "CNAME"
  proxied = true
}

# todo this might be doing more harm then good since it makes requests take longer should they need redirecting.
resource "cloudflare_page_rule" "http_to_https_redirect" {
  zone_id  = data.cloudflare_zone.this.zone_id
  target   = "http://*${var.domain_name}*"
  priority = 1
  status   = "active"

  actions {
    always_use_https = true
  }
}