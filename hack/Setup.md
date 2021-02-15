# Local Setup

Please follow the following instruction to setup the necessary E2E Testing environment:
> Assumes Helm + kubectl are correctly setup & configured

1. Get [K3D](https://k3d.io/) following [this instructions](https://github.com/rancher/k3d#get).
2. Install [arkade](https://github.com/alexellis/arkade), **Important** for _windows user_ is to leverage Git Bash or another Bash.
3. Start a new k8s cluster `$ k3d cluster create openfaas`.
4. Install OpenFaaS to newly created Cluster `$ ark install openfaas --basic-auth=false`, make sure to update your kubectl config prior to use the k3d cluster.
5. Add Helm Repo for Rabbit MQ `$ helm repo add bitnami https://charts.bitnami.com/bitnami` & `$ helm repo update`.
6. Deploy RabbitMQ using helm `$ helm install rabbit -f ./hack/rabbit-values.yaml bitnami/rabbitmq`.
7. Now you can forward the relevant ports from openfaas (8080) & rabbitmq (5672).
8. Deploy Function listening under the relevant Topics `$ faas-cli store deploy figlet --annotation topic="Foo,Bar,Dead,Beef"`
9. You can use the HTTP Api exposed via 15672 to publish messages
