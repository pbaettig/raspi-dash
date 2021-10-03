#!/bin/bash
set -e

executable_name='raspi-dash'
raspi_executable_path="/usr/local/bin/$executable_name"

GOOS=linux GOARCH=arm GOARM=5 go build -o raspi-dash main.go
scp ./raspi-dash raspi:/tmp/
scp ./raspi-dash.service raspi:/tmp/
ssh raspi \
"echo RASPI_DASH_USER_PASCAL=123456 | sudo tee /etc/default/$executable_name; \
sudo mv /tmp/raspi-dash.service /etc/systemd/system/raspi-dash.service; \
sudo systemctl daemon-reload; \
sudo systemctl stop $executable_name.service; \
sudo mv /tmp/$executable_name $raspi_executable_path; \
sudo chmod a+x $raspi_executable_path; \
sudo systemctl start $executable_name.service
"
# sudo systemctl start $executable_name.service
