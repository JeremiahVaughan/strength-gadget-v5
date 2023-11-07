#!/bin/bash
sudo yum update
sudo yum install amazon-cloudwatch-agent -y
sudo systemctl enable amazon-cloudwatch-agent
cat > /opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json <<'EOF'
${config}
EOF
sudo systemctl start amazon-cloudwatch-agent
