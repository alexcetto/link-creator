# Project

This is a research project to create links with little database usage.

## Goals

- provide redirection to connected URL
- connect click to a profile
- privacy
- reliability
- fast to create and to redirect

## Other considerations

- private domain
- delete profile

## proposed solution

- hashing url
- with profile
- with a nonce saved in the activity

## Usage

With go installed on your machine, run `make run`.

You can use with either aes or elliptic curves using the `MODE=[aes|ec]` environment variable.

```bash
curl -X POST --location "http://localhost:8067/sendingid" \
    -H "Accept: application/json" \
    -d "{
\"url\": \"https://eu-west-1.console.aws.amazon.com/cloudwatch/home?region=eu-west-1"
}"
```
