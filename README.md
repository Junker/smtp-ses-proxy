# SMTP-SES-Proxy
SMTP server that convert SMTP message to AWS SES API Call

### Build
	go build smtp.go

### Arguments
```
  --host HOST            SMTP Host (a.g. 127.0.0.1)
  --port PORT            SMTP port [default: 10025]
  --noauth               disable SMTP authentication
  --plainauth            enable SMTP PLAIN authentication
  --anonauth             enable SMTP anonymous authentication
  --user USER            SMTP username
  --password PASSWORD    SMTP password
  --region REGION        AWS region (a.g. eu-west-1)
  --sourcearn ARN        AWS Source ARN
  --fromarn ARN          AWS From ARN
  --returnpatharn ARN    AWS Return Path ARN
  --accesskey ACCESSKEY  AWS Access Key
  --secretkey SECRETKEY  AWS Secret Key
  --help, -h             display this help and exit
```
