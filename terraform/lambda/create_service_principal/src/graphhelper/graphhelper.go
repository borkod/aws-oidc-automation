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
	"github.com/microsoftgraph/msgraph-sdk-go/serviceprincipals"
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

// CreateServicePrincipal creates a service principal for the given app ID
func (g *GraphHelper) CreateServicePrincipal(appId string) (models.ServicePrincipalable, error) {
	requestBody := models.NewServicePrincipal()
	requestBody.SetAppId(&appId)

	servicePrincipal, err := g.appClient.ServicePrincipals().
		Post(context.Background(), requestBody, nil)
	if err != nil {
		return nil, err
	}
	return servicePrincipal, nil
}

// CreateAppWithServicePrincipal creates both an app registration and its service principal
func (g *GraphHelper) CreateAppWithServicePrincipal(name string) (appId string, servicePrincipalId string, err error) {
	// First, create the application registration
	app, err := g.CreateApp(name)
	if err != nil {
		return "", "", fmt.Errorf("failed to create app: %w", err)
	}

	appIdPtr := app.GetAppId()
	if appIdPtr == nil {
		return "", "", fmt.Errorf("app ID is nil after creation")
	}
	appId = *appIdPtr

	// Then, create the service principal for this app
	sp, err := g.CreateServicePrincipal(appId)
	if err != nil {
		return appId, "", fmt.Errorf("failed to create service principal for app %s: %w", appId, err)
	}

	spIdPtr := sp.GetId()
	if spIdPtr == nil {
		return appId, "", fmt.Errorf("service principal ID is nil after creation")
	}
	servicePrincipalId = *spIdPtr

	return appId, servicePrincipalId, nil
}

// SetApplicationIdUri sets the Application ID URI (identifier URI) for an app registration
// This is used to "Expose an API" in the Azure Portal
func (g *GraphHelper) SetApplicationIdUri(appId string, applicationIdUri string) error {
	// Get the application's object ID first
	filter := fmt.Sprintf("appId eq '%s'", appId)
	requestParameters := &applications.ApplicationsRequestBuilderGetQueryParameters{
		Filter: &filter,
		Select: []string{"id", "appId", "identifierUris"},
	}
	configuration := &applications.ApplicationsRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParameters,
	}

	appsResponse, err := g.appClient.Applications().Get(context.Background(), configuration)
	if err != nil {
		return fmt.Errorf("failed to get application: %w", err)
	}

	apps := appsResponse.GetValue()
	if len(apps) == 0 {
		return fmt.Errorf("no application found with app ID %s", appId)
	}

	if len(apps) > 1 {
		return fmt.Errorf("multiple applications found with app ID %s", appId)
	}

	objectId := apps[0].GetId()
	if objectId == nil {
		return fmt.Errorf("application object ID is nil")
	}

	// Update the application with the identifier URI
	requestBody := models.NewApplication()
	identifierUris := []string{applicationIdUri}
	requestBody.SetIdentifierUris(identifierUris)

	_, err = g.appClient.Applications().ByApplicationId(*objectId).Patch(context.Background(), requestBody, nil)
	if err != nil {
		return fmt.Errorf("failed to update application ID URI: %w", err)
	}

	return nil
}

// SetApplicationIdUriByName sets the Application ID URI for an app registration by name
func (g *GraphHelper) SetApplicationIdUriByName(name string, applicationIdUri string) error {
	appId, err := g.GetApp(name)
	if err != nil {
		return fmt.Errorf("failed to get app: %w", err)
	}

	return g.SetApplicationIdUri(appId, applicationIdUri)
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

// GetServicePrincipalByAppId retrieves a service principal by app ID
func (g *GraphHelper) GetServicePrincipalByAppId(appId string) (models.ServicePrincipalable, error) {
	filter := fmt.Sprintf("appId eq '%s'", appId)
	requestParameters := &serviceprincipals.ServicePrincipalsRequestBuilderGetQueryParameters{
		Filter: &filter,
		Select: []string{"id", "appId", "displayName"},
	}
	configuration := &serviceprincipals.ServicePrincipalsRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParameters,
	}

	spResponse, err := g.appClient.ServicePrincipals().Get(context.Background(), configuration)
	if err != nil {
		return nil, err
	}

	sps := spResponse.GetValue()
	if len(sps) == 0 {
		return nil, fmt.Errorf("no service principal found for app ID %s", appId)
	}

	if len(sps) > 1 {
		return nil, fmt.Errorf("multiple service principals found for app ID %s", appId)
	}

	return sps[0], nil
}

// DeleteServicePrincipalByAppId deletes a service principal by app ID
func (g *GraphHelper) DeleteServicePrincipalByAppId(appId string) error {
	sp, err := g.GetServicePrincipalByAppId(appId)
	if err != nil {
		return err
	}

	spId := sp.GetId()
	if spId == nil {
		return fmt.Errorf("service principal ID is nil")
	}

	err = g.appClient.ServicePrincipals().ByServicePrincipalId(*spId).Delete(context.Background(), nil)
	if err != nil {
		return err
	}

	return nil
}

// DeleteAppWithServicePrincipal deletes both the service principal and app registration
func (g *GraphHelper) DeleteAppWithServicePrincipal(name string) (string, error) {
	// First, get the app to find its appId
	appId, err := g.GetApp(name)
	if err != nil {
		return "", fmt.Errorf("failed to get app: %w", err)
	}

	// Delete the service principal first (if it exists)
	err = g.DeleteServicePrincipalByAppId(appId)
	if err != nil {
		// Log but don't fail if service principal deletion fails
		fmt.Printf("Warning: failed to delete service principal for app %s: %v\n", appId, err)
	}

	// Then delete the app registration
	err = g.DeleteApp(name)
	if err != nil {
		return appId, fmt.Errorf("failed to delete app: %w", err)
	}

	return appId, nil
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
