A new parameter "geolocation_url" has been added to the config.

The handler code uses a free service to determine geolocation "https://api.findip.net /".
Register and receive an IP token.
Set the parameter:
"geolocation_url": "https://api.findip.net/%s/?token=<_your_token_>"

Example: "geolocation_url": "https://api.findip.net/%s/?token=200000000000000000000"

If the "geolocation_url" parameter is an empty string, "geolocation_url": "", the handler will skip the geolocation code block.
