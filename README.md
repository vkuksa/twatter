# Description

Twitter feed back end 
Use docker-compose and any programming language
- Implement an endpoint to add message
- Implement an endpoint to get feed (get existing messages and stream new ones - use HTTP streaming)
- Implement back pressure for message creation (use RabbitMQ/Kafka)
- Use Cockroachdb(at least 2-node cluster) as a database
- Implement a bot to generate messages (at configurable speed)
- CRITICAL: Project must start with one command (bash file) without installing anything except docker

Result is a link to a git project

# Overview

Project has two modules: twatterd and spammer.

Twatterd is a backend daemon, that adds incoming messages into storage, and displays feed from them.

Spammer is a bot, which connects to twatterd and generates messages, that populate the feed. 

# REST

There were no REST guidelines used for this task. Endpoints are strictly matching task  

# Usage

## Clone
```
$ git clone https://github.com/vkuksa/twatter.git
```
## Docker compose
```
$ make
```
## Local
```
$ make rebuild 
$ make init
$ ./bin/twatterd 
```
Requires local instances of cockroachDB to be installed and environment variable DATABASE_URL with a complete url to database set

# Tests

No integration testing completed for this project yet