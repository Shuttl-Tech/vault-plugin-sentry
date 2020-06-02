package main

const getOrgResp = `
{
  "id": "5", 
  "slug": "%s", 
  "name": "%s"
}
`

const getProjectResp = `
{
  "id": "5", 
  "slug": "%s", 
  "name": "%s", 
  "status": "active"
}
`

const getKeysResp = `
[
  {
    "label": "%s", 
    "name": "%s", 
    "id": "5", 
    "isActive": true, 
    "public": "%s", 
    "secret": "%s",
    "dsn": {
      "cdn": "https://sentry.io/js-sdk-loader/%s.min.js", 
      "csp": "https://sentry.io/api/2/csp-report/?sentry_key=%s", 
      "minidump": "https://sentry.io/api/2/minidump/?sentry_key=%s", 
      "public": "https://%s@sentry.io/2", 
      "secret": "https://%s:%s@sentry.io/2", 
      "security": "https://sentry.io/api/2/security/?sentry_key=%s"
    }
  }
]
`

const singleKey = `
{
  "label": "%s", 
  "name": "%s"
  "dsn": {
    "public": "https://%s@sentry.io/2", 
  }, 
}
`