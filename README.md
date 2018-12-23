# srelapd
SRE LDAP Server

SuperLight weight LDAP Server.

## Configuration
Create a configuration .JSON file

```json
{
  "baseDN": "dc=tsocial,dc=com",
  "users": [
    {
      "name": "ldap10",
      "givenname": "Test",
      "sn": "Account",
      "mail": "test@trustingsocial.com",
      "unixid": 5002,
      "primarygroup": 5501,
      "loginShell": "/bin/bash",
      "homeDir": "/home/ldap10",
      "passsha256": "652c7dc687d98c9889304ed2e408c74b611e86a40caa51c4b43f1dd5913c5cd0",
    },
    {
      "name": "otpuser",
      "unixid": 5004,
      "primarygroup": 5501,
      "passsha256": "652c7dc687d98c9889304ed2e408c74b611e86a40caa51c4b43f1dd5913c5cd0",
      "otpsecret": "3hnvnk4ycv44glziedqfefw6s25j4dougs3rk"
    }
  ],
  "groups": [
    {
      "name": "sre",
      "unixid": 5501
    }
  ]
}
```

## Binary

```bash
$: make build
```

Will output a `dist/srelapd`

## Start

`./srelapd --config=/config.json`

## TLS

You can choose to keep a Bind password or use TLS for certs

```bash
usage: srelapd --config=CONFIG [<flags>]

Flags:
  --help                         Show context-sensitive help (also try --help-long and --help-man).
  --listen="9899"                Listening on port
  --skip-tls                     Skip using TLS while connecting to the server.
  --root-cert-file=ROOT-CERT-FILE  
                                 Root Cert File
  --cert-file=CERT-FILE          Cert File
  --key-file=KEY-FILE            Key File
  --upstream-addr="localhost:8080"  
                                 Upstream server address, the main server.
  --tls-server-name="localhost"  ServerName in tls Config
  --config=CONFIG                Config File
```
