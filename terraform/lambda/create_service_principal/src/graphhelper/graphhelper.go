package graphhelper

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	abstractions "github.com/microsoft/kiota-abstractions-go"
	auth "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/applications"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
)

type GraphHelper struct {
	clientSecretCredential *azidentity.ClientSecretCredential
	appClient              *msgraphsdk.GraphServiceClient
}

func NewGraphHelper() *GraphHelper {
	g := &GraphHelper{}
	return g
}

func (g *GraphHelper) InitializeGraphForAppAuth(clientId string, tenantId string, clientSecret string) error {

	credential, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, nil)
	if err != nil {
		return err
	}

	g.clientSecretCredential = credential

	// Create an auth provider using the credential
	authProvider, err := auth.NewAzureIdentityAuthenticationProviderWithScopes(g.clientSecretCredential, []string{
		"https://graph.microsoft.com/.default",
	})
	if err != nil {
		return err
	}

	// Create a request adapter using the auth provider
	adapter, err := msgraphsdk.NewGraphRequestAdapter(authProvider)
	if err != nil {
		return err
	}

	// Create a Graph client using request adapter
	client := msgraphsdk.NewGraphServiceClient(adapter)
	g.appClient = client

	return nil
}

func (g *GraphHelper) GetAppToken() (*string, error) {
	token, err := g.clientSecretCredential.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{
			"https://graph.microsoft.com/.default",
		},
	})
	if err != nil {
		return nil, err
	}

	return &token.Token, nil
}

func (g *GraphHelper) GetUsers() (models.UserCollectionResponseable, error) {
	var topValue int32 = 25
	query := users.UsersRequestBuilderGetQueryParameters{
		// Only request specific properties
		Select: []string{"displayName", "id", "mail"},
		// Get at most 25 results
		Top: &topValue,
		// Sort by display name
		Orderby: []string{"displayName"},
	}

	return g.appClient.Users().
		Get(context.Background(),
			&users.UsersRequestBuilderGetRequestConfiguration{
				QueryParameters: &query,
			})
}

func (g *GraphHelper) ListApps() (models.ApplicationCollectionResponseable, error) {
	var topValue int32 = 25
	query := applications.ApplicationsRequestBuilderGetQueryParameters{
		// Only request specific properties
		Select: []string{"displayName", "id", "appId"},
		// Get at most 25 results
		Top: &topValue,
		// Sort by display name
		//Orderby: []string{"displayName"},
	}

	return g.appClient.Applications().
		Get(context.Background(),
			&applications.ApplicationsRequestBuilderGetRequestConfiguration{
				QueryParameters: &query,
			})
}

func (g *GraphHelper) CreateApp(name string) (models.Applicationable, error) {
	requestBody := models.NewApplication()
	requestBody.SetDisplayName(&name)

	applications, err := g.appClient.Applications().
		Post(context.Background(), requestBody, nil)
	if err != nil {
		return nil, err
	}
	return applications, nil
}

func (g *GraphHelper) DeleteApp(name string) error {
	headers := abstractions.NewRequestHeaders()
	headers.Add("ConsistencyLevel", "eventual")

	//requestFilter := "startswith(displayName, 'a')"
	requestSearch := "\"displayName:" + name + "\""
	requestCount := true
	//requestTop := int32(1)

	requestParameters := &applications.ApplicationsRequestBuilderGetQueryParameters{
		Search:  &requestSearch,
		Count:   &requestCount,
		Orderby: []string{"displayName"},
		Select:  []string{"appId", "identifierUris", "displayName"},
	}
	configuration := &applications.ApplicationsRequestBuilderGetRequestConfiguration{
		Headers:         headers,
		QueryParameters: requestParameters,
	}

	// To initialize your graphClient, see https://learn.microsoft.com/en-us/graph/sdks/create-client?from=snippets&tabs=go
	appsResponse, err := g.appClient.Applications().Get(context.Background(), configuration)
	if err != nil {
		return err
	}

	apps := appsResponse.GetValue()
	if len(apps) > 1 {
		return fmt.Errorf("multiple apps found with name %s", name)
	}

	if len(apps) == 0 {
		return fmt.Errorf("no apps found with name %s", name)
	}

	appId := apps[0].GetAppId()
	if appId == nil {
		return fmt.Errorf("appId is nil")
	}

	err = g.appClient.ApplicationsWithAppId(appId).Delete(context.Background(), nil)
	if err != nil {
		return err
	}
	return nil
}

func (g *GraphHelper) CheckAppExists(name string) (bool, error) {
	headers := abstractions.NewRequestHeaders()
	headers.Add("ConsistencyLevel", "eventual")

	//requestFilter := "startswith(displayName, 'a')"
	requestSearch := "\"displayName:" + name + "\""
	requestCount := true
	//requestTop := int32(1)

	requestParameters := &applications.ApplicationsRequestBuilderGetQueryParameters{
		Search:  &requestSearch,
		Count:   &requestCount,
		Orderby: []string{"displayName"},
		Select:  []string{"appId", "identifierUris", "displayName"},
	}
	configuration := &applications.ApplicationsRequestBuilderGetRequestConfiguration{
		Headers:         headers,
		QueryParameters: requestParameters,
	}

	// To initialize your graphClient, see https://learn.microsoft.com/en-us/graph/sdks/create-client?from=snippets&tabs=go
	appsResponse, err := g.appClient.Applications().Get(context.Background(), configuration)
	if err != nil {
		return false, err
	}

	apps := appsResponse.GetValue()

	if len(apps) == 0 {
		return false, nil
	}

	if len(apps) > 1 {
		return true, nil
	}

	appId := apps[0].GetAppId()
	if appId == nil {
		return false, fmt.Errorf("appId is nil")
	}

	return true, nil
}

func (g *GraphHelper) GetApp(name string) (string, error) {
	headers := abstractions.NewRequestHeaders()
	headers.Add("ConsistencyLevel", "eventual")

	//requestFilter := "startswith(displayName, 'a')"
	requestSearch := "\"displayName:" + name + "\""
	requestCount := true
	//requestTop := int32(1)

	requestParameters := &applications.ApplicationsRequestBuilderGetQueryParameters{
		Search:  &requestSearch,
		Count:   &requestCount,
		Orderby: []string{"displayName"},
		Select:  []string{"appId", "identifierUris", "displayName"},
	}
	configuration := &applications.ApplicationsRequestBuilderGetRequestConfiguration{
		Headers:         headers,
		QueryParameters: requestParameters,
	}

	// To initialize your graphClient, see https://learn.microsoft.com/en-us/graph/sdks/create-client?from=snippets&tabs=go
	appsResponse, err := g.appClient.Applications().Get(context.Background(), configuration)
	if err != nil {
		return "", err
	}

	apps := appsResponse.GetValue()
	if len(apps) > 1 {
		return "", fmt.Errorf("multiple apps found with name %s", name)
	}

	if len(apps) == 0 {
		return "", fmt.Errorf("no apps found with name %s", name)
	}

	appId := apps[0].GetAppId()
	if appId == nil {
		return "", fmt.Errorf("appId is nil")
	}

	return *appId, nil
}
