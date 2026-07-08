[Unit]
Description=Couchdev hub
After=network.target

[Service]
Type=simple
User=couchdev
Group=couchdev
ExecStart=@PREFIX@/bin/couchdev serve --config @PREFIX@/etc/config.json
Restart=on-failure
RestartSec=5
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes
ReadWritePaths=@PREFIX@/data

[Install]
WantedBy=multi-user.target
