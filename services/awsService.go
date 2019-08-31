package services

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/dgoujard/uptimeWorker/config"
	"log"
	"strings"
)

type AwsService struct {
	session *session.Session
}

func CreateAwsService(config *config.AwsConfig) *AwsService {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials(config.AccessKey,config.SecretKey,""),
	})
	if err != nil {
		log.Fatalln(err)
	}
	return &AwsService{session:sess}
}

func (a *AwsService)CallUptimeCheckLambdaFunc(arn string,site *SiteBdd) (error,*CheckSiteResponse) {
	tmpListArmParams := strings.Split(arn,":")
	client := lambda.New(a.session, &aws.Config{Region: aws.String(tmpListArmParams[3])})

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