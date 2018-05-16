package move

import (
	"log"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/satori/go.uuid"
)


func Move(src string, dest string, awsRegion string, awsProfile string, messageGroupId string) {

	log.Printf("source queue : %v", src)
	log.Printf("destination queue : %v", dest)
	log.Printf("region : %v", awsRegion)
	log.Printf("profile : %v", awsProfile)
	log.Printf("messageGroupId : %v", messageGroupId)

	client := sqs.New(session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: &awsRegion},
		Profile:           awsProfile,
		SharedConfigState: session.SharedConfigEnable,
	})))

	maxMessages := int64(10)
	waitTime := int64(0)
	messageAttributeNames := aws.StringSlice([]string{"All"})

	rmin := &sqs.ReceiveMessageInput{
		QueueUrl:              &src,
		MaxNumberOfMessages:   &maxMessages,
		WaitTimeSeconds:       &waitTime,
		MessageAttributeNames: messageAttributeNames,
	}

	// loop as long as there are messages on the queue
	for {
		resp, err := client.ReceiveMessage(rmin)

		if err != nil {
			panic(err)
		}

		if len(resp.Messages) == 0 {
			log.Printf("done")
			return
		}

		log.Printf("received %v messages...", len(resp.Messages))

		var wg sync.WaitGroup
		wg.Add(len(resp.Messages))

		for _, m := range resp.Messages {
			go func(m *sqs.Message) {
				defer wg.Done()

				// write the message to the destination queue
				if messageGroupId == "" {
					smi := sqs.SendMessageInput{
						MessageAttributes: 			m.MessageAttributes,
						MessageBody:       			m.Body,
						QueueUrl:          			&dest,
					}

					_, err := client.SendMessage(&smi)

					if err != nil {
						log.Printf("ERROR sending message to destination %v", err)
						return
					}

				} else { // Add MessageGroupId and MessageDeduplicationIdn for fifo
					msgDeduplicationId, err1 := uuid.NewV4()
					if err1 != nil {
						panic(err1)
					}

					smi := sqs.SendMessageInput{
						MessageAttributes: 			m.MessageAttributes,
						MessageBody:       			m.Body,
						QueueUrl:          			&dest,
						MessageGroupId:    			&messageGroupId,
						MessageDeduplicationId: aws.String(msgDeduplicationId.String()),
					}

					_, err := client.SendMessage(&smi)

					if err != nil {
						log.Printf("ERROR sending message to destination %v", err)
						return
					}
				}

				// message was sent, dequeue from source queue
				dmi := &sqs.DeleteMessageInput{
					QueueUrl:      &src,
					ReceiptHandle: m.ReceiptHandle,
				}

				if _, err := client.DeleteMessage(dmi); err != nil {
					log.Printf("ERROR dequeueing message ID %v : %v",
						*m.ReceiptHandle,
						err)
				}
			}(m)
		}

		// wait for all jobs from this batch...
		wg.Wait()
	}
}
