# Secret Management
It's not good practice to keep secrets in plain texts in configuration objects. Symphony recommends keeping secrets in secret stores of your choice (such as Azure Key Vault and Kubernetes secret stores), and use the `$secret()` expression to refer to them in your artifacts such as configurations. 

Because the `$secret()` expression is universally supported in Symphony artifact types, you don't have to use a Catalog object to refer to a secret. Instead, you can directly refer to your secrets in other artifacts such as Solutions and Targets.