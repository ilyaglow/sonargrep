sonargrep
---------

## Install

```
go install github.com/ilyaglow/sonargrep
```

## Usage

Grep IPs from `sonar.https` dataset that contain wordpress in their response
body without saving a 50GB file to the disk:
```
curl -L -s https://opendata.rapid7.com/sonar.https/2018-04-24-1524531601-https_get_443.json.gz \
    | sonargrep -w wordpress -i \
    | jq -r '.ip'
```
