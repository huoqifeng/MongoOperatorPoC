## Steps to create similar project:
a. Clone and build Operator-SDK
    https://github.com/operator-framework/operator-sdk
b. new an operator project
    https://github.com/operator-framework/getting-started
    operator-sdk new MongoOperatorPoC
	cd MongoOperatorPoC
c. add api (CRD, resource type)
	operator-sdk add api --api-version=dbaas.k8s.ibm.com/v1alpha1 --kind=MongoDB
d. update the generated code for that resource after *_types.go updated	
	operator-sdk generate k8s
e. add a new controller
    operator-sdk add controller --api-version=dbaas.k8s.ibm.com/v1alpha1 --kind=MongoDB
f. build and run the operator
    f1. create crd in kubernetes cluster
	    kubectl create -f deploy/crds/dbaas_v1alpha1_mongodb_crd.yaml
	f2. run inside k8s as deployment (skip)
	f3. run locally outside kubernetes cluster
	    export OPERATOR_NAME=MongoOperatorPoC
		operator-sdk up local --namespace=default
g. create a cr
    kubectl apply -f deploy/crds/dbaas_v1alpha1_mongodb_cr.yaml
h. (todo) add more operations in controller.
    https://github.com/operator-framework/operator-sdk/blob/master/doc/user/client.md	
i. (todo) make operator robust (Operator Lifecycle Manager)	
    template: https://github.com/operator-framework/getting-started/blob/master/memcachedoperator.0.0.1.csv.yaml