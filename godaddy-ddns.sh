#!/bin/bash

current_version="vZ.2.1"

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
## */5 * * * * <Path-to-godaddy-ddns>/godaddy-ddns/godaddy-ddns.sh >/dev/null 2>&1  ##
## @reboot <Path-to-godaddy-ddns>/godaddy-ddns/godaddy-ddns.sh >/dev/null 2>&1  ##
function initialize()
{
    # Present directory of this file
    DIR=$(cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd)
    # Properties file to store value of domain, subdomain, keys and others
    properties_file="$DIR/godaddy-ddns.properties"
    logfile="$DIR/godaddy-ddns.log"

    # Source the values from properties file
    source $properties_file

    # API header Authorisation
    headers="Authorization: sso-key $key:$secret"
    return 0
}

function getDNSRecord()
{
    # Get the current data from GoDaddy DNS record
    result=$(curl -s -X GET -H "$headers" \
    "https://api.godaddy.com/v1/domains/$domain/records/A/$name")
    # checkconnection is non-zero if connection fails
    checkconnection=$?
    echo $result | grep -w "code" | grep -w "message"
    # checkerror is zero if any error message with error code is returned.
    checkerror=$?
    if [ $checkerror -ne 0 -a $checkconnection -eq 0 ]; then
        dnsIp=$(echo $result | grep -oE "\b([0-9]{1,3}\.){3}[0-9]{1,3}\b")
        existingTtl=$(echo $result | cut -d "," -f 3 | cut -d ":" -f 2)
        # If record is not created ttl value to initialise with a number
        [[ "$existingTtl" == "[]" ]] && existingTtl=0
        return 0
    fi
    return 1
}

function getPubIP()
{
    # Get public ip address there are several websites that can do this.
    ret=$(curl -s GET "http://ipinfo.io/json")
    checkret=$?
    if [[ $checkret -eq 0 ]]; then
        currentIp=$(echo "$ret" | grep -w -m 1 "\"ip\"" | sed -r -e 's/^ *//' -e 's/.*: *"//' -e 's/",?//')
        return 0
    fi
    return 1
}

function setDNSRecord()
{
    if [ "$dnsIp" != "$currentIp" -o $existingTtl -ne $ttl ];
    then
        # If IP and ttl value mismatch
       request='{"data":"'$currentIp'","ttl":'$ttl'}'
       nresult=$(curl -i -s -X PUT \
       -H "$headers" \
       -H "Content-Type: application/json" \
       -d [$request] "https://api.godaddy.com/v1/domains/$domain/records/A/$name")
       # Fetch out status from output and REMOVES \r character which is automatically getting suffixed in output of API
       stat=$(echo "$nresult" | grep -i http | awk '{first=$1;$1="";print $0;first;}')
       st=$(echo $stat|awk '{print $NF}'|sed 's/\r$//')
       if [[ "$st" == "OK" ]]
       then
           # If status of returned output is OK
           return 0
       else
           # If any error
           return 1
       fi
    else
        # Ips and ttl are equal
       return 0
   fi
}

function writeLog()
{
    if [[ $1 -eq 0 ]]; then
        echo "DNS Name: "$name.$domain"" > "$logfile"
        echo "DNS IP: $currentIp" >> "$logfile"
        echo "TTL: $ttl" >> "$logfile"
        echo "Status: OK" >> "$logfile"
        return 0
    elif [[ $1 -eq 100 ]]; then
        echo "DNS Name: "$name.$domain"" > "$logfile"
        echo "DNS IP: " >> "$logfile"
        echo "TTL: " >> "$logfile"
        echo "Status: Unknown - "$2"" >> "$logfile"
        return 100
    else
        echo "DNS Name: "$name.$domain"" > "$logfile"
        echo "DNS IP: " >> "$logfile"
        echo "TTL: " >> "$logfile"
        echo "Status: NOT OK - "$2"" >> "$logfile"
        return 1
    fi
}

function addCronJobs()
{
    crontab -l > $DIR/godaddy-ddns.cron 2>/dev/null
    grep -v "^#" $DIR/godaddy-ddns.cron | grep -i "$DIR/godaddy-ddns.sh" > /dev/null
    croncheck=$?
    if [[ $croncheck -ne 0 ]]; then
        echo "*/5 * * * * $DIR/godaddy-ddns.sh >/dev/null 2>&1" >> $DIR/godaddy-ddns.cron
        echo "@reboot $DIR/godaddy-ddns.sh >/dev/null 2>&1" >> $DIR/godaddy-ddns.cron
        crontab "$DIR/godaddy-ddns.cron"
    fi
    # Check new status of cron
    crontab -l > $DIR/godaddy-ddns.cron 2>/dev/null
    grep -v "^#" $DIR/godaddy-ddns.cron | grep -i "$DIR/godaddy-ddns.sh" > /dev/null
    croncheck=$?
    if [[ $croncheck -ne 0 ]]; then
        # If cron task not created succesfully
        rm -f $DIR/godaddy-ddns.cron
        return 1
    fi
    rm -f $DIR/godaddy-ddns.cron
    return 0
}

function version()
{
    echo "GoDaddy-DDNS $current_version"
    exit 0
}

# Check version

if [ "$1" == "version" ]; then
    version
fi

# Initialising Variables
getDNSRecordStatus=1000
getPubIPStatus=1000
setDNSRecordStatus=1000
addCronJobsStatus=1000

# Main - Function call
initialize

# Get DNS Record
getDNSRecord
getDNSRecordStatus=$?

# Get Public IPv4 address of server/instance
if [[ $getDNSRecordStatus -eq 0 ]]; then
    getPubIP
    getPubIPStatus=$?
fi

# Set DNS record
if [ $getDNSRecordStatus -eq 0 -a $getPubIPStatus -eq 0 ]; then
    setDNSRecord
    setDNSRecordStatus=$?
fi

# Write Log
if [[ $checkconnection -ne 0 ]]; then
    writeLog 100 "Connectivity Failed"
elif [[ $checkerror -eq 0 ]]; then
    writeLog 100 "$result"
elif [[ $getPubIPStatus -ne 0 ]]; then
    writeLog 100 "Connection Failed"
elif [[ $setDNSRecordStatus -eq 1 ]]; then
    writeLog 1 "$st"
elif [[ $setDNSRecordStatus -eq 0 ]]; then
    writeLog 0
else
    writeLog 100 "Unknown Error: Run command: bash -x godaddy-ddns.sh and report us with output after OMMITING key and secret."
fi

# Add cron job
if [[ $setDNSRecordStatus -eq 0 ]]; then
    addCronJobs
    addCronJobsStatus=$?
fi
