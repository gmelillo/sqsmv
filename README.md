# sqsmv

Move all messages from one SQS queue, to another.

## Installation

    go get github.com/nathandeamer/sqsmv


## Configuration

AWS credentials file ~/.aws/credentials with `aws_access_key_id` and `aws_secret_access_key`  
AWS config file ~/.aws/config with required profiles

## Usage

Supply source and destination URL endpoints, region and profile.  
If FIFO also include messageGroupId (optional)

    sqsmv -src https://region.queue.amazonaws.com/123/queue-a -dest https://region.queue.amazonaws.com/123/queue-b -region eu-west-1 -profile test/prod

## Seeing is believing :)

Create some SQS messages to play with using the AWS CLI.

    for i in {0..24..1}; do
        aws sqs send-message \
            --queue-url https://ap-southeast-2.queue.amazonaws.com/123/wat-a
            --region eu-west-1
            --profile test/prod
            --message-body "{\"id\": $i}"
    done

## License

The MIT License (MIT)

Copyright (c) 2016 Scott Barr

See [LICENSE.md](LICENSE.md)
