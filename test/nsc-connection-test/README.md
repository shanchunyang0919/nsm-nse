# **NSC Connection Test**


This is a test validating the NSE behavior in scenarios where a NSM NetworkService client is repeatedly attempting to connect to the NetworkService (e.g. client crashlooping).

Creation of a test which continuously restarts client pods.  The parameters for iteration:
- Restart rate (or wait time between restarts)
- Restart iteration count
- Restart iteration time period (mutually exclusive from iteration count)

<img width="644" alt="Screen Shot 2021-02-12 at 11 03 30 AM" src="https://user-images.githubusercontent.com/71080192/107810764-08ddc780-6d22-11eb-9633-9ca49a4b86db.png">

Environment setup
```bash
$ ./run_nsc_tests.sh 
```

Environment setup with multiple NSEs. 
```bash
$ NSE=<number> ./run_nsc_tests.sh 
```

Command line (run full test)
```bash 
$ INIT=on LOG=on TIMEOUT=300 ./run_nsc_tests.sh 
```

Connectivty test only
```bash 
$ INIT=off LOG=on TIMEOUT=300 go test --short
```

Clean up
```bash
$ kind delete cluster --name kind-1-demo
```
