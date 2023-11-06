# cli-password-autocomplete

Allow you ssh to your servers using your bitwarden password automatically.

### Preqrequisites:
- Installed [Bitwarden CLI tool](https://bitwarden.com/help/cli/#download-and-install)
- Installed sshpass tool
- BITWARDEN_MASTER_PASSWORD="your master password" in OS env
- [bw cli tool](https://bitwarden.com/help/cli/) configured to use your server (bw config server "your bitwarden server")
- [bw cli tool](https://bitwarden.com/help/cli/) login runned (bw login "your-email")
- ssh users that you want to login to your server - added to "ssh-keys" folder in bitwarden

### Build:
- clone this repo
- ```go build -o ./sshbw .```
- ```mv ./sshbw /usr/local/bin/sshbw && chmod +x /usr/local/bin/sshbw```

### Usage:
- ```sshbw username@hostname```



### ToDo:
- [ ] Add support for ssh keys?
- [ ] Folder/Org/Collection choice?
- [ ] --help flag with usage