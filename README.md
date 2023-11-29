# Introduction

This is an oci hook to start AWS [mountpoint](https://github.com/awslabs/mountpoint-s3#readme).
Mountpoint allows mounting of S3 bucket on Linux. 

This oci hook is primarily meant to be used with Kata containers runtime to mount S3 bucket inside the Kata VM and making the mountpoint available to the pod.

## Details

To configure the hook as a Kata prestart hook, you'll need to copy the hook binary in the oci hook directory of the Kata VM image.
The default location is `/usr/share/oci/hooks/prestart`.
Further, you'll need to add a hook configuration file (default: `/usr/share/oci/hooks/mountpoint_hookconfig.json`).

The hook expects a configuration file which is a json. An example configuration file is shown below.
```
{
  "activation_flag": "HOOK",
  "program_path": "/usr/bin/mount-s3",
  "host_mountpoint": "/s3data",
  "container_mountpoint": "/s3data"
}
```

The value of the `activation_flag` (ie `HOOK` as shown above) needs to be provided as environment variable to the container.
The `container_mountpoint` can also be provided as an environment variable (`CONTAINER_MOUNTPOINT`) in the container using this hook.

Further, the mountpoint auth and other related parameters need to be provided as environment variables. Details of the mountpoint environment variables can be found [here](https://github.com/awslabs/mountpoint-s3#readme).

Following is an example Kubernetes pod using this hook.
```
apiVersion: v1
kind: Pod
metadata:
  name: test
  labels:
    app: test
spec:
  runtimeClassName: kata-remote
  containers:
    - name: ubuntu
      image: ubuntu
      command: ["sleep"]
      args: ["infinity"]
      env:      
        - name: HOOK
          value: "true"
        - name: AWS_ACCESS_KEY_ID
          valueFrom:
            secretKeyRef:
              name: s3-secret
              key: AWS_ACCESS_KEY_ID
        - name: AWS_SECRET_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: s3-secret
              key: AWS_SECRET_ACCESS_KEY
        - name: S3_BUCKET
          value: "my_bucket"          
```

## Build

```
GOOS=linux go build -o mountpoint_hook *.go
```

## Test

```
GOOS=linux go go test ./...
```

## Static checks


```
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...
```

```
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck --tests=false ./...
```