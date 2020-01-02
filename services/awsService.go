package services

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/dgoujard/uptimeWorker/config"
	"log"
	"strings"
)

type AwsService struct {
	Session *session.Session
}

func CreateAwsService(config *config.AwsConfig) *AwsService {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials(config.AccessKey,config.SecretKey,""),
	})
	if err != nil {
		log.Fatalln(err)
	}
	return &AwsService{Session:sess}
}

func (a *AwsService)CallUptimeCheckLambdaFunc(arn string,site *SiteBdd) (error,*CheckSiteResponse) {
	tmpListArmParams := strings.Split(arn,":")
	client := lambda.New(a.Session, &aws.Config{Region: aws.String(tmpListArmParams[3])})

	payload, err := json.Marshal(site)
	if err != nil {
		log.Println("Error marshalling request")
		return err, nil
	}

	result, err := client.Invoke(&lambda.InvokeInput{FunctionName: aws.String(tmpListArmParams[6]), Payload: payload})
	if err != nil {
		log.Println("Error calling Lambda")
		return err, nil
	}
	var resultUptime CheckSiteResponse
	err = json.Unmarshal(result.Payload,&resultUptime)
	if err != nil {
		log.Println("Error decoding réponse uptimeCheck")
		return err, nil
	}
	return nil,&resultUptime
}

func (a *AwsService) SendEmail(from string, to string,subject string, Htmlmessage string, Txtmessage string) error {
	// Create an SES session.
	svc := ses.New(a.Session)
	// Assemble the email.
	body := &ses.Body{}
	if len(Htmlmessage) > 0 {
		body.Html = &ses.Content{
			Data:    aws.String(Htmlmessage),
		}
	}
	if len(Txtmessage) > 0 {
		body.Text = &ses.Content{
			Data:    aws.String(Txtmessage),
		}
	}

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(to),
			},
		},
		Message: &ses.Message{
			Body: body,
			Subject: &ses.Content{
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(from),
		// Uncomment to use a configuration set
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}

	// Attempt to send the email.
	_, err := svc.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		return err
	}
	return nil
}