package test_helper

import(
	"bytes"
	"log"
	"regexp"

	docker "github.com/fsouza/go-dockerclient"
	kubeapi "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/clientgo"
)

const(
	imageName = "busybox:1.28"
	//vl3NSELabel = "networkservicemesh.io/app=vl3-nse-vl3-service"

)

var(
	//64 bytes from 10.244.1.10: icmp_seq=1 ttl=63 time=0.068 ms
	regExp = regexp.MustCompile("\n([0-9]+) bytes from")
)

// simple docker helps making pings from specific container under particular namespace
type DockerAgent struct{
	ContainerID string
	DestIPs []string
	dockerClient *docker.Client
}

func getPodIPList(kClient *kubeapi.KubernetesClientEndpoint, labels string) []string{
	podList := kClient.GetPodList(labels)
	IPList := make([]string, 0)
	for _, pod := range podList.Items{
		IPList = append(IPList, kClient.GetPodIP(pod.Name))
	}
	return IPList
}

func (d *DockerAgent) Ping(destIP string) bool{

	stdout, err := d.execCommand("ping", destIP)
	if err != nil{
		log.Fatal(err)
	}

	match := regExp.FindString(stdout)
	if match == "" {
		// cannot match
		return false
	}
	return true
}



func InitDockerAgent(kClient *kubeapi.KubernetesClientEndpoint, podName string, imageName string, labels string) (d *DockerAgent){
	// connect to the docker daemon
	var err error
	log.Print("herhehrhehrehrherhehr")
	d.dockerClient, err = docker.NewClientFromEnv()
	if err != nil {
		log.Fatalf("failed to get docker client instance from the environment variables: %v", err)
	}
	log.Print("herhehrhehrehrherhehr")
	d.ContainerID = kClient.GetContainerID(podName, imageName)
	d.DestIPs = getPodIPList(kClient, labels)

	return d
}

// reference from microservice.go, this functions exec into the container and ping to the dest IP
func (d *DockerAgent) execCommand(command string, destIP string) (output string, err error){
	exec, err := d.dockerClient.CreateExec(docker.CreateExecOptions{
		AttachStdout: true,
		Cmd: append([]string{command}, destIP),
		Container: d.ContainerID,
	})
	if err != nil{
		log.Fatal(err)
	}
	var stdout bytes.Buffer
	err = d.dockerClient.StartExec(exec.ID, docker.StartExecOptions{
		OutputStream: &stdout,
	})
	return stdout.String(), err
}