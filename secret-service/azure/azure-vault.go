package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"

	"github.com/gerdooshell/tax-logger/environment"
	"github.com/gerdooshell/tax-logger/lib/cache/lrucache"
	secretService "github.com/gerdooshell/tax-logger/secret-service"
)

var secretServiceCache = lrucache.NewLRUCache[string](20)

func NewSecretService(uri string) secretService.SecretService {
	service, err := secretServiceCache.Read(uri)
	if err == nil {
		return service.(*azureSecretService)
	}
	newService := &azureSecretService{
		uri:   uri,
		cache: lrucache.NewLRUCache[string](1000),
	}
	secretServiceCache.Add(uri, newService)
	return newService
}

type azureSecretService struct {
	uri    string
	client *azsecrets.Client
	cache  lrucache.LRUCache[string]
}

func (az *azureSecretService) connectToVault() (err error) {
	var cred azcore.TokenCredential
	if environment.GetEnvironment() == environment.Dev {
		cred, err = azidentity.NewDefaultAzureCredential(nil)
	} else {
		cred, err = azidentity.NewManagedIdentityCredential(nil)
	}
	if err != nil {
		return err
	}
	options := azsecrets.ClientOptions{
		DisableChallengeResourceVerification: true,
	}
	client, err := azsecrets.NewClient(az.uri, cred, &options)
	if err != nil {
		return err
	}
	az.client = client
	return nil
}

func (az *azureSecretService) GetSecretValue(ctx context.Context, secretKey string) (<-chan string, <-chan error) {
	out := make(chan string)
	errChan := make(chan error)
	go func() {
		defer close(out)
		defer close(errChan)
		cachedValue, err := az.cache.Read(secretKey)
		if err == nil {
			out <- cachedValue.(string)
			return
		}
		if err := az.connectToVault(); err != nil {
			errChan <- err
		}
		version := ""
		resp, err := az.client.GetSecret(ctx, secretKey, version, nil)
		if err != nil {
			errChan <- err
			return
		}
		value := resp.Value
		if value == nil {
			errChan <- fmt.Errorf("secret key not found")
		}
		az.cache.Add(secretKey, *value)
		out <- *value
	}()

	return out, errChan
}
