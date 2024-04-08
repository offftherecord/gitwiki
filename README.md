# Gitwiki

This tool is intended to help discovery Github repositories that have writeable wikis. Writeable wikis can be abused by adversaries to assist with social engineering attacks. This can happen by modifying, or adding wiki pages that instruct users to perform malicious activity. Because users typically trust organization repositories, they are inherently trusting of the wiki and may follow instructions blindly.


## Installing Gitwiki

Gitwiki requires  **go1.21**  to install successfully. Run the following command to install the latest version -
```
go install -v github.com/offftherecord/gitwiki@latest
```


### Usage
```
echo single_repo| gitwiki
cat list_of_repos | gitwiki
gitwiki single_repo
```
Gitwiki will accept repositories via stdin or as an argument

# License
This project is licensed under the MIT License. See the [LICENSE.md](LICENSE.md) file for details.
