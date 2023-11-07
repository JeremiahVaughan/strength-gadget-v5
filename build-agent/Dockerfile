FROM ubuntu:23.04
ARG DEBIAN_FRONTEND=noninteractive

USER root
RUN cat /etc/resolv.conf


RUN apt-get update && apt-get install wget unzip zip git curl xz-utils awscli -y

## Install Docker ###
RUN apt-get install ca-certificates curl gnupg lsb-release -y
RUN mkdir -m 0755 -p /etc/apt/keyrings
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
RUN echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
RUN apt-get update
RUN apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin -y
# can't print the docker version here because docker needs to be mounted to the /var/run/docker.sock:/var/run/docker.sock directory first

### Install Terraform ###
# Get hashes from https://releases.hashicorp.com/terraform/1.3.8/terraform_1.3.8_SHA256SUMS
ENV EXPECTED_SHAW="9d9e7d6a9b41cef8b837af688441d4fbbd84b503d24061d078ad662441c70240"
ENV TERRAFORM_VERSION="1.3.8"
ENV FILE_NAME="terraform_${TERRAFORM_VERSION}_linux_amd64.zip"
RUN wget "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/${FILE_NAME}"
RUN echo "$EXPECTED_SHAW $FILE_NAME" | sha256sum --check
RUN unzip $FILE_NAME -d terraform
RUN mv terraform/terraform /usr/bin/
RUN terraform --version


### Install Terragrunt
# Get hashes from https://github.com/gruntwork-io/terragrunt/releases
ENV EXPECTED_SHAW="d9886c7de21fbe7127251735599f634eb7899f635e7cd73015a7b1a0a25772fc"
ENV TERRAGRUNT_VERSION="0.43.2"
ENV FILE_NAME="terragrunt_linux_amd64"
RUN wget "https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64"
RUN echo "$EXPECTED_SHAW $FILE_NAME" | sha256sum --check
RUN chmod 700 $FILE_NAME
RUN mv $FILE_NAME /usr/bin/terragrunt
RUN terragrunt --version


### Install Golang ###
# Get hashes from https://go.dev/dl/
ENV EXPECTED_SHAW="4eaea32f59cde4dc635fbc42161031d13e1c780b87097f4b4234cfce671f1768"
ENV GOLANG_VERSION="1.20.2"
ENV FILE_NAME="go${GOLANG_VERSION}.linux-amd64.tar.gz"
RUN wget "https://go.dev/dl/${FILE_NAME}"
RUN echo "$EXPECTED_SHAW $FILE_NAME" | sha256sum --check
RUN rm -rf /usr/local/go && tar -C /usr/local -xzf "go${GOLANG_VERSION}.linux-amd64.tar.gz"
ENV PATH=$PATH:/usr/local/go/bin
RUN go version


### Install Node ###
# Get hashes from https://nodejs.org/en/download/
ENV EXPECTED_SHAW="44d93d9b4627fe5ae343012d855491d62c7381b236c347f7666a7ad070f26548"
ENV NODE_VERSION="18.16.0"
ENV FILE_NAME="node-v${NODE_VERSION}-linux-x64.tar.xz"
ENV UNPACKED_DIR_NAME="node-v${NODE_VERSION}-linux-x64"
# https://nodejs.org/dist/v18.16.0/node-v18.16.0-linux-x64.tar.xz
RUN wget "https://nodejs.org/dist/v${NODE_VERSION}/${FILE_NAME}"
RUN echo "$EXPECTED_SHAW $FILE_NAME" | sha256sum --check
RUN tar -xf ${FILE_NAME}
ENV PATH=$PATH:"/${UNPACKED_DIR_NAME}/bin"
RUN node -v


# dependencies
RUN mkdir -p /workspace/deployer
WORKDIR /workspace/deployer
COPY go.mod .
COPY go.sum .
RUN go mod download && go mod verify

# Deployer build
WORKDIR /workspace/deployer
COPY . .
RUN go test ./...
RUN go build -o /usr/bin/deploy .








