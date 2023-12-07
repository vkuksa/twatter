# Description

Twitter-like feed back-env.

# Overview

Project has two modules: twatterd and spammer.

Twatterd is a backend daemon, that adds incoming messages into storage, and displays feed from them.

Spammer is a bot, which connects to twatterd and generates messages, that populate the feed. 

# REST

There were no REST guidelines used for this task. Endpoints are strictly matching task  

# Setup

## From scratch only with docker
Navigate to scripts/ folder of this repository and download install_and_run.sh script
```
$ bash install_and_run.sh
```
It will clone the repository and perform a start up

## Download source
```
$ git clone https://github.com/vkuksa/twatter.git
```
## Docker compose
Initialise and start service:
```
$ make
```
Stop service
```
$ make stop
```

Requires local instances of cockroachDB and RabbitMQ to be installed and corresponding environment variables initialised

## Non-container installation
```
$ make rebuild 
$ make init
$ ./bin/twatterd 
```

# Usage
App has two components: daemon and message generator.

View feed of messages:
```
$ curl localhost:9876/feed
```
Add message to storage:
```
$ curl -X POST -F 'content="example"' localhost:9876/add
```

## Twatter

Twatter has a `-address` option, which specifies a port to run app on (default `:9876`)

## Spammer

Has option `-destination` which specifies the address to spam messages to (default `http://localhost:9876/add`)

Has option `-pace` which specifies the intencity to generate messages (default `1s`)

# Tests

Does not have integration tests yet.
