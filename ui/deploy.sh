#!/bin/bash
sed -i "s/\"VERSION_PLACEHOLDER\"/\"$CIRCLE_WORKFLOW_ID\"/" ./public/health.json
npm i
npx nx build --configuration=production