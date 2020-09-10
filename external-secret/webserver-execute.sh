cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: nisingh
  labels:
    run: webserver
  name: webserver
spec:
  replicas: 1
  selector:
    matchLabels:
      run: webserver
  template:
    metadata:
      annotations:
        secrets.k8s.aws/sidecarInjectorWebhook: enabled
        secrets.k8s.aws/secret-arn: arn:aws:secretsmanager:us-west-2:972956032059:secret:test-secret-YVUgzu
      labels:
        run: webserver
    spec:
      serviceAccountName: webserver-service-account
      containers:
      - image: busybox:1.28
        name: webserver
        command: ['sh', '-c', 'echo $(cat /tmp/secret) && sleep 3600']
EOF
