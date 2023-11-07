#Bucket that stores our terraform state
resource "aws_s3_bucket" "terraform_state" {
  bucket = "strength-gadget-terraform-state"
  force_destroy = false
}

#Enable versioning so we can see a full history of our state files
resource "aws_s3_bucket_versioning" "terraform_state_versioning" {
  bucket = aws_s3_bucket.terraform_state.id
  versioning_configuration {
    status = "Enabled"
  }
}


#Enable server side encryption by default
resource "aws_s3_bucket_server_side_encryption_configuration" "terraform_state_encryption" {
  bucket = aws_s3_bucket.terraform_state.bucket

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.terraform_state_key.arn
      sse_algorithm = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

resource "aws_kms_key" "terraform_state_key" {
  description = "This key encrypts the terraform state for strength-gadget"
  deletion_window_in_days = 7
}


#Locking table to prevent concurrent updates to the state files
resource "aws_dynamodb_table" "terraform_locks" {
  name = "strength-gadget-terraform-state"
  billing_mode = "PAY_PER_REQUEST"
  hash_key = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }
}

