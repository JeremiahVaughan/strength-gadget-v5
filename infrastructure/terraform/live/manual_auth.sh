# sign in with:
eval "$(op signin)"

# copy paste into terminal:
export AWS_TF_STATE_ACCESS_KEY_ID
AWS_TF_STATE_ACCESS_KEY_ID=$(op item get AWS_TF_STATE_ACCESS_KEY_ID --fields password)
export AWS_TF_STATE_BUCKET_SECRET
AWS_TF_STATE_BUCKET_SECRET=$(op item get AWS_TF_STATE_BUCKET_SECRET --fields password)
export AWS_TF_STATE_BUCKET_REGION
AWS_TF_STATE_BUCKET_REGION=$(op item get AWS_TF_STATE_BUCKET_REGION --fields password)
export TERRAFORM_STATE_BUCKET_REGION
TERRAFORM_STATE_BUCKET_REGION=$(op item get AWS_TF_STATE_BUCKET_REGION --fields password)
export TF_VAR_aws_region
TF_VAR_aws_region=$(op item get AWS_WEB_HOST_REGION --fields password)
export TF_VAR_aws_access_key_id
TF_VAR_aws_access_key_id=$(op item get AWS_WEB_HOST_ACCESS_KEY_ID --fields password)
export TF_VAR_aws_secret_access_key
TF_VAR_aws_secret_access_key=$(op item get AWS_WEB_HOST_SECRET --fields password)
export CLOUDFLARE_API_TOKEN
CLOUDFLARE_API_TOKEN=$(op item get CLOUDFLARE_API_TOKEN --fields password)
export PUB_SSH_KEY_PATH
PUB_SSH_KEY_PATH="$HOME/.ssh/id_rsa.pub"
export TF_VAR_app_name
TF_VAR_app_name="strengthgadget"
export TF_VAR_build_number
TF_VAR_build_number="9"
export TF_VAR_database_connection_string
TF_VAR_database_connection_string="dummy"
export TF_VAR_database_root_ca
TF_VAR_database_root_ca="dummy"
export TF_VAR_domain_name
TF_VAR_domain_name="strengthgadget.com"
export TF_VAR_email_root_ca
TF_VAR_email_root_ca="dummy"
export TF_VAR_registration_email_from
TF_VAR_registration_email_from="dummy"
export TF_VAR_registration_email_from_password
TF_VAR_registration_email_from_password="dummy"

# init first in artifacts, then ecs, then cloudfront dir with:
terragrunt init -reconfigure -upgrade -backend-config="access_key=$AWS_TF_STATE_ACCESS_KEY_ID" -backend-config="secret_key=$AWS_TF_STATE_BUCKET_SECRET" -backend-config="region=$AWS_TF_STATE_BUCKET_REGION"

