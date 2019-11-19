{
  build(S3Region, S3Bucket, S3Prefix='', SecretArn, LambdaRoleArn='', Tags={},):: {
    local LambdaTags = (if std.length(Tags) > 0 then { Tags: Tags } else {}),

    AWSTemplateFormatVersion: '2010-09-09',
    Transform: 'AWS::Serverless-2016-10-31',
    Description: 'https://github.com/m-mizutani/gsuite-log-exporter',
    Resources: {
      Main: {
        Type: 'AWS::Serverless::Function',
        Properties: {
          Runtime: 'go1.x',
          CodeUri: 'build',
          Handler: 'main',
          Timeout: 900,
          Role: (if LambdaRoleArn != '' then LambdaRoleArn else { Ref: 'LambdaRole' }),
          MemorySize: 256,
          Environment: {
            Variables: {
              SECRET_ARN: SecretArn,
              S3_REGION: S3Region,
              S3_BUCKET: S3Bucket,
              S3_PREFIX: S3Prefix,
            },
          },
          Events: {
            Every5min: {
              Type: 'Schedule',
              Properties: {
                Schedule: 'rate(5 minutes)',
              },
            },
          },
        },
      } + LambdaTags,
    } + (
      if LambdaRoleArn != '' then {} else {
        LambdaRole: {
          Type: 'AWS::IAM::Role',
          Condition: 'LambdaRoleRequired',
          Properties: {
            AssumeRolePolicyDocument: {
              Version: '2012-10-17',
              Statement: [
                {
                  Effect: 'Allow',
                  Principal: {
                    Service: [
                      'lambda.amazonaws.com',
                    ],
                  },
                  Action: [
                    'sts:AssumeRole',
                  ],
                },
              ],
            },
            Path: '/',
            ManagedPolicyArns: [
              'arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole',
            ],
            Policies: [
              {
                PolicyName: 'S3Writable',
                PolicyDocument: {
                  Version: '2012-10-17',
                  Statement: [
                    {
                      Effect: 'Allow',
                      Action: [
                        's3:PutObject',
                        's3:GetObject',
                        's3:ListBucket',
                      ],
                      Resource: [
                        {
                          'Fn::Sub': [
                            'arn:aws:s3:::${bucket}/${prefix}*',
                            {
                              bucket: {
                                Ref: 'S3Bucket',
                              },
                              prefix: {
                                Ref: 'S3Prefix',
                              },
                            },
                          ],
                        },
                        {
                          'Fn::Sub': [
                            'arn:aws:s3:::${bucket}',
                            {
                              bucket: {
                                Ref: 'S3Bucket',
                              },
                            },
                          ],
                        },
                      ],
                    },
                  ],
                },
              },
              {
                PolicyName: 'ReadSecret',
                PolicyDocument: {
                  Version: '2012-10-17',
                  Statement: [
                    {
                      Effect: 'Allow',
                      Action: [
                        'secretsmanager:GetSecretValue',
                      ],
                      Resource: [
                        {
                          Ref: 'SecretArn',
                        },
                      ],
                    },
                  ],
                },
              },
            ],
          },
        },
      }
    ),
  },
}
