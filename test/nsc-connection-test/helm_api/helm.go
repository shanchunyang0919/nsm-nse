package helm_api

import(
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
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
	HELM_DRIVER = "HELM_DRIVER"
)

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
	//ChartName string
	Namespace string
}

func initRelease() (r *Release){
	return &Release{


	}

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
	if err := act.Init(settings.RESTClientGetter(), namespace, os.Getenv(HELM_DRIVER), log.Printf); err != nil{
		log.Fatal("#{err}")
	}
	return act
}

/*
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
*/
