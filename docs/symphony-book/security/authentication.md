# Authentication

By default, Symphony provides a simple user store with basic password-based authentication. In a production environment, we recommend that you use an external identity provider (IdP), such as [Microsoft Entra ID](https://learn.microsoft.com/entra/fundamentals/whatis), [Google Accounts](https://accounts.google.com/), [Microsoft Accounts](https://account.microsoft.com/account), [Twitter Accounts](https://twitter.com/home), and many others.

For example, when you deploy to Kubernetes, you can configure a trust relationship between the selected IdP and your ingress and redirect all unauthenticated requests to the IdP. When authentication succeeds, the bearer token is passed to Symphony API calls. Then, you can configure your Symphony access policies to map security token claims to Symphony roles in your API configuration.
