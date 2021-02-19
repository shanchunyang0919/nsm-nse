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

#### Test Introduction
- **Connectivity Test** - After the NSC bouncing, **exec** into NSC container and try to **ping** NSEs

#### Prerequisite
- **Go 1.14+**

#### Environmnet Variables (set by default)

| ENV VAR | Default       | Description |
| ---------- |:-------------:| -----:|
| `NSE` | `1`| Specify the numbers of NSEs to be deployed.|
| `LOG`      | `on` | Enable logging mode.|
| `TIMEMOUT` | `300`      |   Timeout flag for go test. The unit is **second** |
| `PING_LOG` | `on` | Enable to print out logs when NSC pods ping NSE pods.| 
| `NSE_LOG` | `0` | Prints out recent lines of Network Service Endpoint pod logs.|
| `NSMGR_LOG` | `0` | Prints out recent lines of Network Service Manager Pod logs.|


### Connectivity Test
Perform a ping command from NSC busybox container to the IP address if NSE `memif` interface. For example, according 
to the graph below, we will `exec` into the NSC pod container and execute `ping 192.168.22.2` command.
<img width="1870" alt="Screen Shot 2021-02-18 at 7 13 47 AM" src="https://user-images.githubusercontent.com/71080192/108377543-01eb0500-71b9-11eb-9f84-61826c42d5f4.png">

The next approach will be `exec` into one NSC pod and `ping` other NSC pod's `nsm0` IP Address.
<img width="703" alt="Screen Shot 2021-02-18 at 12 40 44 PM" src="https://user-images.githubusercontent.com/71080192/108421054-789ff680-71e9-11eb-9b43-ddb5294f9d89.png">


#### Demo
This will setup the environment, run tests, and clean up. 
```bash
$ make run-all 
```

##### Multiple NSEs setup
Specify the numbers of NSEs set up the environment.  
```bash
$ make NSE=<number> setup 
```

##### Print logs with different options. (After setup)
This test will print recent 20 lines of NSE logs, recent 5 lines of NSMGR logs, and no logs from the pinging test.
```bash
$ make NSE_LOG=20 NSMGR_LOG=5 PING_LOG=off run-test
```
This will disable the log mode.
```bash
$ make LOG=off run-test
```
Clean up kind cluster
```bash
$ make clean
```


#### Extras

**Run test with specific environment variables (Manual)**
```bash 
$ LOG=on TIMEOUT=300 NSE_LOG=20 NSMGR_LOG=5 PING_LOG=off go test
```

**Clean up (Manual)**
```bash
$ kind delete cluster --name kind-1-demo
```

## Reference
- **client-go**: https://github.com/kubernetes/client-go
- **Kubernetes logs CLI**: https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#logs
- **Busybox DockerHub**: https://hub.docker.com/_/busybox