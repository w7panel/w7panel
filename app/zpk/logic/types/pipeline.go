package types

func NewPipeline(app *PackageApp) *Pipeline {
	return &Pipeline{}
}

type Pipeline struct {
}

func (p *Pipeline) ToJob() {

}

// func (p *Pipeline) ToCreateSiteContainer() *corev1.Container {
// 	return &corev1.Container{
// 		{
// 			Name:            "docker-build",
// 			Image:           "ccr.ccs.tencentyun.com/afan/k8s-offline:1.0.6",
// 			Env:             []corev1.EnvVar{
// 				{
// 					Name:  "THIRDPARTY_CD_TOKEN",
// 					Value: p.app.Option.ThirdpartyCDToken,
// 				},
// 				{
// 					Name:  "DOMAIN_HOST",
// 					Value: p.app.Option.DomainHost,
// 				},
// 			},
// 			ImagePullPolicy: corev1.PullAlways,
// 			Command:         []string{"/kaniko/start.sh"}
// 		}
// 	}
// }

// func (p *Pipeline) ToShellContainer() *corev1.Container {
// 	return &corev1.Container{
// 		{
// 			Name:            "shell",
// 			Image:           p.app.GetImage(),
// 			Env:             []corev1.EnvVar{
// 				{
// 					Name:  "DOMAIN_URL",
// 					Value: p.app.Option.DomainHost,
// 				},
// 			},
// 			ImagePullPolicy: corev1.PullAlways,
// 			Command:         []string{"/start.sh"}
// 		}
// 	}
// }

// func (p *Pipeline) ToBuildContainer() *corev1.Container {
// 	return &corev1.Container{
// 		{
// 			Name:            "docker-build",
// 			Image:           "ccr.ccs.tencentyun.com/afan-public/kaniko:w7console-new5-15",
// 			Env:             option.ToEnv(),
// 			WorkingDir:      "/workspace",
// 			ImagePullPolicy: corev1.PullAlways,
// 			Command:         []string{"/kaniko/start.sh"},
// 			VolumeMounts:    option.GetVolumeMounts()
// 		}
// 	}
// }
