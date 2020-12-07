
## Launch a Bottlerocket Node

```sh
export CLUSTER_NAME=water && kube:env $CLUSTER_NAME
export BOTTLEROCKET_VARIANT=aws-k8s-1.18
export ARCH=x86_64
source $REPOS/mybr/scripts/tom.sh
export AMI_ID=$(aws ssm get-parameter \
  --region "us-west-2" \
  --name "/aws/service/bottlerocket/${BOTTLEROCKET_VARIANT}/${ARCH}/latest/image_id" \
  --query Parameter.Value --output text)
launch_file=/tmp/$(uuidgen).json
launch $CLUSTER_NAME $AMI_ID | tee "${launch_file}"
cat "${launch_file}" | sed -e '1,2d' > "${launch_file}"
export INSTANCE_ID=$(cat "${launch_file}" | jq -r '.Instances | .[0] | .InstanceId')
export IP_ADDRESS=$(aws ec2 describe-instances --filters Name=instance-id,Values=$INSTANCE_ID \
  --query "Reservations[*].Instances[*].PublicIpAddress" \
  --output=text)
```

## Deploy Daemonset

```sh
export CLUSTER_NAME=water && kube:env $CLUSTER_NAME
export CSI=$REPOS/aws-efs-csi-driver
cd $CSI/deploy/kubernetes/overlays/stable/ecr
kubectl apply -k $CSI/deploy/kubernetes/overlays/stable/ecr
```

## Various Commands

```sh
kubectl get pods -n kube-system
export CSI_POD=$(kubectl get pods -n kube-system | grep efs-csi-node | blah not worth it)
kubectl describe pod efs-csi-node-w57j7 -n kube-system
kubectl delete -k $CSI/deploy/kubernetes/overlays/stable/ecr
```

#### Notes from a test run
```text
efs-csi-node-jw2cf -> 192.168.43.17 -> Bottlerocket
efs-csi-node-xfl6m -> 192.168.58.240 -> AL2
```

## Example App

Create a filesystem following [these steps](https://docs.aws.amazon.com/eks/latest/userguide/efs-csi.html)

Then...

```sh
export CLUSTER_NAME=water && kube:env $CLUSTER_NAME
export VPC_ID=$(aws eks describe-cluster --name $CLUSTER_NAME --query "cluster.resourcesVpcConfig.vpcId" --output text)
export CIDR=$(aws ec2 describe-vpcs --vpc-ids $VPC_ID --query "Vpcs[].CidrBlock" --output text)
aws ec2 create-security-group \
  --description $CLUSTER_NAME-efs-sg \
  --group-name $CLUSTER_NAME-efs-sg \
  --vpc-id $VPC_ID \
  --output text
export FS_SG_ID=$(aws ec2 describe-security-groups \
  --filters "Name=group-name,Values=$CLUSTER_NAME-efs-sg" \
  --output json | jq -r '.SecurityGroups | .[0] | .GroupId')
aws ec2 authorize-security-group-ingress \
  --group-id "${FS_SG_ID}" \
  --protocol "all" \
  --port "2049" \
  --cidr "${CIDR}"
export FS_ID=$(aws efs describe-file-systems \
  --output json | \
  jq -r ".FileSystems[] | select(.Name==\"${CLUSTER_NAME}-efs\").FileSystemId")
cd $REPOS
export EXAMPLE_REPO=$REPOS/aws-efs-csi-example
rm -rf $EXAMPLE_REPO && git clone https://github.com/kubernetes-sigs/aws-efs-csi-driver.git $EXAMPLE_REPO
cd $EXAMPLE_REPO/examples/kubernetes/multiple_pods/
sed -i.backup "s/fs-4af69aab/$FS_ID/" specs/pv.yaml
rm specs/pv.yaml.backup
cat $EXAMPLE_REPO/examples/kubernetes/multiple_pods/specs/pv.yaml
kubectl apply -f $EXAMPLE_REPO/examples/kubernetes/multiple_pods/specs
kubectl get pv
kubectl describe pv efs-pv
kubectl get pods --watch
kubectl describe pod app1 | grep Node && kubectl get nodes
kubectl describe pod app2 | grep Node && kubectl get nodes
kubectl exec -ti app1 -- tail /data/out1.txt
kubectl exec -ti app2 -- tail /data/out1.txt
kubectl describe pod app1
```

Delete it

```sh
kubectl delete -f $EXAMPLE_REPO/examples/kubernetes/multiple_pods/specs
```

## Retrieve Logdog

```sh
rm -rf "${TMP}/dog" && mkdir -p "${TMP}/dog"
cd "${TMP}/dog"
ssh -i "~/.ssh/brigmatt-key.pem" "ec2-user@${IP_ADDRESS}" \
  "cat /.bottlerocket/rootfs/tmp/bottlerocket-logs.tar.gz" \
   > "bottlerocket-logs.tar.gz"
tar xvf "bottlerocket-logs.tar.gz"
rm -f "bottlerocket-logs.tar.gz"
cd $CSI
subl "${TMP}/dog/bottlerocket-logs"
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

## Deploying the 'Old' Version

```sh
kubectl apply -k "github.com/kubernetes-sigs/aws-efs-csi-driver/deploy/kubernetes/overlays/stable/ecr/?ref=release-1.0"
```