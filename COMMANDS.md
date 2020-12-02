
## Launch a Bottlerocket Node

```sh
export CLUSTER_NAME=grog && kube:env $CLUSTER_NAME
export BOTTLEROCKET_VARIANT=aws-k8s-1.16
export ARCH=x86_64
source $REPOS/mybr/scripts/tom.sh
export AMI_ID=$(aws ssm get-parameter \
  --region "us-west-2" \
  --name "/aws/service/bottlerocket/${BOTTLEROCKET_VARIANT}/${ARCH}/latest/image_id" \
  --query Parameter.Value --output text)
launch $CLUSTER_NAME $AMI_ID
export INSTANCE_ID=i-090fbb5b83ba10ced
export IP_ADDRESS=$(aws ec2 describe-instances --filters Name=instance-id,Values=$INSTANCE_ID \
  --query "Reservations[*].Instances[*].PublicIpAddress" \
  --output=text)
```

## Deploy Daemonset

```sh
export CLUSTER_NAME=grog && kube:env $CLUSTER_NAME
export CSI=$REPOS/aws-efs-csi-driver
cd $CSI/deploy/kubernetes/overlays/stable/ecr
kubectl apply -k $CSI/deploy/kubernetes/overlays/stable/ecr

```

## Delete Daemonset

```sh
kubectl delete -k $CSI/deploy/kubernetes/overlays/stable/ecr
```

## Prep to deploy Example App

```sh
export CLUSTER_NAME=grog && kube:env $CLUSTER_NAME
export VPC_ID=$(aws eks describe-cluster --name $CLUSTER_NAME --query "cluster.resourcesVpcConfig.vpcId" --output text)
export CIDR=$(aws ec2 describe-vpcs --vpc-ids $VPC_ID --query "Vpcs[].CidrBlock" --output text)
export FS_ID=$(aws efs describe-file-systems --query "FileSystems[*].FileSystemId" --output text)
cd $REPOS
export EXAMPLE_REPO=$REPOS/aws-efs-csi-example
rm -rf $EXAMPLE_REPO && git clone https://github.com/kubernetes-sigs/aws-efs-csi-driver.git $EXAMPLE_REPO
cd $EXAMPLE_REPO/examples/kubernetes/multiple_pods/
sed -i.backup "s/fs-4af69aab/$FS_ID/" specs/pv.yaml
rm specs/pv.yaml.backup
```


arch=x86_64
variant=aws-k8s-1.18
ami_id=$(aws ssm get-parameter \
  --region "us-west-2" \
  --name "/aws/service/bottlerocket/${variant}/${arch}/latest/image_id" \
  --query Parameter.Value --output text)

persistentvolumeclaim/efs-claim unchanged
pod/app1 created
kubectl delete pod/app2 created
kubectl delete persistentvolume/efs-pv unchanged
storageclass.storage.k8s.io/efs-sc unchanged