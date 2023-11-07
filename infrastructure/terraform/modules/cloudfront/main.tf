locals {
  domain_name = var.domain_name
  ui_dir      = "${var.workspace_dir}/ui-react-2/dist/ui-react-2"
  mime_types = {
    "html" = "text/html",
    "css"  = "text/css",
    "js"   = "application/javascript",
    "json" = "application/json",
    "png"  = "image/png",
    "jpg"  = "image/jpeg",
    "gif"  = "image/gif",
    "svg"  = "image/svg+xml"
    "ico"  = "image/x-icon"
    "zip"  = "application/zip"
  }
}

data "cloudflare_zone" "dns_zone" {
  name       = local.domain_name
  account_id = "d9c82ad1adf99452890374a6bf5a879a"
}


resource "aws_acm_certificate" "strength_gadget" {
  domain_name       = local.domain_name
  validation_method = "DNS"
}

resource "cloudflare_record" "strength_gadget" {
  zone_id = data.cloudflare_zone.dns_zone.id
  name    = "@"
  type    = "CNAME"
  value   = aws_cloudfront_distribution.strength_gadget_distribution.domain_name
  ttl     = 300
}


resource "cloudflare_record" "cloudflare_cert_validation_record" {
  for_each = {
    for dvo in aws_acm_certificate.strength_gadget.domain_validation_options : dvo.domain_name => {
      name  = dvo.resource_record_name
      value = dvo.resource_record_value
      type  = dvo.resource_record_type
    }
  }
  zone_id = data.cloudflare_zone.dns_zone.id
  name    = each.value.name
  value   = each.value.value
  type    = each.value.type
  ttl     = 60
}

resource "aws_acm_certificate_validation" "strength_gadget_aws_acm_certificate_validation" {
  certificate_arn         = aws_acm_certificate.strength_gadget.arn
  validation_record_fqdns = [for v in cloudflare_record.cloudflare_cert_validation_record : v.name]
}

resource "aws_cloudfront_distribution" "strength_gadget_distribution" {
  aliases = [local.domain_name]
  origin {
    domain_name              = aws_s3_bucket.static_html_bucket.bucket_regional_domain_name
    origin_access_control_id = aws_cloudfront_origin_access_control.strength_gadget.id
    origin_id                = aws_s3_bucket.static_html_bucket.bucket

    connection_attempts = 3
    connection_timeout  = 10

    #   todo find out if origin shield has a free tier

    #    origin_shield {
    #      enabled              = true
    #      origin_shield_region = "*********"
    #    }
  }

  enabled             = true
  is_ipv6_enabled     = true
  default_root_object = "index.html"

  custom_error_response {
#    todo look into tweaking this caching rule to see if it makes sense to do so
    error_caching_min_ttl = 0

    error_code            = 404
    response_code         = 200
    response_page_path    = "/index.html"
  }

  # Define the behavior for how CloudFront should handle requests for HTML assets
  default_cache_behavior {
    target_origin_id = aws_s3_bucket.static_html_bucket.bucket

    viewer_protocol_policy = "redirect-to-https"

    allowed_methods = ["GET", "HEAD", "OPTIONS"]
    cached_methods  = ["GET", "HEAD"]

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }
  }

  #  todo figure out logging
  #   Define your logging options
  logging_config {
    bucket          = aws_s3_bucket.cloudfront_logging_bucket.bucket_domain_name
    include_cookies = false
    prefix          = "cloudfront-logs/"
  }

  # Define your distribution options
  price_class = "PriceClass_All"

  # Define your default SSL certificate
  viewer_certificate {
    acm_certificate_arn = aws_acm_certificate.strength_gadget.arn
    ssl_support_method  = "sni-only"
  }
  restrictions {
    geo_restriction {
      locations        = []
      restriction_type = "none"
    }
  }
}

resource "aws_s3_bucket" "static_html_bucket" {
  bucket = local.domain_name
}

resource "aws_s3_bucket" "cloudfront_logging_bucket" {
  bucket = "cloudfront-strengthgadget-logging-${var.env}"
}


resource "aws_s3_bucket_acl" "origin_bucket_acl" {
  bucket = aws_s3_bucket.static_html_bucket.id
  acl    = "public-read"
}

resource "aws_s3_bucket_public_access_block" "origin_bucket_public_access_block" {
  bucket = aws_s3_bucket.static_html_bucket.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_policy" "origin_bucket_policy" {
  bucket = aws_s3_bucket.static_html_bucket.id

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": {
        "Sid": "AllowCloudFrontServicePrincipalReadOnly",
        "Effect": "Allow",
        "Principal": {
            "Service": "cloudfront.amazonaws.com"
        },
        "Action": "s3:GetObject",
        "Resource": "arn:aws:s3:::${aws_s3_bucket.static_html_bucket.bucket}/*",
        "Condition": {
            "StringEquals": {
                "AWS:SourceArn": "${aws_cloudfront_distribution.strength_gadget_distribution.arn}"
            }
        }
    }
}
EOF
}

# todo I can still access the bucket with strengthgadget.com.s3.amazonaws.com so not really sure how this is supposed to work
resource "aws_cloudfront_origin_access_control" "strength_gadget" {
  name                              = "strength-gadget"
  description                       = "Cloud front origin access control for Strength Gadget"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}


# todo do the same thing for lambdas, maybe it will fix the hashing issue where the lambdas are deploying every time despite no changes
resource "aws_s3_object" "html_assets" {
  for_each = fileset(local.ui_dir, "**")

  bucket       = aws_s3_bucket.static_html_bucket.bucket
  key          = each.key
  source       = "${local.ui_dir}/${each.value}"
  content_type = local.mime_types[split(".", each.value)[1]]
  etag = filemd5("${local.ui_dir}/${each.value}")
}

output "ui_artifact_path" {
  value = local.ui_dir
}
