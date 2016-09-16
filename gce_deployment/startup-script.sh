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

# Start of Mongodb configuration

set -e

DEFAULT_UPTIME_DEADLINE="300"  # 5 minutes

metadata_value() {
  curl --retry 5 -sfH "Metadata-Flavor: Google" \
       "http://metadata/computeMetadata/v1/$1"
}

access_token() {
  metadata_value "instance/service-accounts/default/token" \
  | python -c "import sys, json; print json.load(sys.stdin)['access_token']"
}

uptime_seconds() {
  seconds="$(cat /proc/uptime | cut -d' ' -f1)"
  echo ${seconds%%.*}  # delete floating point.
}

config_url() { metadata_value "instance/attributes/status-config-url"; }
instance_id() { metadata_value "instance/id"; }
variable_path() { metadata_value "instance/attributes/status-variable-path"; }
project_name() { metadata_value "project/project-id"; }
uptime_deadline() {
  metadata_value "instance/attributes/status-uptime-deadline" \
      || echo $DEFAULT_UPTIME_DEADLINE
}

config_name() {
  python - $(config_url) <<EOF
import sys, urlparse
parsed = urlparse.urlparse(sys.argv[1])
print '/'.join(parsed.path.rstrip('/').split('/')[-4:])
EOF
}

variable_body() {
  encoded_value=$(echo "$2" | base64)
  printf '{"name":"%s", "value":"%s"}\n' "$1" "$encoded_value"
}

post_result() {
  var_subpath=$1
  var_value=$2
  var_path="$(config_name)/variables/$var_subpath/$(instance_id)"

  curl --retry 5 -sH "Authorization: Bearer $(access_token)" \
      -H "Content-Type: application/json" \
      -X POST -d "$(variable_body "$var_path" "$var_value")" \
      "$(config_url)/variables"
}

post_success() {
  post_result "$(variable_path)/success" "${1:-Success}"
}

post_failure() {
  post_result "$(variable_path)/failure" "${1:-Failure}"
}

# The contents of initScript are contained within this function.
custom_init() (
  return 0
)

# The contents of checkScript are contained within this function.
check_success() (
  failed=$(/etc/init.d/bitnami status \
      | grep "not running" | cut -d" " -f1 | tr "\n" " ")
  if [ ! -z "$failed" ]; then
    echo "Processes failed to start: $failed"
    exit 1
  fi
)

check_success_with_retries() {
  deadline="$(uptime_deadline)"
  while [ "$(uptime_seconds)" -lt "$deadline" ]; do
    message=$(check_success)
    case $? in
    0)
      # Success.
      return 0
      ;;
    1)
      # Not ready; continue loop
      ;;
    *)
      # Failure; abort.
      echo $message
      return 1
      ;;
    esac

    sleep 5
  done
}

do_init() {
  # Run the init script first. If no init script was specified, this
  # is a no-op.
  echo "software-status: initializing..."

  set +e
  message="$(custom_init)"
  result=$?
  set -e

  if [ $result -ne 0 ]; then
    echo "software-status: init failure"
    post_failure "$message"
    return 1
  fi
}

do_check() {
  # Poll for success.
  echo "software-status: waiting for software to become ready..."
  set +e
  message="$(check_success_with_retries)"
  result=$?
  set -e

  if [ $result -eq 0 ]; then
    echo "software-status: success"
    post_success
  else
    echo "software-status: failed with message: $message"
    post_failure "$message"
  fi
}

# Run the initialization script synchronously.
do_init || exit $?

# The actual software initialization might come after google's init.d
# script that executes our startup script. Thus, launch this script
# into the background so that it does not block init and eventually
# timeout while waiting for software to start.
do_check & disown

# End of Mongodb configuration
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
