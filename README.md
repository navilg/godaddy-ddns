# README #

**What is this repository for?**

* Public IP address of many hosting servers or self-hosted server in homelab is dynamic and it changes based on availability. For e.g. AWS EC2 instance (Without Elastic IP), Home Servers, GCP Cloud, etc.
* This require updating the IP address in GoDaddy DNS record so that requests are forwarded to correct IP address.
* This script uses cron jobs to check the current Public IP address of server and update the IP address in GoDaddy Manage DNS.
* OS: Linux. Tested on Ubuntu server 20.04 LTS with BASH shell.

**Features**

* Automatically updates Public IPv4 address in GoDaddy A Record
* Minimal and easy setup.
* Minimal dependency on extra tools.
* Lightweight
* Minimal resource requirement

## How to setup

**Docker Way (Recommended)**

* Required tools: Docker

* If you have docker installed on your machine. Proceed with this step.

* Pull image

```
docker pull linuxshots/godaddy-ddns:latest
```

* Run a container using below command. Replace the values with correct value.

For DNS *myserver.example.com* with ttl 1200 seconds

```
docker run --name myserver.example.com --restart --env GD_NAME=myserver --env GD_DOMAIN=example.com --env GD_TTL=1200 --env GD_KEY=key-value-from-godaddy-developer-console --env GD_SECRET=secret-key-value-from-godaddy-developer-console linuxshots/godaddy-ddns:latest
```

* Check the log.

```
docker logs myserver.example.com
```

* And, You are done.

* Make sure to use --restart option while running docker run. This make sure container starts automatically when your machine boots up.

* To add another dns in same machine (Max 5 recommended due to rate limitations), Run the same docker run command with different container name and other values.

**Non-Docker way**

* Required utility tools: curl and BASH. Root/Sudo access required

* Download installer script.

```
curl -LO https://raw.githubusercontent.com/navilg/godaddy-ddns/master/godaddyddns-go/assets/install.sh
```

* Run the script with sudo.

```
sudo bash install.sh
```

To install specific version. E.g. Version 1.0.0

```
sudo bash install.sh 1.0.0
```

* Check status of service

```
sudo systemctl status godaddy-ddns.service
```

* Add a dns record

For DNS *myserver.example.com* with ttl 1200 seconds

```
godaddyddns add --domain='example.com' --name='myserver' --ttl=1200 --key='kEyGeneratedFr0mG0DaddY' --secret='s3cRe7GeneratedFr0mG0DaddY'
```

* Check logs

```
tail -200f $HOME/.config/godaddy-ddns/log/godaddy-ddns.log
```

* To uninstall

```
sudo godaddyddns-uninstall.sh
```

**NOTES:**

* Version Z.\*.\* is no more supported. Please use version 1.0.0+
* Key and Secret can be generated from GoDaddy Developer page. <https://developer.godaddy.com/getstarted>
* **DNS server updation depends on TTL value. Its good to have short TTL value for highly dynamic IP address. After DNS record is updated with new IP address it may take TTL time to update DNS lookup cache servers.**


## Contribution guidelines

* You can create issues here --> <https://github.com/navilg/godaddy-ddns/issues>
* Pull requests are welcomed for review and merge --> <https://github.com/navilg/godaddy-ddns/pulls>

**How to reach me ?**

* Author: [Navratan Lal Gupta](mailto:navilg0409@gmail.com)
* You can reach out to me on navilg0409@gmail.com for any feedback.
