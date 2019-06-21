# Go API Server for Infra Provisioner

Infra Provisioner brings up a test bed by spawning a set of services requested by the test client.

## Overview
- API version: 1.0.0
- Build date: 2019-06-12T08:47:33.654Z
- Contributors: Arun, Vibhore

Provides following functionalities as of now:
 - Environment creation multiple Containers based on input
 - Get details about environment
 - Delete environment
    - Stop running container
    - Kill running container
    - Delete Mongo record

### Running the server
To run the server, follow these simple steps:

```
go run main.go
```

### Help
To see the help related to supported calls browse following link

```
Creates a new testbed environment -

http://<server-ip>:<server-port>/set/createnv/ 
POST body: {"name" : "testbed", "containers" : ["mongo", "redis", "mysql"]}
```

```
Get whole environment detail

http://<server-ip>:<server-port>/get/getenv/
```

```
Get details about a particular test bed based on tag

http://<server-ip>:<server-port>/get/getenv/{tag}
```

```
Stop a container

http://<server-ip>:<server-port>/update/stop/{tag}
```

```
Delete a container

http://<server-ip>:<server-port>/delete/container/{tag}
```
