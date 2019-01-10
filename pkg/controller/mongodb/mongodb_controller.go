package mongodb

import (
	dbaasv1alpha1 "github.com/huoqifeng/MongoOperatorPoC/pkg/apis/dbaas/v1alpha1"
	"context"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_mongodb")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new MongoDB Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMongoDB{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("mongodb-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource MongoDB
	err = c.Watch(&source.Kind{Type: &dbaasv1alpha1.MongoDB{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Deployment and requeue the owner MongoDB
	// err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &dbaasv1alpha1.MongoDB{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileMongoDB{}

// ReconcileMongoDB reconciles a MongoDB object
type ReconcileMongoDB struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

func (r *ReconcileMongoDB) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MongoDB.")

	// Fetch the MongoDB instance
	instance := &dbaasv1alpha1.MongoDB{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("MongoDB resource not found. Ignoring since object must be deleted.")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get MongoDB.")
		return reconcile.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new StatefulSet
		dep := r.statefulsetForMongoDB(instance)
		reqLogger.Info("Creating a new StatefulSet.", "StatefulSet.Namespace", dep.Namespace, "StatefulSet.Name", dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			reqLogger.Error(err, "Failed to create new StatefulSet.", "StatefulSet.Namespace", dep.Namespace, "StatfulSet.Name", dep.Name)
			return reconcile.Result{}, err
		}
		// StatefulSet created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get StatefulSet.")
		return reconcile.Result{}, err
	}

	// Ensure the StatefulSet size is the same as the spec
	size := instance.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.client.Update(context.TODO(), found)
		if err != nil {
			reqLogger.Error(err, "Failed to update StatefulSet.", "StatefulSet.Namespace", found.Namespace, "StatefulSet.Name", found.Name)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}

	// Update the MongoDB status with the pod names
	// List the pods for this mongodb's StatefulSet
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(labelsForMongoDB(instance.Name))
	listOps := &client.ListOptions{
		Namespace:     instance.Namespace,
		LabelSelector: labelSelector,
	}
	err = r.client.List(context.TODO(), listOps, podList)
	if err != nil {
		reqLogger.Error(err, "Failed to list pods.", "MongoDB.Namespace", instance.Namespace, "MongoDB.Name", instance.Name)
		return reconcile.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, instance.Status.Nodes) {
		instance.Status.Nodes = podNames
		err := r.client.Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update MongoDB status.")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// statefulsetForMomngoDB returns a mongodb StatefulSet object
func (r *ReconcileMongoDB) statefulsetForMongoDB(m *dbaasv1alpha1.MongoDB) *appsv1.StatefulSet {
	ls := labelsForMongoDB(m.Name)
	replicas := m.Spec.Size

    // only hostPath works for minikube so far.
	//const pvname string = "example-local-claim"
	//const storageclass string = "local-storage"
	const pvname string = "pv0001"
	const storageclass string = "standard"

	rc := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceStorage: resource.MustParse("0.5Gi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceStorage: resource.MustParse("0.5Gi"),
		},
	}
	
	stateset := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   "busybox",
						Name:    "busybox",
						Command: []string{"sleep", "3600"},
						// add
        				VolumeMounts: []corev1.VolumeMount{{
        					Name: pvname, 
        					MountPath: "/opt/data", 
        					ReadOnly: false,
					    }},
					}},
				},
			},
			// add
    		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				newPersistentVolumeClaim(m, rc, pvname, storageclass),
		    },
		},
	}
	// Set MongoDB instance as the owner and controller
	controllerutil.SetControllerReference(m, stateset, r.scheme)
	return stateset
}

// newPersistentVolumeClaim returns a Persistent Volume Claims for Mongod pod
func newPersistentVolumeClaim(m *dbaasv1alpha1.MongoDB, resources corev1.ResourceRequirements, pvname, storageClass string) corev1.PersistentVolumeClaim {
	vc := corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvname,
			Namespace: m.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resources.Requests[corev1.ResourceStorage],
				},
			},
		},
	}
	if storageClass != "" {
		vc.Spec.StorageClassName = &storageClass
	}
	return vc
}

// labelsForMongoDB returns the labels for selecting the resources
// belonging to the given mongodb CR name.
func labelsForMongoDB(name string) map[string]string {
	return map[string]string{"app": "mongodb", "mongodb_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
