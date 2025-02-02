/*
Copyright 2018 Anevia.
*/

package cdncluster

import (
	"testing"
	"time"

	clusterv1 "github.com/feloy/operator/pkg/apis/cluster/v1"
	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var c client.Client

var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo", Namespace: "default"}}
var depKey = types.NamespacedName{Name: "foo-deployment", Namespace: "default"}

const timeout = time.Second * 5

func TestReconcile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	instance := &clusterv1.CdnCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: clusterv1.CdnClusterSpec{
			Sources: []clusterv1.CdnClusterSource{},
		},
	}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mgr.GetClient()

	recFn, requests := SetupTestReconcile(newReconciler(mgr, record.NewFakeRecorder(1024)))
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())
	defer close(StartTestManager(mgr, g))

	// Create the CdnCluster object and expect the Reconcile and Deployment to be created
	err = c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	deploy := &appsv1.Deployment{}
	g.Eventually(func() error { return c.Get(context.TODO(), depKey, deploy) }, timeout).
		Should(gomega.Succeed())

	// Delete the Deployment and expect Reconcile to be called for Deployment deletion
	g.Expect(c.Delete(context.TODO(), deploy)).NotTo(gomega.HaveOccurred())
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(func() error { return c.Get(context.TODO(), depKey, deploy) }, timeout).
		Should(gomega.Succeed())

	// Manually delete Deployment since GC isn't enabled in the test control plane
	g.Expect(c.Delete(context.TODO(), deploy)).To(gomega.Succeed())

}

func TestReconcile2(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Setup the Manager and Controller.
	// Wrap the Controller Reconcile function
	// so it writes each request to a channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c := mgr.GetClient()

	recFn, requests := SetupTestReconcile(newReconciler(mgr, record.NewFakeRecorder(1024)))
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())
	defer close(StartTestManager(mgr, g))

	instance := &clusterv1.CdnCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo2",
			Namespace: "default",
		},
		Spec: clusterv1.CdnClusterSpec{
			Sources: []clusterv1.CdnClusterSource{},
		},
	}

	// Create the CdnCluster object
	// and expect the Reconcile to be called
	// with the instance namespace and name as parameter
	err = c.Create(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)

	var expectedRequest = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "foo2",
			Namespace: "default",
		},
	}
	const timeout = time.Second * 5

	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	// Expect that a Deployment is created
	deploy := &appsv1.Deployment{}
	var depKey = types.NamespacedName{
		Name:      "foo2-deployment",
		Namespace: "default",
	}
	g.Eventually(func() error {
		return c.Get(context.TODO(), depKey, deploy)
	}, timeout).Should(gomega.Succeed())

	// Delete the Deployment and expect Reconcile
	// to be called for Deployment deletion
	// and Deployment to be created again
	g.Expect(c.Delete(context.TODO(), deploy)).NotTo(gomega.HaveOccurred())
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(func() error {
		return c.Get(context.TODO(), depKey, deploy)
	}, timeout).Should(gomega.Succeed())
}
func TestReconcileCreatedAfterSource(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Setup the Manager and Controller.
	// Wrap the Controller Reconcile function
	// so it writes each request to a channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c := mgr.GetClient()
	recFn, requests := SetupTestReconcile(newReconciler(mgr, record.NewFakeRecorder(1024)))
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())
	defer close(StartTestManager(mgr, g))

	// Create the CdnCluster object
	// and expect the Reconcile to be called
	// with the instance namespace and name as parameter
	instanceParent := &clusterv1.CdnCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo3",
			Namespace: "default",
		},
		Spec: clusterv1.CdnClusterSpec{
			Sources: []clusterv1.CdnClusterSource{
				{
					Name:          "asource",
					PathCondition: "/live/",
				},
			},
		},
	}
	err = c.Create(context.TODO(), instanceParent)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instanceParent)
	var expectedRequest = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "foo3",
			Namespace: "default",
		},
	}
	const timeout = time.Second * 5
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	// Expect that a Deployment is not created
	deploy := &appsv1.Deployment{}
	var depKey = types.NamespacedName{
		Name:      "foo3-deployment",
		Namespace: "default",
	}
	g.Eventually(func() error {
		return c.Get(context.TODO(), depKey, deploy)
	}, timeout).ShouldNot(gomega.Succeed())

	// Create the CdnCluster object
	// and expect the Reconcile to be called
	// with the instance namespace and name as parameter
	instanceSource := &clusterv1.CdnCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "asource",
			Namespace: "default",
		},
		Spec: clusterv1.CdnClusterSpec{
			Sources: []clusterv1.CdnClusterSource{},
		},
	}
	err = c.Create(context.TODO(), instanceSource)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instanceSource)
	var expectedRequestSource = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "asource",
			Namespace: "default",
		},
	}
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequestSource)))

	// Expect that a Deployment is created
	deploy = &appsv1.Deployment{}
	depKey = types.NamespacedName{
		Name:      "asource-deployment",
		Namespace: "default",
	}
	g.Eventually(func() error {
		return c.Get(context.TODO(), depKey, deploy)
	}, timeout).Should(gomega.Succeed())

	// Expect the Reconcile function to be called for the parent cluster
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	// Expect that a Deployment is created
	deploy = &appsv1.Deployment{}
	depKey = types.NamespacedName{
		Name:      "foo3-deployment",
		Namespace: "default",
	}
	g.Eventually(func() error {
		return c.Get(context.TODO(), depKey, deploy)
	}, timeout).Should(gomega.Succeed())
}
