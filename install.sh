#! /bin/bash

installSoftware() {
    apt -qq -y -t $(lsb_release -sc)-backports install golang-go
}

installFEH() {
    curl -Lo- https://github.com/sunshineplan/feh-go/archive/v1.0.tar.gz | tar zxC /etc
    mv /etc/feh-go* /etc/feh-go
    cd /etc/feh-go
    go build
}

configFEH() {
    read -p 'Please enter metadata server: ' server
    read -p 'Please enter VerifyHeader header: ' header
    read -p 'Please enter VerifyHeader value: ' value
    sed "s,\$server,$server," /etc/feh-go/config.ini.default > /etc/feh-go/config.ini
    sed -i "s/\$header/$header/" /etc/feh-go/config.ini
    sed -i "s/\$value/$value/" /etc/feh-go/config.ini
}

createCronTask() {
    cp -s /etc/feh-go/feh-go.cron /etc/cron.d/feh-go
    chmod 644 /etc/feh-go/feh-go.cron
}

main() {
    installSoftware
    installFEH
    configFEH
    createCronTask
}

main
