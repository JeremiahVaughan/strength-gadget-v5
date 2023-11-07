### Environment Structure ###

- Each environment gets its own AWS account for these two reasons:
    - Easier to understand cost
    - Safety in that each terraform apply is not able to touch another environment
- AWS accounts:
    - piegarden@gmail.com is STAGING
    - jeremiah.t.vaughan.com is PRODUCTION
- Cloudflare accounts:
    - piegarden@gmail.com is STAGING
    - jeremiah.t.vaughan.com is PRODUCTION
- Auth0
    - piegarden@gmail.com is STAGING and PRODUCTION (different tenants are used for both environments)

### Project Dependencies ###

- Terraform (https://learn.hashicorp.com/tutorials/terraform/install-cli)
- Terragrunt (https://terragrunt.gruntwork.io/docs/getting-started/install/)

### Perquisite Steps ### 

- Create IAM access user/group/keys for AWS
- Create a connection manually to GitHub called github-connection with GitHub app: 26455730. According
  to https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/codestarconnections_connection,
  terraform can create a connection, but it cannot automate the authentication process. So destroying it would cause me
  to have to reauthenticate every time I spin up staging, so leaving this as a manual step for now.
- Set up terraform state in all desired environments via this module:
  `./aws/terraform-prerequisites`
- Note: These state resources must be maintained manually after creation since we would need to create another state
  backend resource to manage these properly.
    - Delete state files (ONLY for this module) in between executions to avoid errors.
      `terraform.tfstate`

### Operation Steps ###

- Setup environments per https://github.com/JeremiahVaughan/strength-gadget-environment-swapper
- Place configs in environment-configs folder
- Create state backends as needed via aws/terraform-prerequisites
- Switch to desired environment. See make file in same directory
- From project root run (the option --terragrunt-source allows us to redirect to the local modules folder)
  `terragrunt run-all init`
  `terragrunt run-all plan`
  `terragrunt run-all apply`
  `terragrunt run-all destroy`

### Error ###

- Error acquiring the state lock (if the following command doesn't work you will need to go into the table and delete
  the locking record manually)
  `terragrunt force-unlock -force 57b8a892-377b-8f4a-75ff-0556a2dafcd1` -- ID of lock info record

### Nuke AWS Resources ###
#### Using
https://github.com/gruntwork-io/cloud-nuke

#### How
- Switch to AWS CLI profile named staging: `export AWS_PROFILE=staging`
- Run this command `cloud-nuke aws` or specific regions (faster) use `cloud-nuke aws --region us-east-1`
- Check these resources if they did not delete, had issues with them not deleting in the past
  - security groups
  - iam roles

### Swapping Environments

Ensure AWS CLI is installed

Ensure to emplace your ~/.aws folder

From the terraform-live directory run one of the following to switch to your desired environment:
- `make swap-to-dev`
- `make swap-to-staging`
- `make swap-to-production`

### Verify you are Switched to the new Environment
- In zsh you should see an AWS profile indicator at the far left of your current line in the terminal window

### Note
- The current_config.yaml file is a link to the original file. Meaning changes to the original file can be seen immediately in the current_config.yaml.

