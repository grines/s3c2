# S3C2 - AWS S3 C2 Framework

This C2 utilizes an s3 bucket for comms. Constists of a CLI and Payload

## Status
- This is still an active work in progress.
- Not ready for production

- **Maintained by**:
[grines](https://github.com/)

# Features
- [X] Command Completion
- [X] Dynamic resource listing
- [X] Command history
- [X] S3 bucket c2 channel (for those hard to reach places)

## Installation

s3c2 is written in golang so its easy to ship around as a binary.

- Grab a tightly scoped aws key limited to read /write a single bucket. This will go into /test/payload.go.
- Build Payload

- s3c2 cli
- load AWS creds that has access to the same bucket inside of .aws/credentials
- select profile with token autocomplete. (s3c2 will list profiles availaible in your .aws/credentials file)
- choose bucket

# Demo

![](https://github.com/grines/s3c2/blob/main/demo.gif)

