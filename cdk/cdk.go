package main

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"

	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3deployment"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type CdkStackProps struct {
	awscdk.StackProps
}

func NewCdkStack(scope constructs.Construct, id string, props *CdkStackProps, env string, developer string) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// リソース名に環境名と開発者名を含める
	prefix := fmt.Sprintf("-%s-%s", env, developer)

	// The code that defines your stack goes here

	// example resource
	// queue := awssqs.NewQueue(stack, jsii.String("CdkQueue"), &awssqs.QueueProps{
	// 	VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(300)),
	// })

	// DynamoDB
	table := awsdynamodb.NewTable(stack, jsii.String("AwsCdkTable"), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("PK"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("SK"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		BillingMode:   awsdynamodb.BillingMode_PAY_PER_REQUEST,
		RemovalPolicy: awscdk.RemovalPolicy_RETAIN, // 誤ってcdk destroyした際に削除されないようにする。
	})

	// Lambda
	apiFunction := awslambda.NewFunction(stack, jsii.String("HelloAWSCDKFunction"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("bootstrap"),
		Code: awslambda.Code_FromAsset(jsii.String("../backend"), &awss3assets.AssetOptions{
			Exclude: &[]*string{
				jsii.String("**/*.go"),
				jsii.String("**/go.mod"),
				jsii.String("**/go.sum"),
				jsii.String("**/.gitignore"),
			},
		}),
		// テーブル名はcdk deployのたびに変わる可能性があるため、環境変数に設定する。
		Environment: &map[string]*string{
			"TABLE_NAME": table.TableName(),
		},
	})

	table.GrantReadWriteData(apiFunction)

	// APIGateway (REST API)
	restApi := awsapigateway.NewRestApi(stack, jsii.String("RestApi"), &awsapigateway.RestApiProps{
		RestApiName:   jsii.String("MyAwsCdkApi" + prefix), // コンソール上で見分けをつけるためにprefixをつける
		EndpointTypes: &[]awsapigateway.EndpointType{awsapigateway.EndpointType_REGIONAL},
		// CORS設定の場合は以下を記載
		// DefaultCorsPreflightOptions: &awsapigateway.CorsOptions{
		// 	AllowOrigins: awsapigateway.Cors_ALL_ORIGINS(),
		// 	AllowMethods: awsapigateway.Cors_ALL_METHODS(),
		// 	AllowHeaders: awsapigateway.Cors_DEFAULT_HEADERS(),
		// },
	})

	integration := awsapigateway.NewLambdaIntegration(apiFunction, nil)

	api := restApi.Root().AddResource(jsii.String("api"), nil)
	proxy := api.AddResource(jsii.String("{proxy+}"), nil)
	proxy.AddMethod(jsii.String("ANY"), integration, nil)

	// フロントエンドのプログラムをS3に配置しCloudFrontで公開するための設定
	websiteBucket := awss3.NewBucket(stack, jsii.String("WebsiteBucket"), &awss3.BucketProps{
		RemovalPolicy:     awscdk.RemovalPolicy_RETAIN, // 誤ってcdk destroyした際に削除されないようにする。
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
	})

	// Cloud Front
	distribution := awscloudfront.NewDistribution(stack, jsii.String("Distribution"), &awscloudfront.DistributionProps{
		DefaultBehavior: &awscloudfront.BehaviorOptions{
			Origin:               awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(websiteBucket, nil),
			ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
			CachePolicy:          awscloudfront.CachePolicy_CACHING_OPTIMIZED(),
		},
		// 追加: /api/* → API Gateway
		AdditionalBehaviors: &map[string]*awscloudfront.BehaviorOptions{
			"/api/*": {
				Origin:               awscloudfrontorigins.NewRestApiOrigin(restApi, nil),
				ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
				CachePolicy:          awscloudfront.CachePolicy_CACHING_DISABLED(),
				AllowedMethods:       awscloudfront.AllowedMethods_ALLOW_ALL(),
			},
		},
		DefaultRootObject: jsii.String("index.html"),
		ErrorResponses: &[]*awscloudfront.ErrorResponse{
			{
				HttpStatus:         jsii.Number(404),
				ResponseHttpStatus: jsii.Number(200),
				ResponsePagePath:   jsii.String("/index.html"),
			},
		},
	})

	// S3にフロントエンドのプログラムを配置する。
	awss3deployment.NewBucketDeployment(stack, jsii.String("DeployWebsite"), &awss3deployment.BucketDeploymentProps{
		Sources:           &[]awss3deployment.ISource{awss3deployment.Source_Asset(jsii.String("../frontend/dist"), nil)},
		DestinationBucket: websiteBucket,
		Distribution:      distribution, // デプロイ後にCloudFrontキャッシュを自動でクリア
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	// 環境名と開発者名をcdk.jsonから取得
	envName := getContext(app, "env", "dev")
	developer := getContext(app, "developer", "local")
	// 環境名を組み合わせてスタック名を決定
	stackName := fmt.Sprintf("CdkStack-%s-%s", envName, developer)

	NewCdkStack(app, stackName, &CdkStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	}, envName, developer)

	app.Synth(nil)
}

func getContext(app awscdk.App, key, defaultVal string) string {
	if val := app.Node().TryGetContext(jsii.String(key)); val != nil {
		return val.(string)
	}
	return defaultVal
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
