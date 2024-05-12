#!/bin/bash
sed -i "s/\"VERSION_PLACEHOLDER\"/\"$CIRCLE_WORKFLOW_ID\"/" ./public/health.json
echo "NX_CLOUD_ACCESS_TOKEN=$TF_VAR_nx_cloud_access_token" > nx-cloud.env
npm i
npx nx build --configuration=production