# README #

**What is this repository for?**

* Public IP address of some servers or self-hosted server in homelab is dynamic and it keep changing over time. For e.g. AWS EC2 instance (Without Elastic IP), Home Servers, GCP Cloud, etc.
* This makes it difficult to set DNS for the server and require updating the IP address in GoDaddy DNS record so that requests are forwarded to correct IP address.
* This tool and docker container will automatically update the DNS record with new IP address whenever IP address of server changes.

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

Docker repository: https://hub.docker.com/r/linuxshots/godaddy-ddns

```
docker pull linuxshots/godaddy-ddns:latest
```

* Run a container using below command. Replace the values with correct value.

For DNS *myserver.example.com* with ttl 1200 seconds

```
docker run --name myserver.example.com -d --restart unless-stopped --env GD_NAME=myserver --env GD_DOMAIN=example.com --env GD_TTL=1200 --env GD_KEY=key-value-from-godaddy-developer-console --env GD_SECRET=secret-key-value-from-godaddy-developer-console linuxshots/godaddy-ddns:latest
```

* Check the log.

```
docker logs myserver.example.com
```

* And, You are done.

* Make sure to use --restart option while running docker run. This make sure container starts automatically when your machine boots up.

* To add another dns for same machine (Max 5 recommended due to rate limitations), Run the same docker run command with another container name and values.

**Non-Docker way**

* Required tools: curl and BASH. Root/Sudo access required

* Download installer script.

```
curl -LO https://raw.githubusercontent.com/navilg/godaddy-ddns/master/assets/install.sh
```

* Run the script with sudo.

```
sudo bash install.sh
```

To install specific version. E.g. Version v1.0.0

```
sudo bash install.sh v1.0.0
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

**GoDaddy DDNS usage (For Non-docker way)**

* Print help message

```
godaddyddns -h
```

* Add a record

```
godaddyddns add --domain='example.com' --name='myserver' --ttl=1200 --key='kEyGeneratedFr0mG0DaddY' --secret='s3cRe7GeneratedFr0mG0DaddY'
```

Output:

```
INFO 2022/03/26 19:34:00 myserver.example.com Record created/updated (ttl: 1200, ip: 222.48.150.132, key: ****, secret: ****)
```

* List all configured records

```
godaddyddns list
```

Output:

```
+---+-----------+-------------+------+
| # | NAME      | DOMAIN      |  TTL |
+---+-----------+-------------+------+
| 1 | myserver  | example.com | 1200 |
| 2 | myserver2 | example.com | 1300 |
+---+-----------+-------------+------+
```

* Update existing configured record

```
godaddyddns update --domain='example.com' --name='myserver' --ttl=1300 --key='kEyGeneratedFr0mG0DaddY' --secret='s3cRe7GeneratedFr0mG0DaddY'
```

Output:

```
INFO 2022/03/26 19:34:27 myserver.example.com Record created/updated (ttl: 1300, ip: 222.48.150.132, key: ****, secret: ****)
```

* Delete a record from configuration

```
godaddyddns delete --domain='example.com' --name='myserver'
```

Output:

```
INFO 2022/03/26 19:35:34 myserver.example.com Record removed from configuration. If not in use, delete the record manually from GoDaddy console.
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
