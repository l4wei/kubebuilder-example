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
	"my.domain/example/util"
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

const (
	demoMicroServiceFinalizer string = "demomicroservice.finalizers.devops.my.domain"
)

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

	if dms.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("进入到 apply 这个 DemoMicroService CR 的逻辑")
		log.Info("此时必须确保 resource 的 finalizers 里有控制器指定的 finalizer")
		if !util.ContainsString(dms.ObjectMeta.Finalizers, demoMicroServiceFinalizer) {
			dms.ObjectMeta.Finalizers = append(dms.ObjectMeta.Finalizers, demoMicroServiceFinalizer)
			if err := r.Update(ctx, dms); err != nil {
				return ctrl.Result{}, err
			}
		}

		if _, err := r.applyDeployment(ctx, req, dms); err != nil {
			return ctrl.Result{}, nil
		}
	} else {
		log.Info("进入到删除这个 DemoMicroService CR 的逻辑")
		if util.ContainsString(dms.ObjectMeta.Finalizers, demoMicroServiceFinalizer) {
			log.Info("如果 finalizers 被清空，则该 DemoMicroService CR 就已经不存在了，所以必须在次之前删除 Deployment")
			if err := r.cleanDeployment(ctx, req); err != nil {
				return ctrl.Result{}, nil
			}
		}
		log.Info("清空 finalizers，在此之后该 DemoMicroService CR 才会真正消失")
		dms.ObjectMeta.Finalizers = util.RemoveString(dms.ObjectMeta.Finalizers, demoMicroServiceFinalizer)
		if err := r.Update(ctx, dms); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *DemoMicroServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1.DemoMicroService{}).
		Complete(r)
}

func (r *DemoMicroServiceReconciler) applyDeployment(ctx context.Context, req ctrl.Request, dms *devopsv1.DemoMicroService) (*appv1.Deployment, error) {
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
	oldDeployment := &appv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, oldDeployment); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// 如果 deployment 不存在，则创建它
			if err := r.Create(ctx, &deployment); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	// 此时表示 deployment 已经存在，则更新它
	if err := r.Update(ctx, &deployment); err != nil {
		return nil, err
	}
	return &deployment, nil
}

func (r *DemoMicroServiceReconciler) cleanDeployment(ctx context.Context, req ctrl.Request) error {
	deployment := &appv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, deployment); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// 既然已经没了，do nothing
			return nil
		} else {
			return err
		}
	}
	if err := r.Delete(ctx, deployment); err != nil {
		return err
	}
	return nil
}
