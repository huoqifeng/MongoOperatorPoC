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
- (todo) add more operations in controller.  
    https://github.com/operator-framework/operator-sdk/blob/master/doc/user/client.md	
- (todo) make operator robust (Operator Lifecycle Manager)  
    template: https://github.com/operator-framework/getting-started/blob/master/memcachedoperator.0.0.1.csv.yaml