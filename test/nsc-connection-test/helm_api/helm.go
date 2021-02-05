package helm_api

import (
	"log"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

const (
	HELMDRIVER = "HELM_DRIVER"
)

type Util interface {
	InstallChart() error
	UninstallChart()
	ReinstallChart()
}


type HelmClientEndpoint struct{
	Release *Release
	Chart *chart.Chart
	Values *map[string]interface{}
	Actions *action.Configuration
}

// stores helm release info
type Release struct{
	ReleaseName string
	ChartPath string
	Namespace string
}

func (hc *HelmClientEndpoint) InstallChart() error{
	client := action.NewInstall(hc.Actions)
	client.ReleaseName = hc.Release.ReleaseName
	client.Namespace = hc.Release.ChartPath

	rel, err := client.Run(hc.Chart, *hc.Values)
	if err != nil{
		return err

	}
	log.Printf("installed Chart from path %s in namespace %s...\n", rel.Name, rel.Namespace)
	return nil

}

func (hc *HelmClientEndpoint) UninstallChart(){
	client := action.NewUninstall(hc.Actions)
	_, err := client.Run(hc.Release.ReleaseName)
	if err != nil{
		log.Fatalf("error uninstall release %s", hc.Release.ReleaseName)
	}
}


func (hc *HelmClientEndpoint) ReinstallChart(){
	hc.UninstallChart()
	hc.InstallChart()
}


func InitHelmClientEndpoint(values *map[string]interface{}, r *Release) *HelmClientEndpoint{
	return &HelmClientEndpoint{
		Release: r,
		Chart:  createChart(r.ChartPath),
		Values: values,
		Actions: initActionConfig(r.Namespace),
	}
}


func CreateValues(svcname string, rc string, wttime string) *map[string]interface{}{
	return &map[string]interface{}{
		"nsm": map[string]interface{}{
			"serviceName": svcname,
		},
		"restartWaitTime": wttime,
		"replicaCount": rc,
	}
}

func CreateReleaseInfo(releaseName string, chartPath string, namespace string) *Release{
	return &Release{
		ReleaseName: releaseName,
		ChartPath: chartPath,
		Namespace: namespace,
	}
}

func createChart(chartPath string) *chart.Chart{
	chart, err := loader.Load(chartPath)
	if err != nil{
		log.Fatalf("error creating chart with path %s", chartPath)
	}
	return chart
}


func initActionConfig(namespace string) (act *action.Configuration){
	settings := cli.New()
	act = new(action.Configuration)
	if err := act.Init(settings.RESTClientGetter(), namespace, os.Getenv(HELMDRIVER), log.Printf); err != nil{
		log.Fatal("#{err}")
	}
	return act
}

