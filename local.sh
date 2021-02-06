#!/usr/bin/env bash

end() {
    minikube delete
}

trap end exit

minikube start --kubernetes-version=v1.17.13 --driver=virtualbox

kubectl create namespace istio-system

istioctl operator init --watchedNamespaces=istio-system --tag 1.8.1

kubectl apply -f - <<EOF
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  namespace: istio-system
  name: oidc-istiocontrolplane
spec:
  profile: default
EOF

while true; do
    if [ "$(kubectl get svc istio-ingressgateway -n istio-system -o=jsonpath='{.spec.clusterIP}' 2>/dev/null)" != "" ]; then
        break
    fi
    sleep 1
done

MINIKUBE_IP=$(kubectl get svc istio-ingressgateway -n istio-system -o=jsonpath='{.spec.clusterIP}')

cat <<EOF >/usr/local/etc/lcl/base/hosts.txt
/oidc.l$/     $MINIKUBE_IP
/^[^.]*$/   127.0.0.1
*.localhost 127.0.0.1
EOF

# install -> https://github.com/ken109/lcl
lcl base stop
lcl base start dns

kubectl create namespace oidc

kubectl config set-context "$(kubectl config current-context)" --namespace=oidc

kubectl label namespace oidc istio-injection=enabled

skaffold run

minikube tunnel
