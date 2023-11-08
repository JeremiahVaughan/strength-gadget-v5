# Development Getting Started
### Prerequisite
1. Install mkcert<br>
```sudo apt update && sudo apt install mkcert```
2. Have mkcert install to local trust store automatically<br>
```mkcert -install```
3. Generate a CA for local host:<br>
```mkcert localhost```
4. Update the project.json file with the location of your cert (localhost.pem) and key (localhost-key.pem) that was created in the previous step
5. Copy the cert (localhost.pem) and key (localhost-key.pem) to<br>
```./backend/strengthgadget/kind```
6. Copy the cert (localhost.pem) and key (localhost-key.pem) to<br>
```./ui```

### Debugging Dockerfiles
Execute docker commands with verbose output. For example: 
`BUILDKIT_PROGRESS=plain docker-compose -f docker-compose.debug.yml build`

# Speed Considerations
### Keep the docker context small
This ensures only what is needed gets sent to the docker daemon.
1. Ensure the context root only has what is needed. For example, monorepos will need a context big enough to cover the project and all its libraries.
2. Use .dockerignore to exclude unneeded directories. For example, node_modules isn't used by docker in our case, just local development. So its waste to copy it, since it is several GB in size.
### .dockerignore
Only applys if it is in the root of the docker context. That is why there is another .dockerignore in the ui directory.

### CICD
Any directory that contains a deployment_config.json file must have a unique name within the mono-repo. If not then unexpected deployment behavior will happen.

# Agent
Commit to the staging repo:
https://github.com/JeremiahVaughan/agents

# Jenkins
~/strength-gadget-v3/infrastructure/jenkins/docker-compose.yml

Install docker plugin:
Docker Pipeline (Version 563.vd5d2e5c4007f)

Create new "Freestyle Project"

Create build user
https://www.middlewareinventory.com/blog/jenkins-remote-build-trigger-url/
Note: you have to be logged into the user to create an api token for it



### Source code
In the "Source Code Management" section

Enter this into the repository URL
https://<username>:<personal access token>@git.jetbrains.space/strengthgadget/strengthgadget/strength-gadget-v4.git
https://jeremiah.t.vaughan:<personal access token>@git.jetbrains.space/strengthgadget/strengthgadget/strength-gadget-v4.git

### Env Vars
In build "Build Environment" section check the "Use secret text or file" box and select bindings of "Secret Text"

### Email notifications
Go to "Manage Jenkins" > "Configure System" and locate the "Extended E-mail Notification" section. If you can't find this section, you may need to install the "Email Extension Plugin" from "Manage Plugins."

Fill in the following fields:

- SMTP server: Enter your email server's SMTP address.
- Default user E-mail suffix: If your email addresses have a common domain, you can enter it here.
- Use SMTP Authentication: Check this box if your SMTP server requires authentication. Provide your email and APP PASSWORD (not regular password) for the SMTP server.
- Use SSL / TLS: Check this box if your email server requires a secure connection.
- SMTP Port: Enter the appropriate port number for your email server (e.g., 465 for SSL or 587 for TLS).
- Charset: Set it to UTF-8.
- Default Content Type: Choose "HTML" or "Plain Text" based on your preference.
- Default Recipients: Enter the email addresses that should receive notifications by default.


# Hardened Proxy (currently using cloudflare Zero Trust)
Mtls, ratelimiting, and whitelisting would probably be a good start
https://jeremiah.t.vaughan:akw7BDV1vft1nej*gyt@git.jetbrains.space/strengthgadget/strengthgadget/strength-gadget-v4.git


[//]: # (todo setup proxy so I can encrypt traffic to the jenkins instance.)

[//]: # (todo setup jenkins via script: http://localhost:8080/manage/cli/)
