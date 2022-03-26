#!/usr/bin/env bash

set -e

basedir=$(cd $(dirname $0) && pwd)

if [ $EUID -ne 0 ]; then
    echo "Run with sudo"
    exit 1
fi

if [ "$1" == "" ]; then
    # cp $basedir/godaddyddns /usr/local/bin/godaddyddns
    curl https://raw.githubusercontent.com/navilg/godaddy-ddns/master/assets/godaddyddns -o /usr/local/bin/godaddyddns
else
    # cp $basedir/godaddyddns /usr/local/bin/godaddyddns
    curl https://raw.githubusercontent.com/navilg/godaddy-ddns/$1/assets/godaddyddns -o /usr/local/bin/godaddyddns
fi

if [ "$SUDO_USER" == "" ]; then
    installer_user=$USER
else
    installer_user=$SUDO_USER
fi

chmod +x /usr/local/bin/godaddyddns

cat <<EOF > $basedir/godaddy-ddns.service
[Unit]
Description=GoDaddy DDNS service
After=multi-user.target

[Service]
Type=simple
ExecStart=/usr/local/bin/godaddyddns daemon
User=$installer_user
Group=$installer_user

[Install]
WantedBy=multi-user.target
EOF

mv $basedir/godaddy-ddns.service /etc/systemd/system/godaddy-ddns.service
systemctl --system daemon-reload
systemctl start godaddy-ddns.service
systemctl enable godaddy-ddns.service

cat <<EOF > $basedir/godaddyddns-uninstall.sh
#!/usr/bin/env bash

set -e
if [ $EUID -ne 0 ]; then
    echo "Run with sudo"
    exit 1
fi

rm -rfv /usr/local/bin/godaddyddns

systemctl stop godaddy-ddns.service
systemctl disable godaddy-ddns.service
rm -rfv /etc/systemd/system/godaddy-ddns.service
rm -rfv /usr/local/bin/godaddyddns-uninstall.sh
echo "Uninstalled."

EOF

mv $basedir/godaddyddns-uninstall.sh /usr/local/bin/godaddyddns-uninstall.sh
chmod +x /usr/local/bin/godaddyddns-uninstall.sh
echo "Installation successful."
echo "For help Run, godaddyddns -h"