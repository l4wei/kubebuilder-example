/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"github.com/go-logr/logr"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	devopsv1 "my.domain/example/api/v1"
)

// DemoMicroServiceReconciler reconciles a DemoMicroService object
type DemoMicroServiceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=devops.my.domain,resources=demomicroservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=devops.my.domain,resources=demomicroservices/status,verbs=get;update;patch

func (r *DemoMicroServiceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("demomicroservice", req.NamespacedName)

	dms := &devopsv1.DemoMicroService{}
	if err := r.Get(ctx, req.NamespacedName, dms); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			log.Info("此时没有找到对应的 DemoMicroService resource, 即此处进入了 resource 被删除成功后的生命周期")
			return ctrl.Result{}, nil
		} else {
			log.Error(err, "不是未找到的错误，那么就是意料之外的错误，所以这里直接返回错误")
			return ctrl.Result{}, err
		}
	}

	log.Info("走到这里意味着 DemoMicroService resource 被找到，即该 resource 被成功创建，进入到了可根据该 resource 来执行逻辑的主流程")
	podLabels := map[string]string{
		"app": req.Name,
	}
	deployment := appv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Spec: appv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            req.Name,
							Image:           dms.Spec.Image,
							ImagePullPolicy: "Always",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 18080,
								},
							},
						},
					},
				},
			},
		},
	}
	if err := r.Create(ctx, &deployment); err != nil {
		log.Error(err, "创建 Deployment resource 出错")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *DemoMicroServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1.DemoMicroService{}).
		Complete(r)
}
