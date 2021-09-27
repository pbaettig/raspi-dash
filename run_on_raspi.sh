#!/bin/bash
set -e

executable_name='raspi-dash'
raspi_executable_path="/tmp/$executable_name"

GOOS=linux GOARCH=arm GOARM=5 go build -o raspi-dash main.go
scp ./raspi-dash raspi:/tmp/
ssh raspi "chmod a+x $raspi_executable_path; sudo $raspi_executable_path"
ssh raspi "sudo pkill $executable_name"

rm "$executable_name"