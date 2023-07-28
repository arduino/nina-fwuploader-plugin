#!/usr/bin/bash

for profile in mkrwifi1010 nano_33_iot uno2018 nanorp2040connect
do
    arduino-cli compile -e --profile $profile
done

