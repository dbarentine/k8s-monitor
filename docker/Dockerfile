FROM golang:1.8.3-alpine

# https://github.com/dbarentine/k8s-monitor
ADD bin/ /opt/k8s-monitor

ENTRYPOINT ["/opt/k8s-monitor/k8s-monitor"]