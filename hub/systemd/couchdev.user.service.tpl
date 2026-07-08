[Unit]
Description=Couchdev hub
After=network.target

[Service]
Type=simple
ExecStart=@PREFIX@/bin/couchdev serve --config @PREFIX@/etc/config.json
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
