# gsuite-log-exporter

This tool export G Suite audit logs and save them to AWS S3. It is deployed by AWS CloudFormation and runs as Serverless Application.

## Prerequisite

- go >= 1.11
- awscli >= 1.16.90
- automake >= 3.81

## Setup

### Get API credential of G Suite

1. Go to https://console.cloud.google.com and create a new project.
2. Go to https://console.cloud.google.com/apis/credentials and create a new OAuth 2.0 client ID.
3. Download a credential JSON file and save it as `client.json`

### Create OAuth token

Clone this repository and build a helper tool.

```bash
$ git clone git@github.com:m-mizutani/gsuite-log-exporter
$ cd gsuite-log-exporter
$ make build/helper
```

After building the helper tool, you can retrieve OAuth token by the tool and Web browser.


```bash
$ ./build/helper oauth /path/to/client.json token.json
Go to the following link in your browser then type the authorization code:
https://accounts.google.com/o/oauth2/auth?access_type=offline&client_id=xxxxxxxxxxxxxxxxxxxxxx.apps.googleusercontent.com&redirect_uri=urn%3Aietf%3Awg%3Aoauth%3A2.0%3Aoob&response_type=code&scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fadmin.reports.audit.readonly&state=state-token
```

This tool displays URL to authorize. Open the URL by your Web browser. Click `Allow` button after confirming required permissions. Then the page displays temprorary code for authorization. Copy and paste it to terminal. Then `token.json` should be created.

### Create a secret item of AWS SecretsManager

Create a secret of [AWS Secrets Manager](https://console.aws.amazon.com/secretsmanager) and put following items. Then copy ARN of the secret such as `arn:aws:secretsmanager:ap-northeast-1:1234567890:secret:gsuite-log-exporter-xxxxxxxxxxx`

- `gsuite_client`: Content of `client.json`
- `gsuite_token`:  Content of `token.json`

### Create a config file

Then create a config file like following JSON. Save it as `config.json`.

```json
{
    "StackName": "your-gsuite-log-exporeter-stack",
    "CodeS3Bucket": "your-bucket-to-store-code",
    "CodeS3Prefix": "some-prefix",

    "SecretArn": "arn:aws:secretsmanager:ap-northeast-1:1234567890:secret:gsuite-log-exporter-xxxxxxxxxxx",
    "S3Region": "ap-northeast-1",
    "S3Bucket": "bucket-to-save-logs",
    "S3Prefix": "prefix-to-save-logs/"
}
```

Finally, you can invoke deployment command.

```bash
$ env CONFIG_FILE=config.json make deploy
```

## License

- Author: Masayoshi Mizutani < mizutani@sfc.wide.ad.jp >
- License: The 3-Clause BSD License
