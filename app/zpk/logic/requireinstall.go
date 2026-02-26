package logic

import (
	"errors"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/app/zpk/logic/types"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	helm "gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk"
	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RequireInstallJob struct {
	app *types.PackageApp
	sdk *k8s.Sdk
}

func NewRequireInstallJob(app *types.PackageApp, sdk *k8s.Sdk) *RequireInstallJob {
	return &RequireInstallJob{
		app: app,
		sdk: sdk,
	}
}

func (r *RequireInstallJob) Run() error {
	mnames := r.app.GetOutModuleNames()
	for _, mname := range mnames {
		r.RunByModule(mname, r.app.Namespace, r.app.ReleaseName)
	}
	return nil
}

func (r *RequireInstallJob) getPackage(identifie string) (*types.ManifestPackage, error) {
	repoUrl := "https://zpk.w7.cc/respo/info/" + strings.ReplaceAll(identifie, "-", "_")
	repo := NewRepo(repoUrl, "", "")

	mPackage, err := repo.Load()
	if err != nil {
		return nil, errors.New("package not found")
	}
	return mPackage, nil

}

func (r *RequireInstallJob) RunByModule(identifie string, namespace string, releaseName string) error {

	requirePackage, err := r.getPackage(identifie)
	if err != nil {
		return errors.New("package not found")
	}
	shell := requirePackage.Manifest.GetShellByType("requireinstall")
	if shell == nil {
		return errors.New("requireInstall not found")
	}
	deploymentApps, err := r.sdk.GetDeploymentAppByIdentifie(namespace, strings.ReplaceAll(identifie, "_", "-"))
	if err != nil {
		return errors.New("deployment not found")
	}
	if len(deploymentApps.Items) > 0 {
		deployment := deploymentApps.Items[0]
		deploymentReleaseName := deployment.GetLabels()["w7.cc/release-name"]
		job := r.toJob(requirePackage, deployment, identifie, shell, namespace, deploymentReleaseName)
		if job == nil {
			return errors.New("job not found")
		}
		_, err = r.sdk.ClientSet.BatchV1().Jobs(namespace).Create(r.sdk.Ctx, job, metav1.CreateOptions{})
		if err != nil {
			return errors.New("job create error")
		}
	}
	return nil
}

func (r *RequireInstallJob) toJob(mPackage *types.ManifestPackage, deployment v1.Deployment, identifie string, shell helm.ManifestShellInterface, namespace string, releaseName string) *batchv1.Job {

	_, dbName := r.app.RequireCreateDb()
	_, dbuser, dbpassword := r.app.RequireCreateDbUser()
	envKv := []types.EnvKv{}
	for _, v := range deployment.Spec.Template.Spec.Containers[0].Env {
		ek := types.EnvKv{
			Name:  v.Name,
			Value: v.Value,
		}
		envKv = append(envKv, ek)
	}
	installOption := &types.InstallOption{
		Namespace:   namespace,
		Identifie:   identifie,
		EnvKv:       envKv,
		ReleaseName: releaseName,
		PvcName:     "default-volume",
	}

	packageApp := types.NewPackageApp(mPackage, installOption)
	packageApp.Manifest.Platform.Container.Env = append(packageApp.Manifest.Platform.Container.Env, types.Env{Name: "DB_NAME", Value: dbName})
	packageApp.Manifest.Platform.Container.Env = append(packageApp.Manifest.Platform.Container.Env, types.Env{Name: "HOST", Value: string(deployment.GetName())})

	if dbuser != "" && dbpassword != "" {
		packageApp.Manifest.Platform.Container.Env = append(packageApp.Manifest.Platform.Container.Env, types.Env{Name: "DB_USERNAME", Value: dbuser})
		packageApp.Manifest.Platform.Container.Env = append(packageApp.Manifest.Platform.Container.Env, types.Env{Name: "DB_PASSWORD", Value: dbpassword})
	}

	job := helm.ToShellJob(packageApp, shell)
	return job
}

func RequireInstall(secretName string, namespace string, releaseName string, dbName string, dbuser string, dbpassword string) error {
	sdk := k8s.NewK8sClient()

	secret, err := sdk.ClientSet.CoreV1().Secrets(namespace).Get(sdk.Ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return errors.New("secret not found")
	}
	moduleName, ok := secret.Labels["w7.cc/identifie"]
	if !ok {
		return errors.New("identifier not found")
	}
	repoUrl := "https://zpk.w7.cc/respo/info/" + strings.ReplaceAll(moduleName, "-", "_")
	repo := NewRepo(repoUrl, "", "")

	mPackage, err := repo.Load()
	if err != nil {
		return errors.New("package not found")
	}
	shell := mPackage.Manifest.GetShellByType("requireinstall")
	if shell == nil {
		return errors.New("requireInstall not found")
	}
	envKv := []types.EnvKv{}
	for k, v := range secret.Data {
		ek := types.EnvKv{
			Name:  k,
			Value: string(v),
		}
		envKv = append(envKv, ek)
	}
	installOption := &types.InstallOption{
		Namespace:   namespace,
		Identifie:   moduleName,
		EnvKv:       envKv,
		ReleaseName: moduleName,
		PvcName:     "default-volume",
	}

	packageApp := types.NewPackageApp(mPackage, installOption)
	packageApp.Manifest.Platform.Container.Env = append(packageApp.Manifest.Platform.Container.Env, types.Env{Name: "DB_NAME", Value: dbName})
	packageApp.Manifest.Platform.Container.Env = append(packageApp.Manifest.Platform.Container.Env, types.Env{Name: "HOST", Value: string(secret.Data["HOST"])})

	if dbuser != "" && dbpassword != "" {
		packageApp.Manifest.Platform.Container.Env = append(packageApp.Manifest.Platform.Container.Env, types.Env{Name: "DB_USERNAME", Value: dbuser})
		packageApp.Manifest.Platform.Container.Env = append(packageApp.Manifest.Platform.Container.Env, types.Env{Name: "DB_PASSWORD", Value: dbpassword})
	}

	job := helm.ToShellJob(packageApp, shell)
	if job == nil {
		return errors.New("job not found")
	}
	// data, err := yaml.Marshal(job)
	// if err != nil {
	// 	return errors.New("job marshal failed")
	// }
	// data2 := (string(data))
	// os.WriteFile("job.yaml", []byte(data2), 0644)
	// println(data2)

	// return nil
	job, err = sdk.ClientSet.BatchV1().Jobs(namespace).Create(sdk.Ctx, job, metav1.CreateOptions{})

	if err != nil {
		return errors.New("job create failed")
	}
	return nil

}
