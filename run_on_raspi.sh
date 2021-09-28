#!/bin/bash
set -e

executable_name='raspi-dash'
raspi_executable_path="/usr/local/bin/$executable_name"

GOOS=linux GOARCH=arm GOARM=5 go build -o raspi-dash main.go
scp ./raspi-dash raspi:/tmp/
ssh raspi \
"sudo systemctl stop $executable_name.service; \
sudo mv /tmp/$executable_name $raspi_executable_path; \
sudo chmod a+x $raspi_executable_path; \
sudo systemctl start $executable_name.service"

