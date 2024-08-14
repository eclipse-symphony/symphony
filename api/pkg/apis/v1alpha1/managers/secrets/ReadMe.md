# Secret Manager

The Secret Manager implements the `IExtSecretProvider` interface and is part of the `EvaluationContext`, which is used in expression parsing. It only has one function, which is reading the secret object, field field with a local evaluation context if provided.

The Secret Manager decides which secret provider to use when processing the parsing request. It has three options:

1. **Secret Provider Name Specified**: If the secret provider name is specified like `Get("k8s::sampleobject", "samplefield", nil)`, then that specific secret provider is used.

2. **Single Secret Provider**: If the secret provider name is not specified but the Secret Manager only has one secret provider, then that secret provider is used by default.

3. **Multiple Secret Providers**: If the secret provider name is not specified and the Secret Manager has multiple secret providers, then the manager tries to parse the expression in the order of precedence. The first successful result is returned.