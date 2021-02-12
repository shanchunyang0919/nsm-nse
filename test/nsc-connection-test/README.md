# **NSC Connection Test**


This is a test validating the NSE's behavior in scenarios where a NSM NetworkService client is repeatedly attempting 
to connect to the NetworkService (e.g. NSC **CrashLoopBackOff**).

Three factors deciding how **NSCs** are restarting. The **Restart Rate** is how long the **"busybox"** container is going to 
live. The **Starting Iteration Count** is the count of how many times the **NSC** pod is getting deleted, whereas **Restart 
Iteration Time Period** is how long the pod is going to live until it gets deleted. 
- **Restart Rate**
- **Restart Iteration Count**
- **Restart Iteration Time Period** (mutually exclusive from iteration count)

<img width="644" alt="Screen Shot 2021-02-12 at 11 03 30 AM" src="https://user-images.githubusercontent.com/71080192/107810764-08ddc780-6d22-11eb-9633-9ca49a4b86db.png">

####Test Introduction
- **Logs Test** - Capture and assert logs from vl3 **NSE** pods, we are taking last 20 lines of logs by default
- **Connectivity Test** - After the NSC bouncing, **exec** into NSC container and try to **ping** NSEs


####Prerequisite
- **Go 1.14+**

####Environmnet Variables (set by default)
- **LOG**=on (Logging Mode. It could be set to **off**) 
- **TIMEMOUT**=300 (Timeout flag for go test. The unit is **second**)
- **INIT**=on (After running the script, could be set to **off** to run **go test CLI**)
- **NSE**=1 (Specify the numbers of NSEs to be deployed.)

####Demo 
```bash
$ ./run_nsc_tests.sh 
```

#####Run tests with multiple NSEs 
```bash
$ NSE=<number> ./run_nsc_tests.sh 
```



####Extras

**Connectivty test only (need to run script first)**
```bash 
$ INIT=off LOG=on TIMEOUT=300 go test --short
```

**Clean up (Manual)**
```bash
$ kind delete cluster --name <cluster_name>
```

##Reference
- **client-go**: https://github.com/kubernetes/client-go
- **Kubernetes logs CLI**: https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#logs
- **Busybox DockerHub**: https://hub.docker.com/_/busybox