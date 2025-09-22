package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/borkod/poc-aws-azure-oidc/tf-infra/lambda/delete_service_principal/src/graphhelper"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// Response structure
type Response struct {
	StatusCode int    `json:"statusCode"`
	AppID      string `json:"appId,omitempty"`
}

type eventStruct struct {
	Account   string `json:"account"`
	EventName string `json:"eventName"`
	RoleName  string `json:"roleName"`
}

var (
	ssmClient *ssm.Client
)

func init() {
	// Initialize the S3 client outside of the handler, during the init phase
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	ssmClient = ssm.NewFromConfig(cfg)
}

func handleRequest(ctx context.Context, event json.RawMessage) (Response, error) {

	clientID := os.Getenv("CLIENT_ID")
	tenantID := os.Getenv("TENANT_ID")
	paramName := os.Getenv("CLIENT_SECRET_SSM")

	clientSecret, err := getSSMParamValue(ctx, paramName)
	if err != nil {
		log.Println("Error getting SSM parameter:", err)
		return Response{StatusCode: 500}, err
	}

	var evt eventStruct
	err = json.Unmarshal(event, &evt)
	if err != nil {
		log.Println("Error unmarshalling event:", err)
		return Response{StatusCode: 400}, err
	}

	graphHelper := graphhelper.NewGraphHelper()

	err = initializeGraph(graphHelper, clientID, tenantID, clientSecret)
	if err != nil {
		log.Println("Error initializing graph:", err)
		return Response{StatusCode: 500}, err
	}

	appName := "aws-" + evt.Account + "-" + evt.RoleName

	appID, err := graphHelper.GetApp(appName)
	if err != nil {
		log.Println("Error getting app:", err)
		return Response{StatusCode: 500}, err
	}

	err = graphHelper.DeleteApp(appName)
	if err != nil {
		log.Println("Error deleting app:", err)
		return Response{StatusCode: 500}, err
	}

	return Response{
		StatusCode: 200,
		AppID:      appID,
	}, nil
}

func initializeGraph(graphHelper *graphhelper.GraphHelper, clientID, tenantID, clientSecret string) error {
	err := graphHelper.InitializeGraphForAppAuth(clientID, tenantID, clientSecret)
	if err != nil {
		log.Println("Error initializing Graph for app auth: ", err)
		return err
	}
	return nil
}

func getSSMParamValue(ctx context.Context, name string) (string, error) {
	withDecryption := true
	resp, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           &name,
		WithDecryption: &withDecryption,
	})
	if err != nil {
		log.Println("Error getting parameter:", err)
		return "", err
	}
	if resp == nil || resp.Parameter == nil {
		log.Println("Parameter not found")
		return "", errors.New("parameter not found")
	}
	return *resp.Parameter.Value, nil
}

func main() {
	lambda.Start(handleRequest)
}
