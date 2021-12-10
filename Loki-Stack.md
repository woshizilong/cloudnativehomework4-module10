# Loki-Stack Helm Chart

安装 grafana repo

```shell
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
```

预览安装 loki-stack

```shell
helm upgrade --dry-run --install loki grafana/loki-stack --set grafana.enabled=true,prometheus.enabled=true,prometheus.alertmanager.persistentVolume.enabled=false,prometheus.server.persistentVolume.enabled=false
```

如同孟老师的说明报了如下错误

```shell
Release "loki" does not exist. Installing it now.
Error: unable to build kubernetes objects from release manifest: [unable to recognize "": no matches for kind "ClusterRole" in version "rbac.authorization.k8s.io/v1beta1", unable to recognize "": no matches for kind "ClusterRoleBinding" in version "rbac.authorization.k8s.io/v1beta1"]
```

手工下载吧

```shell
helm pull grafana/loki-stack
tar -xvf loki-stack-2.5.0.tgz
```

替换版本号

```shell
cd loki-stack
grep -rl 'rbac.authorization.k8s.io/v1beta1' ./ --include '*.yaml'  | xargs sed -i "" s#rbac.authorization.k8s.io/v1beta1#rbac.authorization.k8s.io/v1#g
```

预览安装本地 loki-stack

```shell
helm upgrade --dry-run --install loki ./loki-stack --set grafana.enabled=true,prometheus.enabled=true,prometheus.alertmanager.persistentVolume.enabled=false,prometheus.server.persistentVolume.enabled=false
```

安装本地 loki-stack

```shell
helm upgrade --install loki ./loki-stack --set grafana.enabled=true,prometheus.enabled=true,prometheus.alertmanager.persistentVolume.enabled=false,prometheus.server.persistentVolume.enabled=false

# 输出如下
Release "loki" does not exist. Installing it now.
W1210 19:25:00.814577   92959 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
W1210 19:25:00.819172   92959 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
W1210 19:25:00.824773   92959 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
W1210 19:25:00.828311   92959 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
W1210 19:25:01.208443   92959 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
W1210 19:25:01.214686   92959 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
W1210 19:25:01.214691   92959 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
W1210 19:25:01.217663   92959 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
NAME: loki
LAST DEPLOYED: Fri Dec 10 19:24:59 2021
NAMESPACE: default
STATUS: deployed
REVISION: 1
NOTES:
The Loki stack has been deployed to your cluster. Loki can now be added as a datasource in Grafana.

See http://docs.grafana.org/features/datasources/loki/ for more detail.
```

查看 grafana 的登录密码

```shell
kubectl get secret --namespace default loki-grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo
```

设定 kubeproxy 访问 [grafana](http://localhost:3000/login)

```shell
kubectl port-forward --namespace default service/loki-grafana 3000:80
```

导入Dashboard

- [Loki stack monitoring (Promtail, Loki)](https://grafana.com/grafana/dashboards/14055)
- [Kubernetes Cluster (Prometheus)](https://grafana.com/grafana/dashboards/6417)

设定 kubeproxy 访问 [prometheus](http://localhost:9090/graph)

```shell
kubectl port-forward --namespace default service/loki-prometheus-server 9090:80
```
