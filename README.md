# cdk-dagger-example

Deploys a [sample Typescript CDK Stack](https://github.com/aws-samples/aws-cdk-examples/tree/master/typescript/ec2-instance) using a combination of [Dagger](https://dagger.io/) with the Go SDK and [Mage](https://magefile.org/).

The Dagger pipeline builds, lints and deploys the CDK Stack to AWS.

The pipeline can be run locally using the `go run main.go -w ../ deploy` command inside the `magefiles` folder.

Thanks to Mage, the individual steps of the build process (install, lint or deploy) can also be run using `go run main.go -w ../ {step}` inside the `magefiles` folder or by directly using Mage and running `mage {step}` in the `magefiles` folder.
