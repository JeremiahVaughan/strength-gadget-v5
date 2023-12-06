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
  type    = "A"
  proxied = true
}

