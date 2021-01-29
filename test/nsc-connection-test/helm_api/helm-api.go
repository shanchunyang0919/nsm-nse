package helm_api

import(
	"helm.sh/helm/v3/pkg/chart"
	"os"
	//"time"
	//"fmt"
	"log"

	"helm.sh/helm/v3/pkg/action"
	//"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	//"helm.sh/helm/v3/pkg/kube"
	//"helm.sh/helm/v3/pkg/release"

)

const (
	HelmDriver = "HELM_DRIVER"
	//defaultTimeout = 120
)

type ReleaseInfo struct {
	ReleaseName string
	ChartPath string
	//ChartName string
	Namespace string
	Chart *chart.Chart
	Values map[string]interface{}
}

func (r *ReleaseInfo) initActionConfig() (act *action.Configuration){
	settings := cli.New()
	act = new(action.Configuration)
	if err := act.Init(settings.RESTClientGetter(), r.Namespace, os.Getenv(HelmDriver), log.Printf); err != nil{
		log.Printf("#{err}")
		os.Exit(1)
	}
	return act
}


func (r *ReleaseInfo) InstallChart() error{
	actionConfig:= r.initActionConfig()
	client := action.NewInstall(actionConfig)
	client.ReleaseName = r.ReleaseName
	client.Namespace = r.Namespace
	rel, err := client.Run(r.Chart, r.Values)

	if err != nil{
		return err

	}
	log.Printf("Installed Chart from path: %s in namespace: %s\n", rel.Name, rel.Namespace)
	return nil
}