# helm-chart-learinging

RBAC of secrets for pods access.

POC for accessing the external secrets in AWS secret manager with Kubernetes and below are steps and finding-

https://github.com/aws-samples/aws-secret-sidecar-injector
https://aws.amazon.com/blogs/containers/aws-secrets-controller-poc/

Use case for POC: Deploy mock web server that need a certificate to access the API. And that certificate is stored as secret the AWS Manager.
Actually secret is retrieved by an init container that get injected into the pod by our mutating webhook.

Steps to create use external secret .

Pre-req
AWS secret admission controller required for the running admission controller and mutating webhook configurations.

1. Required secret-injection admission controller webhook. (Can be added through helm repository)
   helm repo add secret-inject https://aws-samples.github.io/aws-secret-sidecar-injector

2. Install AWS secret Controller
   helm install secret-inject/secret-inject

Step : -

    1. Create a secret (with certificate data) in AWS Secret Manager.
    	check - aws secretsmanager list-secrets   — note down the Arn of created secret.

    2.  Create an IAM policy and AWS role to access the secret in  AWS Secret Manger

    	a. 	Policy-
    		{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "secretsmanager:GetResourcePolicy",
                "secretsmanager:GetSecretValue",
                "secretsmanager:DescribeSecret",
                "secretsmanager:ListSecretVersionIds"
            ],
            "Resource": [
                "arn:aws:secretsmanager:us-west-2:972956032059:secret:test-secret-YVUgzu",
            ]
        },
        {
            "Sid": "VisualEditor1",
            "Effect": "Allow",
            "Action": [
                "secretsmanager:GetRandomPassword",
                "secretsmanager:ListSecrets"
            ],
            "Resource": "*"
        }
    ]

}

    	b.  	we’ve created a policy, we need to create an IAM role, which our pod will assume. We have to use IAM Roles for Service Accounts (IRSA) to enable fine grained access to read the 				secrets from AWS Secrets Manager.

    	c.	Create  trust policy for the role so the OIDC federate user can assume the role.
    		{

"Version": "2012-10-17",
"Statement": [
{
"Effect": "Allow",
"Principal": {
"Federated": "arn:aws:iam::972956032059:oidc-provider/oidc-c2-us-west-2-aws-dp-waf-dev-f5-cs-com.s3.us-west-2.amazonaws.com"
},
"Action": "sts:AssumeRoleWithWebIdentity",
"Condition": {
"StringEquals": {
"oidc-c2-us-west-2-aws-dp-waf-dev-f5-cs-com.s3.us-west-2.amazonaws.com:sub": "system:serviceaccount:nisingh:webserver-service-account",
"oidc-c2-us-west-2-aws-dp-waf-dev-f5-cs-com.s3.us-west-2.amazonaws.com:aud": "sts.amazonaws.com"
}
}
}]
}
d. Attach the IAM policy created earlier to the create IAM role.

    3. 	Create a service account and map it to our role. Add an annotation for the service account that references the IAM role we created earlier. In our case, we’re using the ARN of the role we created earlier.

apiVersion: v1
kind: ServiceAccount
metadata:
annotations:
iam.amazonaws.com/role-arn: arn:aws:iam::972956032059:role/web-secretmanager-role
creationTimestamp: "2020-09-03T05:21:42Z"
name: webserver-service-account
namespace: nisingh
resourceVersion: "90849889"
selfLink: /api/v1/namespaces/nisingh/serviceaccounts/webserver-service-account
uid: 5a061b70-eda5-11ea-ba0a-12c812ad84b7
secrets:

- name: webserver-service-account-token-5lkbb

      4.   Now lets create a pod that need to use/access the secret with the annotations and service account created above are specified in the pod specification

  apiVersion: apps/v1
  kind: Deployment
  metadata:
  namespace: nisingh
  labels:
  run: webapp
  name: webapp
  spec:
  replicas: 1
  selector:
  matchLabels:
  run: webapp
  template:
  metadata:
  annotations:
  secrets.k8s.aws/sidecarInjectorWebhook: enabled
  secrets.k8s.aws/secret-arn: arn:aws:secretsmanager:us-west-2:972956032059:secret:test-secret-YVUgzu
  labels:
  run: webapp
  spec:
  serviceAccountName: webserver-service-account
  containers: - image: busybox:1.28
  name: webapp
  command: ["sh", "-c", "echo $(cat /tmp/secret) && sleep 3600"]

      5. 	In above running pod, the mutating webhook creates a emptyDir volume secret-vol and mounts it at /tmp/ location for all the containers in the pod. The decrypted secret is written to /tmp/secret.

When a pod with the requisite annotations is deployed to the cluster, the webhook will update the pod to run the init container.
Provided the ServiceAccount specified in the podSpec has access to the secret referenced in the “secrets.k8s.aws/secret-arn: dummy-secret” the init container will retrieve the secret from Secrets Manager and write it to a RAM disk. This is done by specifying using Memory as the medium for an emptyDir volume. This prevents the secrets from being persisted to disk after pod is terminated.

After the init container exits, the application container starts and mounts the RAM disk as a volume. When the application needs to read the secret, it reads it from the mounted volume.

Finding -

1. Polices can used to restrict and allow the fine grain access.
2. Even AWS secret manager having resource policies to restrict the access/permission for the particular user/IAM role
3. Godaddy’s Kubernetes External Secret works on same principle of IAM Roles for Service Accounts with fine grain IAM policy and created an extra object of Kind: External secret. Only advantage is it can integrate with any Service provider like AWS secret manager, Azure Key Vault, GCP, HashiCorp Vault etc.

Other oprion explored
https://github.com/doitintl/kube-secrets-init
https://blog.doit-intl.com/kubernetes-and-secrets-management-in-cloud-858533c20dca
