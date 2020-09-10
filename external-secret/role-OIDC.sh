AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)

OIDC_PROVIDER=$(aws eks describe-cluster --name "cluster.name" --query "cluster.identity.oidc.issuer" --output text | sed -e "s/^https:\/\///")

read -r -d '' TRUST_RELATIONSHIP <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::${AWS_ACCOUNT_ID}:oidc-provider/${OIDC_PROVIDER}"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "${OIDC_PROVIDER}:sub": "system:serviceaccount:default:webserver-service-account",
          "${OIDC_PROVIDER}:aud": "sts.amazonaws.com"
        }
      }
    }
  ]
}
EOF
echo "${TRUST_RELATIONSHIP}" >trust.json

aws iam create-role --role-name webserver-secrets-role --assume-role-policy-document file://trust.json --description "IAM Role to access webserver secret"

aws iam attach-role-policy --role-name webserver-secrets-role --policy-arn=arn:aws:iam::123456789012:policy/webserver-secret-policy
