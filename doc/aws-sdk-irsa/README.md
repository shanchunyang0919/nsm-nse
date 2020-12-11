# IRSA (IAM Role & Service Account) Topology
![alt tag](https://user-images.githubusercontent.com/71080192/101626782-49dbf780-39d2-11eb-967f-477277740043.png)

## Prequesite
Install **kubectl**, **AWS CLI**, and **eksctl**. Make sure we have admin rights to install software.\
\
**Install AWS CLI Reference:** https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html \
**Install kubectl Reference:** https://kubernetes.io/docs/tasks/tools/install-kubectl/ \
**Install eksctl Reference:** https://docs.aws.amazon.com/eks/latest/userguide/getting-started-eksctl.html 

**For MacOSX (Homebrew)**:
```bash
$ brew install eksctl
$ brew install kubectl
$ brew install awscli
```
Configure AWS credentials with awscli (You will need to provide **Access Key ID**, **Secret Key**, **region**, and **format**). \
\
**AWS CLI configuration Reference:** https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html \
\
We can put **us-west-2** as an example for **region** and **json** as an example for **format**.
```bash
$ aws configure
```

## Demo 
**Note**: This is a local testing (not yet for NSE).

Create AWS EKS cluster (the **IRSA** feature works with EKS clusters 1.13 and above).
```bash
$ eksctl create cluster <cluster-name>
```

**IAM OIDC identity providers** are entities in IAM that describe an external identity provider (IdP) service that supports the OpenID Connect (OIDC) standard. You use an IAM OIDC identity provider when you want to establish trust between an OIDC-compatible IdP and your AWS account. <br />
\
**OIDC Reference:** https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html \
\
OIDC federation access allows you to assume IAM roles via the Secure Token Service (STS), enabling authentication with an OIDC provider, receiving a JSON Web Token (JWT), which in turn can be used to assume an IAM role. Kubernetes, on the other hand, can issue so-called projected service account tokens, which happen to be valid OIDC JWTs for pods. \
\
**AWS STS(Security Token Service) Reference:** https://docs.aws.amazon.com/STS/latest/APIReference/welcome.html \
\
In order to activate the **IRSA** feature, we have to associate OIDC provider to our EKS cluster.
```bash
$ eksctl utils associate-iam-oidc-provider --name <cluster-name> --approve
```
For testing purposes, we should create a S3 with **AWS CLI** (make sure the name is not already used by others).
```bash
$ aws s3 mb s3://<bucket-name>
```

Create a service account and **IAM Role** and attach the policy **AmazonS3ReadOnlyAccess** as an example.\
\
(If you are following my sample codes, the service account name will be set to **my-serviceaccount-s3yaml**)
```bash
$ eksctl create iamserviceaccount \
                --name <service-account-name> \
                --namespace default \
                --cluster <cluster-name> \
                --attach-policy-arn arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess \
                --approve
```
### Overriding service account with eksctl (Optional) 
This is important if we already have an existing service accout or if we want to attach different **IAM Role** to the service account.
```bash
$ eksctl create iamserviceaccount \
    --name <service-account-name> \
    --namespace default \
    --cluster <cluster-name> \
    --attach-policy-arn arn:aws:iam::<aws-account-id>:policy/<policy-id> \
    --override-existing-serviceaccounts \
    --approve
```

This is basically what a service-account with ARN annotations looks like.
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::<AWS_ACCOUNT_ID>:role/<IAM_ROLE_NAME>
  ...
```

This line specifies **IAM Role** name.
```yaml
arn:aws:iam::<AWS_ACCOUNT_ID>:role/<IAM_ROLE_NAME>
```


In Helm deployment template spec, we must add the field **serviceAccount** and the corresponding name of it.
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    run: myapp
  name: myapp
spec:
  ...
  template:
    ...
    spec:
      serviceAccount: {{ .Values.serviceName }}-service-account 
      containers:
      - image: myapp:latest
        name: myapp
        ...
``` 

For testing purposes, we can simply apply the sample deployment manifest. This yaml file will deploy a few pods that contain AWS-SDK binary images.
```bash
$ kubectl apply -f <deployment-manifest-file>
```

**Extra**:

List all EKS clusters.
```bash
$ eksctl get cluster
```

List all service accounts.
```bash
$ kubectl get sa
```

**IRSA Studies Reference:** https://aws.amazon.com/tw/blogs/opensource/introducing-fine-grained-iam-roles-service-accounts/

# AWS-SDK for golang

![alt tag](https://user-images.githubusercontent.com/71080192/101692969-d5cb3f00-3a25-11eb-9b16-1a466006526e.png)

**AWS-SDK Guide Reference:** https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/welcome.html

## Prequesite
If we haven't had our AWS credentials setup up yet, we should do it. 
```bash
$ aws configure
```
Install the latest version of Docker. 

**For MacOSX (Homebrew)**:
```bash
$ brew install docker
```
After setting up all the AWS credentials, all the credeientials info will be stored under **.aws/** directory. Now, we can start to run our **AWS-SDK** files. \
\
**AWS-SDK Code Referece:** https://github.com/awsdocs/aws-doc-sdk-examples/tree/master/go/example_code/s3 


## Run AWS-SDK at local machine
If we are going to run it locally, we will have to install all the AWS go dependencies.
```bash
$ go get github.com/aws/aws-sdk-go/aws
$ go get github.com/aws/aws-sdk-go/aws/session
$ go get github.com/aws/aws-sdk-go/service/s3
```
We will see all the S3 buckets listed if we run this API.
```bash
$ go run <aws-sdk-go-file>
```

## Run AWS-SDK at Docker-level
Build an image with a **Dockerfile**.
```bash
$ docker build -t <image-name> .

```
We need to provide AWS credientials in order to run the container (at Docker-level). 
```bash
$ docker run \
        -e AWS_ACCESS_KEY_ID=<iam-user-access-key-id> \
        -e AWS_SECRET_ACCESS_KEY=<iam-user-secret-key> \
        -e AWS_DEFAULT_REGION=<region> \
        <image-name>
```

## Run AWS-SDK at Kubernetes-level

If we are running containers in Kubernetes pods, passing AWS credentials into containers will be problematic because most of the ways are not secure enough. The best way to solve this is **use IAM Role and do not deal with credentials at all**.\
\
**IAM Role Reference:** http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html \
\
We can check out the logs to see if the AWS resources are successfully accessed by the pods.
```bash
$ kubectl logs <pod-name>
```
If you encounter the **CrashLoopBackOff** error, in this case, it is because pods start up then immediately exit after the applications are done, thus Kubernetes restarts and the cycle continues. \
\
To solve this, we could add two fields (**command** and **args**) into our deployment manifest.
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    run: myapp
  name: myapp
spec:
  ...
  template:
    ...
    spec:
      serviceAccount: {{ .Values.serviceName }}-service-account 
      containers:
      - image: myapp:latest
        name: myapp
        command: [ "sleep" ]
        args: [ "infinity" ]
        ...
```
Apply the changes with **kubectl**.
```
$ kubectl apply -f <deployment-manifest-file>
```

**Extra**:

List AWS resources.
```bash
$ aws ec2 describe-instances
$ aws s3 ls
$ aws iam list-users
```
