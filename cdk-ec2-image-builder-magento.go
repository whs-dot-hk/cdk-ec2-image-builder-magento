package main

import (
	"log"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsimagebuilder"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type CdkEc2ImageBuilderMagentoStackProps struct {
	awscdk.StackProps
}

func NewCdkEc2ImageBuilderMagentoStack(scope constructs.Construct, id string, props *CdkEc2ImageBuilderMagentoStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	content, err := os.ReadFile("component.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// The code that defines your stack goes here
	component := awsimagebuilder.NewCfnComponent(stack, jsii.String("Component"), &awsimagebuilder.CfnComponentProps{
		Data:     jsii.String(string(content)),
		Name:     jsii.String("install-magento"),
		Platform: jsii.String("Linux"),
		Version:  jsii.String("1.0.0"),
	})

	imageRecipe := awsimagebuilder.NewCfnImageRecipe(stack, jsii.String("ImageRecipe"), &awsimagebuilder.CfnImageRecipeProps{
		Name:        jsii.String("magento"),
		ParentImage: awscdk.Fn_Sub(jsii.String("arn:${AWS::Partition}:imagebuilder:${AWS::Region}:aws:image/amazon-linux-2-x86/x.x.x"), nil),
		Version:     jsii.String("1.0.0"),
		BlockDeviceMappings: []interface{}{
			&awsimagebuilder.CfnImageRecipe_InstanceBlockDeviceMappingProperty{
				DeviceName: jsii.String("/dev/xvda"),
				Ebs: &awsimagebuilder.CfnImageRecipe_EbsInstanceBlockDeviceSpecificationProperty{
					DeleteOnTermination: jsii.Bool(true),
					VolumeSize:          jsii.Number(20),
				},
			},
		},
		Components: []interface{}{
			&awsimagebuilder.CfnImageRecipe_ComponentConfigurationProperty{
				ComponentArn: component.AttrArn(),
			},
		},
	})

	bucket := awss3.NewBucket(stack, jsii.String("Bucket"), &awss3.BucketProps{
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
	})

	role := awsiam.NewRole(stack, jsii.String("Role"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ec2.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonSSMManagedInstanceCore")),
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("EC2InstanceProfileForImageBuilder")),
		},
	})

	role.AttachInlinePolicy(awsiam.NewPolicy(stack, jsii.String("Policy"), &awsiam.PolicyProps{
		Statements: &[]awsiam.PolicyStatement{
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Actions: &[]*string{
					jsii.String("s3:PutObject"),
				},
				Resources: &[]*string{
					bucket.ArnForObjects(jsii.String("*")),
				},
			}),
		},
	}))

	instanceProfile := awsiam.NewCfnInstanceProfile(stack, jsii.String("InstanceProfile"), &awsiam.CfnInstanceProfileProps{
		Roles: &[]*string{
			role.RoleName(),
		},
	})

	infrastructureConfiguration := awsimagebuilder.NewCfnInfrastructureConfiguration(stack, jsii.String("InfrastructureConfiguration"), &awsimagebuilder.CfnInfrastructureConfigurationProps{
		InstanceProfileName: instanceProfile.Ref(),
		Name:                jsii.String("magento"),
		Logging: &awsimagebuilder.CfnInfrastructureConfiguration_LoggingProperty{
			S3Logs: &awsimagebuilder.CfnInfrastructureConfiguration_S3LogsProperty{
				S3BucketName: bucket.BucketName(),
			},
		},
	})

	awsimagebuilder.NewCfnImagePipeline(stack, jsii.String("ImagePipline"), &awsimagebuilder.CfnImagePipelineProps{
		ImageRecipeArn:                 imageRecipe.AttrArn(),
		InfrastructureConfigurationArn: infrastructureConfiguration.AttrArn(),
		Name:                           jsii.String("magento"),
		Status:                         jsii.String("DISABLED"),
	})

	// example resource
	// queue := awssqs.NewQueue(stack, jsii.String("CdkEc2ImageBuilderMagentoQueue"), &awssqs.QueueProps{
	// 	VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(300)),
	// })

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewCdkEc2ImageBuilderMagentoStack(app, "CdkEc2ImageBuilderMagentoStack", &CdkEc2ImageBuilderMagentoStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}