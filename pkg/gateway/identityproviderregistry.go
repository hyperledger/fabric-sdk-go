package gateway


const defaultProviders []IdentityProvider

type IdentityProviderRegistry struct {
	providers map[string]IdentityProvider
}

func (ipr IdentityProviderRegistry) GetProvider(registrytype string) (IdentityProvider, error){
	provider := ipr.providers[registrytype]
	if (provider == nil) {
		return nil, error.Wrap("No Identity Provider Registry was found")
	}
	return provider, nil
}



func (ipr IdentityProviderRegistry) DefaultProviderRegistry() (IdentityProviderRegistry) {
	registry := IdentityProviderRegistry{}
	for _, provider := range defaultProviders {
	}
}