#! /bin/bash

# Copyright 2015, Google, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# [START startup]
set -ex

# Talk to the metadata server to get the project id and location of application binary.
PROJECTID=$(curl -s "http://metadata.google.internal/computeMetadata/v1/project/project-id" -H "Metadata-Flavor: Google")
FARDO_DEPLOY_LOCATION="gs://go-server/fardo-beta.tar"

# Install logging monitor. The monitor will automatically pickup logs send to
# syslog.
# [START logging]
curl -s "https://storage.googleapis.com/signals-agents/logging/google-fluentd-install.sh" | bash
service google-fluentd restart &
# [END logging]

# Install dependencies from apt
apt-get update
apt-get install -yq ca-certificates supervisor

# Get the application tar from the GCS bucket.
gsutil cp $FARDO_DEPLOY_LOCATION /app.tar
if [ -d "/app" ]; then
rm -rf /app
else
useradd -m -d /home/goapp goapp
fi
mkdir /app
tar -x -f /app.tar -C /app
chmod +x /app/app

# Create a goapp user. The application will run as this user.

chown -R goapp:goapp /app


# Configure supervisor to run the Go app.
cat >/etc/supervisor/conf.d/goapp.conf << EOF
[program:goapp]
directory=/app
command=/app/app
autostart=true
autorestart=true
user=goapp
environment=HOME="/home/goapp",USER="goapp"
stdout_logfile=syslog
stderr_logfile=syslog
EOF

service supervisor stop
service supervisor start

# Application should now be running under supervisor
# [END startup]
