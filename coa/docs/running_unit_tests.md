# Running Unit Tests

## In Visual Studio Code
Once you have Visual Studio Code Go extension installed, you should see a ```run test|debug test``` link on top of each unit test. Setting up a break point and debugging through a unit test case is a good way to understand how the code works.

## Use Go
To run all unit tests, use ```go test``` under the root folder (this is also executed upon Git pushes):
```bash
go test ./... -v
```
> **NOTE**: Some platform-specific tests are skipped by default. If you want to run these tests, you need to set up corresponding environment variables. Please see instructions in ```scripts/run_tests.sh``` for more details.

## Use the test script
The ```scripts/run_tests.sh``` script can be used to run all unit test cases, especially the ones that are usually skipped. To test all COA cases, use:
```bash
./run_tests.sh coa
```
> **NOTE**: The script also run Symphony API tests and Symphony K8s tests when ```all``` is passed as parameter.