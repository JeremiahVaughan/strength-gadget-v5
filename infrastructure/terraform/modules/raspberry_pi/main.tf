
data "cloudflare_zone" "this" {
  name = "strengthgadget.com"
}

resource "cloudflare_record" "this" {
  zone_id = data.cloudflare_zone.this.zone_id
  name    = var.domain_name
  value   = var.static_ip
  type    = "A"
  proxied = true
}

