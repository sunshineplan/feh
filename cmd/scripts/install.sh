#! /bin/bash

installSoftware() {
    apt -qq -y install mongodb-org-tools
}

installFEH() {
    mkdir -p /etc/feh
    cd /etc/feh
    curl -LO https://github.com/sunshineplan/feh/releases/latest/download/feh
    curl -LO https://raw.githubusercontent.com/sunshineplan/feh/main/cmd/scripts/feh.cron
    curl -LO https://raw.githubusercontent.com/sunshineplan/feh/main/cmd/config.ini.default
    chmod +x feh
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
    cp -s /etc/feh/feh.cron /etc/cron.d/feh
    chmod 644 /etc/feh/feh.cron
}

main() {
    installSoftware
    installFEH
    configFEH
    createCronTask
}

main
