#! /bin/bash

installSoftware() {
    apt -qq -y -t $(lsb_release -sc)-backports install golang-go
}

installFEH() {
    curl -Lo- https://github.com/sunshineplan/feh/archive/v1.0.tar.gz | tar zxC /etc
    mv /etc/feh* /etc/feh
    cd /etc/feh
    go build -ldflags "-s -w" -o feh
}

configFEH() {
    read -p 'Please enter metadata server: ' server
    read -p 'Please enter VerifyHeader header: ' header
    read -p 'Please enter VerifyHeader value: ' value
    sed "s,\$server,$server," /etc/feh/config.ini.default > /etc/feh/config.ini
    sed -i "s/\$header/$header/" /etc/feh/config.ini
    sed -i "s/\$value/$value/" /etc/feh/config.ini
}

createCronTask() {
    cp -s /etc/feh/scripts/feh.cron /etc/cron.d/feh
    chmod 644 /etc/feh/feh.cron
}

main() {
    installSoftware
    installFEH
    configFEH
    createCronTask
}

main
