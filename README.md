sonargrep
---------

## Install

```
go install github.com/ilyaglow/sonargrep
```

## Usage

```
curl -L -s https://opendata.rapid7.com/sonar.https/2018-04-24-1524531601-https_get_443.json.gz | sonargrep -w wordpress -i | jq -r '.ip'
```

The example achieves:
* Greps records that contain wordpress (case insensitive) in their http response body from `sonar.https` dataset.
* Extracts IPs using [jq](https://stedolan.github.io/jq/).
* All this without saving a 50GB file to the disk.

## Flow

Basically, here is how my research flow looks like:

* Spin up a digital ocean droplet
* Start grepping
* Wait for an hour or something, while playing with the results that come on the fly
* Kill the droplet
