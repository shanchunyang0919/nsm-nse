apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox-{{ .Values.nsm.serviceName }}
  labels:
    version: v1
  #annotations:
   # ns.networkservicemesh.io: {{ .Values.nsm.serviceName }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      app: busybox-{{ .Values.nsm.serviceName }}
      version: v1
  template:
    metadata:
      labels:
        app: busybox-{{ .Values.nsm.serviceName }}
        version: v1
    spec:
      containers:
        - name: busybox
          image: busybox:1.28
          command:
            - sleep
            - "{{ .Values.restartWaitTime }}"
          imagePullPolicy: IfNotPresent
      restartPolicy: Always
