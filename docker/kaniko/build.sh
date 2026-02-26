
cat - <<EOF
FROM nginx:latest
EOF >> /workspace/Dockerfile

/kaniko/executor --context=/workspace \
--dockerfile=/workspace/Dockerfile --force --skip-tls-verify --cache=true --cache-dir=/tmp \
--snapshotMode=redo --destination=$PUSH_IMAGE 

