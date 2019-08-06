import cdk = require("@aws-cdk/core");
import sam = require('@aws-cdk/aws-sam');

class S2TStack extends cdk.Stack {
  public TelegramToken: string;

}

interface sns2TelegramStackProps extends cdk.StackProps {
  TelegramToken: string;
}

export class CdkS2TStack extends S2TStack {
  constructor(scope: cdk.App, id: string, props: sns2TelegramStackProps) {
    super(scope, id, props);

    const sns2telegramApp = new sam.CfnApplication(this, 'sns2telegram', {
      location: {
        applicationId: 'arn:aws:serverlessrepo:us-east-1:903779448426:applications/sns2telegram',
        semanticVersion: '1.0.2',
      },
      parameters: {
        TelegramToken: props.TelegramToken,
      }
    })
    new cdk.CfnOutput(this, 'BaseURL', {
      value: sns2telegramApp.getAtt('Outputs.BaseURL').toString(),
    })
  
    new cdk.CfnOutput(this, 'DynamoDBTableName', {
      value: sns2telegramApp.getAtt('Outputs.DynamoDBTableName').toString(),
    })

    new cdk.CfnOutput(this, 'HealthzURL', {
      value: sns2telegramApp.getAtt('Outputs.HealthzURL').toString(),
    })

    new cdk.CfnOutput(this, 'WebHookURL', {
      value: sns2telegramApp.getAtt('Outputs.WebHookURL').toString(),
    })
  }

}
