# README #


**What is this repository for?**

* Public IP address of many hosting servers are dynamic and it changes based on availability. Some examples include AWS EC2 instance, Home Servers, GCP Cloud, etc.
* This require updating the IP address in GoDaddy DNS record so that requests are forwarded to correct IP address.
* This script uses cron jobs to check the current Public IP address of server and update the IP address in GoDaddy Manage DNS.
* OS: Linux. Tested on Ubuntu server 18.04 LTS with BASH shell.
* Required utility tools: curl, BASH shell.

**Features**

* Automatically updates Public IPv4 address in GoDaddy A Record
* Minimal and easy setup.
* Minimal dependency on extra tools.
* Lightweight
* Minimal resource requirement

**How do I get set up ?**

* Clone this repository in your server.
* Open godaddyDDNS/godaddy-ddns.properties file
* Update the properties values as below:

domain=domain.com

name=subdomainORwww

ttl=ttlValueInSeconds

key=key-value-from-godaddy-developer-console

secret=secret-key-value-from-godaddy-developer-console


e.g. To update DNS Name *navi.example.com*
```
domain=example.com

name=navi

key=cvvgfvfd54jghgz8s

ttl=3600

secret=dfgsx6daflx5]gkhhi8yjxf
```
* Key and Secret can be generated from GoDaddy Developer page. https://developer.godaddy.com/getstarted
* Make sure record is **'A' type** in godaddy records list. If record is created with some other type, it will fail with error. 
* If record is not already created, Running this script will create it.
* After updating properties file. Run the script.
*./godaddy-ddns.sh*
* Check the log in godaddy-ddns.log file
* If log status is OK. Its working fine. Verify the DNS record in GoDaddy account.
* **Updating the DNS record in GoDaddy doesn't mean DNS server is also updated. DNS server updation completely depends on TTL value. Its good to have short TTL value for highly dynamic IP address.**
* Create a crontab entry to run the godaddy-ddns.sh file every five minutes.

```
crontab -e
```
Add below lines

```
*/5 * * * * /Path/to/godaddy-ddns/godaddy-ddns.sh >/dev/null 2>&1

@reboot /Path/to/godaddy-ddns/godaddy-ddns.sh >/dev/null 2>&1
```

Save it

* 'How to' Youtube video --> https://youtu.be/lnPPdYexf4E

**NOTES:**
GoDaddy supports 60 requests per minutes. This scripts uses 2 requests. So, Make sure to create cronjob accordingly so that request doesn't exceed 60 request per minute or else it will fail.


**Contribution guidelines**

* You can create issue here --> https://github.com/navilg/godaddy-ddns/issues
* Your pull requests are invited for review and merge --> https://github.com/navilg/godaddy-ddns/pulls

**How to reach me ?**

* Repo owner or admin: [Navratan Gupta](mailto:navilg0409@gmail.com)
* You can reach out to me on navilg0409@gmail.com for any feedback.

**Special mention to https://github.com/markafox/GoDaddy_Powershell_DDNS which is base for this script.**
