# Write a Target Provider using Go

## High-Level Workflow
To create a new provider:

1. Create a New Folder:
    Under the `api/pkg/apis/v1alpha1/providers/target` directory, create a new folder for your provider (e.g., `myprovider`).

2. Create Implementation and Test Files:
    Inside your new provider folder, create two files:

    * One file for the provider implementation (e.g., `myprovider.go`).
    * Another file for unit tests (e.g., `myprovider_test.go`).

    > **Tip:** A quick way to start is to copy an existing provider implementation and modify it to fit your needs.

3. Implement the Provider Interface:
    In your provider source code, implement the target provider interface. Typically, a provider defines an associated configuration type, which will be injected into the provider instance during initialization.

4. Write Unit Tests:
    Implement relevant unit test cases to ensure your provider functions as expected.

5. Update the Provider Factory:
    Modify `api/pkg/apis/v1alpha1/providers/providerfactory.go` to include the creation of your provider based on its name (this update is usually required in two places).

## Packaing the Provider
Once your provider is created, it will be automatically built and packaged as part of the standard Symphony build pipeline.