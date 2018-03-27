
## Description

Tool for automatic dump HAR file by using **mitmproxy**.

## Usage 

### Install

install **mitpruxy** and **har2case** first 

`pip install mitmproxy`

`pip install har2case`

### Launch proxy server

go to the root path, then run 

`mitmdump -s proxy/mitmproxy/har_dump.py --set hardump=result.har --set request_host=localhost --set request_path=/api/user -p 7777`

* `hardump` required. Dump file name.

* `request_host` optional. Filter the host in request url, i.e. only dump the url match `request_host`.

* `request_path` optional. Filter the path in request url, i.e. only dump the url match `request_path`.

### Config browser proxy setting, and request 

### Stop proxy server

### Convert HAR file to test case 

`har2case result.har test_case.yaml`

