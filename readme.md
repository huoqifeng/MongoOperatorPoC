This is just a k8s operator sample which uses busybox image, no real implementation for mongodb or any other database...

## Steps to create this project:
- Clone and build Operator-SDK  
    https://github.com/operator-framework/operator-sdk  
- new an operator project  
    https://github.com/operator-framework/getting-started  
    `operator-sdk new MongoOperatorPoC`    
	 `cd MongoOperatorPoC`
- add api (CRD, resource type)  
	`operator-sdk add api --api-version=dbaas.k8s.ibm.com/v1alpha1 --kind=MongoDB`
- update the generated code for that resource after *_types.go updated  
    `operator-sdk generate k8s`
- add a new controller  
    `operator-sdk add controller --api-version=dbaas.k8s.ibm.com/v1alpha1 --kind=MongoDB`
- build and run the operator  
   1. create crd in kubernetes cluster  
	    `kubectl create -f deploy/crds/dbaas_v1alpha1_mongodb_crd.yaml`
	2. run inside k8s as deployment (skip)  
	3. run locally outside kubernetes cluster  
    `export OPERATOR_NAME=MongoOperatorPoC`  
	`operator-sdk up local --namespace=default`  
- create a cr  
    `kubectl apply -f deploy/crds/dbaas_v1alpha1_mongodb_cr.yaml`
- (Todo) add more operations in controller.  
    https://github.com/operator-framework/operator-sdk/blob/master/doc/user/client.md	
- (Todo) make operator robust (Operator Lifecycle Manager)  
    template: https://github.com/operator-framework/getting-started/blob/master/memcachedoperator.0.0.1.csv.yaml


## Steps to test the mongodb replicasets
- Install minikube
- Start the operator  
`export OPERATOR_NAME=MongoOperatorPoC`  
`operator-sdk up local --namespace=default`  
- Create pod reader role and role-bind for mongod sidecar  
`kubectl create -f deploy/poc/pods-reader-role-rolebind.yaml`
- Create PV firstly.  
`kubectl create -f deploy/poc/host-path-pv.yaml`
- Deploy MongoDB cr  
`kubectl create -f deploy/crds/dbaas_v1alpha_mongodb_cr.yaml`
- Verify the resource created  
`kubectl get MongoDB`  
`kubectl get StatefulSet`  
`kubectl get pvc`  
`kubectl get po -w`
- Login to one of the mongo pods  
`kubectl exec -it mongodb-0 /bin/sh`
- Connect to mongo replica within the mongo pod  
`mongo mongodb://mongodb-0.mongo:27017,mongodb-1.mongo:27017,mongodb-2.mongo:27017/?replicaSet=rs0`

If you ge t error like below, may be caused by the dns not work well in minikube.  
```
2019-01-11T03:19:27.719+0000 I NETWORK  [thread1]   getaddrinfo("mongodb-2.mongo") failed: Name or service not known. 
2019-01-11T03:19:27.721+0000 I NETWORK  [thread1] getaddrinfo("mongodb-0.mongo") failed: Name or service not known
2019-01-11T03:19:27.724+0000 I NETWORK  [thread1] getaddrinfo("mongodb-1.mongo") failed: Name or service not known
2019-01-11T03:19:27.724+0000 W NETWORK  [thread1] No primary detected for set rs0
2019-01-11T03:19:27.724+0000 I NETWORK  [thread1] All nodes for set rs0 are down. This has happened for 3 checks in a row.
```
Check with command:  
`minikube addons list`
If DNS is not enabled and you can not solve it, You can then use the IP to verify, to get the IPs, run  
`kubectl logs -f mongodb-0 -c mongo-sidecar`  
And then compose the mongo url with the IPs and then connect via mongo shell, like below:
`mongo mongodb://172.17.0.6:27017,172.17.0.7:27017,172.17.0.8:27017/?replicaSet=rs0`

You can also access the replicaset on your host machine after add minikube route:  
`sudo route -n add 172.17.0.0/16 $(minikube ip)`  
`sudo route -n add 10.0.0.0/24 $(minikube ip)` 

Or, use command  
`sudo route -n add 172.17.0.0 $(minikube ip)`  
`sudo route -n add 10.0.0.0 $(minikube ip)` 

## Dependencies:
This Test depends on the MongoDB sidecar which will maintain the mongo replicasets
https://github.com/cvallance/mongo-k8s-sidecar You may also replace the sidecar with your own.

The sidecar will watch the pods list and init the replicasets, similar as the mongo shell as below:
```
mongo
rs.initiate();
var cfg = rs.conf();
cfg.members[0].host="mongo‑0.mongo:27017";
rs.reconfig(cfg);
rs.add("mongo‑1.mongo:27017");
rs.add("mongo‑2.mongo:27017");
rs.status();
```

## Other mongo k8s operators:

- https://github.com/Percona-Lab/percona-server-mongodb-operator (Apache)
- https://github.com/mongodb/mongodb-enterprise-kubernetes (Official, Commercial)
- https://github.com/kbst/mongodb (replicaSets only, python, 35 stars)
- https://github.com/Ultimaker/k8s-mongo-operator (GPL, python, 9 stars)
