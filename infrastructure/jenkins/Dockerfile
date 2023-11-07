FROM jenkins/jenkins:lts

USER root

# Update the package index and install the necessary dependencies
RUN apt-get update && \
    apt-get install \
        ca-certificates \
        curl \
        gnupg \
        lsb-release -y

# Add the Docker GPG key
RUN mkdir -m 0755 -p /etc/apt/keyrings
RUN curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg

RUN echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
      $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

RUN apt-get update

RUN apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin -y

ENTRYPOINT ["/usr/local/bin/jenkins.sh"]

# todo consider using the agent dockerfile to do the build work instead of pulling it. This would allow us to leverage docker cache. I'm not sure though if we lose our multithreaded capabilities with golang though.
