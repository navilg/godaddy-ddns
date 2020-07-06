#!/bin/bash

# This script is used to check and update your GoDaddy DNS server to the IP address of your current internet connection.
# Special thanks to mfox for his ps script
# https://github.com/markafox/GoDaddy_Powershell_DDNS
#
# First go to GoDaddy developer site to create a developer account and get your key and secret
#
# https://developer.godaddy.com/getstarted
# Be aware that there are 2 types of key and secret - one for the test server and one for the production server
# Get a key and secret for the production server
#
#Update the first 4 variables with your information


# Add below two lines in crontab entry.
## */5 * * * * <Path-to-godaddyDDNS>/godaddyDDNS/godaddyDDNS.sh >/dev/null 2>&1  ##
## @reboot <Path-to-godaddyDDNS>/godaddyDDNS/godaddyDDNS.sh >/dev/null 2>&1  ##

# Present directory of this file
DIR=$(cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd)
# Properties file to store value of domain, subdomain, keys and others
properties_file="$DIR/godaddyDDNS.properties"

# Source the values from properties file
source $properties_file

# API header Authorisation
headers="Authorization: sso-key $key:$secret"

# Get the current data from GoDaddy DNS record
result=$(curl -s -X GET -H "$headers" \
 "https://api.godaddy.com/v1/domains/$domain/records/A/$name")

dnsIp=$(echo $result | grep -oE "\b([0-9]{1,3}\.){3}[0-9]{1,3}\b")
existingTtl=$(echo $result | cut -d "," -f 3 | cut -d ":" -f 2)
# If record is not created ttl value to initialise with a number
[[ "$existingTtl" == "[]" ]] && existingTtl=0

# Get public ip address there are several websites that can do this.
ret=$(curl -s GET "http://ipinfo.io/json")
currentIp=$(echo $ret | grep -oE "\b([0-9]{1,3}\.){3}[0-9]{1,3}\b")

 if [ "$dnsIp" != $currentIp -o $existingTtl -ne $ttl ];
 then
     # If IP and ttl value mismatch
	request='{"data":"'$currentIp'","ttl":'$ttl'}'
	nresult=$(curl -i -s -X PUT \
 	-H "$headers" \
 	-H "Content-Type: application/json" \
 	-d [$request] "https://api.godaddy.com/v1/domains/$domain/records/A/$name")
    # Fetch out status from output and REMOVES \r character which is automatically getting suffixed in output of API
	result=$(echo "$nresult" | grep -i http | awk '{first=$1;$1="";print $0;first;}')
	res=$(echo $result|awk '{print $NF}'|sed 's/\r$//')
	if [[ "$res" == "OK" ]]
	then
        # If status of returned output is OK
   		echo "DNS Name: "$name.$domain"" > $DIR/godaddyDDNS.log
		echo "DNS IP: $currentIp" >> $DIR/godaddyDDNS.log
	    echo "Status: OK" >> $DIR/godaddyDDNS.log
	else
        # If any error
		echo "DNS Name: "$name.$domain"" > $DIR/godaddyDDNS.log
		echo "DNS IP: $currentIp" >> $DIR/godaddyDDNS.log
	    echo "Status: NOT OK - $result" >> $DIR/godaddyDDNS.log
	fi
 else
        # Ips and ttl are equal
	echo "DNS Name: "$name.$domain"" > $DIR/godaddyDDNS.log
	echo "DNS IP: $currentIp" >> $DIR/godaddyDDNS.log
    echo "Status: OK" >> $DIR/godaddyDDNS.log
fi
