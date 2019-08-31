package services

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/dgoujard/uptimeWorker/config"
	"log"
	"os"
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

func (a *AwsService)CallUptimeCheckLambdaFunc(arn string,site SiteBdd) *CheckSiteResponse {
	tmpListArmParams := strings.Split(arn,":")

	client := lambda.New(a.session, &aws.Config{Region: aws.String(tmpListArmParams[2])})

	payload, err := json.Marshal(site)
	if err != nil {
		fmt.Println("Error marshalling request")
		os.Exit(0)
	}

	result, err := client.Invoke(&lambda.InvokeInput{FunctionName: aws.String("uptimeCheck"), Payload: payload})
	if err != nil {
		fmt.Println("Error calling uptimeCheck")
		os.Exit(0)
	}
	var resultUptime CheckSiteResponse
	err = json.Unmarshal(result.Payload,&resultUptime)
	if err != nil {
		fmt.Println("Error decoding réponse uptimeCheck")
		os.Exit(0)
	}
	return &resultUptime
}