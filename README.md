# URL shortener service written in GO

## Description

It uses [Google Shortener API](https://developers.google.com/url-shortener/v1/getting_started) to convert long urls to small ones.

Pre-requisites:
 
 * [requesting API KEY](https://developers.google.com/url-shortener/v1/getting_started#APIKey), see code.
 * **root** privileges to run.

Tested on Ubuntu 17.04 only.

## Build

```
go build shortener.go keymapper.go
```

## Usage

1. Copy (**Ctrl + C**) an URL
2. Press **F10**
3. Paste (**Ctrl + V**) a short URL
